package guard

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Every case of scripts/test-check-worktree-processes.sh at c9b8fde8, one
// subtest per `ok:` label. Real child processes run with their working
// directory inside a t.TempDir() worktree and are killed in t.Cleanup; the
// guard reads them through the real lsof. Every path an assertion compares
// is physical (filepath.EvalSymlinks, the harness's `pwd -P`): lsof reports
// resolved paths, so on macOS a TMPDIR fixture is /private/var/... to lsof
// and /var/... to t.TempDir(), and comparing the latter would report CLEAR
// for every held fixture and prove nothing. Each case hands the guard the
// t.TempDir() spelling, as a caller would, and asserts against the physical
// one. stdout and stderr are read apart: a refusal writes its reason to
// stderr and must leave stdout empty.
//
// The harness's case 7 (KAN-197) re-ran the whole suite against a sed-edited
// copy whose at-or-under decision was forced false. Here the decision is the
// predicate worktreeProcesses takes, so the mutant is that predicate replaced
// by one that never matches, and case 7 requires case 2's HELD assertion to
// fail against it.

// wpRun is one in-process run of the guard with the at-or-under decision
// atOrUnder in place of the real one.
func wpRun(args []string, env Env, atOrUnder func(path, root string) bool) guardResult {
	var all, o, e bytes.Buffer
	rc := worktreeProcesses(args, env, io.MultiWriter(&o, &all), io.MultiWriter(&e, &all), atOrUnder)
	return guardResult{rc, strings.TrimRight(all.String(), "\n"), o.String(), e.String()}
}

func wpEnv() Env { return Env{Getenv: os.Getenv} }

// wpStartIn is the harness's start_in followed by its await_cwd: a live
// `sleep` whose working directory is dir, killed and reaped in t.Cleanup,
// returned once lsof reports that pid's cwd as dir. Polling for the state the
// case needs is deterministic where a fixed sleep is a guess; the bound is 5
// seconds, after which the case fails rather than continue on a premise that
// never became true.
func wpStartIn(t *testing.T, dir string) int {
	t.Helper()
	cmd := exec.Command("sleep", "300")
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	pid := cmd.Process.Pid
	for range 100 {
		out, _ := exec.Command("lsof", "-a", "-d", "cwd", "-p", strconv.Itoa(pid), "-Fn").Output()
		for _, l := range strings.Split(string(out), "\n") {
			if l == "n"+dir {
				return pid
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("lsof never reported pid %d's cwd as %s — the test cannot see its own process", pid, dir)
	return 0
}

// wpWorktree is the harness's new_worktree: the path as t.TempDir() spells
// it (what a caller hands the guard) and its physical form (what lsof and
// the verdict line spell).
func wpWorktree(t *testing.T) (wt, phys string) {
	t.Helper()
	wt = t.TempDir()
	phys, err := filepath.EvalSymlinks(wt)
	if err != nil {
		t.Fatal(err)
	}
	return wt, phys
}

type wpCheck struct {
	label string
	fail  func(r guardResult) string // "" when the assertion holds
}

// wpClear is assert_clear: exit 0 and exactly the CLEAR line naming phys.
func wpClear(label, phys string) wpCheck {
	return wpCheck{label, func(r guardResult) string {
		if r.rc != 0 || r.stdout != "CLEAR: "+phys+"\n" {
			return "expected exit 0 and 'CLEAR: " + phys + "', got rc=" + strconv.Itoa(r.rc) + " out=" + r.out
		}
		return ""
	}}
}

// wpHeld is assert_held: exit 0 and exactly the HELD line naming phys with
// pid paired with cwd — the only process the case started under phys.
func wpHeld(label, phys string, pid int, cwd string) []wpCheck {
	want := "HELD: " + phys + " — " + strconv.Itoa(pid) + " " + cwd + "\n"
	return []wpCheck{
		{label + ": reports HELD for the worktree", func(r guardResult) string {
			if r.rc != 0 || !strings.HasPrefix(r.stdout, "HELD: "+phys+" ") {
				return "expected exit 0 and a HELD verdict for " + phys + ", got rc=" + strconv.Itoa(r.rc) + " out=" + r.out
			}
			return ""
		}},
		{label + ": names pid " + strconv.Itoa(pid) + " with its working directory " + cwd, func(r guardResult) string {
			if r.rc != 0 || r.stdout != want {
				return "expected exit 0 and " + strconv.Quote(want) + ", got rc=" + strconv.Itoa(r.rc) + " stdout=" + strconv.Quote(r.stdout)
			}
			return ""
		}},
	}
}

// wpCannot is assert_cannot_answer: exit 2, an empty stdout, no verdict
// token, and a named reason on stderr — exactly want when want is non-empty.
func wpCannot(label, want string) []wpCheck {
	return []wpCheck{
		{label + ": exits 2", func(r guardResult) string {
			if r.rc != 2 {
				return "expected exit 2, got rc=" + strconv.Itoa(r.rc) + " out=" + r.out
			}
			return ""
		}},
		{label + ": writes nothing to stdout", func(r guardResult) string {
			if r.stdout != "" {
				return "emitted something on stdout: " + r.stdout
			}
			return ""
		}},
		{label + ": prints no verdict token", func(r guardResult) string {
			if strings.Contains(r.stdout, "CLEAR") || strings.Contains(r.stdout, "HELD") {
				return "an inability printed a verdict token: " + r.stdout
			}
			return ""
		}},
		{label + ": names the failure on stderr", func(r guardResult) string {
			if !strings.HasPrefix(r.err, "check-worktree-processes: ") || (want != "" && r.err != want) {
				return "stderr = " + strconv.Quote(r.err) + ", want " + strconv.Quote(want)
			}
			return ""
		}},
	}
}

func wpRunChecks(t *testing.T, r guardResult, checks []wpCheck) {
	t.Helper()
	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if msg := c.fail(r); msg != "" {
				t.Fatal(msg)
			}
		})
	}
}

