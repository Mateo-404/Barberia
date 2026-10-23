package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GoogleConfig holds the OAuth client settings.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Enabled reports whether Google login is configured.
func (c GoogleConfig) Enabled() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// GoogleUser is the verified identity returned by Google.
type GoogleUser struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// GoogleAuthURL builds the authorization URL with the given state.
func (c GoogleConfig) AuthURL(state string) string {
	v := url.Values{}
	v.Set("client_id", c.ClientID)
	v.Set("redirect_uri", c.RedirectURL)
	v.Set("response_type", "code")
	v.Set("scope", "openid email profile")
	v.Set("state", state)
	v.Set("access_type", "online")
	v.Set("prompt", "select_account")
	return "https://accounts.google.com/o/oauth2/v2/auth?" + v.Encode()
}

// ExchangeCode swaps an authorization code for an ID token.
func (c GoogleConfig) ExchangeCode(code string) (string, error) {
	v := url.Values{}
	v.Set("code", code)
	v.Set("client_id", c.ClientID)
	v.Set("client_secret", c.ClientSecret)
	v.Set("redirect_uri", c.RedirectURL)
	v.Set("grant_type", "authorization_code")
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", v)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed: %s", string(body))
	}
	var out struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.IDToken == "" {
		return "", fmt.Errorf("no id_token in response")
	}
	return out.IDToken, nil
}

// VerifyIDToken validates an ID token via Google's tokeninfo endpoint.
func (c GoogleConfig) VerifyIDToken(idToken string) (GoogleUser, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken))
	if err != nil {
		return GoogleUser{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return GoogleUser{}, fmt.Errorf("invalid id token: %s", string(body))
	}
	var raw struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Aud           string `json:"aud"`
		Iss           string `json:"iss"`
		Exp           string `json:"exp"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return GoogleUser{}, err
	}
	if raw.Aud != c.ClientID {
		return GoogleUser{}, fmt.Errorf("id token audience mismatch")
	}
	if raw.Iss != "https://accounts.google.com" && raw.Iss != "accounts.google.com" {
		return GoogleUser{}, fmt.Errorf("id token issuer mismatch")
	}
	if raw.Sub == "" || raw.Email == "" {
		return GoogleUser{}, fmt.Errorf("id token missing sub/email")
	}
	// expiry check
	if exp, err := strconv.ParseInt(raw.Exp, 10, 64); err == nil {
		if time.Now().Unix() > exp {
			return GoogleUser{}, fmt.Errorf("id token expired")
		}
	}
	return GoogleUser{
		Sub:           raw.Sub,
		Email:         raw.Email,
		EmailVerified: raw.EmailVerified == "true",
		Name:          raw.Name,
		Picture:       raw.Picture,
	}, nil
}

// StateToken helpers: state is a signed short-lived JWT-ish string.
// We reuse a simple HMAC-signed payload to avoid extra deps.
func SignState(secret []byte, payload string) string {
	mac := hmacSHA256(secret, payload)
	return payload + "." + base64Raw(mac)
}

func VerifyState(secret []byte, state string) (string, bool) {
	idx := lastDot(state)
	if idx < 0 {
		return "", false
	}
	payload := state[:idx]
	sig := state[idx+1:]
	if !constantTimeEq(sig, base64Raw(hmacSHA256(secret, payload))) {
		return "", false
	}
	return payload, true
}
