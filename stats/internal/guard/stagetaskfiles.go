package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// stageTaskFiles is scripts/stage-task-files.sh: that script's header is the
// contract -- the staging half of a /flow implementer's per-task commit
// (skills/flow/implementer-brief.md, FLOW — COMMIT-PER-TASK), refusing to hand the
// commit a task path git left unstaged.
func init() { Registry["stage-task-files"] = stageTaskFiles }

// stfPlanning is what the clearing resets unstage and the add never names: both
// spec-tree leaf spellings and the superpowers notes.
var stfPlanning = []string{"spectre/changes/", "openspec/changes/", "docs/superpowers/"}

func stageTaskFiles(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprint(stderr, "usage: flow-guard stage-task-files <abs-worktree> <file>...\n")
		return 2
	}
	wt, files := args[0], args[1:]
	if !isDir(smcAbs(env, wt)) {
		fmt.Fprintf(stderr, "stage-task-files: %s is not a directory\n", wt)
		return 2
	}
	gitBin, ok := panelResolveGit(env, "stage-task-files", stderr)
	if !ok {
		return 2
	}
	// The add names the task's own non-planning paths and no exclude
	// pathspec: with spectre/changes/ gitignored and files still tracked
	// under it, git 2.50.1 fails an add carrying ':(exclude)spectre/changes/'
	// on the ignored directory (KAN-1073). The second reset unstages any
	// planning path the add swept in through a named parent directory.
	add := []string{"add", "--"}
	// A path whose deletion is already staged is neither on disk nor in the
	// index, and git add refuses it as an unmatched pathspec; it needs no add.
	for _, f := range files {
		if stfIsPlanning(f) {
			continue
		}
		if _, err := os.Lstat(filepath.Join(smcAbs(env, wt), f)); err == nil {
			add = append(add, f)
		} else if out, _ := panelGit(env, gitBin, wt, "ls-files", "--", f).Output(); len(out) > 0 {
			add = append(add, f)
		}
	}
	unstage := append([]string{"reset", "-q", "--"}, stfPlanning...)
	steps := [][]string{unstage}
	if len(add) > 2 {
		steps = append(steps, add)
	}
	for _, step := range append(steps, unstage) {
		if out, err := panelGit(env, gitBin, wt, step...).CombinedOutput(); err != nil {
			fmt.Fprintf(stderr, "stage-task-files: git %s failed in %s:\n%s", step[0], wt, out)
			return 2
		}
	}
	out, err := panelGit(env, gitBin, wt, append([]string{"status", "--porcelain", "-z", "--untracked-files=all", "--"}, files...)...).Output()
	if err != nil {
		fmt.Fprintf(stderr, "stage-task-files: git status failed in %s\n", wt)
		return 2
	}
	var left []string
	// -z: porcelain quotes a spaced or non-ASCII path, which no planning
	// prefix would match; NUL-separated entries carry it raw. A rename or
	// copy entry is followed by its source path, skipped here.
	entries := strings.Split(strings.TrimRight(string(out), "\x00"), "\x00")
	for i := 0; i < len(entries); i++ {
		l := entries[i]
		if len(l) > 0 && (l[0] == 'R' || l[0] == 'C') {
			i++
		}
		if len(l) < 4 || l[1] == ' ' || stfIsPlanning(l[3:]) {
			continue
		}
		left = append(left, l)
	}
	if len(left) > 0 {
		fmt.Fprintf(stderr, "stage-task-files: git add exited 0 and left %d task path(s) unstaged — commit nothing, report this:\n%s\n", len(left), strings.Join(left, "\n"))
		return 1
	}
	return 0
}

func stfIsPlanning(path string) bool {
	for _, p := range stfPlanning {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
