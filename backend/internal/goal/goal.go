// Package goal implements long-term goals (/api/goals).
package goal

import (
	"github.com/gin-gonic/gin"

	"github.com/Im-Doh/revu/backend/internal/activity"
	"github.com/Im-Doh/revu/backend/internal/app"
	"github.com/Im-Doh/revu/backend/internal/auth"
	"github.com/Im-Doh/revu/backend/internal/database"
	"github.com/Im-Doh/revu/backend/internal/httpx"
	"github.com/Im-Doh/revu/backend/internal/store"
)

const (
	StatusActive    = "ACTIVE"
	StatusCompleted = "COMPLETED"
	StatusArchived  = "ARCHIVED"
)

var statuses = []string{StatusActive, StatusCompleted, StatusArchived}

type CreateRequest struct {
	Title       *string     `json:"title"`
	Description *string     `json:"description"`
	TargetDate  *httpx.Date `json:"targetDate"`
}

func (r CreateRequest) Validate() error {
	var v httpx.Validator
	v.NotBlank("title", r.Title, "Title must not be blank")
	v.MaxLen("title", r.Title, 255, "Title must not exceed 255 characters")
	return v.Err()
}

type UpdateRequest struct {
	Title       *string     `json:"title"`
	Description *string     `json:"description"`
	TargetDate  *httpx.Date `json:"targetDate"`
	Status      *string     `json:"status"`
}

func (r UpdateRequest) Validate() error {
	var v httpx.Validator
	v.NotBlank("title", r.Title, "Title must not be blank")
	v.MaxLen("title", r.Title, 255, "Title must not exceed 255 characters")
	if v.Required("status", r.Status != nil, "Status must not be null") {
		v.OneOf("status", r.Status, statuses)
	}
	return v.Err()
}

type Response struct {
	ID          int64               `json:"id"`
	Title       string              `json:"title"`
	Description *string             `json:"description"`
	TargetDate  *httpx.Date         `json:"targetDate"`
	Status      string              `json:"status"`
	CreatedAt   httpx.LocalDateTime `json:"createdAt"`
	UpdatedAt   httpx.LocalDateTime `json:"updatedAt"`
}

func toResponse(g store.Goal) Response {
	return Response{
		ID:          g.ID,
		Title:       g.Title,
		Description: g.Description,
		TargetDate:  httpx.DatePtr(g.TargetDate),
		Status:      g.Status,
		CreatedAt:   httpx.LocalDateTime(g.CreatedAt),
		UpdatedAt:   httpx.LocalDateTime(g.UpdatedAt),
	}
}

type Handler struct{ *app.Deps }

func Register(r *gin.RouterGroup, d *app.Deps) {
	h := Handler{d}
	g := r.Group("/goals")
	g.GET("", httpx.H(h.list))
	g.GET("/:id", httpx.H(h.get))
	g.POST("", httpx.H(h.create))
	g.PUT("/:id", httpx.H(h.update))
	g.DELETE("/:id", httpx.H(h.delete))
}

func (h Handler) list(c *gin.Context) error {
	uid := auth.UserID(c)
	rows, err := h.Q.ListGoals(c, &uid)
	if err != nil {
		return err
	}
	out := make([]Response, 0, len(rows))
	for _, g := range rows {
		out = append(out, toResponse(g))
	}
	return httpx.OK(c, out)
}

func (h Handler) get(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	uid := auth.UserID(c)
	g, err := h.Q.GetGoal(c, store.GetGoalParams{ID: id, UserID: &uid})
	if database.IsNoRows(err) {
		return httpx.NotFound("Goal", id)
	}
	if err != nil {
		return err
	}
	return httpx.OK(c, toResponse(g))
}

func (h Handler) create(c *gin.Context) error {
	var req CreateRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	uid := auth.UserID(c)
	now := h.Clock.Now()
	var g store.Goal
	err := h.InTx(c, func(q *store.Queries) error {
		var err error
		g, err = q.CreateGoal(c, store.CreateGoalParams{
			Title:       *req.Title,
			Description: req.Description,
			TargetDate:  req.TargetDate.TimePtr(),
			Status:      StatusActive,
			UserID:      &uid,
			Now:         now,
		})
		if err != nil {
			return err
		}
		return activity.Record(c, q, uid, activity.GoalCreated, g.ID, now)
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toResponse(g))
}

func (h Handler) update(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	var req UpdateRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	uid := auth.UserID(c)
	now := h.Clock.Now()
	var g store.Goal
	err = h.InTx(c, func(q *store.Queries) error {
		old, err := q.GetGoal(c, store.GetGoalParams{ID: id, UserID: &uid})
		if database.IsNoRows(err) {
			return httpx.NotFound("Goal", id)
		}
		if err != nil {
			return err
		}
		g, err = q.UpdateGoal(c, store.UpdateGoalParams{
			Title:       *req.Title,
			Description: req.Description,
			TargetDate:  req.TargetDate.TimePtr(),
			Status:      *req.Status,
			Now:         now,
			ID:          id,
			UserID:      &uid,
		})
		if err != nil {
			return err
		}
		if old.Status == StatusActive && g.Status == StatusCompleted {
			return activity.Record(c, q, uid, activity.GoalCompleted, g.ID, now)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return httpx.OK(c, toResponse(g))
}

func (h Handler) delete(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	uid := auth.UserID(c)
	n, err := h.Q.DeleteGoal(c, store.DeleteGoalParams{ID: id, UserID: &uid})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("Goal", id)
	}
	return httpx.NoContent(c)
}
