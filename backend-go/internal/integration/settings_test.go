//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
)

const validTheme = `{"v":2,"name":"Test","tagline":"","logo":"","preset":"classic","mode":"dark","brand":"#c9a24b","font":{"heading":"Playfair Display","body":"Inter"},"shape":{"radius":8,"button":"solid"},"overrides":{"light":{},"dark":{}}}`

// themeWith replaces a fragment of validTheme.
func themeWith(old, new string) string { return strings.Replace(validTheme, old, new, 1) }

// SEC-03: no /api/admin/* route works without authentication.
func TestAdminSettingsRequireAuth(t *testing.T) {
	e := newEnv(t)
	for _, p := range []string{"/api/admin/settings", "/api/admin/settings/versions", "/api/admin/theme/presets", "/api/admin/business"} {
		e.call(t, http.MethodGet, p, "", "").expect(http.StatusUnauthorized, p)
	}
}

// Staff cannot reach any site setting.
func TestAdminSettingsForbiddenForStaff(t *testing.T) {
	e := newEnv(t)
	m, _ := auth.NewJWTManager(jwtSecret, 3600e9)
	staff, err := m.Issue(999, "staff@example.com", e.tenantID, auth.RoleStaff)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"/api/admin/settings", "/api/admin/business", "/api/admin/theme/presets"} {
		e.call(t, http.MethodGet, p, "", staff).expect(http.StatusForbidden, p)
	}
}

// The current theme hash is cacheable forever; anything else revalidates.
func TestThemeCacheHeaders(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	e.call(t, http.MethodPut, "/api/admin/settings/draft", validTheme, token).expect(200, "draft")
	e.call(t, http.MethodPost, "/api/admin/settings/publish", "", token).expect(200, "publish")

	res := e.call(t, http.MethodGet, "/api/public/theme.css", "", "").expect(200, "theme")
	hash := strings.Trim(res.Header().Get("ETag"), `"`)
	if hash == "" {
		t.Fatal("missing ETag")
	}
	if cc := res.Header().Get("Cache-Control"); !strings.Contains(cc, "no-cache") {
		t.Fatalf("without v: Cache-Control = %q", cc)
	}
	if cc := e.call(t, http.MethodGet, "/api/public/theme.css?v="+hash, "", "").Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Fatalf("with v=hash: Cache-Control = %q", cc)
	}
	if ct := res.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Fatalf("Content-Type = %q", ct)
	}
}

func TestPublishIsIdempotentButNotForIdentityChanges(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	versions := func() int {
		var vs []struct{ Version int32 }
		e.call(t, http.MethodGet, "/api/admin/settings/versions", "", token).expect(200, "versions").json(&vs)
		return len(vs)
	}
	e.call(t, http.MethodPut, "/api/admin/settings/draft", validTheme, token).expect(200, "draft")
	e.call(t, http.MethodPost, "/api/admin/settings/publish", "", token).expect(200, "publish 1")
	e.call(t, http.MethodPost, "/api/admin/settings/publish", "", token).expect(200, "publish 2")
	if got := versions(); got != 1 {
		t.Fatalf("publishing the same draft twice created %d versions", got)
	}
	// Renaming changes no color, but it is still a change to publish.
	e.call(t, http.MethodPut, "/api/admin/settings/draft", themeWith(`"name":"Test"`, `"name":"Renamed"`), token).expect(200, "draft")
	e.call(t, http.MethodPost, "/api/admin/settings/publish", "", token).expect(200, "publish 3")
	if got := versions(); got != 2 {
		t.Fatalf("a rename must publish a new version, got %d versions", got)
	}
	var pub struct {
		Theme struct{ Name string } `json:"theme"`
	}
	e.call(t, http.MethodGet, "/api/public/settings", "", "").expect(200, "public").json(&pub)
	if pub.Theme.Name != "Renamed" {
		t.Fatalf("public name = %q", pub.Theme.Name)
	}
}

func TestInvalidDraftListsFields(t *testing.T) {
	e := newEnv(t)
	res := e.call(t, http.MethodPut, "/api/admin/settings/draft", themeWith(`"Playfair Display"`, `"Comic Sans"`), e.ownerToken(t)).
		expect(http.StatusUnprocessableEntity, "invalid draft")
	var p struct {
		Code   string `json:"code"`
		Errors []struct{ Field, Code string }
	}
	res.json(&p)
	if p.Code != "settings.invalid" || len(p.Errors) != 1 || p.Errors[0].Field != "font.heading" || p.Errors[0].Code != "invalid_font" {
		t.Fatalf("problem = %+v", p)
	}
}

