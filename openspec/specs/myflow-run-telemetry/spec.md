# myflow-run-telemetry Specification

## Purpose
TBD - created by archiving change kan-16-myflow-stats-app. Update Purpose after archive.
## Requirements
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

### Requirement: A stage mark carries a correlator that identifies its session

Every `stage begin` SHALL carry a correlator identifying the session it was made in, and the harness
it ran under.

The correlator SHALL identify a **session**, not a mark: one correlator is generated per command run
and passed on every mark that run makes. Once a correlator is bound to a session, every later mark
carrying it SHALL bind immediately, without repeating the work that established the binding.

A correlator SHALL be generated fresh per run. A correlator reused across concurrent runs would
resolve to more than one session and so identify none.

The correlator SHALL be a literal value written by the calling command — never a value produced by
expansion at execution time, which would be recorded identically for every caller and so identify
nothing.

The harness SHALL likewise be the harness the command is actually running under, never a fixed value
written into shared command text, since one source is installed into several harnesses.

**A command SHALL be recognised as a mark by the invocation it performs, not by where that
invocation sits in the command text.** Marks are emitted inside shell blocks carrying variable
assignments, directory changes and other statements; a rule that requires the invocation to lead the
text rejects the shapes actually emitted, and a mark that is not recognised binds nothing at all.

Recognition SHALL NOT be satisfied by a command that merely mentions a correlator. Where recognition
admits a command that only reproduces a mark's text without performing it, that SHALL be documented
rather than left implicit — **a false negative here is silent and total, while a false positive is
narrow and, where two sessions match, already refused by the ambiguity rule.**

#### Scenario: A mark inside a larger shell block

- **WHEN** a mark is emitted after variable assignments, a directory change, or on a later line of a
  multi-statement command
- **THEN** it is recognised, and its stage run binds to the session that emitted it

#### Scenario: A mark is emitted

- **WHEN** a pipeline command marks the start of a stage
- **THEN** the mark carries its run's correlator and the harness it ran under

#### Scenario: Later marks in the same run

- **WHEN** a run makes further marks after its correlator has been bound to a session
- **THEN** those marks are attributed to that session without re-establishing the binding

#### Scenario: Two sessions mark stages in the same directory at the same time

- **WHEN** two sessions each mark a stage while working in one directory
- **THEN** each stage run is bound to the session that marked it, and neither receives the other's
  usage

#### Scenario: A correlator that could identify more than one session

- **WHEN** a correlator matches more than one session
- **THEN** no session is recorded for that stage run, and the ambiguity is reported rather than
  resolved by choosing

### Requirement: A stage run is bound to its session after the mark, within a bounded window

Binding SHALL be permitted to complete after the mark is recorded, because the evidence that
identifies the session may not exist at the moment the mark is made.

Binding SHALL be attempted for a bounded period, after which the stage run SHALL be left
unattributed. An unbounded search SHALL NOT be performed, so that a harness which will never produce
the evidence costs a bounded amount of work rather than a permanent one.

Once bound, usage SHALL be attributed exactly as it is for a stage run whose session was known from
the start; binding SHALL introduce no second attribution path.

#### Scenario: The evidence appears after the mark

- **WHEN** the evidence identifying a stage run's session becomes available after the mark was
  recorded
- **THEN** the stage run is bound to that session, and usage falling in its window is attributed to
  it

#### Scenario: A harness that produces no such evidence

- **WHEN** a stage is marked on a harness that produces no session evidence at all
- **THEN** binding is attempted for the bounded period and then abandoned, and the stage run remains
  recorded and unattributed

#### Scenario: Binding does not disturb what is already attributed

- **WHEN** a stage run is bound to its session
- **THEN** usage already recorded against other stage runs is unchanged

### Requirement: Every documented stage is a stage, not a description of a command

The documented stage list SHALL name discrete stages for every command that marks them. A command's
entry SHALL NOT be prose describing the command as a whole, because a command whose entry is one
sentence can only ever record one stage and its per-stage statistics answer nothing.

#### Scenario: A command's documented stages

