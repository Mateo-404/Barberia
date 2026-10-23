package tenant

import "testing"

// normalizeHost strips the port, lower-cases, and maps localhost to the
// default tenant (ADR-002).
func TestNormalizeHost(t *testing.T) {
	cases := []struct{ in, want string }{
		{"salon2.example.com", "salon2.example.com"},
		{"Salon2.Example.com", "salon2.example.com"},
		{"salon2.example.com:8080", "salon2.example.com"},
		{"  salon2.example.com  ", "salon2.example.com"},
		// In development every frontend talks to the default tenant.
		{"localhost", ""},
		{"localhost:5173", ""},
		{"127.0.0.1:8080", ""},
		{"[::1]:8080", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := normalizeHost(c.in); got != c.want {
			t.Errorf("normalizeHost(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// The tenant is never taken from a header: normalizeHost only sees the Host.
// Also verified end to end in internal/integration (TestTenantIsolation).
