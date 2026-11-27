package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

func TestGuardHealthCarriesItsRowsThrough(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
