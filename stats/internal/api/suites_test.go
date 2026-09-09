package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// errFakeInvalidSuiteRun is the fake's stand-in for the store's own typed
// refusal, which the handler maps to 400.
var errFakeInvalidSuiteRun = store.ErrInvalidSuiteRun

// suiteRunRecord is fakeStore's in-memory stand-in for a suite_runs row --
// the owning project key kept beside the row, the same shape hazardRecord
// uses.
type suiteRunRecord struct {
	run        records.SuiteRun
	projectKey string
}

// InsertSuiteRun mirrors store.Store.InsertSuiteRun: rows are allocated ids
// and stamped when the caller left RanAt zero; an empty suite or host or a
// negative duration is store.ErrInvalidSuiteRun, the store's boundary the
// handler never lets a bad payload reach.
func (f *fakeStore) InsertSuiteRun(_ context.Context, projectKey string, run records.SuiteRun) (records.SuiteRun, error) {
	f.recordCalls++
	if f.insertSuiteRunErr != nil {
		return records.SuiteRun{}, f.insertSuiteRunErr
	}
	if run.Suite == "" || run.Host == "" || run.DurationMs < 0 {
		return records.SuiteRun{}, errFakeInvalidSuiteRun
	}
	f.nextSuiteRunID++
	out := run
	out.ID = f.nextSuiteRunID
	if out.RanAt.IsZero() {
		out.RanAt = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC).Add(time.Duration(f.nextSuiteRunID) * time.Second)
	}
	f.suiteRuns = append(f.suiteRuns, suiteRunRecord{run: out, projectKey: projectKey})
	return out, nil
}

// ListSuiteRuns mirrors store.Store.ListSuiteRuns: newest first, the empty
// suite unfiltered, limit capping the row count.
func (f *fakeStore) ListSuiteRuns(_ context.Context, projectKey, suite string, limit int) ([]records.SuiteRun, error) {
	if f.listSuiteRunsErr != nil {
		return nil, f.listSuiteRunsErr
	}
	if limit <= 0 {
		return nil, errFakeInvalidSuiteRun
	}
	var out []records.SuiteRun
	for i := len(f.suiteRuns) - 1; i >= 0 && len(out) < limit; i-- {
		sr := f.suiteRuns[i]
		if sr.projectKey != projectKey {
			continue
		}
		if suite != "" && sr.run.Suite != suite {
			continue
		}
		out = append(out, sr.run)
	}
	return out, nil
}

// TestPostSuiteRunReturnsTheRecordedRow pins the write route: POST 201 with
// the stored row echoed, daemon-allocated id and stamp included.
func TestPostSuiteRunReturnsTheRecordedRow(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	resp, body := postJSON(t, ts.URL+"/api/v1/suites/proj/runs", map[string]any{
		"suite": "guard-tests", "host": "laptop", "durationMs": 52400, "exitCode": 0,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST suite runs = %d (%s), want 201", resp.StatusCode, body)
	}
	var got records.SuiteRun
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if got.ID == 0 || got.Suite != "guard-tests" || got.RanAt.IsZero() {
		t.Errorf("response = %+v, want an allocated id, the suite and a stamp", got)
	}
}

// TestGetSuiteRunsReadsTheArray pins the read route's query contract: the
// suite filter and limit travel to the store, and absent parameters use the
// defaults (every suite, limit 20).
func TestGetSuiteRunsReadsTheArray(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	for _, in := range []map[string]any{
		{"suite": "guard-tests", "host": "laptop", "durationMs": 52400, "exitCode": 0},
		{"suite": "guard-tests", "host": "ci", "durationMs": 90000, "exitCode": 0},
		{"suite": "stats-go", "host": "laptop", "durationMs": 24000, "exitCode": 0},
	} {
		if resp, body := postJSON(t, ts.URL+"/api/v1/suites/proj/runs", in); resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed POST = %d (%s), want 201", resp.StatusCode, body)
		}
	}

	status, body := doGet(t, ts, "/api/v1/suites/proj/runs?suite=guard-tests&limit=2")
	if status != http.StatusOK {
		t.Fatalf("GET suite runs = %d (%s), want 200", status, body)
	}
	var listed []records.SuiteRun
	if err := json.Unmarshal([]byte(body), &listed); err != nil {
		t.Fatalf("decode listing %s: %v", body, err)
	}
	if len(listed) != 2 {
		t.Fatalf("suite=guard-tests&limit=2 listing = %d rows, want 2", len(listed))
	}
	if listed[0].Host != "ci" {
		t.Errorf("newest row host = %s, want ci (newest first)", listed[0].Host)
	}
	for _, run := range listed {
		if run.Suite != "guard-tests" {
			t.Errorf("filtered listing carried suite %s", run.Suite)
		}
	}
}

// TestPostSuiteRunRejectsAnEmptySuite pins the refusal path: a payload that
// cannot identify what it describes is a 400 before the store is touched,
// the same caller-mistake shape every record handler uses.
func TestPostSuiteRunRejectsAnEmptySuite(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")

	resp, body := postJSON(t, ts.URL+"/api/v1/suites/proj/runs", map[string]any{
		"suite": "", "host": "laptop", "durationMs": 52400, "exitCode": 0,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST with empty suite = %d (%s), want 400", resp.StatusCode, body)
	}

	resp, body = postJSON(t, ts.URL+"/api/v1/suites/proj/runs", map[string]any{
		"suite": "guard-tests", "host": "laptop", "durationMs": -1, "exitCode": 0,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST with negative duration = %d (%s), want 400", resp.StatusCode, body)
	}

	if fs.recordCalls != 0 {
		t.Errorf("store record calls = %d, want 0: a refused payload never reaches the store", fs.recordCalls)
	}
}

// TestGetSuiteRunsDefaultsAndRejectsBadLimit pins the limit handling the
// handler owns: an absent limit uses the default (20 -- a store refusing
// limit <= 0 makes a zeroed default loud), and a limit that is not a
// positive integer is a 400 before the store is touched.
func TestGetSuiteRunsDefaultsAndRejectsBadLimit(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")

	if resp, body := postJSON(t, ts.URL+"/api/v1/suites/proj/runs", map[string]any{
		"suite": "guard-tests", "host": "laptop", "durationMs": 52400, "exitCode": 0,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("seed POST = %d (%s), want 201", resp.StatusCode, body)
	}

	status, body := doGet(t, ts, "/api/v1/suites/proj/runs")
	if status != http.StatusOK {
		t.Fatalf("GET without limit = %d (%s), want 200 -- the default must reach the store", status, body)
	}
	var listed []records.SuiteRun
	if err := json.Unmarshal([]byte(body), &listed); err != nil {
		t.Fatalf("decode listing %s: %v", body, err)
	}
	if len(listed) != 1 {
		t.Errorf("default-limit listing = %d rows, want the seeded row", len(listed))
	}
	if fs.listSuiteRunsErr != nil {
		t.Errorf("the default-limit read errored: %v -- a zeroed default would surface here as the store's refusal", fs.listSuiteRunsErr)
	}

	status, body = doGet(t, ts, "/api/v1/suites/proj/runs?limit=abc")
	if status != http.StatusBadRequest {
		t.Fatalf("GET limit=abc = %d (%s), want 400", status, body)
	}
	if !strings.Contains(body, "limit must be a positive integer") {
		t.Errorf("bad-limit body = %s, want the refusal sentence", body)
	}

	status, _ = doGet(t, ts, "/api/v1/suites/proj/runs?limit=0")
	if status != http.StatusBadRequest {
		t.Fatalf("GET limit=0 = %d, want 400", status)
	}
}
