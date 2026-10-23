package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// defaultMaxBody limita el tamaño del body cuando MAX_REQUEST_BODY no está
// configurado (el valor efectivo lo aplica el handler desde cfg).
const defaultMaxBody = 1 << 20

// decodificarJSONObligatorio parsea el body con la semántica de Jackson: se
// ignoran campos desconocidos; JSON ilegible, vacío o gigante → 400 réplica.
func decodificarJSONObligatorio(w http.ResponseWriter, r *http.Request, dst any) bool {
	return decodificarJSONConLimite(w, r, dst, defaultMaxBody)
}

// decodificarJSONConLimite aplica el tope de tamaño configurado.
func decodificarJSONConLimite(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) bool {
	if maxBytes <= 0 {
		maxBytes = defaultMaxBody
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		badRequestValidacion(w, r)
		return false
	}
	// Rechaza contenido extra tras el primer valor JSON.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		badRequestValidacion(w, r)
		return false
	}
	return true
}

func responderJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
