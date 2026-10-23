package httpapi

import (
	"net/http"
	"regexp"
	"strings"
)

// emailValido aproxima @Email de Jakarta Bean Validation (Hibernate Validator):
// local@domain con puntos opcionales en el dominio.
var emailValido = regexp.MustCompile(
	`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// Límites de longitud para campos de texto (no existen en Java, cuyas columnas
// son TEXT; evitan almacenar payloads desproporcionados).
const (
	maxNombre       = 120
	maxEmail        = 254
	maxTelefono     = 32
	maxTipoServicio = 120
)

// validarNotBlank replica @NotBlank: null o solo espacios es inválido.
// Responde el 400 genérico y devuelve false.
func validarNotBlank(w http.ResponseWriter, r *http.Request, v *string) bool {
	if v == nil || strings.TrimSpace(*v) == "" {
		badRequestValidacion(w, r)
		return false
	}
	return true
}

// validarEmail replica @Email: null es válido; no-null debe cumplir el patrón.
func validarEmail(w http.ResponseWriter, r *http.Request, v *string) bool {
	if v == nil {
		return true
	}
	if !emailValido.MatchString(*v) {
		badRequestValidacion(w, r)
		return false
	}
	return true
}

// validarMinLength replica @Size(min=n) sobre String.
func validarMinLength(w http.ResponseWriter, r *http.Request, v *string, min int) bool {
	if v == nil || len(*v) < min {
		badRequestValidacion(w, r)
		return false
	}
	return true
}

// validarMaxLength acota la longitud de un texto: null es válido, no-null no
// puede superar max. No hay equivalente en Java (sus columnas son TEXT), pero
// evita almacenar payloads desproporcionados por campos de texto.
func validarMaxLength(w http.ResponseWriter, r *http.Request, v *string, max int) bool {
	if v != nil && len(*v) > max {
		badRequestValidacion(w, r)
		return false
	}
	return true
}

// validarOpcionalMinLength aplica @Size(min=n) solo cuando el campo viene
// presente (PATCH parcial), en línea con el PATCH sin validación de Java.
func validarOpcionalMinLength(w http.ResponseWriter, r *http.Request, v *string, min int) bool {
	if v != nil && len(*v) < min {
		badRequestValidacion(w, r)
		return false
	}
	return true
}

// validarOpcionalEmail valida @Email solo si el email viene presente.
func validarOpcionalEmail(w http.ResponseWriter, r *http.Request, v *string) bool {
	if v != nil && !emailValido.MatchString(*v) {
		badRequestValidacion(w, r)
		return false
	}
	return true
}

// validarOpcionalPositivo replica @Positive solo si el número viene presente.
func validarOpcionalPositivo(w http.ResponseWriter, r *http.Request, v *float32) bool {
	if v != nil && *v <= 0 {
		badRequestValidacion(w, r)
		return false
	}
	return true
}
