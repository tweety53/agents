package guard

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// closeTask is scripts/close-task.sh: that script's header is the contract.
// It runs the task-close sequence of skills/flow/implement.md (**The next
// implementer overlaps the guard.**, step 2), composing the guards
// in-process through Registry.
func init() { Registry["close-task"] = closeTask }

func closeTask(args []string, env Env, stdout, stderr io.Writer) (exit int) {
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
		rec := []string{"record", "dispatch", "end", "-change", name, "-C", canonical, "-key", *endKey, "-session-token", *token}
		if *endCommit != "" {
			// The store takes a full sha only; expand an abbreviated one,
			// passing it unchanged where it does not resolve.
			sha := *endCommit
			if full, ok := capture(envGit(env)("-C", canonical, "rev-parse", "--verify", "--quiet", sha+"^{commit}")); ok && full != "" {
				sha = full
			}
			rec = append(rec, "-commit", sha)
		}
		rec = append(rec, "-outcome", "completed")
		if ctFlow(env, stdout, stderr, rec...) != 0 {
			fmt.Fprintln(stderr, "close-task: the dispatch end did not reach the store — continuing; a record never blocks")
			// Printed last on every exit from here, and exit 3 where it
			// would be 0 (design.md: close-task-unrecorded-exit-3).
			defer func() {
				fmt.Fprintf(stderr, "close-task: UNRECORDED — the dispatch end for %s did not reach the store; run: flow %s\n", *endKey, strings.Join(rec, " "))
				if exit == 0 {
					exit = 3
				}
			}()
		}
	}

	// SIGNAL FOUR at close time (scripts/check-unfinished-work.sh): an
	// appended task no dispatch row names is refused here, nothing ticked
	// or pushed, while its row can still be recorded. A warning was ticked
	// past (kan-982: Tasks 19-21 closed with none) and surfaced only at
	// integrate's gate. A store that cannot be read refuses nothing --
	// ctUndispatched reads it as no finding: a record never blocks.
	var closed []string
	for _, t := range tasks {
		closed = append(closed, t.id)
	}
	undispatched := ctUndispatched(env, canonical, name, closed)
	for _, id := range undispatched {
		fmt.Fprintf(stderr, "close-task: appended task %s has no dispatch row naming it — record its dispatch (`flow record dispatch backfill … -task %s -started-at <instant> -ended-at <instant> -outcome <outcome>`), then re-run without the -end-* flags\n", id, id)
	}

	// Both commit guards run before anything is ticked or pushed, so a
	// refused commit is never pushed (design.md: close-task-push-after-guards).
	worst := 0
	for _, t := range tasks {
		worst = max(worst, Registry["check-task-commit-fields"]([]string{canonical, t.id, t.commit, "", canonical, name}, env, stdout, stderr))
		worst = max(worst, ctCheckTrailer(env, stdout, stderr, canonical, name, t.id, t.commit))
	}
	worst = max(worst, Registry["check-task-commit-planning-paths"]([]string{canonical, mergeBase}, env, stdout, stderr))
	// The push below sends every commit origin lacks: one outside every task
	// needs its Flow-Key: trailer before it leaves (git-boundaries.md,
	// Commits outside every task; kan-941 pushed four with none). A worktree
	// whose branch origin does not hold yet has nothing to judge.
	for _, wt := range worktrees {
		remote := "refs/remotes/origin/spectre/" + name
		if _, ok := capture(envGit(env)("-C", wt, "rev-parse", "--verify", "-q", remote)); ok {
			worst = max(worst, Registry["check-untasked-commits"]([]string{wt, remote, "--not-remotes"}, env, stdout, stderr))
		}
	}
	switch worst {
	case 0:
	case 1:
		fmt.Fprintln(stderr, "close-task: RE-COMMIT — a commit guard refused; nothing was ticked or pushed")
		return 1
	default:
		fmt.Fprintln(stderr, "close-task: STOP — a commit guard could not judge; nothing was ticked or pushed")
		return 2
	}
	if len(undispatched) > 0 {
		fmt.Fprintln(stderr, "close-task: RECORD — an appended task has no dispatch row; nothing was ticked or pushed")
		return 1
	}

	var quiet, all, fired []string
	for _, t := range tasks {
		var verdict bytes.Buffer
		rc := Registry["check-review-gate"](append([]string{canonical, t.id, t.commit, canonical, name}, undeclared[t.id]...), env, io.MultiWriter(stdout, &verdict), stderr)
		if rc != 0 {
			fmt.Fprintln(stderr, "close-task: STOP — the review gate could not judge; nothing was ticked or pushed")
			return 2
		}
		if strings.HasPrefix(verdict.String(), "QUIET:") {
			quiet = append(quiet, t.id)
		} else {
			fired = append(fired, t.id)
		}
		all = append(all, t.id)
	}
	ticks := quiet
	if len(fired) > 0 {
		switch ctNewestClass(env, name, canonical) {
		case "micro", "small", "regular":
			ticks = all
		case "big":
		default:
			for _, id := range fired {
				fmt.Fprintf(stdout, "close-task: task %s tick waits — decision class unreadable\n", id)
			}
		}
	}

	rc := 0
	for _, id := range ticks {
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

// ctNewestClass is the change's newest decision class, "" when it cannot be
// read. Below big no gated reviewer runs and a fired task is the panel's
// review focus, so it ticks here as a quiet one does (implement.md, **The
// guard's pass ticks an ungated task.**); on big its tick waits for its
// reviewer; unreadable, the tick waits and says so, since a big fallback
// would strand a fired task on a class no reviewer runs on.
func ctNewestClass(env Env, name, canonical string) string {
	out, _, rc := pcFlow(env, false, "record", "decisions", "-change", name, "-C", canonical)
	if rc != 0 {
		return ""
	}
	c, ok := trsdNewestClass(out)
	if !ok {
		return ""
	}
	return c
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

var ctTaskNumberRE = regexp.MustCompile(`[0-9]+`)

// ctCheckTrailer refuses (1) a task commit — every pair of a commit map —
// whose Task-Id: trailer names neither the task nor a task its Build: red
// fold joins (the red's commit carries its partner's id, implement.md). The
// trailer, not the sha close-task was handed, is how a task's commit is
// found later: flow-task-list's landed check, check-task-commit-planning-
// paths and the panel's fix round all read it, so a commit without it closes
// a task nothing can find again. 2 when git cannot read a commit.
func ctCheckTrailer(env Env, stdout, stderr io.Writer, canonical, name, id, commit string) int {
	accepted := []string{id}
	if tasksMD, _ := tcfResolveTasksMD(env, io.Discard, "close-task", canonical, name, canonical); tasksMD != "" {
		if raw, err := os.ReadFile(tasksMD); err == nil {
			if lines, err := tcfPlanLines(raw); err == nil {
				if task, ok := tcfParseTask(lines, id); ok {
					for _, red := range tcfFoldReds(lines, task, tcfRedTasks(lines)) {
						redTask, _ := tcfParseTask(lines, red)
						accepted = append(append(accepted, red), tcfJoinedPartners(lines, redTask)...)
					}
				}
			}
		}
	}
	pairs := [][2]string{{canonical, commit}}
	if strings.Contains(commit, "=") {
		pairs = nil
		for _, pair := range strings.Split(commit, ",") {
			wt, sha, _ := strings.Cut(strings.TrimSpace(pair), "=")
			pairs = append(pairs, [2]string{wt, sha})
		}
	}
	git := envGit(env)
	rc := 0
	for _, p := range pairs {
		cmd := git("-C", p[0], "log", "-1", "--format=%(trailers:key=Task-Id,valueonly)", p[1])
		cmd.Stderr = stderr
		out, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(stderr, "close-task: cannot read the Task-Id trailer of %s in %s\n", p[1], p[0])
			return 2
		}
		named := ctTaskNumberRE.FindAllString(string(out), -1)
		if !slices.ContainsFunc(named, func(n string) bool { return slices.Contains(accepted, n) }) {
			fmt.Fprintf(stdout, "task %s: commit %s carries no Task-Id: trailer naming task %s (found %q) — re-commit with `Task-Id: %s`\n", id, p[1], id, strings.TrimSpace(string(out)), id)
			rc = 1
		}
	}
	return rc
}

// ctUndispatched is signal four's predicate (uwAppended, uwRowTasks) at
// close time: the closed tasks the plan's **Tasks appended:** line marks as
// appended that no dispatch row of the change names. Nil when no closed task
// is appended -- the store is then never read -- or when the plan or the
// store cannot be read: a warning that cannot judge stays silent.
func ctUndispatched(env Env, canonical, name string, closed []string) []string {
	tasksMD, _ := tcfResolveTasksMD(env, io.Discard, "close-task", canonical, name, canonical)
	if tasksMD == "" {
		return nil
	}
	appended, ok := uwAppended(tasksMD, false)
	if !ok {
		return nil
	}
	var judged []string
	for _, id := range closed {
		if slices.Contains(appended, id) {
			judged = append(judged, id)
		}
	}
	if len(judged) == 0 {
		return nil
	}
	out, _, rc := pcFlow(env, false, "record", "dispatches", "-change", name, "-C", canonical)
	var rows []struct {
		Key    string `json:"key"`
		TaskID string `json:"taskId"`
	}
	if rc != 0 || json.Unmarshal(out, &rows) != nil {
		return nil
	}
	named := map[string]bool{}
	for _, d := range rows {
		for _, id := range uwRowTasks(d.Key, d.TaskID) {
			named[id] = true
		}
	}
	var missing []string
	for _, id := range judged {
		if !named[id] {
			missing = append(missing, id)
		}
	}
	return missing
}
