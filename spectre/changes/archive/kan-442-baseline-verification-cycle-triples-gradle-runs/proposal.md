# kan-442-baseline-verification-cycle-triples-gradle-runs

## Why

Self-review finding from KAN-423 (flow-cost angle). `scripts/check-task-commit-fields.py`
verifies a task's `Regression:` and `Baseline:` fields at runtime by mutating the canonical
worktree: `_uncommitted_work_protected` stashes everything, untracked files included;
`_commit_reverted` runs `git revert --no-commit` and, on the way out, `git reset --hard`; and in
between the project's whole `## test` command runs — twice for `Baseline:` (at the commit and
reverted at its parent) and once more with `--only <name>` for `Regression:`. On KAN-423 that
was three full Gradle suites per task, two of them the guard's.

The two suites buy nothing. Both checks need `## test` to be one command printing `COUNT: <N>`
and `RESULT <name>: pass|fail` — a protocol invented for the guard that no real runner speaks.
This repository's three-command list skips before running anything; KAN-423's Gradle command
paid both suites and then skipped on no `COUNT:` line. No project has ever had a field verified.

The mutation is also the mechanism of KAN-423's incident. `skills/flow/implement.md` runs the
guard while the next bundle's implementer works in the same worktree, on the stated ground that
the guard "reads git objects and `tasks.md` only, so it is safe while the tree changes" — which
it does not. A Gradle suite outran the Bash tool's 120s timeout mid-revert, the conductor
re-invoked the guard on top of it, and the tree was left in a stuck revert with the untracked
`spectre/changes/<name>/` directory swallowed into three orphaned protective stash entries:
roughly 55 minutes of hand recovery and one wasted conductor dispatch (~15M cache-read tokens).
<!-- measured: KAN-442's own description, quoting KAN-423's self-review; a live incident, not re-runnable -->

KAN-438 (flow-fix) and KAN-447 (flow-automation) describe the same incident and propose the same
root fix. This change resolves KAN-438 (a) and (c) and KAN-447 (1); it leaves out KAN-438 (b)'s
lock file and KAN-447 (2)'s recovery script — once the guard mutates nothing there is no
re-entrancy to serialize and no stuck state to recover from.

## What changes

- `scripts/check-task-commit-fields.py` loses the runtime `Regression:`/`Baseline:` verification
  entirely: `check_regression`, `check_baseline`, `_commit_reverted`,
  `_uncommitted_work_protected`, `_STASH_MESSAGE`, `read_single_test_command`,
  `_run_test_command`, the `COUNT:`/`RESULT` regexes and the `.flow/project.md` `## test`
  reader, and `CheckOutcome`. `check_task_commit` returns violations alone; `main` prints
  them and nothing else. The guard's git use is `diff`, `log` and `rev-parse` — read-only, so
  the overlap contract in `implement.md` becomes true rather than asserted.
- `Baseline:` and `Regression:` stay as plan declarations: still written by the planner, still
  parsed by `parse_task_fields` and `check-plan-shape.py`. No plan-side change.
- `scripts/test-check-task-commit-fields.sh` drops the cases and fixtures that exist only for
  the cycle (cases 11–14, 16, 83–86; `write_project_md_test_section`, `write_test_runner`,
  `write_side_effect_test_runner`, `write_double_side_effect_test_runner`,
  `write_unsupported_test_runner`) and gains one case proving the property that replaces
  them: on a worktree carrying staged, unstaged and untracked changes, the guard leaves HEAD,
  index, working tree and `git stash list` exactly as it found them.
- `skills/flow/implement.md`'s conductor step 2 gains a prose-only paragraph (KAN-438 c): after
  any guard call that times out, `git status --porcelain=v2 --branch` and `git stash list` run
  before anything else; a reverting state, dirt the run did not make, or a stash entry the
  conductor did not push ends the turn with `## Question` — never a retry.
