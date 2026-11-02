package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-late-fix-trigger.sh shim (KAN-1058).
//
// flow-guard check-late-fix-trigger — decide whether a fix run's pass 1 takes the
// late-fix reduction (skills/flow/review-panel-late-fix.md).
//
// Usage: flow-guard check-late-fix-trigger <change> <tasks.md> <worktree> <since-close-sha|-> <base-verdict> \
//          [<worktree> <since-close-sha|-> <base-verdict>...]
//
// One triple per worktree in the resolved set, the canonical worktree's
// first. <since-close-sha> is the sha the stage's last clean close reviewed
// in that worktree, `-` where there was no earlier clean close;
// <base-verdict> is that worktree's flow-guard check-base-moved line this round.
//
// Every condition must hold to reduce:
//   1  already clean, already verified — flow-guard check-panel-findings-closed
//      <canonical-worktree> <change> exits 0 (run in process), and every
//      worktree names a since-close sha;
//   2  the base has not moved since that close — every base verdict is CLEAR;
//   3  the delta is small — `git diff --numstat <since-close-sha>` per
//      worktree, insertions and deletions summed across the set, is at most
//      40 changed lines; a binary entry (`-`) cannot be counted and fails;
//   4  no scope growth — <tasks.md> carries no task line
//      (`^- \[[ xX]\] [0-9]+\.`) whose task number is absent from
//      <canonical-worktree>/.superpowers/sdd/<change>/tasks-at-close.md, the
//      copy of the plan the review panel's close step writes at each clean
//      close (the plan is gitignored, so no commit holds it); a missing copy
//      fails, naming it on stderr;
//   5  the panel's own machinery is untouched — no path in
//      `git diff --name-only <since-close-sha>` matches
//      skills/flow/review-panel*.md, scripts/check-panel-*.sh,
//      skills/flow/*-reviewer-prompt.md or skills/flow/engineering-principles.md.
//
// Prints on stdout:
//   exit 0  late-fix reduction: <n> changed lines since <canonical since-close-sha>
//   exit 1  full path: condition <k> — <reason>   (one line per failed condition)
//   exit 3  append scope: <n> changed lines since <canonical since-close sha>
//   exit 4  small append: <n> changed lines since <canonical since-close sha>
//
// Exit codes:
//   0  reduce
//   1  full path
//   2  cannot answer — usage, a change name outside the allowlist, a
//      worktree/sha pair that fails panel_validate_worktree, a git failure,
//      an unreadable <tasks.md>, or check-panel-findings-closed
//      exit 2 (its stderr passed through); the caller reads it as full path
//   3  append scope — conditions 1, 2 and 5 hold and only 3 and/or 4 failed,
//      and the delta is over 250 counted lines or has a binary entry:
//      pass 1 runs the decided roster on the since-close delta
//      (skills/flow/review-panel-late-fix.md, The append scope)
//   4  small append — as 3, but no binary entry and the summed delta is at
//      most 250 changed lines: pass 1 runs `primary` alone on the since-close
//      delta (skills/flow/review-panel-late-fix.md, The small append)
//
// WHY 250. Appended fix work is reviewed at the size it lands at, and most
// lands well under this: KAN-930 ran the full roster on 185- and 222-line
// appends that came back clean or Minor-only, while its 350- and 537-line
// late-task appends — real new scope — stay on the decided roster.

// checkLateFixTrigger is scripts/check-late-fix-trigger.sh: that script's
// header comment is the contract -- the five conditions of the late-fix
// reduction (skills/flow/review-panel-late-fix.md), exit 0 reduce (the
// `late-fix reduction:` line), 1 full path (one line per failed condition),
// 2 cannot answer, 3 append scope (the `append scope:` line -- only
// conditions 3 and/or 4 failed), 4 small append (the `small append:` line --
// the same, on a countable delta of at most lftSmallAppendLines). Condition
// 4 compares tasks.md with the canonical worktree's
// .superpowers/sdd/<change>/tasks-at-close.md, the copy the review panel's
// close step writes (KAN-1073: the plan is gitignored, so no commit holds it).
func init() {
	Registry["check-late-fix-trigger"] = checkLateFixTrigger
}

