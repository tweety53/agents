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

func TestGuardActivityMergesRunsAndVerdicts(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-guards-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	day := func(d int) time.Time { return time.Date(2026, 8, d, 12, 0, 0, 0, time.UTC) }
	for _, in := range []records.GuardRun{
		guardRun("check-plan-shape", "clear", 0, 100, day(2)),
		guardRun("check-plan-shape", "fired", 1, 300, day(3)),
		guardRun("check-plan-shape", "cannot-answer", 2, 200, day(4)),
		guardRun("check-quiet", "clear", 0, 10, day(5)),
		// Outside the period: never counted.
		guardRun("check-plan-shape", "fired", 1, 999, time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)),
	} {
		if _, err := st.RecordGuardRun(ctx, projectKey, in); err != nil {
			t.Fatalf("RecordGuardRun: %v", err)
		}
	}
	if _, err := st.RecordVerdict(ctx, projectKey, "kan-1", baseVerdict("check-unfinished-work", "OUTSTANDING: x", day(6))); err != nil {
		t.Fatalf("RecordVerdict: %v", err)
	}
	if _, err := st.FlagVerdictFalsePositive(ctx, projectKey, "kan-1", records.VerdictFlag{Guard: "check-unfinished-work", Reason: "hand-verified"}); err != nil {
		t.Fatalf("FlagVerdictFalsePositive: %v", err)
	}

	rows, err := st.GuardActivity(ctx, healthPeriod, &projectKey)
	if err != nil {
		t.Fatalf("GuardActivity: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("rows = %+v, want 3 guards", rows)
	}
	shape := rows[0]
	if shape.Guard != "check-plan-shape" || shape.Runs != 3 || shape.Fired != 1 || shape.CannotAnswer != 1 {
		t.Errorf("check-plan-shape = %+v, want 3 runs, 1 fired, 1 cannot-answer", shape)
	}
	if shape.LastFiredAt == nil || !shape.LastFiredAt.Equal(day(3)) {
		t.Errorf("LastFiredAt = %v, want %v", shape.LastFiredAt, day(3))
	}
	if shape.MedianDurationMS == nil || *shape.MedianDurationMS != 200 {
		t.Errorf("MedianDurationMS = %v, want 200", shape.MedianDurationMS)
	}
	quiet := rows[1]
	if quiet.Guard != "check-quiet" || quiet.Runs != 1 || quiet.Fired != 0 || quiet.LastFiredAt != nil {
		t.Errorf("check-quiet = %+v, want 1 run that never fired", quiet)
	}
	verdictOnly := rows[2]
	if verdictOnly.Guard != "check-unfinished-work" || verdictOnly.Runs != 0 || verdictOnly.Verdicts != 1 ||
		verdictOnly.FalsePositives != 1 || verdictOnly.LastRunAt != nil {
		t.Errorf("check-unfinished-work = %+v, want a verdict-only row with one false positive", verdictOnly)
	}
}

func TestStageRedoCountsReentries(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-redo-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	seedChange(t, st, projectKey, "kan-2")

	start := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	// kan-1 enters the panel twice (attempt 1 and 2); kan-2 once.
	for i, spec := range []struct {
		change string
		dur    time.Duration
	}{{"kan-1", time.Minute}, {"kan-1", 3 * time.Minute}, {"kan-2", 2 * time.Minute}} {
		in := baseBeginInput(projectKey, spec.change, "/flow", "flow.review-panel")
		in.StartedAt = start.Add(time.Duration(i) * time.Hour)
		runStage(t, st, in, nil, in.StartedAt.Add(spec.dur), "completed")
	}

	rows, err := st.StageRedo(ctx, healthPeriod, &projectKey)
	if err != nil {
		t.Fatalf("StageRedo: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %+v, want 1", rows)
	}
	r := rows[0]
	if r.Runs != 3 || r.Changes != 2 || r.Reentries != 1 || r.ReenteredChanges != 1 {
		t.Errorf("row = %+v, want 3 runs over 2 changes, 1 re-entry on 1 change", r)
	}
	if r.MedianSeconds == nil || *r.MedianSeconds != 120 {
		t.Errorf("MedianSeconds = %v, want 120", r.MedianSeconds)
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
