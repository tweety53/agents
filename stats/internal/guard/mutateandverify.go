package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

// mutateAndVerify is scripts/mutate-and-verify.sh: that script's header
// comment is the contract — apply a mutation, run each harness before and
// after, report the new-failure set per harness, then unconditionally restore
// and verify the touched files are clean; exit 0 ran clean, 2 refused before
// mutating anything or post-restore drift, 3 could not fully restore, 4
// cannot answer. The reasoning for each step, moved here from the bash body
// it replaced (d71a2327), sits beside the code it explains.
func init() {
	Registry["mutate-and-verify"] = mutateAndVerify
}

func mutateAndVerify(args []string, env Env, stdout, stderr io.Writer) int {
	cannotAnswer := func(msg string) int {
		fmt.Fprintf(stderr, "mutate-and-verify: cannot answer — %s\n", msg)
		return 4
	}
	refuse := func(msg string) int {
		fmt.Fprintf(stderr, "mutate-and-verify: refused — %s — nothing was mutated\n", msg)
		return 2
	}
	if len(args) < 2 {
		fmt.Fprintln(stderr, "mutate-and-verify: usage: mutate-and-verify.sh <patch-file> <harness>...")
		return 4
	}

	// Patch and harness arguments resolve to absolute paths against the
	// caller's own cwd BEFORE anything runs from the repository root — a
	// relative argument must keep meaning what the caller meant by it. Joined
	// as bash joined them, "$ORIG_PWD/$1", uncleaned: a refusal names the path
	// exactly as the bash did.
	origPwd := env.Dir
	toAbs := func(p string) string {
		if strings.HasPrefix(p, "/") {
			return p
		}
		return origPwd + "/" + p
	}
	patchArg := args[0]
	patch := toAbs(patchArg)
	var harnesses []string
	for _, h := range args[1:] {
		harnesses = append(harnesses, toAbs(h))
	}

	// Step 1: resolve the repository root; every git call and harness below
	// runs from it, as the bash's `cd -- "$REPO_ROOT"` made them.
	top, rc := gitExec(origPwd, nil, io.Discard, "rev-parse", "--show-toplevel")
	if rc != 0 {
		return cannotAnswer("not inside a git worktree: " + origPwd)
	}
	root := strings.TrimRight(top, "\n")

	// Step 2: the patch file must exist and be readable.
	if !isFile(patch) || syscall.Access(patch, 4) != nil {
		return cannotAnswer("cannot read patch file: " + patchArg)
	}

	// Step 3: git apply --check — refuse (exit 2) without touching anything.
	check := exec.Command("git", "apply", "--check", "--whitespace=nowarn", patch)
	check.Dir = root
	if out, err := check.CombinedOutput(); err != nil {
		fmt.Fprintln(stderr, "mutate-and-verify: git apply --check failed:")
		fmt.Fprintln(stderr, strings.TrimRight(string(out), "\n"))
		return refuse(patchArg + " does not apply cleanly")
	}

	// Step 4: the touched file list is git apply --numstat's third column
	// (cut -f3: a line without a tab is kept whole).
	numstat, _ := gitExec(root, nil, stderr, "apply", "--numstat", patch)
	var touched []string
	for _, line := range strings.Split(numstat, "\n") {
		if f := strings.Split(line, "\t"); len(f) >= 3 {
			line = f[2]
		} else if len(f) == 2 {
			line = ""
		}
		if line != "" {
			touched = append(touched, line)
		}
	}
	if len(touched) == 0 {
		return cannotAnswer(patchArg + " names no touched file")
	}

	// Step 5: every named harness must exist and be executable.
	for _, h := range harnesses {
		if !isFile(h) {
			return refuse("harness does not exist: " + h)
		}
		if syscall.Access(h, 1) != nil {
			return refuse("harness is not executable: " + h)
		}
	}

	// Step 6: none of the touched files may already carry uncommitted
	// changes — never conflate a pre-existing dirty file with the mutation's
	// own diff at restore time. A failing git ends the run with its own
	// status, as `set -e` ended the bash.
	status := append([]string{"status", "--porcelain", "--untracked-files=normal", "--"}, touched...)
	dirty, rc := gitExec(root, nil, stderr, status...)
	if rc != 0 {
		return rc
	}
	if dirty = strings.TrimRight(dirty, "\n"); dirty != "" {
		fmt.Fprintln(stderr, "mutate-and-verify: touched files already have uncommitted changes:")
		fmt.Fprintln(stderr, dirty)
		return refuse("a file the patch touches is not clean")
	}

	// Step 6b: snapshot the tree the mutation starts from — status and stash
	// list — so the exit path can tell post-restore residue from the state
	// the run actually left.
	snapshot, rc := snapshotTreeState(root, stderr)
	if rc != 0 {
		return rc
	}

	// Step 7: from here every exit goes through r.exit, the bash's EXIT
	// trap, installed before anything mutates — a signal included: restore
	// and report as for a run whose last status was 0, then die of it.
	//
	// ponytail: SIGINT restores at once, as SIGTERM does; bash waited for a
	// running harness and carried on if that harness survived the SIGINT. A
	// Ctrl-C reaches the harness too, so both end in the same restore.
	r := &mvRun{root: root, touched: touched, snapshot: snapshot, stdout: stdout, stderr: stderr}
	defer trapExitSignals(func() {
		r.dying.Store(true)
		r.exit(0)
	})()
	// Only the main flow's own writes halt on EPIPE: the trap's restore
	// report, and `git apply`'s output written while mu is held, go through
	// r.stdout, where a failed write carries on -- halting either would
	// leave the trap waiting forever.
	return r.finish(r.main(patchArg, patch, harnesses, env, mvHaltOnEPIPE{stdout}, cannotAnswer))
}

