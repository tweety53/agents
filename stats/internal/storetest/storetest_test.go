package storetest_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/storetest"
)

func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }

// TestMigratedDatabasesAreCopiesOfOneTemplate pins the mechanism the suite's
// setup rests on: two migrated databases carry every embedded migration
// with identical applied_at stamps, which only a copy of one migrated
// template produces. A database migrated per test stamps each migration
// with its own transaction's now(), so the stamps differ and this fails.
func TestMigratedDatabasesAreCopiesOfOneTemplate(t *testing.T) {
	want, err := store.EmbeddedMigrationCount()
	if err != nil {
		t.Fatal(err)
	}
	first := migrationStamps(t, storetest.NewMigratedDatabase(t))
	second := migrationStamps(t, storetest.NewMigratedDatabase(t))

	if len(first) != want {
		t.Fatalf("migrated database records %d migrations, want %d", len(first), want)
	}
	for name, stamp := range first {
		if !second[name].Equal(stamp) {
			t.Fatalf("migration %s applied at %v in one database and %v in the other: each was migrated on its own, not copied from the template",
				name, stamp, second[name])
		}
	}
}

func migrationStamps(t *testing.T, dsn string) map[string]time.Time {
	t.Helper()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(ctx, "SELECT filename, applied_at FROM schema_migrations")
	if err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	stamps := make(map[string]time.Time)
	for rows.Next() {
		var name string
		var at time.Time
		if err := rows.Scan(&name, &at); err != nil {
			t.Fatalf("scan schema_migrations: %v", err)
		}
		stamps[name] = at
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	return stamps
}
