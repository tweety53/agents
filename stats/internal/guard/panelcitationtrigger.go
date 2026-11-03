package guard

import "io"

// The contract below moved here from the header of the deleted
// scripts/check-panel-citation-trigger.sh shim (KAN-1058).
//
// flow-guard check-panel-citation-trigger — decide whether the review panel's
// citation pre-check should run, in a guard instead of prose.
//
// Usage: flow-guard check-panel-citation-trigger <worktree> <merge-base>
//
// Prints nothing; the exit code is the whole answer.
//   0  the change's own paths include at least one path ending .md or .mdc
//   1  none do
//   2  usage error: a missing argument, <worktree> not a directory or not a
//      git worktree, <merge-base> not resolving, or a git invocation failed
//
// THE CHANGE'S OWN PATHS INCLUDE THE INDEX AND THE WORKING TREE (design.md:
// touched-paths-include-index-and-worktree, the same rule flow-guard check-base-moved
// already applies): the union of what HEAD carries since the merge base,
// what is staged, and what is unstaged.
//
// The logic is the Go port in stats/internal/guard/panelcitationtrigger.go.

// checkPanelCitationTrigger is `flow-guard check-panel-citation-trigger`: the comment above is the contract -- nothing printed, exit 0 when
// the change's own paths include one ending .md or .mdc, 1 when none do, 2
// on a usage error or a git invocation that failed. The change's own paths
// include the index and the working tree (design.md:
// touched-paths-include-index-and-worktree); resolving git, validating the
// arguments and collecting the paths are paneltouchedpaths.go's, shared with
// plan-class.
func init() {
	Registry["check-panel-citation-trigger"] = checkPanelCitationTrigger
}

func checkPanelCitationTrigger(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-panel-citation-trigger"
	var worktree, mergebase string
	if len(args) > 0 {
		worktree = args[0]
	}
	if len(args) > 1 {
		mergebase = args[1]
	}
	gitBin, ok := panelResolveGit(env, prog, stderr)
	if !ok || !panelValidateWorktree(env, prog, worktree, mergebase, gitBin, stderr) {
		return 2
	}
	paths, ok := panelTouchedPaths(env, prog, worktree, mergebase, gitBin, stderr)
	if !ok {
		return 2
	}
	for _, p := range paths {
		if isDocPath(p) {
			return 0
		}
	}
	return 1
}
