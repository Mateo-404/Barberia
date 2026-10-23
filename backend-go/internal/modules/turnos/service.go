// Package turnos implementa reservas con reglas de negocio y protección
// atómica contra doble reserva (UNIQUE por tenant + fecha_hora).
package turnos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
	"github.com/Mateo-404/Barberia/backend-go/internal/pgutil"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/timeutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Cliente es la asociación (puede ser parcial en PATCH).
type Cliente struct {
	Telefono *string
	Nombre   *string
	Apellido *string
	Email    *string
}

// Servicio es la asociación (puede ser parcial en PATCH).
type Servicio struct {
	ID     *int64
	Tipo   *string
	Precio *float32
}

// Turno es el modelo de dominio.
type Turno struct {
	ID        int64
	FechaHora time.Time
	Cliente   *Cliente
	Servicio  *Servicio
}

// Input son los datos de creación (TurnoRequestDTO).
type Input struct {
	FechaHora       time.Time
	IDServicio      int64
	TelefonoCliente string
	NombreCliente   string
	ApellidoCliente string
	EmailCliente    *string
}

// Patch son los campos presentes en el body del PATCH (sin validación de
// reglas, igual que Java).
type Patch struct {
	FechaHora    *time.Time
	Cliente      *Cliente
	Servicio     *Servicio
	ClienteDado  bool
	ServicioDado bool
}

// Service agrupa las operaciones de turnos.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
	now  func() time.Time
}

// New construye el servicio.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service {
	return &Service{pool: pool, q: q, now: time.Now}
}

func aDominio(r sqlc.ListTurnosRow) Turno {
	t := Turno{ID: r.ID, FechaHora: r.FechaHora.Time}
	if r.Telefono != nil || r.Nombre != nil || r.Apellido != nil || r.Email != nil {
		t.Cliente = &Cliente{Telefono: r.Telefono, Nombre: r.Nombre, Apellido: r.Apellido, Email: r.Email}
	}
	if r.ServicioID != nil || r.ServicioTipo != nil || r.ServicioPrecio != nil {
		t.Servicio = &Servicio{ID: r.ServicioID, Tipo: r.ServicioTipo, Precio: r.ServicioPrecio}
	}
	return t
}

func aDominioGet(r sqlc.GetTurnoRow) Turno {
	return aDominio(sqlc.ListTurnosRow(r))
}

// List devuelve los turnos del tenant ordenados por id.
func (s *Service) List(ctx context.Context, tenantID int64) ([]Turno, error) {
	var out []Turno
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListTurnos(ctx, tenantID)
		if err != nil {
			return err
		}
		out = make([]Turno, 0, len(rows))
		for _, r := range rows {
			out = append(out, aDominio(r))
		}
		return nil
	})
	return out, err
}

// Get devuelve un turno o nil.
func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Turno, error) {
	var out *Turno
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetTurno(ctx, sqlc.GetTurnoParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			out = nil
			return nil
		}
		if err != nil {
			return err
		}
		t := aDominioGet(r)
		out = &t
		return nil
	})
	return out, err
}

// Page devuelve una página 0-based (entidades anidadas).
func (s *Service) Page(ctx context.Context, tenantID int64, page, size int) ([]Turno, int64, error) {
	var out []Turno
	var total int64
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		t, err := q.CountTurnos(ctx, tenantID)
		if err != nil {
			return err
		}
		total = t
		rows, err := q.PageTurnos(ctx, sqlc.PageTurnosParams{
			TenantID: tenantID, Limite: int32(size), Offset: int32(page * size),
		})
		if err != nil {
			return err
		}
		out = make([]Turno, 0, len(rows))
		for _, r := range rows {
			out = append(out, aDominio(sqlc.ListTurnosRow(r)))
		}
		return nil
	})
	return out, total, err
}

// Create aplica las reglas de Java y protege la doble reserva.
func (s *Service) Create(ctx context.Context, tenantID int64, in Input) (*Turno, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantIDStr(tenantID)); err != nil {
		return nil, err
	}
	t, err := s.crearEnTx(ctx, s.q.WithTx(tx), tx, tenantID, in)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return t, nil
}

