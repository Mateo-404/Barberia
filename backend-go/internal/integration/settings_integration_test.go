//go:build integration

package integration

import (
	"net/http"
	"strings"
	"testing"
)

// SEC-03: ninguna ruta /api/admin/* funciona sin autenticación.
func TestSec03AdminSinAuth(t *testing.T) {
	e, _ := nuevoEntorno(t)
	for _, p := range []string{"/api/admin/settings", "/api/admin/settings/versions"} {
		rec := e.req(t, http.MethodGet, p, "", "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s sin token: %d %s", p, rec.Code, rec.Body.String())
		}
	}
}

// S-13: un staff recibe 403 en /api/admin/settings*.
func TestS13StaffForbidden(t *testing.T) {
	e, ctx := nuevoEntorno(t)
	var tid int64
	_ = e.pool.QueryRow(ctx, `SELECT id FROM tenants WHERE slug='default'`).Scan(&tid)
	jm := jwtDeTest(t)
	// El middleware solo valida el claim "rol"; no hace falta fila en BD.
	staffToken, err := jm.Issue(999, "staff@x.com", tid, "staff")
	if err != nil {
		t.Fatal(err)
	}
	rec := e.req(t, http.MethodGet, "/api/admin/settings", "", staffToken)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("staff en /api/admin/settings: %d %s", rec.Code, rec.Body.String())
	}
	_ = e.login(t)
}

// S-10: cache headers de /api/public/theme.css.
func TestS10ThemeCacheHeaders(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)
	// publicar un borrador para tener un hash
	rec := e.req(t, http.MethodPut, "/api/admin/settings/draft", `{"name":"Test","tagline":"","logo":"","c":{"primary":"#c9a24b","accent":"#e8d9a8","bg":"#161616","surface":"#222222","text":"#f4f0e8"},"hf":"Inter","bf":"Inter","r":8,"btn":"solid","biz":{"whatsapp":"","instagram":"","address":"","currency":"ARS","slot":30,"notice":1,"open":"09:00","close":"20:00","tz":"America/Argentina/Buenos_Aires"}}`, token)
	if rec.Code != 200 {
		t.Fatalf("draft: %d %s", rec.Code, rec.Body.String())
	}
	rec = e.req(t, http.MethodPost, "/api/admin/settings/publish", "", token)
	if rec.Code != 200 {
		t.Fatalf("publish: %d %s", rec.Code, rec.Body.String())
	}
	var pub struct {
		Hash string `json:"hash"`
	}
	// obtener hash del public settings? usar preview css hash via ETag
	rec = e.req(t, http.MethodGet, "/api/public/theme.css", "", "")
	if rec.Code != 200 {
		t.Fatalf("theme.css: %d %s", rec.Code, rec.Body.String())
	}
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("ETag ausente")
	}
	hash := strings.Trim(etag, `"`)
	// con v=hash → immutable
	rec2 := e.req(t, http.MethodGet, "/api/public/theme.css?v="+hash, "", "")
	if !strings.Contains(rec2.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("immutable ausente: %s", rec2.Header().Get("Cache-Control"))
	}
	// sin v → no-cache
	rec3 := e.req(t, http.MethodGet, "/api/public/theme.css", "", "")
	if !strings.Contains(rec3.Header().Get("Cache-Control"), "no-cache") {
		t.Fatalf("no-cache ausente: %s", rec3.Header().Get("Cache-Control"))
	}
	_ = pub
}

// S-08: publicar dos veces lo mismo no crea versión nueva.
func TestS08PublishIdempotent(t *testing.T) {
	e, _ := nuevoEntorno(t)
	token := e.login(t)
	draft := `{"name":"Test","tagline":"","logo":"","c":{"primary":"#c9a24b","accent":"#e8d9a8","bg":"#161616","surface":"#222222","text":"#f4f0e8"},"hf":"Inter","bf":"Inter","r":8,"btn":"solid","biz":{"whatsapp":"","instagram":"","address":"","currency":"ARS","slot":30,"notice":1,"open":"09:00","close":"20:00","tz":"America/Argentina/Buenos_Aires"}}`
	rec := e.req(t, http.MethodPut, "/api/admin/settings/draft", draft, token)
	if rec.Code != 200 {
		t.Fatalf("draft: %d %s", rec.Code, rec.Body.String())
	}
	r1 := e.req(t, http.MethodPost, "/api/admin/settings/publish", "", token)
	r2 := e.req(t, http.MethodPost, "/api/admin/settings/publish", "", token)
	if r1.Code != 200 || r2.Code != 200 {
		t.Fatalf("publish: %d %d", r1.Code, r2.Code)
	}
	// versiones: solo 1
	rec = e.req(t, http.MethodGet, "/api/admin/settings/versions", "", token)
	if strings.Count(rec.Body.String(), `"version"`) != 1 {
		t.Fatalf("versiones: %s", rec.Body.String())
	}
}
