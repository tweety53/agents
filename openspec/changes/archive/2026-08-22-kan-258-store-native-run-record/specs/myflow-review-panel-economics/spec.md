## MODIFIED Requirements

### Requirement: Panel findings are recorded in a table, one row per finding

The review panel record SHALL carry every finding raised in a table, one row per finding, with
columns for the finding's identifier, the slot that raised it, its severity, its location, and a
note.

**The table SHALL be rendered from the store's finding rows, never hand-written.** A finding is
recorded by writing a row as the panel raises it; the table is a view of those rows. No skill SHALL
instruct an agent to author the table, and no run SHALL maintain a hand-written table alongside the
rows — one source, so the two cannot disagree.

The identifier SHALL be the **first** cell of every row and SHALL read `F<n>`, where `<n>` is a
decimal number **unique within the record**. That cell is how a row is paired with the marker line
described in the next requirement, and it is the only part of the table's shape any tool reads. An
identifier used by more than one row SHALL count as outstanding: two findings sharing one identifier
share one recorded state, and the second one's state is then unrecorded. **Uniqueness SHALL be
enforced by the store's own constraint on the finding's reference within its change**, so a duplicate
is refused at the write rather than detected at the read.

The table SHALL NOT carry a status column. A finding's state is recorded once, on its marker line,
and nowhere else. A status cell beside the marker is a second surface that can silently disagree
with the line that actually governs: the machine's direction is protected — a marker reading `open`
blocks regardless of what any cell says — but nothing protects a reader who sees `fixed` in the
table and takes it at face value. One statement of a fact, in one place.

The table is the record's **human-readable** rendering of its findings. It SHALL NOT be the record
a tool counts: no tool reads its column order, its header, its cell boundaries, or where it begins
and ends. A `|` inside a cell therefore needs no escaping and SHALL NOT change how any finding is
counted.

Free prose SHALL NOT be the record of a finding's state. A finding whose state can only be inferred
from surrounding text cannot be counted, and a bar that cannot be counted cannot be enforced.

**A fix round SHALL update an existing finding's row rather than append a second one.** The record
SHALL therefore never be cumulative, and a record's finding count SHALL always describe the change's
current findings rather than the union of every round's.

#### Scenario: Every finding has a row

- **WHEN** a panel slot raises a finding
- **THEN** a row for it exists in the panel record carrying its `F<n>` identifier, the slot,
  severity, location and note, and its state is on its marker line rather than in the row

#### Scenario: An identifier naming two findings counts as outstanding

- **WHEN** a second finding is recorded for a change under an identifier that change already uses
- **THEN** the write is refused by the store's constraint, so the record can no longer be rendered
  with two rows sharing one identifier
- **AND** were such a record to reach the guard by any other route, it still counts as outstanding
  and the reused identifier is still named — the rule is unchanged, the write is simply refused
  earlier

#### Scenario: The table's shape does not change any count

- **WHEN** the findings table carries an unescaped `|` inside a cell, a reordered header, a row
  written without its boundary pipes, or a row separated from the table by a blank line
- **THEN** the number of findings counted from the record is unchanged

#### Scenario: A fix round does not grow the record

- **WHEN** a fix round resolves `F1` and the record is re-rendered
- **THEN** the record carries exactly one row for `F1`, with its state updated
- **AND** the record's declared total counts the change's findings, not every round's findings summed

## ADDED Requirements

### Requirement: The rendered record satisfies the marker contract the guard reads

The renderer that produces the review panel record from the store SHALL emit the total line and the
marker lines exactly as the requirements governing them specify — same labels, same anchoring at the
first character of the line, same one-unbroken-span placement for the `finding-status:` block, and
the reproducer block kept separate from it.

**The marker contract SHALL NOT be altered by the move into the store, and the record's location
SHALL move with it.** These are two separate statements about two separate things, and both hold.

The **format** a guard parses SHALL be unchanged — the labels, their anchoring at the first character
of the line, the one-unbroken-span placement of the `finding-status:` block, the separate reproducer
block, and the declared totals — and a guard SHALL reach the same verdict on a rendered record that
it reaches on a hand-written one carrying the same findings.

The **location** a guard reads SHALL be the path the record is rendered to,
`<worktree>/docs/superpowers/reviews/<YYYY-MM-DD>-<change>-panel.md`. A guard SHALL resolve it by an
**anchored** match on both the date prefix and the change name — the same rule the renderer applies
when it reuses an existing dated file — so that one change's record can never be read as another's,
and SHALL read the same file the renderer would write.

A guard SHALL NOT read the panel record from `<worktree>/.superpowers/sdd/final-review-panel.md`, and
SHALL NOT fall back to it when the rendered record is absent. Once the findings are rows, nothing
writes a findings table or a marker block into that file: it survives as the pass log alone — the
mode, the slots, the diff path, the fix round's recorded mutations and the bounces — and declares no
findings total, so a guard reading it would report every change as carrying unfinished work.

**The record SHALL be rendered to one location, not two.** Writing it to both the rendered
destination and the worktree state directory SHALL NOT be done: two copies of one record can
disagree, which is the defect the move into the store exists to remove.

A guard that resolves the record from a change name SHALL validate that name against the allowlist
the renderer applies — one leading letter or digit, then letters, digits, `.`, `_` and `-` — before
the name is used to resolve any path, and SHALL refuse with its own cannot-answer status rather than
reach a verdict when the name is outside it.

The renderer SHALL NOT emit a marker label inside prose, a table cell, or a fenced block, since a
validly-formatted marker written as an example reads the same as a real one. Where a finding's note
or location contains text that would form such a label, the renderer SHALL neutralise it rather than
emit it verbatim.

**The record SHALL be rendered when the panel closes**, before any guard reads it — not at
integration. A guard that runs during implementation SHALL find a current record to read.

#### Scenario: The guard reads a rendered record

- **WHEN** the unfinished-work guard runs during `/myflow-do` against a record rendered from store
  rows
- **THEN** it parses the total line and the marker block exactly as it does a hand-written record,
  and reaches the same verdict

#### Scenario: The pass log is not read as a findings record

- **WHEN** a worktree carries a complete panel record at `.superpowers/sdd/final-review-panel.md` and
  none at the rendered destination
- **THEN** the guard reports that there is no review panel record, rather than reading the pass log
- **AND** a finding written into the pass log beside a clean rendered record reaches no verdict

#### Scenario: Another change's record is not this change's record

- **WHEN** the reviews directory holds a dated record whose change name merely ends in the name being
  resolved, and none for that name
- **THEN** the guard reports that there is no review panel record for it

#### Scenario: A finding's note would form a marker label

- **WHEN** a finding's note or location contains text matching a marker label
- **THEN** the renderer neutralises it, and the record's counted findings are unchanged

#### Scenario: The record exists before the guard runs

- **WHEN** the review panel closes
- **THEN** the record is rendered at that point, so the guard running later in the same run has a
  current record to read
