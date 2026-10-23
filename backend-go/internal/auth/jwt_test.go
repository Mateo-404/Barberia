package auth

import (
	"testing"
	"time"
)

func TestJWTIssueParse(t *testing.T) {
	m, err := NewJWTManager("test-secret-de-al-menos-32-bytes!!", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := m.Issue(7, "admin@admin.com", 3, "owner")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "7" || claims.Email != "admin@admin.com" || claims.TenantID != 3 {
		t.Fatalf("claims inesperados: %+v", claims)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) < 23*time.Hour {
		t.Fatalf("expiración inesperada: %v", claims.ExpiresAt)
	}
}

func TestJWTTokenLegadoSinTenant(t *testing.T) {
	m, err := NewJWTManager("test-secret-de-al-menos-32-bytes!!", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	// Token generado por el Java: solo sub, email, iat, exp.
	legacy, err := m.Issue(1, "admin@admin.com", 0, "owner")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.Parse(legacy)
	if err != nil {
		t.Fatalf("token legado debería aceptarse: %v", err)
	}
	if claims.TenantID != 0 {
		t.Fatalf("tenant debería resolverse por BD (0 en claims), got %d", claims.TenantID)
	}
}

func TestJWTRechazaFirmaInvalida(t *testing.T) {
	m, _ := NewJWTManager("test-secret-de-al-menos-32-bytes!!", time.Hour)
	otro, _ := NewJWTManager("otro-secreto-tambien-largo-32bytes!", time.Hour)
	token, _ := otro.Issue(1, "a@b.com", 1, "owner")
	if _, err := m.Parse(token); err == nil {
		t.Fatal("token firmado con otra clave debería rechazarse")
	}
}

func TestJWTRechazaAlgDistintoDeHS256(t *testing.T) {
	m, _ := NewJWTManager("test-secret-de-al-menos-32-bytes!!", time.Hour)
	if _, err := m.Parse("eyJhbGciOiJub25lIn0.eyJzdWIiOiIxIn0."); err == nil {
		t.Fatal("alg none debería rechazarse")
	}
	if _, err := m.Parse("token-basura"); err == nil {
		t.Fatal("token basura debería rechazarse")
	}
	_ = m
}
