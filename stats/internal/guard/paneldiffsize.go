package guard

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-panel-diff-size.sh shim (KAN-1058).
//
// flow-guard check-panel-diff-size — report whether a branch's diff is within the
// review panel's reading cap.
//
// Usage: flow-guard check-panel-diff-size <worktree> <merge-base> [cap]
//
// Prints two lines to stdout: the changed-line count, then the verdict
// ("under cap" or "over cap") — so a caller can quote either line rather
// than re-derive the number.
//
// Exit codes: 0 at or under the cap; 1 over the cap; 2 cannot answer at all.
//
// WHY THE CAP IS 5000. KAN-109's own measurement found a seven-slot review
// panel costing 948k tokens to read a 9877-line diff — about 96 tokens per
// diff line across that whole panel, or roughly 14 per line per slot. The
// cap was 2000 on that basis, chosen as roughly a fifth of the measured
// diff.
//
// The operator raised it to 5000 on 2026-08-28, during KAN-341, whose own
// cross-repo diff ran past 2000 and whose panel they chose to run anyway.
// Raising the default records that choice once rather than re-answering the
// same over-cap prompt on every change of that size. On the rate above, a
// four-slot panel at 5000 lines projects to roughly 275k tokens — the same
// order the 2000 cap was set to protect — while a seven-slot panel at that
// length would be about 480k, so a large roster against a full cap is still
// worth a second thought even though the guard now permits it.
//
// That projection is arithmetic from KAN-109's one measurement, not a second
// measurement: nothing has re-measured a real panel at 5000 lines.
//
// WHY BOTH COMMITTED WORK AND THE WORKING TREE COUNT. `final-review.diff`,
// the text the panel actually reads, carries everything that changed since
// the merge base: commits already made plus whatever is still staged or
// unstaged in the working tree. Measuring only `merge-base..HEAD` would
// report zero for a run whose work sits entirely uncommitted, understating
// the exact diff the panel is about to be handed.

// checkPanelDiffSize is scripts/check-panel-diff-size.sh: that script's
// header comment is the contract (and carries why the cap is 5000 and why
// committed work and the working tree both count) -- two stdout lines, the
// changed-line count then "under cap" or "over cap"; exit 0 at or under the
// cap, 1 over it, 2 cannot answer. The reasoning for each step, moved here
// from the bash body it replaced (c9b8fde8), sits beside the code it
// explains.
func init() {
	Registry["check-panel-diff-size"] = checkPanelDiffSize
}

var (
	pdsInsertions = regexp.MustCompile(`([0-9]+) insertion`)
	pdsDeletions  = regexp.MustCompile(`([0-9]+) deletion`)
	pdsDecimal    = regexp.MustCompile(`^[0-9]+$`)
)

func checkPanelDiffSize(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktree, mergePoint, capArg := arg(0), arg(1), arg(2)
	if capArg == "" { // ${3:-5000}: unset or empty reads as the default
		capArg = "5000"
	}
	fail := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-panel-diff-size: "+format+"\n", a...)
		return 2
	}
	if worktree == "" {
		return fail("usage: flow-guard check-panel-diff-size <worktree> <merge-base> [cap]")
	}
	if !isDir(smcAbs(env, worktree)) {
		return fail("%s is not a directory", worktree)
	}
	git := envGit(env)
	if git("-C", worktree, "rev-parse", "--git-dir").Run() != nil {
		return fail("%s is not a directory containing a git repository", worktree)
	}
	if mergePoint == "" {
		return fail("merge base argument is required")
	}
	// Every ref this guard did not choose itself arrives from a caller, so a
	// value beginning with `-` is read as a ref and rejected rather than
	// parsed as a git option.
	if git("-C", worktree, "rev-parse", "--verify", "--end-of-options", mergePoint).Run() != nil {
		return fail("merge base does not resolve in %s: '%s'", worktree, mergePoint)
	}
	if !pdsDecimal.MatchString(capArg) {
		return fail("cap is not a non-negative decimal integer: '%s'", capArg)
	}

	// `git diff <commit>` (no second ref) compares the named commit against
	// the current working tree, which already folds in staged and unstaged
	// changes alongside every commit made since -- one shortstat covers
	// committed work and the working tree together.
	// --no-renames pins the measured size against the machine's git config:
	// with rename detection a pure rename counts as zero changed lines,
	// without it as a full delete+add -- the same tree would measure
	// differently.
	// git's stderr passes through. The bash ran this under `set -e`, so a
	// failing diff ended it with git's own exit code -- and a git exit 1
	// read as over cap; the port answers 2, as the header promises.
	// The worktree's `## panel diff exclude` globs are left out of the
	// count, as they are out of the panel diff.
	cmd := git(append([]string{"-C", worktree, "diff", "--no-renames", "--shortstat", "--end-of-options", mergePoint}, wpdExcludes(smcAbs(env, worktree))...)...)
	cmd.Stderr = stderr
	out, err := cmd.Output()
	if err != nil {
		return fail("git diff failed in %s -- cannot answer", worktree)
	}
	shortstat := strings.TrimRight(string(out), "\n")
	total := 0
	for _, re := range []*regexp.Regexp{pdsInsertions, pdsDeletions} {
		if m := re.FindStringSubmatch(shortstat); m != nil {
			n, _ := strconv.Atoi(m[1])
			total += n
		}
	}
	fmt.Fprintln(stdout, total)

	// `[ "$TOTAL" -le "$CAP" ]`: a cap past the integer range makes bash's
	// test fail, which lands in the else branch -- over cap.
	limit, err := strconv.ParseInt(capArg, 10, 64)
	if err == nil && int64(total) <= limit {
		fmt.Fprintln(stdout, "under cap")
		return 0
	}
	fmt.Fprintln(stdout, "over cap")
	return 1
}
