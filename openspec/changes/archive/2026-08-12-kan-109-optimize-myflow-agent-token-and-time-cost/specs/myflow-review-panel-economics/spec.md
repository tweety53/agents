## ADDED Requirements

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
