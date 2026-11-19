package guard

import (
	"fmt"
	"io"
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
//   0  every touched path ends .md or .mdc, or is an .html under docs/
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
// touched ends .md or .mdc, or is an .html under docs/, 1 with the first path that is neither printed to
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
		Reason(env, "invalid-worktree")
		return 2
	}
	paths, ok := panelTouchedPaths(env, prog, worktree, mergebase, gitBin, stderr)
	if !ok {
		Reason(env, "cannot-list-touched-paths")
		return 2
	}
	for _, p := range paths {
		// isDocPath is the one path classifier, shared with
		// check-panel-citation-trigger and fix-run-floor.
		if !isDocPath(p) {
			fmt.Fprintln(stdout, p)
			Reason(env, "non-doc-path-touched")
			return 1
		}
	}
	if len(paths) == 0 {
		Reason(env, "no-path-touched")
		return 1
	}
	return 0
}

// isDocPath reports a documentation path: `\.mdc?$` anywhere, or an .html
// under docs/ -- a rendered brief or design page, which no build, test or
// served page reads. An .html anywhere else (stats/web/index.html) is source.
func isDocPath(p string) bool {
	return strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".mdc") ||
		(strings.HasPrefix(p, "docs/") && strings.HasSuffix(p, ".html"))
}
