// Package settingsmod implementa la configuración de marca por tenant
// (SPEC §8): borrador, publicación versionada, rollback y render del tema.
//
// Todas las operaciones de escritura usan database.WithTenantTx para que la
// conexión quede atada al tenant resuelto por el middleware (RLS activa).
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

// Service expone las operaciones de marca de un tenant.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
	now  func() time.Time
}

// New construye el service.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service {
	return &Service{pool: pool, q: q, now: time.Now}
}

// Estado es la vista que devuelve GET /api/admin/settings.
type Estado struct {
	Onboarded        bool             `json:"onboarded"`
	Draft            *json.RawMessage `json:"draft"`
	PublishedVersion int32            `json:"published_version"`
	Hash             string           `json:"hash"`
}

// Version es una entrada del historial.
type Version struct {
	Version   int32  `json:"version"`
	Hash      string `json:"hash"`
	CreatedAt string `json:"created_at"`
}

// ResultadoDraft es la respuesta de PUT /api/admin/settings/draft.
type ResultadoDraft struct {
	Draft    json.RawMessage   `json:"draft"`
	Warnings []settings.Detail `json:"warnings"`
}

// ResultadoPublish es la respuesta de POST /api/admin/settings/publish.
type ResultadoPublish struct {
	Version int32  `json:"version"`
	Hash    string `json:"hash"`
}

// ErrVersionNoEncontrada indica que la versión a restaurar no existe.
var ErrVersionNoEncontrada = errors.New("versión no encontrada")

// defaultSettings son los valores por defecto cuando el tenant nunca configuró nada.
func defaultSettings() json.RawMessage {
	return json.RawMessage(`{"name":"Barbería","tagline":"","logo":"","c":{"primary":"#c9a24b","accent":"#e8d9a8","bg":"#161616","surface":"#222222","text":"#f4f0e8"},"hf":"Inter","bf":"Inter","r":8,"btn":"solid","biz":{"whatsapp":"","instagram":"","address":"","currency":"ARS","slot":30,"notice":1,"open":"09:00","close":"20:00","tz":"America/Argentina/Buenos_Aires"}}`)
}

// GetAdmin devuelve el estado del tenant (defaults si no hay fila).
func (s *Service) GetAdmin(ctx context.Context, tenantID int64) (*Estado, error) {
	var est *Estado
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetTenantSettings(ctx, tenantID)
		if errors.Is(err, pgx.ErrNoRows) {
			def := defaultSettings()
			est = &Estado{Onboarded: false, Draft: &def, PublishedVersion: 0, Hash: settings.Hash(settings.RenderCSS(mustSettings(def)))}
			return nil
		}
		if err != nil {
			return err
		}
		e := &Estado{Onboarded: row.Onboarded, PublishedVersion: row.PublishedVersion}
		if len(row.Draft) > 0 {
			d := json.RawMessage(row.Draft)
			e.Draft = &d
		}
		if len(row.Published) > 0 {
			e.Hash = settings.Hash(settings.RenderCSS(mustSettings(json.RawMessage(row.Published))))
		} else if len(row.Draft) > 0 {
			e.Hash = settings.Hash(settings.RenderCSS(mustSettings(json.RawMessage(row.Draft))))
		}
		est = e
		return nil
	})
	if err != nil {
		return nil, err
	}
	return est, nil
}

// SaveDraft valida y guarda el borrador. No toca el publicado.
func (s *Service) SaveDraft(ctx context.Context, tenantID int64, raw json.RawMessage) (*ResultadoDraft, error) {
	res := settings.Validate(raw)
	if len(res.Details) > 0 {
		return nil, &ErrorValidacion{Details: res.Details}
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
	return &ResultadoDraft{Draft: normalized, Warnings: warnings}, nil
}

// Publish copia el borrador al publicado, version N+1. Idempotente si el hash no cambio.
func (s *Service) Publish(ctx context.Context, tenantID int64) (*ResultadoPublish, error) {
	var out *ResultadoPublish
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetTenantSettings(ctx, tenantID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSinBorrador
		}
		if err != nil {
			return err
		}
		if len(row.Draft) == 0 {
			return ErrSinBorrador
		}
		draft := json.RawMessage(row.Draft)
		css := settings.RenderCSS(mustSettings(draft))
		hash := settings.Hash(css)

		// Idempotencia: si el hash publicado es el mismo, no se crea versión nueva.
		if len(row.Published) > 0 {
			pubHash := settings.Hash(settings.RenderCSS(mustSettings(json.RawMessage(row.Published))))
			if pubHash == hash {
				out = &ResultadoPublish{Version: row.PublishedVersion, Hash: hash}
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
		out = &ResultadoPublish{Version: upd.PublishedVersion, Hash: hash}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ErrSinBorrador indica que no hay borrador para publicar.
var ErrSinBorrador = errors.New("no hay borrador para publicar")

// Versions devuelve las últimas 20 versiones.
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

// Rollback crea una NUEVA versión con el contenido de la elegida (no reescribe historia).
func (s *Service) Rollback(ctx context.Context, tenantID int64, version int32) (*ResultadoPublish, error) {
	var out *ResultadoPublish
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetSettingsVersion(ctx, sqlc.GetSettingsVersionParams{
			TenantID: tenantID,
			Version:  version,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrVersionNoEncontrada
		}
		if err != nil {
			return err
		}
		content := json.RawMessage(row.Settings)
		hash := settings.Hash(settings.RenderCSS(mustSettings(content)))
		// La nueva versión se calcula sobre el published_version actual.
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
		out = &ResultadoPublish{Version: upd.PublishedVersion, Hash: hash}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Published devuelve el settings publicado (o defaults) del tenant.
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

// Draft devuelve el borrador (o defaults).
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

// ErrorValidacion agrupa los errores de validación del settings.
type ErrorValidacion struct {
	Details []settings.Detail
}

func (e *ErrorValidacion) Error() string { return "configuración inválida" }

// mustSettings decodifica un settings validado (los defaults también validas).
func mustSettings(raw json.RawMessage) *settings.Settings {
	res := settings.Validate(raw)
	if res.Settings == nil {
		// Default de emergencia: nunca debería ocurrir porque los defaults validan.
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
