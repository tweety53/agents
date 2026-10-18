package reconcile_test

import (
	"context"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/reconcile"
	"github.com/tweety53/agents/stats/internal/store"
)

// TestWatchReplaysOnHealthyTick pins KAN-827: a mark the CLI journalled
// during a brief HTTP-level outage, while the database stayed up, is
// replayed by Watch's next healthy tick -- not only on a ping that goes
// from failing to succeeding, which such an outage never produces.
func TestWatchReplaysOnHealthyTick(t *testing.T) {
	st := newTestStore(t)
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	// Stop Watch and wait for it before the store's cleanup closes it.
	defer func() { cancel(); <-done }()

	applied := make(chan reconcile.Result, 1)
	rec := reconcile.New(st, st, st, root, nil)
	go func() {
		defer close(done)
		rec.Watch(ctx, func(context.Context) error { return nil }, 10*time.Millisecond, func(r reconcile.Result, err error) {
			if err == nil && r.Applied > 0 {
				select {
				case applied <- r:
				default:
				}
			}
		})
	}()

	appendEntry(t, root, "proj-watch", "chg-watch", "STARTED", "2026-10-18T10:00:00Z", "flow-start")

	select {
	case <-applied:
	case <-time.After(5 * time.Second):
		t.Fatal("Watch never replayed a journal entry written while the store stayed healthy")
	}
	got, err := st.GetChange(context.Background(), "proj-watch", "chg-watch")
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.State != store.StateStarted {
		t.Fatalf("stored state = %q, want STARTED", got.State)
	}
	if n := pendingCount(t, root, "proj-watch", "chg-watch"); n != 0 {
		t.Fatalf("pending entries after Watch replay = %d, want 0", n)
	}
}
