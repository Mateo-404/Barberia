// Package pgutil clasifica errores de PostgreSQL por SQLSTATE.
package pgutil

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation indica violación de unicidad (23505).
func IsUniqueViolation(err error) bool { return codigo(err) == "23505" }

// IsForeignKeyViolation indica violación de clave foránea (23503).
func IsForeignKeyViolation(err error) bool { return codigo(err) == "23503" }

func codigo(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func IsExclusionViolation(err error) bool { return codigo(err) == "23P01" }
func IsCheckViolation(err error) bool     { return codigo(err) == "23514" }
