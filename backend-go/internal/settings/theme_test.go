package settings

import (
	"encoding/json"
	"strings"
	"testing"
)

// mergePatch applies a JSON Merge Patch (RFC 7396) to base.
func mergePatch(t *testing.T, base, patch json.RawMessage) json.RawMessage {
	t.Helper()
	var b, p map[string]any
	if err := json.Unmarshal(base, &b); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(patch, &p); err != nil {
		t.Fatal(err)
	}
	mergeMaps(b, p)
	out, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if v == nil {
			delete(dst, k)
			continue
		}
		sub, isMap := v.(map[string]any)
		cur, curIsMap := dst[k].(map[string]any)
		if isMap && curIsMap {
			mergeMaps(cur, sub)
		} else {
			dst[k] = v
		}
	}
}

type fixture struct {
	Base     json.RawMessage `json:"base"`
	Variants []struct {
		Name  string          `json:"name"`
		Patch json.RawMessage `json:"patch"`
	} `json:"variants"`
	Cases []struct {
		Name        string          `json:"name"`
		Patch       json.RawMessage `json:"patch"`
		ExpectField string          `json:"expect_field"`
	} `json:"cases"`
}

func loadFixture(t *testing.T, name string) fixture {
	t.Helper()
	var f fixture
	if err := json.Unmarshal(readTestdata(t, name), &f); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestValidThemes(t *testing.T) {
	f := loadFixture(t, "theme_valid.json")
	if res := ValidateTheme(f.Base); res.Theme == nil {
		t.Fatalf("base is invalid: %+v", res.Details)
	}
	for _, v := range f.Variants {
		res := ValidateTheme(mergePatch(t, f.Base, v.Patch))
		if res.Theme == nil {
			t.Errorf("variant %q is invalid: %+v", v.Name, res.Details)
		}
	}
}

func TestInvalidThemesFailOnTheExpectedField(t *testing.T) {
	base := loadFixture(t, "theme_valid.json").Base
	for _, c := range loadFixture(t, "theme_invalid.json").Cases {
		res := ValidateTheme(mergePatch(t, base, c.Patch))
		if res.Theme != nil {
			t.Errorf("case %q should fail", c.Name)
			continue
		}
		found := false
		for _, d := range res.Details {
			found = found || d.Field == c.ExpectField
		}
		if !found {
			t.Errorf("case %q: details %+v do not mention %q", c.Name, res.Details, c.ExpectField)
		}
	}
}

func TestBrandIsNormalizedToLowerCase(t *testing.T) {
	f := loadFixture(t, "theme_valid.json")
	res := ValidateTheme(mergePatch(t, f.Base, json.RawMessage(`{"brand":"#C9A24B"}`)))
	if res.Theme == nil || res.Theme.Brand != "#c9a24b" {
		t.Fatalf("brand = %+v", res.Theme)
	}
}

func TestScriptInNameIsPreservedAndEscapedOnDemand(t *testing.T) {
	f := loadFixture(t, "theme_valid.json")
	res := ValidateTheme(mergePatch(t, f.Base, json.RawMessage(`{"name":"<script>alert(1)</script>"}`)))
	if res.Theme == nil || res.Theme.Name != "<script>alert(1)</script>" {
		t.Fatalf("name = %+v", res.Theme)
	}
	if got := Escape(res.Theme.Name); strings.Contains(got, "<script>") {
		t.Fatalf("Escape left the tag: %s", got)
	}
}

func TestLowContrastOverrideWarnsAndUnreadableOneBlocks(t *testing.T) {
	f := loadFixture(t, "theme_valid.json")
	// ~3.7:1 text on background: allowed, with a warning.
	warn := ValidateTheme(mergePatch(t, f.Base, json.RawMessage(`{"overrides":{"dark":{"bg":"#161616","text":"#7c7c7c"}}}`)))
	if warn.Theme == nil || len(warn.Warnings) == 0 || warn.Warnings[0].Code != "low_contrast" {
		t.Fatalf("expected a low_contrast warning, got theme=%v warnings=%+v", warn.Theme != nil, warn.Warnings)
	}
	// ~1.1:1: blocked.
	block := ValidateTheme(mergePatch(t, f.Base, json.RawMessage(`{"overrides":{"dark":{"bg":"#161616","text":"#1c1c1c"}}}`)))
	if block.Theme != nil || block.Details[0].Code != "insufficient_contrast" {
		t.Fatalf("expected insufficient_contrast, got %+v", block.Details)
	}
}

func TestValidBusinesses(t *testing.T) {
	f := loadFixture(t, "business_valid.json")
	if res := ValidateBusiness(f.Base); res.Business == nil {
		t.Fatalf("base is invalid: %+v", res.Details)
	}
	for _, v := range f.Variants {
		res := ValidateBusiness(mergePatch(t, f.Base, v.Patch))
		if res.Business == nil {
			t.Errorf("variant %q is invalid: %+v", v.Name, res.Details)
		} else if res.Business.Tz == "" {
			t.Errorf("variant %q: timezone default not applied", v.Name)
		}
	}
}

func TestInvalidBusinesses(t *testing.T) {
	base := loadFixture(t, "business_valid.json").Base
	for _, c := range loadFixture(t, "business_invalid.json").Cases {
		res := ValidateBusiness(mergePatch(t, base, c.Patch))
		found := false
		for _, d := range res.Details {
			found = found || d.Field == c.ExpectField
		}
		if res.Business != nil || !found {
			t.Errorf("case %q: business=%v details=%+v, want a failure on %q", c.Name, res.Business != nil, res.Details, c.ExpectField)
		}
	}
}
