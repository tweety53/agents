package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/store"
)

// adminDSN is the DSN used to create and drop per-test databases. It points
// at the myflow-postgres compose stack from task 1, on host port 5433, and
// can be overridden for environments that run Postgres elsewhere.
func adminDSN() string {
	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
		return v
	}
	return "postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable"
}

// testDSN returns the DSN for a database created by newTestDatabase.
func testDSN(dbName string) string {
	return fmt.Sprintf("postgres://myflow:myflow@localhost:5433/%s?sslmode=disable", dbName)
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
		t.Skipf("myflow-postgres compose stack not reachable: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("myflow-postgres compose stack not reachable: %v", err)
	}

	dbName := fmt.Sprintf("myflow_test_%d_%d", os.Getpid(), time.Now().UnixNano())
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
