package guard

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// check-task-build-green fails when a tasks.md's tasks carry no **Build:**
// tag, or a `red` tag whose task has no **Squash-with:** field, an empty one,
// one naming an absent partner, or one naming a partner that is itself `red`.
//
// Rule (canonical definition: skills/flow-contracts/build-green.md — do not
// restate it here; a second copy is a Single Source of Truth violation, the
// same class of drift planprovenance.go's ported docstring warns against).
//
// Arguments: one tasks.md path checks that file; none scans every live
// change's tasks.md under <root>/<spec leaf>/changes/*/tasks.md, and that of
// every archived change the base does not carry yet (unlanded-archives, run
// in-process), aggregating exit codes — 2 beats 1 beats 0 — and printing each
// file's own violations. <root> is CHECK_TASK_BUILD_GREEN_ROOT when set (a
// harness's sandboxed fixture tree), else FLOW_GUARD_REPO_ROOT, which
// scripts/check-task-build-green.sh exports as its own real parent's parent
// so the scan works from any cwd and through a skill directory's symlink.
//
// Exit codes:
//
//	0  clean — every task in the file carries a resolvable **Build:** tag
//	   (including a file with zero tasks).
//	1  violations found — one or more of: a duplicate task id, a task with
//	   no tag, a malformed tag, a `red` task with no **Squash-with:** field
//	   at all, a **Squash-with:** field naming zero ids, a named partner that
//	   does not exist in this file, or a named partner that is itself `red`.
//	   Printed on stdout one per line as `file:line: message`.
//	2  invocation error — wrong argument count, or the file cannot be read
//	   (a file that is not UTF-8 included: the Python let that decode error
//	   escape as a traceback and exit 1; the header named it exit 2, and the
//	   header wins).
//
// Every piece of tasks.md grammar this guard reads is planparse.go's —
// tcfIterTasks, tcfSelectBuildTag, tcfSelectSquashWith, tcfUnclosedFence —
// which check-task-commit-fields reads with too; nothing about the shape of
// a tasks.md is spelled out in this file.
//
// Fenced code blocks are opaque: a task line, a **Build:** line or a
// **Squash-with:** line inside a fence opens no task, closes no body, and is
// never read as a real tag or field — a tasks.md is free to show worked
// examples of the tag grammar. A fence a task body OPENS and never CLOSES is
// itself a violation (0b below, fix round 11, F22): the fields under it are
// code and correctly not read, but reporting the task as untagged would name
// the consequence and hide the cause.
//
// Parsing model: a task begins at a non-fenced `- [ ] <id>. ` or `- [x]
// <id>. ` line whose id is a flat integer (a dotted `1.1` line is no task; a
// dotted Squash-with partner names no task and is reported as absent). Its
// body runs to the next task line, the next level-2 or level-3 heading, or
// end of file, its indented step checkboxes included. The FIRST non-fenced
// **Build:** line is the tag, its kind the `green`/`red` keyword opening the
// value after whitespace, anything after it ignored (fix round 9, F20;
// KAN-394). A first **Build:** line opening with neither keyword is a
// MALFORMED tag, reported with its line and value — never folded into "no
// tag", and never overridden by a well-formed line further down. The first
// non-fenced **Squash-with:** line whose value gates as `Task <id>[,
// <id>...]` is the field, line-scoped; a non-gating candidate ahead of it is
// skipped, and a body with only non-gating candidates has no field at all
// (fix round 6, F13; fix round 8, F19).
//
// Validation, over every task in document order:
//
//  0. A duplicate id → `<file>:<this task line>: task <id> is defined
//     more than once (first at line <first>)`, for every occurrence after
//     the first; all duplicates are reported before any other finding.
//     0b. An unclosed fence → `<file>:<fence line>: task <id> opens a code
//     fence here that is never closed in its body, ...`, INSTEAD of 1-5.
//  1. No **Build:** tag → `<file>:<task line>: task <id> has no
//     **Build:** tag`.
//  2. `red` with no Squash-with field → at the tag line.
//  3. `red` with a field naming zero ids → at the field line.
//  4. A partner that is no task in this file → at the field line.
//  5. A partner that is itself `red` → at the field line, reported once
//     from the task carrying the field.
//  6. A `red` task whose partner is `green`, or one nothing references, is
//     no violation.
//
// A partner named twice in one field is validated once (the list is
// de-duplicated, order preserving), and every id lookup resolves to the
// first task carrying it.
func init() { Registry["check-task-build-green"] = checkTaskBuildGreen }

