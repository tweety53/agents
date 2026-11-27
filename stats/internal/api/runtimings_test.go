package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// UnmarshalJSON reads the dollars MarshalJSON writes back to micro-dollars,
// so the route tests can decode a response into the result types.
func (u *usd) UnmarshalJSON(b []byte) error {
	f, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	*u = usd(math.Round(f * 1e6))
	return nil
}

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
	return records.GuardRun{Guard: "run-verify-list", Worktree: "/x/gymie-worktrees/kan-1-a", RecordedAt: rtAt(to), DurationMS: int(rtMin(to - from))}
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
	in := rtInput()
	// A dispatch starting after the last stage run ended clips to an empty
	// span at the wall clock's end, never past it.
	in.Dispatches = append(in.Dispatches, rtDispatch("task-9-implementer", "implementer", 170, 180, "1"))
	got := computeRun(in, rtAt(200))
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
		if c.Cost != 1_500_000 || c.ShareMs != rtMin(8.75) || !c.Estimated || c.Group != "13+14" {
			t.Errorf("task %s = %+v, want $1.50, 8.75m, estimated", id, c)
		}
	}
	if c := byTask["1"]; c.Cost != 3_000_000 || c.Rows != 2 || c.Estimated {
		t.Errorf("task 1 = %+v, want $3.00 over 2 rows, measured", c)
	}
}

