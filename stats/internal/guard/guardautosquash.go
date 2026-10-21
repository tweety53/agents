package guard

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// guardAutosquash is scripts/guard-autosquash.sh: that script's header
// comment is the contract -- `targets` checks every fixup target resolves
// as a commit and is an ancestor of HEAD, `after` checks the base and every
// sha tasks.md records after a rebase; exit 0 every check passed, 1
// violations (one line each on stderr), 2 cannot answer. fold-fixup execs
// it through its shim (foldfixup.go), reading those exit codes and stderr
// lines. The reasoning moved here from the bash body it replaced
// (c9b8fde8).
func init() {
	Registry["guard-autosquash"] = guardAutosquash
}

func guardAutosquash(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "guard-autosquash"
	usage := func() int {
		fmt.Fprintf(stderr, "usage: %s targets <worktree> <sha> [<sha>...]\n", prog)
		fmt.Fprintf(stderr, "       %s after <worktree> <base-ref> <tasks-md>\n", prog)
		return 2
	}
	if len(args) < 1 {
		return usage()
	}
	sub, rest := args[0], args[1:]
	switch {
	case sub == "targets" && len(rest) >= 2:
	case sub == "after" && len(rest) == 3:
	default:
		return usage()
	}
	wt, rest := rest[0], rest[1:]
	cannot := func(msg string) int {
		fmt.Fprintf(stderr, "%s: %s\n", prog, msg)
		return 2
	}
	violations := 0
	violation := func(msg string) {
		fmt.Fprintf(stderr, "%s: %s\n", prog, msg)
		violations++
	}

	git := envGit(env)
	if git("-C", wt, "rev-parse", "--git-dir").Run() != nil {
		return cannot(wt + " is not a git repository")
	}
	head := git("-C", wt, "rev-parse", "--verify", "-q", "HEAD")
	head.Stderr = stderr
	if head.Run() != nil {
		return cannot(wt + " has no resolvable HEAD")
	}
	isCommit := func(sha string) bool { return git("-C", wt, "cat-file", "-e", sha+"^{commit}").Run() == nil }
	isAncestor := func(sha string) bool {
		return git("-C", wt, "merge-base", "--is-ancestor", sha, "HEAD").Run() == nil
	}

	if sub == "targets" {
		for _, sha := range rest {
			if !isCommit(sha) {
				violation("fixup target " + sha + " does not resolve as a commit object — refusing the autosquash")
				continue
			}
			if !isAncestor(sha) {
				violation("fixup target " + sha + " is not an ancestor of HEAD — refusing the autosquash")
			}
		}
	} else {
		base, tasksMD := rest[0], rest[1]
		// `[ -f ]`: a directory passes a readability test but not a
		// regular-file test, so the sweep refuses rather than fail open.
		fi, err := os.Stat(smcAbs(env, tasksMD))
		if err != nil || !fi.Mode().IsRegular() {
			return cannot("tasks file " + tasksMD + " is missing, not a regular file, or unreadable")
		}
		if !isAncestor(base) {
			violation("base " + base + " is no longer an ancestor of HEAD")
		}
		// The bash read the file with `<"$TASKS_MD"` after the -f test
		// alone: a regular file it cannot read printed the shell's error and
		// swept no tokens. Kept as it was.
		body, err := os.ReadFile(smcAbs(env, tasksMD))
		if err != nil {
			fmt.Fprintf(stderr, "%s: %v\n", tasksMD, err)
		}
		for _, token := range gaTokens(body) {
			if !isCommit(token) {
				violation(token + " does not resolve as a commit object")
				continue
			}
			if !isAncestor(token) {
				violation("recorded sha " + token + " is not reachable from HEAD — stale after a rewrite?")
			}
		}
	}

	if violations > 0 {
		fmt.Fprintf(stderr, "%s: %d violation(s)\n", prog, violations)
		return 1
	}
	fmt.Fprintf(stdout, "%s: %s ok\n", prog, sub)
	return 0
}

// gaTokens is `tr -c '0-9a-f' '\n' | awk 'length($0) >= 7' | sort -u`:
// maximal runs of lowercase hex characters, 7 or more long -- the span an
// abbreviated-to-full git sha occupies, with no upper bound, so a sha glued
// into a longer hex run fails resolution and is reported, never silently
// skipped. Lowercase only: git object names are lowercase, so uppercase hex
// words in prose never trip the sweep. The tokens hold only [0-9a-f], which
// every locale's collation orders as bytes do, so sort.Strings keeps
// `sort -u`'s order.
func gaTokens(body []byte) []string {
	runs := strings.FieldsFunc(string(body), func(r rune) bool {
		return (r < '0' || r > '9') && (r < 'a' || r > 'f')
	})
	seen := map[string]bool{}
	var out []string
	for _, r := range runs {
		if len(r) >= 7 && !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	sort.Strings(out)
	return out
}
