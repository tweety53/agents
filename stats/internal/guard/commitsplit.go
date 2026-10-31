package guard

import (
	"fmt"
	"io"
	"os"
)

// commitSplit is scripts/commit-split.sh: that script's header is the
// contract — the guarded two-commit chain of
// skills/flow-contracts/pipeline.md's "Git boundaries" section, behind
// check-planning-commit-location (in-process). Every git step's own output passes
// through and its exit code is the guard's, as `set -euo pipefail` made it.
func init() { Registry["commit-split"] = commitSplit }

func commitSplit(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 4 {
		fmt.Fprintln(stderr, "usage: commit-split.sh <worktree> <name> <impl-msg> <plan-msg>")
		return 2
	}
	worktree, name, implMsg, planMsg := args[0], args[1], args[2], args[3]
	planDir := specRootLeaf(pcAbs(env, worktree), stderr) + "/changes/"

	if rc := checkPlanningCommitLocation([]string{worktree, name}, env, stdout, stderr); rc != 0 {
		return rc
	}

	git := envGit(env)
	run := func(args ...string) int {
		cmd := git(append([]string{"-C", worktree}, args...)...)
		cmd.Stdout, cmd.Stderr = stdout, stderr
		return exitCode(cmd.Run())
	}
	// Planning is the change's own folder and each <name>-fix-<N> sub-change
	// folder the worktree holds, never the whole changes tree: every other
	// path under it — an archive deletion, another change's directory or
	// link.md — is implementation (kan-1072 self-review, F1). The reset
	// unstages the whole tree, so whatever the exclude-add leaves out is
	// what the planning commit's final `add -A` picks up.
	//
	// A tracked symlink at a planning path makes the exclude-add exit 128
	// (the header's tracked-symlink paragraph): its code and message are
	// returned as-is, with no retry or reinterpretation around the call.
	//
	// The change's link.md is re-added after the reset, never before, so the
	// reset never strips it back out; it needs its own add because a
	// `:(exclude)` pathspec always wins over a positive pathspec in the same
	// call (verified by running git). `git add -A -- <pathspec>` exits 128
	// when the pathspec matches nothing, so the file is looked for first —
	// most commits touch no link.md, and that must stay a no-op.
	abs := pcAbs(env, worktree) + "/" + planDir
	add := []string{"add", "-A", "--", ".", ":(exclude)" + planDir + name}
	if entries, err := os.ReadDir(abs); err == nil {
		for _, e := range entries {
			if isFixOf(e.Name(), name) {
				add = append(add, ":(exclude)"+planDir+e.Name())
			}
		}
	}
	steps := [][]string{{"reset", "-q", "--", planDir}, add}
	if _, err := os.Lstat(abs + name + "/link.md"); err == nil {
		steps = append(steps, []string{"add", "-A", "--", planDir + name + "/link.md"})
	}
	for _, s := range steps {
		if rc := run(s...); rc != 0 {
			return rc
		}
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
