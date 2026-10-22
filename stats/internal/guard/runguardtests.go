package guard

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// runGuardTests is scripts/run-guard-tests.sh: that script's header comment
// is the contract -- discover <test root>/test-*.sh, refuse an uncompanioned
// check-*.sh, run the harnesses concurrently, print one ok/FAIL line each and
// a summary, replay only the failures, and fail a suite that changed the
// watched repository's tree; exit 0 clean, 1 a failing harness, a companion
// gap or a tree change, 2 cannot answer. It replaced the script's body and
// the bash spawn/capture/replay library whose last
// caller it was (KAN-362, KAN-921).
func init() { Registry["run-guard-tests"] = runGuardTests }

const rgtName = "run-guard-tests"

// rgtJob is one harness's run: its exit code, its wall time and the file
// its merged stdout+stderr was captured into.
type rgtJob struct {
	rc, elapsed int
	out         string
}

func runGuardTests(_ []string, env Env, stdout, stderr io.Writer) int {
	lookup := env.LookupEnv
	if lookup == nil {
		lookup = func(k string) (string, bool) { v := env.Getenv(k); return v, v != "" }
	}
	repoRoot := env.Getenv("FLOW_GUARD_REPO_ROOT")
	scriptDir := filepath.Join(repoRoot, "scripts")

	// RUN_GUARD_TESTS_ROOT and GUARD_TESTS_REPO_ROOT are explicit, opt-in
	// overrides honoured only when set, and refused when set but empty:
	// this runner's own Go test points both at fixtures, so without them a
	// case would re-enter the real suite or gate on this repository's tree.
	testRoot, set := lookup("RUN_GUARD_TESTS_ROOT")
	switch {
	case set && testRoot == "":
		fmt.Fprintln(stderr, "RUN_GUARD_TESTS_ROOT is set but empty")
		return 2
	case !set:
		if repoRoot == "" {
			fmt.Fprintf(stderr, "%s: FLOW_GUARD_REPO_ROOT is unset — run scripts/run-guard-tests.sh, which sets it\n", rgtName)
			return 2
		}
		testRoot = scriptDir
	}
	watched, set := lookup("GUARD_TESTS_REPO_ROOT")
	switch {
	case set && watched == "":
		fmt.Fprintln(stderr, "GUARD_TESTS_REPO_ROOT is set but empty")
		return 2
	case !set:
		watched, _ = capture(rgtGit(env, scriptDir, "rev-parse", "--show-toplevel"))
	}

	// The validation reads rev-parse's VERDICT, never its exit code alone:
	// inside a bare repository `--is-inside-work-tree` exits 0 and prints
	// `false`, and treating that as a pass failed the before-snapshot's
	// `git status` where the documented skip was owed (kan-584's panel,
	// F1/P1).
	if watched != "" {
		if v, _ := capture(rgtGit(env, watched, "rev-parse", "--is-inside-work-tree")); v != "true" {
			fmt.Fprintf(stderr, "%s: %s is not a git work tree — skipping the tree gate\n", rgtName, watched)
			watched = ""
		}
	}

	if fi, err := os.Stat(testRoot); err != nil || !fi.IsDir() {
		fmt.Fprintf(stderr, "%s: not a directory: %s\n", rgtName, testRoot)
		return 2
	}

	// DISCOVERY IS BY GLOB, <test root>/test-*.sh, flat -- never recursive,
	// never a hand-maintained list -- so a new harness nobody wires into
	// .flow/project.md is never silently skipped. Byte order, where the
	// script's glob sorted by LC_COLLATE: the two differ only on names mixing
	// case or punctuation, which no harness name does, and only where no
	// recorded duration orders the list.
	harnesses, _ := filepath.Glob(filepath.Join(testRoot, "test-*.sh"))

	// COMPANION PRESENCE (kan-387): every check-*.sh guard gets a
	// test-check-*.sh companion, or -- a guard ported to Go (kan-760), whose
	// file carries `flow_guard_exec <name> ` -- a
	// <test root>/../stats/internal/guard/<name, - as _>_test.go, which
	// scripts/test-go-guards.sh runs. A gap is a detected violation of the
	// suite's contract, exit 1, refused before anything runs.
	guards, _ := filepath.Glob(filepath.Join(testRoot, "check-*.sh"))
	var missing []string
	for _, g := range guards {
		name := strings.TrimSuffix(filepath.Base(g), ".sh")
		if rgtExists(filepath.Join(testRoot, "test-"+name+".sh")) {
			continue
		}
		body, _ := os.ReadFile(g)
		goTest := filepath.Join(testRoot, "..", "stats", "internal", "guard", strings.ReplaceAll(name, "-", "_")+"_test.go")
		if strings.Contains(string(body), "flow_guard_exec "+name+" ") && rgtExists(goTest) {
			continue
		}
		missing = append(missing, filepath.Base(g))
	}
	if len(missing) > 0 {
		for _, g := range missing {
			fmt.Fprintf(stderr, "%s: %s has no companion test-%s\n", rgtName, g, g)
		}
		fmt.Fprintf(stderr, "%s: every check-*.sh guard gets a test-check-*.sh companion — add the missing harnesses\n", rgtName)
		return 1
	}

	if len(harnesses) == 0 {
		fmt.Fprintf(stderr, "%s: no test-*.sh harnesses found under %s\n", rgtName, testRoot)
		return 2
	}

	// Longest first: jobs start in list order, so a slow harness that starts
	// late ends the run on its own tail -- in glob order the suite took
	// ~150s wall for ~945s of work on ten cores. Each harness's last duration
	// is kept in the repository's git dir (never committed, rewritten every
	// run); a harness with no recorded duration sorts first, as if it were
	// the slowest.
	durations := ""
	if watched != "" {
		common, _ := capture(rgtGit(env, watched, "rev-parse", "--path-format=absolute", "--git-common-dir"))
		durations = common + "/guard-test-durations"
		if body, err := os.ReadFile(durations); err == nil {
			last := map[string]int{}
			for _, line := range strings.Split(string(body), "\n") {
				d, name, ok := strings.Cut(line, "\t")
				if _, seen := last[name]; ok && d != "" && !seen {
					last[name], _ = strconv.Atoi(d) // non-numeric reads 0, as sort -n reads it
				}
			}
			recorded := func(h string) int {
				if d, ok := last[filepath.Base(h)]; ok {
					return d
				}
				return 999999
			}
			sort.SliceStable(harnesses, func(i, j int) bool { return recorded(harnesses[i]) > recorded(harnesses[j]) })
		}
	}

	// An interrupt removes the temp dir, then dies of that signal, as
	// the bash parallel library's EXIT/INT/TERM traps did for its capture directory: a
	// deferred RemoveAll alone never runs in a signal-killed process. The
	// context also kills every harness still running.
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	var caught syscall.Signal
	watching := make(chan struct{})
	go func() {
		defer close(watching)
		select {
		case s := <-sigs:
			caught = s.(syscall.Signal)
			cancel()
		case <-ctx.Done():
		}
	}()
	tmp := env.Getenv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}
	tmpDir, err := os.MkdirTemp(tmp, rgtName+"-time.")
	if err != nil {
		signal.Stop(sigs)
		cancel()
		fmt.Fprintf(stderr, "%s: cannot create a temporary directory under %s\n", rgtName, tmp)
		return 2
	}
	defer func() {
		signal.Stop(sigs)
		cancel()
		<-watching
		os.RemoveAll(tmpDir)
		if caught != 0 {
			signal.Reset(caught)
			_ = syscall.Kill(os.Getpid(), caught)
			time.Sleep(time.Second) // reached only when the signal was ignored at start
			os.Exit(128 + int(caught))
		}
	}()

	// FLOW-GUARD CACHE (kan-760): a ported guard's shim builds flow-guard
	// into FLOW_GUARD_CACHE_DIR, pointed here at this run's temp dir so no
	// harness writes into the operator's real cache. Telemetry off keeps
	// harnesses that exec flow-guard from posting guard rows to the dev
	// store.
	childEnv := rgtChildEnv(lookup,
		"FLOW_GUARD_CACHE_DIR="+filepath.Join(tmpDir, "flow-guard-cache"),
		"FLOW_GUARD_TELEMETRY=off")

	// THE TREE GATE (kan-584) compares STATUS LINES, not content: edits made
	// and restored inside the run net to zero, and mutating an
	// already-modified or already-untracked file in place nets the same
	// porcelain line it had before -- neither is caught. Its contract is
	// structural change to the tree the suite leaves behind.
	var before string
	if watched != "" {
		var ok bool
		if before, ok = rgtStatus(env, watched); !ok {
			fmt.Fprintf(stderr, "%s: cannot read the status of %s — skipping the tree gate\n", rgtName, watched)
			watched = ""
		}
	}

	start := time.Now()
	jobs, ok := rgtJobCount(lookup, stderr)
	if !ok {
		fmt.Fprintf(stderr, "%s: could not run the harnesses (see above)\n", rgtName)
		return 2
	}
	results := rgtRun(ctx, harnesses, jobs, tmpDir, env.Dir, childEnv)
	if ctx.Err() != nil {
		return 2 // interrupted: the deferred cleanup ends the process
	}

	var failed []int
	for i, h := range harnesses {
		r := results[i]
		if r.rc == 0 {
			fmt.Fprintf(stdout, "ok:   %s (%ds)\n", filepath.Base(h), r.elapsed)
		} else {
			fmt.Fprintf(stdout, "FAIL: %s (%ds)\n", filepath.Base(h), r.elapsed)
			failed = append(failed, i)
		}
	}
	if durations != "" {
		var b strings.Builder
		for i, h := range harnesses {
			fmt.Fprintf(&b, "%d\t%s\n", results[i].elapsed, filepath.Base(h))
		}
		_ = os.WriteFile(durations, []byte(b.String()), 0o644)
	}
	fmt.Fprintf(stdout, "\n%d harnesses, %d passed, %d failed, %ds wall\n",
		len(harnesses), len(harnesses)-len(failed), len(failed), time.Now().Unix()-start.Unix())

	dirty := false
	var after string
	if watched != "" {
		var ok bool
		if after, ok = rgtStatus(env, watched); !ok {
			fmt.Fprintf(stderr, "%s: cannot re-read the status of %s — the tree gate is skipped\n", rgtName, watched)
		} else {
			dirty = after != before
		}
	}

	if len(failed) == 0 && !dirty {
		return 0
	}

	if len(failed) > 0 {
		fmt.Fprint(stdout, "\nFAILED:\n")
		for _, i := range failed {
			fmt.Fprintf(stdout, "  %s\n", filepath.Base(harnesses[i]))
		}
		fmt.Fprint(stdout, "\n")
		// REPLAY IS IN LIST ORDER, NEVER COMPLETION ORDER, each failure's
		// capture whole and contiguous under a `---- <name> ----` header
		// and followed by a blank line. Merged stdout+stderr, because a
		// replay reproduces what a reader would have seen watching that
		// harness alone in a terminal.
		for _, i := range failed {
			fmt.Fprintf(stdout, "---- %s ----\n", filepath.Base(harnesses[i]))
			out, _ := os.ReadFile(results[i].out)
			stdout.Write(out)
			fmt.Fprint(stdout, "\n")
		}
	}

	if dirty {
		fmt.Fprintf(stdout, "\n%s: the watched repository tree changed during the suite (%s):\n", rgtName, watched)
		// Name only the delta: entries present after the suite and absent
		// before; a pure-deletion delta prints the fallback line instead.
		was := map[string]bool{}
		for _, l := range strings.Split(before, "\n") {
			was[l] = true
		}
		var delta []string
		for _, l := range strings.Split(after, "\n") {
			if l != "" && !was[l] {
				delta = append(delta, l)
			}
		}
		if len(delta) > 0 {
			fmt.Fprintln(stdout, strings.Join(delta, "\n"))
		} else {
			fmt.Fprintln(stdout, "  (entries removed relative to the pre-suite snapshot)")
		}
	}
	return 1
}

