# myflow-state-advance Specification

## Purpose
TBD - created by archiving change kan-10-myflow-economical-updates. Update Purpose after archive.
## Requirements
### Requirement: A script performs the mechanical state write

`skills/myflow-state-advance/state-advance.sh` SHALL accept `--name`, `--target`, `--accepted`,
and `--by`, and on the happy path SHALL resolve the state file per the State file contract, verify
the current stage is in `--accepted`, write the full state object, print the handoff, and exit `0`.

The script SHALL NOT prompt, SHALL NOT lower a gate value, SHALL NOT contact Jira, and SHALL NOT
write a stage when the current stage is outside `--accepted`.

#### Scenario: Happy-path advance

- **WHEN** the script runs with `--name X --target do-done --accepted awaiting-do-review,do-review-started --by /myflow-do-done`
  and X's state file records `awaiting-do-review`
- **THEN** the file records `stage: do-done`, a current `updatedAt`, and `updatedBy: /myflow-do-done`
- **AND** the script exits `0`

#### Scenario: Unowned fields are carried forward verbatim

- **WHEN** the script advances a change whose state file carries `artifactUrl`, `jiraIssue`,
  `fastPath`, `REVIEWED_TREE`, `MERGE_BASE`, and all four gate values
- **THEN** every one of those fields is byte-identical in the written file
- **AND** no gate value is lowered

#### Scenario: Jira is never contacted

- **WHEN** the script advances a change whose `jiraIssue` is set
- **THEN** no network call is made and `jiraIssue` is unchanged

### Requirement: The script escalates with distinct exit codes

The script SHALL exit with a distinct non-zero code for each condition it cannot mechanically
resolve, and SHALL write nothing when it escalates.

| Code | Condition |
|------|-----------|
| 2 | Invalid arguments — an unknown flag, a flag with no value, a `--name` outside `[a-z0-9-]`, or a `--target`/`--accepted` value that is not a pipeline stage — or `jq` not installed |
| 3 | State file missing, unparseable, not a JSON object, missing a **structural** key (`stage`, or `gates` with its four gate values), or a non-null `worktree` that its own repository no longer lists |
| 4 | Current stage not in `--accepted` |
| 5 | Dynamic target and `originStage` is `null` or missing |
| 6 | Dynamic target and `originStage` holds a value outside the six legal origins |

#### Scenario: Stage mismatch escalates without writing

- **WHEN** the current stage is `proposal-done` and `--accepted` is `awaiting-do-review`
- **THEN** the script exits `4`
- **AND** the state file is unmodified

#### Scenario: Missing state file escalates

- **WHEN** no state file exists for the named change
- **THEN** the script exits `3` and creates no file

#### Scenario: Stale worktree escalates

- **WHEN** the state file records a non-null `worktree` whose directory does not exist
- **THEN** the script exits `3`
- **AND** the state file is unmodified

#### Scenario: A live worktree is probed in its own repository, not in cwd's

- **WHEN** the state file records a non-null `worktree` that its own repository lists, and the
  script is run from a directory in no repository at all
- **THEN** the script exits `0` and writes the stage
- **AND** it does not treat the recorded path as stale

#### Scenario: An unanswerable worktree probe is not a contradiction

- **WHEN** the recorded `worktree` directory exists but git cannot answer for it
- **THEN** the script treats the check as unknown and does not escalate

#### Scenario: A structurally invalid state file escalates rather than raising a jq error

- **WHEN** the state file's top level is an array, a string, or a number
- **THEN** the script exits `3`, never jq's own status `5`

#### Scenario: A partial state file escalates rather than being perpetuated

- **WHEN** the state file is a JSON object missing a structural key — a `{"stage": …}` fragment,
  or one whose `gates` lacks a gate value
- **THEN** the script exits `3`
- **AND** the file is unmodified

#### Scenario: An optional field the contract allows to be absent is accepted

- **WHEN** the state file omits `fastPath`, which the State file contract documents as
  "`null`/absent"
- **THEN** the script exits `0` and writes the stage
- **AND** the omitted field is carried forward as omitted, not invented

#### Scenario: Arguments are validated before they reach a path or a stage

- **WHEN** `--name` contains anything outside `[a-z0-9-]`, or `--target` is not a pipeline stage
- **THEN** the script exits `2`, writes nothing, and creates nothing outside the state directory

#### Scenario: Corrupt originStage escalates rather than being repaired

- **WHEN** `--target originStage` is given and `originStage` is `awaiting-fix-review`
- **THEN** the script exits `6`
- **AND** it does not write any stage

### Requirement: Dynamic target resolution matches the fix re-entry contract

When `--target originStage` is given, the script SHALL target the recorded `originStage`, except
that `do-review-started` SHALL resolve to `awaiting-do-review`. On a successful write under this
form the script SHALL also set `originStage` to `null`.

#### Scenario: do-review-started resolves to awaiting-do-review

- **WHEN** `--target originStage` is given and `originStage` is `do-review-started`
- **THEN** the written stage is `awaiting-do-review`
- **AND** `originStage` is `null` in the written file

#### Scenario: The other five origins target themselves

- **WHEN** `--target originStage` is given and `originStage` is `manual-test-done`
- **THEN** the written stage is `manual-test-done`
- **AND** `originStage` is `null` in the written file

### Requirement: Commands run the script first and the skill only on escalation

Each of the seven pure-state-write commands, in both `commands/` and `commands-claude/`, SHALL
invoke `state-advance.sh` first, report its output and stop on exit `0`, and load the
`myflow-state-advance` skill only on a non-zero exit.

#### Scenario: Both command trees are updated

- **WHEN** the seven `*-done` / `*-manual-review` command files are read in either tree
- **THEN** each instructs the agent to run `state-advance.sh` before loading the skill

#### Scenario: Any other non-zero exit routes to the skill

- **WHEN** the script exits with a status the command file does not enumerate — `1`, `126` or
  `127`, as a missing or non-executable script produces
- **THEN** the command loads the `myflow-state-advance` skill and proceeds as if the script were
  unavailable

#### Scenario: Escalation preserves existing behavior

- **WHEN** the script exits `4`
- **THEN** the agent loads the skill and emits the stage-mismatch handoff with the override prompt
  defaulting to "No — run the suggested command instead"

### Requirement: Self-heal narrowing is documented at the point of change

The `myflow-state-advance` skill and `skills/myflow-contracts/state-self-heal.md` SHALL each state
that the artifact-contradiction checks do not run on the script's happy path for these seven
commands, and why that is acceptable.

#### Scenario: The narrowing is stated, not silent

- **WHEN** `skills/myflow-contracts/state-self-heal.md` is read
- **THEN** it states which checks the script performs, which it skips, and that `/myflow-review`
  and `/myflow-finish` still verify independently

#### Scenario: The one uncovered dropped check is recorded as a known gap

- **WHEN** `skills/myflow-contracts/state-self-heal.md` is read
- **THEN** it records that the dropped "commits beyond `MERGE_BASE` and a PR exists" row was the
  only detector of a backward-stale state file, that nothing downstream catches it, and why a
  stage-monotonicity check in the script would break `--target originStage`

