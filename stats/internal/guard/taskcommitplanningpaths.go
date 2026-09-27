package guard

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// checkTaskCommitPlanningPaths is scripts/check-task-commit-planning-paths.sh:
// that script's header comment is the contract -- fail any task commit in
// <base>..HEAD that swept the planning paths (the spec tree's changes
// directory, leaf resolved through specroot.go, and docs/superpowers/).
// Exit 0 clean, 1 swept, 2 with nothing on stdout when it cannot answer.
func init() {
	Registry["check-task-commit-planning-paths"] = checkTaskCommitPlanningPaths
}

func checkTaskCommitPlanningPaths(args []string, env Env, stdout, stderr io.Writer) int {
	const p = "check-task-commit-planning-paths.sh: "
	refuse := func(msg string) int {
		fmt.Fprintln(stderr, p+msg)
		return 2
	}
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: check-task-commit-planning-paths.sh <worktree> <base>")
		return 2
	}
	wt, base := args[0], args[1]
	if !isDir(wt) {
		return refuse("not a readable directory: " + wt)
	}
	git := envGit(env)
	if git("-C", wt, "rev-parse", "--is-inside-work-tree").Run() != nil {
		return refuse("not a git repository: " + wt)
	}
	verify := func(rev string) bool {
		cmd := git("-C", wt, "rev-parse", "--verify", "--quiet", rev)
		cmd.Stderr = stderr
		return cmd.Run() == nil
	}
	if !verify("HEAD") {
		return refuse("HEAD does not resolve in: " + wt)
	}
	if !verify(base + "^{commit}") {
		return refuse("base does not resolve to a commit: " + base)
	}

	leaf := specRootLeaf(wt, stderr)

	checked, swept := 0, 0
	var sweepLines strings.Builder

	// Both walks run as captured commands whose failure is tested: a
	// discarded git failure printed a clean verdict the guard could not know.
	// Git's own stderr is left visible on a refusal — the guard's line names
	// the stage, git's names the cause. One line per commit: `<sha> <Task-Id
	// value>` — line-wise, not NUL-separated: `git log` emits a newline
	// between entries whatever the format ends with, and a NUL-split record
	// put that newline at the head of every sha after the first — diff-tree
	// refused it and the swept commit went unflagged. A commit with no
	// Task-Id trailer reads as an empty value and is not a task commit.
	logCmd := git("-C", wt, "log", "--format=%H %(trailers:key=Task-Id,valueonly)", base+"..HEAD")
	logCmd.Stderr = stderr
	logWalk, ok := capture(logCmd)
	if !ok {
		return refuse("git log refused the walk in: " + wt)
	}
	for _, line := range strings.Split(logWalk, "\n") {
		if line == "" {
			continue
		}
		sha, _, _ := strings.Cut(line, " ")
		rest := line
		if _, r, found := strings.Cut(line, " "); found {
			rest = r
		}
		// `tr -d '[:space:]'`.
		tid := strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, rest)
		if tid == "" {
			continue
		}
		checked++
		// -c: without a merge flag diff-tree prints nothing for merges, so a
		// Task-Id evil merge was counted and never diffed, answering CLEAN
		// over its smuggled paths (KAN-553 F3, deferred; KAN-607).
		dt := git("-C", wt, "diff-tree", "--no-renames", "-r", "--no-commit-id", "--name-only", "-c", sha)
		dt.Stderr = stderr
		paths, ok := capture(dt)
		if !ok {
			return refuse("git diff-tree refused the walk on " + sha)
		}
		// Newline-separated, deliberately without -z, as the bash read it.
		commitSwept := false
		for _, path := range strings.Split(paths, "\n") {
			if strings.HasPrefix(path, leaf+"/changes/") || strings.HasPrefix(path, "docs/superpowers/") {
				fmt.Fprintf(&sweepLines, "TASK-COMMIT-SWEEP: %.12s %s %s\n", sha, tid, path)
				commitSwept = true
			}
		}
		if commitSwept {
			swept++
		}
	}

	if swept > 0 {
		fmt.Fprint(stdout, sweepLines.String())
		fmt.Fprintf(stdout, "PLANNING-PATHS-SWEPT: %s — %d task commit(s)\n", wt, swept)
		return 1
	}
	fmt.Fprintf(stdout, "PLANNING-PATHS-CLEAN: %s — %d task commit(s) checked\n", wt, checked)
	return 0
}
