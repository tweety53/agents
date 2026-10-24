package guard

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// checkBaseBehind is scripts/check-base-behind.sh: that script's header
// comment is the contract. Each worktree's recorded base is compared with the
// branch origin/HEAD names (kan-1007).
func init() { Registry["check-base-behind"] = checkBaseBehind }

func checkBaseBehind(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-base-behind"
	if len(args) == 0 {
		fmt.Fprintf(stderr, "%s: usage: check-base-behind.sh <worktree>...\n", prog)
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
