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
// clients translate (invalid_json, invalid_length, invalid_format,
// invalid_color, invalid_font, invalid_value, low_contrast,
// insufficient_contrast).
type Detail struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

// ValidationResult holds the decoded settings (nil when invalid), blocking
// details and non-blocking warnings.
type ValidationResult struct {
	Settings *Settings
	Details  []Detail
	Warnings []Detail
}

// Validate decodes raw strictly (unknown fields are rejected) and checks
// every field and the color contrast.
func Validate(raw json.RawMessage) *ValidationResult {
	res := &ValidationResult{Details: []Detail{}, Warnings: []Detail{}}
	var s Settings
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		field := "*"
		msg := err.Error()
		// encoding/json: unknown field "evil"
		if i := strings.Index(msg, "unknown field"); i >= 0 {
			rest := msg[i:]
			if q := strings.Index(rest, "\""); q >= 0 {
				rest = rest[q+1:]
				if q2 := strings.Index(rest, "\""); q2 >= 0 {
					field = rest[:q2]
				}
			}
		}
		res.Details = append(res.Details, Detail{Field: field, Code: "invalid_json"})
		return res
	}

	// name
	if utf8.RuneCountInString(s.Name) < 2 || utf8.RuneCountInString(s.Name) > 60 {
		res.Details = append(res.Details, Detail{Field: "name", Code: "invalid_length"})
	}
	// tagline
	if utf8.RuneCountInString(s.Tagline) > 120 {
		res.Details = append(res.Details, Detail{Field: "tagline", Code: "invalid_length"})
	}
	// logo
	if s.Logo != "" {
		if len(s.Logo) > 400000 {
			res.Details = append(res.Details, Detail{Field: "logo", Code: "demasiado largo"})
		} else if strings.HasPrefix(s.Logo, "https://") {
			// ok
		} else if strings.HasPrefix(s.Logo, "data:image/") {
			if !strings.Contains(s.Logo, ";base64,") {
				res.Details = append(res.Details, Detail{Field: "logo", Code: "invalid_format"})
			} else {
				mime := s.Logo
				if !strings.Contains(mime, "data:image/png;base64,") && !strings.Contains(mime, "data:image/jpeg;base64,") && !strings.Contains(mime, "data:image/webp;base64,") {
					res.Details = append(res.Details, Detail{Field: "logo", Code: "invalid_format"})
				}
			}
		} else {
			res.Details = append(res.Details, Detail{Field: "logo", Code: "invalid_format"})
		}
	}

	// colors
	checkColor := func(field, val string) {
		if !reHex.MatchString(val) {
			res.Details = append(res.Details, Detail{Field: field, Code: "invalid_color"})
			return
		}
		// lower-casing happens later, once valid
	}
	checkColor("c.primary", s.C.Primary)
	checkColor("c.accent", s.C.Accent)
	checkColor("c.bg", s.C.Bg)
	checkColor("c.surface", s.C.Surface)
	checkColor("c.text", s.C.Text)

	if len(res.Details) == 0 {
		s.C.Primary = strings.ToLower(s.C.Primary)
		s.C.Accent = strings.ToLower(s.C.Accent)
		s.C.Bg = strings.ToLower(s.C.Bg)
		s.C.Surface = strings.ToLower(s.C.Surface)
		s.C.Text = strings.ToLower(s.C.Text)
	}

	// fonts
	if _, ok := FontAllowlist[s.HF]; !ok || s.HF == "" {
		res.Details = append(res.Details, Detail{Field: "hf", Code: "invalid_font"})
	}
	if _, ok := FontAllowlist[s.BF]; !ok || s.BF == "" {
		res.Details = append(res.Details, Detail{Field: "bf", Code: "invalid_font"})
	}

	// radius
	if s.R < 0 || s.R > 24 {
		res.Details = append(res.Details, Detail{Field: "r", Code: "invalid_value"})
	}
	// btn
	if s.Btn != "solid" && s.Btn != "outline" {
		res.Details = append(res.Details, Detail{Field: "btn", Code: "invalid_value"})
	}

	// biz
	if s.Biz.Whatsapp != "" {
		// allow +, digits, spaces, parentheses and dashes
		clean := strings.ReplaceAll(s.Biz.Whatsapp, " ", "")
		clean = strings.ReplaceAll(clean, "(", "")
		clean = strings.ReplaceAll(clean, ")", "")
		clean = strings.ReplaceAll(clean, "-", "")
		if len(clean) < 6 || len(clean) > 21 {
			res.Details = append(res.Details, Detail{Field: "biz.whatsapp", Code: "invalid_value"})
		} else if clean[0] != '+' && (clean[0] < '0' || clean[0] > '9') {
			res.Details = append(res.Details, Detail{Field: "biz.whatsapp", Code: "invalid_value"})
		}
	}
	if s.Biz.Instagram != "" {
		if !reIG.MatchString(s.Biz.Instagram) {
			res.Details = append(res.Details, Detail{Field: "biz.instagram", Code: "invalid_value"})
		}
	}
	if s.Biz.Currency != "ARS" && s.Biz.Currency != "USD" && s.Biz.Currency != "EUR" && s.Biz.Currency != "UYU" && s.Biz.Currency != "CLP" {
		res.Details = append(res.Details, Detail{Field: "biz.currency", Code: "invalid_value"})
	}
	if s.Biz.Slot < 10 || s.Biz.Slot > 240 {
		res.Details = append(res.Details, Detail{Field: "biz.slot", Code: "invalid_value"})
	}
	if s.Biz.Notice < 0 || s.Biz.Notice > 168 {
		res.Details = append(res.Details, Detail{Field: "biz.notice", Code: "invalid_value"})
	}
	if !reTime.MatchString(s.Biz.Open) || !reTime.MatchString(s.Biz.Close) {
		res.Details = append(res.Details, Detail{Field: "biz.open", Code: "invalid_value"})
	} else {
		if s.Biz.Open >= s.Biz.Close {
			res.Details = append(res.Details, Detail{Field: "biz.open", Code: "invalid_value"})
		}
	}
	if s.Biz.Tz != nil && *s.Biz.Tz != "" {
		if _, err := time.LoadLocation(*s.Biz.Tz); err != nil {
			res.Details = append(res.Details, Detail{Field: "biz.tz", Code: "invalid_value"})
		}
	} else {
		def := "America/Argentina/Buenos_Aires"
		s.Biz.Tz = &def
	}

	// contrast
	if len(res.Details) == 0 {
		if ratio, err := Contrast(s.C.Text, s.C.Bg); err == nil {
			if ratio < 3.0 {
				res.Details = append(res.Details, Detail{Field: "c.text", Code: "insufficient_contrast"})
			} else if ratio < 4.5 {
				res.Warnings = append(res.Warnings, Detail{Field: "c.text", Code: "low_contrast"})
			}
		}
		if ratio, err := Contrast(s.C.Text, s.C.Surface); err == nil {
			if ratio < 3.0 {
				res.Warnings = append(res.Warnings, Detail{Field: "c.text", Code: "low_contrast"})
			}
		}
		if ratio, err := Contrast(s.C.Primary, s.C.Bg); err == nil {
			if ratio < 3.0 {
				res.Warnings = append(res.Warnings, Detail{Field: "c.primary", Code: "low_contrast"})
			}
		}
	}

	if len(res.Details) > 0 {
		return res
	}
	s.Raw = raw
	res.Settings = &s
	return res
}
