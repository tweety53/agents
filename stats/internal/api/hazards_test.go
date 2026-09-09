package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// hazardRecord is fakeStore's in-memory stand-in for a hazards row -- the
// owning identity kept beside the row, the same shape incidentRecord uses.
type hazardRecord struct {
	hazard     records.Hazard
	projectKey string
}

// AddHazard mirrors store.Store.AddHazard: rows are allocated ids, active,
// stamped when the caller left CreatedAt zero, and a (project, name) the
// project already holds is store.ErrHazardDuplicate.
func (f *fakeStore) AddHazard(_ context.Context, projectKey string, h records.Hazard) (records.Hazard, error) {
	if f.addHazardErr != nil {
		return records.Hazard{}, f.addHazardErr
	}
	for _, hr := range f.hazards {
		if hr.projectKey == projectKey && hr.hazard.Name == h.Name {
			return records.Hazard{}, fmt.Errorf("%w: %s/%s", store.ErrHazardDuplicate, projectKey, h.Name)
		}
	}
	f.nextHazardID++
	out := h
	out.ID = f.nextHazardID
	out.Active = true
	if out.CreatedAt.IsZero() {
		out.CreatedAt = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC).Add(time.Duration(f.nextHazardID) * time.Second)
	}
	f.hazards = append(f.hazards, hazardRecord{hazard: out, projectKey: projectKey})
	return out, nil
}

// ListHazards mirrors store.Store.ListHazards: newest first, the closed-set
// shape filtering to applies IN ('all', shape), the empty shape unfiltered,
// includeInactive lifting the active filter, and a non-empty shape outside
// the vocabulary store.ErrInvalidHazardShape.
func (f *fakeStore) ListHazards(_ context.Context, projectKey, shape string, includeInactive bool) ([]records.Hazard, error) {
	if f.listHazardsErr != nil {
		return nil, f.listHazardsErr
	}
	if shape != "" && shape != "all" && shape != "cross-repo" && shape != "single-repo" {
		return nil, store.ErrInvalidHazardShape
	}
	var out []records.Hazard
	for i := len(f.hazards) - 1; i >= 0; i-- {
		hr := f.hazards[i]
		if hr.projectKey != projectKey {
			continue
		}
		if shape != "" && hr.hazard.Applies != "all" && hr.hazard.Applies != shape {
			continue
		}
		if !hr.hazard.Active && !includeInactive {
			continue
		}
		out = append(out, hr.hazard)
	}
	return out, nil
}

// RetireHazard mirrors store.Store.RetireHazard: active=false on the named
// row, returned; an unknown name is store.ErrHazardNotFound.
func (f *fakeStore) RetireHazard(_ context.Context, projectKey, name string) (records.Hazard, error) {
	for i := range f.hazards {
		if f.hazards[i].projectKey == projectKey && f.hazards[i].hazard.Name == name {
			f.hazards[i].hazard.Active = false
			return f.hazards[i].hazard, nil
		}
	}
	return records.Hazard{}, fmt.Errorf("%w: %s/%s", store.ErrHazardNotFound, projectKey, name)
}

// hazardBody is the wire body a hazard POST carries.
func hazardBody(name, body, applies string) map[string]any {
	return map[string]any{"name": name, "body": body, "applies": applies}
}

