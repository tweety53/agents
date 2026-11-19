package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A STARTED change goes IN_PROGRESS with no run instructions and no open
// flow.write-in-progress run.
func TestStateSetInProgressNeedsNoRunInstructions(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)
	put := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/stage-runs":
			_, _ = w.Write([]byte(`{"stageRuns":[]}`))
		case r.Method == http.MethodPut:
			put = true
			_, _ = w.Write([]byte(`{"state":"IN_PROGRESS"}`))
		default:
			_, _ = w.Write([]byte(`{"state":"STARTED"}`))
		}
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		env.get, strings.NewReader(`{"state":"IN_PROGRESS"}`), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !put {
		t.Error("store not written")
	}
}
