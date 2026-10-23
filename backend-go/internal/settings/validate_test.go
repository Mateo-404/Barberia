package settings

import (
	"encoding/json"
	"testing"
)

func TestValidateFromFixtures(t *testing.T) {
	b := readTestdata(t, "settings_valid.json")
	var v struct {
		Base     json.RawMessage  `json:"base"`
		Variants []map[string]any `json:"variants"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	// base
	if res := Validate(v.Base); len(res.Details) != 0 {
		t.Fatalf("base should be valid: %+v", res.Details)
	}
	// variants simplified
}

func TestValidateInvalidFixtures(t *testing.T) {
	b := readTestdata(t, "settings_invalid.json")
	var v struct {
		Cases []struct {
			Name        string          `json:"name"`
			Patch       json.RawMessage `json:"patch"`
			ExpectField string          `json:"expect_field"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Cases) == 0 {
		t.Fatal("no cases")
	}
}
