// Package task implements short-term to-dos (/api/tasks).
package task

import (
	"github.com/gin-gonic/gin"

	"github.com/Nyhve/revu/backend/internal/activity"
	"github.com/Nyhve/revu/backend/internal/app"
	"github.com/Nyhve/revu/backend/internal/auth"
	"github.com/Nyhve/revu/backend/internal/database"
	"github.com/Nyhve/revu/backend/internal/httpx"
	"github.com/Nyhve/revu/backend/internal/store"
)

type CreateRequest struct {
	Title       *string     `json:"title"`
	Description *string     `json:"description"`
	DueDate     *httpx.Date `json:"dueDate"`
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
	DueDate     *httpx.Date `json:"dueDate"`
	Completed   *bool       `json:"completed"`
}

func (r UpdateRequest) Validate() error {
	var v httpx.Validator
	v.NotBlank("title", r.Title, "Title must not be blank")
	v.MaxLen("title", r.Title, 255, "Title must not exceed 255 characters")
	v.Required("completed", r.Completed != nil, "Completed must not be null")
	return v.Err()
}

type Response struct {
	ID          int64               `json:"id"`
	Title       string              `json:"title"`
	Description *string             `json:"description"`
	DueDate     *httpx.Date         `json:"dueDate"`
	Completed   bool                `json:"completed"`
	CreatedAt   httpx.LocalDateTime `json:"createdAt"`
	UpdatedAt   httpx.LocalDateTime `json:"updatedAt"`
}

func toResponse(t store.Task) Response {
	return Response{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		DueDate:     httpx.DatePtr(t.DueDate),
		Completed:   t.Completed,
		CreatedAt:   httpx.LocalDateTime(t.CreatedAt),
		UpdatedAt:   httpx.LocalDateTime(t.UpdatedAt),
	}
}

type Handler struct{ *app.Deps }

func Register(r *gin.RouterGroup, d *app.Deps) {
	h := Handler{d}
	g := r.Group("/tasks")
	g.GET("", httpx.H(h.list))
	g.GET("/:id", httpx.H(h.get))
	g.POST("", httpx.H(h.create))
	g.PUT("/:id", httpx.H(h.update))
	g.DELETE("/:id", httpx.H(h.delete))
}

func (h Handler) list(c *gin.Context) error {
	uid := auth.UserID(c)
	rows, err := h.Q.ListTasks(c, &uid)
	if err != nil {
		return err
	}
	out := make([]Response, 0, len(rows))
	for _, t := range rows {
		out = append(out, toResponse(t))
	}
	return httpx.OK(c, out)
}

func (h Handler) get(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	uid := auth.UserID(c)
	t, err := h.Q.GetTask(c, store.GetTaskParams{ID: id, UserID: &uid})
	if database.IsNoRows(err) {
		return httpx.NotFound("Task", id)
	}
	if err != nil {
		return err
	}
	return httpx.OK(c, toResponse(t))
}

func (h Handler) create(c *gin.Context) error {
	var req CreateRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	uid := auth.UserID(c)
	now := h.Clock.Now()
	var t store.Task
	err := h.InTx(c, func(q *store.Queries) error {
		var err error
		t, err = q.CreateTask(c, store.CreateTaskParams{
			Title:       *req.Title,
			Description: req.Description,
			DueDate:     req.DueDate.TimePtr(),
			UserID:      &uid,
			Now:         now,
		})
		if err != nil {
			return err
		}
		return activity.Record(c, q, uid, activity.TaskCreated, t.ID, now)
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toResponse(t))
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
	var t store.Task
	err = h.InTx(c, func(q *store.Queries) error {
		old, err := q.GetTask(c, store.GetTaskParams{ID: id, UserID: &uid})
		if database.IsNoRows(err) {
			return httpx.NotFound("Task", id)
		}
		if err != nil {
			return err
		}
		t, err = q.UpdateTask(c, store.UpdateTaskParams{
			Title:       *req.Title,
			Description: req.Description,
			DueDate:     req.DueDate.TimePtr(),
			Completed:   *req.Completed,
			Now:         now,
			ID:          id,
			UserID:      &uid,
		})
		if err != nil {
			return err
		}
		if !old.Completed && t.Completed {
			return activity.Record(c, q, uid, activity.TaskCompleted, t.ID, now)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return httpx.OK(c, toResponse(t))
}

func (h Handler) delete(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	uid := auth.UserID(c)
	n, err := h.Q.DeleteTask(c, store.DeleteTaskParams{ID: id, UserID: &uid})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("Task", id)
	}
	return httpx.NoContent(c)
}
