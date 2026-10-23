package settings

import (
	"encoding/json"
	"math"
	"testing"
)

func TestContrastVectors(t *testing.T) {
	b := readTestdata(t, "contrast_vectors.json")
	var v struct {
		Tolerance float64 `json:"tolerance"`
		Contrast  []struct {
			FG    string  `json:"fg"`
			BG    string  `json:"bg"`
			Ratio float64 `json:"ratio"`
		} `json:"contrast"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
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
