package guard

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

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
