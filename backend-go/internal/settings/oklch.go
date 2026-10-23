package settings

import (
	"fmt"
	"math"
)

// OKLCH is the perceptually uniform color space the palette generator works
// in (Björn Ottosson's Oklab, in polar form). Moving L changes how light a
// color looks without shifting its hue, which sRGB/HSL cannot promise.
type oklch struct{ L, C, H float64 }

func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func linearToSRGB(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1/2.4) - 0.055
}

type oklab struct{ L, A, B float64 }

func linearRGBToOklab(r, g, b float64) oklab {
	l := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*b)
	m := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*b)
	s := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*b)
	return oklab{
		L: 0.2104542553*l + 0.7936177850*m - 0.0040720468*s,
		A: 1.9779984951*l - 2.4285922050*m + 0.4505937099*s,
		B: 0.0259040371*l + 0.7827717662*m - 0.8086757660*s,
	}
}

func oklabToLinearRGB(c oklab) (r, g, b float64) {
	l := c.L + 0.3963377774*c.A + 0.2158037573*c.B
	m := c.L - 0.1055613458*c.A - 0.0638541728*c.B
	s := c.L - 0.0894841775*c.A - 1.2914855480*c.B
	l, m, s = l*l*l, m*m*m, s*s*s
	return 4.0767416621*l - 3.3077115913*m + 0.2309699292*s,
		-1.2684380046*l + 2.6097574011*m - 0.3413193965*s,
		-0.0041960863*l - 0.7034186147*m + 1.7076147010*s
}

func hexToOklab(hex string) (oklab, error) {
	r, g, b, err := hexToRGB(hex)
	if err != nil {
		return oklab{}, err
	}
	return linearRGBToOklab(srgbToLinear(r), srgbToLinear(g), srgbToLinear(b)), nil
}

func oklabToHex(c oklab) string {
	r, g, b := oklabToLinearRGB(c)
	to8 := func(v float64) int { return int(math.Round(math.Min(1, math.Max(0, linearToSRGB(v))) * 255)) }
	return fmt.Sprintf("#%02x%02x%02x", to8(r), to8(g), to8(b))
}

func toOklch(c oklab) oklch {
	h := math.Atan2(c.B, c.A) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return oklch{L: c.L, C: math.Hypot(c.A, c.B), H: h}
}

func (c oklch) lab() oklab {
	rad := c.H * math.Pi / 180
	return oklab{L: c.L, A: c.C * math.Cos(rad), B: c.C * math.Sin(rad)}
}

func inGamut(c oklab) bool {
	const eps = 1e-4
	r, g, b := oklabToLinearRGB(c)
	return r >= -eps && r <= 1+eps && g >= -eps && g <= 1+eps && b >= -eps && b <= 1+eps
}

// hexOf renders an OKLCH color as hex, lowering chroma until it fits sRGB
// (keeps lightness and hue, which is what the eye notices first).
func hexOf(c oklch) string {
	c.L = math.Min(1, math.Max(0, c.L))
	if inGamut(c.lab()) {
		return oklabToHex(c.lab())
	}
	lo, hi := 0.0, c.C
	for i := 0; i < 24; i++ {
		mid := (lo + hi) / 2
		if inGamut(oklch{c.L, mid, c.H}.lab()) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return oklabToHex(oklch{c.L, lo, c.H}.lab())
}

// mix interpolates two hex colors in Oklab; t=0 is a, t=1 is b.
func mix(a, b string, t float64) string {
	ca, errA := hexToOklab(a)
	cb, errB := hexToOklab(b)
	if errA != nil || errB != nil {
		return a
	}
	return oklabToHex(oklab{
		L: ca.L + (cb.L-ca.L)*t,
		A: ca.A + (cb.A-ca.A)*t,
		B: ca.B + (cb.B-ca.B)*t,
	})
}
