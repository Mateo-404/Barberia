// Package appointments implements bookings with business rules and atomic
// double-booking protection (UNIQUE per tenant + starts_at).
package appointments

import (
	"context"
	"errors"
	"strconv"
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

// Opening hours and slot size of a booking.
const (
	openMinute    = 13 * 60
	closeMinute   = 20 * 60
	slotMinutes   = 30
	openingLabel  = "13:00"
	closingLabel  = "20:00"
	savepointName = "sp_customer"
)

// Customer is the association (may be partial in a PATCH).
type Customer struct {
	Phone     *string
	FirstName *string
	LastName  *string
	Email     *string
}

// Offering is the service association (may be partial in a PATCH).
type Offering struct {
	ID    *int64
	Name  *string
	Price *float32
}

// Appointment is the domain model.
type Appointment struct {
	ID       int64
	StartsAt time.Time
	Customer *Customer
	Service  *Offering
}

// Input is the data needed to create a booking.
type Input struct {
	StartsAt          time.Time
	ServiceID         int64
	CustomerPhone     string
	CustomerFirstName string
	CustomerLastName  string
	CustomerEmail     *string
}

// Patch holds the fields present in a PATCH body. Business rules are not
// re-validated on PATCH (admin corrections).
type Patch struct {
	StartsAt     *time.Time
	Customer     *Customer
	Service      *Offering
	CustomerSent bool
	ServiceSent  bool
}

// Service groups appointment operations.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
	now  func() time.Time
}

// New builds the service.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service {
	return &Service{pool: pool, q: q, now: time.Now}
}

func toDomain(r sqlc.ListAppointmentsRow) Appointment {
	a := Appointment{ID: r.ID, StartsAt: r.StartsAt.Time}
	if r.Phone != nil || r.FirstName != nil || r.LastName != nil || r.Email != nil {
		a.Customer = &Customer{Phone: r.Phone, FirstName: r.FirstName, LastName: r.LastName, Email: r.Email}
	}
	if r.ServiceID != nil || r.ServiceName != nil || r.ServicePrice != nil {
		a.Service = &Offering{ID: r.ServiceID, Name: r.ServiceName, Price: r.ServicePrice}
	}
	return a
}

// List returns the tenant's appointments ordered by id.
func (s *Service) List(ctx context.Context, tenantID int64) ([]Appointment, error) {
	var out []Appointment
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListAppointments(ctx, tenantID)
		if err != nil {
			return err
		}
		out = make([]Appointment, 0, len(rows))
		for _, r := range rows {
			out = append(out, toDomain(r))
		}
		return nil
	})
	return out, err
}

// Get returns an appointment, or nil when it does not exist.
func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Appointment, error) {
	var out *Appointment
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAppointment(ctx, sqlc.GetAppointmentParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		a := toDomain(sqlc.ListAppointmentsRow(r))
		out = &a
		return nil
	})
	return out, err
}

// Page returns a 0-based page and the total count.
func (s *Service) Page(ctx context.Context, tenantID int64, page, size int) ([]Appointment, int64, error) {
	var out []Appointment
	var total int64
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		t, err := q.CountAppointments(ctx, tenantID)
		if err != nil {
			return err
		}
		total = t
		rows, err := q.PageAppointments(ctx, sqlc.PageAppointmentsParams{
			TenantID: tenantID, RowLimit: int32(size), RowOffset: int32(page * size),
		})
		if err != nil {
			return err
		}
		out = make([]Appointment, 0, len(rows))
		for _, r := range rows {
			out = append(out, toDomain(sqlc.ListAppointmentsRow(r)))
		}
		return nil
	})
	return out, total, err
}

// Create applies the booking rules and guards against double booking.
func (s *Service) Create(ctx context.Context, tenantID int64, in Input) (*Appointment, error) {
	out, err := s.CreateAll(ctx, tenantID, []Input{in})
	if err != nil {
		return nil, err
	}
	return &out[0], nil
}

