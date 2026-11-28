package guard

// flow-guard design-repo — the project's `design repo` row, read from the
// checkout it is given (a change's worktree, so the row of the change's own
// base branch), never from another.
//
// Usage: design-repo <worktree>
//
// Exit 0 prints `<clone> <remote>`, the clone resolved against <worktree>;
// 1 no `design repo` row of two fields under `## visual verification`;
// 2 cannot answer: a usage error, <worktree> not a directory, or its
// .flow/project.md unreadable.

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/gitops"
)

func init() { Registry["design-repo"] = designRepo }

// designRepoRow is the one parser of the row, shared with render-frames.
func designRepoRow(root string, cells map[string]string) (clone, remote string, ok bool) {
	f := strings.Fields(strings.ReplaceAll(cells["design repo"], "`", ""))
	if len(f) != 2 {
		return "", "", false
	}
	return cellPath(root, f[0]), f[1], true
}

func designRepo(args []string, env core.Env, stdout, stderr io.Writer) int {
	exit := func(code int, id, format string, a ...any) int {
		core.Reason(env, id)
		fmt.Fprintf(stderr, "design-repo: "+format+"\n", a...)
		return code
	}
	if len(args) != 1 {
		return exit(2, "usage", "usage: flow-guard design-repo <worktree>")
	}
	root := gitops.WorkspaceAbs(env, args[0])
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return exit(2, "worktree-not-a-directory", "%s is not a directory", args[0])
	}
	cfg, err := os.ReadFile(filepath.Join(root, ".flow", "project.md"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return exit(2, "cannot-read", "cannot read the project configuration: %v", err)
	}
	clone, remote, ok := designRepoRow(root, vvCells(string(cfg)))
	if !ok {
		return exit(1, "design-repo-row-missing", "%s/.flow/project.md declares no `design repo` row of `<clone> <remote>` under `## visual verification`", args[0])
	}
	fmt.Fprintln(stdout, clone, remote)
	return 0
}
