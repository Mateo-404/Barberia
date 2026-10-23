package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewPage(t *testing.T) {
	p := newPage([]string{"a", "b", "c"}, 25, 2, 10)
	if p.Total != 25 || p.TotalPages != 3 || p.Page != 2 || p.Size != 10 {
		t.Errorf("page = %+v", p)
	}
	empty := newPage([]string{}, 0, 0, 10)
	b, _ := json.Marshal(empty)
	if empty.TotalPages != 0 || !strings.Contains(string(b), `"items":[]`) {
		t.Errorf("empty page = %s", b)
	}
}

func TestDecodeJSON(t *testing.T) {
	var dst struct {
		Name  string   `json:"name"`
		Price *float32 `json:"price"`
	}
	ok := func(body string) bool {
		rec := httptest.NewRecorder()
		return decodeJSON(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), &dst, 0)
	}
	if !ok(`{"name":"Cut","price":5000,"unknown":true}`) || dst.Name != "Cut" || *dst.Price != 5000 {
		t.Errorf("unknown fields must be ignored: %+v", dst)
	}
	if ok("") {
		t.Error("an empty body must be rejected")
	}
	if ok(`{"name":"A"}{"name":"B"}`) {
		t.Error("trailing JSON must be rejected")
	}
	if ok(`{"name":"` + strings.Repeat("a", 2<<20) + `"}`) {
		t.Error("a body over 1 MiB must be rejected")
	}
}

func TestValidatorEmail(t *testing.T) {
	for _, e := range []string{"a@b.com", "user.name+tag@example.com.ar", "x@y"} {
		var v validator
		v.email("email", &e)
		if len(v.errs) != 0 {
			t.Errorf("%q should be valid", e)
		}
	}
	for _, e := range []string{"plain", "a@", "@b.com", "a@b.c.", "a b@c.com", "a@@b.com"} {
		var v validator
		v.email("email", &e)
		if len(v.errs) != 1 {
			t.Errorf("%q should be invalid", e)
		}
	}
	var v validator
	v.email("email", nil)
	if len(v.errs) != 0 {
		t.Error("an absent email is valid")
	}
}

func TestValidatorOptionalRules(t *testing.T) {
	blank, short := "  ", "short"
	negative := float32(-1)
	var v validator
	v.required("a", nil)
	v.required("b", &blank)
	v.minLen("c", &short, minPassword)
	v.minLen("d", nil, minPassword)
	v.positive("e", &negative)
	v.positive("f", nil)
	v.maxLen("g", &short, 3)
	want := []fieldProblem{{"a", codeRequired}, {"b", codeRequired}, {"c", codeTooShort}, {"e", codeMustBePositive}, {"g", codeTooLong}}
	if len(v.errs) != len(want) {
		t.Fatalf("errs = %v, want %v", v.errs, want)
	}
	for i := range want {
		if v.errs[i] != want[i] {
			t.Errorf("errs[%d] = %v, want %v", i, v.errs[i], want[i])
		}
	}
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:4242"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.1")
	if got := (&Server{}).clientIP(req); got != "10.0.0.9" {
		t.Errorf("X-Forwarded-For must be ignored by default, got %q", got)
	}
	if got := (&Server{trustProxyHeaders: true}).clientIP(req); got != "1.2.3.4" {
		t.Errorf("with TRUST_PROXY_HEADERS the first forwarded IP is used, got %q", got)
	}
}

func TestIsPublic(t *testing.T) {
	cases := []struct {
		method, path string
		public       bool
	}{
		{http.MethodGet, "/healthz", true},
		{http.MethodGet, "/readyz", true},
		{http.MethodPost, "/api/appointments", true},
		{http.MethodGet, "/api/appointments/booked-slots", true},
		{http.MethodGet, "/api/services", true},
		{http.MethodGet, "/api/services/1", true},
		{http.MethodPost, "/api/auth/login", true},
		{http.MethodGet, "/api/public/theme.css", true},
		{http.MethodOptions, "/api/appointments", true},
		{http.MethodGet, "/api/appointments", false},
		{http.MethodPost, "/api/services", false},
		{http.MethodPatch, "/api/services/1", false},
		{http.MethodGet, "/api/admins", false},
		{http.MethodGet, "/api/stats/dashboard", false},
		{http.MethodDelete, "/api/appointments/1", false},
		{http.MethodGet, "/api/admin/settings", false},
	}
	for _, c := range cases {
		if got := isPublic(httptest.NewRequest(c.method, c.path, nil)); got != c.public {
			t.Errorf("isPublic(%s %s) = %v, want %v", c.method, c.path, got, c.public)
		}
	}
}
