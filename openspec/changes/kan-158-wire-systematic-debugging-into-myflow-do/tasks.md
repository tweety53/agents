> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Give `superpowers:systematic-debugging` a named slot in `/myflow-do`'s workflow, and name
it explicitly on the two dispatch prompts where a subagent could otherwise freelance a fix instead
of investigating: the implementer dispatch (section 4) and the fix-subagent dispatch (section 5).

**Architecture:** One prose-only edit to one file, `skills/myflow-do/SKILL.md` — a new row in the
Superpowers Basic Workflow table, plus one bullet added to each of two existing dispatch prompt
templates. No code, no guard script, no state file shape change.

**Tech Stack:** Markdown skill file. Verification is the repository's own prose guards
(`scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
`scripts/check-contract-budget.sh`) plus a read-through against the delta spec's scenarios.

## Global Constraints

- **No change to the commit-per-task model, the state machine, or `/myflow-finish`.**
- **No change to the review panel's roster, escalation ladder, or handoff bar.**
- **`skills/myflow-fast/SKILL.md` and `skills/myflow-start/SKILL.md` are not edited.** Neither
  dispatches implementer or fix subagents directly; both chain into `/myflow-do`'s sections 4/5,
  which already carry the fix.
- **The new table row and both dispatch bullets name their trigger precisely** — an unexpected test
  failure, or a review finding confirmed as a real defect — never "on any red test" or "on any
  finding," so the ordinary TDD loop and ordinary fix rounds are unaffected.

## Baseline

<!-- verified: wc -c skills/myflow-do/SKILL.md; budget row read from scripts/check-contract-budget.sh -->

| Measure | Now | After this change |
|---------|-----|-------------------|
| `skills/myflow-do/SKILL.md` | 53425 bytes, budget row 58623 (5198 headroom) | still under the row |

The addition is roughly one table row and two short bullets — well under the headroom above, so
**no task re-anchors the budget table**.

---

### 1 `skills/myflow-do/SKILL.md` — name systematic-debugging in the table and both dispatch prompts

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: nothing later tasks read.

- [x] **Step 1: Fill the step-7 slot in the Superpowers Basic Workflow table**

Add a row between the existing step-6 and step-8 rows:

```markdown unverified:authored for this task; not yet applied
| **7** | **superpowers:systematic-debugging** | An unexpected test failure during implementation, or a review-panel finding confirmed as a real defect |
```

- [x] **Step 2: Name it on the section 4 implementer dispatch**

In section 4's implementer dispatch block, next to the existing
`> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development` bullet, add:

```markdown unverified:authored for this task; not yet applied
> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, invoke
> superpowers:systematic-debugging before writing a fix. An expected RED step needs no invocation.
```

- [x] **Step 3: Name it on the section 5 fix-subagent dispatch**

In section 5, at "Give the surviving findings — every dispatched finding from the union above — to
**one** fix subagent as the combined list," add one sentence naming the same skill for a finding
confirmed as a real defect, before the fix subagent writes a fix for it — distinct from a finding
that turns out to be a style or principles nit, which needs no investigation step.

**Tests:** No automated test — skill prose, the same verification shape section 5 of
kan-153-kan-108-follow-up's own plan used for its prose-only task. Verification is running the
repository's own reference, markdown-integrity and contract-budget guards, all exiting clean
against the edited file, plus a read-through against this change's delta spec covering the
table-row scenario, the unexpected-failure scenario, the expected-RED scenario, and the
fix-subagent scenario.
**Regression:** Reverting this task restores the current state, in which no dispatch prompt names
`superpowers:systematic-debugging`, so a subagent facing an unexpected failure or a confirmed-defect
finding has nothing routing it through a debugging discipline before it writes a fix.
**Baseline:** before=0 after=0 automated cases; `skills/myflow-do/SKILL.md` stays under its
58623-byte budget row. <!-- verified: budget row read from scripts/check-contract-budget.sh -->
**Commit:** `feat(kan-158-wire-systematic-debugging-into-myflow-do): route unexpected failures and confirmed-defect findings through systematic-debugging`
