package settings

import (
	"math/rand"
	"testing"
)

func TestOnPrimaryProperty(t *testing.T) {
	r := rand.New(rand.NewSource(12345))
	for i := 0; i < 10000; i++ {
		rr := r.Intn(256)
		gg := r.Intn(256)
		bb := r.Intn(256)
		bg := "#" + tohex(rr) + tohex(gg) + tohex(bb)
		op := OnPrimary(bg)
		rat, _ := Contrast(op, bg)
		if rat < 4.3-1e-6 {
			t.Fatalf("bad ratio %f bg %s op %s", rat, bg, op)
		}
	}
}

func tohex(x int) string {
	const h = "0123456789abcdef"
	return string([]byte{h[x/16], h[x%16]})
}
