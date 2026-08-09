## Why

`/myflow-do` implements every task uncommitted, reviews it via a checkpoint/diff snapshot
(`checkpoint`, `uncommitted-review-package`), and leaves everything staged for the human gate.
`/myflow-finish` run 1 makes the branch's only two commits. Nothing mechanically checks a task's
claims — which files it touched, whether declared tests exist, what regresses on revert, the real
baseline test count — against anything real; today those only ever exist in an agent's self-report.
KAN-71 surfaced concrete cases where that self-report was wrong and nothing caught it before review.

Real per-task commits give guards something to check *against*.

## What Changes

- `/myflow-do` implementers commit each task (`task(<n>): <subject>` + `Task-Id:` trailer) right
  after finishing it, before review. The `NO-COMMITS` dispatch clause is replaced by
  `COMMIT-PER-TASK`.
- Review reads a real commit diff (`git diff <task-base>..<task-sha>`) instead of a checkpoint
  snapshot. The `checkpoint` and `uncommitted-review-package` scripts are retired.
- Fix-round changes commit as `git commit --fixup=<task-sha>` and are autosquashed into the task
  commit immediately, before re-review.
- New per-task tags in `tasks.md` — `Files:`, `Tests:`, `Regression:`, `Baseline:`, `Squash-with:`,
  `Commit:` — written during `/myflow-start`'s writing-plans stage and checked by a new **runtime**
  guard against the real commit, right after `/myflow-do` commits each task.
- `Squash-with:` replaces `Build: red`'s inline `— merges with Task <N>` text. **BREAKING** for any
  in-flight plan using the old inline syntax — `myflow-build-green`'s guard now reads the partner
  from `Squash-with:` instead.
- `/myflow-finish` run 1 gains one step before its existing two-commit sequence:
  `git reset --soft <recorded-merge-base>`, collapsing every task/fixup commit back into the
  working tree. Everything after that is unchanged — the branch still ends with exactly two
  commits.

## Capabilities

### New Capabilities

- `myflow-task-commits`: per-task commit format during `/myflow-do`, the fixup-and-autosquash rule
  for fix rounds and `Build: red` partners, and review reading a real commit diff instead of a
  checkpoint snapshot.
- `myflow-task-commit-fields`: the `Files:`/`Tests:`/`Regression:`/`Baseline:`/`Squash-with:`/
  `Commit:` tags, written at plan time and checked by a runtime guard against the real commit.

### Modified Capabilities

- `myflow-uncommitted-review-package`: the checkpoint/diff-snapshot mechanism it specifies is
  retired in full, superseded by `myflow-task-commits`.
- `myflow-build-green`: the `red` tag's inline `— merges with Task <N>` text is replaced by the
  `Squash-with:` field from `myflow-task-commit-fields`; the guard's partner-resolution reads it
  from there.
- `myflow-finish-cleanup`: run 1 gains a squash-to-merge-base step before its existing two-commit
  sequence.

## Impact

- `skills/myflow-do/SKILL.md` — commit model, dispatch clause, review-diff source, git boundaries.
- `skills/myflow-do/scripts/checkpoint`, `skills/myflow-do/scripts/uncommitted-review-package` —
  removed.
- `skills/myflow-contracts/pipeline.md` — Git boundaries table.
- `skills/myflow-contracts/build-green.md` — tag syntax and guard.
- `skills/myflow-contracts/finish-contract.md` — run 1's step sequence.
- `skills/myflow-start/SKILL.md` — writing-plans enrichment gains the new tag family.
- New guard script(s) for the mechanical per-task fields.
