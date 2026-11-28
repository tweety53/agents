package guard

import (
	"os"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

func TestUnchanged(t *testing.T) {
	t.Parallel()
	repo := t.TempDir() + "/r"
	var g guardtest.FxGit
	g.CopyShape(repo, guardtest.RepoShape{Branch: "main", Commit: "base"})
	g.Git(repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	g.Err = os.MkdirAll(repo+"/docs", 0o755)
	g.Write(repo+"/docs/a.md", "a")
	g.Git(repo, "add", ".")
	g.Git(repo, "commit", "-q", "-m", "docs")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	env := core.Env{Getenv: verifyFullGetenv(""), Dir: repo}
	specs := []string{".", ":!docs/"}

	if r := runGuard("unchanged", specs, env); r.Rc != 0 || !strings.Contains(r.Stdout, "suite skipped") {
		t.Errorf("only an excluded path changed: rc=%d out=%q, want 0 and the skip line", r.Rc, r.Out)
	}
	g.Write(repo+"/x", "x")
	if r := runGuard("unchanged", specs, env); r.Rc != 1 {
		t.Errorf("an untracked matching file: rc=%d, want 1", r.Rc)
	}
	g.Git(repo, "add", "x")
	g.Git(repo, "commit", "-q", "-m", "x")
	if r := runGuard("unchanged", specs, env); g.Err != nil || r.Rc != 1 {
		t.Errorf("a committed matching file: rc=%d err=%v, want 1", r.Rc, g.Err)
	}
	if r := runGuard("unchanged", nil, env); r.Rc != 2 || !strings.Contains(r.Err, "usage:") {
		t.Errorf("no pathspec: rc=%d err=%q, want 2 and usage", r.Rc, r.Err)
	}
	if r := runGuard("unchanged", specs, core.Env{Getenv: verifyFullGetenv(""), Dir: t.TempDir()}); r.Rc != 2 {
		t.Errorf("not a repository: rc=%d, want 2", r.Rc)
	}
}

// verifyFullGetenv answers FLOW_VERIFY_FULL with full and everything else from os.Getenv.
func verifyFullGetenv(full string) func(string) string {
	return func(k string) string {
		if k == "FLOW_VERIFY_FULL" {
			return full
		}
		return os.Getenv(k)
	}
}

func TestUnchangedAnswersChangedUnderVerifyFull(t *testing.T) {
	t.Parallel()
	repo := t.TempDir() + "/r"
	var g guardtest.FxGit
	g.CopyShape(repo, guardtest.RepoShape{Branch: "main", Commit: "base"})
	g.Git(repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	g.Err = os.MkdirAll(repo+"/docs", 0o755)
	g.Write(repo+"/docs/a.md", "a")
	g.Git(repo, "add", ".")
	g.Git(repo, "commit", "-q", "-m", "docs")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	specs := []string{".", ":!docs/"}
	if r := runGuard("unchanged", specs, core.Env{Getenv: verifyFullGetenv("1"), Dir: repo}); r.Rc != 1 || r.Stdout != "" {
		t.Errorf("FLOW_VERIFY_FULL=1: rc=%d out=%q, want 1 and an empty stdout", r.Rc, r.Out)
	}
	if r := runGuard("unchanged", specs, core.Env{Getenv: verifyFullGetenv(""), Dir: repo}); r.Rc != 0 || !strings.Contains(r.Stdout, "suite skipped") {
		t.Errorf("FLOW_VERIFY_FULL empty: rc=%d out=%q, want 0 and the skip line", r.Rc, r.Out)
	}
	if r := runGuard("unchanged", specs, core.Env{Getenv: verifyFullGetenv("1"), Dir: t.TempDir()}); r.Rc != 1 {
		t.Errorf("FLOW_VERIFY_FULL=1, not a repository: rc=%d out=%q, want 1", r.Rc, r.Out)
	}
	if r := runGuard("unchanged", nil, core.Env{Getenv: verifyFullGetenv("1"), Dir: repo}); r.Rc != 2 || !strings.Contains(r.Err, "usage:") {
		t.Errorf("FLOW_VERIFY_FULL=1, no pathspec: rc=%d err=%q, want 2 and usage", r.Rc, r.Err)
	}
}
