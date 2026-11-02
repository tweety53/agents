package guard

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

// The contract below moved here from the header of the deleted
// scripts/check-worktree-location.sh shim (KAN-1058).
//
// flow-guard check-worktree-location — refuse any registered worktree that lives
// outside <parent>/<project>-worktrees/, the sibling of the main checkout every
// /flow worktree is created under (kan-916 design.md's sibling-worktrees-layout
// decision; the strict guard itself is KAN-462 §11). Outside the repository
// `git check-ignore` exits 128, so Claude Code's LSP filter drops nothing.
//
// Usage: flow-guard check-worktree-location <project>
//
// Prints one STRAY line per offender, then ONE verdict line:
//   STRAY: <path> (<branch>|detached)      per worktree outside <project>-worktrees/
//   FOREIGN: <path> (<branch>|detached)    per harness agent worktree, never counted
//   LOCATION-OK: <project>                 every worktree is at or under it
//   LOCATION-STRAY: <project> — <n>        <n> worktree(s) are not
//
// Exit 0 on LOCATION-OK, 1 on LOCATION-STRAY, 2 with NOTHING on stdout when
// <project> is not a readable directory or `git worktree list` fails — an
// inability is never reported as a verdict.
//
// THE FIRST `worktree` ENTRY IS ALWAYS THE MAIN CHECKOUT. `git worktree list
// --porcelain` prints it first, unconditionally, so this guard skips it by
// position rather than by comparing it against <project> — comparing paths
// would need the same physical-form resolution the strays already require,
// for a fact the porcelain format already guarantees.
//
// PATHS ARE COMPARED IN PHYSICAL FORM. The project root is resolved with
// `cd … && pwd -P` before the comparison, matching what `git worktree list`
// itself already reports — git resolves a worktree's path (through /tmp's
// macOS symlink to /private/tmp, for instance) at `add` time, so the entries
// read from porcelain need no separate resolution of their own.
//
// "AT OR UNDER" IS NOT "STARTS WITH". A path matches when it equals
// <project>-worktrees or begins with it plus a slash — a bare string-prefix
// test would report <project>-worktrees-old/x as in-tree. <project>/.worktrees/,
// the retired in-repo layout, is STRAY like any other path.
//
// A HARNESS AGENT'S WORKTREE IS NOT FLOW'S (KAN-883). Claude Code creates a
// subagent's isolated worktree under <project>/.claude/worktrees/ and locks
// it while that session lives; no flow command creates one there, and a run
// cannot remove another session's. Each is named on a FOREIGN line, never
// counted, so it neither stops a kickoff nor reaches check-finish-preflight's
// STRAY relay. Every other path outside <project>-worktrees/ is STRAY.
//
// ANOTHER SESSION'S SCRATCHPAD IS NOT THIS RUN'S EITHER (kan-1057 self-review,
// F7). A worktree under a Claude Code session scratchpad,
// …/claude-<uid>/<project slug>/<session id>/scratchpad/…, whose session id
// is not $CLAUDE_CODE_SESSION_ID is named on a FOREIGN line, never counted:
// this run cannot remove it. This session's own, or any with the variable
// unset, stays STRAY.

// checkWorktreeLocation is `flow-guard check-worktree-location`: the comment above is the contract -- one STRAY line per registered worktree
// outside <parent>/<project>-worktrees/ (design.md:
// sibling-worktrees-layout), then ONE verdict line, LOCATION-OK (exit 0)
// or LOCATION-STRAY (exit 1); exit 2 with NOTHING on stdout when <project>
// is not a readable directory or `git worktree list` fails -- an inability
// is never reported as a verdict.
func init() { Registry["check-worktree-location"] = checkWorktreeLocation }

// wlScratchpad matches a path under a Claude Code session scratchpad,
// …/claude-<uid>/<project slug>/<session id>/scratchpad/…, capturing the
// session id.
var wlScratchpad = regexp.MustCompile(`/claude-[0-9]+/[^/]+/([^/]+)/scratchpad(/|$)`)

