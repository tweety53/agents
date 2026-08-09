# myflow-fast-command Specification

## Purpose
TBD - created by archiving change kan-111-myflow-fast. Update Purpose after archive.
## Requirements
### Requirement: `/myflow-fast` accepts no state or `IN_PROGRESS`

`/myflow-fast` SHALL accept no state (creating a change) or `IN_PROGRESS`. On any other state it
SHALL emit the same wrong-state handoff every other pipeline command emits, naming the actual
state, the states `/myflow-fast` expects, and asking for an explicit override before proceeding.

#### Scenario: A mismatched invocation does not advance the state

- **WHEN** `/myflow-fast` is invoked for a change at `STARTED` with no worktree yet created
- **THEN** it reports the mismatch, offers the explicit-override choice, and writes no state
  unless the operator overrides

### Requirement: The first invocation chains brainstorming into implementation

`/myflow-fast`, invoked with no existing state file, SHALL run `/myflow-start`'s brainstorming
stage in full — the same interactive, one-question-at-a-time dialogue, ending with the operator
approving the design — exactly as `/myflow-start` runs it, with no auto-answering.

Once brainstorming converges and the OpenSpec artifacts are created, `/myflow-fast` SHALL continue,
within the same invocation and without requiring a further command from the operator, into
`/myflow-do`'s SDD+TDD implementation and review panel, exactly as `/myflow-do` runs them. The run
SHALL end at `IN_PROGRESS` with the same staged-diff-plus-run-instructions handoff `/myflow-do`
produces.

`/myflow-fast` SHALL skip publishing the proposal HTML artifact that `/myflow-start` publishes,
because the operator who answers the brainstorming questions is present in the same session that
produces the design.

#### Scenario: One invocation reaches IN_PROGRESS from nothing

- **WHEN** `/myflow-fast <description>` is invoked for a change with no state file
- **THEN** the operator is asked brainstorming's design questions interactively within that same
  run, and once approved, implementation and the review panel run unattended within the same run,
  ending at `IN_PROGRESS` with no additional command required in between

#### Scenario: No proposal artifact is published

- **WHEN** `/myflow-fast` completes the first invocation for a change
- **THEN** the state file's `artifactUrl` is `null` and no HTML artifact was published

### Requirement: Re-invoking at `IN_PROGRESS` disambiguates fix from integrate by argument presence

`/myflow-fast`, invoked for a change at `IN_PROGRESS` with an argument, SHALL treat that argument as
fix instructions and resume the worktree exactly as re-running `/myflow-do` at `IN_PROGRESS` does
today, leaving the state unchanged.

`/myflow-fast`, invoked bare (no argument) for a change at `IN_PROGRESS`, SHALL proceed to the
integrate question — open PR, merge and push, or manual — exactly as `/myflow-finish` run 1 asks
it.

#### Scenario: An argument at IN_PROGRESS is a fix

- **WHEN** `/myflow-fast <fix instructions>` is invoked for a change at `IN_PROGRESS`
- **THEN** the worktree is resumed with those instructions as fix guidance, and the state file
  still reads `IN_PROGRESS` afterward

#### Scenario: A bare call at IN_PROGRESS proceeds to integrate

- **WHEN** `/myflow-fast` is invoked with no argument for a change at `IN_PROGRESS`
- **THEN** the operator is asked the open-PR / merge-and-push / manual question

### Requirement: Merge-and-push chains straight through to archive

When the operator chooses merge-and-push at the integrate question, `/myflow-fast` SHALL run
`/myflow-finish`'s run 1 (merge, push) and, within the same invocation and without a further
command from the operator, continue into run 2 (archive, delta-spec sync, worktree and branch
removal, Jira transition), ending at `FINISHED`.

When the operator chooses open PR or manual, `/myflow-fast` SHALL stop after run 1 completes,
printing the same handoff `/myflow-finish` run 1 prints, because each of those routes requires an
action outside this command's control (an external merge, or the operator's own manual steps)
before archiving can happen.

#### Scenario: Merge-and-push reaches FINISHED in one invocation

- **WHEN** the operator chooses merge-and-push at the integrate question
- **THEN** the same `/myflow-fast` invocation merges, pushes, archives, syncs delta specs, removes
  the worktree and branch, and ends at `FINISHED`, with no further command needed

#### Scenario: Open PR stops and hands off

- **WHEN** the operator chooses open PR at the integrate question
- **THEN** `/myflow-fast` opens the PR, prints its URL, and stops at `IN_PROGRESS`; the next bare
  `/myflow-fast` call, once the PR is merged, runs archive (run 2)

#### Scenario: Manual stops and hands off

- **WHEN** the operator chooses manual at the integrate question
- **THEN** `/myflow-fast` hands off as `/myflow-finish` run 1 does today; the next bare
  `/myflow-fast` call, once the operator has integrated the branch, runs archive (run 2)

### Requirement: Recorded defaults favor speed, still overridable

On the run that creates a change, `/myflow-fast` SHALL ask the same planning-effort, model and
review-panel-roster questions `/myflow-start` already asks, recommending `sonnet` for
`models.implementation`, `models.reviewPanel` and `models.panelFix`, and `light` for
`reviewPanelRoster` — recommendations only; an operator answer overrides them exactly as it would
under `/myflow-start`.

#### Scenario: Defaults are recommended, not forced

- **WHEN** `/myflow-fast` asks the model and roster questions on a creating run
- **THEN** Sonnet is marked the recommended answer for all three model roles and `light` is marked
  the recommended roster, and an operator choosing a different answer has that answer recorded
  instead

