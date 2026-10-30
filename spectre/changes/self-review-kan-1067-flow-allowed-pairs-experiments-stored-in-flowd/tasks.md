# self-review-kan-1067-flow-allowed-pairs-experiments-stored-in-flowd

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** close kan-1067's six self-review findings — 1-A (Task 1), 1-B (Task 2), 1-C (Task 3),
1-D (Task 4), 1-E (Task 5), 1-F (Task 6), as `proposal.md` names them.

**Architecture:** a wider commit set in `check-task-records`; an unconditional plan-shape call in
`close-task`; a key refusal in `run-verify-list`; a source-file read in one Go test; a regex and a
name mapping in one script; one writing-plans sentence.

**Tech Stack:** Go (`stats/`), Bash, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-30-self-review-kan-1067-flow-allowed-pairs-experiments-stored-in-flowd-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/`. Never merge or push.
- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd` (`<wt>`), branch `spectre/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd`; `**Files:**` paths are relative to `<wt>`. Never touch a main checkout, and never create a worktree.
- Go: `gofmt -w` every Go file you change; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Every prose edit lands the exact sentence this plan gives: the change's `verbatim-moves.txt` already lists it as `scripts/check-verbatim-moves.sh` prints it, so a reworded sentence fails that guard.
- Mutation proofs restore with `scripts/break-and-prove.sh` or `scripts/mutate-and-verify.sh` — never `git checkout --` on a file carrying uncommitted edits.
- Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database, and never run tests against the live database.
- Every Go test run sets a Bash `timeout` of 600000.

## Review Focus

- Task 1: a `Task-Id:` commit naming another task adds nothing; `Task-Id: 1+2` counts for both; a fix commit's undeclared path is still reported; a merge-only subject commit is still accepted silently.
- Task 2: the plan-shape call runs before any commit guard and its exit 1 still ticks and pushes nothing.
- Task 3: the refusal happens before the `begin` call and before any command; the two-argument form is never judged; `verify-3d-viewer` in a repository named `3d-viewer` is admitted.

- [ ] 1. Join a task's Task-Id fix commits in check-task-records
  - [ ] **Step 1: Write the test.** `TestCheckTaskRecordsJoinsTaskIdFixCommits` in `stats/internal/guard/check_task_records_test.go`, on `trRepo`. Each case writes a ticked task 1 whose `**Files:**` names `src/f1.go` and `src/f2.go` (`writeFile` of `fx.tasks`, as case 5 of `TestCheckTaskRecords` does), commits the plan, then `fx.commitFile("src/f1.go", "feat(a): one")` with task 1's `**Commit:**` `feat(a): one`, then one fix commit through `fx.commitFile(<path>, "fix(a): review fix\n\nTask-Id: <ids>")`. Cases: `src/f2.go` with `Task-Id: 1` → exit 0, silent; `src/f2.go` with `Task-Id: 1+2` → exit 0; `src/f2.go` with `Task-Id: 2` → exit 1, a line naming `src/f2.go is not touched` (and the fix commit, carrying no task subject, is otherwise ignored); `src/f9.go` with `Task-Id: 1` → exit 1, lines naming `src/f9.go is not declared` and `src/f2.go is not touched`. Run `cd stats && go test ./internal/guard -run 'TestCheckTaskRecordsJoinsTaskIdFixCommits' -count=1` — the first, second and last cases fail.
  - [ ] **Step 2: Implement.** In `stats/internal/guard/taskrecords.go`, per decision `records-join-task-id-commits`, add `trTaskIDCommits(git func(...string) *exec.Cmd, repo, base string, ids []string, self string) []string`: `capture(git("-C", repo, "log", "--no-merges", "--format=%H %(trailers:key=Task-Id,valueonly,separator=%x2C)", base+"..HEAD"))`, returning each sha other than `self` whose trailer value's `ctTaskNumberRE.FindAllString` holds one of `ids`; a failed `git log` returns nil. In `trCheck`'s group loop, after the merge-commit `continue`, append `trCommitFiles(git, g.repo, sha)` for each such sha to `changed` (no duplicates, `trAppendNew`) before `tcfCheckFiles` and the untouched-file loop. Extend the header's check 2 paragraph with: "The files of every non-merge `<base-ref>..HEAD` commit whose `Task-Id:` trailer names one of the group's tasks are joined to the subject commit's (the rule check-task-commit-fields' tcfTaskIDCommits applies, KAN-976), so a gated review's fix commit is judged with its task (kan-1067 self-review, 1-A)."
  - [ ] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, make `trTaskIDCommits` return nil — the test fails; restore.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckTaskRecords' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/guard/taskrecords.go`, `stats/internal/guard/check_task_records_test.go`
**Tests:** `TestCheckTaskRecordsJoinsTaskIdFixCommits`
**Regression:** `TestCheckTaskRecordsJoinsTaskIdFixCommits` fails — a declared path only a `Task-Id:` fix commit touched is reported untouched again, contradicting check-task-commit-fields.
**Commit:** fix(stats): join a task's Task-Id fix commits in check-task-records
**After:** none

