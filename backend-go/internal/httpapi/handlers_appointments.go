package httpapi

import (
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/modules/appointments"
	"github.com/Mateo-404/Barberia/backend-go/internal/timeutil"
)

// appointmentDTO is the flat appointment shape returned by the API.
type appointmentDTO struct {
	ID                int64   `json:"id"`
	StartsAt          string  `json:"startsAt"`
	CustomerPhone     *string `json:"customerPhone"`
	CustomerFirstName *string `json:"customerFirstName"`
	CustomerLastName  *string `json:"customerLastName"`
	CustomerEmail     *string `json:"customerEmail"`
	ServiceID         *int64  `json:"serviceId"`
	ServiceName       *string `json:"serviceName"`
	ServicePrice      float32 `json:"servicePrice"`
}

type appointmentCreateReq struct {
	StartsAt          *string `json:"startsAt"`
	ServiceID         *int64  `json:"serviceId"`
	CustomerPhone     *string `json:"customerPhone"`
	CustomerFirstName *string `json:"customerFirstName"`
	CustomerLastName  *string `json:"customerLastName"`
	CustomerEmail     *string `json:"customerEmail"`
}

// appointmentPatchReq: a present "customer"/"service" object replaces the
// association; null clears it.
type appointmentPatchReq struct {
	StartsAt *string          `json:"startsAt"`
	Customer *customerPatchIn `json:"customer"`
	Service  *servicePatchIn  `json:"service"`
}

type customerPatchIn struct {
	Phone     *string `json:"phone"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email"`
}

type servicePatchIn struct {
	ID    *int64   `json:"id"`
	Name  *string  `json:"name"`
	Price *float32 `json:"price"`
}

func toAppointmentDTO(a appointments.Appointment) appointmentDTO {
	d := appointmentDTO{ID: a.ID, StartsAt: timeutil.FormatDateTime(a.StartsAt)}
	if a.Customer != nil {
		d.CustomerPhone = a.Customer.Phone
		d.CustomerFirstName = a.Customer.FirstName
		d.CustomerLastName = a.Customer.LastName
		d.CustomerEmail = a.Customer.Email
	}
	if a.Service != nil {
		d.ServiceID = a.Service.ID
		d.ServiceName = a.Service.Name
		if a.Service.Price != nil {
			d.ServicePrice = *a.Service.Price
		}
	}
	return d
}

func (s *Server) handleAppointmentList(w http.ResponseWriter, r *http.Request) {
	list, err := s.appointments.List(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(list, toAppointmentDTO))
}

func (s *Server) handleAppointmentPage(w http.ResponseWriter, r *http.Request) {
	number, size, ok := pageParams(w, r)
	if !ok {
		return
	}
	list, total, err := s.appointments.Page(r.Context(), tenantID(r), number, size)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPage(mapSlice(list, toAppointmentDTO), total, number, size))
}

func (s *Server) handleAppointmentCreate(w http.ResponseWriter, r *http.Request) {
	var req appointmentCreateReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	var v validator
	var in appointments.Input
	if v.required("startsAt", req.StartsAt) {
		at, err := timeutil.ParseLocalDateTime(*req.StartsAt)
		switch {
		case err != nil:
			v.add("startsAt", codeInvalidDate)
		case !at.After(s.now()):
			v.add("startsAt", codeMustBeFuture)
		default:
			in.StartsAt = at
		}
	}
	if req.ServiceID == nil {
		v.add("serviceId", codeRequired)
	} else {
		in.ServiceID = *req.ServiceID
	}
	v.required("customerPhone", req.CustomerPhone)
	v.required("customerFirstName", req.CustomerFirstName)
	v.required("customerLastName", req.CustomerLastName)
	v.maxLen("customerPhone", req.CustomerPhone, maxPhone)
	v.maxLen("customerFirstName", req.CustomerFirstName, maxName)
	v.maxLen("customerLastName", req.CustomerLastName, maxName)
	v.email("customerEmail", req.CustomerEmail)
	v.maxLen("customerEmail", req.CustomerEmail, maxEmail)
	if !v.ok(w, r) {
		return
	}
	in.CustomerPhone, in.CustomerFirstName, in.CustomerLastName = *req.CustomerPhone, *req.CustomerFirstName, *req.CustomerLastName
	in.CustomerEmail = req.CustomerEmail

	a, err := s.appointments.Create(r.Context(), tenantID(r), in)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAppointmentDTO(*a))
}

func (s *Server) handleAppointmentGet(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	a, err := s.appointments.Get(r.Context(), tenantID(r), id)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if a == nil {
		resourceNotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, toAppointmentDTO(*a))
}

func (s *Server) handleAppointmentPatch(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var req appointmentPatchReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	var v validator
	p := appointments.Patch{}
	if req.StartsAt != nil {
		at, err := timeutil.ParseLocalDateTime(*req.StartsAt)
		if err != nil {
			v.add("startsAt", codeInvalidDate)
		}
		p.StartsAt = &at
	}
	if c := req.Customer; c != nil {
		v.maxLen("customer.phone", c.Phone, maxPhone)
		v.maxLen("customer.firstName", c.FirstName, maxName)
		v.maxLen("customer.lastName", c.LastName, maxName)
		v.email("customer.email", c.Email)
		v.maxLen("customer.email", c.Email, maxEmail)
		p.CustomerSent = true
		p.Customer = &appointments.Customer{Phone: c.Phone, FirstName: c.FirstName, LastName: c.LastName, Email: c.Email}
	}
	if sv := req.Service; sv != nil {
		v.maxLen("service.name", sv.Name, maxServiceName)
		v.positive("service.price", sv.Price)
		p.ServiceSent = true
		p.Service = &appointments.Offering{ID: sv.ID, Name: sv.Name, Price: sv.Price}
	}
	if !v.ok(w, r) {
		return
	}
	a, err := s.appointments.Update(r.Context(), tenantID(r), id, p)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if a == nil {
		resourceNotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, toAppointmentDTO(*a))
}

func (s *Server) handleAppointmentDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	deleted, err := s.appointments.Delete(r.Context(), tenantID(r), id)
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

func (s *Server) handleBookedSlots(w http.ResponseWriter, r *http.Request) {
	slots, err := s.appointments.BookedSlots(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, slots)
}
