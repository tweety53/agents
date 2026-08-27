package store_test

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

// adminDSN is the DSN used to create and drop per-test databases. It points
// at the flow-postgres compose stack from task 1, on host port 5433, and
// can be overridden for environments that run Postgres elsewhere.
func adminDSN() string {
	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
		return v
	}
	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
}

// testDSN returns the DSN for a database created by newTestDatabase.
//
// IT DERIVES FROM adminDSN RATHER THAN HARDCODING THE SAME CREDENTIALS AGAIN,
// so that FLOW_STATS_ADMIN_DSN moves BOTH connections together. Hardcoding them
// here made the override only half-work: the admin connection followed the
// environment while the per-test connection stayed pinned to the literal, so
// pointing the suite at a Postgres with different credentials failed on every
// test while the admin step succeeded.
//
// That is not hypothetical. During the flow rename this literal and adminDSN's
// were updated together to `flow:flow@.../flow`, which is what a fresh compose
// stack creates -- but the operator's own container is not renamed until the
// documented cutover step. With no working override, the whole suite reached an
// unreachable Postgres and SKIPPED: `internal/store` reported `ok` while running
// 4 tests and skipping 155. A green package that ran almost nothing is the exact
// failure this repository's REPRODUCE, DON'T READ rule exists to catch.
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

// newTestDatabase creates a uniquely-named, empty (unmigrated) database
// against the compose stack's PostgreSQL, and registers a cleanup that
// drops it once the test finishes. It skips cleanly, with a clear message,
// when the stack is not reachable.
func newTestDatabase(t *testing.T) string {
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

	return testDSN(dbName)
}

// newTestStore creates a uniquely-named database, applies every migration
// to it, and registers a cleanup that closes the store and drops the
// database once the test finishes. It skips cleanly, with a clear message,
// when the stack is not reachable.
func newTestStore(t *testing.T) *store.Store {
	t.Helper()

	dsn := newTestDatabase(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	if err := st.RunMigrations(ctx); err != nil {
		st.Close()
		t.Fatalf("run migrations: %v", err)
	}

	t.Cleanup(st.Close)

	return st
}
