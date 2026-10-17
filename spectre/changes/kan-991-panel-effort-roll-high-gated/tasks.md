# kan-991-panel-effort-roll-high-gated

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

- [x] 1. Planner and implement prose: panel effort roll, gated high, auto-decided display formatting
  - [x] **Step 1: In `skills/flow/brainstorm-planner.md` **Model and effort**, replace the sentence running from "**When `effort_roll < 90`, the implementer, the fixer and every implementer group take `low`" to "its `reason` naming the seam." with: "**When `effort_roll < 90`, the implementer, the fixer and every implementer group take `low` — `medium` where the pair carries a hard seam.** **Every pass-1 panel dispatch and every gated per-task reviewer bundle takes `low` when `effort_roll < 60` and `medium` otherwise, whatever the band** — 60/40 per change. **`high` reaches one of them only two ways:** the operator's explicit **Yes** at the plan gate (**Plan review gate** below), or a fix run's re-Decide when the fix instructions report that an earlier, lower-effort attempt failed to solve the same issue, the pair's `reason` naming that attempt." Then change the following "When `effort_roll ≥ 90`, every other `effort` is the planner's own choice, one of `low`/`medium`/`high` — `low`/`medium` for an implementer, fixer or group —" to "When `effort_roll ≥ 90`, the implementer, fixer and group efforts are the planner's own choice, one of `low`/`medium` —" leaving the rest of that paragraph unchanged.**
  - [x] **Step 2: In the same section's bullet "**every gated per-task reviewer bundle**", replace "at its group's effort, or `high` where its group carries a hard seam, the group's `reason` naming the seam" with "at the panel roll's effort, `high` only as the paragraph below allows".**
  - [x] **Step 3: In **Plan review gate**, after the paragraph ending "with exactly two options: **Yes** *(recommended)* and **No (plan needs updates)**." add before "On **No**," a new paragraph: "**A dispatch the planner judges worth `high`** — a hard seam the roll's effort may not carry — adds one question per such dispatch to the same AskUserQuestion call: **Raise `<dispatch>` to `high` — `<seam>`?**, **No** *(default, recommended)* / **Yes**. Only an explicit **Yes** re-records that pair at `high` through the `flow.decide` record sequence before implementation; anything else leaves the roll's effort. Under the `## decisions: recommended` mode the question is not asked and **No** is taken." Keep the gate's own two options unchanged.**
  - [x] **Step 4: In section **B**'s checklist, after the "A ticket that names a practice or a brief resolves it through the lessons home" bullet, add: "- **Display-only formatting is decided, never asked.** A choice that changes only how a value is shown — numeric precision, an empty-cell placeholder, and the like — and that the request, ticket or mockup leaves open is chosen by you and recorded under `## Decisions`; no question is asked for it. A mockup or ticket that states it still governs."**
  - [x] **Step 5: In `skills/flow/implement.md`, replace "on `opus` at that group's effort from the decision's `groups` entry, or `high` where that group carries a hard seam — a first-pass review" with "on `opus` at the panel roll's effort, `high` only as **Model and effort** allows — a first-pass review".**
  - [x] **Step 6: Verify: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-markdown-integrity.py && scripts/check-plan-shape.sh spectre/changes/kan-991-panel-effort-roll-high-gated/tasks.md`.**
**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/implement.md`
**Tests:** **none**
**Regression:** none — the task declares no tests
**Commit:** feat(flow): roll panel effort low/medium, gate high, auto-decide display formatting
**After:** none

**Decision:** panel-effort-roll-60-40
**Decision:** panel-high-operator-gated
**Decision:** display-formatting-auto-decided
