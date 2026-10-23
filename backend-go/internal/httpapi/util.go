package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// chiURLParam es un alias corto para URLParam.
func chiURLParam(r *http.Request, nombre string) string { return chi.URLParam(r, nombre) }
