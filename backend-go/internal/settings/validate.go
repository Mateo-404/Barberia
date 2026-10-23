package settings

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	reHex  = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	reIG   = regexp.MustCompile(`^@?[A-Za-z0-9._]{1,30}$`)
	reTime = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
)

// Detail is one validation finding. Code is a stable identifier that
// clients translate (invalid_json, invalid_length, too_long, invalid_format,
// invalid_color, invalid_font, invalid_value, low_contrast,
// insufficient_contrast).
type Detail struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

// ThemeResult holds the normalized theme (nil when invalid), blocking
// details and non-blocking warnings.
type ThemeResult struct {
	Theme    *Theme
	Details  []Detail
	Warnings []Detail
}

// BusinessResult is the outcome of validating business data.
type BusinessResult struct {
	Business *Business
	Details  []Detail
}

// strictDecode rejects unknown fields and reports which one it was.
func strictDecode(raw json.RawMessage, dst any) (Detail, bool) {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		field := "*"
		if _, rest, ok := strings.Cut(err.Error(), `unknown field "`); ok {
			if name, _, ok := strings.Cut(rest, `"`); ok {
				field = name
			}
		}
		return Detail{Field: field, Code: "invalid_json"}, false
	}
	return Detail{}, true
}

type findings struct{ details, warnings []Detail }

func (f *findings) fail(field, code string) { f.details = append(f.details, Detail{field, code}) }
func (f *findings) warn(field, code string) { f.warnings = append(f.warnings, Detail{field, code}) }

// ValidateTheme decodes raw strictly (unknown fields are rejected), checks
// every field and the contrast of the resulting palette, and returns the
// theme with colors lower-cased.
func ValidateTheme(raw json.RawMessage) *ThemeResult {
	res := &ThemeResult{Details: []Detail{}, Warnings: []Detail{}}
	var t Theme
	if d, ok := strictDecode(raw, &t); !ok {
		res.Details = append(res.Details, d)
		return res
	}
	var f findings

	if t.V != ThemeVersion {
		f.fail("v", "invalid_value")
	}
	if n := utf8.RuneCountInString(t.Name); n < 2 || n > 60 {
		f.fail("name", "invalid_length")
	}
	if utf8.RuneCountInString(t.Tagline) > 120 {
		f.fail("tagline", "invalid_length")
	}
	checkLogo(&f, t.Logo)

	if _, known := PresetByID(t.Preset); !known && t.Preset != CustomPreset {
		f.fail("preset", "invalid_value")
	}
	if t.Mode != ModeLight && t.Mode != ModeDark && t.Mode != ModeAuto {
		f.fail("mode", "invalid_value")
	}
	if !reHex.MatchString(t.Brand) {
		f.fail("brand", "invalid_color")
	}
	if _, ok := FontAllowlist[t.Font.Heading]; !ok {
		f.fail("font.heading", "invalid_font")
	}
	if _, ok := FontAllowlist[t.Font.Body]; !ok {
		f.fail("font.body", "invalid_font")
	}
	if t.Shape.Radius < 0 || t.Shape.Radius > 24 {
		f.fail("shape.radius", "invalid_value")
	}
	if t.Shape.Button != "solid" && t.Shape.Button != "outline" {
		f.fail("shape.button", "invalid_value")
	}
	checkOverrides(&f, "overrides.light", &t.Overrides.Light)
	checkOverrides(&f, "overrides.dark", &t.Overrides.Dark)

	t.Brand = strings.ToLower(t.Brand)
	if len(f.details) == 0 {
		checkContrast(&f, "overrides.light", &t, false)
		checkContrast(&f, "overrides.dark", &t, true)
	}
	res.Details = append(res.Details, f.details...)
	res.Warnings = append(res.Warnings, f.warnings...)
	if len(res.Details) == 0 {
		res.Theme = &t
	}
	return res
}

