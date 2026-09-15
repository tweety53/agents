package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// errFakeInvalidSpecRun is the fake's stand-in for the store's own typed
// refusal, which the handler maps to 400 -- errFakeInvalidSuiteRun's same
// split.
var errFakeInvalidSpecRun = store.ErrInvalidSpecRun

// specRunRecord is fakeStore's in-memory stand-in for a spec_lastruns row
// -- the owning project key kept beside the row, the same shape
// suiteRunRecord uses.
type specRunRecord struct {
	row        records.SpecLastRun
	projectKey string
}

// RecordSpecRun mirrors store.Store.RecordSpecRun: one row per (project,
// spec), the LATEST timestamp kept -- an out-of-order older call never
// drags the row backwards -- and an empty spec refused at the boundary.
func (f *fakeStore) RecordSpecRun(_ context.Context, projectKey, spec string, ranAt time.Time) (records.SpecLastRun, error) {
	f.recordCalls++
	if f.recordSpecRunErr != nil {
		return records.SpecLastRun{}, f.recordSpecRunErr
	}
	if spec == "" {
		return records.SpecLastRun{}, errFakeInvalidSpecRun
	}
	if ranAt.IsZero() {
		ranAt = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC).Add(time.Duration(len(f.specRuns)) * time.Second)
	}
	for i := range f.specRuns {
		if f.specRuns[i].projectKey != projectKey || f.specRuns[i].row.Spec != spec {
			continue
		}
		if f.specRuns[i].row.LastRanAt == nil || ranAt.After(*f.specRuns[i].row.LastRanAt) {
			stamp := ranAt
			f.specRuns[i].row.LastRanAt = &stamp
		}
		return f.specRuns[i].row, nil
	}
	stamp := ranAt
	out := records.SpecLastRun{Spec: spec, LastRanAt: &stamp}
	f.specRuns = append(f.specRuns, specRunRecord{row: out, projectKey: projectKey})
	return out, nil
}

// ListSpecLastRuns mirrors store.Store.ListSpecLastRuns: named specs answer
// one row each, recorded or not, a never-recorded name carrying nil
// LastRanAt and the project's whole change count; the empty slice lists
// recorded specs alone. changesSince counts the project's changes whose
// UpdatedAt postdates the row -- the seeded changes carry zero times, so a
// recorded row's figure here is 0; the real count is pinned in the store's
// own tests.
func (f *fakeStore) ListSpecLastRuns(_ context.Context, projectKey string, specs []string) ([]records.SpecLastRun, error) {
	if f.listSpecRunsErr != nil {
		return nil, f.listSpecRunsErr
	}

	total := 0
	for _, c := range f.changes {
		if c.ProjectKey == projectKey {
			total++
		}
	}

	withRow := func(spec string) (records.SpecLastRun, bool) {
		for i := range f.specRuns {
			if f.specRuns[i].projectKey == projectKey && f.specRuns[i].row.Spec == spec {
				return f.specRuns[i].row, true
			}
		}
		return records.SpecLastRun{}, false
	}

	var out []records.SpecLastRun
	if len(specs) == 0 {
		for i := range f.specRuns {
			if f.specRuns[i].projectKey == projectKey {
				out = append(out, f.specRuns[i].row)
			}
		}
		return out, nil
	}
	for _, spec := range specs {
		if row, ok := withRow(spec); ok {
			out = append(out, row)
			continue
		}
		out = append(out, records.SpecLastRun{Spec: spec, LastRanAt: nil, ChangesSince: total})
	}
	// The store answers ORDER BY spec; the fake mirrors it rather than the
	// request order, so the handler's contract is what the test pins.
	sort.Slice(out, func(i, j int) bool { return out[i].Spec < out[j].Spec })
	return out, nil
}

// TestPostSpecRunRecordsAndListReadsTheArray pins the route pair: a POST
// 201 echoes the stored row with its stamp, and the GET carries named
// specs through -- a never-recorded name answering the never-run row, nil
// stamp and whole-history count included.
func TestPostSpecRunRecordsAndListReadsTheArray(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	resp, body := postJSON(t, ts.URL+"/api/v1/specs/proj/runs", map[string]any{"spec": "baseline.spec.ts"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST spec runs = %d (%s), want 201", resp.StatusCode, body)
	}
	var got records.SpecLastRun
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if got.Spec != "baseline.spec.ts" || got.LastRanAt == nil {
		t.Errorf("response = %+v, want the spec and a stamp", got)
	}

	resp, body = postJSON(t, ts.URL+"/api/v1/specs/proj/runs", map[string]any{"spec": "views.spec.ts"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("seed POST = %d (%s), want 201", resp.StatusCode, body)
	}

	status, listing := doGet(t, ts, "/api/v1/specs/proj/lastruns")
	if status != http.StatusOK {
		t.Fatalf("GET lastruns = %d (%s), want 200", status, listing)
	}
	var listed []records.SpecLastRun
	if err := json.Unmarshal([]byte(listing), &listed); err != nil {
		t.Fatalf("decode listing %s: %v", listing, err)
	}
	if len(listed) != 2 || listed[0].Spec != "baseline.spec.ts" {
		t.Fatalf("unnamed listing = %+v, want the two recorded specs ordered by spec", listed)
	}

	status, named := doGet(t, ts, "/api/v1/specs/proj/lastruns?spec=views.spec.ts&spec=controls.spec.ts")
	if status != http.StatusOK {
		t.Fatalf("named GET lastruns = %d (%s), want 200", status, named)
	}
	if err := json.Unmarshal([]byte(named), &listed); err != nil {
		t.Fatalf("decode named listing %s: %v", named, err)
	}
	if len(listed) != 2 {
		t.Fatalf("named listing = %d rows, want 2", len(listed))
	}
	if listed[0].Spec != "controls.spec.ts" || listed[0].LastRanAt != nil {
		t.Errorf("first named row = %+v, want the never-recorded spec with a nil stamp", listed[0])
	}
	if listed[0].ChangesSince != 1 {
		t.Errorf("never-run changesSince = %d, want 1 (the project's one seeded change)", listed[0].ChangesSince)
	}
	if listed[1].Spec != "views.spec.ts" {
		t.Errorf("second named row = %+v, want views.spec.ts", listed[1])
	}
}

// TestPostSpecRunRejectsAnEmptySpec pins the refusal path: a payload that
// cannot identify what it describes is a 400 before the store is touched,
// the same caller-mistake shape every record handler uses.
func TestPostSpecRunRejectsAnEmptySpec(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")

	resp, body := postJSON(t, ts.URL+"/api/v1/specs/proj/runs", map[string]any{"spec": ""})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST with empty spec = %d (%s), want 400", resp.StatusCode, body)
	}
	if fs.recordCalls != 0 {
		t.Errorf("store record calls = %d, want 0: a refused payload never reaches the store", fs.recordCalls)
	}
}
