# self-review-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1044's five self-review findings — F1-a (Task 1), F1-b (Task 2), F1-c (Task 3),
F1-e (Task 4), A8-a (Task 5), as `proposal.md` names them.

**Architecture:** a store re-end flag for the transcript restamp; an inline end that derives its
start, two duplicate refusals and a row-drop command in `flow record` and `flowd`; a panel-fix
trailer signal in `check-unfinished-work`; a scrim crop in `measure-visual-properties`; one prose
paragraph; live store repairs through `flowd` and the back-fill CLI, Task 1's first.

**Tech Stack:** Go (`stats/`), Bash, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-28-self-review-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/`. Never merge or push.
- Every task runs in `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets` (`<wt>` below), branch `self-review-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets`. Every `**Files:**` path is relative to `<wt>`. Never touch the main checkout, and never create a worktree.
- Go: `gofmt -w` every Go file you change; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Mutation proofs restore with `break-and-prove.sh` or `mutate-and-verify.sh`, or a saved diff re-applied with `git apply` — never `git checkout --` on a file carrying uncommitted edits.
- Live store repairs (Tasks 1–3) run `cd <wt>/stats && go run ./cmd/flowd <subcommand> …` or `go run ./cmd/flow record …`, which reach the dev `flow` database. They run in task order: Task 1's repair before Task 2's and Task 3's. Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database, and never run tests against the live database.
- Store reads: `docker exec flow-postgres psql -U flow -d flow -tAc "<query>"`. Go test runs set a Bash `timeout` of 600000.

## Review Focus

- Task 1: `ReEnd` is unreachable from JSON; a first end and a re-point still keep a closed row's end; `restamp-ends` never moves an inline row, a whole-second or millisecond end, or an end to before its start.
- Task 2: a row begun at its work's start keeps that start; a non-implementer inline row and a non-inline row are unaffected; a failed record read refuses nothing; `drop-dispatch` refuses rows of two changes.
- Task 3: the signal reads the pin only when it resolves; `flow.commit-two` and non-panel-fix trailers are ignored; an archived change's re-run is not asked.
- Task 4: a frame without a scrim measures byte-identically to before.

- [ ] 1. The transcript restamp re-ends a stamped row
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tAc "select count(*) from dispatches where ended_at >= '2026-10-28T15:30:00Z' and agent_id is not null and agent_id <> 'inline' and ended_at <> date_trunc('milliseconds', ended_at)"` prints 66 <!-- measured: that query @ 2026-10-28T21:39Z -->, and the three panel-10 reviewer rows of `kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets` share one end, 17:26:30.6Z.
  - [ ] **Step 2: Write the tests** in `stats/internal/store/records_test.go`, on `newTestStore`: `TestApplyDispatchEndAtRestampsOnTheStore` — begin an agent row, call `api.ApplyDispatchEndAt` with no `EndedAt` and an `agentEnd` returning an instant after the start and before now; the row's `ended_at` is that instant. `TestRestampStampedEnds` — rows ended after `since` with a microsecond end, a millisecond end, a whole-second end, an inline row and one whose agent end precedes its start; `RestampStampedEnds` moves only the first, and returns 1. Run `cd stats && go test ./internal/store -run 'TestApplyDispatchEndAtRestampsOnTheStore|TestRestampStampedEnds' -count=1` — both fail (the second does not compile).
  - [ ] **Step 3: Fix.** Add `ReEnd bool` with tag `json:"-"` to `records.DispatchEnd` (`stats/internal/records/types.go`), doc comment saying only `ApplyDispatchEndAt` sets it. In `Store.EndDispatch` (`stats/internal/store/records.go`), write `ended_at = CASE WHEN $12 THEN $7 ELSE COALESCE(d.ended_at, $7) END`, passing `in.ReEnd`. In `ApplyDispatchEndAt` (`stats/internal/api/records.go`), set `in.ReEnd = true` before the second call. Add `Store.RestampStampedEnds(ctx, since time.Time, agentEnd func(string) (time.Time, bool)) (int, error)` per decision `restamp-reends`, and `flowd restamp-ends <since RFC3339>` in `stats/cmd/flowd/main.go`, shaped as `retime-dispatch`, wiring `harvest.AgentTranscriptEnd` over `harvest.DefaultTranscriptsRoot()` as `serve` does and printing `restamped: <n>` (kan-1044 self-review, F1-a).
  - [ ] **Step 4: Prove the tests pin the fix.** With `break-and-prove.sh`, drop `in.ReEnd = true` — `TestApplyDispatchEndAtRestampsOnTheStore` fails; restore.
  - [ ] **Step 5: Repair.** `cd stats && go run ./cmd/flowd restamp-ends 2026-10-28T15:30:00Z`; re-run Step 1's query — below 66 by the printed count; the panel-10 rows no longer share one end.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l internal cmd` prints nothing; `cd stats && go test ./internal/store -run 'TestApplyDispatchEndAt|TestRestampStampedEnds|TestEndDispatch' -count=1` and `cd stats && go test ./internal/api -run 'TestEndDispatchRoute' -count=1` pass. Commit.

