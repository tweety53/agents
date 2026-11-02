package guard

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/close-task.sh shim (KAN-1058).
//
// flow-guard close-task — run the task-close sequence of skills/flow/implement.md
// (**A group boundary is one wake, one Bash call and one launch
// message.**, step 2) for every task a
// returning implementer committed.
//
// Usage: flow-guard close-task [-end-key <key> -session-token <token> [-end-commit <sha>]]
//          [-undeclared <task-id>=<path>[,<path>…]]…
//          <canonical-worktree> <name> <merge-base> <task-id>:<sha|worktree=sha[,worktree=sha…]>…
//
// In this order:
//   1. With -end-key: `flow record dispatch end -change <name> -key <key>
//      -session-token <token> [-commit <sha>] -outcome completed`. A record
//      never blocks: a failed call prints a warning and the sequence goes on,
//      and every exit after it prints `close-task: UNRECORDED — the dispatch
//      end for <key> did not reach the store; run: flow record dispatch end
//      …` (the full command) as the last stderr line, turning exit 0 into 3.
//      Then, for every task it closes that the plan's `**Tasks appended:**`
//      line counts or whose title opens `In-run fix <k> — `,
//      `flow record dispatches -change <name> -C <canonical>`:
//      an appended task no row names, or whose naming rows all carry no
//      commit, is refused at step 4 (exit 1); a store that cannot be read
//      refuses nothing.
//   2. Each task's commit first gains one <worktree>=<sha> pair per worktree
//      of `flow state get -C <canonical> <name>`'s worktrees map the argument
//      left out whose newest commit since its base carries the task's
//      Task-Id: trailer — a bare sha is the canonical worktree's pair — each
//      named on stderr as `close-task: task <id>: joined <worktree>=<sha>`
//      (kan-1008's self-review: a map omitting the backend repository left
//      its commits unjudged). Then
//      flow-guard check-task-commit-fields <canonical> <id> <sha|map> "" <canonical>
//      <name> for every task, and its commit's — every map pair's — Task-Id:
//      trailer must name the task or a task its Build: red fold joins (exit 1
//      otherwise): the trailer is how a task's commit is found again.
//      A commit — any map pair's — whose diff-tree lists a path named
//      full-app-suite.zip has that zip rebuilt in its worktree by
//      flow-guard rebuild-full-app-suite-zip, the zip's directory as <screenshots
//      root>, and the committed bytes written back — save a zip a later
//      commit of the same close, in the same worktree, changes again, which
//      is that commit's to judge (kan-1051): a difference prints
//      `close-task: task <id>: <path> is not flow-guard rebuild-full-app-suite-zip's
//      output — rebuild it with that script and re-commit` (exit 1); a
//      screenshots root git status reports dirty, or a rebuild exit other
//      than 0, is exit 2 (kan-1016's self-review: a hand-built zip).
//      When a task it closes is one step 1 counts as appended or in-run,
//      flow-guard check-plan-shape <the change's tasks.md> runs too, its exit
//      folded in at step 4 (kan-1046 self-review, 1f: an in-run fix closed
//      with a **Files:** the plan guard refuses).
//   3. flow-guard check-task-commit-planning-paths <canonical> <merge-base>, then
//      flow-guard check-untasked-commits <worktree> refs/remotes/origin/spectre/<name>
//      once per distinct worktree whose branch origin already holds: the push
//      at step 7 sends those commits, and one outside every task carries its
//      `Flow-Key:` trailer before it leaves (kan-941).
//   4. Any guard exit 2 → exit 2; else any exit 1 → exit 1; else an
//      appended task step 1 found no row naming → exit 1. Nothing is ticked
//      or pushed either way, so a refused commit is never pushed.
//   5. flow-guard check-review-gate <canonical> <id> <sha|map> <canonical> <name>
//      [<the task's -undeclared paths>…] for every task; its FIRE:/QUIET:
//      line is printed. A gate exit 2 → exit 2, nothing ticked or pushed.
//   6. `flow tasks tick -C <canonical> <name> <id>` for every task the gate
//      left QUIET — and, when a gate fired, for every task once `flow record
//      decisions -change <name> -C <canonical>` names the newest class micro,
//      small or regular: no gated reviewer runs there, the whole-branch panel
//      reviews a fired task. On big the tick waits; on an unreadable class it
//      waits too and prints `close-task: task <id> tick waits — decision
//      class unreadable` per fired task.
//   7. `git -C <worktree> push origin spectre/<name>` once per distinct
//      worktree: the canonical one for a bare sha, each map pair's otherwise.
//   A failed tick or push → exit 2; every tick and push is still attempted.
//
// The guards run in-process in flow-guard; their own lines are printed as
// they print them.
//
// Exit 0 closed (read each FIRE: line: on big that task's tick waits for its
// reviewer; elsewhere it is the panel's review focus); 1 a commit guard refused — re-commit, then re-run without the
// -end-* flags — or an appended task has no dispatch row — record it, then
// re-run without the -end-* flags; 2 stop: a guard could not judge, a tick or push failed, or
// a usage error; 3 closed as exit 0 — every tick and push done — but step 1's
// dispatch end never reached the store: run the command the UNRECORDED line
// prints.

