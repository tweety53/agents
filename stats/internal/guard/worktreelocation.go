package guard

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"syscall"
)

// checkWorktreeLocation is scripts/check-worktree-location.sh: that script's
// header comment is the contract -- one STRAY line per registered worktree
// outside <parent>/<project>-worktrees/ (design.md:
// sibling-worktrees-layout), then ONE verdict line, LOCATION-OK (exit 0)
// or LOCATION-STRAY (exit 1); exit 2 with NOTHING on stdout when <project>
// is not a readable directory or `git worktree list` fails -- an inability
// is never reported as a verdict.
func init() { Registry["check-worktree-location"] = checkWorktreeLocation }

func checkWorktreeLocation(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-worktree-location: "+format+"\n", a...)
		return 2
	}
	if len(args) != 1 {
		return die("usage: check-worktree-location.sh <project>")
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
	inTree := siblingRoot(root)
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
			if n > 1 && w != inTree && !strings.HasPrefix(w, inTree+"/") {
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

// siblingRoot is where every worktree of the repository whose main checkout
// is main lives: <dirname main>/<basename main>-worktrees (design.md:
// sibling-worktrees-layout). Kickoff creates there, this guard accepts only
// there, and migrate-worktrees moves there — one rule, one spelling.
func siblingRoot(main string) string {
	return filepath.Join(filepath.Dir(main), filepath.Base(main)+"-worktrees")
}
