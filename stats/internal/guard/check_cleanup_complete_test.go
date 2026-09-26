package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Every case of scripts/test-check-cleanup-complete.sh at 0747740, one leaf
// subtest per ok: label, nested under the case it belongs to; each skip:
// path stays a t.Skip with the harness's reason. The harness's header,
// which is not carried here, states why each case exists — read it with
// `git show 0747740:scripts/test-check-cleanup-complete.sh`; the comments here
// say only what the port changed. The guard's own branch reasoning is in
// cleanupcomplete.go.
//
// Survivor-timeout cases inject Env.SurvivorsTimeout/SurvivorsKillGrace as
// durations (Decision: inject-deadlines-in-process). The cases whose
// assertions do not depend on the fixture having forked anything use a
// sub-second bound. The ones that do -- 28 asserts on output printed, 28b
// and 28d on children forked, BEFORE the kill -- fire the bound on readiness
// through Env.SurvivorsExpire (ccExpireWhen), never on the wall clock: macOS
// spends ~0.2s on the first exec of a freshly written script (measured:
// 0.201s cold, 0.009s warm), and a wall-clock bound, even 3s, lost that race
// under the package's parallel -race load. 29 needs the bound never to fire.
const (
	ccHangBound = 300 * time.Millisecond
	ccForkBound = 3 * time.Second
	ccTestGrace = 100 * time.Millisecond
)

// ccExpireWhen is a survivors bound that fires once every path exists (or
// after a 30s backstop, so a fixture that never gets there cannot hang the
// test; the case's own readiness assertion then fails).
func ccExpireWhen(paths ...string) <-chan time.Time {
	ch := make(chan time.Time, 1)
	go func() {
		deadline := time.Now().Add(30 * time.Second)
		for _, p := range paths {
			for !isFile(p) && time.Now().Before(deadline) {
				time.Sleep(5 * time.Millisecond)
			}
		}
		ch <- time.Now()
	}()
	return ch
}

// ccRepoRoot is this repository's root, for the two cases that read a
// canonical file rather than a fixture copy of it (the harness's header).
const ccRepoRoot = "../../.."

type ccFx struct {
	base, repo, state string
	env               map[string]string
	timeout, grace    time.Duration
	expire            <-chan time.Time // Env.SurvivorsExpire
}

type ccResult struct {
	rc       int
	out, err string // out is stdout only; the harness captures the two apart
}

// ccTemplate is new_fixture's repository after a successful run-2 cleanup of
// "demo", built once and copied by every case.
func ccTemplate(t *testing.T) string {
	dir := t.TempDir() + "/repo"
	gitRun(t, "", "init", "-q", "-b", "main", dir)
	gitRun(t, dir, "config", "user.email", "test@example.invalid")
	gitRun(t, dir, "config", "user.name", "Test")
	mkdir(t, dir+"/spectre/changes/archive/2026-01-01-demo")
	writeFile(t, dir+"/f.txt", "base\n")
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-qm", "base")
	return dir
}

