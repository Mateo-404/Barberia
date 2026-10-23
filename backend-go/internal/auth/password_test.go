package auth

import (
	"strings"
	"testing"
)

// Hash real generado por el backend Java (Argon2PasswordEncoder de Spring)
// para la contraseña "admin1234" del admin semilla.
const hashSpringAdmin1234 = "$argon2id$v=19$m=19456,t=2,p=1$xHKcOLv7KqB1+lnW/1JLZQ$NjphuKT3aw5qkFQRCefzuMd/38ZZ/8Q94xVReKgxRDM"

func TestVerifyPasswordHashRealDeSpring(t *testing.T) {
	ok, err := VerifyPassword("admin1234", hashSpringAdmin1234)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("la contraseña admin1234 debería verificar contra el hash de Spring")
	}

	ok, err = VerifyPassword("clave-equivocada", hashSpringAdmin1234)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if ok {
		t.Fatal("una contraseña distinta no debería verificar")
	}
}

func TestHashPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("otraClave123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("prefijo inesperado: %s", hash)
	}
	ok, err := VerifyPassword("otraClave123", hash)
	if err != nil || !ok {
		t.Fatalf("roundtrip falló: ok=%v err=%v", ok, err)
	}
}

func TestVerifyPasswordMalformado(t *testing.T) {
	for _, h := range []string{"", "plaintext", "$2b$10$abcdefghijklmnop", "$argon2id$v=19$m=0,t=2,p=1$AA$AA"} {
		if _, err := VerifyPassword("x", h); err == nil {
			t.Errorf("hash %q debería fallar", h)
		}
	}
}
