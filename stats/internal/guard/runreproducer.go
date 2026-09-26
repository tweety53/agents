package guard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// runReproducer is scripts/run-reproducer.sh: that script's header comment
// is the contract -- arguments, verdict lines, the stdout/stderr split and
// the exit codes 0 demonstrated, 1 not demonstrated, 2 refused, 3
// unverifiable, 4 cannot answer.
//
// Where the bash needed plumbing Go does not, the plumbing is gone and the
// behaviour stays: SysProcAttr.Setsid replaces the python3 setsid shim (so
// a missing python3 is no longer a refusal), and a failed Start replaces
// the sentinel pipe -- a vanished worktree and an exec the kernel refused
// are still exit 4 with the bash's own lines. The bound and the grace are
// timers; the descendant snapshot stays periodic, because a child that
// calls setsid itself is visible only while its parent is alive.
func init() { Registry["run-reproducer"] = runReproducer }

const (
	rrUsage = "run-reproducer: usage: run-reproducer.sh <worktree> <reproducer-command-line> [--pre-fix-verdict <demonstrated|not-demonstrated>] [--reproducer-sha <sha>]"
	// rrPoll is the bash guard's POLL: how often the descendant tree is
	// snapshotted while the reproducer runs. Sub-second, as the bash's was
	// where the platform's sleep accepted one; a Go ticker needs no probe.
	rrPoll       = 200 * time.Millisecond
	rrMarker     = "# mutation-reproducer"
	rrMaxDepth   = 64
	rrLivePoll   = 10 * time.Millisecond
	rrExecOK     = 0x1 // access(2)'s X_OK
	rrDefaultBnd = "20"
	rrDefaultGrc = "2"
)

