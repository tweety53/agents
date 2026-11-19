package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The close-task fixture: the change "demo" on branch spectre/demo, pushed
// to a real bare origin, with close-task run in-process against real
// commits. Only `flow` is a stub; `git` on PATH is a wrapper that logs each
// push into the same log as the stub, then execs the real git.

type ctFx struct {
	wt, peer, base, log string
	env                 Env
	origin              bool // the repositories carry an origin
}

// ctRepo is a repository on spectre/demo with one base commit and a bare
// origin of its own (none when origin is false).
func ctRepo(t *testing.T, dir string, origin bool) {
	t.Helper()
	mkdir(t, dir)
	fixtureRepo(t, dir, repoShape{Branch: "spectre/demo"})
	writeFile(t, dir+"/base.txt", "base\n")
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-q", "-m", "base")
	if origin {
		fixtureRepo(t, dir+".origin", repoShape{Bare: true})
		gitRun(t, dir, "remote", "add", "origin", dir+".origin")
	}
}

// ctFixture builds the canonical worktree, names a peer repository withPeer
// builds, and the flow stub, whose `flow tasks tick` exits tickRC.
func ctFixture(t *testing.T, origin bool, tickRC string) *ctFx {
	t.Helper()
	root := t.TempDir()
	fx := &ctFx{wt: root + "/wt", peer: root + "/peer", log: root + "/calls.log", origin: origin}
	ctRepo(t, fx.wt, origin)
	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
	bin := root + "/bin"
	// Both stubs find root from their own path, so their bodies are shared
	// across fixtures and each is assessed once (execFixtures).
	writeExec(t, bin+"/flow", "#!/bin/sh\necho \"flow $*\" >> \"${0%/bin/flow}/calls.log\"\nexit "+tickRC+"\n")
	writeExec(t, bin+"/git", "#!/bin/sh\ncase \" $* \" in *\" push \"*) echo \"git $*\" >> \"${0%/bin/git}/calls.log\";; *) exec "+fixtureGit+" \"$@\";; esac\nexec "+fixtureGit+" \"$@\"\n")
	fx.env = Env{Getenv: pathEnv(bin), Dir: fx.wt}
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
		writeFile(t, dir+"/"+files[i], files[i+1])
		paths = append(paths, files[i])
	}
	gitRun(t, dir, append([]string{"add", "-f", "--"}, paths...)...)
	gitRun(t, dir, "commit", "-q", "-m", subject, "-m", "Task-Id: "+taskID)
	return tcfGit(t, dir, "rev-parse", "HEAD")
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

func (fx *ctFx) run(args ...string) guardResult {
	return runGuard("close-task", args, fx.env)
}

func ctEqual(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("calls:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// smoke: core — a task's close: its commit checked, ticked and pushed
func TestCloseTask(t *testing.T) {
	t.Parallel()
	t.Run("clean: every task ticked, one push", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", "b\n")
		if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1, "2:"+s2); r.rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
		}
		ctEqual(t, fx.calls(t), []string{
			"flow tasks tick -C " + fx.wt + " demo 1",
			"flow tasks tick -C " + fx.wt + " demo 2",
			"git -C " + fx.wt + " push origin spectre/demo",
		})
		if got := tcfGit(t, fx.wt+".origin", "rev-parse", "spectre/demo"); got != s2 {
			t.Errorf("origin spectre/demo at %s, want %s", got, s2)
		}
	})
	t.Run("a commit off HEAD or unknown: exit 1, no tick, no push", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		gitRun(t, fx.wt, "reset", "-q", "--hard", fx.base)
		for _, sha := range []string{s1, "0000000"} {
			if r := fx.run(fx.wt, "demo", fx.base, "1:"+sha); r.rc != 1 || !strings.Contains(r.err, "RE-COMMIT") {
				t.Fatalf("%s: exit %d, want 1 and RE-COMMIT\n%s", sha, r.rc, r.out)
			}
		}
		ctEqual(t, fx.calls(t), nil)
	})
	t.Run("a bare task id resolves through its Task-Id trailer", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		fx.commit(t, fx.wt, "two", "2", "b.txt", "b\n")
		if r := fx.run(fx.wt, "demo", fx.base, "1"); r.rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
		}
		if got := tcfGit(t, fx.wt+".origin", "log", "--format=%H", "spectre/demo"); !strings.Contains(got, s1) {
			t.Errorf("origin lacks task 1's commit %s", s1)
		}
	})
	t.Run("a task argument without a commit is a usage error", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0")
		for _, a := range []string{"1", "1:", ":abc"} { // "1": no commit carries Task-Id: 1
			if r := fx.run(fx.wt, "demo", fx.base, a); r.rc != 2 {
				t.Errorf("%q: exit %d, want 2\n%s", a, r.rc, r.out)
			}
		}
		ctEqual(t, fx.calls(t), nil)
	})
	for _, c := range []struct {
		name   string
		origin bool
		tickRC string
	}{
		{"a failed tick is exit 2, the push still runs", true, "1"},
		{"a failed push is exit 2", false, "0"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fx := ctFixture(t, c.origin, c.tickRC)
			s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
			if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1); r.rc != 2 {
				t.Fatalf("exit %d, want 2\n%s", r.rc, r.out)
			}
			ctEqual(t, fx.calls(t), []string{"flow tasks tick -C " + fx.wt + " demo 1", "git -C " + fx.wt + " push origin spectre/demo"})
		})
	}
	t.Run("map form: one push per distinct worktree", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0").withPeer(t)
		s3 := fx.commit(t, fx.wt, "three", "3", "c.txt", "c\n")
		p3 := fx.commit(t, fx.peer, "three", "3", "d.txt", "d\n")
		if r := fx.run(fx.wt, "demo", fx.base, "3:"+fx.wt+"="+s3+","+fx.peer+"="+p3); r.rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
		}
		ctEqual(t, fx.calls(t), []string{
			"flow tasks tick -C " + fx.wt + " demo 3",
			"git -C " + fx.wt + " push origin spectre/demo",
			"git -C " + fx.peer + " push origin spectre/demo",
		})
		if got := tcfGit(t, filepath.Clean(fx.peer+".origin"), "rev-parse", "spectre/demo"); got != p3 {
			t.Errorf("peer origin at %s, want %s", got, p3)
		}
	})
}
