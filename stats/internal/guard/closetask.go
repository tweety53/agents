package guard

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
)

// closeTask is scripts/close-task.sh: that script's header is the contract.
// It runs the task-close sequence of skills/flow/implement.md (**The next
// implementer overlaps the guard.**, step 2), composing the guards
// in-process through Registry.
func init() { Registry["close-task"] = closeTask }

func closeTask(args []string, env Env, stdout, stderr io.Writer) int {
	usage := func(msg string) int {
		fmt.Fprintf(stderr, "close-task: %s\nusage: close-task.sh [-end-key <key> -session-token <token> [-end-commit <sha>]] [-undeclared <task-id>=<path>[,<path>…]]… <canonical-worktree> <name> <merge-base> <task-id>:<sha|worktree=sha[,worktree=sha…]>…\n", msg)
		return 2
	}
	fset := flag.NewFlagSet("close-task", flag.ContinueOnError)
	fset.SetOutput(stderr)
	endKey := fset.String("end-key", "", "the implementer dispatch's key, to close its record")
	token := fset.String("session-token", "", "the run's literal session token, for the dispatch end")
	endCommit := fset.String("end-commit", "", "the commit the implementer produced, for the dispatch end")
	undeclared := map[string][]string{}
	fset.Func("undeclared", "<task-id>=<path>[,<path>…]: paths the fields guard's refusal named for that task", func(v string) error {
		id, paths, ok := strings.Cut(v, "=")
		if !ok || id == "" || paths == "" {
			return fmt.Errorf("not <task-id>=<path>[,<path>…]: %q", v)
		}
		undeclared[id] = append(undeclared[id], strings.Split(paths, ",")...)
		return nil
	})
	if fset.Parse(args) != nil {
		return 2
	}
	pos := fset.Args()
	if len(pos) < 4 {
		return usage(fmt.Sprintf("got %d positional argument(s)", len(pos)))
	}
	canonical, name, mergeBase := pos[0], pos[1], pos[2]
	type task struct{ id, commit string }
	var tasks []task
	var worktrees []string // distinct, in first-named order: one push each
	for _, a := range pos[3:] {
		id, commit, ok := strings.Cut(a, ":")
		if !ok || id == "" || commit == "" {
			return usage("task argument is not <task-id>:<sha|map>: " + a)
		}
		tasks = append(tasks, task{id, commit})
		wts := []string{canonical}
		if strings.Contains(commit, "=") {
			wts = nil
			for _, pair := range strings.Split(commit, ",") {
				wt, _, _ := strings.Cut(strings.TrimSpace(pair), "=")
				wts = append(wts, wt)
			}
		}
		for _, wt := range wts {
			if !slices.Contains(worktrees, wt) {
				worktrees = append(worktrees, wt)
			}
		}
	}

	// A record never blocks: a failed dispatch end is reported, never a stop.
	if *endKey != "" {
		rec := []string{"record", "dispatch", "end", "-change", name, "-key", *endKey, "-session-token", *token}
		if *endCommit != "" {
			rec = append(rec, "-commit", *endCommit)
		}
		if ctFlow(env, stdout, stderr, append(rec, "-outcome", "completed")...) != 0 {
			fmt.Fprintln(stderr, "close-task: the dispatch end did not reach the store — continuing; a record never blocks")
		}
	}

	// Both commit guards run before anything is ticked or pushed, so a
	// refused commit is never pushed (design.md: close-task-push-after-guards).
	worst := 0
	for _, t := range tasks {
		worst = max(worst, Registry["check-task-commit-fields"]([]string{canonical, t.id, t.commit, "", canonical, name}, env, stdout, stderr))
	}
	worst = max(worst, Registry["check-task-commit-planning-paths"]([]string{canonical, mergeBase}, env, stdout, stderr))
	switch worst {
	case 0:
	case 1:
		fmt.Fprintln(stderr, "close-task: RE-COMMIT — a commit guard refused; nothing was ticked or pushed")
		return 1
	default:
		fmt.Fprintln(stderr, "close-task: STOP — a commit guard could not judge; nothing was ticked or pushed")
		return 2
	}

	var quiet []string
	for _, t := range tasks {
		var verdict bytes.Buffer
		rc := Registry["check-review-gate"](append([]string{canonical, t.id, t.commit, canonical, name}, undeclared[t.id]...), env, io.MultiWriter(stdout, &verdict), stderr)
		if rc != 0 {
			fmt.Fprintln(stderr, "close-task: STOP — the review gate could not judge; nothing was ticked or pushed")
			return 2
		}
		if strings.HasPrefix(verdict.String(), "QUIET:") {
			quiet = append(quiet, t.id)
		}
	}

	rc := 0
	for _, id := range quiet {
		if ctFlow(env, stdout, stderr, "tasks", "tick", "-C", canonical, name, id) != 0 {
			fmt.Fprintf(stderr, "close-task: STOP — flow tasks tick failed for task %s\n", id)
			rc = 2
		}
	}
	git := envGit(env)
	for _, wt := range worktrees {
		cmd := git("-C", wt, "push", "origin", "spectre/"+name)
		cmd.Stdout, cmd.Stderr = stdout, stderr
		if cmd.Run() != nil {
			fmt.Fprintf(stderr, "close-task: STOP — git push failed in %s\n", wt)
			rc = 2
		}
	}
	return rc
}

// ctFlow runs the flow CLI from env's PATH, its output relayed; non-zero
// when it is missing or fails.
func ctFlow(env Env, stdout, stderr io.Writer, args ...string) int {
	flow, ok := lookPath(env, "flow")
	if !ok {
		fmt.Fprintln(stderr, "close-task: no flow binary on PATH")
		return 127
	}
	cmd := exec.Command(flow, args...)
	cmd.Dir = env.Dir
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if cmd.Run() != nil {
		return 1
	}
	return 0
}
