package api_test

import (
	"net/http"
	"testing"
)

// TestPostPromptChange pins POST /api/v1/prompt-changes: a valid row is
// upserted and answered 204; a missing sha or an area outside
// review|impl|vv|fix is 400 and reaches no store.
func TestPostPromptChange(t *testing.T) {
	sts := &statsFake{}
	ts := newStatsTestServer(t, sts)
	url := ts.URL + "/api/v1/prompt-changes"

	row := map[string]any{"sha": "abc:skills/flow/test-sanity-reviewer-prompt.md", "at": "2026-10-27T10:00:00Z",
		"kind": "tuning", "area": "review", "slot": "test-sanity", "target": "skills/flow/test-sanity-reviewer-prompt.md",
		"summary": "missed assertion", "caseId": "kan-1-x-1"}
	resp, body := postJSON(t, url, row)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d, body %s; want 204", resp.StatusCode, body)
	}
	if len(sts.promptChanges) != 1 {
		t.Fatalf("upserted %d rows, want 1", len(sts.promptChanges))
	}
	got := sts.promptChanges[0]
	if got.SHA != row["sha"] || got.Kind != "tuning" || got.Area != "review" || got.Slot == nil || *got.Slot != "test-sanity" ||
		got.CaseID == nil || *got.CaseID != "kan-1-x-1" || got.At.Format("2006-01-02T15:04:05Z07:00") != "2026-10-27T10:00:00Z" {
		t.Errorf("upserted %+v, want the posted row", got)
	}

	for name, mutate := range map[string]func(map[string]any){
		"unknown area": func(m map[string]any) { m["area"] = "planner" },
		"missing sha":  func(m map[string]any) { delete(m, "sha") },
		"unknown kind": func(m map[string]any) { m["kind"] = "guess" },
	} {
		bad := map[string]any{}
		for k, v := range row {
			bad[k] = v
		}
		mutate(bad)
		resp, body := postJSON(t, url, bad)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: status %d, body %s; want 400", name, resp.StatusCode, body)
		}
	}
	if len(sts.promptChanges) != 1 {
		t.Errorf("a refused body reached the store: %d rows", len(sts.promptChanges))
	}
}
