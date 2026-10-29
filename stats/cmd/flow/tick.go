// Command flow's tickTask implements `flow tasks tick`'s core logic:
// flipping one task's own checkbox and every step checkbox in its body
// from "[ ]" to "[x]", per skills/flow-contracts/build-green.md's
// Placement rule. It touches no file and knows nothing about -C or
// which tasks.md it is reading -- stats/cmd/flow/tasks.go (task 2)
// owns that.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrTaskIDInvalid is returned when the caller's task id is not a bare
// non-negative integer -- build-green.md's Placement rule: a task's own
// id is a flat integer, never dotted, never signed.
var ErrTaskIDInvalid = errors.New("task id must be a bare non-negative integer")

// ErrTaskNotFound is returned when no column-0 task line in the plan
// carries the requested id.
var ErrTaskNotFound = errors.New("task not found")

// ErrTaskAlreadyTicked is returned when the target task's own checkbox
// is already "[x]" -- refused rather than a silent no-op, so a
// double-tick is visible (the ticket's own stated requirement).
var ErrTaskAlreadyTicked = errors.New("task is already ticked")

var taskIDPattern = regexp.MustCompile(`^[0-9]+$`)

// taskLinePattern matches a column-0 task checkbox line: "- [ ] 12. Title"
// or "- [x] 12. Title". Group 1 is the mark (" " or "x"), group 2 is the
// flat integer id.
var taskLinePattern = regexp.MustCompile(`^- \[( |x)\] ([0-9]+)\. `)

// headingPattern matches a level-2 or level-3 Markdown heading, which
// closes a task's body per build-green.md's Placement rule exactly as a
// following task line does. Mirrors stats/internal/guard/planparse.go's
// body-boundary rule exactly: a space OR end-of-line after the "#"s, so a
// bare "##"/"###" with nothing following it still closes a body.
var headingPattern = regexp.MustCompile(`^#{2,3}($|\s)`)

// fencePattern matches a fenced code block delimiter: three or more
// backticks or tildes, with up to three leading spaces of indent.
// Mirrors stats/internal/guard/planparse.go's fence rule exactly.
var fencePattern = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")

// stepCheckboxPattern matches an unticked step line: two columns of
// indent, then a checkbox, per build-green.md's "a step is indented two
// columns beneath its task" rule. It intentionally does not require the
// "**Step N: ...**" text that follows -- build-green.md defines a step
// by its indent and checkbox, not by that label.
var stepCheckboxPattern = regexp.MustCompile(`^  - \[ \] `)

