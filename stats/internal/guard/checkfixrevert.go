package guard

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

// flow-guard check-fix-revert — after a fix commit, refuse one that undoes an earlier fix of
// the same change (KAN-1130: each fixer saw only the current finding and reverted or reinvented
// an earlier round's fix, so the loop ping-ponged).
//
// Usage: flow-guard check-fix-revert <worktree> <base>
//
// The fix commit is <worktree>'s HEAD. The earlier fixes are the commits of `<base>..HEAD^`
// carrying a `Flow-Key:` trailer that names a fix (fixKey) (`skills/flow/fix.md` step 2). A line HEAD removes
// that an earlier fix added, or adds that an earlier fix removed — compared per file, on
// whitespace-trimmed content, blank lines ignored — is an inverted line.
//
// Prints, per earlier fix with more than maxInvertedFixLines inverted lines:
//   FIX-REVERT: <HEAD sha> inverts <n> lines of <earlier sha> (<its Flow-Key>)
// Exit 0 none, 1 at least one — the pair goes to the fixer as a CONFLICT to reconcile —
// 2 cannot answer (usage, not a worktree, a failing git call), the cause on stderr.

// maxInvertedFixLines is the number of an earlier fix's lines a fix may invert before the
// guard calls it a revert.
const maxInvertedFixLines = 3

func init() { Registry["check-fix-revert"] = checkFixRevert }

// fixKey is a fix commit's Flow-Key: any key with a `fix` segment — `fix-<k>-implementer`,
// `panel-fix-<round>`, `full-suite-fix-<n>`, `task-<n>-implementer-fix-<k>`, `minor-fix`.
var fixKey = regexp.MustCompile(`(^|-)fix(-|$)`)

func checkFixRevert(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-fix-revert"
	if len(args) < 1 || len(args) > 2 {
		fmt.Fprintf(stderr, "%s: usage: flow-guard check-fix-revert <worktree> [<base>]\n", prog)
		Reason(env, "usage")
		return 2
	}
	wt := args[0]
	git := envGit(env)
	base := ""
	if len(args) == 2 {
		base = args[1]
	} else if br, ok := capture(git("-C", wt, "symbolic-ref", "--short", "HEAD")); ok {
		// The branch's flow base, as taskrecords.go resolves it: `branch.<current>.flowBase`, else origin/HEAD.
		base, _ = capture(git("-C", wt, "config", "--get", "branch."+br+".flowBase"))
	}
	if base == "" {
		base = "refs/remotes/origin/HEAD"
	}
	head, ok := capture(git("-C", wt, "rev-parse", "HEAD"))
	if !ok {
		fmt.Fprintf(stderr, "%s: not a git worktree: %s\n", prog, wt)
		Reason(env, "not-a-worktree")
		return 2
	}
	log, ok := capture(git("-C", wt, "log", "--format=%H %(trailers:key=Flow-Key,valueonly,separator=%x2C)", base+"..HEAD^"))
	if !ok {
		fmt.Fprintf(stderr, "%s: git log %s..HEAD^ failed in %s\n", prog, base, wt)
		Reason(env, "git-failed")
		return 2
	}
	headAdd, headDel, ok := commitLines(git, wt, head)
	if !ok {
		fmt.Fprintf(stderr, "%s: git show %s failed\n", prog, head)
		Reason(env, "git-failed")
		return 2
	}
	rc := 0
	for _, row := range strings.Split(log, "\n") {
		sha, key, _ := strings.Cut(row, " ")
		if !fixKey.MatchString(key) {
			continue
		}
		add, del, ok := commitLines(git, wt, sha)
		if !ok {
			fmt.Fprintf(stderr, "%s: git show %s failed\n", prog, sha)
			Reason(env, "git-failed")
			return 2
		}
		n := overlap(headDel, add) + overlap(headAdd, del)
		if n > maxInvertedFixLines {
			fmt.Fprintf(stdout, "FIX-REVERT: %s inverts %d lines of %s (%s)\n", head, n, sha, key)
			rc = 1
		}
	}
	if rc != 0 {
		Reason(env, "fix-conflict")
	}
	return rc
}

// commitLines is a commit's added and removed lines, each keyed `<file>\x00<trimmed line>`
// and counted, blank lines left out.
func commitLines(git func(...string) *exec.Cmd, wt, sha string) (add, del map[string]int, ok bool) {
	out, ok := capture(git("-C", wt, "show", "--format=", "-U0", "--no-color", sha))
	if !ok {
		return nil, nil, false
	}
	add, del = map[string]int{}, map[string]int{}
	file := ""
	for _, l := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(l, "+++ "), strings.HasPrefix(l, "--- "):
			if p := strings.TrimPrefix(l, "+++ "); p != l && p != "/dev/null" {
				file = strings.TrimPrefix(p, "b/")
			} else if p := strings.TrimPrefix(l, "--- "); p != l && p != "/dev/null" {
				file = strings.TrimPrefix(p, "a/")
			}
		case strings.HasPrefix(l, "+"), strings.HasPrefix(l, "-"):
			if t := strings.TrimSpace(l[1:]); t != "" {
				m := del
				if l[0] == '+' {
					m = add
				}
				m[file+"\x00"+t]++
			}
		}
	}
	return add, del, true
}

// overlap counts the lines a and b share, each line as many times as both hold it.
func overlap(a, b map[string]int) int {
	n := 0
	for k, c := range a {
		n += min(c, b[k])
	}
	return n
}