// rgtRun runs each harness as `bash <harness>` in dir, at most jobs at a
// time, starting them in list order. Results are indexed by list position,
// so every later walk is in list order by construction. stdout and stderr
// share one capture file -- one descriptor, so the harness's own write order
// is kept -- and a file rather than a pipe, so a background process a
// harness leaves holding its output never delays the run.
func rgtRun(ctx context.Context, harnesses []string, jobs int, tmpDir, dir string, childEnv []string) []rgtJob {
	results := make([]rgtJob, len(harnesses))
	sem := make(chan struct{}, jobs)
	var wg sync.WaitGroup
	for i, h := range harnesses {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() { <-sem; wg.Done() }()
			r := &results[i]
			r.out = filepath.Join(tmpDir, strconv.Itoa(i)+".out")
			f, err := os.Create(r.out)
			if err != nil {
				r.rc = 126
				return
			}
			defer f.Close()
			cmd := exec.CommandContext(ctx, "bash", h)
			cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = dir, childEnv, f, f
			// Per-harness wall time is whole Unix seconds end minus start,
			// what the script's `$(date +%s)` arithmetic computed -- never
			// a truncated time.Since, which differs whenever a run
			// straddles a second boundary.
			s := time.Now().Unix()
			r.rc = exitCode(cmd.Run())
			r.elapsed = int(time.Now().Unix() - s)
		}()
	}
	wg.Wait()
	return results
}

