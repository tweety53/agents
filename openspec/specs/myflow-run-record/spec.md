# myflow-run-record Specification

## Purpose
TBD - created by archiving change kan-258-store-native-run-record. Update Purpose after archive.
## Requirements
### Requirement: A change's derived records live in the store, not in worktree files

A change's **derived** records — every subagent dispatch and every review-panel finding — SHALL be
recorded in the stats store. They SHALL NOT be authored as files in a worktree.

A record SHALL be attached to the change by the store's own `change_id`, never by a filesystem path.
There SHALL therefore be no path for two components to disagree about, and "the store holds no rows
of this kind for this change" SHALL be a distinct, reportable value from "a record exists somewhere
that was not read".

**Authored artifacts SHALL NOT move.** `tasks.md`, `proposal.md`, `design.md` and the delta specs
remain Markdown in git. Task **completion** SHALL remain recorded by `tasks.md`'s checkbox alone; the
store SHALL NOT become a second source of completion state, and no guard SHALL read completion from
it.

#### Scenario: A change has no record of a kind

- **WHEN** a render is asked for a kind the store holds no rows of for that change
- **THEN** it reports that no rows exist, naming the change and the kind, and exits 0 — a value
  distinct from any failure

#### Scenario: A worktree is removed

- **WHEN** finish run 2 removes a change's worktree
- **THEN** no dispatch record and no finding is lost with it, because neither was ever stored there

### Requirement: One dispatch row carries task, commit, model and cost together

A dispatch's task, its role, the model it ran on, the commit it produced and the tokens it spent
SHALL be recorded as **one row**, not as separate records in separate places.

Each dispatch row SHALL carry: the change, the stage run it belongs to where one is known, its
append order within the change, the task id where the role runs against a task, the role, the panel
slot where the role is a panel slot, the model, the commit sha, the outcome, the run's session token,
the start instant, the end instant, and a metrics bag.

The append order SHALL be unique per change, and the constraint enforcing it SHALL be the race
detector: two concurrent writes that compute the same next value SHALL collide, and the loser SHALL
retry rather than silently overwrite the winner.

The row SHALL be written in two calls: one as the dispatch **starts**, carrying everything known
then, and one as it **finishes**, carrying the commit, the outcome and the end instant. The opening
call SHALL NOT be deferred to the close. Cost attribution reads each transcript record exactly once,
at the harvest cycle that consumes it, and never re-reads behind the offset it has committed — so a
row that does not exist while the dispatch is running has no window for that cycle to attribute to,
and the usage is lost or credited to an unrelated earlier dispatch.

The closing call SHALL record the end instant. A dispatch whose end instant is absent SHALL be
treated as still running, so a dispatch that is never closed SHALL NEVER stop being a candidate for
later usage.

A dispatch SHALL carry a caller-written key, unique within the run its session token names, and the
two calls SHALL name the same row by it rather than by the store-allocated append order — which a
caller whose opening write was journalled has never been told. That key SHALL make the opening write
**idempotent**: the same dispatch recorded twice SHALL be one row, so a write replayed after a lost
response cannot double-count one dispatch's cost.

#### Scenario: A dispatch is still running when the harvester passes over it

- **WHEN** a harvest cycle processes transcript records timestamped inside a dispatch that has not
  finished
- **THEN** that usage is attributed to the dispatch, because its row and its attribution window
  already exist

#### Scenario: A dispatch has finished

- **WHEN** transcript records fall after a dispatch's end instant and inside no other dispatch
- **THEN** they are attributed to no dispatch, rather than to the one that has already closed

#### Scenario: An opening write is replayed after a lost response

- **WHEN** the same dispatch's opening write reaches the store twice under the same key
- **THEN** the change holds exactly one row for it, carrying the append order the first write
  allocated

#### Scenario: Two dispatches are recorded concurrently

- **WHEN** two writers record a dispatch for the same change at the same instant
- **THEN** both rows exist with distinct append positions, and neither has overwritten the other

#### Scenario: A dispatch's model is not knowable

- **WHEN** a slot dispatched by `subagent_type` resolves its model from its own agent definition
- **THEN** the row records the literal `unknown (agent-defined)` and never a plausible-looking slug

### Requirement: A finding is a row whose status is updated in place

Each review-panel finding SHALL be one row carrying the change, the dispatch that raised it, its
reference, the round it was raised in, the slot, the severity, the location, the note, its status and
its reproducer.

A finding's reference SHALL be unique **per change**, not per round. A fix round SHALL update an
existing finding's status in place rather than appending a second row for the same reference, so the
record of a change's findings SHALL NEVER be cumulative.

#### Scenario: A fix round resolves a finding

