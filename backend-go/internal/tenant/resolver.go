// Package tenant resolves the tenant of each request (ADR-002 / SPEC §4):
//   - exact Host in tenant_domains with verified_at IS NOT NULL;
//   - otherwise exact match on tenants.host;
//   - otherwise (and for localhost) the default tenant.
//
// The tenant is never taken from a client-supplied header.
package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/jackc/pgx/v5"
)

// Tenant is the resolved scope.
type Tenant struct {
	ID   int64
	Slug string
}

// Resolver resolves tenants by Host and caches the result.
type Resolver struct {
	q           *sqlc.Queries
	defaultSlug string

	mu            sync.RWMutex
	cache         map[string]Tenant
	misses        map[string]struct{}
	cacheDefault  Tenant
	defaultLoaded bool
}

// NewResolver builds the resolver.
func NewResolver(q *sqlc.Queries, defaultSlug string) *Resolver {
	return &Resolver{
		q: q, defaultSlug: defaultSlug,
		cache: map[string]Tenant{}, misses: map[string]struct{}{},
	}
}

// ErrDefaultTenantMissing means the default tenant does not exist yet.
var ErrDefaultTenantMissing = errors.New("default tenant missing; run the migrations")

// Resolve returns the tenant for a request Host.
func (r *Resolver) Resolve(ctx context.Context, host string) (Tenant, error) {
	host = normalizeHost(host)
	if host == "" {
		return r.defaultTenant(ctx)
	}
	if t, ok := r.cacheGet(host); ok {
		return t, nil
	}
	if r.isMiss(host) {
		return r.defaultTenant(ctx)
	}

	if r.q != nil {
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
		return Tenant{}, fmt.Errorf("resolve tenant by host: %w", err)
	}

	r.mu.Lock()
	r.misses[host] = struct{}{}
	r.mu.Unlock()
	return r.defaultTenant(ctx)
}

func (r *Resolver) cacheGet(host string) (Tenant, bool) {
	r.mu.RLock()
	t, ok := r.cache[host]
	r.mu.RUnlock()
	return t, ok
}

func (r *Resolver) isMiss(host string) bool {
	r.mu.RLock()
	_, ok := r.misses[host]
	r.mu.RUnlock()
	return ok
}

// defaultTenant returns the default tenant (missing or unknown Host).
func (r *Resolver) defaultTenant(ctx context.Context) (Tenant, error) {
	r.mu.RLock()
	if r.defaultLoaded {
		t := r.cacheDefault
		r.mu.RUnlock()
		return t, nil
	}
	r.mu.RUnlock()

	if r.q == nil {
		return Tenant{}, ErrDefaultTenantMissing
	}
	t, err := r.q.GetTenantBySlug(ctx, r.defaultSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrDefaultTenantMissing
	}
	if err != nil {
		return Tenant{}, fmt.Errorf("resolve default tenant: %w", err)
	}
	res := Tenant{ID: t.ID, Slug: t.Slug}

	r.mu.Lock()
	r.cacheDefault = res
	r.defaultLoaded = true
	r.mu.Unlock()
	return res, nil
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if h, _, ok := strings.Cut(host, ":"); ok && !strings.Contains(host, "]") {
		host = h
	} else if strings.HasPrefix(host, "[") {
		if i := strings.Index(host, "]"); i >= 0 {
			host = host[1:i]
		}
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		// In development every frontend talks to the default tenant.
		return ""
	}
	return host
}
