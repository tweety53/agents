package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/store"
)

// newSpecRunStore returns a freshly-migrated store and a raw pool, plus a
// project row to hang spec runs off: a spec run, like a suite run, never
// carries a change, so seedChange's change row would be dead weight here.
// newSuiteRunStore's reason, restated for the spec inventory.
func newSpecRunStore(t *testing.T) (*store.Store, *pgxpool.Pool, string) {
	t.Helper()

	st, pool := newRecordStore(t)
	projectKey := fmt.Sprintf("proj-specs-%d", time.Now().UnixNano())
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO projects (project_key, main_checkout_path) VALUES ($1, $2)
		 ON CONFLICT (project_key) DO NOTHING`,
		projectKey, t.TempDir(),
	); err != nil {
		t.Fatalf("seed project %s: %v", projectKey, err)
	}
	return st, pool, projectKey
}

// TestRecordSpecRunUpsertsKeepsLatest covers the write path: one row per
// (project, spec), an out-of-order older call never drags the timestamp
// back, and the returned row carries the stored timestamp plus the
// changes-since figure the inventory stat is built from.
func TestRecordSpecRunUpsertsKeepsLatest(t *testing.T) {
	st, pool, projectKey := newSpecRunStore(t)
	ctx := context.Background()

	older := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	first, err := st.RecordSpecRun(ctx, projectKey, "baseline.spec.ts", newer)
	if err != nil {
		t.Fatalf("RecordSpecRun: %v", err)
	}
	if first.Spec != "baseline.spec.ts" || !first.LastRanAt.Equal(newer) {
		t.Errorf("row = %s at %s, want baseline.spec.ts at %s", first.Spec, first.LastRanAt, newer)
	}

	back, err := st.RecordSpecRun(ctx, projectKey, "baseline.spec.ts", older)
	if err != nil {
		t.Fatalf("older RecordSpecRun: %v", err)
	}
	if !back.LastRanAt.Equal(newer) {
		t.Errorf("older call moved last_ran_at to %s, want it kept at %s", back.LastRanAt, newer)
	}

	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM spec_lastruns WHERE project_key = $1`, projectKey).Scan(&n); err != nil {
		t.Fatalf("count spec_lastruns: %v", err)
	}
	if n != 1 {
		t.Errorf("spec_lastruns rows = %d, want 1 (one row per spec)", n)
	}
}

// TestRecordSpecRunRejectsInvalidRow pins the store boundary: an empty spec
// is the typed ErrInvalidSpecRun refusal -- the API's validation is UX,
// this one is the boundary, ErrInvalidSuiteRun's same split.
func TestRecordSpecRunRejectsInvalidRow(t *testing.T) {
	st, _, projectKey := newSpecRunStore(t)
	ctx := context.Background()

	if _, err := st.RecordSpecRun(ctx, projectKey, "", time.Now()); !errors.Is(err, store.ErrInvalidSpecRun) {
		t.Errorf("empty spec: err = %v, want ErrInvalidSpecRun", err)
	}

	if _, err := st.RecordSpecRun(ctx, "proj-never-seeded", "baseline.spec.ts", time.Now()); err == nil {
		t.Fatalf("unknown project: err = nil, want a foreign-key violation")
	} else {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
			t.Errorf("unknown project err = %v, want a 23503 foreign-key violation", err)
		}
	}
}

// TestListSpecLastRunsCountsChangesSince asserts the inventory stat: a
// recorded spec's changes-since counts the changes updated after its last
// run, a named-but-never-recorded spec answers nil last_ran_at with the
// project's whole change count, and an unnamed read lists recorded specs
// alone.
func TestListSpecLastRunsCountsChangesSince(t *testing.T) {
	st, _, projectKey := newSpecRunStore(t)
	ctx := context.Background()

	// Two changes the project has recorded, updated now: every spec run set
	// in the past counts both, a run set in the future counts none.
	changed := time.Now().UTC()
	for _, name := range []string{"kan-1", "kan-2"} {
		if err := st.PutChange(ctx, store.Change{
			ProjectKey: projectKey, MainCheckoutPath: t.TempDir(), Name: name, State: store.StateStarted,
			UpdatedAt: changed, UpdatedBy: "test",
		}); err != nil {
			t.Fatalf("seed change %s: %v", name, err)
		}
	}

	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	if _, err := st.RecordSpecRun(ctx, projectKey, "runs.spec.ts", past); err != nil {
		t.Fatalf("RecordSpecRun runs.spec.ts: %v", err)
	}
	if _, err := st.RecordSpecRun(ctx, projectKey, "views.spec.ts", future); err != nil {
		t.Fatalf("RecordSpecRun views.spec.ts: %v", err)
	}

	rows, err := st.ListSpecLastRuns(ctx, projectKey, []string{"runs.spec.ts", "views.spec.ts", "controls.spec.ts"})
	if err != nil {
		t.Fatalf("ListSpecLastRuns (named): %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("named list = %d rows, want 3 (every named spec, recorded or not)", len(rows))
	}
	if rows[0].Spec != "controls.spec.ts" || rows[0].LastRanAt != nil {
		t.Errorf("controls row = %+v, want a never-recorded spec (nil last_ran_at)", rows[0])
	}
	if rows[0].ChangesSince != 2 {
		t.Errorf("never-recorded changesSince = %d, want 2 (the project's whole history)", rows[0].ChangesSince)
	}
	if rows[2].Spec != "views.spec.ts" {
		t.Errorf("rows not ordered by spec: %+v", rows)
	}
	if rows[1].ChangesSince != 2 {
		t.Errorf("past-run changesSince = %d, want 2 (both changes postdate the run)", rows[1].ChangesSince)
	}
	if rows[2].ChangesSince != 0 {
		t.Errorf("future-run changesSince = %d, want 0 (no change postdates the run)", rows[2].ChangesSince)
	}

	recordedOnly, err := st.ListSpecLastRuns(ctx, projectKey, nil)
	if err != nil {
		t.Fatalf("ListSpecLastRuns (unnamed): %v", err)
	}
	if len(recordedOnly) != 2 {
		t.Errorf("unnamed list = %d rows, want 2 (recorded specs alone, never-run ones absent)", len(recordedOnly))
	}
}

// TestRecordSpecRunStampsZeroTimeWithNow pins the zero-time default: a
// caller that does not name a time -- the API handler's every POST -- is
// stamped with the store's own now(), never a zero or stale stamp. The
// mutant killer for the COALESCE($3, now()) default.
func TestRecordSpecRunStampsZeroTimeWithNow(t *testing.T) {
	st, _, projectKey := newSpecRunStore(t)
	ctx := context.Background()

	before := time.Now().UTC().Add(-time.Minute)
	row, err := st.RecordSpecRun(ctx, projectKey, "stamp.spec.ts", time.Time{})
	if err != nil {
		t.Fatalf("RecordSpecRun zero time: %v", err)
	}
	after := time.Now().UTC().Add(time.Minute)
	if row.LastRanAt == nil {
		t.Fatalf("LastRanAt = nil, want the store's now() stamp")
	}
	if row.LastRanAt.Before(before) || row.LastRanAt.After(after) {
		t.Errorf("LastRanAt = %s, want within [%s, %s]", row.LastRanAt, before, after)
	}
}