func checkTaskBuildGreen(args []string, env Env, stdout, stderr io.Writer) int {
	switch len(args) {
	case 1:
		return tbgCheckFile(args[0], env, stdout, stderr)
	case 0:
	default:
		fmt.Fprintln(stderr, "usage: check-task-build-green.sh [path-to-tasks.md]")
		return 2
	}
	root := env.Getenv("CHECK_TASK_BUILD_GREEN_ROOT")
	if root == "" {
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintln(stderr, "check-task-build-green: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-task-build-green.sh, which sets it")
			return 2
		}
	}
	files, ok := tcfScanPlans(env, root, stderr)
	if !ok {
		return 2
	}
	status := 0
	for _, f := range files {
		status = psWorse(status, tbgCheckFile(f, env, stdout, stderr))
	}
	return status
}

func tbgCheckFile(path string, env Env, stdout, stderr io.Writer) int {
	b, err := os.ReadFile(psAbs(env, path))
	var lines []string
	if err == nil {
		lines, err = tcfPlanLines(b)
	} else if pe := (*fs.PathError)(nil); errors.As(err, &pe) {
		pe.Path = path // str(OSError) names the path as given
		err = tcfOSError(path, pe)
	}
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", path, err)
		return 2
	}
	violations := tbgCheck(lines, path)
	for _, v := range violations {
		fmt.Fprintln(stdout, v)
	}
	if len(violations) > 0 {
		return 1
	}
	return 0
}

// tbgTask is one task's resolved tag and field; lines are 1-based, 0 when
// absent.
type tbgTask struct {
	id                                   string
	taskLine, tagLine, squashLine, fence int
	kind, value                          string
	partners                             []string
}

func tbgCheck(lines []string, file string) []string {
	var tasks []tbgTask
	for _, found := range tcfIterTasks(lines) {
		task := tbgTask{id: found.id, taskLine: found.taskLine}
		if tag := tcfSelectBuildTag(found.lines); tag != nil {
			task.tagLine, task.kind, task.value = found.taskLine+tag.offset+1, tag.kind, tag.value
		}
		// A value that does not gate is not a Squash-with field at all — a
		// red task carrying one is reported as having no field, not as
		// having an empty one.
		if _, partners, off := tcfSelectSquashWith(found.lines); off >= 0 {
			task.squashLine, task.partners = found.taskLine+off+1, partners
		}
		if off := tcfUnclosedFence(found.lines); off >= 0 {
			task.fence = found.taskLine + off + 1
		}
		tasks = append(tasks, task)
	}

	// The full index, not tcfSelectTask (fix round 11, F23): every id for
	// one-hit partner lookups, and every duplicate occurrence so each one
	// after the first is reported.
	byID := map[string]*tbgTask{}
	var violations []string
	for i := range tasks {
		task := &tasks[i]
		if first, ok := byID[task.id]; ok {
			violations = append(violations, fmt.Sprintf("%s:%d: task %s is defined more than once (first at line %d)", file, task.taskLine, task.id, first.taskLine))
		} else {
			byID[task.id] = task
		}
	}
	at := func(line int, task tbgTask, msg string, a ...any) {
		violations = append(violations, fmt.Sprintf("%s:%d: task %s ", file, line, task.id)+fmt.Sprintf(msg, a...))
	}
	for _, task := range tasks {
		switch {
		case task.fence > 0:
			at(task.fence, task, "opens a code fence here that is never closed in its body, so every field below it is inside the fence and was not read")
		case task.tagLine > 0 && task.kind == "":
			at(task.tagLine, task, "has a **Build:** line reading \"%s\", which is neither green nor red", task.value)
		case task.kind == "":
			at(task.taskLine, task, "has no **Build:** tag")
		case task.kind != "red":
		case task.squashLine == 0:
			at(task.tagLine, task, "is red with no **Squash-with:** field")
		case len(task.partners) == 0:
			at(task.squashLine, task, "is red with no merge partner named")
		default:
			seen := map[string]bool{}
			for _, partner := range task.partners {
				if seen[partner] {
					continue
				}
				seen[partner] = true
				if p, ok := byID[partner]; !ok {
					at(task.squashLine, task, "merges with Task %s, which does not exist in this plan", partner)
				} else if p.kind == "red" {
					at(task.squashLine, task, "merges with Task %s, which is itself red", partner)
				}
			}
		}
	}
	return violations
}
