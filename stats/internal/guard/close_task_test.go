package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The close-task fixture: the change "demo" on branch spectre/demo, pushed
// to a real bare origin, with the real guards run in-process against real
// commits. Only `flow` is a stub; `git` on PATH is a wrapper that logs each
// push into the same log as the stub, then execs the real git.

var ctPlan = bt(`- [ ] 1. One

**Files:** ¤a.txt¤
**Tests:** none
**Commit:** one
**Build:** green

- [ ] 2. Two

**Files:** ¤b.txt¤
**Tests:** none
**Commit:** two
**Build:** green

- [ ] 3. Spanning

**Files:** ¤c.txt¤, ¤d.txt¤
**Tests:** none
**Commit:** three
**Build:** green
`)

type ctFx struct {
	wt, peer, base, log, decisions string
	env                            Env
}

// ctRepo is a repository on spectre/demo with one base commit and a bare
// origin of its own (none when origin is false).
func ctRepo(t *testing.T, dir string, origin bool) {
	t.Helper()
	mkdir(t, dir)
	gitRun(t, dir, "init", "-q", "-b", "spectre/demo")
	writeFile(t, dir+"/base.txt", "base\n")
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-q", "-m", "base")
	if origin {
		gitRun(t, "", "init", "-q", "--bare", dir+".origin")
		gitRun(t, dir, "remote", "add", "origin", dir+".origin")
	}
}

// ctFixture builds the canonical worktree (plan on disk, never committed),
// a peer repository, and the flow stub: tickRC and recordRC are the exit
// codes of `flow tasks tick` and `flow record …`; `flow record decisions`
// prints fx.decisions when a test wrote it, nothing otherwise.
func ctFixture(t *testing.T, origin bool, tickRC, recordRC string) *ctFx {
	t.Helper()
	root := t.TempDir()
	fx := &ctFx{wt: root + "/wt", peer: root + "/peer", log: root + "/calls.log", decisions: root + "/decisions.json"}
	ctRepo(t, fx.wt, origin)
	ctRepo(t, fx.peer, origin)
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", ctPlan)
	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
	bin := root + "/bin"
	writeExec(t, bin+"/flow", "#!/bin/sh\necho \"flow $*\" >> "+fx.log+"\ncase \"$1 $2\" in\n\"tasks tick\") exit "+tickRC+";;\n\"record decisions\") cat "+fx.decisions+" 2>/dev/null; exit "+recordRC+";;\nrecord*) exit "+recordRC+";;\nesac\n")
	writeExec(t, bin+"/git", "#!/bin/sh\ncase \" $* \" in *\" push \"*) echo \"git $*\" >> "+fx.log+";; esac\nexec "+fixtureGit+" \"$@\"\n")
	fx.env = Env{Getenv: pathEnv(bin), Dir: fx.wt}
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

func TestCloseTaskOrder(t *testing.T) {
	t.Parallel()
	t.Run("clean: gate, tick of QUIET tasks only, one push", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", rgLines(45))
		r := fx.run("-end-key", "task-1-implementer", "-session-token", "mf-t", "-end-commit", s2,
			fx.wt, "demo", fx.base, "1:"+s1, "2:"+s2)
		if r.rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
		}
		for _, want := range []string{"QUIET: task 1 — 1 changed lines, every path declared\n", "FIRE: task 2 — 45 changed lines (more than 40)\n"} {
			if !strings.Contains(r.stdout, want) {
				t.Errorf("stdout lacks %q:\n%s", want, r.stdout)
			}
		}
		ctEqual(t, fx.calls(t), []string{
			"flow record dispatch end -change demo -key task-1-implementer -session-token mf-t -commit " + s2 + " -outcome completed",
			"flow record decisions -change demo -C " + fx.wt,
			"flow tasks tick -C " + fx.wt + " demo 1",
			"git -C " + fx.wt + " push origin spectre/demo",
		})
		if got := tcfGit(t, fx.wt+".origin", "rev-parse", "spectre/demo"); got != s2 {
			t.Errorf("origin spectre/demo at %s, want %s", got, s2)
		}
	})
	t.Run("fields refusal: exit 1, no tick, no push", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		s1 := fx.commit(t, fx.wt, "wrong subject", "1", "a.txt", "a\n")
		r := fx.run(fx.wt, "demo", fx.base, "1:"+s1)
		if r.rc != 1 {
			t.Fatalf("exit %d, want 1\n%s", r.rc, r.out)
		}
		ctEqual(t, fx.calls(t), nil)
	})
	t.Run("planning-paths refusal: exit 1, no tick, no push", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		fx.commit(t, fx.wt, "swept", "9", "spectre/changes/demo/notes.md", "n\n")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		r := fx.run(fx.wt, "demo", fx.base, "1:"+s1)
		if r.rc != 1 {
			t.Fatalf("exit %d, want 1\n%s", r.rc, r.out)
		}
		ctEqual(t, fx.calls(t), nil)
	})
	t.Run("undeclared path from a refusal fires the gate", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		r := fx.run("-undeclared", "1=a.txt", fx.wt, "demo", fx.base, "1:"+s1)
		if r.rc != 0 || !strings.Contains(r.stdout, "FIRE: task 1 — undeclared paths: a.txt\n") {
			t.Fatalf("exit %d, want 0 and a fired gate\n%s", r.rc, r.out)
		}
		ctEqual(t, fx.calls(t), []string{"flow record decisions -change demo -C " + fx.wt, "git -C " + fx.wt + " push origin spectre/demo"})
	})
	// KAN-934: no gated reviewer runs below big, so a fired task ticks with
	// the quiet ones and the panel reviews it.
	t.Run("class small: a fired task ticks too", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		writeFile(t, fx.decisions, `[{"decision":{"class":"small"}},{"decision":{"class":"big"}}]`)
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", rgLines(45))
		r := fx.run(fx.wt, "demo", fx.base, "1:"+s1, "2:"+s2)
		if r.rc != 0 || !strings.Contains(r.stdout, "FIRE: task 2 — 45 changed lines (more than 40)\n") {
			t.Fatalf("exit %d, want 0 and a fired gate\n%s", r.rc, r.out)
		}
		ctEqual(t, fx.calls(t), []string{
			"flow record decisions -change demo -C " + fx.wt,
			"flow tasks tick -C " + fx.wt + " demo 1",
			"flow tasks tick -C " + fx.wt + " demo 2",
			"git -C " + fx.wt + " push origin spectre/demo",
		})
	})
	t.Run("decisions read fails: a fired task's tick waits and says so", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "1")
		writeFile(t, fx.decisions, `[{"decision":{"class":"small"}}]`)
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", rgLines(45))
		r := fx.run(fx.wt, "demo", fx.base, "1:"+s1, "2:"+s2)
		if r.rc != 0 || !strings.Contains(r.stdout, "close-task: task 2 tick waits — decision class unreadable\n") {
			t.Fatalf("exit %d, want 0 and the waiting-tick line\n%s", r.rc, r.out)
		}
		ctEqual(t, fx.calls(t), []string{
			"flow record decisions -change demo -C " + fx.wt,
			"flow tasks tick -C " + fx.wt + " demo 1",
			"git -C " + fx.wt + " push origin spectre/demo",
		})
	})
	t.Run("class big: a fired task's tick waits", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		writeFile(t, fx.decisions, `[{"decision":{"class":"big"}}]`)
		s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", rgLines(45))
		if r := fx.run(fx.wt, "demo", fx.base, "2:"+s2); r.rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
		}
		ctEqual(t, fx.calls(t), []string{"flow record decisions -change demo -C " + fx.wt, "git -C " + fx.wt + " push origin spectre/demo"})
	})
}