- **WHEN** a fix round fixes the finding referenced `F1`
- **THEN** `F1`'s single row has its status updated, and the change still holds exactly one row for
  `F1`

### Requirement: A record write never blocks the pipeline

A record write SHALL NEVER block, delay or alter the work it records. On any store failure the CLI
SHALL journal the intent, print one warning line, and exit 0 — the same guarantee a state write and a
stage mark already give.

A caller SHALL NOT branch on the record command's exit code as a signal about the record: a write
that could not reach the store still exits 0.

The only non-zero exit SHALL be a caller mistake — which is a defect in the call, fixed by
correcting the call rather than worked around. A caller mistake is one of two kinds, and both exit
non-zero without journalling:

- one the CLI itself judges, before any network call: an unrecognised role, a missing required flag,
  or a session token carrying a shell substitution;
- one **the daemon** judges, having been reached and having refused the request — a malformed body,
  or a reference that names nothing.

The second kind SHALL NOT be journalled, for the same reason it exits non-zero: the store was
reached and refused, so a replay would be refused identically, forever. Only a write that could not
reach the store is journalled.

**One refusal is exempt**, and only one: a dispatch's closing write naming a key the store holds no
dispatch under SHALL be journalled and replayed rather than refused. Unlike a reference that names
nothing, this refusal has an ordinary transient cause — the opening write for that same key may
itself be journalled and still queued ahead of it — and discarding the close would leave the
dispatch's attribution window open permanently, which is the failure the closing call exists to
prevent. The reconciler SHALL replay journal entries for one change in the order they were written,
which is what resolves the pair.

The journal SHALL reuse the existing journal-append mechanism and its file naming convention, and
the reconciler SHALL replay record entries alongside the entries it already replays.

#### Scenario: The daemon is unreachable during a dispatch

- **WHEN** a dispatch is recorded and the store cannot be reached
- **THEN** the intent is journalled, one warning line is printed, the command exits 0, and the
  dispatch proceeds unaffected

#### Scenario: The daemon refuses the request

- **WHEN** a record write reaches the store and the store refuses it as malformed, or names a
  reference that does not exist
- **THEN** the command exits non-zero, nothing is journalled, and the refusal is reported — a replay
  of it could never succeed

#### Scenario: A dispatch is closed before its opening write has landed

- **WHEN** a dispatch's closing write reaches the store and names a key no dispatch exists under
- **THEN** the intent is journalled and the command exits 0, and a later replay applies the opening
  write first and the close after it

#### Scenario: A journalled record write is replayed

- **WHEN** the reconciler runs after the daemon becomes reachable again
- **THEN** the journalled record write is applied to the store and retired from the journal

### Requirement: A never-block guarantee is never also a silent one

The `IN_PROGRESS` handoff SHALL name how many record writes for the change are still sitting in the
journal, so that a record which did not reach the store is visible to the operator at the gate they
actually read.

A run whose record writes all reached the store SHALL say so rather than omitting the line, so that
its absence can never be mistaken for a clean run.

#### Scenario: A run journalled two record writes

- **WHEN** `/myflow-do` hands off after two record writes fell back to the journal
- **THEN** the handoff names that two writes are journalled

### Requirement: Every Markdown record is rendered from the store

The SDD ledger and the review panel record SHALL be **generated output**. No skill SHALL instruct an
agent to hand-write either of them, and no run SHALL write both a file and a store row for the same
record — one source, so two cannot disagree.

The renderer SHALL write into the repository directly. There SHALL be no intermediate copy in a
worktree, and no step that copies a rendered record from one path to another.

The destination date SHALL be fixed at the **first** render for a change: an existing rendered file
for that change SHALL be reused, so a fix round overwrites in place rather than leaving one dated
duplicate per round.

The renderer SHALL validate the change name against an allowlist of the characters a real change name
uses, and SHALL require the destination to be contained within the repository root. Both checks SHALL
be explicit in the renderer, never an incidental consequence of how a path is assembled.

**A rendered record SHALL be named to match the archive it joins**, not to a convention of its own. A
rendering that introduced a second naming convention into a directory already holding records under
one would make that directory unsearchable by the very name it is organised by.

**The panel record SHALL always be written when a panel closes, even when the panel raised nothing.**
Reporting "no rows" for it SHALL NOT be treated as equivalent to writing a record that declares zero
findings: the guard reading that record treats silence as outstanding, so an unwritten record would
report a clean change as unfinished. Reporting "no rows" remains correct for a kind whose absence is
a genuine fact rather than a declaration.

#### Scenario: A panel closes having raised no finding

- **WHEN** the panel record is rendered for a change whose panel raised nothing
- **THEN** a record is written declaring zero findings, and the guard that reads it reports the change
  clear rather than outstanding

