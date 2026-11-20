package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// TestChangeHeadlessReadsTheChangesWorktreeCalls: a headless call whose
// worktree is the change's counts, another change's or a plain guard run does not.
func TestChangeHeadlessReadsTheChangesWorktreeCalls(t *testing.T) {
	t.Parallel()
	st, _ := newRecordStore(t)
	ctx := context.Background()
	seedChange(t, st, "p", "kan-1")
	at := time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC)
	usd := 0.05
	call := func(guard, wt, role string) records.GuardRun {
		return records.GuardRun{Guard: guard, Worktree: wt, Outcome: "clear", RecordedAt: at,
			Headless: &records.HeadlessCall{Role: role, Model: "opus", Effort: "low", CostUSD: &usd, DurationMS: 3000}}
	}
	for _, in := range []records.GuardRun{
		call("plan-triage", "/r-worktrees/kan-1", "plan-triage"),
		call("pair-judge", "/r-worktrees/kan-1/.superpowers/sdd/pair-judge", "pair-judge"),
		call("pair-judge", "/r-worktrees/kan-2", "pair-judge"),
		{Guard: "check-references", Worktree: "/r-worktrees/kan-1", Outcome: "clear", RecordedAt: at},
	} {
		if _, err := st.RecordGuardRun(ctx, "p", in); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := st.ChangeHeadless(ctx, "p", "kan-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Role != "plan-triage" || rows[1].Role != "pair-judge" || rows[0].CostUSD == nil || *rows[0].CostUSD != usd {
		t.Fatalf("rows = %+v, want kan-1's plan-triage and pair-judge only", rows)
	}
}