func runReproducer(args []string, env Env, stdout, stderr io.Writer) int {
	usage := func() int { fmt.Fprintln(stderr, rrUsage); return 4 }
	refuse := func(msg string) int {
		fmt.Fprintf(stderr, "run-reproducer: refused — %s — never executed\n", msg)
		return 2
	}
	if len(args) < 2 || args[0] == "" || args[1] == "" {
		return usage()
	}
	worktreeArg, cmdText := args[0], args[1]
	// --pre-fix-verdict <demonstrated|not-demonstrated> — the verdict this
	// reproducer produced when the caller ran it against the defect-present
	// code, named in this guard's own printed vocabulary: `demonstrated` is
	// the verdict a dispatch-time exit 0 carries, `not-demonstrated` the one an
	// exit 1 carries — exactly what the panel parent holds from its
	// `; echo "F<n>: exit $?"` record of the dispatch-time run. On the
	// fix-round re-run, a verdict here that MATCHES the pre-fix one is refused:
	// a reproducer whose verdict is identical pre-fix and post-fix is ambiguous
	// under either convention, and nothing built on it can be verified. Every
	// OTHER value — every number included — is a usage failure, reported and
	// never ignored: a raw exit is not a verdict, and a raw `1` passed where
	// the dispatch-time run printed `demonstrated` is the exact misfeed
	// KAN-614 records three times across runs (kan-542, kan-546, kan-556), so
	// the misfeed is caught here at the flag, before anything executes,
	// instead of surfacing as a spurious ambiguity after a real run. A bare
	// run (two arguments, no flag) is the dispatch-time decision and decides
	// exactly as it always has.
	//
	// --reproducer-sha <sha> — the reproducer sha the dispatch-time run printed,
	// carried by the fix-round re-run so the verdict comparison is pinned to the
	// SAME file: a reproducer re-authored between rounds (its
	// mutation-convention declaration included) is refused before it executes.
	// A hex-string value only; anything else is a usage failure.
	preFix, expectSHA := "", ""
	for rest := args[2:]; len(rest) > 0; rest = rest[2:] {
		if len(rest) < 2 {
			return usage()
		}
		switch rest[0] {
		case "--pre-fix-verdict":
			if rest[1] != "demonstrated" && rest[1] != "not-demonstrated" {
				return usage()
			}
			preFix = rest[1]
		case "--reproducer-sha":
			// `printf '%s' "$2" | tr '[:upper:]' '[:lower:]'` in a command
			// substitution: lowercased, trailing newlines dropped.
			expectSHA = strings.ToLower(strings.TrimRight(rest[1], "\n"))
			if expectSHA == "" || strings.Trim(expectSHA, "0123456789abcdef") != "" {
				return usage()
			}
		default:
			return usage()
		}
	}

	wt := worktreeArg
	if !filepath.IsAbs(wt) {
		wt = filepath.Join(env.Dir, wt)
	}
	if !isDir(wt) {
		fmt.Fprintf(stderr, "run-reproducer: not a directory: %s\n", worktreeArg)
		return 4
	}
	// `cd -- "$WORKTREE" && pwd -P`: `..` taken logically, then physical.
	// Canonicalised before anything is concatenated onto it, for the same
	// reason check-panel-reproducers.sh does: a relative worktree argument
	// beginning with `-` would otherwise make every path built from it look
	// like an option to whatever reads it. Physical, not logical: a TMPDIR
	// that is itself a symlink (`/tmp` on Darwin, to `/private/tmp`) left the
	// bash's logical `pwd` and realpath's own physical answer for the same
	// file disagreeing on every case — every reproducer refused as "outside
	// the worktree" against a worktree it was always inside.
	worktree, err := filepath.EvalSymlinks(filepath.Clean(wt))
	if err != nil {
		fmt.Fprintf(stderr, "run-reproducer: worktree vanished before it could be resolved: %s\n", worktreeArg)
		return 4
	}

	// Lexical checks — mirrors check-panel-reproducers.sh's own reproducer-line
	// rules, repeated here because a record can be edited after that guard last
	// ran, and because this guard must refuse on its own even when it is
	// handed a command line from somewhere other than a validated record.
	//
	// An embedded newline is refused outright: every check below reads
	// cmdText as one line, and a second line hidden inside the argument would
	// ride along unchecked by every one of them.
	if strings.Contains(cmdText, "\n") {
		return refuse("the command line carries an embedded newline")
	}
	// Split on space AND tab, as check-panel-reproducers.sh's own tokenizer
	// does, and the path token is the split's first element. The bash guard
	// learned this the hard way: its `${CMD_TEXT%% *}` split on a literal space
	// only, so a tab-separated command line yielded a path token containing
	// the tab and everything after it — refused as "does not exist" rather than
	// resolved correctly. Deriving the path token from the tokenized list keeps
	// the two splits in agreement. (The bash also had to avoid an unquoted
	// `for tok in $CMD_TEXT`, whose pathname expansion would stat the
	// filesystem ahead of the deliberate touch below; FieldsFunc expands
	// nothing.)
	tokens := strings.FieldsFunc(cmdText, func(r rune) bool { return r == ' ' || r == '\t' })
	if len(tokens) == 0 {
		return refuse("the command line names no path token")
	}
	pathToken := tokens[0]
	if strings.HasPrefix(pathToken, "-") {
		return refuse(fmt.Sprintf("the path token '%s' begins with '-' — a reproducer names a path, never an option", pathToken))
	}
	if strings.Contains(cmdText, "://") {
		return refuse("the command names a URL — a reproducer is a bare path inside the worktree, never a network location")
	}
	// The same banned-character set check-panel-reproducers.sh enforces,
	// including both quote characters: a direct exec never expands any of
	// these, so their presence only ever means the author expected a shell to
	// see them. reproducerMetachars (metachars.go) is this set, single-sourced
	// with check-panel-reproducers.sh's bash copy in
	// scripts/reproducer-metachars.sh — see that file's header for why: this
	// set has already drifted between the two scripts twice — and
	// TestMetacharsMatchBashSource pins the two together.
	for _, c := range reproducerMetachars {
		if strings.ContainsRune(cmdText, c) {
			return refuse(fmt.Sprintf("the command carries the shell metacharacter '%c' — a reproducer is a bare path optionally followed by plain arguments, never a shell command line", c))
		}
	}
	for _, tok := range tokens {
		if strings.HasPrefix(tok, "/") {
			return refuse(fmt.Sprintf("the token '%s' is an absolute path — a reproducer names a path relative to the worktree, on the path token or any argument", tok))
		}
		if hasDotDotSegment(tok) {
			return refuse(fmt.Sprintf("the token '%s' carries a '..' path segment — a reproducer must stay inside the worktree, on the path token or any argument", tok))
		}
	}

	// Resolved containment — the check check-panel-reproducers.sh's own header
	// says it deliberately leaves to whatever runs the reproducer against a
	// real worktree. Every token is resolved to its physical path and required
	// to stay inside the worktree; filepath.EvalSymlinks (the bash's `realpath`)
	// follows `..`, `.` and symlinks in one step, so a relative path with no
	// lexical `..` segment that escapes through a symlink is caught here.
	inside := func(p string) bool { return p == worktree || strings.HasPrefix(p, worktree+"/") }
	candidate := worktree + "/" + pathToken
	if _, err := os.Stat(candidate); err != nil {
		return refuse(fmt.Sprintf("the path token '%s' does not exist inside the worktree", pathToken))
	}
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return refuse(fmt.Sprintf("the path token '%s' could not be resolved", pathToken))
	}
	if !inside(resolved) {
		return refuse(fmt.Sprintf("the path token '%s' resolves to '%s', outside the worktree — a symlink escape", pathToken, resolved))
	}
	if !isFile(resolved) {
		return refuse(fmt.Sprintf("the path token '%s' does not resolve to a regular file", pathToken))
	}
	if syscall.Access(resolved, rrExecOK) != nil {
		return refuse(fmt.Sprintf("the path token '%s' resolves to a file with no execute permission", pathToken))
	}

	// The declaration of the mutation-reproducer convention (KAN-568), read from
	// the file the path token resolved to — the only place a reproducer can
	// state its own convention, since neither the panel record's
	// `finding-reproducer:` line nor the runner's argument vector carries a
	// field for it. Read before execution so the verdict mapping below is
	// decided by the same containment checks every reproducer passes, never a
	// second path around them.
	mutation := declaresMutation(resolved)
	// The reproducer's identity, taken from the same resolved file every
	// containment check above passed, and taken before execution so a
	// --reproducer-sha pin refuses a re-authored file in the never-executed
	// shape class. Printed with every verdict below, so the caller holding a
	// dispatch-time verdict also holds the sha its re-run must pin. It comes
	// from the repository's one sha256 helper (sha256.go) — a second wrapper
	// here would be exactly the drift the reproducer-metachars.sh extraction
	// exists to stop.
	sha, err := sha256HexFile(resolved)
	if err != nil {
		fmt.Fprintf(stderr, "run-reproducer: cannot take a sha of '%s' — no SHA-256 tool on this machine\n", resolved)
		return 4
	}
	if expectSHA != "" && expectSHA != sha {
		return refuse(fmt.Sprintf("the reproducer file is not the one the dispatch-time run read (pinned sha %s, found %s) — a reproducer re-authored between rounds, its mutation-convention declaration included, is a different reproducer answering in a vocabulary the pre-fix verdict never carried; re-run it against the defect-present code and carry its fresh verdict and sha", expectSHA, sha))
	}

	// Every argument after the path token is resolved and checked the same way,
	// but only when it names something that exists: a plain flag such as
	// `--strict` is not a path at all, and resolving a token that names nothing
	// on disk answers a question this guard was never asked.
	cmdArgs := tokens[1:]
	for _, arg := range cmdArgs {
		c := worktree + "/" + arg
		if _, err := os.Stat(c); err != nil {
			continue
		}
		r, err := filepath.EvalSymlinks(c)
		if err != nil {
			return refuse(fmt.Sprintf("the argument '%s' could not be resolved", arg))
		}
		if !inside(r) {
			return refuse(fmt.Sprintf("the argument '%s' resolves to '%s', outside the worktree — a symlink escape", arg, r))
		}
	}

	// Execution, bounded, with a direct exec — the guarantee this guard
	// exists to make true rather than merely claim.
	//
	// The bound and grace are this rule's own numbers: 20 seconds, plus a
	// 2-second SIGTERM-to-SIGKILL grace matching check-cleanup-complete's own
	// survivors kill grace (ccDefaultGrace). RUN_REPRODUCER_* overrides
	// exist for the same reason CHECK_CLEANUP_SURVIVORS_TIMEOUT does in that
	// guard: driving a real 20-second wait through a test harness costs real
	// wall clock per case, and the override is monotone in the safe direction —
	// it can only ever shrink the wait, producing MORE timeouts, never fewer,
	// so no value of it can turn an unverified run into a verified one. Never
	// set either for a normal invocation; the Go tests inject
	// Env.ReproducerBound and Env.ReproducerGrace instead.
	bound, boundLabel := rrDeadline(env.ReproducerBound, env.Getenv("RUN_REPRODUCER_BOUND_SECONDS"), rrDefaultBnd)
	grace, _ := rrDeadline(env.ReproducerGrace, env.Getenv("RUN_REPRODUCER_GRACE_SECONDS"), rrDefaultGrc)
	// RUN_REPRODUCER_BOUND_FILE, when set, also expires the bound the moment that
	// path exists. Test-only, and monotone the same way: it can end the wait
	// early, never extend it. It exists because the bash guard's $SECONDS counted
	// whole seconds, so a BOUND of 2 fired anywhere from 1 to 2 seconds after
	// launch — a window a fixture's own startup chain (subshell, setsid wrapper,
	// bash, python, fork) overran under a loaded suite, killing the parent
	// before the process the case exists to observe was ever forked. A file the
	// fixture creates once that process is in place makes the bound fire on the
	// case's condition, not on the machine's load. The Go bound is a timer; the
	// knob keeps that guarantee for harnesses that drive the shim.
	boundFile := env.Getenv("RUN_REPRODUCER_BOUND_FILE")

	tmpdir := env.Getenv("TMPDIR")
	if tmpdir == "" {
		tmpdir = "/tmp"
	}
	outF, err := os.CreateTemp(tmpdir, "run-reproducer-out.*")
	if err != nil {
		fmt.Fprintf(stderr, "run-reproducer: no writable temporary directory — cannot run '%s' at all\n", pathToken)
		return 4
	}
	defer os.Remove(outF.Name())
	defer outF.Close()
	errF, err := os.CreateTemp(tmpdir, "run-reproducer-err.*")
	if err != nil {
		fmt.Fprintf(stderr, "run-reproducer: no writable temporary directory — cannot run '%s' at all\n", pathToken)
		return 4
	}
	defer os.Remove(errF.Name())
	defer errF.Close()
	emit := func(w io.Writer) { emitCaptured(w, outF.Name(), errF.Name()) }

	// argv[0] is the resolved path, as the bash's python3 shim's
	// execvp(sys.argv[1], sys.argv[1:]) gave it. stdin is /dev/null (a nil
	// cmd.Stdin): this guard is non-interactive, and a reproducer that
	// prompts for input must fail rather than wait on a terminal nobody is
	// watching.
	//
	// Setsid (D4) is what puts the exec'd reproducer in a process group of
	// its own — a NEW session, with a NEW process group whose pgid equals
	// the reproducer's own pid — so the group kill in rrSupervise reaches
	// every process the reproducer forks, even one that re-parents to
	// launchd before any poll ever observes it as a descendant. The bash
	// guard needed a python3 `os.setsid()` shim for this, and refused at
	// exit 2 when python3 was missing rather than falling back to an
	// ungrouped exec, because an ungrouped exec is the exact condition in
	// which a re-parented survivor goes unseen. The Go port calls setsid in
	// the forked child before exec — a fresh child is never already a group
	// leader, so the EPERM the bash avoided by running without `set -m`
	// cannot arise — and there is no ungrouped path to refuse.
	cmd := exec.Command(resolved, cmdArgs...)
	cmd.Dir = worktree
	cmd.Stdout, cmd.Stderr = outF, errF
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	// The reproducer inherits the process environment (cmd.Environ keeps
	// PWD at cmd.Dir, as the bash cd did); an LC_ALL the
	// caller's Env sets (the exit-contract guard's `export LC_ALL=C`) wins.
	cmd.Env = cmd.Environ()
	if lc := env.Getenv("LC_ALL"); lc != "" {
		cmd.Env = append(cmd.Env, "LC_ALL="+lc)
	}
	// A failed Start is the Go port's replacement for the bash guard's sentinel
	// pipe: Start reports synchronously whether the child reached exec, so no
	// signal outlives the reproducer for it to tamper with. The bash sentinel's
	// own history is why that matters (F55): its first form was a
	// `mktemp -u ".../run-reproducer-sentinel.XXXXXX"` path under $TMPDIR, a
	// signal in name only — the guard ran with $TMPDIR inherited by the
	// reproducer it launched, and the sentinel's literal prefix sat in the
	// guard's own source, readable by anyone deciding what a reproducer fixture
	// should do. `rm -f "$TMPDIR"/run-reproducer-sentinel.*` as the fixture's
	// first act deleted it before the guard ever looked, and a genuine "defect
	// demonstrated" then read as "cannot answer", silencing the one answer this
	// tool exists to produce. It never forged the opposite direction — a
	// deleted sentinel was never read as a pass — but it defeated the purpose
	// on exactly the finding it was protecting.
	if err := cmd.Start(); err != nil {
		// The child never reached exec because it could not enter the worktree —
		// almost certainly because the worktree vanished between this guard's own
		// earlier resolution of it and the start. Whatever status that carries is
		// the runtime's report of that failure, never the reproducer's, so it is
		// read as "cannot answer" rather than as a pass or a fail.
		if !isDir(worktree) {
			fmt.Fprintf(stderr, "run-reproducer: cannot answer — the worktree vanished before '%s' could be started\n", cmdText)
			return 4
		}
		// The child started but the exec itself failed — the reproducer's
		// interpreter missing, the file not executable in the exec's own eyes, an
		// EACCES/ENOENT at the exec boundary. Measured before the bash guard's fix
		// for it, that failure surfaced as its python3 shim's own exit 1, which the
		// verdict mapping read as the REPRODUCER's exit: a reproducer that could
		// not be started at all answered "defect demonstrated" — a false verdict
		// spent on a script that never ran. The verdict here is not the
		// reproducer's, so the answer is "cannot answer", never a pass or a fail.
		emit(stderr)
		fmt.Fprintf(stderr, "run-reproducer: cannot answer — exec could not run '%s' (interpreter missing or the exec refused); its exit status was never the reproducer's\n", resolved)
		return 4
	}
	table := env.ProcTable
	if table == nil {
		table = rrProcTable
	}
	res := rrSupervise(cmd, bound, grace, boundFile, table)
	rc := res.rc

	// A reproducer killed at the bound, or one whose child survived the retry,
	// may already have written to the worktree before it was killed, so the
	// worktree is re-checked (rrGitStatusNote) before this guard's own run
	// continues any further. `/flow`'s implement phase re-checks it again itself
	// when the operator resumes, since nothing between this exit and that resume
	// observed what a still-alive survivor went on to write — that second check
	// is the caller's responsibility and is out of this single invocation's
	// scope.
	if len(res.survivors) > 0 {
		rrGitStatusNote(worktree, stderr)
		emit(stderr)
		pids := strings.Trim(fmt.Sprint(res.survivors), "[]")
		if res.timedOut {
			fmt.Fprintf(stderr, "run-reproducer: unverifiable — '%s' was still running at the %ss bound and was killed, and a detached child survived the process-group kill; surviving process pid(s): %s — find and kill it manually, and re-check the worktree again when you resume\n", cmdText, boundLabel, pids)
		} else {
			fmt.Fprintf(stderr, "run-reproducer: unverifiable — '%s' exited %d inside the %ss bound, but it forked a detached child that outlived it and survived the kill retry; surviving process pid(s): %s — find and kill it manually, and re-check the worktree again when you resume\n", cmdText, rc, boundLabel, pids)
		}
		return 3
	}
	if res.timedOut {
		rrGitStatusNote(worktree, stderr)
		emit(stderr)
		fmt.Fprintf(stderr, "run-reproducer: unverifiable — '%s' was still running at the %ss bound and was killed; its exit status is read as neither a pass nor a fail\n", cmdText, boundLabel)
		return 3
	}
	// A process-table read that failed may have missed a detached child, so
	// the reproducer's exit is no verdict (survivor-detection-deterministic).
	if res.tableErr != nil {
		rrGitStatusNote(worktree, stderr)
		emit(stderr)
		fmt.Fprintf(stderr, "run-reproducer: cannot answer — the process table could not be read (%v), so a detached child of '%s' could not be ruled out; its exit status is read as neither a pass nor a fail\n", res.tableErr, cmdText)
		return 4
	}
	// The child did reach exec, but exit codes 126 and 127 are the shell's own
	// conventional reports of "found but not executable" and "not found" for a
	// command it tried to run, emitted only when exec itself never replaced the
	// process image (a race after this guard's own executability check: the
	// file was removed, replaced, or its permissions changed in the gap). A
	// reproducer that actually ran can and occasionally does exit with either
	// number on its own, so this is a heuristic, not a certainty — but a false
	// "cannot answer" on a real 126/127 exit costs a re-run, while reading a
	// genuine plumbing failure as "defect demonstrated" costs a fix built on
	// nothing.
	if rc == 126 || rc == 127 {
		fmt.Fprintf(stderr, "run-reproducer: cannot answer — exec of '%s' itself failed (exit %d), which is never read as the reproducer's own exit status\n", resolved, rc)
		return 4
	}

	// THE VERDICT, mapped from the reproducer's own exit status under the
	// convention its declaration named (KAN-568): generic — any non-zero exit is
	// the defect present; mutation — exit 0 is the defect present, the build
	// having succeeded with the mutation landed. Computed once, ahead of both
	// readers below: the ambiguity refusal compares THIS, never the raw exit
	// code, and the final disposition reports it. The sha line precedes both,
	// so every verdict the caller reads carries the file identity it must pin
	// the fix-round re-run with.
	fmt.Fprintf(stderr, "run-reproducer: reproducer sha %s\n", sha)
	demonstrated := rc != 0
	if mutation {
		demonstrated = rc == 0
	}
	// The verdict's own name, in the vocabulary --pre-fix-verdict speaks: the
	// ambiguity refusal compares THIS against the caller's word, never a raw
	// exit code, so the comparison reads identically under either convention.
	verdictName := "not-demonstrated"
	if demonstrated {
		verdictName = "demonstrated"
	}
	// The ambiguity refusal (KAN-524), checked at the verdict point — after the
	// timeout (exit 3) and cannot-answer (exit 4) dispositions above have had
	// their say, since those runs produced no verdict to compare against. The
	// reproducer HAS run here, which is what separates this exit-2 class from
	// the shape refusals' "never executed" claim above.
	if preFix != "" && preFix == verdictName {
		emit(stderr)
		if mutation {
			fmt.Fprintf(stderr, "run-reproducer: refused — ambiguous reproducer: '%s' exited %d here, the same verdict it produced against the defect-present code (pre-fix verdict %s) — under the mutation-reproducer convention a reproducer must exit 0 while the defect is present (the build succeeds with the mutation landed) and non-zero once it is fixed; one that answers identically pre-fix and post-fix demonstrates nothing under either convention\n", cmdText, rc, preFix)
		} else {
			fmt.Fprintf(stderr, "run-reproducer: refused — ambiguous reproducer: '%s' exited %d here, the same verdict it produced against the defect-present code (pre-fix verdict %s) — a reproducer must exit non-zero while the defect is present and 0 once it is fixed; one that answers identically pre-fix and post-fix demonstrates nothing under either convention\n", cmdText, rc, preFix)
		}
		return 2
	}

	emit(stdout)
	switch {
	case demonstrated && mutation:
		fmt.Fprintf(stdout, "run-reproducer: defect demonstrated — '%s' exited 0 (mutation-reproducer convention: exit 0 = defect present)\n", cmdText)
		return 0
	case demonstrated:
		fmt.Fprintf(stdout, "run-reproducer: defect demonstrated — '%s' exited %d\n", cmdText, rc)
		return 0
	case mutation:
		fmt.Fprintf(stdout, "run-reproducer: defect not demonstrated — '%s' exited %d (mutation-reproducer convention: non-zero = the mutation did not survive)\n", cmdText, rc)
	default:
		fmt.Fprintf(stdout, "run-reproducer: defect not demonstrated — '%s' exited 0\n", cmdText)
	}
	return 1
}

