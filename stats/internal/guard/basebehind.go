package guard

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The contract below moved here from the header of the deleted
// scripts/check-base-behind.sh shim (KAN-1058).
//
// flow-guard check-base-behind — compare each worktree's recorded base with the branch
// origin tracks, before a stale base turns into failures (kan-1007: a feature
// base behind develop was papered over by merging develop into the change
// branches, and integrate's reshape folded KAN-1010 into the squash).
//
// Usage: flow-guard check-base-behind <worktree>...
//
// Per worktree: the recorded base is `branch.<current branch>.flowBase`; the
// tracked branch is the branch `refs/remotes/origin/HEAD` names. Once a base is
// recorded, the guard runs `git fetch -q origin`, once per repository, unless a fetch of it
// succeeded within the last 60s (resolve-base-branch's WHY THE FETCH IS SKIPPED); a failing fetch is
// ignored and the refs already on hand are compared. sync-base-behind, which pushes the base, always
// fetches.
//
// Prints to stdout, one block per worktree in argument order:
//   BASE-CURRENT: <wt> — no recorded base
//   BASE-CURRENT: <wt> — <base> is the tracked branch
//   BASE-CURRENT: <wt> — origin/<base> holds every commit of origin/<tracked>
//   BASE-BEHIND: <wt> — origin/<base> lacks <n> commits of origin/<tracked>; sync the base, never merge <tracked> into the change branch (kan-1007):
//     flow-guard sync-base-behind <wt>
// which does the whole sync itself — merge-tree, commit-tree, the push of the base, sync-onto-base —
// so no raw `git push` is ever typed (kan-1156; syncbasebehind.go).
//
// Exit 0 no BASE-BEHIND, 1 at least one, 2 cannot answer (no argument, a path
// that is not a git worktree, no origin/HEAD, a failing rev-list or
// merge-base) — stdout empty, the cause on stderr.
//
// Go, in flow-guard (stats/internal/guard/basebehind.go).

// checkBaseBehind is `flow-guard check-base-behind`: the comment above is the contract. Each worktree's recorded base is compared with the
// branch origin/HEAD names (kan-1007).
func init() { Registry["check-base-behind"] = checkBaseBehind }

func checkBaseBehind(args []string, env core.Env, stdout, stderr io.Writer) int {
	return baseBehind(args, env, stdout, stderr, false)
}

// baseBehind is check-base-behind; force fetches even after a recent fetch.
func baseBehind(args []string, env core.Env, stdout, stderr io.Writer, force bool) int {
	const prog = "check-base-behind"
	if len(args) == 0 {
		fmt.Fprintf(stderr, "%s: usage: flow-guard check-base-behind <worktree>...\n", prog)
		core.Reason(env, "usage")
		return 2
	}
	git := core.EnvGit(env)
	var out strings.Builder
	rc := 0
	fetched := map[string]bool{}
	for _, wt := range args {
		if _, ok := core.Capture(git("-C", wt, "rev-parse", "--is-inside-work-tree")); !ok {
			fmt.Fprintf(stderr, "%s: not a git worktree: %s\n", prog, wt)
			core.Reason(env, "not-a-worktree")
			return 2
		}
		branch, _ := core.Capture(git("-C", wt, "symbolic-ref", "--short", "-q", "HEAD"))
		base := ""
		if branch != "" {
			base = recordedBase(git, wt, branch)
		}
		if base == "" {
			fmt.Fprintf(&out, "BASE-CURRENT: %s — no recorded base\n", wt)
			continue
		}
		// Worktrees of one repository share their origin: fetch it once. A failed fetch compares the refs on hand.
		common, _ := core.Capture(git("-C", wt, "rev-parse", "--path-format=absolute", "--git-common-dir"))
		if !fetched[common] && (force || !fetchedRecently(git, wt)) {
			_ = git("-C", wt, "fetch", "-q", "origin").Run()
			fetched[common] = true
		}
		head, ok := core.Capture(git("-C", wt, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"))
		if !ok {
			fmt.Fprintf(stderr, "%s: %s has no origin/HEAD\n", prog, wt)
			core.Reason(env, "has-no-origin-head")
			return 2
		}
		tracked := strings.TrimPrefix(head, "origin/")
		if base == tracked {
			fmt.Fprintf(&out, "BASE-CURRENT: %s — %s is the tracked branch\n", wt, base)
			continue
		}
		count, ok := core.Capture(git("-C", wt, "rev-list", "--count", "origin/"+base+"..origin/"+tracked))
		n, err := strconv.Atoi(count)
		if !ok || err != nil {
			fmt.Fprintf(stderr, "%s: %s: git rev-list --count origin/%s..origin/%s failed\n", prog, wt, base, tracked)
			core.Reason(env, "rev-list-failed")
			return 2
		}
		if n == 0 {
			fmt.Fprintf(&out, "BASE-CURRENT: %s — origin/%s holds every commit of origin/%s\n", wt, base, tracked)
			continue
		}
		rc = 1
		core.Reason(env, "base-behind")
		fmt.Fprintf(&out, "BASE-BEHIND: %s — origin/%s lacks %d commits of origin/%s; sync the base, never merge %s into the change branch (kan-1007):\n", wt, base, n, tracked, tracked)
		fmt.Fprintf(&out, "  flow-guard sync-base-behind %s\n", wt)
	}
	io.WriteString(stdout, out.String())
	return rc
}
