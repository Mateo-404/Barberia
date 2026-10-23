package settings

var FontAllowlist = map[string]string{
	"Playfair Display":   "serif",
	"Lora":               "serif",
	"Merriweather":       "serif",
	"Cinzel":             "serif",
	"Cormorant Garamond": "serif",
	"Libre Baskerville":  "serif",
	"Inter":              "sans-serif",
	"DM Sans":            "sans-serif",
	"Montserrat":         "sans-serif",
	"Poppins":            "sans-serif",
	"Raleway":            "sans-serif",
	"Roboto":             "sans-serif",
	"Rubik":              "sans-serif",
	"Work Sans":          "sans-serif",
	"Nunito":             "sans-serif",
	"Josefin Sans":       "sans-serif",
	"Space Grotesk":      "sans-serif",
	"Bebas Neue":         "display",
	"Oswald":             "sans-serif",
	"Barlow Condensed":   "sans-serif",
	"Archivo Black":      "sans-serif",
	"Anton":              "sans-serif",
}

func fontGeneric(f string) string {
	if g, ok := FontAllowlist[f]; ok {
		if g == "display" {
			return "sans-serif"
		}
		return g
	}
	return "sans-serif"
}
