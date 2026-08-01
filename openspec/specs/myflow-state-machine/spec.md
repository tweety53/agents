# myflow-state-machine Specification

## Purpose
TBD - created by archiving change kan-8-myflow-updates. Update Purpose after archive.
## Requirements
### Requirement: The pipeline has exactly three states

myflow SHALL track a change in exactly one of three states: `STARTED`, `IN_PROGRESS`, `FINISHED`.
Each pipeline command SHALL end in the state named after it, so the state names and the command
names are the same vocabulary.

The human gate SHALL be a property of the state rather than a separate stage: `STARTED` means the
proposal awaits reading, `IN_PROGRESS` means the staged diff and the test guide await the human,
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

`/myflow-do` SHALL produce **both** the staged implementation diff **and** the manual test guide
at `docs/manual-test/<name>.md` in the same run, so that the human reviews the diff and runs the
apps at a single stop.

No separate command SHALL exist for generating the test guide, and no state SHALL exist between
implementation and finishing.

#### Scenario: One run produces both review surfaces

- **WHEN** `/myflow-do` completes
- **THEN** the worktree contains a staged diff and a staged manual test guide, and the handoff
  names the absolute path of each

#### Scenario: A fix refreshes both

- **WHEN** `/myflow-do` is re-run as a fix at `IN_PROGRESS`
- **THEN** the guide is refreshed alongside the code, preserving already-ticked boxes, so the two
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
`planningEffort`, `models`, `prUrl`, `updatedAt` and `updatedBy`, and SHALL NOT contain `gates`,
`tested`, `originStage`, `REVIEWED_TREE`, `fastPath`, separate `worktree` and `MERGE_BASE` fields,
or `effort`. That is a rule about what a **write** emits: a file this pipeline produces never carries
the retired key, which is a different question from what a read does with one that already does.

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

`planningEffort` and `models` SHALL each read as "not recorded" when absent, rather than making the
file unparseable. This extends the exception the effort field already carried, and for the same
reason: routing every file written before a field existed through self-heal announces unrecovered
fields and rewrites from artifact inference, which is a loud correction for a value nobody had the
opportunity to set. The exception covers a key that is **absent**, which is a different thing from a
key that is present and null.

A file carrying the retired `effort` key SHALL be read as recording the equivalent level rather than
as unparseable, and SHALL be rewritten under `planningEffort` on the next write it receives. Where
both keys are present, `planningEffort` SHALL win; a value outside the mapped three SHALL read as
"not recorded" and SHALL NOT make the file unparseable. The mapping is
`openspec/specs/myflow-planning-effort/spec.md`'s, and is not restated here.

**A self-heal rebuild SHALL recover the `worktrees` keys rather than emptying the map.** When the
prior file cannot be read, the paths SHALL be recovered by scanning each affected repository for the
worktrees on the change's branch, and each recovered entry SHALL carry `null` for the merge base,
which the scan cannot produce. An emptied map is not a neutral loss: `/myflow-finish`'s preflight and
its unfinished-work gate are both defined as once per recorded worktree, so an empty map makes both
pass having examined none — for a change that may still hold an unmerged worktree. The rebuild SHALL
name `worktrees` in the correction announcement as recovered without merge bases.

The file SHALL continue to live outside the repository at
`/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`, resolved via `--git-common-dir`,
and SHALL never be staged, committed or archived.

#### Scenario: Writing state carries forward unowned fields

- **WHEN** any command writes the state file
- **THEN** it renders the whole object, re-emitting `artifactUrl`, `jiraIssue`, `prUrl`,
  `planningEffort`, `models` and `worktrees` as read, rather than resetting a field it does not own

#### Scenario: Multi-repo worktrees are enumerable

- **WHEN** a change affects two repositories
- **THEN** `worktrees` has one absolute-path key per repository, and `/myflow-finish` cleans up
  every key

#### Scenario: A rebuild recovers the worktree paths

- **WHEN** self-heal rebuilds a state file it could not read, for a change that still holds a
  worktree on its branch
- **THEN** the rewritten `worktrees` map carries that worktree's absolute path with a `null` merge
  base, rather than being emptied
- **AND** the correction announcement names `worktrees` as recovered without merge bases
- **AND** the next `/myflow-finish` refuses on that worktree rather than passing with none examined

#### Scenario: No field records testing

- **WHEN** the state file is read after `/myflow-do` has produced a test guide
- **THEN** there is no `tested` field, because no command observes whether the human ran the apps

#### Scenario: An absent optional field is legal

- **WHEN** a command reads a state file carrying neither `planningEffort` nor `models`
- **THEN** the file parses normally with both read as not recorded
- **AND** no self-heal correction is announced on that account

#### Scenario: The retired key does not make the file unparseable

- **WHEN** a command reads a state file carrying `effort`
- **THEN** the file parses and the level is read as the equivalent under the current key
- **AND** the next write carries `planningEffort` and no `effort` key
- **AND** no correction is announced on that account

### Requirement: State writes are monotonic with one exception

No command SHALL write a state earlier than the one it read, except that a change whose state file
records a `prUrl` MAY have that `prUrl` cleared when the PR's non-existence is **conclusively**
established by a usable PR CLI for the host answering that no PR exists.

An inconclusive probe — no PR CLI, no network, no remote — SHALL be treated as unknown and SHALL
clear nothing.

#### Scenario: A closed-unmerged PR does not strand a change

- **WHEN** `gh` reports no PR exists for the branch of a change whose state file records a `prUrl`
- **THEN** `prUrl` is cleared, the state remains `IN_PROGRESS`, and the correction is announced

#### Scenario: A missing PR CLI is not a contradiction

- **WHEN** PR state cannot be determined
- **THEN** `prUrl` is left as recorded and nothing is cleared

### Requirement: Self-heal validates state against artifacts

Artifacts SHALL remain the source of truth and the state file a cache. Every command SHALL read
the file, validate it against on-disk artifacts, and on a missing, unparseable or contradicted
file SHALL rewrite it with the inferred truth and announce the correction as
`⚠ state corrected: <old> → <new> (reason)`.

Artifacts SHALL be read from the apply worktree whenever one exists, and from the main checkout
only when none does.

#### Scenario: A state claim contradicted by a worktree is corrected

- **WHEN** the file says `STARTED` but a worktree for branch `openspec/<name>` exists
- **THEN** the state is corrected to `IN_PROGRESS` and the correction is announced

#### Scenario: A state file predating this model is rewritten, not migrated

- **WHEN** a state file carries a `stage` field instead of a `state` field
- **THEN** it is treated as unparseable and rewritten from artifacts, with no mapping table from
  the retired vocabulary
