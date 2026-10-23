// Package settingsmod implements the per-tenant site settings: the theme
// (draft, versioned publish, rollback, live preview), the business data and
// the owner's admin-panel preferences.
//
// Every database operation runs in database.WithTenantTx so the connection is
// bound to the tenant resolved by the middleware (RLS active).
package settingsmod

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/settings"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service exposes a tenant's settings operations.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

// New builds the service.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service {
	return &Service{pool: pool, q: q}
}

// Preferences are the owner's switches for the admin panel.
type Preferences struct {
	ThemeEditorEnabled bool `json:"themeEditorEnabled"`
	OnboardingEnabled  bool `json:"onboardingEnabled"`
}

// State is the view returned by GET /api/admin/settings.
type State struct {
	Onboarded        bool           `json:"onboarded"`
	Draft            settings.Theme `json:"draft"`
	PublishedVersion int32          `json:"publishedVersion"`
	Hash             string         `json:"hash"`
	Preferences      Preferences    `json:"preferences"`
}

// Version is one history entry.
type Version struct {
	Version   int32  `json:"version"`
	Hash      string `json:"hash"`
	CreatedAt string `json:"createdAt"`
}

// DraftResult is the response of PUT /api/admin/settings/draft.
type DraftResult struct {
	Draft    settings.Theme    `json:"draft"`
	Warnings []settings.Detail `json:"warnings"`
}

// PublishResult is the response of publish and rollback.
type PublishResult struct {
	Version int32  `json:"version"`
	Hash    string `json:"hash"`
}

// PreviewResult is the stylesheet of an unsaved theme.
type PreviewResult struct {
	CSS      string            `json:"css"`
	Hash     string            `json:"hash"`
	Warnings []settings.Detail `json:"warnings"`
}

// PresetView is a preset with its generated palettes, for the gallery.
type PresetView struct {
	ID    string          `json:"id"`
	Theme settings.Theme  `json:"theme"`
	Light settings.Scheme `json:"light"`
	Dark  settings.Scheme `json:"dark"`
}

// Public is what the public site needs: the published theme and the
// business data.
type Public struct {
	Theme    settings.Theme    `json:"theme"`
	Business settings.Business `json:"business"`
	Hash     string            `json:"hash"`
}

var (
	// ErrVersionNotFound means the version to restore does not exist.
	ErrVersionNotFound = errors.New("version not found")
	// ErrNoDraft means there is no draft to publish.
	ErrNoDraft = errors.New("no draft to publish")
)

// ValidationError groups the validation errors of a settings document.
type ValidationError struct {
	Details []settings.Detail
}

func (e *ValidationError) Error() string { return "invalid settings" }

// loadTheme reads a stored document (v1 or v2); anything unreadable falls
// back to the default theme so a damaged row can never take the site down.
func loadTheme(raw []byte) settings.Theme {
	if len(raw) == 0 {
		return settings.DefaultTheme()
	}
	t, err := settings.NormalizeTheme(raw)
	if err != nil {
		return settings.DefaultTheme()
	}
	return t
}

func themeHash(t settings.Theme) string { return settings.Hash(settings.RenderCSS(&t)) }

func canonical(t settings.Theme) []byte {
	b, _ := json.Marshal(t)
	return b
}

// rowPreferences reads the switches of a settings row.
func rowPreferences(themeEditor, onboarding bool) Preferences {
	return Preferences{ThemeEditorEnabled: themeEditor, OnboardingEnabled: onboarding}
}

// getRow reads the settings row, or the defaults when there is none.
func getRow(ctx context.Context, q *sqlc.Queries, tenantID int64) (sqlc.TenantSetting, error) {
	row, err := q.GetTenantSettings(ctx, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.TenantSetting{TenantID: tenantID, ThemeEditorEnabled: true, OnboardingEnabled: true}, nil
	}
	return row, err
}

