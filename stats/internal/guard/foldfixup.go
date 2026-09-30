package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// foldFixup is scripts/fold-fixup.sh: that script's header comment is the
// contract -- a fix folded into the unpushed task commit it targets
// (skills/flow/review-panel-fix-round.md, "Rewrite-based folding is for
// unpushed history only"), bracketed by guard-autosquash.sh; exit 0 folded
// (FOLDED) or dropped (DROPPED), 1 a guard-autosquash refusal, 2 cannot
// answer, 3 a conflict left in progress (CONFLICT) for a hand resolution and
// `--finish`.
func init() {
	Registry["fold-fixup"] = foldFixup
}

const ffUsage = "usage: fold-fixup.sh <worktree> <task-sha> <tasks-md> <path>...\n" +
	"       fold-fixup.sh --finish <worktree> <task-sha> <tasks-md>\n"

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
	// guard-autosquash stays bash, exec'd from beside the shim, which
	// exports its $SCRIPT_DIR/guard-autosquash.sh spelling.
	autosquash := env.Getenv("FLOW_GUARD_AUTOSQUASH")
	if autosquash == "" {
		return refuse("FLOW_GUARD_AUTOSQUASH is unset — run scripts/fold-fixup.sh, which sets it")
	}
	// guard runs one guard-autosquash check, its lines on stderr; the exit
	// code is the caller's: 0 passed, 1 refused, 2 cannot answer.
	guard := func(a ...string) int {
		cmd := exec.Command(autosquash, a...)
		cmd.Dir = env.Dir
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
	var parent, aside string
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
		var out bytes.Buffer
		rc := asidePlanningArtifacts([]string{"aside", wt}, env, &out, stderr)
		fmt.Fprint(stderr, out.String())
		if rc != 0 {
			return refuse("the planning paths could not be set aside in %s — the fixup commit stands unfolded", wt)
		}
		// PLANNING-ARTIFACTS-ASIDE: <worktree> — <stash sha>; CLEAN set none.
		if line := strings.TrimSpace(out.String()); strings.HasPrefix(line, "PLANNING-ARTIFACTS-ASIDE:") {
			aside = line[strings.LastIndex(line, " ")+1:]
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
				if err := ffRestore(env, wt, aside, stderr); err != nil {
					fmt.Fprintf(stderr, "fold-fixup: %v\n", err)
				}
				return refuse("git refused to rebase %s onto %s — the fixup commit stands unfolded", wt, parent)
			}
		}
	} else {
		var ok bool
		if parent, ok = capture(git("-C", wt, "rev-parse", "--verify", "--end-of-options", task+"^")); !ok {
			return refuse("%s has no parent in %s", task, wt)
		}
		// The aside the conflicted fold recorded: its own stash sha, or none.
		rec, err := asideRecord(env, wt, "flow-fold-aside")
		if err != nil {
			return refuse("%v", err)
		}
		b, err := os.ReadFile(rec)
		if err != nil {
			return refuse("no record of the conflicted fold's aside in %s — restore the planning paths by hand from git stash list", wt)
		}
		if aside = strings.TrimSpace(string(b)); aside == "none" {
			aside = ""
		}
	}
	rec, err := asideRecord(env, wt, "flow-fold-aside")
	if err != nil {
		return refuse("%v", err)
	}

	dropped, unmerged, err := ffSettle(env, wt, stderr)
	if err != nil {
		return refuse("%v", err)
	}
	if unmerged != nil {
		kept := aside
		if kept == "" {
			kept = "none"
		}
		if err := os.WriteFile(rec, []byte(kept+"\n"), 0o644); err != nil {
			return refuse("cannot record the aside in %s: %v — the rebase is left in progress", rec, err)
		}
		fmt.Fprintf(stdout, "CONFLICT: %s — unmerged: %s; resolve keeping both sides, `git rebase --continue`, then fold-fixup.sh --finish %s %s %s\n",
			wt, strings.Join(unmerged, ", "), wt, task, tasksMD)
		return 3
	}
	if err := ffRestore(env, wt, aside, stderr); err != nil {
		return refuse("%v", err)
	}
	if err := os.Remove(rec); err != nil && !os.IsNotExist(err) {
		return refuse("cannot remove %s: %v", rec, err)
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

// ffRestore restores the aside this fold set aside (none when aside is ""),
// and only that stash: the stash list is shared by every worktree.
func ffRestore(env Env, wt, aside string, stderr io.Writer) error {
	if aside == "" {
		return nil
	}
	return restoreAside(env, wt, aside, stderr)
}
