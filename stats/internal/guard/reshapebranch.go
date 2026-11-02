package guard

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/reshape-branch.sh shim (KAN-1058).
//
// flow-guard reshape-branch — integrate's reshape: fold every commit a change branch
// made since <merge-base> — task, fixup and planning commits alike — back
// into the index with `git reset --soft <merge-base>`, for flow-guard commit-split to
// commit on top as the implementation commit and the one planning commit.
//
// Usage: flow-guard reshape-branch <worktree> <name> <merge-base>
//
// skills/flow-contracts/finish-contract-run1.md is canonical for the reshape;
// this script is what runs it. Only HEAD moves: the index and working tree
// keep everything they held, operator edits at the human gate included.
//
// A reset rewrites the change's branch, so it runs the in-process
// change-worktree location check first and stops on its verdict, printing
// its own lines.
//
// A merge that brought in commits <merge-base> does not reach stops the
// reshape before HEAD moves (kan-941: a develop merge folded develop's
// commits into the change commit): the pipeline never merges into a change
// branch, so every commit of <merge-base>..HEAD off the first-parent chain
// is another branch's. Prints `RESHAPE-FOREIGN-MERGE: <worktree> — <n>
// commit(s) …`, one `  <sha12> <subject>` line per such commit and the
// remedy line, and exits 1.
//
// Prints `RESHAPED: <worktree> — reset to <short sha>` and exits 0. Exits 2
// with a reason on stderr on bad arguments or a base that does not resolve,
// the guard's own code on a guard refusal, and git's own code on a git
// failure.

// reshapeBranch is scripts/reshape-branch.sh: that script's header is the
// contract — `reset --soft <merge-base>` behind
// checkChangeWorktreeLocation (in-process). Every git step's stderr passes
// through and its exit code is the guard's, as `set -euo pipefail` made it.
func init() { Registry["reshape-branch"] = reshapeBranch }

func reshapeBranch(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: flow-guard reshape-branch <worktree> <name> <merge-base>")
		return 2
	}
	wt, name, base := args[0], args[1], args[2]
	if rc := checkChangeWorktreeLocation([]string{wt, name}, env, stdout, stderr); rc != 0 {
		return rc
	}
	git := envGit(env)
	var b bytes.Buffer
	cmd := git("-C", wt, "rev-parse", "--verify", "--quiet", base+"^{commit}")
	cmd.Stdout, cmd.Stderr = &b, stderr
	if exitCode(cmd.Run()) != 0 {
		fmt.Fprintf(stderr, "reshape-branch.sh: base does not resolve to a commit: %s\n", base)
		return 2
	}
	baseSHA := strings.TrimRight(b.String(), "\n")
	// A FOREIGN MERGE STOPS THE RESHAPE (kan-941): a commit a merge's side
	// parent brought that <merge-base> does not reach would fold into the
	// change's implementation commit under the change's subject. The
	// pipeline never merges into a change branch -- its sync rebases -- so
	// any such commit is foreign: every commit of <merge-base>..HEAD off
	// the first-parent chain.
	all, ok1 := capture(git("-C", wt, "rev-list", baseSHA+"..HEAD"))
	own, ok2 := capture(git("-C", wt, "rev-list", "--first-parent", baseSHA+"..HEAD"))
	if !ok1 || !ok2 {
		fmt.Fprintf(stderr, "reshape-branch.sh: git cannot list %s..HEAD in %s\n", baseSHA[:12], wt)
		return 2
	}
	mine := map[string]bool{}
	for _, c := range strings.Fields(own) {
		mine[c] = true
	}
	var foreign []string
	for _, c := range strings.Fields(all) {
		if !mine[c] {
			foreign = append(foreign, c)
		}
	}
	if len(foreign) > 0 {
		fmt.Fprintf(stdout, "RESHAPE-FOREIGN-MERGE: %s — %d commit(s) a merge brought in that %s does not reach would fold into the change commit:\n", wt, len(foreign), baseSHA[:12])
		for _, c := range foreign {
			subj, _ := capture(git("-C", wt, "log", "-1", "--format=%s", c))
			fmt.Fprintf(stdout, "  %s %s\n", c[:12], sanitizeDisplay(strings.TrimSpace(subj)))
		}
		fmt.Fprintln(stdout, "  merge that branch into the change's base first, or rebase the change branch onto a base that holds them, then re-run")
		return 1
	}
	reset := git("-C", wt, "reset", "-q", "--soft", baseSHA)
	reset.Stdout, reset.Stderr = stderr, stderr
	if rc := exitCode(reset.Run()); rc != 0 {
		return rc
	}
	fmt.Fprintf(stdout, "RESHAPED: %s — reset to %s\n", wt, baseSHA[:12])
	return 0
}
