package guard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestCheckTaskRecords ports scripts/test-check-task-records.sh at
// bbfe685e: one subtest per ok: label, the guard run in-process against real
// git fixtures copied from one base repository (trBase). The harness's two
// wrapper cases -- a flow-guard that cannot be built, a missing
// lib/flow-guard.sh -- exec a copy of the real shim.
func TestCheckTaskRecords(t *testing.T) {
	t.Parallel()
	type check struct{ label, fail string }
	var checks []check
	want := func(label string, ok bool, r guardResult) {
		c := check{label: label}
		if !ok {
			c.fail = fmt.Sprintf("rc=%d out=%q", r.rc, r.out)
		}
		checks = append(checks, c)
	}
	has := func(r guardResult, s string) bool { return strings.Contains(r.out, s) }
	run := func(env Env, args ...string) guardResult {
		if env.Getenv == nil {
			env.Getenv = os.Getenv
		}
		return runGuard("check-task-records", args, env)
	}

	fx := trRepo(t)
	fx.task(1, "x", "feat(a): one", "green")
	fx.task(2, " ", "-", "green")
	fx.commitAll("docs(plan): fixture plan")
	fx.commitFile("src/f1.go", "feat(a): one")
	r := run(Env{}, fx.root)
	want("case 1 clean plan exits 0 silent", r.rc == 0 && r.out == "", r)

	fx = trRepo(t)
	fx.task(1, "x", "feat(a): missing", "green")
	r = run(Env{}, fx.root)
	want("case 2 ticked without commit reported", r.rc == 1 && has(r, "task 1 ticked but no commit"), r)

	fx = trRepo(t)
	fx.task(1, " ", "feat(a): landed", "green")
	fx.commitAll("docs(plan): fixture plan")
	fx.commitFile("src/f1.go", "feat(a): landed")
	r = run(Env{}, fx.root)
	want("case 3 commit without tick reported", r.rc == 1 && has(r, "task 1 not ticked but commit"), r)

	fx = trRepo(t)
	fx.task(1, "x", "feat(a): sneaky", "green")
	fx.commitAll("docs(plan): fixture plan")
	fx.commitFile("src/f1.go", "feat(a): declared")
	fx.commitFile("src/undeclared.go", "feat(a): sneaky")
	r = run(Env{}, fx.root)
	want("case 4 undeclared commit file reported", r.rc == 1 && has(r, "src/undeclared.go is not declared"), r)

	fx = trRepo(t)
	writeFile(t, fx.tasks, "- [x] 1. Task 1\n**Files:** `src/f1.go` `src/f9.go`\n**Tests:** none\n**Commit:** `feat(a): partial`\n**Build:** green\n**After:** none\n")
	fx.commitAll("docs(plan): fixture plan")
	fx.commitFile("src/f1.go", "feat(a): partial")
	r = run(Env{}, fx.root)
	want("case 5 declared-but-untouched file reported", r.rc == 1 && has(r, "src/f9.go is not touched"), r)

	fx = trRepo(t)
	fx.task(1, "x", "feat(a): collateral", "green", "**Allowed-collateral:** `gen/**`")
	fx.commitAll("docs(plan): fixture plan")
	writeFile(t, fx.root+"/src/f1.go", "f1\n")
	writeFile(t, fx.root+"/gen/generated.go", "gen\n")
	fx.commitAll("feat(a): collateral")
	r = run(Env{}, fx.root)
	want("case 6 allowed-collateral glob passes", r.rc == 0, r)

	fx = trRepo(t)
	fx.task(1, " ", "-", "yellow")
	r = run(Env{}, fx.root)
	want("case 7 malformed Build tag reported", r.rc == 1 && has(r, "malformed **Build:** tag"), r)

	fx = trRepo(t)
	writeFile(t, fx.tasks, "- [ ] 1. Task 1\n**Files:** `src/f1.go`\n**Tests:** none\n**After:** none\n")
	r = run(Env{}, fx.root)
	want("case 8 missing Build tag reported", r.rc == 1 && has(r, "has no **Build:** tag"), r)

	fx = trRepo(t)
	if err := os.RemoveAll(fx.root + "/spectre"); err != nil {
		t.Fatal(err)
	}
	r = run(Env{}, fx.root)
	want("case 9 no plan exits 0 silent", r.rc == 0 && r.out == "", r)

	fx = trRepo(t)
	r = run(Env{}, fx.root, "no-such-ref")
	want("case 10 unusable base-ref refused exit 2", r.rc == 2 && has(r, "COULD NOT JUDGE"), r)

	fx = trRepo(t)
	fx.task(1, "x", "feat(fold): folded", "green")
	fx.task(2, "x", "feat(fold): folded", "red", "**Squash-with:** Task 1", "**After:** Task 1")
	fx.commitAll("docs(plan): fixture plan")
	writeFile(t, fx.root+"/src/f1.go", "f1\n")
	writeFile(t, fx.root+"/src/f2.go", "f2\n")
	fx.commitAll("feat(fold): folded")
	r = run(Env{}, fx.root)
	want("case 11 red fold union passes", r.rc == 0, r)

	fx = trRepo(t)
	fx.task(1, "x", "feat(a): twice", "green")
	fx.commitAll("docs(plan): fixture plan")
	fx.commitFile("src/f1.go", "feat(a): twice")
	fx.commitFile("src/f9.go", "feat(a): twice")
	r = run(Env{}, fx.root)
	want("case 13 duplicate-subject commits reported", r.rc == 1 && has(r, "carried by 2 commits"), r)

	fx = trRepo(t)
	fx.task(1, "x", "feat(a): landed", "green")
	fx.commitAll("docs(plan): fixture plan")
	fx.commitFile("src/f1.go", "feat(a): landed")
	gitRun(t, fx.root, "update-ref", "refs/remotes/origin/main", "HEAD")
	r = run(Env{}, fx.root)
	want("case 14 landed plan past the base reads clean", r.rc == 0 && r.out == "", r)

	fx = trRepo(t)
	fx.archived("old-change", " ", "feat(a): ghost")
	fx.commitAll("chore(spectre): archive old-change")
	gitRun(t, fx.root, "update-ref", "refs/remotes/origin/main", "HEAD")
	fx.commitFile("src/f1.go", "feat(a): ghost")
	r = run(Env{}, fx.root)
	want("case 15 plan archived at the base ignored", r.rc == 0 && r.out == "", r)

	fx = trRepo(t)
	fx.archived("new-change", " ", "feat(a): ghost")
	fx.commitAll("chore(spectre): archive new-change")
	fx.commitFile("src/f1.go", "feat(a): ghost")
	r = run(Env{}, fx.root)
	want("case 15b plan archived on this branch checked", r.rc == 1 && has(r, "spectre/changes/archive/new-change/tasks.md: task 1 not ticked"), r)

	fx = trRepo(t)
	fx.archived("new-change", "x", "feat(a): collapsed")
	fx.commitAll("chore(spectre): archive new-change")
	r = run(Env{}, fx.root)
	want("case 15c archived ticked task with a collapsed commit reads clean", r.rc == 0 && r.out == "", r)

	fx = trRepo(t)
	fx.task(1, "x", "feat(a): twelve", "green")
	fx.commitAll("docs(plan): fixture plan")
	fx.commitFile("src/f1.go", "feat(a): twelve")
	r = run(Env{}, fx.root, "main")
	want("case 12 bare base-ref argument accepted", r.rc == 0, r)

	fx = trRepo(t)
	gitRun(t, fx.root, "config", "branch.main.flowBase", "no-such-ref")
	r = run(Env{}, fx.root)
	want("case 16 recorded base read before origin/HEAD", r.rc == 2 && has(r, "COULD NOT JUDGE"), r)

	// Cases 17-19: a cross-repo change (link.md); a stub `flow` on PATH
	// stands in for the store.
	link := "## Merge order\n\n1. `.`\n2. `peer`\n"
	stub := func(body string) Env {
		dir := t.TempDir()
		writeExec(t, dir+"/flow", "#!/bin/sh\n"+body)
		return Env{Getenv: pathEnv(dir)}
	}
	fx = trRepo(t)
	writeFile(t, fx.root+"/spectre/changes/fixture-change/link.md", link)
	fx.task(1, "x", "fix(peer): one", "green")
	fx.commitAll("docs(plan): fixture plan")
	peer := t.TempDir()
	gitRun(t, peer, "init", "-q", "-b", "main")
	writeFile(t, peer+"/src/f1.go", "src/f1.go\n")
	gitRun(t, peer, "add", "-A")
	gitRun(t, peer, "commit", "-qm", "fix(peer): one")
	r = run(stub(fmt.Sprintf("printf '%%s\\n' '{\"worktrees\":{\"%s\":\"x\",\"%s\":\"y\"}}'\n", peer, fx.root)), fx.root)
	want("case 17 ticked task landed in a peer worktree exits 0 silent", r.rc == 0 && r.out == "", r)

	fx = trRepo(t)
	writeFile(t, fx.root+"/spectre/changes/fixture-change/link.md", link)
	fx.task(1, "x", "fix(peer): one", "green")
	fx.commitAll("docs(plan): fixture plan")
	r = run(stub("exit 1\n"), fx.root)
	want("case 18 unreadable state record keeps the violation", r.rc == 1 && has(r, "no commit with subject 'fix(peer): one'"), r)

	fx = trRepo(t)
	writeFile(t, fx.root+"/spectre/changes/fixture-change/link.md", link)
	fx.task(1, "x", "fix(peer): one", "green")
	fx.commitAll("docs(plan): fixture plan")
	r = run(stub(fmt.Sprintf("printf '%%s\\n' '{\"worktrees\":{\"%s\":\"x\"}}'\n", t.TempDir())), fx.root)
	want("case 19 non-repository record path is skipped", r.rc == 1 && has(r, "no commit with subject 'fix(peer): one'"), r)

	fx = trRepo(t)
	writeFile(t, fx.root+"/spectre/changes/fixture-change/link.md", link)
	fx.task(1, "x", "fix(peer): one", "green")
	fx.commitAll("docs(plan): fixture plan")
	sdir := t.TempDir()
	writeExec(t, sdir+"/flow", "#!/bin/sh\n"+fmt.Sprintf("[ \"$FLOW_ADDR\" = http://127.0.0.1:4173 ] || exit 1\nprintf '%%s\\n' '{\"worktrees\":{\"%s\":\"x\",\"%s\":\"y\"}}'\n", peer, fx.root))
	addrs := map[string]string{"FLOW_ADDR": "http://127.0.0.1:6783", "FLOW_RECORDS_ADDR": "http://127.0.0.1:4173"}
	base := pathEnv(sdir)
	r = run(Env{
		Getenv: func(k string) string {
			if v, ok := addrs[k]; ok {
				return v
			}
			return base(k)
		},
		Environ: func() []string {
			return append(os.Environ(), "FLOW_ADDR="+addrs["FLOW_ADDR"], "FLOW_RECORDS_ADDR="+addrs["FLOW_RECORDS_ADDR"])
		},
	}, fx.root)
	want("case 20 peer state read follows FLOW_RECORDS_ADDR", r.rc == 0 && r.out == "", r)

	// Unlanded archives: a shim copy with no stats/ beside it cannot build
	// flow-guard; the same copy with no lib/flow-guard.sh cannot load it.
	u := t.TempDir()
	mkdir(t, u+"/root/spectre/changes/some-change")
	psCopy(t, trShim, u+"/scripts/check-task-records.sh")
	psCopy(t, psScripts+"/lib/flow-guard.sh", u+"/scripts/lib/flow-guard.sh")
	uenv := []string{"FLOW_GUARD_CACHE_DIR=" + t.TempDir()}
	r = psExec(t, uenv, u+"/scripts/check-task-records.sh", u+"/root")
	want("unlanded archives: flow-guard that cannot be built exits 2", r.rc == 2 && has(r, "check-task-records: cannot build flow-guard"), r)
	if err := os.Remove(u + "/scripts/lib/flow-guard.sh"); err != nil {
		t.Fatal(err)
	}
	r = psExec(t, uenv, u+"/scripts/check-task-records.sh", u+"/root")
	want("unlanded archives: a missing lib/flow-guard.sh exits 2", r.rc == 2 && r.err == "check-task-records: cannot load lib/flow-guard.sh beside "+u+"/scripts/check-task-records.sh\n", r)

	// Beyond the harness: the outputs it never pinned, as the wrapper and
	// check-task-records.py printed them at bbfe685e.
	fx = trRepo(t)
	fx.task(1, "x", "feat(a): sneaky", "yellow")
	fx.task(2, " ", "feat(a): early", "green")
	fx.task(3, "x", "feat(a): gone", "green")
	fx.task(4, "x", "feat(a): partial", "green", "**Allowed-collateral:** `gen/**`")
	fx.commitAll("docs(plan): fixture plan")
	fx.commitFile("src/undeclared.go", "feat(a): sneaky")
	sneaky := fx.head(t)[:12]
	fx.commitFile("src/f2.go", "feat(a): early")
	early := fx.head(t)[:12]
	fx.commitFile("gen/x.go", "feat(a): partial")
	fx.commitFile("gen/y.go", "feat(a): partial")
	rel := "spectre/changes/fixture-change/tasks.md"
	r = run(Env{Dir: fx.root})
	want("pin: every violation shape on stdout, in order, from the working directory's root", r.rc == 1 && r.err == "" && r.stdout == ""+
		rel+": task 1 has a malformed **Build:** tag: 'yellow'\n"+
		rel+": task 2 not ticked but commit with subject 'feat(a): early' exists in origin/main..HEAD ("+early+")\n"+
		rel+": task 3 ticked but no commit with subject 'feat(a): gone' reachable from HEAD\n"+
		rel+": task 1: file src/undeclared.go is not declared in Files: and not covered by Allowed-collateral:\n"+
		rel+": task 1: declared file src/f1.go is not touched by its commit "+sneaky+"\n"+
		rel+": task 4: subject 'feat(a): partial' is carried by 2 commits in origin/main..HEAD; expected one\n"+
		rel+": task 4: declared file src/f4.go is not touched by its commit "+fx.head(t)[:12]+"\n", r)
	r = run(Env{}, "a", "b", "c")
	want("pin: the usage line", r.rc == 2 && r.stdout == "" && r.err == "usage: check-task-records.sh [repo-root] [base-ref]\n", r)
	missing := t.TempDir() + "/nope"
	r = run(Env{}, missing)
	want("pin: a missing root exits 2", r.rc == 2 && r.err == "check-task-records: COULD NOT JUDGE — not a record verdict: repo root not found: "+trResolve(missing)+"\n", r)
	fx = trRepo(t)
	r = run(Env{}, fx.root, "no-such-ref")
	want("pin: the unresolved base-ref line", r.err == "check-task-records: COULD NOT JUDGE — not a record verdict: base-ref does not resolve: no-such-ref\n", r)
	gitRun(t, fx.root, "symbolic-ref", "-d", "refs/remotes/origin/HEAD")
	r = run(Env{}, fx.root)
	want("pin: no origin/HEAD and no base-ref exits 2", r.rc == 2 &&
		r.err == "check-task-records: COULD NOT JUDGE — not a record verdict: refs/remotes/origin/HEAD does not resolve and no base-ref argument was given\n", r)
	// Python let read_text()'s UnicodeDecodeError escape as a traceback
	// (exit 1); the header's "an unreadable plan" is exit 2, and the header
	// wins.
	fx = trRepo(t)
	writeFile(t, fx.tasks, "- [ ] 1. \xff\n")
	r = run(Env{}, fx.root)
	want("pin: a plan that is not UTF-8 exits 2", r.rc == 2 && r.stdout == "" &&
		strings.HasPrefix(r.err, "check-task-records: COULD NOT JUDGE — not a record verdict: cannot read "+trResolve(fx.tasks)+": 'utf-8' codec can't decode byte 0xff"), r)
	fx = trRepo(t)
	fx.task(1, "x", "feat(a): missing", "green")
	sandbox := t.TempDir()
	psCopy(t, trShim, sandbox+"/scripts/check-task-records.sh")
	psLink(t, psScripts+"/lib", sandbox+"/scripts/lib")
	psLink(t, "../../../scripts/lib", sandbox+"/skills/flow/scripts/lib")
	psLink(t, "../../../scripts/check-task-records.sh", sandbox+"/skills/flow/scripts/check-task-records.sh")
	r = psExec(t, []string{"FLOW_GUARD_CACHE_DIR=" + guardCache(t)}, sandbox+"/skills/flow/scripts/check-task-records.sh", fx.root)
	want("pin: the symlinked shim judges the root it is given", r.rc == 1 &&
		r.stdout == rel+": task 1 ticked but no commit with subject 'feat(a): missing' reachable from HEAD\n", r)

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.fail != "" {
				t.Error(c.fail)
			}
		})
	}
}