func TestCloseTaskExitContract(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name             string
		origin           bool
		tickRC, recordRC string
		args             func(fx *ctFx, s1 string) []string
		rc               int
		calls            func(fx *ctFx) []string
	}{
		{"fields could not judge", true, "0", "0",
			func(fx *ctFx, s1 string) []string { return []string{fx.wt, "demo", fx.base, "1:0000000"} },
			2, func(*ctFx) []string { return nil }},
		{"planning-paths could not judge", true, "0", "0",
			func(fx *ctFx, s1 string) []string { return []string{fx.wt, "demo", "no-such-base", "1:" + s1} },
			2, func(*ctFx) []string { return nil }},
		{"failed tick", true, "1", "0",
			func(fx *ctFx, s1 string) []string { return []string{fx.wt, "demo", fx.base, "1:" + s1} },
			2, func(fx *ctFx) []string {
				return []string{"flow tasks tick -C " + fx.wt + " demo 1", "git -C " + fx.wt + " push origin spectre/demo"}
			}},
		{"failed push", false, "0", "0",
			func(fx *ctFx, s1 string) []string { return []string{fx.wt, "demo", fx.base, "1:" + s1} },
			2, func(fx *ctFx) []string {
				return []string{"flow tasks tick -C " + fx.wt + " demo 1", "git -C " + fx.wt + " push origin spectre/demo"}
			}},
		{"a failed record never blocks", true, "0", "1",
			func(fx *ctFx, s1 string) []string {
				return []string{"-end-key", "k", "-session-token", "mf-t", fx.wt, "demo", fx.base, "1:" + s1}
			},
			0, func(fx *ctFx) []string {
				return []string{"flow record dispatch end -change demo -key k -session-token mf-t -outcome completed",
					"flow tasks tick -C " + fx.wt + " demo 1", "git -C " + fx.wt + " push origin spectre/demo"}
			}},
		{"task argument without a sha", true, "0", "0",
			func(fx *ctFx, s1 string) []string { return []string{fx.wt, "demo", fx.base, "1"} },
			2, func(*ctFx) []string { return nil }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fx := ctFixture(t, c.origin, c.tickRC, c.recordRC)
			s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
			r := fx.run(c.args(fx, s1)...)
			if r.rc != c.rc {
				t.Fatalf("exit %d, want %d\n%s", r.rc, c.rc, r.out)
			}
			ctEqual(t, fx.calls(t), c.calls(fx))
		})
	}
	t.Run("map form: one push per distinct worktree", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		s3 := fx.commit(t, fx.wt, "three", "3", "c.txt", "c\n")
		p3 := fx.commit(t, fx.peer, "three", "3", "d.txt", "d\n")
		r := fx.run(fx.wt, "demo", fx.base, "3:"+fx.wt+"="+s3+","+fx.peer+"="+p3)
		if r.rc != 0 || !strings.Contains(r.stdout, "QUIET: task 3 — 2 changed lines, every path declared\n") {
			t.Fatalf("exit %d, want 0 and a quiet gate\n%s", r.rc, r.out)
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
