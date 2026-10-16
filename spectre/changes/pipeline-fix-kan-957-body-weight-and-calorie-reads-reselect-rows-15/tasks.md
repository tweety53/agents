# pipeline-fix-kan-957-body-weight-and-calorie-reads-reselect-rows-15

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

- [x] 1. Scope the explain-then-ask turn split to flow-eval candidates

**Build:** green
**After:** none
**Files:** `skills/flow-contracts/operator-prompts.md`, `skills/flow-self-review/SKILL.md`, `skills/flow/unfinished-work-gate.md`
**Allowed-collateral:** `spectre/changes/pipeline-fix-kan-957-body-weight-and-calorie-reads-reselect-rows-15/**`
**Tests:** **none** — prose contract; guarded by check-verbatim-moves and check-references.
**Regression:** none
**Commit:** fix(flow-contracts): split explanation and ask across turns for flow-eval candidates only

  - [x] **Step 1:** Rewrite the section, rename it, update its citers.
  - [x] **Step 2:** `scripts/pipeline-fix.sh verify <worktree>`.
