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
`myflow-model-policy` is canonical for how such an override is given, recorded and applied.

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


In the panel's **Full** escalation mode, `/myflow-do` SHALL re-run every **required** slot — each
against the diff *A Full-mode re-run gives each slot only the diff it has not read* assigns it — and
SHALL re-run a **conditional** slot — Security, Adversarial, or an extra principles lens — only when
that slot's own row in the optional-slot trigger table still fires against the fix diff.

A conditional slot whose trigger did not fire SHALL NOT be re-run, and its previous pass's result
SHALL stand, recorded in the panel record as not re-run because its subject was unchanged. The record
SHALL distinguish that outcome from a slot that was re-run and from a slot whose trigger never fired
at all.

This SHALL be a definition of what **non-stale** means for a slot with a bounded subject, not a
waiver of the requirement that the final pass show a non-stale clean result for every slot in the
roster. A result is stale when the diff it read has since changed in the region that slot reads; a
conditional slot's region is exactly its trigger's subject, so a fix touching nothing in that subject
leaves its reading current.

**Trigger-based scoping SHALL NOT reach any required slot.** A required slot is never scoped out by
whether some trigger fired: it is scoped, if at all, only by its own delta being empty, which *A
required delta slot with an empty delta is not re-dispatched* governs. The two SHALL remain separate
mechanisms with separate recorded reasons.

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
- **THEN** the Adversarial slot re-runs, against the diff assigned to its slot

#### Scenario: Required slots are never scoped out by a trigger

- **WHEN** a full escalation pass runs after a one-line fix
- **THEN** every required slot in the roster whose delta is non-empty re-runs regardless of what the
  fix touched

#### Scenario: A slot that is not re-run cannot clear its own finding

- **WHEN** a conditional slot raised an open finding on an earlier pass and its trigger does not fire
  on the fix diff
- **THEN** that finding remains open and the handoff is still blocked

#### Scenario: The record tells the two silences apart

- **WHEN** one conditional slot's trigger never fired at all and another's fired on pass 1 but not on
  the fix diff
- **THEN** the panel record distinguishes the slot that was never selected from the slot that was
  selected and not re-run

### Requirement: The auto-escalate trigger set is stated and discriminates

`/myflow-do` SHALL escalate a panel re-run from **Targeted** to **Full** automatically, without
asking and stating the reason in the panel record, when any of the following holds:

- the fix touched a file outside the set named in the findings;
- the fix altered a delta spec, a migration, or a guard's behaviour;
- a targeted re-run surfaced a **new** Critical finding;
- three or more fix rounds have already run;
- the fix diff exceeds approximately 150 changed lines **and** the diff adds a new file.

**Every trigger in that set SHALL discriminate between fixes in the repository it runs in.** A
condition that fires on every fix round of every change in a given repository selects nothing there:
it converts the ladder's default from Targeted to Full while still reading as a trigger. Where a
condition is found not to discriminate, the correct repair is to narrow the condition, not to remove
the escalation.

The clause naming `a guard's behaviour` SHALL replace the earlier clause naming `a public contract`,
which did not discriminate in a repository whose product is contracts. The narrowed clause still fires
on a fix that changes a script's exit codes or its output, which is a real coverage reason and is not
implied by the delta-spec clause: a guard's behaviour can change with no spec edit at all.

**The size clause SHALL be an amplifier and SHALL NOT be an independent trigger.** A diff's length
carries no risk signal on its own — a mechanical rename is large and harmless, and a one-line change
to a guard's exit code is small and dangerous — so size alone selects almost as indiscriminately as
the vacuous clause that preceded this narrowing. Once the earlier clause was narrowed, size became
the dominant path to a full re-run, which is the same defect in a second place, and the same repair
applies: narrow the condition rather than remove the escalation.

**The signal size is paired with SHALL be `adds a new file`, and SHALL NOT restate the other risk
signals in this set.** Each of those — a delta spec, a migration, a guard's behaviour, a file outside
the set named in the findings — already escalates on its own clause above, so a conjunction naming
them reaches no case the ladder does not already reach and is dead text in a trigger list, which is
the same defect as a trigger that fires on everything. `adds a new file` is the one signal with no
clause of its own, and the gap it closes is real: a large body of brand-new, wholly unreviewed code
in a file the findings themselves named, which no other clause sees. A trigger set SHALL be checked
for this on every edit — a condition that can never fire independently is as much a defect as one
that always fires.

Escalation SHALL remain a coverage decision and never a waiver: a targeted re-run SHALL still dispatch
no fewer than two slots, and handoff SHALL still require zero open findings at any severity from every
slot that has run.

#### Scenario: A contract edit no longer forces a full re-run

- **WHEN** a fix round in a repository made of contract documents alters a contract file but no delta
  spec, no migration, and no guard's behaviour
- **THEN** the re-run is Targeted, and the panel record states that no escalation trigger fired

#### Scenario: A fix that changes a guard's behaviour escalates

- **WHEN** a fix round changes a guard script's exit codes or its reported output
- **THEN** the re-run escalates to Full, and the panel record names that trigger

#### Scenario: A trigger that fires on every fix is a defect in the trigger

- **WHEN** a trigger is observed to fire on every fix round of every change in a repository
- **THEN** the condition is narrowed, and the escalation itself is not removed

#### Scenario: A large mechanical fix carrying no risk signal stays Targeted

- **WHEN** a fix round's diff exceeds approximately 150 changed lines and adds no new file, and no
  other trigger in the set has fired
