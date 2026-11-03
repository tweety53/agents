package guard

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-review-gate.sh shim (KAN-1058).
//
// flow-guard check-review-gate — judge the per-task review gate of skills/flow/implement.md
// (**The review gate.**) for one task commit flow-guard check-task-commit-fields passed.
//
// Usage: flow-guard check-review-gate <worktree> <task-id> <commit-sha|worktree=sha[,worktree=sha…]>
//          <canonical-worktree> <change-name> [refused-path…]
//
// The first five arguments are flow-guard check-task-commit-fields's, less its empty
// parent placeholder, and the plan resolves the same way (one shared helper,
// so the two cannot drift). The commit map sums changed lines and unions paths
// across its repositories. Each trailing refused path is one the fields guard
// named undeclared before a Files: widening was transcribed: it stays
// undeclared, so the widening cannot disarm the gate.
//
// The gate fires when the commit changes more than 40 lines (inserted plus
// deleted, `git diff --no-renames --numstat`, a binary file's `-` counted as
// 0) or touches any path outside the task's declared set — its Files: plus
// its Allowed-collateral: globs, widened by a squash fold. The 40 lives in
// stats/internal/guard/reviewgate.go, the one site to re-tune it.
//
// Prints ONE verdict line to stdout:
//   FIRE: task <id> — <n> changed lines (more than 40)[; undeclared paths: <p>, …]
//   FIRE: task <id> — undeclared paths: <p>, …
//   QUIET: task <id> — <n> changed lines, every path declared
//
// Exit 0 a verdict was reached; exit 2 it could not judge (an unreadable
// plan, a task it cannot find, a commit git cannot resolve, a usage error),
// with "check-review-gate: COULD NOT JUDGE —" on stderr and nothing on stdout.

// checkReviewGate is `flow-guard check-review-gate`: the comment above is the contract. It judges the per-task review gate of skills/flow/implement.md
// (**The review gate.**) for one task commit the fields guard passed.
func init() { Registry["check-review-gate"] = checkReviewGate }

// rgMaxLines is the gate's size arm: a commit changing more lines than this
// fires it. Forty changed lines is the boundary below which a diff still is
// one glance. This constant is the one site to re-tune it.
const rgMaxLines = 40

func checkReviewGate(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(detail string) int {
		fmt.Fprintln(stderr, "check-review-gate: COULD NOT JUDGE — "+detail)
		return 2
	}
	if len(args) < 5 {
		return refuse(fmt.Sprintf("usage: flow-guard check-review-gate <worktree> <task-id> <commit-sha|worktree=sha[,worktree=sha…]> <canonical-worktree> <change-name> [refused-path…] — got %d argument(s)", len(args)))
	}
	worktree, taskID, commit, canonical, name, refused := args[0], args[1], args[2], args[3], args[4], args[5:]
	if !isDir(worktree) {
		return refuse("worktree not found: " + worktree)
	}

	repos := []tcfRepoCommit{{worktree, commit}}
	if strings.Contains(commit, "=") {
		repos = nil
		for _, entry := range strings.Split(commit, ",") {
			wt, sha, _ := strings.Cut(strings.TrimSpace(entry), "=")
			if wt == "" || sha == "" || !isDir(wt) {
				return refuse(fmt.Sprintf("map entry is not <worktree>=<sha> naming an existing worktree: %q", entry))
			}
			repos = append(repos, tcfRepoCommit{wt, sha})
		}
	}

	tasksMD, detail := tcfResolveTasksMD(env, stderr, "check-review-gate", worktree, name, canonical)
	if tasksMD == "" {
		if detail == "" {
			return 2
		}
		return refuse(detail)
	}
	raw, err := os.ReadFile(tasksMD)
	if err != nil {
		return refuse(err.Error())
	}
	lines, err := tcfPlanLines(raw)
	if err != nil {
		return refuse(err.Error())
	}
	task, found := tcfParseTask(lines, taskID)
	if !found {
		return refuse(fmt.Sprintf("task %s not found in %s", taskID, tasksMD))
	}
	if task.unclosedFenceLine != 0 {
		return refuse(fmt.Sprintf("task %s's body opens a code fence at tasks.md line %d it never closes, so its fields cannot be read", taskID, task.unclosedFenceLine))
	}
	// The declared set is the fields guard's own: Files: plus the
	// Allowed-collateral: globs, widened by a squash fold.
	folded, violations := tcfResolveFolded(lines, task)
	if violations != nil {
		return refuse("the task's fold cannot be resolved: " + strings.Join(violations, "; "))
	}
	declared := func(path string) bool {
		return slices.Contains(folded.files, path) ||
			slices.ContainsFunc(folded.allowedCollateral, func(g string) bool { return tcfFnmatch(path, g) })
	}

	// --no-renames pins both facts against the caller's git config, as the
	// fields guard pins its changed-file list: a rename is its source deleted
	// and its destination added, both paths judged.
	total := 0
	var undeclared []string
	for _, r := range repos {
		out, err := tcfRunGit(env, r.worktree, "diff", "--no-renames", "--numstat", r.commit+"^.."+r.commit)
		if err != nil {
			return refuse(err.Error())
		}
		for _, line := range tcfSplitLines(out) {
			if line == "" {
				continue
			}
			f := strings.SplitN(line, "\t", 3)
			if len(f) != 3 {
				return refuse("unreadable numstat line: " + line)
			}
			// A binary file's counts are "-": it counts as 0 lines.
			for _, n := range f[:2] {
				if n != "-" {
					v, err := strconv.Atoi(n)
					if err != nil {
						return refuse("unreadable numstat line: " + line)
					}
					total += v
				}
			}
			// A path the fields guard refused stays undeclared, so a Files:
			// widening transcribed from that refusal cannot disarm the gate.
			if (slices.Contains(refused, f[2]) || !declared(f[2])) && !slices.Contains(undeclared, f[2]) {
				undeclared = append(undeclared, f[2])
			}
		}
	}

	var arms []string
	if total > rgMaxLines {
		arms = append(arms, fmt.Sprintf("%d changed lines (more than %d)", total, rgMaxLines))
	}
	if len(undeclared) > 0 {
		arms = append(arms, "undeclared paths: "+strings.Join(undeclared, ", "))
	}
	if arms != nil {
		fmt.Fprintf(stdout, "FIRE: task %s — %s\n", taskID, strings.Join(arms, "; "))
	} else {
		fmt.Fprintf(stdout, "QUIET: task %s — %d changed lines, every path declared\n", taskID, total)
	}
	return 0
}
