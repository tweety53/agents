package sweep_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/storetest"
)

// TestMain drops the package's template database once every test has run.
func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }

// newTestStore opens a store on a throwaway, migrated database -- a copy of
// the package's one migrated template (internal/storetest) -- and registers
// its Close and the database's drop as test cleanup. It skips cleanly when
// the stack is not reachable.
func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	return storetest.NewStore(t)
}

// seedChange puts a minimal, valid change under projectKey/name so a
// stage run has somewhere to attach -- exactly what BeginStage requires.
func seedChange(t *testing.T, st *store.Store, projectKey, name string) {
	t.Helper()
	ctx := context.Background()
	c := store.Change{
		ProjectKey:       projectKey,
		MainCheckoutPath: "/Users/tweety53/Projects/" + projectKey,
		Name:             name,
		State:            store.StateStarted,
		UpdatedAt:        time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC),
		UpdatedBy:        "sweep-test",
	}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("seed change %s/%s: %v", projectKey, name, err)
	}
}

func ptr[T any](v T) *T { return &v }
