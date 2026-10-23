package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
	"github.com/Mateo-404/Barberia/backend-go/internal/pgutil"
	"github.com/go-chi/chi/v5"
)

// writeError maps domain and database errors to problem details.
func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	if pgutil.IsExclusionViolation(err) {
		fromAppError(w, r, apperr.New(apperr.KindConflict, "appointment.slot_taken", nil, "the selected time is not available"))
		return
	}
	if pgutil.IsForeignKeyViolation(err) {
		fromAppError(w, r, apperr.New(apperr.KindConflict, "resource.in_use", nil, "the resource is referenced by other records"))
		return
	}
	if pgutil.IsCheckViolation(err) {
		fromAppError(w, r, apperr.New(apperr.KindValidation, "resource.invalid", nil, "the data violates a database constraint"))
		return
	}
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		fromAppError(w, r, appErr)
		return
	}
	s.logger.Error("internal error", "err", err, "path", r.URL.Path, "request_id", requestID(r))
	internalServerError(w, r)
}

// pathID parses an int64 path parameter; a non-numeric value answers 404,
// because no resource can have that id.
func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	v, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil {
		resourceNotFound(w, r)
		return 0, false
	}
	return v, true
}

// maxPageSize caps page size: an unbounded size multiplies database and
// serialization work (trivial DoS).
const maxPageSize = 100

// pageParams reads ?page (0-based, default 0) and ?size (1–100, default 20).
func pageParams(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	page, size := 0, 20
	var v validator
	if raw := r.URL.Query().Get("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			v.add("page", "invalid_value")
		}
		page = n
	}
	if raw := r.URL.Query().Get("size"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > maxPageSize {
			v.add("size", "invalid_value")
		}
		size = n
	}
	return page, size, v.ok(w, r)
}

// tenantID returns the tenant resolved by the auth middleware.
func tenantID(r *http.Request) int64 {
	t, _ := tenantFrom(r)
	return t.ID
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// page is the JSON envelope of a paged list.
type page[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

func newPage[T any](items []T, total int64, number, size int) page[T] {
	return page[T]{
		Items: items, Page: number, Size: size, Total: total,
		TotalPages: int((total + int64(size) - 1) / int64(size)),
	}
}

// mapSlice converts every element with f.
func mapSlice[A, B any](in []A, f func(A) B) []B {
	out := make([]B, 0, len(in))
	for _, a := range in {
		out = append(out, f(a))
	}
	return out
}

// requestID returns the correlation id set by middlewareRequestID.
func requestID(r *http.Request) string {
	if v, ok := r.Context().Value(ctxKeyRequestID).(string); ok {
		return v
	}
	return ""
}
