package integrate

import (
	"fmt"
	"io"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/gitops"
)

// flow-guard workspace-remove — tear down one parallel request's worktree,
// `<main>-worktrees/<change>` on branch <change> (a `<name>-r<n>`), once it
// has landed: what integrate-change's cleanup does for a change, for one
// branch.
//
// Usage: flow-guard workspace-remove <main-checkout> <change-or-branch>
//
// The worktree is the one `git -C <main> worktree list --porcelain` lists on
// refs/heads/<change>. Before anything is removed, when that worktree's
// directory exists:
//   - it holds no uncommitted change and no untracked, unignored file;
//   - its HEAD is an ancestor of origin/<base>, <base> the branch's recorded
//     branch.<change>.flowBase, else origin/HEAD's branch.
//
// With the worktree gone and the branch left, the branch's tip faces the
// ancestor check alone.
//
// Then, in order:
//   - every process whose working directory lies inside the worktree is
//     stopped (`STOPPED:` lines, flow-guard stop-worktree-processes);
//   - <main>'s `## workspace isolation` `remove` row runs with
//     <id> = the workspace id of <change> (integrate-change cleanup's
//     removeWorkspace: `WORKSPACE-REMOVED:`, `WORKSPACE-REMOVE-FAILED:`, or a
//     `WORKSPACE: … skipped` line when no row is declared). Only the
//     per-change resource the row names is touched — never the dev
//     workspace's flowd, flow-postgres or default `flow` database;
//   - `git worktree remove` WITHOUT --force, then `worktree prune`;
//   - `git branch -D <change>` when the branch exists;
//   - `git push origin --delete <change>` only when origin carries it.
//
// Prints, one line each:
//
//	REFUSED: <worktree> — <reason>
//	REMOVED: <worktree> | <worktree> — already gone | <change>
//	REMOVE-FAILED: <worktree or change> — <git's message>
//	REMOTE-DELETED: origin/<change> | REMOTE-GONE: origin/<change>
//	REMOTE-REFUSED: origin/<change> — <git's message>
//
// Idempotent: a re-run with every part already gone exits 0.
//
// Exit 0 every part removed or already gone; 1 refused (dirty, not landed —
// nothing removed) or a removal failed; 2 usage, a change name that is not a
// plain change name, or a <main> that is not a git repository.
func init() { core.Registry["workspace-remove"] = workspaceRemove }

func workspaceRemove(args []string, env core.Env, stdout, stderr io.Writer) int {
	refuse := func(id, format string, a ...any) int {
		core.Reason(env, id)
		fmt.Fprintf(stderr, "workspace-remove: "+format+"\n", a...)
		return 2
	}
	if len(args) != 2 {
		return refuse("usage", "usage: flow-guard workspace-remove <main-checkout> <change-or-branch>")
	}
	repo, branch := core.CleanAbs(env, args[0]), args[1]
	if !core.PlainChangeName(branch) {
		return refuse("bad-change-name", "change name %q is not a plain change name", branch)
	}
	git := core.EnvGit(env, "LC_ALL=C")
	if !core.IsDir(repo) || git("-C", repo, "rev-parse", "--git-dir").Run() != nil {
		return refuse("not-a-git-repository", "%s is not a git repository", repo)
	}
	porcelain, ok := core.Capture(git("-C", repo, "worktree", "list", "--porcelain"))
	if !ok {
		return refuse("cannot-list-the-worktrees", "cannot list the worktrees of %s", repo)
	}
	wt := ""
	for _, block := range strings.Split(porcelain, "\n\n") {
		if strings.Contains("\n"+block+"\n", "\nbranch refs/heads/"+branch+"\n") {
			wt = strings.TrimPrefix(strings.SplitN(block, "\n", 2)[0], "worktree ")
		}
	}
	hasBranch := git("-C", repo, "show-ref", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil

	// The checks, all before anything is removed.
	base := gitops.RecordedBase(git, repo, branch)
	if base == "" {
		head, _ := core.Capture(git("-C", repo, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"))
		base = strings.TrimPrefix(head, "origin/")
	}
	refused := func(id, what, format string, a ...any) int {
		core.Reason(env, id)
		fmt.Fprintf(stdout, "REFUSED: %s — %s\n", what, gitops.Flat(fmt.Sprintf(format, a...)))
		return 1
	}
	live := wt != "" && core.IsDir(wt)
	if live {
		if s, ok := core.Capture(git("-C", wt, "status", "--porcelain", "--untracked-files=all")); !ok {
			return refused("status-unreadable", wt, "git status failed")
		} else if s != "" {
			return refused("dirty-worktree", wt, "uncommitted or untracked changes: %s", s)
		}
	}
	if live || hasBranch {
		tip := "refs/heads/" + branch
		if live {
			tip = "HEAD"
		}
		dir := repo
		if live {
			dir = wt
		}
		if base == "" || git("-C", dir, "merge-base", "--is-ancestor", tip, "origin/"+base).Run() != nil {
			return refused("not-landed", branch, "not an ancestor of origin/%s", base)
		}
	}

	failed := false
	if live {
		gitops.StopProcesses(env, []string{wt}, stdout)
	}
	c := &ic{env: env, stdout: stdout, stderr: stderr, prog: "workspace-remove", childEnv: core.ChildEnv(env)}
	if _, f := c.removeWorkspace(repo, gitops.WorkspaceID(branch)); f {
		failed = true
	}
	switch {
	case live:
		if out, err := git("-C", repo, "worktree", "remove", wt).CombinedOutput(); err != nil {
			failed = true
			fmt.Fprintf(stdout, "REMOVE-FAILED: %s — %s\n", wt, gitops.Flat(string(out)))
		} else {
			fmt.Fprintf(stdout, "REMOVED: %s\n", wt)
		}
	case wt != "":
		fmt.Fprintf(stdout, "REMOVED: %s — already gone\n", wt)
	default:
		fmt.Fprintf(stdout, "REMOVED: %s-worktrees/%s — already gone\n", repo, branch)
	}
	_ = git("-C", repo, "worktree", "prune").Run()
	if hasBranch {
		if out, err := git("-C", repo, "branch", "-D", branch).CombinedOutput(); err != nil {
			failed = true
			fmt.Fprintf(stdout, "REMOVE-FAILED: %s — %s\n", branch, gitops.Flat(string(out)))
		} else {
			fmt.Fprintf(stdout, "REMOVED: %s\n", branch)
		}
	}
	// ls-remote --exit-code exits 2 when origin does not carry the branch.
	switch err := git("-C", repo, "ls-remote", "--exit-code", "--heads", "origin", branch).Run(); core.ExitCode(err) {
	case 0:
		if out, err := git("-C", repo, "push", "origin", "--delete", branch).CombinedOutput(); err != nil {
			failed = true
			fmt.Fprintf(stdout, "REMOTE-REFUSED: origin/%s — %s\n", branch, gitops.Flat(string(out)))
		} else {
			fmt.Fprintf(stdout, "REMOTE-DELETED: origin/%s\n", branch)
		}
	case 2:
		fmt.Fprintf(stdout, "REMOTE-GONE: origin/%s\n", branch)
	default:
		failed = true
		fmt.Fprintf(stdout, "REMOTE-REFUSED: origin/%s — cannot read origin\n", branch)
	}
	if failed {
		core.Reason(env, "removal-failed")
		return 1
	}
	return 0
}