// TestRunTimingsTaskInTwoGroupsKeepsTheFirst: a task two group keys name
// is labelled with the first group it appears in, not the last.
func TestRunTimingsTaskInTwoGroupsKeepsTheFirst(t *testing.T) {
	t.Parallel()
	got := costPerTask([]row{{tasks: []string{"3", "4"}}, {tasks: []string{"3", "5"}}})
	if got[0].Task != "3" || got[0].Group != "3+4" {
		t.Errorf("task 3 = %+v, want group 3+4", got[0])
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

// runTimingsFake serves rtInput for kan-1-a and its first three rows for
// kan-2-b; gone lists a finished change no store holds.
type runTimingsFake struct {
	gone bool
}

func (f runTimingsFake) RunInputs(_ context.Context, refs []store.ChangeRef) ([]store.RunInputs, error) {
	out := make([]store.RunInputs, len(refs))
	for i, r := range refs {
		ds := rtInput().Dispatches
		switch r.Name {
		case "kan-1-a":
		case "kan-2-b":
			ds = ds[:3]
		default:
			return nil, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, r.ProjectKey, r.Name)
		}
		out[i] = store.RunInputs{Dispatches: ds, Stages: rtInput().Stages}
	}
	return out, nil
}

func (f runTimingsFake) ChangeGuardRuns(_ context.Context, ws []store.GuardWindow) ([]store.ChangeGuards, error) {
	out := make([]store.ChangeGuards, len(ws))
	for i := range out {
		out[i] = store.ChangeGuards{Claimed: rtInput().Guards, Unclaimed: 1}
	}
	return out, nil
}

// TestRunTimingsRouteServesTheResult: the per-run route answers the
// computed result, and 404 for a change the store does not hold.
func TestRunTimingsRouteServesTheResult(t *testing.T) {
	t.Parallel()
	srv, err := New(config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}, nil, nil, nil, nil, nil, WithRunTimings(runTimingsFake{}))
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

// rtDataQualityInput is rtInput carrying every gap data quality names:
// rows open, aborted (one with tokens) and blocked, verify-fix,
// visual-verify-fix and panel-fix rounds, a row without tokens with a
// reason and one without, an unlinked row with an other: key, a retimed
// row, an open and a superseded stage run, unclaimed guard runs and a
// tasks.md readable nowhere. openCost prices the open stage run: at 2.5 the
// stage runs' cost equals the dispatch rows' $12.50.
func rtDataQualityInput(openCost string) runInput {
	in := rtInput()
	aborted, abortedRan, blocked := rtDispatch("task-6-implementer", "implementer", 142, 143, ""), rtDispatch("task-5-implementer", "implementer", 140, 142, "1"), rtDispatch("task-7-implementer", "implementer", 143, 144, "1")
	aborted.Outcome, abortedRan.Outcome, blocked.Outcome = "aborted", "aborted", "blocked"
	reasoned := rtDispatch("task-8-implementer", "implementer", 144, 145, "")
	reasoned.Metrics = json.RawMessage(`{"unattributed":{"reason":"harness lost it"}}`)
	in.Dispatches = append(in.Dispatches,
		rtDispatch("visual-verify-fix-2-gymie", "implementer", 130, 135, "1"),
		rtDispatch("panel-fix-1-gymie", "implementer", 135, 140, "1"),
		abortedRan, aborted, blocked, reasoned,
		rtDispatch("mystery-1", "oddball", 145, 146, "0.5"),
	)
	link := int64(1)
	for i := range in.Dispatches {
		if in.Dispatches[i].Key != "mystery-1" {
			in.Dispatches[i].StageRunID = &link
		}
	}
	in.Dispatches[1].Retimed = &records.Retime{Reason: "clock skew", At: rtAt(50), StartedAt: rtAt(19)}
	open := store.StageRun{ID: 3, Stage: "flow.integrate", StartedAt: rtAt(160), Metrics: json.RawMessage(`{"cost_usd":` + openCost + `}`)}
	superseded := rtStage("flow.implement", 0, 10, "0")
	superseded.ID, superseded.Outcome = 4, new("superseded")
	in.Stages = append(in.Stages, open, superseded)
	in.Unclaimed = 3
	in.Findings = []records.Finding{{Severity: "Critical"}, {Severity: "important"}, {Severity: "Minor"}, {Severity: "minor"}}
	in.VerifyRuns = []records.VerifyRun{{Flaky: new(2)}, {}}
	in.Departures = []roundDepartures{{Round: "visual-verify-1-gymie", Departures: 3}}
	in.PlanRead = false
	return in
}

func rtDataQuality(out RunTimings) map[string]dqItem {
	got := map[string]dqItem{}
	for _, d := range out.DataQuality {
		got[d.ID] = d
	}
	return got
}

// TestRunTimingsCountsAndDataQuality: the counts fold fix rounds, findings,
// flaky retries and row outcomes; data quality names every gap the fixture
// carries; the unreconciled-cost item fires from one cent; the journal is
// counted at the state root and reads unknown when it cannot be read.
func TestRunTimingsCountsAndDataQuality(t *testing.T) {
	t.Parallel()
	out := computeRun(rtDataQualityInput("2.5"), rtAt(200))
	c := out.Counts
	if c.FixRounds.Value != (fixRoundCounts{Verify: 2, Panel: 1}) {
		t.Errorf("fix rounds = %+v", c.FixRounds.Value)
	}
	if c.Findings.Value != (severityCounts{Critical: 1, Important: 1, Minor: 2}) {
		t.Errorf("findings = %+v", c.Findings.Value)
	}
	if c.FlakyRetries.Value != (flakyCounts{Retries: 2, Runs: 1}) {
		t.Errorf("flaky = %+v", c.FlakyRetries.Value)
	}
	if c.Rows.Value != (outcomeCounts{Open: 1, Aborted: 2, Blocked: 1}) {
		t.Errorf("rows = %+v", c.Rows.Value)
	}
	if len(c.Departures.Value) != 1 || c.Departures.Value[0].Departures != 3 {
		t.Errorf("departures = %+v", c.Departures.Value)
	}
	if out.Totals.UnclaimedGuardRuns != 3 {
		t.Errorf("unclaimed guard runs total = %d, want 3", out.Totals.UnclaimedGuardRuns)
	}
	want := map[string]int{
		"open-stage-runs": 1, "superseded-stage-runs": 1, "unlinked": 1, "no-tokens": 3, "aborted-ran": 1,
		"uncategorised": 1, "unreadable-openers": 8, "unreconciled-cost": 0, "journal": 0, "retimed": 1,
		"unclaimed-guard-runs": 3,
	}
	got := rtDataQuality(out)
	if len(out.DataQuality) != len(want) {
		t.Errorf("data quality items = %d, want %d", len(out.DataQuality), len(want))
	}
	for id, n := range want {
		if d, ok := got[id]; !ok || d.Count != n {
			t.Errorf("%s = %+v, want count %d", id, d, n)
		}
	}
	for id, detail := range map[string]string{
		"open-stage-runs": "flow.integrate (run 3)", "superseded-stage-runs": "flow.implement (run 4)",
		"unlinked": "mystery-1", "aborted-ran": "task-5-implementer", "uncategorised": "mystery-1 · other:oddball",
		"retimed": "task-1-reviewer", "no-tokens": "task-8-implementer · harness lost it",
	} {
		if !slices.Contains(got[id].Detail, detail) {
			t.Errorf("%s detail = %q, want it to name %q", id, got[id].Detail, detail)
		}
	}
	if !got["journal"].Unknown {
		t.Errorf("journal = %+v, want unknown without a state root", got["journal"])
	}

	planned := rtDataQualityInput("2.5")
	planned.PlanRead = true
	if n := rtDataQuality(computeRun(planned, rtAt(200)))["unreadable-openers"].Count; n != 0 {
		t.Errorf("tasks.md read: unreadable openers = %d, want 0", n)
	}
	for openCost, n := range map[string]int{"2.49": 1, "2.490001": 0} {
		if d := rtDataQuality(computeRun(rtDataQualityInput(openCost), rtAt(200)))["unreconciled-cost"]; d.Count != n {
			t.Errorf("open stage $%s: unreconciled cost = %+v, want count %d", openCost, d, n)
		}
	}
	rtJournalAtStateRoot(t)
}

// rtJournalAtStateRoot: the route counts the change's record journal under
// WithStateRoot, and serves unknown when the journal cannot be read.
func rtJournalAtStateRoot(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	path := fallback.RecordJournalPath(root, "p", "kan-1-a")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{\"a\":1}\n{\"a\":2}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(fallback.RecordJournalPath(root, "p", "kan-2-b"), 0o755); err != nil {
		t.Fatal(err)
	}
	srv, err := New(config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}, nil, nil, nil, nil, nil, WithRunTimings(runTimingsFake{}), WithStateRoot(root))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	for change, want := range map[string]string{"kan-1-a": "2", "kan-2-b": `"unknown"`} {
		resp, err := http.Get(ts.URL + "/api/v1/stats/run/p/" + change)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var got struct {
			Counts struct {
				JournalPending struct {
					Value json.RawMessage `json:"value"`
				} `json:"journalPending"`
			} `json:"counts"`
			DataQuality []dqItem `json:"dataQuality"`
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("%s: decode %s: %v", change, body, err)
		}
		if string(got.Counts.JournalPending.Value) != want {
			t.Errorf("%s: journal pending = %s, want %s", change, got.Counts.JournalPending.Value, want)
		}
		j := rtDataQuality(RunTimings{DataQuality: got.DataQuality})["journal"]
		if wantN := map[string]int{"kan-1-a": 2}[change]; j.Count != wantN || j.Unknown != (change == "kan-2-b") {
			t.Errorf("%s: journal item = %+v", change, j)
		}
	}
}

