package settings

// Preset is a curated starting point: brand color, typography, shape and
// color mode that already work well together. Picking one fills the editor;
// the saved document is self-contained and only remembers the preset's id.
type Preset struct {
	ID    string `json:"id"`
	Theme Theme  `json:"theme"`
}

func preset(id, brand, mode, heading, body string, radius int, button string) Preset {
	return Preset{ID: id, Theme: Theme{
		V: ThemeVersion, Preset: id, Mode: mode, Brand: brand,
		Font:  Font{Heading: heading, Body: body},
		Shape: Shape{Radius: radius, Button: button},
	}}
}

// Presets lists the starting points, in display order. The first is the
// default used for tenants that never configured anything.
func Presets() []Preset {
	return []Preset{
		preset("classic", "#c9a24b", ModeDark, "Playfair Display", "Inter", 8, "solid"),
		preset("vintage", "#9c3b3b", ModeLight, "Cormorant Garamond", "Lora", 4, "solid"),
		preset("minimal", "#2f5d8a", ModeLight, "DM Sans", "DM Sans", 10, "solid"),
		preset("urban", "#f2711c", ModeDark, "Bebas Neue", "Barlow Condensed", 2, "solid"),
		preset("natural", "#4c7a52", ModeLight, "Lora", "Work Sans", 14, "solid"),
		preset("navy", "#b08d57", ModeAuto, "Cinzel", "Montserrat", 6, "outline"),
	}
}

// PresetByID returns a preset or false when the id is unknown.
func PresetByID(id string) (Preset, bool) {
	for _, p := range Presets() {
		if p.ID == id {
			return p, true
		}
	}
	return Preset{}, false
}

// CustomPreset is the id of a document that was not built from a preset
// (or was edited so far from it that the id no longer describes it).
const CustomPreset = "custom"

// DefaultTheme is the document of a tenant that has not configured anything.
func DefaultTheme() Theme {
	t := Presets()[0].Theme
	t.Name = "Barbería"
	return t
}
