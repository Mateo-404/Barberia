// Package administradores implementa el CRUD, el login y el seed inicial.
package administradores

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

// Administrador es el modelo de dominio (sin hash expuesto).
type Administrador struct {
	ID       int64
	Nombre   string
	Apellido string
	Email    string
	// Rol del administrador: owner | admin | staff (SPEC §8).
	Rol string
}

// Input son los datos de creación.
type Input struct {
	Nombre      string
	Apellido    string
	Email       string
	Contrasenia string
	// Rol del administrador: owner | admin | staff (SPEC §8).
	Rol string
}

// Patch contiene solo los campos presentes en el body.
// Contrasenia se hashea antes de persistir (ADR-004).
type Patch struct {
	Nombre      *string
	Apellido    *string
	Email       *string
	Contrasenia *string
	// Rol opcional: owner | admin | staff (SPEC §8).
	Rol *string
}

// Service agrupa las operaciones de administradores.
type Service struct {
	pool   *pgxpool.Pool
	q      *sqlc.Queries
	jwt    *auth.JWTManager
	logger *slog.Logger
}

// New construye el servicio.
func New(pool *pgxpool.Pool, q *sqlc.Queries, jwt *auth.JWTManager, logger *slog.Logger) *Service {
	return &Service{pool: pool, q: q, jwt: jwt, logger: logger}
}

func aDominio(id int64, nombre, apellido, email *string, rol string) Administrador {
	if rol == "" {
		rol = auth.RolOwner
	}
	return Administrador{ID: id, Nombre: deref(nombre), Apellido: deref(apellido), Email: deref(email), Rol: rol}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// List devuelve todos los administradores del tenant.
func (s *Service) List(ctx context.Context, tenantID int64) ([]Administrador, error) {
	var out []Administrador
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		rows, err := q.ListAdministradores(ctx, tenantID)
		if err != nil {
			return err
		}
		out = make([]Administrador, 0, len(rows))
		for _, r := range rows {
			out = append(out, aDominio(r.ID, r.Nombre, r.Apellido, r.Email, r.Role))
		}
		return nil
	})
	return out, err
}

// Get devuelve un administrador o nil.
func (s *Service) Get(ctx context.Context, tenantID, id int64) (*Administrador, error) {
	var out *Administrador
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAdministrador(ctx, sqlc.GetAdministradorParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			out = nil
			return nil
		}
		if err != nil {
			return err
		}
		a := aDominio(r.ID, r.Nombre, r.Apellido, r.Email, r.Role)
		out = &a
		return nil
	})
	return out, err
}

// Create crea un administrador hasheando la contraseña.
func (s *Service) Create(ctx context.Context, tenantID int64, in Input) (*Administrador, error) {
	var out *Administrador
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		existe, err := existsByEmailTx(ctx, q, tenantID, in.Email)
		if err != nil {
			return err
		}
		if existe {
			return yaExiste(in.Email)
		}
		hash, err := auth.HashPassword(in.Contrasenia)
		if err != nil {
			return err
		}
		rol := in.Rol
		if rol == "" {
			rol = auth.RolOwner
		}
		r, err := q.CreateAdministrador(ctx, sqlc.CreateAdministradorParams{
			TenantID: tenantID, Nombre: &in.Nombre, Apellido: &in.Apellido,
			Email: &in.Email, Contrasenia: &hash, Role: rol,
		})
		if pgutil.IsUniqueViolation(err) {
			return yaExiste(in.Email)
		}
		if err != nil {
			return err
		}
		a := aDominio(r.ID, r.Nombre, r.Apellido, r.Email, r.Role)
		out = &a
		return nil
	})
	return out, err
}

// CreateAll crea todos en una transacción; 409 al primer email duplicado.
func (s *Service) CreateAll(ctx context.Context, tenantID int64, ins []Input) ([]Administrador, error) {
	var out []Administrador
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		out = make([]Administrador, 0, len(ins))
		for _, in := range ins {
			if _, err := q.GetAdministradorByEmail(ctx, sqlc.GetAdministradorByEmailParams{TenantID: tenantID, Email: &in.Email}); err == nil {
				return yaExiste(in.Email)
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			hash, err := auth.HashPassword(in.Contrasenia)
			if err != nil {
				return err
			}
			rol := in.Rol
			if rol == "" {
				rol = auth.RolOwner
			}
			r, err := q.CreateAdministrador(ctx, sqlc.CreateAdministradorParams{
				TenantID: tenantID, Nombre: &in.Nombre, Apellido: &in.Apellido,
				Email: &in.Email, Contrasenia: &hash, Role: rol,
			})
			if pgutil.IsUniqueViolation(err) {
				return yaExiste(in.Email)
			}
			if err != nil {
				return err
			}
			out = append(out, aDominio(r.ID, r.Nombre, r.Apellido, r.Email, r.Role))
		}
		return nil
	})
	return out, err
}

