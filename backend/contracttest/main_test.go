package contracttest

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Nyhve/revu/backend/internal/app"
	"github.com/Nyhve/revu/backend/internal/apptime"
	"github.com/Nyhve/revu/backend/internal/auth"
	"github.com/Nyhve/revu/backend/internal/database"
	"github.com/Nyhve/revu/backend/internal/server"
)

// TestMain starts the Go API in-process when only TEST_DATABASE_URL is given
// (a database migrated with Flyway), so `go test ./...` covers the full stack.
func TestMain(m *testing.M) {
	if baseURL == "" {
		if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
			stop, err := startInProcess(dsn)
			if err != nil {
				fmt.Fprintln(os.Stderr, "contracttest:", err)
				os.Exit(1)
			}
			code := m.Run()
			stop()
			os.Exit(code)
		}
	}
	os.Exit(m.Run())
}

func startInProcess(dsn string) (func(), error) {
	ctx := context.Background()
	pool, err := database.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := database.CheckSchema(ctx, pool, "14"); err != nil {
		pool.Close()
		return nil, err
	}
	tokens, err := auth.NewTokenService([]byte("contract-test-secret-key-with-at-least-32-bytes"), time.Hour)
	if err != nil {
		pool.Close()
		return nil, err
	}
	loc, _ := time.LoadLocation("Europe/Berlin")
	deps := app.New(pool, apptime.New(loc), tokens)
	srv := httptest.NewServer(server.New(deps, server.Options{AllowedOrigins: []string{"http://localhost:5173"}}))
	baseURL = srv.URL
	return func() {
		srv.Close()
		pool.Close()
	}, nil
}
