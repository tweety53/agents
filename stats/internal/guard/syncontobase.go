package guard

import (
	"fmt"
	"io"
	"path/filepath"
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
//             `REBASED: <worktree> — onto <sha>`, exit 0.
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
// loop.
//
// Exit 0 CLEAN or REBASED; 1 CONFLICT; 2 everything else — a REFUSE verdict,
// a question it cannot answer, a rebase git refused to start (nothing
// changed), or a base still moving after 3 rebases. Exit 2 means stop and
// ask.
//
// The guard is Go: stats/internal/guard/syncontobase.go, over the shared
// baseMoved verdict (basemoved.go) and rebaseOntoTip core (baserebase.go).

// syncOntoBase implements the contract above -- a finishing worktree synced
// onto its base's tip, every
// check-base-moved line echoed on stdout, then one of CLEAN / REBASED (exit
// 0) or CONFLICT (exit 1, left mid-rebase); exit 2 on a REFUSE, a
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
		if v, ok = check(base); !ok {
			return 2
		}
		if v.kind == "CLEAR" {
			fmt.Fprintf(stdout, "CLEAN: %s — nothing to rebase\n", wt)
			return 0
		}
	}

	// Rebase while the base is MOVED, re-checking against each new tip.
	onto := base
	for n := 0; v.kind == "MOVED"; n++ {
		if n == sobMaxRebases {
			return refuse("%s moved again after %d rebases of %s — stop and ask", v.ref, n, wt)
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