// Update aplica copyProperties sin nulos; hashea contrasenia si viene (ADR-004).
func (s *Service) Update(ctx context.Context, tenantID, id int64, p Patch) (*Administrador, error) {
	var out *Administrador
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAdministrador(ctx, sqlc.GetAdministradorParams{TenantID: tenantID, ID: id})
		if errors.Is(err, pgx.ErrNoRows) {
			out = nil
			return nil
		}
		if err != nil {
			return err
		}
		actual := aDominio(r.ID, r.Nombre, r.Apellido, r.Email, r.Role)
		hashActual := deref(r.Contrasenia)
		if p.Nombre != nil {
			actual.Nombre = *p.Nombre
		}
		if p.Apellido != nil {
			actual.Apellido = *p.Apellido
		}
		if p.Email != nil {
			actual.Email = *p.Email
		}
		if p.Contrasenia != nil {
			if hashActual, err = auth.HashPassword(*p.Contrasenia); err != nil {
				return err
			}
		}
		rol := actual.Rol
		if p.Rol != nil {
			rol = *p.Rol
		}
		updated, err := q.UpdateAdministrador(ctx, sqlc.UpdateAdministradorParams{
			TenantID: tenantID, ID: id, Nombre: &actual.Nombre, Apellido: &actual.Apellido,
			Email: &actual.Email, Contrasenia: &hashActual, Role: rol,
		})
		if pgutil.IsUniqueViolation(err) {
			return yaExiste(actual.Email)
		}
		if err != nil {
			return err
		}
		a := aDominio(updated.ID, updated.Nombre, updated.Apellido, updated.Email, updated.Role)
		out = &a
		return nil
	})
	return out, err
}

// Delete elimina un administrador; false si no existía.
func (s *Service) Delete(ctx context.Context, tenantID, id int64) (bool, error) {
	var ok bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		n, err := q.DeleteAdministrador(ctx, sqlc.DeleteAdministradorParams{TenantID: tenantID, ID: id})
		if err != nil {
			return err
		}
		ok = n > 0
		return nil
	})
	return ok, err
}

// Exists indica si el administrador existe en el tenant.
func (s *Service) Exists(ctx context.Context, tenantID, id int64) (bool, error) {
	var exists bool
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		_, err := q.GetAdministrador(ctx, sqlc.GetAdministradorParams{TenantID: tenantID, ID: id})
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

// Page devuelve una página 0-based (sin hash, ADR-004).
func (s *Service) Page(ctx context.Context, tenantID int64, page, size int) ([]Administrador, int64, error) {
	var out []Administrador
	var total int64
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		t, err := q.CountAdministradores(ctx, tenantID)
		if err != nil {
			return err
		}
		total = t
		rows, err := q.PageAdministradores(ctx, sqlc.PageAdministradoresParams{
			TenantID: tenantID, Limite: int32(size), Offset: int32(page * size),
		})
		if err != nil {
			return err
		}
		out = make([]Administrador, 0, len(rows))
		for _, r := range rows {
			out = append(out, aDominio(r.ID, r.Nombre, r.Apellido, r.Email, r.Role))
		}
		return nil
	})
	return out, total, err
}

// Login valida credenciales y emite un JWT.
func (s *Service) Login(ctx context.Context, tenantID int64, email, contrasenia string) (Administrador, string, error) {
	var admin Administrador
	var token string
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		r, err := q.GetAdministradorByEmail(ctx, sqlc.GetAdministradorByEmailParams{TenantID: tenantID, Email: &email})
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.New(apperr.KindInvalidCredentials, "Credenciales inválidas")
		}
		if err != nil {
			return err
		}
		ok, err := auth.VerifyPassword(contrasenia, deref(r.Contrasenia))
		if err != nil || !ok {
			return apperr.New(apperr.KindInvalidCredentials, "Credenciales inválidas")
		}
		admin = aDominio(r.ID, r.Nombre, r.Apellido, r.Email, r.Role)
		t, err := s.jwt.Issue(admin.ID, admin.Email, tenantID, admin.Rol)
		if err != nil {
			return err
		}
		token = t
		return nil
	})
	return admin, token, err
}

// Seed crea el administrador inicial si no hay ninguno en el tenant.
func (s *Service) Seed(ctx context.Context, tenantID int64, email, password, nombre, apellido string) error {
	ctx = tenant.WithID(ctx, tenantID)
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		total, err := q.CountAdministradores(ctx, tenantID)
		if err != nil {
			return err
		}
		if total > 0 {
			s.logger.Info("ya existen administradores en el tenant; se omite seed", "tenant_id", tenantID)
			return nil
		}
		if len(password) < 8 {
			return errors.New("ADMIN_SEED_PASSWORD debe tener al menos 8 caracteres")
		}
		hash, err := auth.HashPassword(password)
		if err != nil {
			return err
		}
		if _, err := q.CreateAdministrador(ctx, sqlc.CreateAdministradorParams{
			TenantID: tenantID, Nombre: &nombre, Apellido: &apellido, Email: &email, Contrasenia: &hash,
			Role: auth.RolOwner,
		}); err != nil {
			return err
		}
		s.logger.Info("administrador semilla creado", "email", email)
		return nil
	})
	return err
}

func existsByEmailTx(ctx context.Context, q *sqlc.Queries, tenantID int64, email string) (bool, error) {
	_, err := q.GetAdministradorByEmail(ctx, sqlc.GetAdministradorByEmailParams{TenantID: tenantID, Email: &email})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func yaExiste(email string) error {
	return apperr.New(apperr.KindAlreadyExists, "Ya existe un administrador con el email %s", email)
}
