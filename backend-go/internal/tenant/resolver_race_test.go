package tenant

import (
	"sync"
	"testing"
)

// TestResolverConcurrenteSinRace ejerce la caché desde muchas goroutines: antes
// del RWMutex el map se escribía en paralelo y Go abortaba el proceso por
// "concurrent map writes" (fallo ALTO de la auditoría de seguridad).
func TestResolverConcurrenteSinRace(t *testing.T) {
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
		t.Errorf("caché = %d entradas, quiero %d", got, len(hosts))
	}
}
