# myflow-review-panel-economics Specification

## Purpose
TBD - created by archiving change kan-10-myflow-economical-updates. Update Purpose after archive.
## Requirements
### Requirement: Every review-panel slot runs on Sonnet

Every review-panel slot SHALL run on **Sonnet**. The roster and the trigger table that selects
optional slots are unchanged; only model selection is.

Slots the panel spawns directly — the primary reviewer, the principles reviewer, the adversarial
reviewer, and any extra principle lens — SHALL pass Sonnet explicitly. Slots dispatched by
`subagent_type` (Bugbot, Security Review) carry their own agent definitions and SHALL NOT be
given a model override.

No slot SHALL inherit the parent model, and no slot SHALL resolve a model from the parent's
provider family.

#### Scenario: Every directly-spawned slot names Sonnet

- **WHEN** the panel dispatches the primary, principles, adversarial or extra-lens reviewers
- **THEN** each is given Sonnet explicitly, and none omits the model in order to inherit the
  parent's

#### Scenario: A stronger parent model does not raise the panel's cost

- **WHEN** the parent agent is running on Opus
- **THEN** the panel still runs entirely on Sonnet

### Requirement: No provider-family model mapping survives

There SHALL be no mapping from a parent provider family to an economy tier, because there is no
longer more than one tier. Surviving prose SHALL state only that every slot runs on Sonnet, and
SHALL NOT describe per-slot tier reasoning, provider detection, or a fallback slug.

#### Scenario: No provider-family table survives

- **WHEN** the apply skill and its reviewer prompts are searched
- **THEN** no table maps a parent provider family to a model slug, and no instruction tells the
  agent to detect its own provider

### Requirement: Skills that describe the panel defer to the canonical roster

Every skill that dispatches or describes the review panel SHALL state that all slots run on
Sonnet, and SHALL agree with the roster and trigger table in `skills/myflow-do/SKILL.md`, which is
canonical for the panel.

#### Scenario: The panel is described in exactly one place

- **WHEN** a skill needs to describe the panel
- **THEN** it defers to `skills/myflow-do/SKILL.md` rather than restating the roster


### Requirement: Panel findings are recorded in a table, one row per finding

The review panel record SHALL carry every finding raised in a table, one row per finding, with
columns for the finding's identifier, the slot that raised it, its severity, its location, and a
note.

The identifier SHALL be the **first** cell of every row and SHALL read `F<n>`, where `<n>` is a
decimal number **unique within the record**. That cell is how a row is paired with the marker line
described in the next requirement, and it is the only part of the table's shape any tool reads. An
identifier used by more than one row SHALL count as outstanding: two findings sharing one identifier
share one recorded state, and the second one's state is then unrecorded.

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

#### Scenario: Every finding has a row

- **WHEN** a panel slot raises a finding
- **THEN** a row for it exists in the panel record carrying its `F<n>` identifier, the slot,
  severity, location and note, and its state is on its marker line rather than in the row

#### Scenario: An identifier naming two findings counts as outstanding

- **WHEN** two rows of the findings table carry the same `F<n>`
- **THEN** the record counts as outstanding, and the reused identifier is named

#### Scenario: The table's shape does not change any count

- **WHEN** the findings table carries an unescaped `|` inside a cell, a reordered header, a row
  written without its boundary pipes, or a row separated from the table by a blank line
- **THEN** the number of findings counted from the record is unchanged

### Requirement: Every finding's state is recorded on an anchored marker line

Alongside the table, the review panel record SHALL carry one **marker line** per finding. A marker
line SHALL begin at the first character of its line and SHALL read exactly:

```
finding-status: F<n> <status>
```

where `F<n>` is the identifier of the row it belongs to. The marker line, not the table, SHALL be
the record of a finding's state.

`<status>` SHALL be one of exactly three values — `open`, `fixed`, or `withdrawn` — compared **byte
for byte**, never through the collating sequence of whatever locale a tool happens to run under. Any
other value SHALL count as outstanding: a different case such as `Open`, one of the three with
commentary after it such as `open (needs discussion)`, an empty value, and a value carrying a
zero-width or other invisible character are none of the three, and none of them SHALL be read as
closed.

`withdrawn` SHALL be followed by the reason the finding was retracted, on the same line. A
`withdrawn` marker with nothing after the status SHALL count as outstanding: the reason is part of
the state, and a status nothing checks is the escape hatch this bar exists to close.

A line that names `finding-status:` without being a marker line — indented, inside a blockquote,
missing its `F<n>` identifier, or written into prose — SHALL count as outstanding rather than be
skipped. The record SHALL NOT quote the marker format inside itself.

