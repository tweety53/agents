package guard

import (
	"fmt"
	"io"
)

// commitSplit is scripts/commit-split.sh: that script's header is the
// contract — the guarded commit chain of
// skills/flow-contracts/pipeline.md's "Git boundaries" section, behind
// checkChangeWorktreeLocation (in-process). Every git step's own output passes
// through and its exit code is the guard's, as `set -euo pipefail` made it.
func init() { Registry["commit-split"] = commitSplit }

func commitSplit(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 4 {
		fmt.Fprintln(stderr, "usage: flow-guard commit-split <worktree> <name> <impl-msg> <plan-msg> [<trailer>…]")
		return 2
	}
	worktree, name, implMsg, planMsg, trailers := args[0], args[1], args[2], args[3], args[4:]

	if rc := checkChangeWorktreeLocation([]string{worktree, name}, env, stdout, stderr); rc != 0 {
		return rc
	}

	git := envGit(env)
	run := func(args ...string) int {
		cmd := git(append([]string{"-C", worktree}, args...)...)
		cmd.Stdout, cmd.Stderr = stdout, stderr
		return exitCode(cmd.Run())
	}
	// `git add -A` keeps the change directory out only through the ignore
	// file: a project that does not ignore it would land its plan on the
	// default branch.
	changes := specRootLeaf(worktree, stderr) + "/changes/"
	if exitCode(git("-C", worktree, "check-ignore", "-q", "--no-index", changes).Run()) != 0 {
		fmt.Fprintf(stderr, "commit-split: %s is not gitignored in %s — add it to .gitignore before integrating\n", changes, worktree)
		return 2
	}
	// An ignore rule never removes a path the index already tracks, and a
	// change planned under the old tooling committed its own directory: it is
	// untracked here, and stays on disk.
	if rc := run("rm", "-r", "-q", "-f", "--cached", "--ignore-unmatch", "--", changes+name); rc != 0 {
		return rc
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
	// Each <trailer> — a squashed commit's `Task-Id: <n>` — joins the
	// Flow-Key in the trailer block, so a re-run's Task-Id search still
	// finds every task the squash folded in.
	implArgs := []string{"-m", implMsg, "--trailer", "Flow-Key: flow.commit-two"}
	for _, t := range trailers {
		implArgs = append(implArgs, "--trailer", t)
	}
	if rc := commit(implArgs...); rc != 0 {
		return rc
	}
	if rc := run("add", "-A"); rc != 0 {
		return rc
	}
	return commit("-m", planMsg)
}
