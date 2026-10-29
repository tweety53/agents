package guard

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"os/exec"
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
	wt, peer, base, log, decisions, dispatches, state string
	env                                               Env
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
// and `flow record dispatches` print fx.decisions and fx.dispatches when a
// test wrote them, nothing otherwise; `flow state get` prints fx.state the
// same way and is never logged.
func ctFixture(t *testing.T, origin bool, tickRC, recordRC string) *ctFx {
	t.Helper()
	root := t.TempDir()
	fx := &ctFx{wt: root + "/wt", peer: root + "/peer", log: root + "/calls.log", decisions: root + "/decisions.json", dispatches: root + "/dispatches.json", state: root + "/state.json"}
	ctRepo(t, fx.wt, origin)
	ctRepo(t, fx.peer, origin)
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", ctPlan)
	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
	bin := root + "/bin"
	writeExec(t, bin+"/flow", "#!/bin/sh\n[ \"$1 $2\" = \"state get\" ] && { cat "+fx.state+" 2>/dev/null; exit 0; }\necho \"flow $*\" >> "+fx.log+"\ncase \"$1 $2\" in\n\"tasks tick\") exit "+tickRC+";;\n\"record decisions\") cat "+fx.decisions+" 2>/dev/null; exit "+recordRC+";;\n\"record dispatches\") cat "+fx.dispatches+" 2>/dev/null; exit "+recordRC+";;\nrecord*) exit "+recordRC+";;\nesac\n")
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
			"flow record dispatch end -change demo -C " + fx.wt + " -key task-1-implementer -session-token mf-t -commit " + s2 + " -outcome completed",
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
	t.Run("untasked unpushed commit: exit 1, no tick, no push", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		gitRun(t, fx.wt, "push", "-q", "-u", "origin", "spectre/demo")
		writeFile(t, fx.wt+"/m.txt", "m\n")
		gitRun(t, fx.wt, "add", "--", "m.txt")
		gitRun(t, fx.wt, "commit", "-q", "-m", "fix(app): review Minors")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		r := fx.run(fx.wt, "demo", fx.base, "1:"+s1)
		if r.rc != 1 || !strings.Contains(r.stdout, "UNTASKED: 1 commit(s) carry no Task-Id: or Flow-Key: trailer") {
			t.Fatalf("exit %d, want 1 and the UNTASKED verdict\n%s", r.rc, r.out)
		}
		ctEqual(t, fx.calls(t), nil)
	})
	t.Run("a rebase onto a moved base: the base's commits are not judged", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		gitRun(t, fx.wt, "push", "-q", "-u", "origin", "spectre/demo")
		writeFile(t, fx.wt+"/m.txt", "m\n")
		gitRun(t, fx.wt, "add", "--", "m.txt")
		gitRun(t, fx.wt, "commit", "-q", "-m", "fix(build): someone else's work")
		gitRun(t, fx.wt, "push", "-q", "origin", "HEAD:refs/heads/main")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1); r.rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
		}
	})
	t.Run("a Flow-Key unpushed commit pushes", func(t *testing.T) {
		t.Parallel()
		fx := ctFixture(t, true, "0", "0")
		gitRun(t, fx.wt, "push", "-q", "-u", "origin", "spectre/demo")
		writeFile(t, fx.wt+"/m.txt", "m\n")
		gitRun(t, fx.wt, "add", "--", "m.txt")
		gitRun(t, fx.wt, "commit", "-q", "-m", "fix(app): review Minors", "-m", "Flow-Key: panel-fix-1")
		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
		if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1); r.rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
		}
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
			3, func(fx *ctFx) []string {
				return []string{"flow record dispatch end -change demo -C " + fx.wt + " -key k -session-token mf-t -outcome completed",
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

// The Task-Id trailer, not the sha close-task is handed, is how a task's
// commit is found again, so a commit without one naming the task — or the
// green partner a red task folds into — is refused before any tick or push.
func TestCloseTaskTaskIdTrailer(t *testing.T) {
	t.Parallel()
	fold := ctPlan + bt(`
- [ ] 4. Red half

**Files:** ¤e.txt¤
**Tests:** none
**Commit:** four
**Build:** red

**Squash-with:** Task 5

- [ ] 5. Green half

**Files:** ¤f.txt¤
**Tests:** none
**Commit:** five
**Build:** green
`)
	cases := []struct {
		name, task, trailer, subject string
		files                        []string
		rc                           int
		untrailered                  bool // a second repository's commit of the task with no trailer, passed as a commit map
	}{
		{name: "trailer names the task", task: "1", trailer: "Task-Id: 1", subject: "one", files: []string{"a.txt", "a\n"}, rc: 0},
		{name: "no trailer", task: "1", subject: "one", files: []string{"a.txt", "a\n"}, rc: 1},
		{name: "trailer names another task", task: "1", trailer: "Task-Id: 2", subject: "one", files: []string{"a.txt", "a\n"}, rc: 1},
		{name: "red task closed on its partner's commit", task: "4", trailer: "Task-Id: 5", subject: "five", files: []string{"e.txt", "e\n", "f.txt", "f\n"}, rc: 0},
		{name: "trailer names several tasks", task: "1", trailer: "Task-Id: 1+2", subject: "one", files: []string{"a.txt", "a\n"}, rc: 0},
		{name: "commit map pair without a trailer", task: "1", trailer: "Task-Id: 1", subject: "one", files: []string{"a.txt", "a\n"}, rc: 1, untrailered: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fx := ctFixture(t, true, "0", "0")
			writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", fold)
			var paths []string
			for i := 0; i < len(c.files); i += 2 {
				writeFile(t, fx.wt+"/"+c.files[i], c.files[i+1])
				paths = append(paths, c.files[i])
			}
			gitRun(t, fx.wt, append([]string{"add", "--"}, paths...)...)
			msg := []string{"commit", "-q", "-m", c.subject}
			if c.trailer != "" {
				msg = append(msg, "-m", c.trailer)
			}
			gitRun(t, fx.wt, msg...)
			sha := tcfGit(t, fx.wt, "rev-parse", "HEAD")
			commit := sha
			if c.untrailered {
				wt2 := t.TempDir() + "/wt2"
				gitRun(t, fx.wt, "worktree", "add", "-q", "-b", "other", wt2, fx.base)
				writeFile(t, wt2+"/"+c.files[0], "again\n")
				gitRun(t, wt2, "add", "--", c.files[0])
				gitRun(t, wt2, "commit", "-q", "-m", c.subject)
				commit = fx.wt + "=" + sha + "," + wt2 + "=" + tcfGit(t, wt2, "rev-parse", "HEAD")
			}
			r := fx.run(fx.wt, "demo", fx.base, c.task+":"+commit)
			if r.rc != c.rc {
				t.Fatalf("exit %d, want %d\n%s", r.rc, c.rc, r.out)
			}
			if c.rc == 1 && (len(fx.calls(t)) != 0 || !strings.Contains(r.stdout, "carries no Task-Id: trailer naming task "+c.task)) {
				t.Errorf("want a trailer refusal and no tick or push\n%s\ncalls: %v", r.out, fx.calls(t))
			}
		})
	}
}

// Signal four at close time: closing an appended task no dispatch row names
// is refused, exit 1, nothing ticked or pushed (kan-982 self-review); a
// plan-time task never reads the store at all.
func TestCloseTaskRefusesAnUndispatchedAppendedTask(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, rows string
		warn       bool
	}{
		{"an appended task no row names is refused", `[{"key":"task-1-implementer","taskId":""}]`, true},
		{"an appended task its key names is silent", `[{"key":"task-2-implementer","taskId":"","commitSha":"abc"}]`, false},
		{"an appended task its taskId names is silent", `[{"key":"impl","taskId":"2","commitSha":"abc"}]`, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fx := ctFixture(t, true, "0", "0")
			writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "> **Tasks appended:** 2\n\n"+ctPlan)
			writeFile(t, fx.dispatches, c.rows)
			s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", "b\n")
			r := fx.run(fx.wt, "demo", fx.base, "2:"+s2)
			const want = "close-task: appended task 2 has no dispatch row naming it"
			if got := strings.Contains(r.err, want); got != c.warn {
				t.Errorf("refusal present %v, want %v\n%s", got, c.warn, r.out)
			}
			if c.warn {
				if r.rc != 1 || !strings.Contains(r.err, "close-task: RECORD —") {
					t.Fatalf("exit %d, want 1 and the RECORD line\n%s", r.rc, r.out)
				}
				ctEqual(t, fx.calls(t), []string{"flow record dispatches -change demo -C " + fx.wt})
				return
			}
			if r.rc != 0 {
				t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
			}
			ctEqual(t, fx.calls(t), []string{
				"flow record dispatches -change demo -C " + fx.wt,
				"flow tasks tick -C " + fx.wt + " demo 2",
				"git -C " + fx.wt + " push origin spectre/demo",
			})
		})
	}
}