// TestEveryRunTimingsMetricCarriesADefinition: every object with a value
// in the result carries a non-empty definition and kind.
func TestEveryRunTimingsMetricCarriesADefinition(t *testing.T) {
	b, err := json.Marshal(computeRun(rtInput(), rtAt(200)))
	if err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	n := 0
	var walk func(path string, v any)
	walk = func(path string, v any) {
		switch x := v.(type) {
		case map[string]any:
			if _, ok := x["value"]; ok {
				n++
				if d, _ := x["definition"].(string); d == "" {
					t.Errorf("%s: no definition", path)
				}
			}
			if _, ok := x["definition"]; ok {
				if k, _ := x["kind"].(string); !slices.Contains([]string{kindMeasured, kindDerived, kindEstimated}, k) {
					t.Errorf("%s: kind %q is not measured|derived|estimated", path, k)
				}
			}
			for k, c := range x {
				walk(path+"."+k, c)
			}
		case []any:
			for i, c := range x {
				walk(fmt.Sprintf("%s[%d]", path, i), c)
			}
		}
	}
	walk("$", v)
	var out struct {
		Definitions []definition `json:"definitions"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, d := range out.Definitions {
		if seen[d.Name] {
			t.Errorf("definitions list %q twice", d.Name)
		}
		seen[d.Name] = true
	}
	if n < 12 {
		t.Errorf("metrics found = %d, want at least 12", n)
	}
}

// rtMs is a dispatch of role over [from, to) milliseconds after rtT0.
func rtMs(role string, from, to int64, cost string) records.Dispatch {
	d := rtDispatch("x-"+role, role, 0, 0, cost)
	d.StartedAt = rtT0.Add(time.Duration(from) * time.Millisecond)
	end := rtT0.Add(time.Duration(to) * time.Millisecond)
	d.EndedAt = &end
	return d
}

// TestRunTimingsSweepGivesTheRemainderToTheEarliestRow: two rows sharing
// a 3 ms segment get 2 ms and 1 ms, so the shares still sum to wall clock.
func TestRunTimingsSweepGivesTheRemainderToTheEarliestRow(t *testing.T) {
	t.Parallel()
	st := rtStage("flow.implement", 0, 0, "0")
	end := rtT0.Add(3 * time.Millisecond)
	st.EndedAt = &end
	out := computeRun(runInput{
		Stages:     []store.StageRun{st},
		Dispatches: []records.Dispatch{rtMs("implementer", 0, 3, ""), rtMs("reviewer", 0, 3, "")},
	}, rtAt(10))
	b := out.Breakdowns["role"]
	if a, r := rtRow(t, b, "implementer").ShareMs, rtRow(t, b, "reviewer").ShareMs; a != 2 || r != 1 {
		t.Errorf("shares = %d, %d; want 2, 1", a, r)
	}
}

// TestRunTimingsParentSessionCostFloorsAtZero: dispatch rows costing more
// than the stage runs leave the parent session at zero, not negative.
func TestRunTimingsParentSessionCostFloorsAtZero(t *testing.T) {
	t.Parallel()
	out := computeRun(runInput{
		Stages:     []store.StageRun{rtStage("flow.implement", 0, 10, "1")},
		Dispatches: []records.Dispatch{rtDispatch("task-1-implementer", "implementer", 0, 5, "3")},
	}, rtAt(20))
	if got := out.Totals.ParentSessionCost.Value; got != 0 {
		t.Errorf("parent session cost = %d, want 0", got)
	}
	if got := out.Totals.Cost.Value; got != 3_000_000 {
		t.Errorf("cost = %d, want 3000000", got)
	}
}

// TestRunTimingsCountsRowsWithoutTokens: rtInput's open task-4 row carries
// no tokens and is the one row Totals.NoTokens counts.
func TestRunTimingsCountsRowsWithoutTokens(t *testing.T) {
	t.Parallel()
	if got := computeRun(rtInput(), rtAt(170)).Totals.NoTokens; got != 1 {
		t.Errorf("no tokens = %d, want 1", got)
	}
}

// TestRunTimingsWallClockFallsBackToTheDispatchRows: a run with no stage
// run spans its dispatch rows' extent.
func TestRunTimingsWallClockFallsBackToTheDispatchRows(t *testing.T) {
	t.Parallel()
	out := computeRun(runInput{Dispatches: []records.Dispatch{
		rtDispatch("task-1-implementer", "implementer", 10, 40, "1"),
		rtDispatch("task-2-implementer", "implementer", 20, 60, "1"),
	}}, rtAt(100))
	if got := out.Totals.WallClockMs.Value; got != rtMin(50) || !out.From.Equal(rtAt(10)) {
		t.Errorf("wall clock = %d from %v, want %d from %v", got, out.From, rtMin(50), rtAt(10))
	}
}

// TestRunTimingsRoundsCostToTheNearestMicroDollar: $0.0000017 is 2
// micro-dollars, not the truncated 1.
func TestRunTimingsRoundsCostToTheNearestMicroDollar(t *testing.T) {
	t.Parallel()
	out := computeRun(runInput{
		Stages:     []store.StageRun{rtStage("flow.implement", 0, 10, "0")},
		Dispatches: []records.Dispatch{rtDispatch("task-1-implementer", "implementer", 0, 5, "0.0000017")},
	}, rtAt(20))
	if got := out.Totals.Cost.Value; got != 2 {
		t.Errorf("cost = %d micro-dollars, want 2", got)
	}
}

// TestRunTimingsCountsIdenticalGuardRunsOnce: two guard runs with the same
// span outside every dispatch are one parent-command run.
func TestRunTimingsCountsIdenticalGuardRunsOnce(t *testing.T) {
	t.Parallel()
	out := computeRun(runInput{
		Stages: []store.StageRun{rtStage("flow.implement", 0, 10, "0")},
		Guards: []records.GuardRun{rtGuard(2, 4), rtGuard(2, 4)},
	}, rtAt(20))
	if got := out.Totals.ParentCommandRuns; got != 1 {
		t.Errorf("parent command runs = %d, want 1", got)
	}
}

// The SPA's NOT_CATEGORIES (web/src/api.ts) is the one copy of the rows this
// file draws outside the categories; it must name exactly labelParent and labelIdle.
func TestSPANotCategoriesMatchesTheServerLabels(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "web", "src", "api.ts"))
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("NOT_CATEGORIES = new Set([%q, %q])", labelParent, labelIdle)
	if !strings.Contains(string(src), want) {
		t.Fatalf("web/src/api.ts lacks %s", want)
	}
}

// TestRunTimingsSumsStoredAgentSplits: totals.agentTimeSplit sums the
// dispatches' stored time splits, per tool included; a dispatch with none
// adds nothing, and with none stored at all the field is absent -- never
// recomputed from a transcript.
func TestRunTimingsSumsStoredAgentSplits(t *testing.T) {
	t.Parallel()
	in := rtInput()
	in.Dispatches[0].TimeSplit = json.RawMessage(`{"llmMs":100,"toolMs":30,"humanMs":5,"byTool":{"Bash make":20,"Read":10}}`)
	in.Dispatches[2].TimeSplit = json.RawMessage(`{"llmMs":50,"toolMs":7,"humanMs":0,"byTool":{"Bash make":7}}`)
	got, err := json.Marshal(computeRun(in, rtAt(200)).Totals.AgentTimeSplit)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"llmMs":150,"toolMs":37,"humanMs":5,"byTool":{"Bash make":27,"Read":10}}`; string(got) != want {
		t.Errorf("agentTimeSplit = %s, want %s", got, want)
	}
	if s := computeRun(rtInput(), rtAt(200)).Totals.AgentTimeSplit; s != nil {
		t.Errorf("none stored: agentTimeSplit = %+v, want absent", s)
	}
}

// TestRunTimingsSplitsParentIdle: totals.idleSplit is the parent
// transcript's classified gaps kept only where they lie inside an idle /
// parent segment -- a gap straddling a dispatch counts only its idle part --
// and parentGaps reads every linked transcript but a subagent's; with no
// linked parent transcript the field is absent.
func TestRunTimingsSplitsParentIdle(t *testing.T) {
	t.Parallel()
	gap := func(class, tool string, from, to float64) harvest.Gap {
		return harvest.Gap{From: rtAt(from), To: rtAt(to), Class: class, Tool: tool}
	}
	in := runInput{
		Stages:     []store.StageRun{rtStage("flow.implement", 0, 100, "1")},
		Dispatches: []records.Dispatch{rtDispatch("task-1-implementer", "implementer", 20, 40, "1")},
		ParentGaps: []harvest.Gap{
			gap(harvest.GapLLM, "", 10, 30), gap(harvest.GapTool, "Agent", 30, 60), gap(harvest.GapHuman, "", 90, 110),
		},
	}
	got := computeRun(in, rtAt(200))
	want := harvest.TimeSplit{LLMMs: rtMin(10), ToolMs: rtMin(20), HumanMs: rtMin(10), ByTool: map[string]int64{"Agent": rtMin(20)}}
	if got.Totals.IdleMs.Value != rtMin(80) || got.Totals.IdleSplit == nil || !reflect.DeepEqual(*got.Totals.IdleSplit, want) {
		t.Errorf("idle %d split %+v, want 80 min and %+v", got.Totals.IdleMs.Value, got.Totals.IdleSplit, want)
	}
	in.ParentGaps = nil
	if s := computeRun(in, rtAt(200)).Totals.IdleSplit; s != nil {
		t.Errorf("no parent transcript: idleSplit = %+v, want absent", s)
	}

	dir := t.TempDir()
	parent, sub := filepath.Join(dir, "s.jsonl"), filepath.Join(dir, "s", "subagents", "agent-a.jsonl")
	if err := os.MkdirAll(filepath.Dir(sub), 0o755); err != nil {
		t.Fatal(err)
	}
	line := func(typ string, min float64) string {
		return `{"type":"` + typ + `","timestamp":` + strconv.Quote(rtAt(min).Format(time.RFC3339Nano)) + `,"message":{"content":"x"}}` + "\n"
	}
	if err := os.WriteFile(parent, []byte(line("user", 0)+line("assistant", 5)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sub, []byte(line("user", 0)+line("assistant", 7)), 0o644); err != nil {
		t.Fatal(err)
	}
	stages := []store.StageRun{{TranscriptPath: &parent}, {TranscriptPath: &parent}, {TranscriptPath: &sub}, {}}
	if gaps := parentGaps(stages, nil); len(gaps) != 1 || !gaps[0].To.Equal(rtAt(5)) {
		t.Errorf("parentGaps = %+v, want the parent transcript's one gap", gaps)
	}
	if gaps := parentGaps(stages[2:], nil); gaps != nil {
		t.Errorf("parentGaps of subagent-only links = %+v, want nil", gaps)
	}
}
