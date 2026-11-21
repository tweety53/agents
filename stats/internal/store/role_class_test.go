package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/storetest"
)

// TestDispatchClassMatchesRecordsClass holds the SQL dispatch_class (0094) to
// records.Class: the queries fold a stored roles-table role to its coarse
// class in SQL, the Go consumers in Go, and the two must never disagree.
func TestDispatchClassMatchesRecordsClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dsn := storetest.NewMigratedDatabase(t)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open raw pool: %v", err)
	}
	t.Cleanup(pool.Close)

	roles := []string{"implementer", "reviewer", "verifier", "explorer", "decider", "unknown-role"}
	for role := range records.RoleClasses() {
		roles = append(roles, role)
	}
	for _, role := range roles {
		var got string
		if err := pool.QueryRow(ctx, `SELECT dispatch_class($1::text)`, role).Scan(&got); err != nil {
			t.Fatalf("dispatch_class(%q): %v", role, err)
		}
		if want := records.Class(role); got != want {
			t.Errorf("dispatch_class(%q) = %q, records.Class = %q", role, got, want)
		}
	}
}
