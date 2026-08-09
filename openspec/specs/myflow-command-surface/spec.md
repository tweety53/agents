# myflow-command-surface Specification

## Purpose
TBD - created by archiving change kan-8-myflow-updates. Update Purpose after archive.
## Requirements
### Requirement: Each command declares the states it accepts

The accepted states SHALL be: `/myflow-start` — none or `STARTED`; `/myflow-do` — `STARTED` or
`IN_PROGRESS`; `/myflow-finish` — `IN_PROGRESS`; `/myflow-fast` — none or `IN_PROGRESS`;
`/myflow-status` — any.

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

#### Scenario: A mismatched fast invocation does not advance the state

- **WHEN** `/myflow-fast` is invoked for a change at `STARTED`
- **THEN** it reports the mismatch, suggests `/myflow-do`, and writes no state

### Requirement: Git actions are bounded by state

`/myflow-start` SHALL perform no git actions beyond writing planning artifacts. `/myflow-do` SHALL
create or resume a worktree and stage with `git add`, and SHALL commit and push **only** when the
state file already records a `prUrl`. `/myflow-finish` SHALL commit, push, open a PR or merge on
its first run, and commit and push the archive and remove worktrees, branches and the remote branch
on its second.

No command other than `/myflow-finish`, and `/myflow-do` when a PR is already open, SHALL create a
commit.

Every staging pass `/myflow-do` performs SHALL exclude the paths the review diff excludes, so the
staging area holds implementation only. `/myflow-finish` SHALL stage those paths on its first run
and commit them **separately** from the implementation, implementation first.

`/myflow-do`'s commit-and-push exception SHALL make the same two commits, so a fix pushed to an open
PR keeps its code commits free of planning artifacts.

#### Scenario: A fix before integration stays staged

- **WHEN** `/myflow-do` completes for a change at `IN_PROGRESS` with no `prUrl` recorded
- **THEN** the changes are staged and uncommitted, and `git status` shows them as staged

#### Scenario: A fix after integration reaches the PR

- **WHEN** `/myflow-do` completes for a change whose state file records a `prUrl`
- **THEN** the fix is committed and pushed to the PR branch as two commits, implementation first

#### Scenario: Staging holds implementation only

- **WHEN** `/myflow-do` stages at the end of any run
- **THEN** no path under `openspec/` or `docs/superpowers/` is staged

#### Scenario: Finish commits both, in order

- **WHEN** `/myflow-finish` commits on its first run
- **THEN** the implementation is committed first and the planning artifacts second

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

#### Scenario: Do carries the run-instruction resolution

- **WHEN** `/myflow-do` runs after the restructure
- **THEN** the step that resolves the run instructions is present in `skills/myflow-do/`, and
  `openspec-manual-test-superpowers` no longer exists

### Requirement: The opsx commands that duplicate pipeline steps are removed

`/opsx:propose`, `/opsx:apply`, `/opsx:archive`, `/opsx:sync-specs` and `/opsx:update` SHALL be
removed together with their skills. `/opsx:explore` and `skills/openspec-explore/` SHALL be kept,
because it is a thinking mode with no myflow equivalent that touches no state.

#### Scenario: Explore survives the removal

- **WHEN** `commands/` is listed after the change
- **THEN** `opsx-explore.md` is present and no other `opsx-*.md` file is

### Requirement: Change name resolution enumerates every state-tracked change

The shared change name resolution used by every `/myflow-*` command SHALL enumerate open changes
from the union of `openspec list --json` and the names of the files in the project's state
directory (`/Users/tweety53/Agents/myflow/state/<project-key>/*.json`), minus any name whose
`openspec/changes/<name>/` directory has reached `archive/`.

This resolution SHALL be defined once, in the shared `## Change name resolution` section every
`/myflow-*` command reads, so no command's enumeration can drift from another's.

A state-directory file that cannot be read SHALL be reported and skipped from the union, and SHALL
NOT be silently dropped without being named.

#### Scenario: A worktree-only change is discovered

- **WHEN** a change's `openspec/changes/<name>/` directory exists only in a worktree, staged but
  never committed to the main checkout, and its state file exists in the project's state directory
- **THEN** `/myflow-status` lists the change

#### Scenario: Every command shares the same enumeration

- **WHEN** `/myflow-do` or `/myflow-finish` is invoked with no change name and more than one match
  exists after the union
- **THEN** the operator is asked which change, exactly as `/myflow-status` would enumerate the same
  set

#### Scenario: An archived change is not re-surfaced

- **WHEN** a change's state file still exists in the state directory but its
  `openspec/changes/<name>/` directory has moved to `archive/`
- **THEN** it is not included in the enumeration

#### Scenario: An unreadable state file is reported, not silently dropped

- **WHEN** a file in the project's state directory cannot be parsed
- **THEN** it is skipped from the union and named in the report, rather than causing that change to
  be invisible with no explanation

### Requirement: The command surface is three pipeline commands, one composite command, plus one read-only one

myflow SHALL expose `/myflow-start`, `/myflow-do` and `/myflow-finish` as pipeline commands,
`/myflow-fast` as the composite command that chains their stage content across state transitions
that carry no human gate, and `/myflow-status` as the only read-only command.

`/myflow-info` SHALL NOT exist. Its sole job was to read `skills/myflow-contracts/pipeline.md` at
invocation time and explain the pipeline from it; that explanation now lives in `README.md`, where a
human reads it directly and no command pays for it.

Every command SHALL exist in both `commands/` and `commands-claude/` with the same name and a
description that agrees with the skill it points at.

#### Scenario: The retired commands are absent from both trees

- **WHEN** `commands/` and `commands-claude/` are listed
- **THEN** no file exists for `/myflow-info`, `/myflow-full`, `/myflow-fast-path`,
  `/myflow-manual-test`, `/myflow-review`, `/myflow-start-fix`, `/myflow-start-done`,
  `/myflow-do-fix`, `/myflow-do-manual-review`, `/myflow-do-done`, `/myflow-do-fix-manual-review`,
  `/myflow-do-fix-done`, `/myflow-manual-test-done` or `/myflow-review-done`

#### Scenario: A command and its skill agree

- **WHEN** a command file states which states it accepts
- **THEN** the skill it delegates to states the same set

#### Scenario: The info skill is gone with its commands

- **WHEN** `skills/` is listed
- **THEN** no `myflow-info` directory exists

#### Scenario: `/myflow-fast` exists in both command trees

- **WHEN** `commands/` and `commands-claude/` are listed
- **THEN** a `myflow-fast` file exists in both, and its description agrees with
  `skills/myflow-fast/SKILL.md`

