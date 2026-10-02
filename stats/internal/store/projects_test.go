package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
