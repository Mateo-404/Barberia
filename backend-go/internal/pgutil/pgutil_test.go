package pgutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestClassifiesSQLState(t *testing.T) {
	unique := &pgconn.PgError{Code: "23505", Message: "duplicate key"}
	fk := &pgconn.PgError{Code: "23503", Message: "foreign key"}
	other := &pgconn.PgError{Code: "42P01", Message: "undefined table"}

	if !IsUniqueViolation(unique) {
		t.Error("23505 should be a unique violation")
	}
	if IsUniqueViolation(fk) || IsUniqueViolation(other) {
		t.Error("other SQLSTATEs are not unique violations")
	}
	if !IsForeignKeyViolation(fk) {
		t.Error("23503 should be an FK violation")
	}
	if IsForeignKeyViolation(unique) {
		t.Error("23505 is not an FK violation")
	}
	if !IsExclusionViolation(&pgconn.PgError{Code: "23P01"}) || !IsCheckViolation(&pgconn.PgError{Code: "23514"}) {
		t.Error("exclusion/check violations not detected")
	}
}

func TestUnwrapsWrappedErrors(t *testing.T) {
	wrapped := fmt.Errorf("tx: %w", fmt.Errorf("create appointment: %w", &pgconn.PgError{Code: "23505"}))
	if !IsUniqueViolation(wrapped) {
		t.Error("errors.As should see through double wrapping")
	}
}

func TestNonPostgresErrors(t *testing.T) {
	if IsUniqueViolation(nil) {
		t.Error("nil is not a unique violation")
	}
	if IsUniqueViolation(errors.New("generic")) || IsForeignKeyViolation(errors.New("generic")) {
		t.Error("generic errors must not be classified")
	}
}
