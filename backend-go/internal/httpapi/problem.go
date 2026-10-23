package httpapi

import (
	"encoding/json"
	"net/http"
)

// problem replica spring ProblemDetail. El orden de campos es irrelevante.
type problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func escribirProblem(w http.ResponseWriter, p problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

// unauthorizedSinToken replica el entry point de Spring (sin instance).
func unauthorizedSinToken(w http.ResponseWriter) {
	escribirProblem(w, problem{
		Type: "/errors/unauthorized", Title: "No autorizado", Status: http.StatusUnauthorized,
		Detail: "Se requiere autenticación para acceder a este recurso",
	})
}

func unauthorizedTokenInvalido(w http.ResponseWriter) {
	escribirProblem(w, problem{
		Type: "/errors/unauthorized", Title: "No autorizado", Status: http.StatusUnauthorized,
		Detail: "Token JWT inválido o expirado",
	})
}

// badRequestValidacion replica el 400 genérico de Spring 6.2 para @Valid y
// JSON ilegible.
func badRequestValidacion(w http.ResponseWriter, r *http.Request) {
	escribirProblem(w, problem{
		Type: "about:blank", Title: "Bad Request", Status: http.StatusBadRequest,
		Detail: "Invalid request content.", Instance: r.URL.Path,
	})
}

func businessRule(w http.ResponseWriter, r *http.Request, detalle string) {
	escribirProblem(w, problem{
		Type: "/errors/business-rule-violation", Title: "Violación de regla de negocio",
		Status: http.StatusBadRequest, Detail: detalle, Instance: r.URL.Path,
	})
}

func notFoundConCuerpo(w http.ResponseWriter, r *http.Request, detalle string) {
	escribirProblem(w, problem{
		Type: "/errors/resource-not-found", Title: "Recurso no encontrado",
		Status: http.StatusNotFound, Detail: detalle, Instance: r.URL.Path,
	})
}

func alreadyExists(w http.ResponseWriter, r *http.Request, detalle string) {
	escribirProblem(w, problem{
		Type: "/errors/resource-already-exists", Title: "El recurso ya existe",
		Status: http.StatusConflict, Detail: detalle, Instance: r.URL.Path,
	})
}

func invalidCredentials(w http.ResponseWriter, r *http.Request, detalle string) {
	escribirProblem(w, problem{
		Type: "/errors/invalid-credentials", Title: "Credenciales inválidas",
		Status: http.StatusUnauthorized, Detail: detalle, Instance: r.URL.Path,
	})
}

func rateLimited(w http.ResponseWriter, r *http.Request) {
	escribirProblem(w, problem{
		Type: "/errors/rate-limited", Title: "Demasiadas solicitudes",
		Status:   http.StatusTooManyRequests,
		Detail:   "Demasiados intentos de inicio de sesión. Intente nuevamente en un minuto.",
		Instance: r.URL.Path,
	})
}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	escribirProblem(w, problem{
		Type: "/errors/internal-server-error", Title: "Error interno del servidor",
		Status:   http.StatusInternalServerError,
		Detail:   "Ocurrió un error inesperado. Intente nuevamente más tarde.",
		Instance: r.URL.Path,
	})
}

// notFoundVacio replica el 404 sin cuerpo del CRUD genérico.
func notFoundVacio(w http.ResponseWriter) { w.WriteHeader(http.StatusNotFound) }

// notFoundRecurso replica el 404 de Spring para rutas no mapeadas.
func notFoundRecurso(w http.ResponseWriter, r *http.Request) {
	escribirProblem(w, problem{
		Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound,
		Detail: "No static resource " + r.URL.Path + ".", Instance: r.URL.Path,
	})
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	escribirProblem(w, problem{
		Type: "about:blank", Title: "Method Not Allowed", Status: http.StatusMethodNotAllowed,
		Detail: "Method '" + r.Method + "' is not supported.", Instance: r.URL.Path,
	})
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	escribirProblem(w, problem{
		Type: "/errors/forbidden", Title: "Prohibido",
		Status: http.StatusForbidden, Detail: "Acceso prohibido", Instance: r.URL.Path,
	})
}

func conflictSlotTaken(w http.ResponseWriter, r *http.Request) {
	escribirProblemCode(w, problemWithCode{
		Type: "/errors/slot-taken", Title: "Horario no disponible",
		Status: http.StatusConflict, Detail: "El horario seleccionado no está disponible", Code: "slot_taken", Instance: r.URL.Path,
	})
}

func unprocessable(w http.ResponseWriter, r *http.Request, detalle string) {
	escribirProblem(w, problem{
		Type: "/errors/unprocessable-entity", Title: "Entidad no procesable",
		Status: http.StatusUnprocessableEntity, Detail: detalle, Instance: r.URL.Path,
	})
}

type problemWithCode struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code,omitempty"`
}

func escribirProblemCode(w http.ResponseWriter, p problemWithCode) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}
