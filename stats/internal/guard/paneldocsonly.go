package guard

import (
	"fmt"
	"io"
	"path"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-panel-docs-only.sh shim (KAN-1058).
//
// flow-guard check-panel-docs-only — decide whether this change's own paths are
// entirely documentation, in a guard instead of prose.
//
// Usage: flow-guard check-panel-docs-only <worktree> <merge-base>
//
// Exit codes:
//   0  every touched path ends .md or .mdc, or is a change's
//      verbatim-moves.txt under <spec-root>/changes/ — the acknowledgement
//      list flow-guard check-verbatim-moves reads, documentation though not Markdown
//   1  at least one touched path is neither — that first path is printed to
//      stdout — including an empty touched-path set
//   2  cannot answer: a missing argument, <worktree> not a directory or not
//      a git worktree, <merge-base> not resolving, or a git invocation
//      failed
//
// THE CHANGE'S OWN PATHS are the union of committed-since-merge-base,
// staged and unstaged paths, exactly as flow-guard check-panel-citation-trigger
// collects them.
//
// WHY THIS GUARD EXISTS (KAN-312): a docs-only branch has no code seam for a
// second slot to find; the implementer's self-review and the prose guards
// cover it.
//
// The logic is the Go port in stats/internal/guard/paneldocsonly.go; the
// touched-path collection is stats/internal/guard/paneltouchedpaths.go's.

// checkPanelDocsOnly is `flow-guard check-panel-docs-only`: the comment above is the contract -- exit 0 when every path this change
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