- **WHEN** a command that emits stage marks is examined against the documented list
- **THEN** its entry names discrete stages, each of which a command can begin and end

#### Scenario: A command whose recorded stages span its whole run

- **WHEN** a command records a single stage covering substantially its entire run
- **THEN** that is a defect in the documented stage list rather than a property of the command

### Requirement: A stage is identified by a stable key, not by its prose name

Every documented stage SHALL have a key that is unique across the whole pipeline and a human-readable
name. A mark SHALL identify its stage by key; the interface SHALL present the name.

A key SHALL be declared rather than derived from the name. A key derived from the name changes when
the name is improved, which is the outcome keys exist to prevent.

A stage's name MAY be reworded without changing what is recorded. Rewording SHALL NOT split that
stage's recorded history.

A key SHALL NOT be reused for a different stage. Two stages sharing a key would merge two different
things in every statistic, silently.

Where one command runs a stage another command defines, both SHALL record the same key, so runs of
the same stage under different commands are directly comparable.

#### Scenario: A stage's name is improved

- **WHEN** a documented stage's name is reworded
- **THEN** runs recorded before and after the rewording remain the same stage

#### Scenario: The same stage under two commands

- **WHEN** two different commands each run the same documented stage
- **THEN** both record it under the same key, and their runs are comparable as one stage

#### Scenario: A duplicate key

- **WHEN** two documented stages are given the same key
- **THEN** that is reported as a defect rather than accepted

#### Scenario: A mark naming an unknown key

- **WHEN** a mark names a key no documented stage declares
- **THEN** it is rejected, naming the documented alternatives

### Requirement: A state gate reads the state before it marks

A command that both gates on a change's state and marks its own stages SHALL read that state
**before** emitting its first mark.

A mark SHALL NOT cause the state it is about to be gated on to come into existence. Where marking
creates a record for a change that has none, a command that marks first observes a state its own
mark authored, and can refuse to proceed on the strength of it.

A record whose only author is a mark's own side effect SHALL NOT satisfy a state gate that expects a
state a pipeline command wrote.

#### Scenario: A creating run marks its state gate

- **WHEN** a command that accepts no existing state marks its own state-gate stage
- **THEN** it still proceeds as a creating run, rather than refusing on a state its mark created

### Requirement: A mark names a resolved change, never a guess

The `<change>` argument of a stage mark SHALL be a resolved change name. A command SHALL NOT emit a
mark naming a change name it has not yet resolved — a best guess, a bare issue key awaiting its
slug, or any other stand-in for a name the run is still working out.

Marking is not a read. Where the store has no such change, the begin handler bootstraps a change row
so the mark has something to attach to, so a guessed name becomes a change row that outlives the run
that guessed it: it appears among the open changes, carries a next command, and is never archived,
because no change directory bears that name.

This is the sibling of "A state gate reads the state before it marks", and rests on the same
property of marking: a mark writes. That requirement keeps a command from *reading* a state its own
mark authored; this one keeps a command from *creating* a change nobody named.

The bootstrap itself SHALL remain unchanged. A synthetic row is how an unattributable mark is made
visible rather than dropped, and suppressing it — in the store, in a query, or in a view — would
remove the report while leaving the cause in place.

A skill's `stage begin` call sites SHALL be checked mechanically for a change argument written as a
placeholder that names a guess, alongside the existing checks on that call's session token and
harness.

#### Scenario: A command resolves its change name after gating on state

- **WHEN** a command must read a change's state before that change's name is finally resolved
- **THEN** it performs the read with whatever name it has, and emits its state-gate marks only once
  the name is resolved

#### Scenario: A skill names a guess in a mark

- **WHEN** a `stage begin` call in skill source writes its change argument as a placeholder naming a
  guess
- **THEN** the guard over stage-mark calls rejects it, naming the file and line

#### Scenario: A mark names a change the store has never seen

- **WHEN** a mark carries a resolved change name the store does not hold
- **THEN** the synthetic change row is bootstrapped exactly as before, and the mark is recorded
  rather than dropped

