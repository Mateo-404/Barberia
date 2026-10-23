package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// defaultMaxBody caps the body when MAX_REQUEST_BODY is not configured.
const defaultMaxBody = 1 << 20

// decodeJSON parses the body into dst. Unknown fields are ignored; an
// unreadable, empty, oversized or trailing-data body answers 400.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) bool {
	if maxBytes <= 0 {
		maxBytes = defaultMaxBody
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		badRequest(w, r)
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		badRequest(w, r)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
