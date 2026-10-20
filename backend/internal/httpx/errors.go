// Package httpx contains the HTTP plumbing shared by all modules:
// error responses, JSON decoding, validation and wire types.
package httpx

import (
	"fmt"
	"net/http"
)

// Error is an error that maps to an HTTP status and a client facing message.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

func NewError(status int, format string, args ...any) *Error {
	return &Error{Status: status, Message: fmt.Sprintf(format, args...)}
}

func BadRequest(format string, args ...any) *Error {
	return NewError(http.StatusBadRequest, format, args...)
}

func Unauthorized(format string, args ...any) *Error {
	return NewError(http.StatusUnauthorized, format, args...)
}

func Conflict(format string, args ...any) *Error {
	return NewError(http.StatusConflict, format, args...)
}

// NotFound builds the "<Entity> not found with id: <id>" message used by the API.
func NotFound(entity string, id int64) *Error {
	return NewError(http.StatusNotFound, "%s not found with id: %d", entity, id)
}
