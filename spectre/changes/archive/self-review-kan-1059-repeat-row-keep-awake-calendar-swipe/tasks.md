# self-review-kan-1059-repeat-row-keep-awake-calendar-swipe

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** close kan-1059's seven self-review findings — 1a (Task 1), 1b (Task 2), 1c (Task 3),
1f (Task 4), 5a (Task 5), 1d (Task 6), 1e (Task 7), as `proposal.md` names them.

**Architecture:** a one-round-low trailer match in `check-panel-fix-single-dispatch`; two refusals
on a live re-review's `flow record dispatch begin`; one gate paragraph; a segment match for
`Class.method` test tokens; a commit-less-row refusal in close-task plus a one-off store repair; a
tick bound in one gymie-playwright spec; a derived default origin in gymie-playwright's `api.ts`.

**Tech Stack:** Go (`stats/`), Markdown, TypeScript (gymie-playwright).

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-29-self-review-kan-1059-repeat-row-keep-awake-calendar-swipe-design.md` in this
worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/`. Never merge or push.
- Tasks 1–5 run in `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1059-repeat-row-keep-awake-calendar-swipe` (`<wt>`), branch `self-review-kan-1059-repeat-row-keep-awake-calendar-swipe`; their `**Files:**` paths are relative to `<wt>`.
- Tasks 6–7 run in `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe` (`<pw>`), branch `spectre/kan-1059-repeat-row-keep-awake-calendar-swipe`; their `**Files:**` paths are relative to `<pw>`.
- Never touch a main checkout, and never create a worktree.
- Go: `gofmt -w` every Go file you change; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Mutation proofs restore with `scripts/break-and-prove.sh` or `scripts/mutate-and-verify.sh`, or a saved diff re-applied with `git apply` — never `git checkout --` on a file carrying uncommitted edits.
- Task 5's live repair runs `cd <wt>/stats && make -s web-dist && go run ./cmd/flowd …`, which reaches the dev `flow` database — dev-workspace storage, not production. Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database, and never run tests against the live database.
- Task 6 runs its spec through the gymie `verify` row (`/Users/tweety53/Projects/gymie/.flow/project.md`, `## visual verification`) from `<pw>`, which brings up the kan-1059 workspace's dev stack with `devStackFresh`; nothing else starts, stops or rebuilds that stack. Load runs only under `run-capped.sh` with a bounded worker count.
- Every Go test, Gradle or Playwright run sets a Bash `timeout` of 600000.

## Review Focus

- Task 1: a trailer two rounds low is still refused; the begin check reads author time and skips on a failed read.
- Task 2: a round with only Minor fixed findings, or with a `-mutated none` row, is admitted; a back-fill (`-started-at`) is never refused.
- Task 4: a literal hit still wins; a path token such as `bar.spec.ts` is unaffected.
- Task 5: an appended task with one row carrying a commit passes; a non-appended task is never judged.
- Task 6: no wall-clock bound remains and the throttled run is shown failing before the fix.

