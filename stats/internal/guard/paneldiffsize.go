package guard

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

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
		return fail("usage: check-panel-diff-size.sh <worktree> <merge-base> [cap]")
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
	cmd := git("-C", worktree, "diff", "--no-renames", "--shortstat", "--end-of-options", mergePoint)
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
