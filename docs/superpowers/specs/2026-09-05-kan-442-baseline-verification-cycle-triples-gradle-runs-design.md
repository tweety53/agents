# KAN-442 — remove the guard's revert/test/reset cycle

**Jira:** KAN-442. Resolves KAN-438 (a) and (c) and KAN-447 (1) by the same change.

## Problem

`scripts/check-task-commit-fields.py` verifies `Regression:` and `Baseline:` at runtime by
mutating the canonical worktree: `_uncommitted_work_protected` stashes everything (untracked
included), `_commit_reverted` runs `git revert --no-commit` then `git reset --hard`, and in
between the project's whole `## test` command runs — twice for `Baseline:`, once more with
`--only` for `Regression:`. Three consequences:

- Two extra full suites per task on top of the implementer's own run (three Gradle suites per
  task on KAN-423).
- The checks verify nothing in practice: they need a single `## test` command that prints
  `COUNT: <N>` / `RESULT <name>: pass|fail`, a protocol no real runner speaks. This repository's
  three-command list skips; KAN-423's Gradle command paid two suites, then skipped on no
  `COUNT:` line.
- `skills/flow/implement.md` runs the guard while bundle N+2's implementer works in the same
  tree and states the guard "reads git objects and `tasks.md` only" — false, and the mechanism
  of KAN-423's incident: a 120s tool timeout mid-suite, a re-entered guard, a stuck revert, the
  untracked `spectre/changes/<name>/` swallowed by the stash, three orphaned stash entries, ~55
  minutes of hand recovery.

## Change

1. **Delete the runtime verification.** Remove `check_regression`, `check_baseline`,
   `_commit_reverted`, `_uncommitted_work_protected`, `_STASH_MESSAGE`,
   `read_single_test_command`, `_run_test_command`, the `TARGETED_RESULT_RE`, `TOTAL_COUNT_RE`,
   `PROJECT_TEST_SECTION_RE`, `MARKDOWN_HEADING_RE`, `FENCE_LINE_RE` constants, the
   `CheckOutcome` type, the `contextlib`/`shlex`/`subprocess`-for-tests imports they alone
   used, and `check_task_commit`'s `notices` return. `main` prints violations only. The module
   docstring and exit-code table stop describing the two checks; the citation of the frozen
   `openspec/specs/myflow-task-commit-fields/spec.md` requirement "Regression and Baseline checks
   skip, rather than fail" is reworded to record that the check was removed and why. The guard
   then runs `git diff`/`git log` only — the property `implement.md` already claims.
2. **Delete the harness cases and fixtures that exist only for the cycle:** cases 11, 12, 13,
   14, 16, 83, 84, 85, 86 and the five fixtures `write_project_md_test_section`, `write_test_runner`,
   `write_side_effect_test_runner`, `write_double_side_effect_test_runner`,
   `write_unsupported_test_runner` in `scripts/test-check-task-commit-fields.sh`. One
   new case asserts the property that replaces them: the guard leaves HEAD, index, working tree
   and `git stash list` untouched on a worktree carrying staged, unstaged and untracked changes.
3. **The fields stay as plan declarations.** `Baseline:` and `Regression:` are still written by
   the planner (`skills/flow/brainstorm-planner.md`), still parsed by `parse_task_fields`
   (`task.baseline` is kept; `check-plan-shape.py`'s F1 duplicate check reads the same
   grammar). Nothing in this change touches the plan side.
4. **Conductor rule for a hung guard (KAN-438 c).** `skills/flow/implement.md`'s "The next
   implementer overlaps the guard" step 2 gains one paragraph: after any guard call that times
   out, run `git status --porcelain=v2 --branch` and `git stash list` in that worktree before
   anything else; a `# ... reverting` / `# stash` state, a dirty tree the run did not produce,
   or a stash entry the conductor did not make is a stop that ends the turn with `## Question`,
   never a retry. Prose only, no `SHALL`/`MUST`; the file stays under its 32500-byte budget
   (29818 today).

## Out of scope

- A lock file against re-entrant guard runs (KAN-438 b): nothing mutates, so there is nothing to
  serialize.
- A recovery script for stuck reverts and orphaned stashes (KAN-447 2): the state it recovers
  from can no longer be produced by the guard.
- Any git-object substitute for `Baseline:` (KAN-438 a / KAN-447 1's `git grep -c "@Test"`):
  annotation counts never equal a runner's executed-test totals, so declared `before=/after=`
  would fail real tasks.
- The frozen `openspec/specs/` tree is not edited; `spectre/specs/` stays empty. The live
  requirement is the guard's docstring and `implement.md`.

## Decisions

- **delete-not-relocate** — chosen over running the cycle in a throwaway `git worktree add
  --detach` (no mutation but still two suites for an unspoken protocol) and over the git-object
  count (false violations).
- **fold-in-438c** — the conductor's post-timeout check is prose-only and lands in the same
  file the guard's overlap contract lives in.

## Open questions

None.
