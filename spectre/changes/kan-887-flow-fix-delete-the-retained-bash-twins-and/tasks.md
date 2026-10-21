# kan-887-flow-fix-delete-the-retained-bash-twins-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** delete the bash twins of Go-ported guards that nothing sources any more, and the bash
harnesses whose every case a Go table test already carries.

**Architecture:** pure deletion plus comment rewrites; Go parity tests that read a deleted file
read it at the pinned base `ef6f87cd` through `historicalFile` (`stats/internal/guard/helpers_test.go`).

**Spec:** `design.md` beside this file.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-887-flow-fix-delete-the-retained-bash-twins-and`, branch `spectre/kan-887-flow-fix-delete-the-retained-bash-twins-and`; every `**Files:**` path is relative to it. Never touch the main checkout, never create a worktree, never merge or push.
- `scripts/lib/spec-root.sh` is not deleted (Decision `keep-spec-root-for-kan-923`).
- A rewritten comment that cited a deleted file names the Go twin; a Go twin's header records "its bash twin was deleted (KAN-887)" without naming the deleted path (Decision `record-in-go-header`).
- Run `cd stats && gofmt -w .` on Go changes before committing. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- Every task's verify also runs `scripts/check-references.sh`, which must exit 0.

- [ ] 1. Delete reproducer-metachars.sh
  - [ ] **Step 1: Confirm no live consumer.** `git grep -n -F 'reproducer-metachars.sh' -- scripts stats hooks mods skills rules` shows only comments and `runreproducer_test.go`/`check_panel_reproducers_test.go`.
  - [ ] **Step 2:** `git rm scripts/reproducer-metachars.sh`; delete `TestMetacharsMatchBashSource` and `rrMetachars` from `stats/internal/guard/runreproducer_test.go` and the bash half of case 19 in `stats/internal/guard/check_panel_reproducers_test.go` (Decision `drop-metachars-bash-pin`); case 20's comment no longer names the file.
  - [ ] **Step 3:** rewrite the comments in `scripts/check-panel-reproducers.sh`, `stats/internal/guard/metachars.go`, `stats/internal/guard/panelreproducers.go`, `stats/internal/guard/runreproducer.go` per Global constraints.
  - [ ] **Step 4: Verify.** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `cd stats && go test ./internal/guard/ -run 'TestCheckPanelReproducers|TestRunReproducer' -count=1` passes; `scripts/check-references.sh` exits 0; `git grep -n -F -e 'reproducer-metachars.sh' -e 'rrMetachars' -- scripts stats` prints nothing.
  - [ ] **Step 5: Commit** with the subject below.

**Build:** green
**Files:** `scripts/reproducer-metachars.sh`, `scripts/check-panel-reproducers.sh`, `stats/internal/guard/metachars.go`, `stats/internal/guard/panelreproducers.go`, `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`, `stats/internal/guard/check_panel_reproducers_test.go`
**Tests:** none — repairs existing tests: `TestCheckPanelReproducers` (case 19 keeps its literal pin), `TestMetacharsMatchBashSource` removed.
**Regression:** none — deletion of a dead copy.
**Commit:** refactor(scripts): delete the reproducer-metachars bash twin
**After:** none

**Decision:** drop-metachars-bash-pin
**Decision:** record-in-go-header

- [ ] 2. Delete change-plan.sh and its harness
  - [ ] **Step 1: Confirm no live consumer.** `git grep -n -E '(source|\.) .*change-plan\.sh' -- scripts stats hooks mods` prints only lines inside `scripts/lib/change-plan.sh` and `scripts/test-lib-change-plan.sh`; confirm `stats/internal/guard/helpers_test.go` carries every `ok:` case of `scripts/test-lib-change-plan.sh`, as its own comment claims — a missing case is added to `helpers_test.go` first and named in this task's `**Tests:**`. unverified:every test-lib-change-plan.sh case has a helpers_test.go counterpart.
  - [ ] **Step 2:** `git rm scripts/lib/change-plan.sh scripts/test-lib-change-plan.sh`.
  - [ ] **Step 3:** rewrite every comment naming either file in `scripts/check-task-commit-fields.sh`, `scripts/check-unfinished-work.sh`, `scripts/gather-dispatch-context.sh`, `scripts/lib/spec-root.sh`, `stats/cmd/flow/state.go`, `stats/cmd/flow/tasks.go`, `stats/internal/api/changes.go`, `stats/internal/client/client_test.go`, `stats/internal/guard/changename.go`, `stats/internal/guard/changeplan.go`, `stats/internal/guard/gatherdispatch.go`, `stats/internal/guard/helpers_test.go`, `stats/internal/guard/taskcommitfields.go`; in `skills/flow-contracts/pipeline.md` drop "and its bash twin `<agents repo>/scripts/lib/change-plan.sh`" from the `changePlanDir` sentence.
  - [ ] **Step 4: Verify.** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `cd stats && go build ./... && go test ./internal/guard/ -run 'TestChangePlan|TestSpecRoot|TestGatherDispatchContext' -count=1` passes; `scripts/check-references.sh` and `scripts/check-installed-citations.sh` exit 0; `git grep -n -F -e 'change-plan.sh' -e 'test-lib-change-plan' -- scripts stats skills rules hooks mods` prints nothing.
  - [ ] **Step 5: Commit** with the subject below.

**Build:** green
**Files:** `scripts/lib/change-plan.sh`, `scripts/test-lib-change-plan.sh`, `scripts/check-task-commit-fields.sh`, `scripts/check-unfinished-work.sh`, `scripts/gather-dispatch-context.sh`, `scripts/lib/spec-root.sh`, `skills/flow-contracts/pipeline.md`, `stats/cmd/flow/state.go`, `stats/cmd/flow/tasks.go`, `stats/internal/api/changes.go`, `stats/internal/client/client_test.go`, `stats/internal/guard/changename.go`, `stats/internal/guard/changeplan.go`, `stats/internal/guard/gatherdispatch.go`, `stats/internal/guard/helpers_test.go`, `stats/internal/guard/taskcommitfields.go`
**Tests:** none — repairs existing tests: `TestChangePlan`'s comments, unless Step 1 adds a missing case.
**Regression:** none — deletion of a dead copy.
**Commit:** refactor(scripts): delete the change-plan bash twin and its harness
**After:** Task 1

**Decision:** record-in-go-header

- [ ] 3. Delete project-section.sh
  - [ ] **Step 1: Confirm no live consumer.** `git grep -n -F 'project-section.sh' -- scripts hooks mods` shows only `scripts/lib/strip-bom.sh`'s comment and the file itself.
  - [ ] **Step 2:** `git rm scripts/lib/project-section.sh`; add to `stats/internal/guard/libtwins_test.go` a `bashLibAt` variant of `bashLibIO` that writes the named libs at `ef6f87cd` (`historicalFile`) into a temp dir and sources from there, and point `TestProjectSectionParity` at it with `project-section.sh` and `strip-bom.sh` (Decision `parity-at-historical-ref`).
  - [ ] **Step 3:** rewrite the comments in `scripts/lib/strip-bom.sh`, `stats/internal/guard/projectsection.go`, `stats/internal/guard/gatherdispatch.go`, `stats/internal/guard/visualsection.go` per Global constraints.
  - [ ] **Step 4: Verify.** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `cd stats && go test ./internal/guard/ -run 'TestProjectSectionParity|TestProjectGet|TestStripBOMParity' -count=1 -v 2>&1 | grep -c -- '--- SKIP'` prints 0 and the run passes; `scripts/check-references.sh` exits 0; `git grep -n -F 'project-section.sh' -- scripts stats` prints only `historicalFile`/`bashLibAt` arguments in `libtwins_test.go` and `project_get_test.go`.
  - [ ] **Step 5: Commit** with the subject below.

**Build:** green
**Files:** `scripts/lib/project-section.sh`, `scripts/lib/strip-bom.sh`, `stats/internal/guard/projectsection.go`, `stats/internal/guard/gatherdispatch.go`, `stats/internal/guard/visualsection.go`, `stats/internal/guard/libtwins_test.go`
**Tests:** none — repairs existing tests: `TestProjectSectionParity`, now against the pinned copy.
**Regression:** none — deletion of a dead copy.
**Commit:** refactor(scripts): delete the project-section bash twin
**After:** Task 2

**Decision:** delete-project-section
**Decision:** parity-at-historical-ref
**Decision:** record-in-go-header

- [ ] 4. Delete panel-touched-paths.sh
  - [ ] **Step 1: Confirm no live consumer.** `git grep -n -F 'panel-touched-paths.sh' -- scripts hooks mods` shows only `scripts/write-panel-diff.sh`'s comment and the file itself.
  - [ ] **Step 2:** `git rm scripts/lib/panel-touched-paths.sh`; point `TestPanelTouchedPathsParity`'s three `bashLib`/`bashLibEnv` calls at Task 3's `bashLibAt` with `panel-touched-paths.sh`.
  - [ ] **Step 3:** rewrite the comments in `scripts/write-panel-diff.sh` and `stats/internal/guard/paneltouchedpaths.go` per Global constraints.
  - [ ] **Step 4: Verify.** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `cd stats && go test ./internal/guard/ -run 'TestPanelTouchedPathsParity' -count=1 -v 2>&1 | grep -c -- '--- SKIP'` prints 0 and the run passes; `scripts/check-references.sh` exits 0; `git grep -n -F 'panel-touched-paths.sh' -- scripts stats` prints only `bashLibAt` arguments in `libtwins_test.go`.
  - [ ] **Step 5: Commit** with the subject below.

**Build:** green
**Files:** `scripts/lib/panel-touched-paths.sh`, `scripts/write-panel-diff.sh`, `stats/internal/guard/paneltouchedpaths.go`, `stats/internal/guard/libtwins_test.go`
**Tests:** none — repairs existing tests: `TestPanelTouchedPathsParity`, now against the pinned copy.
**Regression:** none — deletion of a dead copy.
**Commit:** refactor(scripts): delete the panel-touched-paths bash twin
**After:** Task 3

**Decision:** parity-at-historical-ref
**Decision:** record-in-go-header

- [ ] 5. Delete test-git-shim.sh and its harness
  - [ ] **Step 1: Confirm no live consumer.** `git grep -n -F 'test-git-shim' -- scripts stats hooks mods` shows only the two files and a comment in `stats/internal/guard/check_base_moved_test.go`.
  - [ ] **Step 2:** `git rm scripts/lib/test-git-shim.sh scripts/test-lib-test-git-shim.sh`; rewrite the comment in `stats/internal/guard/check_base_moved_test.go`.
  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `scripts/check-references.sh` exits 0; `git grep -n -F 'test-git-shim' -- scripts stats` prints nothing.
  - [ ] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `scripts/lib/test-git-shim.sh`, `scripts/test-lib-test-git-shim.sh`, `stats/internal/guard/check_base_moved_test.go`
**Tests:** none — repairs existing tests: `check_base_moved_test.go`'s comment.
**Regression:** none — deletion of dead code.
**Commit:** refactor(scripts): delete the unused test-git-shim helper
**After:** none

- [ ] 6. Delete the test-aside-planning-artifacts.sh harness
  - [ ] **Step 1: Audit before touching anything; do not guess.** List every case `scripts/test-aside-planning-artifacts.sh` asserts (exit code, stdout, stderr) and name the `stats/internal/guard/aside_planning_artifacts_test.go` case asserting the same; any case with no counterpart is added to `aside_planning_artifacts_test.go` first, and its name appended to this task's `**Tests:**` field. Record the mapping in the commit body.
  - [ ] **Step 2:** `git rm scripts/test-aside-planning-artifacts.sh`; drop the harness's mention from `scripts/test-guard-autosquash.sh`'s comment.
  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `cd stats && go test ./internal/guard/ -run 'TestAsidePlanningArtifactsParity' -count=1` passes; `scripts/check-references.sh` exits 0; `git grep -n -F 'test-aside-planning-artifacts.sh' -- scripts stats skills` prints nothing.
  - [ ] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `scripts/test-aside-planning-artifacts.sh`, `stats/internal/guard/aside_planning_artifacts_test.go`, `scripts/test-guard-autosquash.sh`
**Tests:** none — repairs existing tests: `TestAsidePlanningArtifactsParity`, unless Step 1 adds a missing case.
**Regression:** none — deletion of a duplicate harness.
**Commit:** test(guard): delete the bash harness duplicating aside_planning_artifacts_test.go
**After:** none

**Decision:** harness-case-audit

- [ ] 7. Delete the test-check-done-when-paths.sh harness
  - [ ] **Step 1: Audit before touching anything; do not guess.** List every case `scripts/test-check-done-when-paths.sh` asserts (exit code, stdout, stderr) and name the `stats/internal/guard/donewhenpaths_test.go` case asserting the same; any case with no counterpart is added to `donewhenpaths_test.go` first, and its name appended to this task's `**Tests:**` field. Record the mapping in the commit body.
  - [ ] **Step 2:** `git rm scripts/test-check-done-when-paths.sh`; `git mv stats/internal/guard/donewhenpaths_test.go stats/internal/guard/check_done_when_paths_test.go` — `scripts/run-guard-tests.sh` accepts a `check-*.sh` shim without a bash companion only when `stats/internal/guard/<name with - as _>_test.go` exists.
  - [ ] **Step 3: Verify.** `scripts/run-guard-tests.sh` does not print `run-guard-tests: check-done-when-paths.sh has no companion`; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `cd stats && go test ./internal/guard/ -run 'TestCheckDoneWhenPaths' -count=1` passes; `scripts/check-references.sh` exits 0; `git grep -n -F 'test-check-done-when-paths.sh' -- scripts stats skills` prints nothing.
  - [ ] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `scripts/test-check-done-when-paths.sh`, `stats/internal/guard/donewhenpaths_test.go`, `stats/internal/guard/check_done_when_paths_test.go`
**Tests:** none — repairs existing tests: `TestCheckDoneWhenPaths`, unless Step 1 adds a missing case.
**Regression:** none — deletion of a duplicate harness.
**Commit:** test(guard): delete the bash harness duplicating donewhenpaths_test.go
**After:** none

**Decision:** harness-case-audit

- [ ] 8. Delete the test-check-hand-notes-in-step.sh harness
  - [ ] **Step 1: Audit before touching anything; do not guess.** List every case `scripts/test-check-hand-notes-in-step.sh` asserts (exit code, stdout, stderr) and name the `stats/internal/guard/check_hand_notes_test.go` case asserting the same; any case with no counterpart is added to `check_hand_notes_test.go` first, and its name appended to this task's `**Tests:**` field. Record the mapping in the commit body.
  - [ ] **Step 2:** `git rm scripts/test-check-hand-notes-in-step.sh`; `git mv stats/internal/guard/check_hand_notes_test.go stats/internal/guard/check_hand_notes_in_step_test.go` — `scripts/run-guard-tests.sh` accepts a `check-*.sh` shim without a bash companion only when `stats/internal/guard/<name with - as _>_test.go` exists.
  - [ ] **Step 3: Verify.** `scripts/run-guard-tests.sh` does not print `run-guard-tests: check-hand-notes-in-step.sh has no companion`; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `cd stats && go test ./internal/guard/ -run 'TestCheckHandNotesInStep' -count=1` passes; `scripts/check-references.sh` exits 0; `git grep -n -F 'test-check-hand-notes-in-step.sh' -- scripts stats skills` prints nothing.
  - [ ] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `scripts/test-check-hand-notes-in-step.sh`, `stats/internal/guard/check_hand_notes_test.go`, `stats/internal/guard/check_hand_notes_in_step_test.go`
**Tests:** none — repairs existing tests: `TestCheckHandNotesInStep`, unless Step 1 adds a missing case.
**Regression:** none — deletion of a duplicate harness.
**Commit:** test(guard): delete the bash harness duplicating check_hand_notes_test.go
**After:** none

**Decision:** harness-case-audit

- [ ] 9. Delete the test-project-get.sh harness
  - [ ] **Step 1: Audit before touching anything; do not guess.** List every case `scripts/test-project-get.sh` asserts (exit code, stdout, stderr) and name the `stats/internal/guard/project_get_test.go` case asserting the same; any case with no counterpart is added to `project_get_test.go` first, and its name appended to this task's `**Tests:**` field. Record the mapping in the commit body.
  - [ ] **Step 2:** `git rm scripts/test-project-get.sh`.
  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `cd stats && go test ./internal/guard/ -run 'TestProjectGetParity|TestProjectGetEnum' -count=1` passes; `scripts/check-references.sh` exits 0; `git grep -n -F 'test-project-get.sh' -- scripts stats skills` prints nothing.
  - [ ] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `scripts/test-project-get.sh`, `stats/internal/guard/project_get_test.go`
**Tests:** none — repairs existing tests: `TestProjectGetParity`, `TestProjectGetEnum`, unless Step 1 adds a missing case.
**Regression:** none — deletion of a duplicate harness.
**Commit:** test(guard): delete the bash harness duplicating project_get_test.go
**After:** none

**Decision:** harness-case-audit

- [ ] 10. Delete the test-check-fast-route-record.sh harness
  - [ ] **Step 1: Audit before touching anything; do not guess.** List every case `scripts/test-check-fast-route-record.sh` asserts (exit code, stdout, stderr) and name the `stats/internal/guard/fastrouterecord_test.go` case asserting the same; any case with no counterpart is added to `fastrouterecord_test.go` first, and its name appended to this task's `**Tests:**` field. Record the mapping in the commit body.
  - [ ] **Step 2:** `git rm scripts/test-check-fast-route-record.sh`; `git mv stats/internal/guard/fastrouterecord_test.go stats/internal/guard/check_fast_route_record_test.go` — `scripts/run-guard-tests.sh` accepts a `check-*.sh` shim without a bash companion only when `stats/internal/guard/<name with - as _>_test.go` exists.
  - [ ] **Step 3: Verify.** `scripts/run-guard-tests.sh` does not print `run-guard-tests: check-fast-route-record.sh has no companion`; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0; `cd stats && go test ./internal/guard/ -run 'TestFastRouteRecord' -count=1` passes; `scripts/check-references.sh` exits 0; `git grep -n -F 'test-check-fast-route-record.sh' -- scripts stats skills` prints nothing.
  - [ ] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `scripts/test-check-fast-route-record.sh`, `stats/internal/guard/fastrouterecord_test.go`, `stats/internal/guard/check_fast_route_record_test.go`
**Tests:** none — repairs existing tests: `TestFastRouteRecord*`, unless Step 1 adds a missing case.
**Regression:** none — deletion of a duplicate harness.
**Commit:** test(guard): delete the bash harness duplicating fastrouterecord_test.go
**After:** none

**Decision:** harness-case-audit
