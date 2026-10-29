# self-review-kan-1045-flow-one-allowed-model-effort-list-per-subtask

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** close kan-1045's six self-review findings — F-A (Task 1), F-B (Task 2), F-C (Task 3),
F-D (Task 4), F-E (Task 5), F-F (Task 6), as `proposal.md` names them.

**Architecture:** one project-config sentence; a `remove` verb on `flow tasks files`; a merger
refusal on `flow record dispatch begin`; four citation fixes; one shared flow-guard cache in the
repotest hook runner; one flow-fast sentence.

**Tech Stack:** Go (`stats/`), Markdown, Bash comments.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-29-self-review-kan-1045-flow-one-allowed-model-effort-list-per-subtask-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/`. Never merge or push.
- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1045-flow-one-allowed-model-effort-list-per-subtask` (`<wt>`), branch `spectre/kan-1045-flow-one-allowed-model-effort-list-per-subtask`; `**Files:**` paths are relative to `<wt>`. Never touch a main checkout, and never create a worktree.
- Go: `gofmt -w` every Go file you change; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Every prose edit lands the exact sentence this plan gives: the change's `verbatim-moves.txt` already lists each one as `scripts/check-verbatim-moves.sh` prints it, so a reworded sentence fails that guard.
- Mutation proofs restore with `scripts/break-and-prove.sh` or `scripts/mutate-and-verify.sh` — never `git checkout --` on a file carrying uncommitted edits.
- Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database, and never run tests against the live database. Load runs only under `scripts/run-capped.sh`.
- Every Go test run sets a Bash `timeout` of 600000.

## Review Focus

- Task 2: removing the last entry writes `none`; a fenced `**Files:**` line and another task's identical line stay untouched; `add` behaves exactly as before.
- Task 3: a back-filled `panel-fix-<r>` is refused too; a merger row still open, or ended `failed`, does not satisfy the check; a failed record read refuses nothing; a re-review of round 0 is never checked.
- Task 3: every existing admitted-begin fixture in `stats/cmd/flow/record_test.go` keeps passing for its own reason, not because the new check short-circuits first.
- Task 5: a caller passing its own `FLOW_GUARD_CACHE_DIR` or `HOME` still wins; the shared directory is removed after the run.

- [x] 1. Name a ported guard's Go companion test with `-` replaced by `_`
  - [x] **Step 1: Edit.** In `.flow/project.md`, replace "a Go test in `stats/internal/guard/<name>_test.go` standing in for it (kan-903)." with "a Go test in `stats/internal/guard/<name, - replaced by _>_test.go` standing in for it — `check_dispatch_pair_test.go` for `check-dispatch-pair.sh` (kan-903)."
  - [x] **Step 2: Verify.** `git grep -n -F -e 'stats/internal/guard/<name>_test.go' -- .flow/project.md` prints nothing; `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** none — a project-config sentence.
**Regression:** none — prose; the sweep in Step 2 is the check.
**Commit:** docs(project-config): name a ported guard's Go companion test with - replaced by _
**After:** none

- [x] 2. Narrow a task's Files field with `flow tasks files … remove`
  - [x] **Step 1: Write the test.** `TestTasksFilesRemoveNarrowsOneTask` in `stats/cmd/flow/tasks_test.go`, on `TestTasksFilesAddWidensOneTask`'s fixture shape: task 1's Files line naming x.go, a.go and b.go, a fenced Files line naming a.go inside task 1, and task 3 carrying task 1's identical line. `files("1", "remove", "a.go")` exits 0 and task 1's line names x.go and b.go, the fenced line and task 3 unchanged; `files("1", "remove", "x.go", "b.go")` leaves task 1's line reading none; `files("1", "remove", "q.go")` exits 1 and the plan is byte-identical; `files("9", "remove", "a.go")` exits 1; `files("1", "remove")` exits 2. Run `cd stats && go test ./cmd/flow -run 'TestTasksFiles' -count=1` — the new test fails.
  - [x] **Step 2: Implement.** In `stats/cmd/flow/tick.go`, per decision `tasks-files-remove`, add `filesRemove(plan []byte, id string, paths []string) ([]byte, error)` beside `filesAdd`; both call one shared `editFilesField(plan []byte, id string, edit func(field string) (string, error)) ([]byte, error)` holding the existing locate loop. In `stats/cmd/flow/tasks.go`, `runTasksFiles` accepts `add` or `remove` in `pos[2]` (usage error otherwise, exit 2), prints `files of task <id>: <paths>` as before; the usage line becomes `flow tasks files [-C dir] <change> <task-id> add|remove <path>...`, and the help gains, after the `add` paragraph: "tasks files ... remove narrows it, dropping each backticked <path> entry -- a field left empty becomes "none" -- and refuses (exit 1), leaving the plan unchanged, a <path> the field does not name." Update `runTasksFiles`' doc comment to name both verbs.
  - [x] **Step 3: Edit the contract.** In `skills/flow/implement.md`, replace "The parent transcribes a widening with `flow tasks files -C <worktree> <change> <id> add <path>…`, which edits that one task's `**Files:**` field, never with a hand `sed` over `tasks.md`." with "The parent transcribes a widening with `flow tasks files -C <worktree> <change> <id> add <path>…`, and a narrowing — a declared path the commit legitimately leaves untouched — with `flow tasks files -C <worktree> <change> <id> remove <path>…`, each editing that one task's `**Files:**` field, never with a hand `sed` over `tasks.md`." Wrap at the file's width.
  - [x] **Step 4: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, make `runTasksFiles` refuse `remove` again — the test fails; restore.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestTasksFiles' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/tick.go`, `stats/cmd/flow/tasks.go`, `stats/cmd/flow/tasks_test.go`, `skills/flow/implement.md`
