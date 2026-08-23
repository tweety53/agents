# myflow-run-record delta — kan-212-persist-per-dispatch-cost-tokens-model-and-role

## MODIFIED Requirements

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

## ADDED Requirements

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
