package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestStageCompletedCLI pins `flow stage completed`: exit 0 when flowd
// answers completed, 1 when it answers not, 2 when it cannot be reached,
// and the request names the project, change and stage.
func TestStageCompletedCLI(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	answer := `{"completed":true}`
	var gotPath string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(answer))
	}))
	defer srv.Close()

	call := func(addr string) int {
		var stdout, stderr strings.Builder
		return run(context.Background(), []string{"stage", "completed", "-addr", addr, "-timeout", "300ms", "-C", repo, "-stage", "flow.self-review", "demo"},
			strings.NewReader(""), &stdout, &stderr)
	}
	if code := call(srv.URL); code != 0 {
		t.Errorf("completed: exit = %d, want 0", code)
	}
	if !strings.HasSuffix(gotPath, "/demo/flow.self-review/completed") {
		t.Errorf("request path = %s", gotPath)
	}
	answer = `{"completed":false}`
	if code := call(srv.URL); code != 1 {
		t.Errorf("not completed: exit = %d, want 1", code)
	}
	if code := call(deadPortAddr(t)); code != 2 {
		t.Errorf("unreachable: exit = %d, want 2", code)
	}
}