**Build:** green
**Files:** `stats/internal/records/types.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/internal/api/records.go`, `stats/cmd/flowd/main.go`
**Tests:** `TestApplyDispatchEndAtRestampsOnTheStore`, `TestRestampStampedEnds`
**Regression:** `TestApplyDispatchEndAtRestampsOnTheStore` fails — a stamped end stays at the close call; `TestRestampStampedEnds` fails — the repair moves the wrong rows.
**Baseline:** before=112 after=114 <!-- measured: grep -c -E '^func Test' stats/internal/store/records_test.go @ 3aab64d6 -->
**Commit:** fix(stats): let the transcript restamp re-end a stamped dispatch row
**After:** none

**Decision:** restamp-reends

- [ ] 2. Inline implementer rows record the parent's work once, from its start
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tAc "select id, dispatch_key, outcome, commit_sha, started_at, ended_at, metrics->'tokens' is not null from dispatches where id in (8410,8411,8434,8437,8449,8458)"` — 8410 aborted with tokens, 8411 completed on `37c868ea` without, 8434 and 8437 both completed on `2ca8e5bd`, 8449 begun 16:09:21 and 8458 at 16:46:14, seconds before their commits (16:09:21Z, 16:46:20Z).
  - [ ] **Step 2: Write the tests.** In `stats/internal/store/records_test.go`: `TestEndDispatchTakesAnEarlierStart` — an end carrying `StartedAt` before the row's start moves it there, one after leaves it; `TestDropDispatchMovesTokensToTheKeptRow` — the drop row's `metrics.tokens` lands on a keep row without tokens, the drop row is gone, two rows of different changes are refused. In `stats/cmd/flow/record_test.go`, on the existing back-fill and begin fixtures: `TestRecordDispatchEndInlineDerivesItsStart` — a `completed` end of an inline `task-9-implementer` row with `-commit` sends the derived start; `TestRecordDispatchBackfillRefusesADuplicateTaskCommit` — a `task-8-implementer-retry` back-fill on the commit a completed `task-8-implementer` row carries exits 2; `TestRecordDispatchBeginRefusesACompletedTask` — an inline begin of `task-7-implementer` exits 2 while a completed `task-7-implementer-retry` row exists. Run `cd stats && go test ./internal/store ./cmd/flow -run 'TestEndDispatchTakesAnEarlierStart|TestDropDispatchMovesTokensToTheKeptRow|TestRecordDispatchEndInlineDerivesItsStart|TestRecordDispatchBackfillRefusesADuplicateTaskCommit|TestRecordDispatchBeginRefusesACompletedTask' -count=1` — all fail.
  - [ ] **Step 3: Fix.** Per decisions `inline-end-derives-start` and `one-completed-row-per-task-commit`: add `StartedAt *time.Time` (`json:"startedAt,omitempty"`) to `records.DispatchEnd`; in `Store.EndDispatch`, apply `LEAST(d.started_at, $start)` and use the corrected start in its window checks; extract the inline derivation of `runRecordDispatchBackfill` (`stats/cmd/flow/record.go`) into `inlineDerivedWindow` and call it from `runRecordDispatchEnd`; add the two refusals; add `Store.DropDispatch` and `flowd drop-dispatch <drop-id> <keep-id>` (`stats/cmd/flowd/main.go`), shaped as `drop-pass` (kan-1044 self-review, F1-b).
  - [ ] **Step 4: Prove the tests pin the fix.** With `break-and-prove.sh`, remove the derived start from the end call — `TestRecordDispatchEndInlineDerivesItsStart` fails; remove the back-fill refusal — `TestRecordDispatchBackfillRefusesADuplicateTaskCommit` fails; restore each.
  - [ ] **Step 5: Repair (after Task 1's Step 5).** `cd stats && go run ./cmd/flowd drop-dispatch 8410 8411` and `go run ./cmd/flowd drop-dispatch 8437 8434`. For 8449 and 8458, per decision `retime-late-inline-rows`: read the latest `ended_at` at or before the row's commit time (16:09:21Z, 16:46:20Z) of a row of the change with `agent_id = 'inline'` or role `verifier` and a key starting `visual-verify`, other than the row itself, and run `go run ./cmd/flowd retime-dispatch <id> <that end> <the row's ended_at> <the row's commit_sha>`. <!-- unverified:the derived starts — about 15:57:06Z and 16:29Z once Task 1 has moved the verifier ends --> Re-run Step 1's query: 8410 and 8437 gone, 8411 carries tokens, 8449 and 8458 start at those ends.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l internal cmd` prints nothing; `cd stats && go test ./internal/store -run 'TestEndDispatch|TestDropDispatch' -count=1` and `cd stats && go test ./cmd/flow -run 'TestRecordDispatch' -count=1` pass. Commit.

