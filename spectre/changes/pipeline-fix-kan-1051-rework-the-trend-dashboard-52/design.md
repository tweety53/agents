# Design: stage-task-files stages a task's deleted paths

**ID:** first-review-effort

## Decision

In `stageTaskFiles`, filter the add's path list: keep a path when it exists under the worktree
(`os.Lstat`, so a dangling symlink counts) or when `git ls-files -- <path>` prints anything (a
tracked path deleted only from disk, whose deletion `git add` stages). Drop the rest — a path whose
deletion is already in the index. Skip the `git add` step when the filtered list is empty. The
status check still runs over every named path, so a genuinely unstaged path is still refused; a
staged deletion reports `D ` (second column blank) and passes.

## Rejected alternatives

- `git add -A -- <paths>`: fails with the same pathspec error on a path in neither tree nor index
  (verified with git on this machine).
- `git add --ignore-missing`: only valid with `--dry-run`.
- `git rm --cached` for missing paths: errors on a path already removed from the index, and adds a
  step with nothing to do.

## Scope

`stats/internal/guard/stagetaskfiles.go`, `stats/internal/guard/stage_task_files_test.go`. No
contract text changes: the script's header contract already says it stages the task's paths.
Nothing the defect wrote needs repair — it only refused, and kan-1051's hand commit is correct.
