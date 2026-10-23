package auth

import (
	"sync"
	"time"
)

// LoginLimiter limits attempts per key (IP) in one-minute windows. It is
// in-memory and per process: fine for one instance; with several replicas
// the effective limit multiplies (documented).
type LoginLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	now     func() time.Time
	entries map[string]*window
}

type window struct {
	start time.Time
	count int
}

// NewLoginLimiter allows limit attempts per minute.
func NewLoginLimiter(limit int) *LoginLimiter {
	return &LoginLimiter{limit: limit, window: time.Minute, now: time.Now, entries: map[string]*window{}}
}

// Allow reports whether the key may try now.
func (l *LoginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	v, ok := l.entries[key]
	if !ok || now.Sub(v.start) >= l.window {
		l.entries[key] = &window{start: now, count: 1}
		l.prune(now)
		return true
	}
	if v.count >= l.limit {
		return false
	}
	v.count++
	return true
}

// Reset forgets the key (e.g. after a successful login).
func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

// prune keeps the map from growing without bound.
func (l *LoginLimiter) prune(now time.Time) {
	if len(l.entries) < 1024 {
		return
	}
	for k, v := range l.entries {
		if now.Sub(v.start) >= l.window {
			delete(l.entries, k)
		}
	}
}
