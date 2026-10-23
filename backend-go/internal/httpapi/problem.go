package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
)

// problem is an RFC 7807 problem detail. Code (plus Params) is the stable
// identifier clients translate; Title and Detail are English, for
// developers.
type problem struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Code     string         `json:"code"`
	Params   apperr.Params  `json:"params,omitempty"`
	Errors   []fieldProblem `json:"errors,omitempty"`
}

// fieldProblem is one invalid field of a request body.
type fieldProblem struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

func writeProblem(w http.ResponseWriter, r *http.Request, p problem) {
	if r != nil && p.Instance == "" {
		p.Instance = r.URL.Path
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, problem{
		Type: "/errors/unauthorized", Title: "Unauthorized", Status: http.StatusUnauthorized,
		Detail: "authentication is required", Code: "auth.unauthorized",
	})
}

func invalidToken(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, problem{
		Type: "/errors/invalid-token", Title: "Unauthorized", Status: http.StatusUnauthorized,
		Detail: "invalid or expired JWT", Code: "auth.invalid_token",
	})
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, problem{
		Type: "/errors/forbidden", Title: "Forbidden", Status: http.StatusForbidden,
		Detail: "access denied", Code: "auth.forbidden",
	})
}

// badRequest is a malformed body (unreadable JSON, trailing data, too big).
func badRequest(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, problem{
		Type: "/errors/bad-request", Title: "Bad Request", Status: http.StatusBadRequest,
		Detail: "invalid request content", Code: "request.malformed",
	})
}

// validationFailed lists every invalid field of a well-formed body.
func validationFailed(w http.ResponseWriter, r *http.Request, errs []fieldProblem) {
	writeProblem(w, r, problem{
		Type: "/errors/validation", Title: "Bad Request", Status: http.StatusBadRequest,
		Detail: "one or more fields are invalid", Code: "request.invalid_fields", Errors: errs,
	})
}

func rateLimited(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, problem{
		Type: "/errors/rate-limited", Title: "Too Many Requests", Status: http.StatusTooManyRequests,
		Detail: "too many attempts; try again in a minute", Code: "request.rate_limited",
	})
}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, problem{
		Type: "/errors/internal-server-error", Title: "Internal Server Error", Status: http.StatusInternalServerError,
		Detail: "unexpected error; try again later", Code: "server.internal",
	})
}

func routeNotFound(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, problem{
		Type: "/errors/not-found", Title: "Not Found", Status: http.StatusNotFound,
		Detail: "no route for " + r.URL.Path, Code: "request.not_found",
	})
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, problem{
		Type: "/errors/method-not-allowed", Title: "Method Not Allowed", Status: http.StatusMethodNotAllowed,
		Detail: "method " + r.Method + " is not supported", Code: "request.method_not_allowed",
	})
}

// resourceNotFound is a 404 for a missing entity (GET/PATCH/DELETE by id).
func resourceNotFound(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, problem{
		Type: "/errors/resource-not-found", Title: "Not Found", Status: http.StatusNotFound,
		Detail: "resource not found", Code: "resource.not_found",
	})
}

// fromAppError maps a domain error to its problem detail.
func fromAppError(w http.ResponseWriter, r *http.Request, e *apperr.Error) {
	p := problem{Detail: e.Message, Code: e.Code, Params: e.Params}
	switch e.Kind {
	case apperr.KindBusinessRule:
		p.Type, p.Title, p.Status = "/errors/business-rule-violation", "Business Rule Violation", http.StatusBadRequest
	case apperr.KindNotFound:
		p.Type, p.Title, p.Status = "/errors/resource-not-found", "Not Found", http.StatusNotFound
	case apperr.KindAlreadyExists:
		p.Type, p.Title, p.Status = "/errors/resource-already-exists", "Conflict", http.StatusConflict
	case apperr.KindConflict:
		p.Type, p.Title, p.Status = "/errors/conflict", "Conflict", http.StatusConflict
	case apperr.KindInvalidCredentials:
		p.Type, p.Title, p.Status = "/errors/invalid-credentials", "Unauthorized", http.StatusUnauthorized
	default:
		p.Type, p.Title, p.Status = "/errors/validation", "Bad Request", http.StatusBadRequest
	}
	writeProblem(w, r, p)
}
