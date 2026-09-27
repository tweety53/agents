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
	if !isDir(smcAbs(env, worktree)) {
		fmt.Fprintf(stderr, "check-base-moved: %s is not a directory — cannot determine anything\n", worktree)
		return 2
	}
	// The bash exported LC_ALL=C for its whole run; git inherits it here too.
	git := envGit(env, "LC_ALL=C")
	if git("-C", worktree, "rev-parse", "--git-dir").Run() != nil {
		fmt.Fprintf(stderr, "check-base-moved: %s is not a git worktree — cannot determine anything\n", worktree)
		return 2
	}

	// No recorded merge base: an honest unknown, never an inferred verdict.
	if recorded == "-" {
		fmt.Fprintf(stdout, "REFUSE: no merge base recorded for %s — cannot tell whether the base has moved\n", worktree)
		return 0
	}

	// Every ref this guard did not choose itself is passed after
	// --end-of-options, so a value beginning with `-` is read as a ref and
	// rejected rather than parsed as a git option.
	recordedSHA, ok := capture(git("-C", worktree, "rev-parse", "--verify", "--end-of-options", recorded+"^{commit}"))
	if !ok {
		fmt.Fprintf(stdout, "REFUSE: recorded merge base '%s' does not resolve in %s\n", recorded, worktree)
		return 0
	}

	ref := resolveRemoteBase(git, worktree, baseRef)
	if git("-C", worktree, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}").Run() != nil {
		fmt.Fprintf(stdout, "REFUSE: base ref '%s' does not resolve in %s — cannot tell whether the base has moved\n", ref, worktree)
		return 0
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
		return 2
	}
	if count == "0" {
		fmt.Fprintf(stdout, "CLEAR: %s — %s has not moved since the recorded merge base\n", worktree, ref)
		return 0
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
		return 2
	}
	if uncarried == "0" {
		fmt.Fprintf(stdout, "CLEAR: %s — the %s commits %s gained since the recorded merge base are all already carried by this branch — nothing to rebase\n", worktree, count, ref)
		return 0
	}

	moved, ok := capture(git("-C", worktree, "diff", "--no-renames", "--name-only", "--end-of-options", rng))
	if !ok {
		fmt.Fprintf(stderr, "check-base-moved: cannot list paths changed on %s in %s\n", ref, worktree)
		return 2
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
		return 2
	}
	staged, ok := capture(git("-C", worktree, "diff", "--no-renames", "--name-only", "--cached"))
	if !ok {
		fmt.Fprintf(stderr, "check-base-moved: cannot list this change's staged paths in %s\n", worktree)
		return 2
	}
	unstaged, ok := capture(git("-C", worktree, "diff", "--no-renames", "--name-only"))
	if !ok {
		fmt.Fprintf(stderr, "check-base-moved: cannot list this change's unstaged paths in %s\n", worktree)
		return 2
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
		fmt.Fprintf(stdout, "MOVED: %s — %s commits on %s since the recorded merge base; no overlap with this change's paths\n", worktree, count, ref)
		return 0
	}
	shown := strings.Join(overlap, ", ")
	if len(overlap) > 10 {
		shown = fmt.Sprintf("%s (+%d more)", strings.Join(overlap[:10], ", "), len(overlap)-10)
	}
	fmt.Fprintf(stdout, "MOVED: %s — %s commits on %s since the recorded merge base; overlaps: %s\n", worktree, count, ref, shown)
	return 0
}