// CreateAll creates every appointment in a single transaction (ADR-005).
func (s *Service) CreateAll(ctx context.Context, tenantID int64, ins []Input) ([]Appointment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.FormatInt(tenantID, 10)); err != nil {
		return nil, err
	}
	qtx := s.q.WithTx(tx)
	out := make([]Appointment, 0, len(ins))
	for _, in := range ins {
		a, err := s.createInTx(ctx, qtx, tx, tenantID, in)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) createInTx(ctx context.Context, q *sqlc.Queries, tx pgx.Tx, tenantID int64, in Input) (*Appointment, error) {
	svc, err := q.GetService(ctx, sqlc.GetServiceParams{TenantID: tenantID, ID: in.ServiceID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, serviceNotFound(in.ServiceID)
	}
	if err != nil {
		return nil, err
	}
	if err := validateStartsAt(in.StartsAt); err != nil {
		return nil, err
	}
	taken, err := s.slotTaken(ctx, q, tenantID, in.StartsAt)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, slotTakenError(in.StartsAt)
	}

	customer, err := findOrCreateCustomer(ctx, q, tx, tenantID, in)
	if err != nil {
		return nil, err
	}

	created, err := q.CreateAppointment(ctx, sqlc.CreateAppointmentParams{
		TenantID: tenantID, StartsAt: ts(in.StartsAt), CustomerPhone: &in.CustomerPhone, ServiceID: &in.ServiceID,
	})
	if pgutil.IsUniqueViolation(err) {
		return nil, slotTakenError(in.StartsAt)
	}
	if err != nil {
		return nil, err
	}

	return &Appointment{
		ID:       created.ID,
		StartsAt: created.StartsAt.Time,
		Customer: &Customer{Phone: &customer.Phone, FirstName: customer.FirstName, LastName: customer.LastName, Email: customer.Email},
		Service:  &Offering{ID: &svc.ID, Name: svc.Name, Price: svc.Price},
	}, nil
}

// findOrCreateCustomer uses a savepoint so that a race on the same phone does
// not abort the transaction (23505 → roll back to the savepoint and re-read
// the customer the other request created).
func findOrCreateCustomer(ctx context.Context, q *sqlc.Queries, tx pgx.Tx, tenantID int64, in Input) (sqlc.GetCustomerRow, error) {
	params := sqlc.GetCustomerParams{TenantID: tenantID, Phone: in.CustomerPhone}
	customer, err := q.GetCustomer(ctx, params)
	if !errors.Is(err, pgx.ErrNoRows) {
		return customer, err
	}

	if _, err := tx.Exec(ctx, "SAVEPOINT "+savepointName); err != nil {
		return customer, err
	}
	_, err = q.CreateCustomer(ctx, sqlc.CreateCustomerParams{
		TenantID: tenantID, Phone: in.CustomerPhone, FirstName: &in.CustomerFirstName,
		LastName: &in.CustomerLastName, Email: in.CustomerEmail,
	})
	if err != nil {
		if _, rbErr := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+savepointName); rbErr != nil {
			return customer, rbErr
		}
		if !pgutil.IsUniqueViolation(err) {
			return customer, err
		}
	}
	// Always re-read: after a race, the customer that won is returned.
	return q.GetCustomer(ctx, params)
}

func (s *Service) slotTaken(ctx context.Context, q *sqlc.Queries, tenantID int64, startsAt time.Time) (bool, error) {
	// Past slots are not checked here; the unique index still guards them.
	if startsAt.Before(s.now()) {
		return false, nil
	}
	return q.AppointmentExistsAt(ctx, sqlc.AppointmentExistsAtParams{TenantID: tenantID, StartsAt: ts(startsAt)})
}