func TestCheckWorktreeProcesses(t *testing.T) {
	t.Parallel()
	if Registry["check-worktree-processes"] == nil {
		t.Fatal("check-worktree-processes is not registered")
	}
	run := func(args ...string) guardResult { return runGuard("check-worktree-processes", args, wpEnv()) }

	// Case 1: the ordinary pre-removal state — nothing was ever started here.
	wt, phys := wpWorktree(t)
	wpRunChecks(t, run(wt), []wpCheck{wpClear("case 1: a worktree with no process in it is CLEAR", phys)})

	// Case 2: a process whose cwd IS the worktree. This is the incident.
	wt2, phys2 := wpWorktree(t)
	pid2 := wpStartIn(t, phys2)
	r2 := run(wt2)
	case2 := wpHeld("case 2: a process whose cwd is the worktree is HELD", phys2, pid2, phys2)
	wpRunChecks(t, r2, append(case2, wpCheck{"case 2: a reached verdict exits 0", func(r guardResult) string {
		if r.rc != 0 {
			return "expected exit 0 for a reached verdict, got rc=" + strconv.Itoa(r.rc)
		}
		return ""
	}}))

	// Case 3: a process in a SUBDIRECTORY of the worktree — a backend started
	// from its own module root.
	wt3, phys3 := wpWorktree(t)
	mkdir(t, phys3+"/stats")
	pid3 := wpStartIn(t, phys3+"/stats")
	wpRunChecks(t, run(wt3), wpHeld("case 3: a process under the worktree is HELD", phys3, pid3, phys3+"/stats"))

	// Case 4: siblings sharing a name prefix, as spectre-kan-1 and
	// spectre-kan-12 do; a bare string-prefix test reports the first HELD.
	sib, sibPhys := wpWorktree(t)
	mkdir(t, sibPhys+"/spectre-kan-1")
	mkdir(t, sibPhys+"/spectre-kan-12")
	pid4 := wpStartIn(t, sibPhys+"/spectre-kan-12")
	wpRunChecks(t, run(sib+"/spectre-kan-1"), []wpCheck{
		wpClear("case 4: a sibling sharing a name prefix is not HELD", sibPhys+"/spectre-kan-1"),
		{"case 4: the sibling's pid is absent from the verdict", func(r guardResult) string {
			if strings.Contains(r.out, strconv.Itoa(pid4)) {
				return "the sibling's pid " + strconv.Itoa(pid4) + " reached the verdict: " + r.out
			}
			return ""
		}},
	})

	// Cases 5a–5c: inputs it cannot answer about, never failing open.
	work := t.TempDir()
	missing := work + "/no-such-worktree"
	wpRunChecks(t, run(missing), wpCannot("case 5a: a worktree path that does not exist",
		"check-worktree-processes: not a directory, so it cannot be scanned: "+missing+"\n"))

	writeFile(t, work+"/regular-file", "not a directory\n")
	wpRunChecks(t, run(work+"/regular-file"), wpCannot("case 5b: a worktree path that is a regular file",
		"check-worktree-processes: not a directory, so it cannot be scanned: "+work+"/regular-file\n"))

	// Root ignores the mode bits, so case 5c skips there rather than pass
	// vacuously.
	unreadable := t.TempDir() + "/unreadable"
	mkdir(t, unreadable)
	if err := os.Chmod(unreadable, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o755) })
	r5c := run(unreadable)
	for _, c := range wpCannot("case 5c: an unreadable worktree directory",
		"check-worktree-processes: cannot resolve the worktree to a physical path — unreadable directory: "+unreadable+"\n") {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if os.Geteuid() == 0 {
				t.Skip("running as root, which ignores the mode bits")
			}
			if msg := c.fail(r5c); msg != "" {
				t.Fatal(msg)
			}
		})
	}

	// Case 6: lsof is absent from the PATH on Env — an inability, never an
	// empty result, and the message names the tool.
	noLsof := t.TempDir()
	wt6, _ := wpWorktree(t)
	r6 := runGuard("check-worktree-processes", []string{wt6}, Env{Getenv: func(k string) string {
		if k == "PATH" {
			return noLsof
		}
		return os.Getenv(k)
	}})
	wpRunChecks(t, r6, append(wpCannot("case 6: lsof is not on PATH",
		"check-worktree-processes: lsof is not on PATH — this guard has no other way to read the process table\n"),
		wpCheck{"case 6: the message names the missing tool", func(r guardResult) string {
			if !strings.Contains(r.err, "lsof") {
				return "the message does not name lsof: " + r.err
			}
			return ""
		}}))

	// The refusals the harness never named: the wrong number of arguments.
	usage := "check-worktree-processes: usage: check-worktree-processes.sh <worktree>\n"
	wpRunChecks(t, run(), wpCannot("no arguments", usage))
	wpRunChecks(t, run(wt, wt), wpCannot("two arguments", usage))
	wpRunChecks(t, run(""), wpCannot("an empty worktree argument",
		"check-worktree-processes: not a directory, so it cannot be scanned: \n"))

	// An lsof that ran and failed is an inability, never an empty CLEAR; its
	// exit status is read, not its (empty) output.
	stubEnv := func(body string) Env {
		bin := t.TempDir()
		writeExec(t, bin+"/lsof", body)
		return Env{Getenv: pathEnv(bin)}
	}
	wpRunChecks(t, runGuard("check-worktree-processes", []string{wt}, stubEnv("#!/bin/sh\nexit 3\n")),
		wpCannot("a failing lsof", "check-worktree-processes: lsof -d cwd -Fpn exited 3 — the process table could not be read, which is not the same as finding nothing\n"))

	// Several holders join with "; " in lsof's order, each `n` line paired
	// with the `p` line above it; `f` lines and a prefix sibling are ignored.
	table := "#!/bin/sh\nprintf 'p11\\nfcwd\\nn%s\\np22\\nfcwd\\nn%s/a\\np33\\nfcwd\\nn%s-2\\n' '" + phys + "' '" + phys + "' '" + phys + "'\n"
	wpRunChecks(t, runGuard("check-worktree-processes", []string{wt}, stubEnv(table)), []wpCheck{
		{"several holders are listed in lsof's order", func(r guardResult) string {
			if want := "HELD: " + phys + " — 11 " + phys + "; 22 " + phys + "/a\n"; r.rc != 0 || r.stdout != want {
				return "expected exit 0 and " + strconv.Quote(want) + ", got rc=" + strconv.Itoa(r.rc) + " out=" + r.out
			}
			return ""
		}},
	})

	// Case 7 (KAN-197): the guard's one decision forced false, so every
	// verdict becomes CLEAR, must fail case 2's HELD assertion.
	never := func(string, string) bool { return false }
	mutant := wpRun([]string{wt2}, wpEnv(), never)
	wpRunChecks(t, mutant, []wpCheck{
		{"case 7 (KAN-197 mutation): the mutation applied", func(guardResult) string {
			if wpAtOrUnder(phys2, phys2) == never(phys2, phys2) {
				return "the mutant decides a process in the worktree root as the guard does, so nothing was mutated"
			}
			return ""
		}},
		{"case 7 (KAN-197 mutation): a guard that always prints CLEAR fails this suite", func(r guardResult) string {
			for _, c := range case2 {
				if c.fail(r) != "" {
					return ""
				}
			}
			return "a guard that always prints CLEAR passed case 2, which therefore detects nothing: " + r.out
		}},
	})
}
