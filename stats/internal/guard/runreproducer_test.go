package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// rrRepo is a committed repository with an uncommitted edit to a.txt: the
// edit a reviewer's command must see in its copy and must never revert.
func rrRepo(t *testing.T) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), "wt")
	guardtest.Mkdir(t, wt)
	guardtest.FixtureRepo(t, wt, guardtest.RepoShape{})
	guardtest.WriteFile(t, wt+"/a.txt", "a\n")
	guardtest.GitRun(t, wt, "add", "-A")
	guardtest.GitRun(t, wt, "commit", "-q", "-m", "base")
	guardtest.WriteFile(t, wt+"/a.txt", "a edited\n")
	return wt
}

func TestRunReproducer(t *testing.T) {
	t.Parallel()
	t.Run("usage is exit 2", func(t *testing.T) {
		t.Parallel()
		wt := rrRepo(t)
		for _, args := range [][]string{{wt}, {wt, "sh"}, {wt, "-t", "0", "--", "true"}, {wt, "x", "--", "true"}} {
			if r := runGuard("run-reproducer", args, core.Env{Getenv: os.Getenv}); r.Rc != 2 {
				t.Errorf("%q: exit %d, want 2\n%s", args, r.Rc, r.Out)
			}
		}
	})
	t.Run("the command runs in a copy: it sees the edit, its writes never reach the worktree, its exit status passes through", func(t *testing.T) {
		t.Parallel()
		wt := rrRepo(t)
		script := `grep -q edited a.txt || exit 9; git checkout -q -- a.txt; rm -f a.txt; echo x > new.txt; exit 3`
		r := runGuard("run-reproducer", []string{wt, "--", "sh", "-c", script}, core.Env{Getenv: os.Getenv})
		if r.Rc != 3 {
			t.Fatalf("exit %d, want the command's 3\n%s", r.Rc, r.Out)
		}
		if b, err := os.ReadFile(wt + "/a.txt"); err != nil || string(b) != "a edited\n" {
			t.Errorf("a.txt = %q, %v; want the uncommitted edit kept", b, err)
		}
		if core.IsFile(wt + "/new.txt") {
			t.Error("the command's new.txt reached the worktree")
		}
		if out := guardtest.GitC(t, wt, "worktree", "list", "--porcelain"); strings.Count(out, "worktree ") != 1 {
			t.Errorf("the copy was not removed:\n%s", out)
		}
	})
	t.Run("a command outliving the bound is killed with its children: exit 124", func(t *testing.T) {
		t.Parallel()
		wt := rrRepo(t)
		start := time.Now()
		r := runGuard("run-reproducer", []string{wt, "-t", "1", "--", "sh", "-c", "sleep 30; true"}, core.Env{Getenv: os.Getenv})
		if r.Rc != 124 || !strings.Contains(r.Err, "TIMED OUT after 1s") {
			t.Fatalf("exit %d, want 124 and the TIMED OUT line\n%s", r.Rc, r.Out)
		}
		if d := time.Since(start); d > 20*time.Second {
			t.Errorf("returned after %s: the child sleep outlived the kill", d)
		}
	})
}
