package guard

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/sync-onto-base.sh shim (KAN-1058).
//
// flow-guard sync-onto-base — sync a finishing worktree onto its base's tip, in a
// guard instead of the hand-typed recipe skills/flow/sync-onto-base.md
// carried (KAN-860).
//
// Usage:
//
//   flow-guard sync-onto-base <worktree> <base-ref> <merge-base>
//   flow-guard sync-onto-base --resume <worktree> <base-ref> <onto>
//
// <base-ref> is the base branch, bare (main) or remote-tracking
// (origin/main), resolved exactly as flow-guard check-base-moved resolves it; the
// caller has already fetched (flow-guard resolve-base-branch), and this guard fetches
// nothing. Every flow-guard check-base-moved line it reaches is echoed on stdout.
//
// The first form runs flow-guard check-base-moved against <merge-base>:
//   CLEAR  -> `CLEAN: <worktree> — nothing to rebase`, exit 0.
//   MOVED  -> the worktree is rebased onto the base's tip — resolved to a
//             sha first, so a fetch landing mid-run cannot move it.
//             flow-guard check-base-moved then re-runs against
//             that sha; while it answers MOVED (the base moved during the
//             rebase) the rebase runs again, at most 3 rebases in all. Then
//             one line per path any MOVED verdict named under `overlaps:`,
//             sorted:
//               GUARD-TEST: <path> — <agents repo>/scripts/test-<stem>.sh
//               NO-GUARD-TEST: <path>
//             (<stem> is the path's basename without its extension — the
//             guard test flow-guard run-guard-tests discovers by glob), and
//             `REBASED: <worktree> — onto <sha>`, exit 0. The caller runs each
//             named test; this guard runs none.
//   A rebase that stops on a conflict prints
//             `CONFLICT: <worktree> — onto <sha>; unmerged: <paths>` and exits
//             1, the worktree left mid-rebase exactly as git left it —
//             resolve in place, then --resume. Each
//             unmerged path that is a PNG under a `*-snapshots/` directory, or
//             `full-app-suite.zip`, follows as `RECAPTURE: <path>`, in the
//             CONFLICT line's order — resolved by a targeted recapture
//             (skills/flow/sync-onto-base.md), never a stop.
//
// --resume runs once an in-place resolution has finished the rebase: it
// refuses while a rebase is still in progress and unless <onto> — the sha the
// CONFLICT line named — is an ancestor of HEAD, and runs the same re-check
// loop. It prints no GUARD-TEST lines: a resolved rebase
// runs the project's whole lint and test lists instead.
//
// Exit 0 CLEAN or REBASED; 1 CONFLICT; 2 everything else — a REFUSE verdict,
// a question it cannot answer, a rebase git refused to start (nothing
// changed), or a base still moving after 3 rebases. Exit 2 means stop and
// ask.
//
// The guard is Go: stats/internal/guard/syncontobase.go, over the shared
// baseMoved verdict (basemoved.go) and rebaseOntoTip core (baserebase.go).
// The binary does not live in this checkout, so this shim exports
// FLOW_GUARD_REPO_ROOT — the agents repository, derived through lib/'s
// physical path so a skill's symlink to this shim still answers the
// repository and never the skill directory — for the GUARD-TEST lookup.

// syncOntoBase is scripts/sync-onto-base.sh: that script's header comment is
// the contract -- a finishing worktree synced onto its base's tip, every
// check-base-moved line echoed on stdout, then one of CLEAN / REBASED (exit
// 0, REBASED preceded by a NO-GUARD-TEST line per overlap
// path) or CONFLICT (exit 1, left mid-rebase); exit 2 on a REFUSE, a
// question it cannot answer.
func init() {
	Registry["sync-onto-base"] = syncOntoBase
}

const (
	sobPrefix = "sync-onto-base: "
	// sobMaxRebases caps the re-check loop: a base still moving after this
	// many rebases is a question for the operator, not another rebase.
	sobMaxRebases = 3
)

