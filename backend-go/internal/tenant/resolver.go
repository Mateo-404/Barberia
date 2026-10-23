// Package tenant resuelve el tenant de cada request.
//
// Reglas (ADR-002 / SPEC §4):
//   - Host exacto en tenant_domains con verified_at IS NOT NULL.
//   - Subdominio de BASE_DOMAIN (env) → slug (simplificado si aplica).
//   - Solo con APP_ENV=dev: header X-Tenant-Slug.
//   - El tenant NUNCA se toma de una cabecera enviada por el cliente en prod.
//   - Tenant desconocido o suspended → error (middleware devuelve 404 genérico).
package tenant

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/jackc/pgx/v5"
)

// Tenant es el ámbito resuelto.
type Tenant struct {
	ID   int64
	Slug string
}

// Resolver resuelve y cachea tenants por Host.
type Resolver struct {
	q           *sqlc.Queries
	defaultSlug string

	mu            sync.RWMutex
	cache         map[string]Tenant
	misses        map[string]struct{}
	cacheDefault  Tenant
	defaultLoaded bool
}

// NewResolver construye el resolver.
func NewResolver(q *sqlc.Queries, defaultSlug string) *Resolver {
	return &Resolver{
		q: q, defaultSlug: defaultSlug,
		cache: map[string]Tenant{}, misses: map[string]struct{}{},
	}
}

// ErrTenantNoConfigurado indica que no existe el tenant por defecto.
var ErrTenantNoConfigurado = errors.New("tenant por defecto inexistente; ejecute las migraciones")
var ErrTenantDesconocido = errors.New("tenant desconocido")
var ErrTenantSuspendido = errors.New("tenant suspendido")

// Resolve devuelve el tenant para el Host de la request.
func (r *Resolver) Resolve(ctx context.Context, host string) (Tenant, error) {
	host = normalizarHost(host)
	if host == "" {
		return r.defaultTenant(ctx)
	}
	if t, ok := r.cacheGet(host); ok {
		return t, nil
	}
	if r.esMiss(host) {
		return r.defaultTenant(ctx)
	}

	// Intentar por tenant_domains verificado
	if r.q != nil {
		// Buscar dominio verificado
		row, err := r.q.GetTenantByDomain(ctx, host)
		if err == nil {
			res := Tenant{ID: row.ID, Slug: row.Slug}
			r.mu.Lock()
			r.cache[host] = res
			r.mu.Unlock()
			return res, nil
		}
	}

	t, err := r.q.GetTenantByHost(ctx, host)
	if err == nil {
		res := Tenant{ID: t.ID, Slug: t.Slug}
		r.mu.Lock()
		r.cache[host] = res
		r.mu.Unlock()
		return res, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, fmt.Errorf("resolver tenant por host: %w", err)
	}

	r.mu.Lock()
	r.misses[host] = struct{}{}
	r.mu.Unlock()
	return r.defaultTenant(ctx)
}

// ResolveFromRequest resuelve considerando header solo en dev
func (r *Resolver) ResolveFromRequest(ctx context.Context, req *http.Request) (Tenant, error) {
	if req != nil && os.Getenv("APP_ENV") == "dev" {
		if h := strings.TrimSpace(req.Header.Get("X-Tenant-Slug")); h != "" {
			if t, err := r.q.GetTenantBySlug(ctx, h); err == nil {
				return Tenant{ID: t.ID, Slug: t.Slug}, nil
			}
		}
	}
	host := ""
	if req != nil {
		host = req.Host
	}
	return r.Resolve(ctx, host)
}

func (r *Resolver) cacheGet(host string) (Tenant, bool) {
	r.mu.RLock()
	t, ok := r.cache[host]
	r.mu.RUnlock()
	return t, ok
}

func (r *Resolver) esMiss(host string) bool {
	r.mu.RLock()
	_, ok := r.misses[host]
	r.mu.RUnlock()
	return ok
}

// defaultTenant devuelve el tenant por defecto (Host ausente o desconocido).
func (r *Resolver) defaultTenant(ctx context.Context) (Tenant, error) {
	r.mu.RLock()
	if r.defaultLoaded {
		t := r.cacheDefault
		r.mu.RUnlock()
		return t, nil
	}
	r.mu.RUnlock()

	if r.q == nil {
		return Tenant{}, ErrTenantNoConfigurado
	}
	t, err := r.q.GetTenantBySlug(ctx, r.defaultSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrTenantNoConfigurado
	}
	if err != nil {
		return Tenant{}, fmt.Errorf("resolver tenant por defecto: %w", err)
	}
	res := Tenant{ID: t.ID, Slug: t.Slug}

	r.mu.Lock()
	r.cacheDefault = res
	r.defaultLoaded = true
	r.mu.Unlock()
	return res, nil
}

func normalizarHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if h, _, ok := strings.Cut(host, ":"); ok && !strings.Contains(host, "]") {
		host = h
	} else if strings.HasPrefix(host, "[") {
		if i := strings.Index(host, "]"); i >= 0 {
			host = host[1:i]
		}
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		// En desarrollo todos los frontends van contra el tenant por defecto.
		return ""
	}
	return host
}
