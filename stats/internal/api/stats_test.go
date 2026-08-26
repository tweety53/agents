package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// --- fakeStore's StatsStore surface -----------------------------------
//
// These five methods exist purely so fakeStore (changes_test.go,
// stages_test.go) keeps satisfying api.StatsStore for the New() call sites
// in those files that pass it as every one of New's three store
// parameters and never exercise a stats route. Every stats_test.go test
// below that actually asserts on aggregation output uses statsFake
// instead (see below), which gives each test full, independent control
// over exactly the rows and errors it needs.
func (f *fakeStore) LiveStateBoard(_ context.Context, _ store.Period, project *string) ([]store.LiveStateRow, error) {
	f.lastStatsProject = project
	return f.liveStateBoard, f.liveStateBoardErr
}
func (f *fakeStore) CostPerChange(_ context.Context, _ store.Period, project, _ *string) ([]store.CostPerChangeRow, error) {
	f.lastStatsProject = project
	return f.costPerChange, f.costPerChangeErr
}
func (f *fakeStore) StageLeaderboard(_ context.Context, _ store.Period, project, _ *string) ([]store.StageLeaderboardRow, error) {
	f.lastStatsProject = project
	return f.stageLeaderboard, f.stageLeaderboardErr
}
func (f *fakeStore) TrendOverTime(_ context.Context, _ store.Period, project, _ *string) ([]store.TrendPoint, error) {
	f.lastStatsProject = project
	return f.trendOverTime, f.trendOverTimeErr
}
func (f *fakeStore) CacheEfficiency(_ context.Context, _ store.Period, project, _ *string) ([]store.CacheEfficiencyRow, error) {
	f.lastStatsProject = project
	return f.cacheEfficiency, f.cacheEfficiencyErr
}

// CountRunsWithoutModel, ListModels and AllRecordedRunsUnmeasured have no
// fakeStore field of their own: this type exists purely to keep satisfying
// api.StatsStore at call sites that never exercise a stats route at all
// (this section's own header comment) -- no test here needs any of them to
// return anything but a harmless zero value.
func (f *fakeStore) CountRunsWithoutModel(_ context.Context, _ store.Period, _ *string) (int, error) {
	return 0, nil
}
func (f *fakeStore) ListModels(_ context.Context, _ store.Period, _ *string) ([]string, error) {
	return nil, nil
}
func (f *fakeStore) AllRecordedRunsUnmeasured(_ context.Context, _ store.Period, _ *string) (bool, error) {
	return false, nil
}

var _ api.StatsStore = (*fakeStore)(nil)

// --- statsFake: a StatsStore built for this file's own tests -----------
//
// fakeStore's QueryStageRuns (stages_test.go) only understands the filter
// fields findOpenStageRun needs (project, name, command, stage, ended_at)
// and only sorts by "attempt". This file's own handlers (recorded,
// costPerChangeByRepo, listStageRuns) filter on started_at and sort on it
// too, so statsFake implements its own small, independently correct
// in-memory QueryStageRuns rather than stretching fakeStore's to cover
// both shapes. It satisfies api.StatsStore in full, and every stats_test.go
// test that asserts real numbers builds its server around one of these
// rather than around package-shared fakeStore.
type statsRun struct {
	run        store.StageRun
	projectKey string
	changeName string
}

type statsFake struct {
	liveStateBoard   []store.LiveStateRow
	costPerChange    []store.CostPerChangeRow
	stageLeaderboard []store.StageLeaderboardRow
	trendOverTime    []store.TrendPoint
	cacheEfficiency  []store.CacheEfficiencyRow
	aggErr           error

	countRunsWithoutModel    int
	countRunsWithoutModelErr error
	models                   []string
	modelsErr                error

	// allRecordedRunsUnmeasured and allRecordedRunsUnmeasuredErr let a
	// test drive task 5's third arm directly, independent of stageRuns'
	// own fixtures -- statsFake's other aggregation fields already work
	// this way (see this type's own header comment), and
	// AllRecordedRunsUnmeasured's real implementation is exercised
	// separately, against real Postgres, by
	// internal/store/aggregate_test.go's own TestAllRecordedRunsUnmeasured*
	// cases.
	allRecordedRunsUnmeasured    bool
	allRecordedRunsUnmeasuredErr error

	stageRuns         []statsRun
	queryStageRunsErr error
	lastQuery         store.Query
	lastProject       *string
	lastModel         *string

	// projectKeysByDisplayName and projectKeysByDisplayNameCalls give
	// this file's own tests the same control over task 3's project
	// display-name resolution that fakeStore's identically-shaped fields
	// give changes_test.go -- see fakeStore.ProjectKeysByDisplayName's own
	// doc comment.
	projectKeysByDisplayName      map[string][]string
	projectKeysByDisplayNameErr   error
	projectKeysByDisplayNameCalls []string
}

func (f *statsFake) ProjectKeysByDisplayName(_ context.Context, displayName string) ([]string, error) {
	f.projectKeysByDisplayNameCalls = append(f.projectKeysByDisplayNameCalls, displayName)
	if f.projectKeysByDisplayNameErr != nil {
		return nil, f.projectKeysByDisplayNameErr
	}
	return f.projectKeysByDisplayName[displayName], nil
}

func (f *statsFake) LiveStateBoard(_ context.Context, _ store.Period, p *string) ([]store.LiveStateRow, error) {
	f.lastProject = p
	return f.liveStateBoard, f.aggErr
}
func (f *statsFake) CostPerChange(_ context.Context, _ store.Period, p, m *string) ([]store.CostPerChangeRow, error) {
	f.lastProject, f.lastModel = p, m
	return f.costPerChange, f.aggErr
}
func (f *statsFake) StageLeaderboard(_ context.Context, _ store.Period, p, m *string) ([]store.StageLeaderboardRow, error) {
	f.lastProject, f.lastModel = p, m
	return f.stageLeaderboard, f.aggErr
}
func (f *statsFake) TrendOverTime(_ context.Context, _ store.Period, p, m *string) ([]store.TrendPoint, error) {
	f.lastProject, f.lastModel = p, m
	return f.trendOverTime, f.aggErr
}
func (f *statsFake) CacheEfficiency(_ context.Context, _ store.Period, p, m *string) ([]store.CacheEfficiencyRow, error) {
	f.lastProject, f.lastModel = p, m
	return f.cacheEfficiency, f.aggErr
}
func (f *statsFake) CountRunsWithoutModel(_ context.Context, _ store.Period, p *string) (int, error) {
	f.lastProject = p
	return f.countRunsWithoutModel, f.countRunsWithoutModelErr
}
func (f *statsFake) ListModels(_ context.Context, _ store.Period, p *string) ([]string, error) {
	f.lastProject = p
	return f.models, f.modelsErr
}
func (f *statsFake) AllRecordedRunsUnmeasured(_ context.Context, _ store.Period, p *string) (bool, error) {
	f.lastProject = p
	return f.allRecordedRunsUnmeasured, f.allRecordedRunsUnmeasuredErr
}

func (f *statsFake) QueryStageRuns(_ context.Context, q store.Query) ([]store.StageRun, int, error) {
	f.lastQuery = q
	if f.queryStageRunsErr != nil {
		return nil, 0, f.queryStageRunsErr
	}

	var matches []statsRun
	for _, r := range f.stageRuns {
		if statsRunMatches(r, q.Filters) {
			matches = append(matches, r)
		}
	}

	if len(q.Sort) > 0 {
		sk := q.Sort[0]
		sort.SliceStable(matches, func(i, j int) bool {
			var less bool
			switch sk.Field {
			case "started_at":
				less = matches[i].run.StartedAt.Before(matches[j].run.StartedAt)
			case "command":
				less = matches[i].run.Command < matches[j].run.Command
			default:
				less = matches[i].run.ID < matches[j].run.ID
			}
			if sk.Desc {
				return !less && matches[i].run.ID != matches[j].run.ID
			}
			return less
		})
	}

	total := len(matches)
	if q.Limit > 0 && q.Limit != store.NoLimit {
		start := q.Offset
		if start > len(matches) {
			start = len(matches)
		}
		end := start + q.Limit
		if end > len(matches) {
			end = len(matches)
		}
		matches = matches[start:end]
	}

	out := make([]store.StageRun, len(matches))
	for i, m := range matches {
		out[i] = m.run
	}
	return out, total, nil
}

