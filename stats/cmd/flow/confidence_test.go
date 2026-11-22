package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
)

// confidenceRun drives `flow confidence <args>` against a daemon answering
// with handler and returns the exit code and both streams.
func confidenceRun(t *testing.T, handler http.HandlerFunc, args ...string) (int, string, string) {
	t.Helper()
	srv := httptest.NewServer(genuineDaemon(handler))
	t.Cleanup(srv.Close)
	var stdout, stderr strings.Builder
	argv := append([]string{"confidence", args[0], "-addr", srv.URL, "-timeout", "500ms"}, args[1:]...)
	code := run(context.Background(), argv, newTestEnv(t).get, strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// TestConfidenceRecordPostsTheEstimate: every flag reaches the store's route
// for the change, each -gap and -tag in order.
func TestConfidenceRecordPostsTheEstimate(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	var gotPath string
	var got records.ConfidenceEstimate
	code, _, stderr := confidenceRun(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1}`))
	}, "record", "-C", repo, "-change", "kan-1", "-stage", "implement", "-task", "2+3", "-score", "85",
		"-gap", "the fixture is missing", "-gap", "task 3's order is unstated", "-tag", "missing-fixture", "-tag", "ordering")
	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr)
	}
	if !strings.HasPrefix(gotPath, "POST /api/v1/confidence/") || !strings.HasSuffix(gotPath, "/kan-1") {
		t.Errorf("request = %s", gotPath)
	}
	if got.Stage != "implement" || got.Task != "2+3" || got.Score != 85 ||
		strings.Join(got.Gaps, "|") != "the fixture is missing|task 3's order is unstated" || strings.Join(got.Tags, "|") != "missing-fixture|ordering" {
		t.Errorf("body = %+v", got)
	}
}

// TestConfidenceRecordExits: 100 writes nothing and exits 0; a store refusal
// is exit 2; a store that cannot take the row is one warning and exit 0, so
// an implementer never stops on it; a missing -change is exit 2.
func TestConfidenceRecordExits(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	base := []string{"record", "-C", repo, "-change", "kan-1", "-stage", "plan", "-disposition", "later", "-gap", "g", "-tag", "t"}
	calls := 0
	answer := func(status int) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			calls++
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":"no"}`))
		}
	}
	if code, stdout, _ := confidenceRun(t, answer(http.StatusCreated), append(base, "-score", "100")...); code != 0 || calls != 0 || !strings.Contains(stdout, "nothing to file") {
		t.Errorf("score 100: exit %d, %d calls, stdout %q — want exit 0, no call", code, calls, stdout)
	}
	if code, _, stderr := confidenceRun(t, answer(http.StatusBadRequest), append(base, "-score", "70")...); code != 2 || !strings.Contains(stderr, "refused") {
		t.Errorf("refused: exit %d, stderr %q — want exit 2", code, stderr)
	}
	if code, _, stderr := confidenceRun(t, answer(http.StatusInternalServerError), append(base, "-score", "70")...); code != 0 || !strings.Contains(stderr, "not recorded") {
		t.Errorf("store down: exit %d, stderr %q — want a warning and exit 0", code, stderr)
	}
	if code, _, _ := confidenceRun(t, answer(http.StatusCreated), "record", "-C", repo, "-stage", "plan", "-score", "70"); code != 2 {
		t.Errorf("no -change: exit %d, want 2", code)
	}
}

// TestConfidenceTagsPrintsCountAndTag: one "<count> <tag>" line per tag in
// the daemon's order; a store that cannot answer is exit 1.
func TestConfidenceTagsPrintsCountAndTag(t *testing.T) {
	t.Parallel()
	code, stdout, stderr := confidenceRun(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/confidence-tags" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`[{"tag":"missing-fixture","count":3},{"tag":"ordering","count":1}]`))
	}, "tags")
	if code != 0 || stdout != "3 missing-fixture\n1 ordering\n" {
		t.Errorf("exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	if code, _, _ := confidenceRun(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }, "tags"); code != 1 {
		t.Errorf("store down: exit %d, want 1", code)
	}
}
