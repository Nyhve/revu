// Package user implements registration, login and the per-user feature toggles.
package user

import (
	"github.com/gin-gonic/gin"

	"github.com/Im-Doh/revu/backend/internal/app"
	"github.com/Im-Doh/revu/backend/internal/auth"
	"github.com/Im-Doh/revu/backend/internal/database"
	"github.com/Im-Doh/revu/backend/internal/httpx"
	"github.com/Im-Doh/revu/backend/internal/store"
)

type RegisterRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

func (r RegisterRequest) Validate() error {
	var v httpx.Validator
	if v.NotBlank("email", r.Email, "Email must not be blank") {
		v.Email("email", r.Email, "Email must be valid")
	}
	if v.NotBlank("password", r.Password, "Password must not be blank") {
		v.MinLen("password", r.Password, 8, "Password must be at least 8 characters long")
	}
	// bcrypt only considers the first 72 bytes; reject longer input explicitly.
	if r.Password != nil && len(*r.Password) > 72 {
		return httpx.BadRequest("password: Password must not exceed 72 bytes")
	}
	return v.Err()
}

type LoginRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

func (r LoginRequest) Validate() error {
	var v httpx.Validator
	v.NotBlank("email", r.Email, "Email must not be blank")
	v.NotBlank("password", r.Password, "Password must not be blank")
	return v.Err()
}

type Response struct {
	ID        int64               `json:"id"`
	Email     string              `json:"email"`
	CreatedAt httpx.LocalDateTime `json:"createdAt"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Preferences is used for both the request and the response body.
// Missing booleans in a request count as false (as with the Spring backend).
type Preferences struct {
	GoalsEnabled    bool `json:"goalsEnabled"`
	TasksEnabled    bool `json:"tasksEnabled"`
	HabitsEnabled   bool `json:"habitsEnabled"`
	FitnessEnabled  bool `json:"fitnessEnabled"`
	LearningEnabled bool `json:"learningEnabled"`
	FinanceEnabled  bool `json:"financeEnabled"`
}

func preferencesFrom(p store.UserPreference) Preferences {
	return Preferences{
		GoalsEnabled:    p.GoalsEnabled,
		TasksEnabled:    p.TasksEnabled,
		HabitsEnabled:   p.HabitsEnabled,
		FitnessEnabled:  p.FitnessEnabled,
		LearningEnabled: p.LearningEnabled,
		FinanceEnabled:  p.FinanceEnabled,
	}
}

type Handler struct{ *app.Deps }

// RegisterPublic mounts /auth (no authentication).
func RegisterPublic(r *gin.RouterGroup, d *app.Deps) {
	h := Handler{d}
	g := r.Group("/auth")
	g.POST("/register", httpx.H(h.register))
	g.POST("/login", httpx.H(h.login))
}

// Register mounts /user (authenticated).
func Register(r *gin.RouterGroup, d *app.Deps) {
	h := Handler{d}
	g := r.Group("/user")
	g.GET("/preferences", httpx.H(h.getPreferences))
	g.PUT("/preferences", httpx.H(h.updatePreferences))
}

func (h Handler) register(c *gin.Context) error {
	var req RegisterRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	email := *req.Email
	exists, err := h.Q.UserExistsByEmail(c, email)
	if err != nil {
		return err
	}
	if exists {
		return httpx.Conflict("Email already registered: %s", email)
	}
	hash, err := auth.HashPassword(*req.Password)
	if err != nil {
		return err
	}
	u, err := h.Q.CreateUser(c, store.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		Now:          h.Clock.Now(),
	})
	if database.IsUniqueViolation(err) {
		return httpx.Conflict("Email already registered: %s", email)
	}
	if err != nil {
		return err
	}
	return httpx.Created(c, Response{ID: u.ID, Email: u.Email, CreatedAt: httpx.LocalDateTime(u.CreatedAt)})
}

func (h Handler) login(c *gin.Context) error {
	var req LoginRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	invalid := httpx.Unauthorized("Invalid email or password")
	u, err := h.Q.GetUserByEmail(c, *req.Email)
	if database.IsNoRows(err) {
		// Burn comparable time so response timing does not reveal registered e-mails.
		auth.BurnPasswordCheck(*req.Password)
		return invalid
	}
	if err != nil {
		return err
	}
	if !auth.CheckPassword(u.PasswordHash, *req.Password) {
		return invalid
	}
	token, err := h.Tokens.Generate(u.ID, u.Email)
	if err != nil {
		return err
	}
	return httpx.OK(c, LoginResponse{Token: token})
}

func (h Handler) getPreferences(c *gin.Context) error {
	userID := auth.UserID(c)
	if err := h.Q.EnsurePreferences(c, userID); err != nil {
		return err
	}
	p, err := h.Q.GetPreferences(c, userID)
	if err != nil {
		return err
	}
	return httpx.OK(c, preferencesFrom(p))
}

func (h Handler) updatePreferences(c *gin.Context) error {
	var req Preferences
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	p, err := h.Q.UpsertPreferences(c, store.UpsertPreferencesParams{
		UserID:          auth.UserID(c),
		GoalsEnabled:    req.GoalsEnabled,
		TasksEnabled:    req.TasksEnabled,
		HabitsEnabled:   req.HabitsEnabled,
		FitnessEnabled:  req.FitnessEnabled,
		LearningEnabled: req.LearningEnabled,
		FinanceEnabled:  req.FinanceEnabled,
	})
	if err != nil {
		return err
	}
	return httpx.OK(c, preferencesFrom(p))
}
