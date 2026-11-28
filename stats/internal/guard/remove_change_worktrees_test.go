package guard

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// rcwFx is a main checkout whose change branch spectre/demo is merged into
// main and pushed to a bare origin, with the change's apply worktree checked
// out beside it -- the state run 2 reaches its cleanup step in.
type rcwFx struct {
	dir, repo, wt, mergeBase string
	g                        fxGit
}

func rcwNewFx(t *testing.T) *rcwFx {
	t.Helper()
	dir := t.TempDir()
	fx := &rcwFx{dir: dir, repo: dir + "/repo", wt: dir + "/wt/demo"}
	fx.g.copyShape(dir+"/origin.git", repoShape{Branch: "main", Bare: true})
	fx.g.copyShape(fx.repo, repoShape{Branch: "main"})
	for _, kv := range [][2]string{{"user.name", "Test"}, {"user.email", "test@example.invalid"}, {"commit.gpgsign", "false"}} {
		fx.g.git(fx.repo, "config", kv[0], kv[1])
	}
	fx.g.write(fx.repo+"/.gitignore", "*.log\nbuild/\n.env\n__pycache__/\n*.pyc\n*.tsbuildinfo")
	fx.g.write(fx.repo+"/base.txt", "base")
	fx.g.git(fx.repo, "add", "-A")
	fx.g.git(fx.repo, "commit", "-qm", "base")
	fx.mergeBase = fx.g.git(fx.repo, "rev-parse", "HEAD")
	fx.g.git(fx.repo, "remote", "add", "origin", dir+"/origin.git")
	fx.g.git(fx.repo, "push", "-q", "-u", "origin", "main")
	fx.g.git(fx.repo, "remote", "set-head", "origin", "main")
	fx.g.git(fx.repo, "checkout", "-q", "-b", "spectre/demo")
	fx.g.write(fx.repo+"/work.txt", "work")
	fx.g.git(fx.repo, "add", "work.txt")
	fx.g.git(fx.repo, "commit", "-qm", "work")
	fx.g.git(fx.repo, "push", "-q", "origin", "spectre/demo")
	fx.g.git(fx.repo, "checkout", "-q", "main")
	fx.g.git(fx.repo, "merge", "-q", "--ff-only", "spectre/demo")
	fx.g.git(fx.repo, "push", "-q", "origin", "main")
	fx.g.git(fx.repo, "worktree", "add", "-q", fx.wt, "spectre/demo")
	return fx
}

// copy adds a detached mutation copy of the apply worktree, as
// implement.md makes one.
func (fx *rcwFx) copy() string {
	c := fx.wt + "-mutation-2"
	fx.g.git(fx.repo, "worktree", "add", "-q", "--detach", c, "HEAD")
	return c
}

// run runs remove-change-worktrees in-process.
func (fx *rcwFx) run(t *testing.T, tweak func(*core.Env), args ...string) (int, string, string) {
	t.Helper()
	if fx.g.err != nil {
		t.Fatal(fx.g.err)
	}
	env := core.Env{Dir: fx.dir, Getenv: os.Getenv}
	if tweak != nil {
		tweak(&env)
	}
	var out, errb bytes.Buffer
	code := removeChangeWorktrees(append([]string{fx.repo, "demo"}, args...), env, &out, &errb)
	return code, out.String(), errb.String()
}

func rcwExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (fx *rcwFx) hasRef(ref string) bool {
	return exec.Command(fixtureGit, "-C", fx.repo, "show-ref", "--verify", "--quiet", ref).Run() == nil
}

// assertUntouched is "any failed check leaves every worktree alone".
func (fx *rcwFx) assertUntouched(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range append([]string{fx.wt}, paths...) {
		if !rcwExists(p) {
			t.Errorf("%s was removed", p)
		}
	}
	if !fx.hasRef("refs/heads/spectre/demo") {
		t.Error("the local branch was deleted")
	}
	if !fx.hasRef("refs/remotes/origin/spectre/demo") {
		t.Error("the remote branch was deleted")
	}
}

func rcwLines(out, prefix string) []string {
	var got []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, prefix) {
			got = append(got, l)
		}
	}
	return got
}

