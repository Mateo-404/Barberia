// Package apperr defines the domain errors that the HTTP layer turns into
// RFC 7807 problem details.
//
// Every error carries a stable machine-readable Code (plus Params) that
// clients translate into user-facing text; Message is English and meant for
// developers and logs.
package apperr

import "fmt"

// Kind is the error category, which decides the HTTP status.
type Kind int

const (
	// KindBusinessRule → 400 /errors/business-rule-violation.
	KindBusinessRule Kind = iota + 1
	// KindNotFound → 404 /errors/resource-not-found.
	KindNotFound
	// KindAlreadyExists → 409 /errors/resource-already-exists.
	KindAlreadyExists
	// KindInvalidCredentials → 401 /errors/invalid-credentials.
	KindInvalidCredentials
	// KindValidation → 400 /errors/validation.
	KindValidation
	// KindConflict → 409 /errors/conflict.
	KindConflict
)

// Params are the values a client needs to render the message for Code.
type Params map[string]any

// Error is a domain error.
type Error struct {
	Kind    Kind
	Code    string
	Params  Params
	Message string
}

func (e *Error) Error() string { return e.Message }

// New builds a domain error. Code is a stable dotted identifier such as
// "appointment.outside_hours"; params may be nil.
func New(kind Kind, code string, params Params, format string, args ...any) *Error {
	return &Error{Kind: kind, Code: code, Params: params, Message: fmt.Sprintf(format, args...)}
}