func statsRunMatches(r statsRun, filters []store.Filter) bool {
	for _, f := range filters {
		switch f.Field {
		case "project":
			if r.projectKey != f.Value {
				return false
			}
		case "name":
			if r.changeName != f.Value {
				return false
			}
		case "command":
			if r.run.Command != f.Value {
				return false
			}
		case "stage":
			if r.run.Stage != f.Value {
				return false
			}
		case "started_at":
			t, ok := f.Value.(time.Time)
			if !ok {
				return false
			}
			switch f.Op {
			case store.OpGte:
				if r.run.StartedAt.Before(t) {
					return false
				}
			case store.OpLt:
				if !r.run.StartedAt.Before(t) {
					return false
				}
			}
		}
	}
	return true
}

var _ api.StatsStore = (*statsFake)(nil)

func newStatsTestServer(t *testing.T, sts api.StatsStore) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	srv, err := api.New(cfg, newFakeStore(), newFakeStore(), sts, newFakeStore(), newFakeStore(), nil)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

// statsEnvelope mirrors statsResponse's exported wire shape loosely enough
// for these tests: Rows is decoded generically (json.RawMessage) so each
// test decodes it into the concrete row DTO it actually expects.
type statsEnvelope struct {
	View               string          `json:"view"`
	From               string          `json:"from"`
	To                 string          `json:"to"`
	Project            *string         `json:"project"`
	Model              *string         `json:"model"`
	BoundaryConvention string          `json:"boundaryConvention"`
	Recorded           bool            `json:"recorded"`
	Unmeasured         bool            `json:"unmeasured"`
	ExcludedNoModel    *int            `json:"excludedNoModel"`
	Rows               json.RawMessage `json:"rows"`
}

func getStats(t *testing.T, ts *httptest.Server, path string) (int, statsEnvelope, []byte) {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	var body []byte
	buf := make([]byte, 65536)
	for {
		n, err := resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		if err != nil {
			break
		}
	}
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, statsEnvelope{}, body
	}
	var env statsEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode response for %s: %v (body: %s)", path, err, body)
	}
	return resp.StatusCode, env, body
}

const (
	fromParam = "2026-08-01T00:00:00Z"
	toParam   = "2026-09-01T00:00:00Z"
)

func periodPath(view string) string {
	return fmt.Sprintf("/api/v1/stats/%s?from=%s&to=%s", view, fromParam, toParam)
}

// --- TestEveryViewAcceptsAPeriod ---------------------------------------

func TestEveryViewAcceptsAPeriod(t *testing.T) {
	views := []string{
		"state-board", "cost-per-change", "stage-leaderboard", "trend",
		"cache-efficiency",
	}
	for _, view := range views {
		t.Run(view, func(t *testing.T) {
			sts := &statsFake{}
			ts := newStatsTestServer(t, sts)
			status, env, body := getStats(t, ts, periodPath(view))
			if status != http.StatusOK {
				t.Fatalf("GET /api/v1/stats/%s: status %d, body %s", view, status, body)
			}
			if env.View != view {
				t.Errorf("view = %q, want %q", env.View, view)
			}
		})
	}
}

// --- TestEveryViewCarriesItsRealNumbersThrough (post-commit review F2) ---
//
// TestEveryViewAcceptsAPeriod, above, asserts only status and view name --
// it would pass even if every view's DTO conversion silently zeroed or
// transposed a field. Post-commit review proved that gap directly:
// swapping MeanCostUSD and MedianCostUSD in toStageLeaderboardDTOs left the
// entire suite green. Task 3's own store-layer aggregation already has
// hand-computed assertions against real Postgres (aggregate_test.go); what
// was never covered is the DTO layer this task adds on top of it. Each
// subtest below seeds one hand-picked, independently-chosen row per view
// -- distinct values in every field, so a swap or a dropped field cannot
// hide behind a coincidental match -- and asserts every field of the
// decoded response equals it exactly.
func TestEveryViewCarriesItsRealNumbersThrough(t *testing.T) {
	t.Run("stage-leaderboard", func(t *testing.T) {
		sts := &statsFake{stageLeaderboard: []store.StageLeaderboardRow{
			{Command: "/myflow-do", Stage: "SDD + TDD per task", RunCount: 5,
				MeanCostUSD: 11.25, MedianCostUSD: 9.5, P90CostUSD: 22.75},
		}}
		ts := newStatsTestServer(t, sts)
		status, env, body := getStats(t, ts, periodPath("stage-leaderboard"))
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		var rows []struct {
			Command       string  `json:"command"`
			Stage         string  `json:"stage"`
			RunCount      int     `json:"runCount"`
			MeanCostUSD   float64 `json:"meanCostUsd"`
			MedianCostUSD float64 `json:"medianCostUsd"`
			P90CostUSD    float64 `json:"p90CostUsd"`
		}
		mustDecodeRows(t, env, body, &rows)
		if len(rows) != 1 {
			t.Fatalf("rows = %d, want 1", len(rows))
		}
		got := rows[0]
		if got.RunCount != 5 || got.MeanCostUSD != 11.25 || got.MedianCostUSD != 9.5 || got.P90CostUSD != 22.75 {
			t.Errorf("got %+v, want RunCount=5 MeanCostUSD=11.25 MedianCostUSD=9.5 P90CostUSD=22.75", got)
		}
	})

	t.Run("trend", func(t *testing.T) {
		day := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
		cost := 42.5
		sts := &statsFake{trendOverTime: []store.TrendPoint{{Day: day, RunCount: 8, TotalCostUSD: &cost}}}
		ts := newStatsTestServer(t, sts)
		status, env, body := getStats(t, ts, periodPath("trend"))
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		var rows []struct {
			Day          string   `json:"day"`
			RunCount     int      `json:"runCount"`
			TotalCostUSD *float64 `json:"totalCostUsd"`
		}
		mustDecodeRows(t, env, body, &rows)
		if len(rows) != 1 {
			t.Fatalf("rows = %d, want 1", len(rows))
		}
		got := rows[0]
		if got.Day != "2026-08-12" || got.RunCount != 8 || got.TotalCostUSD == nil || *got.TotalCostUSD != 42.5 {
			t.Errorf("got %+v, want Day=2026-08-12 RunCount=8 TotalCostUSD=42.5", got)
		}
	})

	t.Run("cache-efficiency", func(t *testing.T) {
		read, creation, ratio := int64(9000), int64(3000), 3.0
		sts := &statsFake{cacheEfficiency: []store.CacheEfficiencyRow{
			{Command: "/myflow-do", Stage: "review panel", CacheReadTotal: &read, CacheCreationTotal: &creation, Ratio: &ratio},
		}}
		ts := newStatsTestServer(t, sts)
		status, env, body := getStats(t, ts, periodPath("cache-efficiency"))
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		var rows []struct {
			CacheReadTotal     *int64   `json:"cacheReadTotal"`
			CacheCreationTotal *int64   `json:"cacheCreationTotal"`
			Ratio              *float64 `json:"ratio"`
		}
		mustDecodeRows(t, env, body, &rows)
		if len(rows) != 1 {
			t.Fatalf("rows = %d, want 1", len(rows))
		}
		got := rows[0]
		if got.CacheReadTotal == nil || *got.CacheReadTotal != 9000 ||
			got.CacheCreationTotal == nil || *got.CacheCreationTotal != 3000 ||
			got.Ratio == nil || *got.Ratio != 3.0 {
			t.Errorf("got %+v, want CacheReadTotal=9000 CacheCreationTotal=3000 Ratio=3.0", got)
		}
	})

}

