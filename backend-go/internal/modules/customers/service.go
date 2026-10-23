// Package customers implements tenant-scoped customer CRUD.
package customers

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

// Customer is the domain model exposed by the API. The phone is the key.
type Customer struct {
	Phone     string
	FirstName string
	LastName  string
	Email     *string
}

// Input is the data for create.
type Input struct {
	Phone     string
	FirstName string
	LastName  string
	Email     *string
}

// Patch holds only the fields present in the request body.
type Patch struct {
	FirstName *string
	LastName  *string
	Email     *string
}

// Service groups customer operations.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

// New builds the service.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service { return &Service{pool: pool, q: q} }

func toDomain(phone string, firstName, lastName, email *string) Customer {
	return Customer{Phone: phone, FirstName: deref(firstName), LastName: deref(lastName), Email: email}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// List returns every customer of the tenant ordered by phone.
func (s *Service) List(ctx context.Context, tenantID int64) ([]Customer, error) {
	var out []Customer
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListCustomers(ctx, tenantID)
		if err != nil {
			return err
		}
		out = make([]Customer, 0, len(rows))
		for _, r := range rows {
			out = append(out, toDomain(r.Phone, r.FirstName, r.LastName, r.Email))
		}
		return nil
	})
	return out, err
}

// Get returns a customer, or nil when it does not exist.
func (s *Service) Get(ctx context.Context, tenantID int64, phone string) (*Customer, error) {
	var out *Customer
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetCustomer(ctx, sqlc.GetCustomerParams{TenantID: tenantID, Phone: phone})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		c := toDomain(r.Phone, r.FirstName, r.LastName, r.Email)
		out = &c
		return nil
	})
	return out, err
}

// Create inserts a customer; 409 when the phone already exists in the tenant.
func (s *Service) Create(ctx context.Context, tenantID int64, in Input) (*Customer, error) {
	var out *Customer
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		c, err := create(ctx, q, tenantID, in)
		out = c
		return err
	})
	return out, err
}

// CreateAll inserts every customer in a single transaction.
func (s *Service) CreateAll(ctx context.Context, tenantID int64, ins []Input) ([]Customer, error) {
	var out []Customer
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		out = make([]Customer, 0, len(ins))
		for _, in := range ins {
			c, err := create(ctx, q, tenantID, in)
			if err != nil {
				return err
			}
			out = append(out, *c)
		}
		return nil
	})
	return out, err
}

func create(ctx context.Context, q *sqlc.Queries, tenantID int64, in Input) (*Customer, error) {
	if _, err := q.GetCustomer(ctx, sqlc.GetCustomerParams{TenantID: tenantID, Phone: in.Phone}); err == nil {
		return nil, alreadyExists(in.Phone)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	r, err := q.CreateCustomer(ctx, sqlc.CreateCustomerParams{
		Phone: in.Phone, FirstName: &in.FirstName, LastName: &in.LastName, Email: in.Email, TenantID: tenantID,
	})
	if pgutil.IsUniqueViolation(err) {
		return nil, alreadyExists(in.Phone)
	}
	if err != nil {
		return nil, err
	}
	c := toDomain(r.Phone, r.FirstName, r.LastName, r.Email)
	return &c, nil
}

// Update applies the non-nil patch fields. The phone (key) cannot change.
func (s *Service) Update(ctx context.Context, tenantID int64, phone string, p Patch) (*Customer, error) {
	var out *Customer
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r0, err := q.GetCustomer(ctx, sqlc.GetCustomerParams{TenantID: tenantID, Phone: phone})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		cur := toDomain(r0.Phone, r0.FirstName, r0.LastName, r0.Email)
		if p.FirstName != nil {
			cur.FirstName = *p.FirstName
		}
		if p.LastName != nil {
			cur.LastName = *p.LastName
		}
		if p.Email != nil {
			cur.Email = p.Email
		}
		r, err := q.UpdateCustomer(ctx, sqlc.UpdateCustomerParams{
			TenantID: tenantID, Phone: phone,
			FirstName: &cur.FirstName, LastName: &cur.LastName, Email: cur.Email,
		})
		if err != nil {
			return err
		}
		c := toDomain(r.Phone, r.FirstName, r.LastName, r.Email)
		out = &c
		return nil
	})
	return out, err
}

// Delete removes a customer; false when it did not exist. A customer with
// appointments is protected by an FK (RESTRICT) and surfaces as an error.
func (s *Service) Delete(ctx context.Context, tenantID int64, phone string) (bool, error) {
	var ok bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		n, err := q.DeleteCustomer(ctx, sqlc.DeleteCustomerParams{TenantID: tenantID, Phone: phone})
		if err != nil {
			return err
		}
		ok = n > 0
		return nil
	})
	return ok, err
}

// Exists reports whether the customer exists in the tenant.
func (s *Service) Exists(ctx context.Context, tenantID int64, phone string) (bool, error) {
	c, err := s.Get(ctx, tenantID, phone)
	return c != nil, err
}

// Page returns a 0-based page and the total count.
func (s *Service) Page(ctx context.Context, tenantID int64, page, size int) ([]Customer, int64, error) {
	var out []Customer
	var total int64
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		t, err := q.CountCustomers(ctx, tenantID)
		if err != nil {
			return err
		}
		total = t
		rows, err := q.PageCustomers(ctx, sqlc.PageCustomersParams{
			TenantID: tenantID, RowLimit: int32(size), RowOffset: int32(page * size),
		})
		if err != nil {
			return err
		}
		out = make([]Customer, 0, len(rows))
		for _, r := range rows {
			out = append(out, toDomain(r.Phone, r.FirstName, r.LastName, r.Email))
		}
		return nil
	})
	return out, total, err
}

func alreadyExists(phone string) error {
	return apperr.New(apperr.KindAlreadyExists, "customer.already_exists", apperr.Params{"phone": phone},
		"a customer with phone %s already exists", phone)
}