- [x] 1. Accept a panel-fix trailer one round low, and refuse an off-round trailer before the re-review
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `scripts/check-panel-fix-single-dispatch.sh /Users/tweety53/Projects/gymie-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe kan-1059-repeat-row-keep-awake-calendar-swipe mf-k1059-20261029T094524Z` exits 1 naming panel-fix-5 and panel-fix-7 <!-- measured: that run → exit 1 @ a5e4f917 -->.
  - [x] **Step 2: Write the guard test.** `TestCheckPanelFixSingleDispatchAcceptsATrailerOneRoundLow` in `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, on the file's fixtures: a row `panel-fix-5` whose commit carries `Flow-Key: panel-fix-4` passes; `panel-fix-5-2` with `panel-fix-4-2` and `panel-fix-5-retry` with `panel-fix-4` pass; `panel-fix-5` with `panel-fix-3` is still refused. Run `cd stats && go test ./internal/guard -run 'TestCheckPanelFixSingleDispatchAcceptsATrailerOneRoundLow' -count=1` — it fails.
  - [x] **Step 3: Fix the guard.** In `pfdCheckCommits` (`stats/internal/guard/panelfixsingledispatch.go`), per decision `pfd-one-round-low`, also accept the trailer `panel-fix-<r-1>` plus the row's own chunk suffix, `-retry` stripped first; the function comment names signal eleven of `check-unfinished-work` and kan-1059 self-review, 1a.
  - [x] **Step 4: Write the begin test.** `TestRecordDispatchRefusesAReReviewOverAnOffRoundPanelFixTrailer` in `stats/cmd/flow/record_test.go`, on `TestRecordDispatchRefusesAPanelFixKeyedARoundLow`'s fakes and a temporary git repository the change's state names as a worktree: round 4's reviewer rows started at T; a commit authored after T carrying `Flow-Key: panel-fix-4` — a live `begin -key panel-5-primary -role reviewer` exits 2 naming the commit and `panel-fix-5`; with the trailer `panel-fix-5` it is admitted; with `-started-at` it is admitted; with the trailer commit authored before T it is admitted. Run `cd stats && go test ./cmd/flow -run 'TestRecordDispatchRefusesAReReviewOverAnOffRoundPanelFixTrailer' -count=1` — it fails.
  - [x] **Step 5: Fix the begin.** In `runRecordDispatchBegin` (`stats/cmd/flow/record.go`), add the check of decision `rereview-refuses-offround-trailer` as a function beside `panelFixKeyedARoundLow`, called for a live key; a failed record or git read refuses nothing. In `skills/flow/review-panel-fix-round.md`, directly after "and `flow record dispatch` refuses `panel-fix-0` (kan-1040).", add the sentence: "A fix commit's `Flow-Key:` trailer names that same fix round, and `flow record dispatch begin` refuses the re-review's reviewer rows while a commit authored since the reviewed round began carries a `Flow-Key: panel-fix-<k>` trailer of any other round (kan-1059 self-review, 1a)."
  - [x] **Step 6: Prove both tests pin the fix.** With `scripts/break-and-prove.sh`, drop the one-round-low acceptance — the guard test fails; restore. Then make the begin check return admitted — the begin test fails; restore.
  - [x] **Step 7: Verify.** `cd stats && gofmt -l internal/guard cmd/flow` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckPanelFixSingleDispatch' -count=1` and `cd stats && go test ./cmd/flow -run 'TestRecordDispatch' -count=1` pass; Step 1's command exits 0; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/panelfixsingledispatch.go`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/review-panel-fix-round.md`
**Tests:** `TestCheckPanelFixSingleDispatchAcceptsATrailerOneRoundLow`, `TestRecordDispatchRefusesAReReviewOverAnOffRoundPanelFixTrailer`
**Regression:** `TestCheckPanelFixSingleDispatchAcceptsATrailerOneRoundLow` fails — signal eleven's prescribed back-fill leaves the panel-close guard red; `TestRecordDispatchRefusesAReReviewOverAnOffRoundPanelFixTrailer` fails — a re-review begins over a fix commit keyed a round low.
**Commit:** fix(stats): accept a panel-fix trailer one round low and refuse an off-round one before the re-review
**After:** none

**Decision:** pfd-one-round-low, rereview-refuses-offround-trailer

