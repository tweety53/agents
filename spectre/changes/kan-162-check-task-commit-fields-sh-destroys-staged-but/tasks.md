# kan-162-check-task-commit-fields-sh-destroys-staged-but

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

- [x] 1. Protect uncommitted work around the Regression:/Baseline: revert cycle

  - [ ] **Step 1: Write the failing tests first (RED)** — in
    `scripts/test-check-task-commit-fields.sh`:
    - **Revise case 16** ("the initial `git revert` inside `_commit_reverted` itself fails"). Under
      the new stash-first behavior, the fixture's premise — a local uncommitted modification to
      `suite.txt` conflicting with the revert — no longer produces a revert conflict, because the
      dirty state is now stashed away before the revert runs, so the revert always applies cleanly
      against the now-clean tree. Change the assertions to: the guard now **succeeds** (`RC` 0),
      HEAD is unchanged, the worktree is clean immediately after the guard returns, and the
      previously-dirty `suite.txt` modification (`alpha_test_MODIFIED_LOCALLY`) is **restored** —
      proving the stash/pop round-tripped it rather than losing it. Keep the case's number and its
      surrounding comment block; rewrite the comment to state the new premise and the new
      assertions, per this file's `verified:`/`unverified:` convention where a command's real
      output changed.
    - **New case 83 — staged-but-uncommitted file survives.** Reuse case 11's fixture
      (`write_test_runner`, a task declaring `Tests: alpha_test`). Before calling `run_guard`,
      write a second file (e.g. `planning.txt`) and `git add` it without committing. Run the guard
      against case 11's own commit shape. Assert: `RC` 0 (regression check still passes), and
      `git status --porcelain` still reports `planning.txt` staged (`A  planning.txt`) after the
      guard returns — the exact KAN-144 failure mode, now proven not to reproduce.
    - **New case 84 — untracked file survives.** Same fixture, but instead write an untracked file
      (no `git add`) before `run_guard`. Assert `RC` 0 and that the untracked file still exists on
      disk with its original content afterward, proving `--include-untracked` covers it.
    - **New case 85 — a stash-pop conflict refuses rather than resetting.** Construct a test runner
      command (a variant of `write_test_runner`) whose script, as a side effect of running,
      recreates an **untracked** file with content that differs from what will be in the outer
      stash — since `git reset --hard` never touches untracked files, this leaves a real collision
      for the final `git stash pop` to hit (`error: Untracked working tree file … would be
      overwritten`). Stage the same-named untracked file before `run_guard` so it is what the outer
      stash captures. Assert: `RC` 2, stderr names `git stash list` as the recovery path, `git
      stash list` in the fixture repo still shows one entry (the stash was never dropped), and the
      guard did **not** run `git reset --hard` over it — assert this by checking the conflicting
      file still carries the test runner's side-effect content, not the stashed original silently
      discarded.
    - Run `bash scripts/test-check-task-commit-fields.sh` and confirm every new/revised assertion
      fails for the expected reason (case 16 fails because the guard still reports non-zero today;
      cases 83/84 fail because `planning.txt`/the untracked file are gone after `run_guard`; case
      85 fails because there is no stash-conflict handling to trigger yet — the guard's `_run`
      calls will not produce the untracked-file collision at all since nothing stashes today).

  - [ ] **Step 2: Implement (GREEN)** — in `scripts/check-task-commit-fields.py`:
    - Add a new context manager, e.g. `_uncommitted_work_protected(worktree)`, that on entry runs
      `git stash push --include-untracked -m "check-task-commit-fields: protecting uncommitted
      work"` via `run_git`, and inspects its output: `git stash push` prints `No local changes to
      save` on a no-op (exit 0, nothing stashed) — detect this and skip the pop. Otherwise, in a
      `finally`, run `git stash pop`; if that raises (non-zero exit from `run_git`), do **not**
      catch it into a `reset --hard` — let it propagate as the guard's own hard failure, with a
      message naming `git stash list` as the recovery path, and exit 2 from `main`'s exception
      handling (or wherever this repository's guards already turn an unexpected exception into
      exit 2 with a message — follow the existing pattern in this file rather than inventing a new
      one).
    - In `check_task_commit`, wrap the existing `for outcome in (check_regression(...),
      check_baseline(...))` loop's construction in `with _uncommitted_work_protected(worktree):` —
      one stash/pop cycle around both checks, not one per check. `_commit_reverted` itself is
      unchanged: it still does its own `git revert` / `git reset --hard commit_sha` pair, now
      running against a tree the outer stash has already made clean of anything worth protecting.
    - Run `bash scripts/test-check-task-commit-fields.sh` again; every case must pass, including
      the untouched ones (case 11, 13, etc.) — the stash/pop must be transparent to the existing
      passing behavior.

  - [ ] **Step 3: Refactor** — re-read the new context manager and its call site against this
    repository's existing comment conventions in this file (a "WHY" block explaining the
    non-obvious choice, matching `_commit_reverted`'s own docstring style) before committing.

**Files:** `scripts/check-task-commit-fields.py`, `scripts/test-check-task-commit-fields.sh`
**Tests:** `case 16 (revised)`, `case 83`, `case 84`, `case 85`
**Regression:** reverting this task's commit restores the unconditional `git reset --hard`
without a protecting stash, so case 83 and case 84 fail again (the staged/untracked file is gone
after `run_guard`), and case 85 fails because no stash exists to refuse over (the guard would
instead silently succeed or hit a plain revert conflict, never reporting the stash-list recovery
message).
**Baseline:** before=202 after=213 (predicted: unverified — 202 measured by running
`bash scripts/test-check-task-commit-fields.sh` against the pre-change file on 2026-08-30; 213 is
a prediction of 202 plus roughly 4 new assertions in case 83, 4 in case 84, 4 in case 85 net of
case 16's assertion count staying the same; the real count is whatever the implementer's actual
assertions total, measured the same way at GREEN)
**Commit:** fix(scripts): stash uncommitted work around the task-commit-fields revert cycle
**Build:** green
