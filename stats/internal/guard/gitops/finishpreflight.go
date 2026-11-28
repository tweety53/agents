package gitops

import (
	"fmt"
	"io"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The contract below moved here from the header of the deleted
// scripts/check-finish-preflight.sh shim (KAN-1058).
//
// flow-guard check-finish-preflight — decide whether /flow should integrate
// (run 1) or clean up (run 2), or refuse because it cannot tell.
//
// Usage: flow-guard check-finish-preflight <worktree> <base-ref> <recorded-merge-base|->
//
// Prints ONE verdict line to stdout, and after a RUN1 a second, check-base-moved's
// own CLEAR / MOVED / REFUSE line about the same worktree:
//   RUN1: <reason>    integrate — the branch has not reached the base branch
//   RUN2: <reason>    clean up  — the branch is merged and nothing is outstanding
//   REFUSE: <reason>  stop and ask the operator
//
// Exit 0 whenever a verdict was reached; exit 2 when the tree cannot be read.
// The VERDICT carries the answer, not the exit status: an exit-code-only
// protocol makes "cannot determine" indistinguishable from "violation", which
// is the exact confusion this script exists to remove. Exit 2 keeps the
// meaning this repository's other guards give it.
//
// WHY SIGNAL ORDER IS THE WHOLE FIX. A branch with no commits of its own is an
// ancestor of EVERY branch, so `merge-base --is-ancestor` answers "merged" on a
// branch whose work is staged and never committed — the normal IN_PROGRESS
// state, since /flow's implement phase may not commit before a PR exists. Run 2 would then
// archive the change and `git worktree remove --force` the worktree holding all
// of it. Comparing HEAD with the merge base RECORDED IN THE STATE FILE catches
// it; counting commits ahead of the base branch does NOT, because a genuinely
// merged branch is also zero ahead once its commit joins the base branch.
//
// Base-branch resolution is owned by flow-guard resolve-base-branch, not by this
// script — the caller resolves it and passes it in. That guard already
// carries a hard-won assertion against HEAD@{upstream}; a second copy here
// could drift from it.
//
// What THIS script owns (KAN-88, design.md:
// preflight-resolves-remote-tracking): the base ref it was HANDED may name a
// local branch that has fallen behind its remote-tracking counterpart — the
// caller composes `origin/$BASE`, but a caller that regresses to a bare name
// must not silently feed a stale local branch into the RUN1/RUN2 decision.
// EFFECTIVE_REF (resolveRemoteBase in stats/internal/guard/gitops/resolveremotebase.go)
// substitutes `refs/remotes/origin/<base-ref>` for BASE_REF when it resolves
// (design.md: remote-lookup-is-a-preference-not-a-rewrite — a preference, not
// a rewrite: `origin/main` looks up `origin/origin/main`, finds nothing, and
// passes through unchanged); every verdict line names EFFECTIVE_REF rather
// than the raw argument, so each names the ref the test actually ran against.
//
// HOW TO HAND-VERIFY A REFUSE VERDICT (KAN-446). Re-derive the refused signal
// by hand: which of the branch states or the worktree's uncommitted entries
// fired — fix the cause, never the verdict.

// checkFinishPreflight is `flow-guard check-finish-preflight`: the comment above is the contract -- one RUN1/RUN2/REFUSE verdict line on
// stdout, exit 0 whenever a verdict was reached, exit 2 when the tree cannot
// be read. The reasoning for each step, moved here from the bash body it
// replaced (d71a2327), sits beside the code it explains.
func init() {
	core.Registry["check-finish-preflight"] = checkFinishPreflight
}

func checkFinishPreflight(args []string, env core.Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktree, baseRef, recorded := arg(0), arg(1), arg(2)
	if worktree == "" || baseRef == "" || recorded == "" {
		fmt.Fprint(stderr, baseRefUsage("flow-guard check-finish-preflight"))
		core.Reason(env, "usage")
		return 2
	}
	if !core.IsDir(core.Abs(env, worktree)) {
		fmt.Fprintf(stderr, "check-finish-preflight: %s is not a directory — cannot determine anything\n", worktree)
		core.Reason(env, "not-a-directory")
		return 2
	}
	git := core.EnvGit(env)
	if git("-C", worktree, "rev-parse", "--git-dir").Run() != nil {
		fmt.Fprintf(stderr, "check-finish-preflight: %s is not a git worktree — cannot determine anything\n", worktree)
		core.Reason(env, "not-a-worktree")
		return 2
	}

	// (a) No recorded merge base: an honest unknown, never an inferred verdict.
	if recorded == "-" {
		fmt.Fprintf(stdout, "REFUSE: no merge base recorded for %s — cannot tell an unmerged branch from a merged one\n", worktree)
		return 0
	}

	// Every ref this guard did not choose itself arrives from a state file
	// and is passed after --end-of-options, so a value beginning with `-` is
	// read as a ref and rejected rather than parsed as a git option.
	headSHA, ok := core.Capture(git("-C", worktree, "rev-parse", "--verify", "--end-of-options", "HEAD^{commit}"))
	if !ok {
		fmt.Fprintf(stderr, "check-finish-preflight: cannot resolve HEAD in %s\n", worktree)
		core.Reason(env, "head-unresolved")
		return 2
	}
	recordedSHA, ok := core.Capture(git("-C", worktree, "rev-parse", "--verify", "--end-of-options", recorded+"^{commit}"))
	if !ok {
		fmt.Fprintf(stdout, "REFUSE: recorded merge base '%s' does not resolve in %s\n", recorded, worktree)
		return 0
	}

	// (b) HEAD is still the merge base: the branch has no commits of its own.
	//     This MUST be tested before the ancestor test. It reads only HEAD and
	//     the recorded merge base, so it is deliberately answered BEFORE the
	//     base ref is resolved below — an unresolvable base ref must not hide
	//     this shape, which is the one this guard exists to catch.
	if headSHA == recordedSHA {
		return runOne(env, stdout, stderr, worktree, baseRef, recorded, "RUN1: HEAD is still the recorded merge base — the branch has no commits of its own")
	}

	// Resolve the effective base ref (KAN-88) before testing whether it
	// resolves, so the substitution — not the raw argument — is what gets
	// tested and named. A base ref that does not resolve would make the
	// ancestor test below fail for an environmental reason and read as "not
	// merged" — a RUN1 verdict reached by accident — so it is named instead.
	ref := ResolveRemoteBase(git, worktree, baseRef)
	if git("-C", worktree, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}").Run() != nil {
		fmt.Fprintf(stdout, "REFUSE: base ref '%s' does not resolve in %s — cannot test whether HEAD reached it\n", ref, worktree)
		return 0
	}

	// (c) The ancestor test. Only exit 1 means "not an ancestor"; anything
	//     else is git failing to answer, which must not be read as a RUN1
	//     verdict. git's stderr is discarded so no chatter can prefix the
	//     verdict line for a caller that merges the two streams.
	if rc := core.ExitCode(git("-C", worktree, "merge-base", "--is-ancestor", "--end-of-options", headSHA, ref).Run()); rc == 1 {
		return runOne(env, stdout, stderr, worktree, baseRef, recorded, fmt.Sprintf("RUN1: HEAD is not an ancestor of %s — not merged", ref))
	} else if rc != 0 {
		fmt.Fprintf(stderr, "check-finish-preflight: merge-base failed in %s (exit %d) — cannot determine anything\n", worktree, rc)
		core.Reason(env, "merge-base-failed")
		return 2
	}

	// (d) Merged by ancestry, so nothing should be outstanding. Tracked
	//     changes and untracked-unignored files both count; ignored files do
	//     not, because they are disclosed separately at removal time rather
	//     than gating. An unreadable status (a held index lock, a transient
	//     I/O error, a worktree that became unreadable mid-run) is the same
	//     class of failure as an unreadable worktree — exit 2, named — never a
	//     silent 0 that falls through to RUN2, the destructive verdict.
	status, ok := core.Capture(git("-C", worktree, "status", "--porcelain", "--untracked-files=normal"))
	if !ok {
		fmt.Fprintf(stderr, "check-finish-preflight: cannot read the worktree status in %s — cannot determine anything\n", worktree)
		core.Reason(env, "cannot-read-the-worktree-status")
		return 2
	}
	if status != "" {
		fmt.Fprintf(stdout, "REFUSE: %s contains HEAD, but %s has %d uncommitted entries — a merged change should have nothing left to commit\n",
			ref, worktree, strings.Count(status, "\n")+1)
		return 0
	}

	fmt.Fprintf(stdout, "RUN2: HEAD is an ancestor of %s, differs from the recorded merge base, and the worktree is clean\n", ref)
	return 0
}

// runOne prints a RUN1 verdict and, on the line after it, check-base-moved's
// answer about the same worktree (CLEAR, MOVED or REFUSE, its own grammar): the
// base only matters to a branch that is about to integrate, so one call answers
// both. The base-moved answer is read first, so a tree it cannot read is exit 2
// with nothing on stdout, never a RUN1 without its companion line.
func runOne(env core.Env, stdout, stderr io.Writer, worktree, baseRef, recorded, run1 string) int {
	bm := baseMoved(env, worktree, baseRef, recorded, stderr)
	if bm.code != 0 {
		core.Reason(env, "base-moved-unreadable")
		return bm.code
	}
	fmt.Fprintln(stdout, run1)
	fmt.Fprintln(stdout, bm.line)
	return 0
}
