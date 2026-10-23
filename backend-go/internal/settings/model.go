package settings

import "encoding/json"

type Colors struct {
	Primary string `json:"primary"`
	Accent  string `json:"accent"`
	Bg      string `json:"bg"`
	Surface string `json:"surface"`
	Text    string `json:"text"`
}

type Biz struct {
	Whatsapp  string  `json:"whatsapp"`
	Instagram string  `json:"instagram"`
	Address   string  `json:"address"`
	Currency  string  `json:"currency"`
	Slot      int     `json:"slot"`
	Notice    int     `json:"notice"`
	Open      string  `json:"open"`
	Close     string  `json:"close"`
	Tz        *string `json:"tz"`
}

type Settings struct {
	Name    string          `json:"name"`
	Tagline string          `json:"tagline"`
	Logo    string          `json:"logo"`
	C       Colors          `json:"c"`
	HF      string          `json:"hf"`
	BF      string          `json:"bf"`
	R       int             `json:"r"`
	Btn     string          `json:"btn"`
	Biz     Biz             `json:"biz"`
	Raw     json.RawMessage `json:"-"`
}