// ccNew is new_fixture. Every pid a fixture script appends to <base>/pids is
// SIGKILLed with its process group at cleanup, so a failed assertion never
// leaves a fixture running on the machine.
func ccNew(t *testing.T, tmpl string) *ccFx {
	base := t.TempDir()
	f := &ccFx{base: base, repo: base + "/repo", state: base + "/state"}
	gdcCopyTree(t, tmpl, f.repo)
	mkdir(t, f.state)
	mkdir(t, base+"/tmp")
	f.env = map[string]string{"PATH": os.Getenv("PATH"), "TMPDIR": base + "/tmp"}
	t.Cleanup(func() {
		b, _ := os.ReadFile(base + "/pids")
		own := syscall.Getpgrp()
		for _, s := range strings.Fields(string(b)) {
			pid, err := strconv.Atoi(s)
			if err != nil || pid <= 1 {
				continue
			}
			if pgid, err := syscall.Getpgid(pid); err == nil && pgid > 1 && pgid != own {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})
	return f
}

func (f *ccFx) run(t *testing.T, args ...string) ccResult {
	t.Helper()
	fn := Registry["check-cleanup-complete"]
	if fn == nil {
		t.Fatal("check-cleanup-complete is not registered")
	}
	var o, e bytes.Buffer
	env := Env{Getenv: func(k string) string { return f.env[k] }, Dir: f.repo,
		SurvivorsTimeout: f.timeout, SurvivorsKillGrace: f.grace, SurvivorsExpire: f.expire}
	rc := fn(args, env, &o, &e)
	return ccResult{rc: rc, out: o.String(), err: e.String()}
}

func (f *ccFx) guard(t *testing.T, name string) ccResult {
	t.Helper()
	return f.run(t, f.repo, name, f.state)
}

// ccWorktree is add_worktree: a linked worktree on branch, under a holder
// directory of its own.
func (f *ccFx) worktree(t *testing.T, branch, leaf string) string {
	wt := t.TempDir() + "/" + leaf
	gitRun(t, f.repo, "worktree", "add", "-q", "-b", branch, wt)
	return wt
}

// isolation is declare_isolation: the survivors cell is written exactly as a
// project would write it; an empty one declares no survivors row.
func (f *ccFx) isolation(t *testing.T, cell string) {
	s := "# fixture project configuration\n\n## workspace isolation\n\n" +
		"| Resource | Variable | Default | In a workspace |\n" +
		"|----------|----------|---------|----------------|\n" +
		"| `database` | `DB_URL` | `appdb` | `appdb_<id_underscored>` |\n\n" +
		"| Command | Runs |\n|---------|------|\n" +
		fmt.Sprintf("| `create` | `touch '%s/create-ran'` |\n", f.repo) +
		fmt.Sprintf("| `remove` | `touch '%s/remove-ran'` |\n", f.repo)
	if cell != "" {
		s += fmt.Sprintf("| `survivors` | %s |\n", cell)
	}
	writeFile(t, f.repo+"/.flow/project.md", s)
}

// script is write_script: an executable bash script in the repository root.
func (f *ccFx) script(t *testing.T, name, body string) {
	writeExec(t, f.repo+"/"+name, "#!/usr/bin/env bash\n"+body+"\n")
}

// recordPid is the fixture line that registers the script's own pid for
// ccNew's cleanup.
func (f *ccFx) recordPid() string { return fmt.Sprintf("echo $$ >> '%s/pids'", f.base) }

func ccCheck(t *testing.T, label string, ok bool, format string, a ...any) {
	t.Helper()
	t.Run(label, func(t *testing.T) {
		if !ok {
			t.Fatalf(format, a...)
		}
	})
}

// ccVerdict is assert_verdict: exit 0, exactly one stdout line, beginning want.
func ccVerdict(t *testing.T, r ccResult, want, label string) {
	t.Helper()
	line := strings.TrimRight(r.out, "\n")
	var why string
	switch {
	case r.rc != 0:
		why = fmt.Sprintf("expected exit 0, got rc=%d out=%s err=%s", r.rc, r.out, r.err)
	case line == "":
		why = "expected a verdict line, got empty stdout (err=" + r.err + ")"
	case strings.Contains(line, "\n"):
		why = "expected exactly one stdout line, got: " + r.out
	case !strings.HasPrefix(line, want):
		why = "expected a line beginning " + want + ", got: " + r.out
	}
	ccCheck(t, label, why == "", "%s", why)
}

func ccHas(t *testing.T, r ccResult, needle, label string) {
	t.Helper()
	ccCheck(t, label, strings.Contains(r.out, needle), "the verdict line does not carry %q: %s", needle, r.out)
}

func ccLacks(t *testing.T, r ccResult, needle, label string) {
	t.Helper()
	ccCheck(t, label, !strings.Contains(r.out, needle), "the verdict line carries %q: %s", needle, r.out)
}

func ccAbsent(t *testing.T, path, label string) {
	t.Helper()
	_, err := os.Lstat(path)
	ccCheck(t, label, err != nil, "%s exists, so the command that creates it was run", path)
}

func ccPresent(t *testing.T, path, label string) {
	t.Helper()
	ccCheck(t, label, isFile(path), "%s was never created, so the command was not run", path)
}

// ccNoVerdict is assert_no_verdict: the refusal shape.
func ccNoVerdict(t *testing.T, r ccResult, label string) {
	t.Helper()
	ccCheck(t, label+": exits non-zero", r.rc != 0, "expected a non-zero exit, got rc=%d out=%s", r.rc, r.out)
	ccCheck(t, label+": writes nothing to stdout", r.out == "", "emitted a verdict line: %s", r.out)
	ccCheck(t, label+": names the failure on stderr", strings.Contains(r.err, "check-cleanup-complete: "),
		"no named message on stderr: %s", r.err)
}

// ccGone waits for a pid to be gone: a SIGKILLed orphan is reaped by launchd
// asynchronously, so absence is awaited under a deadline, never assumed.
func ccGone(pid int) bool {
	deadline := time.Now().Add(5 * time.Second)
	for syscall.Kill(pid, 0) == nil {
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(20 * time.Millisecond)
	}
	return true
}

func ccReadPid(t *testing.T, path string) int {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s was never written, so the fixture never reached its fork before the bound fired: %v", path, err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return pid
}

func TestCheckCleanupComplete(t *testing.T) {
	t.Parallel()
	tmpl := ccTemplate(t)
	cases := []struct {
		name string
		fn   func(t *testing.T, f *ccFx)
	}{
		{"1", func(t *testing.T, f *ccFx) {
			ccVerdict(t, f.guard(t, "demo"), "COMPLETE:", "a fully cleaned repository is COMPLETE")
		}},
		{"2", func(t *testing.T, f *ccFx) {
			f.worktree(t, "spectre/demo", "tree")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a surviving worktree is LEFTOVER")
			ccHas(t, r, "worktree", "a surviving worktree names its row")
		}},
		{"2b", func(t *testing.T, f *ccFx) {
			wt := f.worktree(t, "spectre/demo", "tree")
			if err := os.RemoveAll(wt); err != nil {
				t.Fatal(err)
			}
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a registered but pruned-away worktree is LEFTOVER")
			ccHas(t, r, "worktree", "a registered but pruned-away worktree names its row")
		}},
		{"2c", func(t *testing.T, f *ccFx) {
			wt := f.worktree(t, "spectre/demo", "tree with space")
			real, err := filepath.EvalSymlinks(wt)
			if err != nil {
				t.Fatal(err)
			}
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a worktree whose path contains a space is LEFTOVER")
			ccHas(t, r, real, "the breakdown names the worktree's whole path")
		}},
		{"3", func(t *testing.T, f *ccFx) {
			gitRun(t, f.repo, "branch", "spectre/demo")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a surviving local branch is LEFTOVER")
			ccHas(t, r, "local branch", "a surviving local branch names its row")
			ccLacks(t, r, "remote-tracking ref", "a surviving local branch is not read as a remote one")
		}},
		{"4", func(t *testing.T, f *ccFx) {
			gitRun(t, f.repo, "update-ref", "refs/remotes/origin/spectre/demo", "HEAD")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a surviving remote-tracking ref is LEFTOVER")
			ccHas(t, r, "remote-tracking ref", "a surviving remote-tracking ref names its row")
			ccLacks(t, r, "local branch", "a surviving remote-tracking ref is not read as a local branch")
		}},
		{"4b", func(t *testing.T, f *ccFx) {
			gitRun(t, f.repo, "branch", "spectre/demo")
			writeFile(t, f.repo+"/.git/refs/heads/spectre/demo", "garbage\n")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a corrupt local branch ref does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a corrupt local branch ref is reported as skipped")
			ccHas(t, r, "refs/heads/spectre/demo", "the skip names the ref it could not read")
		}},
		{"4c", func(t *testing.T, f *ccFx) {
			gitRun(t, f.repo, "update-ref", "refs/remotes/origin/spectre/demo", "HEAD")
			writeFile(t, f.repo+"/.git/refs/remotes/origin/spectre/demo", "garbage\n")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a corrupt remote-tracking ref does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a corrupt remote-tracking ref is reported as skipped")
			ccHas(t, r, "refs/remotes/origin/spectre/demo", "the skip names the remote-tracking ref it could not read")
		}},
		{"4d", func(t *testing.T, f *ccFx) {
			gitRun(t, f.repo, "branch", "spectre/demo")
			ref := f.repo + "/.git/refs/heads/spectre/demo"
			const label = "an unreadable local branch ref is reported as skipped"
			if !isFile(ref) {
				t.Run(label, func(t *testing.T) { t.Skip("this git stores the ref packed rather than loose") })
				return
			}
			if err := os.Chmod(ref, 0); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Chmod(ref, 0o644) }()
			if syscall.Access(ref, 4) == nil {
				t.Run(label, func(t *testing.T) { t.Skip("this process can read a mode-000 file") })
				return
			}
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "an unreadable local branch ref does not block the terminal state")
			ccHas(t, r, "SKIPPED", label)
			ccHas(t, r, "refs/heads/spectre/demo", "the skip names the unreadable ref")
		}},
		{"4e", func(t *testing.T, f *ccFx) {
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a repository with no branch at all is still COMPLETE")
			ccLacks(t, r, "SKIPPED", "a readable ref store produces no skip note")
		}},
		{"5", func(t *testing.T, f *ccFx) {
			mkdir(t, f.repo+"/spectre/changes/demo")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "an unarchived change directory is LEFTOVER")
			ccHas(t, r, "spectre/changes/demo", "an unarchived change directory names its row")
		}},
		{"5b", func(t *testing.T, f *ccFx) {
			mkdir(t, f.repo+"/spectre/changes/demo-fix-1")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "an unarchived demo-fix-1 sub-change is LEFTOVER")
			ccHas(t, r, "spectre/changes/demo-fix-1", "an unarchived sub-change names its own row")
			ccHas(t, r, "a sub-change of demo", "the sub-change row says whose sub-change it is")
		}},
		{"5c", func(t *testing.T, f *ccFx) {
			mkdir(t, f.repo+"/spectre/changes/archive/demo")
			mkdir(t, f.repo+"/spectre/changes/demo-fix-2")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "parent archived but sub-change left behind is LEFTOVER")
			ccHas(t, r, "spectre/changes/demo-fix-2", "the surviving sub-change is named")
		}},
		{"5d", func(t *testing.T, f *ccFx) {
			mkdir(t, f.repo+"/spectre/changes/demo-fix-1")
			mkdir(t, f.repo+"/spectre/changes/demo-fix-2")
			r := f.guard(t, "demo")
			ccHas(t, r, "spectre/changes/demo-fix-1", "the first sub-change is named")
			ccHas(t, r, "spectre/changes/demo-fix-2", "the second sub-change is named")
		}},
		{"5e", func(t *testing.T, f *ccFx) {
			mkdir(t, f.repo+"/spectre/changes/demo-fix-the-parser")
			mkdir(t, f.repo+"/spectre/changes/demo-fixup")
			ccVerdict(t, f.guard(t, "demo"), "COMPLETE:", "a differently-named change beginning demo-fix is not a sub-change")
		}},
		{"5f", func(t *testing.T, f *ccFx) {
			writeFile(t, f.repo+"/spectre/changes/demo/link.md", "## Part of\n`spectre:demo`\n")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a leftover satellite (link.md-only) change directory is LEFTOVER")
			ccHas(t, r, "spectre/changes/demo", "a leftover satellite directory names its row")
		}},
		{"5g", func(t *testing.T, f *ccFx) {
			writeFile(t, f.repo+"/spectre/changes/archive/demo/link.md", "## Part of\n`spectre:demo`\n")
			ccVerdict(t, f.guard(t, "demo"), "COMPLETE:", "an archived satellite (link.md-only) change directory is COMPLETE")
		}},
		{"6", func(t *testing.T, f *ccFx) {
			writeFile(t, f.state+"/demo-proposal-artifact.html", "")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a surviving proposal artifact source is LEFTOVER")
			ccHas(t, r, "proposal artifact source", "a surviving artifact source names its row")
		}},
		{"7", func(t *testing.T, f *ccFx) {
			f.worktree(t, "spectre/demo-other", "tree")
			gitRun(t, f.repo, "update-ref", "refs/remotes/origin/spectre/demo-other", "HEAD")
			mkdir(t, f.repo+"/spectre/changes/demo-other")
			writeFile(t, f.state+"/demo-other-proposal-artifact.html", "")
			ccVerdict(t, f.guard(t, "demo"), "COMPLETE:", "another change's artifacts are not this change's leftovers")
		}},
		{"8", func(t *testing.T, f *ccFx) {
			gitRun(t, f.repo, "branch", "spectre/demo")
			mkdir(t, f.repo+"/spectre/changes/demo")
			writeFile(t, f.state+"/demo-proposal-artifact.html", "")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "several leftovers still produce exactly one line")
			ccHas(t, r, "local branch", "the combined breakdown names the branch")
			ccHas(t, r, "spectre/changes/demo", "the combined breakdown names the change directory")
			ccHas(t, r, "proposal artifact source", "the combined breakdown names the artifact source")
		}},
		{"9", func(t *testing.T, f *ccFx) {
			ccNoVerdict(t, f.run(t, "/nonexistent/repo", "demo", f.state), "an unreadable repository")
		}},
		{"9b", func(t *testing.T, f *ccFx) {
			ccNoVerdict(t, f.run(t, t.TempDir(), "demo", f.state), "a directory that is not a git repository")
		}},
		{"10", func(t *testing.T, f *ccFx) {
			ccNoVerdict(t, f.run(t, f.repo, "demo", "/nonexistent/state"), "an unreadable state directory")
		}},
		{"11", func(t *testing.T, f *ccFx) {
			r := f.run(t, "", "", "")
			ccCheck(t, "missing arguments -> exit 2", r.rc == 2, "expected exit 2, got rc=%d out=%s", r.rc, r.out)
			ccCheck(t, "missing arguments: emits no verdict line", r.out == "", "emitted a verdict line: %s", r.out)
		}},
		{"12", func(t *testing.T, f *ccFx) {
			for _, bad := range []string{"../../nonexistent-decoy", "demo*", "demo/../demo", ".hidden", "demo?x"} {
				r := f.guard(t, bad)
				ccCheck(t, "a change name outside the allowlist ("+bad+") -> exit 2", r.rc == 2, "expected exit 2, got rc=%d out=%s", r.rc, r.out)
				ccCheck(t, "a change name outside the allowlist ("+bad+") emits no verdict line", r.out == "", "emitted a verdict line: %s", r.out)
				ccCheck(t, "a rejected change name ("+bad+") names the failure", strings.Contains(r.err, "check-cleanup-complete: change name"),
					"no named message on stderr: %s", r.err)
			}
		}},
		{"12b", func(t *testing.T, f *ccFx) {
			mkdir(t, f.repo+"/spectre/changes/demo")
			writeFile(t, f.state+"/demo-proposal-artifact.html", "")
			r := f.guard(t, "../../nonexistent-decoy")
			ccCheck(t, "a traversal-shaped name cannot report COMPLETE over a live change", r.out == "",
				"the guard produced a verdict for a name outside the repository: %s", r.out)
		}},
		{"13", func(t *testing.T, f *ccFx) {
			realGit, err := exec.LookPath("git")
			if err != nil {
				t.Fatal(err)
			}
			stub := f.base + "/gitstub"
			writeExec(t, stub+"/git", fmt.Sprintf(`#!/usr/bin/env bash
for arg in "$@"; do
  if [ "$arg" = "worktree" ]; then
    echo "fatal: injected failure — cannot list worktrees" >&2
    exit 128
  fi
done
exec %q "$@"
`, realGit))
			f.env["PATH"] = stub + ":" + f.env["PATH"]
			ccNoVerdict(t, f.guard(t, "demo"), "a failed worktree listing")
		}},
		{"14", func(t *testing.T, _ *ccFx) { ccRegistryCoupling(t) }},
		{"15", func(t *testing.T, f *ccFx) {
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a project with no .flow/project.md is COMPLETE")
			ccLacks(t, r, "workspace", "a project with no configuration says nothing about a workspace")
		}},
		{"15b", func(t *testing.T, f *ccFx) {
			writeFile(t, f.repo+"/.flow/project.md", "# fixture project configuration\n\n## test\n\n"+
				"| Command | Runs |\n|---------|------|\n"+
				fmt.Sprintf("| `survivors` | `touch '%s/decoy-ran'` |\n", f.repo))
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a project declaring no isolation is COMPLETE")
			ccLacks(t, r, "workspace", "a project declaring no isolation says nothing about a workspace")
			ccAbsent(t, f.repo+"/decoy-ran", "a command outside the isolation section is never run")
		}},
		{"16", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", fmt.Sprintf("touch '%s/survivors-ran'", f.repo))
			f.isolation(t, "`./survivors.sh`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "an empty survivor report is COMPLETE")
			ccPresent(t, f.repo+"/survivors-ran", "the declared survivors command is actually run")
			ccAbsent(t, f.repo+"/remove-ran", "the guard never runs the project's remove command")
			ccAbsent(t, f.repo+"/create-ran", "the guard never runs the project's create command")
			ccHas(t, r, "survivor report", "an empty survivor report is reported as verified")
		}},
		{"16b", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", `printf '\n\n'`)
			f.isolation(t, "`./survivors.sh`")
			ccVerdict(t, f.guard(t, "demo"), "COMPLETE:", "a survivor report of blank lines is COMPLETE")
		}},
		{"17", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", `printf 'db-survivor-one\ndb-survivor-two\n'`)
			f.isolation(t, "`./survivors.sh`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a reported survivor is LEFTOVER")
			ccHas(t, r, "db-survivor-one", "the breakdown names the first survivor")
			ccHas(t, r, "db-survivor-two", "the breakdown names the second survivor")
			ccLacks(t, r, "timed out", "an ordinary survivor report is not reported as a timeout")
		}},
		{"18", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", "exit 7")
			f.isolation(t, "`./survivors.sh`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "an unreachable service does not block the terminal state")
			ccHas(t, r, "SKIPPED", "an unreachable service is reported as skipped")
			ccHas(t, r, "./survivors.sh", "the skip names the command")
			ccHas(t, r, "exited 7", "the skip carries the exit code")
			ccLacks(t, r, "timed out", "an ordinary non-zero exit is not reported as a timeout")
		}},
		{"18b", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", "printf 'noise-not-a-survivor\\n'; exit 3")
			f.isolation(t, "`./survivors.sh`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "output from a failed survivor report is not a leftover")
			ccLacks(t, r, "noise-not-a-survivor", "output from a failed survivor report is not read as a survivor")
			ccHas(t, r, "SKIPPED", "a failed survivor report that printed something is still a skip")
		}},
		{"19", func(t *testing.T, f *ccFx) {
			f.isolation(t, "")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "no survivors command does not block the terminal state")
			ccHas(t, r, "SKIPPED", "no survivors command is reported as skipped, not as passed")
			ccAbsent(t, f.repo+"/remove-ran", "no survivors command: the removal is not run in its place")
			ccAbsent(t, f.repo+"/create-ran", "no survivors command: nothing else in the table is run either")
		}},
		{"19b", func(t *testing.T, f *ccFx) {
			writeFile(t, f.repo+"/.flow/project.md", "## workspace isolation\n\n| Command | Runs |\n|---------|------|\n"+
				fmt.Sprintf("| `remove` | `touch '%s/remove-ran'` |\n", f.repo)+"| `survivors` |  |\n")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "an empty survivors cell does not block the terminal state")
			ccHas(t, r, "SKIPPED", "an empty survivors cell is reported as skipped")
			ccAbsent(t, f.repo+"/remove-ran", "an empty survivors cell does not fall back to the removal")
		}},
		{"19c", func(t *testing.T, f *ccFx) {
			writeFile(t, f.repo+"/.flow/project.md", "## workspace isolation\n\n| Command | Runs |\n|---------|------|\n"+
				"| `create` | `true` |\n\n## test\n\n| Command | Runs |\n|---------|------|\n"+
				fmt.Sprintf("| `survivors` | `touch '%s/later-section-ran'` |\n", f.repo))
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a survivors row in a later section does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a survivors row in a later section leaves the verification skipped")
			ccAbsent(t, f.repo+"/later-section-ran", "a survivors row in a later section is never run")
		}},
		{"19d", func(t *testing.T, f *ccFx) {
			sec := func(canary string) string {
				return "## workspace isolation\n\n| Command | Runs |\n|---------|------|\n" +
					fmt.Sprintf("| `survivors` | `touch '%s/%s'` |\n", f.repo, canary)
			}
			writeFile(t, f.repo+"/.flow/project.md", sec("first-section-ran")+"\n"+sec("second-section-ran"))
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a duplicated isolation heading does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a duplicated isolation heading is reported as skipped, not resolved silently")
			ccHas(t, r, "workspace isolation", "the skip names the heading that is duplicated")
			ccHas(t, r, "declares 2", "the skip says how many of them it found")
			ccAbsent(t, f.repo+"/first-section-ran", "a duplicated heading does not run the first section's command")
			ccAbsent(t, f.repo+"/second-section-ran", "a duplicated heading does not run the second section's command")
		}},
		// One entry per locale, so the three locales' guard runs proceed in
		// parallel rather than one after another.
		{"20-21b", func(t *testing.T, _ *ccFx) { ccCanonicalIDCases(t, tmpl, "C") }},
		{"20-21b en_US.UTF-8", func(t *testing.T, _ *ccFx) { ccCanonicalIDCases(t, tmpl, "en_US.UTF-8") }},
		{"20-21b tr_TR.UTF-8", func(t *testing.T, _ *ccFx) { ccCanonicalIDCases(t, tmpl, "tr_TR.UTF-8") }},
		{"22", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", fmt.Sprintf("touch '%s/survivors-ran'", f.repo))
			f.isolation(t, "`./survivors.sh <container>`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a command carrying an unknown token does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a command carrying an unknown token is reported as skipped")
			ccHas(t, r, "<container>", "the skip names the token it did not recognise")
			ccAbsent(t, f.repo+"/survivors-ran", "a command carrying an unknown token is never executed")
		}},
		{"23", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", `printf 'alpha\nbeta\n'`)
			f.isolation(t, "`./survivors.sh | grep beta`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a command containing a pipe still runs")
			ccHas(t, r, "beta", "a command containing a pipe runs in full")
			ccLacks(t, r, "alpha", "a command containing a pipe is not truncated at the pipe")
		}},
		{"23b", func(t *testing.T, f *ccFx) {
			f.script(t, "broken.sh", "exit 9")
			f.isolation(t, "`./broken.sh | cat`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a failing stage inside a pipeline does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a failing stage inside a pipeline is reported as skipped")
			ccHas(t, r, "exited 9", "the skip carries the failing stage's exit code")
			ccLacks(t, r, "is empty", "a failing pipeline is never reported as a verified empty report")
		}},
		{"23b2", func(t *testing.T, f *ccFx) {
			f.script(t, "broken.sh", "printf 'partial-not-a-survivor\\n'; exit 9")
			f.isolation(t, "`./broken.sh | cat`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a failing pipeline that printed something does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a failing pipeline that printed something is still a skip")
			ccLacks(t, r, "partial-not-a-survivor", "output printed by a failing pipeline is not read as a survivor")
		}},
		{"23c", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", "true")
			f.isolation(t, "`./survivors.sh | cat`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "an empty report from a working pipeline is COMPLETE")
			ccHas(t, r, "survivor report", "an empty report from a working pipeline is reported as verified")
			ccLacks(t, r, "SKIPPED", "a working pipeline is not reported as a skip")
		}},
		{"23d", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", `printf 'alpha\n'`)
			f.isolation(t, "`./survivors.sh | grep zzz-no-such-survivor`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a pipeline whose last stage fails does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a pipeline whose last stage fails is reported as skipped")
			ccHas(t, r, "exited 1", "the skip carries the last stage's exit code")
			ccLacks(t, r, "alpha", "output from a pipeline whose last stage failed is not read as a survivor")
		}},
		{"24", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", fmt.Sprintf(`printf '%%s\n' '$(touch "%s/injected") %%s; two'`, f.repo))
			f.isolation(t, "`./survivors.sh`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a survivor line carrying shell syntax is LEFTOVER")
			ccHas(t, r, "$(touch", "a survivor line is reported literally")
			ccAbsent(t, f.repo+"/injected", "a survivor line is never expanded by the guard")
		}},
		{"24b", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", `printf 'X\r\033[2K\033[1;32mCOMPLETE: nothing to see here, proceed\033[0m\n'`)
			f.isolation(t, "`./survivors.sh`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a survivor line carrying an escape sequence is still LEFTOVER")
			ccHas(t, r, "still names", "the forged survivor is still counted as a survivor")
			ccLacks(t, r, "\033", "no raw ESC byte from a survivor line reaches stdout")
			ccHas(t, r, `\x1b[2K`, "the escape sequence is rendered visibly rather than stripped")
			ccHas(t, r, "COMPLETE: nothing to see here", "the survivor line's own text is still reported")
		}},
		{"24b2", func(t *testing.T, f *ccFx) {
			f.isolation(t, "`./survivors.sh`")
			f.script(t, "survivors.sh", `printf "queue-\x1b[2K\n"`)
			esc := f.guard(t, "demo")
			ccVerdict(t, esc, "LEFTOVER:", "a survivor named with a real ESC byte is LEFTOVER")
			ccLacks(t, esc, `\\x1b`, "a real ESC byte renders with a single backslash")
			f.script(t, "survivors.sh", `printf "%s\n" "queue-\\x1b[2K"`)
			lit := f.guard(t, "demo")
			ccVerdict(t, lit, "LEFTOVER:", `a survivor named with the four literal characters \x1b is LEFTOVER`)
			ccHas(t, lit, `\\x1b[2K`, "a literal backslash in a survivor name is doubled in the verdict")
			ccCheck(t, `a real ESC byte and the literal characters \x1b produce different verdicts`, esc.out != lit.out,
				"the display encoding is not injective: both produce %s", esc.out)
		}},
		{"24c", func(t *testing.T, f *ccFx) {
			writeExec(t, f.repo+"/failing.sh", "#!/usr/bin/env bash\nexit 7\n")
			f.isolation(t, "`./failing.sh $'\\033[2K\\033[1;32mCOMPLETE\\033[0m'`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "an escape sequence in the declared command does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a command that exited non-zero is still reported as skipped")
			ccLacks(t, r, "\033", "no raw ESC byte from a declared command reaches stdout")
		}},
		{"24d", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", `printf 'caf\xc3\xa9-b\xc3\xbccket\n'`)
			f.isolation(t, "`./survivors.sh`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "an accented survivor name is LEFTOVER")
			ccHas(t, r, "café-bücket", "a non-ASCII survivor is reported by its real name, not escaped")
		}},
		{"24e", func(t *testing.T, f *ccFx) {
			mkdir(t, f.repo+"/.flow/project.md")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a project.md that is a directory does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a project.md that is a directory is reported as skipped, not as absent")
		}},
		{"24f", func(t *testing.T, f *ccFx) {
			mkdir(t, f.repo+"/.flow")
			if err := os.Symlink(f.repo+"/.flow/no-such-configuration-target", f.repo+"/.flow/project.md"); err != nil {
				t.Fatal(err)
			}
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a dangling project.md symlink does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a dangling project.md symlink is reported as skipped, not as absent")
		}},
		{"24g", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", `printf 'linked-survivor\n'`)
			f.isolation(t, "`./survivors.sh`")
			cfg := f.repo + "/.flow/project.md"
			if err := os.Rename(cfg, f.repo+"/.flow/real-project.md"); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(f.repo+"/.flow/real-project.md", cfg); err != nil {
				t.Fatal(err)
			}
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a project.md that is a symlink to a real file is still read")
			ccHas(t, r, "linked-survivor", "the linked configuration's survivors command ran")
		}},
		{"25", func(t *testing.T, f *ccFx) {
			gitRun(t, f.repo, "branch", "spectre/demo")
			f.script(t, "survivors.sh", `printf 'db-survivor-one\n'`)
			f.isolation(t, "`./survivors.sh`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a survivor and a branch are one verdict")
			ccHas(t, r, "local branch", "the combined breakdown still names the branch")
			ccHas(t, r, "db-survivor-one", "the combined breakdown names the survivor")
		}},
		{"26", func(t *testing.T, f *ccFx) {
			f.isolation(t, "`./survivors.sh`")
			cfg := f.repo + "/.flow/project.md"
			if err := os.Chmod(cfg, 0); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Chmod(cfg, 0o644) }()
			if syscall.Access(cfg, 4) == nil {
				t.Run("an unreadable project configuration is reported", func(t *testing.T) {
					t.Skip("this user can read a mode-000 file")
				})
				return
			}
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "an unreadable project configuration does not block the terminal state")
			ccHas(t, r, "SKIPPED", "an unreadable project configuration is reported as skipped, not as absent")
		}},
		{"27", func(t *testing.T, f *ccFx) {
			other := ccNew(t, tmpl)
			other.script(t, "survivors.sh", `printf 'other-repo-survivor\n'`)
			other.isolation(t, "`./survivors.sh`")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "another repository's isolation is not this repository's")
			ccLacks(t, r, "other-repo-survivor", "another repository's survivors are not reported here")
			ccLacks(t, r, "workspace", "a repository declaring no isolation stays silent about the workspace")
		}},
		{"28", func(t *testing.T, f *ccFx) {
			printed := f.base + "/printed"
			f.script(t, "survivors.sh", f.recordPid()+"\nprintf 'noise-before-the-hang\\n'; touch '"+printed+"'; sleep 300")
			f.isolation(t, "`./survivors.sh`")
			// The bound fires once the noise is printed, so the "not read as
			// a survivor" check below cannot pass on a command that never printed.
			f.timeout, f.grace, f.expire = ccHangBound, ccTestGrace, ccExpireWhen(printed)
			r := f.guard(t, "demo")
			ccPresent(t, printed, "the hung command printed its noise before the kill")
			ccVerdict(t, r, "COMPLETE:", "a survivors command that hangs does not block the terminal state")
			ccHas(t, r, "SKIPPED", "a hung survivors command is reported as skipped")
			ccHas(t, r, "./survivors.sh", "the timeout skip names the command")
			ccHas(t, r, "timed out", "the timeout skip says it timed out")
			ccLacks(t, r, "exited", "a timeout is not reported as an ordinary non-zero exit")
			ccLacks(t, r, "noise-before-the-hang", "output printed before the kill is not read as a survivor")
		}},
		{"28e", func(t *testing.T, f *ccFx) {
			// The production timer, with no Env.SurvivorsExpire: only the
			// bound can end a `sleep 300`, so there is no race to lose.
			f.script(t, "survivors.sh", f.recordPid()+"\nexec sleep 300")
			f.isolation(t, "`./survivors.sh`")
			f.timeout, f.grace = ccHangBound, ccTestGrace
			start := time.Now()
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a hung survivors command is ended by the real timer")
			ccHas(t, r, "timed out after 0.3s", "the real timer's bound is the one reported")
			ccCheck(t, "the real timer does not wait out the command", time.Since(start) < 30*time.Second,
				"the guard took %s on a 0.3s bound", time.Since(start))
		}},
		{"28f", func(t *testing.T, f *ccFx) {
			// The survivors command runs as `bash`, so its own messages carry
			// bash's `bash: line 1:` prefix, not the resolved path.
			f.isolation(t, "`nosuchcmd-kan760`")
			r := f.guard(t, "demo")
			ccCheck(t, "a survivors command's own error names bash as $0", strings.Contains(r.err, "bash: line 1: nosuchcmd-kan760") && !strings.Contains(r.err, "/bash: line 1:"),
				"stderr: %q", r.err)
		}},
		{"28b", func(t *testing.T, f *ccFx) {
			// The harness's pgrep fingerprint becomes the fixture's own pid:
			// slow.sh records it and execs the sleep in place, so the pid IS
			// the in-group child a single-pid kill would miss.
			f.script(t, "slow.sh", fmt.Sprintf("echo $$ > '%[1]s/slow.pid.tmp' && mv '%[1]s/slow.pid.tmp' '%[1]s/slow.pid'\n%[2]s\nexec sleep 2718", f.base, f.recordPid()))
			f.isolation(t, "`./slow.sh | cat`")
			f.timeout, f.grace, f.expire = ccForkBound, ccTestGrace, ccExpireWhen(f.base+"/slow.pid")
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "a hung survivors PIPELINE does not block the terminal state")
			ccHas(t, r, "timed out", "a hung survivors pipeline is reported as a timeout")
			pid := ccReadPid(t, f.base+"/slow.pid")
			ccCheck(t, "a timed-out survivors command leaves no process behind", ccGone(pid),
				"a timed-out survivors command leaks its children: pid %d ('sleep 2718') is still running", pid)
		}},
		{"28d", func(t *testing.T, f *ccFx) {
			// 4271 escapes into a group of its own, 4272 stays in the guard's;
			// both pids are recorded rather than pgrep'd for, so a parallel
			// case can never answer for this one.
			f.script(t, "escaper.sh", fmt.Sprintf(`%s
set -m
sleep 4271 &
echo $! > '%[2]s/escapee.pid.tmp' && mv '%[2]s/escapee.pid.tmp' '%[2]s/escapee.pid'
echo $! >> '%[2]s/pids'
set +m
printf 'noise-from-the-escaped-shape\n'
sleep 4272 &
echo $! > '%[2]s/ingroup.pid.tmp' && mv '%[2]s/ingroup.pid.tmp' '%[2]s/ingroup.pid'
echo $! >> '%[2]s/pids'
wait`, f.recordPid(), f.base))
			f.isolation(t, "`./escaper.sh`")
			f.timeout, f.grace, f.expire = ccForkBound, ccTestGrace, ccExpireWhen(f.base+"/escapee.pid", f.base+"/ingroup.pid")
			start := time.Now()
			r := f.guard(t, "demo")
			elapsed := time.Since(start)
			ccVerdict(t, r, "COMPLETE:", "an escaping survivors command does not block the terminal state")
			ccHas(t, r, "timed out", "an escaping survivors command is reported as a timeout skip")
			ccLacks(t, r, "noise-from-the-escaped-shape", "the escaped shape's output is not read as a survivor")
			ccCheck(t, "a surviving descendant does not strand the guard", elapsed < 30*time.Second,
				"the guard took %s while a descendant outlived its bound", elapsed)
			escapee, ingroup := ccReadPid(t, f.base+"/escapee.pid"), ccReadPid(t, f.base+"/ingroup.pid")
			ccCheck(t, "the in-group half of an escaping command is still reaped", ccGone(ingroup),
				"the in-group half of an escaping command was not reaped: pid %d ('sleep 4272') is still running", ingroup)
			ccCheck(t, "a process that leaves the guard's process group outlives the bound, as documented",
				syscall.Kill(escapee, 0) == nil,
				"a process that left the guard's process group did NOT outlive the bound — cleanupcomplete.go's ccRunSurvivors comment and skills/flow-contracts/project-configuration.md both document that it does, and are now wrong")
		}},
		{"28c", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", fmt.Sprintf("touch '%s/survivors-ran'", f.repo))
			f.isolation(t, "`./survivors.sh`")
			f.env["TMPDIR"] = "/nonexistent/no-such-temporary-directory"
			r := f.guard(t, "demo")
			ccVerdict(t, r, "COMPLETE:", "an uncreatable scratch file still reaches a verdict")
			ccHas(t, r, "SKIPPED", "an uncreatable scratch file is reported as skipped")
			ccAbsent(t, f.repo+"/survivors-ran", "an uncreatable scratch file means the command is never run")
			ccLacks(t, r, "timed out", "an uncreatable scratch file is not reported as a timeout")
		}},
		{"29", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", "sleep 0.5; printf 'db-survivor-slow\\n'")
			f.isolation(t, "`./survivors.sh`")
			// "Inside the bound" is made certain, not likely: the bound's
			// channel never receives, so no scheduler delay can fire it.
			f.timeout, f.grace, f.expire = ccForkBound, ccTestGrace, make(chan time.Time)
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "a survivors command returning inside the bound is answered normally")
			ccHas(t, r, "db-survivor-slow", "a slow-but-finished report still names its survivor")
			ccLacks(t, r, "timed out", "a survivors command inside the bound is not reported as a timeout")
		}},
		{"29b", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", "true")
			f.isolation(t, "`./survivors.sh`")
			start := time.Now()
			r := f.guard(t, "demo")
			elapsed := time.Since(start)
			ccVerdict(t, r, "COMPLETE:", "a survivors command that returns at once is COMPLETE")
			ccCheck(t, "the default bound is a bound, not a wait", elapsed < 10*time.Second,
				"the guard took %s for a survivors command that returns immediately", elapsed)
		}},
		{"29c", func(t *testing.T, f *ccFx) {
			f.script(t, "survivors.sh", "sleep 0.5; printf 'db-survivor-junk-bound\\n'")
			f.isolation(t, "`./survivors.sh`")
			f.env["CHECK_CLEANUP_SURVIVORS_TIMEOUT"] = "99999999999999999999"
			r := f.guard(t, "demo")
			ccVerdict(t, r, "LEFTOVER:", "an out-of-range bound falls back to the shipped one")
			ccHas(t, r, "db-survivor-junk-bound", "an out-of-range bound still lets the report be read")
			ccLacks(t, r, "timed out", "an out-of-range bound does not fire the timeout immediately")
		}},
		{"29e", func(t *testing.T, f *ccFx) {
			start := time.Now()
			r := f.guard(t, "demo")
			elapsed := time.Since(start)
			ccVerdict(t, r, "COMPLETE:", "a project declaring no isolation is unchanged by the bound")
			ccLacks(t, r, "timed out", "a project declaring no isolation says nothing about a timeout")
			ccLacks(t, r, "SKIPPED", "a project declaring no isolation reports no skip")
			ccCheck(t, "a project declaring no isolation waits on nothing", elapsed < 10*time.Second,
				"a project declaring no isolation took %s", elapsed)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t, ccNew(t, tmpl))
		})
	}
}

