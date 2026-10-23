// Package clientes implementa el CRUD de clientes con ámbito de tenant.
package clientes

import (
	"context"
	"errors"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
	"github.com/Mateo-404/Barberia/backend-go/internal/pgutil"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Cliente es el modelo de dominio expuesto por la API.
type Cliente struct {
	Telefono string
	Nombre   string
	Apellido string
	Email    *string
}

// Input son los datos de creación/actualización.
type Input struct {
	Telefono string
	Nombre   string
	Apellido string
	Email    *string
}

// Patch contiene solo los campos presentes en el body (copyProperties).
type Patch struct {
	Nombre   *string
	Apellido *string
	Email    *string
}

// Service agrupa las operaciones de clientes.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

// New construye el servicio.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service { return &Service{pool: pool, q: q} }

func aDominio(telefono string, nombre, apellido, email *string) Cliente {
	return Cliente{Telefono: telefono, Nombre: deref(nombre), Apellido: deref(apellido), Email: email}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// List devuelve todos los clientes del tenant ordenados por teléfono.
func (s *Service) List(ctx context.Context, tenantID int64) ([]Cliente, error) {
	var out []Cliente
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListClientes(ctx, tenantID)
		if err != nil {
			return err
		}
		out = make([]Cliente, 0, len(rows))
		for _, r := range rows {
			out = append(out, aDominio(r.Telefono, r.Nombre, r.Apellido, r.Email))
		}
		return nil
	})
	return out, err
}

// Get devuelve un cliente o nil si no existe.
func (s *Service) Get(ctx context.Context, tenantID int64, telefono string) (*Cliente, error) {
	var out *Cliente
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetCliente(ctx, sqlc.GetClienteParams{TenantID: tenantID, Telefono: telefono})
		if errors.Is(err, pgx.ErrNoRows) {
			out = nil
			return nil
		}
		if err != nil {
			return err
		}
		c := aDominio(r.Telefono, r.Nombre, r.Apellido, r.Email)
		out = &c
		return nil
	})
	return out, err
}

// Create inserta un cliente; 409 si el teléfono ya existe en el tenant.
func (s *Service) Create(ctx context.Context, tenantID int64, in Input) (*Cliente, error) {
	var out *Cliente
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		if _, err := q.GetCliente(ctx, sqlc.GetClienteParams{TenantID: tenantID, Telefono: in.Telefono}); err == nil {
			return yaExiste(in.Telefono)
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		r, err := q.CreateCliente(ctx, sqlc.CreateClienteParams{
			Telefono: in.Telefono, Nombre: ptr(in.Nombre), Apellido: ptr(in.Apellido), Email: in.Email, TenantID: tenantID,
		})
		if pgutil.IsUniqueViolation(err) {
			return yaExiste(in.Telefono)
		}
		if err != nil {
			return err
		}
		c := aDominio(r.Telefono, r.Nombre, r.Apellido, r.Email)
		out = &c
		return nil
	})
	return out, err
}

// CreateAll inserta todos los clientes en una transacción.
func (s *Service) CreateAll(ctx context.Context, tenantID int64, ins []Input) ([]Cliente, error) {
	var out []Cliente
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		out = make([]Cliente, 0, len(ins))
		for _, in := range ins {
			if _, err := q.GetCliente(ctx, sqlc.GetClienteParams{TenantID: tenantID, Telefono: in.Telefono}); err == nil {
				return yaExiste(in.Telefono)
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			r, err := q.CreateCliente(ctx, sqlc.CreateClienteParams{
				Telefono: in.Telefono, Nombre: ptr(in.Nombre), Apellido: ptr(in.Apellido), Email: in.Email, TenantID: tenantID,
			})
			if pgutil.IsUniqueViolation(err) {
				return yaExiste(in.Telefono)
			}
			if err != nil {
				return err
			}
			out = append(out, aDominio(r.Telefono, r.Nombre, r.Apellido, r.Email))
		}
		return nil
	})
	return out, err
}

// Update aplica copyProperties sin nulos. El teléfono (PK) no se modifica.
func (s *Service) Update(ctx context.Context, tenantID int64, telefono string, p Patch) (*Cliente, error) {
	var out *Cliente
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r0, err := q.GetCliente(ctx, sqlc.GetClienteParams{TenantID: tenantID, Telefono: telefono})
		if errors.Is(err, pgx.ErrNoRows) {
			out = nil
			return nil
		}
		if err != nil {
			return err
		}
		actual := aDominio(r0.Telefono, r0.Nombre, r0.Apellido, r0.Email)
		if p.Nombre != nil {
			actual.Nombre = *p.Nombre
		}
		if p.Apellido != nil {
			actual.Apellido = *p.Apellido
		}
		if p.Email != nil {
			actual.Email = p.Email
		}
		r, err := q.UpdateCliente(ctx, sqlc.UpdateClienteParams{
			TenantID: tenantID, Telefono: telefono,
			Nombre: ptr(actual.Nombre), Apellido: ptr(actual.Apellido), Email: actual.Email,
		})
		if err != nil {
			return err
		}
		c := aDominio(r.Telefono, r.Nombre, r.Apellido, r.Email)
		out = &c
		return nil
	})
	return out, err
}

// Delete elimina un cliente; false si no existía. FK restrict → error → 500.
func (s *Service) Delete(ctx context.Context, tenantID int64, telefono string) (bool, error) {
	var ok bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		n, err := q.DeleteCliente(ctx, sqlc.DeleteClienteParams{TenantID: tenantID, Telefono: telefono})
		if err != nil {
			return err
		}
		ok = n > 0
		return nil
	})
	return ok, err
}

// Exists indica si el cliente existe en el tenant.
func (s *Service) Exists(ctx context.Context, tenantID int64, telefono string) (bool, error) {
	var exists bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		_, err := q.GetCliente(ctx, sqlc.GetClienteParams{TenantID: tenantID, Telefono: telefono})
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
func (s *Service) Page(ctx context.Context, tenantID int64, page, size int) ([]Cliente, int64, error) {
	var out []Cliente
	var total int64
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		t, err := q.CountClientes(ctx, tenantID)
		if err != nil {
			return err
		}
		total = t
		rows, err := q.PageClientes(ctx, sqlc.PageClientesParams{
			TenantID: tenantID, Limite: int32(size), Offset: int32(page * size),
		})
		if err != nil {
			return err
		}
		out = make([]Cliente, 0, len(rows))
		for _, r := range rows {
			out = append(out, aDominio(r.Telefono, r.Nombre, r.Apellido, r.Email))
		}
		return nil
	})
	return out, total, err
}

func yaExiste(telefono string) error {
	return apperr.New(apperr.KindAlreadyExists, "Ya existe un cliente con el teléfono %s", telefono)
}

func ptr(s string) *string { return &s }
