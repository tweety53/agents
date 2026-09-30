package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// syncOntoBase is scripts/sync-onto-base.sh: that script's header comment is
// the contract -- a finishing worktree synced onto its base's tip, every
// check-base-moved line echoed on stdout, then one of CLEAN / REBASED (exit
// 0, REBASED preceded by a GUARD-TEST or NO-GUARD-TEST line per overlap
// path) or CONFLICT (exit 1, left mid-rebase); exit 2 on a REFUSE, a
// question it cannot answer, or an aside it cannot restore.
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
		return refuse("usage: sync-onto-base.sh [--resume] <worktree> <base-ref> <merge-base|onto>")
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
		// The aside the conflicted sync recorded: its own stash sha, or
		// none -- never whichever marker stash another worktree left on top.
		rec, err := asideRecord(env, wt, "flow-sync-aside")
		if err != nil {
			return refuse("%v", err)
		}
		b, err := os.ReadFile(rec)
		if err != nil {
			return refuse("no record of the conflicted sync's aside in %s — restore the planning paths by hand from git stash list", wt)
		}
		if busy, err := inProgress(env, wt, "rebase-merge", "rebase-apply", "MERGE_HEAD", "CHERRY_PICK_HEAD"); err != nil {
			return refuse("%v", err)
		} else if busy {
			return refuse("restore refused — a rebase/merge/cherry-pick is still in progress in: %s", wt)
		}
		if sha := strings.TrimSpace(string(b)); sha != "none" {
			if err := restoreAside(env, wt, sha, stderr); err != nil {
				return refuse("%v", err)
			}
		}
		if err := os.Remove(rec); err != nil {
			return refuse("cannot remove %s: %v", rec, err)
		}
		if v, ok = check(base); !ok {
			return 2
		}
	} else {
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			return refuse("FLOW_GUARD_REPO_ROOT is unset — run scripts/sync-onto-base.sh, which sets it")
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
		o, err := rebaseOntoTip(env, wt, v.ref, true, stderr)
		if err != nil {
			return refuse("%v", err)
		}
		switch o.kind {
		case "conflict":
			rec, err := asideRecord(env, wt, "flow-sync-aside")
			if err != nil {
				return refuse("%v", err)
			}
			aside := o.aside
			if aside == "" {
				aside = "none"
			}
			if err := os.WriteFile(rec, []byte(aside+"\n"), 0o644); err != nil {
				return refuse("cannot record the aside in %s: %v — the rebase is left in progress", rec, err)
			}
			fmt.Fprintf(stdout, "CONFLICT: %s — onto %s; unmerged: %s\n", wt, o.sha, strings.Join(o.unmerged, ", "))
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
		for _, p := range paths {
			name := filepath.Base(p)
			test := root + "/scripts/test-" + strings.TrimSuffix(name, filepath.Ext(name)) + ".sh"
			if isFile(test) {
				fmt.Fprintf(stdout, "GUARD-TEST: %s — %s\n", p, test)
			} else {
				fmt.Fprintf(stdout, "NO-GUARD-TEST: %s\n", p)
			}
		}
	}
	fmt.Fprintf(stdout, "REBASED: %s — onto %s\n", wt, onto)
	return 0
}
