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

func TestGuardHealthCarriesItsRowsThrough(t *testing.T) {
	fired := time.Date(2026, 8, 3, 14, 5, 0, 0, time.FixedZone("x", 3600))
	sts := &statsFake{guardHealth: []store.GuardHealthRow{
		{Guard: "check-a", Runs: 9, Fired: 2, CannotAnswer: 1, MedianMs: 120, TotalMs: 4000, LastFiredAt: &fired},
		{Guard: "check-b", Runs: 3, MedianMs: 5, TotalMs: 15},
	}}
	ts := newStatsTestServer(t, sts)
	status, env, body := getStats(t, ts, periodPath("guard-health")+"&project=p")
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}
	if sts.lastProject == nil || *sts.lastProject != "p" {
		t.Errorf("project reaching the store = %v, want p", sts.lastProject)
	}
	var rows []map[string]any
	mustDecodeRows(t, env, body, &rows)
	if len(rows) != 2 {
		t.Fatalf("rows = %v, want 2", rows)
	}
	want := map[string]any{
		"guard": "check-a", "runs": 9.0, "fired": 2.0, "cannotAnswer": 1.0,
		"medianMs": 120.0, "totalMs": 4000.0, "lastFiredAt": "2026-08-03T13:05:00Z",
	}
	for k, v := range want {
		if rows[0][k] != v {
			t.Errorf("rows[0][%q] = %v, want %v", k, rows[0][k], v)
		}
	}
	if v, ok := rows[1]["lastFiredAt"]; !ok || v != nil {
		t.Errorf("rows[1].lastFiredAt = %v (present %v), want an explicit null", v, ok)
	}

	// No guard ran: an empty list, never null.
	ts = newStatsTestServer(t, &statsFake{})
	_, _, body = getStats(t, ts, periodPath("guard-health"))
	var raw struct {
		Rows json.RawMessage `json:"rows"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decode: %v (body %s)", err, body)
	}
	if string(raw.Rows) != "[]" {
		t.Errorf("rows = %s, want []", raw.Rows)
	}
}

// TestGuardHealthAnswersProjects: guard-health?projects=1 carries each
// project's guard-run count in the store's order; without it the count
// does not run (decision guard-health-projects-param).
func TestGuardHealthAnswersProjects(t *testing.T) {
	sts := &statsFake{guardHealthProjects: []store.ProjectGuardRuns{{ProjectKey: "z", Runs: 5}, {ProjectKey: "a", Runs: 2}}}
	ts := newStatsTestServer(t, sts)

	status, _, body := getStats(t, ts, periodPath("guard-health")+"&project=a&projects=1")
	if status != http.StatusOK {
		t.Fatalf("status %d, body %s", status, body)
	}
	var got struct {
		Projects json.RawMessage `json:"projects"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v (body %s)", err, body)
	}
	if want := `[{"projectKey":"z","runs":5},{"projectKey":"a","runs":2}]`; string(got.Projects) != want {
		t.Errorf("projects = %s, want %s", got.Projects, want)
	}

	sts.guardHealthProjectsCalled = false
	_, _, body = getStats(t, ts, periodPath("guard-health"))
	var absent map[string]json.RawMessage
	if err := json.Unmarshal(body, &absent); err != nil {
		t.Fatalf("decode: %v (body %s)", err, body)
	}
	if _, ok := absent["projects"]; ok {
		t.Errorf("projects present without projects=1: %s", body)
	}
	if sts.guardHealthProjectsCalled {
		t.Error("GuardHealthProjects ran without projects=1")
	}
}

// TestHealthViewsAreNeverUnmeasured: "no stage run carried a token
// measurement" says nothing about changes, guard runs or panel rounds, and
// the SPA hides a panel's rows when it is set.
func TestHealthViewsAreNeverUnmeasured(t *testing.T) {
	for _, view := range []string{"guard-health", "self-review"} {
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
	for _, view := range []string{"guard-health", "self-review"} {
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
