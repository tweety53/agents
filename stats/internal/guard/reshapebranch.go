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
	reset := git("-C", wt, "reset", "-q", "--soft", baseSHA)
	reset.Stdout, reset.Stderr = stderr, stderr
	if rc := exitCode(reset.Run()); rc != 0 {
		return rc
	}
	fmt.Fprintf(stdout, "RESHAPED: %s — reset to %s\n", wt, baseSHA[:12])
	return 0
}