// CreateAll crea todos los turnos en una transacción (ADR-005).
func (s *Service) CreateAll(ctx context.Context, tenantID int64, ins []Input) ([]Turno, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantIDStr(tenantID)); err != nil {
		return nil, err
	}
	qtx := s.q.WithTx(tx)
	out := make([]Turno, 0, len(ins))
	for _, in := range ins {
		t, err := s.crearEnTx(ctx, qtx, tx, tenantID, in)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) crearEnTx(ctx context.Context, q *sqlc.Queries, tx pgx.Tx, tenantID int64, in Input) (*Turno, error) {
	servicio, err := q.GetServicio(ctx, sqlc.GetServicioParams{TenantID: tenantID, ID: in.IDServicio})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.New(apperr.KindNotFound, "No existe un servicio con el id %d", in.IDServicio)
	}
	if err != nil {
		return nil, err
	}
	if err := validarFechaHora(in.FechaHora); err != nil {
		return nil, err
	}
	duplicado, err := s.existeDuplicado(ctx, q, tenantID, in.FechaHora)
	if err != nil {
		return nil, err
	}
	if duplicado {
		return nil, turnoDuplicado(in.FechaHora)
	}

	cliente, err := s.obtenerOCrearCliente(ctx, q, tx, tenantID, in)
	if err != nil {
		return nil, err
	}

	creado, err := q.CreateTurno(ctx, sqlc.CreateTurnoParams{
		TenantID: tenantID, FechaHora: ts(in.FechaHora), ClienteID: &in.TelefonoCliente, ServicioID: &in.IDServicio,
	})
	if pgutil.IsUniqueViolation(err) {
		return nil, turnoDuplicado(in.FechaHora)
	}
	if err != nil {
		return nil, err
	}

	t := Turno{ID: creado.ID, FechaHora: creado.FechaHora.Time}
	t.Cliente = &Cliente{Telefono: &cliente.Telefono, Nombre: cliente.Nombre, Apellido: cliente.Apellido, Email: cliente.Email}
	t.Servicio = &Servicio{ID: &servicio.ID, Tipo: servicio.Tipo, Precio: servicio.Precio}
	return &t, nil
}

// obtenerOCrearCliente replica el find-or-create de TurnoService.save con
// savepoints: una carrera por el mismo teléfono no aborta la transacción
// (23505 → rollback al savepoint y relectura del cliente ya creado).
func (s *Service) obtenerOCrearCliente(ctx context.Context, q *sqlc.Queries, tx pgx.Tx, tenantID int64, in Input) (sqlc.GetClienteRow, error) {
	const sp = "sp_cliente"
	params := sqlc.GetClienteParams{TenantID: tenantID, Telefono: in.TelefonoCliente}
	cliente, err := q.GetCliente(ctx, params)
	if !errors.Is(err, pgx.ErrNoRows) {
		return cliente, err
	}

	nombre, apellido := in.NombreCliente, in.ApellidoCliente
	if tx != nil {
		if _, err := tx.Exec(ctx, "SAVEPOINT "+sp); err != nil {
			return cliente, err
		}
	}
	_, err = q.CreateCliente(ctx, sqlc.CreateClienteParams{
		TenantID: tenantID, Telefono: in.TelefonoCliente, Nombre: &nombre,
		Apellido: &apellido, Email: in.EmailCliente,
	})
	if err != nil {
		if tx != nil {
			if _, rbErr := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+sp); rbErr != nil {
				return cliente, rbErr
			}
		}
		if !pgutil.IsUniqueViolation(err) {
			return cliente, err
		}
	}
	// Relee siempre: si hubo una carrera, gana el cliente ya creado.
	return q.GetCliente(ctx, params)
}

func (s *Service) existeDuplicado(ctx context.Context, q *sqlc.Queries, tenantID int64, fechaHora time.Time) (bool, error) {
	// Java solo detecta duplicados >= ahora (findDateTimes(ahora)+equals).
	if fechaHora.Before(s.now()) {
		return false, nil
	}
	return q.ExistsTurnoEnFecha(ctx, sqlc.ExistsTurnoEnFechaParams{TenantID: tenantID, FechaHora: ts(fechaHora)})
}