// hasDotDotSegment is the bash `..|../*|*/..|*/../*` case pattern.
func hasDotDotSegment(p string) bool {
	return p == ".." || strings.HasPrefix(p, "../") || strings.HasSuffix(p, "/..") || strings.Contains(p, "/../")
}

// headLines is `head -n <n> -- <path>`, one entry per line without its
// newline; an unreadable file has no lines.
func headLines(path string, n int) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	r := bufio.NewReader(f)
	for len(out) < n {
		line, err := r.ReadString('\n')
		if line != "" {
			out = append(out, strings.TrimSuffix(line, "\n"))
		}
		if err != nil {
			break
		}
	}
	return out
}

// declaresMutation is `head -n 10 -- <path> | grep -qx '# mutation-reproducer'`:
// the KAN-568 declaration is that exact line within the first 10. Exact line,
// within that window: a line that merely CONTAINS the marker, or one buried
// below the shebang and its immediate commentary, is prose, not a
// declaration — the generic convention applies. rrMarker and the headLines
// call below are the canonical declaration
// scripts/check-mutation-reproducer-pin.sh pins against review-panel.md and
// run-reproducer.sh's header.
func declaresMutation(path string) bool {
	for _, l := range headLines(path, 10) {
		if l == rrMarker {
			return true
		}
	}
	return false
}

