package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

var healthPeriod = store.Period{
	From: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	To:   time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
}

func guardRun(guard, outcome string, exit, ms int, at time.Time) records.GuardRun {
	return records.GuardRun{Guard: guard, Worktree: "/tmp/wt", ExitCode: exit, Outcome: outcome, DurationMS: ms, RecordedAt: at}
}

func TestRecordGuardRunRefusesAnUnknownProject(t *testing.T) {
	st := newTestStore(t)
	_, err := st.RecordGuardRun(context.Background(), "no-such-project", guardRun("g", "clear", 0, 1, time.Now()))
	if !errors.Is(err, store.ErrProjectNotFound) {
		t.Fatalf("err = %v, want ErrProjectNotFound", err)
	}
}

func TestRecordGuardRunRefusesAnInvalidRun(t *testing.T) {
	st := newTestStore(t)
	projectKey := fmt.Sprintf("proj-grinv-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	for name, in := range map[string]records.GuardRun{
		"no guard":         guardRun("", "clear", 0, 1, time.Now()),
		"unknown outcome":  guardRun("g", "exploded", 3, 1, time.Now()),
		"negative runtime": guardRun("g", "clear", 0, -1, time.Now()),
	} {
		if _, err := st.RecordGuardRun(context.Background(), projectKey, in); !errors.Is(err, store.ErrInvalidGuardRun) {
			t.Errorf("%s: err = %v, want ErrInvalidGuardRun", name, err)
		}
	}
}

func TestPanelRoundsPerChange(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-rounds-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	seedChange(t, st, projectKey, "kan-2")
	seedChange(t, st, projectKey, "kan-3")

	start := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	for i, change := range []string{"kan-1", "kan-2"} {
		in := baseBeginInput(projectKey, change, "/flow", "flow.review-panel")
		in.StartedAt = start.Add(time.Duration(i) * time.Hour)
		runStage(t, st, in, nil, in.StartedAt.Add(time.Minute), "completed")
	}
	// kan-3 never reached the panel: not a row, whatever it recorded.
	for _, f := range []records.Finding{
		{Ref: "F1", Round: 0, Slot: "primary", Severity: "Important", Note: "a", Status: "fixed"},
		{Ref: "F2", Round: 0, Slot: "primary", Severity: "minor", Note: "b", Status: "fixed"},
		{Ref: "F3", Round: 1, Slot: "bugbot", Severity: "Critical", Note: "c", Status: "fixed"},
	} {
		if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", f); err != nil {
			t.Fatalf("UpsertFinding: %v", err)
		}
	}
	// kan-1's clean round-2 pass is what makes it three rounds.
	if _, err := st.RecordPass(ctx, projectKey, "kan-1", records.Pass{Round: 2, Note: "clean"}); err != nil {
		t.Fatalf("RecordPass: %v", err)
	}

	rows, err := st.PanelRounds(ctx, healthPeriod, &projectKey)
	if err != nil {
		t.Fatalf("PanelRounds: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %+v, want 2 (kan-3 never entered the panel)", rows)
	}
	// Newest first: kan-2 started an hour after kan-1.
	if rows[0].Change != "kan-2" || rows[0].Rounds != nil || rows[0].Findings != 0 {
		t.Errorf("rows[0] = %+v, want kan-2 with no recorded rounds", rows[0])
	}
	k1 := rows[1]
	if k1.Change != "kan-1" || k1.Rounds == nil || *k1.Rounds != 3 {
		t.Fatalf("rows[1] = %+v, want kan-1 with 3 rounds", k1)
	}
	if k1.Findings != 3 || k1.Critical != 1 || k1.Important != 1 || k1.Minor != 1 {
		t.Errorf("kan-1 counts = %+v, want 3 findings, one per severity", k1)
	}
}