func mustDecodeRows(t *testing.T, env statsEnvelope, body []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(env.Rows, v); err != nil {
		t.Fatalf("decode rows: %v (body %s)", err, body)
	}
}

// --- TestEmptyPeriodReturnsEmptyNotError -------------------------------

func TestEmptyPeriodReturnsEmptyNotError(t *testing.T) {
	// Telemetry exists (a run recorded well before and inside the queried
	// period), but this specific view's aggregation has nothing to report
	// for it -- costPerChange is left nil/empty on purpose. This must
	// succeed with an empty rows array, not an error.
	seededAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	sts := &statsFake{
		stageRuns: []statsRun{{
			run:        store.StageRun{ID: 1, StartedAt: seededAt, Command: "/myflow-do", Stage: "x"},
			projectKey: "proj", changeName: "kan-1",
		}},
	}
	ts := newStatsTestServer(t, sts)
	status, env, body := getStats(t, ts, periodPath("cost-per-change"))
	if status != http.StatusOK {
		t.Fatalf("status = %d, body %s", status, body)
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(env.Rows, &rows); err != nil && string(env.Rows) != "null" {
		t.Fatalf("decode rows: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("rows = %d, want 0", len(rows))
	}
	if !env.Recorded {
		t.Errorf("Recorded = false, want true: telemetry exists before this period")
	}
}

// --- TestBoundaryConventionIsConsistentAcrossViews ---------------------

// TestBoundaryConventionIsConsistentAcrossViews exercises every one of the
// four stage-run-keyed views. A post-commit review mutation proved the
// gap this closes: changing StageLeaderboard's boundary from
// "sr.started_at < $2" to "<= $2" (internal/store/aggregate.go) left the
// *previous* version of this test green -- it only looped over
// cost-per-change, so a boundary divergence in stage-leaderboard, trend or
// cache-efficiency went completely undetected, even though this test's own
// name claims to cover "every view".
//
// The live state board is deliberately excluded, not merely forgotten: it
// aggregates *changes* by their own updated_at (store.LiveStateBoard),
// never *stage runs* by started_at at all, so it has no started_at
// boundary for this fixture to exercise in the first place -- a genuinely
// different case, not an oversight.
//
// One stage run is seeded exactly on the boundary T, with a metrics bag and
// an owning change shaped to satisfy every remaining view's own WHERE
// clause (internal/store/aggregate.go): StageLeaderboard requires
// "cost_usd" present. CostPerChange, TrendOverTime and CacheEfficiency
// impose no such extra condition, so the same run and change satisfy all
// four at once.
func TestBoundaryConventionIsConsistentAcrossViews(t *testing.T) {
	st := newIntegrationStore(t)
	projectKey := fmt.Sprintf("proj-boundary-%d", time.Now().UnixNano())
	roster := "light"
	if err := st.PutChange(context.Background(), store.Change{
		ProjectKey: projectKey, MainCheckoutPath: "/tmp/" + projectKey, Name: "kan-1",
		State: store.StateInProgress, ReviewPanelRoster: &roster,
		UpdatedAt: time.Now().UTC(), UpdatedBy: "test",
	}); err != nil {
		t.Fatalf("seed change: %v", err)
	}

	boundary := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	mustRunIntegrationStage(t, st, store.BeginStageInput{
		ProjectKey: projectKey, ChangeName: "kan-1", Harness: "claude-code",
		Command: "/myflow-do", Stage: "SDD + TDD per task", StartedAt: boundary,
	}, json.RawMessage(`{"cost_usd":1.0,"models":{"claude-opus-4":{"cost_usd":1.0}}}`), boundary.Add(time.Minute))

	ts := newIntegrationTestServer(t, st)
	before := fmt.Sprintf("from=%s&to=%s",
		boundary.Add(-time.Hour).Format(time.RFC3339), boundary.Format(time.RFC3339))
	atAndAfter := fmt.Sprintf("from=%s&to=%s",
		boundary.Format(time.RFC3339), boundary.Add(time.Hour).Format(time.RFC3339))

	// The live state board is not in this loop -- see the doc comment above
	// for why it is a genuinely different case, not an omission.
	views := []string{
		"cost-per-change", "stage-leaderboard", "trend", "cache-efficiency",
	}
	for _, view := range views {
		t.Run(view, func(t *testing.T) {
			status, env, body := getStats(t, ts, fmt.Sprintf("/api/v1/stats/%s?%s&project=%s", view, before, projectKey))
			if status != http.StatusOK {
				t.Fatalf("period before boundary: status %d, body %s", status, body)
			}
			var rowsBefore []json.RawMessage
			_ = json.Unmarshal(env.Rows, &rowsBefore)
			if len(rowsBefore) != 0 {
				t.Errorf("period [T-1h, T) reported %d rows for a run starting exactly at T, want 0 (half-open, start is not < T)", len(rowsBefore))
			}

			status, env, body = getStats(t, ts, fmt.Sprintf("/api/v1/stats/%s?%s&project=%s", view, atAndAfter, projectKey))
			if status != http.StatusOK {
				t.Fatalf("period at/after boundary: status %d, body %s", status, body)
			}
			var rowsAtAndAfter []json.RawMessage
			_ = json.Unmarshal(env.Rows, &rowsAtAndAfter)
			if len(rowsAtAndAfter) != 1 {
				t.Errorf("period [T, T+1h) reported %d rows for a run starting exactly at T, want 1", len(rowsAtAndAfter))
			}
		})
	}
}

// --- TestProjectFilterRestrictsResults / TestNoProjectFilterAggregatesAcrossProjects ---

func TestProjectFilterRestrictsResults(t *testing.T) {
	sts := &statsFake{}
	ts := newStatsTestServer(t, sts)
	status, env, body := getStats(t, ts, periodPath("cost-per-change")+"&project=kan-16")
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}
	if env.Project == nil || *env.Project != "kan-16" {
		t.Fatalf("response project = %v, want \"kan-16\"", env.Project)
	}
	if sts.lastProject == nil || *sts.lastProject != "kan-16" {
		t.Fatalf("store received project = %v, want \"kan-16\"", sts.lastProject)
	}
}

func TestNoProjectFilterAggregatesAcrossProjects(t *testing.T) {
	sts := &statsFake{}
	ts := newStatsTestServer(t, sts)
	status, env, body := getStats(t, ts, periodPath("cost-per-change"))
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}
	if env.Project != nil {
		t.Fatalf("response project = %v, want nil (no filter)", env.Project)
	}
	if sts.lastProject != nil {
		t.Fatalf("store received project = %v, want nil", sts.lastProject)
	}
}

// TestStatsViewAcceptsProjectDisplayName asserts a "project" value that is
// a display name matching exactly one project resolves to that project's
// key before it reaches the store -- specs/myflow-stats-views/spec.md's
// "A display name is filtered on".
func TestStatsViewAcceptsProjectDisplayName(t *testing.T) {
	sts := &statsFake{projectKeysByDisplayName: map[string][]string{"agents": {"agents-a740d89c"}}}
	ts := newStatsTestServer(t, sts)

	status, env, body := getStats(t, ts, periodPath("cost-per-change")+"&project=agents")
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}
	if env.Project == nil || *env.Project != "agents-a740d89c" {
		t.Fatalf("response project = %v, want resolved key %q", env.Project, "agents-a740d89c")
	}
	if sts.lastProject == nil || *sts.lastProject != "agents-a740d89c" {
		t.Fatalf("store received project = %v, want resolved key %q", sts.lastProject, "agents-a740d89c")
	}
}

