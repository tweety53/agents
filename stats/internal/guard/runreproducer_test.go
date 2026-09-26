package guard

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Every case of scripts/test-run-reproducer.sh at 0747740, one subtest per
// ok: label; a "(never executed)" or "(ran …)" ok: line is a nested
// subtest of the case it follows. The bound and grace are injected on Env as
// durations (Decision: inject-deadlines-in-process); every case that leaves
// a survivor still asserts the survivor is gone afterwards.
//
// The bound is a backstop, never the trigger: macOS spends ~0.2s on the
// first exec of a freshly written script (measured: 0.201s cold, 0.009s
// warm), so a tight wall-clock bound times out verdict cases under load.
// Every timeout case fires the bound through RUN_REPRODUCER_BOUND_FILE, on
// its fixture's own condition.
const rrTestBound, rrTestGrace = 30 * time.Second, 100 * time.Millisecond

// rrWorktree is make_worktree: a fresh worktree with scripts/ inside. Every
// process group a fixture records in <wt>/groups is SIGKILLed at cleanup, so
// a failed assertion never leaves a survivor running on the machine.
func rrWorktree(t *testing.T) string {
	t.Helper()
	wt := t.TempDir()
	mkdir(t, filepath.Join(wt, "scripts"))
	t.Cleanup(func() {
		b, _ := os.ReadFile(filepath.Join(wt, "groups"))
		for _, f := range strings.Fields(string(b)) {
			var pgid int
			if _, err := fmt.Sscan(f, &pgid); err == nil && pgid > 1 {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
		}
	})
	return wt
}

// rrFixture is the harness's fixture(): an executable bash script that drops
// a RAN marker before its body, so a refusal can prove nothing executed, and
// records its own process group (it leads one) for rrWorktree's cleanup --
// on the RAN line, so the harness's line arithmetic (case 28) holds. Both
// paths are named from the script's own (fixtureRoot), so every case with
// the same rel and body writes the same bytes and shares one assessed inode
// (writeExec).
// ponytail: the markers land beside whatever path was exec'd, so a guard that
// ran a copy of a refused reproducer would drop RAN beside the copy and "never
// executed" would pass. Absolute markers make every body unique: 20.7s
// package wall against 10.3s. Pin an absolute marker if a guard ever copies
// a reproducer before exec.
func rrFixture(t *testing.T, wt, rel, body string) {
	t.Helper()
	root := fixtureRoot(rel)
	writeExec(t, filepath.Join(wt, rel), fmt.Sprintf("#!/usr/bin/env bash\ntouch \"%sRAN\"; echo $$ >> \"%sgroups\"\n%s\n",
		root, root, body))
}

// rrEnv is the environment a case runs under: TMPDIR inside the test's own
// temporary directory, plus the case's own knobs.
func rrEnv(t *testing.T, kv ...string) Env {
	m := map[string]string{"TMPDIR": t.TempDir(), "PATH": os.Getenv("PATH")}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return Env{Getenv: func(k string) string { return m[k] }, Dir: t.TempDir(),
		ReproducerBound: rrTestBound, ReproducerGrace: rrTestGrace}
}

// rrRun runs the registered guard in-process, stdout and stderr in one
// buffer the way the harness's `2>&1` reads them.
func rrRun(t *testing.T, env Env, args ...string) (int, string) {
	t.Helper()
	fn := Registry["run-reproducer"]
	if fn == nil {
		t.Fatal("run-reproducer is not registered")
	}
	var out bytes.Buffer
	return fn(args, env, &out, &out), out.String()
}

func rrExpect(t *testing.T, got int, out string, want int, needle string) {
	t.Helper()
	if got != want {
		t.Fatalf("exit %d, want %d\n%s", got, want, out)
	}
	if !strings.Contains(out, needle) {
		t.Fatalf("output does not name %q:\n%s", needle, out)
	}
}

func rrNotRan(t *testing.T, wt string) {
	t.Run("never executed", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(wt, "RAN")); err == nil {
			t.Fatalf("the reproducer ran despite being refused (found %s/RAN)", wt)
		}
	})
}

