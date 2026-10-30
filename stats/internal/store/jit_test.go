package store_test

import (
	"context"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestOpenTurnsJITOff asserts every pooled connection runs with JIT off: the
// live state board's query is costed above jit_above_cost, and compiling it
// took ~1s of its ~1.3s on every poll.
func TestOpenTurnsJITOff(t *testing.T) {
	st := newTestStore(t)
	var jit string
	if err := store.PoolOf(st).QueryRow(context.Background(), `SHOW jit`).Scan(&jit); err != nil {
		t.Fatal(err)
	}
	if jit != "off" {
		t.Fatalf("jit = %q, want off", jit)
	}
}
