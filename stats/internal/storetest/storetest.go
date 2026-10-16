// Package storetest gives tests a throwaway, migrated database on the
// flow-postgres compose stack. It is imported by _test.go files only.
//
// A migrated database is a copy of one template database the package
// process migrates once, never a fresh database migrated per test. Migrating
// per test was the contention the suite flaked on: every test of every
// package under `go test ./...` ran all of internal/store's migrations --
// one transaction and one commit flush each -- against the one shared
// server, and under that load a 3-second setup deadline ran out mid
// migration ("run migrations: ... timeout: context deadline exceeded").
// CREATE DATABASE ... TEMPLATE copies the migrated files instead, so a test
// pays one database copy and the server sees one migration run per
// package process.
//
// Setup work -- creating, migrating, copying, opening and dropping
// databases -- carries no wall-clock bound: how long it takes depends on
// server load the test does not control, so any fixed deadline races it.
// Only the reachability probe is bounded, by reachTimeout; a hang in the
// work falls to `go test -timeout`, which reports it with a goroutine dump.
package storetest

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
	"github.com/tweety53/agents/stats/internal/store"
)

// reachTimeout bounds the reachability probe alone: a ping on a reachable
// server answers in milliseconds and a down stack refuses at once, so the
// bound does not depend on load. It never bounds setup work.
const reachTimeout = 3 * time.Second

// AdminDSN is the DSN used to create and drop test databases: the
// flow-postgres compose stack on host port 5433, overridable through
// FLOW_STATS_ADMIN_DSN for environments that run Postgres elsewhere.
func AdminDSN() string {
	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
		return v
	}
	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
}

// DSN returns AdminDSN with its database replaced by dbName, so
// FLOW_STATS_ADMIN_DSN moves the admin and the per-test connections
// together. It panics rather than hand back a corrupted DSN: a test that
// then connects reports whatever that connection says.
func DSN(dbName string) string {
	out, err := dsnutil.ForDatabase(AdminDSN(), dbName)
	if err != nil {
		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
	}
	return out
}

var (
	templateOnce sync.Once
	templateName string
	templateErr  error
	unreachable  error
)

// Main runs the package's tests, then drops the template, if a test built
// one, and every test database this process created whose cleanup drop
// failed and would otherwise outlive the run. Databases are matched by this
// process's PID, which is unique among processes on one host; two runs in
// separate PID namespaces sharing one server through FLOW_STATS_ADMIN_DSN
// could collide. Every
// package calling NewStore, NewMigratedDatabase or NewDatabase calls it from
// TestMain: `os.Exit(storetest.Main(m))`.
func Main(m *testing.M) int {
	code := m.Run()
	if err := dropLeftoverDatabases(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "storetest: drop leftover test databases: %v\n", err)
	}
	return code
}

// dropLeftoverDatabases drops every test and template database this
// process created that still exists. It runs once no test is left to use one. An unreachable
// stack means nothing was created, so it is not an error.
func dropLeftoverDatabases(ctx context.Context) error {
	admin, err := connectAdmin(ctx)
	if err != nil {
		return nil
	}
	defer admin.Close()

	rows, err := admin.Query(ctx, "SELECT datname FROM pg_database WHERE datname ~ $1",
		fmt.Sprintf(`^flow_(test|template)_%d_[0-9]+$`, os.Getpid()))
	if err != nil {
		return err
	}
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}
	for _, name := range names {
		if _, err := admin.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)"); err != nil {
			return fmt.Errorf("drop %s: %w", name, err)
		}
	}
	return nil
}

// NewDatabase creates an empty, unmigrated database and registers its drop
// as test cleanup. It skips the test when the stack is not reachable.
func NewDatabase(t testing.TB) string {
	t.Helper()
	return createDatabase(t, "")
}

// NewMigratedDatabase creates a database carrying every embedded migration,
// copied from the package's template, and registers its drop as test
// cleanup. It skips the test when the stack is not reachable.
func NewMigratedDatabase(t testing.TB) string {
	t.Helper()
	templateOnce.Do(buildTemplate)
	if unreachable != nil {
		t.Skipf("flow-postgres compose stack not reachable: %v", unreachable)
	}
	if templateErr != nil {
		t.Fatalf("build template database: %v", templateErr)
	}
	return createDatabase(t, templateName)
}

// NewStore opens a store on a NewMigratedDatabase and registers its Close
// as test cleanup.
func NewStore(t testing.TB) *store.Store {
	t.Helper()
	dsn := NewMigratedDatabase(t)

	st, err := store.Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(st.Close)
	return st
}

// buildTemplate creates and migrates the package's template database, then
// closes every connection to it: CREATE DATABASE ... TEMPLATE refuses a
// template another session is connected to.
func buildTemplate() {
	ctx := context.Background()
	admin, err := connectAdmin(ctx)
	if err != nil {
		unreachable = err
		return
	}
	name := uniqueName("flow_template")
	_, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize())
	admin.Close()
	if err != nil {
		templateErr = fmt.Errorf("create %s: %w", name, err)
		return
	}
	templateName = name
	templateErr = migrateDatabase(name)
}

// migrateDatabase runs every embedded migration on dbName and closes its
// connections before returning.
func migrateDatabase(dbName string) error {
	ctx := context.Background()
	st, err := store.Open(ctx, DSN(dbName))
	if err != nil {
		return fmt.Errorf("open %s: %w", dbName, err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		return fmt.Errorf("migrate %s: %w", dbName, err)
	}
	return nil
}

func createDatabase(t testing.TB, template string) string {
	t.Helper()

	ctx := context.Background()
	admin, err := connectAdmin(ctx)
	if err != nil {
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}
	defer admin.Close()

	dbName := uniqueName("flow_test")
	stmt := "CREATE DATABASE " + pgx.Identifier{dbName}.Sanitize()
	if template != "" {
		stmt += " TEMPLATE " + pgx.Identifier{template}.Sanitize()
	}
	if _, err := admin.Exec(ctx, stmt); err != nil {
		t.Fatalf("create test database %s: %v", dbName, err)
	}

	t.Cleanup(func() {
		if err := dropDatabase(context.Background(), dbName); err != nil {
			t.Logf("drop test database %s: %v", dbName, err)
		}
	})

	return DSN(dbName)
}

func connectAdmin(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, AdminDSN())
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, reachTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func dropDatabase(ctx context.Context, dbName string) error {
	admin, err := connectAdmin(ctx)
	if err != nil {
		return fmt.Errorf("reconnect: %w", err)
	}
	defer admin.Close()
	_, err = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{dbName}.Sanitize()+" WITH (FORCE)")
	return err
}

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, os.Getpid(), time.Now().UnixNano())
}
