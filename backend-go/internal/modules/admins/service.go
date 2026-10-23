// Package admins implements admin CRUD, login and the initial seed.
package admins

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

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
	pool    *pgxpool.Pool
	q       *sqlc.Queries
	jwt     *auth.JWTManager
	logger  *slog.Logger
	totpKey string
}

// New builds the service.
func New(pool *pgxpool.Pool, q *sqlc.Queries, jwt *auth.JWTManager, logger *slog.Logger, totpKey string) *Service {
	return &Service{pool: pool, q: q, jwt: jwt, logger: logger, totpKey: totpKey}
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

// Login checks the credentials and issues a JWT. When the admin has TOTP
// enabled, it returns a short-lived pending token and requires2FA=true.
func (s *Service) Login(ctx context.Context, tenantID int64, email, password string) (Admin, string, bool, error) {
	var admin Admin
	var token string
	var requires2FA bool
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
		// TOTP check
		totpRow, err := q.GetAdminTOTP(ctx, sqlc.GetAdminTOTPParams{TenantID: tenantID, ID: r.ID})
		if err == nil && totpRow.TotpEnabled {
			tok, err := s.jwt.IssuePending2FA(admin.ID, admin.Email, tenantID, admin.Role)
			if err != nil {
				return err
			}
			token = tok
			requires2FA = true
			return nil
		}
		t, err := s.jwt.Issue(admin.ID, admin.Email, tenantID, admin.Role)
		if err != nil {
			return err
		}
		token = t
		return nil
	})
	return admin, token, requires2FA, err
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

// LoginWithGoogle finds or creates an admin from a verified Google identity.
// Returns requires2FA=true with a pending token when TOTP is enabled.
func (s *Service) LoginWithGoogle(ctx context.Context, tenantID int64, gu auth.GoogleUser) (Admin, string, bool, error) {
	var admin Admin
	var token string
	var requires2FA bool
	var adminID int64
	var adminEmail, adminRole string
	var firstName, lastName *string
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		// 1. Find by google_sub.
		r, err := q.GetAdminByGoogleSub(ctx, sqlc.GetAdminByGoogleSubParams{TenantID: tenantID, GoogleSub: &gu.Sub})
		if errors.Is(err, pgx.ErrNoRows) {
			// 2. Link by email if an admin with that email exists.
			r2, err2 := q.GetAdminByEmail(ctx, sqlc.GetAdminByEmailParams{TenantID: tenantID, Email: &gu.Email})
			if err2 == nil {
				linked, err := q.LinkAdminGoogle(ctx, sqlc.LinkAdminGoogleParams{TenantID: tenantID, ID: r2.ID, GoogleSub: &gu.Sub})
				if err != nil {
					return err
				}
				adminID = linked.ID
				firstName, lastName = linked.FirstName, linked.LastName
				adminEmail = deref(linked.Email)
				adminRole = linked.Role
			} else if errors.Is(err2, pgx.ErrNoRows) {
				// 3. Create a new staff admin.
				created, err := q.CreateAdminGoogle(ctx, sqlc.CreateAdminGoogleParams{
					FirstName:    strPtr(firstNameFrom(gu.Name)),
					LastName:     strPtr(""),
					Email:        &gu.Email,
					PasswordHash: nil,
					TenantID:     tenantID,
					Role:         auth.RoleStaff,
					GoogleSub:    &gu.Sub,
					AuthProvider: "google",
				})
				if err != nil {
					return err
				}
				adminID = created.ID
				firstName, lastName = created.FirstName, created.LastName
				adminEmail = deref(created.Email)
				adminRole = created.Role
			} else {
				return err2
			}
		} else if err != nil {
			return err
		} else {
			adminID = r.ID
			firstName, lastName = r.FirstName, r.LastName
			adminEmail = deref(r.Email)
			adminRole = r.Role
		}
		admin = Admin{ID: adminID, FirstName: deref(firstName), LastName: deref(lastName), Email: adminEmail, Role: adminRole}
		if admin.Role == "" {
			admin.Role = auth.RoleOwner
		}

		totpRow, err := q.GetAdminTOTP(ctx, sqlc.GetAdminTOTPParams{TenantID: tenantID, ID: adminID})
		if err == nil && totpRow.TotpEnabled {
			tok, err := s.jwt.IssuePending2FA(admin.ID, admin.Email, tenantID, admin.Role)
			if err != nil {
				return err
			}
			token = tok
			requires2FA = true
			return nil
		}
		t, err := s.jwt.Issue(admin.ID, admin.Email, tenantID, admin.Role)
		if err != nil {
			return err
		}
		token = t
		return nil
	})
	return admin, token, requires2FA, err
}

