// Package activity records user actions and serves the activity feed.
package activity

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Im-Doh/revu/backend/internal/app"
	"github.com/Im-Doh/revu/backend/internal/apptime"
	"github.com/Im-Doh/revu/backend/internal/auth"
	"github.com/Im-Doh/revu/backend/internal/httpx"
	"github.com/Im-Doh/revu/backend/internal/store"
)

type Type string

const (
	GoalCreated              Type = "GOAL_CREATED"
	GoalCompleted            Type = "GOAL_COMPLETED"
	TaskCreated              Type = "TASK_CREATED"
	TaskCompleted            Type = "TASK_COMPLETED"
	HabitCompleted           Type = "HABIT_COMPLETED"
	WorkoutCompleted         Type = "WORKOUT_COMPLETED"
	LearningSessionCompleted Type = "LEARNING_SESSION_COMPLETED"
)

// Record stores an activity. Call it with the transaction of the triggering change.
func Record(ctx context.Context, q *store.Queries, userID int64, t Type, referenceID int64, now time.Time) error {
	return q.CreateActivity(ctx, store.CreateActivityParams{
		UserID:      userID,
		Type:        string(t),
		ReferenceID: referenceID,
		CreatedAt:   now,
	})
}

type Response struct {
	ID          int64               `json:"id"`
	Type        string              `json:"type"`
	ReferenceID int64               `json:"referenceId"`
	CreatedAt   httpx.LocalDateTime `json:"createdAt"`
}

type DailyCount struct {
	Date  httpx.Date `json:"date"`
	Count int64      `json:"count"`
}

type Handler struct{ *app.Deps }

func Register(r *gin.RouterGroup, d *app.Deps) {
	h := Handler{d}
	g := r.Group("/activities")
	g.GET("", httpx.H(h.feed))
	g.GET("/daily", httpx.H(h.daily))
}

func (h Handler) feed(c *gin.Context) error {
	rows, err := h.Q.ListActivities(c, auth.UserID(c))
	if err != nil {
		return err
	}
	out := make([]Response, 0, len(rows))
	for _, a := range rows {
		out = append(out, Response{
			ID:          a.ID,
			Type:        a.Type,
			ReferenceID: a.ReferenceID,
			CreatedAt:   httpx.LocalDateTime(a.CreatedAt),
		})
	}
	return httpx.OK(c, out)
}

// daily returns the number of activities per calendar day (user time zone)
// for the last ?days=30 days, sorted by date; days without activity are omitted.
func (h Handler) daily(c *gin.Context) error {
	days, err := httpx.QueryInt(c, "days", 30, 0, 3660)
	if err != nil {
		return err
	}
	from := h.Clock.Now().AddDate(0, 0, -days)
	times, err := h.Q.ListActivityTimesSince(c, store.ListActivityTimesSinceParams{
		UserID:    auth.UserID(c),
		CreatedAt: from,
	})
	if err != nil {
		return err
	}
	return httpx.OK(c, groupByDay(h.Clock, times))
}

func groupByDay(clock *apptime.Clock, times []time.Time) []DailyCount {
	out := make([]DailyCount, 0)
	for _, t := range times { // sorted ascending by the query
		d := clock.LocalDate(t)
		if n := len(out); n > 0 && time.Time(out[n-1].Date).Equal(d) {
			out[n-1].Count++
			continue
		}
		out = append(out, DailyCount{Date: httpx.Date(d), Count: 1})
	}
	return out
}
