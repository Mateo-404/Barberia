package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims son los claims emitidos por el backend. Los cuatro primeros son
// idénticos a los del Java; "tenant" es aditivo (ADR-002).
type Claims struct {
	Email    string `json:"email"`
	TenantID int64  `json:"tenant,omitempty"`
	Rol      string `json:"rol,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager emite y valida tokens HS256.
type JWTManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewJWTManager crea el emisor. secret debe tener al menos 32 bytes (validado
// en config); aquí se vuelve a comprobar para uso defensivo.
func NewJWTManager(secret string, ttl time.Duration) (*JWTManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET debe tener al menos 32 bytes")
	}
	if ttl <= 0 {
		return nil, errors.New("ttl de JWT inválido")
	}
	return &JWTManager{secret: []byte(secret), ttl: ttl, now: time.Now}, nil
}

// Issue genera un token para el administrador dado.
func (m *JWTManager) Issue(adminID int64, email string, tenantID int64, rol string) (string, error) {
	now := m.now()
	if rol == "" {
		rol = RolOwner
	}
	claims := Claims{
		Email:    email,
		TenantID: tenantID,
		Rol:      rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(adminID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse valida firma, algoritmo y expiración, y devuelve los claims.
func (m *JWTManager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, errors.New("token inválido")
	}
	if claims.Subject == "" {
		return nil, errors.New("token sin sub")
	}
	return claims, nil
}

// ErrTokenInvalido error para tokens inválidos
var ErrTokenInvalido = errors.New("token inválido")