- [x] 2. Refuse a re-review before its fix round's mutation rows
  - [x] **Step 1: Write the test.** `TestRecordDispatchRefusesAReReviewBeforeItsMutationRows` in `stats/cmd/flow/record_test.go`, on the same fakes: round 4 holds an Important finding with status `fixed` and no mutation row — a live `begin -key panel-5-primary -role reviewer` exits 2 naming round 4 and `flow record mutation -round 5`; with a mutation row at round 5, or a `-mutated none` row at round 5, it is admitted; with only a row at round 4 it exits 2; with only a Minor fixed finding it is admitted; with `-started-at` it is admitted. Run `cd stats && go test ./cmd/flow -run 'TestRecordDispatchRefusesAReReviewBeforeItsMutationRows' -count=1` — it fails.
  - [x] **Step 2: Fix.** In `runRecordDispatchBegin` (`stats/cmd/flow/record.go`), add the check of decision `rereview-needs-mutation-rows`, on the run record Task 1's check already reads where it can share the read. In `skills/flow/review-panel-fix-round.md`, directly after "Each row is recorded before the round's re-review begins, and `flow record mutation` refuses one after (kan-1046 self-review, 1d).", add the sentence: "`flow record dispatch begin` refuses that re-review's reviewer rows while the reviewed round holds a fixed Critical or Important finding and no fix-mutation row at the fix round (kan-1059 self-review, 1b)."
  - [x] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, skip the check — the test fails; restore.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestRecordDispatch|TestRecordMutation' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/review-panel-fix-round.md`
**Tests:** `TestRecordDispatchRefusesAReReviewBeforeItsMutationRows`
**Regression:** `TestRecordDispatchRefusesAReReviewBeforeItsMutationRows` fails — a re-review begins while its fix round's mutation proof is unrecorded, found only as a late row afterwards.
**Commit:** fix(stats): refuse a re-review before its fix round's mutation rows
**After:** Task 1

**Decision:** rereview-needs-mutation-rows

- [x] 3. Fix a non-recurring verify failure instead of closing it as a false positive
  - [x] **Step 1: Edit.** In `skills/flow/unfinished-work-gate.md`, replace the paragraph's opening, from "**An item no change to the tree can finish** — a verify failure" through "and the run continues to the landing route.", with: "**An item no change to the tree can finish** — a verdict the guard reaches over a report line that names a passing spec — is recorded as a guard false positive per **A verdict the operator calls structural** below, its `-reason` naming that spec's passing run, and the run continues to the landing route. A verify failure that does not recur is never such an item: it is fixed anyway, as **The loop** (`skills/flow/verify-fix-loop.md`) states for an observed failure that does not reproduce (kan-1059 self-review, 1c)." — the rest of the paragraph, from "The outstanding items and their dispositions", unchanged; wrap at the file's width. Per decision `gate-keeps-loop-rule`.
  - [x] **Step 2: Verify.** `git grep -n -F -e 'single-worker, isolated re-runs' -e 'carrying that run count' -- skills/flow/unfinished-work-gate.md` prints nothing; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-normative-inventory.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/unfinished-work-gate.md`
**Tests:** none — a prose rule; `check-verbatim-moves.sh` holds the reworded sentence to `verbatim-moves.txt`.
**Regression:** none
**Commit:** docs(flow): fix a non-recurring verify failure instead of closing it as a false positive
**After:** none

**Decision:** gate-keeps-loop-rule

- [x] 4. Find a Class.method test token by both its segments
  - [x] **Step 1: Write the test.** `TestCheckTaskCommitFieldsFindsAClassDotMethodTest` in `stats/internal/guard/check_task_commit_fields_test.go`, on the file's repository fixtures: a task declaring `**Tests:** \`FooUiTest.barKeepsBaz\`` whose commit adds `class FooUiTest` with `fun barKeepsBaz()` passes both the diff and the tree check; one whose commit adds `class FooUiTest` with only `fun other()` is refused naming the token. Run `cd stats && go test ./internal/guard -run 'TestCheckTaskCommitFieldsFindsAClassDotMethodTest' -count=1` — it fails.
  - [x] **Step 2: Fix.** Per decision `tests-class-dot-method`, in `stats/internal/guard/taskcommitfields.go` (and `planparse.go` if the diff pattern is built there), a token not found literally that matches `^[A-Z][A-Za-z0-9_]*\.[a-z][A-Za-z0-9_]*$` is found when both segments are, in `tcfCheckTests` and `tcfMissingTreeNames`; `tcfTestsParseRule` adds ", a `Class.method` token is found when both its segments are". In `skills/flow/brainstorm-planner.md`'s `**Tests:**` bullet, directly after "write what covers the task outside the field, never inside it.", add the sentence: "A token written `Class.method` is found when both its segments are, so a test may be named as the pipeline's `fix-mutation:` lines name it (kan-1059 self-review, 1f)."
  - [x] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, drop the segment fallback — the test fails; restore.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckTaskCommitFields|TestTcf' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/taskcommitfields.go`, `stats/internal/guard/check_task_commit_fields_test.go`, `skills/flow/brainstorm-planner.md`