// The editor renders unsaved themes through the API and nothing is stored.
func TestPreviewRendersWithoutSaving(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	var out struct {
		CSS  string `json:"css"`
		Hash string `json:"hash"`
	}
	e.call(t, http.MethodPost, "/api/admin/theme/preview", themeWith(`"brand":"#c9a24b"`, `"brand":"#2f5d8a"`), token).expect(200, "preview").json(&out)
	if !strings.Contains(out.CSS, "--color-primary:#2f5d8a") || out.Hash == "" {
		t.Fatalf("preview = %+v", out)
	}
	var state struct {
		Draft struct{ Brand string } `json:"draft"`
	}
	e.call(t, http.MethodGet, "/api/admin/settings", "", token).expect(200, "state").json(&state)
	if state.Draft.Brand == "#2f5d8a" {
		t.Fatal("preview must not save the draft")
	}
	e.call(t, http.MethodPost, "/api/admin/theme/preview", themeWith(`"mode":"dark"`, `"mode":"sepia"`), token).expect(422, "invalid preview")
}

func TestPresetsComeWithGeneratedPalettes(t *testing.T) {
	e := newEnv(t)
	var presets []struct {
		ID    string `json:"id"`
		Theme struct{ Brand, Mode string }
		Light struct{ Bg, Text, Primary string }
		Dark  struct{ Bg, Text, Primary string }
	}
	e.call(t, http.MethodGet, "/api/admin/theme/presets", "", e.ownerToken(t)).expect(200, "presets").json(&presets)
	if len(presets) < 4 || presets[0].ID != "classic" {
		t.Fatalf("presets = %+v", presets)
	}
	for _, p := range presets {
		if p.Light.Bg == "" || p.Dark.Bg == "" || p.Light.Primary == "" || p.Theme.Brand == "" {
			t.Errorf("preset %s is incomplete: %+v", p.ID, p)
		}
	}
}

func TestPreferencesDefaultOnAndCanBeSwitchedOff(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	type state struct {
		Preferences struct {
			ThemeEditorEnabled bool `json:"themeEditorEnabled"`
			OnboardingEnabled  bool `json:"onboardingEnabled"`
		} `json:"preferences"`
	}
	var s state
	e.call(t, http.MethodGet, "/api/admin/settings", "", token).expect(200, "state").json(&s)
	if !s.Preferences.ThemeEditorEnabled || !s.Preferences.OnboardingEnabled {
		t.Fatalf("defaults must be on: %+v", s.Preferences)
	}
	e.call(t, http.MethodPatch, "/api/admin/settings/preferences", `{"themeEditorEnabled":false}`, token).expect(200, "patch")
	e.call(t, http.MethodGet, "/api/admin/settings", "", token).expect(200, "state").json(&s)
	if s.Preferences.ThemeEditorEnabled || !s.Preferences.OnboardingEnabled {
		t.Fatalf("only the editor switch should change: %+v", s.Preferences)
	}
	// Switching the menu off does not touch what the public site shows.
	e.call(t, http.MethodGet, "/api/public/theme.css", "", "").expect(200, "public css")
}

func TestBusinessIsSavedDirectlyAndServedPublicly(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	const business = `{"whatsapp":"+54 9 3492 123456","instagram":"@donmateo","address":"Calle 1","currency":"ARS","slot":30,"notice":1,"open":"10:00","close":"19:00","tz":""}`
	var saved struct{ Open, Tz string }
	e.call(t, http.MethodPut, "/api/admin/business", business, token).expect(200, "save").json(&saved)
	if saved.Open != "10:00" || saved.Tz != "America/Argentina/Buenos_Aires" {
		t.Fatalf("saved = %+v", saved)
	}
	var pub struct {
		Business struct{ Open, Close string } `json:"business"`
	}
	e.call(t, http.MethodGet, "/api/public/settings", "", "").expect(200, "public").json(&pub)
	if pub.Business.Open != "10:00" || pub.Business.Close != "19:00" {
		t.Fatalf("public business = %+v", pub.Business)
	}
	e.call(t, http.MethodPut, "/api/admin/business", strings.Replace(business, `"slot":30`, `"slot":5`, 1), token).expect(422, "invalid business")
}

