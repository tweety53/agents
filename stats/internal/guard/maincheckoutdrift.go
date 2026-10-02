package guard

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"syscall"
)

// checkMainCheckoutDrift is scripts/check-main-checkout-drift.sh: that
// script's header comment is the contract -- DRIFT-BRANCH and DRIFT-DIRTY
// findings in that order, DRIFT-CLEAN exactly when neither prints, exit 0 on
// any verdict; exit 2 with NOTHING on stdout when it cannot answer. Every
// git read is silenced, as the bash's 2>/dev/null silenced it: the named
// cause on stderr is the whole of a refusal.
func init() {
	Registry["check-main-checkout-drift"] = checkMainCheckoutDrift
}

func checkMainCheckoutDrift(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-main-checkout-drift"
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, self+": "+format+"\n", a...)
		return 2
	}
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: check-main-checkout-drift.sh <main-checkout>")
		return 2
	}

	// THE VERDICT NAMES THE PHYSICAL PATH: `cd "$1" && pwd -P` -- `cd`
	// cleans `..` lexically (its default -L), then every symlink resolves.
	// `cd` also needs search permission, which a bare stat does not test.
	root, err := filepath.EvalSymlinks(pcAbs(env, args[0]))
	if err != nil || !isDir(root) || syscall.Access(root, 1) != nil {
		return die("%s is not a directory", args[0])
	}
	git := envGit(env)
	if git("-C", root, "rev-parse", "--git-dir").Run() != nil {
		return die("%s is not a git repository", root)
	}

	// THE DEFAULT BRANCH is what refs/remotes/origin/HEAD points at; with no
	// such ref the guard cannot answer rather than guessing `main`.
	def, ok := capture(git("-C", root, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD"))
	if !ok {
		return die("%s has no resolvable refs/remotes/origin/HEAD — the default branch cannot be named", root)
	}
	def = strings.TrimPrefix(def, "refs/remotes/origin/")
	// A symbolic ref can dangle -- `git update-ref -d` on the branch it
	// points at leaves the symref behind -- and a name that resolves to
	// nothing is not an answer. Verify the target before trusting it.
	if git("-C", root, "rev-parse", "--verify", "--quiet", "refs/remotes/origin/"+def+"^{commit}").Run() != nil {
		return die("%s's refs/remotes/origin/HEAD dangles — %s does not resolve", root, def)
	}

	branch, ok := capture(git("-C", root, "branch", "--show-current"))
	if !ok {
		return die("cannot read the current branch of %s", root)
	}
	if branch == "" {
		branch = "(detached HEAD)"
	}

	// WHY THE STATUS SHAPE IS THE CONTENT MARKER: one cheap read, untracked
	// files hidden exactly as the KAN-546 guard's identical read hides them.
	status, ok := capture(git("-C", root, "status", "--porcelain", "--untracked-files=no"))
	if !ok {
		return die("cannot read the status of %s", root)
	}
	entries := 0
	if status != "" {
		entries = strings.Count(status, "\n") + 1
	}

	if branch != def {
		fmt.Fprintf(stdout, "DRIFT-BRANCH: %s — on %s, not %s\n", root, branch, def)
	}
	if entries != 0 {
		fmt.Fprintf(stdout, "DRIFT-DIRTY: %s — %d tracked entries\n", root, entries)
	}
	if branch == def && entries == 0 {
		fmt.Fprintf(stdout, "DRIFT-CLEAN: %s — %s\n", root, def)
	}
	return 0
}