// TestCloseTaskExpandsAnAbbreviatedEndCommit pins that an abbreviated
// -end-commit reaches `flow record dispatch end` as the full sha, which is
// all the store accepts.
func TestCloseTaskExpandsAnAbbreviatedEndCommit(t *testing.T) {
	t.Parallel()
	fx := ctFixture(t, true, "0", "0")
	s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
	r := fx.run("-end-key", "task-1-implementer", "-session-token", "mf-t", "-end-commit", s1[:8],
		fx.wt, "demo", fx.base, "1:"+s1)
	if r.rc != 0 {
		t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
	}
	want := "flow record dispatch end -change demo -C " + fx.wt + " -key task-1-implementer -session-token mf-t -commit " + s1 + " -outcome completed"
	if calls := fx.calls(t); len(calls) == 0 || calls[0] != want {
		t.Errorf("first call = %q, want %q", calls, want)
	}
}

func TestCloseTaskSurfacesAFailedDispatchEnd(t *testing.T) {
	t.Parallel()
	fx := ctFixture(t, true, "0", "1")
	s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
	r := runGuard("close-task", []string{"-end-key", "k", "-session-token", "mf-t", fx.wt, "demo", fx.base, "1:" + s1}, fx.env)
	if r.rc != 3 {
		t.Fatalf("exit %d, want 3\n%s", r.rc, r.out)
	}
	calls := strings.Join(fx.calls(t), "\n")
	if !strings.Contains(calls, "flow tasks tick -C "+fx.wt+" demo 1") || !strings.Contains(calls, "git -C "+fx.wt+" push origin spectre/demo") {
		t.Fatalf("tick or push missing:\n%s", calls)
	}
	lines := strings.Split(strings.TrimRight(r.err, "\n"), "\n")
	last := lines[len(lines)-1]
	if !strings.HasPrefix(last, "close-task: UNRECORDED — the dispatch end for k did not reach the store; run: flow record dispatch end -change demo -C ") || !strings.HasSuffix(last, "-outcome completed") {
		t.Fatalf("last stderr line %q", last)
	}
}