const (
	lftUsage = "usage: flow-guard check-late-fix-trigger <change> <tasks.md> <worktree> <since-close-sha|-> <base-verdict> [<worktree> <since-close-sha|-> <base-verdict>...]\n"
	// lftMaxLines is condition 3's cap on the delta's changed lines.
	lftMaxLines = 40
	// lftSmallAppendLines caps the small append's countable delta.
	lftSmallAppendLines = 250
)

var (
	// lftTaskLine is a plan task line, its task number captured: condition
	// 4's added task is a number in tasks.md the at-close copy lacks, so a
	// tick since the close is no growth.
	lftTaskLine = regexp.MustCompile(`^- \[[ xX]\] ([0-9]+)\.`)
	// lftMachinery is condition 5's panel machinery, as repository-relative
	// path.Match patterns.
	lftMachinery = []string{"skills/flow/review-panel*.md", "scripts/check-panel-*.sh",
		"skills/flow/*-reviewer-prompt.md", "skills/flow/engineering-principles.md"}
)

// lftTaskNumbers is the set of task numbers plan's task lines carry.
func lftTaskNumbers(plan string) map[string]bool {
	out := map[string]bool{}
	for _, l := range strings.Split(plan, "\n") {
		if m := lftTaskLine.FindStringSubmatch(l); m != nil {
			out[m[1]] = true
		}
	}
	return out
}

