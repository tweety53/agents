package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// TestDropLeftoverDatabasesRemovesALeftover pins Main's sweep: a test
// database whose cleanup never dropped it, and a template whose process
// never reached its own cleanup, are both gone once dropLeftoverDatabases
// has run.
func TestDropLeftoverDatabasesRemovesALeftover(t *testing.T) {
	ctx := context.Background()
	admin, err := connectAdmin(ctx)
	if err != nil {
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}
	defer admin.Close()

	leftovers := []string{uniqueName("flow_test"), uniqueName("flow_template")}
	for _, name := range leftovers {
		if _, err := admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
			t.Fatalf("create leftover database %s: %v", name, err)
		}
	}

	if err := dropLeftoverDatabases(ctx); err != nil {
		t.Fatalf("dropLeftoverDatabases: %v", err)
	}

	for _, name := range leftovers {
		var exists bool
		if err := admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
			t.Fatalf("look up leftover database %s: %v", name, err)
		}
		if exists {
			t.Fatalf("database %s outlived dropLeftoverDatabases", name)
		}
	}
}

// TestMigrationWaitsOutSlowSetupInsteadOfTimingOut pins that template
// migration carries no wall-clock bound. Another session holds the
// migration advisory lock on the database for longer than reachTimeout --
// the deterministic stand-in for a server too loaded to finish within a
// fixed deadline -- and migrateDatabase must wait for it and succeed. With a
// reachTimeout-sized deadline on the migration, it fails with
// "context deadline exceeded".
func TestMigrationWaitsOutSlowSetupInsteadOfTimingOut(t *testing.T) {
	ctx := context.Background()
	dbName := NewDatabase(t)

	holder, err := pgx.Connect(ctx, dbName)
	if err != nil {
		t.Fatalf("connect lock holder: %v", err)
	}
	defer holder.Close(ctx)
	const migrationsLockKey = 725_016_001 // internal/store's migration lock
	if _, err := holder.Exec(ctx, "SELECT pg_advisory_lock($1)", int64(migrationsLockKey)); err != nil {
		t.Fatalf("take migration lock: %v", err)
	}
	released := make(chan error, 1)
	go func() {
		time.Sleep(reachTimeout + time.Second)
		_, err := holder.Exec(ctx, "SELECT pg_advisory_unlock($1)", int64(migrationsLockKey))
		released <- err
	}()

	cfg, err := pgx.ParseConfig(dbName)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	if err := migrateDatabase(cfg.Database); err != nil {
		t.Fatalf("migrateDatabase while setup is slow: %v", err)
	}
	if err := <-released; err != nil {
		t.Fatalf("release migration lock: %v", err)
	}
}