// kan-1008's self-review: a recorded worktree the commit map leaves out
// joins the verdict through its newest commit carrying the task's trailer,
// so a commit there touching an undeclared path is refused; a worktree with
// no such commit adds nothing.
func TestCloseTaskJoinsARecordedWorktreeTheMapLeftOut(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, peerFile, peerTask string
		rc                       int
	}{
		{"a peer commit outside Files is refused", "z.txt", "1", 1},
		{"a peer commit inside Files passes", "a.txt", "1", 0},
		{"a peer commit of another task adds nothing", "z.txt", "2", 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fx := ctFixture(t, true, "0", "0")
			gitRun(t, fx.peer, "branch", "main")
			writeFile(t, fx.state, `{"worktrees":{"`+fx.wt+`":"`+fx.base+`","`+fx.peer+`":"-"}}`)
			fx.commit(t, fx.peer, "one", c.peerTask, c.peerFile, "p\n")
			s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
			r := fx.run(fx.wt, "demo", fx.base, "1:"+s1)
			if r.rc != c.rc {
				t.Fatalf("exit %d, want %d\n%s", r.rc, c.rc, r.out)
			}
			joined := strings.Contains(r.err, "close-task: task 1: joined "+fx.peer+"=")
			if joined != (c.peerTask == "1") {
				t.Errorf("joined line present = %v, want %v\n%s", joined, c.peerTask == "1", r.out)
			}
			if c.rc == 1 && !strings.Contains(r.stdout, "z.txt") {
				t.Errorf("want the refusal to name z.txt\n%s", r.out)
			}
		})
	}
}

