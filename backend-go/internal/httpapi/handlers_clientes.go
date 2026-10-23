package httpapi

import (
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/modules/clientes"
)

// clienteDTO replica ClienteResponseDTO.
type clienteDTO struct {
	Telefono string  `json:"telefono"`
	Nombre   string  `json:"nombre"`
	Apellido string  `json:"apellido"`
	Email    *string `json:"email"`
}

type clienteCreateReq struct {
	Telefono *string `json:"telefono"`
	Nombre   *string `json:"nombre"`
	Apellido *string `json:"apellido"`
	Email    *string `json:"email"`
}

type clientePatchReq struct {
	// El teléfono (PK) se ignora deliberadamente; ver matriz de compatibilidad.
	Telefono *string `json:"telefono"`
	Nombre   *string `json:"nombre"`
	Apellido *string `json:"apellido"`
	Email    *string `json:"email"`
}

func (s *Server) hClienteList(w http.ResponseWriter, r *http.Request) {
	tid, _ := tenantID(r)
	lista, err := s.clientes.List(r.Context(), tid)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]clienteDTO, 0, len(lista))
	for _, c := range lista {
		out = append(out, clienteDTO{Telefono: c.Telefono, Nombre: c.Nombre, Apellido: c.Apellido, Email: c.Email})
	}
	responderJSON(w, http.StatusOK, out)
}

func (s *Server) hClienteCreate(w http.ResponseWriter, r *http.Request) {
	var req clienteCreateReq
	if !decodificarJSONConLimite(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	if !validarNotBlank(w, r, req.Telefono) || !validarNotBlank(w, r, req.Nombre) ||
		!validarNotBlank(w, r, req.Apellido) || !validarEmail(w, r, req.Email) ||
		!validarMaxLength(w, r, req.Telefono, maxTelefono) ||
		!validarMaxLength(w, r, req.Nombre, maxNombre) ||
		!validarMaxLength(w, r, req.Apellido, maxNombre) ||
		!validarMaxLength(w, r, req.Email, maxEmail) {
		return
	}
	tid, _ := tenantID(r)
	c, err := s.clientes.Create(r.Context(), tid, clientes.Input{
		Telefono: *req.Telefono, Nombre: *req.Nombre, Apellido: *req.Apellido, Email: req.Email,
	})
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	responderJSON(w, http.StatusCreated, clienteDTO{Telefono: c.Telefono, Nombre: c.Nombre, Apellido: c.Apellido, Email: c.Email})
}

func (s *Server) hClienteCreateAll(w http.ResponseWriter, r *http.Request) {
	var reqs []clienteCreateReq
	if !decodificarJSONConLimite(w, r, &reqs, s.cfg.MaxRequestBody) {
		return
	}
	inputs := make([]clientes.Input, 0, len(reqs))
	for _, req := range reqs {
		if !validarNotBlank(w, r, req.Telefono) || !validarNotBlank(w, r, req.Nombre) ||
			!validarNotBlank(w, r, req.Apellido) || !validarEmail(w, r, req.Email) ||
			!validarMaxLength(w, r, req.Telefono, maxTelefono) ||
			!validarMaxLength(w, r, req.Nombre, maxNombre) ||
			!validarMaxLength(w, r, req.Apellido, maxNombre) ||
			!validarMaxLength(w, r, req.Email, maxEmail) {
			return
		}
		inputs = append(inputs, clientes.Input{
			Telefono: *req.Telefono, Nombre: *req.Nombre, Apellido: *req.Apellido, Email: req.Email,
		})
	}
	tid, _ := tenantID(r)
	lista, err := s.clientes.CreateAll(r.Context(), tid, inputs)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]clienteDTO, 0, len(lista))
	for _, c := range lista {
		out = append(out, clienteDTO{Telefono: c.Telefono, Nombre: c.Nombre, Apellido: c.Apellido, Email: c.Email})
	}
	responderJSON(w, http.StatusCreated, out)
}

func (s *Server) hClienteGet(w http.ResponseWriter, r *http.Request) {
	telefono := chiURLParam(r, "telefono")
	tid, _ := tenantID(r)
	c, err := s.clientes.Get(r.Context(), tid, telefono)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	if c == nil {
		notFoundVacio(w)
		return
	}
	responderJSON(w, http.StatusOK, clienteDTO{Telefono: c.Telefono, Nombre: c.Nombre, Apellido: c.Apellido, Email: c.Email})
}

func (s *Server) hClientePatch(w http.ResponseWriter, r *http.Request) {
	telefono := chiURLParam(r, "telefono")
	var req clientePatchReq
	if !decodificarJSONConLimite(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	// Validación de los campos presentes (PATCH parcial sin reglas en Java).
	if !validarOpcionalEmail(w, r, req.Email) ||
		!validarMaxLength(w, r, req.Nombre, maxNombre) ||
		!validarMaxLength(w, r, req.Apellido, maxNombre) ||
		!validarMaxLength(w, r, req.Email, maxEmail) {
		return
	}
	tid, _ := tenantID(r)
	c, err := s.clientes.Update(r.Context(), tid, telefono, clientes.Patch{
		Nombre: req.Nombre, Apellido: req.Apellido, Email: req.Email,
	})
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	if c == nil {
		notFoundVacio(w)
		return
	}
	responderJSON(w, http.StatusOK, clienteDTO{Telefono: c.Telefono, Nombre: c.Nombre, Apellido: c.Apellido, Email: c.Email})
}

func (s *Server) hClienteDelete(w http.ResponseWriter, r *http.Request) {
	telefono := chiURLParam(r, "telefono")
	tid, _ := tenantID(r)
	borrado, err := s.clientes.Delete(r.Context(), tid, telefono)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	if !borrado {
		notFoundVacio(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) hClientePage(w http.ResponseWriter, r *http.Request) {
	page, size, ok := s.paramPagina(w, r)
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	lista, total, err := s.clientes.Page(r.Context(), tid, page, size)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]clienteDTO, 0, len(lista))
	for _, c := range lista {
		out = append(out, clienteDTO{Telefono: c.Telefono, Nombre: c.Nombre, Apellido: c.Apellido, Email: c.Email})
	}
	responderJSON(w, http.StatusOK, construirPagina(out, total, page, size))
}

func (s *Server) hClienteExists(w http.ResponseWriter, r *http.Request) {
	telefono := chiURLParam(r, "telefono")
	tid, _ := tenantID(r)
	existe, err := s.clientes.Exists(r.Context(), tid, telefono)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	responderJSON(w, http.StatusOK, existe)
}
