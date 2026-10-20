// Package config loads the runtime configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// devJWTSecret mirrors the development default of the former Spring backend.
// It is only used when APP_ENV=dev and JWT_SECRET is not set.
const devJWTSecret = "dev-only-secret-key-change-in-production-min-32-chars"

type Config struct {
	Env                   string
	Port                  string
	DatabaseURL           string
	JWTSecret             []byte
	JWTExpiration         time.Duration
	CORSAllowedOrigins    []string
	Location              *time.Location
	ExpectedSchemaVersion string
	SkipSchemaCheck       bool
}

func (c Config) IsDev() bool { return c.Env == "dev" }

// Load reads the configuration. Supported variables:
//
//	APP_ENV                 "dev" or "prod" (default "prod")
//	PORT                    HTTP port (default 8080; set automatically by Cloud Run)
//	DATABASE_URL            postgres URL or key/value DSN; alternatively DB_HOST, DB_PORT,
//	                        DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE
//	JWT_SECRET              HMAC secret, at least 32 bytes (required outside dev)
//	JWT_EXPIRATION_MS       token lifetime in milliseconds (default 86400000 = 24h)
//	CORS_ALLOWED_ORIGINS    comma separated list of allowed origins
//	APP_TIMEZONE            IANA zone used to decide what "today" is (default Europe/Berlin)
//	SCHEMA_VERSION          Flyway version the app expects (default 14)
//	SKIP_SCHEMA_CHECK       "true" disables the Flyway version check
func Load() (Config, error) {
	cfg := Config{
		Env:                   getenv("APP_ENV", "prod"),
		Port:                  getenv("PORT", "8080"),
		ExpectedSchemaVersion: getenv("SCHEMA_VERSION", "14"),
		SkipSchemaCheck:       strings.EqualFold(os.Getenv("SKIP_SCHEMA_CHECK"), "true"),
	}

	dbURL, err := databaseURL(cfg.IsDev())
	if err != nil {
		return cfg, err
	}
	cfg.DatabaseURL = dbURL

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		if !cfg.IsDev() {
			return cfg, errors.New("JWT_SECRET must be set")
		}
		secret = devJWTSecret
	}
	if len(secret) < 32 {
		return cfg, errors.New("JWT_SECRET must be at least 32 bytes long")
	}
	cfg.JWTSecret = []byte(secret)

	expMs, err := strconv.ParseInt(getenv("JWT_EXPIRATION_MS", "86400000"), 10, 64)
	if err != nil || expMs <= 0 {
		return cfg, errors.New("JWT_EXPIRATION_MS must be a positive integer")
	}
	cfg.JWTExpiration = time.Duration(expMs) * time.Millisecond

	origins := getenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	for _, o := range strings.Split(origins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, strings.TrimRight(o, "/"))
		}
	}

	loc, err := time.LoadLocation(getenv("APP_TIMEZONE", "Europe/Berlin"))
	if err != nil {
		return cfg, fmt.Errorf("invalid APP_TIMEZONE: %w", err)
	}
	cfg.Location = loc

	return cfg, nil
}

func databaseURL(dev bool) (string, error) {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v, nil
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		if !dev {
			return "", errors.New("DATABASE_URL or DB_HOST must be set")
		}
		return "postgres://overvu:overvu@localhost:5432/overvu?sslmode=disable", nil
	}
	// Key/value form so that Cloud SQL unix sockets (/cloudsql/...) work as host.
	parts := []string{
		"host=" + quote(host),
		"port=" + quote(getenv("DB_PORT", "5432")),
		"dbname=" + quote(getenv("DB_NAME", "overvu")),
		"user=" + quote(getenv("DB_USER", "overvu")),
		"sslmode=" + quote(getenv("DB_SSLMODE", "disable")),
	}
	if pw := os.Getenv("DB_PASSWORD"); pw != "" {
		parts = append(parts, "password="+quote(pw))
	}
	return strings.Join(parts, " "), nil
}

func quote(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `'`, `\'`)
	return "'" + v + "'"
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// RedactedDatabaseURL hides the password for log output.
func RedactedDatabaseURL(raw string) string {
	if u, err := url.Parse(raw); err == nil && u.Scheme != "" {
		return u.Redacted()
	}
	if i := strings.Index(raw, "password="); i >= 0 {
		return raw[:i] + "password=xxxxx"
	}
	return raw
}