// Update replica el PATCH de Java: copyProperties sin nulos, sin validar
// reglas, respuesta con el objeto en memoria (sin refrescar asociaciones).
func (s *Service) Update(ctx context.Context, tenantID, id int64, p Patch) (*Turno, error) {
	var out *Turno
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetTurno(ctx, sqlc.GetTurnoParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			out = nil
			return nil
		}
		if err != nil {
			return err
		}
		actual := aDominioGet(r)
		res := actual

		var clienteID *string
		if actual.Cliente != nil {
			clienteID = actual.Cliente.Telefono
		}
		var servicioID *int64
		if actual.Servicio != nil {
			servicioID = actual.Servicio.ID
		}

		if p.FechaHora != nil {
			res.FechaHora = *p.FechaHora
		}
		if p.ClienteDado {
			res.Cliente = p.Cliente
			if p.Cliente != nil {
				clienteID = p.Cliente.Telefono
			} else {
				clienteID = nil
			}
		}
		if p.ServicioDado {
			res.Servicio = p.Servicio
			if p.Servicio != nil {
				servicioID = p.Servicio.ID
			} else {
				servicioID = nil
			}
		}

		if persistir := servicioID; persistir != nil {
			if _, err := q.GetServicio(ctx, sqlc.GetServicioParams{TenantID: tenantID, ID: *persistir}); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return apperr.New(apperr.KindNotFound, "No existe un servicio con el id %d", *persistir)
				}
				return err
			}
		}

		persistirServicio := servicioID
		if p.ServicioDado && p.Servicio != nil && p.Servicio.ID == nil {
			persistirServicio = nil
			if actual.Servicio != nil {
				persistirServicio = actual.Servicio.ID
			}
		}

		_, err = q.UpdateTurno(ctx, sqlc.UpdateTurnoParams{
			TenantID: tenantID, ID: id, FechaHora: ts(res.FechaHora),
			ClienteID: clienteID, ServicioID: persistirServicio,
		})
		if pgutil.IsUniqueViolation(err) {
			return turnoDuplicado(res.FechaHora)
		}
		if err != nil {
			return err
		}
		out = &res
		return nil
	})
	return out, err
}

// Delete elimina un turno; false si no existía.
func (s *Service) Delete(ctx context.Context, tenantID, id int64) (bool, error) {
	var ok bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		n, err := q.DeleteTurno(ctx, sqlc.DeleteTurnoParams{TenantID: tenantID, ID: id})
		if err != nil {
			return err
		}
		ok = n > 0
		return nil
	})
	return ok, err
}

// Exists indica si el turno existe en el tenant.
func (s *Service) Exists(ctx context.Context, tenantID, id int64) (bool, error) {
	var exists bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		_, err := q.GetTurno(ctx, sqlc.GetTurnoParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			exists = false
			return nil
		}
		if err != nil {
			return err
		}
		exists = true
		return nil
	})
	return exists, err
}

// FindDateTimes replica /turnos/findDateTimes: fecha_hora >= ahora, formato
// LocalDateTime.toString() (sin segundos cuando son cero).
func (s *Service) FindDateTimes(ctx context.Context, tenantID int64) ([]string, error) {
	var out []string
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.FindDateTimesDesde(ctx, sqlc.FindDateTimesDesdeParams{TenantID: tenantID, Desde: ts(s.now())})
		if err != nil {
			return err
		}
		out = make([]string, 0, len(rows))
		for _, f := range rows {
			out = append(out, timeutil.FormatJavaToString(f.Time))
		}
		return nil
	})
	return out, err
}

func validarFechaHora(f time.Time) error {
	hora := f.Hour()*60 + f.Minute()
	if hora < 13*60 || hora > 20*60 {
		return apperr.New(apperr.KindBusinessRule, "La hora debe estar entre 13:00 y 20:00.")
	}
	if f.Minute() != 0 && f.Minute() != 30 {
		return apperr.New(apperr.KindBusinessRule, "La hora debe ser en intervalos de 30 minutos.")
	}
	return nil
}

func turnoDuplicado(f time.Time) error {
	return apperr.New(apperr.KindConflict,
		"Ya existe un turno para la fecha y hora especificada. Fecha y hora: %s", timeutil.FormatJavaToString(f))
}

func ts(t time.Time) pgtype.Timestamp { return pgtype.Timestamp{Time: t, Valid: true} }

func tenantIDStr(t int64) string {
	return fmt.Sprintf("%d", t)
}
