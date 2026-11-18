package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

var rtT0 = time.Date(2034, 1, 1, 10, 0, 0, 0, time.UTC)

func rtAt(min float64) time.Time { return rtT0.Add(time.Duration(min * float64(time.Minute))) }

func rtMin(min float64) int64 { return int64(min * 60_000) }

func rtDispatch(key, role string, from, to float64, cost string) records.Dispatch {
	d := records.Dispatch{Key: key, Role: role, Model: "opus", Effort: "high", StartedAt: rtAt(from), Metrics: json.RawMessage(`{}`)}
	if to >= 0 {
		end := rtAt(to)
		d.EndedAt = &end
		d.Outcome = "done"
	}
	if cost != "" {
		d.Metrics = json.RawMessage(`{"tokens":{"input":1},"cost_usd":` + cost + `}`)
	}
	return d
}

func rtStage(stage string, from, to float64, cost string) store.StageRun {
	end, outcome := rtAt(to), "done"
	root := "/x/gymie-worktrees/kan-1-a"
	return store.StageRun{Stage: stage, StartedAt: rtAt(from), EndedAt: &end, Outcome: &outcome, RepoRoot: &root, Metrics: json.RawMessage(`{"cost_usd":` + cost + `}`)}
}

func rtGuard(from, to float64) records.GuardRun {
	return records.GuardRun{Guard: "run-capped", Worktree: "/x/gymie-worktrees/kan-1-a", RecordedAt: rtAt(to), DurationMS: int(rtMin(to - from))}
}

// rtInput is one run of 160 minutes: overlapping task rows, a group key, a
// frontend verify row, a visual verifier, an open row without tokens, a
// guard run inside a dispatch and one nested in another outside every
// dispatch.
func rtInput() runInput {
	return runInput{
		Stages: []store.StageRun{rtStage("flow.implement", 0, 100, "6"), rtStage("flow.verify", 100, 160, "4")},
		Dispatches: []records.Dispatch{
			rtDispatch("task-1-implementer", "implementer", 0, 30, "2"),
			rtDispatch("task-1-reviewer", "reviewer", 20, 40, "1"),
			rtDispatch("task-13+14-implementer", "implementer", 35, 55, "3"),
			rtDispatch("verify-fix-3-gymie-frontend", "verifier", 100, 120, "1.5"),
			rtDispatch("visual-verify-1-gymie", "visual-verifier", 120, 130, "0.5"),
			rtDispatch("task-4-implementer", "implementer", 150, -1, ""),
		},
		Guards:  []records.GuardRun{rtGuard(5, 6), rtGuard(60, 70), rtGuard(62, 64)},
		Openers: map[string]string{"4": "verify-fix-3-gymie-frontend"},
		Repos:   []string{"gymie", "gymie-frontend"},
	}
}

func rtRow(t *testing.T, b breakdown, label string) breakdownRow {
	t.Helper()
	for _, r := range b.Rows {
		if r.Label == label {
			return r
		}
	}
	t.Fatalf("no row %q in %+v", label, b.Rows)
	return breakdownRow{}
}

// TestRunTimingsBreakdownsSumToTheirTotals: every group-by's rows add up to
// its total exactly, and the total is the run's wall clock and cost.
func TestRunTimingsBreakdownsSumToTheirTotals(t *testing.T) {
	t.Parallel()
	got := computeRun(rtInput(), rtAt(200))
	for _, by := range []string{"category", "role", "model", "repo", "round"} {
		b, ok := got.Breakdowns[by]
		if !ok {
			t.Fatalf("no %s breakdown", by)
		}
		var sum breakdownRow
		for _, r := range b.Rows {
			sum.Rows += r.Rows
			sum.ShareMs += r.ShareMs
			sum.Cost += r.Cost
			sum.NoTokens += r.NoTokens
		}
		if sum.Rows != b.Total.Rows || sum.ShareMs != b.Total.ShareMs || sum.Cost != b.Total.Cost || sum.NoTokens != b.Total.NoTokens {
			t.Errorf("%s: rows sum to %+v, total is %+v", by, sum, b.Total)
		}
		if b.Total.ShareMs != got.Totals.WallClockMs.Value || b.Total.Cost != got.Totals.Cost.Value {
			t.Errorf("%s: total %+v, want share %d and cost %d", by, b.Total, got.Totals.WallClockMs.Value, got.Totals.Cost.Value)
		}
	}
	if want := usd(10_000_000); got.Totals.Cost.Value != want {
		t.Errorf("cost = %d, want %d (dispatches 8 + parent session 2)", got.Totals.Cost.Value, want)
	}
}

