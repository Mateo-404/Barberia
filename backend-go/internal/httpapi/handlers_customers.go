package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Mateo-404/Barberia/backend-go/internal/modules/customers"
)

type customerDTO struct {
	Phone     string  `json:"phone"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     *string `json:"email"`
}

type customerCreateReq struct {
	Phone     *string `json:"phone"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email"`
}

type customerPatchReq struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email"`
}

func toCustomerDTO(c customers.Customer) customerDTO {
	return customerDTO{Phone: c.Phone, FirstName: c.FirstName, LastName: c.LastName, Email: c.Email}
}

func (s *Server) handleCustomerList(w http.ResponseWriter, r *http.Request) {
	list, err := s.customers.List(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(list, toCustomerDTO))
}

func (s *Server) handleCustomerPage(w http.ResponseWriter, r *http.Request) {
	number, size, ok := pageParams(w, r)
	if !ok {
		return
	}
	list, total, err := s.customers.Page(r.Context(), tenantID(r), number, size)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPage(mapSlice(list, toCustomerDTO), total, number, size))
}

func (s *Server) handleCustomerCreate(w http.ResponseWriter, r *http.Request) {
	var req customerCreateReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	var v validator
	v.required("phone", req.Phone)
	v.required("firstName", req.FirstName)
	v.required("lastName", req.LastName)
	v.maxLen("phone", req.Phone, maxPhone)
	v.maxLen("firstName", req.FirstName, maxName)
	v.maxLen("lastName", req.LastName, maxName)
	v.email("email", req.Email)
	v.maxLen("email", req.Email, maxEmail)
	if !v.ok(w, r) {
		return
	}
	c, err := s.customers.Create(r.Context(), tenantID(r), customers.Input{
		Phone: *req.Phone, FirstName: *req.FirstName, LastName: *req.LastName, Email: req.Email,
	})
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toCustomerDTO(*c))
}

func (s *Server) handleCustomerGet(w http.ResponseWriter, r *http.Request) {
	c, err := s.customers.Get(r.Context(), tenantID(r), chi.URLParam(r, "phone"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if c == nil {
		resourceNotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, toCustomerDTO(*c))
}

func (s *Server) handleCustomerPatch(w http.ResponseWriter, r *http.Request) {
	var req customerPatchReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	var v validator
	v.maxLen("firstName", req.FirstName, maxName)
	v.maxLen("lastName", req.LastName, maxName)
	v.email("email", req.Email)
	v.maxLen("email", req.Email, maxEmail)
	if !v.ok(w, r) {
		return
	}
	c, err := s.customers.Update(r.Context(), tenantID(r), chi.URLParam(r, "phone"), customers.Patch{
		FirstName: req.FirstName, LastName: req.LastName, Email: req.Email,
	})
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if c == nil {
		resourceNotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, toCustomerDTO(*c))
}

func (s *Server) handleCustomerDelete(w http.ResponseWriter, r *http.Request) {
	deleted, err := s.customers.Delete(r.Context(), tenantID(r), chi.URLParam(r, "phone"))
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
