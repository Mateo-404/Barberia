package settings

// Theme is the appearance of a tenant's public site (settings v2). It is
// small on purpose: one brand color generates the whole accessible palette
// (see palette.go); Overrides are an escape hatch for specific tokens.
type Theme struct {
	V         int       `json:"v"`
	Name      string    `json:"name"`
	Tagline   string    `json:"tagline"`
	Logo      string    `json:"logo"`
	Preset    string    `json:"preset"`
	Mode      string    `json:"mode"`
	Brand     string    `json:"brand"`
	Font      Font      `json:"font"`
	Shape     Shape     `json:"shape"`
	Overrides Overrides `json:"overrides"`
}

// Font names the heading and body typefaces (from FontAllowlist).
type Font struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}

// Shape sets the corner radius in px (0–24) and the primary button style.
type Shape struct {
	Radius int    `json:"radius"`
	Button string `json:"button"`
}

// Overrides replaces generated tokens per color scheme.
type Overrides struct {
	Light TokenOverrides `json:"light"`
	Dark  TokenOverrides `json:"dark"`
}

// TokenOverrides are optional hex colors; empty means "use the generated one".
type TokenOverrides struct {
	Bg      string `json:"bg,omitempty"`
	Surface string `json:"surface,omitempty"`
	Text    string `json:"text,omitempty"`
	Primary string `json:"primary,omitempty"`
	Accent  string `json:"accent,omitempty"`
}

// Business holds the operating data of the business. It is not part of the
// theme: it is not versioned or published, a change applies when saved.
type Business struct {
	Whatsapp  string `json:"whatsapp"`
	Instagram string `json:"instagram"`
	Address   string `json:"address"`
	Currency  string `json:"currency"`
	Slot      int    `json:"slot"`
	Notice    int    `json:"notice"`
	Open      string `json:"open"`
	Close     string `json:"close"`
	Tz        string `json:"tz"`
}

// Modes of the color scheme.
const (
	ModeLight = "light"
	ModeDark  = "dark"
	ModeAuto  = "auto"
)

// ThemeVersion is the current settings document version.
const ThemeVersion = 2

// DefaultTimezone applies when a business does not set one.
const DefaultTimezone = "America/Argentina/Buenos_Aires"
