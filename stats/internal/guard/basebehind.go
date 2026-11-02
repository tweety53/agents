package guard

import (
	"fmt"
	"io"
	"strconv"
	"strings"
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
// recorded, the guard runs `git fetch -q origin`; a failing fetch is ignored and
// the refs already on hand are compared.
//
// Prints to stdout, one block per worktree in argument order:
//   BASE-CURRENT: <wt> — no recorded base
//   BASE-CURRENT: <wt> — <base> is the tracked branch
//   BASE-CURRENT: <wt> — origin/<base> holds every commit of origin/<tracked>
//   BASE-BEHIND: <wt> — origin/<base> lacks <n> commits of origin/<tracked>; sync the base, never merge <tracked> into the change branch (kan-1007):
//     T=$(git -C <wt> merge-tree --write-tree origin/<base> origin/<tracked>) && C=$(git -C <wt> commit-tree "$T" -p origin/<base> -p origin/<tracked> -m 'Merge <tracked> into <base>') && git -C <wt> push origin "${C}:refs/heads/<base>"
//     git -C <wt> fetch -q origin && flow-guard sync-onto-base <wt> origin/<base> <merge-base>
// where <merge-base> is the sha `git merge-base HEAD origin/<base>` prints. The
// first command line syncs the base with no checkout (merge-tree exits 1 on a
// conflict and stops the chain); its push runs unasked (operator-prompts-auto-resolution.md).
//
// Exit 0 no BASE-BEHIND, 1 at least one, 2 cannot answer (no argument, a path
// that is not a git worktree, no origin/HEAD, a failing rev-list or
// merge-base) — stdout empty, the cause on stderr.
//
// Go, in flow-guard (stats/internal/guard/basebehind.go).

// checkBaseBehind is scripts/check-base-behind.sh: that script's header
// comment is the contract. Each worktree's recorded base is compared with the
// branch origin/HEAD names (kan-1007).
func init() { Registry["check-base-behind"] = checkBaseBehind }

func checkBaseBehind(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-base-behind"
	if len(args) == 0 {
		fmt.Fprintf(stderr, "%s: usage: flow-guard check-base-behind <worktree>...\n", prog)
		return 2
	}
	git := envGit(env)
	var out strings.Builder
	rc := 0
	for _, wt := range args {
		if _, ok := capture(git("-C", wt, "rev-parse", "--is-inside-work-tree")); !ok {
			fmt.Fprintf(stderr, "%s: not a git worktree: %s\n", prog, wt)
			return 2
		}
		branch, _ := capture(git("-C", wt, "symbolic-ref", "--short", "-q", "HEAD"))
		base := ""
		if branch != "" {
			base = recordedBase(git, wt, branch)
		}
		if base == "" {
			fmt.Fprintf(&out, "BASE-CURRENT: %s — no recorded base\n", wt)
			continue
		}
		_ = git("-C", wt, "fetch", "-q", "origin").Run() // a failed fetch compares the refs on hand
		head, ok := capture(git("-C", wt, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"))
		if !ok {
			fmt.Fprintf(stderr, "%s: %s has no origin/HEAD\n", prog, wt)
			return 2
		}
		tracked := strings.TrimPrefix(head, "origin/")
		if base == tracked {
			fmt.Fprintf(&out, "BASE-CURRENT: %s — %s is the tracked branch\n", wt, base)
			continue
		}
		count, ok := capture(git("-C", wt, "rev-list", "--count", "origin/"+base+"..origin/"+tracked))
		n, err := strconv.Atoi(count)
		if !ok || err != nil {
			fmt.Fprintf(stderr, "%s: %s: git rev-list --count origin/%s..origin/%s failed\n", prog, wt, base, tracked)
			return 2
		}
		if n == 0 {
			fmt.Fprintf(&out, "BASE-CURRENT: %s — origin/%s holds every commit of origin/%s\n", wt, base, tracked)
			continue
		}
		mb, ok := capture(git("-C", wt, "merge-base", "HEAD", "origin/"+base))
		if !ok {
			fmt.Fprintf(stderr, "%s: %s: git merge-base HEAD origin/%s failed\n", prog, wt, base)
			return 2
		}
		rc = 1
		fmt.Fprintf(&out, "BASE-BEHIND: %s — origin/%s lacks %d commits of origin/%s; sync the base, never merge %s into the change branch (kan-1007):\n", wt, base, n, tracked, tracked)
		fmt.Fprintf(&out, "  T=$(git -C %s merge-tree --write-tree origin/%s origin/%s) && C=$(git -C %s commit-tree \"$T\" -p origin/%s -p origin/%s -m 'Merge %s into %s') && git -C %s push origin \"${C}:refs/heads/%s\"\n",
			wt, base, tracked, wt, base, tracked, tracked, base, wt, base)
		fmt.Fprintf(&out, "  git -C %s fetch -q origin && sync-onto-base.sh %s origin/%s %s\n", wt, wt, base, mb)
	}
	io.WriteString(stdout, out.String())
	return rc
}
