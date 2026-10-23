// Package database manages the pgx pool and runs migrations.
package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/db"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Options configures the connection and the pool.
type Options struct {
	ConnectionString string
	MaxConns         int32
	MaxConnLifetime  time.Duration
	ConnTimeout      time.Duration
	// StatementTimeout caps every query on the server (0 = no limit).
	StatementTimeout time.Duration
}

// NewPool creates a pgxpool and checks the connection.
func NewPool(ctx context.Context, opts Options) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(opts.ConnectionString)
	if err != nil {
		// The connection string is left out on purpose: it carries the
		// password and this error ends up in startup logs.
		return nil, errors.New("invalid DATABASE_URL: the connection string could not be parsed")
	}
	if opts.MaxConns > 0 {
		cfg.MaxConns = opts.MaxConns
	}
	if opts.MaxConnLifetime > 0 {
		cfg.MaxConnLifetime = opts.MaxConnLifetime
	}
	if opts.ConnTimeout <= 0 {
		opts.ConnTimeout = 5 * time.Second
	}
	cfg.ConnConfig.ConnectTimeout = opts.ConnTimeout
	if opts.StatementTimeout > 0 {
		if cfg.ConnConfig.RuntimeParams == nil {
			cfg.ConnConfig.RuntimeParams = map[string]string{}
		}
		cfg.ConnConfig.RuntimeParams["statement_timeout"] =
			strconv.FormatInt(opts.StatementTimeout.Milliseconds(), 10)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, opts.ConnTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping a PostgreSQL: %w", err)
	}
	return pool, nil
}

// RunMigrations applies the embedded migrations up to the latest version.
func RunMigrations(ctx context.Context, connString string, logger *slog.Logger) error {
	src, err := iofs.New(db.MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("open migrations source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, connString)
	if err != nil {
		return fmt.Errorf("inicializar migrate: %w", err)
	}
	defer func() {
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			logger.Warn("cerrar migrate", "source", sourceErr, "db", dbErr)
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("schema unchanged")
			return nil
		}
		return fmt.Errorf("aplicar migraciones: %w", err)
	}
	v, dirty, verErr := m.Version()
	if verErr != nil && !errors.Is(verErr, migrate.ErrNilVersion) {
		logger.Warn("could not read the migration version", "err", verErr)
	}
	logger.Info("migraciones aplicadas", "version", v, "dirty", dirty)
	return nil
}