// TestRunTimingsShareSumsToWallClock: dispatch shares, parent commands and
// idle add up to wall clock; an overlapped instant is split, never counted
// twice; busy is a row's own union.
func TestRunTimingsShareSumsToWallClock(t *testing.T) {
	t.Parallel()
	got := computeRun(rtInput(), rtAt(200))
	tot := got.Totals
	if tot.WallClockMs.Value != rtMin(160) {
		t.Fatalf("wall clock = %d, want %d", tot.WallClockMs.Value, rtMin(160))
	}
	var share int64
	for _, r := range got.Breakdowns["category"].Rows {
		share += r.ShareMs
	}
	if share != tot.WallClockMs.Value {
		t.Errorf("Σ share = %d, want wall clock %d", share, tot.WallClockMs.Value)
	}
	if tot.AgentTimeMs.Value != rtMin(95) || tot.IdleMs.Value != rtMin(55) {
		t.Errorf("agent time %d, idle %d; want %d, %d", tot.AgentTimeMs.Value, tot.IdleMs.Value, rtMin(95), rtMin(55))
	}
	planned := rtRow(t, got.Breakdowns["category"], "planned implementer")
	if planned.ShareMs != rtMin(42.5) || planned.BusyMs != rtMin(50) {
		t.Errorf("planned implementer share %d busy %d, want %d %d", planned.ShareMs, planned.BusyMs, rtMin(42.5), rtMin(50))
	}
	rev := rtRow(t, got.Breakdowns["category"], "gated reviewer")
	if rev.ShareMs != rtMin(12.5) || rev.BusyMs != rtMin(20) {
		t.Errorf("gated reviewer share %d busy %d, want %d %d", rev.ShareMs, rev.BusyMs, rtMin(12.5), rtMin(20))
	}
}

// TestGuardRunInsideADispatchAddsNoParentTime: a guard run inside a
// dispatch window is the agent's own; a run nested in another collapses
// into it, so parent commands are the outer run alone.
func TestGuardRunInsideADispatchAddsNoParentTime(t *testing.T) {
	t.Parallel()
	got := computeRun(rtInput(), rtAt(200))
	if got.Totals.ParentCommandsMs.Value != rtMin(10) {
		t.Errorf("parent commands = %d, want %d", got.Totals.ParentCommandsMs.Value, rtMin(10))
	}
	if got.Totals.ParentCommandRuns != 1 {
		t.Errorf("parent command runs = %d, want 1", got.Totals.ParentCommandRuns)
	}
	if r := rtRow(t, got.Breakdowns["category"], "parent commands"); r.ShareMs != rtMin(10) || r.Rows != 1 {
		t.Errorf("parent commands row = %+v", r)
	}
}

// TestRunTimingsSplitsAGroupKeyEvenly: task-13+14's cost and time go half
// to each task, marked estimated; a single-task row is not.
func TestRunTimingsSplitsAGroupKeyEvenly(t *testing.T) {
	t.Parallel()
	got := computeRun(rtInput(), rtAt(200))
	byTask := map[string]taskCost{}
	for _, c := range got.CostPerTask {
		byTask[c.Task] = c
	}
	for _, id := range []string{"13", "14"} {
		c := byTask[id]
		if c.Cost != 1_500_000 || c.ShareMs != rtMin(8.75) || !c.Estimated {
			t.Errorf("task %s = %+v, want $1.50, 8.75m, estimated", id, c)
		}
	}
	if c := byTask["1"]; c.Cost != 3_000_000 || c.Rows != 2 || c.Estimated {
		t.Errorf("task 1 = %+v, want $3.00 over 2 rows, measured", c)
	}
}

// TestRunTimingsRepoFromTheKeySuffix: the longest repository suffix wins,
// and a key naming none is change-wide.
func TestRunTimingsRepoFromTheKeySuffix(t *testing.T) {
	t.Parallel()
	repos := repoNames([]string{"/x/gymie-worktrees/kan-1-a", "/x/gymie-frontend-worktrees/kan-1-a/sub", "/x/gymie"})
	for key, want := range map[string]string{
		"verify-fix-3-gymie-frontend": "gymie-frontend",
		"visual-verify-1-gymie":       "gymie",
		"task-4-implementer":          "change-wide",
	} {
		if got := repoOf(key, repos); got != want {
			t.Errorf("repoOf(%q, %v) = %q, want %q", key, repos, got, want)
		}
	}
	got := computeRun(rtInput(), rtAt(200))
	if r := rtRow(t, got.Breakdowns["repo"], "gymie-frontend"); r.Rows != 1 || r.Cost != 1_500_000 {
		t.Errorf("gymie-frontend row = %+v", r)
	}
}

type runTimingsFake struct{}

func (runTimingsFake) RunRecord(_ context.Context, project, change string) (records.Run, error) {
	if change != "kan-1-a" {
		return records.Run{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, project, change)
	}
	return records.Run{Change: change, Dispatches: rtInput().Dispatches}, nil
}

func (runTimingsFake) QueryStageRuns(context.Context, store.Query) ([]store.StageRun, int, error) {
	return rtInput().Stages, 2, nil
}

func (runTimingsFake) ChangeGuardRuns(context.Context, string, string, time.Time, time.Time) ([]records.GuardRun, int, error) {
	return rtInput().Guards, 1, nil
}

// TestRunTimingsRouteServesTheResult: the per-run route answers the
// computed result, and 404 for a change the store does not hold.
func TestRunTimingsRouteServesTheResult(t *testing.T) {
	t.Parallel()
	srv, err := New(config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}, nil, nil, nil, nil, nil, nil, WithRunTimings(runTimingsFake{}))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	resp, err := http.Get(ts.URL + "/api/v1/stats/run/p/kan-1-a")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, k := range []string{"totals", "lanes", "breakdowns", "costPerTask", "costPerVerifyRound"} {
		if v, ok := got[k]; !ok || string(v) == "null" {
			t.Errorf("response lacks %s: %s", k, body)
		}
	}

	resp, err = http.Get(ts.URL + "/api/v1/stats/run/p/kan-404")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown change: status %d, want 404", resp.StatusCode)
	}
}
