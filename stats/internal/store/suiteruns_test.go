package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// newSuiteRunStore returns a freshly-migrated store and a raw pool, plus a
// project row to hang suite runs off: a suite run, like a hazard, never
// carries a change, so seedChange's change row would be dead weight here.
// The pool exists because seeding the project row is not reachable through
// the typed API -- PutChange bootstraps the project as a side effect of a
// change, and a suite-runs test wants the project alone.
func newSuiteRunStore(t *testing.T) (*store.Store, *pgxpool.Pool, string) {
	t.Helper()

	st, pool := newRecordStore(t)
	projectKey := fmt.Sprintf("proj-suites-%d", time.Now().UnixNano())
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO projects (project_key, main_checkout_path) VALUES ($1, $2)
		 ON CONFLICT (project_key) DO NOTHING`,
		projectKey, t.TempDir(),
	); err != nil {
		t.Fatalf("seed project %s: %v", projectKey, err)
	}
	return st, pool, projectKey
}

// TestInsertSuiteRunReturnsRecordedRow covers the write path: the returned
// row carries what was recorded plus the store's own stamps (id, ran_at),
// and a second insert is a second row -- no upsert, one row per run.
func TestInsertSuiteRunReturnsRecordedRow(t *testing.T) {
	st, pool, projectKey := newSuiteRunStore(t)
	ctx := context.Background()

	added, err := st.InsertSuiteRun(ctx, projectKey, records.SuiteRun{
		Suite: "guard-tests", Host: "laptop", DurationMs: 52_400, ExitCode: 0,
	})
	if err != nil {
		t.Fatalf("InsertSuiteRun: %v", err)
	}
	if added.Suite != "guard-tests" || added.Host != "laptop" {
		t.Errorf("row = %s/%s, want guard-tests/laptop", added.Suite, added.Host)
	}
	if added.DurationMs != 52_400 || added.ExitCode != 0 {
		t.Errorf("row = %d ms exit %d, want 52400 ms exit 0", added.DurationMs, added.ExitCode)
	}
	if added.ID == 0 {
		t.Errorf("ID = 0, want the allocated row id")
	}
	if added.RanAt.IsZero() {
		t.Errorf("RanAt = zero time, want the column's now() default")
	}

	second, err := st.InsertSuiteRun(ctx, projectKey, records.SuiteRun{
		Suite: "guard-tests", Host: "laptop", DurationMs: 53_100, ExitCode: 0,
	})
	if err != nil {
		t.Fatalf("second InsertSuiteRun: %v", err)
	}
	if second.ID == added.ID {
		t.Errorf("second insert reused row id %d, want a new row per run", added.ID)
	}

	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM suite_runs WHERE project_key = $1`, projectKey).Scan(&n); err != nil {
		t.Fatalf("count suite_runs: %v", err)
	}
	if n != 2 {
		t.Errorf("suite_runs rows = %d, want 2", n)
	}
}

// TestInsertSuiteRunRejectsInvalidRow pins the store boundary: an empty
// suite, an empty host or a negative duration is the typed
// ErrInvalidSuiteRun refusal, the same caller-mistake split
// ErrInvalidHazardShape draws -- the API's validation is UX, this one is
// the boundary, and it is covered here rather than only through a fake.
func TestInsertSuiteRunRejectsInvalidRow(t *testing.T) {
	st, _, projectKey := newSuiteRunStore(t)
	ctx := context.Background()

	for name, run := range map[string]records.SuiteRun{
		"empty suite":       {Suite: "", Host: "laptop", DurationMs: 1_000, ExitCode: 0},
		"empty host":        {Suite: "guard-tests", Host: "", DurationMs: 1_000, ExitCode: 0},
		"negative duration": {Suite: "guard-tests", Host: "laptop", DurationMs: -1, ExitCode: 0},
	} {
		if _, err := st.InsertSuiteRun(ctx, projectKey, run); !errors.Is(err, store.ErrInvalidSuiteRun) {
			t.Errorf("%s: err = %v, want ErrInvalidSuiteRun", name, err)
		}
	}
}

// TestListSuiteRunsFiltersBySuiteAndLimit asserts the read path: an empty
// suite lists every suite, a named suite lists only that suite's rows, both
// newest first, and limit caps the row count.
func TestListSuiteRunsFiltersBySuiteAndLimit(t *testing.T) {
	st, _, projectKey := newSuiteRunStore(t)
	ctx := context.Background()

	seed := []records.SuiteRun{
		{Suite: "guard-tests", Host: "laptop", DurationMs: 52_000, ExitCode: 0},
		{Suite: "guard-tests", Host: "ci", DurationMs: 90_000, ExitCode: 0},
		{Suite: "stats-go", Host: "laptop", DurationMs: 24_000, ExitCode: 0},
	}
	for i, run := range seed {
		if _, err := st.InsertSuiteRun(ctx, projectKey, run); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	all, err := st.ListSuiteRuns(ctx, projectKey, "", 50)
	if err != nil {
		t.Fatalf("ListSuiteRuns (all): %v", err)
	}
	if len(all) != 3 {
		t.Errorf("ListSuiteRuns (all) = %d rows, want 3", len(all))
	}
	if all[0].Suite != "stats-go" {
		t.Errorf("newest row = %s, want stats-go (reverse insertion order)", all[0].Suite)
	}

	guard, err := st.ListSuiteRuns(ctx, projectKey, "guard-tests", 50)
	if err != nil {
		t.Fatalf("ListSuiteRuns (guard-tests): %v", err)
	}
	if len(guard) != 2 {
		t.Errorf("ListSuiteRuns (guard-tests) = %d rows, want 2", len(guard))
	}
	for _, run := range guard {
		if run.Suite != "guard-tests" {
			t.Errorf("filtered list carried suite %s", run.Suite)
		}
	}

	capped, err := st.ListSuiteRuns(ctx, projectKey, "", 2)
	if err != nil {
		t.Fatalf("ListSuiteRuns (limit 2): %v", err)
	}
	if len(capped) != 2 {
		t.Errorf("ListSuiteRuns (limit 2) = %d rows, want 2", len(capped))
	}
}

// TestSuiteRunsRejectUnknownProject asserts the foreign key surfaces: an
// insert naming a project with no projects row fails with Postgres' FK
// violation (23503), never a silently stored row -- the boundary that keeps
// a mistyped -C from filing runtimes under a project that does not exist.
func TestSuiteRunsRejectUnknownProject(t *testing.T) {
	st, _, _ := newSuiteRunStore(t)
	ctx := context.Background()

	_, err := st.InsertSuiteRun(ctx, "proj-never-seeded", records.SuiteRun{
		Suite: "guard-tests", Host: "laptop", DurationMs: 1_000, ExitCode: 0,
	})
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Fatalf("InsertSuiteRun unknown project err = %v, want a 23503 foreign-key violation", err)
	}
}
