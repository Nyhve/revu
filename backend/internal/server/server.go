// Package server wires all modules into one HTTP handler.
package server

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Nyhve/revu/backend/internal/activity"
	"github.com/Nyhve/revu/backend/internal/app"
	"github.com/Nyhve/revu/backend/internal/auth"
	"github.com/Nyhve/revu/backend/internal/finance"
	"github.com/Nyhve/revu/backend/internal/fitness"
	"github.com/Nyhve/revu/backend/internal/goal"
	"github.com/Nyhve/revu/backend/internal/habit"
	"github.com/Nyhve/revu/backend/internal/httpx"
	"github.com/Nyhve/revu/backend/internal/learning"
	"github.com/Nyhve/revu/backend/internal/task"
	"github.com/Nyhve/revu/backend/internal/user"
)

type Options struct {
	AllowedOrigins []string
}

func New(d *app.Deps, opts Options) http.Handler {
	r := gin.New()
	r.ContextWithFallback = true // request cancellation reaches database calls
	r.HandleMethodNotAllowed = true
	r.RedirectTrailingSlash = false
	r.Use(recovery(), requestLog(), cors(opts.AllowedOrigins))

	r.NoRoute(func(c *gin.Context) {
		httpx.WriteError(c, httpx.NewError(http.StatusNotFound, "No endpoint %s %s", c.Request.Method, c.Request.URL.Path))
	})
	r.NoMethod(func(c *gin.Context) {
		httpx.WriteError(c, httpx.NewError(http.StatusMethodNotAllowed, "Request method '%s' is not supported", c.Request.Method))
	})

	r.GET("/healthz", func(c *gin.Context) {
		if err := d.Pool.Ping(c); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	api := r.Group("/api")
	user.RegisterPublic(api, d)

	protected := api.Group("", auth.Middleware(d.Tokens))
	user.Register(protected, d)
	goal.Register(protected, d)
	task.Register(protected, d)
	activity.Register(protected, d)
	habit.Register(protected, d)
	fitness.Register(protected, d)
	learning.Register(protected, d)
	finance.Register(protected, d)

	return r
}

// cors allows the configured origins for /api, like the former Spring config.
func cors(allowed []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" || !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}
		h := c.Writer.Header()
		h.Add("Vary", "Origin")
		h.Add("Vary", "Access-Control-Request-Method")
		h.Add("Vary", "Access-Control-Request-Headers")
		if !slices.Contains(allowed, origin) {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
			return
		}
		h.Set("Access-Control-Allow-Origin", origin)
		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			h.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
			if req := c.GetHeader("Access-Control-Request-Headers"); req != "" {
				h.Set("Access-Control-Allow-Headers", req)
			}
			h.Set("Access-Control-Max-Age", "1800")
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

func recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		slog.Error("panic", "path", c.Request.URL.Path, "error", err)
		httpx.WriteError(c, httpx.NewError(http.StatusInternalServerError, "Internal server error"))
	})
}

func requestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if c.Request.URL.Path == "/healthz" {
			return
		}
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}
