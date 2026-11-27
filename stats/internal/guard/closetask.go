package guard

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// flow-guard close-task — the task-close step of skills/flow/implement.md: tick
// every task a returning implementer committed. It never pushes: the branch is
// pushed once at flow.implement's end, and before any stop inside it (Branch
// backup, skills/flow-contracts/git-boundaries.md).
//
// Usage: flow-guard close-task <canonical-worktree> <name> <merge-base> <task-id>[:<sha|worktree=sha[,worktree=sha…]>]…
//
// A bare <task-id> resolves to the newest commit since <merge-base> in the
// canonical worktree whose Task-Id: trailer names it, none a usage error. A
// bare <sha> is the canonical worktree's commit; the map form names one
// commit per worktree of a cross-repository task.
//
// In this order:
//  1. Every task's commit — every map pair's — must exist in its worktree
//     and sit on that worktree's HEAD, the branch the stage-end push sends; a
//     commit that does not is refused (exit 1) before anything is ticked, so
//     a task never ticks over work that would not leave the machine.
//     A task whose **Files:** touch the project's `ui paths` is refused the
//     same way (exit 1) unless check-ui-task-measured passes it: its
//     task-<id>-measure.json exists, is no older than its last source commit
//     and carries no flagged row. The refusal prints the command to run.
//  2. `flow tasks tick -C <canonical> <name> <id>` for every task.
//
// A failed tick is exit 2; every tick is still attempted.
//
// Exit 0 closed; 1 a commit is missing or off the branch — commit it, then
// re-run — or a UI task is unmeasured, stale or flagged; 2 a tick failed, or a usage error.
//
// A call that times out is inspected before it is retried: `git status
// --porcelain=v2 --branch` and `git stash list` in the worktree first. A
// reverting, rebasing or merging state on the `# branch` lines, a change the
// run did not make, or a stash entry the run did not push means the tree is
// not the one the run left — the run stops with both outputs verbatim,
// never re-running the guard on top of it.
func init() { Registry["close-task"] = closeTask }

func closeTask(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 4 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard close-task <canonical-worktree> <name> <merge-base> <task-id>[:<sha|worktree=sha[,worktree=sha…]>]…")
		return 2
	}
	canonical, name, mergeBase := args[0], args[1], args[2]
	git := envGit(env)
	var ids []string
	refused, unmeasured := false, false
	for _, a := range args[3:] {
		id, commit, given := strings.Cut(a, ":")
		if id == "" || given && commit == "" {
			Reason(env, "bad-task-argument")
			fmt.Fprintln(stderr, "close-task: task argument is not <task-id>[:<sha|map>]: "+a)
			return 2
		}
		if !given {
			if commit = ctTrailerCommit(git, canonical, mergeBase, id); commit == "" {
				Reason(env, "no-task-commit")
				fmt.Fprintf(stderr, "close-task: task %s: no commit since %s carries Task-Id: %s\n", id, mergeBase, id)
				return 2
			}
		}
		ids = append(ids, id)
		pairs := []string{canonical + "=" + commit}
		if strings.Contains(commit, "=") {
			pairs = strings.Split(commit, ",")
		}
		measured := []string{pcAbs(env, canonical)} // the canonical worktree, then the task's others
		for _, p := range pairs {
			wt, sha, _ := strings.Cut(strings.TrimSpace(p), "=")
			if abs := pcAbs(env, wt); !slices.Contains(measured, abs) {
				measured = append(measured, abs)
			}
			if _, ok := capture(git("-C", wt, "merge-base", "--is-ancestor", sha, "HEAD")); !ok {
				fmt.Fprintf(stderr, "close-task: task %s: %s is not a commit on %s's HEAD — commit the task there, then re-run\n", id, sha, wt)
				refused = true
			}
		}
		if why, msg := utmJudge(env, measured, name, id); why == "measure-missing" || why == "measure-stale" || why == "measure-not-clean" {
			fmt.Fprintf(stderr, "UI-TASK-UNMEASURED: %s — %s\n", id, msg)
			unmeasured = true
		}
	}
	if unmeasured {
		Reason(env, "ui-task-not-measured")
		fmt.Fprintln(stderr, "close-task: MEASURE — nothing was ticked")
		return 1
	}
	if refused {
		Reason(env, "commit-not-on-branch")
		fmt.Fprintln(stderr, "close-task: RE-COMMIT — nothing was ticked")
		return 1
	}

	rc := 0
	for _, id := range ids {
		out, errb, code := pcFlow(env, false, "tasks", "tick", "-C", canonical, name, id)
		_, _ = stdout.Write(out)
		_, _ = stderr.Write(errb)
		if code != 0 {
			fmt.Fprintf(stderr, "close-task: STOP — flow tasks tick failed for task %s\n", id)
			Reason(env, "tick-failed")
			rc = 2
		}
	}
	return rc
}

// ctTaskNumberRE finds the task numbers a Task-Id: trailer names.
var ctTaskNumberRE = regexp.MustCompile(`[0-9]+`)

// ctTrailerCommit is the newest commit in wt's base..HEAD whose Task-Id:
// trailer names id, "" when none does or git cannot read the range.
func ctTrailerCommit(git func(...string) *exec.Cmd, wt, base, id string) string {
	out, ok := capture(git("-C", wt, "log", "--no-merges", "--format=%H %(trailers:key=Task-Id,valueonly,separator=%x2C)", base+"..HEAD"))
	if !ok {
		return ""
	}
	for _, line := range strings.Split(out, "\n") {
		sha, ids, _ := strings.Cut(strings.TrimSpace(line), " ")
		if sha != "" && slices.Contains(ctTaskNumberRE.FindAllString(ids, -1), id) {
			return sha
		}
	}
	return ""
}
