// Package pgutil classifies PostgreSQL errors by SQLSTATE.
package pgutil

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation reports a unique violation (23505).
func IsUniqueViolation(err error) bool { return sqlState(err) == "23505" }

// IsForeignKeyViolation reports a foreign key violation (23503).
func IsForeignKeyViolation(err error) bool { return sqlState(err) == "23503" }

// IsExclusionViolation reports an exclusion constraint violation (23P01).
func IsExclusionViolation(err error) bool { return sqlState(err) == "23P01" }

// IsCheckViolation reports a check constraint violation (23514).
func IsCheckViolation(err error) bool { return sqlState(err) == "23514" }

func sqlState(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
