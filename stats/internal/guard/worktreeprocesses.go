package guard

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

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
		return die("usage: check-worktree-processes.sh <worktree>")
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
