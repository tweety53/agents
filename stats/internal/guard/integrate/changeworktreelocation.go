package integrate

import (
	"fmt"
	"io"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// checkChangeWorktreeLocation refuses a worktree a change's branch is
// rewritten in unless it is a linked worktree on spectre/<name>:
// WORKTREE-LOCATION-MAIN-CHECKOUT and/or WORKTREE-LOCATION-WRONG-BRANCH at exit
// 1, or WORKTREE-LOCATION-OK at exit 0; exit 2 with nothing on stdout
// when it cannot answer. integrate-change's commit calls it in-process
// before it commits: a rewrite in the main checkout, or on another
// change's branch, lands where no step of this change ever sees it.
func checkChangeWorktreeLocation(args []string, env core.Env, stdout, stderr io.Writer) int {
	die := func(id, format string, a ...any) int {
		core.Reason(env, id)
		fmt.Fprintf(stderr, "change-worktree-location: "+format+"\n", a...)
		return 2
	}
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return die("usage", "usage: <worktree> <name>")
	}
	wt, name := args[0], args[1]
	if !core.IsDir(core.Abs(env, wt)) {
		return die("not-a-readable-directory", "not a readable directory: %s", wt)
	}
	git := core.EnvGit(env)
	if inside, _ := core.Capture(git("-C", wt, "rev-parse", "--is-inside-work-tree")); inside != "true" {
		return die("not-a-git-work-tree", "not a git work tree: %s", wt)
	}
	// "MAIN CHECKOUT" IS THE WORKTREE WHOSE GIT DIR IS THE COMMON DIR — the
	// same test flow-guard hook block-unsafe-commands makes. A linked worktree's
	// `--git-dir` is `<common>/worktrees/<id>`, so it differs from
	// `--git-common-dir`; both are asked for with `--path-format=absolute` so
	// a relative `.git` from a main checkout never compares unequal to its own
	// absolute form.
	gitDir, ok := core.Capture(git("-C", wt, "rev-parse", "--path-format=absolute", "--git-dir"))
	if !ok {
		return die("cannot-resolve-the-git-dir", "cannot resolve the git dir of: %s", wt)
	}
	common, ok := core.Capture(git("-C", wt, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	if !ok {
		return die("cannot-resolve-common-git-dir", "cannot resolve the common git dir of: %s", wt)
	}

	// symbolic-ref exits 1 on a detached HEAD, which is an answer here, not an
	// inability. Its stderr is not silenced, as the bash's was not.
	sym := git("-C", wt, "symbolic-ref", "--short", "-q", "HEAD")
	sym.Stderr = stderr
	branch, ok := core.Capture(sym)
	if !ok {
		branch = "detached"
	}

	violation := false
	if gitDir == common {
		fmt.Fprintf(stdout, "WORKTREE-LOCATION-MAIN-CHECKOUT: %s\n", wt)
		violation = true
	}
	if branch != "spectre/"+name {
		fmt.Fprintf(stdout, "WORKTREE-LOCATION-WRONG-BRANCH: %s on %s — expected spectre/%s\n", wt, branch, name)
		violation = true
	}
	if violation {
		core.Reason(env, "worktree-location-wrong")
		return 1
	}
	fmt.Fprintf(stdout, "WORKTREE-LOCATION-OK: %s on spectre/%s\n", wt, name)
	return 0
}
