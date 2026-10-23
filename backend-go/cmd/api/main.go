// Command api es el backend de la barbería (Go).
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/config"
	"github.com/Mateo-404/Barberia/backend-go/internal/httpapi"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/administradores"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/tenant"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		slog.Error("arranque fallido", "err", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Chequeo de rol al arrancar si no es dev
	if cfg.AppEnv != "dev" {
		if err := checkAppUserRole(context.Background(), cfg.DatabaseURL); err != nil {
			return err
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, database.Options{
		ConnectionString: cfg.DatabaseURL,
		MaxConns:         cfg.PoolMaxConns,
		ConnTimeout:      cfg.ConnTimeout,
		StatementTimeout: cfg.StatementTimeout,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		if err := database.RunMigrations(ctx, cfg.DatabaseURL, logger); err != nil {
			return err
		}
	}

	jwtManager, err := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		return err
	}

	q := sqlc.New(pool)
	resolver := tenant.NewResolver(q, cfg.DefaultTenantSlug)

	// Seed del administrador inicial (paridad con AdminSeedRunner de Java).
	if cfg.AdminSeed {
		seedAdmins := administradores.New(pool, q, jwtManager, logger)
		t, err := resolver.Resolve(ctx, "")
		if err != nil && !errors.Is(err, tenant.ErrTenantNoConfigurado) {
			return err
		}
		if err == nil {
			if err := seedAdmins.Seed(ctx, t.ID, cfg.AdminSeedEmail, cfg.AdminSeedPassword,
				cfg.AdminSeedNombre, cfg.AdminSeedApellido); err != nil {
				return err
			}
		} else {
			logger.Warn("seed omitido: tenant por defecto inexistente (ejecute las migraciones)")
		}
	}

	srv := httpapi.NewServer(httpapi.Deps{
		Cfg: cfg, Logger: logger, Pool: pool, JWT: jwtManager, Resolver: resolver,
	})

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("servidor escuchando", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	logger.Info("apagando servidor")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdownCtx)
}

func checkAppUserRole(ctx context.Context, connStr string) error {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}
	defer pool.Close()
	var rolsuper, rolbypassrls bool
	err = pool.QueryRow(ctx, "SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user").Scan(&rolsuper, &rolbypassrls)
	if err != nil {
		return err
	}
	if rolsuper || rolbypassrls {
		return fmt.Errorf("app se conecta con rol superusuario o con BYPASSRLS (rolsuper=%v rolbypassrls=%v)", rolsuper, rolbypassrls)
	}
	return nil
}
