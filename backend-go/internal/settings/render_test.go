package settings

import (
	"strings"
	"testing"
)

func themeWithMode(mode string) *Theme {
	t := DefaultTheme()
	t.Mode = mode
	return &t
}

func TestRenderCSSByMode(t *testing.T) {
	cases := []struct {
		mode        string
		contains    []string
		notContains []string
	}{
		{ModeLight, []string{"--theme-mode:light", `:root[data-scheme="dark"]`}, []string{"prefers-color-scheme", `data-scheme="light"`}},
		{ModeDark, []string{"--theme-mode:dark", `:root[data-scheme="light"]`}, []string{"prefers-color-scheme", `data-scheme="dark"`}},
		{ModeAuto, []string{"--theme-mode:auto", `:root[data-scheme="dark"]`, "@media (prefers-color-scheme:dark){:root:not([data-scheme])"}, nil},
	}
	for _, c := range cases {
		css := RenderCSS(themeWithMode(c.mode))
		for _, s := range c.contains {
			if !strings.Contains(css, s) {
				t.Errorf("%s: missing %q in %s", c.mode, s, css)
			}
		}
		for _, s := range c.notContains {
			if strings.Contains(css, s) {
				t.Errorf("%s: unexpected %q in %s", c.mode, s, css)
			}
		}
	}
}

func TestRenderCSSIsDeterministicAndSafe(t *testing.T) {
	theme := DefaultTheme()
	theme.Name = `"};</style><script>alert(1)</script>` // never reaches the CSS
	a, b := RenderCSS(&theme), RenderCSS(&theme)
	if a != b || Hash(a) != Hash(b) {
		t.Fatal("rendering must be deterministic")
	}
	if strings.ContainsAny(a, "<>") || strings.Contains(a, "script") {
		t.Fatalf("css contains markup: %s", a)
	}
	for _, token := range []string{"--color-bg", "--color-surface", "--color-text", "--color-muted", "--color-primary", "--color-on-primary", "--color-accent", "--color-border", "--font-heading", "--font-body", "--radius:8px"} {
		if !strings.Contains(a, token) {
			t.Errorf("missing %s", token)
		}
	}
}

func TestChangingTheBrandChangesTheHash(t *testing.T) {
	a, b := DefaultTheme(), DefaultTheme()
	b.Brand = "#2f5d8a"
	if Hash(RenderCSS(&a)) == Hash(RenderCSS(&b)) {
		t.Fatal("different themes must hash differently")
	}
}

func TestEveryPresetValidates(t *testing.T) {
	for _, p := range Presets() {
		theme := p.Theme
		theme.Name = "Test"
		raw, err := jsonOf(theme)
		if err != nil {
			t.Fatal(err)
		}
		if res := ValidateTheme(raw); res.Theme == nil {
			t.Errorf("preset %s does not validate: %+v", p.ID, res.Details)
		}
	}
}