// mvRun is the state the bash's EXIT trap read: whether the patch is applied,
// what it touched, and the pre-mutation snapshot.
type mvRun struct {
	root, snapshot string
	touched        []string
	stdout, stderr io.Writer

	mu      sync.Mutex // held across `git apply` and the restore, so neither interleaves
	applied bool
	once    sync.Once
	rc      int
	dying   atomic.Bool // a signal is ending the process
}

// main is steps 8-14: the baseline pass, the mutation, the mutated pass and
// the verdict. It returns the exit code the trap starts from.
func (r *mvRun) main(patchArg, patch string, harnesses []string, env Env, stdout io.Writer, cannotAnswer func(string) int) int {
	fmt.Fprintf(stdout, "mutate-and-verify: patch %s touches:\n", patchArg)
	for _, t := range r.touched {
		fmt.Fprintf(stdout, "  %s\n", t)
	}

	// Step 9: baseline pass — every harness, before applying the patch.
	fmt.Fprint(stdout, "\nmutate-and-verify: baseline pass (before mutation)\n")
	base := make([][]string, len(harnesses))
	for i, h := range harnesses {
		name := filepath.Base(h)
		ok, fails, answered := mvRunHarness(h, r.root)
		if !answered {
			return cannotAnswer("harness " + name + " produced no ok:/FAIL: line on the baseline pass")
		}
		base[i] = fails
		fmt.Fprintf(stdout, "  %s: baseline ok=%d fail=%d\n", name, ok, len(fails))
		if len(fails) > 0 {
			fmt.Fprintf(stdout, "  %s: baseline NOT clean — %s\n", name, strings.Join(fails, " "))
		}
	}

	// Step 10: apply the mutation. A failing git apply ends the run with its
	// own status, as `set -e` did, before the patch counts as applied. A
	// signal already being handled has run -- or is waiting on mu to run --
	// the restore, so nothing is applied after it (bash deferred a trapped
	// signal until its foreground command returned; the lock is that order).
	rc, ok := r.apply(r.stdout, patch)
	if !ok {
		select {} // the signal goroutine ends the process
	}
	if rc != 0 {
		return rc
	}

	// Step 11: mutated pass — every harness, after applying the patch.
	fmt.Fprint(stdout, "\nmutate-and-verify: mutated pass (after mutation)\n")
	mut := make([][]string, len(harnesses))
	for i, h := range harnesses {
		name := filepath.Base(h)
		ok, fails, answered := mvRunHarness(h, r.root)
		if !answered {
			return cannotAnswer("harness " + name + " produced no ok:/FAIL: line on the mutated pass")
		}
		mut[i] = fails
		fmt.Fprintf(stdout, "  %s: mutated ok=%d fail=%d\n", name, ok, len(fails))
	}

	// Step 13: the blast-radius bound — MUTATE_AND_VERIFY_MAX_NEW_FAILURES
	// when it is all digits, else 5.
	bound := env.Getenv("MUTATE_AND_VERIFY_MAX_NEW_FAILURES")
	if bound == "" || strings.Trim(bound, "0123456789") != "" {
		bound = "5"
	}
	fmt.Fprintf(stdout, "\nmutate-and-verify: verdict (blast-radius bound: %s)\n", bound)

	for i, h := range harnesses {
		name := filepath.Base(h)
		// Step 12: new failures = the mutated FAIL-name set minus the
		// baseline's, by whole name — a case name can contain spaces ("guard
		// exits 0"). An empty name (a bare `FAIL: ` line) is counted in the
		// pass's fail= total but never takes part in the difference, as the
		// bash's newline-joined sets dropped it.
		var fresh []string
		for _, n := range mut[i] {
			if n != "" && !slices.Contains(base[i], n) {
				fresh = append(fresh, n)
			}
		}
		// Step 14: per-harness verdict. A bound bash's `-le` cannot read (out
		// of range) is no bound, so the verdict falls to FLAG as the bash's did;
		// the bash's own "integer expression expected" line on stderr is not
		// reproduced.
		limit, err := strconv.Atoi(bound)
		switch {
		case len(fresh) == 0:
			fmt.Fprintf(stdout, "  %s: surviving mutant — 0 new failures\n", name)
		case err == nil && len(fresh) <= limit:
			fmt.Fprintf(stdout, "  %s: caught — %d new failure(s): %s\n", name, len(fresh), strings.Join(fresh, " "))
		default:
			fmt.Fprintf(stdout, "  %s: FLAG suspicious blast radius — %d new failures (bound %s): %s\n",
				name, len(fresh), bound, strings.Join(fresh, " "))
		}
	}

	// Step 15's restore-and-verify — and the "ran clean" report, printed only
	// once the restore is verified — is the exit path's (r.exit).
	fmt.Fprint(stdout, "\nmutate-and-verify: restoring touched files\n")
	return 0
}

