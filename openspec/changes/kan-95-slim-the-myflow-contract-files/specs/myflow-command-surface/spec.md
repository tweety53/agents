## MODIFIED Requirements

### Requirement: Each command declares the states it accepts

The accepted states SHALL be: `/myflow-start` — none or `STARTED`; `/myflow-do` — `STARTED` or
`IN_PROGRESS`; `/myflow-finish` — `IN_PROGRESS`; `/myflow-status` — any.

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

## ADDED Requirements

### Requirement: The command surface is three pipeline commands plus one read-only one

myflow SHALL expose exactly `/myflow-start`, `/myflow-do` and `/myflow-finish` as pipeline
commands, plus `/myflow-status` as the only read-only one.

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

## REMOVED Requirements

### Requirement: The command surface is three pipeline commands plus two read-only ones

**Reason**: Renamed. `/myflow-info` is removed, so the count in the title is false. This is the
same correction this change already made to `State writes are monotonic with one exception` in
`myflow-state-machine`, and for the same reason: a title that states something false about the
rule beneath it is not a title worth archiving.

**Migration**: **The command surface is three pipeline commands plus one read-only one** above
carries the surface unchanged apart from the removal of `/myflow-info`. No behaviour is lost.