// TestHazardRoutesAddListRetire pins the row's whole life over HTTP: POST
// 201 with the stored row echoed, GET listing newest first with the shape
// filter the store documents, PATCH retiring rather than deleting, and the
// retired row leaving the default listing.
func TestHazardRoutesAddListRetire(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + "/api/v1/hazards/proj"

	resp, body := postJSON(t, url, hazardBody("guard-reverts-head-only", "inspect before retrying", "all"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST hazards = %d (%s), want 201", resp.StatusCode, body)
	}
	var got records.Hazard
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if got.ID == 0 || !got.Active {
		t.Errorf("response = %+v, want an allocated id and active=true", got)
	}
	second, secondBody := postJSON(t, url, hazardBody("guard-single-repo", "pass the change's own worktree", "single-repo"))
	if second.StatusCode != http.StatusCreated {
		t.Fatalf("second POST hazards = %d (%s), want 201", second.StatusCode, secondBody)
	}

	gstatus, gbody := doGet(t, ts, "/api/v1/hazards/proj?shape=single-repo")
	if gstatus != http.StatusOK {
		t.Fatalf("GET hazards = %d (%s), want 200", gstatus, gbody)
	}
	var listed []records.Hazard
	if err := json.Unmarshal([]byte(gbody), &listed); err != nil {
		t.Fatalf("decode listing %s: %v", gbody, err)
	}
	if len(listed) != 2 || listed[0].Name != "guard-single-repo" || listed[1].Name != "guard-reverts-head-only" {
		t.Errorf("shape=single-repo listing = %+v, want newest first with the all row beside the single-repo row", listed)
	}

	gstatus, gbody = doGet(t, ts, "/api/v1/hazards/proj?shape=cross-repo")
	if gstatus != http.StatusOK {
		t.Fatalf("GET hazards shape=cross-repo = %d (%s), want 200", gstatus, gbody)
	}
	listed = nil
	if err := json.Unmarshal([]byte(gbody), &listed); err != nil {
		t.Fatalf("decode listing %s: %v", gbody, err)
	}
	if len(listed) != 1 || listed[0].Name != "guard-reverts-head-only" {
		t.Errorf("shape=cross-repo listing = %+v, want only the all row", listed)
	}

	req, err := http.NewRequest(http.MethodPatch, url+"/guard-reverts-head-only", nil)
	if err != nil {
		t.Fatalf("build PATCH: %v", err)
	}
	patchResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH hazards: %v", err)
	}
	defer patchResp.Body.Close()
	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("PATCH hazards = %d, want 200", patchResp.StatusCode)
	}

	gstatus, gbody = doGet(t, ts, "/api/v1/hazards/proj")
	if gstatus != http.StatusOK {
		t.Fatalf("GET hazards after retire = %d (%s), want 200", gstatus, gbody)
	}
	listed = nil
	if err := json.Unmarshal([]byte(gbody), &listed); err != nil {
		t.Fatalf("decode listing %s: %v", gbody, err)
	}
	if len(listed) != 1 || listed[0].Name != "guard-single-repo" {
		t.Errorf("listing after retire = %+v, want only the active row", listed)
	}
	if len(fs.hazards) != 2 {
		t.Errorf("store holds %d hazard rows, want 2 -- retire is active=false, never a delete", len(fs.hazards))
	}
}

// TestHazardRouteRejectsInvalidApplies pins the closed-vocabulary refusal
// on both POST's body and GET's shape query -- a caller mistake caught
// before the store is touched.
func TestHazardRouteRejectsInvalidApplies(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + "/api/v1/hazards/proj"

	before := len(fs.hazards)
	resp, body := postJSON(t, url, hazardBody("probe", "b", "always"))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST hazards applies=always = %d (%s), want 400", resp.StatusCode, body)
	}
	if len(fs.hazards) != before {
		t.Errorf("the store was reached for an applies outside the closed vocabulary")
	}

	gstatus, gbody := doGet(t, ts, "/api/v1/hazards/proj?shape=bogus")
	if gstatus != http.StatusBadRequest {
		t.Fatalf("GET hazards shape=bogus = %d (%s), want 400", gstatus, gbody)
	}
}

// TestHazardRouteDuplicateConflict pins the unique constraint's answer: a
// second POST under the same (project, name) is a 409, not a duplicate row.
func TestHazardRouteDuplicateConflict(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + "/api/v1/hazards/proj"

	if resp, body := postJSON(t, url, hazardBody("dup", "b", "all")); resp.StatusCode != http.StatusCreated {
		t.Fatalf("first POST hazards = %d (%s), want 201", resp.StatusCode, body)
	}
	resp, body := postJSON(t, url, hazardBody("dup", "b again", "all"))
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("second POST hazards = %d (%s), want 409", resp.StatusCode, body)
	}
	if len(fs.hazards) != 1 {
		t.Errorf("store holds %d hazard rows, want 1 after the refused duplicate", len(fs.hazards))
	}
}

// TestHazardRouteRetireUnknownName pins the 404 for a retire naming no
// hazard the project holds.
func TestHazardRouteRetireUnknownName(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	req, err := http.NewRequest(http.MethodPatch, ts.URL+"/api/v1/hazards/proj/never-recorded", nil)
	if err != nil {
		t.Fatalf("build PATCH: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH hazards: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("PATCH unknown hazard = %d, want 404", resp.StatusCode)
	}
}

// TestHazardRouteListAllIncludesRetired pins the all=true query lifting the
// active filter -- the operator's -all view.
func TestHazardRouteListAllIncludesRetired(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + "/api/v1/hazards/proj"

	if resp, body := postJSON(t, url, hazardBody("retired", "b", "all")); resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST hazards = %d (%s), want 201", resp.StatusCode, body)
	}
	req, _ := http.NewRequest(http.MethodPatch, url+"/retired", nil)
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("PATCH hazards: %v", err)
	}

	gstatus, gbody := doGet(t, ts, "/api/v1/hazards/proj?all=true")
	if gstatus != http.StatusOK {
		t.Fatalf("GET hazards all=true = %d (%s), want 200", gstatus, gbody)
	}
	var listed []records.Hazard
	if err := json.Unmarshal([]byte(gbody), &listed); err != nil {
		t.Fatalf("decode listing %s: %v", gbody, err)
	}
	if len(listed) != 1 || listed[0].Active {
		t.Errorf("all=true listing = %+v, want the retired row included", listed)
	}
}
