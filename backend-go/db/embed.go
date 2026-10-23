// Package db embeds the SQL migrations in the binary.
package db

import "embed"

// MigrationsFS exposes the migrations to golang-migrate.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
