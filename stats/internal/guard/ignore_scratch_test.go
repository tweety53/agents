package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// TestIgnoreScratch pins flow-guard ignore-scratch (kan-1081 self-review,
// F2): a worktree that does not ignore .superpowers/ gets one index-only
// commit ignoring and untracking it, the scratch kept on disk; an ignored
// one is left alone; a pre-staged index is refused.
func TestIgnoreScratch(t *testing.T) {
	smoke.Heavy(t)
	for k, v := range map[string]string{"GIT_AUTHOR_NAME": "test", "GIT_AUTHOR_EMAIL": "test@example.com", "GIT_COMMITTER_NAME": "test", "GIT_COMMITTER_EMAIL": "test@example.com"} {
		t.Setenv(k, v)
	}
	repo := func(t *testing.T) (string, func(args ...string) string) {
		t.Helper()
		dir, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		var g guardtest.FxGit
		g.CopyShape(dir, guardtest.RepoShape{Commit: "init"})
		if g.Err != nil {
			t.Fatal(g.Err)
		}
		return dir, func(args ...string) string {
			t.Helper()
			var g guardtest.FxGit
			out := g.Git(dir, args...)
			if g.Err != nil {
				t.Fatal(g.Err)
			}
			return out
		}
	}
	run := func(t *testing.T, args ...string) guardtest.GuardResult {
		t.Helper()
		var o, e bytes.Buffer
		rc := ignoreScratch(args, core.Env{Getenv: os.Getenv}, &o, &e)
		return guardtest.GuardResult{Rc: rc, Stdout: o.String(), Err: e.String(), Out: o.String() + e.String()}
	}

	t.Run("already ignored", func(t *testing.T) {
		t.Parallel()
		dir, git := repo(t)
		guardtest.WriteFile(t, filepath.Join(dir, ".gitignore"), ".superpowers/\n")
		git("add", ".gitignore")
		git("commit", "-q", "-m", "ignore")
		head := git("rev-parse", "HEAD")
		if r := run(t, dir); r.Rc != 0 || r.Stdout != "" || git("rev-parse", "HEAD") != head {
			t.Fatalf("rc=%d out=%q, HEAD moved=%v", r.Rc, r.Out, git("rev-parse", "HEAD") != head)
		}
	})
	t.Run("tracked scratch", func(t *testing.T) {
		t.Parallel()
		dir, git := repo(t)
		guardtest.WriteFile(t, filepath.Join(dir, ".superpowers/sdd/x/land.sh"), "echo\n")
		git("add", ".superpowers")
		git("commit", "-q", "-m", "scratch")
		head := git("rev-parse", "HEAD")
		r := run(t, dir)
		if r.Rc != 0 || !strings.HasPrefix(r.Stdout, "ignore-scratch: committed ") {
			t.Fatalf("rc=%d out=%q", r.Rc, r.Out)
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
		guardtest.WriteFile(t, filepath.Join(dir, ".gitignore"), "node_modules")
		git("add", ".gitignore")
		git("commit", "-q", "-m", "ignore")
		if r := run(t, dir); r.Rc != 0 {
			t.Fatalf("rc=%d out=%q", r.Rc, r.Out)
		}
		if b, _ := os.ReadFile(filepath.Join(dir, ".gitignore")); string(b) != "node_modules\n.superpowers/\n" {
			t.Errorf(".gitignore = %q", b)
		}
	})
	t.Run("staged index", func(t *testing.T) {
		t.Parallel()
		dir, git := repo(t)
		guardtest.WriteFile(t, filepath.Join(dir, "other.txt"), "x\n")
		git("add", "other.txt")
		head := git("rev-parse", "HEAD")
		if r := run(t, dir); r.Rc != 1 || !strings.Contains(r.Err, "staged index") || git("rev-parse", "HEAD") != head {
			t.Fatalf("rc=%d out=%q, want 1 naming the staged index, HEAD unchanged", r.Rc, r.Out)
		}
	})
	t.Run("usage", func(t *testing.T) {
		t.Parallel()
		if r := run(t); r.Rc != 2 {
			t.Errorf("no argument: rc=%d out=%q", r.Rc, r.Out)
		}
		f := filepath.Join(t.TempDir(), "file")
		guardtest.WriteFile(t, f, "x")
		if r := run(t, f); r.Rc != 2 {
			t.Errorf("a file: rc=%d out=%q", r.Rc, r.Out)
		}
	})
}
