package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

func TestHealthViewsCarryTheirNumbersThrough(t *testing.T) {
	lastRun := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	lastFired := time.Date(2026, 8, 9, 9, 0, 0, 0, time.UTC)
	median := 812.5
	medSec, p90Sec := 61.0, 340.0
	rounds := 3
	sts := &statsFake{
		guardActivity: []store.GuardActivityRow{
			{Guard: "check-plan-shape", Runs: 40, Fired: 7, CannotAnswer: 2,
				LastRunAt: &lastRun, LastFiredAt: &lastFired, MedianDurationMS: &median,
				Verdicts: 5, FalsePositives: 1},
			{Guard: "check-unfinished-work", Verdicts: 3},
		},
		stageRedo: []store.StageRedoRow{
			{Command: "/flow", Stage: "flow.review-panel", Runs: 9, Changes: 5,
				Reentries: 4, ReenteredChanges: 3, MedianSeconds: &medSec, P90Seconds: &p90Sec},
		},
		panelRounds: []store.PanelRoundsRow{
			{Project: "agents", Change: "kan-1", StartedAt: lastRun, Rounds: &rounds,
				Findings: 6, Critical: 1, Important: 2, Minor: 3},
			{Project: "agents", Change: "kan-2", StartedAt: lastFired},
		},
	}
	ts := newStatsTestServer(t, sts)

	t.Run("guards", func(t *testing.T) {
		status, env, body := getStats(t, ts, periodPath("guards"))
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		var rows []map[string]any
		mustDecodeRows(t, env, body, &rows)
		if len(rows) != 2 {
			t.Fatalf("rows = %v, want 2", rows)
		}
		want := map[string]any{
			"guard": "check-plan-shape", "runs": 40.0, "fired": 7.0, "cannotAnswer": 2.0,
			"lastRunAt": "2026-08-10T09:00:00Z", "lastFiredAt": "2026-08-09T09:00:00Z",
			"medianDurationMs": 812.5, "verdicts": 5.0, "falsePositives": 1.0,
		}
		for k, v := range want {
			if rows[0][k] != v {
				t.Errorf("rows[0][%q] = %v, want %v", k, rows[0][k], v)
			}
		}
		// A guard known only from a verdict has no runs: its run-derived
		// fields are explicit nulls, never a zero time or a zero duration.
		for _, k := range []string{"lastRunAt", "lastFiredAt", "medianDurationMs"} {
			v, ok := rows[1][k]
			if !ok || v != nil {
				t.Errorf("rows[1][%q] = %v (present %v), want an explicit null", k, v, ok)
			}
		}
	})

	t.Run("stage-redo", func(t *testing.T) {
		status, env, body := getStats(t, ts, periodPath("stage-redo"))
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		var rows []map[string]any
		mustDecodeRows(t, env, body, &rows)
		want := map[string]any{
			"command": "/flow", "stage": "flow.review-panel", "runs": 9.0, "changes": 5.0,
			"reentries": 4.0, "reenteredChanges": 3.0, "medianSeconds": 61.0, "p90Seconds": 340.0,
		}
		if len(rows) != 1 {
			t.Fatalf("rows = %v, want 1", rows)
		}
		for k, v := range want {
			if rows[0][k] != v {
				t.Errorf("rows[0][%q] = %v, want %v", k, rows[0][k], v)
			}
		}
	})

	t.Run("panel-rounds", func(t *testing.T) {
		status, env, body := getStats(t, ts, periodPath("panel-rounds"))
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		var rows []map[string]any
		mustDecodeRows(t, env, body, &rows)
		if len(rows) != 2 {
			t.Fatalf("rows = %v, want 2", rows)
		}
		want := map[string]any{
			"project": "agents", "change": "kan-1", "startedAt": "2026-08-10T09:00:00Z",
			"rounds": 3.0, "findings": 6.0, "critical": 1.0, "important": 2.0, "minor": 3.0,
		}
		for k, v := range want {
			if rows[0][k] != v {
				t.Errorf("rows[0][%q] = %v, want %v", k, rows[0][k], v)
			}
		}
		if v, ok := rows[1]["rounds"]; !ok || v != nil {
			t.Errorf("rows[1].rounds = %v (present %v), want an explicit null for an unrecorded panel", v, ok)
		}
	})
}

