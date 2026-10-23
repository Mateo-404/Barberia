package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	settingsmod "github.com/Mateo-404/Barberia/backend-go/internal/modules/settings"
	"github.com/Mateo-404/Barberia/backend-go/internal/settings"
)

// hAdminSettingsGet responde GET /api/admin/settings.
func (s *Server) hAdminSettingsGet(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenantID(r)
	if !ok {
		forbidden(w, r)
		return
	}
	est, err := s.settings.GetAdmin(r.Context(), tid)
	if err != nil {
		s.logger.Error("leer settings", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
		return
	}
	responderJSON(w, http.StatusOK, est)
}

// hAdminSettingsDraft responde PUT /api/admin/settings/draft.
func (s *Server) hAdminSettingsDraft(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenantID(r)
	if !ok {
		forbidden(w, r)
		return
	}
	raw, ok := leerJSONRaw(w, r, s.cfg.MaxRequestBody)
	if !ok {
		return
	}
	res, err := s.settings.SaveDraft(r.Context(), tid, raw)
	if err != nil {
		var verr *settingsmod.ErrorValidacion
		if errors.As(err, &verr) {
			escribirValidacion(w, r, verr.Details)
			return
		}
		s.logger.Error("guardar borrador", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
		return
	}
	responderJSON(w, http.StatusOK, res)
}

// hAdminSettingsPublish responde POST /api/admin/settings/publish.
func (s *Server) hAdminSettingsPublish(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenantID(r)
	if !ok {
		forbidden(w, r)
		return
	}
	res, err := s.settings.Publish(r.Context(), tid)
	if err != nil {
		if errors.Is(err, settingsmod.ErrSinBorrador) {
			escribirErrorJSON(w, r, http.StatusUnprocessableEntity, "validation_failed", "No hay borrador para publicar")
			return
		}
		s.logger.Error("publicar settings", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
		return
	}
	responderJSON(w, http.StatusOK, res)
}

// hAdminSettingsVersions responde GET /api/admin/settings/versions.
func (s *Server) hAdminSettingsVersions(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenantID(r)
	if !ok {
		forbidden(w, r)
		return
	}
	vs, err := s.settings.Versions(r.Context(), tid)
	if err != nil {
		s.logger.Error("listar versiones", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
		return
	}
	responderJSON(w, http.StatusOK, vs)
}

// hAdminSettingsRollback responde POST /api/admin/settings/rollback.
func (s *Server) hAdminSettingsRollback(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenantID(r)
	if !ok {
		forbidden(w, r)
		return
	}
	var body struct {
		Version *int32 `json:"version"`
	}
	if !decodificarJSONConLimite(w, r, &body, s.cfg.MaxRequestBody) {
		return
	}
	if body.Version == nil {
		escribirErrorJSON(w, r, http.StatusUnprocessableEntity, "validation_failed", "Falta 'version'")
		return
	}
	res, err := s.settings.Rollback(r.Context(), tid, *body.Version)
	if err != nil {
		if errors.Is(err, settingsmod.ErrVersionNoEncontrada) {
			notFoundConCuerpo(w, r, "La versión indicada no existe")
			return
		}
		s.logger.Error("rollback settings", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
		return
	}
	responderJSON(w, http.StatusOK, res)
}

// hAdminSettingsPreview responde GET /api/admin/preview/theme.css (CSS del borrador).
func (s *Server) hAdminSettingsPreview(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenantID(r)
	if !ok {
		forbidden(w, r)
		return
	}
	draft, err := s.settings.Draft(r.Context(), tid)
	if err != nil {
		s.logger.Error("preview settings", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
		return
	}
	css, err := settings.CSSFromRaw(draft)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write([]byte(css))
}

// hPublicSettings responde GET /api/public/settings (publicado o defaults).
func (s *Server) hPublicSettings(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenantID(r)
	if !ok {
		forbidden(w, r)
		return
	}
	pub, err := s.settings.Published(r.Context(), tid)
	if err != nil {
		s.logger.Error("leer settings publicos", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	responderJSON(w, http.StatusOK, json.RawMessage(pub))
}

// hPublicThemeCSS responde GET /api/public/theme.css?v=<hash>.
func (s *Server) hPublicThemeCSS(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenantID(r)
	if !ok {
		forbidden(w, r)
		return
	}
	pub, err := s.settings.Published(r.Context(), tid)
	if err != nil {
		s.logger.Error("leer theme public", "err", err, "path", r.URL.Path, "request_id", requestID(r))
		internalServerError(w, r)
		return
	}
	css, err := settings.CSSFromRaw(pub)
	if err != nil {
		internalServerError(w, r)
		return
	}
	hash := settings.Hash(css)
	w.Header().Set("ETag", `"`+hash+`"`)
	if v := r.URL.Query().Get("v"); v != "" && v == hash {
		// Mismo hash: cacheable para siempre.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write([]byte(css))
}

// escribirValidacion responde 422 con details por campo.
func escribirValidacion(w http.ResponseWriter, r *http.Request, details []settings.Detail) {
	escribirErrorJSON(w, r, http.StatusUnprocessableEntity, "validation_failed", "La configuración no es válida", details)
}

// escribirErrorJSON responde JSON estándar de error.
func escribirErrorJSON(w http.ResponseWriter, r *http.Request, status int, code, mensaje string, details ...any) {
	body := map[string]any{"error": mensaje, "code": code}
	if len(details) > 0 {
		body["details"] = details[0]
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// leerJSONRaw lee el cuerpo como JSON crudo con límite de tamaño.
func leerJSONRaw(w http.ResponseWriter, r *http.Request, maxBytes int64) (json.RawMessage, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		badRequestValidacion(w, r)
		return nil, false
	}
	return raw, true
}
