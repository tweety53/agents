# self-review-kan-1046-review-dynamics-dashboard

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1046's nine self-review findings — 1a (Task 1), 1b (Task 2), 1c (Task 3),
1d (Task 4), 1e (Task 5), 1f (Task 6), 4a (Task 7), 8c (Task 8), 8d (Task 9), as `proposal.md`
names them.

**Architecture:** a store guard on re-ending a closed row; two back-fill floor fixes and a
mutation-order refusal in `flow record`; a brainstorm-end gate in `flow stage`; the plan guard in
`close-task`; three prose sentences; a Chromium rasterisation pin with a one-time recapture; six
live store repairs through `flowd`.

**Tech Stack:** Go (`stats/`), Bash, Markdown, TypeScript/Playwright (`stats/web`).

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-28-self-review-kan-1046-review-dynamics-dashboard-design.md` in this
worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/`. Never merge or push.
- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1046-review-dynamics-dashboard` (`<wt>` below), branch `spectre/kan-1046-review-dynamics-dashboard`, committing after its existing commits and rewriting none. Every `**Files:**` path is relative to `<wt>`. Never touch the main checkout, and never create a worktree.
- Go: `gofmt -w` every Go file you change; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Mutation proofs restore with `break-and-prove.sh` or `mutate-and-verify.sh`, or a saved diff re-applied with `git apply` — never `git checkout --` on a file carrying uncommitted edits.
- Live store repairs (Tasks 1–5) run `cd <wt>/stats && go run ./cmd/flowd <subcommand> …`, which opens the dev `flow` database directly. Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database, and never run tests against the live database.
- Task 9 runs Playwright against the disposable UI-test stack on 4174 through the `.flow/project.md` `## visual verification` `capture` and `verify` rows, with a Bash `timeout` of 600000.

## Review Focus

- Task 1: a first `end` on an open row still stores its `-ended-at`; the end-before-start clamp still applies to a first end.
- Task 2: a verifier row ending after the commit is no floor; a `visual-verify` row of another change is never read.
- Task 3: a `full-suite-fix-<n>` key and a non-inline back-fill are unaffected; a failed record read refuses nothing.
- Task 4: round 0 and a round whose re-review has not begun pass; a failed record read refuses nothing.
- Task 5: only `/flow`'s `flow.brainstorm` end is gated; a non-`completed` outcome is gated too, since an abandoned brainstorm leaves the same open row.
- Task 9: no `maxDiffPixels`, threshold or retry anywhere in the diff.

- [ ] 1. A re-end of a closed dispatch row keeps its end
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tAc "select id, ended_at from dispatches where id in (8415, 8419)"` prints 14:54:33.63 and 14:54:33.75 — the re-point instant, after commits authored 14:37:26Z and 14:41:35Z.
  - [ ] **Step 2: Write the test** `TestEndDispatchKeepsClosedRowEnd` in `stats/internal/store/records_test.go`, on the existing `EndDispatch` fixtures: begin a row, end it at T1 with commit A, end it again at T2 > T1 with commit B and outcome `completed` — the row reads `ended_at` T1, commit B. Run `cd stats && go test ./internal/store -run 'TestEndDispatchKeepsClosedRowEnd' -count=1` — it fails.
  - [ ] **Step 3: Fix.** In `Store.EndDispatch` (`stats/internal/store/records.go`), set `ended_at = COALESCE(d.ended_at, $7)` in the `UPDATE`, and say in the doc comment that a closed row's end moves only through `flowd retime-dispatch` (kan-1046 self-review, 1a).
  - [ ] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, restore `ended_at = $7` — the test fails; restore.
  - [ ] **Step 5: Repair.** `cd stats && go run ./cmd/flowd retime-dispatch 8415 2026-10-28T14:26:41Z 2026-10-28T14:37:26Z 6a8a6faed865bf293d23fd9602e7f2ee3569a8c6` and `go run ./cmd/flowd retime-dispatch 8419 2026-10-28T14:37:39Z 2026-10-28T14:41:35Z 0acfc59094f51a01bd1b59232f6cee0a3d4eb9e4` (starts unchanged; Tasks 2 and 3 move them). Re-run Step 1's query: 14:37:26 and 14:41:35.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l internal/store` prints nothing; `cd stats && go test ./internal/store -run 'TestEndDispatch' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestEndDispatchKeepsClosedRowEnd`
