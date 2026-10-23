package database

import (
	"context"
	"fmt"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/tenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTenantTx opens a transaction, sets app.tenant_id from ctx (for RLS)
// and runs fn with *sqlc.Queries bound to that transaction. Handlers cannot
// pass an arbitrary tenant.
func WithTenantTx(ctx context.Context, pool *pgxpool.Pool, fn func(*sqlc.Queries) error) error {
	tid, ok := tenant.ID(ctx)
	if !ok || tid == 0 {
		return fmt.Errorf("tenant_id ausente en context")
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", fmt.Sprintf("%d", tid)); err != nil {
		return err
	}

	q := sqlc.New(pool).WithTx(tx)
	if err := fn(q); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
