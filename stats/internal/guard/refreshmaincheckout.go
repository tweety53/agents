package guard

import (
	"fmt"
	"io"
	"strings"
)

// refreshMainCheckout is scripts/refresh-main-checkout.sh: that script's
// header comment is the contract -- 0 REFRESH-DONE or REFRESH-CURRENT, 1
// REFRESH-REFUSED with nothing touched, 2 usage, every verdict on stdout.
// A git read that fails outright -- an unmerged index `write-tree` cannot
// write, an unborn HEAD -- is exit 2 with git's own message on stderr, as
// cannot-answer; the bash died there under `set -e` with git's status
// (128), a code its contract did not name.
func init() {
	Registry["refresh-main-checkout"] = refreshMainCheckout
}

func refreshMainCheckout(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 || !isDir(pcAbs(env, args[0])) {
		fmt.Fprintln(stderr, "usage: refresh-main-checkout.sh <main-checkout> <base>")
		return 2
	}
	repo, base := args[0], args[1]
	git := envGit(env)
	// g is the bash's g(): git -C <repo>, its stderr reaching the caller.
	g := func(a ...string) (string, bool) {
		cmd := git(append([]string{"-C", repo}, a...)...)
		cmd.Stderr = stderr
		return capture(cmd)
	}
	refuse := func(reason string) int {
		fmt.Fprintf(stdout, "REFRESH-REFUSED: %s %s — nothing touched\n", repo, reason)
		return 1
	}

	cur, _ := capture(git("-C", repo, "symbolic-ref", "-q", "--short", "HEAD"))
	if cur == "" {
		return refuse("is detached")
	}
	if cur != base {
		return refuse("is on " + cur + ", not " + base)
	}
	// Staleness never produces an unstaged change -- the worktree and the
	// index go stale together -- so any is real work.
	if _, ok := g("diff", "--quiet"); !ok {
		return refuse("has unstaged changes")
	}

	indexTree, ok := g("write-tree")
	if !ok {
		return 2
	}
	tip, ok := g("rev-parse", "--short", "HEAD")
	if !ok {
		return 2
	}
	if headTree, _ := g("rev-parse", "HEAD^{tree}"); indexTree == headTree {
		fmt.Fprintf(stdout, "REFRESH-CURRENT: %s already at %s\n", repo, tip)
		return 0
	}

	// An index tree equal to the tree of a commit the branch has already
	// moved past is the reversal artifact, not staged work. One rev-list
	// prints each of the last 300 commits with its tree, where the bash ran
	// a rev-parse per commit: `commit <sha>` then `<tree>`.
	list, _ := g("rev-list", "--max-count=300", "--skip=1", "--format=%T", "HEAD")
	match := ""
	lines := strings.Split(list, "\n")
	for i := 0; i+1 < len(lines); i += 2 {
		if lines[i+1] == indexTree {
			match = strings.TrimPrefix(lines[i], "commit ")
			break
		}
	}
	if match == "" {
		return refuse("has staged changes that match no recent " + base + " tip")
	}

	// The reset is safe: the index tree is byte-identical to a commit the
	// branch has already moved past, and the worktree is identical to the
	// index, so there is no edit anywhere the reset could lose. `reset
	// --hard` never touches untracked files.
	if _, ok := g("reset", "-q", "--hard"); !ok {
		return 2
	}
	short, _ := g("rev-parse", "--short", match)
	fmt.Fprintf(stdout, "REFRESH-DONE: %s index was the tree of %s; now at %s\n", repo, short, tip)
	return 0
}
