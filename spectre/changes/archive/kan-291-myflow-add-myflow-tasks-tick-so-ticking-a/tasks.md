# kan-291-myflow-add-myflow-tasks-tick-so-ticking-a

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

Two tasks on one branch: the pure tick logic (task 1), then the CLI subcommand that wires it up
(task 2). `design.md` is canonical for every decision below.

**Baseline, measured before any edit:**

- `stats/cmd/flow` carries 88 test functions.
  <!-- measured: go test -list '.*' ./cmd/flow/... @ bacc2f4 -->

- [x] 1. Tick logic — flip a task's own checkbox and its steps' checkboxes

  Add `stats/cmd/flow/tick.go`, `package main`, no new imports beyond the standard library
  (`bytes`, `errors`, `fmt`, `regexp`).

  1. Write the failing tests first, in a new `stats/cmd/flow/tick_test.go`:

  ```go verified:matches this package's existing table-driven test style (state_test.go)
  package main

  import (
  	"errors"
  	"strings"
  	"testing"
  )

  func TestTickTaskFlipsOwnCheckboxAndSteps(t *testing.T) {
  	plan := "# c\n\n- [ ] 1. First\n\n  - [ ] **Step 1: do a thing**\n\n  - [ ] **Step 2: do another**\n\n**Files:** `x`\n"
  	out, steps, err := tickTask([]byte(plan), "1")
  	if err != nil {
  		t.Fatalf("tickTask: %v", err)
  	}
  	if steps != 2 {
  		t.Fatalf("steps = %d, want 2", steps)
  	}
  	got := string(out)
  	if !strings.Contains(got, "- [x] 1. First") {
  		t.Errorf("task line not ticked:\n%s", got)
  	}
  	if strings.Contains(got, "- [ ] **Step") {
  		t.Errorf("a step was left unticked:\n%s", got)
  	}
  	if !strings.Contains(got, "**Files:** `x`") {
  		t.Errorf("field line was disturbed:\n%s", got)
  	}
  }

  func TestTickTaskZeroSteps(t *testing.T) {
  	plan := "- [ ] 1. First\n\nno steps here.\n\n**Files:** `x`\n"
  	out, steps, err := tickTask([]byte(plan), "1")
  	if err != nil {
  		t.Fatalf("tickTask: %v", err)
  	}
  	if steps != 0 {
  		t.Fatalf("steps = %d, want 0", steps)
  	}
  	if !strings.Contains(string(out), "- [x] 1. First") {
  		t.Errorf("task line not ticked:\n%s", string(out))
  	}
  }

  func TestTickTaskStopsAtNextTaskLine(t *testing.T) {
  	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n- [ ] 2. Second\n\n  - [ ] **Step 1: b**\n"
  	out, steps, err := tickTask([]byte(plan), "1")
  	if err != nil {
  		t.Fatalf("tickTask: %v", err)
  	}
  	if steps != 1 {
  		t.Fatalf("steps = %d, want 1 (task 2's step must not be touched)", steps)
  	}
  	got := string(out)
  	if !strings.Contains(got, "- [x] 1. First") {
  		t.Errorf("task 1 not ticked:\n%s", got)
  	}
  	if strings.Contains(got, "- [x] 2. Second") {
  		t.Errorf("task 2 was ticked, must not be:\n%s", got)
  	}
  	if !strings.Contains(got, "- [ ] **Step 1: b**") {
  		t.Errorf("task 2's step was flipped, must not be:\n%s", got)
  	}
  }

  func TestTickTaskStopsAtNextHeading(t *testing.T) {
  	plan := "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n## Part 2\n\n- [ ] 2. Second\n\n  - [ ] **Step 1: b**\n"
  	_, steps, err := tickTask([]byte(plan), "1")
  	if err != nil {
  		t.Fatalf("tickTask: %v", err)
  	}
  	if steps != 1 {
  		t.Fatalf("steps = %d, want 1 (must stop at the ## heading)", steps)
  	}
  }

  func TestTickTaskAlreadyTickedRefuses(t *testing.T) {
  	plan := "- [x] 1. First\n\n  - [ ] **Step 1: a**\n"
  	_, _, err := tickTask([]byte(plan), "1")
  	if !errors.Is(err, ErrTaskAlreadyTicked) {
  		t.Fatalf("err = %v, want ErrTaskAlreadyTicked", err)
  	}
  }

  func TestTickTaskNotFound(t *testing.T) {
  	plan := "- [ ] 1. First\n"
  	_, _, err := tickTask([]byte(plan), "2")
  	if !errors.Is(err, ErrTaskNotFound) {
  		t.Fatalf("err = %v, want ErrTaskNotFound", err)
  	}
  }

  func TestTickTaskInvalidID(t *testing.T) {
  	plan := "- [ ] 1. First\n"
  	for _, id := range []string{"1.1", "one", "-1", "", "1 "} {
  		if _, _, err := tickTask([]byte(plan), id); !errors.Is(err, ErrTaskIDInvalid) {
  			t.Errorf("id %q: err = %v, want ErrTaskIDInvalid", id, err)
  		}
  	}
  }

  func TestTickTaskLeavesRestByteIdentical(t *testing.T) {
  	plan := "# Title\n\nprose before.\n\n- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n- [ ] 2. Second\n\nprose after.\n"
  	out, _, err := tickTask([]byte(plan), "1")
  	if err != nil {
  		t.Fatalf("tickTask: %v", err)
  	}
  	want := "# Title\n\nprose before.\n\n- [x] 1. First\n\n  - [x] **Step 1: a**\n\n- [ ] 2. Second\n\nprose after.\n"
  	if string(out) != want {
  		t.Fatalf("out =\n%s\nwant\n%s", out, want)
  	}
  }
  ```

  2. Run `cd stats && go test ./cmd/flow/... -run TestTickTask -v` to confirm they fail with
     `undefined: tickTask` — `tick.go` does not exist yet.

  3. Write `stats/cmd/flow/tick.go`:

  ```go unverified:confirm the regexes below match every existing tasks.md in this repository's spectre/changes/ tree before merging
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
  // following task line does.
  var headingPattern = regexp.MustCompile(`^(## |### )`)

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
  // already applies it when scanning the same file.
  func tickTask(plan []byte, taskID string) ([]byte, int, error) {
  	if !taskIDPattern.MatchString(taskID) {
  		return nil, 0, fmt.Errorf("%w: %q", ErrTaskIDInvalid, taskID)
  	}

  	lines := bytes.Split(plan, []byte("\n"))

  	start := -1
  	alreadyTicked := false
  	for i, line := range lines {
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
  	for i := start + 1; i < len(lines); i++ {
  		if taskLinePattern.Match(lines[i]) || headingPattern.Match(lines[i]) {
  			end = i
  			break
  		}
  	}

  	lines[start] = append([]byte("- [x] "), lines[start][len("- [ ] "):]...)

  	steps := 0
  	for i := start + 1; i < end; i++ {
  		if stepCheckboxPattern.Match(lines[i]) {
  			lines[i] = append([]byte("  - [x] "), lines[i][len("  - [ ] "):]...)
  			steps++
  		}
  	}

  	return bytes.Join(lines, []byte("\n")), steps, nil
  }
  ```

  4. Run `cd stats && go test ./cmd/flow/... -run TestTickTask -v` to confirm all 7 subtests pass.

  5. Run `cd stats && go vet ./cmd/flow/... && gofmt -l cmd/flow/tick.go cmd/flow/tick_test.go && go
     test ./cmd/flow/...` to confirm `gofmt -l` prints nothing and the whole package passes (96
     tests total).

  6. Commit:

  ```bash verified:authored in-tree for this change
  git add stats/cmd/flow/tick.go stats/cmd/flow/tick_test.go
  git commit -m "feat(flow): add tickTask, the core logic for flipping a task's checkboxes"
  ```

**Files:** `stats/cmd/flow/tick.go`
**Test:** `stats/cmd/flow/tick_test.go`
**Tests:** `TestTickTaskFlipsOwnCheckboxAndSteps`, `TestTickTaskZeroSteps`,
`TestTickTaskStopsAtNextTaskLine`, `TestTickTaskStopsAtNextHeading`, `TestTickTaskAlreadyTickedRefuses`,
`TestTickTaskNotFound`, `TestTickTaskInvalidID`, `TestTickTaskLeavesRestByteIdentical`
**Regression:** reverting this commit removes `tickTask` and its sentinel errors entirely, so every
test above fails to compile (`undefined: tickTask`) rather than merely failing an assertion.
**Baseline:** before=88 after=96 tests in `stats/cmd/flow`
<!-- measured: go test -list '.*' ./cmd/flow/... @ bacc2f4 -->
**Commit:** `feat(flow): add tickTask, the core logic for flipping a task's checkboxes`
**Build:** green

- [x] 2. `flow tasks tick` — the CLI subcommand

  Add `stats/cmd/flow/tasks.go`, `package main`, wiring `tickTask` (task 1) up to a `-C dir`
  argument and the on-disk plan file, and register it in `main.go`.

  1. Write the failing tests first, in a new `stats/cmd/flow/tasks_test.go`:

  ```go verified:matches this package's existing CLI-level test style (state_test.go's use of run())
  package main

  import (
  	"bytes"
  	"context"
  	"os"
  	"path/filepath"
  	"strings"
  	"testing"
  )

  func writePlan(t *testing.T, dir, changesLeaf, change, body string) {
  	t.Helper()
  	planDir := filepath.Join(dir, changesLeaf, "changes", change)
  	if err := os.MkdirAll(planDir, 0o755); err != nil {
  		t.Fatalf("MkdirAll: %v", err)
  	}
  	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(body), 0o644); err != nil {
  		t.Fatalf("WriteFile: %v", err)
  	}
  }

  func TestTasksTickWritesFile(t *testing.T) {
  	dir := t.TempDir()
  	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n")

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1"}, nil, &stdout, &stderr)
  	if code != 0 {
  		t.Fatalf("code = %d, stderr = %s", code, stderr.String())
  	}
  	if !strings.Contains(stdout.String(), "ticked task 1 (1 steps)") {
  		t.Errorf("stdout = %q", stdout.String())
  	}
  	got, err := os.ReadFile(filepath.Join(dir, "spectre", "changes", "c", "tasks.md"))
  	if err != nil {
  		t.Fatalf("ReadFile: %v", err)
  	}
  	if !strings.Contains(string(got), "- [x] 1. First") || !strings.Contains(string(got), "- [x] **Step 1: a**") {
  		t.Errorf("file not ticked:\n%s", got)
  	}
  }

  func TestTasksTickAlreadyTickedExitsNonZero(t *testing.T) {
  	dir := t.TempDir()
  	writePlan(t, dir, "spectre", "c", "- [x] 1. First\n")

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1"}, nil, &stdout, &stderr)
  	if code == 0 {
  		t.Fatalf("code = 0, want non-zero; stderr = %s", stderr.String())
  	}
  	if !strings.Contains(stderr.String(), "already ticked") {
  		t.Errorf("stderr = %q", stderr.String())
  	}
  }

  func TestTasksTickTaskNotFoundExitsNonZero(t *testing.T) {
  	dir := t.TempDir()
  	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n")

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "9"}, nil, &stdout, &stderr)
  	if code == 0 {
  		t.Fatalf("code = 0, want non-zero; stderr = %s", stderr.String())
  	}
  }

  func TestTasksTickInvalidIDExitsTwo(t *testing.T) {
  	dir := t.TempDir()
  	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n")

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1.1"}, nil, &stdout, &stderr)
  	if code != 2 {
  		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
  	}
  }

  func TestTasksTickMissingPlanExitsNonZero(t *testing.T) {
  	dir := t.TempDir()

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1"}, nil, &stdout, &stderr)
  	if code == 0 {
  		t.Fatalf("code = 0, want non-zero; stderr = %s", stderr.String())
  	}
  }

  func TestTasksTickBadArgCountExitsTwo(t *testing.T) {
  	dir := t.TempDir()
  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c"}, nil, &stdout, &stderr)
  	if code != 2 {
  		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
  	}
  }

  func TestSpecRootLeafPrefersSpectre(t *testing.T) {
  	dir := t.TempDir()
  	if err := os.MkdirAll(filepath.Join(dir, "spectre", "changes"), 0o755); err != nil {
  		t.Fatal(err)
  	}
  	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
  		t.Fatal(err)
  	}
  	if got := specRootLeaf(dir); got != "spectre" {
  		t.Errorf("specRootLeaf = %q, want spectre", got)
  	}
  }

  func TestSpecRootLeafFallsBackToOpenspec(t *testing.T) {
  	dir := t.TempDir()
  	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
  		t.Fatal(err)
  	}
  	if got := specRootLeaf(dir); got != "openspec" {
  		t.Errorf("specRootLeaf = %q, want openspec", got)
  	}
  }

  func TestSpecRootLeafDefaultsToSpectreWhenNeitherExists(t *testing.T) {
  	dir := t.TempDir()
  	if got := specRootLeaf(dir); got != "spectre" {
  		t.Errorf("specRootLeaf = %q, want spectre", got)
  	}
  }
  ```

  2. Run `cd stats && go test ./cmd/flow/... -run 'TestTasksTick|TestSpecRootLeaf' -v` to confirm
     they fail to compile — `run` does not recognise `"tasks"`, and `specRootLeaf` is undefined.

  3. Write `stats/cmd/flow/tasks.go`:

  ```go verified:the -C flag, its cwd default and its usage string follow parseStateFlags in state.go exactly
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
  ```

     In `main.go`: add `case "tasks": return runTasks(ctx, args[1:], stdout, stderr)` to `run`'s
     switch, and a `tasks tick <change> <task-id>  flip a task's checkbox and its steps'
     checkboxes` line to the `usage` const, alongside the existing `state`/`stage`/`record`/
     `journal`/`settings` lines.

  4. Run `cd stats && go test ./cmd/flow/... -run 'TestTasksTick|TestSpecRootLeaf' -v` to confirm
     all 9 subtests pass.

  5. Run `cd stats && go vet ./cmd/flow/... && gofmt -l cmd/flow/tasks.go cmd/flow/tasks_test.go
     cmd/flow/main.go && go test ./cmd/flow/...` to confirm `gofmt -l` prints nothing and the whole
     package passes (105 tests total).
     <!-- predicted: go test -list '.*' ./cmd/flow/... after task 2 -->

  6. Commit:

  ```bash verified:authored in-tree for this change
  git add stats/cmd/flow/tasks.go stats/cmd/flow/tasks_test.go stats/cmd/flow/main.go
  git commit -m "feat(flow): add the tasks tick CLI subcommand"
  ```

**Files:** `stats/cmd/flow/tasks.go`, `stats/cmd/flow/main.go`
**Test:** `stats/cmd/flow/tasks_test.go`
**Tests:** `TestTasksTickWritesFile`, `TestTasksTickAlreadyTickedExitsNonZero`,
`TestTasksTickTaskNotFoundExitsNonZero`, `TestTasksTickInvalidIDExitsTwo`,
`TestTasksTickMissingPlanExitsNonZero`, `TestTasksTickBadArgCountExitsTwo`,
`TestSpecRootLeafPrefersSpectre`, `TestSpecRootLeafFallsBackToOpenspec`,
`TestSpecRootLeafDefaultsToSpectreWhenNeitherExists`
**Regression:** reverting this commit removes the `tasks` case from `main.go`'s switch, so `flow
tasks tick` fails with `unknown command "tasks"` instead of ticking anything — every test above
either fails to compile or gets exit 2 from the unknown-command path.
**Baseline:** before=96 after=105 tests in `stats/cmd/flow`
<!-- predicted: go test -list '.*' ./cmd/flow/... after task 2 -->
**Commit:** `feat(flow): add the tasks tick CLI subcommand`
**Build:** green