**Tests:** `TestCheckTaskCommitFieldsFindsAClassDotMethodTest`
**Regression:** `TestCheckTaskCommitFieldsFindsAClassDotMethodTest` fails — a test declared `Class.method` is reported missing though the commit adds it.
**Commit:** fix(guard): find a Class.method test token by both its segments
**After:** none

**Decision:** tests-class-dot-method

- [x] 5. Refuse closing an appended task whose dispatch rows carry no commit, and repair kan-1059's two rows
  - [x] **Step 1: Write the test.** `TestCloseTaskRefusesAnAppendedTaskRowWithNoCommit` in `stats/internal/guard/close_task_test.go`, on `TestCloseTaskRefusesAnUndispatchedAppendedTask`'s fixture: an appended task whose one naming row has no `commitSha` is refused, exit 1, nothing ticked, stderr naming the row id and `flowd retime-dispatch`; with one naming row carrying a commit it closes; a non-appended task with a commit-less row closes. Run `cd stats && go test ./internal/guard -run 'TestCloseTaskRefusesAnAppendedTaskRowWithNoCommit' -count=1` — it fails.
  - [x] **Step 2: Fix.** Per decision `closetask-appended-needs-commit`, in `stats/internal/guard/closetask.go` extend `ctUndispatched`'s row read with `id` and `commitSha` and report an appended task none of whose naming rows carries a commit, with its own message; the exit stays 1 through the existing `RECORD` path. Update `scripts/close-task.sh`'s step 1 header to say an appended task whose rows carry no commit is refused too.
  - [x] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, drop the commit condition — the test fails; restore.
  - [x] **Step 4: Repair kan-1059's rows.** From `<wt>/stats`: `make -s web-dist && go run ./cmd/flowd retime-dispatch 8623 2026-10-29T14:01:55+02:00 2026-10-29T14:18:42+02:00 f595824df4329065b6fbc58d9c2239048e4b45ec`, then `go run ./cmd/flowd retime-dispatch 8632 2026-10-29T15:05:00+02:00 2026-10-29T15:06:55+02:00 c3f7de89d1c2f416dcc39d24248afd3e96ffa614`, then `go run ./cmd/flowd restamp-dispatch 8623 /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/19e8be9a-7414-433b-8e3e-a443871be918.jsonl` and the same for 8632. Confirm with `flow record dispatches -change kan-1059-repeat-row-keep-awake-calendar-swipe -C /Users/tweety53/Projects/gymie-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe`: rows 8623 and 8632 carry those commits, `agentId` `inline`, instants unchanged. <!-- measured: rows 8623 and 8632 read commitSha null, agentId null, started 14:01:55+02:00/15:05:00+02:00, ended 14:18:42+02:00/15:06:55+02:00 @ 2026-10-29 dev store -->
  - [x] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCloseTask' -count=1` passes; `scripts/check-references.sh` exits 0. Commit; the body names the two repaired rows and their commits.

**Build:** green
**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `scripts/close-task.sh`
**Tests:** `TestCloseTaskRefusesAnAppendedTaskRowWithNoCommit`
**Regression:** `TestCloseTaskRefusesAnAppendedTaskRowWithNoCommit` fails — an appended task closes on a row that records no commit, and the ledger reads "no commit".
**Commit:** fix(guard): refuse closing an appended task whose dispatch rows carry no commit
**After:** none

**Decision:** closetask-appended-needs-commit

