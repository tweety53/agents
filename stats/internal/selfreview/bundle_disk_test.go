package selfreview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
)

// changeWorktree is a /flow change's worktree in miniature: a main checkout
// with an origin/main base, and a linked worktree on spectre/<name>, its
// change directory gitignored.
func changeWorktree(t *testing.T, name string) (main, wt string) {
	t.Helper()
	main = gitRepo(t)
	commitAll(t, main, ".gitignore", "spectre/changes/\n", "chore: ignore the change directory")
	runGit(t, main, "update-ref", "refs/remotes/origin/main", "main")
	runGit(t, main, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	wt = filepath.Join(t.TempDir(), "wt")
	runGit(t, main, "worktree", "add", "-q", "-b", "spectre/"+name, wt)
	return main, wt
}

// TestBundleReadsChangeFilesFromTheWorktree pins the disk read: every file
// of an uncommitted, gitignored change directory is served under its
// repository-relative label.
func TestBundleReadsChangeFilesFromTheWorktree(t *testing.T) {
	_, wt := changeWorktree(t, "demo")
	files := map[string]string{
		"tasks.md":               "# demo tasks\n",
		"design.md":              "# demo design\n",
		"narrative.md":           "# demo narrative\n",
		"visual-verification.md": "# demo visual\n",
		"live-verification.md":   "# demo live\n",
	}
	for file, body := range files {
		p := filepath.Join(wt, "spectre/changes/demo", file)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if out, err := (ExecRunner{}).Output(wt, "status", "--porcelain"); err != nil || len(out) != 0 {
		t.Fatalf("the change directory must be ignored, status = %q, %v", out, err)
	}

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{wt}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	for file, body := range files {
		if !strings.Contains(bundle, "## spectre/changes/demo/"+file+"\n\n"+body) {
			t.Errorf("bundle missing %s served from disk:\n%s", file, bundle)
		}
	}
	if strings.Contains(bundle, wt) {
		t.Errorf("bundle leaks the worktree path:\n%s", bundle)
	}
}

// TestBundleFinishRangeIsMergeBaseToHead pins the git-log source: the
// worktree's first-parent merge-base..HEAD, the base's own commit out of it,
// every commit of the change branch in, followed by the merge-order peer
// spectre/peers names relative to the main checkout, not the worktree.
func TestBundleFinishRangeIsMergeBaseToHead(t *testing.T) {
	peer := gitRepo(t)
	runGit(t, peer, "update-ref", "refs/remotes/origin/main", "main")
	runGit(t, peer, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	runGit(t, peer, "checkout", "-q", "-b", "spectre/demo")
	peerSHA := commitAll(t, peer, "app.go", "package main\n", "feat(app): the satellite's half")
	main, wt := changeWorktree(t, "demo")
	rel, err := filepath.Rel(main, peer)
	if err != nil {
		t.Fatal(err)
	}
	baseSHA := strings.TrimSpace(string(must(ExecRunner{}.Output(main, "rev-parse", "HEAD"))))
	first := commitAll(t, wt, "spectre/peers", "front "+rel+"\n", "chore: declare the peer")
	second := commitAll(t, wt, "b.go", "package a\n", "fix(demo): second")
	if err := os.MkdirAll(filepath.Join(wt, "spectre/changes/demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, "spectre/changes/demo/link.md"), []byte("## Merge order\n\n1. `.`\n2. `front`\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{wt}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	gitLog := bundle[strings.Index(bundle, "## git log --stat"):]
	for _, sha := range []string{first, second} {
		if !strings.Contains(gitLog, "commit "+sha) {
			t.Errorf("git log missing the change's commit %s:\n%s", sha, gitLog)
		}
	}
	if strings.Contains(gitLog, "commit "+baseSHA) {
		t.Errorf("git log carries the merge base %s:\n%s", baseSHA, gitLog)
	}
	if !strings.Contains(gitLog, "### front") || !strings.Contains(gitLog, "commit "+peerSHA) {
		t.Errorf("git log missing the peer's commit %s under ### front:\n%s", peerSHA, gitLog)
	}
}

func must(b []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return b
}