func syncOntoBase(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, sobPrefix+format+"\n", a...)
		return 2
	}
	resume := len(args) > 0 && args[0] == "--resume"
	if resume {
		args = args[1:]
	}
	if len(args) != 3 || args[0] == "" || args[1] == "" || args[2] == "" {
		return refuse("usage: flow-guard sync-onto-base [--resume] <worktree> <base-ref> <merge-base|onto>")
	}
	wt, baseRef, base := args[0], args[1], args[2]

	// check echoes one check-base-moved answer; ok is false when the caller
	// must exit 2 (a REFUSE, or the tree could not be read).
	check := func(recorded string) (v bmVerdict, ok bool) {
		v = baseMoved(env, wt, baseRef, recorded, stderr)
		if v.line != "" {
			fmt.Fprintln(stdout, v.line)
		}
		return v, v.code == 0 && v.kind != "REFUSE"
	}

	var v bmVerdict
	var ok bool
	root := ""
	if resume {
		if envGit(env)("-C", wt, "merge-base", "--is-ancestor", "--end-of-options", base, "HEAD").Run() != nil {
			return refuse("%s is not an ancestor of HEAD in %s — the rebase it resumes did not finish onto it", base, wt)
		}
		if busy, err := inProgress(env, wt, "rebase-merge", "rebase-apply", "MERGE_HEAD", "CHERRY_PICK_HEAD"); err != nil {
			return refuse("%v", err)
		} else if busy {
			return refuse("resume refused — a rebase/merge/cherry-pick is still in progress in: %s", wt)
		}
		if v, ok = check(base); !ok {
			return 2
		}
	} else {
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			return refuse("FLOW_GUARD_REPO_ROOT is unset — run flow-guard sync-onto-base, which sets it")
		}
		if v, ok = check(base); !ok {
			return 2
		}
		if v.kind == "CLEAR" {
			fmt.Fprintf(stdout, "CLEAN: %s — nothing to rebase\n", wt)
			return 0
		}
	}

	// Rebase while the base is MOVED, re-checking against each new tip.
	overlap := map[string]bool{}
	onto := base
	for n := 0; v.kind == "MOVED"; n++ {
		if n == sobMaxRebases {
			return refuse("%s moved again after %d rebases of %s — stop and ask", v.ref, n, wt)
		}
		for _, p := range v.overlap {
			overlap[p] = true
		}
		o, err := rebaseOntoTip(env, wt, v.ref, stderr)
		if err != nil {
			return refuse("%v", err)
		}
		switch o.kind {
		case "conflict":
			fmt.Fprintf(stdout, "CONFLICT: %s — onto %s; unmerged: %s\n", wt, o.sha, strings.Join(o.unmerged, ", "))
			for _, u := range o.unmerged {
				if sobRecapture(u) {
					fmt.Fprintf(stdout, "RECAPTURE: %s\n", u)
				}
			}
			return 1
		case "refused":
			return refuse("git refused to rebase %s onto %s — nothing changed", wt, o.sha)
		}
		onto = o.sha
		if v, ok = check(onto); !ok {
			return 2
		}
	}

	// A resumed sync prints no guard-test lines: the resolution it follows
	// runs the whole lint and test lists instead.
	if !resume {
		paths := make([]string, 0, len(overlap))
		for p := range overlap {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		// The per-guard bash harnesses (scripts/test-<stem>.sh) this looked
		// up are gone — every guard's test is Go, run by the guard-tests
		// suite — so no overlap path has a scoped test to name.
		for _, p := range paths {
			fmt.Fprintf(stdout, "NO-GUARD-TEST: %s\n", p)
		}
	}
	fmt.Fprintf(stdout, "REBASED: %s — onto %s\n", wt, onto)
	return 0
}

// sobRecapture reports whether an unmerged path is a captured screenshot
// baseline -- a PNG under a Playwright `*-snapshots/` directory -- or the full
// app suite's archive: a conflict the caller resolves by a targeted recapture,
// never by merging bytes or stopping to ask.
func sobRecapture(p string) bool {
	if filepath.Base(p) == "full-app-suite.zip" {
		return true
	}
	if !strings.EqualFold(filepath.Ext(p), ".png") {
		return false
	}
	for _, dir := range strings.Split(filepath.Dir(p), "/") {
		if strings.HasSuffix(dir, "-snapshots") {
			return true
		}
	}
	return false
}