// TestHealthViewsAreNeverUnmeasured: "no stage run carried a token
// measurement" says nothing about guard runs, attempts or panel rounds, and
// the SPA hides a panel's rows when it is set.
func TestHealthViewsAreNeverUnmeasured(t *testing.T) {
	for _, view := range []string{"guards", "stage-redo", "panel-rounds"} {
		t.Run(view, func(t *testing.T) {
			ts := newStatsTestServer(t, &statsFake{
				allRecordedRunsUnmeasured: true,
				stageRuns:                 []statsRun{{run: store.StageRun{StartedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)}}},
			})
			status, env, body := getStats(t, ts, periodPath(view))
			if status != http.StatusOK {
				t.Fatalf("status %d, body %s", status, body)
			}
			if !env.Recorded || env.Unmeasured {
				t.Fatalf("recorded=%v unmeasured=%v, want recorded and never unmeasured", env.Recorded, env.Unmeasured)
			}
		})
	}
}

func TestHealthViewsRefuseAModelRestriction(t *testing.T) {
	for _, view := range []string{"guards", "stage-redo", "panel-rounds"} {
		t.Run(view, func(t *testing.T) {
			ts := newStatsTestServer(t, &statsFake{})
			status, _, body := getStats(t, ts, periodPath(view)+"&model=opus")
			if status != http.StatusBadRequest {
				t.Fatalf("status %d, want 400; body %s", status, body)
			}
		})
	}
}

// guardRunFake is the GuardRunStore the route's tests wire through
// api.WithGuardRuns.
type guardRunFake struct {
	project string
	got     records.GuardRun
	err     error
}

func (f *guardRunFake) RecordGuardRun(_ context.Context, projectKey string, in records.GuardRun) (records.GuardRun, error) {
	f.project, f.got = projectKey, in
	if f.err != nil {
		return records.GuardRun{}, f.err
	}
	in.ID = 7
	return in, nil
}

func newGuardRunServer(t *testing.T, f api.GuardRunStore) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	var opts []api.Option
	if f != nil {
		opts = append(opts, api.WithGuardRuns(f))
	}
	srv, err := api.New(cfg, newFakeStore(), newFakeStore(), newFakeStore(), newFakeStore(), newFakeStore(), nil, opts...)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func postGuardRun(t *testing.T, ts *httptest.Server, project string, in records.GuardRun) (int, []byte) {
	t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(ts.URL+"/api/v1/guard-runs/"+project, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body
}

func TestRecordGuardRunRoute(t *testing.T) {
	run := records.GuardRun{Guard: "check-plan-shape", Worktree: "/w/kan-1", ExitCode: 1,
		Outcome: "fired", DurationMS: 120, RecordedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)}

	t.Run("records the run against the project", func(t *testing.T) {
		f := &guardRunFake{}
		ts := newGuardRunServer(t, f)
		status, body := postGuardRun(t, ts, "agents-1234abcd", run)
		if status != http.StatusCreated {
			t.Fatalf("status %d, want 201; body %s", status, body)
		}
		if f.project != "agents-1234abcd" || f.got.Guard != run.Guard || f.got.Outcome != "fired" || f.got.DurationMS != 120 {
			t.Errorf("store got %s %+v", f.project, f.got)
		}
	})

	for _, tc := range []struct {
		err  error
		want int
	}{
		{fmt.Errorf("%w: x", store.ErrProjectNotFound), http.StatusNotFound},
		{fmt.Errorf("%w: x", store.ErrInvalidGuardRun), http.StatusBadRequest},
	} {
		t.Run(fmt.Sprintf("maps %v to %d", tc.err, tc.want), func(t *testing.T) {
			ts := newGuardRunServer(t, &guardRunFake{err: tc.err})
			if status, body := postGuardRun(t, ts, "p", run); status != tc.want {
				t.Fatalf("status %d, want %d; body %s", status, tc.want, body)
			}
		})
	}

	t.Run("absent without WithGuardRuns", func(t *testing.T) {
		ts := newGuardRunServer(t, nil)
		if status, body := postGuardRun(t, ts, "p", run); status != http.StatusNotFound {
			t.Fatalf("status %d, want 404; body %s", status, body)
		}
	})
}
