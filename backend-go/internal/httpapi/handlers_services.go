package httpapi

import (
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/modules/services"
)

type serviceDTO struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float32 `json:"price"`
}

type serviceReq struct {
	Name  *string  `json:"name"`
	Price *float32 `json:"price"`
}

func toServiceDTO(o services.Offering) serviceDTO {
	return serviceDTO{ID: o.ID, Name: o.Name, Price: o.Price}
}

func (s *Server) handleServiceList(w http.ResponseWriter, r *http.Request) {
	list, err := s.services.List(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(list, toServiceDTO))
}

func (s *Server) handleServicePage(w http.ResponseWriter, r *http.Request) {
	number, size, ok := pageParams(w, r)
	if !ok {
		return
	}
	list, total, err := s.services.Page(r.Context(), tenantID(r), number, size)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPage(mapSlice(list, toServiceDTO), total, number, size))
}

func (s *Server) handleServiceCreate(w http.ResponseWriter, r *http.Request) {
	var req serviceReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	var v validator
	v.required("name", req.Name)
	v.maxLen("name", req.Name, maxServiceName)
	if req.Price == nil {
		v.add("price", codeRequired)
	}
	v.positive("price", req.Price)
	if !v.ok(w, r) {
		return
	}
	o, err := s.services.Create(r.Context(), tenantID(r), services.Input{Name: *req.Name, Price: *req.Price})
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toServiceDTO(*o))
}

func (s *Server) handleServiceGet(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	o, err := s.services.Get(r.Context(), tenantID(r), id)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if o == nil {
		resourceNotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, toServiceDTO(*o))
}

func (s *Server) handleServicePatch(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var req serviceReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	var v validator
	if req.Name != nil {
		v.required("name", req.Name)
	}
	v.maxLen("name", req.Name, maxServiceName)
	v.positive("price", req.Price)
	if !v.ok(w, r) {
		return
	}
	o, err := s.services.Update(r.Context(), tenantID(r), id, services.Patch{Name: req.Name, Price: req.Price})
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if o == nil {
		resourceNotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, toServiceDTO(*o))
}

func (s *Server) handleServiceDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	deleted, err := s.services.Delete(r.Context(), tenantID(r), id)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if !deleted {
		resourceNotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
