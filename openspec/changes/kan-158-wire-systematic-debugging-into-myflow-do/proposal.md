## Why

`skills/myflow-do/SKILL.md`'s Superpowers Basic Workflow table names steps 2–6 and 8, but none of
them is `superpowers:systematic-debugging`. When an implementer subagent hits a test failure for a
reason RED-GREEN-REFACTOR did not plan, or a review-panel finding turns out to be a confirmed
defect, nothing in the dispatch chain routes that investigation through the debugging skill — a
dispatched subagent starts fresh and never sees the global `using-superpowers` rule that would
otherwise trigger it.

## What Changes

- Fill the unused step-7 slot in the Superpowers Basic Workflow table
  (`skills/myflow-do/SKILL.md`) with `superpowers:systematic-debugging`, scoped to an unexpected
  test failure and a confirmed-defect review finding.
- Add a `REQUIRED SUB-SKILL` bullet to the section 4 implementer dispatch prompt, alongside the
  existing TDD bullet, naming the trigger for invoking it.
- Add an equivalent instruction to the section 5 fix-subagent dispatch, scoped to a finding
  confirmed as a real defect.

No new commit type, no state or state-machine change, no change to the review panel roster or
`/myflow-finish`.

## Capabilities

### New Capabilities

- `myflow-debugging-dispatch`: implementer and fix-subagent dispatches route an unexpected test
  failure or a confirmed-defect finding through `superpowers:systematic-debugging` before writing a
  fix.

### Modified Capabilities

(none — no existing capability's requirements change; this adds a new, previously unstated one)

## Impact

- `skills/myflow-do/SKILL.md` — the Superpowers Basic Workflow table (step 7) and the section 4 and
  section 5 dispatch prompt templates.
- No code, no guard script, no state file shape change.
