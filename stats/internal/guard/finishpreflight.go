package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"syscall"
)

// The contract below moved here from the header of the deleted
// scripts/check-finish-preflight.sh shim (KAN-1058).
//
// flow-guard check-finish-preflight — decide whether /flow should integrate
// (run 1) or clean up (run 2), or refuse because it cannot tell.
//
// Usage: flow-guard check-finish-preflight <worktree> <base-ref> <recorded-merge-base|->
//
// Prints ONE verdict line to stdout:
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
// THE STRAY-WORKTREE ASSERTION. A would-be RUN2 verdict is REFUSEd instead
// — `REFUSE: stray worktree(s): …` — when `flow-guard check-worktree-location
// <main-checkout>` exits 1, relaying its STRAY lines; the main checkout is
// resolved from <worktree> via `git rev-parse --git-common-dir`, never taken
// as an argument. This check runs last, immediately before the RUN2 line it
// can still turn into a REFUSE — it never affects a RUN1 or an earlier
// REFUSE. The main checkout's own branch and tracked state are not asserted:
// run 2 only fast-forwards it, flow-guard refresh-main-checkout refuses by name what
// it cannot fast-forward, and that refusal never blocks FINISHED.
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
// EFFECTIVE_REF (resolveRemoteBase in stats/internal/guard/resolveremotebase.go)
// substitutes `refs/remotes/origin/<base-ref>` for BASE_REF when it resolves
// (design.md: remote-lookup-is-a-preference-not-a-rewrite — a preference, not
// a rewrite: `origin/main` looks up `origin/origin/main`, finds nothing, and
// passes through unchanged); every verdict line names EFFECTIVE_REF rather
// than the raw argument, so each names the ref the test actually ran against.
//
// HOW TO HAND-VERIFY A REFUSE VERDICT (KAN-446). Re-derive the refused signal
// by hand: which of the branch states, the worktree's uncommitted entries, or
// stray worktrees fired — fix the cause, never the verdict.

// checkFinishPreflight is `flow-guard check-finish-preflight`: the comment above is the contract -- one RUN1/RUN2/REFUSE verdict line on
// stdout, exit 0 whenever a verdict was reached, exit 2 when the tree cannot
// be read. The reasoning for each step, moved here from the bash body it
// replaced (d71a2327), sits beside the code it explains.
func init() {
	Registry["check-finish-preflight"] = checkFinishPreflight
}