// GetAdmin returns the tenant's state (defaults when there is no row).
func (s *Service) GetAdmin(ctx context.Context, tenantID int64) (*State, error) {
	var out *State
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := getRow(ctx, q, tenantID)
		if err != nil {
			return err
		}
		draft := loadTheme(row.Draft)
		if len(row.Draft) == 0 {
			draft = loadTheme(row.Published)
		}
		hash := themeHash(draft)
		if len(row.Published) > 0 {
			hash = themeHash(loadTheme(row.Published))
		}
		out = &State{
			Onboarded: row.Onboarded, Draft: draft, PublishedVersion: row.PublishedVersion,
			Hash: hash, Preferences: rowPreferences(row.ThemeEditorEnabled, row.OnboardingEnabled),
		}
		return nil
	})
	return out, err
}

// SaveDraft validates and stores the draft. The published copy is untouched.
func (s *Service) SaveDraft(ctx context.Context, tenantID int64, raw json.RawMessage) (*DraftResult, error) {
	res := settings.ValidateTheme(raw)
	if res.Theme == nil {
		return nil, &ValidationError{Details: res.Details}
	}
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		_, err := q.UpsertTenantSettingsDraft(ctx, sqlc.UpsertTenantSettingsDraftParams{
			TenantID: tenantID, Draft: canonical(*res.Theme),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &DraftResult{Draft: *res.Theme, Warnings: res.Warnings}, nil
}

// Preview renders an unsaved theme without touching the database.
func (s *Service) Preview(raw json.RawMessage) (*PreviewResult, error) {
	res := settings.ValidateTheme(raw)
	if res.Theme == nil {
		return nil, &ValidationError{Details: res.Details}
	}
	css := settings.RenderCSS(res.Theme)
	return &PreviewResult{CSS: css, Hash: settings.Hash(css), Warnings: res.Warnings}, nil
}

// Presets lists the starting points with their generated palettes.
func (s *Service) Presets() []PresetView {
	presets := settings.Presets()
	out := make([]PresetView, 0, len(presets))
	for _, p := range presets {
		theme := p.Theme
		theme.Name = settings.DefaultTheme().Name
		light, dark := settings.Generate(theme.Brand, theme.Overrides)
		out = append(out, PresetView{ID: p.ID, Theme: theme, Light: light, Dark: dark})
	}
	return out
}

// Publish copies the draft to published as version N+1. Publishing a draft
// identical to the published theme is a no-op.
func (s *Service) Publish(ctx context.Context, tenantID int64) (*PublishResult, error) {
	var out *PublishResult
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetTenantSettings(ctx, tenantID)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && len(row.Draft) == 0) {
			return ErrNoDraft
		}
		if err != nil {
			return err
		}
		draft := loadTheme(row.Draft)
		if len(row.Published) > 0 && bytes.Equal(canonical(loadTheme(row.Published)), canonical(draft)) {
			out = &PublishResult{Version: row.PublishedVersion, Hash: themeHash(draft)}
			return nil
		}
		out, err = publishTheme(ctx, q, tenantID, row.PublishedVersion+1, draft)
		return err
	})
	return out, err
}

// publishTheme stores version n and makes the draft the published theme.
func publishTheme(ctx context.Context, q *sqlc.Queries, tenantID int64, n int32, theme settings.Theme) (*PublishResult, error) {
	hash := themeHash(theme)
	if _, err := q.InsertSettingsVersion(ctx, sqlc.InsertSettingsVersionParams{
		TenantID: tenantID, Version: n, Hash: hash, Settings: canonical(theme),
	}); err != nil {
		return nil, err
	}
	updated, err := q.PublishTenantSettings(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &PublishResult{Version: updated.PublishedVersion, Hash: hash}, nil
}

// Versions returns the last 20 versions.
func (s *Service) Versions(ctx context.Context, tenantID int64) ([]Version, error) {
	var out []Version
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListSettingsVersions(ctx, tenantID)
		if err != nil {
			return err
		}
		out = make([]Version, 0, len(rows))
		for _, r := range rows {
			out = append(out, Version{Version: r.Version, Hash: r.Hash, CreatedAt: r.CreatedAt.Time.UTC().Format(time.RFC3339)})
		}
		return nil
	})
	return out, err
}