**Decision:** records-join-task-id-commits

- [ ] 2. Run check-plan-shape on every close-task
  - [ ] **Step 1: Write the test.** `TestCloseTaskRunsPlanShapeOnAWidenedPlannedTask` in `stats/internal/guard/close_task_test.go`, on `TestCloseTaskRunsPlanShapeOnInRunFix`'s shape: `ctFixture(t, true, "0", "0")`, the plan `ctPlan` with task 2's `**Files:** ¤b.txt¤` replaced by `**Files:** ¤b.txt¤, ¤b_test.go¤` (its `**Tests:**` stays `none`), task 2 committed touching `b.txt` and `b_test.go`; closing task 2 exits 1, stdout+stderr carry "opens with `none` but **Files:** names a test-shaped path", and no `flow tasks tick` call is logged. In `TestCloseTaskRunsPlanShapeOnInRunFix`, rename the case "a planned task never runs the guard" to "a planned task over the same plan is refused" with `refused` true, and change the doc comment's last clause to "and a planned task over the same bad plan is refused too (kan-1067 self-review, 1-B)". Run `cd stats && go test ./internal/guard -run 'TestCloseTaskRunsPlanShape' -count=1` — both fail.
  <!-- measured: with the uwAppended condition replaced by `true`, `cd stats && go test ./internal/guard -run TestCloseTask -count=1` failed only TestCloseTaskRunsPlanShapeOnInRunFix/a_planned_task_never_runs_the_guard @ ca792d7b -->
  - [ ] **Step 2: Implement.** In `stats/internal/guard/closetask.go`, per decision `plan-shape-on-every-close`, drop the `uwAppended(...)` condition so `check-plan-shape` runs whenever `tcfResolveTasksMD` resolves a `tasks.md`; replace the comment above it with: "The plan guard runs on every close, before any commit guard: an appended or in-run fix task was written after it last ran (kan-1046 self-review, 1f), and a `flow tasks files add` widening of a planned task can break plan shape too (kan-1067 self-review, 1-B: a test path under `Tests: none`, found only at verify)."
  - [ ] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, restore the `uwAppended` condition — `TestCloseTaskRunsPlanShapeOnAWidenedPlannedTask` fails; restore.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCloseTask' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`
**Tests:** `TestCloseTaskRunsPlanShapeOnAWidenedPlannedTask`
**Regression:** `TestCloseTaskRunsPlanShapeOnAWidenedPlannedTask` fails — a widened planned task breaking plan shape closes and surfaces only at verify.
**Commit:** fix(stats): run check-plan-shape on every close-task
**After:** none

**Decision:** plan-shape-on-every-close

- [ ] 3. Refuse an off-contract verify key in run-verify-list
  - [ ] **Step 1: Write the test.** `TestRunVerifyListRefusesAnOffContractKey` in `stats/internal/guard/run_verify_list_test.go`, on the recording fixture of the test holding `rec := []string{"kan-1", "verify", …}` (fake `flow` logging its calls, `project("true")` as a `lint` list). `wt` is `git init`-ed. Admitted (exit 0, the log opens with the `begin` carrying that key): `verify`, `verify-2`, `verify-fix-1`, `verify-` + `filepath.Base(wt)`, `verify-fix-2-` + `filepath.Base(wt)`. Refused (exit 2, stderr names the key and `verify-and-handoff.md`, the log empty): `verify-fix-1b`, `verify-2b`, `verify-other`. Repository-name form: `git -C wt worktree add <tmp>/other` (one commit first), run on `<tmp>/other` with key `verify-` + `filepath.Base(wt)` — admitted. Run `cd stats && go test ./internal/guard -run 'TestRunVerifyListRefusesAnOffContractKey' -count=1` — it fails.
  - [ ] **Step 2: Implement.** In `stats/internal/guard/runverifylist.go`, per decision `verify-key-refused-at-run`, inside the `len(args) >= 7` block right after `key, token, model, effort := …`, add `verifyKeyOK(env, abs(args[0]), key) bool`: `g := vgFixSuffix.ReplaceAllString(key, "verify$2")`; true when `g == "verify"`, `g == "verify-"+filepath.Base(wt)`, or `g == "verify-"+<repo>`, `<repo>` the basename of `filepath.Dir` of `git -C <wt> rev-parse --path-format=absolute --git-common-dir`'s output (skipped when git fails). On false print `run-verify-list: key "<key>" does not fold to verify or verify-<repository name> (skills/flow/verify-and-handoff.md, **Recording.**) — no command ran` on stderr and return 2. Extend the header's exit-2 line with ", or a recording <key> that does not fold, by check-verify-green's vgFixSuffix, to `verify` or `verify-<s>`, <s> the worktree's basename or its repository's name (kan-1067 self-review, 1-C: `verify-fix-1b` formed its own group and verify stayed NOT-GREEN)".
  - [ ] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, make `verifyKeyOK` return true — the test fails; restore.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestRunVerifyList' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`
