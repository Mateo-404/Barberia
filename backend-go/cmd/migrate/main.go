// Command migrate aplica las migraciones embebidas con un rol dueño del esquema
// (MIGRATE_DATABASE_URL), separado del rol de la app (ver docs/DEPLOY_NOTES.md).
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
		logger.Error("MIGRATE_DATABASE_URL o DATABASE_URL requerido")
		os.Exit(1)
	}
	if err := database.RunMigrations(context.Background(), dsn, logger); err != nil {
		logger.Error("migrar", "err", err)
		os.Exit(1)
	}
}