#### Scenario: A change name carries a path separator or a glob metacharacter

- **WHEN** a render is asked for a change name carrying `/` or a glob metacharacter
- **THEN** the render is refused, names why, and writes nothing

#### Scenario: A change is rendered twice

- **WHEN** a fix round renders a record a change already has a rendered file for
- **THEN** the existing file is overwritten in place, and no second dated file is created

### Requirement: The panel record is rendered when the panel closes

The review panel record SHALL be rendered at panel close, before any guard reads it — not at finish
run 1.

The rendered record SHALL carry the same marker blocks, in the same format, that the guard reading it
parses today. **The format SHALL NOT be altered by this move, and the record's location SHALL move
with it.** These are two separate statements about two separate things, and both hold: a guard's
parsing is untouched, while the path it resolves the record from is the path the renderer writes to,
so a guard that hard-coded the old path is edited to read the new one.

#### Scenario: The unfinished-work guard reads a rendered record

- **WHEN** the guard runs during `/myflow-do` against a panel record rendered from the store
- **THEN** it parses the marker block exactly as it does a hand-written one, and reaches the same
  verdict

### Requirement: Per-dispatch cost is derived from the transcript, never self-reported

A dispatch row's model, role, task and slot SHALL be **recorded intent**, written by the dispatcher.
Its token figures SHALL be **derived**, attributed from the harness transcript by the same mechanism
that already attributes usage to a stage run.

Attribution to dispatch rows SHALL be a second, independent pass. It SHALL NOT change what the
existing stage-run attribution records, and the two SHALL be understood as different grains over the
same usage rather than a double count.

No agent SHALL be asked to report its own token consumption, because two of the three supported
harnesses write no transcript and the figure would be absent or guessed there.

Where two or more dispatches run **concurrently** against one session, their attribution intervals
overlap, and an interval alone cannot say which dispatch a record belongs to. A dispatch SHALL
therefore record the harness's own identifier for the subagent it dispatched, where the harness
exposes one.

**The identifier SHALL attribute a record regardless of the interval.** Where a record's reported
identifier equals a dispatch's recorded identifier, that record SHALL be attributed to that dispatch
whether or not its timestamp falls inside that dispatch's window. A dispatch's `started_at` and
`ended_at` are written by the dispatching agent at the moment it marks, so the window is approximate
by construction; the identifier is exact, and SHALL NOT be gated behind an approximation.

The interval rule SHALL remain the fallback, reached only where no identifier is reported on one side
or the other. A harness that exposes no such identifier SHALL NOT be degraded: the interval rule
remains correct for dispatches that do not overlap. An absent identifier SHALL mean "not reported"
and SHALL NEVER match another absent identifier.

**Where two dispatch rows record the same non-empty identifier, that identifier SHALL attribute a
record to neither of them.** An identifier naming two dispatches has said nothing about which one
incurred the record, and a tie broken by interval or by row order would reintroduce, one layer down,
the silent misattribution this rule exists to remove.

**Where the interval rule is reached and two or more windows contain the record, the record SHALL be
attributed to no dispatch at all**, and the candidates' cost SHALL be recorded as unattributed,
carrying the number of candidates that could not be told apart. Attributing to the latest-started
candidate SHALL NOT be done: it charges one dispatch with a whole concurrent group's spend while its
siblings record zero, and presents the result as a measurement. Splitting the record's tokens across
the candidates SHALL NOT be done either, because an even split is a fabricated measurement.

An ambiguity resolved to no dispatch SHALL NOT change what the stage-grain pass records for the same
usage. What stops is the split across dispatches, never the total.

#### Scenario: Two panel slots run concurrently

- **WHEN** two review-panel slots are dispatched at the same time against one session, and the
  harness reports an identifier for each
- **THEN** each slot's usage is attributed to its own dispatch, rather than both landing on whichever
  slot started later

#### Scenario: A record's identifier falls outside its dispatch's recorded window

- **WHEN** a record reports an identifier equal to a dispatch's recorded identifier, and its
  timestamp falls outside that dispatch's `started_at`–`ended_at` interval
- **THEN** the record is attributed to that dispatch

#### Scenario: Two dispatches record the same identifier

- **WHEN** two dispatch rows carry the same non-empty identifier and a record reports it
- **THEN** the record is attributed to neither, and the ambiguity is recorded

#### Scenario: Overlapping windows with no identifier to separate them

- **WHEN** a record falls inside two or more dispatch windows and no identifier is reported on either
  side
- **THEN** no dispatch is credited with it
- **AND** those dispatches record their cost as unattributed, naming how many candidates could not be
  told apart
- **AND** the stage run's own metrics are exactly what they would have been

#### Scenario: The harness reports no identifier

