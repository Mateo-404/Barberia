package settings

// FontAllowlist is every typeface a tenant may pick, with the CSS generic
// family used as fallback. The frontend ships a self-hosted copy of each
// (frontend/src/lib/fonts.ts): keep both lists in sync.
var FontAllowlist = map[string]string{
	"Playfair Display":   "serif",
	"Lora":               "serif",
	"Merriweather":       "serif",
	"Cinzel":             "serif",
	"Cormorant Garamond": "serif",
	"Inter":              "sans-serif",
	"DM Sans":            "sans-serif",
	"Montserrat":         "sans-serif",
	"Poppins":            "sans-serif",
	"Raleway":            "sans-serif",
	"Work Sans":          "sans-serif",
	"Bebas Neue":         "sans-serif",
	"Oswald":             "sans-serif",
	"Barlow Condensed":   "sans-serif",
}

func fontStack(name string) string {
	generic, ok := FontAllowlist[name]
	if !ok {
		generic = "sans-serif"
	}
	return `"` + name + `",` + generic
}
