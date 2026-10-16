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

func TestFlowSummaryCarriesItsNumbersThrough(t *testing.T) {
	meanCost, rounds := 3.5, 2.0
	lead := int64(1_126_800_000)
	sts := &statsFake{flowSummary: store.FlowSummaryRow{
		ChangesStarted: 2, ChangesFinished: 3, TotalCostUsd: 6.5, MeanCostPerFinished: &meanCost,
		MedianLeadTimeMs: &lead, MeanPanelRounds: &rounds, Critical: 1, Important: 2, Minor: 4, GuardFires: 7,
	}}
	ts := newStatsTestServer(t, sts)
	status, env, body := getStats(t, ts, periodPath("flow-summary"))
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}
	var row map[string]any
	mustDecodeRows(t, env, body, &row)
	want := map[string]any{
		"changesStarted": 2.0, "changesFinished": 3.0, "totalCostUsd": 6.5, "meanCostPerFinished": 3.5,
		"medianLeadTimeMs": 1126800000.0, "meanPanelRounds": 2.0,
		"critical": 1.0, "important": 2.0, "minor": 4.0, "guardFires": 7.0,
	}
	for k, v := range want {
		if row[k] != v {
			t.Errorf("row[%q] = %v, want %v", k, row[k], v)
		}
	}

	// Nothing finished: the means are explicit nulls, never a zero.
	ts = newStatsTestServer(t, &statsFake{})
	_, env, body = getStats(t, ts, periodPath("flow-summary"))
	row = nil
	mustDecodeRows(t, env, body, &row)
	for _, k := range []string{"meanCostPerFinished", "medianLeadTimeMs", "meanPanelRounds"} {
		if v, ok := row[k]; !ok || v != nil {
			t.Errorf("empty row[%q] = %v (present %v), want an explicit null", k, v, ok)
		}
	}
}

// TestHealthViewsAreNeverUnmeasured: "no stage run carried a token
// measurement" says nothing about changes, guard runs or panel rounds, and
// the SPA hides a panel's rows when it is set.
func TestHealthViewsAreNeverUnmeasured(t *testing.T) {
	for _, view := range []string{"flow-summary", "self-review"} {
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
	for _, view := range []string{"flow-summary", "self-review"} {
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
