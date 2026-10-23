// Package settingsmod implements per-tenant brand settings (SPEC §8): draft,
// versioned publish, rollback and theme rendering.
//
// Every operation runs in database.WithTenantTx so the connection is bound
// to the tenant resolved by the middleware (RLS active).
package settingsmod

import (
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

// Service exposes a tenant's brand operations.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
	now  func() time.Time
}

// New builds the service.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service {
	return &Service{pool: pool, q: q, now: time.Now}
}

// State is the view returned by GET /api/admin/settings.
type State struct {
	Onboarded        bool             `json:"onboarded"`
	Draft            *json.RawMessage `json:"draft"`
	PublishedVersion int32            `json:"published_version"`
	Hash             string           `json:"hash"`
}

// Version is one history entry.
type Version struct {
	Version   int32  `json:"version"`
	Hash      string `json:"hash"`
	CreatedAt string `json:"created_at"`
}

// DraftResult is the response of PUT /api/admin/settings/draft.
type DraftResult struct {
	Draft    json.RawMessage   `json:"draft"`
	Warnings []settings.Detail `json:"warnings"`
}

// PublishResult is the response of POST /api/admin/settings/publish.
type PublishResult struct {
	Version int32  `json:"version"`
	Hash    string `json:"hash"`
}

// ErrVersionNotFound means the version to restore does not exist.
var ErrVersionNotFound = errors.New("version not found")

// defaultSettings is used when the tenant never configured anything.
func defaultSettings() json.RawMessage {
	return json.RawMessage(`{"name":"Barbería","tagline":"","logo":"","c":{"primary":"#c9a24b","accent":"#e8d9a8","bg":"#161616","surface":"#222222","text":"#f4f0e8"},"hf":"Inter","bf":"Inter","r":8,"btn":"solid","biz":{"whatsapp":"","instagram":"","address":"","currency":"ARS","slot":30,"notice":1,"open":"09:00","close":"20:00","tz":"America/Argentina/Buenos_Aires"}}`)
}

// GetAdmin returns the tenant state (defaults when there is no row).
func (s *Service) GetAdmin(ctx context.Context, tenantID int64) (*State, error) {
	var st *State
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetTenantSettings(ctx, tenantID)
		if errors.Is(err, pgx.ErrNoRows) {
			def := defaultSettings()
			st = &State{Onboarded: false, Draft: &def, PublishedVersion: 0, Hash: settings.Hash(settings.RenderCSS(mustSettings(def)))}
			return nil
		}
		if err != nil {
			return err
		}
		e := &State{Onboarded: row.Onboarded, PublishedVersion: row.PublishedVersion}
		if len(row.Draft) > 0 {
			d := json.RawMessage(row.Draft)
			e.Draft = &d
		}
		if len(row.Published) > 0 {
			e.Hash = settings.Hash(settings.RenderCSS(mustSettings(json.RawMessage(row.Published))))
		} else if len(row.Draft) > 0 {
			e.Hash = settings.Hash(settings.RenderCSS(mustSettings(json.RawMessage(row.Draft))))
		}
		st = e
		return nil
	})
	if err != nil {
		return nil, err
	}
	return st, nil
}

