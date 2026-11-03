package guard

import (
	"fmt"
	"io"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/fold-fixup.sh shim (KAN-1058).
//
// flow-guard fold-fixup — fold a fix into the unpushed task commit it targets
// (skills/flow/review-panel-fix-round.md, "Rewrite-based folding is for
// unpushed history only").
//
// Usage: flow-guard fold-fixup <worktree> <task-sha> <tasks-md> <path>...
//        flow-guard fold-fixup --finish <worktree> <task-sha> <tasks-md>
//
// The fold, in order: `git add -- <path>...`; `flow-guard guard-autosquash targets
// <worktree> <task-sha>`; `git commit --fixup=<task-sha> -- <path>...`;
// `GIT_SEQUENCE_EDITOR=: git rebase -i --autosquash <task-sha>^` (git
// ignores a non-interactive --autosquash); the empty-fold drop;
// `flow-guard guard-autosquash after <worktree> <task-sha>^ <tasks-md>`. The rebase's upstream is the task
// commit's own parent, never the base branch re-resolved: replaying only the
// branch's own commits after <task-sha>^ keeps commits that landed on the base
// mid-panel out of the fold. The `after` check's base is that same upstream.
//
// THE EMPTY-FOLD DROP. A fixup whose fold empties its target commit is
// dropped, never kept as a no-op commit: git stops the rebase at that fixup
// ("would make it empty"), and the guard drops the emptied commit with
// `git reset --soft HEAD^` and continues — tip or not, the commits after it
// replay onto its parent.
//
// A CONFLICT IS LEFT IN PROGRESS, never resolved here: it is between two of
// the branch's own commits, resolved by hand keeping both sides, then
// `git rebase --continue`, then `flow-guard fold-fixup --finish` — which settles any
// further empty-fold stop and runs the `after` check.
//
// Prints one line on stdout:
//   FOLDED:   <worktree> — HEAD <sha>
//   DROPPED:  <worktree> — the fold emptied <task-sha>; HEAD <sha>
//   CONFLICT: <worktree> — unmerged: <paths>; resolve keeping both sides, …
//
// Exit codes:
//   0  folded, or the emptied commit dropped
//   1  flow-guard guard-autosquash refused (its lines on stderr) — the round stops
//   2  cannot answer — usage, a git refusal, or a rebase stop that is
//      neither a conflict nor an empty fold
//   3  a conflict is left in progress (CONFLICT line)

// foldFixup is `flow-guard fold-fixup`: the comment above is the contract -- a fix folded into the unpushed task commit it targets
// (skills/flow/review-panel-fix-round.md, "Rewrite-based folding is for
// unpushed history only"), bracketed by guard-autosquash.sh; exit 0 folded
// (FOLDED) or dropped (DROPPED), 1 a guard-autosquash refusal, 2 cannot
// answer, 3 a conflict left in progress (CONFLICT) for a hand resolution and
// `--finish`.
func init() {
	Registry["fold-fixup"] = foldFixup
}

const ffUsage = "usage: flow-guard fold-fixup <worktree> <task-sha> <tasks-md> <path>...\n" +
	"       flow-guard fold-fixup --finish <worktree> <task-sha> <tasks-md>\n"

func foldFixup(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "fold-fixup: "+format+"\n", a...)
		return 2
	}
	finish := len(args) > 0 && args[0] == "--finish"
	if finish {
		args = args[1:]
	}
	if (finish && len(args) != 3) || (!finish && len(args) < 4) {
		fmt.Fprint(stderr, ffUsage)
		return 2
	}
	wt, task, tasksMD := args[0], args[1], args[2]
	for _, a := range args[:3] {
		if a == "" {
			fmt.Fprint(stderr, ffUsage)
			return 2
		}
	}
	// guard-autosquash runs in this same binary (runSibling), or as the
	// executable FLOW_GUARD_AUTOSQUASH names.
	// guard runs one guard-autosquash check, its lines on stderr; the exit
	// code is the caller's: 0 passed, 1 refused, 2 cannot answer.
	guard := func(a ...string) int {
		cmd := siblingCmd(env, "FLOW_GUARD_AUTOSQUASH", "guard-autosquash", a...)
		cmd.Stdout, cmd.Stderr = stderr, stderr
		switch rc := exitCode(cmd.Run()); rc {
		case 0, 1:
			return rc
		default:
			return 2
		}
	}
	git := envGit(env)
	quiet := func(a ...string) bool { // `git ... >&2`
		cmd := git(a...)
		cmd.Stdout, cmd.Stderr = stderr, stderr
		return cmd.Run() == nil
	}
	var parent string
	if !finish {
		paths := args[3:]
		if !quiet(append([]string{"-C", wt, "add", "--"}, paths...)...) {
			return refuse("git add refused the fix's paths in %s", wt)
		}
		if rc := guard("targets", wt, task); rc != 0 {
			return rc
		}
		// The parent is resolved before anything rewrites the task commit.
		var ok bool
		if parent, ok = capture(git("-C", wt, "rev-parse", "--verify", "--end-of-options", task+"^")); !ok {
			return refuse("%s has no parent in %s — cannot fold into a root commit", task, wt)
		}
		if !quiet(append([]string{"-C", wt, "commit", "-q", "--fixup=" + task, "--"}, paths...)...) {
			return refuse("git commit --fixup=%s refused in %s", task, wt)
		}
		// A non-interactive --autosquash is ignored by git; -i with a no-op
		// sequence editor runs the autosquashed todo as written.
		cmd := envGit(env, "GIT_SEQUENCE_EDITOR=:", "GIT_EDITOR=true")("-C", wt, "rebase", "-q", "-i", "--autosquash", parent)
		cmd.Stdout, cmd.Stderr = stderr, stderr
		if cmd.Run() != nil {
			in, err := inProgress(env, wt, "rebase-merge", "rebase-apply")
			if err != nil {
				return refuse("%v", err)
			}
			if !in {
				return refuse("git refused to rebase %s onto %s — the fixup commit stands unfolded", wt, parent)
			}
		}
	} else {
		var ok bool
		if parent, ok = capture(git("-C", wt, "rev-parse", "--verify", "--end-of-options", task+"^")); !ok {
			return refuse("%s has no parent in %s", task, wt)
		}
	}

	dropped, unmerged, err := ffSettle(env, wt, stderr)
	if err != nil {
		return refuse("%v", err)
	}
	if unmerged != nil {
		fmt.Fprintf(stdout, "CONFLICT: %s — unmerged: %s; resolve keeping both sides, `git rebase --continue`, then flow-guard fold-fixup --finish %s %s %s\n",
			wt, strings.Join(unmerged, ", "), wt, task, tasksMD)
		return 3
	}
	if rc := guard("after", wt, parent, tasksMD); rc != 0 {
		return rc
	}
	head, ok := capture(git("-C", wt, "rev-parse", "HEAD"))
	if !ok {
		return refuse("HEAD does not resolve in %s", wt)
	}
	if dropped {
		fmt.Fprintf(stdout, "DROPPED: %s — the fold emptied %s; HEAD %s\n", wt, task, head)
	} else {
		fmt.Fprintf(stdout, "FOLDED: %s — HEAD %s\n", wt, head)
	}
	return 0
}