// Update applies the patch without re-validating business rules and
// returns the in-memory result (associations are not reloaded).
func (s *Service) Update(ctx context.Context, tenantID, id int64, p Patch) (*Appointment, error) {
	var out *Appointment
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAppointment(ctx, sqlc.GetAppointmentParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		cur := toDomain(sqlc.ListAppointmentsRow(r))
		res := cur

		var customerPhone *string
		if cur.Customer != nil {
			customerPhone = cur.Customer.Phone
		}
		var serviceID *int64
		if cur.Service != nil {
			serviceID = cur.Service.ID
		}

		if p.StartsAt != nil {
			res.StartsAt = *p.StartsAt
		}
		if p.CustomerSent {
			res.Customer = p.Customer
			customerPhone = nil
			if p.Customer != nil {
				customerPhone = p.Customer.Phone
			}
		}
		if p.ServiceSent {
			res.Service = p.Service
			serviceID = nil
			if p.Service != nil {
				serviceID = p.Service.ID
			}
		}

		if serviceID != nil {
			if _, err := q.GetService(ctx, sqlc.GetServiceParams{TenantID: tenantID, ID: *serviceID}); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return serviceNotFound(*serviceID)
				}
				return err
			}
		}

		// A service object without id keeps the stored service.
		storedServiceID := serviceID
		if p.ServiceSent && p.Service != nil && p.Service.ID == nil {
			storedServiceID = nil
			if cur.Service != nil {
				storedServiceID = cur.Service.ID
			}
		}

		_, err = q.UpdateAppointment(ctx, sqlc.UpdateAppointmentParams{
			TenantID: tenantID, ID: id, StartsAt: ts(res.StartsAt),
			CustomerPhone: customerPhone, ServiceID: storedServiceID,
		})
		if pgutil.IsUniqueViolation(err) {
			return slotTakenError(res.StartsAt)
		}
		if err != nil {
			return err
		}
		out = &res
		return nil
	})
	return out, err
}

// Delete removes an appointment; false when it did not exist.
func (s *Service) Delete(ctx context.Context, tenantID, id int64) (bool, error) {
	var ok bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		n, err := q.DeleteAppointment(ctx, sqlc.DeleteAppointmentParams{TenantID: tenantID, ID: id})
		if err != nil {
			return err
		}
		ok = n > 0
		return nil
	})
	return ok, err
}

// Exists reports whether the appointment exists in the tenant.
func (s *Service) Exists(ctx context.Context, tenantID, id int64) (bool, error) {
	a, err := s.Get(ctx, tenantID, id)
	return a != nil, err
}

// BookedSlots returns the booked start times from now on, formatted without
// zero seconds ("2026-10-22T15:00").
func (s *Service) BookedSlots(ctx context.Context, tenantID int64) ([]string, error) {
	var out []string
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListBookedSlotsFrom(ctx, sqlc.ListBookedSlotsFromParams{TenantID: tenantID, FromTs: ts(s.now())})
		if err != nil {
			return err
		}
		out = make([]string, 0, len(rows))
		for _, t := range rows {
			out = append(out, timeutil.FormatShort(t.Time))
		}
		return nil
	})
	return out, err
}

func validateStartsAt(t time.Time) error {
	minute := t.Hour()*60 + t.Minute()
	if minute < openMinute || minute > closeMinute {
		return apperr.New(apperr.KindBusinessRule, "appointment.outside_hours",
			apperr.Params{"open": openingLabel, "close": closingLabel},
			"the time must be between %s and %s", openingLabel, closingLabel)
	}
	if t.Minute()%slotMinutes != 0 {
		return apperr.New(apperr.KindBusinessRule, "appointment.invalid_interval",
			apperr.Params{"minutes": slotMinutes},
			"the time must be on %d-minute intervals", slotMinutes)
	}
	return nil
}

func slotTakenError(t time.Time) error {
	at := timeutil.FormatShort(t)
	return apperr.New(apperr.KindConflict, "appointment.slot_taken", apperr.Params{"startsAt": at},
		"an appointment already exists at %s", at)
}

func serviceNotFound(id int64) error {
	return apperr.New(apperr.KindNotFound, "service.not_found", apperr.Params{"id": id},
		"no service with id %d", id)
}

func ts(t time.Time) pgtype.Timestamp { return pgtype.Timestamp{Time: t, Valid: true} }
