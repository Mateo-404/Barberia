package httpapi

import (
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
)

// requireRole lets the request through only when the token's role is one of
// roles (the auth middleware already validated the token).
func requireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, _ := r.Context().Value(ctxKeyClaims).(*auth.Claims)
			if claims == nil {
				unauthorized(w, r)
				return
			}
			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			forbidden(w, r)
		})
	}
}
