package settings

import (
	"encoding/json"
	"strings"
)

// legacyV1 is the first settings document: one flat palette of five colors
// plus the business data in the same document.
type legacyV1 struct {
	Name    string `json:"name"`
	Tagline string `json:"tagline"`
	Logo    string `json:"logo"`
	C       struct {
		Primary string `json:"primary"`
		Accent  string `json:"accent"`
		Bg      string `json:"bg"`
		Surface string `json:"surface"`
		Text    string `json:"text"`
	} `json:"c"`
	HF  string `json:"hf"`
	BF  string `json:"bf"`
	R   int    `json:"r"`
	Btn string `json:"btn"`
	Biz struct {
		Whatsapp  string  `json:"whatsapp"`
		Instagram string  `json:"instagram"`
		Address   string  `json:"address"`
		Currency  string  `json:"currency"`
		Slot      int     `json:"slot"`
		Notice    int     `json:"notice"`
		Open      string  `json:"open"`
		Close     string  `json:"close"`
		Tz        *string `json:"tz"`
	} `json:"biz"`
}

// IsLegacy reports whether raw is a v1 document (it has no "v" field).
func IsLegacy(raw json.RawMessage) bool {
	var probe map[string]json.RawMessage
	if json.Unmarshal(raw, &probe) != nil {
		return false
	}
	_, hasVersion := probe["v"]
	return !hasVersion
}

// UpgradeV1 converts a v1 document into a theme and the business data it
// carried. The v1 palette becomes overrides of the scheme matching its
// background, so the site keeps looking the same until the owner edits it;
// versions already stored remain valid for rollback.
func UpgradeV1(raw json.RawMessage) (Theme, Business, error) {
	var old legacyV1
	if err := json.Unmarshal(raw, &old); err != nil {
		return Theme{}, Business{}, err
	}

	mode := ModeDark
	if bg, err := hexToOklab(old.C.Bg); err == nil && bg.L > 0.6 {
		mode = ModeLight
	}
	font := func(name string) string {
		if _, ok := FontAllowlist[name]; ok {
			return name
		}
		return "Inter"
	}
	t := Theme{
		V: ThemeVersion, Name: old.Name, Tagline: old.Tagline, Logo: old.Logo,
		Preset: CustomPreset, Mode: mode, Brand: strings.ToLower(old.C.Primary),
		Font:  Font{Heading: font(old.HF), Body: font(old.BF)},
		Shape: Shape{Radius: min(max(old.R, 0), 24), Button: old.Btn},
	}
	kept := TokenOverrides{
		Bg: strings.ToLower(old.C.Bg), Surface: strings.ToLower(old.C.Surface),
		Text: strings.ToLower(old.C.Text), Accent: strings.ToLower(old.C.Accent),
	}
	if mode == ModeLight {
		t.Overrides.Light = kept
	} else {
		t.Overrides.Dark = kept
	}

	b := Business{
		Whatsapp: old.Biz.Whatsapp, Instagram: old.Biz.Instagram, Address: old.Biz.Address,
		Currency: old.Biz.Currency, Slot: old.Biz.Slot, Notice: old.Biz.Notice,
		Open: old.Biz.Open, Close: old.Biz.Close, Tz: DefaultTimezone,
	}
	if old.Biz.Tz != nil && *old.Biz.Tz != "" {
		b.Tz = *old.Biz.Tz
	}
	return t, b, nil
}

// NormalizeTheme returns the v2 theme of a stored document, upgrading v1 on
// the fly. Stored documents were validated when saved, so this does not
// re-validate; callers that accept client input use ValidateTheme.
func NormalizeTheme(raw json.RawMessage) (Theme, error) {
	if IsLegacy(raw) {
		t, _, err := UpgradeV1(raw)
		return t, err
	}
	var t Theme
	err := json.Unmarshal(raw, &t)
	return t, err
}
