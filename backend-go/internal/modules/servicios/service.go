// Package servicios implementa el CRUD de servicios con ámbito de tenant.
package servicios

import (
	"context"
	"errors"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Servicio es el modelo de dominio.
type Servicio struct {
	ID     int64
	Tipo   string
	Precio float32
}

// Input son los datos de creación.
type Input struct {
	Tipo   string
	Precio float32
}

// Patch contiene solo los campos presentes en el body.
type Patch struct {
	Tipo   *string
	Precio *float32
}

// Service agrupa las operaciones.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

// New construye el servicio.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service { return &Service{pool: pool, q: q} }

func aDominio(id int64, tipo *string, precio *float32) Servicio {
	s := Servicio{ID: id}
	if tipo != nil {
		s.Tipo = *tipo
	}
	if precio != nil {
		s.Precio = *precio
	}
	return s
}

// List devuelve todos los servicios del tenant.
func (s *Service) List(ctx context.Context, tenantID int64) ([]Servicio, error) {
	var out []Servicio
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListServicios(ctx, tenantID)
		if err != nil {
			return err
		}
		out = make([]Servicio, 0, len(rows))
		for _, r := range rows {
			out = append(out, aDominio(r.ID, r.Tipo, r.Precio))
		}
		return nil
	})
	return out, err
}

// Get devuelve un servicio o nil.
func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Servicio, error) {
	var out *Servicio
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetServicio(ctx, sqlc.GetServicioParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			out = nil
			return nil
		}
		if err != nil {
			return err
		}
		v := aDominio(r.ID, r.Tipo, r.Precio)
		out = &v
		return nil
	})
	return out, err
}

// Create inserta un servicio.
func (s *Service) Create(ctx context.Context, tenantID int64, in Input) (*Servicio, error) {
	var out *Servicio
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		tipo, precio := in.Tipo, in.Precio
		r, err := q.CreateServicio(ctx, sqlc.CreateServicioParams{
			Tipo: &tipo, Precio: &precio, TenantID: tenantID,
		})
		if err != nil {
			return err
		}
		v := aDominio(r.ID, r.Tipo, r.Precio)
		out = &v
		return nil
	})
	return out, err
}

// CreateAll inserta todos los servicios en una transacción.
func (s *Service) CreateAll(ctx context.Context, tenantID int64, ins []Input) ([]Servicio, error) {
	var out []Servicio
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		out = make([]Servicio, 0, len(ins))
		for _, in := range ins {
			tipo, precio := in.Tipo, in.Precio
			r, err := q.CreateServicio(ctx, sqlc.CreateServicioParams{Tipo: &tipo, Precio: &precio, TenantID: tenantID})
			if err != nil {
				return err
			}
			out = append(out, aDominio(r.ID, r.Tipo, r.Precio))
		}
		return nil
	})
	return out, err
}

// Update aplica copyProperties sin nulos.
func (s *Service) Update(ctx context.Context, tenantID, id int64, p Patch) (*Servicio, error) {
	var out *Servicio
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r0, err := q.GetServicio(ctx, sqlc.GetServicioParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			out = nil
			return nil
		}
		if err != nil {
			return err
		}
		actual := aDominio(r0.ID, r0.Tipo, r0.Precio)
		if p.Tipo != nil {
			actual.Tipo = *p.Tipo
		}
		if p.Precio != nil {
			actual.Precio = *p.Precio
		}
		r, err := q.UpdateServicio(ctx, sqlc.UpdateServicioParams{
			TenantID: tenantID, ID: id, Tipo: &actual.Tipo, Precio: &actual.Precio,
		})
		if err != nil {
			return err
		}
		v := aDominio(r.ID, r.Tipo, r.Precio)
		out = &v
		return nil
	})
	return out, err
}

// Delete elimina un servicio; false si no existía.
func (s *Service) Delete(ctx context.Context, tenantID, id int64) (bool, error) {
	var ok bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		n, err := q.DeleteServicio(ctx, sqlc.DeleteServicioParams{TenantID: tenantID, ID: id})
		if err != nil {
			return err
		}
		ok = n > 0
		return nil
	})
	return ok, err
}

// Exists indica si el servicio existe en el tenant.
func (s *Service) Exists(ctx context.Context, tenantID, id int64) (bool, error) {
	var exists bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		_, err := q.GetServicio(ctx, sqlc.GetServicioParams{TenantID: tenantID, ID: id})
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

// Page devuelve una página 0-based.
func (s *Service) Page(ctx context.Context, tenantID int64, page, size int) ([]Servicio, int64, error) {
	var out []Servicio
	var total int64
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		t, err := q.CountServicios(ctx, tenantID)
		if err != nil {
			return err
		}
		total = t
		rows, err := q.PageServicios(ctx, sqlc.PageServiciosParams{
			TenantID: tenantID, Limite: int32(size), Offset: int32(page * size),
		})
		if err != nil {
			return err
		}
		out = make([]Servicio, 0, len(rows))
		for _, r := range rows {
			out = append(out, aDominio(r.ID, r.Tipo, r.Precio))
		}
		return nil
	})
	return out, total, err
}
