package guard

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// checkLateFixTrigger is scripts/check-late-fix-trigger.sh: that script's
// header comment is the contract -- the five conditions of the late-fix
// reduction (skills/flow/review-panel-late-fix.md), exit 0 reduce (the
// `late-fix reduction:` line), 1 full path (one line per failed condition),
// 2 cannot answer, 3 append scope (the `append scope:` line -- only
// conditions 3 and/or 4 failed).
func init() {
	Registry["check-late-fix-trigger"] = checkLateFixTrigger
}

const (
	lftUsage = "usage: check-late-fix-trigger.sh <change> <tasks.md> <worktree> <since-close-sha|-> <base-verdict> [<worktree> <since-close-sha|-> <base-verdict>...]\n"
	// lftMaxLines is condition 3's cap on the delta's changed lines.
	lftMaxLines = 40
)

var (
	lftTaskLine = regexp.MustCompile(`^\+- \[[ xX]\] [0-9]+\.`)
	// lftMachinery is condition 5's panel machinery, as repository-relative
	// path.Match patterns.
	lftMachinery = []string{"skills/flow/review-panel*.md", "scripts/check-panel-*.sh",
		"skills/flow/*-reviewer-prompt.md", "skills/flow/engineering-principles.md"}
)

func checkLateFixTrigger(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 5 || (len(args)-2)%3 != 0 {
		fmt.Fprint(stderr, lftUsage)
		return 2
	}
	change, tasks, triples := args[0], args[1], args[2:]
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
		why = append(why, "the change's findings are not all closed (check-panel-findings-closed.sh exited 1)")
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
	lines, binary, taskLine, untracked := 0, []string{}, false, false
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
			if panelGit(env, gitBin, wt, "ls-files", "--error-unmatch", "--", tasks).Run() != nil {
				untracked = true
				continue
			}
			diff, ok := gitOut(wt, "diff", "--no-renames", since, "--", tasks)
			if !ok {
				return 2
			}
			for _, l := range strings.Split(diff, "\n") {
				if lftTaskLine.MatchString(l) {
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
	case untracked:
		fail(4, "%s is not tracked in %s", tasks, canon)
	case taskLine:
		fail(4, "%s adds a task line since the close", tasks)
	}
	if len(machinery) > 0 {
		fail(5, "the delta touches the panel's own machinery: %s", strings.Join(machinery, ", "))
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
