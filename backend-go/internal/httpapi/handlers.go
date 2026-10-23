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

// escribirError mapea errores de dominio/BD a los ProblemDetail exactos.
func (s *Server) escribirError(w http.ResponseWriter, r *http.Request, err error) {
	if pgutil.IsExclusionViolation(err) {
		conflictSlotTaken(w, r)
		return
	}
	if pgutil.IsForeignKeyViolation(err) {
		// Java: DataIntegrityViolationException sin handler → 500 (mantener), pero mapear a 422 si aplica
		s.logger.Error("violación de integridad referencial", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
		return
	}
	if pgutil.IsCheckViolation(err) {
		unprocessable(w, r, err.Error())
		return
	}
	var appErr *apperr.Error
	switch {
	case errors.As(err, &appErr):
		switch appErr.Kind {
		case apperr.KindBusinessRule:
			businessRule(w, r, appErr.Message)
		case apperr.KindNotFound:
			notFoundConCuerpo(w, r, appErr.Message)
		case apperr.KindAlreadyExists:
			alreadyExists(w, r, appErr.Message)
		case apperr.KindConflict:
			conflictSlotTaken(w, r)
		case apperr.KindInvalidCredentials:
			invalidCredentials(w, r, appErr.Message)
		case apperr.KindValidation:
			badRequestValidacion(w, r)
		default:
			alreadyExists(w, r, appErr.Message)
		}
	default:
		s.logger.Error("error interno", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
	}
}

// paramID parsea un path param int64. Un valor no numérico replica el 400
// de Spring (MethodArgumentTypeMismatch).
func (s *Server) paramID(w http.ResponseWriter, r *http.Request, nombre string) (int64, bool) {
	v, err := strconv.ParseInt(chi.URLParam(r, nombre), 10, 64)
	if err != nil {
		badRequestValidacion(w, r)
		return 0, false
	}
	return v, true
}

// maxPaginaSize acota el tamaño de página: sin tope, un size enorme multiplica
// el trabajo de la BD y el de serialización (DoS trivial).
const maxPaginaSize = 100

// paramPagina parsea page/size; igual que PageRequest.of, un valor inválido
// termina en 500. Como en Spring, page/size son int32: un valor fuera de rango
// (o un size desmedido) da 400 (MethodArgumentTypeMismatch).
func (s *Server) paramPagina(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	rawPage := chi.URLParam(r, "page")
	rawSize := chi.URLParam(r, "size")
	page, err1 := strconv.ParseInt(rawPage, 10, 32)
	size, err2 := strconv.ParseInt(rawSize, 10, 32)
	if err1 != nil || err2 != nil {
		badRequestValidacion(w, r)
		return 0, 0, false
	}
	if page < 0 || size < 1 || size > maxPaginaSize {
		internalServerError(w, r)
		return 0, 0, false
	}
	return int(page), int(size), true
}

// tenantID r del tenant resuelto.
func tenantID(r *http.Request) (int64, bool) {
	t, ok := tenantDe(r)
	return t.ID, ok
}

func (s *Server) hPing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Backend funcionando!"))
}

func (s *Server) hHealthz(w http.ResponseWriter, r *http.Request) {
	responderJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) hReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// pagina replica el JSON de Page<T> de Spring Data.
type pagina[T any] struct {
	Content          []T      `json:"content"`
	Pageable         pageable `json:"pageable"`
	TotalElements    int64    `json:"totalElements"`
	TotalPages       int      `json:"totalPages"`
	Last             bool     `json:"last"`
	First            bool     `json:"first"`
	Size             int      `json:"size"`
	Number           int      `json:"number"`
	Sort             sorteo   `json:"sort"`
	NumberOfElements int      `json:"numberOfElements"`
	Empty            bool     `json:"empty"`
}

type pageable struct {
	PageNumber int    `json:"pageNumber"`
	PageSize   int    `json:"pageSize"`
	Sort       sorteo `json:"sort"`
	Offset     int64  `json:"offset"`
	Paged      bool   `json:"paged"`
	Unpaged    bool   `json:"unpaged"`
}

type sorteo struct {
	Sorted   bool `json:"sorted"`
	Unsorted bool `json:"unsorted"`
	Empty    bool `json:"empty"`
}

var sorteoDefault = sorteo{Sorted: false, Unsorted: true, Empty: true}

func construirPagina[T any](content []T, total int64, page, size int) pagina[T] {
	totalPages := 0
	if size > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}
	return pagina[T]{
		Content: content,
		Pageable: pageable{
			PageNumber: page, PageSize: size, Sort: sorteoDefault,
			Offset: int64(page * size), Paged: true, Unpaged: false,
		},
		TotalElements:    total,
		TotalPages:       totalPages,
		Last:             page >= totalPages-1,
		First:            page == 0,
		Size:             size,
		Number:           page,
		Sort:             sorteoDefault,
		NumberOfElements: len(content),
		Empty:            len(content) == 0,
	}
}

// requestID extrae el id de correlación del contexto (asignado por middlewareRequestID).
func requestID(r *http.Request) string {
	if v, ok := r.Context().Value(ctxKeyRequestID).(string); ok {
		return v
	}
	return ""
}
