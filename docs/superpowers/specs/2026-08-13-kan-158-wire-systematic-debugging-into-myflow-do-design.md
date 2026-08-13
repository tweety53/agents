# KAN-158 — Wire systematic-debugging into myflow-do's task loop

**Date:** 2026-08-13
**Jira:** KAN-158
**Change:** `kan-158-wire-systematic-debugging-into-myflow-do`

## Context

`skills/myflow-do/SKILL.md`'s Superpowers Basic Workflow table names steps 2–6 and 8; steps 1 and 7
are unused slots (1 belongs to `/myflow-start`'s brainstorming, run before `/myflow-do` starts).
None of the named steps is `superpowers:systematic-debugging`. When an implementer subagent hits a
test failure for a reason RED-GREEN-REFACTOR did not plan, or a review-panel finding turns out to be
a confirmed defect rather than a style nit, nothing in the dispatch chain routes that investigation
through the debugging skill — the subagent is left to freelance a fix.

The global rule `superpowers:using-superpowers` says "fix this bug" should trigger
`systematic-debugging`, but every subagent `/myflow-do` dispatches starts fresh and inherits no
context from the parent session (per this repository's own caveman-mode carve-out, which states the
same fact about subagent dispatch generally). That global rule only reaches a dispatched subagent if
the dispatch prompt names it explicitly — the same reason the implementer dispatch already carries an
explicit TDD bullet rather than relying on a subagent to remember the global rule on its own.

## Goals

- Give `superpowers:systematic-debugging` a named slot in the Superpowers Basic Workflow table,
  stating when it applies.
- Name it explicitly on the two dispatch prompts where a subagent might otherwise freelance a fix:
  the section 4 implementer dispatch (unexpected test failure) and the section 5 fix-subagent
  dispatch (a confirmed-defect finding).

## Non-goals

- No change to the commit-per-task model, the state machine, or `/myflow-finish` — this only shapes
  how an implementer or fixer investigates before writing a fix.
- No change to the review panel's roster, escalation ladder, or handoff bar.
- No change to `/myflow-fast` or `/myflow-start` — neither dispatches implementer or fix subagents
  directly; both chain into `/myflow-do`'s section 4/5, which already carry the fix.

## Decisions

### D1 — Fill the step-7 slot in the workflow table

**ID:** step-7-systematic-debugging
**Status:** active
**Chosen:** add a row naming `superpowers:systematic-debugging`, scoped to two triggers: an
unexpected test failure during implementation, and a review-panel finding confirmed as a real
defect.
**Considered:**
- *Renumber the table to close the 1/7 gap* — cosmetic, touches every other row's number for no
  behavioural gain, and the gap already carries a stated reason (step 1 belongs to
  `/myflow-start`). Rejected.
- *Add it as prose outside the table* — the table is the one place this skill enumerates which
  Superpowers step governs which moment; a new governing step belongs in it, not beside it.

### D2 — Name it explicitly on both dispatch prompts, not the plan template

**ID:** dispatch-not-plan-template
**Status:** active
**Chosen:** add a `REQUIRED SUB-SKILL` bullet next to the existing TDD bullet in section 4's
implementer dispatch, and an equivalent instruction in section 5's fix-subagent dispatch — both
conditioned on the trigger firing, not unconditionally invoked on every task.
**Considered:**
- *Add it to `tasks.md`'s per-task template instead* — would apply it to every task regardless of
  whether anything unexpected happened, which is the wrong trigger; the failure is discovered during
  execution, not knowable at planning time.
- *One shared bullet text reused verbatim in both places* — the two dispatches address different
  audiences (a task implementer mid-RED-GREEN-REFACTOR vs. a fix subagent reading a findings list),
  so the wording differs slightly even though the invoked skill is the same.

## Open questions

None — the design is small enough that both decisions were settled in one round.

## Risks and trade-offs

- **`skills/myflow-do/SKILL.md` is under a contract budget** (`scripts/check-contract-budget.sh`) →
  both additions are short (one table row, one bullet in each of two dispatch prompts); run the
  guard before publishing and trim wording if it goes over.
- **An implementer could over-invoke the skill on an ordinary expected RED** → the table row and both
  dispatch bullets state the trigger precisely (unexpected failure / confirmed defect), not "on any
  red test," to keep the ordinary TDD loop unaffected.

## Testing

This is a prose-only change to a skill file with no executable surface of its own. Verification is:
- `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` stay green on the edited
  file.
- `scripts/check-contract-budget.sh` stays green on `skills/myflow-do/SKILL.md`.
- Manual read-through confirming the new row and both bullets read consistently with the rest of the
  file's voice and citation style.