**Build:** green
**Files:** `stats/internal/records/types.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/cmd/flowd/main.go`
**Tests:** `TestEndDispatchTakesAnEarlierStart`, `TestDropDispatchMovesTokensToTheKeptRow`, `TestRecordDispatchEndInlineDerivesItsStart`, `TestRecordDispatchBackfillRefusesADuplicateTaskCommit`, `TestRecordDispatchBeginRefusesACompletedTask`
**Regression:** `TestRecordDispatchEndInlineDerivesItsStart` fails — a row begun at its commit records seconds of minutes of work; `TestRecordDispatchBackfillRefusesADuplicateTaskCommit` and `TestRecordDispatchBeginRefusesACompletedTask` fail — a task's work is recorded twice; `TestEndDispatchTakesAnEarlierStart` and `TestDropDispatchMovesTokensToTheKeptRow` fail — the store ignores the derived start, or the repair loses tokens.
**Baseline:** before=268 after=273 <!-- measured: cat stats/internal/store/records_test.go stats/cmd/flow/record_test.go | grep -c -E '^func Test' @ 3aab64d6 -->
**Commit:** fix(stats): record an inline implementer's work once, from its start
**After:** Task 1

**Decision:** inline-end-derives-start

- [ ] 3. The unfinished-work gate flags a panel-fix commit with no row
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets`, `git log --format='%h %(trailers:key=Flow-Key,valueonly)' refs/flow/pre-reshape/kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets | grep panel-fix` lists `3b9a3985 panel-fix-3`, `ae761ec7 panel-fix-6`, `25394c9a panel-fix-10`; the store holds no `panel-fix-3` … `-11` row for the change.
  - [ ] **Step 2: Write the test** `TestUnfinishedWorkFlagsAPanelFixCommitWithNoRow` in `stats/internal/guard/check_unfinished_work_test.go`, on `TestCheckUnfinishedWork`'s fixtures: a range whose commit carries `Flow-Key: panel-fix-3` and dispatch rows keyed `panel-fix-1` only — the gate exits 1 naming the commit and `panel-fix-3`; the same commit reachable only from `refs/flow/pre-reshape/<name>` is flagged too; with a `panel-fix-3` row it passes. Run `cd stats && go test ./internal/guard -run 'TestUnfinishedWorkFlagsAPanelFixCommitWithNoRow' -count=1` — it fails.
  - [ ] **Step 3: Fix.** In `stats/internal/guard/unfinishedwork.go`, add the signal of decision `gate-panel-fix-trailer-row` inside the `anchor == worktree` block beside signal seven, reusing its `dispatches`, `ranges` and `capture(envGit(env)(…))`; the tip is `refs/flow/pre-reshape/<name>` when `git rev-parse -q --verify` resolves it in that repository, else `HEAD`. Name the signal in the file's header comment (kan-1044 self-review, F1-c).
  - [ ] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, read `HEAD` only — the pin case fails; restore.
  - [ ] **Step 5: Repair (after Task 1's Step 5).** From the frontend worktree above, for each of `panel-fix-4 3b9a3985`, `panel-fix-7 ae761ec7`, `panel-fix-11 25394c9a`: `cd <wt>/stats && go run ./cmd/flow record dispatch backfill -C /Users/tweety53/Projects/gymie-frontend-worktrees/kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets -change kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets -key <key> -session-token mf-k1044-20261028T121921Z -role panel-fix -model opus -agent-id inline -commit <sha> -outcome completed`. <!-- unverified:commitBefore reads HEAD, so a pin-only commit's start falls to the base's last commit before it, then panelFixFloor raises it to round r-1's reviewer end --> A refusal is a defect of this task: fix it in `stats/cmd/flow/record.go` with a test, and record it in the commit body. Step 1's store read then shows the three rows.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'UnfinishedWork' -count=1` passes; `scripts/run-guard-tests.sh` passes. Commit.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`
**Allowed-collateral:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestUnfinishedWorkFlagsAPanelFixCommitWithNoRow`
**Regression:** `TestUnfinishedWorkFlagsAPanelFixCommitWithNoRow` fails — a panel fix with no dispatch row lands unflagged.
**Baseline:** before=13 after=14 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_unfinished_work_test.go @ 3aab64d6 -->
**Commit:** fix(guard): flag a panel-fix commit with no dispatch row
**After:** Task 1, 2

