## ADDED Requirements

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

## MODIFIED Requirements

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
