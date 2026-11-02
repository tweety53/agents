package guard

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// The contract below moved here from the header of the deleted
// scripts/check-worktree-processes.sh shim (KAN-1058).
//
// flow-guard check-worktree-processes — report every process whose working directory is
// at or under a worktree, so /flow's cleanup run can verify the project's stack
// actually stopped before it removes anything.
//
// Usage: flow-guard check-worktree-processes <worktree>
//
// Prints ONE verdict line to stdout:
//   CLEAR: <worktree>                       no process has a cwd at or under it
//   HELD: <worktree> — <pid> <cwd>[; ...]   one or more processes do
//
// Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at all —
// lsof absent, the wrong number of arguments, a worktree path that is not a
// readable directory, or an lsof invocation that itself failed. The VERDICT
// carries the answer, not the exit status; flow-guard check-cleanup-complete's header
// records why this repository separates the two. A refusal writes its reason to
// stderr and NOTHING to stdout, so a caller grepping for CLEAR in empty output
// can never read an inability as a pass.
//
// WHY A WORKING DIRECTORY AND NOT A PORT. A port proves nothing: a live
// workspace legitimately holds one, and a stack that has genuinely stopped can
// still leave a port in TIME_WAIT. What a process launched from a worktree has,
// and what ties it to the directory about to be removed, is its working
// directory — which is also why removing the worktree is what makes the process
// unreachable. gymie's own OrphanScan.kt reaches the same conclusion
// independently.
//
// WHY THIS ASKS THE OPERATING SYSTEM AND NOT THE PROJECT. The records a project
// keeps about what it started live inside the worktree being removed, so they
// are exactly the evidence a removal destroys. Reading them to decide whether
// the removal is safe is circular. This guard reads the process table instead.
//
// PROJECT-AGNOSTIC BY CONSTRUCTION. It holds no port list, invokes no build
// tool, and requires no project to declare anything, so it protects every
// project flow is installed into rather than only those that opted in. That
// is deliberate: a guard needing an opt-in would not have protected the project
// this incident happened in.
//
// NO sudo, AND THE LIMIT THAT FOLLOWS FROM IT. lsof without privileges reports
// the invoking user's own processes, which is the correct set — a worktree's
// stack is started by the operator, in the operator's own session. The limit is
// recorded here rather than left to be discovered: a process owned by another
// user, or a root-owned one, is outside what this guard can see, and a CLEAR
// verdict says nothing about those.
//
// THE CALLER'S OWN SHELL COUNTS. A shell whose cwd is inside the worktree is a
// process with a working directory at or under it, and this guard reports it
// like any other — it cannot tell the operator's own `cd` from a stray daemon.
// Run it from outside the worktree.
//
// PATHS ARE COMPARED IN THEIR PHYSICAL FORM. lsof reports resolved paths, so on
// macOS a worktree under /tmp reaches it as /private/tmp/... The worktree
// argument is resolved the same way (the `pwd -P` form) before anything is
// compared, and the verdict line echoes that resolved form rather than the
// argument, so the path in the report is the path the comparison used. Skipping
// this makes every symlinked worktree report CLEAR while holding a process.
//
// "AT OR UNDER" IS NOT "STARTS WITH". A path matches when it equals the worktree
// or begins with the worktree plus a slash. A bare string-prefix test reports
// .../spectre-kan-1 as held by a process running in .../spectre-kan-12, which
// is a false block on a worktree the operator cannot clear — and this pipeline's
// own branch naming produces exactly those pairs.
//
// THE lsof EXIT STATUS IS READ, NEVER THE EMPTY OUTPUT. An lsof that failed and
// an lsof that found nothing both print nothing, and treating them alike would
// turn every failure into a CLEAR verdict — the inversion this guard exists to
// prevent, one layer down.

// checkWorktreeProcesses is scripts/check-worktree-processes.sh: that
// script's header comment is the contract -- one CLEAR/HELD verdict line on
// stdout, exit 0 whenever a verdict was reached, exit 2 with the reason on
// stderr and nothing on stdout when it cannot answer. The reasoning for each
// step, moved here from the bash body it replaced (c9b8fde8), sits beside the
// code it explains.
func init() {
	Registry["check-worktree-processes"] = checkWorktreeProcesses
}

func checkWorktreeProcesses(args []string, env Env, stdout, stderr io.Writer) int {
	return worktreeProcesses(args, env, stdout, stderr, wpAtOrUnder)
}

// wpAtOrUnder is the guard's one decision: path equals root, or begins with
// root plus a slash. Both are needed — the first for a process sitting in the
// worktree root, the second for one in any descendant — and root is compared
// as a literal string, so a worktree path holding a glob character cannot
// widen the match.
func wpAtOrUnder(path, root string) bool {
	return path == root || strings.HasPrefix(path, root+"/")
}

// worktreeProcesses is the guard with its at-or-under decision passed in, so
// the KAN-197 mutation case can force it false.
//
// The bash opened `set -uo pipefail` without `-e`, because its lsof
// assignment read `$?` on the next line; every failure here returns 2
// explicitly through die, so the exit-2 refusals are deliberate rather than
// inherited.
func worktreeProcesses(args []string, env Env, stdout, stderr io.Writer, atOrUnder func(path, root string) bool) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-worktree-processes: "+format+"\n", a...)
		return 2
	}
	if len(args) != 1 {
		return die("usage: flow-guard check-worktree-processes <worktree>")
	}
	wt := args[0]

	// lsof is this guard's only source of an answer, so its absence is an
	// inability and never an empty result. The message names the tool,
	// because "the check failed" leaves the operator nothing to install.
	lsof, ok := lookPath(env, "lsof")
	if !ok {
		return die("lsof is not on PATH — this guard has no other way to read the process table")
	}

	abs := smcAbs(env, wt)
	if wt == "" || !isDir(abs) {
		return die("not a directory, so it cannot be scanned: %s", wt)
	}

	// The bash's `cd -- "$WT" && pwd -P`: cd fails on a directory the caller
	// cannot search, which the directory test above does not catch — a mode
	// the operator set, or one inherited from a parent — so the search bit is
	// checked before the path is resolved to its physical form.
	phys := ""
	if syscall.Access(abs, 1) == nil {
		if a, err := filepath.Abs(abs); err == nil {
			phys, _ = filepath.EvalSymlinks(a)
		}
	}
	if phys == "" {
		return die("cannot resolve the worktree to a physical path — unreadable directory: %s", wt)
	}

	// One pass over every process's working directory. -Fp emits `p<pid>`,
	// -Fn emits `n<path>`, and lsof always emits the `f<fd>` line that opens
	// each file record; with -d cwd the only file record a process has is its
	// working directory, so every `n` line belongs to the `p` line above it.
	// lsof's stderr is discarded, as the bash's 2>/dev/null did; its exit
	// status is read, never its empty output.
	cmd := exec.Command(lsof, "-d", "cwd", "-Fpn")
	cmd.Dir = env.Dir
	out, err := cmd.Output()
	if rc := exitCode(err); rc != 0 {
		return die("lsof -d cwd -Fpn exited %d — the process table could not be read, which is not the same as finding nothing", rc)
	}

	var held []string
	pid := ""
	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "p"):
			pid = line[1:]
		case strings.HasPrefix(line, "n"):
			if path := line[1:]; atOrUnder(path, phys) {
				held = append(held, pid+" "+path)
			}
		}
	}

	if len(held) > 0 {
		fmt.Fprintf(stdout, "HELD: %s — %s\n", phys, strings.Join(held, "; "))
	} else {
		fmt.Fprintf(stdout, "CLEAR: %s\n", phys)
	}
	return 0
}
