// Package estadisticas calcula el panel replicando la aritmética de fechas
// de Java (minusMonths/plusMonths/withDayOfMonth) y las queries del panel.
package estadisticas

import (
	"context"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/timeutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IngresoDiario es un bucket diario del panel.
type IngresoDiario struct {
	Fecha        time.Time
	IngresoTotal float64
}

// ServicioEstadistica es el conteo por servicio.
type ServicioEstadistica struct {
	ID                int64
	Nombre            string
	CantidadRealizado int64
}

// HorarioEstadistica es el conteo por hora del día.
type HorarioEstadistica struct {
	Hora              int32
	CantidadRealizado int64
}

// Panel es el DTO completo.
type Panel struct {
	TurnosHoy           int64
	TurnosAyer          int64
	IngresosMes         float64
	IngresosMesAnterior float64
	CantClientes        int64
	IngresosDiarios     []IngresoDiario
	Servicios           []ServicioEstadistica
	Horarios            []HorarioEstadistica
}

func ts(t time.Time) pgtype.Timestamp { return pgtype.Timestamp{Time: t, Valid: true} }

// Service calcula el panel.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
	now  func() time.Time
}

// New construye el servicio.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service {
	return &Service{pool: pool, q: q, now: time.Now}
}

// PanelCalcula replica PanelEstadisticaDTOService.obtenerPanelEstadisticaDTO.
func (s *Service) PanelCalcula(ctx context.Context, tenantID int64) (*Panel, error) {
	var panel *Panel
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		fechaActual := s.now()

		turnosHoy, err := q.CountTurnosDelDia(ctx, sqlc.CountTurnosDelDiaParams{TenantID: tenantID, Fecha: pgtype.Date{Time: fechaActual, Valid: true}})
		if err != nil {
			return err
		}
		turnosAyer, err := q.CountTurnosDelDia(ctx, sqlc.CountTurnosDelDiaParams{TenantID: tenantID, Fecha: pgtype.Date{Time: fechaActual.AddDate(0, 0, -1), Valid: true}})
		if err != nil {
			return err
		}

		inicioMes := time.Date(fechaActual.Year(), fechaActual.Month(), 1, 0, 0, 0, 0, fechaActual.Location())
		inicioMesSiguiente := timeutil.PlusMonths(inicioMes, 1)
		inicioMesAnterior := timeutil.MinusMonths(inicioMes, 1)

		ingresosMes, err := q.SumIngresosRango(ctx, sqlc.SumIngresosRangoParams{
			TenantID: tenantID, Inicio: ts(inicioMes), Fin: ts(inicioMesSiguiente),
		})
		if err != nil {
			return err
		}
		ingresosMesAnterior, err := q.SumIngresosRango(ctx, sqlc.SumIngresosRangoParams{
			TenantID: tenantID, Inicio: ts(inicioMesAnterior), Fin: ts(inicioMes),
		})
		if err != nil {
			return err
		}

		cantClientes, err := q.CountClientes(ctx, tenantID)
		if err != nil {
			return err
		}

		fechaLimite := timeutil.MinusMonths(fechaActual, 1)
		ingresosDiarios := make([]IngresoDiario, 0, 31)
		for i := 0; ; i++ {
			fecha := timeutil.StartOfDay(fechaActual.AddDate(0, 0, -i))
			if fecha.Before(fechaLimite) {
				break
			}
			total, err := q.SumIngresosRango(ctx, sqlc.SumIngresosRangoParams{
				TenantID: tenantID, Inicio: ts(fecha), Fin: ts(fecha.AddDate(0, 0, 1)),
			})
			if err != nil {
				return err
			}
			ingresosDiarios = append(ingresosDiarios, IngresoDiario{Fecha: fecha, IngresoTotal: total})
		}

		desdeServicios := timeutil.StartOfDay(timeutil.MinusMonths(fechaActual, 1))
		hastaServicios := timeutil.StartOfDay(fechaActual.AddDate(0, 0, 1))
		rowsServicios, err := q.CountServiciosRango(ctx, sqlc.CountServiciosRangoParams{
			TenantID: tenantID, Desde: ts(desdeServicios), Hasta: ts(hastaServicios),
		})
		if err != nil {
			return err
		}
		servicios := make([]ServicioEstadistica, 0, len(rowsServicios))
		for _, r := range rowsServicios {
			nombre := ""
			if r.Tipo != nil {
				nombre = *r.Tipo
			}
			servicios = append(servicios, ServicioEstadistica{ID: r.ID, Nombre: nombre, CantidadRealizado: r.Cantidad})
		}

		// Java: fechaActual.withYear(...).withMonth(1).withDayOfMonth(1) conserva
		// la hora actual del día 1 de enero.
		desdeHorarios := time.Date(fechaActual.Year(), time.January, 1,
			fechaActual.Hour(), fechaActual.Minute(), fechaActual.Second(), fechaActual.Nanosecond(), fechaActual.Location())
		rowsHorarios, err := q.CountTurnosPorHorario(ctx, sqlc.CountTurnosPorHorarioParams{
			TenantID: tenantID, Desde: ts(desdeHorarios),
		})
		if err != nil {
			return err
		}
		horarios := make([]HorarioEstadistica, 0, len(rowsHorarios))
		for _, r := range rowsHorarios {
			horarios = append(horarios, HorarioEstadistica{Hora: r.Hora, CantidadRealizado: r.Cantidad})
		}

		panel = &Panel{
			TurnosHoy: turnosHoy, TurnosAyer: turnosAyer,
			IngresosMes: ingresosMes, IngresosMesAnterior: ingresosMesAnterior,
			CantClientes: cantClientes, IngresosDiarios: ingresosDiarios,
			Servicios: servicios, Horarios: horarios,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return panel, nil
}