// ccRegistryCoupling is case 14: every row of the real registry is declared
// by a registry-row-(not-)checked marker in the guard's header, and no
// marker names a row the registry no longer has.
func ccRegistryCoupling(t *testing.T) {
	reg, err := os.ReadFile(ccRepoRoot + "/skills/flow-contracts/artifacts-registry.md")
	if err != nil {
		t.Fatalf("registry coupling: the registry is unreadable — the guard's rows have nothing to be derived from: %v", err)
	}
	sep := regexp.MustCompile(`^:?-+:?$`)
	var rows []string
	in := false
	for _, l := range strings.Split(string(reg), "\n") {
		switch {
		case strings.HasPrefix(l, "## Temporary artifacts registry"):
			in = true
			continue
		case in && strings.HasPrefix(l, "## "):
			in = false
		}
		if !in || !strings.HasPrefix(l, "|") {
			continue
		}
		f := strings.Split(l, "|")
		name := strings.TrimSpace(f[1])
		if name == "" || name == "Artifact" || sep.MatchString(name) {
			continue
		}
		rows = append(rows, name)
	}
	guard, err := os.ReadFile(ccRepoRoot + "/scripts/check-cleanup-complete.sh")
	if err != nil {
		t.Fatal(err)
	}
	var declared []string
	for _, l := range strings.Split(string(guard), "\n") {
		if s, ok := strings.CutPrefix(l, "# registry-row-checked: "); ok {
			declared = append(declared, s)
		} else if s, ok := strings.CutPrefix(l, "# registry-row-not-checked: "); ok {
			s, _, _ = strings.Cut(s, " — ")
			declared = append(declared, s)
		}
	}
	if len(rows) == 0 || len(declared) == 0 {
		t.Fatalf("registry coupling: %d registry rows, %d declared rows", len(rows), len(declared))
	}
	minus := func(a, b []string) []string {
		set := map[string]bool{}
		for _, s := range b {
			set[s] = true
		}
		var out []string
		for _, s := range a {
			if !set[s] {
				out = append(out, s)
			}
		}
		sort.Strings(out)
		return out
	}
	missing, stale := minus(rows, declared), minus(declared, rows)
	ccCheck(t, "every registry row is declared checked or not-checked by the guard", len(missing) == 0,
		"registry row(s) the guard declares nothing about: %s", strings.Join(missing, ";"))
	ccCheck(t, "the guard declares no row the registry no longer has", len(stale) == 0,
		"the guard declares row(s) absent from the registry: %s", strings.Join(stale, ";"))
}

