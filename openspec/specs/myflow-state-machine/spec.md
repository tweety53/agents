# myflow-state-machine Specification

## Purpose
TBD - created by archiving change kan-8-myflow-updates. Update Purpose after archive.
## Requirements

### Requirement: The pipeline has exactly three states

myflow SHALL track a change in exactly one of three states: `STARTED`, `IN_PROGRESS`, `FINISHED`.
Each pipeline command SHALL end in the state named after it, so the state names and the command
names are the same vocabulary.

The human gate SHALL be a property of the state rather than a separate stage: `STARTED` means the
proposal awaits reading, `IN_PROGRESS` means the staged diff awaits the human's review and the run,
`FINISHED` is terminal.

#### Scenario: Each command lands in its own state

- **WHEN** `/myflow-do` completes from `STARTED`
- **THEN** the state is `IN_PROGRESS`, and no separate command is required to record that a human
  reviewed or tested anything

#### Scenario: No state records a human confirmation

- **WHEN** the state file is read at any point in the pipeline
- **THEN** it contains no field asserting that a human reviewed the proposal, reviewed the diff,
  or ran the tests, and no command exists whose only effect is to write such a field

### Requirement: Reviewing and testing are one gate

`/myflow-do` SHALL produce **both** the staged implementation diff **and** the run instructions in
its handoff, in the same run, so that the human reviews the diff and runs the apps at a single stop.

The run instructions SHALL be printed rather than written to a file, per **The `IN_PROGRESS` handoff
carries run instructions** in `myflow-handoff-output`. No artifact SHALL record what the human ran
or whether they ran it.

No separate command SHALL exist for producing either surface, and no state SHALL exist between
implementation and finishing.

#### Scenario: One run produces both review surfaces

- **WHEN** `/myflow-do` completes
- **THEN** the worktree contains a staged diff, and the handoff names the worktree's absolute path
  and carries the run instructions

#### Scenario: A fix refreshes both

- **WHEN** `/myflow-do` is re-run as a fix at `IN_PROGRESS`
- **THEN** the handoff's run instructions are resolved afresh alongside the code, so the two
  surfaces never drift apart

### Requirement: Every command is re-entrant

Re-invoking a pipeline command SHALL be the supported way to revise its output. `/myflow-start`
re-run at `STARTED` SHALL revise the proposal and republish the artifact to the same URL.
`/myflow-do` re-run at `IN_PROGRESS` SHALL apply a fix in the existing worktree.

No separate `*-fix` command SHALL exist for either.

#### Scenario: Re-running start revises rather than recreating

- **WHEN** `/myflow-start` is invoked for a change already at `STARTED`
- **THEN** the existing artifacts are revised, the proposal artifact is republished to its
  existing URL, and the state remains `STARTED`

#### Scenario: Re-running do resumes the existing worktree

- **WHEN** `/myflow-do` is invoked for a change at `IN_PROGRESS`
- **THEN** the existing worktree and branch are resumed, and no new worktree is created

### Requirement: A fix never moves the state

`/myflow-do` SHALL advance the state only from `STARTED` to `IN_PROGRESS`. From `IN_PROGRESS` it
SHALL leave the state exactly as it found it.

No field SHALL record where a fix was raised. Whether the human re-reviews or re-tests after a fix
SHALL be their decision.

#### Scenario: A fix before integration leaves the state alone

- **WHEN** `/myflow-do` completes for a change at `IN_PROGRESS` with no PR open
- **THEN** the state is still `IN_PROGRESS`, the changes are staged and uncommitted, and no
  `originStage` or equivalent field is written

#### Scenario: A fix after a PR is open reaches the PR

- **WHEN** `/myflow-do` completes for a change at `IN_PROGRESS` whose state file records a `prUrl`
- **THEN** the fix is committed and pushed to the PR branch, because a staged-only fix would be
  invisible on an open PR

### Requirement: The state file carries only fields with a live consumer

The state file SHALL contain `state`, `branch`, `worktrees`, `artifactUrl`, `jiraIssue`,
`planningEffort`, `models`, `reviewPanelRoster`, `prUrl`, `updatedAt` and `updatedBy`, and SHALL NOT
contain `gates`, `tested`, `originStage`, `REVIEWED_TREE`, `fastPath`, separate `worktree` and
`MERGE_BASE` fields, or `effort`. That is a rule about what a **write** emits: a file this pipeline
produces never carries the retired key, which is a different question from what a read does with one
that already does.

`worktrees` SHALL be an object keyed by the **absolute path** of each affected worktree, whose
value is that worktree's merge base, or `null` where no merge base is recorded for that path. This
key set SHALL be the authoritative list of worktrees for a change spanning more than one repository,
and a `null` value SHALL be read exactly as a missing merge base is — never as a licence to infer
one.

