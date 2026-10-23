package auth

import (
	"testing"
	"time"
)

const testSecret = "test-secret-at-least-32-bytes-long!"

func TestJWTIssueParse(t *testing.T) {
	m, err := NewJWTManager(testSecret, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := m.Issue(7, "admin@admin.com", 3, RoleOwner)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "7" || claims.Email != "admin@admin.com" || claims.TenantID != 3 || claims.Role != RoleOwner {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) < 23*time.Hour {
		t.Fatalf("unexpected expiry: %v", claims.ExpiresAt)
	}
}

func TestJWTWithoutTenantParsesWithZeroTenant(t *testing.T) {
	m, err := NewJWTManager(testSecret, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := m.Issue(1, "admin@admin.com", 0, RoleOwner)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("token should parse: %v", err)
	}
	// The HTTP middleware rejects it with 401; Parse only validates the token.
	if claims.TenantID != 0 {
		t.Fatalf("tenant = %d, want 0", claims.TenantID)
	}
}

func TestJWTRejectsWrongSignature(t *testing.T) {
	m, _ := NewJWTManager(testSecret, time.Hour)
	other, _ := NewJWTManager("another-secret-also-32-bytes-long!!", time.Hour)
	token, _ := other.Issue(1, "a@b.com", 1, RoleOwner)
	if _, err := m.Parse(token); err == nil {
		t.Fatal("a token signed with another key must be rejected")
	}
}

func TestJWTRejectsNonHS256(t *testing.T) {
	m, _ := NewJWTManager(testSecret, time.Hour)
	if _, err := m.Parse("eyJhbGciOiJub25lIn0.eyJzdWIiOiIxIn0."); err == nil {
		t.Fatal("alg none must be rejected")
	}
	if _, err := m.Parse("garbage-token"); err == nil {
		t.Fatal("a garbage token must be rejected")
	}
}
