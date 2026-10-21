// Command server runs the RevU REST API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata" // time zone database for minimal container images

	"github.com/gin-gonic/gin"

	"github.com/Im-Doh/revu/backend/internal/app"
	"github.com/Im-Doh/revu/backend/internal/apptime"
	"github.com/Im-Doh/revu/backend/internal/auth"
	"github.com/Im-Doh/revu/backend/internal/config"
	"github.com/Im-Doh/revu/backend/internal/database"
	"github.com/Im-Doh/revu/backend/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.IsDev() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if cfg.SkipSchemaCheck {
		slog.Warn("schema check disabled")
	} else if err := database.CheckSchema(ctx, pool, cfg.ExpectedSchemaVersion); err != nil {
		return err
	}

	tokens, err := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)
	if err != nil {
		return err
	}

	deps := app.New(pool, apptime.New(cfg.Location), tokens)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(deps, server.Options{AllowedOrigins: cfg.CORSAllowedOrigins}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening",
			"port", cfg.Port,
			"database", config.RedactedDatabaseURL(cfg.DatabaseURL),
			"jwt_alg", tokens.Algorithm(),
			"timezone", cfg.Location.String(),
		)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("shutting down")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
