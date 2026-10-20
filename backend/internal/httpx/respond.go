package httpx

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the error body of every failed request:
// {"timestamp": "...", "status": 404, "message": "..."}.
type ErrorResponse struct {
	Timestamp LocalDateTime `json:"timestamp"`
	Status    int           `json:"status"`
	Message   string        `json:"message"`
}

// WriteError renders err. Unknown errors become a logged 500.
func WriteError(c *gin.Context, err error) {
	var he *Error
	if !errors.As(err, &he) {
		slog.Error("unhandled error", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		he = &Error{Status: http.StatusInternalServerError, Message: "Internal server error"}
	}
	c.AbortWithStatusJSON(he.Status, ErrorResponse{
		Timestamp: LocalDateTime(time.Now().UTC()),
		Status:    he.Status,
		Message:   he.Message,
	})
}

// Handler is a gin handler that may return an error.
type Handler func(c *gin.Context) error

// H adapts a Handler to gin.
func H(h Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			WriteError(c, err)
		}
	}
}

// OK writes a 200 JSON response.
func OK(c *gin.Context, body any) error {
	c.JSON(http.StatusOK, body)
	return nil
}

// Created writes a 201 JSON response.
func Created(c *gin.Context, body any) error {
	c.JSON(http.StatusCreated, body)
	return nil
}

// NoContent writes an empty 204 response.
func NoContent(c *gin.Context) error {
	c.Status(http.StatusNoContent)
	return nil
}

// PathID parses a numeric path parameter.
func PathID(c *gin.Context, name string) (int64, error) {
	v, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		return 0, BadRequest("Invalid value for path variable '%s': %s", name, c.Param(name))
	}
	return v, nil
}

// PathInt parses a numeric path parameter as int.
func PathInt(c *gin.Context, name string) (int, error) {
	v, err := strconv.Atoi(c.Param(name))
	if err != nil {
		return 0, BadRequest("Invalid value for path variable '%s': %s", name, c.Param(name))
	}
	return v, nil
}

// QueryInt parses an optional integer query parameter within [min, max].
func QueryInt(c *gin.Context, name string, def, min, max int) (int, error) {
	raw, ok := c.GetQuery(name)
	if !ok || raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, BadRequest("Invalid value for parameter '%s': %s", name, raw)
	}
	if v < min || v > max {
		return 0, BadRequest("Parameter '%s' must be between %d and %d", name, min, max)
	}
	return v, nil
}