**Regression:** `TestEndDispatchKeepsClosedRowEnd` fails — a re-end overwrites a closed row's `ended_at` with the re-end instant.
**Baseline:** before=110 after=111 <!-- measured: grep -c -E '^func Test' stats/internal/store/records_test.go @ 6c2ee105 -->
**Commit:** fix(stats): keep a closed dispatch row's end on a re-end
**After:** none

**Decision:** closed-row-keeps-end

- [ ] 2. An inline back-fill floors at the change's visual-verify verdicts
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tAc "select id, dispatch_key, role, started_at, ended_at from dispatches where id in (8409, 8415)"` — 8415 starts 14:26:41, before 8409 (`visual-verify`, role `verifier`) ended at 14:34:31.338.
  - [ ] **Step 2: Write the test** `TestRecordDispatchBackfillInlineFloorsAtVisualVerify` in `stats/cmd/flow/record_test.go`, on `inlineFloor`'s existing back-fill test fixtures (fake run-record server, scratch git repo): an inline `-commit` back-fill of `task-3-implementer` whose commit-before lies before a `visual-verify` verifier row's end — the derived `-started-at` is that end; a `visual-verify` row ending after the commit is no floor. Run `cd stats && go test ./cmd/flow -run 'TestRecordDispatchBackfillInlineFloorsAtVisualVerify' -count=1` — it fails.
  - [ ] **Step 3: Fix.** In `inlineFloor` (`stats/cmd/flow/record.go`), also take a row whose `Role == "verifier"` and whose key starts `visual-verify` as a floor candidate, under the same `!d.EndedAt.After(endT)` bound; update the doc comment (kan-1046 self-review, 1b).
  - [ ] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, drop the verifier condition — the test fails; restore.
  - [ ] **Step 5: Repair.** `cd stats && go run ./cmd/flowd retime-dispatch 8415 2026-10-28T14:34:31Z 2026-10-28T14:37:26Z 6a8a6faed865bf293d23fd9602e7f2ee3569a8c6`; Step 1's query shows 8415 starting 14:34:31.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestRecordDispatchBackfill' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBackfillInlineFloorsAtVisualVerify`
**Regression:** `TestRecordDispatchBackfillInlineFloorsAtVisualVerify` fails — an in-run visual fix's derived start predates the verdict it fixes.
**Baseline:** before=150 after=151 <!-- measured: grep -c -E '^func Test' stats/cmd/flow/record_test.go @ 6c2ee105 -->
**Commit:** fix(stats): floor an inline back-fill at the change's visual-verify verdicts
**After:** none

**Decision:** inline-floor-visual-verify