**Tests:** `TestTasksFilesRemoveNarrowsOneTask`
**Regression:** `TestTasksFilesRemoveNarrowsOneTask` fails — a narrowing again needs the hand edit of `tasks.md` the contract forbids.
**Commit:** feat(stats): narrow a task's Files field with flow tasks files remove
**After:** none

**Decision:** tasks-files-remove

- [x] 3. Refuse a panel fix or re-review before its round's findings merger
  - [x] **Step 1: Write the test.** `TestRecordDispatchRefusesAFixOrReReviewBeforeItsRoundsMerger` in `stats/cmd/flow/record_test.go`, on the `panelFixBegin` fakes and `TestRecordDispatchRefusesAReReviewBeforeItsMutationRows`' fakes: round 0 holds an Important finding. A live `begin -key panel-fix-1 -role panel-fix` with no `panel-0-merger` row exits 2 naming `panel-0-merger`, nothing posted; with a `panel-0-merger` row of role `merger`, an end and outcome `completed` it is admitted; with that row open, or ended `failed`, it exits 2; a back-filled `panel-fix-1` (`-started-at`/`-ended-at`) with no merger row exits 2; a failed record read is admitted. A live `begin -key panel-1-primary -role reviewer` with no `panel-0-merger` row exits 2 naming it; with the completed row it is admitted; `panel-0-primary` is never checked. Run `cd stats && go test ./cmd/flow -run 'TestRecordDispatchRefusesAFixOrReReviewBeforeItsRoundsMerger' -count=1` — it fails.
  - [x] **Step 2: Implement.** In `stats/cmd/flow/record.go`, per decision `merger-enforced-at-begin`, add `mergerMissing(run records.Run, round int) bool` beside `panelFixBeginRefusal`. In `runRecordDispatchBegin`, for a key matching `panelFixKeyRE` — live or back-filled — read the run record (a failed read refuses nothing) and, before `panelFixBeginRefusal`'s checks, refuse exit 2 with `flow: -key "<key>": round <r-1> holds no completed panel-<r-1>-merger row -- the parent reads a round's findings through one findings merger before it fixes or re-reviews them (skills/flow/review-panel.md)`; open `reReviewBlocked` with the same check for round `r-1`, the same wording.
  - [x] **Step 3: Repair the fixtures.** Run `cd stats && go test ./cmd/flow -run 'TestRecord' -count=1`; every existing test that expects a `panel-fix-<r>` or `panel-<r>-*` (r ≥ 1) begin admitted and now meets the new refusal gains, in its stored record, a `{"key":"panel-<r-1>-merger","role":"merger","outcome":"completed",…}` dispatch with an `endedAt`; a test expecting a refusal keeps refusing for its own reason (its stderr assertion unchanged). Change no assertion.
  - [x] **Step 4: Edit the contract.** In `skills/flow/review-panel.md`, directly after "**The parent reads a round's findings through one findings merger, never the reports one by one.**", add the sentence: "`flow record dispatch begin` refuses a `panel-fix-<r>` row, and a reviewer row of panel round `<r>`, while round `<r-1>` holds no completed `panel-<r-1>-merger` row (kan-1045 self-review, F-C)." Wrap at the file's width.
  - [x] **Step 5: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, make `mergerMissing` return false — the test fails; restore.
  - [x] **Step 6: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestRecordDispatch|TestRecord' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/review-panel.md`