// rgtJobCount is the job count: JOBS when set at all, else the core count.
// A JOBS set but not a positive integer is refused, never coerced into the
// fallback -- this repository's guards reject rather than fold bad input
// into a reassuring default.
func rgtJobCount(lookup func(string) (string, bool), stderr io.Writer) (int, bool) {
	v, set := lookup("JOBS")
	if !set {
		return runtime.NumCPU(), true
	}
	n, err := strconv.Atoi(v)
	switch {
	case err != nil || strings.TrimLeft(v, "0123456789") != "":
		fmt.Fprintf(stderr, "parallel_job_count: JOBS must be a positive integer, got '%s'\n", v)
		return 0, false
	case n == 0:
		fmt.Fprintln(stderr, "parallel_job_count: JOBS must be a positive integer, got 0")
		return 0, false
	}
	return n, true
}

// rgtChildEnv is what each harness runs under: the process environment with
// this runner's own overrides as env carries them -- a harness reads
// GUARD_TESTS_REPO_ROOT, as the script's exported environment passed it --
// and extra on top.
func rgtChildEnv(lookup func(string) (string, bool), extra ...string) []string {
	names := []string{"RUN_GUARD_TESTS_ROOT", "GUARD_TESTS_REPO_ROOT", "JOBS", "TMPDIR"}
	var out []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		keep := true
		for _, n := range names {
			keep = keep && name != n
		}
		if keep {
			out = append(out, kv)
		}
	}
	for _, n := range names {
		if v, ok := lookup(n); ok {
			out = append(out, n+"="+v)
		}
	}
	return append(out, extra...)
}

func rgtGit(env Env, dir string, args ...string) *exec.Cmd {
	return envGit(env)(append([]string{"-C", dir}, args...)...)
}

// rgtStatus is `git status --porcelain --untracked-files=normal` over dir.
func rgtStatus(env Env, dir string) (string, bool) {
	out, err := rgtGit(env, dir, "status", "--porcelain", "--untracked-files=normal").Output()
	return string(out), err == nil
}

func rgtExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
