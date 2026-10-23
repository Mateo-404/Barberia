package httpapi

import (
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/modules/turnos"
	"github.com/Mateo-404/Barberia/backend-go/internal/timeutil"
)

// turnoDTO replica TurnoResponseDTO (aplanado).
type turnoDTO struct {
	ID              int64   `json:"id"`
	FechaHora       string  `json:"fechaHora"`
	TelefonoCliente *string `json:"telefonoCliente"`
	NombreCliente   *string `json:"nombreCliente"`
	ApellidoCliente *string `json:"apellidoCliente"`
	EmailCliente    *string `json:"emailCliente"`
	IDServicio      *int64  `json:"idServicio"`
	TipoServicio    *string `json:"tipoServicio"`
	PrecioServicio  float32 `json:"precioServicio"`
}

// turnoEntidadDTO replica la serialización de la entidad Turno (paginación).
type turnoEntidadDTO struct {
	ID        int64               `json:"id"`
	FechaHora string              `json:"fechaHora"`
	Cliente   *clienteEntidadDTO  `json:"cliente"`
	Servicio  *servicioEntidadDTO `json:"servicio"`
}

type clienteEntidadDTO struct {
	Telefono *string `json:"telefono"`
	Nombre   *string `json:"nombre"`
	Apellido *string `json:"apellido"`
	Email    *string `json:"email"`
}

type servicioEntidadDTO struct {
	ID     *int64   `json:"id"`
	Tipo   *string  `json:"tipo"`
	Precio *float32 `json:"precio"`
}

type turnoCreateReq struct {
	FechaHora       *string `json:"fechaHora"`
	IDServicio      *int64  `json:"idServicio"`
	TelefonoCliente *string `json:"telefonoCliente"`
	NombreCliente   *string `json:"nombreCliente"`
	ApellidoCliente *string `json:"apellidoCliente"`
	EmailCliente    *string `json:"emailCliente"`
}

type turnoPatchReq struct {
	FechaHora *string           `json:"fechaHora"`
	Cliente   *clientePatchObj  `json:"cliente"`
	Servicio  *servicioPatchObj `json:"servicio"`
}

type clientePatchObj struct {
	Telefono *string `json:"telefono"`
	Nombre   *string `json:"nombre"`
	Apellido *string `json:"apellido"`
	Email    *string `json:"email"`
}

type servicioPatchObj struct {
	ID     *int64   `json:"id"`
	Tipo   *string  `json:"tipo"`
	Precio *float32 `json:"precio"`
}

func aTurnoDTO(t turnos.Turno) turnoDTO {
	d := turnoDTO{ID: t.ID, FechaHora: timeutil.FormatFechaHora(t.FechaHora)}
	if t.Cliente != nil {
		d.TelefonoCliente = t.Cliente.Telefono
		d.NombreCliente = t.Cliente.Nombre
		d.ApellidoCliente = t.Cliente.Apellido
		d.EmailCliente = t.Cliente.Email
	}
	if t.Servicio != nil {
		d.IDServicio = t.Servicio.ID
		d.TipoServicio = t.Servicio.Tipo
		if t.Servicio.Precio != nil {
			d.PrecioServicio = *t.Servicio.Precio
		}
	}
	return d
}

func aTurnoEntidadDTO(t turnos.Turno) turnoEntidadDTO {
	d := turnoEntidadDTO{ID: t.ID, FechaHora: timeutil.FormatFechaHora(t.FechaHora)}
	if t.Cliente != nil {
		d.Cliente = &clienteEntidadDTO{
			Telefono: t.Cliente.Telefono, Nombre: t.Cliente.Nombre,
			Apellido: t.Cliente.Apellido, Email: t.Cliente.Email,
		}
	}
	if t.Servicio != nil {
		d.Servicio = &servicioEntidadDTO{ID: t.Servicio.ID, Tipo: t.Servicio.Tipo, Precio: t.Servicio.Precio}
	}
	return d
}

func (s *Server) hTurnoList(w http.ResponseWriter, r *http.Request) {
	tid, _ := tenantID(r)
	lista, err := s.turnos.List(r.Context(), tid)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]turnoDTO, 0, len(lista))
	for _, t := range lista {
		out = append(out, aTurnoDTO(t))
	}
	responderJSON(w, http.StatusOK, out)
}

func (s *Server) hTurnoCreate(w http.ResponseWriter, r *http.Request) {
	in, ok := s.parsearTurnoCreate(w, r)
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	t, err := s.turnos.Create(r.Context(), tid, in)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	responderJSON(w, http.StatusCreated, aTurnoDTO(*t))
}

func (s *Server) hTurnoCreateAll(w http.ResponseWriter, r *http.Request) {
	var reqs []turnoCreateReq
	if !decodificarJSONConLimite(w, r, &reqs, s.cfg.MaxRequestBody) {
		return
	}
	inputs := make([]turnos.Input, 0, len(reqs))
	for _, req := range reqs {
		in, ok := s.validarTurnoCreate(w, r, req)
		if !ok {
			return
		}
		inputs = append(inputs, in)
	}
	tid, _ := tenantID(r)
	lista, err := s.turnos.CreateAll(r.Context(), tid, inputs)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]turnoDTO, 0, len(lista))
	for _, t := range lista {
		out = append(out, aTurnoDTO(t))
	}
	responderJSON(w, http.StatusCreated, out)
}

