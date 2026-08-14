package store_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestMigrationsAreIdempotent asserts that running the migration runner a
// second time against a database it has already migrated is a no-op: no
// error, and the schema_migrations bookkeeping is not duplicated.
func TestMigrationsAreIdempotent(t *testing.T) {
	st := newTestStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("second RunMigrations call: %v", err)
	}

	count, err := st.MigrationCount(ctx)
	if err != nil {
		t.Fatalf("count schema_migrations rows: %v", err)
	}
	want, err := store.EmbeddedMigrationCount()
	if err != nil {
		t.Fatalf("count embedded migration files: %v", err)
	}
	if count != want {
		t.Fatalf("schema_migrations has %d rows after two runs, want %d (one per embedded migration file)", count, want)
	}
}

// TestConcurrentRunMigrationsDoesNotCollide reproduces the failure reported
// against a fresh database: five callers, each with their own Store and
// pool, all invoking RunMigrations at once against the same brand-new
// database. Without an advisory lock serializing the run, two sessions race
// the same CREATE TABLE/CREATE FUNCTION DDL and one fails with a duplicate
// pg_type/pg_class catalog entry.
func TestConcurrentRunMigrationsDoesNotCollide(t *testing.T) {
	dsn := newTestDatabase(t)

	const callers = 5

	var wg sync.WaitGroup
	errs := make([]error, callers)

	for i := range callers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			st, err := store.Open(ctx, dsn)
			if err != nil {
				errs[i] = err
				return
			}
			defer st.Close()

			errs[i] = st.RunMigrations(ctx)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("caller %d: RunMigrations: %v", i, err)
		}
	}
}