// mvHaltOnEPIPE stops the main flow at a write to a closed stdout. The
// SIGPIPE that came with it is the trap's (trapExitSignals), which restores
// and exits; bash died at that write, so no later step -- the next harness
// above all -- may start meanwhile. Only a trapped SIGPIPE returns EPIPE
// here: untrapped, the runtime ends the process at the write.
type mvHaltOnEPIPE struct{ w io.Writer }

func (h mvHaltOnEPIPE) Write(p []byte) (int, error) {
	n, err := h.w.Write(p)
	if errors.Is(err, syscall.EPIPE) {
		select {}
	}
	return n, err
}

// apply is step 10's `git apply`, under mu; ok is false, and nothing is
// applied, once a signal is being handled.
func (r *mvRun) apply(stdout io.Writer, patch string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.dying.Load() {
		return 0, false
	}
	// The output is written once git has exited, so a closed stdout can
	// neither cut the apply short nor hide that it ran: applied is git's
	// own verdict.
	out, rc := gitExec(r.root, nil, r.stderr, "apply", "--whitespace=nowarn", patch)
	r.applied = rc == 0
	_, _ = io.WriteString(stdout, out)
	return rc, true
}

// finish is the main flow's exit: r.exit, except that once a signal is
// ending the process this goroutine never returns an exit code that could
// race the signal's own.
func (r *mvRun) finish(rc int) int {
	rc = r.exit(rc)
	if r.dying.Load() {
		select {} // the signal goroutine ends the process
	}
	return rc
}

// exit is the bash's EXIT trap, run once. If the patch was applied, restore
// the touched files and re-verify they are clean; a residual forces exit 3,
// overriding whatever the run was carrying, so a mutated file is never left
// behind unreported. Then diff the whole tree against the step 6b snapshot:
// any new stash entry or unexpected status line is reported and forces exit
// 2, unless exit 3's residual already took precedence.
func (r *mvRun) exit(rc int) int {
	r.once.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.rc = rc
		if !r.applied {
			return
		}
		gitExec(r.root, r.stdout, io.Discard, append([]string{"checkout", "--"}, r.touched...)...)
		residual, _ := gitExec(r.root, nil, io.Discard,
			append([]string{"status", "--porcelain", "--untracked-files=normal", "--"}, r.touched...)...)
		if residual = strings.TrimRight(residual, "\n"); residual != "" {
			fmt.Fprintln(r.stderr, "mutate-and-verify: could not fully restore — residual status:")
			fmt.Fprintln(r.stderr, residual)
			r.rc = 3
		} else if r.rc == 0 {
			// Only claim success once the restore is verified clean —
			// printing it earlier would contradict a residual reported moments
			// later.
			fmt.Fprintln(r.stdout, "mutate-and-verify: ran clean — touched files restored")
		}
		if drift, _ := checkTreeRestored(r.root, r.snapshot, r.stderr); len(drift) > 0 {
			fmt.Fprintln(r.stderr, "mutate-and-verify: post-restore drift detected:")
			for _, d := range drift {
				fmt.Fprintln(r.stderr, d)
			}
			if r.rc != 3 {
				r.rc = 2
			}
		}
	})
	return r.rc
}

// mvRunHarness is run_harness: the harness run from the repository root with
// the inherited environment and stdin, stdout and stderr captured together,
// its exit status ignored. It reports the `ok: ` line count and every `FAIL: `
// line's name; answered is false when neither kind of line appeared.
//
// ponytail: a harness without a #! line fails to exec here (ENOEXEC) and
// reads as unanswered, where bash ran it as a bash script; every harness this
// runs is a scripts/test-*.sh with one.
func mvRunHarness(h, dir string) (ok int, fails []string, answered bool) {
	cmd := exec.Command(h)
	cmd.Dir, cmd.Stdin = dir, os.Stdin
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	_ = cmd.Run()
	// $(...) dropped every NUL byte (bash 5 also warned "ignored null byte
	// in input" on stderr, the one byte of its output not reproduced here).
	for _, line := range strings.Split(strings.ReplaceAll(out.String(), "\x00", ""), "\n") {
		switch {
		case strings.HasPrefix(line, "ok: "):
			ok++
		case strings.HasPrefix(line, "FAIL: "):
			fails = append(fails, strings.TrimPrefix(line, "FAIL: "))
		}
	}
	return ok, fails, ok > 0 || len(fails) > 0
}
