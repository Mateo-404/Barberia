package tenant

import "testing"

// normalizarHost replica la extracción del Host de la request (ADR-002):
// sin puerto, minúsculas, localhost implícito → tenant por defecto.
func TestNormalizarHost(t *testing.T) {
	casos := []struct{ in, want string }{
		{"salon2.example.com", "salon2.example.com"},
		{"Salon2.Example.com", "salon2.example.com"},
		{"salon2.example.com:8080", "salon2.example.com"},
		{"  salon2.example.com  ", "salon2.example.com"},
		// En desarrollo todos los frontends van contra el tenant por defecto.
		{"localhost", ""},
		{"localhost:5173", ""},
		{"127.0.0.1:8080", ""},
		{"[::1]:8080", ""},
		{"", ""},
	}
	for _, c := range casos {
		if got := normalizarHost(c.in); got != c.want {
			t.Errorf("normalizarHost(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// El tenant NUNCA se toma de una cabecera: normalizarHost solo recibe el Host.
// Verificado también end-to-end en internal/integration (TestAislamientoDeTenant).
