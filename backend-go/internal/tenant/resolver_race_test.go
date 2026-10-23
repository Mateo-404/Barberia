package tenant

import (
	"sync"
	"testing"
)

// TestResolverConcurrentAccess hammers the cache from many goroutines: before
// the RWMutex the map was written concurrently and Go aborted the process
// with "concurrent map writes" (a HIGH finding of the security audit).
func TestResolverConcurrentAccess(t *testing.T) {
	r := NewResolver(nil, "default")
	hosts := []string{"a.example", "b.example", "c.example"}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				h := hosts[(i+j)%len(hosts)]
				_, _ = r.cacheGet(h)
				r.mu.Lock()
				r.cache[h] = Tenant{ID: int64(i + 1), Slug: h}
				r.mu.Unlock()
				r.mu.RLock()
				_, _ = r.cacheDefault, r.defaultLoaded
				r.mu.RUnlock()
			}
		}(i)
	}
	wg.Wait()

	if got := len(r.cache); got != len(hosts) {
		t.Errorf("cache = %d entries, want %d", got, len(hosts))
	}
}
