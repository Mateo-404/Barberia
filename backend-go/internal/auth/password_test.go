package auth

import (
	"strings"
	"testing"
)

// Real hash written by the former Spring backend (Argon2PasswordEncoder) for
// the seed admin password "admin1234".
const springHashAdmin1234 = "$argon2id$v=19$m=19456,t=2,p=1$xHKcOLv7KqB1+lnW/1JLZQ$NjphuKT3aw5qkFQRCefzuMd/38ZZ/8Q94xVReKgxRDM"

func TestVerifyPasswordWithRealSpringHash(t *testing.T) {
	ok, err := VerifyPassword("admin1234", springHashAdmin1234)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("admin1234 should verify against the Spring hash")
	}

	ok, err = VerifyPassword("wrong-password", springHashAdmin1234)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if ok {
		t.Fatal("a different password must not verify")
	}
}

func TestHashPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("anotherPass123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("unexpected prefix: %s", hash)
	}
	ok, err := VerifyPassword("anotherPass123", hash)
	if err != nil || !ok {
		t.Fatalf("roundtrip failed: ok=%v err=%v", ok, err)
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	for _, h := range []string{"", "plaintext", "$2b$10$abcdefghijklmnop", "$argon2id$v=19$m=0,t=2,p=1$AA$AA"} {
		if _, err := VerifyPassword("x", h); err == nil {
			t.Errorf("hash %q should fail", h)
		}
	}
}
