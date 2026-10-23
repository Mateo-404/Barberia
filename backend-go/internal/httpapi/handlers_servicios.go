package httpapi

import (
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/modules/servicios"
)

// servicioDTO replica ServicioResponseDTO.
type servicioDTO struct {
	ID     int64   `json:"id"`
	Tipo   string  `json:"tipo"`
	Precio float32 `json:"precio"`
}

type servicioCreateReq struct {
	Tipo   *string  `json:"tipo"`
	Precio *float32 `json:"precio"`
}

type servicioPatchReq struct {
	ID     *int64   `json:"id"`
	Tipo   *string  `json:"tipo"`
	Precio *float32 `json:"precio"`
}

func (s *Server) hServicioList(w http.ResponseWriter, r *http.Request) {
	tid, _ := tenantID(r)
	lista, err := s.servicios.List(r.Context(), tid)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]servicioDTO, 0, len(lista))
	for _, v := range lista {
		out = append(out, servicioDTO{ID: v.ID, Tipo: v.Tipo, Precio: v.Precio})
	}
	responderJSON(w, http.StatusOK, out)
}

func (s *Server) hServicioCreate(w http.ResponseWriter, r *http.Request) {
	var req servicioCreateReq
	if !decodificarJSONConLimite(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	if !validarNotBlank(w, r, req.Tipo) || !validarMaxLength(w, r, req.Tipo, maxTipoServicio) {
		return
	}
	// @Positive: precio obligatorio y > 0.
	if req.Precio == nil || *req.Precio <= 0 {
		badRequestValidacion(w, r)
		return
	}
	tid, _ := tenantID(r)
	v, err := s.servicios.Create(r.Context(), tid, servicios.Input{Tipo: *req.Tipo, Precio: *req.Precio})
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	responderJSON(w, http.StatusCreated, servicioDTO{ID: v.ID, Tipo: v.Tipo, Precio: v.Precio})
}

func (s *Server) hServicioCreateAll(w http.ResponseWriter, r *http.Request) {
	var reqs []servicioCreateReq
	if !decodificarJSONConLimite(w, r, &reqs, s.cfg.MaxRequestBody) {
		return
	}
	inputs := make([]servicios.Input, 0, len(reqs))
	for _, req := range reqs {
		if !validarNotBlank(w, r, req.Tipo) || !validarMaxLength(w, r, req.Tipo, maxTipoServicio) {
			return
		}
		if req.Precio == nil || *req.Precio <= 0 {
			badRequestValidacion(w, r)
			return
		}
		inputs = append(inputs, servicios.Input{Tipo: *req.Tipo, Precio: *req.Precio})
	}
	tid, _ := tenantID(r)
	lista, err := s.servicios.CreateAll(r.Context(), tid, inputs)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]servicioDTO, 0, len(lista))
	for _, v := range lista {
		out = append(out, servicioDTO{ID: v.ID, Tipo: v.Tipo, Precio: v.Precio})
	}
	responderJSON(w, http.StatusCreated, out)
}

func (s *Server) hServicioGet(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	v, err := s.servicios.Get(r.Context(), tid, id)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	if v == nil {
		notFoundVacio(w)
		return
	}
	responderJSON(w, http.StatusOK, servicioDTO{ID: v.ID, Tipo: v.Tipo, Precio: v.Precio})
}

func (s *Server) hServicioPatch(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	var req servicioPatchReq
	if !decodificarJSONConLimite(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	// Validación de los campos presentes (PATCH parcial sin reglas en Java).
	if !validarMaxLength(w, r, req.Tipo, maxTipoServicio) ||
		!validarOpcionalPositivo(w, r, req.Precio) {
		return
	}
	tid, _ := tenantID(r)
	v, err := s.servicios.Update(r.Context(), tid, id, servicios.Patch{Tipo: req.Tipo, Precio: req.Precio})
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	if v == nil {
		notFoundVacio(w)
		return
	}
	responderJSON(w, http.StatusOK, servicioDTO{ID: v.ID, Tipo: v.Tipo, Precio: v.Precio})
}

func (s *Server) hServicioDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	borrado, err := s.servicios.Delete(r.Context(), tid, id)
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

func (s *Server) hServicioPage(w http.ResponseWriter, r *http.Request) {
	page, size, ok := s.paramPagina(w, r)
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	lista, total, err := s.servicios.Page(r.Context(), tid, page, size)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]servicioDTO, 0, len(lista))
	for _, v := range lista {
		out = append(out, servicioDTO{ID: v.ID, Tipo: v.Tipo, Precio: v.Precio})
	}
	responderJSON(w, http.StatusOK, construirPagina(out, total, page, size))
}

func (s *Server) hServicioExists(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	existe, err := s.servicios.Exists(r.Context(), tid, id)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	responderJSON(w, http.StatusOK, existe)
}