**Decision:** gate-panel-fix-trailer-row

- [ ] 4. Band measurement crops a scrim frame to its sheet
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `/Users/tweety53/Projects/gymie/spectre/changes/archive/kan-938-step-3-frontend-cardio-type-picker-on-add-a/visual-verification.md:59` records "the frame reads as one 1632px band (scrim plus sheet)".
  - [ ] **Step 2: Write the test** `TestMeasureVisualPropertiesBandsUnderScrim` in `stats/internal/guard/measure_visual_properties_test.go`, with `mvpPNG`: 400×800, scrim `#7f7f7f`, a white sheet at x 40–360, y 300–800, dark text rows at y 350–370 and 420–440; `--props bands` reports two bands at the text rows; a two-image run against the same sheet without scrim reports `delta.bands_summary.paired` 2. Run `cd stats && go test ./internal/guard -run 'TestMeasureVisualPropertiesBandsUnderScrim' -count=1` — it fails with one band.
  - [ ] **Step 3: Fix.** In the page branch of `mvpOpen` (`stats/internal/guard/measurevisualproperties.go`), add the scrim crop of decision `scrim-crop`, with its `ponytail:` comment; replace the kan-1015 scrim note in `scripts/measure-visual-properties.sh` with one line saying a flat scrim is cropped to its sheet (kan-1044 self-review, F1-e).
  - [ ] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, skip the crop — the test fails; restore.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'MeasureVisualProperties' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/guard/measurevisualproperties.go`, `stats/internal/guard/measure_visual_properties_test.go`, `scripts/measure-visual-properties.sh`
**Tests:** `TestMeasureVisualPropertiesBandsUnderScrim`
**Regression:** `TestMeasureVisualPropertiesBandsUnderScrim` fails — a modal-sheet frame reads as one band and gets no band or seam pairing.
**Baseline:** before=9 after=10 <!-- measured: grep -c -E '^func Test' stats/internal/guard/measure_visual_properties_test.go @ 3aab64d6 -->
**Commit:** fix(guard): crop a scrim frame to its sheet before measuring bands
**After:** none

**Decision:** scrim-crop

- [ ] 5. A running-clock departure closes on repeated running-clock passes
  - [ ] **Step 1: Edit.** In `skills/flow/verify-fix-loop.md`, after the paragraph **A task fixing a visual departure runs its capture spec before it closes**, add the paragraph of decision `running-clock-repro`, citing kan-1044 self-review A8-a: Tasks 7 and 8 closed green on the pinned-clock capture, and Task 9 on one running-clock pass against a four-in-five failure rate, three extra visual-verify rounds following.
  - [ ] **Step 2: Verify.** `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** none — a prose paragraph.
**Regression:** none — prose.
**Commit:** docs(flow): close a running-clock departure on repeated running-clock passes
**After:** none

**Decision:** running-clock-repro
