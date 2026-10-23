package auth

import (
	"testing"
	"time"
)

// nuevaConReloj permite controlar la ventana del limitador en tests.
func nuevaConReloj(limit int, reloj *time.Time) *LoginLimiter {
	l := NewLoginLimiter(limit)
	l.now = func() time.Time { return *reloj }
	return l
}

func TestLoginLimiterBloqueaTrasLimite(t *testing.T) {
	ahora := time.Now()
	l := nuevaConReloj(3, &ahora)
	for i := 0; i < 3; i++ {
		if !l.Allow("1.1.1.1") {
			t.Fatalf("intento %d debería permitirse", i+1)
		}
	}
	if l.Allow("1.1.1.1") {
		t.Fatal("el cuarto intento debería bloquearse")
	}
}

func TestLoginLimiterClavesIndependientes(t *testing.T) {
	ahora := time.Now()
	l := nuevaConReloj(1, &ahora)
	if !l.Allow("1.1.1.1") || l.Allow("1.1.1.1") {
		t.Fatal("misma clave: solo 1 permitido")
	}
	if !l.Allow("2.2.2.2") {
		t.Fatal("otra clave no debería verse afectada")
	}
}

func TestLoginLimiterVentanaSeResetea(t *testing.T) {
	ahora := time.Now()
	l := nuevaConReloj(1, &ahora)
	if !l.Allow("1.1.1.1") || l.Allow("1.1.1.1") {
		t.Fatal("setup: solo 1 permitido por ventana")
	}
	ahora = ahora.Add(61 * time.Second)
	if !l.Allow("1.1.1.1") {
		t.Fatal("tras la ventana debería permitirse de nuevo")
	}
}

func TestLoginLimiterResetOlvida(t *testing.T) {
	ahora := time.Now()
	l := nuevaConReloj(2, &ahora)
	_ = l.Allow("1.1.1.1")
	l.Reset("1.1.1.1")
	if !l.Allow("1.1.1.1") {
		t.Fatal("tras Reset debería permitirse el primer intento")
	}
	if !l.Allow("1.1.1.1") {
		t.Fatal("tras Reset debería permitirse el segundo intento (ventana nueva)")
	}
}

// TestLoginLimiterLimpieza evita crecimiento sin límite del mapa: al superar
// 1024 claves, las ventanas vencidas se eliminan.
func TestLoginLimiterLimpieza(t *testing.T) {
	ahora := time.Now()
	l := nuevaConReloj(1, &ahora)
	for i := 0; i < 1025; i++ {
		l.Allow("clave-vieja")
		ahora = ahora.Add(time.Second)
	}
	// Tras llenar el mapa con claves vencidas, una clave nueva dispara la
	// limpieza: solo debe quedar (más o menos) la ventana activa.
	l.Allow("nueva")
	l.mu.Lock()
	restantes := len(l.entradas)
	l.mu.Unlock()
	if restantes > 2 {
		t.Fatalf("la limpieza no se ejecutó: quedan %d entradas", restantes)
	}
}
