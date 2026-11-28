package guard

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// wrFx is a main checkout whose project declares a `remove` workspace row
// logging the id it is given, with a parallel request's worktree on
// demo-r2 merged into main and pushed to a bare origin -- the state a
// coordinator tears a request worktree down in.
type wrFx struct {
	dir, repo, wt, log string
	g                  fxGit
}

func wrNewFx(t *testing.T) *wrFx {
	t.Helper()
	dir := t.TempDir()
	fx := &wrFx{dir: dir, repo: dir + "/repo", wt: dir + "/repo-worktrees/demo-r2", log: dir + "/removed.log"}
	fx.g.copyShape(dir+"/origin.git", repoShape{Branch: "main", Bare: true})
	fx.g.copyShape(fx.repo, repoShape{Branch: "main"})
	for _, kv := range [][2]string{{"user.name", "Test"}, {"user.email", "test@example.invalid"}, {"commit.gpgsign", "false"}} {
		fx.g.git(fx.repo, "config", kv[0], kv[1])
	}
	mkdir(t, fx.repo+"/.flow")
	fx.g.write(fx.repo+"/.flow/project.md", "# demo\n\n## workspace isolation\n\n| Command | Runs |\n|---------|------|\n| `remove` | `echo <id> >>"+fx.log+"` |")
	fx.g.write(fx.repo+"/.gitignore", "*.log")
	fx.g.git(fx.repo, "add", "-A")
	fx.g.git(fx.repo, "commit", "-qm", "base")
	fx.g.git(fx.repo, "remote", "add", "origin", dir+"/origin.git")
	fx.g.git(fx.repo, "push", "-q", "-u", "origin", "main")
	fx.g.git(fx.repo, "remote", "set-head", "origin", "main")
	fx.g.git(fx.repo, "worktree", "add", "-q", "-b", "demo-r2", fx.wt, "main")
	fx.g.write(fx.wt+"/work.txt", "work")
	fx.g.git(fx.wt, "add", "work.txt")
	fx.g.git(fx.wt, "commit", "-qm", "work")
	fx.g.git(fx.wt, "push", "-q", "origin", "demo-r2", "demo-r2:main")
	fx.g.write(fx.wt+"/run.log", "ignored")
	return fx
}

func (fx *wrFx) run(t *testing.T) (int, string, string) {
	t.Helper()
	if fx.g.err != nil {
		t.Fatal(fx.g.err)
	}
	var out, errb bytes.Buffer
	code := workspaceRemove([]string{fx.repo, "demo-r2"}, core.Env{Dir: fx.dir, Getenv: os.Getenv}, &out, &errb)
	return code, out.String(), errb.String()
}

func (fx *wrFx) remoteBranch() bool {
	return fx.g.git(fx.dir+"/origin.git", "branch", "--list", "demo-r2") != ""
}

func (fx *wrFx) assertUntouched(t *testing.T) {
	t.Helper()
	if !rcwExists(fx.wt) {
		t.Error("the worktree was removed")
	}
	if fx.g.git(fx.repo, "branch", "--list", "demo-r2") == "" {
		t.Error("the local branch was deleted")
	}
	if !fx.remoteBranch() {
		t.Error("the remote branch was deleted")
	}
	if rcwExists(fx.log) {
		t.Error("the workspace was removed")
	}
}

func TestWorkspaceRemove(t *testing.T) {
	t.Parallel()

	t.Run("landed clean worktree is removed with its workspace, branch and remote branch", func(t *testing.T) {
		t.Parallel()
		fx := wrNewFx(t)
		code, out, errs := fx.run(t)
		if code != 0 {
			t.Fatalf("exit %d, stdout %q, stderr %q", code, out, errs)
		}
		if rcwExists(fx.wt) {
			t.Error("the worktree survived")
		}
		if fx.g.git(fx.repo, "branch", "--list", "demo-r2") != "" {
			t.Error("the local branch survived")
		}
		if fx.remoteBranch() {
			t.Error("the remote branch survived")
		}
		if b, _ := os.ReadFile(fx.log); strings.TrimSpace(string(b)) != ccWorkspaceID("demo-r2") {
			t.Errorf("remove row got %q, want %q", b, ccWorkspaceID("demo-r2"))
		}
		if !strings.Contains(out, "WORKSPACE-REMOVED: "+fx.repo) {
			t.Errorf("stdout %q lacks WORKSPACE-REMOVED", out)
		}
	})

	t.Run("dirty worktree is refused and nothing removed", func(t *testing.T) {
		t.Parallel()
		fx := wrNewFx(t)
		fx.g.write(fx.wt+"/stray.txt", "stray")
		if code, out, _ := fx.run(t); code != 1 {
			t.Fatalf("exit %d, want 1; stdout %q", code, out)
		}
		fx.assertUntouched(t)
	})

	t.Run("worktree not landed on origin/<base> is refused and nothing removed", func(t *testing.T) {
		t.Parallel()
		fx := wrNewFx(t)
		fx.g.write(fx.wt+"/more.txt", "more")
		fx.g.git(fx.wt, "add", "more.txt")
		fx.g.git(fx.wt, "commit", "-qm", "more")
		if code, out, _ := fx.run(t); code != 1 {
			t.Fatalf("exit %d, want 1; stdout %q", code, out)
		}
		fx.assertUntouched(t)
	})

	t.Run("second run is a no-op", func(t *testing.T) {
		t.Parallel()
		fx := wrNewFx(t)
		if code, out, errs := fx.run(t); code != 0 {
			t.Fatalf("first run exit %d, stdout %q, stderr %q", code, out, errs)
		}
		code, out, errs := fx.run(t)
		if code != 0 {
			t.Fatalf("second run exit %d, stdout %q, stderr %q", code, out, errs)
		}
		for _, l := range []string{"REMOVED: " + fx.wt + " — already gone", "REMOTE-GONE: origin/demo-r2"} {
			if !strings.Contains(out, l) {
				t.Errorf("stdout %q lacks %q", out, l)
			}
		}
	})
}
