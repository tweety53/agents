package guard

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
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
//
// Run as the script, it also refuses a commit appending a task already
// ticked: checkPlanningCommit below. Reshape and commit-split call
// checkPlanningCommitLocation alone, since their planning half folds every
// planning commit of the run into one, ticks included.
func init() { Registry["check-planning-commit-location"] = checkPlanningCommit }

// pclTask is one task line of a tasks.md diff hunk: its sign, its checkbox
// and its number.
var pclTask = regexp.MustCompile(`^([-+])- \[([ xX])\] ([0-9]+)\. `)

// checkPlanningCommit is checkPlanningCommitLocation, then: a task line the
// working tree's change-folder tasks.md adds over HEAD's, already ticked,
// whose number HEAD's file did not carry, is
// PLANNING-COMMIT-TICKED-APPEND at exit 1. A task is appended unticked and
// committed before any work on it runs (document-fix.md's fix-run planning
// commit), and ticked once its commit passes check-task-commit-fields.sh —
// kan-1010's fix run appended Task 3 already ticked and skipped implement,
// panel and handoff. A tick of a task HEAD already carries removes that
// task's line in the same diff, so it passes; a tasks.md HEAD lacks is a new
// plan, never an append, and is not read.
func checkPlanningCommit(args []string, env Env, stdout, stderr io.Writer) int {
	rc := checkPlanningCommitLocation(args, env, stdout, stderr)
	if rc == 2 {
		return 2
	}
	wt, name := args[0], args[1]
	diff, ok := capture(envGit(env)("-C", wt, "diff", "HEAD", "--no-renames", "--diff-filter=M", "-U0", "--",
		":(top,glob)**/"+name+"/tasks.md"))
	if !ok {
		fmt.Fprintf(stderr, "check-planning-commit-location.sh: cannot diff %s's tasks.md against HEAD in %s\n", name, wt)
		return 2
	}
	for _, file := range strings.Split(diff, "diff --git ")[1:] {
		path := ""
		had := map[string]bool{}
		var ticked, added []string
		for _, l := range strings.Split(file, "\n") {
			if p, ok := strings.CutPrefix(l, "+++ b/"); ok {
				path = p
			}
			m := pclTask.FindStringSubmatch(l)
			switch {
			case m == nil:
			case m[1] == "-":
				had[m[3]] = true
			default:
				added = append(added, m[3])
				if m[2] != " " {
					ticked = append(ticked, m[3])
				}
			}
		}
		for _, n := range ticked {
			if !had[n] {
				fmt.Fprintf(stdout, "PLANNING-COMMIT-TICKED-APPEND: %s task %s — appended already ticked; append it unticked, commit, and tick it once its commit passes check-task-commit-fields.sh\n", path, n)
				rc = 1
			}
		}
		if slices.ContainsFunc(added, func(n string) bool { return !had[n] }) {
			plan := []string{filepath.Join(wt, path)}
			for _, guard := range []func([]string, Env, io.Writer, io.Writer) int{checkPlanShape, checkTaskBuildGreen} {
				switch guard(plan, env, stdout, stderr) {
				case 2:
					return 2
				case 1:
					rc = 1
				}
			}
		}
	}
	return rc
}

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
