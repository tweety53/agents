package guard

import (
	"fmt"
	"io"
)

// checkPlanningCommitLocation is scripts/check-planning-commit-location.sh:
// that script's header comment is the contract -- PLANNING-COMMIT-MAIN-
// CHECKOUT and/or PLANNING-COMMIT-WRONG-BRANCH at exit 1, or
// PLANNING-COMMIT-LOCATION-OK at exit 0; exit 2 with nothing on stdout when
// it cannot answer.
//
// `spectre link` writes `link.md` on both sides of a cross-repo change, and
// it is run with the working directory at a repository's PRIMARY checkout so
// its peers file resolves — which is exactly where a link commit made from
// the wrong directory lands. A planning commit in the main checkout, or on
// any branch but `spectre/<name>`, puts change-folder content on the landing
// target (or on another change's branch) where no reshape, review or archive
// step of this change ever sees it. Every planning commit therefore runs
// behind this guard; git-boundaries.md's **Planning commits** is canonical
// for where it is called.
func init() { Registry["check-planning-commit-location"] = checkPlanningCommitLocation }

func checkPlanningCommitLocation(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-planning-commit-location.sh: "+format+"\n", a...)
		return 2
	}
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return die("usage: check-planning-commit-location.sh <worktree> <name>")
	}
	wt, name := args[0], args[1]
	if !isDir(smcAbs(env, wt)) {
		return die("not a readable directory: %s", wt)
	}
	git := envGit(env)
	if inside, _ := capture(git("-C", wt, "rev-parse", "--is-inside-work-tree")); inside != "true" {
		return die("not a git work tree: %s", wt)
	}
	// "MAIN CHECKOUT" IS THE WORKTREE WHOSE GIT DIR IS THE COMMON DIR — the
	// same test hooks/protect-main-checkout.py makes. A linked worktree's
	// `--git-dir` is `<common>/worktrees/<id>`, so it differs from
	// `--git-common-dir`; both are asked for with `--path-format=absolute` so
	// a relative `.git` from a main checkout never compares unequal to its own
	// absolute form.
	gitDir, ok := capture(git("-C", wt, "rev-parse", "--path-format=absolute", "--git-dir"))
	if !ok {
		return die("cannot resolve the git dir of: %s", wt)
	}
	common, ok := capture(git("-C", wt, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	if !ok {
		return die("cannot resolve the common git dir of: %s", wt)
	}

	// symbolic-ref exits 1 on a detached HEAD, which is an answer here, not an
	// inability. Its stderr is not silenced, as the bash's was not.
	sym := git("-C", wt, "symbolic-ref", "--short", "-q", "HEAD")
	sym.Stderr = stderr
	branch, ok := capture(sym)
	if !ok {
		branch = "detached"
	}

	violation := false
	if gitDir == common {
		fmt.Fprintf(stdout, "PLANNING-COMMIT-MAIN-CHECKOUT: %s\n", wt)
		violation = true
	}
	if branch != "spectre/"+name {
		fmt.Fprintf(stdout, "PLANNING-COMMIT-WRONG-BRANCH: %s on %s — expected spectre/%s\n", wt, branch, name)
		violation = true
	}
	if violation {
		return 1
	}
	fmt.Fprintf(stdout, "PLANNING-COMMIT-LOCATION-OK: %s on spectre/%s\n", wt, name)
	return 0
}
