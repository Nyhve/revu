package auth

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Im-Doh/revu/backend/internal/httpx"
)

const userIDKey = "auth.userID"

// Middleware rejects requests without a valid "Authorization: Bearer <jwt>" header
// with 401 and stores the user id for the handlers.
func Middleware(tokens *TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			httpx.WriteError(c, httpx.Unauthorized("Authentication required"))
			return
		}
		userID, err := tokens.Parse(strings.TrimSpace(token))
		if err != nil {
			httpx.WriteError(c, httpx.Unauthorized("Invalid or expired token"))
			return
		}
		c.Set(userIDKey, userID)
		c.Next()
	}
}

// UserID returns the authenticated user's id. Only valid behind Middleware.
func UserID(c *gin.Context) int64 {
	return c.MustGet(userIDKey).(int64)
}
