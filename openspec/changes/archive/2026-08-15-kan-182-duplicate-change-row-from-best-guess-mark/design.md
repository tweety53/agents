# Design — one change, two rows

Adapted from `docs/superpowers/specs/2026-08-15-kan-182-duplicate-change-row-from-best-guess-mark-design.md`,
which carries the full walk-through of the mechanism.

## The mechanism

`/myflow-fast`'s **State gate** runs before anything resolves the change name — deciding which
branch the run takes is its whole job — so it refers to the change as `<name-or-best-guess>` in both
the read and the marks.

The read is harmless: `myflow state get` on an unknown name exits 1 and writes nothing, which is
exactly the signal the gate wants. The mark is not: `ApplyBeginStageMark`
(`stats/internal/api/stages.go`) catches `store.ErrChangeNotFound`, bootstraps a change row —
`STARTED`, `updated_by = stages.SyntheticChangeUpdatedBy` — so the mark has something to attach to,
and retries once. The guessed name is now a row. Section A then resolves the real `<key>-<slug>`
name, and the next mark bootstraps a second row.

## The fix, in four parts

1. **`skills/myflow-fast/SKILL.md`, State gate.** The read keeps `<name-or-best-guess>`; the
   `do.state-gate` `begin`/`end` pair takes a resolved `<name>`. On a creating run the pair fires
   back to back the moment section A fixes the name, before `start.resolve-change`'s own marks. At
   `IN_PROGRESS` the name is resolved before the gate in every path, so those marks do not move.
2. **`scripts/check-stage-mark-calls.sh`.** A fourth check beside its three existing ones: a
   `stage begin`'s positional change argument may not be a bracketed placeholder whose text names a
   guess.
3. **`skills/myflow-contracts/pipeline.md`, Stage marks.** One paragraph stating the rule, beside
   the session-token and harness rules the same guard enforces.
4. **The stray row.** One transaction against `myflow` on `localhost:5433` removing
   `gymie-7c1f238a` / `kan-175`.

## Decisions

### Where the fix lands

**ID:** fix-in-skill-ordering
**Status:** active
**Chosen:** Defer `/myflow-fast`'s state-gate mark until the change name is resolved — the cause is
one placeholder in one section of one skill, and `/myflow-start` section A already states the same
rule for its own creating run.
**Considered:**
- *The daemon refuses the bootstrap for pre-resolution marks* — ruled out: the bootstrap is
  kan-174's recorded decision, whose stated purpose is that a synthetic row makes "a defect worth
  seeing in the data" visible. It worked. Removing the report leaves the cause in place, here and
  for every other mistyped change name.
- *The dashboard filters synthetic-only rows out of the open-change list* — ruled out for the same
  reason, one layer further from the cause, and it would hide genuine strays that an operator needs
  to see.
- *Both the skill fix and a daemon guard* — ruled out as unrequested complexity: one command can
  emit an unresolved name, and after this change none does.

### Whether a mark may name something that is not a change

**ID:** marks-name-resolved-changes
**Status:** active
**Chosen:** State the rule once under **Stage marks** (`skills/myflow-contracts/pipeline.md`) and
enforce it in `check-stage-mark-calls.sh`, which already owns the other three `stage begin` rules.
**Considered:** *Leaving it as a fixed placeholder in `/myflow-fast` alone* — ruled out: nothing
would stop the next author moving a mark above a resolution step, which is exactly how this arose.

### How the guard recognises a guess

**ID:** guard-matches-guess-placeholders
**Status:** active
**Chosen:** Reject a bracketed placeholder whose own text contains `guess`, case-insensitively. The
guard cannot know whether `<name>` is resolved at a given call site; it can know that a placeholder
saying "guess" is not.
**Considered:**
- *Requiring the change argument to be exactly `<name>`* — ruled out as too strict for a guard over
  prose skill files, and it would fail call sites that legitimately name the change differently.
- *No guard, prose only* — ruled out: the change would then be unverifiable text, and the placeholder
  returns the first time someone tidies a mark upward.

### Clearing the existing stray row

**ID:** one-off-delete-not-a-command
**Status:** active
**Chosen:** A one-off transaction against the shared development database, ordered
`stage_runs` → `change_repos` → `changes` because neither foreign key cascades.
**Considered:**
- *Adding `myflow state delete <name>`* — ruled out: one row needs removing and the cause is fixed
  in the same change; a delete path into the store is surface area nobody has needed twice.
- *Leaving the row* — ruled out: the reported defect is the duplicate row on screen, so a fix that
  leaves it there has not been done.

## Open questions

*(none — the two questions this change raised, where the fix lands and what happens to the existing
row, were both answered before the design was written and are recorded above.)*
