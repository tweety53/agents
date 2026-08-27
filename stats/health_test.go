// Package stats_test asserts that the stats module compiles and, when the
// dedicated docker-compose PostgreSQL stack is running, that this module can
// open a connection to it. When the stack is not running the test skips with
// a clear message rather than failing the whole suite.
package stats_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// dsn returns the connection string for the dedicated flow-postgres
// container defined in stats/docker-compose.yml. FLOW_STATS_DSN overrides
// it for environments that run Postgres elsewhere.
func dsn() string {
	if v := os.Getenv("FLOW_STATS_DSN"); v != "" {
		return v
	}
	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
}

func TestModuleConnectsToComposeStack(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn())
	if err != nil {
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}
}
