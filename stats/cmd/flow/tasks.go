// Command flow's tasks.go implements `flow tasks tick`, the CLI surface
// over tickTask (tick.go, task 1). Unlike state.go/stage.go/settings.go,
// this subcommand never talks to flowd's HTTP store: ticking a task is
// a filesystem edit inside a change's own worktree, so -C names that
// worktree's root directly rather than a point git resolves a project
// key from.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const tasksUsage = `usage: flow tasks tick [-C dir] <change> <task-id>

tasks tick flips one task's own checkbox and every step checkbox in its
body from "[ ]" to "[x]" in <dir>/<spec-root>/changes/<change>/tasks.md,
where <spec-root> is "spectre" when <dir>/spectre/changes exists,
"openspec" when only <dir>/openspec/changes does, and "spectre"
otherwise. It refuses (exit 1) rather than silently no-opping when the
task is already ticked, or when no task with that id exists in the
plan. <task-id> must be a bare non-negative integer (exit 2 otherwise).
`

func runTasks(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, tasksUsage)
		return 2
	}
	switch args[0] {
	case "tick":
		return runTasksTick(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "flow: unknown tasks command %q\n", args[0])
		fmt.Fprint(stderr, tasksUsage)
		return 2
	}
}

// specRootLeaf ports scripts/lib/spec-root.sh's spec_root_leaf probe:
// "spectre" when <dir>/spectre/changes exists, "openspec" when only
// <dir>/openspec/changes does, "spectre" when neither does (a project
// with no open changes yet is expected to gain the canonical tree, so
// callers already have their own "no plan found" message for that
// case rather than needing an error here).
//
// Deliberately omitted: the shell version's stderr diagnostic for the
// "both spectre/changes and openspec/changes exist" half-migration
// case. specRootLeaf is a pure function with no writer to log to
// today, and adding one is disproportionate to the value -- a judged
// tradeoff, not an oversight, so a future reader doesn't need to
// rediscover the question.
func specRootLeaf(dir string) string {
	if info, err := os.Stat(filepath.Join(dir, "spectre", "changes")); err == nil && info.IsDir() {
		return "spectre"
	}
	if info, err := os.Stat(filepath.Join(dir, "openspec", "changes")); err == nil && info.IsDir() {
		return "openspec"
	}
	return "spectre"
}

// changeNamePattern is the same containment allowlist
// scripts/lib/change-plan.sh's _change_plan_name_ok applies to a change
// name before concatenating it into a path: start with a letter or
// digit, then only letters, digits, '.', '_' and '-' -- so a value like
// "../escape" can never walk planPath outside <dir>/<spec-root>/changes/.
var changeNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// planPath returns <dir>/<spec-root>/changes/<change>/tasks.md, or an
// error when change fails changeNamePattern. It resolves only the
// direct path -- no link.md/peer indirection the way
// scripts/lib/change-plan.sh's change_plan_path offers guards checking
// a different worktree's plan from outside it: flow tasks tick only
// ever runs inside the worktree that owns the plan it is ticking
// (design.md's tasks-tick-no-peer-indirection).
func planPath(dir, change string) (string, error) {
	if !changeNamePattern.MatchString(change) {
		return "", fmt.Errorf("change name %q must start with a letter or digit and contain only letters, digits, '.', '_' and '-'", change)
	}
	return filepath.Join(dir, specRootLeaf(dir), "changes", change, "tasks.md"), nil
}

func runTasksTick(_ context.Context, args []string, stdout, stderr io.Writer) int {
	fset := flag.NewFlagSet("flow tasks tick", flag.ContinueOnError)
	fset.SetOutput(stderr)
	dir := fset.String("C", "", "resolve the plan as if run from this directory (default: cwd)")
	if err := fset.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprint(stderr, tasksUsage)
		return 2
	}
	if fset.NArg() != 2 {
		fmt.Fprintln(stderr, "flow: expected exactly two arguments, the change name and the task id")
		fmt.Fprint(stderr, tasksUsage)
		return 2
	}
	change, taskID := fset.Arg(0), fset.Arg(1)

	resolvedDir := *dir
	if resolvedDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "flow: resolve working directory: %v\n", err)
			return 1
		}
		resolvedDir = wd
	}

	path, err := planPath(resolvedDir, change)
	if err != nil {
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 2
	}

	plan, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "flow: read %s: %v\n", path, err)
		return 1
	}

	out, steps, err := tickTask(plan, taskID)
	if err != nil {
		if errors.Is(err, ErrTaskIDInvalid) {
			fmt.Fprintf(stderr, "flow: %v\n", err)
			return 2
		}
		fmt.Fprintf(stderr, "flow: %v\n", err)
		return 1
	}

	if err := os.WriteFile(path, out, 0o644); err != nil {
		fmt.Fprintf(stderr, "flow: write %s: %v\n", path, err)
		return 1
	}

	fmt.Fprintf(stdout, "ticked task %s (%d steps)\n", taskID, steps)
	return 0
}