var trShim = psScripts + "/check-task-records.sh"

// trFx is one fixture repository: a copy of trBase with its plan at tasks.
type trFx struct {
	t           *testing.T
	root, tasks string
}

// trBase is the harness's new_repo, built once: a base commit recorded at
// refs/remotes/origin/main with origin/HEAD pointing at it, so base..HEAD is
// exactly the commits a case makes after copying it.
var trBase = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp(execFixtures.dir, "task-records")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir+"/spectre/changes/fixture-change", 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(dir+"/base.txt", []byte("base\n"), 0o644); err != nil {
		return "", err
	}
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"add", "-A"},
		{"-c", "user.email=fixture@example.com", "-c", "user.name=Fixture", "commit", "-qm", "chore: base commit"},
		{"update-ref", "refs/remotes/origin/main", "HEAD"},
		{"symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main"},
	} {
		if out, err := trGit(dir, args...).CombinedOutput(); err != nil {
			return "", fmt.Errorf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir, nil
})

func trRepo(t *testing.T) *trFx {
	t.Helper()
	base, err := trBase()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.CopyFS(root, os.DirFS(base)); err != nil {
		t.Fatal(err)
	}
	fx := &trFx{t, root, root + "/spectre/changes/fixture-change/tasks.md"}
	writeFile(t, fx.tasks, "")
	return fx
}

