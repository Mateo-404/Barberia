// Package admins implements admin CRUD, login and the initial seed.
package admins

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/pgutil"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/tenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Admin is the domain model (the password hash is never exposed).
type Admin struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
	// Role: owner | admin | staff (SPEC §8).
	Role string
}

// Input is the data for create.
type Input struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	Role      string
}

// Patch holds only the fields present in the request body. A new password
// is hashed before it is stored (ADR-004).
type Patch struct {
	FirstName *string
	LastName  *string
	Email     *string
	Password  *string
	Role      *string
}

// Service groups admin operations.
type Service struct {
	pool   *pgxpool.Pool
	q      *sqlc.Queries
	jwt    *auth.JWTManager
	logger *slog.Logger
}

// New builds the service.
func New(pool *pgxpool.Pool, q *sqlc.Queries, jwt *auth.JWTManager, logger *slog.Logger) *Service {
	return &Service{pool: pool, q: q, jwt: jwt, logger: logger}
}

func toDomain(id int64, firstName, lastName, email *string, role string) Admin {
	if role == "" {
		role = auth.RoleOwner
	}
	return Admin{ID: id, FirstName: deref(firstName), LastName: deref(lastName), Email: deref(email), Role: role}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// List returns every admin of the tenant.
func (s *Service) List(ctx context.Context, tenantID int64) ([]Admin, error) {
	var out []Admin
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListAdmins(ctx, tenantID)
		if err != nil {
			return err
		}
		out = make([]Admin, 0, len(rows))
		for _, r := range rows {
			out = append(out, toDomain(r.ID, r.FirstName, r.LastName, r.Email, r.Role))
		}
		return nil
	})
	return out, err
}

// Get returns an admin, or nil when it does not exist.
func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Admin, error) {
	var out *Admin
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAdmin(ctx, sqlc.GetAdminParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		a := toDomain(r.ID, r.FirstName, r.LastName, r.Email, r.Role)
		out = &a
		return nil
	})
	return out, err
}

// Create creates an admin, hashing the password.
func (s *Service) Create(ctx context.Context, tenantID int64, in Input) (*Admin, error) {
	var out *Admin
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		a, err := create(ctx, q, tenantID, in)
		out = a
		return err
	})
	return out, err
}

// CreateAll creates every admin in one transaction; 409 on the first
// duplicate email.
func (s *Service) CreateAll(ctx context.Context, tenantID int64, ins []Input) ([]Admin, error) {
	var out []Admin
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		out = make([]Admin, 0, len(ins))
		for _, in := range ins {
			a, err := create(ctx, q, tenantID, in)
			if err != nil {
				return err
			}
			out = append(out, *a)
		}
		return nil
	})
	return out, err
}

func create(ctx context.Context, q *sqlc.Queries, tenantID int64, in Input) (*Admin, error) {
	exists, err := existsByEmail(ctx, q, tenantID, in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, alreadyExists(in.Email)
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	// Least privilege: admins created through the API are staff unless the
	// owner chooses otherwise.
	role := in.Role
	if role == "" {
		role = auth.RoleStaff
	}
	r, err := q.CreateAdmin(ctx, sqlc.CreateAdminParams{
		TenantID: tenantID, FirstName: &in.FirstName, LastName: &in.LastName,
		Email: &in.Email, PasswordHash: &hash, Role: role,
	})
	if pgutil.IsUniqueViolation(err) {
		return nil, alreadyExists(in.Email)
	}
	if err != nil {
		return nil, err
	}
	a := toDomain(r.ID, r.FirstName, r.LastName, r.Email, r.Role)
	return &a, nil
}

// Update applies the non-nil patch fields; a new password is hashed (ADR-004).
func (s *Service) Update(ctx context.Context, tenantID, id int64, p Patch) (*Admin, error) {
	var out *Admin
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAdmin(ctx, sqlc.GetAdminParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		cur := toDomain(r.ID, r.FirstName, r.LastName, r.Email, r.Role)
		hash := deref(r.PasswordHash)
		if p.FirstName != nil {
			cur.FirstName = *p.FirstName
		}
		if p.LastName != nil {
			cur.LastName = *p.LastName
		}
		if p.Email != nil {
			cur.Email = *p.Email
		}
		if p.Password != nil {
			if hash, err = auth.HashPassword(*p.Password); err != nil {
				return err
			}
		}
		if p.Role != nil {
			cur.Role = *p.Role
		}
		updated, err := q.UpdateAdmin(ctx, sqlc.UpdateAdminParams{
			TenantID: tenantID, ID: id, FirstName: &cur.FirstName, LastName: &cur.LastName,
			Email: &cur.Email, PasswordHash: &hash, Role: cur.Role,
		})
		if pgutil.IsUniqueViolation(err) {
			return alreadyExists(cur.Email)
		}
		if err != nil {
			return err
		}
		a := toDomain(updated.ID, updated.FirstName, updated.LastName, updated.Email, updated.Role)
		out = &a
		return nil
	})
	return out, err
}

