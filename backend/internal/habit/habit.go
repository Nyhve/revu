// Package habit implements daily/weekly habits and their completion entries (/api/habits).
package habit

import (
	"math"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Im-Doh/revu/backend/internal/activity"
	"github.com/Im-Doh/revu/backend/internal/app"
	"github.com/Im-Doh/revu/backend/internal/auth"
	"github.com/Im-Doh/revu/backend/internal/database"
	"github.com/Im-Doh/revu/backend/internal/httpx"
	"github.com/Im-Doh/revu/backend/internal/store"
)

const (
	Daily  = "DAILY"
	Weekly = "WEEKLY"
)

var frequencies = []string{Daily, Weekly}

type CreateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Frequency   *string `json:"frequency"`
}

func (r CreateRequest) Validate() error {
	var v httpx.Validator
	v.NotBlank("name", r.Name, "Name must not be blank")
	v.MaxLen("name", r.Name, 255, "Name must not exceed 255 characters")
	if v.Required("frequency", r.Frequency != nil, "Frequency must not be null") {
		v.OneOf("frequency", r.Frequency, frequencies)
	}
	return v.Err()
}

type Response struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	Frequency   string              `json:"frequency"`
	CreatedAt   httpx.LocalDateTime `json:"createdAt"`
}

type EntryResponse struct {
	ID        int64               `json:"id"`
	HabitID   int64               `json:"habitId"`
	Date      httpx.Date          `json:"date"`
	CreatedAt httpx.LocalDateTime `json:"createdAt"`
}

type ConsistencyResponse struct {
	HabitID            int64   `json:"habitId"`
	Name               string  `json:"name"`
	Frequency          string  `json:"frequency"`
	TotalDays          int     `json:"totalDays"`
	CompletedDays      int     `json:"completedDays"`
	ConsistencyPercent float64 `json:"consistencyPercent"`
}

func toResponse(h store.Habit) Response {
	return Response{
		ID:          h.ID,
		Name:        h.Name,
		Description: h.Description,
		Frequency:   h.Frequency,
		CreatedAt:   httpx.LocalDateTime(h.CreatedAt),
	}
}

func toEntryResponse(e store.HabitEntry) EntryResponse {
	return EntryResponse{
		ID:        e.ID,
		HabitID:   e.HabitID,
		Date:      httpx.Date(e.Date),
		CreatedAt: httpx.LocalDateTime(e.CreatedAt),
	}
}

type Handler struct{ *app.Deps }

func Register(r *gin.RouterGroup, d *app.Deps) {
	h := Handler{d}
	g := r.Group("/habits")
	g.GET("", httpx.H(h.list))
	g.GET("/entries/today", httpx.H(h.todaysEntries))
	g.GET("/consistency", httpx.H(h.consistency))
	g.POST("", httpx.H(h.create))
	g.DELETE("/:id", httpx.H(h.delete))
	g.POST("/:id/entries", httpx.H(h.completeToday))
	g.GET("/:id/entries", httpx.H(h.entries))
	g.DELETE("/:id/entries/:entryId", httpx.H(h.deleteEntry))
}

func (h Handler) findHabit(c *gin.Context, q *store.Queries, id int64) (store.Habit, error) {
	habit, err := q.GetHabit(c, store.GetHabitParams{ID: id, UserID: auth.UserID(c)})
	if database.IsNoRows(err) {
		return habit, httpx.NotFound("Habit", id)
	}
	return habit, err
}

func (h Handler) list(c *gin.Context) error {
	rows, err := h.Q.ListHabits(c, auth.UserID(c))
	if err != nil {
		return err
	}
	out := make([]Response, 0, len(rows))
	for _, r := range rows {
		out = append(out, toResponse(r))
	}
	return httpx.OK(c, out)
}

func (h Handler) todaysEntries(c *gin.Context) error {
	rows, err := h.Q.ListHabitEntriesForUserOnDate(c, store.ListHabitEntriesForUserOnDateParams{
		UserID: auth.UserID(c),
		Date:   h.Clock.Today(),
	})
	if err != nil {
		return err
	}
	out := make([]EntryResponse, 0, len(rows))
	for _, e := range rows {
		out = append(out, toEntryResponse(e))
	}
	return httpx.OK(c, out)
}

