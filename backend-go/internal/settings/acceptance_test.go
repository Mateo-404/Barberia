package settings

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

// mergePatch applies a JSON Merge Patch (RFC 7396) to base.
func mergePatch(base, patch json.RawMessage) json.RawMessage {
	var b map[string]any
	var p map[string]any
	_ = json.Unmarshal(base, &b)
	_ = json.Unmarshal(patch, &p)
	mergeMaps(b, p)
	out, _ := json.Marshal(b)
	return out
}

func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if v == nil {
			delete(dst, k)
			continue
		}
		if vm, ok := v.(map[string]any); ok {
			if dm, ok := dst[k].(map[string]any); ok {
				mergeMaps(dm, vm)
				continue
			}
		}
		dst[k] = v
	}
}

// S-01: base + valid variants validate.
func TestAcceptanceS01Valid(t *testing.T) {
	b := readTestdata(t, "settings_valid.json")
	var v struct {
		Base     json.RawMessage  `json:"base"`
		Variants []map[string]any `json:"variants"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if res := Validate(v.Base); len(res.Details) != 0 {
		t.Fatalf("invalid base: %+v", res.Details)
	}
	// each variant is a patch over base
	raw := readTestdata(t, "settings_valid.json")
	var rawv map[string]json.RawMessage
	_ = json.Unmarshal(raw, &rawv)
	// re-read variants as raw
	var doc struct {
		Base     json.RawMessage `json:"base"`
		Variants []struct {
			Name  string          `json:"name"`
			Patch json.RawMessage `json:"patch"`
		} `json:"variants"`
	}
	_ = json.Unmarshal(b, &doc)
	for _, vr := range doc.Variants {
		merged := mergePatch(doc.Base, vr.Patch)
		res := Validate(merged)
		if len(res.Details) != 0 {
			t.Fatalf("variant %q invalid: %+v", vr.Name, res.Details)
		}
	}
}

// S-02: invalid cases fail and details mention expect_field.
func TestAcceptanceS02Invalid(t *testing.T) {
	b := readTestdata(t, "settings_invalid.json")
	var doc struct {
		Cases []struct {
			Name        string          `json:"name"`
			Patch       json.RawMessage `json:"patch"`
			ExpectField string          `json:"expect_field"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	// base taken from settings_valid.json
	vb := readTestdata(t, "settings_valid.json")
	var vdoc struct {
		Base json.RawMessage `json:"base"`
	}
	_ = json.Unmarshal(vb, &vdoc)
	docBase := vdoc.Base
	for _, c := range doc.Cases {
		merged := mergePatch(docBase, c.Patch)
		res := Validate(merged)
		if len(res.Details) == 0 {
			t.Fatalf("case %q should fail", c.Name)
		}
		found := false
		for _, d := range res.Details {
			if d.Field == c.ExpectField {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("caso %q: expect_field %q no encontrado en %+v", c.Name, c.ExpectField, res.Details)
		}
	}
}

// S-03: a logo over 400000 characters is rejected.
func TestAcceptanceS03LogoTooLong(t *testing.T) {
	base := readTestdata(t, "settings_valid.json")
	var doc struct {
		Base json.RawMessage `json:"base"`
	}
	_ = json.Unmarshal(base, &doc)
	var s map[string]any
	_ = json.Unmarshal(doc.Base, &s)
	s["logo"] = "https://" + strings.Repeat("a", 400000)
	raw, _ := json.Marshal(s)
	res := Validate(raw)
	if len(res.Details) == 0 {
		t.Fatal("logo > 400000 should fail")
	}
}

// S-04: contrast_vectors ratios match.
func TestAcceptanceS04ContrastVectors(t *testing.T) {
	b := readTestdata(t, "contrast_vectors.json")
	var v struct {
		Tolerance float64 `json:"tolerance"`
		Contrast  []struct {
			FG    string  `json:"fg"`
			BG    string  `json:"bg"`
			Ratio float64 `json:"ratio"`
		} `json:"contrast"`
	}
	_ = json.Unmarshal(b, &v)
	for _, c := range v.Contrast {
		r, err := Contrast(c.FG, c.BG)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(r-c.Ratio) > v.Tolerance+1e-9 {
			t.Fatalf("mismatch %s/%s got %f want %f", c.FG, c.BG, r, c.Ratio)
		}
	}
}

// S-06: RenderCSS is deterministic.
func TestAcceptanceS06Deterministic(t *testing.T) {
	base := readTestdata(t, "settings_valid.json")
	var doc struct {
		Base json.RawMessage `json:"base"`
	}
	_ = json.Unmarshal(base, &doc)
	r1 := Validate(doc.Base)
	css1 := RenderCSS(r1.Settings)
	h1 := Hash(css1)
	css2 := RenderCSS(r1.Settings)
	h2 := Hash(css2)
	if css1 != css2 || h1 != h2 {
		t.Fatal("RenderCSS no es determinista")
	}
	// only validated values: check there are no dangerous characters
	if strings.Contains(css1, "<") || strings.Contains(css1, ">") {
		t.Fatal("RenderCSS contiene caracteres peligrosos")
	}
}

// S-07: a name containing <script> is accepted as plain text.
func TestAcceptanceS07XSSNameAccepted(t *testing.T) {
	base := readTestdata(t, "settings_valid.json")
	var doc struct {
		Base json.RawMessage `json:"base"`
	}
	_ = json.Unmarshal(base, &doc)
	patch := json.RawMessage(`{"name":"<script>alert(1)</script>"}`)
	merged := mergePatch(doc.Base, patch)
	res := Validate(merged)
	if len(res.Details) != 0 {
		t.Fatalf("a name with script should validate: %+v", res.Details)
	}
	if res.Settings.Name != "<script>alert(1)</script>" {
		t.Fatalf("name not preserved: %q", res.Settings.Name)
	}
}

// S-11: warnings do not block; < 3:1 blocks.
func TestAcceptanceS11ContrastWarnings(t *testing.T) {
	base := readTestdata(t, "settings_valid.json")
	var doc struct {
		Base json.RawMessage `json:"base"`
	}
	_ = json.Unmarshal(base, &doc)
	// ratio 3–4.5: warnings only
	patch := json.RawMessage(`{"c":{"text":"#777777"}}`)
	merged := mergePatch(doc.Base, patch)
	res := Validate(merged)
	if len(res.Details) != 0 {
		t.Fatalf("contrast 3–4.5 must not block: %+v", res.Details)
	}
	// ratio < 3 blocks
	patch2 := json.RawMessage(`{"c":{"text":"#ffff00","bg":"#ffffff"}}`)
	merged2 := mergePatch(doc.Base, patch2)
	res2 := Validate(merged2)
	if len(res2.Details) == 0 {
		t.Fatal("contrast < 3 must block")
	}
}
