package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
	settingsmod "github.com/Mateo-404/Barberia/backend-go/internal/modules/settings"
	"github.com/Mateo-404/Barberia/backend-go/internal/settings"
)

// settingsError maps a settings failure to a problem detail and reports
// whether it handled the error.
func (s *Server) settingsError(w http.ResponseWriter, r *http.Request, err error, invalidCode string) {
	var verr *settingsmod.ValidationError
	switch {
	case errors.As(err, &verr):
		writeProblem(w, r, problem{
			Type: "/errors/validation", Title: "Unprocessable Entity", Status: http.StatusUnprocessableEntity,
			Detail: "the document is invalid", Code: invalidCode,
			Errors: mapSlice(verr.Details, func(d settings.Detail) fieldProblem { return fieldProblem(d) }),
		})
	case errors.Is(err, settingsmod.ErrNoDraft):
		fromAppError(w, r, apperr.New(apperr.KindBusinessRule, "settings.no_draft", nil, "there is no draft to publish"))
	case errors.Is(err, settingsmod.ErrVersionNotFound):
		fromAppError(w, r, apperr.New(apperr.KindNotFound, "settings.version_not_found", nil, "version not found"))
	default:
		s.writeError(w, r, err)
	}
}

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
	var raw json.RawMessage
	if !decodeJSON(w, r, &raw, s.cfg.MaxRequestBody) {
		return
	}
	res, err := s.settings.SaveDraft(r.Context(), tenantID(r), raw)
	if err != nil {
		s.settingsError(w, r, err, "settings.invalid")
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleThemePreview: POST /api/admin/theme/preview renders an unsaved
// theme, so the editor can show it live without writing anything.
func (s *Server) handleThemePreview(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	if !decodeJSON(w, r, &raw, s.cfg.MaxRequestBody) {
		return
	}
	res, err := s.settings.Preview(raw)
	if err != nil {
		s.settingsError(w, r, err, "settings.invalid")
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleThemePresets: GET /api/admin/theme/presets.
func (s *Server) handleThemePresets(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.settings.Presets())
}

// handleSettingsPublish: POST /api/admin/settings/publish.
func (s *Server) handleSettingsPublish(w http.ResponseWriter, r *http.Request) {
	res, err := s.settings.Publish(r.Context(), tenantID(r))
	if err != nil {
		s.settingsError(w, r, err, "settings.invalid")
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
	if err != nil {
		s.settingsError(w, r, err, "settings.invalid")
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleSettingsPreferences: PATCH /api/admin/settings/preferences.
func (s *Server) handleSettingsPreferences(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ThemeEditorEnabled *bool `json:"themeEditorEnabled"`
		OnboardingEnabled  *bool `json:"onboardingEnabled"`
	}
	if !decodeJSON(w, r, &body, s.cfg.MaxRequestBody) {
		return
	}
	res, err := s.settings.UpdatePreferences(r.Context(), tenantID(r), body.ThemeEditorEnabled, body.OnboardingEnabled)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleBusinessGet: GET /api/admin/business.
func (s *Server) handleBusinessGet(w http.ResponseWriter, r *http.Request) {
	b, err := s.settings.GetBusiness(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// handleBusinessPut: PUT /api/admin/business.
func (s *Server) handleBusinessPut(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	if !decodeJSON(w, r, &raw, s.cfg.MaxRequestBody) {
		return
	}
	b, err := s.settings.SaveBusiness(r.Context(), tenantID(r), raw)
	if err != nil {
		s.settingsError(w, r, err, "business.invalid")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// handlePublicSettings: GET /api/public/settings (published theme + business).
func (s *Server) handlePublicSettings(w http.ResponseWriter, r *http.Request) {
	pub, err := s.settings.Published(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	writeJSON(w, http.StatusOK, pub)
}

// handlePublicThemeCSS: GET /api/public/theme.css?v=<hash>. The current hash
// is cacheable forever; anything else revalidates through the ETag.
func (s *Server) handlePublicThemeCSS(w http.ResponseWriter, r *http.Request) {
	pub, err := s.settings.Published(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	css := settings.RenderCSS(&pub.Theme)
	w.Header().Set("ETag", `"`+pub.Hash+`"`)
	if v := r.URL.Query().Get("v"); v != "" && v == pub.Hash {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write([]byte(css))
}