// Rollback creates a NEW version with the content of the chosen one (history
// is never rewritten). A version stored as v1 is upgraded on the way.
func (s *Service) Rollback(ctx context.Context, tenantID int64, version int32) (*PublishResult, error) {
	var out *PublishResult
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetSettingsVersion(ctx, sqlc.GetSettingsVersionParams{TenantID: tenantID, Version: version})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrVersionNotFound
		}
		if err != nil {
			return err
		}
		theme := loadTheme(row.Settings)
		current, err := getRow(ctx, q, tenantID)
		if err != nil {
			return err
		}
		if _, err := q.UpsertTenantSettingsDraft(ctx, sqlc.UpsertTenantSettingsDraftParams{
			TenantID: tenantID, Draft: canonical(theme),
		}); err != nil {
			return err
		}
		out, err = publishTheme(ctx, q, tenantID, current.PublishedVersion+1, theme)
		return err
	})
	return out, err
}

// UpdatePreferences changes the owner's switches; nil keeps the current value.
func (s *Service) UpdatePreferences(ctx context.Context, tenantID int64, themeEditor, onboarding *bool) (*Preferences, error) {
	var out *Preferences
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.UpdateSettingsPreferences(ctx, sqlc.UpdateSettingsPreferencesParams{
			TenantID: tenantID, ThemeEditorEnabled: themeEditor, OnboardingEnabled: onboarding,
		})
		if err != nil {
			return err
		}
		p := rowPreferences(row.ThemeEditorEnabled, row.OnboardingEnabled)
		out = &p
		return nil
	})
	return out, err
}

// businessOf reads the business data of a row, falling back to the data the
// legacy v1 documents carried and finally to the defaults.
func businessOf(row sqlc.TenantSetting) settings.Business {
	if len(row.Business) > 0 {
		var b settings.Business
		if json.Unmarshal(row.Business, &b) == nil {
			return b
		}
	}
	for _, doc := range [][]byte{row.Published, row.Draft} {
		if len(doc) > 0 && settings.IsLegacy(doc) {
			if _, b, err := settings.UpgradeV1(doc); err == nil {
				return b
			}
		}
	}
	return settings.DefaultBusiness()
}

// GetBusiness returns the tenant's business data.
func (s *Service) GetBusiness(ctx context.Context, tenantID int64) (*settings.Business, error) {
	var out *settings.Business
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := getRow(ctx, q, tenantID)
		if err != nil {
			return err
		}
		b := businessOf(row)
		out = &b
		return nil
	})
	return out, err
}

// SaveBusiness validates and stores the business data.
func (s *Service) SaveBusiness(ctx context.Context, tenantID int64, raw json.RawMessage) (*settings.Business, error) {
	res := settings.ValidateBusiness(raw)
	if res.Business == nil {
		return nil, &ValidationError{Details: res.Details}
	}
	stored, _ := json.Marshal(res.Business)
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		_, err := q.UpsertBusiness(ctx, sqlc.UpsertBusinessParams{TenantID: tenantID, Business: stored})
		return err
	})
	if err != nil {
		return nil, err
	}
	return res.Business, nil
}

// Published returns what the public site shows: the published theme (or the
// default) and the business data.
func (s *Service) Published(ctx context.Context, tenantID int64) (*Public, error) {
	var out *Public
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := getRow(ctx, q, tenantID)
		if err != nil {
			return err
		}
		theme := loadTheme(row.Published)
		out = &Public{Theme: theme, Business: businessOf(row), Hash: themeHash(theme)}
		return nil
	})
	return out, err
}
