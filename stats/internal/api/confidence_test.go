package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

func (f *fakeStore) RecordConfidenceEstimate(_ context.Context, projectKey string, in records.ConfidenceEstimate) (records.ConfidenceEstimate, error) {
	if in.Score < 1 || in.Score > 99 {
		return records.ConfidenceEstimate{}, fmt.Errorf("%w: score", store.ErrConfidenceEstimateInvalid)
	}
	in.ID, in.Project = int64(len(f.confidence)+1), projectKey
	f.confidence = append(f.confidence, in)
	return in, nil
}

func (f *fakeStore) ConfidenceTags(context.Context) ([]records.ConfidenceTag, error) {
	return []records.ConfidenceTag{{Tag: "missing-fixture", Count: len(f.confidence)}}, nil
}

func (f *fakeStore) ConfidenceEstimatesPage(_ context.Context, tag string, limit, offset int) ([]records.ConfidenceEstimate, int, error) {
	out := []records.ConfidenceEstimate{}
	for _, e := range f.confidence {
		if tag == "" || slices.Contains(e.Tags, tag) {
			out = append(out, e)
		}
	}
	return out[min(offset, len(out)):min(offset+limit, len(out))], len(out), nil
}

// TestConfidenceRoutes: POST records an estimate under the path's project and
// change (201) and refuses a store-invalid one (400); the tags route lists the
// tags; the page route answers the rows of one tag with the total, the paging
// echo and the tag counts, and refuses a bad limit (400).
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
	if err := json.Unmarshal(raw, &got); err != nil || got.ID != 1 || got.Change != "kan-1" || got.Project != "proj" || got.Task != "2+3" {
		t.Errorf("POST echoed %s (%v), want id 1, change kan-1 from the path, project proj", raw, err)
	}
	if resp, raw := postJSON(t, post, map[string]any{"stage": "implement", "task": "1", "score": 100, "gaps": []string{"g"}, "tags": []string{"x"}}); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("POST score 100 = %d (%s), want 400", resp.StatusCode, raw)
	}

	if code, body := doGet(t, ts, "/api/v1/confidence-tags"); code != http.StatusOK || body != `[{"tag":"missing-fixture","count":1}]`+"\n" {
		t.Errorf("GET tags = %d %q", code, body)
	}

	code, body := doGet(t, ts, "/api/v1/stats/confidence-estimates?tag=missing-fixture&limit=5")
	var page struct {
		Total, Limit, Offset int
		Tags                 []records.ConfidenceTag
		Rows                 []records.ConfidenceEstimate
	}
	if err := json.Unmarshal([]byte(body), &page); code != http.StatusOK || err != nil {
		t.Fatalf("GET page = %d %q (%v)", code, body, err)
	}
	if page.Total != 1 || page.Limit != 5 || page.Offset != 0 || len(page.Rows) != 1 || page.Rows[0].Score != 80 || len(page.Tags) != 1 {
		t.Errorf("page = %+v, want the one row, total 1, limit 5 and the tag counts", page)
	}
	if code, body := doGet(t, ts, "/api/v1/stats/confidence-estimates?tag=other"); code != http.StatusOK || !json.Valid([]byte(body)) || !containsRowsEmpty(body) {
		t.Errorf("GET page for an unused tag = %d %q, want an empty rows array", code, body)
	}
	if code, _ := doGet(t, ts, "/api/v1/stats/confidence-estimates?limit=0"); code != http.StatusBadRequest {
		t.Errorf("GET page limit=0 = %d, want 400", code)
	}
}

func containsRowsEmpty(body string) bool {
	var p struct{ Rows []json.RawMessage }
	return json.Unmarshal([]byte(body), &p) == nil && p.Rows != nil && len(p.Rows) == 0
}