- [ ] 3. Refuse an inline panel-fix back-fill while its round's reviewers are open
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tAc "select id, dispatch_key, started_at, ended_at from dispatches where id between 8416 and 8419"` — 8419 `panel-fix-4` starts 14:37:39, before the `panel-3-*` reviewers began (14:38:10) and before the last ended (14:40:55).
  - [ ] **Step 2: Write the test** `TestRecordDispatchBackfillRefusesPanelFixOverOpenRound` in `stats/cmd/flow/record_test.go`, on `panelFixFloor`'s existing back-fill fixtures: an inline `-commit` back-fill of `panel-fix-2` while a `panel-1-primary` reviewer row has no end exits 2, its stderr carrying `close panel round 1's reviewer rows first`, and journals nothing; with that row ended it passes and floors at its end. Run `cd stats && go test ./cmd/flow -run 'TestRecordDispatchBackfillRefusesPanelFixOverOpenRound' -count=1` — it fails.
  - [ ] **Step 3: Fix.** In the inline back-fill path of `stats/cmd/flow/record.go` (where `panelFixFloor` is applied), before deriving the start: for role `panel-fix` and a key `panel-fix-<r>[-<chunk>]`, read the run record once and, when any reviewer row with `records.PanelDispatchRound == r-1` has `EndedAt == nil`, print `flow: -key %q: close panel round %d's reviewer rows first -- the fix cannot begin before the reports it fixes` and return 2. A failed read refuses nothing; `full-suite-fix-<n>` keys are untouched (kan-1046 self-review, 1c).
  - [ ] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, make the open-row condition false — the test fails; restore.
  - [ ] **Step 5: Repair.** `cd stats && go run ./cmd/flowd retime-dispatch 8419 2026-10-28T14:40:55Z 2026-10-28T14:41:35Z 0acfc59094f51a01bd1b59232f6cee0a3d4eb9e4`; Step 1's query shows 8419 starting 14:40:55.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestRecordDispatchBackfill' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBackfillRefusesPanelFixOverOpenRound`
**Regression:** `TestRecordDispatchBackfillRefusesPanelFixOverOpenRound` fails — a panel-fix back-fill over an open round derives a start before the round's reports.
**Baseline:** before=151 after=152 <!-- measured: grep -c -E '^func Test' stats/cmd/flow/record_test.go @ 6c2ee105 -->
**Commit:** fix(stats): refuse a panel-fix back-fill while its round's reviewers are open
**After:** Task 2

**Decision:** panel-fix-needs-closed-round

- [ ] 4. Refuse a mutation proof recorded after its re-review began
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tAc "select m.id, m.round, m.mutated from panel_mutations m join changes c on c.id = m.change_id where c.name = 'kan-1046-review-dynamics-dashboard' and m.round = 1 order by m.id"` prints 1076–1079; `grep -n 'DropMutation\|drop-mutation' -r stats` prints nothing.
  - [ ] **Step 2: Write the tests.** `TestRecordMutationRefusesAfterReReviewBegan` in `stats/cmd/flow/record_test.go`, on `TestRecordMutationCommand`'s fixtures with a fake run record: `-round 1` while a `panel-1-primary` reviewer row has begun exits 2 naming panel round 1 and records nothing; `-round 2` with no `panel-2-*` row passes. `TestDropMutation` in `stats/internal/store/records_test.go`, beside `DropPass`'s test: a recorded mutation is gone after `DropMutation`, and a second drop of the same id errors `no such mutation`. Run `cd stats && go test ./cmd/flow -run 'TestRecordMutationRefusesAfterReReviewBegan' -count=1` and `cd stats && go test ./internal/store -run 'TestDropMutation' -count=1` — both fail.
  - [ ] **Step 3: Fix.** In `runRecordMutation` (`stats/cmd/flow/record.go`), after the path check and before the store call: read the run record and, when any reviewer row with `records.PanelDispatchRound == *round` has `StartedAt` not after now, print `flow: record mutation: panel round %d's re-review began %s -- a mutation proof is recorded before the re-review it precedes (skills/flow/review-panel-fix-round.md)` and return 2; a failed read refuses nothing. Add `Store.DropMutation(ctx, id)` to `stats/internal/store/records.go` as `DropPass` is, and `flowd drop-mutation <mutation-id>` to `stats/cmd/flowd/main.go` as `drop-pass` is. In `skills/flow/review-panel-fix-round.md`, beside **Record each one**, add one sentence: each row is recorded before the round's re-review begins, and `flow record mutation` refuses one after (kan-1046 self-review, 1d).
  - [ ] **Step 4: Prove the tests pin the fix.** With `break-and-prove.sh`, make the round condition false — `TestRecordMutationRefusesAfterReReviewBegan` fails; make `DropMutation` delete nothing — `TestDropMutation` fails; restore each.
  - [ ] **Step 5: Repair.** `cd stats && go run ./cmd/flowd drop-mutation 1076`, then `1077`, then `1079` — the round-1 rows panel-fix-1 never ran; 1078 stays. Step 1's query prints 1078 alone.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l cmd internal/store` prints nothing; `cd stats && go test ./cmd/flow -run 'TestRecordMutation' -count=1` and `cd stats && go test ./internal/store -run 'TestDrop' -count=1` pass; `cd stats && go build ./cmd/flowd` exits 0; `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flowd/main.go`, `skills/flow/review-panel-fix-round.md`
**Tests:** `TestRecordMutationRefusesAfterReReviewBegan`, `TestDropMutation`
**Regression:** `TestRecordMutationRefusesAfterReReviewBegan` fails — a proof is accepted after the re-review that would contradict it; `TestDropMutation` fails — no command removes a mutation row recorded wrongly.
**Baseline:** before=152 after=153 <!-- measured: grep -c -E '^func Test' stats/cmd/flow/record_test.go @ 6c2ee105 --> (`record_test.go`); before=111 after=112 (`stats/internal/store/records_test.go`, after Task 1)
**Commit:** fix(stats): refuse a mutation proof recorded after its re-review began
**After:** Task 1, 3

**Decision:** mutation-before-re-review

