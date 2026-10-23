package auth

import (
	"testing"
	"time"
)

// withClock lets tests control the limiter window.
func withClock(limit int, clock *time.Time) *LoginLimiter {
	l := NewLoginLimiter(limit)
	l.now = func() time.Time { return *clock }
	return l
}

func TestLoginLimiterBlocksAfterLimit(t *testing.T) {
	now := time.Now()
	l := withClock(3, &now)
	for i := 0; i < 3; i++ {
		if !l.Allow("1.1.1.1") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if l.Allow("1.1.1.1") {
		t.Fatal("the fourth attempt should be blocked")
	}
}

func TestLoginLimiterKeysAreIndependent(t *testing.T) {
	now := time.Now()
	l := withClock(1, &now)
	if !l.Allow("1.1.1.1") || l.Allow("1.1.1.1") {
		t.Fatal("same key: only 1 allowed")
	}
	if !l.Allow("2.2.2.2") {
		t.Fatal("another key must not be affected")
	}
}

func TestLoginLimiterWindowResets(t *testing.T) {
	now := time.Now()
	l := withClock(1, &now)
	if !l.Allow("1.1.1.1") || l.Allow("1.1.1.1") {
		t.Fatal("setup: only 1 allowed per window")
	}
	now = now.Add(61 * time.Second)
	if !l.Allow("1.1.1.1") {
		t.Fatal("after the window it should be allowed again")
	}
}

func TestLoginLimiterResetForgets(t *testing.T) {
	now := time.Now()
	l := withClock(2, &now)
	_ = l.Allow("1.1.1.1")
	l.Reset("1.1.1.1")
	if !l.Allow("1.1.1.1") {
		t.Fatal("after Reset the first attempt should be allowed")
	}
	if !l.Allow("1.1.1.1") {
		t.Fatal("after Reset the second attempt should be allowed (new window)")
	}
}

// TestLoginLimiterPrunes keeps the map bounded: past 1024 keys, expired
// windows are removed.
func TestLoginLimiterPrunes(t *testing.T) {
	now := time.Now()
	l := withClock(1, &now)
	for i := 0; i < 1025; i++ {
		l.Allow("old-key")
		now = now.Add(time.Second)
	}
	// With the map full of expired keys, a new key triggers pruning: roughly
	// only the active window must remain.
	l.Allow("new")
	l.mu.Lock()
	remaining := len(l.entries)
	l.mu.Unlock()
	if remaining > 2 {
		t.Fatalf("pruning did not run: %d entries left", remaining)
	}
}
