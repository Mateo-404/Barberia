package httpapi

import (
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/modules/administradores"
)

// adminDTO replica AdministradorResponseDTO (sin hash, ADR-004).
type adminDTO struct {
	ID       int64  `json:"id"`
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Email    string `json:"email"`
}

// loginDTO replica LoginResponseDTO (los 5 campos siempre emitidos).
type loginDTO struct {
	ID       int64  `json:"id"`
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Email    string `json:"email"`
	Token    string `json:"token"`
	// Rol del administrador: owner | admin | staff (SPEC §8).
	Rol string `json:"rol,omitempty"`
}

type adminCreateReq struct {
	Nombre      *string `json:"nombre"`
	Apellido    *string `json:"apellido"`
	Email       *string `json:"email"`
	Contrasenia *string `json:"contrasenia"`
}

type adminPatchReq struct {
	Nombre      *string `json:"nombre"`
	Apellido    *string `json:"apellido"`
	Email       *string `json:"email"`
	Contrasenia *string `json:"contrasenia"`
}

type loginReq struct {
	Email       *string `json:"email"`
	Contrasenia *string `json:"contrasenia"`
}

func (s *Server) hAdminList(w http.ResponseWriter, r *http.Request) {
	tid, _ := tenantID(r)
	lista, err := s.admins.List(r.Context(), tid)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]adminDTO, 0, len(lista))
	for _, a := range lista {
		out = append(out, adminDTO{ID: a.ID, Nombre: a.Nombre, Apellido: a.Apellido, Email: a.Email})
	}
	responderJSON(w, http.StatusOK, out)
}

func (s *Server) hAdminCreate(w http.ResponseWriter, r *http.Request) {
	var req adminCreateReq
	if !decodificarJSONConLimite(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	if !validarNotBlank(w, r, req.Nombre) || !validarNotBlank(w, r, req.Apellido) ||
		!validarNotBlank(w, r, req.Email) || !validarEmail(w, r, req.Email) ||
		!validarMaxLength(w, r, req.Email, maxEmail) ||
		!validarMinLength(w, r, req.Contrasenia, 8) {
		return
	}
	tid, _ := tenantID(r)
	a, err := s.admins.Create(r.Context(), tid, administradores.Input{
		Nombre: *req.Nombre, Apellido: *req.Apellido, Email: *req.Email, Contrasenia: *req.Contrasenia,
	})
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	responderJSON(w, http.StatusCreated, adminDTO{ID: a.ID, Nombre: a.Nombre, Apellido: a.Apellido, Email: a.Email})
}

func (s *Server) hAdminCreateAll(w http.ResponseWriter, r *http.Request) {
	var reqs []adminCreateReq
	if !decodificarJSONConLimite(w, r, &reqs, s.cfg.MaxRequestBody) {
		return
	}
	inputs := make([]administradores.Input, 0, len(reqs))
	for _, req := range reqs {
		if !validarNotBlank(w, r, req.Nombre) || !validarNotBlank(w, r, req.Apellido) ||
			!validarNotBlank(w, r, req.Email) || !validarEmail(w, r, req.Email) ||
			!validarMaxLength(w, r, req.Email, maxEmail) ||
			!validarMinLength(w, r, req.Contrasenia, 8) {
			return
		}
		inputs = append(inputs, administradores.Input{
			Nombre: *req.Nombre, Apellido: *req.Apellido, Email: *req.Email, Contrasenia: *req.Contrasenia,
		})
	}
	tid, _ := tenantID(r)
	lista, err := s.admins.CreateAll(r.Context(), tid, inputs)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]adminDTO, 0, len(lista))
	for _, a := range lista {
		out = append(out, adminDTO{ID: a.ID, Nombre: a.Nombre, Apellido: a.Apellido, Email: a.Email})
	}
	responderJSON(w, http.StatusCreated, out)
}

func (s *Server) hAdminLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if !decodificarJSONConLimite(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	if !validarNotBlank(w, r, req.Email) || !validarNotBlank(w, r, req.Contrasenia) ||
		!validarMinLength(w, r, req.Contrasenia, 8) {
		return
	}

	ip := s.clienteIP(r)
	if !s.limiter.Allow(ip) {
		rateLimited(w, r)
		return
	}

	tid, _ := tenantID(r)
	admin, token, err := s.admins.Login(r.Context(), tid, *req.Email, *req.Contrasenia)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	s.limiter.Reset(ip)
	responderJSON(w, http.StatusOK, loginDTO{
		ID: admin.ID, Nombre: admin.Nombre, Apellido: admin.Apellido, Email: admin.Email, Token: token,
		Rol: admin.Rol,
	})
}

func (s *Server) hAdminGet(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	a, err := s.admins.Get(r.Context(), tid, id)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	if a == nil {
		notFoundVacio(w)
		return
	}
	responderJSON(w, http.StatusOK, adminDTO{ID: a.ID, Nombre: a.Nombre, Apellido: a.Apellido, Email: a.Email})
}

func (s *Server) hAdminPatch(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	var req adminPatchReq
	if !decodificarJSONConLimite(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	// Validación de presencia: coherente con el POST (Java no valida el PATCH,
	// pero dejar pasar contraseñas de 1 carácter o emails inválidos era un
	// agujero evidente).
	if !validarOpcionalEmail(w, r, req.Email) ||
		!validarMaxLength(w, r, req.Email, maxEmail) ||
		!validarOpcionalMinLength(w, r, req.Contrasenia, 8) {
		return
	}
	tid, _ := tenantID(r)
	a, err := s.admins.Update(r.Context(), tid, id, administradores.Patch{
		Nombre: req.Nombre, Apellido: req.Apellido, Email: req.Email, Contrasenia: req.Contrasenia,
	})
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	if a == nil {
		notFoundVacio(w)
		return
	}
	responderJSON(w, http.StatusOK, adminDTO{ID: a.ID, Nombre: a.Nombre, Apellido: a.Apellido, Email: a.Email})
}

func (s *Server) hAdminDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	borrado, err := s.admins.Delete(r.Context(), tid, id)
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

func (s *Server) hAdminPage(w http.ResponseWriter, r *http.Request) {
	page, size, ok := s.paramPagina(w, r)
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	lista, total, err := s.admins.Page(r.Context(), tid, page, size)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]adminDTO, 0, len(lista))
	for _, a := range lista {
		out = append(out, adminDTO{ID: a.ID, Nombre: a.Nombre, Apellido: a.Apellido, Email: a.Email})
	}
	responderJSON(w, http.StatusOK, construirPagina(out, total, page, size))
}

func (s *Server) hAdminExists(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	existe, err := s.admins.Exists(r.Context(), tid, id)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	responderJSON(w, http.StatusOK, existe)
}