func rrRan(t *testing.T, wt string) {
	t.Run("the reproducer ran before the verdict was read", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(wt, "RAN")); err != nil {
			t.Fatal("the ambiguity refusal answered without running the reproducer")
		}
	})
}

var rrSurvivorPid = regexp.MustCompile(`pid\(s\): ([0-9]+)`)

// rrGone waits for a pid (or, negated, a whole process group) to be gone:
// a SIGKILLed orphan is reaped by launchd asynchronously, so absence is
// awaited under a deadline, never assumed.
func rrGone(t *testing.T, target int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for syscall.Kill(target, 0) == nil {
		if time.Now().After(deadline) {
			_ = syscall.Kill(target, syscall.SIGKILL)
			t.Fatalf("process %d is still alive after the guard returned", target)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// rrSurvivorKilled is the survivor cases' shared assertion: exit 3 naming a
// surviving process by pid, and that pid gone once the guard has returned.
func rrSurvivorKilled(t *testing.T, got int, out, needle string) {
	t.Helper()
	rrExpect(t, got, out, 3, "surviving process")
	rrExpect(t, got, out, 3, needle)
	m := rrSurvivorPid.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("no surviving pid named:\n%s", out)
	}
	var pid int
	fmt.Sscan(m[1], &pid)
	rrGone(t, pid)
}

// rrMetachars reads REPRODUCER_METACHARS the way the bash guards bind it:
// by sourcing scripts/reproducer-metachars.sh.
func rrMetachars(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("bash", "-c", `source "$1" && printf '%s' "$REPRODUCER_METACHARS"`, "_",
		"../../../scripts/reproducer-metachars.sh").Output()
	if err != nil || len(out) == 0 {
		t.Fatalf("sourcing reproducer-metachars.sh: %v", err)
	}
	return string(out)
}

func TestMetacharsMatchBashSource(t *testing.T) {
	t.Parallel()
	if got, want := reproducerMetachars, rrMetachars(t); got != want {
		t.Fatalf("metachars.go = %q, scripts/reproducer-metachars.sh = %q", got, want)
	}
}

func TestRunReproducer(t *testing.T) {
	t.Parallel()
	// Simple verdict and refusal rows: fixture (if any), command line and
	// flags, expected exit and the substring the output must name.
	type row struct {
		label, rel, body string
		args             []string
		want             int
		needle           string
		notRan, ran      bool
	}
	rows := []row{
		{label: "case 1: a failing contained reproducer demonstrates the defect", rel: "scripts/fails.sh", body: "exit 7", args: []string{"scripts/fails.sh"}, want: 0},
		{label: "case 2: a passing contained reproducer is distinguished from a failing one", rel: "scripts/passes.sh", body: "exit 0", args: []string{"scripts/passes.sh"}, want: 1},
		{label: "case 3: an absolute path token is refused", args: []string{"/bin/echo hi"}, want: 2, needle: "absolute", notRan: true},
		{label: "case 4: a `..` path-token segment is refused", args: []string{"../../../etc/passwd"}, want: 2, needle: "'..'", notRan: true},
		{label: "case 5: a `..` argument segment is refused", rel: "scripts/ok.sh", body: "exit 0", args: []string{"scripts/ok.sh ../../../etc/passwd"}, want: 2, needle: "'..'", notRan: true},
		{label: "case 6 control: an ordinary command with no banned metacharacter is accepted", rel: "scripts/ok.sh", body: "exit 0", args: []string{"scripts/ok.sh"}, want: 1},
		{label: "case 7: a nonexistent path is refused", args: []string{"scripts/nonexistent.sh"}, want: 2, needle: "does not exist", notRan: true},
		{label: "case 12: a tab-separated command line still resolves the path token", rel: "scripts/ok.sh", body: "exit 0", args: []string{"scripts/ok.sh\t--strict"}, want: 1},
		{label: "case 19: a post-fix verdict identical to the pre-fix verdict (demonstrated/demonstrated) is refused", rel: "scripts/fails-again.sh", body: "exit 7", args: []string{"scripts/fails-again.sh", "--pre-fix-verdict", "demonstrated"}, want: 2, needle: "convention", ran: true},
		{label: "case 20: a post-fix verdict identical to the pre-fix verdict (not-demonstrated/not-demonstrated) is refused", rel: "scripts/passes-again.sh", body: "exit 0", args: []string{"scripts/passes-again.sh", "--pre-fix-verdict", "not-demonstrated"}, want: 2, needle: "convention", ran: true},
		{label: "case 21: a flipped verdict (demonstrated pre-fix, not-demonstrated post-fix) is the fix verified", rel: "scripts/now-passes.sh", body: "exit 0", args: []string{"scripts/now-passes.sh", "--pre-fix-verdict", "demonstrated"}, want: 1},
		{label: "case 22: a flipped verdict (not-demonstrated pre-fix, demonstrated post-fix) passes through", rel: "scripts/now-fails.sh", body: "exit 5", args: []string{"scripts/now-fails.sh", "--pre-fix-verdict", "not-demonstrated"}, want: 0},
		{label: "case 23.stray: a stray third argument is a usage failure", rel: "scripts/never-runs.sh", body: "exit 0", args: []string{"scripts/never-runs.sh", "extra-arg"}, want: 4, needle: "usage", notRan: true},
		{label: "case 23.: a missing --pre-fix-verdict value is a usage failure", rel: "scripts/never-runs.sh", body: "exit 0", args: []string{"scripts/never-runs.sh", "--pre-fix-verdict"}, want: 4, needle: "usage", notRan: true},
		{label: "case 24: a mutation reproducer exiting 0 demonstrates the defect", rel: "scripts/mutation-survives.sh", body: "# mutation-reproducer\nexit 0", args: []string{"scripts/mutation-survives.sh"}, want: 0},
		{label: "case 25: a mutation reproducer exiting non-zero is not demonstrated", rel: "scripts/mutation-caught.sh", body: "# mutation-reproducer\nexit 3", args: []string{"scripts/mutation-caught.sh"}, want: 1},
		{label: "case 26: a mutation reproducer with verdict demonstrated both sides is refused as ambiguous", rel: "scripts/mutation-still-survives.sh", body: "# mutation-reproducer\nexit 0", args: []string{"scripts/mutation-still-survives.sh", "--pre-fix-verdict", "demonstrated"}, want: 2, needle: "convention", ran: true},
		{label: "case 27: a mutation reproducer flipping demonstrated to not-demonstrated is the fix verified", rel: "scripts/mutation-now-caught.sh", body: "# mutation-reproducer\nexit 3", args: []string{"scripts/mutation-now-caught.sh", "--pre-fix-verdict", "demonstrated"}, want: 1},
		// fixture() prepends the shebang and RAN-marker lines, so N filler
		// lines put the marker at line N+3 (KAN-623).
		{label: "case 28.a: a marker beyond the first 10 lines declares nothing", rel: "scripts/late-marker.sh", body: strings.Repeat("# filler line\n", 12) + "# mutation-reproducer\nexit 0", args: []string{"scripts/late-marker.sh"}, want: 1},
		{label: "case 28.b: a marker line carrying extra text declares nothing", rel: "scripts/marker-with-suffix.sh", body: "# mutation-reproducer: because the build succeeded\nexit 0", args: []string{"scripts/marker-with-suffix.sh"}, want: 1},
		{label: "case 28.c: a marker at exactly line 10 declares", rel: "scripts/edge-marker-10.sh", body: strings.Repeat("# filler line\n", 7) + "# mutation-reproducer\nexit 0", args: []string{"scripts/edge-marker-10.sh"}, want: 0},
		{label: "case 28.d: a marker at line 11 declares nothing", rel: "scripts/edge-marker-11.sh", body: strings.Repeat("# filler line\n", 8) + "# mutation-reproducer\nexit 0", args: []string{"scripts/edge-marker-11.sh"}, want: 1},
		{label: "case 30: a mismatched --reproducer-sha pin is refused without running", rel: "scripts/sha-pinned.sh", body: "exit 5", args: []string{"scripts/sha-pinned.sh", "--reproducer-sha", strings.Repeat("0", 64)}, want: 2, needle: "pinned sha", notRan: true},
		{label: "case 31: a non-hex --reproducer-sha value is a usage failure", rel: "scripts/sha-never-runs.sh", body: "exit 0", args: []string{"scripts/sha-never-runs.sh", "--reproducer-sha", "zz-nothex"}, want: 4, needle: "usage", notRan: true},
		// The bash 3.2 floor cases: the empty-ARGS and the argument paths.
		{label: "case 32: a single-token reproducer decides under /bin/bash (empty ARGS)", rel: "scripts/floor-single-token.sh", body: "exit 3", args: []string{"scripts/floor-single-token.sh"}, want: 0, needle: "defect demonstrated"},
		{label: "case 33: a two-token reproducer decides under /bin/bash (sentinel fd)", rel: "scripts/floor-two-token.sh", body: "exit 3", args: []string{"scripts/floor-two-token.sh arg1"}, want: 0, needle: "defect demonstrated"},
	}
	for _, bad := range []string{"0", "1", "2", "3", "7", "-1", "zero", " "} {
		rows = append(rows, row{label: fmt.Sprintf("case 23.%s: --pre-fix-verdict '%s' is a usage failure", bad, bad),
			rel: "scripts/never-runs.sh", body: "exit 0", args: []string{"scripts/never-runs.sh", "--pre-fix-verdict", bad}, want: 4, needle: "usage", notRan: true})
	}
	for i, c := range rrMetachars(t) {
		rows = append(rows, row{label: fmt.Sprintf("case 6.%d: banned metacharacter '%c' alone is refused", i, c),
			args: []string{fmt.Sprintf("scripts/x%csh", c)}, want: 2, needle: "metacharacter", notRan: true})
	}
	for _, r := range rows {
		t.Run(r.label, func(t *testing.T) {
			t.Parallel()
			wt := rrWorktree(t)
			if r.rel != "" {
				rrFixture(t, wt, r.rel, r.body)
			}
			got, out := rrRun(t, rrEnv(t), append([]string{wt}, r.args...)...)
			rrExpect(t, got, out, r.want, r.needle)
			if r.notRan {
				rrNotRan(t, wt)
			}
			if r.ran {
				rrRan(t, wt)
			}
		})
	}

	t.Run("case 8: a symlink escaping the worktree is refused", func(t *testing.T) {
		t.Parallel()
		wt, outside := rrWorktree(t), t.TempDir()
		rrFixture(t, outside, "secret.sh", "exit 0")
		if err := os.Symlink(filepath.Join(outside, "secret.sh"), filepath.Join(wt, "scripts/escape.sh")); err != nil {
			t.Fatal(err)
		}
		got, out := rrRun(t, rrEnv(t), wt, "scripts/escape.sh")
		rrExpect(t, got, out, 2, "outside the worktree")
		rrNotRan(t, outside)
		// An exec by the unresolved link path drops RAN beside the link (rrFixture).
		rrNotRan(t, filepath.Join(wt, "scripts"))
	})

	t.Run("case 9: a command past the bound is killed and reported unverifiable", func(t *testing.T) {
		t.Parallel()
		wt := rrWorktree(t)
		pgidFile := filepath.Join(wt, "pgid")
		rrFixture(t, wt, "scripts/hangs.sh", `echo $$ > "`+pgidFile+`.tmp"; mv "`+pgidFile+`.tmp" "`+pgidFile+`"; sleep 30`)
		got, out := rrRun(t, rrEnv(t, "RUN_REPRODUCER_BOUND_FILE", pgidFile), wt, "scripts/hangs.sh")
		rrExpect(t, got, out, 3, "unverifiable")
		// The reproducer leads its own process group: the whole group,
		// its `sleep 30` included, is gone.
		b, err := os.ReadFile(pgidFile)
		if err != nil {
			t.Fatal(err)
		}
		var pgid int
		fmt.Sscan(string(b), &pgid)
		rrGone(t, -pgid)
	})

	t.Run("case 10: a detached survivor is named with its pid", func(t *testing.T) {
		t.Parallel()
		// The detached child reports over a pipe once setsid and SIG_IGN
		// are in place; only then does the parent expire the bound through
		// RUN_REPRODUCER_BOUND_FILE, so the bound fires on the case's
		// condition, not on the machine's load. The Env bound is only the
		// backstop.
		wt := rrWorktree(t)
		boundFile := filepath.Join(t.TempDir(), "bound")
		rrFixture(t, wt, "scripts/detach.sh", `exec python3 -c "
import os, time, signal
ready_r, ready_w = os.pipe()
if os.fork() == 0:
    os.setsid()
    open(\"`+filepath.Join(wt, "groups")+`\", \"a\").write(str(os.getpid()) + \"\\n\")
    signal.signal(signal.SIGTERM, signal.SIG_IGN)
    os.write(ready_w, b\"x\")
    time.sleep(30)
    os._exit(0)
else:
    os.read(ready_r, 1)
    open(\"`+boundFile+`\", \"w\").close()
    time.sleep(30)
"`)
		got, out := rrRun(t, rrEnv(t, "RUN_REPRODUCER_BOUND_FILE", boundFile), wt, "scripts/detach.sh")
		rrSurvivorKilled(t, got, out, "still running at the")
	})

	t.Run("case 11: a passing reproducer's output is captured and delimited", func(t *testing.T) {
		t.Parallel()
		wt := rrWorktree(t)
		rrFixture(t, wt, "scripts/passes-with-output.sh", "echo \"some diagnostic output here\"\nexit 0")
		got, out := rrRun(t, rrEnv(t), wt, "scripts/passes-with-output.sh")
		for _, n := range []string{"some diagnostic output here", "captured reproducer output begin", "captured reproducer output end"} {
			rrExpect(t, got, out, 1, n)
		}
	})

	t.Run("case 13: a child detached before the parent exits normally is still caught", func(t *testing.T) {
		t.Parallel()
		// The parent waits until the child has detached and ignores SIGTERM,
		// then lives 0.5s -- several descendant snapshots -- and exits on its
		// own, well inside the bound.
		wt := rrWorktree(t)
		rrFixture(t, wt, "scripts/detach-early.sh", `exec python3 -c "
import os, time, signal
ready_r, ready_w = os.pipe()
if os.fork() == 0:
    os.setsid()
    open(\"`+filepath.Join(wt, "groups")+`\", \"a\").write(str(os.getpid()) + \"\\n\")
    signal.signal(signal.SIGTERM, signal.SIG_IGN)
    os.write(ready_w, b\"x\")
    time.sleep(30)
    os._exit(0)
else:
    os.read(ready_r, 1)
    time.sleep(0.5)
"`)
		got, out := rrRun(t, rrEnv(t), wt, "scripts/detach-early.sh")
		rrSurvivorKilled(t, got, out, "forked a detached child")
	})

	t.Run("case 14: a real double-forked grandchild is named with its pid", func(t *testing.T) {
		t.Parallel()
		wt := rrWorktree(t)
		rrFixture(t, wt, "scripts/doublefork.sh", "( bash -c \"trap \\\"\\\" TERM; sleep 30\" & sleep 0.6 ) &\nsleep 1\nexit 0")
		got, out := rrRun(t, rrEnv(t), wt, "scripts/doublefork.sh")
		rrSurvivorKilled(t, got, out, "surviving process")
	})

	t.Run("a failed process-table read is never a verdict (exit 4)", func(t *testing.T) {
		t.Parallel()
		wt := rrWorktree(t)
		rrFixture(t, wt, "scripts/ok.sh", "exit 0")
		env := rrEnv(t)
		env.ProcTable = func() ([]rrProc, error) { return nil, errors.New("injected read failure") }
		got, out := rrRun(t, env, wt, "scripts/ok.sh")
		rrExpect(t, got, out, 4, "run-reproducer: cannot answer — the process table could not be read (injected read failure)")
		rrRan(t, wt)
	})

	t.Run("case 15: a flow-guard the shim cannot build is never a verdict (shim, exit 4)", func(t *testing.T) {
		t.Parallel()
		// The Go port has no run-time dependency file to lose: its one
		// missing-dependency path is the shim's own, no go to build
		// flow-guard with and nothing cached, which must never read as a
		// verdict.
		wt := rrWorktree(t)
		rrFixture(t, wt, "scripts/ok.sh", "exit 0")
		shim, err := filepath.Abs("../../../scripts/run-reproducer.sh")
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(shim, wt, "scripts/ok.sh")
		cmd.Env = []string{"PATH=/usr/bin:/bin", "TMPDIR=" + t.TempDir(), "FLOW_GUARD_CACHE_DIR=" + t.TempDir()}
		out, err := cmd.CombinedOutput()
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatalf("shim: %v\n%s", err, out)
		}
		rrExpect(t, ee.ExitCode(), string(out), 4, "run-reproducer: cannot build flow-guard from")
		rrNotRan(t, wt)
	})

	t.Run("case 16: a fast double-forked grandchild that re-parents before the first poll is still named with its pid", func(t *testing.T) {
		t.Parallel()
		// The intermediate exits at once, re-parenting the grandchild before
		// any descendant snapshot; the reproducer waits only until the
		// grandchild ignores SIGTERM, so the kill path is deterministic.
		wt := rrWorktree(t)
		ready := filepath.Join(wt, "ready")
		rrFixture(t, wt, "scripts/doublefork-fast.sh", fmt.Sprintf("( bash -c \"trap \\\"\\\" TERM; touch %q; sleep 31\" & )\nuntil [ -e %q ]; do sleep 0.01; done\nexit 0", ready, ready))
		got, out := rrRun(t, rrEnv(t), wt, "scripts/doublefork-fast.sh")
		rrSurvivorKilled(t, got, out, "surviving process")
	})

	t.Run("case 17: an ordinary non-forking reproducer still demonstrates the defect", func(t *testing.T) {
		t.Parallel()
		wt := rrWorktree(t)
		rrFixture(t, wt, "scripts/ordinary17.sh", `echo $$ > "`+wt+`/pgid"; exit 5`)
		got, out := rrRun(t, rrEnv(t), wt, "scripts/ordinary17.sh marker17")
		rrExpect(t, got, out, 0, "")
		t.Run("case 17: no process is left behind", func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(wt, "pgid"))
			if err != nil {
				t.Fatal(err)
			}
			var pgid int
			fmt.Sscan(string(b), &pgid)
			if syscall.Kill(-pgid, 0) == nil {
				t.Fatalf("process group %d survived the run", pgid)
			}
		})
	})

	t.Run("case 18: a fast double-forked grandchild is still named on the timeout path", func(t *testing.T) {
		t.Parallel()
		// Same shape as case 16, but the reproducer outlives the bound; the
		// bound file fires it once the grandchild ignores SIGTERM.
		wt := rrWorktree(t)
		ready := filepath.Join(wt, "ready")
		rrFixture(t, wt, "scripts/doublefork-timeout.sh", fmt.Sprintf("( bash -c \"trap \\\"\\\" TERM; touch %q; sleep 30\" & )\nsleep 30", ready))
		got, out := rrRun(t, rrEnv(t, "RUN_REPRODUCER_BOUND_FILE", ready), wt, "scripts/doublefork-timeout.sh")
		rrSurvivorKilled(t, got, out, "still running at the")
	})

	t.Run("case 29: a matching --reproducer-sha pin is accepted and the sha is printed with the verdict", func(t *testing.T) {
		t.Parallel()
		wt := rrWorktree(t)
		rrFixture(t, wt, "scripts/sha-ok.sh", "exit 5")
		want := shasum(t, filepath.Join(wt, "scripts/sha-ok.sh"))
		got, out := rrRun(t, rrEnv(t), wt, "scripts/sha-ok.sh", "--reproducer-sha", want)
		rrExpect(t, got, out, 0, "reproducer sha "+want)
	})

	t.Run("case 34: a reproducer whose interpreter is missing cannot answer (exec failure)", func(t *testing.T) {
		t.Parallel()
		wt := rrWorktree(t)
		writeExec(t, filepath.Join(wt, "scripts/no-interpreter.sh"), "#!/nonexistent-interpreter\n")
		got, out := rrRun(t, rrEnv(t), wt, "scripts/no-interpreter.sh")
		rrExpect(t, got, out, 4, "exec could not run")
		rrNotRan(t, wt)
	})
}

