package auth

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// The local secret of the old project has 57 bytes; jjwt picked HS384 for it.
const legacySecret = "local-docker-secret-key-change-in-production-min-32-chars"

func TestAlgorithmFollowsSecretLength(t *testing.T) {
	cases := map[int]string{32: "HS256", 47: "HS256", 48: "HS384", 57: "HS384", 64: "HS512", 100: "HS512"}
	for n, want := range cases {
		s, err := NewTokenService([]byte(strings.Repeat("k", n)), time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		if s.Algorithm() != want {
			t.Errorf("%d bytes: got %s want %s", n, s.Algorithm(), want)
		}
	}
	if _, err := NewTokenService([]byte("too-short"), time.Hour); err == nil {
		t.Fatal("short secret must be rejected")
	}
}

func TestRoundTrip(t *testing.T) {
	s, _ := NewTokenService([]byte(legacySecret), time.Hour)
	tok, err := s.Generate(42, "a@b.de")
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.Parse(tok)
	if err != nil || id != 42 {
		t.Fatalf("Parse = %d, %v", id, err)
	}
	header, _ := base64.RawURLEncoding.DecodeString(strings.Split(tok, ".")[0])
	if !strings.Contains(string(header), `"HS384"`) {
		t.Fatalf("unexpected header %s", header)
	}
}

// A token built by hand exactly like jjwt does (HS384, sub/userId/iat/exp) must be accepted,
// so users stay logged in when switching from the Spring backend.
func TestAcceptsJJWTStyleToken(t *testing.T) {
	now := time.Now()
	header := `{"alg":"HS384"}`
	payload, _ := json.Marshal(map[string]any{
		"sub": "user@example.com", "userId": 7, "iat": now.Unix(), "exp": now.Add(time.Hour).Unix(),
	})
	enc := base64.RawURLEncoding
	signingInput := enc.EncodeToString([]byte(header)) + "." + enc.EncodeToString(payload)
	mac := hmac.New(sha512.New384, []byte(legacySecret))
	mac.Write([]byte(signingInput))
	token := signingInput + "." + enc.EncodeToString(mac.Sum(nil))

	s, _ := NewTokenService([]byte(legacySecret), time.Hour)
	id, err := s.Parse(token)
	if err != nil || id != 7 {
		t.Fatalf("Parse = %d, %v", id, err)
	}
}

func TestRejectsInvalidTokens(t *testing.T) {
	s, _ := NewTokenService([]byte(legacySecret), time.Hour)
	other, _ := NewTokenService([]byte(legacySecret+"-other"), time.Hour)
	foreign, _ := other.Generate(1, "x@y.z")

	expired, _ := NewTokenService([]byte(legacySecret), time.Hour)
	expired.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	old, _ := expired.Generate(1, "x@y.z")

	// HS512 requires a 64 byte key; with a 57 byte key it must not be accepted.
	strong, _ := NewTokenService([]byte(strings.Repeat("s", 64)), time.Hour)
	hs512, _ := strong.Generate(1, "x@y.z")

	for name, tok := range map[string]string{
		"garbage": "abc", "foreign signature": foreign, "expired": old, "hs512 with weak key": hs512,
		"alg none": "eyJhbGciOiJub25lIn0.eyJ1c2VySWQiOjEsImV4cCI6OTk5OTk5OTk5OX0.",
	} {
		if _, err := s.Parse(tok); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

// Spring's BCryptPasswordEncoder writes "$2a$10$..." hashes; ours use the same format.
func TestPasswordHashFormat(t *testing.T) {
	h, err := HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$2a$10$") {
		t.Fatalf("unexpected hash format %s", h)
	}
	if !CheckPassword(h, "password123") || CheckPassword(h, "wrong") {
		t.Fatal("check failed")
	}
}