// closeTask is scripts/close-task.sh: that script's header is the contract.
// It runs the task-close sequence of skills/flow/implement.md (**A group
// boundary is one wake, one Bash call and one launch message.**, step 2),
// composing the guards
// in-process through Registry.
func init() { Registry["close-task"] = closeTask }

func closeTask(args []string, env Env, stdout, stderr io.Writer) (exit int) {
	usage := func(msg string) int {
		fmt.Fprintf(stderr, "close-task: %s\nusage: flow-guard close-task [-end-key <key> -session-token <token> [-end-commit <sha>]] [-undeclared <task-id>=<path>[,<path>…]]… <canonical-worktree> <name> <merge-base> <task-id>:<sha|worktree=sha[,worktree=sha…]>…\n", msg)
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
	recorded := ctRecordedWorktrees(env, canonical, name)
	for _, a := range pos[3:] {
		id, commit, ok := strings.Cut(a, ":")
		if !ok || id == "" || commit == "" {
			return usage("task argument is not <task-id>:<sha|map>: " + a)
		}
		commit = ctJoinRecorded(env, stderr, canonical, id, commit, recorded)
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
	for _, line := range undispatched {
		fmt.Fprintln(stderr, line)
	}

	// Both commit guards run before anything is ticked or pushed, so a
	// refused commit is never pushed (design.md: close-task-push-after-guards).
	worst := 0
	// The plan guard runs on every close, before any commit guard: an
	// appended or in-run fix task was written after it last ran (kan-1046
	// self-review, 1f), and a `flow tasks files add` widening of a planned
	// task can break plan shape too (kan-1067 self-review, 1-B: a test path
	// under `Tests: none`, found only at verify).
	if tasksMD, _ := tcfResolveTasksMD(env, io.Discard, "close-task", canonical, name, canonical); tasksMD != "" {
		worst = max(worst, Registry["check-plan-shape"]([]string{tasksMD}, env, stdout, stderr))
	}
	for _, t := range tasks {
		worst = max(worst, Registry["check-task-commit-fields"]([]string{canonical, t.id, t.commit, "", canonical, name}, env, stdout, stderr))
		worst = max(worst, ctCheckTrailer(env, stdout, stderr, canonical, name, t.id, t.commit))
		var others []string
		for _, o := range tasks {
			if o.id != t.id {
				others = append(others, o.commit)
			}
		}
		worst = max(worst, ctCheckSuiteZips(env, stderr, canonical, t.id, t.commit, others))
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
		fmt.Fprintln(stderr, "close-task: RECORD — an appended task has no dispatch row carrying its commit; nothing was ticked or pushed")
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

// ctRecordedWorktrees is the change's recorded worktree set, the keys of
// `flow state get`'s worktrees map, sorted; nil when it cannot be read.
func ctRecordedWorktrees(env Env, canonical, name string) []string {
	out, _, rc := pcFlow(env, false, "state", "get", "-C", canonical, name)
	var rec struct {
		Worktrees map[string]json.RawMessage `json:"worktrees"`
	}
	if rc != 0 || json.Unmarshal(out, &rec) != nil {
		return nil
	}
	var wts []string
	for wt := range rec.Worktrees {
		if wt != "" {
			wts = append(wts, wt)
		}
	}
	slices.Sort(wts)
	return wts
}

// ctJoinRecorded adds to a task's commit — a bare sha is the canonical
// worktree's — one <worktree>=<sha> pair for each recorded worktree the
// caller left out whose newest commit since its base carries the task's
// Task-Id: trailer, and names each on stderr (kan-1008's self-review: the
// parent's map omitted the backend repository, and Task 13's and 14's
// backend commits were never judged against their Files:). The base is
// vmDefaultBase's; a worktree with none, or that git cannot read, adds
// nothing.
func ctJoinRecorded(env Env, stderr io.Writer, canonical, id, commit string, recorded []string) string {
	pairs := []string{canonical + "=" + commit}
	if strings.Contains(commit, "=") {
		pairs = strings.Split(commit, ",")
	}
	have := map[string]bool{}
	for _, p := range pairs {
		wt, _, _ := strings.Cut(strings.TrimSpace(p), "=")
		have[trResolvePath(env, wt)] = true
	}
	git := envGit(env)
	added := false
	for _, wt := range recorded {
		if have[trResolvePath(env, wt)] {
			continue
		}
		base, _, ok := vmDefaultBase(git, wt)
		if !ok {
			continue
		}
		out, ok := capture(git("-C", wt, "log", "--no-merges", "--format=%H %(trailers:key=Task-Id,valueonly,separator=%x2C)", base+"..HEAD"))
		if !ok {
			continue
		}
		for _, line := range strings.Split(out, "\n") {
			sha, ids, _ := strings.Cut(strings.TrimSpace(line), " ")
			if sha != "" && slices.Contains(ctTaskNumberRE.FindAllString(ids, -1), id) {
				fmt.Fprintf(stderr, "close-task: task %s: joined %s=%s — a recorded worktree the commit map left out\n", id, wt, sha)
				pairs = append(pairs, wt+"="+sha)
				added = true
				break
			}
		}
	}
	if !added {
		return commit
	}
	return strings.Join(pairs, ",")
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
// close time, one refusal line per closed task the plan's **Tasks
// appended:** line marks as appended that no dispatch row of the change
// names -- or whose naming rows all carry no commit, which the ledger then
// reads as "no commit" (kan-1059 self-review, 5a). Nil when no closed task
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
		ID        int64  `json:"id"`
		Key       string `json:"key"`
		TaskID    string `json:"taskId"`
		CommitSha string `json:"commitSha"`
	}
	if rc != 0 || json.Unmarshal(out, &rows) != nil {
		return nil
	}
	named := map[string][]int64{}
	committed := map[string]bool{}
	for _, d := range rows {
		for _, id := range uwRowTasks(d.Key, d.TaskID) {
			named[id] = append(named[id], d.ID)
			committed[id] = committed[id] || d.CommitSha != ""
		}
	}
	var lines []string
	for _, id := range judged {
		switch {
		case named[id] == nil:
			lines = append(lines, fmt.Sprintf("close-task: appended task %s has no dispatch row naming it — record its dispatch (`flow record dispatch backfill … -task %s -started-at <instant> -ended-at <instant> -outcome <outcome>`), then re-run without the -end-* flags", id, id))
		case !committed[id]:
			for _, row := range named[id] {
				lines = append(lines, fmt.Sprintf("close-task: appended task %s: dispatch row %d carries no commit — stamp it (`flowd retime-dispatch %d <started-at> <ended-at> <the task's sha>`, or `-commit` on the back-fill), then re-run without the -end-* flags", id, row, row))
			}
		}
	}
	return lines
}

// ctCheckSuiteZips refuses a task commit — every map pair's — whose
// full-app-suite.zip is not rebuild-full-app-suite-zip.sh's output: it
// rebuilds the zip in that worktree, compares the bytes with the committed
// ones and writes the on-disk bytes back (kan-1016's self-review: a zip
// hand-built with archive/zip was committed). The rebuild's <worktree> is the
// canonical worktree, whose .flow/project.md declares where screenshots live
// — the zip may sit in another repository, the regression checkout. A commit
// deleting the zip is not judged, and neither is one a later commit of the
// same close changes again — others holds the close's other task commits —
// since the worktree holds only the last one's screenshots root (kan-1051:
// tasks 6 and 7 both wrote the zip and closed together). 1 a difference; 2 a
// screenshots root that differs from the commit's or a rebuild that did not
// answer 0.
func ctCheckSuiteZips(env Env, stderr io.Writer, canonical, id, commit string, others []string) int {
	pairsOf := func(c string) []string {
		if strings.Contains(c, "=") {
			return strings.Split(c, ",")
		}
		return []string{canonical + "=" + c}
	}
	pairs := pairsOf(commit)
	absOf := func(p string) string {
		if filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(env.Dir, p)
	}
	absCanonical := absOf(canonical)
	git := envGit(env)
	// superseded: a later commit of the close, in the same worktree, changes path too.
	superseded := func(wt, sha, path string) bool {
		for _, o := range others {
			for _, op := range pairsOf(o) {
				owt, osha, _ := strings.Cut(strings.TrimSpace(op), "=")
				if osha == sha || absOf(owt) != absOf(wt) || git("-C", wt, "merge-base", "--is-ancestor", sha, osha).Run() != nil {
					continue
				}
				if out, _ := capture(git("-C", wt, "diff-tree", "--no-renames", "--no-commit-id", "--name-only", "-r", osha, "--", path)); out != "" {
					return true
				}
			}
		}
		return false
	}
	worst := 0
	for _, p := range pairs {
		wt, sha, _ := strings.Cut(strings.TrimSpace(p), "=")
		out, ok := capture(git("-C", wt, "diff-tree", "--no-renames", "--no-commit-id", "--name-only", "--diff-filter=d", "-r", sha))
		if !ok {
			continue // the fields guard already answers for an unreadable commit
		}
		for _, path := range strings.Split(out, "\n") {
			if filepath.Base(path) != "full-app-suite.zip" || superseded(wt, sha, path) {
				continue
			}
			abs := filepath.Join(absOf(wt), path)
			root := filepath.Dir(abs)
			untracked, ok := capture(git("-C", wt, "ls-files", "--others", "--exclude-standard", "--", root))
			if !ok || untracked != "" || git("-C", wt, "diff", "--quiet", sha, "--", root).Run() != nil {
				fmt.Fprintf(stderr, "close-task: task %s: %s is not the commit's — cannot rebuild %s to compare\n", id, root, path)
				worst = max(worst, 2)
				continue
			}
			committed, err := git("-C", wt, "show", sha+":"+path).Output()
			if err != nil {
				fmt.Fprintf(stderr, "close-task: task %s: cannot read %s at %s\n", id, path, sha)
				worst = max(worst, 2)
				continue
			}
			onDisk, err := os.ReadFile(abs)
			if err != nil {
				fmt.Fprintf(stderr, "close-task: task %s: cannot read %s: %v\n", id, path, err)
				worst = max(worst, 2)
				continue
			}
			rbEnv := env
			rbEnv.Dir = absCanonical
			rc := Registry["rebuild-full-app-suite-zip"]([]string{absCanonical, root}, rbEnv, io.Discard, stderr)
			rebuilt, _ := os.ReadFile(abs)
			if err := os.WriteFile(abs, onDisk, 0o644); err != nil {
				fmt.Fprintf(stderr, "close-task: task %s: cannot restore %s: %v\n", id, path, err)
				worst = max(worst, 2)
				continue
			}
			switch {
			case rc != 0:
				fmt.Fprintf(stderr, "close-task: task %s: rebuild-full-app-suite-zip exited %d — cannot judge %s\n", id, rc, path)
				worst = max(worst, 2)
			case !bytes.Equal(rebuilt, committed):
				fmt.Fprintf(stderr, "close-task: task %s: %s is not rebuild-full-app-suite-zip.sh's output — rebuild it with that script and re-commit\n", id, path)
				worst = max(worst, 1)
			}
		}
	}
	return worst
}