// task is the harness's task_block: subject "-" declares no Commit:.
func (fx *trFx) task(n int, state, subject, build string, extra ...string) {
	var b strings.Builder
	fmt.Fprintf(&b, "- [%s] %d. Task %d\n**Files:** `src/f%d.go`\n**Tests:** none\n", state, n, n, n)
	if subject != "-" {
		fmt.Fprintf(&b, "**Commit:** `%s`\n", subject)
	}
	fmt.Fprintf(&b, "**Build:** %s\n", build)
	for _, l := range extra {
		b.WriteString(l + "\n")
	}
	b.WriteString("**After:** none\n\n")
	old, _ := os.ReadFile(fx.tasks)
	writeFile(fx.t, fx.tasks, string(old)+b.String())
}

// archived writes a one-task plan under spectre/changes/archive/<name>/.
func (fx *trFx) archived(name, state, subject string) {
	writeFile(fx.t, fx.root+"/spectre/changes/archive/"+name+"/tasks.md",
		"- ["+state+"] 1. Task 1\n**Files:** `src/f1.go`\n**Tests:** none\n**Commit:** `"+subject+"`\n**Build:** green\n**After:** none\n")
}

func (fx *trFx) commitAll(subject string) {
	gitRun(fx.t, fx.root, "add", "-A")
	gitRun(fx.t, fx.root, "commit", "-qm", subject)
}

// commitFile writes, adds and commits exactly one file.
func (fx *trFx) commitFile(path, subject string) {
	writeFile(fx.t, fx.root+"/"+path, path+"\n")
	gitRun(fx.t, fx.root, "add", path)
	gitRun(fx.t, fx.root, "commit", "-qm", subject)
}

func (fx *trFx) head(t *testing.T) string {
	out, ok := gitOut(Env{Dir: fx.root}, "rev-parse", "HEAD")
	if !ok {
		t.Fatal("rev-parse HEAD failed")
	}
	return out
}

func trGit(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command(fixtureGit, append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), fixtureGitEnv...)
	return cmd
}

// trResolve is Path.resolve(): symlinks resolved as far as the path exists.
func trResolve(p string) string {
	if r, err := filepath.EvalSymlinks(filepath.Dir(p)); err == nil {
		return r + "/" + filepath.Base(p)
	}
	return p
}