// A task commit carrying a full-app-suite.zip the rebuild script did not
// build is refused, and the committed bytes are left on disk; the script's
// own zip closes (design.md suite-zip-at-close-task).
func TestCloseTaskRefusesAHandBuiltSuiteZip(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("zip"); err != nil {
		t.Skip("zip is not on PATH")
	}
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	fx := ctFixture(t, true, "0", "0")
	base := fx.env.Getenv
	fx.env.Getenv = func(k string) string {
		switch k {
		case "FLOW_GUARD_SCRIPT_DIR":
			return filepath.Join(repo, "scripts")
		case "FLOW_GUARD_CACHE_DIR":
			return guardCache(t)
		}
		return base(k)
	}
	writeFile(t, fx.wt+"/.flow/project.md", "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `screenshots` | `shots` |\n")
	writeFile(t, fx.wt+"/shots/full-app-suite.spec.ts-snapshots/a.png", "a")
	gitRun(t, fx.wt, "add", "-f", ".flow", "shots")
	gitRun(t, fx.wt, "commit", "-q", "-m", "shots")
	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
	plan := strings.Replace(ctPlan, "**Files:** `a.txt`", "**Files:** `a.txt`, `shots/full-app-suite.zip`", 1)
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", plan)
	zipPath := fx.wt + "/shots/full-app-suite.zip"

	// Hand-built with archive/zip.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("a.png")
	w.Write([]byte("a"))
	zw.Close()
	writeFile(t, fx.wt+"/a.txt", "a\n")
	if err := os.WriteFile(zipPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, fx.wt, "add", "-f", "a.txt", "shots/full-app-suite.zip")
	gitRun(t, fx.wt, "commit", "-q", "-m", "one", "-m", "Task-Id: 1")
	s1 := tcfGit(t, fx.wt, "rev-parse", "HEAD")
	r := fx.run(fx.wt, "demo", fx.base, "1:"+s1)
	if r.rc != 1 || !strings.Contains(r.out, "close-task: task 1: shots/full-app-suite.zip is not rebuild-full-app-suite-zip.sh's output") {
		t.Fatalf("hand-built zip: exit %d, want 1\n%s", r.rc, r.out)
	}
	if got, _ := os.ReadFile(zipPath); !bytes.Equal(got, buf.Bytes()) {
		t.Errorf("the committed zip's bytes were not written back")
	}

	// The script's own zip closes.
	gitRun(t, fx.wt, "reset", "-q", "--hard", fx.base)
	if rc := Registry["rebuild-full-app-suite-zip"]([]string{fx.wt, fx.wt + "/shots"}, fx.env, io.Discard, os.Stderr); rc != 0 {
		t.Fatalf("rebuild exit %d", rc)
	}
	writeFile(t, fx.wt+"/a.txt", "a\n")
	gitRun(t, fx.wt, "add", "-f", "a.txt", "shots/full-app-suite.zip")
	gitRun(t, fx.wt, "commit", "-q", "-m", "one", "-m", "Task-Id: 1")
	s1 = tcfGit(t, fx.wt, "rev-parse", "HEAD")
	if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1); r.rc != 0 {
		t.Fatalf("script-built zip: exit %d, want 0\n%s", r.rc, r.out)
	}
}

// The zip lives in the regression checkout, a second repository whose own
// tree has no .flow/project.md: the rebuild reads the canonical worktree's
// config, which names that checkout (kan-1016's self-review, C1).
func TestCloseTaskRebuildsASuiteZipInTheRegressionCheckout(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("zip"); err != nil {
		t.Skip("zip is not on PATH")
	}
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	fx := ctFixture(t, true, "0", "0")
	// The resolver answers with git's real paths; the temp dir sits under a symlink on macOS.
	if fx.peer, err = filepath.EvalSymlinks(fx.peer); err != nil {
		t.Fatal(err)
	}
	base := fx.env.Getenv
	fx.env.Getenv = func(k string) string {
		switch k {
		case "FLOW_GUARD_SCRIPT_DIR":
			return filepath.Join(repo, "scripts")
		case "FLOW_GUARD_CACHE_DIR":
			return guardCache(t)
		}
		return base(k)
	}
	writeFile(t, fx.wt+"/.flow/project.md", "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `screenshots` | `shots` |\n| `regression checkout` | `"+fx.peer+"` |\n")
	gitRun(t, fx.wt, "add", "-f", ".flow")
	gitRun(t, fx.wt, "commit", "-q", "-m", "config")
	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
	writeFile(t, fx.peer+"/shots/full-app-suite.spec.ts-snapshots/a.png", "a")
	if rc := Registry["rebuild-full-app-suite-zip"]([]string{fx.wt, fx.peer + "/shots"}, fx.env, io.Discard, os.Stderr); rc != 0 {
		t.Fatalf("rebuild exit %d", rc)
	}
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", strings.Replace(ctPlan, "**Files:** `c.txt`, `d.txt`", "**Files:** `c.txt`, `shots/full-app-suite.zip`, `shots/full-app-suite.spec.ts-snapshots/a.png`", 1))
	s3 := fx.commit(t, fx.wt, "three", "3", "c.txt", "c\n")
	gitRun(t, fx.peer, "add", "-f", "shots")
	gitRun(t, fx.peer, "commit", "-q", "-m", "three", "-m", "Task-Id: 3")
	p3 := tcfGit(t, fx.peer, "rev-parse", "HEAD")
	if r := fx.run(fx.wt, "demo", fx.base, "3:"+fx.wt+"="+s3+","+fx.peer+"="+p3); r.rc != 0 {
		t.Fatalf("script-built zip in the regression checkout: exit %d, want 0\n%s", r.rc, r.out)
	}
}