func checkFinishPreflight(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktree, baseRef, recorded := arg(0), arg(1), arg(2)
	if worktree == "" || baseRef == "" || recorded == "" {
		fmt.Fprint(stderr, baseRefUsage("flow-guard check-finish-preflight"))
		return 2
	}
	if !isDir(smcAbs(env, worktree)) {
		fmt.Fprintf(stderr, "check-finish-preflight: %s is not a directory — cannot determine anything\n", worktree)
		return 2
	}
	git := envGit(env)
	if git("-C", worktree, "rev-parse", "--git-dir").Run() != nil {
		fmt.Fprintf(stderr, "check-finish-preflight: %s is not a git worktree — cannot determine anything\n", worktree)
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
	headSHA, ok := capture(git("-C", worktree, "rev-parse", "--verify", "--end-of-options", "HEAD^{commit}"))
	if !ok {
		fmt.Fprintf(stderr, "check-finish-preflight: cannot resolve HEAD in %s\n", worktree)
		return 2
	}
	recordedSHA, ok := capture(git("-C", worktree, "rev-parse", "--verify", "--end-of-options", recorded+"^{commit}"))
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
		fmt.Fprintln(stdout, "RUN1: HEAD is still the recorded merge base — the branch has no commits of its own")
		return 0
	}

	// Resolve the effective base ref (KAN-88) before testing whether it
	// resolves, so the substitution — not the raw argument — is what gets
	// tested and named. A base ref that does not resolve would make the
	// ancestor test below fail for an environmental reason and read as "not
	// merged" — a RUN1 verdict reached by accident — so it is named instead.
	ref := resolveRemoteBase(git, worktree, baseRef)
	if git("-C", worktree, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}").Run() != nil {
		fmt.Fprintf(stdout, "REFUSE: base ref '%s' does not resolve in %s — cannot test whether HEAD reached it\n", ref, worktree)
		return 0
	}

	// (c) The ancestor test. Only exit 1 means "not an ancestor"; anything
	//     else is git failing to answer, which must not be read as a RUN1
	//     verdict. git's stderr is discarded so no chatter can prefix the
	//     verdict line for a caller that merges the two streams.
	if rc := exitCode(git("-C", worktree, "merge-base", "--is-ancestor", "--end-of-options", headSHA, ref).Run()); rc == 1 {
		fmt.Fprintf(stdout, "RUN1: HEAD is not an ancestor of %s — not merged\n", ref)
		return 0
	} else if rc != 0 {
		fmt.Fprintf(stderr, "check-finish-preflight: merge-base failed in %s (exit %d) — cannot determine anything\n", worktree, rc)
		return 2
	}

	// (d) Merged by ancestry, so nothing should be outstanding. Tracked
	//     changes and untracked-unignored files both count; ignored files do
	//     not, because they are disclosed separately at removal time rather
	//     than gating. An unreadable status (a held index lock, a transient
	//     I/O error, a worktree that became unreadable mid-run) is the same
	//     class of failure as an unreadable worktree — exit 2, named — never a
	//     silent 0 that falls through to RUN2, the destructive verdict.
	status, ok := capture(git("-C", worktree, "status", "--porcelain", "--untracked-files=normal"))
	if !ok {
		fmt.Fprintf(stderr, "check-finish-preflight: cannot read the worktree status in %s — cannot determine anything\n", worktree)
		return 2
	}
	if status != "" {
		fmt.Fprintf(stdout, "REFUSE: %s contains HEAD, but %s has %d uncommitted entries — a merged change should have nothing left to commit\n",
			ref, worktree, strings.Count(status, "\n")+1)
		return 0
	}

	// (e) The stray-worktree assertion — see "THE STRAY-WORKTREE ASSERTION"
	//     in the comment above for the REFUSE grammar below. The main
	//     checkout's branch and tracked state are not asserted: run 2 only
	//     fast-forwards it, and refresh-main-checkout.sh refuses by name
	//     whatever it cannot fast-forward.
	common, ok := capture(git("-C", worktree, "rev-parse", "--git-common-dir"))
	if !ok {
		fmt.Fprintf(stderr, "check-finish-preflight: cannot resolve the main checkout from %s\n", worktree)
		return 2
	}
	if !strings.HasPrefix(common, "/") {
		common = worktree + "/" + common
	}
	// `cd "$(dirname "$COMMON_DIR")" && pwd -P`: the physical path.
	mainCheckout, err := filepath.EvalSymlinks(smcAbs(env, filepath.Dir(common)))
	if err == nil && !isDir(mainCheckout) {
		err = errors.New("not a directory")
	}
	if err != nil {
		fmt.Fprintf(stderr, "check-finish-preflight: cannot resolve the main checkout from %s\n", worktree)
		return 2
	}

	// check-worktree-location runs in this same binary (runSibling), or as
	// the executable FLOW_GUARD_WORKTREE_LOCATION names.
	if location := env.Getenv("FLOW_GUARD_WORKTREE_LOCATION"); location != "" && syscall.Access(location, 1) != nil { // `[ -x ]`
		fmt.Fprintf(stderr, "check-finish-preflight: %s is missing or not executable — cannot determine anything\n", location)
		return 2
	}
	var locOut, locErr bytes.Buffer
	cmd := siblingCmd(env, "FLOW_GUARD_WORKTREE_LOCATION", "check-worktree-location", mainCheckout)
	cmd.Stdout, cmd.Stderr = &locOut, &locErr
	rc := exitCode(cmd.Run())
	switch {
	case rc == 1:
		// The bash's `grep '^STRAY: ' | sed | paste -sd ';' | sed 's/;/; /g'`:
		// every `;` becomes `; `, one inside a path included. With no STRAY
		// line at all, grep's failure aborted the bash under `set -e` with
		// exit 1 and nothing printed; that is kept.
		var strays []string
		for _, l := range strings.Split(strings.TrimRight(locOut.String(), "\n"), "\n") {
			if s, ok := strings.CutPrefix(l, "STRAY: "); ok {
				strays = append(strays, s)
			}
		}
		if len(strays) == 0 {
			return 1
		}
		fmt.Fprintf(stdout, "REFUSE: stray worktree(s): %s\n", strings.ReplaceAll(strings.Join(strays, ";"), ";", "; "))
		return 0
	case rc == 2:
		if e := strings.TrimRight(locErr.String(), "\n"); e != "" {
			fmt.Fprintln(stderr, e)
		}
		return 2
	case rc != 0:
		fmt.Fprintf(stderr, "check-finish-preflight: flow-guard check-worktree-location exited %d against %s — cannot determine anything\n", rc, mainCheckout)
		return 2
	}

	fmt.Fprintf(stdout, "RUN2: HEAD is an ancestor of %s, differs from the recorded merge base, and the worktree is clean\n", ref)
	return 0
}
