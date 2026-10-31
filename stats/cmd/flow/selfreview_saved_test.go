package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/storetest"
)

// linkedWorktree adds a worktree of repo on spectre/<name> and returns its
// path: a /flow change's canonical worktree.
func linkedWorktree(t *testing.T, repo, name string) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), "wt")
	if out, err := exec.Command("git", "-C", repo, "worktree", "add", "-q", "-b", "spectre/"+name, wt).CombinedOutput(); err != nil {
		t.Fatalf("worktree add: %v\n%s", err, out)
	}
	return wt
}

// TestSelfReviewBundleWritesTheContextFile drives the bundle end to end
// against a real flowd and store from inside the change's worktree: the
// daemon reads the gitignored change directory off disk, and the CLI prints
// the bundle and writes the same bytes to
// .superpowers/sdd/<change>/self-review-context.md.
func TestSelfReviewBundleWritesTheContextFile(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	wt := linkedWorktree(t, repo, "demo")
	writeChangeFile(t, wt, "spectre/changes/demo/design.md", "# demo design, never committed\n")

	st := storetest.NewStore(t)
	projectKey, mainCheckout, err := fallback.ProjectKey(wt)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.PutChange(context.Background(), store.Change{
		ProjectKey: projectKey, MainCheckoutPath: mainCheckout, Name: "demo",
		State: store.StateInProgress, UpdatedAt: time.Now(), UpdatedBy: "test",
	}); err != nil {
		t.Fatalf("seed change: %v", err)
	}
	srv, err := api.New(config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}, st, st, st, st, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	var stdout, stderr strings.Builder
	if code := run(context.Background(), []string{"self-review", "bundle", "-addr", ts.URL, "-C", wt, "-change", "demo"},
		strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "## spectre/changes/demo/design.md\n\n# demo design, never committed") {
		t.Errorf("bundle missing the on-disk design.md:\n%s", stdout.String())
	}
	written, err := os.ReadFile(filepath.Join(wt, ".superpowers/sdd/demo/self-review-context.md"))
	if err != nil {
		t.Fatalf("context file not written: %v", err)
	}
	if string(written) != stdout.String() {
		t.Errorf("context file differs from the printed bundle:\n%s", written)
	}
}

// TestSelfReviewBundleWithoutWorktreeReadsTheSavedBundle pins the deferred
// path: run from the main checkout, the CLI asks for the saved bundle --
// never the assembly route -- and prints it; flowd holding none is exit 1.
func TestSelfReviewBundleWithoutWorktreeReadsTheSavedBundle(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	saved := true
	var paths []string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if !saved {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte("# Self-review context bundle for demo (saved)\n"))
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	if code := run(context.Background(), []string{"self-review", "bundle", "-addr", srv.URL, "-C", repo, "-change", "demo"},
		strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if stdout.String() != "# Self-review context bundle for demo (saved)\n" {
		t.Errorf("stdout = %q", stdout.String())
	}
	if len(paths) != 1 || !strings.HasSuffix(paths[0], "/demo/files/self-review-context.md") {
		t.Errorf("requests = %v, want only the saved bundle's file", paths)
	}

	other := linkedWorktree(t, repo, "other")
	stdout.Reset()
	paths = nil
	if code := run(context.Background(), []string{"self-review", "bundle", "-addr", srv.URL, "-C", other, "-change", "demo"},
		strings.NewReader(""), &stdout, &stderr); code != 0 || len(paths) != 1 || !strings.HasSuffix(paths[0], "/demo/files/self-review-context.md") {
		t.Errorf("another change's worktree: exit = %d requests = %v, want the saved bundle only", code, paths)
	}
	if _, err := os.Stat(filepath.Join(other, ".superpowers/sdd/demo/self-review-context.md")); err == nil {
		t.Errorf("another change's worktree gained a context file")
	}

	saved = false
	stdout.Reset()
	if code := run(context.Background(), []string{"self-review", "bundle", "-addr", srv.URL, "-C", repo, "-change", "demo"},
		strings.NewReader(""), &stdout, &stderr); code != 1 || stdout.Len() != 0 {
		t.Errorf("no saved bundle: exit = %d stdout = %q, want 1 and nothing", code, stdout.String())
	}
}
