package settings

import (
	"math/rand"
	"strings"
	"testing"
)

// assertReadable checks every token pair the UI relies on.
func assertReadable(t *testing.T, label string, s Scheme) {
	t.Helper()
	check := func(name, fg, bg string, min float64) {
		if ratio, err := Contrast(fg, bg); err != nil || ratio < min-1e-9 {
			t.Errorf("%s: %s %s on %s = %.2f, want >= %.1f (%v)", label, name, fg, bg, ratio, min, err)
		}
	}
	check("text/bg", s.Text, s.Bg, 7)
	check("text/surface", s.Text, s.Surface, 7)
	check("muted/bg", s.Muted, s.Bg, minTextContrast)
	check("primary/bg", s.Primary, s.Bg, minUIContrast)
	check("onPrimary/primary", s.OnPrimary, s.Primary, minTextContrast)
}

func TestEveryPresetIsAccessibleInBothSchemes(t *testing.T) {
	for _, p := range Presets() {
		light, dark := Generate(p.Theme.Brand, Overrides{})
		assertReadable(t, p.ID+"/light", light)
		assertReadable(t, p.ID+"/dark", dark)
	}
}

// Whatever brand color an owner picks, the generated palette stays readable.
func TestRandomBrandsAreAccessible(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	const hex = "0123456789abcdef"
	for i := 0; i < 2000; i++ {
		var b strings.Builder
		b.WriteByte('#')
		for j := 0; j < 6; j++ {
			b.WriteByte(hex[rng.Intn(16)])
		}
		light, dark := Generate(b.String(), Overrides{})
		assertReadable(t, b.String()+"/light", light)
		assertReadable(t, b.String()+"/dark", dark)
	}
}

func TestBrandColorIsKeptWhenAlreadyVisible(t *testing.T) {
	_, dark := Generate("#c9a24b", Overrides{})
	if dark.Primary != "#c9a24b" {
		t.Fatalf("a visible brand must not be altered, got %s", dark.Primary)
	}
	// A dark brand on a dark background gets lighter, keeping its hue.
	_, darkBlue := Generate("#0a2a66", Overrides{})
	if darkBlue.Primary == "#0a2a66" {
		t.Fatalf("an invisible brand on dark should be lightened")
	}
}

func TestOverridesWinOverGeneratedTokens(t *testing.T) {
	light, _ := Generate("#c9a24b", Overrides{Light: TokenOverrides{Bg: "#ffffff", Primary: "#112233"}})
	if light.Bg != "#ffffff" || light.Primary != "#112233" {
		t.Fatalf("overrides ignored: %+v", light)
	}
	if light.OnPrimary != OnPrimary("#112233") {
		t.Fatalf("onPrimary must follow the overridden primary: %s", light.OnPrimary)
	}
}

func TestOklchRoundTrip(t *testing.T) {
	for _, hex := range []string{"#000000", "#ffffff", "#c9a24b", "#2f5d8a", "#9c3b3b", "#808080"} {
		lab, err := hexToOklab(hex)
		if err != nil {
			t.Fatal(err)
		}
		if got := hexOf(toOklch(lab)); got != hex {
			t.Errorf("round trip %s -> %s", hex, got)
		}
	}
}

func TestHexOfClipsToGamutKeepingLightness(t *testing.T) {
	// Chroma 0.4 is far outside sRGB for any lightness; the result must be a
	// valid color of about the same lightness.
	got := hexOf(oklch{L: 0.6, C: 0.4, H: 145})
	lab, err := hexToOklab(got)
	if err != nil {
		t.Fatal(err)
	}
	if lab.L < 0.58 || lab.L > 0.62 {
		t.Fatalf("lightness drifted: %s has L=%.3f", got, lab.L)
	}
}