func checkLogo(f *findings, logo string) {
	switch {
	case logo == "":
	case len(logo) > 400000:
		f.fail("logo", "too_long")
	case strings.HasPrefix(logo, "https://"):
	case strings.HasPrefix(logo, "data:image/png;base64,"),
		strings.HasPrefix(logo, "data:image/jpeg;base64,"),
		strings.HasPrefix(logo, "data:image/webp;base64,"):
	default:
		f.fail("logo", "invalid_format")
	}
}

// checkOverrides validates and lower-cases every override color in place.
func checkOverrides(f *findings, path string, o *TokenOverrides) {
	for name, value := range map[string]*string{"bg": &o.Bg, "surface": &o.Surface, "text": &o.Text, "primary": &o.Primary, "accent": &o.Accent} {
		if *value == "" {
			continue
		}
		if !reHex.MatchString(*value) {
			f.fail(path+"."+name, "invalid_color")
			continue
		}
		*value = strings.ToLower(*value)
	}
}

// checkContrast reports poor contrast of the generated palette. Without
// overrides it never triggers (the generator guarantees the ratios), so the
// findings point at the override that broke readability.
func checkContrast(f *findings, path string, t *Theme, dark bool) {
	light, darkScheme := Generate(t.Brand, t.Overrides)
	s, over := light, t.Overrides.Light
	if dark {
		s, over = darkScheme, t.Overrides.Dark
	}
	if ratio, _ := Contrast(s.Text, s.Bg); ratio < minUIContrast {
		f.fail(path+".text", "insufficient_contrast")
	} else if ratio < minTextContrast {
		f.warn(path+".text", "low_contrast")
	}
	if ratio, _ := Contrast(s.Text, s.Surface); ratio < minUIContrast {
		f.warn(path+".text", "low_contrast")
	}
	if over.Primary != "" {
		if ratio, _ := Contrast(s.Primary, s.Bg); ratio < minUIContrast {
			f.warn(path+".primary", "low_contrast")
		}
	}
}

// ValidateBusiness checks the operating data of a business.
func ValidateBusiness(raw json.RawMessage) *BusinessResult {
	res := &BusinessResult{Details: []Detail{}}
	var b Business
	if d, ok := strictDecode(raw, &b); !ok {
		res.Details = append(res.Details, d)
		return res
	}
	var f findings

	if b.Whatsapp != "" {
		clean := strings.NewReplacer(" ", "", "(", "", ")", "", "-", "").Replace(b.Whatsapp)
		if len(clean) < 6 || len(clean) > 21 || (clean[0] != '+' && (clean[0] < '0' || clean[0] > '9')) {
			f.fail("whatsapp", "invalid_value")
		}
	}
	if b.Instagram != "" && !reIG.MatchString(b.Instagram) {
		f.fail("instagram", "invalid_value")
	}
	if utf8.RuneCountInString(b.Address) > 160 {
		f.fail("address", "invalid_length")
	}
	switch b.Currency {
	case "ARS", "USD", "EUR", "UYU", "CLP":
	default:
		f.fail("currency", "invalid_value")
	}
	if b.Slot < 10 || b.Slot > 240 {
		f.fail("slot", "invalid_value")
	}
	if b.Notice < 0 || b.Notice > 168 {
		f.fail("notice", "invalid_value")
	}
	if !reTime.MatchString(b.Open) || !reTime.MatchString(b.Close) || b.Open >= b.Close {
		f.fail("open", "invalid_value")
	}
	if b.Tz == "" {
		b.Tz = DefaultTimezone
	} else if _, err := time.LoadLocation(b.Tz); err != nil {
		f.fail("tz", "invalid_value")
	}

	res.Details = append(res.Details, f.details...)
	if len(res.Details) == 0 {
		res.Business = &b
	}
	return res
}

// DefaultBusiness is used until a tenant saves its own.
func DefaultBusiness() Business {
	return Business{Currency: "ARS", Slot: 30, Notice: 1, Open: "13:00", Close: "20:00", Tz: DefaultTimezone}
}