// tickTask flips task <taskID>'s own checkbox and every unticked step
// checkbox in its body, in plan (a whole tasks.md file's bytes), and
// returns the rewritten plan plus how many step checkboxes were
// flipped. Every line outside the task's own body -- including every
// other task -- is returned byte-identical.
//
// A task's body runs from its own task line to the next task line, the
// next level-2/3 heading, or the end of the file -- build-green.md's
// Placement rule, applied here exactly as check-task-build-green
// (stats/internal/guard/taskbuildgreen.go) already applies it when scanning the same file. Fence state is
// tracked across the whole scan (mirroring stats/internal/guard/planparse.go's
// task iteration), so a task line, heading or step checkbox shown inside a
// fenced example opens no task, closes no body and is never ticked --
// it is documentation, not structure.
func tickTask(plan []byte, taskID string) ([]byte, int, error) {
	if !taskIDPattern.MatchString(taskID) {
		return nil, 0, fmt.Errorf("%w: %q", ErrTaskIDInvalid, taskID)
	}

	lines := bytes.Split(plan, []byte("\n"))
	start, end, alreadyTicked, err := taskBody(lines, taskID)
	if err != nil {
		return nil, 0, err
	}
	if alreadyTicked {
		return nil, 0, fmt.Errorf("%w: %s", ErrTaskAlreadyTicked, taskID)
	}

	lines[start] = append([]byte("- [x] "), lines[start][len("- [ ] "):]...)

	steps := 0
	inFence := false
	for i := start + 1; i < end; i++ {
		line := lines[i]
		if fencePattern.Match(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if stepCheckboxPattern.Match(line) {
			lines[i] = append([]byte("  - [x] "), line[len("  - [ ] "):]...)
			steps++
		}
	}

	return bytes.Join(lines, []byte("\n")), steps, nil
}

// taskBody finds taskID's column-0 task line in lines, outside fences, and
// returns its index, the index its body ends before (the next task line or
// heading, or len(lines)), and whether its own checkbox is ticked.
func taskBody(lines [][]byte, taskID string) (start, end int, ticked bool, err error) {
	start = -1
	inFence := false
	for i, line := range lines {
		if fencePattern.Match(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if m := taskLinePattern.FindSubmatch(line); m != nil && string(m[2]) == taskID {
			start, ticked = i, string(m[1]) == "x"
			break
		}
	}
	if start == -1 {
		return 0, 0, false, fmt.Errorf("%w: %s", ErrTaskNotFound, taskID)
	}
	end = len(lines)
	inFence = false
	for i := start + 1; i < len(lines); i++ {
		line := lines[i]
		if fencePattern.Match(line) {
			inFence = !inFence
			continue
		}
		if !inFence && (taskLinePattern.Match(line) || headingPattern.Match(line)) {
			end = i
			break
		}
	}
	return start, end, ticked, nil
}

// filesLinePrefix opens a task's **Files:** field.
const filesLinePrefix = "**Files:** "

// filesAdd adds every path in paths not already on task id's **Files:**
// line to it, as a backticked, comma-separated entry -- a `none` field
// becomes the paths alone. Only that task's first unfenced **Files:** line
// changes; every other byte of plan is kept.
func filesAdd(plan []byte, id string, paths []string) ([]byte, error) {
	return editFilesField(plan, id, func(field string) (string, error) {
		if strings.HasPrefix(field, "none") {
			field = ""
		}
		for _, p := range paths {
			entry := "`" + p + "`"
			if strings.Contains(field, entry) {
				continue
			}
			if field != "" {
				field += ", "
			}
			field += entry
		}
		return field, nil
	})
}

// filesRemove drops every path in paths from task id's **Files:** line --
// a field left empty becomes `none` -- and refuses a path the field does
// not name, the plan left unchanged.
func filesRemove(plan []byte, id string, paths []string) ([]byte, error) {
	return editFilesField(plan, id, func(field string) (string, error) {
		entries := strings.Split(field, ", ")
		for _, p := range paths {
			entry := "`" + p + "`"
			kept := entries[:0:0]
			for _, e := range entries {
				if e != entry {
					kept = append(kept, e)
				}
			}
			if len(kept) == len(entries) {
				return "", fmt.Errorf("task %s's **Files:** field does not name %s", id, entry)
			}
			entries = kept
		}
		if len(entries) == 0 {
			return "none", nil
		}
		return strings.Join(entries, ", "), nil
	})
}

// editFilesField replaces task id's first unfenced **Files:** field with
// edit's result; every other byte of plan is kept.
func editFilesField(plan []byte, id string, edit func(field string) (string, error)) ([]byte, error) {
	if !taskIDPattern.MatchString(id) {
		return nil, fmt.Errorf("%w: %q", ErrTaskIDInvalid, id)
	}
	lines := bytes.Split(plan, []byte("\n"))
	start, end, _, err := taskBody(lines, id)
	if err != nil {
		return nil, err
	}
	inFence := false
	for i := start + 1; i < end; i++ {
		line := string(lines[i])
		if fencePattern.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence || !strings.HasPrefix(line, filesLinePrefix) {
			continue
		}
		field, err := edit(strings.TrimPrefix(line, filesLinePrefix))
		if err != nil {
			return nil, err
		}
		lines[i] = []byte(filesLinePrefix + field)
		return bytes.Join(lines, []byte("\n")), nil
	}
	return nil, fmt.Errorf("task %s has no **Files:** line", id)
}

// untickedTasks returns the ids of the plan's unticked task lines, fenced
// lines skipped as countPlanLines skips them.
func untickedTasks(plan []byte) []string {
	var ids []string
	inFence := false
	for _, line := range bytes.Split(plan, []byte("\n")) {
		if fencePattern.Match(line) {
			inFence = !inFence
			continue
		}
		if m := taskLinePattern.FindSubmatch(line); !inFence && m != nil && string(m[1]) == " " {
			ids = append(ids, string(m[2]))
		}
	}
	return ids
}

// countPlanLines returns the number of plan lines pattern matches outside
// fenced blocks -- taskLinePattern for the task count, correctionPattern
// for the plan-correction count. Fence state is tracked across the whole
// scan exactly as tickTask tracks it, so a line shown inside a fenced
// example is documentation, not structure -- the same rule planparse.go's
// task iteration applies to the same file.
func countPlanLines(plan []byte, pattern *regexp.Regexp) int {
	count := 0
	inFence := false
	for _, line := range bytes.Split(plan, []byte("\n")) {
		if fencePattern.Match(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if pattern.Match(line) {
			count++
		}
	}
	return count
}
