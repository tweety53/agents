# self-review-kan-918-flow-improvement-port-the-git-snapshot-and

## Why

kan-918's self-review pass found six pipeline defects, none `big`:

- **F1 (flow-fix).** The PLAN FIELDS brief tells a panel fixer to widen a task's `**Tests:**`,
  `**Files:**` and `**Baseline:**` on any route, and the fix round's walk refuses a round without
  it; but the normal route lands a fix as an on-top `Flow-Key:` commit outside every task, so
  `check-task-records`' rule 2 refuses the widened fields. kan-918's task 3 hit it at verify and
  the fields were reverted.
- **F2 (flow-fix).** `verify-and-handoff.md` runs the live check "after the lint and test list",
  while the same file starts that list as two detached, polled groups. kan-918 ran the live check
  while the groups ran; it wrote `live-verification.md` into the tree `run-guard-tests.sh` watches,
  which failed the watched-tree check and cost a re-run.
- **F3 (flow-fix).** `check-plan-shape` passes a `**Tests:**` name the task's commit cannot add:
  kan-918's task 3 declared `TestFoldFixup`, at base the prefix of five tests in
  `stats/internal/guard/fold_fixup_test.go`, a file outside the task's `**Files:**`, and
  `check-task-commit-fields` refused the commit.
- **F4 (flow-fix).** Round 1's `primary+principles` panel re-run ended without writing either
  report file.
- **F5 (flow-stats-app).** Recording a round's findings is one `flow record finding` per finding;
  agents hand-build shell loops over them, and kan-918's 21-row loop lost a retry to zsh word
  splitting of an optional `-task`.
- **F6 (flow-speed).** `scripts/test-integrate-change.sh` alone sets the guard suite's wall: its
  cases run serially on one global `$WORK`.

## What changes

- PLAN FIELDS and the fix round's walk bind a fix folded into a task's commit with
  `fold-fixup.sh` only; an on-top `Flow-Key:` fix edits no task's fields.
- The live check starts once both groups of the lint and test list have exited.
- `check-plan-shape` gains F15: a `**Tests:**` name a base-tree test file outside the task's files
  already declares.
- A panel slot's Agent call carries each role's REPORT FILE path beside the pointer to its brief.
- `flow record finding -from <file>` records one finding per JSON line, and `review-panel.md`
  records a round's findings through it.
- `test-integrate-change.sh` runs each case in its own background subshell and sums the tallies.