func checkLateFixTrigger(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 5 || (len(args)-2)%3 != 0 {
		fmt.Fprint(stderr, lftUsage)
		return 2
	}
	change, tasks, triples := args[0], args[1], args[2:]
	if !plainChangeName(change) {
		fmt.Fprintf(stderr, "check-late-fix-trigger: change name '%s' is not a plain change name\n", change)
		return 2
	}
	gitBin, ok := panelResolveGit(env, "check-late-fix-trigger", stderr)
	if !ok {
		return 2
	}
	gitOut := func(wt string, a ...string) (string, bool) {
		out, ok := capture(panelGit(env, gitBin, wt, a...))
		if !ok {
			fmt.Fprintf(stderr, "check-late-fix-trigger: git %s failed in %s\n", strings.Join(a, " "), wt)
		}
		return out, ok
	}
	var failed []string
	appendScope := true // every failed condition is 3 or 4
	fail := func(n int, format string, a ...any) {
		failed = append(failed, fmt.Sprintf("full path: condition %d — "+format, append([]any{n}, a...)...))
		appendScope = appendScope && (n == 3 || n == 4)
	}

	// Condition 1: the findings are closed, and every worktree names a
	// since-close sha.
	canon := triples[0]
	var closedOut, closedErr bytes.Buffer
	var why []string
	switch checkPanelFindingsClosed([]string{canon, change}, env, &closedOut, &closedErr) {
	case 0:
	case 1:
		reason := "the change's findings are not all closed (check-panel-findings-closed.sh exited 1)"
		var lines []string
		for _, l := range strings.Split(closedErr.String(), "\n") {
			if l = strings.TrimSpace(l); l != "" {
				lines = append(lines, l)
			}
		}
		if len(lines) > 0 {
			reason += ": " + strings.Join(lines, "; ")
		}
		why = append(why, reason)
	default:
		fmt.Fprint(stderr, closedErr.String())
		return 2
	}
	var noClose []string
	for i := 0; i < len(triples); i += 3 {
		if triples[i+1] == "-" {
			noClose = append(noClose, triples[i])
		}
	}
	if len(noClose) > 0 {
		why = append(why, "no earlier clean close in "+strings.Join(noClose, ", "))
	}
	if len(why) > 0 {
		fail(1, "%s", strings.Join(why, "; "))
	}

	// Condition 2: every base verdict is CLEAR.
	var moved []string
	for i := 0; i < len(triples); i += 3 {
		if f := strings.Fields(triples[i+2]); len(f) == 0 || strings.TrimSuffix(f[0], ":") != "CLEAR" {
			moved = append(moved, triples[i])
		}
	}
	if len(moved) > 0 {
		fail(2, "the base verdict is not CLEAR in %s", strings.Join(moved, ", "))
	}

	// Conditions 3-5 read each worktree's delta since its close.
	lines, binary, taskLine, missingCopy := 0, []string{}, false, ""
	var machinery []string
	for i := 0; i < len(triples); i += 3 {
		wt, since := triples[i], triples[i+1]
		if since == "-" {
			continue
		}
		if !panelValidateWorktree(env, "check-late-fix-trigger", wt, since, gitBin, stderr) {
			return 2
		}
		numstat, ok := gitOut(wt, "diff", "--no-renames", "--numstat", since)
		if !ok {
			return 2
		}
		for _, l := range strings.Split(numstat, "\n") {
			f := strings.SplitN(l, "\t", 3)
			if len(f) < 3 {
				continue
			}
			if f[0] == "-" || f[1] == "-" {
				binary = append(binary, f[2])
				continue
			}
			add, err1 := strconv.Atoi(f[0])
			del, err2 := strconv.Atoi(f[1])
			if err1 != nil || err2 != nil {
				fmt.Fprintf(stderr, "check-late-fix-trigger: unreadable numstat line in %s: %q\n", wt, l)
				return 2
			}
			lines += add + del
		}
		names, ok := gitOut(wt, "diff", "--no-renames", "--name-only", since)
		if !ok {
			return 2
		}
		for _, p := range strings.Split(names, "\n") {
			for _, m := range lftMachinery {
				if hit, _ := path.Match(m, p); hit {
					machinery = append(machinery, p)
					break
				}
			}
		}
		if i == 0 {
			copyPath := filepath.Join(wt, ".superpowers", "sdd", change, "tasks-at-close.md")
			closed, err := os.ReadFile(copyPath)
			if err != nil {
				missingCopy = copyPath
				continue
			}
			tasksPath := tasks
			if !filepath.IsAbs(tasksPath) {
				tasksPath = filepath.Join(wt, tasksPath)
			}
			cur, err := os.ReadFile(tasksPath)
			if err != nil {
				fmt.Fprintf(stderr, "check-late-fix-trigger: cannot read %s: %v\n", tasksPath, err)
				return 2
			}
			atClose := lftTaskNumbers(string(closed))
			for n := range lftTaskNumbers(string(cur)) {
				if !atClose[n] {
					taskLine = true
				}
			}
		}
	}
	switch {
	case len(binary) > 0:
		fail(3, "a binary change cannot be counted: %s", strings.Join(binary, ", "))
	case lines > lftMaxLines:
		fail(3, "%d changed lines since the close, over %d", lines, lftMaxLines)
	}
	switch {
	case missingCopy != "":
		fmt.Fprintf(stderr, "check-late-fix-trigger: no at-close copy of the plan at %s -- the review panel's close step writes it\n", missingCopy)
		fail(4, "no at-close copy of %s at %s", tasks, missingCopy)
	case taskLine:
		fail(4, "%s adds a task line since the close", tasks)
	}
	if len(machinery) > 0 {
		fail(5, "the delta touches the panel's own machinery: %s", strings.Join(machinery, ", "))
	}

	if len(failed) > 0 && appendScope && len(binary) == 0 && lines <= lftSmallAppendLines {
		fmt.Fprintf(stdout, "small append: %d changed lines since %s\n", lines, triples[1])
		return 4
	}
	if len(failed) > 0 && appendScope {
		fmt.Fprintf(stdout, "append scope: %d changed lines since %s\n", lines, triples[1])
		return 3
	}
	if len(failed) > 0 {
		fmt.Fprintln(stdout, strings.Join(failed, "\n"))
		return 1
	}
	fmt.Fprintf(stdout, "late-fix reduction: %d changed lines since %s\n", lines, triples[1])
	return 0
}
