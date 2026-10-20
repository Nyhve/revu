// Package contracttest is a black-box test suite for the REST API.
//
// It talks HTTP only, so it can verify any implementation of the API:
//
//	CONTRACT_BASE_URL=http://localhost:8080 go test ./contracttest/ -count=1
//
// Set CONTRACT_LEGACY=1 when running against the former Spring backend: tests
// for behaviour that was intentionally fixed in the Go rewrite are skipped then.
package contracttest

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	baseURL = strings.TrimRight(os.Getenv("CONTRACT_BASE_URL"), "/")
	legacy  = os.Getenv("CONTRACT_LEGACY") == "1"
	httpc   = &http.Client{Timeout: 15 * time.Second}
)

func requireServer(t *testing.T) {
	t.Helper()
	if baseURL == "" {
		t.Skip("CONTRACT_BASE_URL not set")
	}
}

// fixed marks a test that checks a bug fix of the Go rewrite.
func fixed(t *testing.T) {
	t.Helper()
	if legacy {
		t.Skip("behaviour intentionally changed in the rewrite")
	}
}

type resp struct {
	Status int
	Body   []byte
}

func (r resp) json(t *testing.T) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(r.Body, &m); err != nil {
		t.Fatalf("response is not a JSON object: %v\n%s", err, r.Body)
	}
	return m
}

func (r resp) list(t *testing.T) []map[string]any {
	t.Helper()
	var l []map[string]any
	if err := json.Unmarshal(r.Body, &l); err != nil {
		t.Fatalf("response is not a JSON array: %v\n%s", err, r.Body)
	}
	if l == nil {
		t.Fatalf("expected [] but got %s", r.Body)
	}
	return l
}

type client struct {
	t     *testing.T
	token string
	email string
}

func anon(t *testing.T) *client { return &client{t: t} }

func (c *client) do(method, path string, body any) resp {
	c.t.Helper()
	var rd io.Reader
	if body != nil {
		var b []byte
		switch v := body.(type) {
		case string:
			b = []byte(v)
		default:
			var err error
			if b, err = json.Marshal(v); err != nil {
				c.t.Fatal(err)
			}
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, baseURL+path, rd)
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := httpc.Do(req)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	return resp{Status: res.StatusCode, Body: data}
}

func (c *client) get(path string) resp         { c.t.Helper(); return c.do("GET", path, nil) }
func (c *client) post(path string, b any) resp { c.t.Helper(); return c.do("POST", path, b) }
func (c *client) put(path string, b any) resp  { c.t.Helper(); return c.do("PUT", path, b) }
func (c *client) del(path string) resp         { c.t.Helper(); return c.do("DELETE", path, nil) }
func (c *client) mustGet(path string) resp     { c.t.Helper(); return expect(c.t, c.get(path), 200) }
func (c *client) mustPost(path string, b any) resp {
	c.t.Helper()
	return expect(c.t, c.post(path, b), 201)
}

func randomEmail() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "contract-" + hex.EncodeToString(b) + "@example.com"
}

func randomName(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return prefix + " " + hex.EncodeToString(b)
}

// newUser registers and logs in a fresh user.
func newUser(t *testing.T) *client {
	t.Helper()
	requireServer(t)
	email, password := randomEmail(), "secret-password"
	a := anon(t)
	expect(t, a.post("/api/auth/register", map[string]string{"email": email, "password": password}), 201)
	login := expect(t, a.post("/api/auth/login", map[string]string{"email": email, "password": password}), 200).json(t)
	token, _ := login["token"].(string)
	if token == "" {
		t.Fatalf("login returned no token: %v", login)
	}
	return &client{t: t, token: token, email: email}
}

func expect(t *testing.T, r resp, status int) resp {
	t.Helper()
	if r.Status != status {
		t.Fatalf("expected status %d, got %d: %s", status, r.Status, r.Body)
	}
	return r
}

// expectError checks status and that the error body carries a message containing want.
func expectError(t *testing.T, r resp, status int, want string) {
	t.Helper()
	expect(t, r, status)
	m := r.json(t)
	msg, _ := m["message"].(string)
	if !strings.Contains(msg, want) {
		t.Fatalf("expected message containing %q, got %q", want, msg)
	}
	if got, _ := m["status"].(float64); int(got) != status {
		t.Fatalf("expected status field %d in body, got %v", status, m["status"])
	}
	if _, ok := m["timestamp"].(string); !ok {
		t.Fatalf("error body has no timestamp: %s", r.Body)
	}
}

func id(t *testing.T, m map[string]any) int64 {
	t.Helper()
	v, ok := m["id"].(float64)
	if !ok {
		t.Fatalf("object has no numeric id: %v", m)
	}
	return int64(v)
}

func num(t *testing.T, m map[string]any, key string) float64 {
	t.Helper()
	v, ok := m[key].(float64)
	if !ok {
		t.Fatalf("field %q is not a number: %v", key, m[key])
	}
	return v
}

func hasKeys(t *testing.T, m map[string]any, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing field %q in %v", k, m)
		}
	}
	if len(m) != len(keys) {
		t.Fatalf("unexpected fields: want exactly %v, got %v", keys, m)
	}
}

func isDate(t *testing.T, v any) {
	t.Helper()
	s, _ := v.(string)
	if _, err := time.Parse("2006-01-02", s); err != nil {
		t.Fatalf("not a date: %v", v)
	}
}

func isTimestamp(t *testing.T, v any) {
	t.Helper()
	s, _ := v.(string)
	if _, err := time.Parse("2006-01-02T15:04:05.999999999", s); err != nil {
		if _, err2 := time.Parse("2006-01-02T15:04", s); err2 != nil {
			t.Fatalf("not a zone-less timestamp: %v", v)
		}
	}
}

func today() string { return time.Now().Format("2006-01-02") }

func p(format string, args ...any) string { return fmt.Sprintf(format, args...) }
