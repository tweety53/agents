# kan-439-check-unfinished-work-sh-false-outstanding-on

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

- [x] 1. Harness cases for same-name canonical resolution and canonical-project findings
**Build:** red
**Squash-with:** Task 2
**Files:** `scripts/test-lib-change-plan.sh`, `scripts/test-check-unfinished-work.sh`
**Tests:** `same-name canonical resolution resolves a treeless satellite`, `same-name resolution never fires without a canonical worktree`, `a failing Part of link is never rescued by same-name resolution`, `signal two reads findings via the canonical worktree`
**Regression:** if this commit is reverted, the four new cases vanish and a future regression of the treeless-satellite false `OUTSTANDING` (and of the permissive satellite findings read) ships unobserved.
**Baseline:** lib harness before=32 after=42; guard harness before=84 after=95
<!-- measured: bash scripts/test-lib-change-plan.sh | grep -c '^ok:' → 32, and bash scripts/test-check-unfinished-work.sh | grep -c '^ok:' → 84 @ branch spectre/kan-439-check-unfinished-work-sh-false-outstanding-on (== main e0a0960) -->
<!-- predicted: both harnesses green after task 2 lands the implementation; counts confirmed by re-running the same commands -->
**Commit:** fix(scripts): resolve a treeless satellite plan through the canonical worktree

  - [x] **Step 1:** extend `scripts/test-lib-change-plan.sh` with three fixtures — a change dir with no `link.md` plus a canonical worktree holding a same-named `tasks.md` (resolves there), a `link.md` with no `## Part of` plus the same canonical worktree (resolves there), and a `## Part of` link whose canonical plan is absent (still unresolvable — no same-name rescue); plus the no-canonical-argument case (unresolvable).
  - [x] **Step 2:** extend `scripts/test-check-unfinished-work.sh` with guard-level cases — treeless satellite + canonical argument tracks the canonical plan's checkboxes (`CLEAR` and `OUTSTANDING`), argument omitted reports `OUTSTANDING: no plan at`, `## Part of` link with a missing canonical plan exits 2, and the stubbed `flow` records receiving `-C <canonical-worktree>` on the findings call when the plan resolved there (`signal two reads findings via the canonical worktree`).
  - [x] **Step 3:** run both harnesses; every new case fails for the missing implementation, every pre-existing case still passes.

- [x] 2. Same-name resolution in change-plan.sh and canonical-project findings read in the guard
**Build:** green
**Files:** `scripts/lib/change-plan.sh`, `scripts/check-unfinished-work.sh`
**Tests:** `same-name canonical resolution resolves a treeless satellite`, `same-name resolution never fires without a canonical worktree`, `a failing Part of link is never rescued by same-name resolution`, `signal two reads findings via the canonical worktree`
**Regression:** if this commit is reverted, `check-unfinished-work.sh` returns to the false `OUTSTANDING: no plan at …` on every treeless cross-repo worktree and its satellite signal two answers a permissive `[]` under the wrong project key; all four cases fail.
**Baseline:** lib harness before=32 after=42; guard harness before=84 after=95
<!-- measured: bash scripts/test-lib-change-plan.sh | grep -c '^ok:' → 32, and bash scripts/test-check-unfinished-work.sh | grep -c '^ok:' → 84 @ branch spectre/kan-439-check-unfinished-work-sh-false-outstanding-on (== main e0a0960) -->
<!-- predicted: counts after this task, both harnesses fully green -->
**Commit:** fix(scripts): resolve a treeless satellite plan through the canonical worktree

  - [x] **Step 1:** `_change_plan_resolve_dir` gains the same-name branch — reached only when the local `tasks.md` is absent and the `link.md` is absent or carries no `## Part of`; with a canonical worktree supplied, resolve `<canonical-worktree>/<spec_root_leaf(canonical)>/changes/<name>` when its `tasks.md` exists; every existing failure shape keeps its current return.
  - [x] **Step 2:** `scripts/check-unfinished-work.sh` — when the resolved plan directory lies under the canonical worktree, run signal two's `flow record findings` with `-C <canonical-worktree>`; the closing `flow record verdict` write keeps `-C <worktree>`; update the header comment's satellite definition for the canonical-repo-only convention.
  - [x] **Step 3:** run both harnesses green, then the lint commands for the touched files — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-contract-budget.sh`.
