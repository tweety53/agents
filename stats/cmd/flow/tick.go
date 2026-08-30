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
// following task line does. Mirrors scripts/lib/plan_grammar.py's
// BODY_BOUNDARY_RE exactly: a space OR end-of-line after the "#"s, so a
// bare "##"/"###" with nothing following it still closes a body.
var headingPattern = regexp.MustCompile(`^#{2,3}($|\s)`)

// fencePattern matches a fenced code block delimiter: three or more
// backticks or tildes, with up to three leading spaces of indent.
// Mirrors scripts/lib/plan_grammar.py's FENCE_RE exactly.
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
// Placement rule, applied here exactly as check-task-build-green.py
// already applies it when scanning the same file. Fence state is
// tracked across the whole scan (mirroring scripts/lib/plan_grammar.py's
// iter_tasks), so a task line, heading or step checkbox shown inside a
// fenced example opens no task, closes no body and is never ticked --
// it is documentation, not structure.
func tickTask(plan []byte, taskID string) ([]byte, int, error) {
	if !taskIDPattern.MatchString(taskID) {
		return nil, 0, fmt.Errorf("%w: %q", ErrTaskIDInvalid, taskID)
	}

	lines := bytes.Split(plan, []byte("\n"))

	start := -1
	alreadyTicked := false
	inFence := false
	for i, line := range lines {
		if fencePattern.Match(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		m := taskLinePattern.FindSubmatch(line)
		if m == nil || string(m[2]) != taskID {
			continue
		}
		start = i
		alreadyTicked = string(m[1]) == "x"
		break
	}
	if start == -1 {
		return nil, 0, fmt.Errorf("%w: %s", ErrTaskNotFound, taskID)
	}
	if alreadyTicked {
		return nil, 0, fmt.Errorf("%w: %s", ErrTaskAlreadyTicked, taskID)
	}

	end := len(lines)
	inFence = false
	for i := start + 1; i < len(lines); i++ {
		line := lines[i]
		if fencePattern.Match(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if taskLinePattern.Match(line) || headingPattern.Match(line) {
			end = i
			break
		}
	}

	lines[start] = append([]byte("- [x] "), lines[start][len("- [ ] "):]...)

	steps := 0
	inFence = false
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
