package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// proveReproducer is scripts/prove-reproducer.sh: that script's header
// comment is the contract -- exit 0 proof held, 1 the proof did not hold,
// 2 cannot answer. The reasoning for each branch, moved here from the bash
// body it replaced (d71a2327), sits beside the code it explains.
//
// EACH LEG IS AN IN-PROCESS runReproducer CALL (runreproducer.go) with the
// argv the bash passed to run-reproducer.sh, its stdout and stderr into one
// buffer as the bash's `>leg.out 2>&1` did, and its return code read
// exactly as the bash read the shim's exit. The bash refused at exit 2 when
// run-reproducer.sh or lib/reproducer-path.sh was unreadable beside it;
// both are compiled in here, so there is nothing left to find missing.
func init() { Registry["prove-reproducer"] = proveReproducer }

const prPrefix = "prove-reproducer: "

func proveReproducer(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, prPrefix+format+"\n", a...)
		return 2
	}
	if len(args) != 3 {
		fmt.Fprint(stderr, "usage: prove-reproducer.sh <worktree> <pre-fix-ref> <reproducer-path>\n"+
			"       <reproducer-path> is relative to <worktree>; the pre-fix leg runs\n"+
			"       it in a detached scratch worktree at <pre-fix-ref>, the post-fix\n"+
			"       leg against <worktree> itself\n")
		return 2
	}
	wt, ref, rel := args[0], args[1], args[2]
	// abs is the path bash's test operators saw: relative to the working
	// directory, never cleaned -- `..` after a symlink resolves on disk.
	abs := func(p string) string { return smcAbs(env, p) }
	git := func(args ...string) *exec.Cmd {
		cmd := exec.Command("git", append([]string{"-C", wt}, args...)...)
		cmd.Dir = env.Dir
		return cmd
	}
	// tool is one of the bash's coreutils steps, its own diagnostic passed
	// through to stderr as the bash left it.
	tool := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Dir, cmd.Stderr = env.Dir, stderr
		return cmd.Run()
	}

	if wt == "" || !isDir(abs(wt)) { // bash: [ -d "" ] is false
		return die("not a directory: %s", wt)
	}
	if git("rev-parse", "--is-inside-work-tree").Run() != nil {
		return die("not a git worktree: %s", wt)
	}
	resolve := git("rev-parse", "--verify", "--quiet", ref+"^{commit}")
	resolve.Stderr = stderr
	out, err := resolve.Output()
	preSHA := strings.TrimRight(string(out), "\n")
	if err != nil || preSHA == "" {
		return die("pre-fix ref does not resolve to a commit: %s", ref)
	}

	// The reproducer path follows the recorded-form refusals at the door:
	// relative, no leading dash, no `..` segment. The copy step resolves it
	// against the scratch root, so a `..` here would escape the scratch the
	// same way F17's absolute ROOT escaped the live tree.
	if reason := reproducerPathRefusal(rel); reason != "" {
		return die("%s", reason)
	}
	if fi, err := os.Stat(abs(wt + "/" + rel)); err != nil || !fi.Mode().IsRegular() {
		return die("no reproducer file at %s in %s", rel, wt)
	}
	if syscall.Access(abs(wt+"/"+rel), rrExecOK) != nil {
		return die("reproducer is not executable: %s/%s", wt, rel)
	}

	// The bash's "${TMPDIR:-/tmp}/prove-reproducer.XXXXXX", slash for slash.
	tmp := env.Getenv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}
	base, err := os.MkdirTemp(tmp+"/", "prove-reproducer.")
	if err != nil {
		return die("no writable temporary directory")
	}
	scratch := base + "/scratch"

	// THE SCRATCH IS REMOVED ON EVERY PATH OUT, as the bash's EXIT trap did.
	// Errors ignored throughout: cleanup runs on every path, including those
	// where the scratch was never created or git is mid-failure, and a
	// cleanup failure must never mask the verdict already printed.
	cleanup := func() {
		_ = git("worktree", "remove", "--force", scratch).Run()
		_ = os.RemoveAll(base)
	}
	defer cleanup()
	defer trapExitSignals(cleanup)()

	add := git("worktree", "add", "--detach", "--quiet", scratch, preSHA)
	var addOut bytes.Buffer
	add.Stdout, add.Stderr = &addOut, &addOut
	if err := add.Run(); err != nil {
		_, _ = stderr.Write(addOut.Bytes())
		return die("could not materialize the scratch worktree at %s", preSHA)
	}

	// A plumbing failure in the copy steps is the documented cannot-answer
	// 2, never the "proof did not hold" 1 with no leg run. dirname, not
	// path.Dir: `a/./r.sh` is created as `a/.`, uncleaned. The copy keeps
	// its mode, so it loses the executable bit only on a noexec TMPDIR.
	dir := gdcDirname(rel)
	if tool("mkdir", "-p", scratch+"/"+dir) != nil {
		return die("cannot create %s/%s in the scratch worktree", scratch, dir)
	}
	if tool("cp", "-p", wt+"/"+rel, scratch+"/"+rel) != nil {
		return die("cannot copy %s/%s into the scratch worktree", wt, rel)
	}
	if syscall.Access(scratch+"/"+rel, rrExecOK) != nil {
		return die("the scratch copy lost its executable bit: %s/%s", scratch, rel)
	}

	// Each leg's own temporary files go under the scratch base, so the
	// cleanup removes them on every path out, a signal's included -- the
	// bash's separate run-reproducer process removed its own as it ended.
	// The reproducer itself still inherits the caller's TMPDIR.
	legEnv := env
	legEnv.Getenv = func(k string) string {
		if k == "TMPDIR" {
			return base
		}
		return env.Getenv(k)
	}
	var leg1Out, leg2Out bytes.Buffer
	leg1 := runReproducer([]string{scratch, rel}, legEnv, &leg1Out, &leg1Out)
	leg2 := runReproducer([]string{wt, rel}, legEnv, &leg2Out, &leg2Out)

	fmt.Fprintf(stdout, "== pre-fix leg — scratch worktree at %s\n", preSHA)
	_, _ = stdout.Write(leg1Out.Bytes())
	fmt.Fprintf(stdout, "== pre-fix leg exit: %d\n", leg1)
	fmt.Fprintf(stdout, "== post-fix leg — %s\n", wt)
	_, _ = stdout.Write(leg2Out.Bytes())
	fmt.Fprintf(stdout, "== post-fix leg exit: %d\n", leg2)

	// A refused or unverifiable leg (2 or 3) or one that could not answer
	// (4) spends the proof: nothing is claimed about the defect either way.
	for _, leg := range []int{leg1, leg2} {
		if leg != 0 && leg != 1 {
			fmt.Fprintf(stderr, "%scannot spend the proof — run-reproducer exited %d on a leg (2 refused, 3 unverifiable, 4 cannot answer)\n", prPrefix, leg)
			return 2
		}
	}
	if leg1 == 0 && leg2 == 1 {
		fmt.Fprintln(stdout, "PROOF HELD — pre-fix: defect demonstrated (exit 0); post-fix: defect not demonstrated (exit 1)")
		return 0
	}
	if leg1 != 0 {
		fmt.Fprintf(stdout, "PROOF FAILED — pre-fix leg read defect not demonstrated: the reproducer does not demonstrate the defect at %s\n", preSHA)
	}
	if leg2 != 1 {
		fmt.Fprintln(stdout, "PROOF FAILED — post-fix leg still reads defect demonstrated: the fix does not stop this reproducer")
	}
	return 1
}