**Tests:** `TestRunVerifyListRefusesAnOffContractKey`
**Regression:** `TestRunVerifyListRefusesAnOffContractKey` fails — an off-contract key is recorded again and check-verify-green reports NOT-GREEN with no hint.
**Commit:** fix(stats): refuse an off-contract verify key in run-verify-list
**After:** none

**Decision:** verify-key-refused-at-run

- [ ] 4. Read the policy table from its source in the experiments tests
  - [ ] **Step 1: Implement.** In `stats/internal/api/experiments_test.go`, per decision `api-test-reads-policy-source`, `newExperimentsTestServer` reads `table, err := os.ReadFile("../../../skills/flow-contracts/model-policy.md")` (`t.Fatalf` on error) and passes `table` to `api.WithExperiments`; drop the `modelpolicy/embedded` import; in `TestRolesRoute`'s doc comment replace "over the real embedded table" with "over the real model-policy table".
  <!-- measured: with stats/internal/modelpolicy/embedded/model-policy.md moved aside, `cd stats && go vet ./internal/api` failed "pattern model-policy.md: no matching files found" @ ca792d7b -->
  - [ ] **Step 2: Prove the fix.** Move `stats/internal/modelpolicy/embedded/model-policy.md` to a temp path; `cd stats && go vet ./internal/api && go test ./internal/api -run 'TestRolesRoute|TestExperiment' -count=1` passes; move the file back.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l internal/api` prints nothing; `cd stats && go test ./internal/api -run 'TestRolesRoute|TestExperiment' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/api/experiments_test.go`
**Tests:** none — repairs existing tests: `TestRolesRoute`, `TestExperimentsRoutes`, `TestExperimentStatsRoute`
**Regression:** none — Step 2 shows `go vet ./internal/api` failing to compile without the make copy if reverted.
**Commit:** test(api): read the policy table from its source in the experiments tests
**After:** none

**Decision:** api-test-reads-policy-source

- [ ] 5. Admit a hyphenated name in live-store-copy
  - [ ] **Step 1: Write the test.** In `scripts/test-live-store-copy.sh`, set `name="lsc-test$$"` and add `db="flow_lsc_test$$_live"`; replace every `flow_${name}_live` in the assertions with `$db`; add `-a` to the bad-name list. Run `scripts/test-live-store-copy.sh` — it fails (`up` exits 2).
  - [ ] **Step 2: Implement.** In `scripts/live-store-copy.sh`, per decision `live-copy-hyphen-maps-to-underscore`, the name check becomes `[[ $name =~ ^[a-z0-9][a-z0-9_-]*$ ]]` with the message `name '$name' is not [a-z0-9][a-z0-9_-]*`, and `db="flow_${name//-/_}_live"`; in the header, after the `up` paragraph's first sentence, state: "A `-` in <name> becomes `_` in the database name only (kan-1067 self-review, 1-E)."; and replace the remaining `flow_<name>_live` mentions in the header with `flow_<name, - as _>_live`.
  - [ ] **Step 3: Verify.** `scripts/test-live-store-copy.sh` prints `all cases passed`. Commit.

**Build:** green
**Files:** `scripts/live-store-copy.sh`, `scripts/test-live-store-copy.sh`
**Tests:** `scripts/test-live-store-copy.sh`
**Regression:** `scripts/test-live-store-copy.sh` fails — `up lsc-test<pid>` exits 2 again.
**Commit:** fix(scripts): admit a hyphenated name in live-store-copy
**After:** none

**Decision:** live-copy-hyphen-maps-to-underscore

- [ ] 6. Verify a Markdown-editing task with the whole lint list
  - [ ] **Step 1: Edit.** In `skills/flow/brainstorm-planner.md`, section D, inside the block quote opening "> **Write each task's verify step as its own lint commands plus targeted tests.**", directly after "A task whose `**Tests:**` is `none` names lint alone and no test command.", add, wrapped at the file's width and continuing the block quote: "A task whose `**Files:**` holds a Markdown file under `skills/` or `rules/` is the exception: its verify step runs the whole list, `run-verify-list.sh <worktree> lint`, because the prose guards that list carries are too many to hand-pick and the one a subset leaves out fails only at `flow.verify` (kan-1067 self-review, 1-F)."
  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh <wt> lint` — every command exits 0 save `scripts/check-task-records.sh`, whose only lines name this change's tasks as not ticked (the ticks are the self-review loop's own step). Commit.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** none — a writing-plans sentence.
**Regression:** none — prose; `scripts/check-verbatim-moves.sh` in Step 2's list pins the exact sentence.
**Commit:** docs(flow): verify a Markdown-editing task with the whole lint list
**After:** none

**Decision:** markdown-tasks-run-whole-lint
