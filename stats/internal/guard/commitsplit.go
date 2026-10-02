package guard

import (
	"fmt"
	"io"
	"os"
	"strings"
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
	// A tracked symlink at a planning path makes the exclude-add exit 128
	// (the header's tracked-symlink paragraph): its code and message are
	// returned as-is, with no retry or reinterpretation around the call.
	//
	// link.md is re-added after the reset, never before, so the reset never
	// strips it back out; it needs its own add because a `:(exclude)` pathspec
	// always wins over a positive pathspec in the same call (verified by
	// running git). `git add -A -- <pathspec>` exits 128 when the pathspec
	// matches nothing, so the files are looked for first — most commits touch
	// no link.md, and that must stay a no-op.
	steps := [][]string{
		{"reset", "-q", "--", planDir},
		{"add", "-A", "--", ".", ":(exclude)" + planDir},
	}
	if hasChangeLinkMD(pcAbs(env, worktree) + "/" + planDir) {
		steps = append(steps, []string{"add", "-A", "--", planDir + "*/link.md"})
	}
	for _, s := range steps {
		if rc := run(s...); rc != 0 {
			return rc
		}
	}
	// `git diff --cached --quiet || git commit`: any non-zero diff — changes
	// staged, or diff itself failing — goes on to the commit.
	commit := func(msg string) int {
		if run("diff", "--cached", "--quiet") != 0 {
			return run("commit", "-m", msg)
		}
		return 0
	}
	if rc := commit(implMsg); rc != 0 {
		return rc
	}
	if rc := run("add", "-A"); rc != 0 {
		return rc
	}
	return commit(planMsg)
}

// hasChangeLinkMD is bash's nullglob `<planDir>*/link.md` having a match:
// some non-dot entry of planDir holds a link.md (a dangling symlink counts,
// as a glob's directory read counts it).
func hasChangeLinkMD(planDir string) bool {
	entries, err := os.ReadDir(planDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if _, err := os.Lstat(planDir + e.Name() + "/link.md"); err == nil {
			return true
		}
	}
	return false
}
