package guard

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// flow-guard sync-base-behind — the whole base sync check-base-behind prints the commands for, run
// by the guard so no raw `git push` is ever typed (kan-1156: the permission classifier denied that
// push and /flow stalled after /clear).
//
// Usage: flow-guard sync-base-behind <worktree>...
//
// Runs check-base-behind over the worktrees; for each BASE-BEHIND block it merges origin/<tracked>
// into origin/<base> with no checkout (merge-tree, commit-tree), pushes that merge fast-forward to
// <base> — never an empty commit, so never a delete — fetches, and runs sync-onto-base <wt>
// origin/<base> <merge-base>. Every line of those guards is echoed on stdout, then one
// `BASE-SYNCED: <wt> — <base> now holds origin/<tracked>` per worktree.
//
// Exit 0 nothing behind, or every behind base synced; 1 a conflict (merge-tree or sync-onto-base's
// CONFLICT, the latter left mid-rebase as its contract says); 2 anything else, the cause on stderr.
//
// Go, in flow-guard (stats/internal/guard/syncbasebehind.go).
func init() { Registry["sync-base-behind"] = syncBaseBehind }

var sbbBlock = regexp.MustCompile(`(?m)^BASE-BEHIND: (.+) — origin/(\S+) lacks [0-9]+ commits of origin/(\S+);`)

func syncBaseBehind(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "sync-base-behind: "
	if len(args) == 0 {
		fmt.Fprintln(stderr, prog+"usage: flow-guard sync-base-behind <worktree>...")
		Reason(env, "usage")
		return 2
	}
	var check bytes.Buffer
	rc := checkBaseBehind(args, env, &check, stderr)
	if rc == 0 {
		io.WriteString(stdout, check.String())
		return 0
	}
	if rc != 1 {
		return rc
	}
	git := envGit(env)
	blocks := sbbBlock.FindAllStringSubmatch(check.String(), -1)
	for _, b := range blocks {
		wt, base, tracked := b[1], b[2], b[3]
		fmt.Fprintf(stdout, "BASE-BEHIND: %s — origin/%s lacks commits of origin/%s\n", wt, base, tracked)
		mb, ok := capture(git("-C", wt, "merge-base", "HEAD", "origin/"+base))
		if !ok {
			fmt.Fprintf(stderr, "%sgit merge-base HEAD origin/%s failed in %s\n", prog, base, wt)
			Reason(env, "merge-base-failed")
			return 2
		}
		tree, ok := capture(git("-C", wt, "merge-tree", "--write-tree", "origin/"+base, "origin/"+tracked))
		if !ok {
			fmt.Fprintf(stdout, "BASE-CONFLICT: %s — origin/%s and origin/%s do not merge cleanly; resolve by hand\n", wt, base, tracked)
			Reason(env, "base-merge-conflict")
			return 1
		}
		merge, ok := capture(git("-C", wt, "commit-tree", tree, "-p", "origin/"+base, "-p", "origin/"+tracked, "-m", "Merge "+tracked+" into "+base))
		if !ok || merge == "" {
			fmt.Fprintf(stderr, "%scommit-tree failed in %s\n", prog, wt)
			Reason(env, "commit-tree-failed")
			return 2
		}
		if out, err := git("-C", wt, "push", "origin", merge+":refs/heads/"+base).CombinedOutput(); err != nil {
			fmt.Fprintf(stderr, "%spush of %s to %s failed: %s\n", prog, merge, base, strings.TrimSpace(string(out)))
			Reason(env, "push-failed")
			return 2
		}
		_ = git("-C", wt, "fetch", "-q", "origin").Run()
		if rc := Registry["sync-onto-base"]([]string{wt, "origin/" + base, mb}, env, stdout, stderr); rc != 0 {
			return rc
		}
		fmt.Fprintf(stdout, "BASE-SYNCED: %s — %s now holds origin/%s\n", wt, base, tracked)
	}
	return 0
}
