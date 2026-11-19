package store_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestRoleDispatchesReadsThePeriodsRows: every dispatch started in the
// period, of the asked project only, with its key, pair, span, cost and
// outcome -- a row with no cost carries none, never a 0.
func TestRoleDispatchesReadsThePeriodsRows(t *testing.T) {
	t.Parallel()
	st, _ := newRecordStore(t)
	ctx := context.Background()
	seedChange(t, st, "p", "kan-1")
	seedChange(t, st, "q", "kan-2")
	at := time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC)
	end := at.Add(10 * time.Minute)
	in := baseDispatch("implementer", "opus")
	in.Key, in.Effort, in.StartedAt, in.EndedAt, in.Outcome = "task-3-implementer", "low", at, &end, "completed"
	in.Metrics = json.RawMessage(`{"cost_usd": 1.5}`)
	if _, err := st.RecordDispatch(ctx, "p", "kan-1", in); err != nil {
		t.Fatal(err)
	}
	unpriced := baseDispatch("reviewer", "sonnet")
	unpriced.Key, unpriced.StartedAt = "panel-2-primary", at.Add(time.Hour)
	if _, err := st.RecordDispatch(ctx, "p", "kan-1", unpriced); err != nil {
		t.Fatal(err)
	}
	early := baseDispatch("planner", "opus")
	early.Key, early.StartedAt = "plan-1", at.Add(-30*24*time.Hour)
	if _, err := st.RecordDispatch(ctx, "p", "kan-1", early); err != nil {
		t.Fatal(err)
	}
	other := baseDispatch("explorer", "opus")
	other.Key = "explore-1"
	if _, err := st.RecordDispatch(ctx, "q", "kan-2", other); err != nil {
		t.Fatal(err)
	}
	p := "p"
	rows, err := st.RoleDispatches(ctx, store.Period{From: at.Add(-time.Hour), To: at.Add(2 * time.Hour)}, &p)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %+v, want the two in-period rows of p", rows)
	}
	r := rows[0]
	if r.Role != "implementer" || r.Key != "task-3-implementer" || r.Model != "opus" || r.Effort != "low" ||
		r.EndedAt == nil || !r.EndedAt.Equal(end) || r.CostUSD == nil || *r.CostUSD != 1.5 || r.Outcome != "completed" {
		t.Errorf("row 0 = %+v", r)
	}
	if rows[1].Key != "panel-2-primary" || rows[1].CostUSD != nil || rows[1].EndedAt != nil {
		t.Errorf("row 1 = %+v, want no cost and no end", rows[1])
	}
	all, err := st.RoleDispatches(ctx, store.Period{From: at.Add(-time.Hour), To: at.Add(2 * time.Hour)}, nil)
	if err != nil || len(all) != 3 {
		t.Fatalf("every project = %d rows, %v; want 3", len(all), err)
	}
}
