package guard

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// landSelfReviewReport is scripts/land-self-review-report.sh: that script's
// header comment is the contract -- the one landing chain for a self-review
// report or context bundle, in its order, with exit 0 landed or nothing to
// land, 1 branch mismatch, 2 usage, 3 foreign staged work, and otherwise
// git's own exit code, unmasked.
//
// Every git call is the bash's g() -- git resolved over env's PATH, run as
// `git -C <repo>` from env.Dir -- with git's own output passed through to
// this guard's stdout and stderr, so whatever the caller has checked out is
// never consulted and a rejected push or a stopped rebase names itself.
func init() {
	Registry["land-self-review-report"] = landSelfReviewReport
}

const lsrrUsage = "usage: land-self-review-report.sh <repo> <branch> <subject> <add-path> [<rm-path>] [--push <base>]\n"

func landSelfReviewReport(args []string, env Env, stdout, stderr io.Writer) int {
	usage := func() int {
		fmt.Fprint(stderr, lsrrUsage)
		return 2
	}
	if len(args) < 4 {
		return usage()
	}
	repo, branch, subject, addPath := args[0], args[1], args[2], args[3]
	var rmPath, pushBase string
	for rest := args[4:]; len(rest) > 0; {
		if rest[0] == "--push" {
			if len(rest) < 2 || rest[1] == "" {
				return usage()
			}
			pushBase, rest = rest[1], rest[2:]
			continue
		}
		if rmPath != "" {
			return usage()
		}
		rmPath, rest = rest[0], rest[1:]
	}

	// GIT_CONFIG_GLOBAL is forwarded from env, as verbatimmoves.go does:
	// the process environment in production (a no-op), and the tests'
	// /dev/null, so no guard-run commit, pull or push reads the operator's
	// ~/.gitconfig.
	var gitEnv []string
	if v, ok := env.LookupEnv("GIT_CONFIG_GLOBAL"); ok {
		gitEnv = append(gitEnv, "GIT_CONFIG_GLOBAL="+v)
	}
	git := envGit(env, gitEnv...)
	g := func(args ...string) *exec.Cmd {
		cmd := git(append([]string{"-C", repo}, args...)...)
		cmd.Stdout, cmd.Stderr = stdout, stderr
		return cmd
	}
	// read is `$(g <args>)` under set -e: stdout with trailing newlines
	// stripped, stderr passed through, and git's status.
	read := func(args ...string) (string, int) {
		cmd := g(args...)
		cmd.Stdout = nil
		out, err := cmd.Output()
		return strings.TrimRight(string(out), "\n"), lsrrStatus(err, stderr)
	}

	// The start-of-run assert cannot be trusted across the run: a concurrent
	// session on a shared checkout can switch the branch between two steps
	// (observed, kan-657). The commit and the pull/push pair each re-check
	// before an irreversible step runs on the wrong branch.
	assertBranch := func(stage string) int {
		found, rc := read("branch", "--show-current")
		if rc != 0 || found == branch {
			return rc
		}
		fmt.Fprintf(stderr, "LAND-BRANCH-MISMATCH: expected %s, found %s %s\n", branch, found, stage)
		return 1
	}

	if rc := assertBranch("— nothing added, committed, pulled or pushed"); rc != 0 {
		return rc
	}
	if rc := lsrrStatus(g("add", "--", addPath).Run(), stderr); rc != 0 {
		return rc
	}
	if rmPath != "" {
		if rc := lsrrStatus(g("rm", "--", rmPath).Run(), stderr); rc != 0 {
			return rc
		}
	}

	// An `if` condition: any non-zero status, git's own error included,
	// reads as "something is staged".
	if g("diff", "--cached", "--quiet").Run() == nil {
		fmt.Fprintf(stdout, "LAND-NOTHING-TO-COMMIT: nothing staged under %s — no commit, no pull, no push\n", repo)
		return 0
	}

	// git commit takes the whole index, and the chain staged only its own
	// paths: a shared checkout holding foreign staged work must not be swept
	// into the commit (kan-657, where 121 foreign paths landed as 66ae176 and
	// reverted a just-merged change). Refuse loudly, index untouched; the
	// own-paths set is deliberately stated twice -- in the add/rm calls and
	// in expected below -- because the staging verbs and this assertion must
	// agree, and each side reads it in its own grammar. --no-renames: a
	// foreign deletion paired as a rename with the chain's own new file
	// would drop out of the name list and reopen the sweep.
	out, rc := read("diff", "--cached", "--name-only", "--no-renames")
	if rc != 0 {
		return rc
	}
	var stagedLines []string
	if out != "" {
		stagedLines = strings.Split(out, "\n")
	}
	own := addPath
	if rmPath != "" {
		own += "\n" + rmPath
	}
	// Both sides sort by an exec'd `sort` under the caller's collation, as
	// the bash's did. No sort on PATH was the shell's 127 under pipefail.
	if _, ok := lookPath(env, "sort"); !ok {
		fmt.Fprintln(stderr, "land-self-review-report: sort: command not found")
		return 127
	}
	staged, err := crSort(env, stagedLines, false)
	if err != nil {
		fmt.Fprintf(stderr, "land-self-review-report: %v\n", err)
		return 2
	}
	expected, err := crSort(env, strings.Split(own, "\n"), false)
	if err != nil {
		fmt.Fprintf(stderr, "land-self-review-report: %v\n", err)
		return 2
	}
	if s, e := strings.Join(staged, "\n"), strings.Join(expected, "\n"); s != e {
		// The expected list is `printf '%s' | tr '\n' ' '`: no trailing
		// space. `printf '%s\n' "$STAGED"` of an empty set is one empty line.
		if len(staged) == 0 {
			staged = []string{""}
		}
		fmt.Fprintf(stderr, "LAND-FOREIGN-STAGED: expected only %s— foreign staged: %s— missing from the index: %s— nothing committed, pulled or pushed; clear the staging or land from a clean checkout\n",
			strings.Join(expected, " "), lsrrWords(lsrrOnlyIn(staged, expected)), lsrrWords(lsrrOnlyIn(expected, staged)))
		return 3
	}

	if rc := assertBranch("before the commit — nothing committed, pulled or pushed"); rc != 0 {
		return rc
	}
	commit := []string{"commit", "-m", subject, "--", addPath}
	if rmPath != "" {
		commit = append(commit, rmPath)
	}
	if rc := lsrrStatus(g(commit...).Run(), stderr); rc != 0 {
		return rc
	}

	// Pull and push sit inside the same guard (F4): they can never act on a
	// branch other than the asserted one. A rejected push leaves the commit
	// local and is never retried; a rebase stopped on a conflict is never
	// aborted here.
	if pushBase != "" {
		if rc := assertBranch("before the pull/push — nothing pulled or pushed"); rc != 0 {
			return rc
		}
		if rc := lsrrStatus(g("pull", "--rebase", "origin", pushBase).Run(), stderr); rc != 0 {
			return rc
		}
		return lsrrStatus(g("push", "origin", pushBase).Run(), stderr)
	}
	return 0
}

// lsrrStatus is the `$?` bash would see: git's own status, 128+n when signal
// n killed it, and 127 -- the cause on stderr -- when git could not run.
func lsrrStatus(err error, stderr io.Writer) int {
	var ee *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &ee):
		return rrExitCode(ee.ProcessState)
	}
	fmt.Fprintf(stderr, "land-self-review-report: %v\n", err)
	return 127
}

// lsrrOnlyIn is `comm -13 <(b) <(a)`: a's lines that b does not pair, in
// a's order. Both come out of the same sort, so walking them as comm does
// reduces to a multiset difference.
func lsrrOnlyIn(a, b []string) []string {
	left := map[string]int{}
	for _, l := range b {
		left[l]++
	}
	var only []string
	for _, l := range a {
		if left[l] > 0 {
			left[l]--
			continue
		}
		only = append(only, l)
	}
	return only
}

// lsrrWords is `tr '\n' ' '` over the lines: each followed by one space.
func lsrrWords(lines []string) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l + " ")
	}
	return b.String()
}
