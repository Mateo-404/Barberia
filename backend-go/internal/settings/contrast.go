package settings

import (
	"fmt"
	"math"
)

func hexToRGB(h string) (float64, float64, float64, error) {
	if len(h) != 7 || h[0] != '#' {
		return 0, 0, 0, fmt.Errorf("invalid hex")
	}
	var r, g, b int
	_, err := fmt.Sscanf(h, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, 0, 0, err
	}
	return float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0, nil
}

func lin(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func luminance(h string) (float64, error) {
	r, g, b, err := hexToRGB(h)
	if err != nil {
		return 0, err
	}
	R := lin(r)
	G := lin(g)
	B := lin(b)
	return 0.2126*R + 0.7152*G + 0.0722*B, nil
}

func Contrast(fg, bg string) (float64, error) {
	L1, err := luminance(fg)
	if err != nil {
		return 0, err
	}
	L2, err := luminance(bg)
	if err != nil {
		return 0, err
	}
	if L1 < L2 {
		L1, L2 = L2, L1
	}
	return (L1 + 0.05) / (L2 + 0.05), nil
}

func OnPrimary(bg string) string {
	c1, _ := Contrast("#111111", bg)
	c2, _ := Contrast("#ffffff", bg)
	if c1 > c2+1e-12 {
		return "#111111"
	}
	if c2 > c1+1e-12 {
		return "#ffffff"
	}
	return "#111111"
}