// Delete removes an admin; false when it did not exist.
func (s *Service) Delete(ctx context.Context, tenantID, id int64) (bool, error) {
	var ok bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		n, err := q.DeleteAdmin(ctx, sqlc.DeleteAdminParams{TenantID: tenantID, ID: id})
		if err != nil {
			return err
		}
		ok = n > 0
		return nil
	})
	return ok, err
}

// Exists reports whether the admin exists in the tenant.
func (s *Service) Exists(ctx context.Context, tenantID, id int64) (bool, error) {
	a, err := s.Get(ctx, tenantID, id)
	return a != nil, err
}

// Page returns a 0-based page (without hashes, ADR-004) and the total count.
func (s *Service) Page(ctx context.Context, tenantID int64, page, size int) ([]Admin, int64, error) {
	var out []Admin
	var total int64
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		t, err := q.CountAdmins(ctx, tenantID)
		if err != nil {
			return err
		}
		total = t
		rows, err := q.PageAdmins(ctx, sqlc.PageAdminsParams{
			TenantID: tenantID, RowLimit: int32(size), RowOffset: int32(page * size),
		})
		if err != nil {
			return err
		}
		out = make([]Admin, 0, len(rows))
		for _, r := range rows {
			out = append(out, toDomain(r.ID, r.FirstName, r.LastName, r.Email, r.Role))
		}
		return nil
	})
	return out, total, err
}

// Login checks the credentials and issues a JWT.
func (s *Service) Login(ctx context.Context, tenantID int64, email, password string) (Admin, string, error) {
	var admin Admin
	var token string
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAdminByEmail(ctx, sqlc.GetAdminByEmailParams{TenantID: tenantID, Email: &email})
		if errors.Is(err, pgx.ErrNoRows) {
			return invalidCredentials()
		}
		if err != nil {
			return err
		}
		ok, err := auth.VerifyPassword(password, deref(r.PasswordHash))
		if err != nil || !ok {
			return invalidCredentials()
		}
		admin = toDomain(r.ID, r.FirstName, r.LastName, r.Email, r.Role)
		t, err := s.jwt.Issue(admin.ID, admin.Email, tenantID, admin.Role)
		if err != nil {
			return err
		}
		token = t
		return nil
	})
	return admin, token, err
}

// Seed creates the initial admin when the tenant has none.
func (s *Service) Seed(ctx context.Context, tenantID int64, email, password, firstName, lastName string) error {
	ctx = tenant.WithID(ctx, tenantID)
	return database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		total, err := q.CountAdmins(ctx, tenantID)
		if err != nil {
			return err
		}
		if total > 0 {
			s.logger.Info("tenant already has admins; seed skipped", "tenant_id", tenantID)
			return nil
		}
		if len(password) < 8 {
			return errors.New("ADMIN_SEED_PASSWORD must be at least 8 characters")
		}
		hash, err := auth.HashPassword(password)
		if err != nil {
			return err
		}
		if _, err := q.CreateAdmin(ctx, sqlc.CreateAdminParams{
			TenantID: tenantID, FirstName: &firstName, LastName: &lastName, Email: &email, PasswordHash: &hash,
			Role: auth.RoleOwner,
		}); err != nil {
			return err
		}
		s.logger.Info("seed admin created", "email", email)
		return nil
	})
}

func existsByEmail(ctx context.Context, q *sqlc.Queries, tenantID int64, email string) (bool, error) {
	_, err := q.GetAdminByEmail(ctx, sqlc.GetAdminByEmailParams{TenantID: tenantID, Email: &email})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func alreadyExists(email string) error {
	return apperr.New(apperr.KindAlreadyExists, "admin.already_exists", apperr.Params{"email": email},
		"an admin with email %s already exists", email)
}

func invalidCredentials() error {
	return apperr.New(apperr.KindInvalidCredentials, "auth.invalid_credentials", nil, "invalid credentials")
}
