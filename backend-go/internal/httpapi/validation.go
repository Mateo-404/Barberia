package httpapi

import (
	"net/http"
	"regexp"
	"strings"
)

// emailPattern: local@domain with optional dotted labels.
var emailPattern = regexp.MustCompile(
	`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// Length caps for text fields (the columns are TEXT; these keep payloads
// proportionate).
const (
	maxName        = 120
	maxEmail       = 254
	maxPhone       = 32
	maxServiceName = 120
	minPassword    = 8
)

// Field error codes, translated by clients.
const (
	codeRequired       = "required"
	codeInvalidEmail   = "invalid_email"
	codeTooLong        = "too_long"
	codeTooShort       = "too_short"
	codeMustBePositive = "must_be_positive"
	codeInvalidDate    = "invalid_datetime"
)

// validator collects every invalid field so the client can show all errors
// at once.
type validator struct {
	errs []fieldProblem
}

func (v *validator) add(field, code string) {
	v.errs = append(v.errs, fieldProblem{Field: field, Code: code})
}

// required: present and not blank.
func (v *validator) required(field string, s *string) bool {
	if s == nil || strings.TrimSpace(*s) == "" {
		v.add(field, codeRequired)
		return false
	}
	return true
}

// maxLen: absent is fine; present must not exceed max bytes.
func (v *validator) maxLen(field string, s *string, max int) {
	if s != nil && len(*s) > max {
		v.add(field, codeTooLong)
	}
}

// minLen: absent is fine; present must have at least min bytes.
func (v *validator) minLen(field string, s *string, min int) {
	if s != nil && len(*s) < min {
		v.add(field, codeTooShort)
	}
}

// email: absent is fine; present must look like an address.
func (v *validator) email(field string, s *string) {
	if s != nil && !emailPattern.MatchString(*s) {
		v.add(field, codeInvalidEmail)
	}
}

// positive: absent is fine; present must be > 0.
func (v *validator) positive(field string, n *float32) {
	if n != nil && *n <= 0 {
		v.add(field, codeMustBePositive)
	}
}

// ok answers 400 with every collected error when there is any.
func (v *validator) ok(w http.ResponseWriter, r *http.Request) bool {
	if len(v.errs) == 0 {
		return true
	}
	validationFailed(w, r, v.errs)
	return false
}