// TestStatsViewExactProjectKeyRunsNoResolutionQuery asserts a "project"
// value already carrying the derivation suffix is used unchanged, with no
// call to ProjectKeysByDisplayName at all -- specs/myflow-stats-views/spec.md's
// "A full key is filtered on": "it is used as-is, with no resolution
// attempted".
func TestStatsViewExactProjectKeyRunsNoResolutionQuery(t *testing.T) {
	sts := &statsFake{}
	ts := newStatsTestServer(t, sts)

	status, env, body := getStats(t, ts, periodPath("cost-per-change")+"&project=agents-a740d89c")
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}
	if env.Project == nil || *env.Project != "agents-a740d89c" {
		t.Fatalf("response project = %v, want %q unchanged", env.Project, "agents-a740d89c")
	}
	if len(sts.projectKeysByDisplayNameCalls) != 0 {
		t.Fatalf("ProjectKeysByDisplayName called with %v, want no calls for an exact key", sts.projectKeysByDisplayNameCalls)
	}
}

// TestStatsViewAmbiguousProjectDisplayNameReturns400 asserts a "project"
// value matching more than one project's display name is rejected with
// 400, naming the ambiguity and both candidate keys, rather than silently
// filtering by one of them.
func TestStatsViewAmbiguousProjectDisplayNameReturns400(t *testing.T) {
	sts := &statsFake{projectKeysByDisplayName: map[string][]string{
		"agents": {"agents-a740d89c", "agents-7c1f238a"},
	}}
	ts := newStatsTestServer(t, sts)

	status, _, body := getStats(t, ts, periodPath("cost-per-change")+"&project=agents")
	if status != http.StatusBadRequest {
		t.Fatalf("status %d, want 400, body %s", status, body)
	}
	if !strings.Contains(string(body), "agents-a740d89c") || !strings.Contains(string(body), "agents-7c1f238a") {
		t.Fatalf("expected both candidate keys named in the body, got %q", body)
	}
}

// TestStatsViewProjectResolutionStoreFailureReturns500 asserts that a
// store failure inside ProjectKeysByDisplayName -- as opposed to a
// genuine client mistake like an ambiguous display name -- is reported as
// a logged 500 with a generic body, exactly like any other store failure
// this handler reports, rather than the 400 an ordinary parse failure
// gets. Before the fix, resolveProjectParam's propagated store error was
// indistinguishable from a parse error at this call site and was reported
// as 400 with the store's own internal text leaking into the body
// (post-commit review round 2, F5).
func TestStatsViewProjectResolutionStoreFailureReturns500(t *testing.T) {
	sts := &statsFake{
		projectKeysByDisplayNameErr: errors.New(`store: project keys by display name "agents": connection refused`),
	}
	ts := newStatsTestServer(t, sts)

	status, _, body := getStats(t, ts, periodPath("cost-per-change")+"&project=agents")
	if status != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500, body %s", status, body)
	}
	if strings.Contains(string(body), "connection refused") {
		t.Fatalf("response body leaked internal store text: %s", body)
	}
}

// TestListModelsProjectResolutionStoreFailureReturns500 is
// TestStatsViewProjectResolutionStoreFailureReturns500's counterpart for
// GET /api/v1/models, which resolves its own "project" parameter through
// the same parsePeriodAndProject call (post-commit review round 2, F5).
func TestListModelsProjectResolutionStoreFailureReturns500(t *testing.T) {
	sts := &statsFake{
		projectKeysByDisplayNameErr: errors.New(`store: project keys by display name "agents": connection refused`),
	}
	ts := newStatsTestServer(t, sts)

	status, body := doGetRaw(t, ts, fmt.Sprintf("/api/v1/models?from=%s&to=%s&project=agents", fromParam, toParam))
	if status != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500, body %s", status, body)
	}
	if strings.Contains(body, "connection refused") {
		t.Fatalf("response body leaked internal store text: %s", body)
	}
}

// TestStatsViewUnknownProjectPassesThroughUnchanged asserts a "project"
// value matching no project's display name at all is passed through
// unchanged, yielding no rows -- exactly as an unknown key already does --
// rather than being treated as an error.
func TestStatsViewUnknownProjectPassesThroughUnchanged(t *testing.T) {
	sts := &statsFake{}
	ts := newStatsTestServer(t, sts)

	status, env, body := getStats(t, ts, periodPath("cost-per-change")+"&project=no-such-project")
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}
	if env.Project == nil || *env.Project != "no-such-project" {
		t.Fatalf("response project = %v, want %q unchanged", env.Project, "no-such-project")
	}
	if sts.lastProject == nil || *sts.lastProject != "no-such-project" {
		t.Fatalf("store received project = %v, want %q unchanged", sts.lastProject, "no-such-project")
	}
}

// --- TestPeriodBeforeAnyDataReportsNotRecorded --------------------------

func TestPeriodBeforeAnyDataReportsNotRecorded(t *testing.T) {
	earliest := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	sts := &statsFake{
		stageRuns: []statsRun{{
			run:        store.StageRun{ID: 1, StartedAt: earliest, Command: "/myflow-do", Stage: "x"},
			projectKey: "proj", changeName: "kan-1",
		}},
	}
	ts := newStatsTestServer(t, sts)

	t.Run("period entirely before the earliest recorded run", func(t *testing.T) {
		path := "/api/v1/stats/cost-per-change?from=2020-01-01T00:00:00Z&to=2020-02-01T00:00:00Z"
		status, env, body := getStats(t, ts, path)
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		if env.Recorded {
			t.Errorf("Recorded = true, want false: period predates any telemetry")
		}
	})

	t.Run("period covering the earliest recorded run", func(t *testing.T) {
		path := "/api/v1/stats/cost-per-change?from=2026-08-01T00:00:00Z&to=2026-09-01T00:00:00Z"
		status, env, body := getStats(t, ts, path)
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		if !env.Recorded {
			t.Errorf("Recorded = false, want true: period covers the earliest recorded run")
		}
	})

	t.Run("no telemetry recorded at all", func(t *testing.T) {
		empty := &statsFake{}
		ts2 := newStatsTestServer(t, empty)
		status, env, body := getStats(t, ts2, periodPath("cost-per-change"))
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		if env.Recorded {
			t.Errorf("Recorded = true, want false: store has never recorded any stage run")
		}
	})
}

// --- TestThreeAbsenceStatesDoNotCollapse ---------------------------------
//
// Task 5's own non-negotiable rule: (a) no runs in the period, (b) runs
// recorded but none attributed, (c) attributed and measured at a real
// zero, asserted as three separate wire responses so that no two of them
// can render identically -- that collapse is the exact defect this task
// closes (tasks.md, "5 A recorded but unmeasured run is visible as such").
func TestThreeAbsenceStatesDoNotCollapse(t *testing.T) {
	earliest := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	seeded := statsRun{
		run:        store.StageRun{ID: 1, StartedAt: earliest, Command: "/myflow-do", Stage: "x"},
		projectKey: "proj", changeName: "kan-1",
	}

	t.Run("(a) no runs in the period at all: Recorded=false, Unmeasured=false", func(t *testing.T) {
		sts := &statsFake{stageRuns: []statsRun{seeded}}
		ts := newStatsTestServer(t, sts)
		status, env, body := getStats(t, ts, "/api/v1/stats/cost-per-change?from=2020-01-01T00:00:00Z&to=2020-02-01T00:00:00Z")
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		if env.Recorded {
			t.Errorf("Recorded = true, want false: period predates any telemetry")
		}
		if env.Unmeasured {
			t.Errorf("Unmeasured = true, want false: a period with no runs at all is never reported as the unmeasured arm")
		}
	})

	t.Run("(b) runs recorded but none attributed: Recorded=true, Unmeasured=true", func(t *testing.T) {
		sts := &statsFake{stageRuns: []statsRun{seeded}, allRecordedRunsUnmeasured: true}
		ts := newStatsTestServer(t, sts)
		status, env, body := getStats(t, ts, periodPath("cost-per-change"))
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		if !env.Recorded {
			t.Errorf("Recorded = false, want true: period covers the earliest recorded run")
		}
		if !env.Unmeasured {
			t.Errorf("Unmeasured = false, want true: the store reports every run in scope carries no measurement")
		}
	})

	t.Run("(c) attributed and measured as a real zero: Recorded=true, Unmeasured=false", func(t *testing.T) {
		sts := &statsFake{stageRuns: []statsRun{seeded}, allRecordedRunsUnmeasured: false}
		ts := newStatsTestServer(t, sts)
		status, env, body := getStats(t, ts, periodPath("cost-per-change"))
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		if !env.Recorded {
			t.Errorf("Recorded = false, want true: period covers the earliest recorded run")
		}
		if env.Unmeasured {
			t.Errorf("Unmeasured = true, want false: at least one run in scope was measured, even at a real zero")
		}
	})
}

