package contracttest

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Users created by the Spring backend have "$2a$10$..." BCrypt hashes. This test
// stores such a hash produced by an independent implementation (PostgreSQL's
// pgcrypto) and logs in through the API.
func TestLoginWithExistingBCryptHash(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" || os.Getenv("CONTRACT_BASE_URL") != "" {
		t.Skip("needs TEST_DATABASE_URL (in-process mode)")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS pgcrypto"); err != nil {
		t.Skipf("pgcrypto not available: %v", err)
	}
	email := randomEmail()
	var hash string
	err = conn.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, crypt('legacy-password', gen_salt('bf', 10)))
		 RETURNING password_hash`, email).Scan(&hash)
	if err != nil {
		t.Fatal(err)
	}
	if hash[:7] != "$2a$10$" {
		t.Fatalf("unexpected hash format %q", hash)
	}
	a := anon(t)
	expect(t, a.post("/api/auth/login", map[string]string{"email": email, "password": "legacy-password"}), 200)
	expect(t, a.post("/api/auth/login", map[string]string{"email": email, "password": "wrong"}), 401)
}

func TestCORS(t *testing.T) {
	requireServer(t)
	req, _ := http.NewRequest(http.MethodOptions, baseURL+"/api/goals", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	res, err := httpc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("preflight status %d", res.StatusCode)
	}
	if got := res.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Allow-Origin = %q", got)
	}

	req, _ = http.NewRequest(http.MethodOptions, baseURL+"/api/goals", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", "GET")
	res, err = httpc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("foreign origin must not be allowed")
	}
}

func TestNotLeakingOtherUsersData(t *testing.T) {
	a := newUser(t)
	b := newUser(t)
	a.mustPost("/api/goals", map[string]any{"title": "geheim"})
	a.mustPost("/api/tasks", map[string]any{"title": "geheim"})
	a.mustPost("/api/habits", map[string]any{"name": "geheim", "frequency": "DAILY"})
	a.mustPost("/api/learning", map[string]any{"title": "geheim", "topic": "x"})
	a.mustPost("/api/finance", map[string]any{"amount": 1, "description": "geheim", "category": "FOOD", "type": "EXPENSE", "date": today()})
	a.mustPost("/api/workouts", map[string]any{"name": "geheim", "date": today()})
	for _, path := range []string{"/api/goals", "/api/tasks", "/api/habits", "/api/learning", "/api/finance", "/api/workouts", "/api/activities"} {
		if l := b.mustGet(path).list(t); len(l) != 0 {
			t.Fatalf("%s leaks data of another user: %v", path, l)
		}
	}
}
