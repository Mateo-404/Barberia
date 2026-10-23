//go:build integration

package integration

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
)

const validDraft = `{"name":"Test","tagline":"","logo":"","c":{"primary":"#c9a24b","accent":"#e8d9a8","bg":"#161616","surface":"#222222","text":"#f4f0e8"},"hf":"Inter","bf":"Inter","r":8,"btn":"solid","biz":{"whatsapp":"","instagram":"","address":"","currency":"ARS","slot":30,"notice":1,"open":"09:00","close":"20:00","tz":"America/Argentina/Buenos_Aires"}}`

// SEC-03: no /api/admin/* route works without authentication.
func TestAdminSettingsRequireAuth(t *testing.T) {
	e := newEnv(t)
	for _, p := range []string{"/api/admin/settings", "/api/admin/settings/versions"} {
		e.call(t, http.MethodGet, p, "", "").expect(http.StatusUnauthorized, p)
	}
}

// S-13: staff gets 403 on /api/admin/settings*.
func TestAdminSettingsForbiddenForStaff(t *testing.T) {
	e := newEnv(t)
	m, _ := auth.NewJWTManager(jwtSecret, 3600e9)
	staff, err := m.Issue(999, "staff@example.com", e.tenantID, auth.RoleStaff)
	if err != nil {
		t.Fatal(err)
	}
	e.call(t, http.MethodGet, "/api/admin/settings", "", staff).expect(http.StatusForbidden, "staff")
}

// S-10: the current theme hash is cacheable forever; otherwise revalidate.
func TestThemeCacheHeaders(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	e.call(t, http.MethodPut, "/api/admin/settings/draft", validDraft, token).expect(200, "draft")
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
}

// S-08: publishing the same draft twice creates a single version.
func TestPublishIsIdempotent(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	e.call(t, http.MethodPut, "/api/admin/settings/draft", validDraft, token).expect(200, "draft")
	e.call(t, http.MethodPost, "/api/admin/settings/publish", "", token).expect(200, "publish 1")
	e.call(t, http.MethodPost, "/api/admin/settings/publish", "", token).expect(200, "publish 2")
	var versions []struct {
		Version int32 `json:"version"`
	}
	e.call(t, http.MethodGet, "/api/admin/settings/versions", "", token).expect(200, "versions").json(&versions)
	if len(versions) != 1 {
		t.Fatalf("versions = %+v", versions)
	}
}

func TestInvalidDraftListsFields(t *testing.T) {
	e := newEnv(t)
	res := e.call(t, http.MethodPut, "/api/admin/settings/draft", strings.Replace(validDraft, `"hf":"Inter"`, `"hf":"Comic Sans"`, 1), e.ownerToken(t)).
		expect(http.StatusUnprocessableEntity, "invalid draft")
	var p struct {
		Code   string `json:"code"`
		Errors []struct {
			Field string `json:"field"`
			Code  string `json:"code"`
		} `json:"errors"`
	}
	res.json(&p)
	if p.Code != "settings.invalid" || len(p.Errors) != 1 || p.Errors[0].Field != "hf" || p.Errors[0].Code != "invalid_font" {
		t.Fatalf("problem = %+v", p)
	}
}