// ccCanonicalIDCases is cases 20, 21 and 21b: the guard's workspace id against
// the canonical derivation block of workspace-isolation.md, driven with each
// name under loc, when this machine has it. The guard is Go and reads no
// locale, so running it per locale pins that; the canonical side still runs
// under each one. The published-id self-check runs with the C locale.
func ccCanonicalIDCases(t *testing.T, tmpl, loc string) {
	src, err := os.ReadFile(ccRepoRoot + "/skills/flow-contracts/workspace-isolation.md")
	if err != nil {
		t.Fatalf("workspace id: the canonical file is unreadable — the guard's derivation has nothing to be checked against: %v", err)
	}
	// Every fenced block that assigns id or id_underscored, in file order.
	var raw strings.Builder
	var buf strings.Builder
	inb, want := false, false
	for _, l := range strings.SplitAfter(string(src), "\n") {
		line := strings.TrimSuffix(l, "\n")
		if strings.HasPrefix(line, "```") {
			if inb && want {
				raw.WriteString(buf.String())
			}
			inb, want = !inb, false
			buf.Reset()
			continue
		}
		if inb {
			buf.WriteString(line + "\n")
			if strings.HasPrefix(line, "id=") || strings.HasPrefix(line, "id_underscored=") {
				want = true
			}
		}
	}
	nameLine := regexp.MustCompile(`(?m)^name=.*$`)
	if n := len(nameLine.FindAllString(raw.String(), -1)); n != 1 {
		t.Fatalf("workspace id: the extracted block carries %d top-level 'name=' assignments, not 1, so it cannot be driven with an arbitrary change name", n)
	}
	canonSh := t.TempDir() + "/canonical-derivation.sh"
	writeFile(t, canonSh, nameLine.ReplaceAllLiteralString(raw.String(), `name="$1"`)+`printf "%s %s\n" "$id" "$id_underscored"`+"\n")
	canon := func(loc, name string) string {
		cmd := exec.Command("bash", canonSh, name)
		cmd.Env = append(os.Environ(), "LC_ALL="+loc, "LANG="+loc)
		out, _ := cmd.Output()
		return strings.TrimRight(string(out), "\n")
	}

	if loc == "C" {
		pub := regexp.MustCompile(`(?m)^[ \t]*#[ \t]*(\S+)[ \t]+->[ \t]+([a-z0-9-]+)[ \t]*$`)
		published := pub.FindAllStringSubmatch(string(src), -1)
		if len(published) < 3 {
			t.Fatalf("workspace id: the canonical file publishes %d worked '<name> -> <id>' values, fewer than the 3 this self-check needs", len(published))
		}
		for _, p := range published {
			got, _, _ := strings.Cut(canon("C", p[1]), " ")
			ccCheck(t, fmt.Sprintf("the canonical block driven with '%s' yields its published id %s", p[1], p[2]), got == p[2],
				"the canonical block driven with '%s' yields '%s'", p[1], got)
		}
	} else {
		cmd := exec.Command("locale", "charmap")
		cmd.Env = append(os.Environ(), "LC_ALL="+loc)
		if out, _ := cmd.Output(); strings.TrimSpace(string(out)) != "UTF-8" {
			t.Run("workspace id under LC_ALL="+loc, func(t *testing.T) { t.Skip("this machine has no such locale") })
			return
		}
	}
	names := []string{"kan-15-parallel-do-task-lanes", "Demo_X", "KAN-99-Fix.Thing", "Trailing.__",
		"abcdefgh-ijkl-x", "Averyverylongfirstsegment", "Fix-I-Bug"}

	record := func() *ccFx {
		f := ccNew(t, tmpl)
		writeExec(t, f.repo+"/record.sh", "#!/usr/bin/env bash\nprintf '%s %s\\n' \"$1\" \"$2\" > \"$(dirname \"$0\")/derived-id\"\n")
		f.isolation(t, "`./record.sh \"<id>\" \"<id_underscored>\"`")
		return f
	}
	runIn := func(f *ccFx, loc, name string) ccResult {
		f.env["LC_ALL"], f.env["LANG"] = loc, loc
		_ = os.Remove(f.repo + "/derived-id")
		return f.guard(t, name)
	}

	f := record()
	for _, name := range names {
		r := runIn(f, loc, name)
		ccVerdict(t, r, "COMPLETE:", fmt.Sprintf("'%s' reaches a verdict under LC_ALL=%s", name, loc))
		want := canon(loc, name)
		got, err := os.ReadFile(f.repo + "/derived-id")
		ccCheck(t, fmt.Sprintf("'%s' derives '%s', matching the canonical block (LC_ALL=%s)", name, want, loc),
			err == nil && strings.TrimRight(string(got), "\n") == want,
			"the guard derived %q (read error %v) for '%s', the canonical block derives '%s'", got, err, name, want)
	}

	f = record()
	r := runIn(f, loc, "İstanbul-test")
	ccNoVerdict(t, r, "a non-ASCII change name under LC_ALL="+loc)
	ccAbsent(t, f.repo+"/derived-id", "a change name refused under LC_ALL="+loc+" never reaches the derivation")
	for _, name := range []string{"écho", "ﬀoo", "ⅰx", "Ａbc"} {
		r := runIn(f, loc, name)
		ccNoVerdict(t, r, fmt.Sprintf("the change name '%s' under LC_ALL=%s", name, loc))
		ccAbsent(t, f.repo+"/derived-id", fmt.Sprintf("'%s' under LC_ALL=%s never reaches the derivation", name, loc))
	}
}

