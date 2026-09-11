# kan-494-flow-9-of-72-review-findings-were-plan-record

> **Execution:** `/flow-fast` implements this plan inline. Mark a task's own checkbox once its
> commit lands.
> **Relocation:** no

- [x] 1. Extend **PLAN FIELDS** to name `**Tests:**` ownership

**Build:** green

**Files:**
- Modify: `skills/flow/review-panel.md`

**Tests:** **none** — prose-only contract change; verified by the repository's markdown guards and
by inspection that the paragraph names all three fields
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): plan-fields obligation owns Tests: field updates`

  - [ ] **Step 1: replace the PLAN FIELDS paragraph**

  In `skills/flow/review-panel.md`, replace the fix subagent's **PLAN FIELDS** paragraph (the
  blockquote beginning "**PLAN FIELDS:** when your fix changes what a task's `**Baseline:**`
  counts…") with:

  ````markdown verified:authored in-tree for this change
  > **PLAN FIELDS:** your fix owns its plan record, as part of the fix itself: when it adds a test
  > case or changes what tests a task names, update that task's `**Tests:**` field; when it changes
  > what a task's `**Baseline:**` counts or `**Files:**` paths declare — a test case added, a file
  > created — update those fields too. All of it lands in the worktree's
  > `<project>/spectre/changes/<name>/tasks.md` in this same pass — never left for a reviewer to
  > catch next round (kan-454, KAN-459). Edit them; do not stage or commit them — the plan record is
  > a planning path and is committed later by the pipeline, never in a fixup.
  ````

  - [ ] **Step 2: verify**

  Run `scripts/check-vocabulary.sh` and `scripts/check-references.sh` — both must exit 0. Confirm
  the paragraph names `**Tests:**`, `**Baseline:**` and `**Files:**` and keeps the
  never-stage-or-commit rule.

- [x] 2. Bind the fix round's close to a mechanical plan-record check

**Build:** green

**Files:**
- Modify: `skills/flow/review-panel.md`

**Tests:** **none** — prose-only contract change; verified by the repository's markdown guards and
by inspection that the close names the guard, its exit rule, and the KAN-511 split
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): fix round re-runs check-task-commit-fields on the folded task sha`

  - [ ] **Step 1: extend the walk's PLAN FIELDS sentence with the guard re-run**

  In `skills/flow/review-panel.md`'s fix-round section, directly after the sentence ending "…does
  not close the round; it goes to the handback." (the one beginning "**The same walk holds the fix
  subagent to its PLAN FIELDS obligation**"), insert:

  ````markdown verified:authored in-tree for this change
  Beside the reproducer re-runs below, the round close re-runs the task-field guard mechanically:
  `check-task-commit-fields.sh <worktree> <task-id> <task-sha> <task-base> <canonical-worktree>
  <name>` for every task a fixup folded into — `<task-id>` from that commit's `Task-Id:` trailer,
  `<task-sha>` the folded commit as it now stands, the remaining arguments resolved the way
  `skills/flow/implement.md`'s task-close step resolves them. A non-zero exit does not close the
  round; it goes to the handback. This catches an undeclared file the fixup added and a declared
  test it removed or renamed, read post-autosquash; a test added to the commit without a
  `**Tests:**` declaration and a stale `**Baseline:**` count remain the walk's judgment until the
  companion checker (KAN-511) lands.
  ````

  - [ ] **Step 2: verify**

  Run `scripts/check-vocabulary.sh` and `scripts/check-references.sh` — both must exit 0. Confirm
  the inserted paragraph names the guard invocation, the non-zero-exit handback rule, and the
  KAN-511 split.
