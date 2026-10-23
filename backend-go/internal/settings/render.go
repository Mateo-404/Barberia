package settings

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"strings"
)

// RenderCSS turns a theme into the stylesheet the frontend loads. Both color
// schemes are always emitted; the mode only decides which one is the default:
//
//	light / dark  one scheme at :root, the other under [data-scheme]
//	auto          light at :root, dark when the visitor's system prefers it
//
// Frontend code switches schemes with the data-scheme attribute and reads
// --theme-mode to know whether the visitor may toggle.
func RenderCSS(t *Theme) string {
	light, dark := Generate(t.Brand, t.Overrides)

	var b strings.Builder
	b.WriteString(":root{")
	fmt.Fprintf(&b, "--theme-mode:%s;--radius:%dpx;--button-style:%s;", t.Mode, t.Shape.Radius, t.Shape.Button)
	fmt.Fprintf(&b, "--font-heading:%s;--font-body:%s;", fontStack(t.Font.Heading), fontStack(t.Font.Body))
	first, second := light, dark
	if t.Mode == ModeDark {
		first, second = dark, light
	}
	b.WriteString(tokens(first))
	b.WriteString("}")

	secondSelector := `:root[data-scheme="dark"]`
	if t.Mode == ModeDark {
		secondSelector = `:root[data-scheme="light"]`
	}
	fmt.Fprintf(&b, "%s{%s}", secondSelector, tokens(second))
	if t.Mode == ModeAuto {
		fmt.Fprintf(&b, "@media (prefers-color-scheme:dark){:root:not([data-scheme]){%s}}", tokens(dark))
	}
	return b.String()
}

func tokens(s Scheme) string {
	return fmt.Sprintf(
		"--color-bg:%s;--color-surface:%s;--color-text:%s;--color-muted:%s;--color-primary:%s;--color-on-primary:%s;--color-accent:%s;--color-border:%s;",
		s.Bg, s.Surface, s.Text, s.Muted, s.Primary, s.OnPrimary, s.Accent, s.Border,
	)
}

// Hash is a short content hash used for cache busting.
func Hash(css string) string {
	h := sha256.Sum256([]byte(css))
	return hex.EncodeToString(h[:])[:12]
}

// Escape HTML-escapes s.
func Escape(s string) string {
	return template.HTMLEscapeString(s)
}