`models` SHALL be an object carrying `implementation`, `reviewPanel` and `panelFix`, each naming the
model chosen for that role. Its live consumer is `/myflow-do`, which dispatches on those values;
`myflow-model-policy` is canonical for the roles, their defaults and how an override applies.

`planningEffort` SHALL carry the level chosen for the change's planning. Its live consumer is
`/myflow-start`, which reuses it on a revision round; `myflow-planning-effort` is canonical for the
levels and the default.

`reviewPanelRoster` SHALL carry the review panel roster preset chosen for the change. Its live
consumer is `/myflow-do`, which selects the panel's required slots and the per-task review's shape
from it; `myflow-review-panel-roster` is canonical for the presets, the default, and what each one
means. A roster is not a model, so the field is top-level rather than nested under `models`.

`planningEffort`, `models` and `reviewPanelRoster` SHALL each read as "not recorded" when absent,
rather than making the file unparseable. This extends the exception the effort field already
carried, and for the same reason: a file written before a field existed would otherwise be reported
as unparseable and its change skipped, which is a loud failure for a value nobody had the
opportunity to set. The exception covers a key that is **absent**, which is a different thing from a
key that is present and null. For `reviewPanelRoster`, "not recorded" resolves to that capability's
stated default rather than leaving the panel unconfigured.

A file carrying the retired `effort` key SHALL be read as recording the equivalent level rather than
as unparseable, and SHALL be rewritten under `planningEffort` on the next write it receives. Where
both keys are present, `planningEffort` SHALL win; a value outside the mapped three SHALL read as
"not recorded" and SHALL NOT make the file unparseable. The mapping is
`openspec/specs/myflow-planning-effort/spec.md`'s, and is not restated here.

**A state file that is missing or unparseable SHALL be reported and skipped, never rebuilt from
inference.** No command infers a state file's contents, so an unreadable file names itself in the
reading command's own output and its change is omitted from that command's report. The `worktrees`
map is therefore never reconstructed by scanning, and a change whose state file cannot be read is
visible as unreadable rather than silently replaced by a guess.

The file SHALL continue to live outside the repository at
`/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`, resolved via `--git-common-dir`,
and SHALL never be staged, committed or archived.

#### Scenario: Writing state carries forward unowned fields

- **WHEN** any command writes the state file
- **THEN** it renders the whole object, re-emitting `artifactUrl`, `jiraIssue`, `prUrl`,
  `planningEffort`, `models`, `reviewPanelRoster` and `worktrees` as read, rather than resetting a
  field it does not own

#### Scenario: Multi-repo worktrees are enumerable

- **WHEN** a change affects two repositories
- **THEN** `worktrees` has one absolute-path key per repository, and `/myflow-finish` cleans up
  every key

#### Scenario: An unreadable state file is named rather than rebuilt

- **WHEN** a command reads a state file that is missing or unparseable
- **THEN** it names that file in its own output and omits the change from its report
- **AND** it writes nothing to that file and reconstructs no `worktrees` map

#### Scenario: No field records testing

- **WHEN** the state file is read after `/myflow-do` has handed off at `IN_PROGRESS`
- **THEN** there is no `tested` field, because no command observes whether the human ran the apps

#### Scenario: An absent optional field is legal

- **WHEN** a command reads a state file carrying neither `planningEffort`, `models` nor
  `reviewPanelRoster`
- **THEN** the file parses normally with all three read as not recorded
- **AND** the file is not reported as unparseable on that account

#### Scenario: An absent roster resolves to its default rather than to nothing

- **WHEN** `/myflow-do` reads a state file carrying no `reviewPanelRoster`
- **THEN** it dispatches the panel at that capability's default preset
- **AND** it does not treat the panel as unconfigured or ask which roster to use

#### Scenario: The retired key does not make the file unparseable

- **WHEN** a command reads a state file carrying `effort`
- **THEN** the file parses and the level is read as the equivalent under the current key
- **AND** the next write carries `planningEffort` and no `effort` key
- **AND** the file is not reported as unparseable on that account

### Requirement: State writes are monotonic

No command SHALL write a state earlier than the one it read. There is no exception.

`prUrl` SHALL NOT be cleared by any command. The one carve-out this rule previously carried allowed
`/myflow-status` to clear a recorded `prUrl` when a PR CLI conclusively answered that no pull request
existed for the branch; that command no longer makes any network call, so the evidence the carve-out
required can no longer be gathered, and a rule whose precondition can never be met is removed rather
than left to read as live guidance.

#### Scenario: A recorded PR URL is never cleared

- **WHEN** any command reads a state file recording a `prUrl`
- **THEN** it carries that value forward verbatim
- **AND** no command clears it, whatever the pull request's actual state

#### Scenario: An earlier state is never written

- **WHEN** a command that accepts `IN_PROGRESS` completes for a change at `IN_PROGRESS`
- **THEN** it writes `IN_PROGRESS` back, and never `STARTED`
