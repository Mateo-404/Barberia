package httpapi

import (
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/timeutil"
)

// panelDTO replica PanelEstadisticaDTO.
type panelDTO struct {
	TurnosHoy           int64              `json:"turnosHoy"`
	TurnosAyer          int64              `json:"turnosAyer"`
	IngresosMes         float64            `json:"ingresosMes"`
	IngresosMesAnterior float64            `json:"ingresosMesAnterior"`
	CantClientes        int64              `json:"cantClientes"`
	IngresosDiarios     []ingresoDiarioDTO `json:"ingresosDiarios"`
	Servicios           []servicioEstDTO   `json:"servicios"`
	Horarios            []horarioEstDTO    `json:"horarios"`
}

type ingresoDiarioDTO struct {
	Fecha        string  `json:"fecha"`
	IngresoTotal float64 `json:"ingresoTotal"`
}

type servicioEstDTO struct {
	ID                int64  `json:"id"`
	Nombre            string `json:"nombre"`
	CantidadRealizado int64  `json:"cantidadRealizado"`
}

type horarioEstDTO struct {
	Hora              int32 `json:"hora"`
	CantidadRealizado int64 `json:"cantidadRealizado"`
}

func (s *Server) hEstadisticasPanel(w http.ResponseWriter, r *http.Request) {
	tid, _ := tenantID(r)
	p, err := s.stats.PanelCalcula(r.Context(), tid)
	if err != nil {
		s.escribirError(w, r, err)
		return
	}
	out := panelDTO{
		TurnosHoy: p.TurnosHoy, TurnosAyer: p.TurnosAyer,
		IngresosMes: p.IngresosMes, IngresosMesAnterior: p.IngresosMesAnterior,
		CantClientes:    p.CantClientes,
		IngresosDiarios: make([]ingresoDiarioDTO, 0, len(p.IngresosDiarios)),
		Servicios:       make([]servicioEstDTO, 0, len(p.Servicios)),
		Horarios:        make([]horarioEstDTO, 0, len(p.Horarios)),
	}
	for _, d := range p.IngresosDiarios {
		out.IngresosDiarios = append(out.IngresosDiarios, ingresoDiarioDTO{
			Fecha: timeutil.FormatFechaHora(d.Fecha), IngresoTotal: d.IngresoTotal,
		})
	}
	for _, sv := range p.Servicios {
		out.Servicios = append(out.Servicios, servicioEstDTO{
			ID: sv.ID, Nombre: sv.Nombre, CantidadRealizado: sv.CantidadRealizado,
		})
	}
	for _, h := range p.Horarios {
		out.Horarios = append(out.Horarios, horarioEstDTO{Hora: h.Hora, CantidadRealizado: h.CantidadRealizado})
	}
	responderJSON(w, http.StatusOK, out)
}
