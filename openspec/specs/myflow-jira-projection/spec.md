# myflow-jira-projection Specification

## Purpose
TBD - created by archiving change kan-17-finish-gate-jira-and-commit-hygiene. Update Purpose after archive.
## Requirements

### Requirement: The pipeline projects its state onto a tracker issue at three points, and nowhere else

When a change has a linked issue, `/myflow-start` SHALL move it to **In Progress**,
`/myflow-finish` SHALL move it to **In Review** on its first run and to **Done** after the archive
is written. No other command SHALL transition the issue.

`/myflow-do` SHALL NOT transition the issue at any time, including on a fix run. It MAY write the
issue **description** when a fix adds scope, which is a separate concern from status.

The linked key SHALL live in the state file's `jiraIssue` field, SHALL be resolved only by
`/myflow-start`, and SHALL be carried forward verbatim by every other command.

#### Scenario: A fix round does not move the issue backwards

- **WHEN** `/myflow-do` runs for a change whose issue is already In Review
- **THEN** no transition is attempted
- **AND** the issue remains In Review

#### Scenario: A change with no linked issue makes no Jira call

- **WHEN** any command runs for a change whose `jiraIssue` is null
- **THEN** no Jira request of any kind is made

### Requirement: The issue reaches In Progress when planning begins, not when it ends

`/myflow-start` SHALL transition the issue to In Progress **immediately after the issue key is
resolved**, before brainstorming produces any artifact — so the board reflects the work while a
long planning run is in flight.

The transition SHALL remain non-blocking: a failure SHALL degrade to a single skipped-with-reason
line, and the run SHALL continue and write its state file exactly as it would have. The earlier
call position SHALL NOT make any state write depend on a Jira call succeeding.

A revision round SHALL find the issue already at or past In Progress and SHALL make no call.

#### Scenario: Planning begins and the board follows

- **WHEN** `/myflow-start` resolves a linked issue that is at To Do
- **THEN** the issue moves to In Progress before the first brainstorming question is asked

#### Scenario: A failed transition does not stop planning

- **WHEN** the transition call fails for any reason
- **THEN** one `⚠ Jira: skipped — <reason>` line is emitted
- **AND** brainstorming continues and the state file is written at the end of the run as normal

#### Scenario: A revision round makes no call

- **WHEN** `/myflow-start` is re-run for a change already at `STARTED` whose issue is In Progress
- **THEN** no transition is attempted and the handoff says the status was already correct

### Requirement: The issue reaches In Review when run 1 completes, whichever route was taken

`/myflow-finish` SHALL transition the issue to In Review at the end of a **successful** first run,
independently of how the branch landed — pull request, merge and push, or handled manually.

The transition SHALL NOT be conditioned on a pull request existing. Conditioning it on a PR is what
allowed a merge-and-push change to reach Done without ever passing through In Review.

A run 1 that stops before its chosen route completes — a rejected push, a merge conflict, a failed
PR creation — SHALL NOT transition the issue.

#### Scenario: Merge and push still reaches In Review

- **WHEN** run 1 completes by merging the branch and pushing, opening no pull request
- **THEN** the issue is transitioned to In Review

#### Scenario: The manual route still reaches In Review

- **WHEN** run 1 completes by pushing the branch only, leaving integration to the operator
- **THEN** the issue is transitioned to In Review

#### Scenario: A failed run 1 does not claim review

- **WHEN** run 1 stops because its push was rejected
- **THEN** no transition is attempted and the failure is reported with the command's own output

### Requirement: Transitions are forward-only, matched by name, and an unrecognised status is asked about

Transitions SHALL be resolved by reading the issue's available transitions and matching the target
**by name**, case-insensitively, allowing the usual spellings. A numeric transition identifier SHALL
NEVER be hardcoded.

The order SHALL be To Do, then In Progress, then In Review, then Done. An issue already at or past
the target SHALL NOT be transitioned.

When the issue's **current** status is not one of those four names, its position in that order is
undefined. The command SHALL NOT infer a position — in particular SHALL NOT infer one from Jira's
`statusCategory` — and SHALL instead show the operator the issue key, the current status and the
intended target, and ask whether to transition. Only an explicit yes SHALL transition it; anything
else SHALL leave the status untouched and emit one skipped-with-reason line.

Inferring order from `statusCategory` would report an issue in a custom `indeterminate` status as
already at In Progress, freezing the board at that status for the whole change.

#### Scenario: A custom status is asked about rather than guessed

- **WHEN** the issue's current status is a workflow-specific name outside the four ordered names
- **THEN** the operator is shown the key, the current status and the target, and asked whether to
  transition

#### Scenario: Declining leaves the status untouched

- **WHEN** the operator declines the transition for an unrecognised status
- **THEN** no transition is made, one skipped-with-reason line is emitted, and the run continues

#### Scenario: A recognised status past the target is left alone

- **WHEN** the target is In Progress and the issue is already In Review
- **THEN** no transition is attempted and no question is asked

### Requirement: Issues the pipeline creates carry the parent issue's labels plus `AI-generated`

When a `/myflow-*` command creates a tracker issue, that issue SHALL be labelled with every label
carried by the change's linked issue, plus `AI-generated`.

No label SHALL be invented. The parent's labels exist by construction; `AI-generated` SHALL be
applied only because it is already in use in the project. When the change has no linked issue, the
created issue SHALL carry `AI-generated` alone.

The created issue SHALL be linked to the change's issue when one exists.

#### Scenario: A follow-up inherits its parent's routing label

- **WHEN** a command creates a follow-up issue during a change whose issue carries a routing label
- **THEN** the created issue carries that label and `AI-generated`

#### Scenario: No parent means no inherited labels

- **WHEN** a command creates an issue for a change with no linked issue
- **THEN** the created issue carries `AI-generated` and no other label

### Requirement: Jira is never a gate

No state write, commit, push, pull request, merge or archive SHALL depend on a Jira call
succeeding. Every failure path — no linked issue, integration unreachable, transition rejected,
target transition name not offered, issue not found — SHALL degrade to exactly one
`⚠ Jira: skipped — <reason>` line, after which the command SHALL continue and write its state
exactly as it would have.

A failure SHALL NOT be retried in a loop, SHALL NOT roll anything back, and SHALL NOT abort the run.

#### Scenario: An unreachable tracker does not stop an archive

- **WHEN** run 2 cannot reach the tracker to set Done
- **THEN** the archive is committed and pushed, the state is written `FINISHED`, and one
  skipped-with-reason line is reported

### Requirement: The issue description is appended to, never rewritten

`/myflow-start` and `/myflow-do` SHALL be the only commands that write the issue description, and
only when the run added scope the issue does not already describe. The write SHALL append a dated
bullet under an `## Added during implementation` heading, creating that heading once at the end of
the description when absent.

Because the underlying operation replaces the whole description, the payload SHALL be asserted
before every such write to contain the description just read as an **exact prefix**, byte for byte,
and to be **strictly longer** than it. If either assertion fails — including a read that was
truncated, elided, or returned in a form that cannot be reproduced verbatim — no write SHALL be made
and one skipped-with-reason line SHALL be emitted instead.

The pre-edit description SHALL be echoed verbatim into the handoff on any run that writes it, so the
transcript is the recovery path.

#### Scenario: A lossy read skips the write

- **WHEN** the existing description cannot be reproduced verbatim
- **THEN** no description write is made
- **AND** `⚠ Jira: description sync skipped — could not reproduce the existing description verbatim`
  is reported

#### Scenario: A run that added no scope writes nothing

- **WHEN** a run ends without having added scope the issue does not describe
- **THEN** the description is not written at all