func TestRemoveChangeWorktrees(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()

	t.Run("clean removal", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.write(fx.wt+"/run.log", "log")
		mkdir(t, fx.wt+"/build")
		fx.g.write(fx.wt+"/build/out.o", "obj")
		mkdir(t, fx.wt+"/__pycache__")
		fx.g.write(fx.wt+"/__pycache__/mod.pyc", "pyc")
		fx.g.write(fx.wt+"/loose.pyc", "pyc")
		fx.g.write(fx.wt+"/tsconfig.tsbuildinfo", "tsbuildinfo")
		fx.g.write(fx.wt+"/build/icon.png", "png")
		// Ignored files go with --force, never asked about.
		fx.g.write(fx.wt+"/.env", "SECRET=1")
		code, out, errb := fx.run(t, nil)
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "REMOVED: "); len(l) != 2 || !strings.HasSuffix(l[0], "/wt/demo") || l[1] != "REMOVED: spectre/demo" {
			t.Errorf("REMOVED lines %q, want the worktree then the branch", l)
		}
		if l := rcwLines(out, "REMOTE-"); len(l) != 1 || l[0] != "REMOTE-DELETED: origin/spectre/demo" {
			t.Errorf("REMOTE lines %q, want REMOTE-DELETED", l)
		}
		if rcwExists(fx.wt) || fx.hasRef("refs/heads/spectre/demo") || fx.hasRef("refs/remotes/origin/spectre/demo") {
			t.Error("the worktree, the branch or the remote branch survived")
		}
		if exec.Command(fixtureGit, "--git-dir", fx.dir+"/origin.git", "show-ref", "--verify", "--quiet", "refs/heads/spectre/demo").Run() == nil {
			t.Error("origin still carries spectre/demo")
		}
	})

	t.Run("remote ref already gone", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.git("", "--git-dir", fx.dir+"/origin.git", "branch", "-D", "spectre/demo")
		code, out, errb := fx.run(t, nil)
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "REMOTE-"); len(l) != 1 || l[0] != "REMOTE-GONE: origin/spectre/demo" {
			t.Errorf("REMOTE lines %q, want REMOTE-GONE", l)
		}
		if fx.hasRef("refs/remotes/origin/spectre/demo") {
			t.Error("the stale tracking ref was not pruned")
		}
	})

	// A forge merged the PR with a merge commit and deleted the remote
	// branch; a fetch --prune removed its tracking ref, and the main
	// checkout has not been refreshed yet — run 2 refreshes it last.
	t.Run("PR merged, remote branch pruned, main checkout behind", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.git(fx.repo, "reset", "-q", "--hard", fx.mergeBase)
		fx.g.git(fx.repo, "checkout", "-q", "--detach")
		fx.g.git(fx.repo, "merge", "-q", "--no-ff", "-m", "Merge pull request #1", "spectre/demo")
		fx.g.git(fx.repo, "push", "-q", "-f", "origin", "HEAD:main")
		fx.g.git(fx.repo, "checkout", "-q", "main")
		fx.g.git(fx.repo, "branch", "-q", "--set-upstream-to=origin/spectre/demo", "spectre/demo")
		fx.g.git("", "--git-dir", fx.dir+"/origin.git", "branch", "-D", "spectre/demo")
		fx.g.git(fx.repo, "fetch", "-q", "--prune", "origin")
		code, out, errb := fx.run(t, nil)
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "REMOVED: "); len(l) != 2 || l[1] != "REMOVED: spectre/demo" {
			t.Errorf("REMOVED lines %q, want the worktree then the branch", l)
		}
		if fx.hasRef("refs/heads/spectre/demo") {
			t.Error("the merged branch survived")
		}
	})

	// -d refuses a branch origin/<base> does not carry; the branch survives
	// with the upstream it had, or none, never origin/<base>.
	for _, tc := range []struct {
		name      string
		upstream  bool
		wantMerge string
	}{
		{"a refused -d restores the branch's upstream", true, "refs/heads/spectre/demo"},
		{"a refused -d unsets an upstream the branch never had", false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := rcwNewFx(t)
			fx.g.git(fx.repo, "reset", "-q", "--hard", fx.mergeBase)
			fx.g.git(fx.repo, "push", "-q", "-f", "origin", "HEAD:main")
			if tc.upstream {
				fx.g.git(fx.repo, "branch", "-q", "--set-upstream-to=origin/spectre/demo", "spectre/demo")
			} else {
				// Check 3 refuses an unmerged worktree with no upstream: only
				// a re-run whose worktree is already gone reaches -d this way.
				fx.g.git(fx.repo, "worktree", "remove", fx.wt)
			}
			code, out, errb := fx.run(t, nil)
			if code != 1 {
				t.Fatalf("exit %d, want 1\nstdout:\n%s\nstderr:\n%s", code, out, errb)
			}
			if l := rcwLines(out, "REFUSED: spectre/demo — git branch -d"); len(l) != 1 {
				t.Errorf("REFUSED lines %q, want the branch delete refused", l)
			}
			merge, _ := core.Capture(exec.Command(fixtureGit, "-C", fx.repo, "config", "--get", "branch.spectre/demo.merge"))
			if merge != tc.wantMerge {
				t.Errorf("branch.spectre/demo.merge = %q, want %q", merge, tc.wantMerge)
			}
		})
	}

	t.Run("a failed gate removes nothing", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.appendLine(fx.wt+"/work.txt", "dirty")
		code, out, _ := fx.run(t, nil)
		if code != 1 || len(rcwLines(out, "REFUSED: ")) != 1 || !strings.Contains(out, "check 1") {
			t.Fatalf("exit %d, stdout %q; want 1 and one check 1 REFUSED line", code, out)
		}
		if len(rcwLines(out, "REMOVED: "))+len(rcwLines(out, "REMOTE-")) != 0 {
			t.Errorf("a refusal still removed something:\n%s", out)
		}
		fx.assertUntouched(t)
	})

	t.Run("check 3 — a commit its upstream lacks", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.git(fx.wt, "branch", "-q", "--set-upstream-to=origin/spectre/demo")
		fx.g.write(fx.wt+"/late.txt", "late")
		fx.g.git(fx.wt, "add", "late.txt")
		fx.g.git(fx.wt, "commit", "-qm", "unpushed")
		code, out, _ := fx.run(t, nil)
		if code != 1 || len(rcwLines(out, "REFUSED: ")) != 1 || !strings.Contains(out, "check 3: commits not on origin/spectre/demo") {
			t.Fatalf("exit %d, stdout %q; want 1 and one check 3 REFUSED line", code, out)
		}
		fx.assertUntouched(t)
	})

	t.Run("a worktree already gone", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		if err := os.RemoveAll(fx.wt); err != nil {
			t.Fatal(err)
		}
		code, out, errb := fx.run(t, nil)
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "REMOVED: "); len(l) != 2 || !strings.HasSuffix(l[0], "/wt/demo — already gone") || l[1] != "REMOVED: spectre/demo" {
			t.Errorf("REMOVED lines %q, want the gone worktree then the branch", l)
		}
		if fx.hasRef("refs/heads/spectre/demo") {
			t.Error("the branch survived its gone worktree")
		}
	})

	t.Run("cannot answer", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		for _, args := range [][]string{{fx.repo}, {fx.repo, "../demo"}, {fx.dir, "demo"}, {fx.repo, "demo", "--force"}} {
			var out, errb bytes.Buffer
			env := core.Env{Dir: fx.dir, Getenv: os.Getenv}
			if code := removeChangeWorktrees(args, env, &out, &errb); code != 2 || out.String() != "" {
				t.Errorf("%q: exit %d, stdout %q; want 2 and nothing on stdout", args, code, out.String())
			}
		}
		fx.assertUntouched(t)
	})

	// The real shim: it asserts the sibling.
	t.Run("through the shim", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		cmd := exec.Command(guardShim(t, "remove-change-worktrees"), fx.repo, "demo")
		cmd.Dir = fx.dir
		cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
		out, err := cmd.Output()
		if err != nil || !strings.Contains(string(out), "REMOVED: spectre/demo\n") {
			t.Fatalf("shim: %v, stdout %q", err, out)
		}
	})
}

