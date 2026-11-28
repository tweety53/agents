package integrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// The close-task fixture: the change "demo" on branch spectre/demo with a
// real bare origin, close-task run in-process against real commits. Only
// `flow` is a stub; `git` on PATH is a wrapper that logs each push into the
// same log as the stub, then execs the real git — so a push close-task must
// not make shows in the log.

// ctRepo is a repository on spectre/demo with one base commit and a bare
// origin of its own (none when origin is false).
func ctRepo(t *testing.T, dir string, origin bool) {
	t.Helper()
	guardtest.Mkdir(t, dir)
	guardtest.FixtureRepo(t, dir, guardtest.RepoShape{Branch: "spectre/demo"})
	guardtest.WriteFile(t, dir+"/base.txt", "base\n")
	guardtest.GitRun(t, dir, "add", "-A")
	guardtest.GitRun(t, dir, "commit", "-q", "-m", "base")
	if origin {
		guardtest.FixtureRepo(t, dir+".origin", guardtest.RepoShape{Bare: true})
		guardtest.GitRun(t, dir, "remote", "add", "origin", dir+".origin")
	}
}

// smoke: core — a task's close: its commit checked and ticked, never pushed
// (the branch is pushed once at flow.implement's end)
func TestCloseTask(t *testing.T) {
	t.Parallel()
	t.Run("clean: every task ticked, nothing pushed", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", "b\n")
		if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1, "2:"+s2); r.Rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.Rc, r.Out)
		}
		ctEqual(t, fx.calls(t), []string{
			"flow tasks tick -C " + fx.wt + " demo 1",
			"flow tasks tick -C " + fx.wt + " demo 2",
		})
		if got := guardtest.GitC(t, fx.wt+".origin", "for-each-ref", "refs/heads/spectre/demo"); got != "" {
			t.Errorf("origin carries spectre/demo (%s), want it unpushed", got)
		}
	})
	t.Run("a commit off HEAD or unknown: exit 1, no tick, no push", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		guardtest.GitRun(t, fx.wt, "reset", "-q", "--hard", fx.base)
		for _, sha := range []string{s1, "0000000"} {
			if r := fx.run(fx.wt, "demo", fx.base, "1:"+sha); r.Rc != 1 || !strings.Contains(r.Err, "RE-COMMIT") {
				t.Fatalf("%s: exit %d, want 1 and RE-COMMIT\n%s", sha, r.Rc, r.Out)
			}
		}
		ctEqual(t, fx.calls(t), nil)
	})
	t.Run("a bare task id resolves through its Task-Id trailer", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0")
		fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		fx.commit(t, fx.wt, "two", "2", "b.txt", "b\n")
		if r := fx.run(fx.wt, "demo", fx.base, "1"); r.Rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.Rc, r.Out)
		}
		ctEqual(t, fx.calls(t), []string{"flow tasks tick -C " + fx.wt + " demo 1"})
	})
	t.Run("a task argument without a commit is a usage error", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0")
		for _, a := range []string{"1", "1:", ":abc"} { // "1": no commit carries Task-Id: 1
			if r := fx.run(fx.wt, "demo", fx.base, a); r.Rc != 2 {
				t.Errorf("%q: exit %d, want 2\n%s", a, r.Rc, r.Out)
			}
		}
		ctEqual(t, fx.calls(t), nil)
	})
	t.Run("a failed tick is exit 2", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "1")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1); r.Rc != 2 {
			t.Fatalf("exit %d, want 2\n%s", r.Rc, r.Out)
		}
		ctEqual(t, fx.calls(t), []string{"flow tasks tick -C " + fx.wt + " demo 1"})
	})
	t.Run("map form — no worktree pushed", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0").withPeer(t)
		s3 := fx.commit(t, fx.wt, "three", "3", "c.txt", "c\n")
		p3 := fx.commit(t, fx.peer, "three", "3", "d.txt", "d\n")
		if r := fx.run(fx.wt, "demo", fx.base, "3:"+fx.wt+"="+s3+","+fx.peer+"="+p3); r.Rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.Rc, r.Out)
		}
		ctEqual(t, fx.calls(t), []string{
			"flow tasks tick -C " + fx.wt + " demo 3",
		})
		if got := guardtest.GitC(t, filepath.Clean(fx.peer+".origin"), "for-each-ref", "refs/heads/spectre/demo"); got != "" {
			t.Errorf("peer origin carries spectre/demo (%s), want it unpushed", got)
		}
	})
}

type ctFx struct {
	wt, peer, base, log string
	env                 core.Env
	origin              bool // the repositories carry an origin
}

// ctFixture builds the canonical worktree, names a peer repository withPeer
// builds, and the flow stub, whose `flow tasks tick` exits tickRC.
func ctFixture(t *testing.T, origin bool, tickRC string) *ctFx {
	t.Helper()
	root := t.TempDir()
	fx := &ctFx{wt: root + "/wt", peer: root + "/peer", log: root + "/calls.log", origin: origin}
	ctRepo(t, fx.wt, origin)
	fx.base = guardtest.GitC(t, fx.wt, "rev-parse", "HEAD")
	bin := root + "/bin"
	// Both stubs find root from their own path, so their bodies are shared
	// across fixtures and each is assessed once (ExecFixtures).
	guardtest.WriteExec(t, bin+"/flow", "#!/bin/sh\necho \"flow $*\" >> \"${0%/bin/flow}/calls.log\"\nexit "+tickRC+"\n")
	guardtest.WriteExec(t, bin+"/git", "#!/bin/sh\ncase \" $* \" in *\" push \"*) echo \"git $*\" >> \"${0%/bin/git}/calls.log\";; *) exec "+guardtest.FixtureGit+" \"$@\";; esac\nexec "+guardtest.FixtureGit+" \"$@\"\n")
	fx.env = core.Env{Getenv: guardtest.PathEnv(bin), Dir: fx.wt}
	return fx
}

// withPeer builds the peer repository at fx.peer, as ctRepo built the
// canonical one: only the tests that read fx.peer pay for it.
func (fx *ctFx) withPeer(t *testing.T) *ctFx {
	t.Helper()
	ctRepo(t, fx.peer, fx.origin)
	return fx
}

// commit writes each path=body pair in dir and commits exactly those paths
// under subject with a Task-Id trailer, returning the sha.
func (fx *ctFx) commit(t *testing.T, dir, subject, taskID string, files ...string) string {
	t.Helper()
	var paths []string
	for i := 0; i < len(files); i += 2 {
		guardtest.WriteFile(t, dir+"/"+files[i], files[i+1])
		paths = append(paths, files[i])
	}
	guardtest.GitRun(t, dir, append([]string{"add", "-f", "--"}, paths...)...)
	guardtest.GitRun(t, dir, "commit", "-q", "-m", subject, "-m", "Task-Id: "+taskID)
	return guardtest.GitC(t, dir, "rev-parse", "HEAD")
}

func (fx *ctFx) calls(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(fx.log)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimRight(string(b), "\n"), "\n")
}

func (fx *ctFx) run(args ...string) guardtest.GuardResult {
	return runGuard("close-task", args, fx.env)
}

func ctEqual(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("calls:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
