package guard

import (
	"fmt"
	"io"
	"path"
	"strings"
)

// checkPanelDocsOnly is scripts/check-panel-docs-only.sh: that script's
// header comment is the contract -- exit 0 when every path this change
// touched ends .md or .mdc or is a change's verbatim-moves.txt under
// <spec-root>/changes/, 1 with the first path that is neither printed to
// stdout (an empty touched-path set is 1 with nothing printed), 2 on a usage
// error or a git invocation that failed. Resolving git, validating the
// arguments and collecting the paths -- the union of committed-since-merge-
// base, staged and unstaged, exactly as check-panel-citation-trigger
// collects them -- are paneltouchedpaths.go's, never re-derived here, so the
// "first" path is the first after that collection's `sort -u` under the
// caller's collation.
func init() {
	Registry["check-panel-docs-only"] = checkPanelDocsOnly
}

func checkPanelDocsOnly(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-panel-docs-only"
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
	changes := specRootLeaf(smcAbs(env, worktree), stderr) + "/changes/"
	for _, p := range paths {
		// A verbatim-moves.txt is check-verbatim-moves.sh's acknowledgement
		// list of reworded prose: documentation, though not Markdown.
		// isDocPath stays `.mdc?$`, shared with plan-class.
		if !isDocPath(p) && !(strings.HasPrefix(p, changes) && path.Base(p) == "verbatim-moves.txt") {
			fmt.Fprintln(stdout, p)
			return 1
		}
	}
	if len(paths) == 0 {
		return 1
	}
	return 0
}
