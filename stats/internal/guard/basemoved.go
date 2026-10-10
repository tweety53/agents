package guard

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// checkBaseMoved is scripts/check-base-moved.sh: that script's header comment
// is the contract -- one CLEAR/MOVED/REFUSE verdict line on stdout, exit 0
// whenever a verdict was reached, exit 2 when the tree cannot be read. The
// reasoning for each step, moved here from the bash body it replaced
// (d71a2327), sits beside the code it explains.
func init() {
	Registry["check-base-moved"] = checkBaseMoved
}

func checkBaseMoved(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktree, baseRef, recorded := arg(0), arg(1), arg(2)
	if worktree == "" || baseRef == "" || recorded == "" {
		fmt.Fprint(stderr, baseRefUsage("check-base-moved.sh"))
		return 2
	}
	v := baseMoved(env, worktree, baseRef, recorded, stderr)
	if v.line != "" {
		fmt.Fprintln(stdout, v.line)
	}
	return v.code
}

// bmVerdict is one check-base-moved answer. code is the guard's exit code:
// 0 a verdict was reached, 2 the tree cannot be read (the cause already on
// stderr, line and kind empty). line is the verdict line the guard prints,
// without its newline; kind is its CLEAR/MOVED/REFUSE token; ref is the base
// ref the answer was about, once resolved; overlap is the full sorted overlap
// a MOVED line names -- the line itself cuts it at 10 paths.
type bmVerdict struct {
	code       int
	line, kind string
	ref        string
	overlap    []string
}

