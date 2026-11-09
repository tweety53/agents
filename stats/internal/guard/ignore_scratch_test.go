package guard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestIgnoreScratch pins flow-guard ignore-scratch (kan-1081 self-review,
// F2): a worktree that does not ignore .superpowers/ gets one index-only
// commit ignoring and untracking it, the scratch kept on disk; an ignored
// one is left alone; a pre-staged index is refused.
func TestIgnoreScratch(t *testing.T) {
	t.Parallel()
	shim := guardShim(t, "ignore-scratch")
	repo := func(t *testing.T) (string, func(args ...string) string) {
		t.Helper()
		dir, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		var g fxGit
		g.git(dir, "init", "-q")
		g.git(dir, "commit", "-q", "--allow-empty", "-m", "init")
		if g.err != nil {
			t.Fatal(g.err)
		}
		return dir, func(args ...string) string {
			t.Helper()
			var g fxGit
			out := g.git(dir, args...)
			if g.err != nil {
				t.Fatal(g.err)
			}
			return out
		}
	}
	run := func(t *testing.T, args ...string) guardResult {
		t.Helper()
		cmd := exec.Command(shim, args...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com", "FLOW_GUARD_CACHE_DIR="+guardCache(t))
		var o, e bytes.Buffer
		cmd.Stdout, cmd.Stderr = &o, &e
		err := cmd.Run()
		return guardResult{rc: exitCode(err), stdout: o.String(), err: e.String(), out: o.String() + e.String()}
	}

	t.Run("already ignored", func(t *testing.T) {
		t.Parallel()
		dir, git := repo(t)
		writeFile(t, filepath.Join(dir, ".gitignore"), ".superpowers/\n")
		git("add", ".gitignore")
		git("commit", "-q", "-m", "ignore")
		head := git("rev-parse", "HEAD")
		if r := run(t, dir); r.rc != 0 || r.stdout != "" || git("rev-parse", "HEAD") != head {
			t.Fatalf("rc=%d out=%q, HEAD moved=%v", r.rc, r.out, git("rev-parse", "HEAD") != head)
		}
	})
	t.Run("tracked scratch", func(t *testing.T) {
		t.Parallel()
		dir, git := repo(t)
		writeFile(t, filepath.Join(dir, ".superpowers/sdd/x/land.sh"), "echo\n")
		git("add", ".superpowers")
		git("commit", "-q", "-m", "scratch")
		head := git("rev-parse", "HEAD")
		r := run(t, dir)
		if r.rc != 0 || !strings.HasPrefix(r.stdout, "ignore-scratch: committed ") {
			t.Fatalf("rc=%d out=%q", r.rc, r.out)
		}
		if p := git("rev-parse", "HEAD^"); p != head {
			t.Errorf("HEAD^ = %s, want one new commit on %s", p, head)
		}
		if s := git("log", "-1", "--format=%s"); s != "chore: ignore .superpowers/ scratch" {
			t.Errorf("subject = %q", s)
		}
		if k := git("log", "-1", "--format=%(trailers:key=Flow-Key,valueonly)"); k != "flow.kickoff" {
			t.Errorf("Flow-Key = %q", k)
		}
		if b, _ := os.ReadFile(filepath.Join(dir, ".gitignore")); string(b) != ".superpowers/\n" {
			t.Errorf(".gitignore = %q", b)
		}
		if ls := git("ls-files", ".superpowers"); ls != "" {
			t.Errorf("still tracked: %q", ls)
		}
		if _, err := os.Stat(filepath.Join(dir, ".superpowers/sdd/x/land.sh")); err != nil {
			t.Errorf("scratch removed from disk: %v", err)
		}
	})
	t.Run("no trailing newline", func(t *testing.T) {
		t.Parallel()
		dir, git := repo(t)
		writeFile(t, filepath.Join(dir, ".gitignore"), "node_modules")
		git("add", ".gitignore")
		git("commit", "-q", "-m", "ignore")
		if r := run(t, dir); r.rc != 0 {
			t.Fatalf("rc=%d out=%q", r.rc, r.out)
		}
		if b, _ := os.ReadFile(filepath.Join(dir, ".gitignore")); string(b) != "node_modules\n.superpowers/\n" {
			t.Errorf(".gitignore = %q", b)
		}
	})
	t.Run("staged index", func(t *testing.T) {
		t.Parallel()
		dir, git := repo(t)
		writeFile(t, filepath.Join(dir, "other.txt"), "x\n")
		git("add", "other.txt")
		head := git("rev-parse", "HEAD")
		if r := run(t, dir); r.rc != 1 || !strings.Contains(r.err, "staged index") || git("rev-parse", "HEAD") != head {
			t.Fatalf("rc=%d out=%q, want 1 naming the staged index, HEAD unchanged", r.rc, r.out)
		}
	})
	t.Run("usage", func(t *testing.T) {
		t.Parallel()
		if r := run(t); r.rc != 2 {
			t.Errorf("no argument: rc=%d out=%q", r.rc, r.out)
		}
		f := filepath.Join(t.TempDir(), "file")
		writeFile(t, f, "x")
		if r := run(t, f); r.rc != 2 {
			t.Errorf("a file: rc=%d out=%q", r.rc, r.out)
		}
	})
}