// TestFlowSummaryTotals pins the eight flow-health tiles (design.md's
// flow-health-tiles): a change starts at its first stage run and finishes
// at its FINISHED write (a withdrawal is not a finish); cost is the stage
// runs started in period, the per-finished mean is whole-life; the panel
// figures come from PanelRounds and guard fires from guard_runs.
func TestFlowSummaryTotals(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-summary-%d", time.Now().UnixNano())
	for _, name := range []string{"kan-1", "kan-2", "kan-3", "kan-4"} {
		seedChange(t, st, projectKey, name)
	}

	stage := func(change, name string, at time.Time, cost float64) {
		t.Helper()
		in := baseBeginInput(projectKey, change, "/flow", name)
		in.StartedAt = at
		costRun(t, st, in, cost)
	}
	// kan-1 started before the period, re-entered the panel inside it.
	kan1Start := time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC)
	stage("kan-1", "flow.kickoff", kan1Start, 1.0)
	stage("kan-1", "flow.review-panel", time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC), 2.0)
	// kan-2 started and finished inside it.
	kan2Start := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	stage("kan-2", "flow.review-panel", kan2Start, 4.0)
	// kan-3 started inside it and was withdrawn.
	stage("kan-3", "flow.kickoff", time.Date(2026, 8, 11, 9, 0, 0, 0, time.UTC), 0.5)

	finish := func(name string, at time.Time, withdrawn bool) {
		t.Helper()
		c := baseChange(projectKey, name)
		c.State, c.Withdrawn, c.UpdatedAt = store.StateFinished, withdrawn, at
		if err := st.PutChange(ctx, c); err != nil {
			t.Fatalf("PutChange %s: %v", name, err)
		}
	}
	kan1End := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	finish("kan-1", kan1End, false)
	finish("kan-2", kan2Start.Add(48*time.Hour), false)
	finish("kan-3", time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC), true)

	for _, f := range []records.Finding{
		{Ref: "F1", Round: 0, Slot: "primary", Severity: "Critical", Note: "a", Status: "fixed"},
		{Ref: "F2", Round: 1, Slot: "primary", Severity: "Important", Note: "b", Status: "fixed"},
	} {
		if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", f); err != nil {
			t.Fatalf("UpsertFinding: %v", err)
		}
	}
	day := func(d int) time.Time { return time.Date(2026, 8, d, 12, 0, 0, 0, time.UTC) }
	for _, in := range []records.GuardRun{
		guardRun("check-plan-shape", "fired", 1, 10, day(2)),
		guardRun("check-plan-shape", "fired", 1, 10, day(3)),
		guardRun("check-plan-shape", "clear", 0, 10, day(4)),
		guardRun("check-plan-shape", "fired", 1, 10, time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)),
	} {
		if _, err := st.RecordGuardRun(ctx, projectKey, in); err != nil {
			t.Fatalf("RecordGuardRun: %v", err)
		}
	}

	got, err := st.FlowSummary(ctx, healthPeriod, &projectKey)
	if err != nil {
		t.Fatalf("FlowSummary: %v", err)
	}
	if got.ChangesStarted != 2 || got.ChangesFinished != 2 {
		t.Errorf("started/finished = %d/%d, want 2/2 (kan-2, kan-3 started; kan-1, kan-2 finished)", got.ChangesStarted, got.ChangesFinished)
	}
	if got.TotalCostUsd != 6.5 {
		t.Errorf("TotalCostUsd = %v, want 6.5 (2.0 + 4.0 + 0.5 started in period)", got.TotalCostUsd)
	}
	if got.MeanCostPerFinished == nil || *got.MeanCostPerFinished != 3.5 {
		t.Errorf("MeanCostPerFinished = %v, want 3.5 (whole-life 3.0 and 4.0)", got.MeanCostPerFinished)
	}
	wantLead := (kan1End.Sub(kan1Start) + 48*time.Hour).Milliseconds() / 2
	if got.MedianLeadTimeMs == nil || *got.MedianLeadTimeMs != wantLead {
		t.Errorf("MedianLeadTimeMs = %v, want %d", got.MedianLeadTimeMs, wantLead)
	}
	if got.MeanPanelRounds == nil || *got.MeanPanelRounds != 2 {
		t.Errorf("MeanPanelRounds = %v, want 2 (kan-1's two rounds; kan-2 recorded none)", got.MeanPanelRounds)
	}
	if got.Critical != 1 || got.Important != 1 || got.Minor != 0 {
		t.Errorf("C/I/Mi = %d/%d/%d, want 1/1/0", got.Critical, got.Important, got.Minor)
	}
	if got.GuardFires != 2 {
		t.Errorf("GuardFires = %d, want 2", got.GuardFires)
	}
}
