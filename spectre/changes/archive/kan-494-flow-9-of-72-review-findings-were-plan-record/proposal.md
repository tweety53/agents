# kan-494-flow-9-of-72-review-findings-were-plan-record

## Why

Nine of KAN-459's 72 review findings (F23, F27, F38, F47, F55, F62, F67, F68, F75) were plan-record
bookkeeping, not defects: every fix round that added tests left that task's `**Tests:**` field and
`**Baseline:**` counts stale, and `tasks.md` sketch text drifted from the actual commits. Each drift
consumed a review-panel round — a reviewer caught it on the next pass. Two causes:

- The fix subagent's **PLAN FIELDS** paragraph (`skills/flow/review-panel.md`) names only
  `**Baseline:**` counts and `**Files:**` paths as fields to update — never `**Tests:**` — even
  though "a test case added" is listed among its triggers. The one field a test case updates is
  absent from the list.
- The parent-side enforcement is a judgment walk at round close, with no mechanical check behind
  it; the paragraph was already ignored once before (kan-454), and the walk missed the drift again.

## What changes

In `skills/flow/review-panel.md`'s fix-round contract:

- **PLAN FIELDS** names `**Tests:**` explicitly: panel-fix owns updating `**Tests:**`,
  `**Baseline:**` and `**Files:**` for tests it adds, in the same pass — part of the fix itself,
  never left for a reviewer to catch next round. The never-commit rule is unchanged:
  `spectre/changes/` is a planning path that stages at integrate, never in a fixup.
- The round close binds mechanically where a guard exists: beside the reproducer re-runs, the
  parent re-runs `check-task-commit-fields.sh` against the folded task sha for the task the fixup
  folded into, and a non-zero exit does not close the round. That catches an undeclared file a
  fixup added and a declared test it removed or renamed, post-autosquash.
- The added-but-undeclared-test direction and the `**Baseline:**` chain stay with the companion
  checker change (KAN-511, "baseline-chain and Tests: field accuracy") — this change states that
  split rather than building the checker.
