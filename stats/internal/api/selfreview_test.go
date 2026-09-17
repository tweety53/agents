package api_test

import (
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// selfReviewPath is the bundle route's path for one project/change; repo
// rides as the query parameter the caller resolves from its own location
// inside the repository.
func selfReviewPath(project, change, repo string) string {
	p := "/api/v1/self-review/" + project + "/" + change + "/bundle"
	if repo != "" {
		p += "?repo=" + repo
	}
	return p
}

// TestSelfReviewBundleHandlerServesAssembledBundle drives the route end to
// end over a real repository: the ledger renders from the recorded
// dispatches, and the archived design.md is read out of the
// chore/archive-<name> branch of the repository the request named — no
// path arrives from the store.
func TestSelfReviewBundleHandlerServesAssembledBundle(t *testing.T) {
	repo := bundleGitRepo(t)
	bundleCommit(t, repo, "spectre/changes/kan-1/tasks.md", "- [ ] 1. do it\n",
		"chore(spectre): plan and session records")
	bundleArchiveBranch(t, repo, "kan-1", map[string]string{
		"tasks.md":  "- [ ] 1. do it\n",
		"design.md": "# kan-1 design\n",
	})

	ts, _ := recordTestServer(t, "proj", "kan-1")
	if resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", dispatchBody("implementer", "opus")); resp.StatusCode != http.StatusCreated {
		t.Fatalf("seed POST dispatches = %d (%s), want 201", resp.StatusCode, body)
	}

	resp, err := http.Get(ts.URL + selfReviewPath("proj", "kan-1", repo))
	if err != nil {
		t.Fatalf("GET bundle: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET bundle = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/markdown") {
		t.Errorf("Content-Type = %q, want text/markdown", ct)
	}
	raw, _ := io.ReadAll(resp.Body)
	body := string(raw)

	for _, want := range []string{
		"# Self-review context bundle for kan-1",
		"## .superpowers/sdd/ledgers/kan-1.md",
		"## spectre/changes/archive/kan-1/design.md",
		"# kan-1 design",
		"## git log --stat",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("bundle missing %q:\n%s", want, body)
		}
	}
}

// TestSelfReviewBundleHandlerUnknownChange carries the gather's "a missing
// source is never fatal" rule through the route: a change the store has
// never heard of still serves a bundle whose store-side sections report
// skipped — ledger and panel both, never an empty panel record nobody
// wrote — not a 404 a caller could read as the endpoint refusing.
func TestSelfReviewBundleHandlerUnknownChange(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	code, body := doGet(t, ts, selfReviewPath("proj", "never-heard", t.TempDir()))
	if code != http.StatusOK {
		t.Fatalf("GET bundle = %d (%s), want 200", code, body)
	}
	if !strings.Contains(body, "skipped: .superpowers/sdd/ledgers/never-heard.md (absent)") {
		t.Errorf("unknown change's ledger not reported skipped:\n%s", body)
	}
	if !strings.Contains(body, "skipped: .superpowers/sdd/reviews/never-heard-panel.md (absent)") {
		t.Errorf("unknown change's panel not reported skipped:\n%s", body)
	}
}

// TestSelfReviewBundleHandlerRefusesRelativeRepo pins the one caller
// mistake the route itself can see: a repo parameter that is not an
// absolute path is a 400, before anything reads or renders.
func TestSelfReviewBundleHandlerRefusesRelativeRepo(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	code, body := doGet(t, ts, selfReviewPath("proj", "kan-1", "relative/path"))
	if code != http.StatusBadRequest {
		t.Fatalf("GET bundle = %d (%s), want 400", code, body)
	}
}

// TestSelfReviewBundleHandlerStoreFailure keeps the one failure the bundle
// cannot absorb: a store read that fails for a real reason is a 5xx, never
// a bundle a caller could mistake for the change's own.
func TestSelfReviewBundleHandlerStoreFailure(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	fs.runRecordErr = errors.New("store exploded")

	code, body := doGet(t, ts, selfReviewPath("proj", "kan-1", t.TempDir()))
	if code != http.StatusInternalServerError {
		t.Fatalf("GET bundle = %d (%s), want 500", code, body)
	}
}

// --- repository fixtures, the selfreview package's own in miniature ---

func bundleGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bundleRunGit(t, dir, "init", "-b", "main")
	bundleRunGit(t, dir, "config", "user.email", "test@example.com")
	bundleRunGit(t, dir, "config", "user.name", "Test")
	bundleCommit(t, dir, "README.md", "repo\n", "initial")
	return dir
}

func bundleArchiveBranch(t *testing.T, repo, name string, files map[string]string) {
	t.Helper()
	for file, body := range files {
		path := filepath.Join(repo, "spectre/changes/archive", name, file)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bundleRunGit(t, repo, "checkout", "-b", "chore/archive-"+name)
	bundleRunGit(t, repo, "add", "-A")
	bundleRunGit(t, repo, "commit", "-m", "chore(spectre): archive "+name)
	bundleRunGit(t, repo, "checkout", "main")
}

func bundleCommit(t *testing.T, repo, file, body, subject string) {
	t.Helper()
	path := filepath.Join(repo, file)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	bundleRunGit(t, repo, "add", "-A")
	bundleRunGit(t, repo, "commit", "-m", subject)
}

func bundleRunGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestSelfReviewBundleHandlerRequiresRepo pins repo as mandatory: without
// it the route could only misreport an archived change's sources as
// absent, so the request is refused before anything is read.
func TestSelfReviewBundleHandlerRequiresRepo(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	code, body := doGet(t, ts, selfReviewPath("proj", "kan-1", ""))
	if code != http.StatusBadRequest {
		t.Fatalf("GET bundle without repo = %d (%s), want 400", code, body)
	}
}