- [ ] 5. Refuse a brainstorm end while a dispatch row of the change is open
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tAc "select id, dispatch_key, started_at, ended_at from dispatches where id between 8371 and 8373"` — 8371 `explore-1` ends 12:57:35, while its agent `a49c030616e139dc5` handed back at 12:36:38Z (the pass's evidence) and its siblings ended at 12:36.
  - [ ] **Step 2: Write the test** `TestStageEndBrainstormRefusesOpenDispatch` in `stats/cmd/flow/stage_test.go`, on the existing stage-end gate fixtures with a fake run-record server: `stage end -command /flow -stage flow.brainstorm -outcome completed` while an `explore-1` row has no end exits 1 naming `explore-1`, and sends no end mark; with the row ended it passes; `-stage flow.create-artifacts` over the same open row passes. Run `cd stats && go test ./cmd/flow -run 'TestStageEndBrainstormRefusesOpenDispatch' -count=1` — it fails.
  - [ ] **Step 3: Fix.** Add a `gateBrainstormDispatchesClosed(ctx, f, stderr)` beside `gateLiveVerification` in `stats/cmd/flow/stage.go`, called with the other gates in `runStageEnd`: for command `/flow` and stage `flow.brainstorm`, read the change's run record and, when any dispatch has `EndedAt == nil`, print `flow: stage end refused: flow.brainstorm ends with dispatch rows still open: <keys> -- end each with flow record dispatch end when its agent hands back` and return 1; a failed read refuses nothing. In **Explore and design agents** (`skills/flow/brainstorm-planner.md`) add one sentence: each planning dispatch's row is ended when its agent hands back, and `flow stage end` of `flow.brainstorm` refuses while one is open (kan-1046 self-review, 1e).
  - [ ] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, make the gate return 0 — the test fails; restore.
  - [ ] **Step 5: Repair.** `cd stats && go run ./cmd/flowd retime-dispatch 8371 2026-10-28T12:35:02.262Z 2026-10-28T12:36:38Z ""`; Step 1's query shows 8371 ending 12:36:38.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestStageEnd' -count=1` passes; `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`, `skills/flow/brainstorm-planner.md`
**Tests:** `TestStageEndBrainstormRefusesOpenDispatch`
**Regression:** `TestStageEndBrainstormRefusesOpenDispatch` fails — the brainstorm stage ends over an open planning row, which a later session closes with a guessed end.
**Baseline:** before=43 after=44 <!-- measured: grep -c -E '^func Test' stats/cmd/flow/stage_test.go @ 6c2ee105 -->
**Commit:** fix(stats): refuse a brainstorm end over an open dispatch row
**After:** none

**Decision:** brainstorm-end-needs-closed-rows

- [ ] 6. close-task runs the plan guard on an appended or in-run fix task
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'check-plan-shape' stats/internal/guard/closetask.go` prints nothing.
  - [ ] **Step 2: Write the test** `TestCloseTaskRunsPlanShapeOnInRunFix` in `stats/internal/guard/close_task_test.go`, on its existing fixtures: closing a task titled `In-run fix 1 — visual-verify: x` whose `**Files:**` omits `visual-verification.md` exits 1 with plan-shape's F21 line and ticks nothing; with the record added to `**Files:**` it passes; a planned (non-appended) task over the same plan never runs the guard. Run `cd stats && go test ./internal/guard -run 'TestCloseTaskRunsPlanShapeOnInRunFix' -count=1` — it fails.
  - [ ] **Step 3: Fix.** In `closeTask` (`stats/internal/guard/closetask.go`), when any task it closes is counted by the plan's `**Tasks appended:**` line (`uwAppended`) or titled `In-run fix <k> — `, run `Registry["check-plan-shape"]` on `<canonical>/spectre/changes/<name>/tasks.md` before anything is ticked and fold its exit into `worst`, refusing on 1. Add the step to `scripts/close-task.sh`'s header list. In **The loop** step 1 (`skills/flow/verify-fix-loop.md`), add that `<changeRoot>/visual-verification.md` is untracked until this task's commit, which `git add`s it (kan-1046 self-review, 1f).
  - [ ] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, skip the guard call — the test fails; restore.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCloseTask' -count=1` passes; `scripts/check-guard-symlinks.sh` and `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `scripts/close-task.sh`, `skills/flow/verify-fix-loop.md`
**Tests:** `TestCloseTaskRunsPlanShapeOnInRunFix`
**Regression:** `TestCloseTaskRunsPlanShapeOnInRunFix` fails — an appended task closes with a plan the shape guard refuses, found only at verify.
**Baseline:** before=11 after=12 <!-- measured: grep -c -E '^func Test' stats/internal/guard/close_task_test.go @ 6c2ee105 -->
**Commit:** fix(stats): run the plan guard when close-task closes an appended task
**After:** none

**Decision:** close-task-plan-shape

- [ ] 7. The live-check paragraph names a project-declared stand-in tool
  - [ ] **Step 1: Edit.** In **Write a live check** (`skills/flow/brainstorm-planner.md`), after the sentence naming the stand-ins, add: when `<project>/.flow/project.md` names a tool for the check's stand-in — this repository's `scripts/live-store-copy.sh` — the section invokes that tool by name and never re-specifies what it does (kan-1046 self-review, 4a).
  - [ ] **Step 2: Verify.** `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** none — a prose sentence.
