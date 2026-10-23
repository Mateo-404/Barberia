// Package apperr define los errores de dominio que el HTTP layer traduce a
// los ProblemDetail exactos del backend Java.
package apperr

import "fmt"

// Kind identifica la categoría de error.
type Kind int

const (
	// KindBusinessRule → 400 /errors/business-rule-violation.
	KindBusinessRule Kind = iota + 1
	// KindNotFound → 404 /errors/resource-not-found (con cuerpo).
	KindNotFound
	// KindAlreadyExists → 409 /errors/resource-already-exists.
	KindAlreadyExists
	// KindInvalidCredentials → 401 /errors/invalid-credentials.
	KindInvalidCredentials
	// KindValidation → 400 about:blank "Invalid request content.".
	KindValidation
	// KindConflict → 409 con detalle genérico (violación de unicidad no
	// mapeada específicamente).
	KindConflict
)

// Error es un error de dominio con mensaje listo para el cliente.
type Error struct {
	Kind    Kind
	Message string
}

func (e *Error) Error() string { return e.Message }

// New construye un error de dominio.
func New(kind Kind, format string, args ...any) *Error {
	return &Error{Kind: kind, Message: fmt.Sprintf(format, args...)}
}