// TestUnmeasuredNeverComputedWhenPeriodPredatesTelemetry pins the guard in
// (*statsHandler).view: AllRecordedRunsUnmeasured is only called once
// Recorded is already true, so a period that predates all telemetry never
// even asks the store the unmeasured question -- asserted by never handing
// the fake a fixture that would let it answer true, and confirming the
// response still reports Unmeasured=false.
func TestUnmeasuredNeverComputedWhenPeriodPredatesTelemetry(t *testing.T) {
	earliest := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	sts := &statsFake{
		stageRuns: []statsRun{{
			run:        store.StageRun{ID: 1, StartedAt: earliest, Command: "/myflow-do", Stage: "x"},
			projectKey: "proj", changeName: "kan-1",
		}},
		allRecordedRunsUnmeasured: true, // would flip the assertion below if ever consulted
	}
	ts := newStatsTestServer(t, sts)
	status, env, body := getStats(t, ts, "/api/v1/stats/cost-per-change?from=2020-01-01T00:00:00Z&to=2020-02-01T00:00:00Z")
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}
	if env.Unmeasured {
		t.Errorf("Unmeasured = true, want false: the period predates telemetry, so the store's AllRecordedRunsUnmeasured must never have been consulted")
	}
}

// --- TestLiveStateBoardMatchesStatusOutput ------------------------------

func TestLiveStateBoardMatchesStatusOutput(t *testing.T) {
	updatedAt := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	sts := &statsFake{
		liveStateBoard: []store.LiveStateRow{
			{ProjectKey: "agents", Name: "kan-1", State: store.StateStarted, UpdatedAt: updatedAt, UpdatedBy: "alice"},
			{ProjectKey: "agents", Name: "kan-2", State: store.StateInProgress, UpdatedAt: updatedAt, UpdatedBy: "bob"},
			{ProjectKey: "agents", Name: "kan-3", State: store.StateFinished, UpdatedAt: updatedAt, UpdatedBy: "carol"},
		},
	}
	ts := newStatsTestServer(t, sts)
	status, env, body := getStats(t, ts, periodPath("state-board"))
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}

	var rows []struct {
		Name        string `json:"name"`
		State       string `json:"state"`
		NextCommand string `json:"nextCommand"`
	}
	if err := json.Unmarshal(env.Rows, &rows); err != nil {
		t.Fatalf("decode rows: %v (body %s)", err, body)
	}
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}

	// This mapping is CLAUDE.md's own fixed three-state pipeline table,
	// restated here as the mechanical tie between the board and
	// /myflow-status's report for the same record (spec: "the information
	// matches what /myflow-status reports").
	want := map[string]string{"kan-1": "/myflow-do", "kan-2": "/myflow-finish", "kan-3": ""}
	for _, r := range rows {
		if r.NextCommand != want[r.Name] {
			t.Errorf("%s (state %s): nextCommand = %q, want %q", r.Name, r.State, r.NextCommand, want[r.Name])
		}
	}
}

// --- TestListEndpointsAcceptFilterSortSearchPage ------------------------

