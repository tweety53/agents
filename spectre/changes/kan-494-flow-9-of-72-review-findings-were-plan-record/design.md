## Context

`skills/flow/review-panel.md` is the only home of both halves that need editing: the fix subagent's
**PLAN FIELDS** dispatch paragraph and the parent's fix-diff walk that holds the fix to it. Nothing
restates them — `skills/flow/implement.md`'s inline mode binds the parent to every implementer/fixer
dispatch paragraph "in the same words", so one edit covers both execution modes. `/flow-fast`
carries its own inline fix loop (`skills/flow-fast/review.md`) and deliberately restates nothing
from review-panel.md's fix-round machinery.

The mechanical guard that already exists, `check-task-commit-fields.sh`, checks `**Files:**`,
`**Tests:**` and `**Commit:**` against the real commit — but `check_tests` is one-directional:
every declared test must appear in the commit's diff; a test added to the commit without a matching
declaration is not flagged. `**Baseline:**` and `**Regression:**` remain plan declarations only
(the runtime check behind them was removed by KAN-442).

## Decisions

### plan-fields-owns-tests

**ID:** plan-fields-owns-tests
**Status:** active
**Chosen:** **PLAN FIELDS** names `**Tests:**` explicitly and states the same-pass ownership — the
field list is what a fixer reads, and a trigger example without its field in the list has now been
ignored twice (kan-454, KAN-459)
**Considered:** leaving `**Tests:**` implicit under "a test case added" — the exact wording that
failed twice; waiting for the KAN-511 checker before touching the text — the ownership wording is
fixable now and the checker is its own change

### same-pass-not-in-fixup

**ID:** same-pass-not-in-fixup
**Status:** active
**Chosen:** the Jira's "as part of the fix commit itself" is read as same-pass ownership — the plan
record is current by the time the round closes — never as committing `tasks.md` inside the fixup
**Considered:** committing `tasks.md` in the fixup commit — violates the canonical git boundary
(`skills/flow-contracts/git-boundaries.md`): `spectre/changes/` is never staged in a task or fixup
commit; the integrate phase commits it

### walk-plus-guard-rerun

**ID:** walk-plus-guard-rerun
**Status:** active
**Chosen:** the round close adds a `check-task-commit-fields.sh` re-run against the folded task
sha, beside the reproducer re-runs — the mechanical half available today (an undeclared file the
fixup added; a declared test it removed or renamed), applied post-autosquash where a judgment
read of the pre-rebase tree would not see the surviving content
**Considered:** the judgment walk alone — it is what missed KAN-459's drift; implementing the
added-undeclared-test direction here — that is KAN-511's declared scope, so this change states the
split instead of duplicating the checker

### flow-fast-out-of-scope

**ID:** flow-fast-out-of-scope
**Status:** active
**Chosen:** `/flow-fast`'s inline fix loop is not extended — the observed waste was `/flow`'s
dispatched panel-fix, and `/flow-fast`'s review file restates nothing from review-panel.md's
fix-round machinery by design
**Considered:** mirroring the ownership sentence into `skills/flow-fast/review.md` — no observed
failure on that path; a separate ask if one appears

## Open questions
