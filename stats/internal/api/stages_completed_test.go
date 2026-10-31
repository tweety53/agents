package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// stageCompletedAnswer GETs the stage-completed route and decodes it.
func stageCompletedAnswer(t *testing.T, base, project, change, stage string) bool {
	t.Helper()
	status, _, body := doRequest(t, http.MethodGet,
		base+"/api/v1/stages/"+project+"/"+change+"/"+stage+"/completed", nil)
	if status != http.StatusOK {
		t.Fatalf("GET completed status %d, body %s", status, body)
	}
	var out struct{ Completed bool }
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return out.Completed
}

func selfReviewBegin(projectKey, change string) store.BeginStageInput {
	session := "session-1"
	return store.BeginStageInput{
		ProjectKey: projectKey, ChangeName: change, Harness: "claude-code",
		SessionID: &session, Command: "/flow", Stage: "flow.self-review",
	}
}

// TestStageCompletedRouteFalseBeforeEnd pins the route's false answer: a
// begun, not yet ended flow.self-review run has not completed.
func TestStageCompletedRouteFalseBeforeEnd(t *testing.T) {
	st := newIntegrationStore(t)
	projectKey := fmt.Sprintf("proj-stage-completed-open-%d", time.Now().UnixNano())
	mustPutIntegrationChange(t, st, projectKey, "kan-1073", nil)
	ts := newIntegrationTestServer(t, st)

	if stageCompletedAnswer(t, ts.URL, projectKey, "kan-1073", "flow.self-review") {
		t.Fatal("completed = true with no stage run")
	}
	if _, err := st.BeginStage(t.Context(), selfReviewBegin(projectKey, "kan-1073")); err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if stageCompletedAnswer(t, ts.URL, projectKey, "kan-1073", "flow.self-review") {
		t.Error("completed = true for a run that has not ended")
	}
}

// TestStageCompletedRouteTrueAfterCompletedEnd pins the true answer after a
// completed end, scoped to the change: a second change's completed run does
// not count for the first.
func TestStageCompletedRouteTrueAfterCompletedEnd(t *testing.T) {
	st := newIntegrationStore(t)
	projectKey := fmt.Sprintf("proj-stage-completed-done-%d", time.Now().UnixNano())
	mustPutIntegrationChange(t, st, projectKey, "kan-1073", nil)
	mustPutIntegrationChange(t, st, projectKey, "other", nil)
	ts := newIntegrationTestServer(t, st)

	mustRunIntegrationStage(t, st, selfReviewBegin(projectKey, "other"), nil, time.Now())
	if stageCompletedAnswer(t, ts.URL, projectKey, "kan-1073", "flow.self-review") {
		t.Fatal("completed = true from another change's run")
	}
	mustRunIntegrationStage(t, st, selfReviewBegin(projectKey, "kan-1073"), nil, time.Now())
	if !stageCompletedAnswer(t, ts.URL, projectKey, "kan-1073", "flow.self-review") {
		t.Error("completed = false after a completed end")
	}
}
