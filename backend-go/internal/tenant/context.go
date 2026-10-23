package tenant

import "context"

type ctxKey int

const key ctxKey = iota

// WithID guarda el tenant_id en el context.
func WithID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, key, id)
}

// ID obtiene el tenant_id del context.
func ID(ctx context.Context) (int64, bool) {
	v := ctx.Value(key)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// MustID obtiene el tenant_id o false si no existe.
func MustID(ctx context.Context) int64 {
	id, _ := ID(ctx)
	return id
}
