// Package app holds the dependencies shared by all feature modules.
package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Im-Doh/revu/backend/internal/apptime"
	"github.com/Im-Doh/revu/backend/internal/auth"
	"github.com/Im-Doh/revu/backend/internal/database"
	"github.com/Im-Doh/revu/backend/internal/store"
)

type Deps struct {
	Pool   *pgxpool.Pool
	Q      *store.Queries
	Clock  *apptime.Clock
	Tokens *auth.TokenService
}

func New(pool *pgxpool.Pool, clock *apptime.Clock, tokens *auth.TokenService) *Deps {
	return &Deps{Pool: pool, Q: store.New(pool), Clock: clock, Tokens: tokens}
}

// InTx runs fn in a database transaction.
func (d *Deps) InTx(ctx context.Context, fn func(q *store.Queries) error) error {
	return database.InTx(ctx, d.Pool, fn)
}
