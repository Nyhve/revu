// Package database owns the PostgreSQL connection pool.
//
// The schema is managed exclusively by the Flyway migrations in /migrations.
// This application never migrates; it only verifies on startup that the
// expected Flyway version has been applied.
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Im-Doh/revu/backend/internal/store"
)

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// CheckSchema verifies that Flyway has successfully applied the expected version.
func CheckSchema(ctx context.Context, pool *pgxpool.Pool, version string) error {
	var ok bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM flyway_schema_history WHERE version = $1 AND success)`,
		version).Scan(&ok)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return errors.New("table flyway_schema_history not found: run the Flyway migrations first")
		}
		return fmt.Errorf("check schema version: %w", err)
	}
	if !ok {
		return fmt.Errorf("flyway migration V%s has not been applied: run the Flyway migrations first", version)
	}
	return nil
}

// InTx runs fn inside a transaction and commits if fn returns nil.
func InTx(ctx context.Context, pool *pgxpool.Pool, fn func(q *store.Queries) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	if err := fn(store.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// IsUniqueViolation reports whether err is a PostgreSQL unique constraint violation.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// IsNoRows reports whether err signals an empty result.
func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