// reproducerPathRefusal is scripts/lib/reproducer-path.sh's (at d71a2327)
// reproducer_path_refusal, moved here with its last caller: "" when rel is
// acceptable, else the one reason line. run-reproducer (runreproducer.go)
// and check-panel-reproducers keep their own, older copies of these
// refusals beside their wider lexical sets (shell metacharacters, URLs, NUL
// bytes, resolved-symlink containment); only what was verbatim-identical
// moved into that library, and their wider sets are not this rule.
func reproducerPathRefusal(rel string) string {
	switch {
	case rel == "":
		return "reproducer path is empty"
	case strings.HasPrefix(rel, "/"):
		return "reproducer path must be relative to the worktree, got an absolute path: " + rel
	case strings.HasPrefix(rel, "-"):
		return "reproducer path may not begin with a dash: " + rel
	case hasDotDotSegment(rel):
		return "reproducer path may not contain a .. segment: " + rel
	}
	return ""
}

// trapExitSignals makes the terminating signals do what they did to a bash
// script with an EXIT trap: the trap -- cleanup here -- ran at once, and
// the script died of the signal, leaving any running child orphaned. A
// signal ignored at start stays ignored, as a non-interactive bash cannot
// trap one. It returns the function that stops listening.
//
// SIGHUP, SIGINT and SIGTERM are re-raised, so the process dies of them.
// The rest cannot be: the Go runtime drops a SIGPIPE it did not raise
// itself, and answers a sent SIGABRT, SIGFPE, SIGSYS, SIGTRAP (and, on
// linux, SIGSEGV, SIGBUS, SIGILL) with a stack dump and exit 2 -- which mutate-and-verify's caller would read as
// "refused, nothing was mutated". They end in exit 128+n, the status a
// shell reads for death by that signal. A SIGPIPE arriving on the caller's
// last write is handled too: the returned function takes a pending signal
// before letting the caller return its own exit code.
//
// ponytail: gaps against bash, each closable only by a trap in the shim,
// if a caller ever depends on it. The other signals bash's trap caught
// (SIGUSR1, SIGUSR2, SIGALRM, ...) the Go runtime ignores, so the run
// carries on to its normal exit. A sent SIGSEGV, SIGBUS or SIGILL is
// trapped on linux but never reaches signal.Notify on darwin, nor does
// darwin's SIGEMT: the runtime crashes with exit 2, and nothing is cleaned
// up. A SIGPIPE ignored at start is trapped anyway (exit 141 after the
// cleanup), where bash carried on to its own exit. A SIGTERM
// ignored at start is handled anyway: Go reports an inherited ignore for
// SIGHUP and SIGINT only. A SIGINT sent to this pid alone ends the run at
// once, where bash waited for its foreground child and carried on if that
// child survived; a Ctrl-C reaches the whole process group, so both end.
func trapExitSignals(cleanup func()) func() {
	var sigs []os.Signal
	for _, s := range []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE,
		syscall.SIGABRT, syscall.SIGBUS, syscall.SIGFPE, syscall.SIGILL, syscall.SIGSEGV, syscall.SIGSYS, syscall.SIGTRAP} {
		if !signal.Ignored(s) {
			sigs = append(sigs, s)
		}
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	die := func(s os.Signal) {
		cleanup()
		sig := s.(syscall.Signal)
		if sig == syscall.SIGHUP || sig == syscall.SIGINT || sig == syscall.SIGTERM {
			signal.Reset(s)
			_ = syscall.Kill(os.Getpid(), sig)
			time.Sleep(time.Second) // reached only if the reset signal did not end the process
		}
		os.Exit(128 + int(sig))
	}
	done, finished := make(chan struct{}), make(chan struct{})
	go func() {
		select {
		case s := <-ch:
			die(s)
		case <-done:
			// signal.Stop has delivered any signal still in flight -- a
			// SIGPIPE raised by the caller's last write, typically -- so
			// one pending now is handled before the caller's own exit.
			select {
			case s := <-ch:
				die(s)
			default:
			}
			close(finished)
		}
	}()
	return func() {
		signal.Stop(ch)
		close(done)
		<-finished
	}
}
