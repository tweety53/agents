package guard

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// check-untasked-commits lists the commits of <base>..HEAD that sit outside
// every task, for /flow's run-1 unfinished-work gate. Its shim's header
// (scripts/check-untasked-commits.sh) is canonical for the contract.
func init() { Registry["check-untasked-commits"] = checkUntaskedCommits }

const ucPrefix = "check-untasked-commits:"

// ucChoreRE is a Conventional Commits `chore` subject: the pipeline's own
// planning, link and archive commits.
var ucChoreRE = regexp.MustCompile(`^chore(\([^)]*\))?!?:`)

// ucPipelineFixBranchRE is the branch a pipeline-fix Flow-Key belongs on.
var ucPipelineFixBranchRE = regexp.MustCompile(`^pipeline-fix-.+-[0-9]+$`)

func checkUntaskedCommits(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 2 || len(args) > 3 || args[0] == "" || args[1] == "" || (len(args) == 3 && args[2] != "--not-remotes") {
		fmt.Fprintf(stderr, "%s usage: check-untasked-commits.sh <worktree> <base> [--not-remotes]\n", ucPrefix)
		return 2
	}
	worktree, base := args[0], args[1]
	rng := []string{"HEAD", "^" + base}
	if len(args) == 3 {
		// A push sends only what origin lacks: after a rebase onto a moved
		// base, the base's own commits sit in <base>..HEAD but are already on
		// origin, never this change's to judge.
		rng = append(rng, "--not", "--remotes=origin")
	}
	if !isDir(worktree) {
		fmt.Fprintf(stderr, "%s worktree not found: %s\n", ucPrefix, worktree)
		return 2
	}
	git := envGit(env)
	// --first-parent: a merged-in branch's commits are that branch's own,
	// never this change's -- reshape-branch refuses such a merge outright.
	out, ok := capture(git(append([]string{"-C", worktree, "log", "--no-merges", "--first-parent",
		"--format=%H%x00%s%x00%(trailers:key=Task-Id,valueonly,separator=%x2C)%x00%(trailers:key=Flow-Key,valueonly,separator=%x2C)%x1e"},
		rng...)...))
	if !ok {
		fmt.Fprintf(stderr, "%s git cannot read %s..HEAD in %s\n", ucPrefix, base, worktree)
		return 2
	}
	// Empty on a detached HEAD, which is judged as a change branch.
	branch, _ := capture(git("-C", worktree, "symbolic-ref", "--short", "-q", "HEAD"))
	pipelineFixBranch := ucPipelineFixBranchRE.MatchString(strings.TrimSpace(branch))
	planning := specRootLeaf(worktree, io.Discard) + "/changes/"
	var listed []string
	checked := 0
	for _, rec := range strings.Split(out, "\x1e") {
		f := strings.Split(strings.Trim(rec, "\n"), "\x00")
		if len(f) != 4 {
			continue
		}
		checked++
		sha, subject := f[0], f[1]
		if !pipelineFixBranch && ucHasPipelineFixKey(f[3]) {
			listed = append(listed, fmt.Sprintf("  %s %s — Flow-Key: pipeline-fix-… on a change branch", sha[:12], strings.TrimSuffix(sanitizeDisplay(subject), "\n")))
			continue
		}
		if strings.TrimSpace(f[2]) != "" || strings.TrimSpace(f[3]) != "" || ucChoreRE.MatchString(subject) {
			continue
		}
		paths, ok := capture(git("-C", worktree, "diff-tree", "--no-commit-id", "--no-renames", "--name-only", "-r", "--root", sha))
		if !ok {
			fmt.Fprintf(stderr, "%s git cannot list the paths of %s\n", ucPrefix, sha)
			return 2
		}
		if ucPlanningOnly(paths, planning) {
			continue
		}
		listed = append(listed, fmt.Sprintf("  %s %s", sha[:12], sanitizeDisplay(subject)))
	}
	if len(listed) == 0 {
		fmt.Fprintf(stdout, "UNTASKED-OK: %d commit(s) checked\n", checked)
		return 0
	}
	fmt.Fprintf(stdout, "UNTASKED: %d commit(s) carry no Task-Id: or Flow-Key: trailer\n", len(listed))
	for _, l := range listed {
		fmt.Fprintln(stdout, l)
	}
	return 1
}

// ucPlanningOnly is whether every path a commit touched lies under the
// planning tree; a commit touching nothing is not.
func ucPlanningOnly(paths, planning string) bool {
	if paths == "" {
		return false
	}
	for _, p := range strings.Split(paths, "\n") {
		if !strings.HasPrefix(p, planning) {
			return false
		}
	}
	return true
}

// ucHasPipelineFixKey is whether any Flow-Key value opens pipeline-fix-.
func ucHasPipelineFixKey(keys string) bool {
	for _, k := range strings.Split(keys, ",") {
		if strings.HasPrefix(strings.TrimSpace(k), "pipeline-fix-") {
			return true
		}
	}
	return false
}
