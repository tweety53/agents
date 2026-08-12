# myflow-review-panel-economics Specification

## Purpose
TBD - created by archiving change kan-10-myflow-economical-updates. Update Purpose after archive.
## Requirements
### Requirement: Every review-panel slot runs on Sonnet

Every review-panel slot SHALL run on **Sonnet** by default. The roster and the trigger table that
selects optional slots are unchanged; only model selection is.

Slots the panel spawns directly — the primary reviewer, the principles reviewer, the adversarial
reviewer, and any extra principle lens — SHALL pass Sonnet explicitly. Slots dispatched by
`subagent_type` (Bugbot, Security Review) carry their own agent definitions and SHALL NOT be
given a model override.

No slot SHALL inherit the parent model, and no slot SHALL resolve a model from the parent's
provider family.

The default SHALL yield to an explicit operator override, and only to that — whether the operator
recorded a review-panel model in the change's state file or instructed one during the session.
`myflow-model-policy` is canonical for how such an override is given, recorded and applied; this
requirement defers to it rather than restating the mechanism.

**Only the slots that take an override are affected.** Bugbot and Security Review are dispatched by
`subagent_type` and SHALL still receive no model override, whatever is recorded.

The default is Sonnet for the reason it always was, and an override does not weaken it: a reviewer
supplies one of many independent readings of a finished diff, so the panel's cost must not scale
with whichever model the operator happens to be running. An override is a deliberate, recorded
decision for one change, not an inheritance path — which is what "no slot SHALL inherit the parent
model" continues to forbid.

#### Scenario: Every directly-spawned slot names Sonnet

- **WHEN** the panel dispatches the primary, principles, adversarial or extra-lens reviewers for a
  change that records no panel model
- **THEN** each is given Sonnet explicitly, and none omits the model in order to inherit the
  parent's

#### Scenario: A stronger parent model does not raise the panel's cost

- **WHEN** the parent agent is running on Opus and no override was given
- **THEN** the panel still runs entirely on Sonnet

#### Scenario: A recorded override raises the panel

- **WHEN** the change's state file records a review-panel model other than Sonnet
- **THEN** the slots that take an override are dispatched on that model
- **AND** this is an override rather than inheritance, because the parent's model played no part in
  selecting it

#### Scenario: An override does not reach the subagent_type slots

- **WHEN** a panel model is recorded and Bugbot or Security Review is dispatched
- **THEN** no model override is passed to that slot

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

### Requirement: Bugbot's brief includes reasoned mutation testing

Where the panel dispatches Bugbot, its dispatch prompt SHALL include a mutation-testing brief: for
each behaviour the diff changes, mutate it — flip a condition, drop a guard, move a boundary, remove
a branch — and establish whether an existing test fails as a result. A mutation no test catches is a
**surviving mutant**.

This SHALL be reasoned mutation testing performed by the reviewer. No mutation-testing framework
SHALL be added, adopted or executed by this requirement, and no mutation score SHALL be computed.

The brief SHALL be carried by the dispatch prompt rather than by the agent definition. Bugbot is
dispatched by `subagent_type` and carries its own definition, which the dispatcher does not edit;
the prompt is therefore the only place this instruction can live, and passing a model override
remains forbidden.

A surviving mutant SHALL be raised as an ordinary finding — an `F<n>` row in the findings table and
a marker line in the marker block — and SHALL block the handoff under the existing zero-open-findings
bar until a test is added or the operator withdraws the finding with a reason. It SHALL NOT be
recorded as an advisory note outside the findings table: a second class of finding that nothing
enforces would sit beside a bar that enforces every other one.

Presets that do not dispatch Bugbot SHALL NOT acquire the mutation reading by another route, and
this requirement SHALL NOT add a slot to any preset. `myflow-review-panel-roster` is canonical for
which presets dispatch Bugbot.

#### Scenario: The brief reaches Bugbot through its prompt

- **WHEN** the panel dispatches Bugbot
- **THEN** its prompt carries the mutation-testing brief
- **AND** no model override is passed to it

#### Scenario: A surviving mutant blocks the handoff

- **WHEN** Bugbot finds a changed behaviour whose mutation no existing test would catch
- **THEN** the panel record carries an `F<n>` row and an `open` marker line for it
- **AND** `/myflow-do` does not hand off until it is fixed or withdrawn with a reason

#### Scenario: No mutation tooling is introduced

- **WHEN** the panel runs with the mutation brief in force
- **THEN** no mutation-testing framework is installed or executed, and no mutation score is reported

#### Scenario: A preset without Bugbot gets no mutation reading

- **WHEN** a change records a preset whose required slots do not include Bugbot, and no conditional
  slot adds it
- **THEN** no slot is briefed for mutation testing

### Requirement: The panel measures the diff it is about to read

Before writing `final-review.diff` and dispatching any slot, `/myflow-do` SHALL measure the branch's
changed-line count — insertions plus deletions across committed work and the unstaged working tree,
the same body of text the panel will read — and SHALL compare it against a **cap**, defaulting to
**2000** changed lines.