// SaveDraft validates and stores the draft. The published copy is untouched.
func (s *Service) SaveDraft(ctx context.Context, tenantID int64, raw json.RawMessage) (*DraftResult, error) {
	res := settings.Validate(raw)
	if len(res.Details) > 0 {
		return nil, &ValidationError{Details: res.Details}
	}
	normalized, err := json.Marshal(res.Settings)
	if err != nil {
		return nil, err
	}
	err = database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		_, err := q.UpsertTenantSettingsDraft(ctx, sqlc.UpsertTenantSettingsDraftParams{
			TenantID: tenantID,
			Draft:    normalized,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	warnings := res.Warnings
	if warnings == nil {
		warnings = []settings.Detail{}
	}
	return &DraftResult{Draft: normalized, Warnings: warnings}, nil
}

// Publish copies the draft to published as version N+1. Idempotent when the hash did not change.
func (s *Service) Publish(ctx context.Context, tenantID int64) (*PublishResult, error) {
	var out *PublishResult
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetTenantSettings(ctx, tenantID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoDraft
		}
		if err != nil {
			return err
		}
		if len(row.Draft) == 0 {
			return ErrNoDraft
		}
		draft := json.RawMessage(row.Draft)
		css := settings.RenderCSS(mustSettings(draft))
		hash := settings.Hash(css)

		// Idempotent: same published hash means no new version.
		if len(row.Published) > 0 {
			pubHash := settings.Hash(settings.RenderCSS(mustSettings(json.RawMessage(row.Published))))
			if pubHash == hash {
				out = &PublishResult{Version: row.PublishedVersion, Hash: hash}
				return nil
			}
		}

		newVersion := row.PublishedVersion + 1
		if _, err := q.InsertSettingsVersion(ctx, sqlc.InsertSettingsVersionParams{
			TenantID: tenantID,
			Version:  newVersion,
			Hash:     hash,
			Settings: draft,
		}); err != nil {
			return err
		}
		upd, err := q.PublishTenantSettings(ctx, tenantID)
		if err != nil {
			return err
		}
		out = &PublishResult{Version: upd.PublishedVersion, Hash: hash}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ErrNoDraft means there is no draft to publish.
var ErrNoDraft = errors.New("no draft to publish")

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
			out = append(out, Version{
				Version:   r.Version,
				Hash:      r.Hash,
				CreatedAt: r.CreatedAt.Time.UTC().Format(time.RFC3339),
			})
		}
		return nil
	})
	return out, err
}

// Rollback creates a NEW version with the content of the chosen one (history is never rewritten).
func (s *Service) Rollback(ctx context.Context, tenantID int64, version int32) (*PublishResult, error) {
	var out *PublishResult
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetSettingsVersion(ctx, sqlc.GetSettingsVersionParams{
			TenantID: tenantID,
			Version:  version,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrVersionNotFound
		}
		if err != nil {
			return err
		}
		content := json.RawMessage(row.Settings)
		hash := settings.Hash(settings.RenderCSS(mustSettings(content)))
		// The new version number builds on the current published_version.
		cur, err := q.GetTenantSettings(ctx, tenantID)
		if err != nil {
			return err
		}
		newVersion := cur.PublishedVersion + 1
		if _, err := q.InsertSettingsVersion(ctx, sqlc.InsertSettingsVersionParams{
			TenantID: tenantID,
			Version:  newVersion,
			Hash:     hash,
			Settings: content,
		}); err != nil {
			return err
		}
		if _, err := q.UpsertTenantSettingsDraft(ctx, sqlc.UpsertTenantSettingsDraftParams{
			TenantID: tenantID,
			Draft:    content,
		}); err != nil {
			return err
		}
		upd, err := q.PublishTenantSettings(ctx, tenantID)
		if err != nil {
			return err
		}
		out = &PublishResult{Version: upd.PublishedVersion, Hash: hash}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Published returns the tenant's published settings (or defaults).
func (s *Service) Published(ctx context.Context, tenantID int64) (json.RawMessage, error) {
	var out json.RawMessage
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetTenantSettings(ctx, tenantID)
		if errors.Is(err, pgx.ErrNoRows) {
			out = defaultSettings()
			return nil
		}
		if err != nil {
			return err
		}
		if len(row.Published) > 0 {
			out = json.RawMessage(row.Published)
		} else {
			out = defaultSettings()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Draft returns the draft (or defaults).
func (s *Service) Draft(ctx context.Context, tenantID int64) (json.RawMessage, error) {
	var out json.RawMessage
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetTenantSettings(ctx, tenantID)
		if errors.Is(err, pgx.ErrNoRows) {
			out = defaultSettings()
			return nil
		}
		if err != nil {
			return err
		}
		if len(row.Draft) > 0 {
			out = json.RawMessage(row.Draft)
		} else {
			out = defaultSettings()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ValidationError groups the settings validation errors.
type ValidationError struct {
	Details []settings.Detail
}

func (e *ValidationError) Error() string { return "invalid settings" }

// mustSettings decodes already-validated settings (the defaults validate too).
func mustSettings(raw json.RawMessage) *settings.Settings {
	res := settings.Validate(raw)
	if res.Settings == nil {
		// Fallback that should never run, because the defaults validate.
		return &settings.Settings{
			C:   settings.Colors{Primary: "#c9a24b", Accent: "#e8d9a8", Bg: "#161616", Surface: "#222222", Text: "#f4f0e8"},
			HF:  "Inter",
			BF:  "Inter",
			R:   8,
			Btn: "solid",
		}
	}
	return res.Settings
}
