package settings

import (
	"encoding/json"
	"testing"
)

func jsonOf(v any) (json.RawMessage, error) { return json.Marshal(v) }

func TestUpgradeV1KeepsTheLookAndSplitsOutBusiness(t *testing.T) {
	var fx struct {
		Document json.RawMessage `json:"document"`
	}
	if err := json.Unmarshal(readTestdata(t, "settings_v1.json"), &fx); err != nil {
		t.Fatal(err)
	}
	if !IsLegacy(fx.Document) {
		t.Fatal("a document without v is legacy")
	}
	theme, business, err := UpgradeV1(fx.Document)
	if err != nil {
		t.Fatal(err)
	}
	if theme.V != ThemeVersion || theme.Preset != CustomPreset || theme.Mode != ModeDark || theme.Brand != "#c9a24b" {
		t.Fatalf("theme = %+v", theme)
	}
	if theme.Overrides.Dark.Bg != "#161616" || theme.Overrides.Dark.Text != "#f4f0e8" || theme.Overrides.Light != (TokenOverrides{}) {
		t.Fatalf("v1 palette must become dark overrides: %+v", theme.Overrides)
	}
	if business.Open != "09:00" || business.Tz != "America/Argentina/Buenos_Aires" || business.Currency != "ARS" {
		t.Fatalf("business = %+v", business)
	}
	raw, _ := jsonOf(theme)
	if res := ValidateTheme(raw); res.Theme == nil {
		t.Fatalf("the upgraded theme must be valid: %+v", res.Details)
	}
}

func TestUpgradeV1PicksLightModeForLightBackgrounds(t *testing.T) {
	doc := json.RawMessage(`{"name":"X","c":{"primary":"#336699","accent":"#aabbcc","bg":"#fafafa","surface":"#ffffff","text":"#111111"},"hf":"Papyrus","bf":"Inter","r":99,"btn":"solid","biz":{}}`)
	theme, _, err := UpgradeV1(doc)
	if err != nil {
		t.Fatal(err)
	}
	if theme.Mode != ModeLight || theme.Overrides.Light.Bg != "#fafafa" {
		t.Fatalf("theme = %+v", theme)
	}
	if theme.Font.Heading != "Inter" || theme.Shape.Radius != 24 {
		t.Fatalf("unknown font must fall back and radius clamp: %+v", theme)
	}
}

func TestNormalizeThemeAcceptsBothVersions(t *testing.T) {
	v2, _ := jsonOf(DefaultTheme())
	got, err := NormalizeTheme(v2)
	if err != nil || got.Brand != DefaultTheme().Brand {
		t.Fatalf("v2: %+v %v", got, err)
	}
	v1 := json.RawMessage(`{"name":"Old","c":{"primary":"#112233","accent":"#445566","bg":"#000000","surface":"#111111","text":"#ffffff"},"hf":"Inter","bf":"Inter","r":8,"btn":"solid","biz":{}}`)
	got, err = NormalizeTheme(v1)
	if err != nil || got.V != ThemeVersion || got.Name != "Old" {
		t.Fatalf("v1: %+v %v", got, err)
	}
}