func (s *Server) parsearTurnoCreate(w http.ResponseWriter, r *http.Request) (turnos.Input, bool) {
	var req turnoCreateReq
	if !decodificarJSONConLimite(w, r, &req, s.cfg.MaxRequestBody) {
		return turnos.Input{}, false
	}
	return s.validarTurnoCreate(w, r, req)
}

func (s *Server) validarTurnoCreate(w http.ResponseWriter, r *http.Request, req turnoCreateReq) (turnos.Input, bool) {
	var in turnos.Input

	if req.FechaHora == nil {
		badRequestValidacion(w, r)
		return in, false
	}
	fecha, err := timeutil.ParseLocalDateTime(*req.FechaHora)
	if err != nil {
		badRequestValidacion(w, r)
		return in, false
	}
	// @Future
	if !fecha.After(s.now()) {
		badRequestValidacion(w, r)
		return in, false
	}
	in.FechaHora = fecha

	if req.IDServicio == nil {
		badRequestValidacion(w, r)
		return in, false
	}
	in.IDServicio = *req.IDServicio

	if !validarNotBlank(w, r, req.TelefonoCliente) || !validarNotBlank(w, r, req.NombreCliente) ||
		!validarNotBlank(w, r, req.ApellidoCliente) || !validarEmail(w, r, req.EmailCliente) ||
		!validarMaxLength(w, r, req.TelefonoCliente, maxTelefono) ||
		!validarMaxLength(w, r, req.NombreCliente, maxNombre) ||
		!validarMaxLength(w, r, req.ApellidoCliente, maxNombre) ||
		!validarMaxLength(w, r, req.EmailCliente, maxEmail) {
		return in, false
	}
	in.TelefonoCliente, in.NombreCliente, in.ApellidoCliente = *req.TelefonoCliente, *req.NombreCliente, *req.ApellidoCliente
	in.EmailCliente = req.EmailCliente
	return in, true
}

func (s *Server) hTurnoGet(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	t, err := s.turnos.Get(r.Context(), tid, id)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	if t == nil {
		notFoundVacio(w)
		return
	}
	responderJSON(w, http.StatusOK, aTurnoDTO(*t))
}

func (s *Server) hTurnoPatch(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	var req turnoPatchReq
	if !decodificarJSONConLimite(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	p := turnos.Patch{}
	if req.FechaHora != nil {
		fecha, err := timeutil.ParseLocalDateTime(*req.FechaHora)
		if err != nil {
			badRequestValidacion(w, r)
			return
		}
		p.FechaHora = &fecha
	}
	if req.Cliente != nil {
		// Validación de los campos presentes (PATCH parcial sin reglas en Java).
		if !validarOpcionalEmail(w, r, req.Cliente.Email) ||
			!validarMaxLength(w, r, req.Cliente.Telefono, maxTelefono) ||
			!validarMaxLength(w, r, req.Cliente.Nombre, maxNombre) ||
			!validarMaxLength(w, r, req.Cliente.Apellido, maxNombre) ||
			!validarMaxLength(w, r, req.Cliente.Email, maxEmail) {
			return
		}
		p.ClienteDado = true
		p.Cliente = &turnos.Cliente{Telefono: req.Cliente.Telefono, Nombre: req.Cliente.Nombre, Apellido: req.Cliente.Apellido, Email: req.Cliente.Email}
	}
	if req.Servicio != nil {
		if !validarMaxLength(w, r, req.Servicio.Tipo, maxTipoServicio) ||
			!validarOpcionalPositivo(w, r, req.Servicio.Precio) {
			return
		}
		p.ServicioDado = true
		p.Servicio = &turnos.Servicio{ID: req.Servicio.ID, Tipo: req.Servicio.Tipo, Precio: req.Servicio.Precio}
	}
	tid, _ := tenantID(r)
	t, err := s.turnos.Update(r.Context(), tid, id, p)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	if t == nil {
		notFoundVacio(w)
		return
	}
	responderJSON(w, http.StatusOK, aTurnoDTO(*t))
}

func (s *Server) hTurnoDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	borrado, err := s.turnos.Delete(r.Context(), tid, id)
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

func (s *Server) hTurnoPage(w http.ResponseWriter, r *http.Request) {
	page, size, ok := s.paramPagina(w, r)
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	lista, total, err := s.turnos.Page(r.Context(), tid, page, size)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := make([]turnoEntidadDTO, 0, len(lista))
	for _, t := range lista {
		out = append(out, aTurnoEntidadDTO(t))
	}
	responderJSON(w, http.StatusOK, construirPagina(out, total, page, size))
}

func (s *Server) hTurnoExists(w http.ResponseWriter, r *http.Request) {
	id, ok := s.paramID(w, r, "id")
	if !ok {
		return
	}
	tid, _ := tenantID(r)
	existe, err := s.turnos.Exists(r.Context(), tid, id)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	responderJSON(w, http.StatusOK, existe)
}

func (s *Server) hTurnoFindDateTimes(w http.ResponseWriter, r *http.Request) {
	tid, _ := tenantID(r)
	fechas, err := s.turnos.FindDateTimes(r.Context(), tid)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	responderJSON(w, http.StatusOK, fechas)
}
