# Proposal: stage-task-files stages a task's deleted paths

## Why

`scripts/stage-task-files.sh` (`stats/internal/guard/stagetaskfiles.go`) runs `git add -- <excludes> <files>`.
A task `**Files:**` path the task deletes and has already `git rm`'d is absent from both the
worktree and the index, so git exits 128 with `fatal: pathspec '<path>' did not match any files`
and the script exits 2. kan-1051 task 6 deleted `stats/web/tests/visual/views.spec.ts` and its
snapshot; its implementer could not use the script and committed from the index by hand.
Reproduced: a path removed with plain `rm` stages fine (exit 0); a path removed with `git rm`
exits 2.

## What changes

- Before `git add`, a named path absent from the worktree and matched by no index entry is dropped
  from the add — its deletion is already staged, so there is nothing to add. When every path is
  dropped, the add step is skipped.
- The planning-path reset and excludes, and the "left unstaged" status check over all named paths,
  are unchanged.
- A regression test in `stats/internal/guard/stage_task_files_test.go` that exits 2 without the fix.