func (h Handler) create(c *gin.Context) error {
	var req CreateRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	habit, err := h.Q.CreateHabit(c, store.CreateHabitParams{
		UserID:      auth.UserID(c),
		Name:        *req.Name,
		Description: req.Description,
		Frequency:   *req.Frequency,
		Now:         h.Clock.Now(),
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toResponse(habit))
}

func (h Handler) delete(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	n, err := h.Q.DeleteHabit(c, store.DeleteHabitParams{ID: id, UserID: auth.UserID(c)})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("Habit", id)
	}
	return httpx.NoContent(c)
}

// completeToday ticks a habit for today. A daily habit can be ticked once per
// day, a weekly habit once per ISO week (Monday to Sunday).
func (h Handler) completeToday(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	uid := auth.UserID(c)
	today := h.Clock.Today()
	now := h.Clock.Now()
	var entry store.HabitEntry
	err = h.InTx(c, func(q *store.Queries) error {
		habit, err := h.findHabit(c, q, id)
		if err != nil {
			return err
		}
		from, to, conflict := today, today, httpx.Conflict("Habit already completed today")
		if habit.Frequency == Weekly {
			from, to = WeekBounds(today)
			conflict = httpx.Conflict("Habit already completed this week")
		}
		done, err := q.HabitEntryExistsBetween(c, store.HabitEntryExistsBetweenParams{
			HabitID: id, FromDate: from, ToDate: to,
		})
		if err != nil {
			return err
		}
		if done {
			return conflict
		}
		entry, err = q.CreateHabitEntry(c, store.CreateHabitEntryParams{HabitID: id, Date: today, CreatedAt: now})
		if database.IsUniqueViolation(err) {
			return conflict
		}
		if err != nil {
			return err
		}
		return activity.Record(c, q, uid, activity.HabitCompleted, id, now)
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toEntryResponse(entry))
}

func (h Handler) entries(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	if _, err := h.findHabit(c, h.Q, id); err != nil {
		return err
	}
	rows, err := h.Q.ListHabitEntries(c, id)
	if err != nil {
		return err
	}
	out := make([]EntryResponse, 0, len(rows))
	for _, e := range rows {
		out = append(out, toEntryResponse(e))
	}
	return httpx.OK(c, out)
}

func (h Handler) deleteEntry(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	entryID, err := httpx.PathID(c, "entryId")
	if err != nil {
		return err
	}
	if _, err := h.findHabit(c, h.Q, id); err != nil {
		return err
	}
	// The entry must belong to this (owned) habit.
	n, err := h.Q.DeleteHabitEntry(c, store.DeleteHabitEntryParams{ID: entryID, HabitID: id})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("Habit entry", entryID)
	}
	return httpx.NoContent(c)
}

func (h Handler) consistency(c *gin.Context) error {
	days, err := httpx.QueryInt(c, "days", 28, 0, 3660)
	if err != nil {
		return err
	}
	from := h.Clock.Today().AddDate(0, 0, -(days - 1))
	rows, err := h.Q.HabitCompletionCounts(c, store.HabitCompletionCountsParams{
		UserID:   auth.UserID(c),
		FromDate: from,
	})
	if err != nil {
		return err
	}
	out := make([]ConsistencyResponse, 0, len(rows))
	for _, r := range rows {
		total := days
		if r.Frequency == Weekly {
			total = days / 7
		}
		out = append(out, ConsistencyResponse{
			HabitID:            r.ID,
			Name:               r.Name,
			Frequency:          r.Frequency,
			TotalDays:          total,
			CompletedDays:      int(r.CompletedDays),
			ConsistencyPercent: Percent(int(r.CompletedDays), total),
		})
	}
	return httpx.OK(c, out)
}

// Percent returns completed/total in percent, rounded half-up to one decimal
// and capped at 100 (identical to the former Java implementation).
func Percent(completed, total int) float64 {
	if total == 0 {
		return 0
	}
	p := math.Floor(float64(completed)*100.0/float64(total)*10.0+0.5) / 10.0
	return math.Min(p, 100.0)
}

// WeekBounds returns Monday and Sunday of the ISO week containing day.
func WeekBounds(day time.Time) (time.Time, time.Time) {
	offset := (int(day.Weekday()) + 6) % 7 // Monday = 0
	monday := day.AddDate(0, 0, -offset)
	return monday, monday.AddDate(0, 0, 6)
}