- **WHEN** a dispatch is recorded on a harness that exposes no subagent identifier, and its window
  overlaps no other
- **THEN** its usage is attributed by the interval rule, exactly as before, and the dispatch is not
  treated as degraded

#### Scenario: A subagent's usage is attributed

- **WHEN** the harvester processes a transcript containing sidechain usage inside a dispatch's window
- **THEN** that usage is accumulated into the dispatch's own metrics bag, and the stage run's metrics
  are exactly what the existing pass would have produced

### Requirement: File-based preservation of session records is retired

The script that copied session records out of a worktree into the repository SHALL be removed,
together with its test harness. No skill, contract or guard-presence list SHALL name it.

Its duty SHALL be discharged by the render, invoked directly. The pipeline stage that performed the
preservation SHALL keep its stage key and its position in the run, so that stage runs already
recorded under that key remain valid and the documented stage table needs no edit.

The render SHALL report one outcome per kind: that the record was rendered and where; that the store
holds no rows of that kind; or that the write was journalled because the store was unreachable. Each
SHALL exit 0. A non-zero exit SHALL keep its single existing meaning — a write attempted and refused
or failed — so a caller branching on exit status never reads an empty record as a failure.

#### Scenario: A contract still names the retired script

- **WHEN** a reference check runs after the retirement
- **THEN** no skill, contract, guard list or installer names the removed script

#### Scenario: Run 1 renders instead of preserving

- **WHEN** finish run 1 reaches the stage that used to preserve session records
- **THEN** it renders the change's ledger from the store into the repository, under the same stage
  key as before

### Requirement: A dispatch's identifier may be recorded when it becomes known

The dispatcher SHALL record the harness's subagent identifier on the opening call where it is known
at that point, and SHALL be able to record it on the closing call where it is not.

The opening call SHALL NOT be delayed in order to obtain an identifier. The row must exist before the
first harvest tick the dispatch runs through — otherwise that tick's usage is dropped or credited to
an earlier dispatch — so waiting for an identifier would reopen the loss the two-call shape closed.

Attribution SHALL read dispatch windows afresh on each cycle, so an identifier recorded on the
closing call still attributes usage read before it arrived.

An identifier SHALL NEVER be invented. A dispatch recorded without one is ordinary, not degraded.

#### Scenario: The identifier is known only at the close

- **WHEN** a dispatcher learns the subagent's identifier only as the dispatch finishes
- **THEN** it records it on the closing call, and the dispatch's usage is attributed by that
  identifier

#### Scenario: The opening call is not delayed

- **WHEN** a dispatcher cannot learn the identifier before dispatching
- **THEN** it writes the opening call without one rather than deferring it

### Requirement: The record says why a dispatch has no cost

A rendered record SHALL distinguish a dispatch that was never measured from one whose cost could not
be attributed, and SHALL name which of the two it is.

At least these SHALL be told apart:

- **not measured** — no transcript exists for this harness, or none has been read yet;
- **cost unattributed — session never bound** — the dispatch's session was never bound, so no
  transcript could be searched for it;
- **cost unattributed — the session token matched N sessions** — the token identified more than one
  session, so binding was refused rather than resolved by choosing;
- **cost unattributed — indistinguishable from N concurrent dispatches** — the record fell inside
  more than one window and no identifier separated them.

**The two ambiguities are separate states and SHALL NOT share a wording.** One counts *sessions* a
token matched; the other counts *dispatches* a record could not be told apart from. A single rendered
phrase covering both would put a session count behind the words "concurrent dispatches", which
misreports the measurement rather than merely abbreviating it. Where a count accompanies a state, the
rendered wording SHALL name the thing counted.

A command whose own run produced any dispatch in an unattributed state SHALL say so in its handoff,
in one line. A loss that is discoverable only by someone who already suspects it is not reported.

An unattributed state SHALL NEVER be rendered as a zero figure.

#### Scenario: A run whose session never bound

- **WHEN** a change's dispatches belong to a session that was never bound
- **THEN** the rendered record says the cost is unattributed because the session never bound, rather
  than `not measured`

#### Scenario: A run reports its own unattributed cost

- **WHEN** a run finishes and one or more of its dispatches are in an unattributed state
- **THEN** its handoff carries one line saying so

#### Scenario: A session token that matched more than one session

- **WHEN** a dispatch's session token identified more than one session, so binding was refused
- **THEN** the rendered record says the token matched more than one session, and names how many
- **AND** it does not render the concurrent-dispatch wording, whose count means something else

#### Scenario: A harness that writes no transcript

- **WHEN** a dispatch is recorded on a harness that writes no transcript at all
- **THEN** the rendered record says `not measured`, which is distinct from every unattributed state

