package guard

import (
	"fmt"
	"io"
)

// commitSplit is scripts/commit-split.sh: that script's header is the
// contract — the guarded two-commit chain of
// skills/flow-contracts/pipeline.md's "Git boundaries" section, behind
// checkChangeWorktreeLocation (in-process). Every git step's own output passes
// through and its exit code is the guard's, as `set -euo pipefail` made it.
func init() { Registry["commit-split"] = commitSplit }

func commitSplit(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 4 {
		fmt.Fprintln(stderr, "usage: commit-split.sh <worktree> <name> <impl-msg> <plan-msg>")
		return 2
	}
	worktree, name, implMsg, planMsg := args[0], args[1], args[2], args[3]

	if rc := checkChangeWorktreeLocation([]string{worktree, name}, env, stdout, stderr); rc != 0 {
		return rc
	}

	git := envGit(env)
	run := func(args ...string) int {
		cmd := git(append([]string{"-C", worktree}, args...)...)
		cmd.Stdout, cmd.Stderr = stdout, stderr
		return exitCode(cmd.Run())
	}
	if rc := run("add", "-A", "--", "."); rc != 0 {
		return rc
	}
	// `git diff --cached --quiet || git commit`: any non-zero diff — changes
	// staged, or diff itself failing — goes on to the commit.
	commit := func(msg ...string) int {
		if run("diff", "--cached", "--quiet") != 0 {
			return run(append([]string{"commit"}, msg...)...)
		}
		return 0
	}
	// The implementation commit is outside every task, so it carries the
	// stage key that runs this chain (git-boundaries.md, "Commits outside
	// every task") — else run 1's gate lists it on a re-run.
	if rc := commit("-m", implMsg, "-m", "Flow-Key: flow.commit-two"); rc != 0 {
		return rc
	}
	if rc := run("add", "-A"); rc != 0 {
		return rc
	}
	return commit("-m", planMsg)
}
