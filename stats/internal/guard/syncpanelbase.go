package guard

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// syncPanelBase is scripts/sync-panel-base.sh: that script's header comment
// is the contract -- resolve-base-branch's fetch and a `BASE:` line, every
// check-base-moved line echoed, a bare rebase (no aside) on MOVED without
// overlap or under --rebase with a `REBASED: … merge base <sha>` line after
// each, then exit 0 settled, 1 CONFLICT, 3 an overlap waiting on the
// operator's prompt, 2 anything it cannot answer.
func init() {
	Registry["sync-panel-base"] = syncPanelBase
}

func syncPanelBase(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "sync-panel-base: "+format+"\n", a...)
		return 2
	}
	rebase := len(args) > 0 && args[0] == "--rebase"
	if rebase {
		args = args[1:]
	}
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return refuse("usage: sync-panel-base.sh [--rebase] <worktree> <working-notes-merge-base>")
	}
	wt, mb := args[0], args[1]

	var name bytes.Buffer
	if rc := resolveBaseBranch([]string{wt}, env, &name, stderr); rc != 0 {
		return refuse("could not resolve the base branch of %s (resolve-base-branch exit %d)", wt, rc)
	}
	base := strings.TrimSpace(name.String())
	fmt.Fprintf(stdout, "BASE: %s\n", base)
	ref := "origin/" + base

	check := func(recorded string) (v bmVerdict, ok bool) {
		v = baseMoved(env, wt, ref, recorded, stderr)
		if v.line != "" {
			fmt.Fprintln(stdout, v.line)
		}
		return v, v.code == 0 && v.kind != "REFUSE"
	}
	v, ok := check(mb)
	if !ok {
		return 2
	}
	// Rebase while MOVED: the first time on no overlap or --rebase, after
	// that only on no overlap -- a fresh overlap re-offers the prompt.
	for n := 0; v.kind == "MOVED"; n++ {
		if len(v.overlap) > 0 && (n > 0 || !rebase) {
			return 3
		}
		if n == sobMaxRebases {
			return refuse("%s moved again after %d rebases of %s — stop and ask", ref, n, wt)
		}
		o, err := rebaseOntoTip(env, wt, ref, false, stderr)
		if err != nil {
			return refuse("%v", err)
		}
		switch o.kind {
		case "conflict":
			fmt.Fprintf(stdout, "CONFLICT: %s — onto %s; unmerged: %s\n", wt, o.sha, strings.Join(o.unmerged, ", "))
			return 1
		case "refused":
			return refuse("git refused to rebase %s onto %s — nothing changed", wt, o.sha)
		}
		fmt.Fprintf(stdout, "REBASED: %s — merge base %s\n", wt, o.sha)
		if v, ok = check(o.sha); !ok {
			return 2
		}
	}
	return 0
}
