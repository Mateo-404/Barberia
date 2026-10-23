package pgutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestClasificacionSQLSTATE(t *testing.T) {
	duplicado := &pgconn.PgError{Code: "23505", Message: "duplicate key"}
	fk := &pgconn.PgError{Code: "23503", Message: "foreign key"}
	otro := &pgconn.PgError{Code: "42P01", Message: "undefined table"}

	if !IsUniqueViolation(duplicado) {
		t.Error("23505 debería ser unique violation")
	}
	if IsUniqueViolation(fk) || IsUniqueViolation(otro) {
		t.Error("otros SQLSTATE no son unique violation")
	}
	if !IsForeignKeyViolation(fk) {
		t.Error("23503 debería ser FK violation")
	}
	if IsForeignKeyViolation(duplicado) {
		t.Error("23505 no es FK violation")
	}
}

func TestErroresEnvolventes(t *testing.T) {
	duplicado := &pgconn.PgError{Code: "23505"}
	envuelto := fmt.Errorf("crear turno: %w", duplicado)
	dobleEnvuelto := fmt.Errorf("tx: %w", envuelto)
	if !IsUniqueViolation(dobleEnvuelto) {
		t.Error("errors.As debería atravesar envoltorios dobles")
	}
}

func TestErroresNoPG(t *testing.T) {
	if IsUniqueViolation(nil) {
		t.Error("nil no es unique violation")
	}
	if IsUniqueViolation(errors.New("genérico")) || IsForeignKeyViolation(errors.New("genérico")) {
		t.Error("errores genéricos no deben clasificarse")
	}
}