The findings table's identifiers and the marker lines' identifiers SHALL name the same findings. A
row with no marker, and a marker with no row, SHALL each count as outstanding. An identifier
carried by two marker lines SHALL count as outstanding on its own, and SHALL be reported even when
the findings table repeats the same identifier — the two lists agreeing is not evidence that either
is right.

The marker lines SHALL occupy **consecutive lines** — one unbroken block. A marker line written
anywhere else in the record, including inside a fenced example, SHALL count as outstanding. Without
this rule a marker quoted elsewhere stands in for a marker that was never written: the identifiers
match, the total matches, and an open finding reads as clear.

#### Scenario: A marker line records each finding's state

- **WHEN** a panel slot raises a finding and the record gives it row `F7`
- **THEN** the record also carries a line reading `finding-status: F7 open` until the finding is
  closed

#### Scenario: An unrecognised status counts as outstanding

- **WHEN** a marker line's status is anything other than `open`, `fixed` or `withdrawn`
- **THEN** that finding counts as outstanding

#### Scenario: The status is compared as bytes, not by collation

- **WHEN** a marker line's status carries a zero-width character
- **THEN** it counts as outstanding under every locale, and the verdict does not depend on which
  locale the tool ran under

#### Scenario: A withdrawal with no reason counts as outstanding

- **WHEN** a marker line reads `finding-status: F7 withdrawn` with nothing after the status
- **THEN** that finding counts as outstanding rather than closed

#### Scenario: A marker that does not begin its line counts as outstanding

- **WHEN** a line names `finding-status:` but is indented, quoted, or missing its identifier
- **THEN** the record counts as outstanding rather than that line being ignored

#### Scenario: The table and the marker lines must name the same findings

- **WHEN** a fix round adds a row to the findings table and does not write its marker line
- **THEN** the record counts as outstanding

#### Scenario: An identifier repeated on both sides is still outstanding

- **WHEN** two distinct findings are both given identifier `F1`, in the table and in the marker
  block alike
- **THEN** the record counts as outstanding, rather than the repeats cancelling out

#### Scenario: A marker written outside the block cannot stand in for a missing one

- **WHEN** a finding's marker line is missing and a marker line naming the same identifier appears
  elsewhere in the record, such as inside a fenced example
- **THEN** the marker lines are not consecutive and the record counts as outstanding

### Requirement: The panel record declares how many findings it carries

The review panel record SHALL carry exactly one **total line**, beginning at the first character of
its line and reading:

```
findings-total: <n>
```

where `<n>` is the number of findings the record carries, written in decimal with no leading zero.
The number of marker lines SHALL equal `<n>`; a disagreement in either direction SHALL count as
outstanding, and SHALL be reported with both numbers so the record can be corrected rather than
guessed at.

A record with **no** total line SHALL count as outstanding, whatever else it contains. Zero findings
SHALL NOT be inferred from a record that never states how many findings it has: a panel that raised
no finding says so with `findings-total: 0`, which is a declaration and clears, where silence is
not. A record declaring a total more than once, or declaring one that is not a plain count, SHALL
likewise count as outstanding.

#### Scenario: A record that never speaks cannot show zero findings

- **WHEN** the panel record exists but carries no total line
- **THEN** it counts as outstanding, however clean its prose

#### Scenario: A panel that raised nothing declares it

- **WHEN** no slot raised a finding
- **THEN** the record carries `findings-total: 0` and no marker lines, and counts as clear

#### Scenario: A total that disagrees with the marker lines counts as outstanding

- **WHEN** the record declares `findings-total: 12` and carries eleven marker lines
- **THEN** it counts as outstanding, and both numbers are named

### Requirement: No handoff while any finding is open, at any severity

`/myflow-do` SHALL NOT hand off while any finding in the panel record has status `open`, regardless
of severity. A minor finding SHALL block the handoff exactly as a critical one does.

The panel's existing escalation ladder SHALL remain the termination guarantee: when fix rounds do
not converge, the run hands back to the operator, who resolves the disagreement — including by
marking a finding `withdrawn` with a reason. That handback is the existing human gate and SHALL NOT
be turned into a routine way to defer findings.

Every round that changes a finding's state SHALL rewrite that finding's marker line in the same
edit as its table row, so that the record's state and its rendering are written together.

The final pass SHALL show a non-stale clean result for every slot in the run's roster.

#### Scenario: A single open minor blocks the handoff

- **WHEN** every critical and important finding is fixed but one marker line still reads `open`
- **THEN** `/myflow-do` does not hand off

#### Scenario: Withdrawal requires a reason

- **WHEN** a finding is closed by withdrawal rather than by a fix
- **THEN** the reason is recorded on its marker line and the withdrawal is visible in the panel
  record

#### Scenario: A stale clean result does not satisfy the bar

- **WHEN** a slot's last recorded result predates the most recent fix
- **THEN** that result does not count as clean and the handoff is still blocked