func firstNameFrom(full string) string {
	full = strings.TrimSpace(full)
	if i := strings.IndexByte(full, ' '); i > 0 {
		return full[:i]
	}
	if full == "" {
		return "User"
	}
	return full
}

func strPtr(s string) *string { return &s }

// SetupTOTP generates a new secret, stores it (not yet enabled) and returns
// the provisioning URI.
func (s *Service) SetupTOTP(ctx context.Context, tenantID, adminID int64) (string, string, error) {
	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		return "", "", err
	}
	enc, err := auth.EncryptTOTP(s.totpKey, secret)
	if err != nil {
		return "", "", err
	}
	err = database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		return q.SetAdminTOTPSecret(ctx, sqlc.SetAdminTOTPSecretParams{TenantID: tenantID, ID: adminID, TotpSecret: &enc})
	})
	if err != nil {
		return "", "", err
	}
	admin, _ := s.GetByID(ctx, tenantID, adminID)
	account := admin.Email
	if account == "" {
		account = fmt.Sprintf("admin-%d", adminID)
	}
	uri := auth.TOTPProvisioningURI("Barberia", account, secret)
	return secret, uri, nil
}

// GetByID returns a single admin.
func (s *Service) GetByID(ctx context.Context, tenantID, adminID int64) (Admin, error) {
	var admin Admin
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAdmin(ctx, sqlc.GetAdminParams{TenantID: tenantID, ID: adminID})
		if err != nil {
			return err
		}
		admin = toDomain(r.ID, r.FirstName, r.LastName, r.Email, r.Role)
		return nil
	})
	return admin, err
}

// VerifyTOTP enables TOTP when the code is valid.
func (s *Service) VerifyTOTP(ctx context.Context, tenantID, adminID int64, code string) error {
	var enc *string
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetAdminTOTP(ctx, sqlc.GetAdminTOTPParams{TenantID: tenantID, ID: adminID})
		if err != nil {
			return err
		}
		enc = row.TotpSecret
		return nil
	})
	if err != nil {
		return err
	}
	if enc == nil || *enc == "" {
		return errors.New("totp not set up")
	}
	secret, err := auth.DecryptTOTP(s.totpKey, *enc)
	if err != nil {
		return err
	}
	if !auth.VerifyTOTP(secret, code, timeNow()) {
		return invalidCredentials()
	}
	return database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		return q.EnableAdminTOTP(ctx, sqlc.EnableAdminTOTPParams{TenantID: tenantID, ID: adminID})
	})
}

// DisableTOTP clears the secret after verifying a code.
func (s *Service) DisableTOTP(ctx context.Context, tenantID, adminID int64, code string) error {
	var enc *string
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		row, err := q.GetAdminTOTP(ctx, sqlc.GetAdminTOTPParams{TenantID: tenantID, ID: adminID})
		if err != nil {
			return err
		}
		enc = row.TotpSecret
		return nil
	})
	if err != nil {
		return err
	}
	if enc == nil || *enc == "" {
		return nil
	}
	secret, err := auth.DecryptTOTP(s.totpKey, *enc)
	if err != nil {
		return err
	}
	if !auth.VerifyTOTP(secret, code, timeNow()) {
		return invalidCredentials()
	}
	return database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		return q.DisableAdminTOTP(ctx, sqlc.DisableAdminTOTPParams{TenantID: tenantID, ID: adminID})
	})
}

// ChallengeTOTP verifies a pending 2FA token + code and issues a full token.
func (s *Service) ChallengeTOTP(ctx context.Context, pendingClaims *auth.Claims, code string) (Admin, string, error) {
	var admin Admin
	var token string
	tenantID := pendingClaims.TenantID
	adminID, err := strconv.ParseInt(pendingClaims.Subject, 10, 64)
	if err != nil {
		return admin, "", auth.ErrInvalidToken
	}
	err = database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAdmin(ctx, sqlc.GetAdminParams{TenantID: tenantID, ID: adminID})
		if err != nil {
			return invalidCredentials()
		}
		admin = toDomain(r.ID, r.FirstName, r.LastName, r.Email, r.Role)
		row, err := q.GetAdminTOTP(ctx, sqlc.GetAdminTOTPParams{TenantID: tenantID, ID: adminID})
		if err != nil || row.TotpSecret == nil || *row.TotpSecret == "" {
			return invalidCredentials()
		}
		secret, err := auth.DecryptTOTP(s.totpKey, *row.TotpSecret)
		if err != nil {
			return err
		}
		if !auth.VerifyTOTP(secret, code, timeNow()) {
			return invalidCredentials()
		}
		t, err := s.jwt.Issue(admin.ID, admin.Email, tenantID, admin.Role)
		if err != nil {
			return err
		}
		token = t
		return nil
	})
	return admin, token, err
}

var timeNow = func() time.Time { return time.Now() }
