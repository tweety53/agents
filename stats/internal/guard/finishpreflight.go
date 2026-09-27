package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// checkFinishPreflight is scripts/check-finish-preflight.sh: that script's
// header comment is the contract -- one RUN1/RUN2/REFUSE verdict line on
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
		fmt.Fprint(stderr, baseRefUsage("check-finish-preflight.sh"))
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

	// (e) The main-checkout assertion — see "THE MAIN-CHECKOUT ASSERTION" in
	//     the shim's header (KAN-462 §4, design.md
	//     main-checkout-is-asserted-not-moved) for the REFUSE grammar below.
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

	strippedBase := strings.TrimPrefix(baseRef, "origin/")
	branch, ok := capture(git("-C", mainCheckout, "branch", "--show-current"))
	if !ok {
		fmt.Fprintf(stderr, "check-finish-preflight: cannot read the current branch of the main checkout %s\n", mainCheckout)
		return 2
	}
	if branch != strippedBase {
		if branch == "" {
			branch = "(detached HEAD)"
		}
		fmt.Fprintf(stdout, "REFUSE: main checkout %s is on %s, not %s\n", mainCheckout, branch, strippedBase)
		return 0
	}

	mcStatus, ok := capture(git("-C", mainCheckout, "status", "--porcelain", "--untracked-files=no"))
	if !ok {
		fmt.Fprintf(stderr, "check-finish-preflight: cannot read the main checkout status in %s\n", mainCheckout)
		return 2
	}
	if mcStatus != "" {
		fmt.Fprintf(stdout, "REFUSE: main checkout %s has tracked changes\n", mainCheckout)
		return 0
	}

	// check-worktree-location stays bash (design.md: exec-unported-siblings),
	// exec'd from beside the shim with the bash's argv. The shim exports its
	// path as $SCRIPT_DIR/check-worktree-location.sh — the bash guard's own
	// spelling, and the one check-guard-symlinks rule 2 reads to require the
	// sibling wherever the shim is carried.
	location := env.Getenv("FLOW_GUARD_WORKTREE_LOCATION")
	if location == "" {
		fmt.Fprintln(stderr, "check-finish-preflight: FLOW_GUARD_WORKTREE_LOCATION is unset — run scripts/check-finish-preflight.sh, which sets it")
		return 2
	}
	if syscall.Access(location, 1) != nil { // `[ -x ]`
		fmt.Fprintf(stderr, "check-finish-preflight: %s is missing or not executable — cannot determine anything\n", location)
		return 2
	}
	var locOut, locErr bytes.Buffer
	cmd := exec.Command(location, mainCheckout)
	cmd.Dir = env.Dir
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
		fmt.Fprintf(stderr, "check-finish-preflight: check-worktree-location.sh exited %d against %s — cannot determine anything\n", rc, mainCheckout)
		return 2
	}

	fmt.Fprintf(stdout, "RUN2: HEAD is an ancestor of %s, differs from the recorded merge base, and the worktree is clean\n", ref)
	return 0
}

// exitCode is `$?` after a command: its exit status, 128+n when a signal
// killed it, and 126 — bash's "found but cannot execute" — when it never
// started.
func exitCode(err error) int {
	var ee *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &ee):
		return rrExitCode(ee.ProcessState)
	}
	return 126
}