// Documents stored by the previous version keep working: they are upgraded on
// read, rollback to them works, and business data moves to its own column.
func TestLegacyV1DocumentsAreUpgraded(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	token := e.ownerToken(t)
	const v1 = `{"name":"Old Shop","tagline":"since 1990","logo":"","c":{"primary":"#336699","accent":"#aabbcc","bg":"#101010","surface":"#202020","text":"#f0f0f0"},"hf":"Lora","bf":"Inter","r":6,"btn":"outline","biz":{"whatsapp":"","instagram":"","address":"","currency":"ARS","slot":45,"notice":2,"open":"11:00","close":"18:00","tz":"America/Argentina/Buenos_Aires"}}`
	if _, err := e.pool.Exec(ctx, `
		INSERT INTO tenant_settings (tenant_id, draft, published, published_version, onboarded)
		VALUES ($1, $2::jsonb, $2::jsonb, 1, TRUE)
		ON CONFLICT (tenant_id) DO UPDATE SET draft = EXCLUDED.draft, published = EXCLUDED.published, published_version = 1, onboarded = TRUE`,
		e.tenantID, v1); err != nil {
		t.Fatal(err)
	}
	if _, err := e.pool.Exec(ctx, `INSERT INTO tenant_settings_versions (tenant_id, version, hash, settings) VALUES ($1, 1, 'legacy', $2::jsonb)`, e.tenantID, v1); err != nil {
		t.Fatal(err)
	}

	var state struct {
		Draft struct {
			V     int    `json:"v"`
			Name  string `json:"name"`
			Mode  string `json:"mode"`
			Brand string `json:"brand"`
			Shape struct {
				Radius int    `json:"radius"`
				Button string `json:"button"`
			} `json:"shape"`
		} `json:"draft"`
	}
	e.call(t, http.MethodGet, "/api/admin/settings", "", token).expect(200, "state").json(&state)
	if state.Draft.V != 2 || state.Draft.Name != "Old Shop" || state.Draft.Mode != "dark" || state.Draft.Brand != "#336699" ||
		state.Draft.Shape.Radius != 6 || state.Draft.Shape.Button != "outline" {
		t.Fatalf("upgraded draft = %+v", state.Draft)
	}

	var pub struct {
		Business struct{ Slot int } `json:"business"`
	}
	e.call(t, http.MethodGet, "/api/public/settings", "", "").expect(200, "public").json(&pub)
	if pub.Business.Slot != 45 {
		t.Fatalf("business should come from the v1 document, got %+v", pub.Business)
	}
	if css := e.call(t, http.MethodGet, "/api/public/theme.css", "", "").Body.String(); !strings.Contains(css, "--color-bg:#101010") {
		t.Fatalf("the v1 palette must survive as overrides: %s", css)
	}

	// Rolling back to a v1 version publishes a v2 document.
	var rolled struct{ Version int32 }
	e.call(t, http.MethodPost, "/api/admin/settings/rollback", `{"version":1}`, token).expect(200, "rollback").json(&rolled)
	if rolled.Version != 2 {
		t.Fatalf("rollback must create version 2, got %d", rolled.Version)
	}
	var stored string
	if err := e.pool.QueryRow(ctx, `SELECT settings->>'v' FROM tenant_settings_versions WHERE version = 2`).Scan(&stored); err != nil || stored != "2" {
		t.Fatalf("stored version 2 has v=%q err=%v", stored, err)
	}
	e.call(t, http.MethodPost, "/api/admin/settings/rollback", `{"version":99}`, token).expect(404, "unknown version")
}

func TestRollbackRestoresAnEarlierTheme(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	for _, brand := range []string{"#c9a24b", "#2f5d8a"} {
		e.call(t, http.MethodPut, "/api/admin/settings/draft", themeWith(`"brand":"#c9a24b"`, fmt.Sprintf(`"brand":%q`, brand)), token).expect(200, "draft")
		e.call(t, http.MethodPost, "/api/admin/settings/publish", "", token).expect(200, "publish")
	}
	e.call(t, http.MethodPost, "/api/admin/settings/rollback", `{"version":1}`, token).expect(200, "rollback")
	if css := e.call(t, http.MethodGet, "/api/public/theme.css", "", "").Body.String(); !strings.Contains(css, "--color-primary:#c9a24b") {
		t.Fatalf("rollback did not restore the first theme: %s", css)
	}
	var vs []struct{ Version int32 }
	e.call(t, http.MethodGet, "/api/admin/settings/versions", "", token).expect(200, "versions").json(&vs)
	if len(vs) != 3 {
		t.Fatalf("history must grow, got %d versions", len(vs))
	}
}

func TestPublishWithoutDraftIsRejected(t *testing.T) {
	e := newEnv(t)
	if code := e.call(t, http.MethodPost, "/api/admin/settings/publish", "", e.ownerToken(t)).expect(400, "no draft").problemCode(); code != "settings.no_draft" {
		t.Fatalf("code = %q", code)
	}
}