The measurement SHALL be performed by a script taking the worktree, the merge base, and an optional
cap. It SHALL exit **0** at or under the cap, **1** over it, and **2** when it cannot answer at all —
a path that is not a worktree, an unresolvable merge base, or a cap that is not a number.

An exit of **2** SHALL stop the run rather than being read as under the cap. A size the guard could
not measure is not a size under the cap.

On an exit of **1**, `/myflow-do` SHALL put the choice to the operator as named options — proceeding
with the panel anyway as the default and recommended answer, and stopping so the change can be split
as the alternative. A stop SHALL end the run at `IN_PROGRESS` with the implementation intact, exactly
as a fix round does; it SHALL NOT discard work.

The measured count, the cap in force, and the operator's answer where one was given SHALL be recorded
in the panel record on **every** run, including runs at or under the cap, so the record always states
the size the panel read.

The cap SHALL NOT move the handoff bar. Proceeding past it SHALL change nothing about the panel's
roster, its slots, its escalation ladder, or the requirement of zero open findings at any severity.
The cap governs how much text one pass reads, never what clears it.

The script SHALL NOT be a lint step: it needs a worktree and a merge base for a change in flight, the
same reason `check-finish-preflight.sh` and `check-unfinished-work.sh` are already excluded from
lint. It SHALL be covered by its own test harness.

#### Scenario: A diff under the cap runs the panel unchanged

- **WHEN** the branch diff measures fewer changed lines than the cap
- **THEN** the panel runs exactly as it does today, and the panel record states the measured size

#### Scenario: A diff over the cap asks the operator

- **WHEN** the branch diff measures more changed lines than the cap
- **THEN** the operator is offered proceeding with the panel — the default and recommended answer —
  or stopping to split the change

#### Scenario: Stopping does not discard work

- **WHEN** the operator chooses to stop and split
- **THEN** the run ends at `IN_PROGRESS` with the implementation committed on the branch

#### Scenario: An unmeasurable diff is not treated as small

- **WHEN** the measuring script exits 2
- **THEN** the run stops rather than dispatching the panel

#### Scenario: Proceeding past the cap does not lower the bar

- **WHEN** the operator proceeds with an over-cap diff and a slot raises a minor finding
- **THEN** the handoff is blocked exactly as it would be under the cap

### Requirement: A conditional slot re-runs only when its own subject changed

In the panel's **Full** escalation mode, `/myflow-do` SHALL re-run every **required** slot against
the rewritten `final-review.diff`, and SHALL re-run a **conditional** slot — Security, Adversarial,
or an extra principles lens — only when that slot's own row in the optional-slot trigger table still
fires against the fix diff.

A conditional slot whose trigger did not fire SHALL NOT be re-run, and its previous pass's result
SHALL stand, recorded in the panel record as not re-run because its subject was unchanged. The record
SHALL distinguish that outcome from a slot that was re-run and from a slot whose trigger never fired
at all.

This SHALL be a definition of what **non-stale** means for a slot with a bounded subject, not a
waiver of the requirement that the final pass show a non-stale clean result for every slot in the
roster. A result is stale when the diff it read has since changed in the region that slot reads; a
conditional slot's region is exactly its trigger's subject, so a fix touching nothing in that subject
leaves its reading current. A required slot has no bounded region — it reads the whole diff — which
is why this scoping SHALL NOT reach any required slot.

The zero-open-findings bar SHALL be untouched. A conditional slot that raised an open finding SHALL
still block the handoff whether or not it re-runs, and not re-running a slot SHALL NOT close, soften
or expire any finding it has already raised.

**Targeted** mode SHALL be unchanged; it is already scoped to the agents that raised findings plus
the integration check.

#### Scenario: An untouched subject is not re-read

- **WHEN** a full escalation pass follows a fix that touched no auth, token, crypto, secret, config,
  query-construction, path-handling, deserialization, HTTP-edge or dependency code
- **THEN** the Security slot is not re-run, and the panel record says its subject was unchanged

#### Scenario: A touched subject is re-read

- **WHEN** the fix diff modifies a migration and the Adversarial slot's trigger names migrations
- **THEN** the Adversarial slot re-runs against the rewritten diff

#### Scenario: Required slots are never scoped out

- **WHEN** a full escalation pass runs after a one-line fix
- **THEN** every required slot in the roster re-runs regardless of what the fix touched

#### Scenario: A slot that is not re-run cannot clear its own finding

- **WHEN** a conditional slot raised an open finding on an earlier pass and its trigger does not fire
  on the fix diff
- **THEN** that finding remains open and the handoff is still blocked

#### Scenario: The record tells the two silences apart

- **WHEN** one conditional slot's trigger never fired at all and another's fired on pass 1 but not on
  the fix diff
- **THEN** the panel record distinguishes the slot that was never selected from the slot that was
  selected and not re-run

