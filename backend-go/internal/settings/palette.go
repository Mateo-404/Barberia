package settings

import "math"

// Scheme is the generated color tokens of one color scheme (light or dark).
type Scheme struct {
	Bg        string `json:"bg"`
	Surface   string `json:"surface"`
	Text      string `json:"text"`
	Muted     string `json:"muted"`
	Primary   string `json:"primary"`
	OnPrimary string `json:"onPrimary"`
	Accent    string `json:"accent"`
	Border    string `json:"border"`
}

// WCAG 2.2 thresholds: 4.5:1 for body text, 3:1 for large text and UI
// components (borders, icons, focus rings).
const (
	minTextContrast = 4.5
	minUIContrast   = 3.0
)

// Lightness anchors per scheme, in OKLCH. Chosen so that text on both bg and
// surface clears AAA (7:1) for any brand hue before any adjustment.
type anchors struct{ bg, surface, text float64 }

var (
	lightAnchors = anchors{bg: 0.965, surface: 0.995, text: 0.22}
	darkAnchors  = anchors{bg: 0.165, surface: 0.215, text: 0.95}
)

// Generate builds both schemes from a brand color. Neutrals carry a whisper
// of the brand hue so the surfaces feel part of the same family; the primary
// keeps the brand color untouched unless it would be hard to see against the
// background, in which case only its lightness moves.
func Generate(brand string, overrides Overrides) (light, dark Scheme) {
	seed, err := hexToOklab(brand)
	if err != nil {
		seed, _ = hexToOklab(Presets()[0].Theme.Brand)
	}
	b := toOklch(seed)
	return buildScheme(b, lightAnchors, false, overrides.Light), buildScheme(b, darkAnchors, true, overrides.Dark)
}

func buildScheme(brand oklch, a anchors, dark bool, over TokenOverrides) Scheme {
	tint := math.Min(brand.C*0.15, 0.02)
	neutral := func(l, chroma float64) string { return hexOf(oklch{L: l, C: chroma, H: brand.H}) }

	s := Scheme{
		Bg:      pick(over.Bg, neutral(a.bg, tint)),
		Surface: pick(over.Surface, neutral(a.surface, tint)),
		Text:    pick(over.Text, neutral(a.text, tint*1.2)),
	}

	primary := over.Primary
	if primary == "" {
		primary = readable(brand, s.Bg, dark)
	}
	s.Primary = primary
	s.OnPrimary = OnPrimary(primary)
	s.Accent = pick(over.Accent, mix(s.Bg, primary, 0.16))
	s.Border = mix(s.Bg, s.Text, 0.16)
	s.Muted = mutedText(s.Text, s.Bg)
	return s
}

func pick(override, generated string) string {
	if override != "" {
		return override
	}
	return generated
}

// readable returns the brand color, moved in lightness only as far as
// needed so that it is visible against bg (3:1, WCAG 1.4.11) and so that
// button text on it reads at 4.5:1: lighter on dark schemes, darker on light
// ones. A brand that already passes is returned as given.
func readable(brand oklch, bg string, lighten bool) string {
	ok := func(hex string) bool {
		ui, _ := Contrast(hex, bg)
		label, _ := Contrast(OnPrimary(hex), hex)
		return ui >= minUIContrast && label >= minTextContrast
	}
	step := -0.01
	if lighten {
		step = 0.01
	}
	for l := brand.L; l >= 0 && l <= 1; l += step {
		if candidate := hexOf(oklch{L: l, C: brand.C, H: brand.H}); ok(candidate) {
			return candidate
		}
	}
	// No lightness of this hue works (never the case for real backgrounds):
	// fall back to plain text color.
	if lighten {
		return "#ffffff"
	}
	return "#111111"
}

// mutedText steps from the background toward the text color until it reads
// at 4.5:1: secondary copy that is still accessible.
func mutedText(text, bg string) string {
	for t := 0.62; t < 1; t += 0.04 {
		candidate := mix(bg, text, t)
		if ratio, _ := Contrast(candidate, bg); ratio >= minTextContrast {
			return candidate
		}
	}
	return text
}
