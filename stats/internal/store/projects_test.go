package store_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

func TestProjectRootsListsRegisteredCheckouts(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	now := time.Now()
	for _, c := range []store.Change{
		{ProjectKey: "ws-b-key", MainCheckoutPath: "/ws/beta", Name: "change-b", State: store.StateStarted, UpdatedAt: now, UpdatedBy: "test"},
		{ProjectKey: "ws-a-key", MainCheckoutPath: "/ws/alpha", Name: "change-a", State: store.StateStarted, UpdatedAt: now, UpdatedBy: "test"},
	} {
		if err := st.PutChange(ctx, c); err != nil {
			t.Fatalf("PutChange %s: %v", c.ProjectKey, err)
		}
	}

	roots, err := st.ProjectRoots(ctx)
	if err != nil {
		t.Fatalf("ProjectRoots: %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("roots = %d, want 2", len(roots))
	}
	if roots[0].ProjectKey != "ws-a-key" || roots[0].MainCheckoutPath != "/ws/alpha" {
		t.Errorf("roots[0] = %+v", roots[0])
	}
	if roots[1].ProjectKey != "ws-b-key" || roots[1].MainCheckoutPath != "/ws/beta" {
		t.Errorf("roots[1] = %+v", roots[1])
	}
}

func TestProjectRootsKeepsFirstCheckoutPath(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	now := time.Now()
	first := store.Change{ProjectKey: "ws-key", MainCheckoutPath: "/ws/first", Name: "change-1", State: store.StateStarted, UpdatedAt: now, UpdatedBy: "test"}
	if err := st.PutChange(ctx, first); err != nil {
		t.Fatalf("PutChange first: %v", err)
	}
	second := first
	second.Name = "change-2"
	second.MainCheckoutPath = "/ws/second"
	second.UpdatedAt = now.Add(time.Minute)
	if err := st.PutChange(ctx, second); err != nil {
		t.Fatalf("PutChange second: %v", err)
	}

	roots, err := st.ProjectRoots(ctx)
	if err != nil {
		t.Fatalf("ProjectRoots: %v", err)
	}
	if len(roots) != 1 || roots[0].MainCheckoutPath != "/ws/first" {
		t.Fatalf("roots = %+v, want the first checkout kept", roots)
	}
}

func TestProjectRootsOnEmptyStore(t *testing.T) {
	st := newTestStore(t)
	roots, err := st.ProjectRoots(context.Background())
	if err != nil {
		t.Fatalf("ProjectRoots: %v", err)
	}
	if len(roots) != 0 {
		t.Fatalf("roots = %v, want none", roots)
	}
	if errors.Is(err, store.ErrChangeNotFound) {
		t.Fatal("a plain read never carries ErrChangeNotFound")
	}
}

// TestEnsureProjectLetsARecordWriteThrough pins KAN-914: a fresh database
// refuses a record write for a project it has no row for, and
// EnsureProject's row lets it through; a second call and an existing row
// are left as they are, and an empty path never creates one.
func TestEnsureProjectLetsARecordWriteThrough(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	run := records.GuardRun{Guard: "check-x", Worktree: "/ws/alpha", Outcome: "clear"}

	if _, err := st.RecordGuardRun(ctx, "ws-alpha-key", run); !errors.Is(err, store.ErrProjectNotFound) {
		t.Fatalf("RecordGuardRun before EnsureProject: err = %v, want ErrProjectNotFound", err)
	}
	if err := st.EnsureProject(ctx, "ws-alpha-key", ""); !errors.Is(err, store.ErrInvalidMainCheckoutPath) {
		t.Fatalf("EnsureProject with an empty path: err = %v, want ErrInvalidMainCheckoutPath", err)
	}
	for range 2 {
		if err := st.EnsureProject(ctx, "ws-alpha-key", "/ws/alpha"); err != nil {
			t.Fatalf("EnsureProject: %v", err)
		}
	}
	if err := st.EnsureProject(ctx, "ws-alpha-key", "/ws/other"); err != nil {
		t.Fatalf("EnsureProject on an existing row: %v", err)
	}
	if _, err := st.RecordGuardRun(ctx, "ws-alpha-key", run); err != nil {
		t.Fatalf("RecordGuardRun after EnsureProject: %v", err)
	}
	roots, err := st.ProjectRoots(ctx)
	if err != nil {
		t.Fatalf("ProjectRoots: %v", err)
	}
	want := []store.ProjectRoot{{ProjectKey: "ws-alpha-key", MainCheckoutPath: "/ws/alpha"}}
	if !reflect.DeepEqual(roots, want) {
		t.Fatalf("ProjectRoots = %+v, want %+v", roots, want)
	}
}
