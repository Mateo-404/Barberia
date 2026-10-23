// Command migrate applies the embedded migrations with a schema-owner role
// (MIGRATE_DATABASE_URL), separate from the app role (see docs/DEPLOY_NOTES.md).
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	dsn := os.Getenv("MIGRATE_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		logger.Error("MIGRATE_DATABASE_URL or DATABASE_URL is required")
		os.Exit(1)
	}
	if err := database.RunMigrations(context.Background(), dsn, logger); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}
}
