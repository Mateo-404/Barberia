package tenant

import "context"

type ctxKey int

const key ctxKey = iota

// WithID stores the tenant id in ctx.
func WithID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, key, id)
}

// ID returns the tenant id stored in ctx.
func ID(ctx context.Context) (int64, bool) {
	v := ctx.Value(key)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// MustID returns the tenant id, or 0 when there is none.
func MustID(ctx context.Context) int64 {
	id, _ := ID(ctx)
	return id
}
