package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// newHazardStore returns a freshly-migrated store and a raw pool, plus a
// project row to hang hazards off: a hazard, unlike an incident, never
// carries a change, so seedChange's change row would be dead weight here.
// The pool exists because seeding the project row is not reachable through
// the typed API -- PutChange bootstraps the project as a side effect of a
// change, and a hazards test wants the project alone.
func newHazardStore(t *testing.T) (*store.Store, *pgxpool.Pool, string) {
	t.Helper()

	st, pool := newRecordStore(t)
	projectKey := fmt.Sprintf("proj-hazards-%d", time.Now().UnixNano())
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO projects (project_key, main_checkout_path) VALUES ($1, $2)
		 ON CONFLICT (project_key) DO NOTHING`,
		projectKey, t.TempDir(),
	); err != nil {
		t.Fatalf("seed project %s: %v", projectKey, err)
	}
	return st, pool, projectKey
}

// TestAddListRetireHazard covers the row's whole life: added active, listed
// newest-first, retired stays listed only under includeInactive.
func TestAddListRetireHazard(t *testing.T) {
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()

	added, err := st.AddHazard(ctx, projectKey, records.Hazard{
		Name: "guard-single-repo", Body: "pass the change's own worktree", Applies: "single-repo",
	})
	if err != nil {
		t.Fatalf("AddHazard: %v", err)
	}
	if !added.Active {
		t.Errorf("Active = false, want true: AddHazard records active rows; retiring is RetireHazard's job")
	}
	if added.CreatedAt.IsZero() {
		t.Errorf("CreatedAt = zero time, want the column's now() default")
	}
	if added.ID == 0 {
		t.Errorf("ID = 0, want the allocated row id")
	}

	// An older row, seeded with an explicit CreatedAt, must list second.
	older, err := st.AddHazard(ctx, projectKey, records.Hazard{
		Name: "guard-reverts-head-only", Body: "inspect before retrying", Applies: "all",
		CreatedAt: added.CreatedAt.Add(-time.Minute),
	})
	if err != nil {
		t.Fatalf("AddHazard (older): %v", err)
	}

	// The empty shape is the operator's unfiltered listing: both rows
	// qualify regardless of applies; the older row, seeded with an explicit
	// CreatedAt, must list second.
	list, err := st.ListHazards(ctx, projectKey, "", false)
	if err != nil {
		t.Fatalf("ListHazards: %v", err)
	}
	if len(list) != 2 || list[0].Name != added.Name || list[1].Name != older.Name {
		t.Errorf("ListHazards = %v, want newest first: [%s %s]", hazardNames(list), added.Name, older.Name)
	}

	retired, err := st.RetireHazard(ctx, projectKey, added.Name)
	if err != nil {
		t.Fatalf("RetireHazard: %v", err)
	}
	if retired.Active {
		t.Errorf("Active = true after retire, want false")
	}

	list, err = st.ListHazards(ctx, projectKey, "", false)
	if err != nil {
		t.Fatalf("ListHazards after retire: %v", err)
	}
	if len(list) != 1 || list[0].Name != older.Name {
		t.Errorf("ListHazards after retire = %v, want only %s", hazardNames(list), older.Name)
	}

	list, err = st.ListHazards(ctx, projectKey, "", true)
	if err != nil {
		t.Fatalf("ListHazards includeInactive: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("ListHazards includeInactive = %v, want both rows", hazardNames(list))
	}
}

// TestListHazardsFiltersByShape asserts the shape filter is applies IN
// ('all', shape): the always-on rows inject beside the shape-specific ones,
// and never the other shape's rows.
func TestListHazardsFiltersByShape(t *testing.T) {
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()

	for _, applies := range []string{"all", "cross-repo", "single-repo"} {
		if _, err := st.AddHazard(ctx, projectKey, records.Hazard{
			Name: "h-" + applies, Body: "b", Applies: applies,
		}); err != nil {
			t.Fatalf("AddHazard %s: %v", applies, err)
		}
	}

	for shape, want := range map[string][]string{
		// All three rows share one created_at instant, so the tiebreak is
		// id DESC -- reverse insertion order.
		"":            {"h-single-repo", "h-cross-repo", "h-all"},
		"all":         {"h-all"},
		"cross-repo":  {"h-cross-repo", "h-all"},
		"single-repo": {"h-single-repo", "h-all"},
	} {
		list, err := st.ListHazards(ctx, projectKey, shape, false)
		if err != nil {
			t.Fatalf("ListHazards %s: %v", shape, err)
		}
		if len(list) != len(want) {
			t.Errorf("ListHazards %s = %v, want %v", shape, hazardNames(list), want)
			continue
		}
		for i, name := range want {
			if list[i].Name != name {
				t.Errorf("ListHazards %s[%d] = %s, want %s", shape, i, list[i].Name, name)
			}
		}
	}
}

// TestListHazardsIncludeInactive isolates the includeInactive lift: a
// retired row reappears only when it is set, active rows regardless.
func TestListHazardsIncludeInactive(t *testing.T) {
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()

	if _, err := st.AddHazard(ctx, projectKey, records.Hazard{
		Name: "retired", Body: "b", Applies: "all",
	}); err != nil {
		t.Fatalf("AddHazard: %v", err)
	}
	if _, err := st.RetireHazard(ctx, projectKey, "retired"); err != nil {
		t.Fatalf("RetireHazard: %v", err)
	}

	list, err := st.ListHazards(ctx, projectKey, "all", true)
	if err != nil {
		t.Fatalf("ListHazards includeInactive: %v", err)
	}
	if len(list) != 1 || list[0].Name != "retired" {
		t.Errorf("ListHazards includeInactive = %v, want the retired row", hazardNames(list))
	}
}

// TestAddHazardDuplicateRefused asserts the unique constraint surfaces as
// ErrHazardDuplicate -- a typed refusal the API maps to 409 -- and that the
// refused insert left exactly one row.
func TestAddHazardDuplicateRefused(t *testing.T) {
	st, pool, projectKey := newHazardStore(t)
	ctx := context.Background()

	in := records.Hazard{Name: "dup", Body: "b", Applies: "all"}
	if _, err := st.AddHazard(ctx, projectKey, in); err != nil {
		t.Fatalf("first AddHazard: %v", err)
	}
	if _, err := st.AddHazard(ctx, projectKey, in); !errors.Is(err, store.ErrHazardDuplicate) {
		t.Fatalf("second AddHazard err = %v, want ErrHazardDuplicate", err)
	}

	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM hazards WHERE project_key = $1`, projectKey).Scan(&n); err != nil {
		t.Fatalf("count hazards: %v", err)
	}
	if n != 1 {
		t.Errorf("hazards rows = %d, want 1 after the refused duplicate", n)
	}

	if _, err := st.ListHazards(ctx, projectKey, "bogus", false); !errors.Is(err, store.ErrInvalidHazardShape) {
		t.Errorf("ListHazards bogus shape err = %v, want ErrInvalidHazardShape", err)
	}

	if _, err := st.RetireHazard(ctx, projectKey, "never-recorded"); !errors.Is(err, store.ErrHazardNotFound) {
		t.Errorf("RetireHazard unknown name err = %v, want ErrHazardNotFound", err)
	}
}

func hazardNames(hs []records.Hazard) []string {
	out := make([]string, len(hs))
	for i, h := range hs {
		out[i] = h.Name
	}
	return out
}
