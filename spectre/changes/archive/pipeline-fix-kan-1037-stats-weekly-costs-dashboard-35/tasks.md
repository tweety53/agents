# pipeline-fix-kan-1037-stats-weekly-costs-dashboard-35

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** the Finished handoff's prompt-tuning line reads `**Prompt tuning:** <index.html path> | NONE`
and is present in every `/flow` and `/flow-fast` run's final message.

**Architecture:** a docs edit to `skills/flow/cleanup.md` and `skills/flow-fast/SKILL.md`, plus a
bash regression test `scripts/test-prompt-tuning-line.sh`.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is this change's `task1.diff`, applied from the repository root with
  `git apply`; it was applied to this worktree at 980bf698 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- Cut, never paraphrase: no sentence beyond the diff's is reworded.
- Read-only against the dev store: never stop flowd or flow-postgres.

## Review Focus

- The `NONE` value carries no reason; `Learning:` keeps its own (`none-without-reason`).
- The always-present rule is stated once, in cleanup.md (`always-in-final-message`).
- No `Prompt-tuning:` spelling remains anywhere in skills, rules, scripts or stats.

- [x] 1. Rename the Prompt tuning line and make it always present
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Copy this change's `task1.diff` to
    `/tmp/pf35-task1.diff`, apply its test hunk alone
    (`git apply --include='scripts/test-*' /tmp/pf35-task1.diff`) and run
    `bash scripts/test-prompt-tuning-line.sh` (the file Step 2 completes): it exits 1 printing four `FAIL:` lines and the four
    `Prompt-tuning:` occurrences still in the skills.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='scripts/test-*' /tmp/pf35-task1.diff`):

    The diff is this change's `task1.diff`, verified: applied with git apply to this worktree at
    980bf698; scripts/test-prompt-tuning-line.sh prints all cases passed with it; check-verbatim-moves
    exits 0 with this change's verbatim-moves.txt; check-references exits 0.

  - [x] **Step 3: Verify.** `bash scripts/test-prompt-tuning-line.sh` exits 0 printing
    `all cases passed`; `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh`
    exit 0.
  - [x] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `scripts/test-prompt-tuning-line.sh`, `skills/flow-fast/SKILL.md`, `skills/flow/cleanup.md`, `skills/flow/integrate.md`, `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/handoff-blocks.md`, `spectre/changes/pipeline-fix-kan-1037-stats-weekly-costs-dashboard-35/verbatim-moves.txt`
**Tests:** `scripts/test-prompt-tuning-line.sh`
**Regression:** scripts/test-prompt-tuning-line.sh — reverting the skills hunks prints `FAIL: skills/flow/cleanup.md: Finished block lacks the '**Prompt tuning:** <path> | NONE' line`.
**Baseline:** before=0 after=5
<!-- measured: grep -c "fail \"" scripts/test-prompt-tuning-line.sh @ HEAD -->
**Commit:** fix(flow): print the Prompt tuning line in every run's final message
**After:** none

**Decision:** prompt-tuning-line-renamed
**Decision:** none-without-reason
**Decision:** always-in-final-message
**Decision:** test-script-guard
**Decision:** verbatim-moves-acknowledged
