## MODIFIED Requirements

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

`/myflow-status` is read-only and SHALL register nothing. Registering steps for a report would put
entries on the operator's task list for work nobody is doing.

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
