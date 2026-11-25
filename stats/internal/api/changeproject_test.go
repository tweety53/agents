package api_test

import (
	"net/http"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// stageMarkBody is a /flow-interactive begin mark for change sent from a checkout of project.
func stageMarkBody(project, change string) map[string]any {
	return map[string]any{
		"projectKey":   project,
		"changeName":   change,
		"harness":      "claude-code",
		"sessionToken": "mf-change-project",
		"command":      "/flow-interactive",
		"stage":        "flow.verify",
		"startedAt":    "2026-11-25T10:00:00Z",
	}
}

// stageEndBody is the end mark matching stageMarkBody.
func stageEndBody(project, change string) map[string]any {
	return map[string]any{
		"projectKey": project,
		"changeName": change,
		"command":    "/flow-interactive",
		"stage":      "flow.verify",
		"endedAt":    "2026-11-25T10:05:00Z",
		"outcome":    "completed",
	}
}

// TestStageMarkFromAnotherProjectLandsOnTheChangesRow pins that a mark
// made from a checkout of another repository -- a /flow-interactive
// executor working in a second repo -- records on the one unfinished row of
// that name under its own project, begin and end alike, and never
// bootstraps a synthetic change under the cwd's project.
func TestStageMarkFromAnotherProjectLandsOnTheChangesRow(t *testing.T) {
	t.Parallel()
	fs := newFakeStore()
	fs.changes[changeKey("agents", "kan-1")] = store.Change{ProjectKey: "agents", Name: "kan-1", State: store.StateStarted}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	if resp, body := postJSON(t, srv.URL+"/api/v1/stages/begin", stageMarkBody("gymie", "kan-1")); resp.StatusCode != http.StatusOK {
		t.Fatalf("begin = %d (%s), want 200", resp.StatusCode, body)
	}
	if _, ok := fs.changes[changeKey("gymie", "kan-1")]; ok {
		t.Fatal("a synthetic gymie/kan-1 row was created; the mark belongs on agents/kan-1")
	}
	if len(fs.stageRuns) != 1 || fs.stageRuns[0].projectKey != "agents" {
		t.Fatalf("stage runs = %+v, want one on agents/kan-1", fs.stageRuns)
	}
	if resp, body := postJSON(t, srv.URL+"/api/v1/stages/end", stageEndBody("gymie", "kan-1")); resp.StatusCode != http.StatusOK {
		t.Fatalf("end = %d (%s), want 200", resp.StatusCode, body)
	}
	if fs.stageRuns[0].run.EndedAt == nil {
		t.Error("the end mark did not close the run on agents/kan-1")
	}
}

// TestStageMarkKeepsTheCwdProjectWhenItsRowExistsOrTheMatchIsAmbiguous
// pins the two cases the rule leaves on the cwd's project: a row of the
// name under the cwd's project wins over any other project's, and two
// unfinished rows elsewhere are never guessed between (the begin
// bootstraps under the cwd's project, as before).
func TestStageMarkKeepsTheCwdProjectWhenItsRowExistsOrTheMatchIsAmbiguous(t *testing.T) {
	t.Parallel()
	fs := newFakeStore()
	fs.changes[changeKey("agents", "own")] = store.Change{ProjectKey: "agents", Name: "own", State: store.StateStarted}
	fs.changes[changeKey("gymie", "own")] = store.Change{ProjectKey: "gymie", Name: "own", State: store.StateFinished}
	fs.changes[changeKey("a", "two")] = store.Change{ProjectKey: "a", Name: "two", State: store.StateStarted}
	fs.changes[changeKey("b", "two")] = store.Change{ProjectKey: "b", Name: "two", State: store.StateInProgress}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	for _, change := range []string{"own", "two"} {
		project := "gymie"
		if change == "two" {
			project = "c"
		}
		if resp, body := postJSON(t, srv.URL+"/api/v1/stages/begin", stageMarkBody(project, change)); resp.StatusCode != http.StatusOK {
			t.Fatalf("begin %s = %d (%s), want 200", change, resp.StatusCode, body)
		}
	}
	if len(fs.stageRuns) != 2 || fs.stageRuns[0].projectKey != "gymie" || fs.stageRuns[1].projectKey != "c" {
		t.Fatalf("stage runs = %+v, want gymie/own then c/two", fs.stageRuns)
	}
}

// TestRecordWriteFromAnotherProjectLandsOnTheChangesRow pins the same rule
// for the record routes: a dispatch recorded from another repository's
// checkout lands on the change's own row instead of a 404.
func TestRecordWriteFromAnotherProjectLandsOnTheChangesRow(t *testing.T) {
	t.Parallel()
	ts, fs := recordTestServer(t, "agents", "kan-1")

	resp, body := postJSON(t, ts.URL+recordsPath("gymie", "kan-1")+"/dispatches", dispatchBody("implementer", "opus"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST dispatches = %d (%s), want 201", resp.StatusCode, body)
	}
	if len(fs.dispatches) != 1 || fs.dispatches[0].projectKey != "agents" {
		t.Errorf("dispatches = %+v, want one on agents/kan-1", fs.dispatches)
	}
}