- **THEN** the re-run is Targeted, and the panel record states that no escalation trigger fired

#### Scenario: A large fix adding a new file escalates

- **WHEN** a fix round's diff exceeds approximately 150 changed lines and adds a new file, and no
  other trigger has fired — the new file's path is one the findings named, so the outside-the-set
  clause does not fire
- **THEN** the re-run escalates to Full, and the panel record names both the size and the new file

#### Scenario: A condition that can never fire independently is a defect in the trigger

- **WHEN** a trigger's condition is found to be reachable only when some other trigger has already
  fired
- **THEN** the condition is narrowed to the case it alone reaches, or removed, rather than left in
  the set as text that selects nothing

#### Scenario: Targeting is never a coverage waiver

- **WHEN** a Targeted re-run runs because no trigger fired
- **THEN** at least two slots are dispatched, and handoff still requires zero open findings at any
  severity

### Requirement: A fix instruction is verified by a failing reproducer before dispatch

Every slot the panel dispatches SHALL be required, in its dispatch prompt, to supply for each finding
it raises either a **runnable command** that demonstrates the defect, or the literal exemption form
`none — <reason>`. Slots dispatched by `subagent_type` SHALL receive this requirement as prompt text,
the way the mutation-testing brief already reaches Bugbot; no agent definition is edited to carry it.

The panel record SHALL carry the reproducers in a marker block of their own — one
`reproducers-total: <n>` count line and one `finding-reproducer: F<n> <command | none — reason>` line
per finding. That block SHALL be separate from the `finding-status:` block, whose lines are required
to occupy one unbroken span; interleaving the two SHALL NOT be done.

