package guard

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/guard-autosquash.sh shim (KAN-1058).
//
// flow-guard guard-autosquash — mechanical ancestry guard for the fixup-to-autosquash
// sequence. A fixup against a sha that resolves but is no ancestor of the
// branch — a stale duplicate object, for instance — hands
// `git rebase --autosquash` a spurious merge base and replays unrelated
// history into the branch; a rebase can likewise strand shas a plan file
// still records. This guard asserts the ancestry the sequence silently
// assumes, before and after.
//
// Subcommands:
//   targets <worktree> <sha> [<sha>...]
//       Run before any autosquash: every fixup target must resolve as a
//       commit object and be an ancestor of the worktree's HEAD.
//   after <worktree> <base-ref> <tasks-md>
//       Run after an autosquash or any rebase: <base-ref> must still be an
//       ancestor of HEAD, and every run of 7 or more lowercase hex characters
//       in <tasks-md> must resolve as a commit object AND be reachable from
//       HEAD. Existence alone passes a dangling sha — the stale
//       recorded-sha shape — so the two are separate checks. There is no
//       length ceiling: a sha glued into a longer hex run fails resolution
//       and is reported, never silently skipped. Lowercase only: git object
//       names are lowercase, so uppercase hex words in prose (DEADBEEF and
//       its like) never trip the sweep.
//
// Exit codes:
//   0  every check passed
//   1  one or more violations, one line each on stderr, each naming the sha
//   2  cannot answer — bad arguments, <worktree> not a git repository,
//      HEAD unresolvable, or <tasks-md> missing, not a regular file, or
//      unreadable

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
	// commits is `cat-file -e <sha>^{commit}` for every sha in one
	// --batch-check process: one output line per input line, in order,
	// `<sha> commit <size>` where -e would have passed and `<input>
	// missing` or `<input> ambiguous` where it would have failed. A failed
	// run reads every sha as unresolved, as a failing -e did.
	commits := func(shas []string) []bool {
		ok := make([]bool, len(shas))
		if len(shas) == 0 {
			return ok
		}
		cmd := git("-C", wt, "cat-file", "--batch-check")
		cmd.Stdin = strings.NewReader(strings.Join(shas, "^{commit}\n") + "^{commit}\n")
		out, err := cmd.Output()
		lines := strings.Split(string(out), "\n")
		for i := range shas {
			ok[i] = err == nil && i < len(lines) && strings.Contains(lines[i], " commit ")
		}
		return ok
	}
	isAncestor := func(sha string) bool {
		return git("-C", wt, "merge-base", "--is-ancestor", sha, "HEAD").Run() == nil
	}

	if sub == "targets" {
		resolved := commits(rest)
		for i, sha := range rest {
			if !resolved[i] {
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
		// The bash read the file after that test alone, so a regular file it
		// could not read swept no tokens and could pass; the read is now part
		// of the same refusal, as the header always stated.
		fi, err := os.Stat(smcAbs(env, tasksMD))
		var body []byte
		if err == nil && fi.Mode().IsRegular() {
			body, err = os.ReadFile(smcAbs(env, tasksMD))
		}
		if err != nil || !fi.Mode().IsRegular() {
			return cannot("tasks file " + tasksMD + " is missing, not a regular file, or unreadable")
		}
		if !isAncestor(base) {
			violation("base " + base + " is no longer an ancestor of HEAD")
		}
		tokens := gaTokens(body)
		resolved := commits(tokens)
		for i, token := range tokens {
			if !resolved[i] {
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
