package settings

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

// mergePatch aplica un JSON Merge Patch (RFC 7396) sobre base.
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

// S-01: base + variantes válidas validan OK.
func TestAcceptanceS01Valid(t *testing.T) {
	b := leerTestdata(t, "settings_valid.json")
	var v struct {
		Base     json.RawMessage  `json:"base"`
		Variants []map[string]any `json:"variants"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if res := Validate(v.Base); len(res.Details) != 0 {
		t.Fatalf("base inválido: %+v", res.Details)
	}
	// variantes: cada una es un patch sobre base
	raw := leerTestdata(t, "settings_valid.json")
	var rawv map[string]json.RawMessage
	_ = json.Unmarshal(raw, &rawv)
	// re-leer variants as raw
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
			t.Fatalf("variante %q inválida: %+v", vr.Name, res.Details)
		}
	}
}

// S-02: casos inválidos fallan con 422 y details menciona expect_field.
func TestAcceptanceS02Invalid(t *testing.T) {
	b := leerTestdata(t, "settings_invalid.json")
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
	// base tomada de settings_valid.json
	vb := leerTestdata(t, "settings_valid.json")
	var vdoc struct {
		Base json.RawMessage `json:"base"`
	}
	_ = json.Unmarshal(vb, &vdoc)
	docBase := vdoc.Base
	for _, c := range doc.Cases {
		merged := mergePatch(docBase, c.Patch)
		res := Validate(merged)
		if len(res.Details) == 0 {
			t.Fatalf("caso %q debería fallar", c.Name)
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

// S-03: logo > 400000 caracteres rechazado.
func TestAcceptanceS03LogoTooLong(t *testing.T) {
	base := leerTestdata(t, "settings_valid.json")
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
		t.Fatal("logo > 400000 debería fallar")
	}
}

// S-04: ratios de contrast_vectors coinciden.
func TestAcceptanceS04ContrastVectors(t *testing.T) {
	b := leerTestdata(t, "contrast_vectors.json")
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

// S-06: RenderCSS determinista.
func TestAcceptanceS06Deterministic(t *testing.T) {
	base := leerTestdata(t, "settings_valid.json")
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
	// solo contiene valores validados: verificar que no haya caracteres peligrosos
	if strings.Contains(css1, "<") || strings.Contains(css1, ">") {
		t.Fatal("RenderCSS contiene caracteres peligrosos")
	}
}

// S-07: name con script se acepta como texto.
func TestAcceptanceS07XSSNameAccepted(t *testing.T) {
	base := leerTestdata(t, "settings_valid.json")
	var doc struct {
		Base json.RawMessage `json:"base"`
	}
	_ = json.Unmarshal(base, &doc)
	patch := json.RawMessage(`{"name":"<script>alert(1)</script>"}`)
	merged := mergePatch(doc.Base, patch)
	res := Validate(merged)
	if len(res.Details) != 0 {
		t.Fatalf("name con script debería validar: %+v", res.Details)
	}
	if res.Settings.Name != "<script>alert(1)</script>" {
		t.Fatalf("name no se conservó: %q", res.Settings.Name)
	}
}

// S-11: warnings no bloquean; <3:1 bloquea.
func TestAcceptanceS11ContrastWarnings(t *testing.T) {
	base := leerTestdata(t, "settings_valid.json")
	var doc struct {
		Base json.RawMessage `json:"base"`
	}
	_ = json.Unmarshal(base, &doc)
	// warning: ratio 3-4.5 → no details, warnings present
	patch := json.RawMessage(`{"c":{"text":"#777777"}}`)
	merged := mergePatch(doc.Base, patch)
	res := Validate(merged)
	if len(res.Details) != 0 {
		t.Fatalf("contraste 3-4.5 no debería bloquear: %+v", res.Details)
	}
	// bloqueo: ratio <3
	patch2 := json.RawMessage(`{"c":{"text":"#ffff00","bg":"#ffffff"}}`)
	merged2 := mergePatch(doc.Base, patch2)
	res2 := Validate(merged2)
	if len(res2.Details) == 0 {
		t.Fatal("contraste <3 debería bloquear")
	}
}
