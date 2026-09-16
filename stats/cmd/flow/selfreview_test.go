package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRunSelfReviewBundlePrintsBundle drives `flow self-review bundle`
// end to end: the project key resolves from -C, the request lands on the
// bundle route for that project and change, and the body is printed
// verbatim on stdout with exit 0.
func TestRunSelfReviewBundlePrintsBundle(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotPath string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Self-review context bundle for demo\n\nfound: 1 of 6 sources\n"))
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"self-review", "bundle", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if !strings.HasSuffix(gotPath, "/demo/bundle") || !strings.Contains(gotPath, "/api/v1/self-review/") {
		t.Errorf("request path = %s", gotPath)
	}
	if !strings.Contains(stdout.String(), "# Self-review context bundle for demo") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

// TestRunSelfReviewBundleRequiresChange pins the caller mistake: no
// -change is exit 2 and a stderr line, never a request.
func TestRunSelfReviewBundleRequiresChange(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"self-review", "bundle", "-addr", "http://127.0.0.1:1", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "-change is required") {
		t.Errorf("stderr = %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want nothing", stdout.String())
	}
}

// TestRunSelfReviewBundleUnreachableStoreExitsNonZero carries the read
// contract through the CLI: a store that cannot answer is reported to
// stderr and exits non-zero -- never a partial bundle on stdout a caller
// could mistake for the change's own.
func TestRunSelfReviewBundleUnreachableStoreExitsNonZero(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"self-review", "bundle", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero (stderr: %s)", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want nothing on a failed read", stdout.String())
	}
	if !strings.Contains(stderr.String(), "self-review bundle") {
		t.Errorf("stderr = %q, want the failure named", stderr.String())
	}
}
