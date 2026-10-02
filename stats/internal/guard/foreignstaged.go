package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// checkForeignStaged is scripts/check-foreign-staged.sh: that script's header
// comment is the contract -- one FOREIGN-STAGED line per staged or unmerged
// entry, then one STAGED-CLEAN/STAGED-FOREIGN verdict line, exit 0 on either;
// exit 2 with nothing on stdout when it cannot answer.
//
// WHAT "FOREIGN" MEANS. Every pipeline run works in a worktree; nothing in
// `/flow` or `/flow-fast` ever stages into a main checkout. So ANY staged
// entry in a main checkout is foreign to the change in flight — residue of
// some other session — and is listed as such, whatever it contains. The
// operator decides what it is and what becomes of it; this guard classifies
// nothing beyond "staged".
//
// WHY ONLY STAGED ENTRIES. Staged work is what a resumed run is tempted to
// `git reset` or `git stash` when the preflight's main-checkout assertion
// refuses — the high-judgment surgery kan-437's run 2 performed inline
// across three repos. Unstaged worktree modifications also refuse that
// assertion, but they are not the reset/stash target class and naming them
// "foreign staged work" would be false; they stay the preflight's own
// finding to report. Untracked files are hidden from the status read
// entirely, exactly as the preflight's assertion hides them.
func init() { Registry["check-foreign-staged"] = checkForeignStaged }

func checkForeignStaged(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-foreign-staged: "+format+"\n", a...)
		return 2
	}
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: check-foreign-staged.sh <main-checkout>")
		return 2
	}
	// The physical path the header's verdict names.
	root, ok := cdPhysical(env, args[0])
	if !ok {
		return die("%s is not a directory", args[0])
	}
	git := envGit(env)
	if git("-C", root, "rev-parse", "--git-dir").Run() != nil {
		return die("%s is not a git repository", root)
	}
	status, ok := capture(git("-C", root, "status", "--porcelain", "--untracked-files=no"))
	if !ok {
		return die("cannot read the status of %s", root)
	}

	count := 0
	if status != "" {
		for _, line := range strings.Split(status, "\n") {
			// An intent-to-add entry (`git add -N`) is index work like any
			// other stage, but porcelain prints its index code blank — ` A
			// <path>` — so the filter also lists a line whose SECOND column is
			// `A`: that second-column A is the only index-work state a blank
			// first column can hide, and the preflight's identical status read
			// refuses on it, so hiding it would leave the surface silent on a
			// state the very next gate stops for.
			if !strings.HasPrefix(line, " ") || strings.HasPrefix(line, " A") {
				fmt.Fprintf(stdout, "FOREIGN-STAGED: %s\n", line)
				count++
			}
		}
	}
	if count == 0 {
		fmt.Fprintf(stdout, "STAGED-CLEAN: %s\n", root)
	} else {
		fmt.Fprintf(stdout, "STAGED-FOREIGN: %s — %d\n", root, count)
	}
	return 0
}

// cdPhysical is the bash's `cd "$p" && pwd -P`: p joined to the caller's
// cwd and made physical, or false where cd refuses -- an empty argument
// (bash 5's "null directory"), a path that is not a searchable directory.
func cdPhysical(env Env, p string) (string, bool) {
	if p == "" {
		return "", false
	}
	dir, err := filepath.EvalSymlinks(pcAbs(env, p))
	if err != nil {
		return "", false
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return "", false
	}
	return dir, syscall.Access(dir, 1) == nil
}
