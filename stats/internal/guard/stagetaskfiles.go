package guard

import (
	"fmt"
	"io"
	"strings"
)

// stageTaskFiles is scripts/stage-task-files.sh: that script's header is the
// contract -- the staging half of a /flow implementer's per-task commit
// (skills/flow/implement.md, FLOW — COMMIT-PER-TASK), refusing to hand the
// commit a task path git left unstaged.
func init() { Registry["stage-task-files"] = stageTaskFiles }

// stfPlanning is what the clearing reset unstages and the add excludes: both
// spec-tree leaf spellings and the superpowers notes.
var stfPlanning = []string{"spectre/changes/", "openspec/changes/", "docs/superpowers/"}

func stageTaskFiles(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprint(stderr, "usage: stage-task-files.sh <abs-worktree> <file>...\n")
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
	add := []string{"add", "--"}
	for _, p := range stfPlanning {
		add = append(add, ":(exclude)"+p)
	}
	// The excludes lead the paths: git 2.50.1 given the task's paths first
	// and an exclude after them stages tracked paths alone and exits 0
	// (kan-1012 self-review).
	for _, step := range [][]string{append([]string{"reset", "-q", "--"}, stfPlanning...), append(add, files...)} {
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
