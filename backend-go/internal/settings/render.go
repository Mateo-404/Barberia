package settings

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
)

func RenderCSS(s *Settings) string {
	onp := OnPrimary(s.C.Primary)
	var b strings.Builder
	b.WriteString(":root{")
	fmt.Fprintf(&b, "--color-primary:%s;", strings.ToLower(s.C.Primary))
	fmt.Fprintf(&b, "--color-accent:%s;", strings.ToLower(s.C.Accent))
	fmt.Fprintf(&b, "--color-bg:%s;", strings.ToLower(s.C.Bg))
	fmt.Fprintf(&b, "--color-surface:%s;", strings.ToLower(s.C.Surface))
	fmt.Fprintf(&b, "--color-text:%s;", strings.ToLower(s.C.Text))
	fmt.Fprintf(&b, "--color-on-primary:%s;", onp)
	fmt.Fprintf(&b, "--radius:%dpx;", s.R)
	gH := fontGeneric(s.HF)
	gB := fontGeneric(s.BF)
	fmt.Fprintf(&b, "--font-heading:\"%s\",%s;--font-body:\"%s\",%s}", s.HF, gH, s.BF, gB)
	return b.String()
}

func Hash(css string) string {
	h := sha256.Sum256([]byte(css))
	return hex.EncodeToString(h[:])[:12]
}

func Escape(s string) string {
	return template.HTMLEscapeString(s)
}

// CSSFromRaw renderiza el CSS de un settings ya validado (o defaults).
func CSSFromRaw(raw json.RawMessage) (string, error) {
	res := Validate(raw)
	if res.Settings == nil {
		return "", fmt.Errorf("settings inválidos")
	}
	return RenderCSS(res.Settings), nil
}