**A runnable `finding-reproducer:` command is constrained, and is dispatched as a direct exec with
an argument vector — without shell interpretation.** `scripts/run-reproducer.sh <worktree>
<reproducer-command-line>` is that argv-exec runner: it resolves the path token to a file that
already exists inside the worktree, executes it with its arguments passed as a real argument
vector — never through a shell — and no reproducer line is ever handed to a shell for
interpretation, regardless of what it contains. It SHALL be a bare path to a file that already
exists inside the worktree, naming a script or test target, optionally followed by plain arguments.
**Every token — the path and every argument — SHALL be resolved against the worktree and required
to pass the same containment test** the `## standards` entry resolution already applies: normalize
`..`, `.` and symlinks, then require the normalized result to stay inside the worktree, so that a
relative path with no lexical `..` segment that escapes only through a symlink is still caught; an
argument that fails containment (for example `../../../etc/passwd`) SHALL make the reproducer
unverifiable exactly as a failing path does. A reproducer line carrying a shell metacharacter
anywhere on it (`` | ; & $ ` < > ( ) { } ~ * ? [ ] # \ ' " ``), a leading `-` on the path token
specifically (not on every token — `script.sh --strict` SHALL NOT be rejected on that account), or a
URL SHALL NOT be run. A reproducer that fails any of these checks SHALL be recorded as unverifiable
and put to the operator, never run — the line naming it is subagent-authored text, and the reviewed
diff itself can carry the injection that reaches it. **The guard script that validates the panel
record SHALL itself reject a runnable reproducer line that is not this bare-path shape**, at exit 1,
naming the violated shape — the constraint SHALL NOT live in prose alone. This mechanical rejection
SHALL extend to a path token or argument that is **absolute** or that carries a **`..` path
segment**: the guard SHALL check this lexically, on the text of the line, and SHALL NOT resolve the
token against a filesystem to do it — the record may be read in a context where the worktree that
wrote it is absent — so it rejects only the shapes that cannot be contained inside any worktree; the
dispatch-time containment test above, run by `run-reproducer.sh` against the worktree that actually
exists at dispatch time, resolves the shapes that can be. A rejected reproducer line SHALL NOT be
silently rewritten to satisfy the guard: it is a refusal, not a defect to fix, since the line may
itself be the injection the guard exists to catch.

**Before dispatching the fix subagent**, `/myflow-do` SHALL run each supplied, constrained
reproducer command through `scripts/run-reproducer.sh <worktree> <reproducer-command-line>`, with
the worktree as the reproducer's working directory, under a bounded wait of **20 seconds** plus a
**2-second** SIGTERM-to-SIGKILL grace — the grace matching `scripts/check-cleanup-complete.sh`'s own
`SURVIVORS_KILL_GRACE`, the 20-second bound being this rule's own number rather than that script's
60-second `SURVIVORS_TIMEOUT`, since a reproducer is one short-lived check under test rather than a
wait for a whole cleanup pass. `run-reproducer.sh`'s own exit code SHALL be read as its documented
contract: **0** demonstrates the defect and clears the finding for dispatch; **1** means the
reproducer exited 0 and demonstrates nothing, so the instruction built on it cannot be verified as a
fix and the finding SHALL NOT be dispatched to the fixer on that pass; **2** is a refusal — the
command line failed a containment or shape check and was never executed at all — recorded
**unverifiable** and put to the operator; **3** is a timeout or a detached survivor — the reproducer
was killed at the bound and SHALL be recorded **unverifiable** and put to the operator, never read as
either a pass or a fail, since a timeout is not an exit; **4** means the run cannot answer at all and
is handled like an unreadable record. A reproducer still running at the bound SHALL be killed. A
reproducer whose process double-forks or detaches SHALL, before being recorded unverifiable, have the
same SIGTERM-then-grace-then-SIGKILL sequence retried against it, found by `run-reproducer.sh`'s own
measured mechanism rather than by a `ps` keyword this platform does not have (`ps -o sid=` does not
exist here, and `ps -o sess=` was measured to report `0` for every process tried); a child still
alive after that attempt SHALL be
named to the operator, by `run-reproducer.sh`'s own exit-3 report, as a **surviving process**, with
its pid, so it can be found and killed manually, rather than merely recorded unverifiable. A
reproducer killed at the bound, or one whose child survives that retry, may already have written to
the worktree; the worktree SHALL be re-checked (`git status`) before the run continues, and, for a
surviving process, SHALL be re-checked again when the operator resumes, since nothing between the
pause and the resume observed what the child wrote.

**`run-reproducer.sh` SHALL launch the reproducer in a process group of its own**, so that a
descendant which re-parents away from it remains findable and killable by group membership rather
than by parentage. Parentage is not a sufficient mechanism: a grandchild whose intermediate parent
exits within milliseconds re-parents to the init process before the first poll, and no
parentage-based lookup can see it afterwards. The group SHALL be established without a shell ever
seeing the reproducer line — the argv vector SHALL pass through untouched — so the direct-exec
guarantee above is preserved. Survivor detection SHALL consult the process group alongside the
descendant walk, and cleanup SHALL signal the group before falling back to that walk. Where the
mechanism that establishes the group is unavailable at run time, the run SHALL refuse — recorded
**unverifiable** and put to the operator — and SHALL NOT fall back to an ungrouped exec, since an
ungrouped exec is the exact condition in which the survivor goes unseen. A reproducer that leaves
the group deliberately, by calling `setsid` itself, is a stated residual limit documented beside the
implementation, and no clean result SHALL claim to have covered it.

A finding whose reproducer passed SHALL be **bounced once** to the slot that raised it, carrying the
reproducer's passing output. If that slot's second reproducer also passes, the run SHALL stop and put
the finding to the operator through the existing handback prompt — take another round, withdraw it
with a reason, or stop the run — and SHALL NOT dispatch it silently.

**Once the fix subagent reports, /myflow-do SHALL re-run every dispatched finding's reproducer**
under the same constraints and SHALL require it now to exit **0**. A reproducer that still exits
non-zero means the fix did not fix it, and the finding SHALL NOT be closed on that pass. **The
passing re-run alone SHALL NOT close a finding: the fix's diff SHALL also touch at least one path
the finding named.** A fix whose diff, for a given finding, touches only the reproducer's own target
and no path the finding named SHALL NOT be treated as a fix; the finding stays open and goes to the
operator through the handback prompt, carrying that fact as the reason.

**The disqualifying condition is "no path the finding named", and the two clauses SHALL be read
together rather than the first alone.** Where the reproducer's target *is* a path the finding named
— the ordinary shape for a finding about a guard script, whose reproducer is that script's own test
harness or the script itself — the fix is material, and a reading that disqualifies it inverts the
condition into one that fails the commonest correct case. Every restatement of this condition in a
skill or contract SHALL carry both clauses; a restatement that carries only the first is a defect in
the restatement, not a narrowing of the requirement.

A finding recorded `none — <reason>` SHALL be dispatched without a run: the rule binds findings
claiming a mechanical defect, and a principles, prose or naming finding has no runnable check to
demand. Where a slot **dispatched by `subagent_type`** supplies nothing at all, the parent SHALL
record `none — not supplied by <slot>` and SHALL dispatch the finding unverified, so the omission is
visible in the record rather than indistinguishable from a verified instruction. This exemption
SHALL NOT apply to a **general-purpose** slot: a general-purpose slot that supplies nothing SHALL be
recorded as its own open finding rather than as `none — not supplied by <slot>`, since the pipeline
fully controls that slot's prompt and it has no third-party excuse for the omission.

The presence of the field SHALL be enforced by a script taking the worktree, exiting **0** when every
`F<n>` named in the `finding-status:` block has exactly one well-formed `finding-reproducer:` line and
`reproducers-total` equals the number of those lines, **1** when violations are found, each reported
with its line, and **2** when it cannot answer at all. Exit **1** SHALL cover two distinct classes,
handled differently by the caller: a **missing or malformed field** (no `finding-reproducer:` line
for some `F<n>`, a malformed `reproducers-total:` count) is a record-completeness defect and the
missing field is added before any fix is dispatched; a **rejected reproducer shape** (a command
carrying a shell metacharacter, an absolute path, a `..` segment, a leading `-` on its path token, a
URL, or a NUL byte) is a refusal, recorded **unverifiable** and put to the operator exactly like a
failed dispatch-time containment check, and SHALL NOT be silently rewritten to satisfy the guard.
It SHALL read the marker block through anchored patterns and SHALL NOT parse the findings table. It
SHALL NOT be a lint step — it needs a worktree for a change in flight — and SHALL be covered by its
own test harness.

#### Scenario: A failing reproducer clears the finding for dispatch

- **WHEN** a finding's reproducer command exits non-zero in the worktree
- **THEN** the finding is dispatched to the fix subagent, and the record carries its reproducer line

#### Scenario: A passing reproducer bounces rather than dispatching

- **WHEN** a finding's reproducer command exits 0
- **THEN** the finding returns to the slot that raised it, carrying the passing output, and reaches no
  fix subagent on that pass

#### Scenario: A second passing reproducer reaches the operator

- **WHEN** a bounced finding's replacement reproducer also exits 0
- **THEN** the run stops at the operator handback with that finding's text and named options

#### Scenario: A prose finding is exempt with its reason stated

- **WHEN** a slot raises a finding about naming, prose, or a principle and records `none — <reason>`
- **THEN** the finding is dispatched without a reproducer run, and the record states the reason

#### Scenario: A slot supplying nothing is recorded as such

- **WHEN** a slot dispatched by `subagent_type` reports a finding with no reproducer and no exemption
- **THEN** the record reads `none — not supplied by <slot>` and the finding is dispatched unverified

#### Scenario: A general-purpose slot supplying nothing is its own finding

- **WHEN** a general-purpose slot reports a finding with no reproducer and no exemption
- **THEN** the omission is recorded as its own open finding, not as `none — not supplied by <slot>`

#### Scenario: An unconstrained reproducer line is never run

- **WHEN** a `finding-reproducer:` line is not a bare path to an existing worktree file, or carries a
  shell metacharacter, an absolute path token, a `..` path segment, a leading `-` on its path token,
  or a URL
- **THEN** the reproducer is never run, and the finding is recorded unverifiable and put to the
  operator

#### Scenario: The record guard rejects a non-bare-path reproducer line mechanically

- **WHEN** a `finding-reproducer:` line's command carries a shell metacharacter, an absolute path
  token, a `..` path segment, a leading `-` on its path token, or a URL
- **THEN** `scripts/check-panel-reproducers.sh` exits 1 against that record, naming the shape it
  violated, rather than accepting it at exit 0, without resolving the token against a filesystem to
  decide it

#### Scenario: A rejected reproducer shape is a refusal, not a defect to fix

- **WHEN** a `finding-reproducer:` line is rejected by the guard's mechanical shape check
- **THEN** the line is recorded unverifiable and put to the operator, and is never silently rewritten
  to satisfy the guard, since the line may itself be the injection the guard exists to catch

#### Scenario: An argument outside the worktree makes a reproducer unverifiable

- **WHEN** a reproducer command's path token is contained but an argument after it, once normalized,
  resolves outside the worktree
- **THEN** the reproducer is never run, and the finding is recorded unverifiable and put to the
  operator

#### Scenario: A reproducer killed at the bound is unverifiable, not a pass or a fail

- **WHEN** a reproducer command is still running when the 20-second bound plus its 2-second grace
  elapses
- **THEN** it is killed, and the finding is recorded unverifiable and put to the operator — its exit
  status is read as neither a pass nor a fail — and the worktree is re-checked (`git status`) before
  the run continues, since the reproducer may already have written to it before being killed

#### Scenario: A detached reproducer child that survives is named to the operator, not merely logged unverifiable

- **WHEN** a reproducer double-forks or detaches a child, and that process is not confirmed gone once
  the SIGTERM-to-SIGKILL grace elapses
- **THEN** the same SIGTERM-then-grace-then-SIGKILL sequence is retried against the reproducer's
  process group, and a child still alive after that attempt is named to the operator as a surviving
  process with its pid rather than merely recorded unverifiable, and the worktree is re-checked
  (`git status`) both now and again when the operator resumes

#### Scenario: A fast double fork is detected rather than missed

- **WHEN** a reproducer double-forks a grandchild whose intermediate parent exits before the first
  poll, so the grandchild has already re-parented and no parentage-based lookup can find it
- **THEN** the process-group lookup finds it anyway, and the finding is recorded unverifiable with the
  surviving process named, rather than reported as "defect not demonstrated"

#### Scenario: A reproducer that leaves the group itself is a stated limit, not a silent miss

- **WHEN** a reproducer calls `setsid` itself and thereby leaves the process group it was launched in
- **THEN** the limitation is documented beside the implementation, and no clean result claims to have
  covered it

#### Scenario: A fixed finding's reproducer must now pass

- **WHEN** the fix subagent reports and a dispatched finding's reproducer is re-run under the same
  constraints
- **THEN** an exit of 0 closes the finding for that pass, and a non-zero exit means the fix did not
  fix it

#### Scenario: A fix that only edits the reproducer does not close the finding

- **WHEN** a dispatched finding's reproducer flips from failing to passing, but the fix subagent's
  diff touches only the reproducer's own target and no path the finding named
- **THEN** the finding is not closed; it stays open and is put to the operator through the handback
  prompt, carrying that fact as the reason

#### Scenario: A fix to a path that is both the finding's target and the reproducer's is material

- **WHEN** a dispatched finding names a path, the finding's reproducer targets that same path — the
  ordinary shape for a finding about a guard script — and the fix's diff makes a non-comment,
  non-whitespace change to it
- **THEN** the fix is material and the finding closes on the reproducer's flip, since the materiality
  condition disqualifies only a diff that touches **no** path the finding named

#### Scenario: A missing reproducer line fails the guard

- **WHEN** the panel record names `F3` in the `finding-status:` block and carries no
  `finding-reproducer:` line for it
- **THEN** the guard exits 1 and names the finding

### Requirement: A bounced finding consumes no fix round

A bounce SHALL be accounted as a **guard-class** failure and not as a review finding, following the
accounting already required of `check-task-commit-fields.sh`: it SHALL NOT consume one of the review
loop's fix-round slots, and SHALL NOT advance the escalation ladder's count of rounds already run.

A bounce SHALL NOT close, soften or expire the finding it concerns. The finding stays open, and the
zero-open-findings bar governs it exactly as before — a bounce changes who is being asked to act on
it, never whether it must be resolved.

**Bounce accounting SHALL key off defect identity — file:line plus theme — never off the finding's
`F<n>` identifier.** The findings union already dedupes on that identity; keying the one-bounce
budget on `F<n>` instead would let the same defect reset its budget the moment it is re-raised under
a fresh identifier in a later pass, and the one-bounce-then-operator bound would never be reached.
**`file:line` is the finding's recorded location taken verbatim from the findings table; `theme` is
the finding's Note reduced to the shortest defect noun phrase within it, with severity words, slot
names and prose padding stripped — a reworded restatement of the same defect at the same file:line
reduces to the same theme, and this reduction SHALL be defined once and cited, not restated, by
every rule keyed on it.**

**Each bounce SHALL be recorded in the panel record's pass log entry, keyed by the finding's defect
identity and carrying the reproducer output that was returned to the raising slot**, so that a run
resumed after an interruption can see that a finding has already had its bounce and goes to the
operator instead of bouncing it again. The pass log entry's own definition (mode, agents, reason,
diff path read) SHALL be amended to name this bounce field explicitly, so that a pass which recorded
a bounce is not left pointing at a field the log's definition never described. This SHALL be recorded
in the pass log this section already requires, and SHALL NOT be a third marker block or a new guard
requirement.

#### Scenario: A re-identified defect does not reset its bounce budget

- **WHEN** a finding is bounced once under identifier `F3`, and the same defect — the same file:line
  and theme — is raised again in a later pass under a fresh identifier
- **THEN** the bounce budget for that defect identity is already spent, and a second passing
  reproducer reaches the operator handback rather than bouncing again

#### Scenario: A resumed run sees a prior bounce

- **WHEN** a run is interrupted after a finding has been bounced once, and is resumed in a later
  session
- **THEN** the panel record's pass log shows that defect identity already bounced, and a second
  passing reproducer for it goes straight to the operator handback

#### Scenario: Two bounces do not escalate the panel

- **WHEN** two findings are each bounced once for a passing reproducer, and no fix round has yet run
- **THEN** the escalation ladder's round count stays at zero and the next re-run is Targeted

#### Scenario: A bounced finding still blocks the handoff

- **WHEN** a finding is bounced and the run reaches its handoff check
- **THEN** the finding is still open and the handoff is blocked

### Requirement: Every panel slot is dispatched with the shared context bundle

Each review-panel slot's dispatch prompt SHALL name the absolute path of
`<worktree>/.superpowers/sdd/dispatch-context.md` and SHALL state that the slot need not go looking
for the change's proposal, design, plan, delta specs or engineering principles.

The same prompt SHALL require the slot to read the review diff itself. The bundle carries planning
context and never the diff's content, so no slot may treat the bundle as evidence about the code
under review.

This requirement SHALL NOT change which slots a roster dispatches, that slots are dispatched
separately and never merged into one prompt, the model each slot runs on, the reproducer each finding
must carry, or the zero-open-findings handoff bar.

#### Scenario: Three slots, one bundle, three dispatches

- **WHEN** a `light` roster dispatches its three required slots
- **THEN** each is a separate dispatch naming the same bundle path, and no two slots are merged into
  one prompt

#### Scenario: A slot dispatched with no bundle available

- **WHEN** the gather script could not produce a bundle
- **THEN** the slot is dispatched with the prompt shape used before this capability existed, and the
  panel is neither reduced nor skipped

### Requirement: The fix dispatch carries each finding's located evidence

The dispatch that repairs panel findings SHALL carry, for every surviving finding, that finding's
identifier, the slot that raised it, its severity, its `file:line` taken verbatim from the findings
table's Location column, its theme, the text of its `finding-reproducer:` line, and any bounce
recorded against its defect identity. It SHALL also name the shared context bundle.

The dispatch prompt SHALL state that these locations were established by the slot that raised them
and are not to be re-derived. The evidence already exists on the other side of the dispatch boundary;
this requirement is what carries it across.

Source excerpts around a finding's `file:line` SHALL NOT be inlined into the dossier. The fix round
edits the code it is given, so an inlined excerpt would be invalidated by the fixer's own work, and a
fixer reasoning from a stale excerpt is a correctness risk that the location alone does not carry.

This requirement SHALL NOT change the reproducer verification that precedes dispatch, the
re-run-and-materiality condition that closes a finding, the bounce accounting, the operator handback,
or the model the fix subagent runs on.

#### Scenario: A surviving finding reaches the fixer located

- **WHEN** a finding at `scripts/foo.sh:42` survives its review pass
- **THEN** the fix dispatch names that location, its theme and its reproducer, rather than a bare
  restatement of the finding's prose

#### Scenario: The fixer reads the real file

- **WHEN** the fix subagent is dispatched
- **THEN** its prompt carries no source excerpt for any finding, and it opens the named file itself

#### Scenario: A bounced finding carries its bounce

- **WHEN** a finding has already been bounced once against its defect identity
- **THEN** the dossier states that bounce alongside the finding

### Requirement: A fix round mutation-proves every behaviour it changes

A panel fix round SHALL mutation-prove every **executable** behaviour its fix changes, not only the
test cases the round adds. Executable behaviour means anything a test could fail on — a guard
script, Go, TypeScript, shell. A fix round that changes only prose or markdown has no executable
behaviour to mutate and SHALL record the exemption form rather than recording nothing.

Each mutation SHALL alter **one** mechanism. Where a single revert would change state that more
than one check reads, the round SHALL split it into surgical mutations, one per mechanism. A
mutation touching shared state can produce a **false pass** — a case that passes by
cross-contamination rather than by the mutated mechanism working — and a false pass reports coverage
that does not exist, which is worse than recording no mutation at all.

The **parent** SHALL perform the mutations, after the fix subagent reports. The fix subagent SHALL
name the executable behaviours its fix changed; the parent SHALL then mutate each one and establish
that an existing test fails. A round SHALL NOT certify its own mutations, for the same reason
`/myflow-do` already re-runs every dispatched finding's reproducer itself rather than accepting the
subagent's account of its own success.

Where a mutation survives — no existing test fails — the round SHALL add the test that catches it
before the round closes. A surviving mutation found by this requirement SHALL NOT be raised as a new
finding and SHALL NOT consume an additional panel pass; this is deliberately a different disposition
from a surviving mutant Bugbot reports, because the round already has the changed behaviour in hand
and a finding would spend a full round re-discovering it.

This requirement SHALL bind every roster preset, including one that dispatches no Bugbot. The
obligation belongs to the fix round, not to a slot. It SHALL NOT add a slot to any preset and SHALL
NOT change any preset's required set. It SHALL NOT introduce a mutation-testing framework, and no
mutation score SHALL be computed.

This requirement SHALL bind a fix round's own changes and nothing else. It SHALL NOT reach task
implementation, the red-task-partner's fixups, or guard changes generally.

#### Scenario: A structural change alongside added cases is proved too

- **WHEN** a fix round both adds test cases and changes a mechanism those cases are written
  alongside — a delimiter, a protocol, a control flow
- **THEN** the parent mutates the changed mechanism as well as running the added cases
- **AND** the round does not close on the added cases passing alone

#### Scenario: A revert touching shared state is split

- **WHEN** reverting a changed behaviour in one step would also alter state a second check reads
- **THEN** the round splits it into surgical mutations, one per mechanism
- **AND** no single mutation is relied on to prove more than one mechanism

#### Scenario: A surviving mutation is repaired inside the round

- **WHEN** the parent mutates a changed behaviour and no existing test fails
- **THEN** the round adds a test that fails under that mutation before it closes
- **AND** no new `F<n>` finding is raised for it and no additional panel pass is spent

#### Scenario: A prose-only fix round

- **WHEN** a fix round changes only prose or markdown
- **THEN** it records the exemption form rather than recording nothing

#### Scenario: A preset without Bugbot is bound too

- **WHEN** a change records a roster preset whose slots do not include Bugbot
- **THEN** its fix rounds carry this obligation unchanged
- **AND** no slot is added to that preset

### Requirement: The pass log records each fix round's mutations

The parent SHALL record each fix round's mutations in the pass log entry it already writes in
`.superpowers/sdd/final-review-panel.md`, one line per executable behaviour the round changed,
followed by a declared count:

```
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

The exemption SHALL use the `none — <reason>` literal form the record already uses for
`finding-reproducer:`, so the record carries one exemption shape rather than two.

These lines SHALL be written in the pass log entry and SHALL NOT be written inside the marker block.
`check-unfinished-work.sh` requires every `finding-status:` marker to occupy one unbroken run of
consecutive lines, and derives the findings table's identifiers from lines matching
`^\|?[[:space:]]*F[0-9]+[[:space:]]*\|`. A line that split that block, or that matched that row
shape, would change a guard's verdict on a record this requirement does not otherwise touch.

The record SHALL NOT be read by a guard script. This requirement SHALL NOT add one, and SHALL NOT
extend an existing guard to parse these lines. What the record exists for is to make the question
askable by the next round's reviewers, which is how the omission this requirement addresses was
found in the first place.

The fix round SHALL NOT close, and `/myflow-do` SHALL NOT reach the handoff, while an executable
behaviour the round changed carries neither a `fix-mutation:` line nor an exemption. This is the
round's own completeness condition, checked where the parent writes the record; it SHALL NOT be
expressed as a new class of finding and SHALL NOT add a gate to the handoff sequence.

#### Scenario: The lines land in the pass log

- **WHEN** a fix round that changed executable behaviour completes
- **THEN** its pass log entry carries one `fix-mutation:` line per changed behaviour and a
  `fix-mutations-total:` count
- **AND** those lines sit outside the marker block

#### Scenario: The existing guard's verdict is unchanged

- **WHEN** `check-unfinished-work.sh` reads a panel record carrying `fix-mutation:` and
  `fix-mutations-total:` lines
- **THEN** its verdict is identical to its verdict on the same record without them

#### Scenario: An undeclared behaviour holds the round open

- **WHEN** a fix round changed an executable behaviour and the pass log carries neither a
  `fix-mutation:` line nor an exemption for it
- **THEN** the parent does not close the fix round
- **AND** the run does not reach the handoff

#### Scenario: No guard is added

- **WHEN** this requirement is in force
- **THEN** no guard script reads `fix-mutation:` or `fix-mutations-total:`

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

### Requirement: No panel slot is dispatched onto a skill or agent that forks

A panel slot SHALL NOT be dispatched onto a skill or an agent that forks its own background agent to
do the work.

The parent dispatcher never observes a forked agent's completion. A slot dispatched that way
therefore cannot report through the panel's contract: its findings arrive on a surface the panel does
not read, or they do not arrive at all, and either way the slot itself neither returns a result nor
reliably ends.

This SHALL be read as a property of forking rather than of any one skill. Where a slot's intended
skill or agent is found to fork, the correct repair is to dispatch that slot on a shape that does
not fork — never to drop the slot, which **No preset moves the handoff bar** forbids.

#### Scenario: A forking skill is not dispatched

- **WHEN** a panel slot's intended skill or agent forks its own background agent
- **THEN** the slot is not dispatched onto it
- **AND** the slot runs on a shape that reports back to the dispatcher directly
- **AND** the slot is not dropped and the panel still runs its full required roster

### Requirement: A panel slot has a wall-clock ceiling and is checked for liveness

Every panel slot SHALL carry a wall-clock ceiling of **15 minutes** from its dispatch.

The dispatcher SHALL NOT block indefinitely on a slot's completion notification. It SHALL track each
in-flight slot's elapsed time and act on the ceiling itself, so that a slot which emits output while
making no progress — the case a stall watchdog cannot see — is still bounded.

This requirement SHALL be satisfied by whatever mechanism the running harness provides for observing
an in-flight dispatch, and no harness has to gain one for the requirement to hold. What is not
optional is that the dispatcher knows a slot's elapsed time and enforces the ceiling; a dispatcher
that waits on a notification alone has dropped the requirement rather than adapted it.

On a breach, the dispatcher SHALL, in order: stop the slot; close its dispatch row with
`-outcome timed-out`; record the breach in the panel record, naming the slot and the elapsed time;
and re-dispatch that one slot once. Other slots SHALL be unaffected.

A second breach of the same slot SHALL stop and put the choice to the operator — re-dispatch again,
proceed without the slot, or stop the run — and SHALL NOT be resolved silently.

That prompt SHALL take the shape **Operator prompts**
(`skills/myflow-contracts/operator-prompts.md`) fixes, which requires exactly one option marked
recommended and names it as what silence selects. **Stop the run SHALL be that option.**
Re-dispatching again on silence risks a loop, and proceeding without the slot on silence weakens
review, so neither may be what an unanswered prompt resolves to. Stopping ends the run at
`IN_PROGRESS` with the implementation committed and nothing lost.

A timed-out slot raises no finding, consumes no fix round, and SHALL NOT be counted as a clean
result. Where the operator chooses to proceed without a slot, the panel record SHALL name that slot
as not run, so a panel missing a slot is never indistinguishable from a clean one.

#### Scenario: A slot exceeds the ceiling and is re-dispatched

- **WHEN** a panel slot has run for 15 minutes without returning
- **THEN** the slot is stopped
- **AND** its dispatch row is closed with `-outcome timed-out`
- **AND** the panel record names the slot and its elapsed time
- **AND** that one slot is re-dispatched, with every other slot unaffected

#### Scenario: A second breach asks the operator

- **WHEN** the same slot exceeds the ceiling a second time
- **THEN** the dispatcher stops and asks the operator whether to re-dispatch again, proceed without
  the slot, or stop the run
- **AND** stop the run is the option marked recommended, and is what silence selects
- **AND** the breach is never resolved silently

#### Scenario: Proceeding without a slot is recorded, never silent

- **WHEN** the operator chooses to proceed without a timed-out slot
- **THEN** the panel record names that slot as not run
- **AND** the slot is not counted as a clean result for the final pass
- **AND** the zero-open-findings bar still governs every slot that did run

#### Scenario: A slot emitting output makes no difference

- **WHEN** a slot has emitted output within the last minute but has run for 15 minutes without
  returning
- **THEN** the ceiling is enforced exactly as for a silent slot

### Requirement: A Full-mode re-run gives each slot only the diff it has not read


In the panel's **Full** escalation mode, `/myflow-do` SHALL give slot 0, the primary reviewer, the
rewritten `final-review.diff`, and SHALL give every other slot that reads a diff file — Principles,
Code review (low), Adversarial, and every extra principles lens — that slot's own **delta** instead.

A slot's delta SHALL be `git diff <the HEAD sha that slot last reviewed> HEAD`, written to its own
file in the worktree's session directory. It SHALL be anchored at the sha that slot itself last
read, never at the current round: in **Targeted** mode only the integration check and the slots that
raised findings re-run, so a slot reaching a Full pass may have missed rounds in between, and the
current round's fix diff alone SHALL NOT be treated as covering them.

The range SHALL be expressed so that it compares two trees without requiring ancestry between them,
so that a base recorded before a `git rebase --autosquash` remains valid after that rebase rewrote
the task commit.

A slot for which no last-reviewed sha is held SHALL be given the whole `final-review.diff`. An
unrecorded base SHALL NOT be read as an empty or small delta.

Each slot's dispatch prompt SHALL state which of the two it is holding, naming the path and, for a
delta, the sha the delta starts from.

**Targeted** mode SHALL be unchanged: it already gives every re-running slot, slot 0 included, the
round's fix diff.

Slots dispatched by `subagent_type` — Bugbot and Security Review — read no diff file and SHALL be
unaffected.

This SHALL NOT move the handoff bar. Coverage holds because pass 1 always runs the full roster
against the full `final-review.diff`, so every slot has a real starting sha and every change made
since sits inside some slot's delta. There SHALL be no exemption for the pass that closes the panel:
the rule applies to every pass after the first.

#### Scenario: The primary slot keeps the whole diff

- **WHEN** a Full escalation pass runs after a fix round
- **THEN** slot 0 is dispatched against the rewritten `final-review.diff`

#### Scenario: A delta slot reads only what changed since it last read

- **WHEN** the Principles slot last read at sha `A` and a Full escalation pass runs at sha `C`
- **THEN** it is dispatched against `git diff A C`, and its prompt names sha `A` as the delta's base

#### Scenario: A slot that skipped a targeted round still sees it

- **WHEN** the Principles slot last read at round 1, rounds 2 and 3 were Targeted and did not re-run
  it, and round 4 escalates to Full
- **THEN** its delta spans rounds 2, 3 and 4, not round 4 alone

#### Scenario: A rewritten task commit does not invalidate a base

- **WHEN** a fix is folded into its task commit by `git commit --fixup` and `git rebase --autosquash`
  after a slot's base sha was recorded
- **THEN** that slot's delta is still computed against the recorded base, and reports the fix as part
  of what changed

#### Scenario: An unrecorded base falls back to the whole diff

- **WHEN** a slot is dispatched in a Full pass and no last-reviewed sha is held for it
- **THEN** it is given the whole `final-review.diff`

#### Scenario: The subagent_type slots are untouched

- **WHEN** a Full escalation pass runs under the `standard` or `full` preset
- **THEN** Bugbot is dispatched exactly as before, with no diff file and no delta

### Requirement: A required delta slot with an empty delta is not re-dispatched


In the panel's **Full** escalation mode, a required slot that reads a delta and whose delta is empty
SHALL NOT be dispatched. The panel record SHALL state that it was not re-run because nothing changed
since its last read, and SHALL distinguish that outcome from a conditional slot not re-run because
its trigger did not fire, from a slot the operator declined, and from a slot whose trigger never
fired at all.

This SHALL be a statement about staleness, not a waiver. A slot's own delta is its bounded region; an
empty delta means the reading it already gave is current. The zero-open-findings bar SHALL be
untouched — a slot that raised an open finding SHALL still block the handoff whether or not it is
re-dispatched — and not re-dispatching a slot SHALL NOT close, soften or expire any finding it has
already raised.

Slot 0 SHALL NOT be scoped out by this rule: it reads the whole diff and has no delta to be empty.

#### Scenario: Nothing changed since a slot last read

- **WHEN** a Full escalation pass runs and the Code review (low) slot already read at the current HEAD
- **THEN** it is not dispatched, and the panel record says it was not re-run because nothing changed
  since its last read

#### Scenario: The record tells the three silences apart

- **WHEN** one conditional slot's trigger did not fire, one required delta slot's delta is empty, and
  one slot was declined by the operator
- **THEN** the panel record states a distinct reason for each

#### Scenario: An open finding survives a skipped re-dispatch

- **WHEN** a delta slot raised an open finding on an earlier pass and its delta is empty on this pass
- **THEN** that finding remains open and the handoff is still blocked

### Requirement: A relocation-declaring plan gets a mechanical passage comparison

`tasks.md`'s header SHALL carry an explicit `**Relocation:** yes — <one-line reason>` or
`**Relocation:** no` field. This field SHALL be required and explicit on every plan — never
omitted, per this repository's "missing rather than dropped" convention.

Where the header declares `yes`, `/flow` SHALL generate a mechanical before/after passage
comparison, scoped to the union of every task's own `**Files:**` field across the plan, before the
review panel is dispatched.

The comparison SHALL classify each passage it finds into exactly one of four categories:

1. **Moved** — the passage's text is identical, and only its file or heading changed.
2. **Repointed** — the passage's text is identical modulo exactly the two edits
   `myflow-contract-economy`'s split requirement already permits a moved passage: repointing a
   citation's backticked path and the bold section token beside it, or deleting a stale `above` or
   `below` position word the move made false.
3. **Added** — the passage has no matching passage before the change.
4. **Removed** — the passage has no matching passage after the change.

Generation SHALL NOT block the panel. Where the generation script fails, `/flow` SHALL print one
line stating the failure and dispatch the panel without the comparison file — the same fallback
shape the dispatch-context bundle's own missing-bundle case already uses.

Every review-panel slot's dispatch prompt SHALL name the comparison file's absolute path, when it
exists, alongside the existing dispatch-context bundle pointer.

#### Scenario: The header is required and explicit

- **WHEN** a plan's `tasks.md` header is enriched by writing-plans
- **THEN** it carries an explicit `**Relocation:** yes — <reason>` or `**Relocation:** no` line,
  never an omitted one

#### Scenario: A `yes` plan scopes the comparison to the plan's own files

- **WHEN** a plan declares `**Relocation:** yes` and its tasks' `**Files:**` fields name three files
  between them
- **THEN** the comparison reads exactly those three files, and no file outside that union

#### Scenario: An identical passage under a new heading is moved

- **WHEN** a passage's text is byte-for-byte unchanged but now sits under a different heading in a
  different file
- **THEN** the comparison classifies it as moved

#### Scenario: A repointed citation is not misclassified as added and removed

- **WHEN** a moved passage's only changes are a repointed backticked path and bold section token, or
  the deletion of a stale `above` or `below`
- **THEN** the comparison classifies it as repointed, not as one removed passage plus one added
  passage

#### Scenario: A genuinely new passage is added

- **WHEN** a passage in the after-state has no matching passage in the before-state
- **THEN** the comparison classifies it as added

#### Scenario: A genuinely deleted passage is removed

- **WHEN** a passage in the before-state has no matching passage in the after-state
- **THEN** the comparison classifies it as removed

#### Scenario: A failed generation does not block the panel

- **WHEN** the generation script exits non-zero
- **THEN** `/flow` prints one line stating the failure and dispatches the panel without a comparison
  file

#### Scenario: Every slot's prompt names the comparison file

- **WHEN** a relocation-declaring plan's comparison file exists at dispatch time
- **THEN** every panel slot's dispatch prompt names that file's absolute path, alongside the
  dispatch-context bundle pointer
