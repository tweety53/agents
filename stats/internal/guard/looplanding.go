package guard

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// check-loop-landing.sh <repo> <branch>: may a pipeline-fix-* or
// self-review-* branch land on the default branch? Only once flowd holds,
// for its loop -- under the pushing repository's project, or, for a loop
// branch with none there, under the project of the agents checkout this
// guard is installed from -- a completed plan, a completed first fix and a
// review, every row closed, the highest review `clean` -- no Critical or Important -- and
// no red-line fix (`fix-v<v>`) after it. protect-main-checkout.py runs it on
// a landing push typed by hand; pipeline-fix.sh land runs it before its own.
// One admission skips the rows: a self-review-* branch whose diff from
// origin/HEAD touches only docs/prompt-tuning/ carries prompt-tuning cases and
// no code, so no loop review had anything to judge (kan-1032 design.md §4).
func init() {
	Registry["check-loop-landing"] = checkLoopLanding
}

var (
	lpPipelineBranch = regexp.MustCompile(`^pipeline-fix-(.+)-([0-9]+)$`)
	lpReviewSuffix   = regexp.MustCompile(`^review-([0-9]+)$`)
)

// lpFromBranch is the loop a branch is, or ok false for any other branch. A
// change's own spectre/<name> is the self-review loop that runs on it, which
// files its rows under <name>; lpOwnBranch tells the two apart.
func lpFromBranch(branch string) (lpLoop, bool) {
	if name, ok := strings.CutPrefix(branch, "spectre/"); ok && plainChangeName(name) {
		return lpLoop{change: name, token: "self-review-" + name, prefix: "self-review-" + name + "-"}, true
	}
	if m := lpPipelineBranch.FindStringSubmatch(branch); m != nil && plainChangeName(branch) {
		return lpLoop{change: branch, token: branch, prefix: "pipeline-fix-" + m[2] + "-"}, true
	}
	if name, ok := strings.CutPrefix(branch, "self-review-"); ok && plainChangeName(name) {
		return lpLoop{change: branch, token: branch, prefix: branch + "-"}, true
	}
	return lpLoop{}, false
}

// lpOwnBranch is whether branch is a change's own branch: a loop only when its
// rows say a self-review loop ran on it.
func lpOwnBranch(branch string) bool { return strings.HasPrefix(branch, "spectre/") }

// lpLandingRefusals is every reason the loop's rows refuse a landing, none
// when it may land.
func lpLandingRefusals(lp lpLoop, rows []lpDispatch) []string {
	var plan, fix bool
	var open, bad []string
	lastReview, lastReviewSeq, lastFixVSeq := -1, -1, -1
	lastReviewOutcome, lastReviewKey := "", ""
	for _, d := range rows {
		suffix := strings.TrimPrefix(d.Key, lp.prefix)
		if d.EndedAt == nil {
			open = append(open, d.Key)
		}
		switch lpFamily(suffix) {
		case "plan":
			plan = plan || d.Outcome == "completed"
		case "fix":
			fix = fix || d.Outcome == "completed"
		case "fix-v":
			lastFixVSeq = d.Seq
		case "review":
			m := lpReviewSuffix.FindStringSubmatch(suffix)
			if m == nil {
				bad = append(bad, d.Key)
				continue
			}
			n, _ := strconv.Atoi(m[1])
			if n >= lastReview {
				lastReview, lastReviewSeq, lastReviewOutcome, lastReviewKey = n, d.Seq, d.Outcome, d.Key
			}
		}
	}
	var out []string
	if !plan {
		out = append(out, fmt.Sprintf("PLAN — no completed %splan dispatch", lp.prefix))
	}
	if !fix {
		out = append(out, fmt.Sprintf("FIX — no completed %sfix dispatch", lp.prefix))
	}
	for _, k := range open {
		out = append(out, fmt.Sprintf("OPEN — %s has not been closed: scripts/pipeline-fix.sh end %s <outcome>", k, k))
	}
	for _, k := range bad {
		out = append(out, fmt.Sprintf("REVIEW — %s carries no numeric round: a review key is %sreview-<r>", k, lp.prefix))
	}
	switch {
	case lastReview < 0:
		out = append(out, fmt.Sprintf("REVIEW — no %sreview-<r> dispatch", lp.prefix))
	case lastReviewOutcome != "clean":
		out = append(out, fmt.Sprintf("REVIEW — %s ended %q, not clean: fix its Critical and Important findings, then the next review", lastReviewKey, lastReviewOutcome))
	case lastFixVSeq > lastReviewSeq:
		out = append(out, fmt.Sprintf("REVIEW — a %sfix-v<v> dispatch came after %s: run the next review", lp.prefix, lastReviewKey))
	}
	return out
}