// A zip deleted by the task commit is not judged (kan-1016's self-review, M1).
func TestCloseTaskSuiteZipAgainstTheTaskCommit(t *testing.T) {
	t.Parallel()
	fx := ctFixture(t, true, "0", "0")
	writeFile(t, fx.wt+"/shots/full-app-suite.zip", "z")
	gitRun(t, fx.wt, "add", "-f", "shots")
	gitRun(t, fx.wt, "commit", "-q", "-m", "zip")
	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", strings.Replace(ctPlan, "**Files:** `a.txt`", "**Files:** `a.txt`, `shots/full-app-suite.zip`", 1))
	writeFile(t, fx.wt+"/a.txt", "a\n")
	gitRun(t, fx.wt, "rm", "-q", "shots/full-app-suite.zip")
	gitRun(t, fx.wt, "add", "-f", "a.txt")
	gitRun(t, fx.wt, "commit", "-q", "-m", "one", "-m", "Task-Id: 1")
	s1 := tcfGit(t, fx.wt, "rev-parse", "HEAD")
	if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1); r.rc != 0 {
		t.Fatalf("deleted zip: exit %d, want 0\n%s", r.rc, r.out)
	}
}

// scripts/close-task.sh is close-task's only caller, and it runs the zip
// rebuild in-process: the shim exports FLOW_GUARD_SCRIPT_DIR, which the
// rebuild needs to find resolve-visual-screenshots.sh (kan-1024 Task 14:
// unset, close-task stopped on exit 2 for every committed zip).
func TestCloseTaskShimExportsTheRebuildsScriptDir(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("zip"); err != nil {
		t.Skip("zip is not on PATH")
	}
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	fx := ctFixture(t, true, "0", "0")
	writeFile(t, fx.wt+"/.flow/project.md", "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `screenshots` | `shots` |\n")
	writeFile(t, fx.wt+"/shots/full-app-suite.spec.ts-snapshots/a.png", "a")
	gitRun(t, fx.wt, "add", "-f", ".flow", "shots")
	gitRun(t, fx.wt, "commit", "-q", "-m", "shots")
	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", strings.Replace(ctPlan, "**Files:** `a.txt`", "**Files:** `a.txt`, `shots/full-app-suite.zip`", 1))
	writeFile(t, fx.wt+"/a.txt", "a\n")
	writeFile(t, fx.wt+"/shots/full-app-suite.zip", "hand-built")
	gitRun(t, fx.wt, "add", "-f", "a.txt", "shots/full-app-suite.zip")
	gitRun(t, fx.wt, "commit", "-q", "-m", "one", "-m", "Task-Id: 1")
	s1 := tcfGit(t, fx.wt, "rev-parse", "HEAD")
	cmd := exec.Command(filepath.Join(repo, "scripts", "close-task.sh"), fx.wt, "demo", fx.base, "1:"+s1)
	cmd.Dir = fx.wt
	cmd.Env = []string{"PATH=" + fx.env.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "FLOW_GUARD_TELEMETRY=off", "FLOW_GUARD_CACHE_DIR=" + guardCache(t)}
	out, err := cmd.CombinedOutput()
	if code := exitCode(err); code != 1 || !strings.Contains(string(out), "is not rebuild-full-app-suite-zip.sh's output") {
		t.Fatalf("close-task.sh on a hand-built zip: exit %d, want 1 (the rebuild ran and differed)\n%s", code, out)
	}
}