// rrDeadline is a RUN_REPRODUCER_* knob: an injected duration wins;
// otherwise the env value when it is all digits, else the default. The label
// is the number the bash prints before "s".
func rrDeadline(injected time.Duration, raw, def string) (time.Duration, string) {
	if injected > 0 {
		return injected, strconv.FormatFloat(injected.Seconds(), 'f', -1, 64)
	}
	if raw == "" || strings.Trim(raw, "0123456789") != "" {
		raw = def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		raw = def
		n, _ = strconv.Atoi(def)
	}
	return time.Duration(n) * time.Second, raw
}

type rrResult struct {
	rc        int
	timedOut  bool
	survivors []int
	tableErr  error // the first failed process-table read, if any
}

// rrSupervise waits for the started reproducer under the bound, snapshots
// its descendant tree while it lives, and runs the bash guard's kill
// sequence: group snapshot before any signal, SIGTERM to the group and
// SIGKILL to the leader on a timeout, a courtesy SIGTERM to the group, then
// the per-pid TERM/grace/KILL retry over every pid any snapshot named, a pid
// still alive when its SIGTERM grace ends being a survivor (rrSweep).
//
// A failed process-table read does not stop the sequence -- the reproducer
// is still killed and swept -- but it is kept in tableErr: a snapshot that
// could not be taken may have missed a survivor, so the caller must not
// read the run as a verdict.
func rrSupervise(cmd *exec.Cmd, bound, grace time.Duration, boundFile string, table func() ([]rrProc, error)) rrResult {
	pid := cmd.Process.Pid
	var tableErr error
	read := func() []rrProc {
		procs, err := table()
		if err != nil && tableErr == nil {
			tableErr = err
		}
		return procs
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()

	depth := map[int]int{} // every descendant ever seen, at its deepest depth
	boundTimer := time.NewTimer(bound)
	defer boundTimer.Stop()
	tick := time.NewTicker(rrPoll)
	defer tick.Stop()
	expired, timedOut := false, false
watch:
	for {
		select {
		case <-done:
			break watch
		default:
		}
		// Snapshotted on EVERY iteration, not only when the bound fires: a
		// reproducer that forks a child, detaches it (e.g. via setsid) and then
		// exits well inside the bound would otherwise leak that child entirely —
		// the child is reparented the instant the leader dies, so a walk taken only
		// after that has nothing left to find. Polling here, while the parent is
		// still alive, is the only window in which the relationship is visible at
		// all. The walk is the FULL descendant tree (rrDescendants), not one level:
		// a double fork — a child that forks a grandchild and exits immediately —
		// orphans the grandchild one hop below what a single-level children lookup
		// can ever see (F54).
		//
		// The iteration on which the bound fires snapshots too, before it breaks:
		// a reproducer that double-forks or detaches may not have spawned that child
		// yet at launch or at an earlier poll, and this is the last moment before
		// the kill sequence that a walk can still see it as this guard's own
		// descendant. Session ids are not used to find it: `ps -o sid=` is not a
		// valid ps(1) keyword on this platform — the citation F44 recorded as wrong
		// — and `ps -o sess=` was measured on this machine to report 0 for every
		// process tried, a real setsid() session leader included, so it cannot be
		// trusted to name one either. Parentage — `pgrep -P <pid>` in the bash, the
		// ppid column here — was measured instead to correctly name a detached child
		// before the kill and to keep naming it, alive, after the parent was killed
		// with SIGKILL, which is the mechanism this guard relies on; walking it
		// recursively is what extends that to every depth, not only the first.
		// The bound is read BEFORE the snapshot: a descendant forked before
		// the bound file appeared is then always in the last snapshot taken
		// (its parent still lives), never lost to a fork landing between
		// the snapshot and the check.
		_, statErr := os.Stat(boundFile)
		hit := expired || boundFile != "" && statErr == nil
		for p, d := range rrDescendants(read(), pid) {
			depth[p] = max(depth[p], d)
		}
		if hit {
			// No kill happens here. Signalling is deliberately deferred past
			// this loop entirely, to the single point below where the pre-kill
			// process group snapshot is taken — see the comment there for why
			// a kill this early would make that snapshot's own "before
			// anything is signalled" claim false on exactly this path.
			timedOut = true
			break
		}
		select {
		case <-done:
			break watch
		case <-tick.C:
		case <-boundTimer.C:
			expired = true
		}
	}

	// GROUP DETECTION, FIRST (D4) — a process-group snapshot taken HERE, at
	// the one point in this function that provably precedes every kill
	// below, on every path, timeout included. A second snapshot site used to
	// sit inside the bash guard's timeout branch and fire only there; that
	// gave this comment's own claim — "before anything is signalled" — two
	// places to be tested and only one of them true, which is how F4 slipped
	// through: the timeout branch's `kill -TERM`/`kill -KILL` used to run
	// before the snapshot rather than after it, so a process reaped between
	// that kill and the snapshot never appeared in it and so was never named
	// to the operator. The watch loop above now only sets timedOut and
	// breaks; every actual kill, timeout or not, happens after this line,
	// which is what makes "before anything is signalled" true on every path
	// rather than only the early-exit one.
	//
	// This is what closes F54's residual gap: a fast double fork, whose
	// intermediate process exits within milliseconds, re-parents its
	// grandchild to launchd before the leader itself ever leaves the watch
	// loop, so the leader exits on its own well inside the bound and
	// timedOut stays false — the exact condition under which the unpatched
	// bash returned "defect not demonstrated" while a live process kept
	// running. The group answers by process group, never by parentage: a
	// process's pgid does not change when its ppid does, so this still finds
	// a process the descendant walk above lost the moment it was re-parented.
	//
	// Taken as a SNAPSHOT before any signal, rather than folded into a
	// separate group-scoped TERM/grace/KILL sequence of its own: the group
	// lookup is a full process-table scan, measurably slower than the
	// `kill -0 <exact pid>` check the retry sweep below already uses, and
	// that extra latency was enough, measured directly, for a just-SIGKILLed
	// process to be reaped before a follow-up group lookup could report it —
	// silently erasing the very survivor this detection exists to name.
	// Feeding the snapshot into the SAME retry sweep as the descendant walk's
	// own kids, below, keeps exactly one kill-and-report mechanism, the one
	// that reads each pid's liveness before its SIGKILL (rrSweep). The
	// leader's pid is the group id throughout: Setsid made them one number.
	var group []int
	for _, p := range read() {
		if p.pgid == pid {
			group = append(group, p.pid)
		}
	}
	// kids is every pid any poll logged, one entry per pid (the deepest depth it
	// was ever seen at — a pid does not change its place in the tree while it
	// stays a descendant, but taking the max is defensive against a reused pid
	// observed at different depths across polls) and ordered DEEPEST FIRST.
	// That ordering is what the kill sweep below relies on: killing a parent
	// before a still-uncollected child would re-orphan that child mid-sweep, the
	// exact defect the walk exists to stop happening by omission.
	//
	// Computed HERE, before the timeout kill sequence below, rather than after
	// it: this is pure data massaging over the logged depths and the pre-kill
	// group snapshot, neither of which any kill below can change, so computing
	// it earlier costs nothing in correctness — and in the bash guard, whose
	// `awk`/`sort` each cost a fork+exec, computing it after the kill measurably
	// lost the race: on Darwin 25.5.0 a same-group descendant killed by a
	// group-wide `kill -KILL -PGID` is typically reaped within roughly a
	// millisecond, which silently emptied the very retry sweep meant to catch
	// the descendant. The escalation kill below is also scoped to the leader
	// alone rather than the whole group for the same measured reason — see the
	// comment there.
	kids := make([]int, 0, len(depth))
	for p := range depth {
		kids = append(kids, p)
	}
	sort.Slice(kids, func(i, j int) bool {
		if depth[kids[i]] != depth[kids[j]] {
			return depth[kids[i]] > depth[kids[j]]
		}
		return kids[i] < kids[j]
	})
	sort.Ints(group)
	// The pre-kill group snapshot is folded into kids here, deduplicated, so
	// the SAME retry-and-report sweep below is the one backstop for both
	// detection paths — the walk still names pids the group pass would only
	// ever report as a bare count. This is also where a class the group
	// deliberately does NOT cover would surface: a reproducer that calls
	// `setsid` itself leaves the group on its own initiative, so the group
	// snapshot cannot see it — only the descendant walk above can, and only if
	// some poll caught it before it re-parented, which is the same residual
	// limit this guard has always had for that one case (see the note below).
	for _, g := range group {
		if _, ok := depth[g]; !ok {
			kids = append(kids, g)
		}
	}

	// The timeout kill sequence itself, deferred to here — strictly after the
	// snapshot and the kids arithmetic above, so that snapshot is provably the
	// first signal-adjacent thing on the timeout path too, exactly as it
	// already was on the early-exit path (nothing above this line sends any
	// signal on either path). The wait for the leader sits below it for the
	// same reason: on the timeout path it must follow the kill, and keeping
	// both paths' wait in one place is what keeps there being a single
	// snapshot point rather than a second one added back beside it.
	//
	// The initial SIGTERM targets the whole group (`-pid`, the pgid form): a
	// courtesy that lets an ordinary, non-detaching descendant exit cleanly.
	// The escalation to SIGKILL, though, is scoped to the leader alone, never
	// the group — this is the second half of closing F4, not an unrelated
	// change: a same-group descendant that ignores SIGTERM (the exact shape the
	// group snapshot exists to catch) used to be reaped by a group-wide
	// `kill -KILL -PGID` before the retry sweep below ever got to it, which
	// measurably outran that sweep's `kill -0` liveness check every time it was
	// tried — the pid was real and the kill was real, but nothing after the
	// kill could still see it, so it was never named. Leaving that pid's kill
	// to the retry sweep instead — the same TERM-then-grace-then-SIGKILL
	// sequence that names a detached survivor (cases 10, 16) — still kills it,
	// just via the one mechanism that reads its liveness before killing it,
	// rather than a second, earlier kill call that only ever wins the race to
	// erase it.
	if timedOut {
		if syscall.Kill(-pid, syscall.SIGTERM) != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
		select {
		case <-done:
		case <-time.After(grace):
		}
		_ = cmd.Process.Kill()
	}
	<-done
	res := rrResult{rc: rrExitCode(cmd.ProcessState), timedOut: timedOut, tableErr: tableErr}

	// A courtesy SIGTERM to the whole process group — "kill by group first",
	// ahead of the per-pid retry sweep below — sent once, unconditionally, and
	// never waited on here: it is a best-effort head start for any group
	// member the retry sweep is about to walk individually anyway, harmless
	// when the group is already empty (a kill on a pgid with nothing left in
	// it simply errors, discarded like every other best-effort signal here).
	// The retry sweep below is what actually waits, escalates to SIGKILL and
	// reports — this line only ever shortens that sweep's own grace window in
	// practice, never replaces it.
	//
	// RESIDUAL GAP, stated honestly rather than implied away. Before the
	// process group (D4), a grandchild that re-parented to launchd before any
	// poll observed it was invisible outright: a parentage walk only ever
	// answers "children of this still-tracked pid right now", and once a
	// process was reparented with no poll having caught it as a descendant
	// first, nothing could walk back to it. The process group established at
	// launch closes that gap for the ordinary case — pgid survives
	// re-parenting even though ppid does not, so the group snapshot above finds
	// it regardless of how fast it detached. What remains open is narrower and
	// different in kind: a reproducer that calls `setsid` (or otherwise calls
	// `setpgid` on itself or a child) LEAVES the group deliberately, on its own
	// initiative, and no group-based lookup can see a process that removed
	// itself from the group being searched. That case still depends on the
	// descendant walk having caught the process before it left — the same limit
	// the parentage walk always had, now scoped to only this one case instead of
	// every fast double fork.
	_ = syscall.Kill(-pid, syscall.SIGTERM)

	// Retry the same SIGTERM-then-grace-then-SIGKILL sequence against every pid
	// any poll or group check above named, then report whichever of them is
	// still alive — checked UNCONDITIONALLY, on every run, not only when
	// timedOut: a child detached before the parent exits normally is exactly as
	// orphaned and exactly as unaccounted-for as one detached before a kill,
	// and the disposition is the same either way. This is the backstop the
	// group kill above falls back to, not a duplicate of it: a pid killed here
	// may already be dead from the group signal, in which case `kill -0` simply
	// finds nothing left to do.
	res.survivors = rrSweep(kids, grace, syscall.Kill)
	return res
}

// rrSweep is the per-pid retry: SIGTERM to every pid of pids still alive,
// the grace, SIGKILL, and the survivors named. kill is syscall.Kill.
func rrSweep(pids []int, grace time.Duration, kill func(int, syscall.Signal) error) (survivors []int) {
	alive := func(p int) bool { return kill(p, 0) == nil }
	var still []int
	for _, k := range pids {
		if alive(k) {
			still = append(still, k)
		}
	}
	for _, k := range still {
		_ = kill(k, syscall.SIGTERM)
	}
	deadline := time.Now().Add(grace)
	for _, k := range still {
		// ponytail: a non-child pid has no portable wait, so liveness is
		// polled under the grace deadline.
		for alive(k) && time.Now().Before(deadline) {
			time.Sleep(rrLivePoll)
		}
	}
	// A pid still alive when the grace ends outlived its SIGTERM: it is a
	// survivor, and its liveness is read HERE, before its SIGKILL, never
	// after. The bash read `kill -0` right after the SIGKILL and counted on
	// the orphan's reaper (launchd) being slower than that read; under load
	// the reaper won, the pid read as gone and the survivor went unnamed
	// (cases 10, 13, 14, 16, 18). Nothing reaps a live process, so the read
	// before the kill cannot lose that race.
	for _, k := range still {
		if alive(k) {
			survivors = append(survivors, k)
		}
		_ = kill(k, syscall.SIGKILL)
	}
	return survivors
}

// rrExitCode is bash's `$?` after `wait`: the exit status, or 128+signal.
func rrExitCode(ps *os.ProcessState) int {
	if ws, ok := ps.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return ps.ExitCode()
}

type rrProc struct{ pid, ppid, pgid int }

// rrDescendants is collect_descendants: every live descendant of root with
// its depth (1 a child), walking the WHOLE tree breadth-first rather than the
// one level `pgrep -P "$root"` alone sees. F54: a real double fork — child
// forks a grandchild, then the child exits immediately — orphans the
// grandchild one hop below what a single-level children lookup can ever
// name; a snapshot taken only at that first level reported a clean verdict
// while the grandchild kept running, unreported, forever. The caller uses the
// depth to kill the deepest processes first, so a mid-sweep kill of a parent
// can never re-orphan a child this same sweep has not reached yet. The walk
// is capped at 64 levels, purely defensively — this repository's fixtures
// never nest that deep — so a pathological or cyclic process table cannot
// spin this loop forever.
func rrDescendants(procs []rrProc, root int) map[int]int {
	children := map[int][]int{}
	for _, p := range procs {
		children[p.ppid] = append(children[p.ppid], p.pid)
	}
	found := map[int]int{}
	frontier := []int{root}
	for d := 1; len(frontier) > 0 && d <= rrMaxDepth; d++ {
		var next []int
		for _, p := range frontier {
			for _, c := range children[p] {
				found[c] = d
				next = append(next, c)
			}
		}
		frontier = next
	}
	return found
}

// rrGitStatusNote is git_status_note: the worktree's status after a kill,
// silent when git cannot answer.
func rrGitStatusNote(worktree string, stderr io.Writer) {
	out, err := exec.Command("git", "-C", worktree, "status", "--porcelain", "--untracked-files=normal").Output()
	if err != nil {
		return
	}
	if s := strings.TrimRight(string(out), "\n"); s != "" {
		fmt.Fprintf(stderr, "run-reproducer: worktree status after the kill:\n%s\n", s)
	} else {
		fmt.Fprintln(stderr, "run-reproducer: worktree status after the kill: clean")
	}
}

// emitCaptured is emit_captured_output: the reproducer's own output between
// fixed delimiter lines, each stream's trailing newlines dropped as command
// substitution drops them; silent when both streams are empty.
//
// The bounce built on a reproducer's result carries "the reproducer's passing
// output" back to whichever review-panel slot raised the finding — but until
// this ran, the caller had nothing to carry: the captured streams were read by
// nothing and deleted, unread. Emitted to the same stream the verdict goes
// to, wrapped in a fixed pair of delimiter lines so a caller can tell where
// the reproducer's own output starts and ends and never mistake it for one of
// this guard's own `run-reproducer:` messages printed around it. Silent when
// the reproducer produced nothing on either stream — an empty pair of
// delimiters would say less than saying nothing.
func emitCaptured(w io.Writer, outPath, errPath string) {
	o, _ := os.ReadFile(outPath)
	e, _ := os.ReadFile(errPath)
	out, errS := strings.TrimRight(string(o), "\n"), strings.TrimRight(string(e), "\n")
	if out == "" && errS == "" {
		return
	}
	fmt.Fprintln(w, "run-reproducer: --- captured reproducer output begin ---")
	if out != "" {
		fmt.Fprintln(w, out)
	}
	if errS != "" {
		fmt.Fprintln(w, "run-reproducer: --- captured reproducer stderr ---")
		fmt.Fprintln(w, errS)
	}
	fmt.Fprintln(w, "run-reproducer: --- captured reproducer output end ---")
}
