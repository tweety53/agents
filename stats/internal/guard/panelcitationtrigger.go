package guard

import "io"

// checkPanelCitationTrigger is scripts/check-panel-citation-trigger.sh: that
// script's header comment is the contract -- nothing printed, exit 0 when
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
