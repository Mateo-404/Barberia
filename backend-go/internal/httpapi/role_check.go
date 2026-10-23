package httpapi

import (
	"net/http"
	"strings"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
)

// middlewareAdminRole exige que el token tenga rol owner/admin para /api/admin/*.
func (s *Server) middlewareAdminRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/admin/") {
			claims, _ := r.Context().Value(ctxKeyClaims).(*auth.Claims)
			if claims == nil || !auth.EsAdminRol(claims.Rol) {
				forbidden(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
