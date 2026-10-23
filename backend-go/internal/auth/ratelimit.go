package auth

import (
	"sync"
	"time"
)

// LoginLimiter limita intentos de login por clave (IP) en ventanas de un
// minuto. Es en memoria y por proceso: suficiente para una instancia; con
// varias réplicas el límite efectivo se multiplica (documentado).
type LoginLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	now      func() time.Time
	entradas map[string]*ventana
}

type ventana struct {
	inicio time.Time
	count  int
}

// NewLoginLimiter crea el limitador con N intentos por minuto.
func NewLoginLimiter(limit int) *LoginLimiter {
	return &LoginLimiter{limit: limit, window: time.Minute, now: time.Now, entradas: map[string]*ventana{}}
}

// Allow indica si la clave puede intentar login ahora.
func (l *LoginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	v, ok := l.entradas[key]
	if !ok || now.Sub(v.inicio) >= l.window {
		l.entradas[key] = &ventana{inicio: now, count: 1}
		l.limpieza(now)
		return true
	}
	if v.count >= l.limit {
		return false
	}
	v.count++
	return true
}

// Reset olvida la clave (p. ej. tras login exitoso).
func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entradas, key)
}

// limpieza evita crecimiento sin límite del mapa.
func (l *LoginLimiter) limpieza(now time.Time) {
	if len(l.entradas) < 1024 {
		return
	}
	for k, v := range l.entradas {
		if now.Sub(v.inicio) >= l.window {
			delete(l.entradas, k)
		}
	}
}