**Regression:** none — prose.
**Commit:** docs(flow): name a project-declared live-check tool in the planner
**After:** Task 5

**Decision:** live-check-names-declared-tool

- [ ] 8. A prototype-specified capture measures before it commits
  - [ ] **Step 1: Edit.** Extend **Write a capture task that baselines against a mockup to measure before it commits** (`skills/flow/brainstorm-planner.md`): the same compose-and-measure steps apply when the view's specification is an approved prototype, rendered as each capture's frame per step 10 of `skills/flow/visual-verify-verifier.md`, with no declared `mockups` root (kan-1046 self-review, 8c).
  - [ ] **Step 2: Verify.** `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** none — a prose sentence.
**Regression:** none — prose.
**Commit:** docs(flow): measure a prototype-specified capture before it commits
**After:** Task 7

**Decision:** prototype-capture-measures

- [ ] 9. Pin Chromium's rasterisation for the stats SPA's visual baselines
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Write a scratch Playwright config outside the tree (plain `module.exports`, the committed config's `use` values, `testDir` the worktree's `stats/web/tests/visual`, `snapshotPathTemplate` and `outputDir` in a scratch dir per run) and run `npx playwright test -c <scratch config> full-app-suite.spec.ts --update-snapshots=all` three times from `stats/web` against the UI-test stack: the PNGs differ pixelwise between runs at (622,139)–(624,141) or (752,156)–(754,158). <!-- measured: three unpinned captures, run1 vs run2 differ at (622,139,624,141) on full-decisions/full-reviewers, (24,115,26,192) on full-run-detail; run1 vs run3 at (752,156,754,158) on full-state-board @ 6c2ee105, 2026-10-28, @playwright/test 1.62.1 -->
  - [ ] **Step 2: Fix.** In `stats/web/playwright.config.ts`, set `use.launchOptions: { args: ["--disable-gpu-rasterization", "--disable-partial-raster", "--disable-gpu"] }` and replace the closing comment's "no rendering pin is needed" with what the pin fixes and why it is not a tolerance (kan-1046 self-review, 8d). <!-- measured: four captures with these three args were byte-identical (cmp) @ 6c2ee105, 2026-10-28 -->
  - [ ] **Step 3: Prove the pin holds.** Re-run Step 1's three captures with the pin in the scratch config: byte-identical (`cmp`). Remove the pin: Step 1's differences return. Restore it.
  - [ ] **Step 4: Recapture.** The pinned render differs from today's baselines, so run the `capture` row of `.flow/project.md` `## visual verification` once per spec in `stats/web/tests/visual/` with `--update-snapshots=all` appended, read every PNG whose bytes changed per **The sweep** (`skills/flow-contracts/known-bugs.md`) — a change other than edge rasterisation is a defect to report, not to commit — and rebuild `full-app-suite.zip` with `scripts/rebuild-full-app-suite-zip.sh`.
  - [ ] **Step 5: Verify.** The `verify` row passes twice in a row; `cd stats/web && npx tsc -b` exits 0. Commit.

**Build:** green
**Files:** `stats/web/playwright.config.ts`, `stats/web/tests/visual/full-app-suite.zip`
**Allowed-collateral:** `stats/web/tests/visual/*-snapshots/*.png`
**Tests:** none — a determinism pin, proved by Step 3.
**Regression:** none — Step 3 shows the captures diverge again without the pin.
**Commit:** fix(stats): pin Chromium rasterisation for the visual baselines
**After:** none

**Decision:** pin-chromium-raster
