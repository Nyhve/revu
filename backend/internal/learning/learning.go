// Package learning implements learning projects and study sessions (/api/learning).
package learning

import (
	"github.com/gin-gonic/gin"

	"github.com/Im-Doh/revu/backend/internal/activity"
	"github.com/Im-Doh/revu/backend/internal/app"
	"github.com/Im-Doh/revu/backend/internal/auth"
	"github.com/Im-Doh/revu/backend/internal/httpx"
	"github.com/Im-Doh/revu/backend/internal/store"
)

type CreateProjectRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Topic       *string `json:"topic"`
}

func (r CreateProjectRequest) Validate() error {
	var v httpx.Validator
	v.NotBlank("title", r.Title, "Title must not be blank")
	v.MaxLen("title", r.Title, 255, "Title must not exceed 255 characters")
	v.NotBlank("topic", r.Topic, "Topic must not be blank")
	v.MaxLen("topic", r.Topic, 255, "Topic must not exceed 255 characters")
	return v.Err()
}

type CreateSessionRequest struct {
	DurationMinutes int         `json:"durationMinutes"`
	Notes           *string     `json:"notes"`
	Date            *httpx.Date `json:"date"`
}

func (r CreateSessionRequest) Validate() error {
	var v httpx.Validator
	v.Min("durationMinutes", r.DurationMinutes, 1, "Duration must be at least 1 minute")
	v.Max("durationMinutes", r.DurationMinutes, 100000, "Duration must not exceed 100000 minutes")
	v.Required("date", r.Date != nil, "Date must not be null")
	return v.Err()
}

type ProjectResponse struct {
	ID           int64               `json:"id"`
	Title        string              `json:"title"`
	Description  *string             `json:"description"`
	Topic        string              `json:"topic"`
	TotalMinutes int                 `json:"totalMinutes"`
	CreatedAt    httpx.LocalDateTime `json:"createdAt"`
}

type SessionResponse struct {
	ID              int64               `json:"id"`
	ProjectID       int64               `json:"projectId"`
	DurationMinutes int                 `json:"durationMinutes"`
	Notes           *string             `json:"notes"`
	Date            httpx.Date          `json:"date"`
	CreatedAt       httpx.LocalDateTime `json:"createdAt"`
}

func toSessionResponse(s store.LearningSession) SessionResponse {
	return SessionResponse{
		ID:              s.ID,
		ProjectID:       s.LearningProjectID,
		DurationMinutes: int(s.DurationMinutes),
		Notes:           s.Notes,
		Date:            httpx.Date(s.Date),
		CreatedAt:       httpx.LocalDateTime(s.CreatedAt),
	}
}

type Handler struct{ *app.Deps }

func Register(r *gin.RouterGroup, d *app.Deps) {
	h := Handler{d}
	g := r.Group("/learning")
	g.GET("", httpx.H(h.list))
	g.POST("", httpx.H(h.create))
	g.DELETE("/:id", httpx.H(h.delete))
	g.GET("/:id/sessions", httpx.H(h.sessions))
	g.POST("/:id/sessions", httpx.H(h.addSession))
	g.DELETE("/:id/sessions/:sessionId", httpx.H(h.deleteSession))
}

func (h Handler) requireProject(c *gin.Context, q *store.Queries, id int64) error {
	ok, err := q.LearningProjectExists(c, store.LearningProjectExistsParams{ID: id, UserID: auth.UserID(c)})
	if err != nil {
		return err
	}
	if !ok {
		return httpx.NotFound("Learning project", id)
	}
	return nil
}

func (h Handler) list(c *gin.Context) error {
	rows, err := h.Q.ListLearningProjects(c, auth.UserID(c))
	if err != nil {
		return err
	}
	out := make([]ProjectResponse, 0, len(rows))
	for _, p := range rows {
		out = append(out, ProjectResponse{
			ID:           p.ID,
			Title:        p.Title,
			Description:  p.Description,
			Topic:        p.Topic,
			TotalMinutes: int(p.TotalMinutes),
			CreatedAt:    httpx.LocalDateTime(p.CreatedAt),
		})
	}
	return httpx.OK(c, out)
}

func (h Handler) create(c *gin.Context) error {
	var req CreateProjectRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	p, err := h.Q.CreateLearningProject(c, store.CreateLearningProjectParams{
		UserID:      auth.UserID(c),
		Title:       *req.Title,
		Description: req.Description,
		Topic:       *req.Topic,
		Now:         h.Clock.Now(),
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, ProjectResponse{
		ID:           p.ID,
		Title:        p.Title,
		Description:  p.Description,
		Topic:        p.Topic,
		TotalMinutes: 0,
		CreatedAt:    httpx.LocalDateTime(p.CreatedAt),
	})
}

func (h Handler) delete(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	n, err := h.Q.DeleteLearningProject(c, store.DeleteLearningProjectParams{ID: id, UserID: auth.UserID(c)})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("Learning project", id)
	}
	return httpx.NoContent(c)
}

func (h Handler) sessions(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	if err := h.requireProject(c, h.Q, id); err != nil {
		return err
	}
	rows, err := h.Q.ListLearningSessions(c, id)
	if err != nil {
		return err
	}
	out := make([]SessionResponse, 0, len(rows))
	for _, s := range rows {
		out = append(out, toSessionResponse(s))
	}
	return httpx.OK(c, out)
}

func (h Handler) addSession(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	var req CreateSessionRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	uid := auth.UserID(c)
	now := h.Clock.Now()
	var s store.LearningSession
	err = h.InTx(c, func(q *store.Queries) error {
		if err := h.requireProject(c, q, id); err != nil {
			return err
		}
		var err error
		s, err = q.CreateLearningSession(c, store.CreateLearningSessionParams{
			ProjectID:       id,
			DurationMinutes: int32(req.DurationMinutes),
			Notes:           req.Notes,
			Date:            req.Date.Time(),
			Now:             now,
		})
		if err != nil {
			return err
		}
		return activity.Record(c, q, uid, activity.LearningSessionCompleted, id, now)
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toSessionResponse(s))
}

func (h Handler) deleteSession(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	sessionID, err := httpx.PathID(c, "sessionId")
	if err != nil {
		return err
	}
	if err := h.requireProject(c, h.Q, id); err != nil {
		return err
	}
	// The session must belong to this (owned) project.
	n, err := h.Q.DeleteLearningSession(c, store.DeleteLearningSessionParams{ID: sessionID, LearningProjectID: id})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("Learning session", sessionID)
	}
	return httpx.NoContent(c)
}
