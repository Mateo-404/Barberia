package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
	settingsmod "github.com/Mateo-404/Barberia/backend-go/internal/modules/settings"
	"github.com/Mateo-404/Barberia/backend-go/internal/settings"
)

// handleSettingsGet: GET /api/admin/settings.
func (s *Server) handleSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := s.settings.GetAdmin(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// handleSettingsDraft: PUT /api/admin/settings/draft.
func (s *Server) handleSettingsDraft(w http.ResponseWriter, r *http.Request) {
	raw, ok := readRawJSON(w, r, s.cfg.MaxRequestBody)
	if !ok {
		return
	}
	res, err := s.settings.SaveDraft(r.Context(), tenantID(r), raw)
	if err != nil {
		var verr *settingsmod.ValidationError
		if errors.As(err, &verr) {
			settingsInvalid(w, r, verr.Details)
			return
		}
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleSettingsPublish: POST /api/admin/settings/publish.
func (s *Server) handleSettingsPublish(w http.ResponseWriter, r *http.Request) {
	res, err := s.settings.Publish(r.Context(), tenantID(r))
	if errors.Is(err, settingsmod.ErrNoDraft) {
		fromAppError(w, r, apperr.New(apperr.KindBusinessRule, "settings.no_draft", nil, "there is no draft to publish"))
		return
	}
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleSettingsVersions: GET /api/admin/settings/versions.
func (s *Server) handleSettingsVersions(w http.ResponseWriter, r *http.Request) {
	vs, err := s.settings.Versions(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, vs)
}

// handleSettingsRollback: POST /api/admin/settings/rollback.
func (s *Server) handleSettingsRollback(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version *int32 `json:"version"`
	}
	if !decodeJSON(w, r, &body, s.cfg.MaxRequestBody) {
		return
	}
	if body.Version == nil {
		validationFailed(w, r, []fieldProblem{{Field: "version", Code: codeRequired}})
		return
	}
	res, err := s.settings.Rollback(r.Context(), tenantID(r), *body.Version)
	if errors.Is(err, settingsmod.ErrVersionNotFound) {
		fromAppError(w, r, apperr.New(apperr.KindNotFound, "settings.version_not_found",
			apperr.Params{"version": *body.Version}, "version %d does not exist", *body.Version))
		return
	}
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleSettingsPreview: GET /api/admin/settings/preview/theme.css (draft CSS).
func (s *Server) handleSettingsPreview(w http.ResponseWriter, r *http.Request) {
	draft, err := s.settings.Draft(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	css, err := settings.CSSFromRaw(draft)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeCSS(w, css)
}

// handlePublicSettings: GET /api/public/settings (published or defaults).
func (s *Server) handlePublicSettings(w http.ResponseWriter, r *http.Request) {
	pub, err := s.settings.Published(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	writeJSON(w, http.StatusOK, json.RawMessage(pub))
}

// handlePublicThemeCSS: GET /api/public/theme.css?v=<hash>. The current hash
// is cacheable forever; anything else revalidates through the ETag.
func (s *Server) handlePublicThemeCSS(w http.ResponseWriter, r *http.Request) {
	pub, err := s.settings.Published(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	css, err := settings.CSSFromRaw(pub)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	hash := settings.Hash(css)
	w.Header().Set("ETag", `"`+hash+`"`)
	if v := r.URL.Query().Get("v"); v != "" && v == hash {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	writeCSS(w, css)
}

func writeCSS(w http.ResponseWriter, css string) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write([]byte(css))
}

// settingsInvalid answers 422 with the invalid settings fields.
func settingsInvalid(w http.ResponseWriter, r *http.Request, details []settings.Detail) {
	writeProblem(w, r, problem{
		Type: "/errors/validation", Title: "Unprocessable Entity", Status: http.StatusUnprocessableEntity,
		Detail: "the settings are invalid", Code: "settings.invalid",
		Errors: mapSlice(details, func(d settings.Detail) fieldProblem { return fieldProblem(d) }),
	})
}

// readRawJSON reads the body as raw JSON, capped at maxBytes.
func readRawJSON(w http.ResponseWriter, r *http.Request, maxBytes int64) (json.RawMessage, bool) {
	var raw json.RawMessage
	return raw, decodeJSON(w, r, &raw, maxBytes)
}
