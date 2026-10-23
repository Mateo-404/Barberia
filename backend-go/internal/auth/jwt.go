package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims are the token claims issued by the backend (ADR-002).
type Claims struct {
	Email    string `json:"email"`
	TenantID int64  `json:"tenant,omitempty"`
	Role     string `json:"role,omitempty"`
	Scope    string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager issues and validates HS256 tokens.
type JWTManager struct {
	secret     []byte
	ttl        time.Duration
	pendingTTL time.Duration
	now        func() time.Time
}

// NewJWTManager builds the issuer. The secret must be at least 32 bytes
// (validated in config; checked again here defensively).
func NewJWTManager(secret string, ttl time.Duration) (*JWTManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 bytes")
	}
	if ttl <= 0 {
		return nil, errors.New("invalid JWT ttl")
	}
	return &JWTManager{secret: []byte(secret), ttl: ttl, pendingTTL: 5 * time.Minute, now: time.Now}, nil
}

// Issue creates a token for the given admin.
func (m *JWTManager) Issue(adminID int64, email string, tenantID int64, role string) (string, error) {
	now := m.now()
	if role == "" {
		role = RoleOwner
	}
	claims := Claims{
		Email:    email,
		TenantID: tenantID,
		Role:     role,
		Scope:    "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(adminID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// IssuePending2FA creates a short-lived token used only to complete a TOTP
// challenge after a successful primary authentication.
func (m *JWTManager) IssuePending2FA(adminID int64, email string, tenantID int64, role string) (string, error) {
	now := m.now()
	if role == "" {
		role = RoleOwner
	}
	claims := Claims{
		Email:    email,
		TenantID: tenantID,
		Role:     role,
		Scope:    "2fa_pending",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(adminID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.pendingTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse validates signature, algorithm and expiry and returns the claims.
func (m *JWTManager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Subject == "" {
		return nil, errors.New("token without sub")
	}
	return claims, nil
}

// ErrInvalidToken is returned for any token that fails validation.
var ErrInvalidToken = errors.New("invalid token")
