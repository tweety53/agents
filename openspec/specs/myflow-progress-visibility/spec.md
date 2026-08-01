# myflow-progress-visibility Specification

## Purpose
TBD - created by archiving change kan-26-operator-output-and-configuration. Update Purpose after archive.
## Requirements
### Requirement: Every pipeline command drives the harness's task-list mechanism

`/myflow-start`, `/myflow-do` and `/myflow-finish` SHALL register their steps with the harness's
task-list mechanism at the start of a run, and SHALL keep each entry's status current as the run
proceeds, so the harness's live progress view — a task count line and one line per task — renders
throughout.

The steps registered SHALL be:

| Command | One entry per |
|---------|---------------|
| `/myflow-start` | its brainstorming checklist item and each artifact it produces |
| `/myflow-do` | each item in `tasks.md`, in plan order |
| `/myflow-finish` | each step of the run it is performing |

`/myflow-status` and `/myflow-info` are read-only and SHALL register nothing. Registering steps for
a report would put entries on the operator's task list for work nobody is doing.

An entry SHALL move to in-progress when its step begins and to completed when that step finishes, so
the count line distinguishes done, in progress and open at every point in the run.

#### Scenario: A plan's tasks appear as the run starts

- **WHEN** `/myflow-do` begins a run for a change whose `tasks.md` carries four items
- **THEN** four entries are registered before implementation starts
- **AND** the harness's progress view reports four tasks with none done

#### Scenario: Status changes as work proceeds

- **WHEN** an implementer is dispatched for the second task
- **THEN** that task's entry is in progress while it runs
- **AND** it is completed once the task passes its spec and quality review

#### Scenario: A read-only command registers nothing

- **WHEN** `/myflow-status` runs
- **THEN** no task entries are created or updated

### Requirement: The progress view is a view, never a record

The harness's task list SHALL NOT be read back by any command, guard or contract as evidence of what
was done. `tasks.md` SHALL remain the single source of truth for a plan's completion state.

No third checkbox marker SHALL be introduced into `tasks.md` to carry an in-progress state. The
in-progress count SHALL come from the harness's task list alone, which no run persists.

`scripts/check-unfinished-work.sh` reads `tasks.md`, and a second source of completion state would be
one that guard cannot see. A marker written into `tasks.md` at dispatch and resolved at completion
would additionally survive a crashed run as a permanently in-progress task, in a file two guards
parse.

#### Scenario: Completion is read from the plan, not the widget

- **WHEN** the finish gate determines whether any plan item is outstanding
- **THEN** it reads `tasks.md`
- **AND** the harness's task list is not consulted

#### Scenario: The plan file gains no new marker

- **WHEN** `/myflow-do` dispatches an implementer for a task
- **THEN** that task's line in `tasks.md` still reads `- [ ]` until the task completes
- **AND** no third marker is written to represent the in-progress state

### Requirement: A harness with no task mechanism prints the equivalent block

Where the harness offers no task-list mechanism, a pipeline command SHALL print the equivalent block
in its output instead: a count line naming how many tasks are done, in progress and open, followed by
one line per task marked as done or not done.

The requirement SHALL be satisfied by whichever mechanism the harness provides, and SHALL NOT name
one harness's tool as the only means. myflow runs in Claude Code, Cursor and Codex, and a requirement
written against one harness's API would be unimplementable in the other two.

#### Scenario: A harness without the mechanism still shows progress

- **WHEN** a pipeline command runs in a harness that offers no task-list mechanism
- **THEN** it prints a count line and one line per task
- **AND** the requirement is met without that harness gaining a task tool

#### Scenario: The mechanism is not named as a single tool

- **WHEN** an implementer reads this requirement to satisfy it in a new harness
- **THEN** it directs them to that harness's own task-list mechanism
- **AND** it does not require the tool name of any particular harness
