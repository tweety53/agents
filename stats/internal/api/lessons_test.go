package api_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// lessonsRepo lays down one fake project root whose tree the scanner can
// read, and points the fake store at it.
func lessonsRepo(t *testing.T) (*httptest.Server, *fakeStore, string) {
	t.Helper()
	root := t.TempDir()
	brief := filepath.Join(root, "docs", "briefs", "flake-bisection.md")
	if err := os.MkdirAll(filepath.Dir(brief), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(brief, []byte("# Flake bisection brief\n\nStart small.\n"), 0o644); err != nil {
		t.Fatalf("write brief: %v", err)
	}

	fs := newFakeStore()
	fs.projectRoots = []store.ProjectRoot{{ProjectKey: "proj-key", MainCheckoutPath: root}}
	return newTestServer(t, fs), fs, root
}

// TestLessonsResolveServesScannedAnswer drives the route end to end: the
// answer is assembled from the registered root's files, never from the
// store, and renders the brief in full.
func TestLessonsResolveServesScannedAnswer(t *testing.T) {
	ts, _, _ := lessonsRepo(t)

	resp, err := http.Get(ts.URL + "/api/v1/lessons/resolve?topic=flake-bisection")
	if err != nil {
		t.Fatalf("GET resolve: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET resolve = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/markdown") {
		t.Errorf("Content-Type = %q, want text/markdown", ct)
	}
	raw, _ := io.ReadAll(resp.Body)
	body := string(raw)
	for _, want := range []string{
		"# Lessons: flake-bisection",
		"found: 1 briefs, 0 narrative mentions",
		"# Flake bisection brief",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("answer missing %q:\n%s", want, body)
		}
	}
}

func TestLessonsResolveMissingTopicIsACallerMistake(t *testing.T) {
	ts, _, _ := lessonsRepo(t)
	resp, err := http.Get(ts.URL + "/api/v1/lessons/resolve")
	if err != nil {
		t.Fatalf("GET resolve: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("GET resolve without topic = %d, want 400", resp.StatusCode)
	}
}

func TestLessonsResolveUnreadableRootIsANoteNotAnError(t *testing.T) {
	ts, fs, _ := lessonsRepo(t)
	fs.projectRoots = append(fs.projectRoots, store.ProjectRoot{ProjectKey: "gone-key", MainCheckoutPath: filepath.Join(t.TempDir(), "missing")})

	resp, err := http.Get(ts.URL + "/api/v1/lessons/resolve?topic=flake")
	if err != nil {
		t.Fatalf("GET resolve: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET resolve = %d, want 200 -- one broken root is not a 5xx", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(raw), "note: ") || !strings.Contains(string(raw), "could not be read") {
		t.Errorf("answer must name the unreadable root:\n%s", raw)
	}
}

func TestLessonsResolveStoreFailureIs5xx(t *testing.T) {
	ts, fs, _ := lessonsRepo(t)
	fs.projectRootsErr = context.DeadlineExceeded
	resp, err := http.Get(ts.URL + "/api/v1/lessons/resolve?topic=flake")
	if err != nil {
		t.Fatalf("GET resolve: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 500 {
		t.Fatalf("GET resolve = %d, want 5xx on a store read failure", resp.StatusCode)
	}
}
