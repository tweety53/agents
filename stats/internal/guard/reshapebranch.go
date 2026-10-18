package guard

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// reshapeBranch is scripts/reshape-branch.sh: that script's header is the
// contract — `reset --soft <merge-base>` behind
// check-planning-commit-location (in-process). Every git step's stderr passes
// through and its exit code is the guard's, as `set -euo pipefail` made it.
func init() { Registry["reshape-branch"] = reshapeBranch }

func reshapeBranch(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: reshape-branch.sh <worktree> <name> <merge-base>")
		return 2
	}
	wt, name, base := args[0], args[1], args[2]
	if rc := checkPlanningCommitLocation([]string{wt, name}, env, stdout, stderr); rc != 0 {
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