// TestRemoveChangeWorktreesDeletesACopyBranch pins that a copy made by
// throwaway-worktree create, on its own flow-copy branch, is found, removed
// and its branch deleted (kan-1090 self-review, F14).
func TestRemoveChangeWorktreesDeletesACopyBranch(t *testing.T) {
	t.Parallel()
	fx := rcwNewFx(t)
	c := fx.wt + "-mutation-2"
	var o, e bytes.Buffer
	if rc := throwawayWorktree([]string{"create", fx.wt, c}, core.Env{Getenv: os.Getenv}, &o, &e); rc != 0 {
		t.Fatalf("create: exit %d: %s", rc, e.String())
	}
	code, out, errb := fx.run(t, nil)
	if code != 0 || rcwExists(c) {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
	}
	if fx.hasRef("refs/heads/flow-copy/demo-mutation-2") {
		t.Error("the copy's flow-copy branch survived")
	}
}

// TestRemoveChangeWorktreesRemovesACopyUnasked pins that a mutation copy is
// removed without asking, before the worktree it was copied from.
func TestRemoveChangeWorktreesRemovesACopyUnasked(t *testing.T) {
	t.Parallel()
	fx := rcwNewFx(t)
	copyWT := fx.copy()
	fx.g.write(copyWT+"/scratch.txt", "scratch")
	code, out, errb := fx.run(t, nil)
	if code != 0 || rcwExists(fx.wt) || rcwExists(copyWT) {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
	}
	if l := rcwLines(out, "REMOVED: "); len(l) != 3 || !strings.HasSuffix(l[0], "-mutation-2") {
		t.Errorf("REMOVED lines %q, want the copy first, then the worktree and the branch", l)
	}
}
