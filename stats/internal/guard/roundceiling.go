package guard

// flow-guard round-ceiling — the round ceiling of every fix loop
// (**The round ceiling**, skills/flow-contracts/pipeline.md): what a loop's
// next round is, given how many of its rounds have failed so far.
//
// Usage: flow-guard round-ceiling <worktree> <failed rounds>
//
// <failed rounds> is the count the loop's site already keeps — a round as the
// **One round** column of its row of **The fix loop** (skills/flow/fix.md)
// defines it — so no counter is stored. The ceiling is the first non-blank
// line of <worktree>/.flow/project.md's `## fix round ceiling`, a positive
// integer optionally followed by prose; absent, it is 6: three rounds on the
// loop's own fixer, then three rethinks, before the item is deferred.
//
// Stdout, one line:
//
//	NEXT: retry    fewer than three failed rounds — the loop's own fixer again
//	NEXT: rethink  three or more, below the ceiling — **Three failed rounds**
//	NEXT: defer    the ceiling reached — the item is deferred, the run continues
//
// Exit codes:
//
//	0  a verdict printed.
//	2  usage — not two arguments, or <failed rounds> not a positive integer;
//	   or a `## fix round ceiling` whose head is not a positive integer.

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

func init() {
	Registry["round-ceiling"] = roundCeiling
}

const defaultRoundCeiling = 6

func roundCeiling(args []string, env Env, stdout, stderr io.Writer) int {
	failed := 0
	if len(args) == 2 {
		failed, _ = strconv.Atoi(args[1])
	}
	if failed < 1 {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard round-ceiling <worktree> <failed rounds, a positive integer>")
		return 2
	}
	ceiling := defaultRoundCeiling
	body := projectSection(filepath.Join(args[0], ".flow", "project.md"), "fix round ceiling")
	if head := strings.Fields(body); len(head) > 0 {
		n, err := strconv.Atoi(head[0])
		if err != nil || n < 1 {
			Reason(env, "malformed-ceiling")
			fmt.Fprintf(stderr, "round-ceiling: `## fix round ceiling` opens with %q, not a positive integer\n", head[0])
			return 2
		}
		ceiling = n
	}
	switch {
	case failed >= ceiling:
		fmt.Fprintf(stdout, "NEXT: defer — %d failed rounds reach the ceiling of %d\n", failed, ceiling)
	case failed >= 3:
		fmt.Fprintf(stdout, "NEXT: rethink — %d failed rounds, ceiling %d\n", failed, ceiling)
	default:
		fmt.Fprintf(stdout, "NEXT: retry — %d failed rounds, ceiling %d\n", failed, ceiling)
	}
	return 0
}