// ffSettle drives an autosquash rebase in wt to its end. At a stop with
// unmerged paths it returns them, the rebase left in progress. At a stop
// whose index equals HEAD^'s tree -- git refusing a fixup that would leave
// the task commit empty -- it drops that commit (`git reset --soft HEAD^`)
// and continues: a fold that empties its target is dropped, never kept.
// Any other stop is an error.
func ffSettle(env Env, wt string, stderr io.Writer) (dropped bool, unmerged []string, err error) {
	git := envGit(env, "GIT_EDITOR=true")
	for {
		in, err := inProgress(env, wt, "rebase-merge", "rebase-apply")
		if err != nil || !in {
			return dropped, nil, err
		}
		paths, ok := capture(git("-C", wt, "diff", "--no-renames", "--name-only", "--diff-filter=U"))
		if !ok {
			return dropped, nil, fmt.Errorf("cannot list the unmerged paths in %s", wt)
		}
		if paths != "" {
			return dropped, strings.Split(paths, "\n"), nil
		}
		if git("-C", wt, "diff", "--cached", "--quiet", "HEAD^").Run() != nil {
			return dropped, nil, fmt.Errorf("the rebase in %s stopped for a reason fold-fixup cannot read — left in progress", wt)
		}
		for _, a := range [][]string{{"reset", "-q", "--soft", "HEAD^"}, {"rebase", "--continue"}} {
			cmd := git(append([]string{"-C", wt}, a...)...)
			cmd.Stdout, cmd.Stderr = stderr, stderr
			if cmd.Run() != nil && a[0] == "reset" {
				return dropped, nil, fmt.Errorf("git reset --soft HEAD^ refused in %s", wt)
			}
		}
		dropped = true
	}
}