// baseMoved is check-base-moved's whole answer for an in-process caller.
func baseMoved(env Env, worktree, baseRef, recorded string, stderr io.Writer) bmVerdict {
	cannot := bmVerdict{code: 2}
	verdict := func(kind, ref string, overlap []string, format string, a ...any) bmVerdict {
		return bmVerdict{line: kind + ": " + fmt.Sprintf(format, a...), kind: kind, ref: ref, overlap: overlap}
	}
	if !isDir(smcAbs(env, worktree)) {
		fmt.Fprintf(stderr, "check-base-moved: %s is not a directory — cannot determine anything\n", worktree)
		return cannot
	}
	// The bash exported LC_ALL=C for its whole run; git inherits it here too.
	git := envGit(env, "LC_ALL=C")
	if git("-C", worktree, "rev-parse", "--git-dir").Run() != nil {
		fmt.Fprintf(stderr, "check-base-moved: %s is not a git worktree — cannot determine anything\n", worktree)
		return cannot
	}

	// No recorded merge base: an honest unknown, never an inferred verdict.
	if recorded == "-" {
		return verdict("REFUSE", "", nil, "no merge base recorded for %s — cannot tell whether the base has moved", worktree)
	}

	// Every ref this guard did not choose itself is passed after
	// --end-of-options, so a value beginning with `-` is read as a ref and
	// rejected rather than parsed as a git option.
	recordedSHA, ok := capture(git("-C", worktree, "rev-parse", "--verify", "--end-of-options", recorded+"^{commit}"))
	if !ok {
		return verdict("REFUSE", "", nil, "recorded merge base '%s' does not resolve in %s", recorded, worktree)
	}

	ref := resolveRemoteBase(git, worktree, baseRef)
	if git("-C", worktree, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}").Run() != nil {
		return verdict("REFUSE", ref, nil, "base ref '%s' does not resolve in %s — cannot tell whether the base has moved", ref, worktree)
	}

	// Every git invocation whose failure would otherwise be read as an
	// answer is captured and checked on its own, never folded into a
	// sort/intersect — the reasoning check-finish-preflight's signal (d)
	// comment records. A failing invocation is exit 2 with a named message,
	// never a CLEAR. The argv shapes are the bash's exactly: a test stub
	// picks each call out by position.
	rng := recordedSHA + ".." + ref
	count, ok := capture(git("-C", worktree, "rev-list", "--count", "--end-of-options", rng))
	if !ok {
		fmt.Fprintf(stderr, "check-base-moved: cannot count commits between %s and %s in %s\n", recordedSHA, ref, worktree)
		return cannot
	}
	if count == "0" {
		return verdict("CLEAR", ref, nil, "%s — %s has not moved since the recorded merge base", worktree, ref)
	}

	// Movement entirely satisfied by commits this branch already carries
	// (KAN-535) — the same commit objects landed upstream by another route
	// while the change was in flight — needs no rebase: rebasing over it
	// replays nothing, so it reads as CLEAR even though the base moved.
	// `^HEAD` rides after --end-of-options as rev syntax (`^rev` is a rev,
	// not an option).
	uncarried, ok := capture(git("-C", worktree, "rev-list", "--count", "--end-of-options", rng, "^HEAD"))
	if !ok {
		fmt.Fprintf(stderr, "check-base-moved: cannot count uncarried commits between %s and %s in %s\n", recordedSHA, ref, worktree)
		return cannot
	}
	if uncarried == "0" {
		// The recorded merge base is stale here -- most often a rebase an
		// earlier, stopped run made (KAN-924) -- so the line names the
		// branch's actual fork point, which the run carries forward as this
		// worktree's <rebased-merge-base> (skills/flow/sync-onto-base.md). A
		// reshape from the stale one would fold base commits into the change.
		fork, ok := capture(git("-C", worktree, "merge-base", "--end-of-options", "HEAD", ref))
		if !ok {
			fmt.Fprintf(stderr, "check-base-moved: cannot find the merge base of HEAD and %s in %s\n", ref, worktree)
			return cannot
		}
		return verdict("CLEAR", ref, nil, "%s — the %s commits %s gained since the recorded merge base are all already carried by this branch — nothing to rebase; merge base now %s", worktree, count, ref, fork)
	}

	moved, ok := capture(git("-C", worktree, "diff", "--no-renames", "--name-only", "--end-of-options", rng))
	if !ok {
		fmt.Fprintf(stderr, "check-base-moved: cannot list paths changed on %s in %s\n", ref, worktree)
		return cannot
	}

	// The change's own paths (design.md: touched-paths-include-index-and-worktree):
	// the union of what HEAD carries since the recorded merge base, what is
	// staged, and what is unstaged. Run 1 — the only run this guard serves —
	// is reached with work staged and uncommitted by design, so committed
	// history alone would report "no overlap" for a change that conflicts.
	// Three captures, each checked before any is used, so a partial failure
	// can never read as an empty — and therefore clean — set.
	committed, ok := capture(git("-C", worktree, "diff", "--no-renames", "--name-only", "--end-of-options", recordedSHA+"..HEAD"))
	if !ok {
		fmt.Fprintf(stderr, "check-base-moved: cannot list this change's committed paths in %s\n", worktree)
		return cannot
	}
	staged, ok := capture(git("-C", worktree, "diff", "--no-renames", "--name-only", "--cached"))
	if !ok {
		fmt.Fprintf(stderr, "check-base-moved: cannot list this change's staged paths in %s\n", worktree)
		return cannot
	}
	unstaged, ok := capture(git("-C", worktree, "diff", "--no-renames", "--name-only"))
	if !ok {
		fmt.Fprintf(stderr, "check-base-moved: cannot list this change's unstaged paths in %s\n", worktree)
		return cannot
	}

	// `sort -u` and `comm -12` under the bash's LC_ALL=C: byte order, which
	// is what sort.Strings gives, so the listed paths keep its order.
	change := map[string]bool{}
	for _, p := range strings.Split(committed+"\n"+staged+"\n"+unstaged, "\n") {
		change[p] = true
	}
	seen := map[string]bool{}
	var overlap []string
	for _, p := range strings.Split(moved, "\n") {
		if p != "" && change[p] && !seen[p] {
			seen[p] = true
			overlap = append(overlap, p)
		}
	}
	sort.Strings(overlap)

	if len(overlap) == 0 {
		return verdict("MOVED", ref, nil, "%s — %s commits on %s since the recorded merge base; no overlap with this change's paths", worktree, count, ref)
	}
	shown := strings.Join(overlap, ", ")
	if len(overlap) > 10 {
		shown = fmt.Sprintf("%s (+%d more)", strings.Join(overlap[:10], ", "), len(overlap)-10)
	}
	return verdict("MOVED", ref, overlap, "%s — %s commits on %s since the recorded merge base; overlaps: %s", worktree, count, ref, shown)
}
