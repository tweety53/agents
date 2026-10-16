package reconcile_test

import (
	"os"
	"testing"

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