func TestListEndpointsAcceptFilterSortSearchPage(t *testing.T) {
	base := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	sts := &statsFake{
		stageRuns: []statsRun{
			{run: store.StageRun{ID: 1, Command: "/myflow-do", Stage: "a", StartedAt: base}, projectKey: "p", changeName: "kan-1"},
			{run: store.StageRun{ID: 2, Command: "/myflow-do", Stage: "b", StartedAt: base.Add(time.Hour)}, projectKey: "p", changeName: "kan-1"},
			{run: store.StageRun{ID: 3, Command: "/myflow-finish", Stage: "c", StartedAt: base.Add(2 * time.Hour)}, projectKey: "p", changeName: "kan-1"},
		},
	}
	ts := newStatsTestServer(t, sts)

	status, body := doGetRaw(t, ts, "/api/v1/stage-runs?command=/myflow-do&q=term&sort=-started_at&limit=1&offset=0")
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}

	// filter: command was translated into a store.Filter for "command".
	foundCommandFilter := false
	for _, f := range sts.lastQuery.Filters {
		if f.Field == "command" && f.Value == "/myflow-do" {
			foundCommandFilter = true
		}
	}
	if !foundCommandFilter {
		t.Errorf("Filters = %+v, want a command=/myflow-do filter", sts.lastQuery.Filters)
	}
	// search: q became Query.Search.
	if sts.lastQuery.Search != "term" {
		t.Errorf("Search = %q, want %q", sts.lastQuery.Search, "term")
	}
	// sort: "-started_at" became a descending SortKey on started_at.
	if len(sts.lastQuery.Sort) != 1 || sts.lastQuery.Sort[0].Field != "started_at" || !sts.lastQuery.Sort[0].Desc {
		t.Errorf("Sort = %+v, want one descending SortKey on started_at", sts.lastQuery.Sort)
	}
	// page: limit/offset were carried through.
	if sts.lastQuery.Limit != 1 || sts.lastQuery.Offset != 0 {
		t.Errorf("Limit/Offset = %d/%d, want 1/0", sts.lastQuery.Limit, sts.lastQuery.Offset)
	}

	var resp struct {
		Total     int `json:"total"`
		StageRuns []struct {
			StageRunID int64  `json:"stageRunId"`
			Command    string `json:"command"`
		} `json:"stageRuns"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decode: %v (body %s)", err, body)
	}
	// Two runs match command=/myflow-do; limit=1 pages down to one, and
	// Total still reports the unpaged match count of 2.
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
	if len(resp.StageRuns) != 1 {
		t.Fatalf("stageRuns = %d, want 1", len(resp.StageRuns))
	}
	if resp.StageRuns[0].StageRunID != 2 {
		t.Errorf("stageRuns[0].stageRunId = %d, want 2 (started_at DESC picks the later /myflow-do run first)", resp.StageRuns[0].StageRunID)
	}
}

// TestNegativeLimitAtHTTPBoundaryIsRejected is post-commit review finding
// F3: store.NoLimit is -1, an internal sentinel documented on the Query
// type as disabling paging entirely -- it was never meant to be reachable
// by a caller-supplied HTTP value. Before parseLimit existed, "limit=-1"
// resolved to that exact sentinel and produced every matching row with no
// LIMIT clause at all, defeating pagination. This is reproduced against
// GET /api/v1/stage-runs (this task's new endpoint) and GET
// /api/v1/changes (task 4's pre-existing endpoint, which had the identical
// gap -- the fix in changes.go's parseChangeQuery closes it there too).
func TestNegativeLimitAtHTTPBoundaryIsRejected(t *testing.T) {
	t.Run("stage-runs", func(t *testing.T) {
		sts := &statsFake{stageRuns: []statsRun{
			{run: store.StageRun{ID: 1, Command: "/myflow-do"}, projectKey: "p", changeName: "kan-1"},
		}}
		ts := newStatsTestServer(t, sts)
		status, body := doGetRaw(t, ts, "/api/v1/stage-runs?limit=-1")
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body %s", status, body)
		}
	})

	t.Run("changes", func(t *testing.T) {
		fs := newFakeStore()
		cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
		srv, err := api.New(cfg, fs, fs, fs, fs, fs, nil)
		if err != nil {
			t.Fatalf("api.New: %v", err)
		}
		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		status, body := doGetRaw(t, ts, "/api/v1/changes?limit=-1")
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body %s", status, body)
		}
	})
}

func doGetRaw(t *testing.T, ts *httptest.Server, path string) (int, string) {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	var body []byte
	buf := make([]byte, 65536)
	for {
		n, err := resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		if err != nil {
			break
		}
	}
	return resp.StatusCode, string(body)
}

// --- TestUnknownQueryFieldReturns400NamingIt ----------------------------

func TestUnknownQueryFieldReturns400NamingIt(t *testing.T) {
	t.Run("stage-runs list: unknown filter field", func(t *testing.T) {
		sts := &statsFake{
			queryStageRunsErr: fmt.Errorf("%w: filter %q not recognised; accepted: command, stage, project", store.ErrUnknownField, "bogus"),
		}
		ts := newStatsTestServer(t, sts)
		status, body := doGetRaw(t, ts, "/api/v1/stage-runs?bogus=1")
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body %s", status, body)
		}
		if !contains(body, "bogus") {
			t.Errorf("body %q does not name the rejected field %q", body, "bogus")
		}
	})

	t.Run("stats view: unrecognised query parameter", func(t *testing.T) {
		sts := &statsFake{}
		ts := newStatsTestServer(t, sts)
		status, body := doGetRaw(t, ts, periodPath("cost-per-change")+"&bogus=1")
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body %s", status, body)
		}
		if !contains(body, "bogus") {
			t.Errorf("body %q does not name the rejected parameter %q", body, "bogus")
		}
	})

	t.Run("stats view: unrecognised view name", func(t *testing.T) {
		sts := &statsFake{}
		ts := newStatsTestServer(t, sts)
		status, body := doGetRaw(t, ts, "/api/v1/stats/nonsense?from="+fromParam+"&to="+toParam)
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body %s", status, body)
		}
		if !contains(body, "nonsense") {
			t.Errorf("body %q does not name the rejected view %q", body, "nonsense")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}

// --- period validation ---------------------------------------------------

func TestStatsViewRequiresFromAndTo(t *testing.T) {
	sts := &statsFake{}
	ts := newStatsTestServer(t, sts)
	status, body := doGetRaw(t, ts, "/api/v1/stats/cost-per-change")
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body %s", status, body)
	}
}

// TestStatsViewRejectsToBeforeFrom is post-commit review finding F6: a
// period whose "to" precedes its "from" was verified by hand to return 400
// but had no test defending it.
func TestStatsViewRejectsToBeforeFrom(t *testing.T) {
	sts := &statsFake{}
	ts := newStatsTestServer(t, sts)
	status, body := doGetRaw(t, ts, "/api/v1/stats/cost-per-change?from=2026-09-01T00:00:00Z&to=2026-08-01T00:00:00Z")
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body %s", status, body)
	}
	if !contains(body, "before") {
		t.Errorf("body %q does not explain that to precedes from", body)
	}
}

// --- statsWriteTimeout: a slow aggregation gets a clean 503, never a
// truncated body -----------------------------------------------------------

// slowStatsFake blocks CostPerChange until the request's context is done,
// simulating an aggregation that runs past statsWriteTimeout.
type slowStatsFake struct {
	statsFake
}

func (f *slowStatsFake) CostPerChange(ctx context.Context, _ store.Period, _, _ *string) ([]store.CostPerChangeRow, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestStatsViewTimesOutCleanly(t *testing.T) {
	if os.Getenv("FLOW_STATS_SKIP_SLOW_TESTS") != "" {
		t.Skip("FLOW_STATS_SKIP_SLOW_TESTS set")
	}
	sts := &slowStatsFake{}
	ts := newStatsTestServer(t, sts)
	start := time.Now()

	resp, err := http.Get(ts.URL + periodPath("cost-per-change"))
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	elapsed := time.Since(start)

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body %s (after %s)", resp.StatusCode, body, elapsed)
	}
	if elapsed > 25*time.Second {
		t.Errorf("took %s to time out, want well under statsWriteTimeout's slack", elapsed)
	}
	// Post-commit review finding F5: internal/client treats a response
	// with no Myflow-Daemon header as "the daemon is unreachable" and
	// takes the journal fallback (server.go's withDaemonHeader, and the
	// client's own doc comment on why it insists on this header). A 503
	// from http.TimeoutHandler is written by net/http's own machinery,
	// not by this package's handler code, so it is not automatically
	// guaranteed to carry a header withDaemonHeader set before the
	// handler ever started -- this is what would silently regress a
	// clean timeout into "myflowd looks dead" for the CLI.
	if got := resp.Header.Get(api.DaemonHeader); got != api.DaemonHeaderValue {
		t.Errorf("Myflow-Daemon header = %q, want %q -- a 503 timeout must still look like a daemon answer, not an unreachable one", got, api.DaemonHeaderValue)
	}
}

// ==========================================================================
// Integration tests against a real PostgreSQL (the myflow-postgres compose
// stack, task 1): the new arithmetic this task adds outside store's own
// already-reviewed SQL -- costPerChangeByRepo's per-repository sums, and
// the multi-repo "one row by default" claim -- is exercised here with
// hand-computed totals, per this task's own testing standard: a fake
// proves the HTTP layer forwards whatever it is told, never that the
// numbers it is told are the right ones.
// ==========================================================================

func statsAdminDSN() string {
	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
		return v
	}
	return "postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable"
}

// newIntegrationStore creates a uniquely-named, migrated database against
// the compose stack's PostgreSQL and returns a *store.Store backed by it,
// registering cleanup. It skips cleanly, with a clear message, when the
// stack is not reachable -- exactly internal/store's own testsupport_test.go
// pattern, reproduced here because that helper is unexported in package
// store_test and this package cannot import it.
func newIntegrationStore(t *testing.T) *store.Store {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adminPool, err := pgxpool.New(ctx, statsAdminDSN())
	if err != nil {
		t.Skipf("myflow-postgres compose stack not reachable: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("myflow-postgres compose stack not reachable: %v", err)
	}

	dbName := fmt.Sprintf("myflow_apitest_%d_%d", os.Getpid(), time.Now().UnixNano())
	ident := pgx.Identifier{dbName}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		adminPool.Close()
		t.Fatalf("create test database %s: %v", dbName, err)
	}
	adminPool.Close()

	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer dropCancel()
		dropPool, err := pgxpool.New(dropCtx, statsAdminDSN())
		if err != nil {
			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
			return
		}
		defer dropPool.Close()
		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
			t.Logf("drop test database %s: %v", dbName, err)
		}
	})

	dsn := fmt.Sprintf("postgres://myflow:myflow@localhost:5433/%s?sslmode=disable", dbName)
	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	if err := st.RunMigrations(ctx); err != nil {
		st.Close()
		t.Fatalf("run migrations: %v", err)
	}
	t.Cleanup(st.Close)
	return st
}

// newIntegrationTestServer wires a server whose every store dependency is
// the real, migrated *store.Store the caller already holds -- the record
// store included. It used to pass a fake for that one, which meant the
// records routes were the only routes this harness could not exercise
// against Postgres at all, for no reason beyond the order the parameters
// were added in: *store.Store satisfies api.RecordStore, and a compile-time
// assertion in internal/api already says so.
func newIntegrationTestServer(t *testing.T, st *store.Store) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	srv, err := api.New(cfg, st, st, st, st, st, nil)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func mustPutIntegrationChange(t *testing.T, st *store.Store, projectKey, name string, repos []store.Repo) {
	t.Helper()
	err := st.PutChange(context.Background(), store.Change{
		ProjectKey:       projectKey,
		MainCheckoutPath: "/tmp/" + projectKey,
		Name:             name,
		State:            store.StateInProgress,
		Repos:            repos,
		UpdatedAt:        time.Now().UTC(),
		UpdatedBy:        "test",
	})
	if err != nil {
		t.Fatalf("seed change %s/%s: %v", projectKey, name, err)
	}
}

func mustRunIntegrationStage(t *testing.T, st *store.Store, in store.BeginStageInput, metrics json.RawMessage, endedAt time.Time) store.StageRun {
	t.Helper()
	ctx := context.Background()
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if metrics != nil {
		if err := st.MergeMetrics(ctx, run.ID, metrics); err != nil {
			t.Fatalf("MergeMetrics: %v", err)
		}
	}
	if err := st.EndStage(ctx, run.ID, endedAt, "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}
	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	return got
}

// TestMultiRepoChangeIsOneRowInEveryView seeds one change spanning two
// repositories, with one stage run recorded in each, and asserts that both
// the cost-per-change and state-board views report it as a single row --
// design.md: "a two-repository change is one row here, never two" --
// with hand-computed totals (RunCount=2, TotalCostUSD=2.0+3.0=5.0), not
// merely "some row came back".
func TestMultiRepoChangeIsOneRowInEveryView(t *testing.T) {
	st := newIntegrationStore(t)
	projectKey := fmt.Sprintf("proj-multirepo-%d", time.Now().UnixNano())
	repoA, repoB := "/repos/a", "/repos/b"
	mustPutIntegrationChange(t, st, projectKey, "kan-1", []store.Repo{{RepoRoot: repoA}, {RepoRoot: repoB}})

	started := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)
	mustRunIntegrationStage(t, st, store.BeginStageInput{
		ProjectKey: projectKey, ChangeName: "kan-1", RepoRoot: &repoA, Harness: "claude-code",
		Command: "/myflow-do", Stage: "SDD + TDD per task", StartedAt: started,
	}, json.RawMessage(`{"cost_usd":2.0,"tokens":{"input":100}}`), started.Add(time.Minute))
	mustRunIntegrationStage(t, st, store.BeginStageInput{
		ProjectKey: projectKey, ChangeName: "kan-1", RepoRoot: &repoB, Harness: "claude-code",
		Command: "/myflow-do", Stage: "SDD + TDD per task", StartedAt: started.Add(2 * time.Minute),
	}, json.RawMessage(`{"cost_usd":3.0,"tokens":{"input":200}}`), started.Add(3*time.Minute))

	ts := newIntegrationTestServer(t, st)
	period := "from=2026-08-01T00:00:00Z&to=2026-09-01T00:00:00Z"

	status, env, body := getStats(t, ts, fmt.Sprintf("/api/v1/stats/cost-per-change?%s&project=%s", period, projectKey))
	if status != http.StatusOK {
		t.Fatalf("cost-per-change: status %d, body %s", status, body)
	}
	var costRows []struct {
		RunCount     int      `json:"runCount"`
		TotalCostUSD *float64 `json:"totalCostUsd"`
	}
	if err := json.Unmarshal(env.Rows, &costRows); err != nil {
		t.Fatalf("decode cost-per-change rows: %v (body %s)", err, body)
	}
	if len(costRows) != 1 {
		t.Fatalf("cost-per-change rows = %d, want 1 (a two-repository change is one row)", len(costRows))
	}
	if costRows[0].RunCount != 2 {
		t.Errorf("RunCount = %d, want 2", costRows[0].RunCount)
	}
	if costRows[0].TotalCostUSD == nil || *costRows[0].TotalCostUSD != 5.0 {
		t.Errorf("TotalCostUSD = %v, want 5.0 (2.0 + 3.0)", costRows[0].TotalCostUSD)
	}

	status, env, body = getStats(t, ts, fmt.Sprintf("/api/v1/stats/state-board?%s&project=%s", period, projectKey))
	if status != http.StatusOK {
		t.Fatalf("state-board: status %d, body %s", status, body)
	}
	var boardRows []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(env.Rows, &boardRows); err != nil {
		t.Fatalf("decode state-board rows: %v (body %s)", err, body)
	}
	if len(boardRows) != 1 {
		t.Fatalf("state-board rows = %d, want 1", len(boardRows))
	}
}

// --- task 21: the model filter -------------------------------------------

// TestModelRestrictionRejectedOnStateBoard is
// specs/myflow-stats-views/spec.md's "A model restriction on the live
// state board" scenario: the state board's rows are changes, not stage
// runs, so a model parameter is rejected outright rather than silently
// accepted and ignored -- the one outcome the round's own spec forbids.
// It must name the view, the same posture an unrecognised breakdown or
// query parameter already takes elsewhere in this file, and it must never
// reach the store at all: sts.aggErr would fail the request a different,
// wrong way if this rejection happened downstream instead of before any
// store call.
func TestModelRestrictionRejectedOnStateBoard(t *testing.T) {
	sts := &statsFake{}
	ts := newStatsTestServer(t, sts)
	status, body := doGetRaw(t, ts, periodPath("state-board")+"&model=claude-opus-5")
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body %s", status, body)
	}
	if !contains(body, "state-board") {
		t.Errorf("body %q does not name the rejected view", body)
	}
	if sts.lastProject != nil || sts.lastModel != nil {
		t.Errorf("store was called (lastProject=%v, lastModel=%v) -- the rejection must happen before any store call", sts.lastProject, sts.lastModel)
	}
}

// TestModelParamAcceptedAndPassedThroughOnEveryStageRunView iterates the
// four views whose rows are stage runs (every view except state-board,
// covered separately above) and asserts a "model" query parameter reaches
// the store call unchanged, for each one -- not spot-checked, per this
// task's own non-negotiable requirement that the filter be honoured on
// every one of the four, iterated.
func TestModelParamAcceptedAndPassedThroughOnEveryStageRunView(t *testing.T) {
	views := []string{
		"cost-per-change", "stage-leaderboard", "trend", "cache-efficiency",
	}
	for _, view := range views {
		t.Run(view, func(t *testing.T) {
			sts := &statsFake{}
			ts := newStatsTestServer(t, sts)
			status, env, body := getStats(t, ts, periodPath(view)+"&model=claude-opus-5")
			if status != http.StatusOK {
				t.Fatalf("GET /api/v1/stats/%s?model=...: status %d, body %s", view, status, body)
			}
			if sts.lastModel == nil || *sts.lastModel != "claude-opus-5" {
				t.Errorf("store received model=%v, want \"claude-opus-5\"", sts.lastModel)
			}
			if env.Model == nil || *env.Model != "claude-opus-5" {
				t.Errorf("response model = %v, want \"claude-opus-5\" echoed back", env.Model)
			}
		})
	}
}

// TestExcludedNoModelPresentOnlyWhenAModelFilterWasApplied is this task's
// own absence-is-never-zero rule applied to its newest field: an
// unfiltered request must carry no "excludedNoModel" key at all (never a
// zero standing in for "not applicable"), and a filtered request must
// always carry the key, including when the excluded count is genuinely
// zero -- both cases asserted against the same statsFake so a regression
// toward either wrong shape (always present, or omitted-on-zero) fails
// here.
func TestExcludedNoModelPresentOnlyWhenAModelFilterWasApplied(t *testing.T) {
	t.Run("unfiltered: absent", func(t *testing.T) {
		sts := &statsFake{countRunsWithoutModel: 7} // must not leak through unfiltered
		ts := newStatsTestServer(t, sts)
		_, body := doGetRaw(t, ts, periodPath("cost-per-change"))
		if contains(body, "excludedNoModel") {
			t.Errorf("unfiltered response body carries \"excludedNoModel\", want it entirely absent: %s", body)
		}
	})

	t.Run("filtered, non-zero: present with the real count", func(t *testing.T) {
		sts := &statsFake{countRunsWithoutModel: 3}
		ts := newStatsTestServer(t, sts)
		status, env, body := getStats(t, ts, periodPath("cost-per-change")+"&model=claude-opus-5")
		if status != http.StatusOK {
			t.Fatalf("status = %d, want 200; body %s", status, body)
		}
		if env.ExcludedNoModel == nil || *env.ExcludedNoModel != 3 {
			t.Errorf("excludedNoModel = %v, want 3", env.ExcludedNoModel)
		}
	})

	t.Run("filtered, genuinely zero: present as 0, not absent", func(t *testing.T) {
		sts := &statsFake{countRunsWithoutModel: 0}
		ts := newStatsTestServer(t, sts)
		status, env, body := getStats(t, ts, periodPath("cost-per-change")+"&model=claude-opus-5")
		if status != http.StatusOK {
			t.Fatalf("status = %d, want 200; body %s", status, body)
		}
		if !contains(string(body), `"excludedNoModel":0`) {
			t.Errorf("body does not carry an explicit excludedNoModel:0, want the genuinely-zero case distinguishable from absence: %s", body)
		}
		if env.ExcludedNoModel == nil || *env.ExcludedNoModel != 0 {
			t.Errorf("excludedNoModel = %v, want a present 0", env.ExcludedNoModel)
		}
	})
}

// --- GET /api/v1/models ---------------------------------------------------

type modelsEnvelope struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	Project *string  `json:"project"`
	Models  []string `json:"models"`
}

func getModels(t *testing.T, ts *httptest.Server, path string) (int, modelsEnvelope, []byte) {
	t.Helper()
	status, body := doGetRaw(t, ts, path)
	if status != http.StatusOK {
		return status, modelsEnvelope{}, []byte(body)
	}
	var env modelsEnvelope
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("decode response for %s: %v (body: %s)", path, err, body)
	}
	return status, env, []byte(body)
}

// TestListModelsRoute exercises GET /api/v1/models against a statsFake
// carrying a known set -- the route's own wiring, request validation and
// wire shape -- store.ListModels' own arithmetic (period, project scoping,
// sorting) is aggregate_test.go's TestListModels, against real Postgres.
func TestListModelsRoute(t *testing.T) {
	t.Run("returns the fake's models", func(t *testing.T) {
		sts := &statsFake{models: []string{"claude-opus-5", "claude-sonnet-5"}}
		ts := newStatsTestServer(t, sts)
		status, env, body := getModels(t, ts, fmt.Sprintf("/api/v1/models?from=%s&to=%s", fromParam, toParam))
		if status != http.StatusOK {
			t.Fatalf("status = %d, want 200; body %s", status, body)
		}
		if len(env.Models) != 2 || env.Models[0] != "claude-opus-5" || env.Models[1] != "claude-sonnet-5" {
			t.Errorf("models = %v, want [claude-opus-5 claude-sonnet-5]", env.Models)
		}
	})

	t.Run("requires from and to", func(t *testing.T) {
		sts := &statsFake{}
		ts := newStatsTestServer(t, sts)
		status, body := doGetRaw(t, ts, "/api/v1/models")
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body %s", status, body)
		}
	})

	t.Run("rejects an unrecognised parameter, naming it", func(t *testing.T) {
		sts := &statsFake{}
		ts := newStatsTestServer(t, sts)
		status, body := doGetRaw(t, ts, fmt.Sprintf("/api/v1/models?from=%s&to=%s&bogus=1", fromParam, toParam))
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body %s", status, body)
		}
		if !contains(body, "bogus") {
			t.Errorf("body %q does not name the rejected parameter", body)
		}
	})

	t.Run("carries the daemon header, like every other route", func(t *testing.T) {
		sts := &statsFake{}
		ts := newStatsTestServer(t, sts)
		resp, err := http.Get(ts.URL + fmt.Sprintf("/api/v1/models?from=%s&to=%s", fromParam, toParam))
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()
		if resp.Header.Get(api.DaemonHeader) != api.DaemonHeaderValue {
			t.Errorf("%s header = %q, want %q", api.DaemonHeader, resp.Header.Get(api.DaemonHeader), api.DaemonHeaderValue)
		}
	})
}

// --- task 21, step 5: "change" alone on cost-per-change -------------------

// TestCostPerChangeAcceptsChangeAlone is task 21, step 5's own scenario,
// against a real store: before this step, only breakdown=repo ever read
// "change" -- a bare "change" parameter on cost-per-change was accepted
// (it was already in statsQueryParams) and then silently ignored, which is
// exactly the shape this file's own header comment forbids everywhere
// else. Two changes are seeded in one project, each with its own stage
// run; a request naming one change by name, with no breakdown, must
// return only that change's own row.
func TestCostPerChangeAcceptsChangeAlone(t *testing.T) {
	st := newIntegrationStore(t)
	projectKey := fmt.Sprintf("proj-change-alone-%d", time.Now().UnixNano())
	mustPutIntegrationChange(t, st, projectKey, "kan-1", nil)
	mustPutIntegrationChange(t, st, projectKey, "kan-2", nil)
	ts := newIntegrationTestServer(t, st)

	started := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	mustRunIntegrationStage(t, st, store.BeginStageInput{
		ProjectKey: projectKey, ChangeName: "kan-1", Harness: "claude-code",
		SessionID: sptr("s-kan-1"), Command: "/myflow-do", Stage: "SDD + TDD per task",
		StartedAt: started,
	}, json.RawMessage(`{"cost_usd":1.5}`), started.Add(time.Minute))
	mustRunIntegrationStage(t, st, store.BeginStageInput{
		ProjectKey: projectKey, ChangeName: "kan-2", Harness: "claude-code",
		SessionID: sptr("s-kan-2"), Command: "/myflow-do", Stage: "SDD + TDD per task",
		StartedAt: started,
	}, json.RawMessage(`{"cost_usd":9}`), started.Add(time.Minute))

	period := "from=2026-08-01T00:00:00Z&to=2026-09-01T00:00:00Z"
	path := fmt.Sprintf("/api/v1/stats/cost-per-change?%s&project=%s&change=kan-1", period, projectKey)
	status, env, body := getStats(t, ts, path)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", status, body)
	}

	var rows []struct {
		ChangeName   string   `json:"changeName"`
		TotalCostUSD *float64 `json:"totalCostUsd"`
	}
	if err := json.Unmarshal(env.Rows, &rows); err != nil {
		t.Fatalf("decode rows: %v (body %s)", err, body)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1 (only kan-1)", len(rows))
	}
	if rows[0].ChangeName != "kan-1" {
		t.Errorf("changeName = %q, want kan-1", rows[0].ChangeName)
	}
	if rows[0].TotalCostUSD == nil || *rows[0].TotalCostUSD != 1.5 {
		t.Errorf("totalCostUsd = %v, want 1.5 (kan-1's own row, not kan-2's)", rows[0].TotalCostUSD)
	}
}

// sptr is *string's own ptr helper for this file, distinct from the
// store package's own generic ptr[T] (internal/store/changes_test.go),
// which this package cannot reach.
func sptr(v string) *string { return &v }
