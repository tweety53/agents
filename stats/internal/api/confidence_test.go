package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

func (f *fakeStore) RecordConfidenceEstimate(_ context.Context, projectKey string, in records.ConfidenceEstimate) (records.ConfidenceEstimate, error) {
	if in.Score < 1 || in.Score > 99 {
		return records.ConfidenceEstimate{}, fmt.Errorf("%w: score", store.ErrConfidenceEstimateInvalid)
	}
	in.ID, in.Project, in.ReviewStatus = int64(len(f.confidence)+1), projectKey, "to-review"
	f.confidence = append(f.confidence, in)
	return in, nil
}

func (f *fakeStore) ConfidenceTags(context.Context) ([]records.ConfidenceTag, error) {
	return []records.ConfidenceTag{{Tag: "missing-fixture", Count: len(f.confidence)}}, nil
}

func (f *fakeStore) ConfidenceEstimatesPage(_ context.Context, status string, limit, offset int) (store.ConfidencePage, error) {
	if status != "to-review" && status != "approved" && status != "declined" {
		return store.ConfidencePage{}, fmt.Errorf("%w: status", store.ErrConfidenceEstimateInvalid)
	}
	p := store.ConfidencePage{}
	out := []records.ConfidenceEstimate{}
	for _, e := range f.confidence {
		switch e.ReviewStatus {
		case "to-review":
			p.ToReview++
		case "approved":
			p.Approved++
		case "declined":
			p.Declined++
		}
		if e.ReviewStatus == status {
			out = append(out, e)
		}
	}
	p.Total, p.Rows = len(out), out[min(offset, len(out)):min(offset+limit, len(out))]
	return p, nil
}

func (f *fakeStore) ConfidenceEstimatesForChange(_ context.Context, projectKey, change string) ([]records.ConfidenceEstimate, error) {
	out := []records.ConfidenceEstimate{}
	for _, e := range f.confidence {
		if e.Project == projectKey && e.Change == change {
			out = append(out, e)
		}
	}
	return out, nil
}

func (f *fakeStore) ReviewConfidenceEstimate(_ context.Context, projectKey string, id int64, status, reason string) error {
	for i, e := range f.confidence {
		if e.Project == projectKey && e.ID == id {
			f.confidence[i].ReviewStatus, f.confidence[i].Reason = status, reason
			return nil
		}
	}
	return store.ErrConfidenceEstimateNotFound
}

// TestConfidenceRoutes: POST records an estimate under the path's project and
// change (201), to-review, and refuses a store-invalid one (400); the tags
// route lists the tags; the page route answers the rows of one review status
// with the total, the paging echo and every status's count, and refuses a
// bad limit or a missing status (400); PATCH review moves the row to another
// status with its reason (204) and answers 404 for an unknown id.
func TestConfidenceRoutes(t *testing.T) {
	t.Parallel()
	fs := newFakeStore()
	ts := newTestServer(t, fs)
	post := ts.URL + "/api/v1/confidence/proj/kan-1"

	resp, raw := postJSON(t, post, map[string]any{"change": "ignored", "stage": "implement", "task": "2+3", "score": 80, "gaps": []string{"g"}, "tags": []string{"missing-fixture"}})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST = %d (%s), want 201", resp.StatusCode, raw)
	}
	var got records.ConfidenceEstimate
	if err := json.Unmarshal(raw, &got); err != nil || got.ID != 1 || got.Change != "kan-1" || got.Project != "proj" || got.Task != "2+3" || got.ReviewStatus != "to-review" {
		t.Errorf("POST echoed %s (%v), want id 1, change kan-1 from the path, project proj, to-review", raw, err)
	}
	if resp, raw := postJSON(t, post, map[string]any{"stage": "implement", "task": "1", "score": 100, "gaps": []string{"g"}, "tags": []string{"x"}}); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("POST score 100 = %d (%s), want 400", resp.StatusCode, raw)
	}

	if code, body := doGet(t, ts, "/api/v1/confidence/proj/kan-1"); code != http.StatusOK || !strings.Contains(body, `"score":80`) {
		t.Errorf("GET change estimates = %d %q, want the recorded row", code, body)
	}
	if code, body := doGet(t, ts, "/api/v1/confidence/proj/other"); code != http.StatusOK || body != "[]\n" {
		t.Errorf("GET another change's estimates = %d %q, want []", code, body)
	}

	if code, body := doGet(t, ts, "/api/v1/confidence-tags"); code != http.StatusOK || body != `[{"tag":"missing-fixture","count":1}]`+"\n" {
		t.Errorf("GET tags = %d %q", code, body)
	}

	type pageDTO struct {
		Total, ToReview, Approved, Declined, Limit, Offset int
		Rows                                               []records.ConfidenceEstimate
	}
	read := func(status string) pageDTO {
		t.Helper()
		code, body := doGet(t, ts, "/api/v1/stats/confidence-estimates?status="+status+"&limit=5")
		var page pageDTO
		if err := json.Unmarshal([]byte(body), &page); code != http.StatusOK || err != nil {
			t.Fatalf("GET page %s = %d %q (%v)", status, code, body, err)
		}
		return page
	}
	if page := read("to-review"); page.Total != 1 || page.ToReview != 1 || page.Approved != 0 || page.Limit != 5 || page.Offset != 0 || len(page.Rows) != 1 || page.Rows[0].Score != 80 {
		t.Errorf("page = %+v, want the one row, total 1, limit 5 and the status counts", page)
	}
	if code, body := doGet(t, ts, "/api/v1/stats/confidence-estimates?status=approved"); code != http.StatusOK || !containsRowsEmpty(body) {
		t.Errorf("GET page for a status with no row = %d %q, want an empty rows array", code, body)
	}
	if code, _ := doGet(t, ts, "/api/v1/stats/confidence-estimates?status=to-review&limit=0"); code != http.StatusBadRequest {
		t.Errorf("GET page limit=0 = %d, want 400", code)
	}
	if code, _ := doGet(t, ts, "/api/v1/stats/confidence-estimates"); code != http.StatusBadRequest {
		t.Errorf("GET page with no status = %d, want 400", code)
	}

	review := ts.URL + "/api/v1/confidence/proj/estimates/"
	if code, raw := patchJSON(t, review+"1/review", map[string]any{"status": "approved", "reason": "worth closing"}); code != http.StatusNoContent {
		t.Fatalf("PATCH review = %d (%s), want 204", code, raw)
	}
	if page := read("approved"); page.Total != 1 || page.ToReview != 0 || page.Approved != 1 || len(page.Rows) != 1 || page.Rows[0].ReviewStatus != "approved" || page.Rows[0].Reason != "worth closing" {
		t.Errorf("approved page = %+v, want the row with its reason", page)
	}
	if code, _ := patchJSON(t, review+"9/review", map[string]any{"status": "approved"}); code != http.StatusNotFound {
		t.Errorf("PATCH review of an unknown id = %d, want 404", code)
	}
	if code, _ := patchJSON(t, review+"x/review", map[string]any{"status": "approved"}); code != http.StatusBadRequest {
		t.Errorf("PATCH review of a non-numeric id = %d, want 400", code)
	}
}

func containsRowsEmpty(body string) bool {
	var p struct{ Rows []json.RawMessage }
	return json.Unmarshal([]byte(body), &p) == nil && p.Rows != nil && len(p.Rows) == 0
}
