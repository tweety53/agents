# myflow-command-surface Specification

## Purpose
TBD - created by archiving change kan-8-myflow-updates. Update Purpose after archive.
## Requirements
### Requirement: The command surface is three pipeline commands plus two read-only ones

myflow SHALL expose exactly `/myflow-start`, `/myflow-do` and `/myflow-finish` as pipeline
commands, plus `/myflow-status` and `/myflow-info` as read-only ones.

Every command SHALL exist in both `commands/` and `commands-claude/` with the same name and a
description that agrees with the skill it points at.

#### Scenario: The retired commands are absent from both trees

- **WHEN** `commands/` and `commands-claude/` are listed
- **THEN** no file exists for `/myflow-full`, `/myflow-fast-path`, `/myflow-manual-test`,
  `/myflow-review`, `/myflow-start-fix`, `/myflow-start-done`, `/myflow-do-fix`,
  `/myflow-do-manual-review`, `/myflow-do-done`, `/myflow-do-fix-manual-review`,
  `/myflow-do-fix-done`, `/myflow-manual-test-done` or `/myflow-review-done`

#### Scenario: A command and its skill agree

- **WHEN** a command file states which states it accepts
- **THEN** the skill it delegates to states the same set

### Requirement: Each command declares the states it accepts

The accepted states SHALL be: `/myflow-start` — none or `STARTED`; `/myflow-do` — `STARTED` or
`IN_PROGRESS`; `/myflow-finish` — `IN_PROGRESS`; `/myflow-status` and `/myflow-info` — any.

On a mismatch a command SHALL stop, report the actual state, the states it expects and the command
to run instead, and SHALL ask for an explicit override whose default and recommended answer is to
run the suggested command instead.

#### Scenario: A mismatched command does not advance the state

- **WHEN** `/myflow-finish` is invoked for a change at `STARTED`
- **THEN** it reports the mismatch, suggests `/myflow-do`, and writes no state

#### Scenario: Override is explicit

- **WHEN** the operator is asked to override a state mismatch
- **THEN** the default answer is to run the suggested command instead, and only an explicit choice
  to override proceeds

### Requirement: Git actions are bounded by state

`/myflow-start` SHALL perform no git actions beyond writing planning artifacts. `/myflow-do` SHALL
create or resume a worktree and stage with `git add`, and SHALL commit and push **only** when the
state file already records a `prUrl`. `/myflow-finish` SHALL commit, push, open a PR or merge on
its first run, and commit and push the archive and remove worktrees on its second.

No command other than `/myflow-finish`, and `/myflow-do` when a PR is already open, SHALL create a
commit.

#### Scenario: A fix before integration stays staged

- **WHEN** `/myflow-do` completes for a change at `IN_PROGRESS` with no `prUrl` recorded
- **THEN** the changes are staged and uncommitted, and `git status` shows them as staged

#### Scenario: A fix after integration reaches the PR

- **WHEN** `/myflow-do` completes for a change whose state file records a `prUrl`
- **THEN** the fix is committed and pushed to the PR branch

### Requirement: No command accepts a flag

No `/myflow-*` command SHALL accept any flag. The flags `automerge`, `skip-review`,
`skip-manual-test`, `skip-propose`, `propose-only`, `checkpoint`, `commit-during-apply` and
`full-panel` SHALL NOT exist.

The only argument any command takes SHALL be the optional change name. Behaviour that a flag
previously selected SHALL either be asked at invocation (the integration choice), derived from the
current state, or fixed at the single sensible default.

#### Scenario: A retired flag is not documented anywhere

- **WHEN** the command files, skills, `README.md`, `CLAUDE.md` and `AGENTS.md` are searched
- **THEN** none of the retired flag names appears as a supported flag, and no document has a
  "Flags" section listing one

#### Scenario: An unrecognized argument is reported, not silently ignored

- **WHEN** a command is invoked with a word that is not a known change name
- **THEN** it reports that the command takes only a change name, rather than treating the word as
  a flag or ignoring it

#### Scenario: Panel breadth is decided by escalation, not by a flag

- **WHEN** `/myflow-do` re-runs the review panel after a fix
- **THEN** it runs targeted by default and escalates to the full whole-branch panel on its own
  triggers, with no flag available to force either mode

### Requirement: One skill per command, named after it

Each pipeline command SHALL load exactly one skill, and that skill's directory SHALL be named
after the command: `skills/myflow-start/`, `skills/myflow-do/`, `skills/myflow-finish/`.

Content that a surviving skill delegates to a removed skill SHALL be inlined into the surviving
skill **before** the removed skill is deleted.

#### Scenario: Finish does not delegate to a deleted skill

- **WHEN** `/myflow-finish` runs after the restructure
- **THEN** the delta-sync and archive-move steps it previously delegated to `openspec-sync-specs`
  and `openspec-archive-change` are present in `skills/myflow-finish/`, and neither removed skill
  is referenced

#### Scenario: Start does not delegate to a deleted skill

- **WHEN** `/myflow-start` runs after the restructure
- **THEN** the artifact-creation steps it previously delegated to `openspec-propose` are present
  in `skills/myflow-start/`

#### Scenario: Do carries the test guide generation

- **WHEN** `/myflow-do` runs after the restructure
- **THEN** the guide-writing steps previously in `openspec-manual-test-superpowers` are present in
  `skills/myflow-do/`, and that skill no longer exists

### Requirement: The opsx commands that duplicate pipeline steps are removed

`/opsx:propose`, `/opsx:apply`, `/opsx:archive`, `/opsx:sync-specs` and `/opsx:update` SHALL be
removed together with their skills. `/opsx:explore` and `skills/openspec-explore/` SHALL be kept,
because it is a thinking mode with no myflow equivalent that touches no state.

#### Scenario: Explore survives the removal

- **WHEN** `commands/` is listed after the change
- **THEN** `opsx-explore.md` is present and no other `opsx-*.md` file is

