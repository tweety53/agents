package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// rrRepo is a committed repository with an uncommitted edit to a.txt: the
// edit a reviewer's command must see in its copy and must never revert.
func rrRepo(t *testing.T) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), "wt")
	mkdir(t, wt)
	fixtureRepo(t, wt, repoShape{})
	writeFile(t, wt+"/a.txt", "a\n")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-q", "-m", "base")
	writeFile(t, wt+"/a.txt", "a edited\n")
	return wt
}

func TestRunReproducer(t *testing.T) {
	t.Parallel()
	t.Run("usage is exit 2", func(t *testing.T) {
		t.Parallel()
		wt := rrRepo(t)
		for _, args := range [][]string{{wt}, {wt, "sh"}, {wt, "-t", "0", "--", "true"}, {wt, "x", "--", "true"}} {
			if r := runGuard("run-reproducer", args, Env{Getenv: os.Getenv}); r.rc != 2 {
				t.Errorf("%q: exit %d, want 2\n%s", args, r.rc, r.out)
			}
		}
	})
	t.Run("the command runs in a copy: it sees the edit, its writes never reach the worktree, its exit status passes through", func(t *testing.T) {
		t.Parallel()
		wt := rrRepo(t)
		script := `grep -q edited a.txt || exit 9; git checkout -q -- a.txt; rm -f a.txt; echo x > new.txt; exit 3`
		r := runGuard("run-reproducer", []string{wt, "--", "sh", "-c", script}, Env{Getenv: os.Getenv})
		if r.rc != 3 {
			t.Fatalf("exit %d, want the command's 3\n%s", r.rc, r.out)
		}
		if b, err := os.ReadFile(wt + "/a.txt"); err != nil || string(b) != "a edited\n" {
			t.Errorf("a.txt = %q, %v; want the uncommitted edit kept", b, err)
		}
		if isFile(wt + "/new.txt") {
			t.Error("the command's new.txt reached the worktree")
		}
		if out := tcfGit(t, wt, "worktree", "list", "--porcelain"); strings.Count(out, "worktree ") != 1 {
			t.Errorf("the copy was not removed:\n%s", out)
		}
	})
	t.Run("a command outliving the bound is killed with its children: exit 124", func(t *testing.T) {
		t.Parallel()
		wt := rrRepo(t)
		start := time.Now()
		r := runGuard("run-reproducer", []string{wt, "-t", "1", "--", "sh", "-c", "sleep 30; true"}, Env{Getenv: os.Getenv})
		if r.rc != 124 || !strings.Contains(r.err, "TIMED OUT after 1s") {
			t.Fatalf("exit %d, want 124 and the TIMED OUT line\n%s", r.rc, r.out)
		}
		if d := time.Since(start); d > 20*time.Second {
			t.Errorf("returned after %s: the child sleep outlived the kill", d)
		}
	})
}