**Tests:** `TestRecordDispatchRefusesAFixOrReReviewBeforeItsRoundsMerger`
**Regression:** `TestRecordDispatchRefusesAFixOrReReviewBeforeItsRoundsMerger` fails — a fix round or re-review begins over a round whose findings no merger read.
**Commit:** fix(stats): refuse a panel fix or re-review before its round's findings merger
**After:** none

**Decision:** merger-enforced-at-begin

- [x] 4. Cite Allowed pairs where the deleted pairs table was named
  - [x] **Step 1: Edit `skills/flow/implement.md`.** Replace "(the pairs table, `skills/flow-self-review/SKILL.md`; kan-1033)" with "(**Allowed pairs**, `skills/flow-contracts/model-policy.md`; kan-1033)".
  - [x] **Step 2: Edit `scripts/pipeline-fix.sh`.** In the `escalate` usage comment, replace "which check-loop-dispatch.sh requires before admitting <key> on its pairs row's escalation type; refused unless the pairs table admits an escalation for <key> and an earlier round of its kind completed." with "which check-loop-dispatch.sh requires before admitting <key> on its role's marked `No-progress escalation` row of **Allowed pairs** (skills/flow-contracts/model-policy.md); refused unless that table gives <key>'s role such a row and an earlier round of its kind completed." — rewrapped as `#       ` comment lines at the block's width.
  - [x] **Step 3: Edit the test comments.** In `stats/internal/guard/pipeline_fix_test.go`, "(the pairs table, a .flow/project.md, …" becomes "(model-policy.md's **Allowed pairs**, a .flow/project.md, …"; in `stats/internal/guard/check_loop_dispatch_test.go`, "an agents repo carrying the real pairs table," becomes "an agents repo carrying the real model-policy.md (**Allowed pairs**),". Rewrap each comment at its width.
  - [x] **Step 4: Verify.** `git grep -n -F -e 'the pairs table' -e 'the real pairs table' -- skills/flow/implement.md scripts/pipeline-fix.sh stats/internal/guard/pipeline_fix_test.go stats/internal/guard/check_loop_dispatch_test.go` prints nothing; `cd stats && gofmt -l internal/guard` prints nothing; `bash -n scripts/pipeline-fix.sh` exits 0; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/implement.md`, `scripts/pipeline-fix.sh`, `stats/internal/guard/pipeline_fix_test.go`, `stats/internal/guard/check_loop_dispatch_test.go`
**Tests:** none — repairs existing tests: comments only in pipeline_fix_test.go and check_loop_dispatch_test.go, no test added.
**Regression:** none — prose; the sweep in Step 4 is the check.
**Commit:** docs(flow): cite Allowed pairs where the deleted pairs table was named
**After:** Task 2

- [x] 5. Share one flow-guard cache across the repotest hook runs
  - [x] **Step 1: Trace — reproduce before touching anything; do not guess.** On the untouched tree, run at most three times, each under `scripts/run-capped.sh --`, `bash -c 'cd stats && (go test ./... -count=1 >/dev/null 2>&1 &) ; go test ./internal/repotest -run TestEnforceAgentBaselineHook -count=5'` — the hook tests under the full suite's load. Record in the commit body whether any hook test failed, quoting its output (a deny on an allow case, or flow-guard's "sources … changed during its build"), or that none failed within the three runs. Then show the per-call builds: `cd stats && time go test ./internal/repotest -run TestEnforceAgentBaseline -count=1` with no `FLOW_GUARD_CACHE_DIR`, and again with it exported to one fresh temp directory — the plan-time figures were 12.302s against 3.955s, one key directory built <!-- measured: cd stats && time go test ./internal/repotest -run TestEnforceAgentBaseline -count=1, unset vs FLOW_GUARD_CACHE_DIR=$(mktemp -d) @ branch spectre/kan-1045-flow-one-allowed-model-effort-list-per-subtask -->; record both times in the commit body.
  - [x] **Step 2: Write the test.** `TestRunHookSharesOneGuardCache` in `stats/internal/repotest/helpers_test.go`: write `<t.TempDir()>/env.py` printing `os.environ.get("FLOW_GUARD_CACHE_DIR", "")`, run it twice through `runHook` (absolute path); both outputs are equal and non-empty, and `filepath.Glob(<that dir>/*/flow-guard)` finds a binary. Run `cd stats && go test ./internal/repotest -run 'TestRunHookSharesOneGuardCache' -count=1` — it fails.
  - [x] **Step 3: Implement.** Per decision `hook-tests-share-guard-cache`, in `stats/internal/repotest/helpers_test.go`: a package `sync.OnceValues` that makes one `os.MkdirTemp` directory and builds flow-guard into it with `bash -c '. "$1" && flow_guard_exec spec-root 2 guard-cache:' _ <repo>/scripts/lib/flow-guard.sh` under `FLOW_GUARD_CACHE_DIR=<dir>`, as `guardCacheBuilt` (`stats/internal/guard/helpers_test.go`) does, returning an error when no `*/flow-guard` was built; `runHook` fails the test on that error and appends `FLOW_GUARD_CACHE_DIR=<dir>` after `HOME` and `goCacheEnv()` and before the caller's `env`; a `TestMain` runs the tests, then removes the directory when it was made. Update `runHook`'s doc comment to name the shared cache and why (kan-1045 self-review, F-E).
  - [x] **Step 4: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, drop the `FLOW_GUARD_CACHE_DIR` append — the test fails; restore. When Step 1 reproduced a failure, re-run its load command once with the fix in place and record that it no longer fails.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l internal/repotest` prints nothing; `cd stats && make -s web-dist && go vet ./internal/repotest` exits 0; `cd stats && go test ./internal/repotest -run 'TestRunHookSharesOneGuardCache|TestEnforceAgentBaselineHook' -count=1` passes. Commit, the body carrying Step 1's trace.

**Build:** green
**Files:** `stats/internal/repotest/helpers_test.go`
**Tests:** `TestRunHookSharesOneGuardCache`
**Regression:** `TestRunHookSharesOneGuardCache` fails — every hook call builds flow-guard into its own `HOME` again, under load the build that can turn an allow into a deny.
**Commit:** test(stats): share one flow-guard cache across the repotest hook runs
**After:** none

**Decision:** hook-tests-share-guard-cache

- [x] 6. Send flow-fast dispatches to the closed list for the Role line too
  - [x] **Step 1: Edit.** In `skills/flow-fast/SKILL.md`, replace "Every dispatch's `description` follows **Dispatch sites — the parent's closed list** (`skills/flow/implement.md`)." with "Every dispatch's `description` and `Role:` line follow **Dispatch sites — the parent's closed list** (`skills/flow/implement.md`)."
  - [x] **Step 2: Verify.** `git grep -n -F -e "Every dispatch's \`description\` follows" -- skills/flow-fast/SKILL.md` prints nothing; `cd stats && go test ./internal/repotest -run 'TestEveryRoleLineIsAPolicyRow' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow-fast/SKILL.md`
**Tests:** none — one citation sentence.
**Regression:** none — prose; the sweep in Step 2 is the check.
**Commit:** docs(flow-fast): send flow-fast dispatches to the closed list for the Role line too
**After:** none
