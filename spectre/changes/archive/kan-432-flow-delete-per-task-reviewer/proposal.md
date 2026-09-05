# kan-432-flow-delete-per-task-reviewer

**Jira:** KAN-432

## Why

Nineteen per-task reviews across seven changes produced one `fix` outcome (kan-402 task 1): a
fence-regex fidelity gap the reviewer itself rated harmless, whose fix landed after task 2 had
already been implemented and reviewed against the unfixed task 1 — the "caught before the next
task builds on it" rationale did not hold in its one hit
(`docs/superpowers/research/flow-speedup.md` §8, "Per-task review deleted").
<!-- measured: 19 dispatch records with -role reviewer and -task across 7 changes, 1 with -outcome fix, per docs/superpowers/research/flow-speedup.md section 8 @ 2a5c000 -->

Its mandate is already carried elsewhere: since KAN-436 the whole-branch `primary` slot is
plan/task alignment for every commit, `code-review-low` and Bugbot are the code-quality half, and
`check-task-commit-fields.sh` checks each task's Files/Tests/Commit mechanically. What the
reviewer costs is one launch, one record pair and one tick decision per bundle boundary (~15 s)
plus the 55–90 s trailing solo reviewer: ~1–2 min and ~$0.2–0.7 per multi-task change.
<!-- measured: boundary and trailing-reviewer durations from the store's dispatch records, per docs/superpowers/research/flow-speedup.md section 7 @ 2a5c000 -->

## What changes

- `skills/flow/implement.md` **4. Execute (SDD + TDD)** dispatches no reviewer per bundle. `flow
  tasks tick` runs in the same Bash call as `check-task-commit-fields.sh`, for every task whose
  commit the guard passed. The pending-fix fold, the re-reviewer, the **Per-task review**
  paragraph and **The last bundle's reviewers run alone** are deleted; the panel's slots dispatch
  once the last bundle's guard passes.
- Every sentence elsewhere that asserts a per-task reviewer exists is corrected:
  `skills/README.md`'s step table, `skills/flow/brainstorm-planner.md`'s `tasks.md` header
  template, `skills/flow/review-panel.md`'s docs-only rationale and fix-round scoping sentence,
  `skills/flow-contracts/finish-contract-run1.md`'s no-verification-gate paragraph, and the
  comments in `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
  and `scripts/check-panel-docs-only.sh`.
- Nothing executable changes: `stats/cmd/flow/tick.go`, `scripts/check-unfinished-work.sh`,
  every guard's table and exit contract, and the store's dispatch schema are untouched. The
  deterrent value of a reviewer following every implementer stays unmeasured and is recorded as
  an open question in `design.md`.
