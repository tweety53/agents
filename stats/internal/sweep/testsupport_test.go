package sweep_test

import (
	"context"
	"fmt"
	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/store"
)

// adminDSN is the DSN used to create and drop per-test databases -- the
// same dedicated flow-postgres compose stack internal/store's own tests
// use (task 1), overridable for environments that run Postgres elsewhere.
func adminDSN() string {
	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
		return v
	}
	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
}

// It derives from adminDSN rather than repeating the credentials, so that
// FLOW_STATS_ADMIN_DSN moves both connections together. Hardcoding them here
// made the override only half-work -- the admin connection followed the
// environment while the per-test one stayed pinned, so a Postgres with
// different credentials failed every test while the admin step succeeded. See
// internal/store/testsupport_test.go's testDSN for the measured case.
func testDSN(dbName string) string {
	out, err := dsnutil.ForDatabase(adminDSN(), dbName)
	if err != nil {
		// Panic rather than return a best-effort string. This helper's
		// predecessor answered confidently when it could not do the job and
		// silently handed back a corrupted DSN; a test that then connects
		// reports whatever that connection says, which is the wrong question
		// answered convincingly.
		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
	}
	return out
}

// newTestStore creates a uniquely-named, migrated database against the
// compose stack, returning a *store.Store whose Close and whose database
// drop are both registered as test cleanup. It skips cleanly, with a clear
// message, when the stack is not reachable -- mirroring
// internal/store/testsupport_test.go's own helper and
// internal/reconcile/testsupport_test.go's copy of it: sweep's own tests
// need the same real-Postgres guarantee design.md's testing strategy
// requires ("store is tested against a real PostgreSQL ... never a mock")
// for the concurrency behaviour under test, which a mock could not
// reproduce.
func newTestStore(t *testing.T) *store.Store {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adminPool, err := pgxpool.New(ctx, adminDSN())
	if err != nil {
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}

	dbName := fmt.Sprintf("flow_test_%d_%d", os.Getpid(), time.Now().UnixNano())
	ident := pgx.Identifier{dbName}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		adminPool.Close()
		t.Fatalf("create test database %s: %v", dbName, err)
	}
	adminPool.Close()

	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer dropCancel()

		dropPool, err := pgxpool.New(dropCtx, adminDSN())
		if err != nil {
			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
			return
		}
		defer dropPool.Close()

		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
			t.Logf("drop test database %s: %v", dbName, err)
		}
	})

	dsn := testDSN(dbName)

	openCtx, openCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer openCancel()

	st, err := store.Open(openCtx, dsn)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	if err := st.RunMigrations(openCtx); err != nil {
		st.Close()
		t.Fatalf("run migrations: %v", err)
	}
	t.Cleanup(st.Close)

	return st
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