// lpPromptTuningOnly is whether branch's diff from origin/HEAD is non-empty
// and every path lies under docs/prompt-tuning/.
func lpPromptTuningOnly(env Env, git, repo, branch string) bool {
	rc, base := kwRun(env, "", git, "-C", repo, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if rc != 0 {
		return false
	}
	rc, out := kwRun(env, "", git, "-C", repo, "diff", "--no-renames", "--name-only", strings.TrimSpace(base)+"..."+branch)
	paths := strings.Fields(out)
	if rc != 0 || len(paths) == 0 {
		return false
	}
	for _, p := range paths {
		if !strings.HasPrefix(p, "docs/prompt-tuning/") {
			return false
		}
	}
	return true
}

func checkLoopLanding(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-loop-landing"
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		fmt.Fprint(stderr, "usage: check-loop-landing.sh <repo> <branch>\n")
		return 2
	}
	lp, ok := lpFromBranch(args[1])
	if !ok {
		fmt.Fprintf(stdout, "LOOP-LANDING-NA %s\n", args[1])
		return 0
	}
	git, ok := panelResolveGit(env, self, stderr)
	if !ok {
		return 2
	}
	main, ok := lpMain(env, git, smcAbs(env, args[0]))
	if !ok {
		fmt.Fprintf(stderr, "%s: %s is not a git checkout\n", self, args[0])
		return 2
	}
	if strings.HasPrefix(args[1], "self-review-") && lpPromptTuningOnly(env, git, smcAbs(env, args[0]), args[1]) {
		fmt.Fprintf(stdout, "LOOP-LANDING-OK %s (prompt-tuning cases only)\n", args[1])
		return 0
	}
	rows, ok := lpRows(env, main, lp)
	if !ok && lpOwnBranch(args[1]) {
		// ponytail: an unreadable flowd lets a change branch land unchecked, so no
		// ordinary /flow landing depends on flowd; its self-review rows go unread then.
		fmt.Fprintf(stdout, "LOOP-LANDING-NA %s (flowd unreadable)\n", args[1])
		return 0
	}
	if !ok {
		fmt.Fprintf(stderr, "%s: flow record dispatches -change %s failed -- cannot answer (is flowd up?)\n", self, lp.change)
		return 2
	}
	// A loop run from the agents checkout files every row under that checkout's
	// project, so its satellite branches -- the same loop branch in another
	// repository it carried a task into -- find their rows there.
	if root := env.Getenv("FLOW_GUARD_REPO_ROOT"); len(rows) == 0 && !lpOwnBranch(args[1]) && root != "" {
		if agents, ok := lpMain(env, git, root); ok && agents != main {
			if r, ok := lpRows(env, agents, lp); ok {
				rows = r
			}
		}
	}
	// A change's own branch holding only the pass row is no loop on it: the pass
	// is filed under the change, while a loop run from the agents checkout on its
	// own self-review-<name> branch files every other row there. Judge that loop,
	// since its satellite commits ride this branch.
	if lpOwnBranch(args[1]) && lpPassOnly(lp, rows) {
		rows = nil
		if root := env.Getenv("FLOW_GUARD_REPO_ROOT"); root != "" {
			if agents, ok := lpMain(env, git, root); ok && agents != main {
				sep := lpLoop{change: lp.token, token: lp.token, prefix: lp.prefix}
				if r, ok := lpRows(env, agents, sep); ok {
					rows = r
				}
			}
		}
	}
	if lpOwnBranch(args[1]) && len(rows) == 0 {
		fmt.Fprintf(stdout, "LOOP-LANDING-NA %s\n", args[1])
		return 0
	}
	refusals := lpLandingRefusals(lp, rows)
	for _, r := range refusals {
		fmt.Fprintf(stdout, "LOOP-LANDING-REFUSED: %s\n", r)
	}
	if len(refusals) > 0 {
		return 1
	}
	fmt.Fprintf(stdout, "LOOP-LANDING-OK %s\n", args[1])
	return 0
}

// lpPassOnly is whether every row is the pass, which no loop step follows.
func lpPassOnly(lp lpLoop, rows []lpDispatch) bool {
	for _, d := range rows {
		if lpFamily(strings.TrimPrefix(d.Key, lp.prefix)) != "pass" {
			return false
		}
	}
	return len(rows) > 0
}
