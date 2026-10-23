// Package settings: branding/config por tenant (schema-per-tenant).
// Asume que un middleware ya resolvió el tenant, hizo SET search_path y
// guardó un Querier en el contexto (ver QFrom). El rol admin lo valida tu middleware de auth.
//
// Migración (una vez por schema, vía Flyway/golang-migrate):
//
//	CREATE TABLE tenant_settings (
//	  id         smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
//	  data       jsonb       NOT NULL,
//	  onboarded  boolean     NOT NULL DEFAULT false,
//	  updated_at timestamptz NOT NULL DEFAULT now()
//	);
package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type Querier interface {
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
}

type ctxKey struct{}

// QFrom devuelve la conexión ya apuntada al schema del tenant (la pone tu middleware).
func QFrom(r *http.Request) Querier { return r.Context().Value(ctxKey{}).(Querier) }

type Business struct {
	Whatsapp  string `json:"whatsapp"`
	Instagram string `json:"instagram"`
	Address   string `json:"address"`
	Currency  string `json:"currency"`
	Slot      int    `json:"slot"`   // minutos por turno
	Notice    int    `json:"notice"` // anticipación mínima en horas
	Open      string `json:"open"`
	Close     string `json:"close"`
}

type Settings struct {
	Name    string            `json:"name"`
	Tagline string            `json:"tagline"`
	Logo    string            `json:"logo"` // https://... o data:image/png|jpeg|webp
	Colors  map[string]string `json:"c"`
	HFont   string            `json:"hf"`
	BFont   string            `json:"bf"`
	Radius  int               `json:"r"`
	Button  string            `json:"btn"`
	Biz     Business          `json:"biz"`
}

var (
	hexRe  = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	fontRe = regexp.MustCompile(`^[A-Za-z0-9 ]{2,40}$`) // evita inyección CSS/URL
	timeRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)
	logoRe = regexp.MustCompile(`^(https://\S+|data:image/(png|jpeg|webp);base64,[A-Za-z0-9+/=]+)$`) // sin SVG (XSS)
	colorKeys = []string{"primary", "accent", "bg", "surface", "text"}
)

func (s *Settings) Validate() error {
	if n := len(strings.TrimSpace(s.Name)); n < 2 || n > 60 {
		return errors.New("name: 2-60 caracteres")
	}
	if len(s.Tagline) > 120 {
		return errors.New("tagline: máx 120")
	}
	if s.Logo != "" && (len(s.Logo) > 400_000 || !logoRe.MatchString(s.Logo)) {
		return errors.New("logo inválido (https o PNG/JPEG/WebP ≤ 300 KB)")
	}
	for _, k := range colorKeys {
		if !hexRe.MatchString(s.Colors[k]) {
			return fmt.Errorf("color %q inválido", k)
		}
	}
	if !fontRe.MatchString(s.HFont) || !fontRe.MatchString(s.BFont) {
		return errors.New("tipografía inválida")
	}
	if s.Radius < 0 || s.Radius > 24 {
		return errors.New("r: 0-24")
	}
	if s.Button != "solid" && s.Button != "outline" {
		return errors.New("btn inválido")
	}
	b := s.Biz
	if !timeRe.MatchString(b.Open) || !timeRe.MatchString(b.Close) || b.Open >= b.Close {
		return errors.New("horario inválido")
	}
	if b.Slot < 10 || b.Slot > 240 || b.Notice < 0 || b.Notice > 168 {
		return errors.New("slot/anticipación fuera de rango")
	}
	return nil
}

func Default() Settings {
	return Settings{
		Name: "Mi Barbería", Tagline: "Reservá tu turno online",
		Colors: map[string]string{"primary": "#c9a24b", "accent": "#e8d9a8", "bg": "#161616", "surface": "#222222", "text": "#f4f0e8"},
		HFont: "Playfair Display", BFont: "Inter", Radius: 8, Button: "solid",
		Biz: Business{Currency: "ARS", Slot: 30, Notice: 1, Open: "09:00", Close: "20:00"},
	}
}

type resp struct {
	Onboarded bool     `json:"onboarded"`
	Settings  Settings `json:"settings"`
}

func load(r *http.Request) (resp, error) {
	var raw []byte
	out := resp{Settings: Default()}
	err := QFrom(r).QueryRowContext(r.Context(), `SELECT data, onboarded FROM tenant_settings WHERE id = 1`).Scan(&raw, &out.Onboarded)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil // primer login: wizard con defaults
	}
	if err != nil {
		return out, err
	}
	return out, json.Unmarshal(raw, &out.Settings)
}

func write(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// GET /api/admin/settings  (auth admin) -> {onboarded, settings}. El front abre el wizard si onboarded=false.
func Get(w http.ResponseWriter, r *http.Request) {
	out, err := load(r)
	if err != nil {
		write(w, 500, map[string]string{"error": "no se pudo leer la configuración"})
		return
	}
	write(w, 200, out)
}

// PUT /api/admin/settings  (auth admin). Marca onboarded=true: sirve para el wizard y para ediciones posteriores.
func Put(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var s Settings
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		write(w, 400, map[string]string{"error": "JSON inválido"})
		return
	}
	if err := s.Validate(); err != nil {
		write(w, 422, map[string]string{"error": err.Error()})
		return
	}
	data, _ := json.Marshal(s)
	_, err := QFrom(r).ExecContext(r.Context(), `
		INSERT INTO tenant_settings (id, data, onboarded) VALUES (1, $1, true)
		ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, onboarded = true, updated_at = now()`, data)
	if err != nil {
		write(w, 500, map[string]string{"error": "no se pudo guardar"})
		return
	}
	write(w, 200, resp{Onboarded: true, Settings: s})
}

// GET /api/public/settings (sin auth; el tenant sale del subdominio/slug). Lo consume la página de reservas.
func Public(w http.ResponseWriter, r *http.Request) {
	out, err := load(r)
	if err != nil {
		write(w, 500, map[string]string{"error": "error"})
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	write(w, 200, out.Settings)
}

// Registro (Go 1.22+):
//   mux.Handle("GET /api/admin/settings", auth(tenant(http.HandlerFunc(settings.Get))))
//   mux.Handle("PUT /api/admin/settings", auth(tenant(http.HandlerFunc(settings.Put))))
//   mux.Handle("GET /api/public/settings", tenant(http.HandlerFunc(settings.Public)))