func checkWorktreeLocation(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-worktree-location: "+format+"\n", a...)
		return 2
	}
	if len(args) != 1 {
		return die("usage: flow-guard check-worktree-location <project>")
	}

	// PATHS ARE COMPARED IN PHYSICAL FORM. The project root is resolved as
	// `cd … && pwd -P` resolves it, matching what `git worktree list` itself
	// already reports -- git resolves a worktree's path (through /tmp's
	// macOS symlink to /private/tmp, for instance) at `add` time, so the
	// entries read from porcelain need no separate resolution of their own.
	// `cd` also needs search permission, which a bare stat does not test,
	// and refuses an empty operand rather than staying put.
	root, err := filepath.EvalSymlinks(pcAbs(env, args[0]))
	if args[0] == "" || err != nil || !isDir(root) || syscall.Access(root, 1) != nil {
		return die("%s is not a directory", args[0])
	}
	porcelain, ok := capture(envGit(env)("-C", root, "worktree", "list", "--porcelain"))
	if !ok {
		return die("cannot list the worktrees of %s", root)
	}

	// THE FIRST `worktree` ENTRY IS ALWAYS THE MAIN CHECKOUT. `git worktree
	// list --porcelain` prints it first, unconditionally, so this guard skips
	// it by position (n == 1 at its record) rather than by comparing it
	// against <project> -- comparing paths would need the same physical-form
	// resolution the strays already require, for a fact the porcelain format
	// already guarantees.
	//
	// THE ACCEPTED ROOT IS THE MAIN CHECKOUT'S SIBLING, <project>-worktrees,
	// outside the repository: there `git check-ignore` exits 128, so Claude
	// Code's LSP result filter drops nothing. <project>/.worktrees/, the
	// retired in-repo layout, is STRAY like any other path.
	//
	// "AT OR UNDER" IS NOT "STARTS WITH". A path matches when it equals
	// <project>-worktrees or begins with it plus a slash -- a bare
	// string-prefix test would report <project>-worktrees-old/x as in-tree.
	//
	// The appended newline restores the blank line that terminates the LAST
	// record, which capture's trailing-newline strip removed (the bash's
	// `printf '%s\n\n'`), so the final worktree's record reaches the check.
	// A HARNESS AGENT'S WORKTREE IS NOT FLOW'S (KAN-883). Claude Code
	// creates a subagent's isolated worktree under <project>/.claude/worktrees/
	// and locks it while that session lives; no flow command creates one
	// there, and a run cannot remove another session's. Such a worktree is
	// named on a FOREIGN line and never counted, so it neither stops a
	// kickoff nor reaches check-finish-preflight's STRAY relay.
	//
	// ANOTHER SESSION'S SCRATCHPAD IS NOT THIS RUN'S EITHER (kan-1057
	// self-review, F7). A worktree under a session scratchpad whose session id
	// is not $CLAUDE_CODE_SESSION_ID is FOREIGN: this run cannot remove it.
	// This session's own, or any with the variable unset, stays STRAY.
	inTree, harness := siblingRoot(root), filepath.Join(root, ".claude", "worktrees")+"/"
	sid := env.Getenv("CLAUDE_CODE_SESSION_ID")
	n, count := 0, 0
	w, b := "", "detached"
	for _, line := range strings.Split(porcelain+"\n", "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			n++
			w, b = line[len("worktree "):], "detached"
		case strings.HasPrefix(line, "branch "):
			if f := strings.Fields(line); len(f) > 1 {
				b = f[1]
			}
		case line == "":
			switch {
			case n <= 1 || w == inTree || strings.HasPrefix(w, inTree+"/"):
			case strings.HasPrefix(w, harness):
				fmt.Fprintf(stdout, "FOREIGN: %s (%s)\n", w, b)
			case otherSessionScratchpad(w, sid):
				fmt.Fprintf(stdout, "FOREIGN: %s (%s)\n", w, b)
			default:
				fmt.Fprintf(stdout, "STRAY: %s (%s)\n", w, b)
				count++
			}
			w = ""
		}
	}

	if count > 0 {
		fmt.Fprintf(stdout, "LOCATION-STRAY: %s — %d\n", root, count)
		return 1
	}
	fmt.Fprintf(stdout, "LOCATION-OK: %s\n", root)
	return 0
}

// otherSessionScratchpad reports whether w sits under the scratchpad of a
// session other than sid; with sid empty no scratchpad is another session's.
func otherSessionScratchpad(w, sid string) bool {
	m := wlScratchpad.FindStringSubmatch(w)
	return m != nil && sid != "" && m[1] != sid
}