// TestCloseTaskRunsPlanShapeOnInRunFix pins the kan-1046 self-review fix
// (1f): closing an in-run visual-verify fix task runs the plan guard first,
// so one whose **Files:** omits visual-verification.md is refused with F21's
// line and nothing ticked; with the record listed it closes; a planned task
// over the same bad plan never runs the guard.
func TestCloseTaskRunsPlanShapeOnInRunFix(t *testing.T) {
	t.Parallel()
	fix := func(files string) string {
		return ctPlan + "\n- [ ] 4. In-run fix 1 — visual-verify: x\n\n**Files:** " + files + "\n**Tests:** none\n**Commit:** four\n**Build:** green\n"
	}
	const f21 = "is an in-run visual-verify fix whose **Files:** omits"
	for _, c := range []struct {
		name, plan, task, subject string
		files                     []string
		refused                   bool
	}{
		{"an in-run fix omitting the record is refused", fix("¤e.txt¤"), "4", "four", []string{"e.txt", "x\n"}, true},
		{"an in-run fix listing the record closes", fix("¤e.txt¤, ¤docs/visual-verification.md¤"), "4", "four",
			[]string{"e.txt", "x\n", "docs/visual-verification.md", "r\n"}, false},
		{"a planned task never runs the guard", fix("¤e.txt¤"), "2", "two", []string{"b.txt", "x\n"}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fx := ctFixture(t, true, "0", "0")
			writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", bt(c.plan))
			writeFile(t, fx.dispatches, `[{"key":"verify-fix-1","taskId":"4","commitSha":"abc"}]`)
			sha := fx.commit(t, fx.wt, c.subject, c.task, c.files...)
			r := fx.run(fx.wt, "demo", fx.base, c.task+":"+sha)
			if got := strings.Contains(r.out+r.err, f21); got != c.refused {
				t.Errorf("F21 line present %v, want %v\n%s\n%s", got, c.refused, r.out, r.err)
			}
			if c.refused {
				if r.rc != 1 {
					t.Fatalf("exit %d, want 1\n%s\n%s", r.rc, r.out, r.err)
				}
				for _, call := range fx.calls(t) {
					if strings.HasPrefix(call, "flow tasks tick") {
						t.Errorf("ticked after a refusal: %s", call)
					}
				}
				return
			}
			if r.rc != 0 {
				t.Fatalf("exit %d, want 0\n%s\n%s", r.rc, r.out, r.err)
			}
		})
	}
}

// TestCloseTaskRefusesAnAppendedTaskRowWithNoCommit pins the kan-1059
// self-review fix, 5a: an appended task none of whose naming rows carries a
// commit is refused, exit 1, nothing ticked, naming the row and the repair;
// one naming row with a commit closes it, and a non-appended task is never
// judged.
func TestCloseTaskRefusesAnAppendedTaskRowWithNoCommit(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, header, rows string
		refused            bool
	}{
		{"an appended task whose row carries no commit is refused", "> **Tasks appended:** 2\n\n", `[{"id":8623,"key":"task-2-implementer","taskId":""}]`, true},
		{"an appended task with one row carrying a commit closes", "> **Tasks appended:** 2\n\n", `[{"id":8623,"key":"task-2-implementer","taskId":""},{"id":8624,"key":"task-2-implementer-retry","taskId":"2","commitSha":"abc"}]`, false},
		{"a non-appended task with a commit-less row closes", "", `[{"id":8623,"key":"task-2-implementer","taskId":""}]`, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fx := ctFixture(t, true, "0", "0")
			writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", c.header+ctPlan)
			writeFile(t, fx.dispatches, c.rows)
			s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", "b\n")
			r := fx.run(fx.wt, "demo", fx.base, "2:"+s2)
			if !c.refused {
				if r.rc != 0 {
					t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
				}
				return
			}
			if r.rc != 1 || !strings.Contains(r.err, "dispatch row 8623 carries no commit") || !strings.Contains(r.err, "flowd retime-dispatch 8623") {
				t.Fatalf("exit %d, want 1 naming row 8623 and flowd retime-dispatch\n%s", r.rc, r.out)
			}
			ctEqual(t, fx.calls(t), []string{"flow record dispatches -change demo -C " + fx.wt})
		})
	}
}
