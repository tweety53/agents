# Tasks: stage-task-files stages a task's deleted paths

## Task 1: Skip named paths whose deletion is already staged

**Files:**
- `stats/internal/guard/stagetaskfiles.go`
- `stats/internal/guard/stage_task_files_test.go`

- [ ] Add `TestStageTaskFilesStagesDeletedPaths` to `stats/internal/guard/stage_task_files_test.go`:
  a fixture repo with tracked `gone-rm.txt` and `gone-plain.txt` committed; `git rm -q gone-rm.txt`
  and plain-delete `gone-plain.txt`; run `stage-task-files` with both paths. Assert exit 0, and that
  `git diff --cached --name-status` lists `D` for both.
- [ ] Run `cd stats && go test ./internal/guard -run TestStageTaskFiles` and confirm the new test
  fails (exit 2, pathspec error).
- [ ] In `stageTaskFiles`, before the reset/add loop, build `addFiles` from `files`: keep `f` when
  `os.Lstat(filepath.Join(<abs worktree>, f))` succeeds or `git ls-files -- f` output is non-empty;
  pass `addFiles` to the add step and skip that step when it is empty. Keep the planning reset, the
  excludes-first order and the status check over all `files` unchanged.
- [ ] Re-run the test filter above: every `TestStageTaskFiles*` passes.
- [ ] `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean; `scripts/check-references.sh`
  and the other `scripts/check-*.sh` guards in `.flow/project.md` `## lint` pass.
- [ ] Commit `fix(stats): stage a task path whose deletion is already staged`.