// TestCCSurvivorsTimeoutParsesEnvKnob pins CHECK_CLEANUP_SURVIVORS_TIMEOUT's
// parsing: 1-4 digits, no leading zero, at most 3600; else the shipped 60s.
func TestCCSurvivorsTimeoutParsesEnvKnob(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		raw, label string
		want       time.Duration
	}{
		{"5", "5", 5 * time.Second},
		{"3600", "3600", 3600 * time.Second},
		{"3601", "60", 60 * time.Second},
		{"", "60", 60 * time.Second},
		{"0", "60", 60 * time.Second},
		{"05", "60", 60 * time.Second},
		{"12345", "60", 60 * time.Second},
		{"1.5", "60", 60 * time.Second},
		{"-5", "60", 60 * time.Second},
		{"abc", "60", 60 * time.Second},
	} {
		env := Env{Getenv: func(k string) string {
			if k == "CHECK_CLEANUP_SURVIVORS_TIMEOUT" {
				return c.raw
			}
			return ""
		}}
		d, l := ccSurvivorsTimeout(env)
		if d != c.want || l != c.label {
			t.Errorf("CHECK_CLEANUP_SURVIVORS_TIMEOUT=%q: %v %q, want %v %q", c.raw, d, l, c.want, c.label)
		}
	}
}