- [x] 6. Bound the paused-clock wait by ticked time, not wall time
  - [x] **Step 1: Read the constants.** Confirm `TAG_TIMEOUT_MS` (`<pw>/fixtures.ts`) and `FRAME_TICK_MS` (`fidelity/kan-932-forms.spec.ts`) still read 15_000 and 100. <!-- measured: grep -n 'TAG_TIMEOUT_MS =' fixtures.ts; grep -n 'FRAME_TICK_MS =' fidelity/kan-932-forms.spec.ts @ gymie-playwright 4ed55ad -->
  - [x] **Step 2: Force the failure before touching the code.** In `fidelity/kan-932-forms.spec.ts`, uncommitted, open a CDP session on wr-E3b's page and send `Emulation.setCPUThrottlingRate` with a rate before its `onPausedClock` call; run the spec alone through the `verify` row with `fidelity/kan-932-forms.spec.ts -g 'wr-E3b'`, raising the rate until it fails with `onPausedClock: … not live within 15000ms`. Keep the failing run's result files. unverified:the throttle rate that exhausts the wall bound — find it here.
  - [x] **Step 3: Fix.** Per decision `paused-clock-tick-bound`, in `onPausedClock` delete `deadline` and throw when `until` is set, `liveAt` is still undefined and `ticked >= TAG_TIMEOUT_MS`, the message saying "not live within ${TAG_TIMEOUT_MS}ms of ticked clock"; the doc comment says the bound is fake-clock progress so a loaded machine cannot exhaust it.
  - [x] **Step 4: Prove.** The same throttled run passes; drop the throttle and run wr-E3b alone `--repeat-each 5 --workers 1` through the `verify` row — 5/5 pass. Remove the throttle before committing.
  - [x] **Step 5: Verify.** `cd <pw> && npx playwright test fidelity/kan-932-forms.spec.ts --list` exits 0; `git -C <pw> diff --cached --name-only` names only the spec. Commit; the body states the forced reproduction — the throttle rate, failing before the fix and passing after.

**Build:** green
**Files:** `fidelity/kan-932-forms.spec.ts`
**Tests:** none — repairs existing tests: wr-E3b of `fidelity/kan-932-forms.spec.ts`, run under a forced CPU throttle as the proof.
**Regression:** none
**Commit:** test(fidelity): bound the paused-clock wait by ticked time, not wall time
**After:** none

**Decision:** paused-clock-tick-bound

- [x] 7. Derive the default API origin from GATEWAY_PORT
  - [x] **Step 1: Write the test.** `<pw>/scripts/api-origin.test.mjs`, `node:test` like `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1059-repeat-row-keep-awake-calendar-swipe/scripts/fresh-identity.test.mjs`: test `api origin follows GATEWAY_PORT` sets `GATEWAY_PORT=18080`, deletes `BACKEND_URL`, dynamically imports `../api.ts` with a cache-busting query and asserts `API` is `http://localhost:18080`; a second case with `BACKEND_URL` set asserts it wins. Run `cd <pw> && node --test scripts/api-origin.test.mjs` — the first case fails, reading `http://localhost:8080`. <!-- measured: the same import in a scratch node test, node v24.12.0 → actual 'http://localhost:8080' @ gymie-playwright 4ed55ad -->
  - [x] **Step 2: Fix.** Per decision `api-default-gateway-port`, in `api.ts` the default becomes `` `http://localhost:${process.env.GATEWAY_PORT ?? '8080'}` ``.
  - [x] **Step 3: Prove the test pins the fix.** Revert the line by hand — the test fails; restore.
  - [x] **Step 4: Verify.** `cd <pw> && npm run test:scripts` passes. Commit.

**Build:** green
**Files:** `api.ts`, `scripts/api-origin.test.mjs`
**Tests:** `api origin follows GATEWAY_PORT`
**Regression:** `api origin follows GATEWAY_PORT` fails — a run exporting only `GATEWAY_PORT` points `API` at 8080 and `fixtures.ts` throws.
**Commit:** fix(api): derive the default API origin from GATEWAY_PORT
**After:** none

**Decision:** api-default-gateway-port
