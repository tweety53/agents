package guard

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"syscall"
)

// The contract below moved here from the header of the deleted
// scripts/check-main-checkout-drift.sh shim (KAN-1058).
//
// flow-guard check-main-checkout-drift — name a main checkout's drift from its
// expected post-merge state, before a /flow run's preflight reaches it
// (KAN-647, observed in kan-574's archive run).
//
// Usage: flow-guard check-main-checkout-drift <main-checkout>
//
// Prints its findings, then at most one verdict line, all to stdout:
//   DRIFT-BRANCH: <path> — on <branch>, not <default>   not on the default branch
//   DRIFT-DIRTY: <path> — <n> tracked entries            tracked changes present
//   DRIFT-CLEAN: <path> — <default>                      neither finding
//
// DRIFT-BRANCH and DRIFT-DIRTY are findings: either or both may print, in
// that order. DRIFT-CLEAN prints exactly when neither does. A detached HEAD
// is named `on (detached HEAD), not <default>` rather than left blank.
//
// Exit 0 on any verdict; exit 2 with NOTHING on stdout when the argument
// is missing, is not a readable directory, is not a git repository, the
// repository has no resolvable `refs/remotes/origin/HEAD`, or `git status`
// fails there — an inability is never reported as a verdict.
//
// WHAT DRIFT MEANS. Every pipeline run works in a worktree, and the main
// checkout's expected state is the post-merge one: on the repository's
// default branch with nothing tracked modified, staged or unmerged — the
// only state from which a landing or a cleanup can proceed and in which
// the checkout reflects what actually landed. kan-574's archive run found
// two "reverse image of a landing" incidents — tracked content silently
// reverted to pre-merge state — and a main checkout left on an unrelated
// feature branch; both were discovered ad hoc, mid-run. This guard names
// both shapes before any work starts.
//
// WHY THE STATUS SHAPE IS THE CONTENT MARKER. `git status --porcelain
// --untracked-files=no` is the cheap read: one invocation, no object
// database walk. A checkout that is clean but BEHIND its default branch is
// ordinary not-pulled state, indistinguishable from benign by any local
// marker and never this guard's finding — the cleanup run's own refresh step
// owns bringing the checkout forward. Untracked files are hidden from the
// read entirely, exactly as the KAN-546 guard's identical read hides them:
// the pipeline never creates them in a main checkout, and the operator's
// own residue is not drift this gate stops for.
//
// THE DEFAULT BRANCH is what `refs/remotes/origin/HEAD` points at — the
// repository's own answer, not a name invented here. A repository without
// one cannot answer, so it exits 2 rather than guessing `main`.
//
// THE VERDICT NAMES THE PHYSICAL PATH, resolved with `cd … && pwd -P`, so a
// caller that passed a symlinked path still sees the real checkout named —
// the same resolution `git worktree list` applies when it prints paths.
//
// HOW TO HAND-VERIFY A DRIFT VERDICT. Run the same reads by hand:
// `git -C <main-checkout> branch --show-current` against the branch
// `git -C <main-checkout> symbolic-ref refs/remotes/origin/HEAD` names with
// its `refs/remotes/origin/` prefix stripped, and
// `git -C <main-checkout> status --porcelain --untracked-files=no` — those,
// and only those, are the inputs the findings report. A main checkout
// mid-rebase or carrying an unrelated staged file is a typical structural
// cause — fix the cause, never the verdict.

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
		fmt.Fprintln(stderr, "usage: flow-guard check-main-checkout-drift <main-checkout>")
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
