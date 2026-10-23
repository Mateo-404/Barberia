// Package services implements tenant-scoped CRUD for the services a
// business offers (haircut, beard, …).
package services

import (
	"context"
	"errors"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Offering is the domain model of a bookable service.
type Offering struct {
	ID    int64
	Name  string
	Price float32
}

// Input is the data for create.
type Input struct {
	Name  string
	Price float32
}

// Patch holds only the fields present in the request body.
type Patch struct {
	Name  *string
	Price *float32
}

// Service groups the operations.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

// New builds the service.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service { return &Service{pool: pool, q: q} }

func toDomain(id int64, name *string, price *float32) Offering {
	o := Offering{ID: id}
	if name != nil {
		o.Name = *name
	}
	if price != nil {
		o.Price = *price
	}
	return o
}

// List returns every service of the tenant.
func (s *Service) List(ctx context.Context, tenantID int64) ([]Offering, error) {
	var out []Offering
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListServices(ctx, tenantID)
		if err != nil {
			return err
		}
		out = make([]Offering, 0, len(rows))
		for _, r := range rows {
			out = append(out, toDomain(r.ID, r.Name, r.Price))
		}
		return nil
	})
	return out, err
}

// Get returns a service, or nil when it does not exist.
func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Offering, error) {
	var out *Offering
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetService(ctx, sqlc.GetServiceParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		o := toDomain(r.ID, r.Name, r.Price)
		out = &o
		return nil
	})
	return out, err
}

// Create inserts a service.
func (s *Service) Create(ctx context.Context, tenantID int64, in Input) (*Offering, error) {
	var out *Offering
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.CreateService(ctx, sqlc.CreateServiceParams{Name: &in.Name, Price: &in.Price, TenantID: tenantID})
		if err != nil {
			return err
		}
		o := toDomain(r.ID, r.Name, r.Price)
		out = &o
		return nil
	})
	return out, err
}

// CreateAll inserts every service in a single transaction.
func (s *Service) CreateAll(ctx context.Context, tenantID int64, ins []Input) ([]Offering, error) {
	var out []Offering
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		out = make([]Offering, 0, len(ins))
		for _, in := range ins {
			r, err := q.CreateService(ctx, sqlc.CreateServiceParams{Name: &in.Name, Price: &in.Price, TenantID: tenantID})
			if err != nil {
				return err
			}
			out = append(out, toDomain(r.ID, r.Name, r.Price))
		}
		return nil
	})
	return out, err
}

// Update applies the non-nil patch fields.
func (s *Service) Update(ctx context.Context, tenantID, id int64, p Patch) (*Offering, error) {
	var out *Offering
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r0, err := q.GetService(ctx, sqlc.GetServiceParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		cur := toDomain(r0.ID, r0.Name, r0.Price)
		if p.Name != nil {
			cur.Name = *p.Name
		}
		if p.Price != nil {
			cur.Price = *p.Price
		}
		r, err := q.UpdateService(ctx, sqlc.UpdateServiceParams{
			TenantID: tenantID, ID: id, Name: &cur.Name, Price: &cur.Price,
		})
		if err != nil {
			return err
		}
		o := toDomain(r.ID, r.Name, r.Price)
		out = &o
		return nil
	})
	return out, err
}

// Delete removes a service; false when it did not exist.
func (s *Service) Delete(ctx context.Context, tenantID, id int64) (bool, error) {
	var ok bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		n, err := q.DeleteService(ctx, sqlc.DeleteServiceParams{TenantID: tenantID, ID: id})
		if err != nil {
			return err
		}
		ok = n > 0
		return nil
	})
	return ok, err
}

// Exists reports whether the service exists in the tenant.
func (s *Service) Exists(ctx context.Context, tenantID, id int64) (bool, error) {
	o, err := s.Get(ctx, tenantID, id)
	return o != nil, err
}

// Page returns a 0-based page and the total count.
func (s *Service) Page(ctx context.Context, tenantID int64, page, size int) ([]Offering, int64, error) {
	var out []Offering
	var total int64
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		t, err := q.CountServices(ctx, tenantID)
		if err != nil {
			return err
		}
		total = t
		rows, err := q.PageServices(ctx, sqlc.PageServicesParams{
			TenantID: tenantID, RowLimit: int32(size), RowOffset: int32(page * size),
		})
		if err != nil {
			return err
		}
		out = make([]Offering, 0, len(rows))
		for _, r := range rows {
			out = append(out, toDomain(r.ID, r.Name, r.Price))
		}
		return nil
	})
	return out, total, err
}
