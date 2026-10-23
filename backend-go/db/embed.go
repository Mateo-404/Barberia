// Package db contiene las migraciones SQL embebidas en el binario.
package db

import "embed"

// MigrationsFS expone las migraciones para golang-migrate.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