// TestRRDeadlineParsesEnvKnob pins the RUN_REPRODUCER_*_SECONDS parsing:
// all digits is taken, anything else falls back to the default.
func TestRRDeadlineParsesEnvKnob(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		raw, def, label string
		want            time.Duration
	}{
		{"7", "60", "7", 7 * time.Second},
		{"", "60", "60", 60 * time.Second},
		{"1.5", "60", "60", 60 * time.Second},
		{"-3", "5", "5", 5 * time.Second},
		{"12abc", "5", "5", 5 * time.Second},
	} {
		d, l := rrDeadline(0, c.raw, c.def)
		if d != c.want || l != c.label {
			t.Errorf("rrDeadline(0, %q, %q) = %v %q, want %v %q", c.raw, c.def, d, l, c.want, c.label)
		}
	}
}

// TestRunReproducerChildPWDIsWorktree: the bash guard's `cd -- "$WORKTREE"`
// gave the reproducer PWD=<worktree>. A non-bash reproducer is used because
// bash resets PWD on its own.
func TestRunReproducerChildPWDIsWorktree(t *testing.T) {
	t.Parallel()
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not on PATH")
	}
	wt := rrWorktree(t)
	writeExec(t, filepath.Join(wt, "scripts/pwd.py"), "#!"+py+`
import os, sys
sys.exit(9 if os.path.realpath(os.environ.get("PWD", "")) == os.getcwd() else 0)
`)
	got, out := rrRun(t, rrEnv(t), wt, "scripts/pwd.py")
	if got != 0 {
		t.Fatalf("exit %d, want 0 (reproducer saw PWD=<worktree>):\n%s", got, out)
	}
}

// TestRunReproducerSurvivorNamedWhenReapedAtKill pins the interleaving that
// machine load only sometimes produced (cases 10, 13, 14, 16, 18): the
// survivor's reaper collects it the instant its SIGKILL lands, before the
// sweep can look again. The survivor is this test's own child, so the kill
// hook is its reaper; the process and every signal are real.
func TestRunReproducerSurvivorNamedWhenReapedAtKill(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("/bin/sh", "-c", `trap "" TERM; echo ready; exec sleep 30`)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	if _, err := bufio.NewReader(stdout).ReadString('\n'); err != nil {
		t.Fatalf("survivor never reported SIGTERM ignored: %v", err)
	}
	pid := cmd.Process.Pid
	reapAtKill := func(p int, sig syscall.Signal) error {
		err := syscall.Kill(p, sig)
		if sig == syscall.SIGKILL {
			_ = cmd.Wait()
		}
		return err
	}
	if got := rrSweep([]int{pid}, rrTestGrace, reapAtKill); len(got) != 1 || got[0] != pid {
		t.Fatalf("survivors %v, want [%d]: a pid that outlived its SIGTERM grace went unnamed", got, pid)
	}
}
