# pipeline-fix-kan-1017-templates-step-3-admin-templates-queue-create-32

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

- [x] 1. Pipeline: a run's size is never a stop
  - [ ] **Step 1: In `skills/flow-contracts/pipeline.md`, section **Fewest operator actions**, add the paragraph opening **A run's size is never a stop.** with the two sentences recorded in `.superpowers/sdd/pipeline-fix-kan-1017-templates-step-3-admin-templates-queue-create-32/verbatim-moves.txt`: tokens, wall-clock time, context length, rounds or tasks ahead, a checkpoint, and open items each carrying a recommended option are never reasons to end a turn mid-run; a fix reaching a shared source is made there in the same run, never offered as "a separate change" (kan-1017).**
  - [ ] **Step 2: Verify: `scripts/check-installed-citations.sh` exits 0.**
**Build:** green
**Files:** `skills/flow-contracts/pipeline.md`
**Tests:** none — prose rule, covered by the citations guard run.
**Regression:** reverting the commit removes the rule; a run may again stop mid-run citing its size.
**Commit:** fix(flow-contracts): a run's size is never a stop
**After:** none

**Decision:** run-size-never-a-stop

- [x] 2. Cleanup: mark the learning-note path as an illustration
  - [ ] **Step 1: In `skills/flow/cleanup.md`, append ` <!-- citations-guard:allow -->` to the line citing the learning-note placeholder path in the `Learning:` paragraph.**
  - [ ] **Step 2: Verify: `scripts/check-installed-citations.sh` exits 0.**
  <!-- measured: `scripts/check-installed-citations.sh` @ origin/main exit 1, "INSTALLED-CITATIONS-INVALID … 1 violation(s)"; with the allow marker applied exit 0 -->
**Build:** green
**Files:** `skills/flow/cleanup.md`
**Tests:** none — covered by `scripts/check-installed-citations.sh`.
**Regression:** reverting the commit makes `scripts/check-installed-citations.sh` fail on `skills/flow/cleanup.md:67`.
**Commit:** fix(flow): mark the learning-note path in cleanup as an illustration
**After:** none

**Decision:** learning-path-allow-marker
