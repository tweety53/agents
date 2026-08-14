## ADDED Requirements

### Requirement: Every stage of every command is marked

Each `/myflow-*` skill SHALL emit a begin mark when a stage starts and an end mark when it finishes,
naming the command, the stage, and the change.

The stage names SHALL be those in **Level 1 — the stages of each command** (`README.md`). No skill
SHALL invent a stage name absent from that list, so the vocabulary recorded in the store and the
vocabulary in the documentation are one list rather than two that can drift.

A mark SHALL NEVER block, delay, or alter the stage it marks. A failed mark SHALL degrade exactly as
a failed state write does: one warning line, a journal entry, and the run continues.

#### Scenario: A stage is marked at both ends

- **WHEN** `/myflow-do` runs its SDD + TDD stage to completion
- **THEN** a stage run is recorded with that command, that stage name, a start instant, an end
  instant, and an outcome

#### Scenario: A mark names a stage the documentation does not

- **WHEN** a skill would emit a stage name absent from the README Level 1 list
- **THEN** that is a defect in the skill, corrected by using the documented name or by adding the
  stage to the documented list — never by recording an undocumented name

#### Scenario: Marking fails

- **WHEN** a mark cannot reach the store
- **THEN** one warning line is printed, the intent is journalled, and the stage proceeds unaffected

### Requirement: Re-entrant stages are recorded as separate attempts

A stage run SHALL be identified by change, command, stage and attempt number together. Re-running a
command SHALL record a new attempt rather than overwrite the previous one, so the record of a fix
round is preserved.

#### Scenario: A fix re-runs an already-recorded stage

- **WHEN** `/myflow-do` is re-run at `IN_PROGRESS` as a fix
- **THEN** its stages are recorded as attempt 2, and attempt 1 remains readable
- **AND** statistics that count rework read the attempt number rather than inferring rework from
  timestamps

### Requirement: Cost metrics are harvested, never self-reported

Token metrics SHALL be derived from the harness's own session transcript, not from values an agent
supplies about itself. The daemon SHALL read Claude Code session transcripts and attribute each
assistant message's recorded usage, model and reasoning effort to the open stage window whose
session matches and whose interval contains the message's timestamp.

Main-thread and subagent usage SHALL accumulate separately, so a dispatching command's own cost is
distinguishable from the cost of what it dispatched.

Attribution SHALL be idempotent: re-reading a transcript SHALL NOT double-count usage already
attributed.

#### Scenario: A stage's tokens are attributed from the transcript

- **WHEN** a stage window is open and assistant messages are written to the matching session
  transcript
- **THEN** those messages' input, output, cache-creation, cache-read and thinking token counts are
  attributed to that stage run

#### Scenario: Subagent cost is distinguishable

- **WHEN** a stage dispatches subagents whose messages are marked as sidechain
- **THEN** their usage is accumulated separately from the main thread's, and both are readable for
  that stage run

#### Scenario: The harvester restarts

- **WHEN** the daemon restarts and re-opens a transcript it has already partly consumed
- **THEN** usage already attributed is not attributed a second time

### Requirement: An unavailable metric is recorded as unavailable

A metric that could not be measured SHALL be recorded as absent or explicitly unavailable, and
SHALL NEVER be recorded as zero. On a harness that writes no machine-readable transcript, token
metrics SHALL be marked unavailable while state, marks and wall-clock duration are still recorded.

#### Scenario: A run on a harness with no transcript

- **WHEN** a stage runs under a harness that writes no readable transcript
- **THEN** the stage run records its duration, command, stage and outcome
- **AND** its token metrics are marked unavailable rather than recorded as zero

#### Scenario: A view encounters an unavailable metric

- **WHEN** a statistics view aggregates over runs whose token metrics are unavailable
- **THEN** those runs are excluded from the token figures and their exclusion is visible in the
  output, rather than being averaged in as zeros

### Requirement: The metric structure absorbs new metrics without migration

Measured values SHALL be stored in a schema-flexible structure, so that recording a new metric
requires writing a new key and never a database migration. Identity and lifecycle fields — project,
change, command, stage, attempt, timestamps and outcome — SHALL remain typed columns.

Recording a metric SHALL combine the new values with the existing ones **recursively**: where both
hold an object under the same key, their contents SHALL be merged rather than one replacing the
other, and only a non-object value SHALL replace what was there. A writer SHALL NOT be required to
know which other values exist under a shared parent in order to avoid destroying them.

#### Scenario: A new metric is added

- **WHEN** a new measurement is introduced for a stage
- **THEN** it is stored as a new key alongside the existing ones
- **AND** no migration is required, and existing rows remain valid without it

#### Scenario: Two writers touch different keys under the same parent

- **WHEN** one writer records a value under a nested key and a later writer records a different
  nested key under the same parent
- **THEN** both values are present afterwards
- **AND** neither writer needed to know what the other had written in order to preserve it

#### Scenario: A recorded value is replaced by a later write of the same key

- **WHEN** a writer records a value for a key that already holds a non-object value
- **THEN** the later value replaces the earlier one

### Requirement: An abandoned stage is recorded, not discarded

A stage whose run ends without an end mark SHALL be closed by the daemon after its session has been
silent past a timeout, with an outcome recording that it was abandoned.

An abandoned stage is a statistic, not an error condition to suppress.

#### Scenario: A run dies mid-stage

- **WHEN** a session stops without emitting the end mark for its open stage
- **THEN** the daemon closes that stage run with an abandoned outcome after the timeout
- **AND** the rework-rate view can read abandoned stages directly

### Requirement: Cost in currency is reproducible

A stage run's monetary cost SHALL be stored alongside the token counts it was derived from and the
version of the pricing data used to derive it, so that history remains stable when displayed and can
be recomputed when prices change.

#### Scenario: Prices change

- **WHEN** model pricing is updated
- **THEN** previously recorded costs continue to display the figure they were computed with
- **AND** the stored token counts and pricing version make recomputation possible
