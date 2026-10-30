# kan-1061-flow-add-flow-interactive-fast-conversational

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** `/flow-interactive` — a conversational session that lands each request as it goes, every request recorded in flowd, a guard on every landed request's end, and the live board drawn as the approved prototype (KAN-1061).

**Architecture:** a flowd `requests` table and `stage_runs.request_n` behind store, API, client, journal replay and `flow record request`; `/flow-interactive` on eleven existing stage keys; the board API serving requests as the row's tasks; the SPA card and run view made interactive-aware; two hook changes; the skill, its command file and its registration.

**Tech Stack:** Go (`stats/`), PostgreSQL migrations, React/TypeScript (`stats/web`), Playwright, Python 3 hooks, Markdown.

**Spec:** `design.md` beside this file — its `## Design` section is the behaviour every task implements; the approved prototype `.superpowers/sdd/planner/prototype-board.html` is the board's specification.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational` (`<wt>`), branch `spectre/kan-1061-flow-add-flow-interactive-fast-conversational`; `**Files:**` paths are relative to `<wt>`. Never touch the main checkout, never create a worktree, never write `spectre/changes/`.
- Go: `gofmt -w` every Go file changed; `cd stats && gofmt -l . && make -s web-dist && go vet ./...` exits clean. SPA: `cd stats && make -s web-dist && cd web && npx tsc -b` exits 0.
- Store tests use the package's test-database helper (`newTestStore`, `stats/internal/store/testsupport_test.go`), never the live `flow` database. Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database.
- Every Go test run sets a Bash `timeout` of 600000.
- Mutation proofs restore with `scripts/break-and-prove.sh` — never `git checkout --` over uncommitted edits.
- A task that writes Markdown under `skills/` or `commands-claude/` reports every `scripts/check-verbatim-moves.sh` FAIL line its own new or reworded sentences raise, exactly as printed after `::`, as that task's verbatim-moves correction; the parent appends them to this change's `verbatim-moves.txt` before the task closes, and the guard then exits 0.

## Review Focus

- Task 4: the guard runs before the write and before the store is contacted; a store that cannot be reached skips only the stage-run read; a defect-finding UI check (`failed`) never satisfies a later end; `-commit`/`-base` usage is exact per outcome.
- Task 1: `received_at` survives a second `begin`; an omitted title or reason never blanks a stored one; request state is derived on read by one helper, never stored.
- Task 3: `-request` is refused off `/flow-interactive` at the CLI and the API alike, and survives a journalled replay.
- Tasks 7 and 8: every interactive state matches the prototype; a `/flow` or `/flow-fast` change renders exactly as before.
- Task 10: the begun-row check now reaches a `/flow-fast` session's dispatches.

- [x] 1. Store requests and request-tagged stage runs
  - [x] **Step 1: Write the tests.** `stats/internal/store/requests_test.go`, on `newTestStore`, seeding a change with `PutChange`: `TestRequestBeginUpserts` — `BeginRequest(ctx, p, c, records.RequestBegin{N: 1, Title: "a"})` returns the row with `ReceivedAt` set and `State` `queued`; a second begin with only `BrainstormReason: "r"` keeps title `a`, sets the reason and keeps `ReceivedAt`; the change's `updated_at` is not earlier than the begin instant; `N` ≤ 0 is `ErrInvalidRequest`; an unknown change is `ErrChangeNotFound`. `TestRequestEndRecordsOutcome` — `EndRequest` with `Outcome: "landed", CommitSHA, BaseSHA` sets `EndedAt` and both shas and `State` `landed`; `answered` and `dropped` with no shas succeed; `landed` missing either sha, `answered` carrying one, and an outcome outside the three are each `ErrInvalidRequest`; an end for a request never begun inserts it. `TestListRequestsDerivesState` — three requests: one ended (`State` = its outcome), one carried by a stage run begun with `RequestN` (`in progress`), one with neither (`queued`), returned in `N` order. In `stats/internal/store/stageruns_test.go`: `TestBeginStageStoresRequestN` — `BeginStage` with `RequestN: &n` stores it and `QueryStageRuns` returns it; without it the field is nil. Run `cd stats && go test ./internal/store -run 'TestRequest|TestListRequestsDerivesState|TestBeginStageStoresRequestN' -count=1` — fails to compile.
  - [x] **Step 2: Migration.** `stats/internal/store/migrations/0056_requests.sql`, header comment in the house style citing KAN-1061: `CREATE TABLE IF NOT EXISTS requests (id BIGSERIAL PRIMARY KEY, change_id BIGINT NOT NULL REFERENCES changes(id), n INT NOT NULL CHECK (n > 0), title TEXT NOT NULL DEFAULT '', brainstorm_reason TEXT NOT NULL DEFAULT '', received_at TIMESTAMPTZ NOT NULL DEFAULT now(), ended_at TIMESTAMPTZ, outcome TEXT CHECK (outcome IN ('landed','answered','dropped')), commit_sha TEXT, base_sha TEXT, UNIQUE (change_id, n), CHECK ((ended_at IS NULL) = (outcome IS NULL)))` and `ALTER TABLE stage_runs ADD COLUMN IF NOT EXISTS request_n INT`.
  - [x] **Step 3: Implement.** `stats/internal/records/types.go`: `RequestBegin{N int; Title, BrainstormReason string; ReceivedAt *time.Time}`, `RequestEnd{N int; Outcome, CommitSHA, BaseSHA string; EndedAt *time.Time}`, `Request{N int; Title, BrainstormReason string; ReceivedAt time.Time; EndedAt *time.Time; Outcome, CommitSHA, BaseSHA, State string}`, JSON tags camelCase, doc comments in the file's style. `stats/internal/store/requests.go`: `ErrInvalidRequest`; `BeginRequest` (one transaction: `INSERT … ON CONFLICT (change_id, n) DO UPDATE` keeping `received_at` and every field the call left empty, then `UPDATE changes SET updated_at = GREATEST(updated_at, <instant>)`); `EndRequest` (upsert setting `ended_at`, `outcome`, `commit_sha`, `base_sha`); `ListRequests(ctx, projectKey, change)`; a supplied instant is kept, else `now()`; `requestState(outcome string, started bool) string` — the one derivation of design.md's request state, used by every reader. `stats/internal/store/stageruns.go`: `BeginStageInput.RequestN *int` written by `insertStageRunAndSupersede`'s insert; `StageRun.RequestN *int` read by every select that fills a `StageRun`.
  - [x] **Step 4: Prove.** `scripts/break-and-prove.sh` with `BeginRequest`'s conflict update overwriting `received_at` — `TestRequestBeginUpserts` fails; restore.
  - [x] **Step 5: Verify.** Step 1's command passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/store/migrations/0056_requests.sql`, `stats/internal/store/requests.go`, `stats/internal/store/requests_test.go`, `stats/internal/store/stageruns.go`, `stats/internal/store/stageruns_test.go`, `stats/internal/records/types.go`
**Allowed-collateral:** `stats/internal/store/*_test.go`
**Tests:** `TestRequestBeginUpserts`, `TestRequestEndRecordsOutcome`, `TestListRequestsDerivesState`, `TestBeginStageStoresRequestN`
**Regression:** `TestListRequestsDerivesState` fails — a request's state no longer follows its outcome and its stage runs.
**Commit:** feat(store): store session requests and tag stage runs with their request
**After:** none

**Decision:** session-change-request-record

- [x] 2. Add `/flow-interactive` to the stage vocabulary
  - [x] **Step 1: Write the tests.** `stats/internal/stages/names_test.go`: `TestFlowInteractiveStageKeys` — `stages.ForCommand(stages.FlowInteractive)` is exactly `flow.kickoff`, `flow.brainstorm`, `flow.load-context`, `flow.sdd-tdd`, `flow.verify`, `flow.visual-verify`, `flow.self-review`, `flow.landing-routes`, `flow.cleanup`, `flow.write-finished`, `flow.refresh-main-checkout` in Table order, and `stages.Valid(stages.FlowInteractive, "flow.decide")` is false. `stats/internal/api/stages_test.go`: `TestApplyBeginStageMarkFinishesAFlowInteractiveChange`, on `TestApplyBeginStageMarkFinishesAFlowFastChange`'s shape — a `/flow-interactive` `flow.write-finished` begin moves the bootstrapped change to `FINISHED`. `stats/cmd/flow/stage_test.go`: `TestStageEndRefusesOpenSelfReviewPassInteractive`, on `TestStageEndRefusesOpenSelfReviewPass`'s shape with `-command /flow-interactive` — an open pass row exits 1, an ended one 0. Run `cd stats && go test ./internal/stages ./internal/api ./cmd/flow -run 'TestFlowInteractiveStageKeys|TestApplyBeginStageMarkFinishesAFlowInteractiveChange|TestStageEndRefusesOpenSelfReviewPassInteractive' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/internal/stages/names.go`: `FlowInteractive Command = "/flow-interactive"`, added to the `Commands` of the eleven rows Step 1 names. `README.md`'s **Level 1 — the stages of each command** table: `` `/flow-interactive` `` appended to the Commands cell of the same eleven rows. `stats/internal/api/stages.go` `ApplyBeginStageMark`: the `flow.write-finished` → `FINISHED` rule fires for `stages.FlowInteractive` as for `stages.FlowFast`, its comment naming both. `stats/cmd/flow/stage.go` `gateBrainstormDispatchesClosed`: the `selfReview` condition admits `/flow-interactive`, its doc comment naming it.
  - [x] **Step 3: Verify.** Step 1's command and `cd stats && go test ./internal/stages -run 'TestStagesMatchReadmeLevelOne' -count=1` pass; `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0; lint. Commit.

**Build:** green
**Files:** `stats/internal/stages/names.go`, `stats/internal/stages/names_test.go`, `README.md`, `stats/internal/api/stages.go`, `stats/internal/api/stages_test.go`, `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`
**Allowed-collateral:** `stats/internal/stages/*_test.go`
**Tests:** `TestFlowInteractiveStageKeys`, `TestApplyBeginStageMarkFinishesAFlowInteractiveChange`, `TestStageEndRefusesOpenSelfReviewPassInteractive`
**Regression:** `TestFlowInteractiveStageKeys` fails — every `/flow-interactive` mark is refused as an undocumented stage.
**Commit:** feat(stages): add /flow-interactive to the stage vocabulary
**After:** none

**Decision:** reuse-flow-stage-keys
**Decision:** self-review-gate-interactive

- [x] 3. Serve requests and request-tagged stage runs through the API, client and journal
  - [x] **Step 1: Write the tests.** `stats/internal/api/requests_test.go` on `records_test.go`'s fake-store shape: `TestRequestRoutes` — `POST /api/v1/records/{project}/{change}/requests` 200 with the row, `POST …/requests/end` 200, `GET …/requests` the list; `ErrInvalidRequest` → 400, `ErrChangeNotFound` → 404. `stats/internal/api/stages_test.go`: `TestApplyBeginStageMarkRefusesRequestOffInteractive` — a begin mark carrying `Request` with `/flow` or `/flow-fast` is `ErrInvalidMark`, a non-positive one too, and a `/flow-interactive` one reaches `BeginStage` with `RequestN` set. `stats/internal/api/stats_test.go`: `TestListStageRunsServesRequest` — a stage run with `RequestN` serves `"request"`, one without omits it. `stats/internal/reconcile/record_test.go`: `TestReplayRequestEntries` — journalled `request-begin` and `request-end` bodies replay through `ApplyRequestBegin`/`ApplyRequestEnd` with their instants. `stats/internal/reconcile/stage_test.go`: `TestReplayBeginMarkKeepsRequest` — a journalled begin carrying `request` replays with it. `stats/cmd/flow/stage_test.go`: `TestStageBeginRequestFlag` — `-request 3` with `-command /flow-interactive` sends `"request":3`; with `-command /flow` exits 2 naming the flag and sends nothing; `-request -1` exits 2. Run `cd stats && go test ./internal/api ./internal/reconcile ./cmd/flow -run 'TestRequestRoutes|TestApplyBeginStageMarkRefusesRequestOffInteractive|TestListStageRunsServesRequest|TestReplayRequestEntries|TestReplayBeginMarkKeepsRequest|TestStageBeginRequestFlag' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/internal/api/records.go`: `RecordStore` gains `BeginRequest`, `EndRequest`, `ListRequests`. `stats/internal/api/requests.go`: `ApplyRequestBegin`, `ApplyRequestEnd` (the store call; `ErrInvalidRequest` → `ErrInvalidRecord`) and the three handlers on `recordHandler`, routed in `stats/internal/api/server.go` beside `task-counts`. `stats/internal/api/stages.go`: `stageBeginRequest` and `BeginStageMark` gain `Request *int` (`json:"request,omitempty"`); `ApplyBeginStageMark` refuses a request off `/flow-interactive` or ≤ 0 with `ErrInvalidMark` and passes it as `BeginStageInput.RequestN`. `stats/internal/api/stats.go`: `stageRunDTO.Request *int` (`json:"request,omitempty"`). `stats/internal/client/client.go`: `BeginRequest`, `EndRequest`, `ListRequests`; `BeginStageRequest.Request *int`; `StageRunRow` gains `Request *int` and `StartedAt string`. `stats/internal/reconcile/reconcile.go`: the record replay switch gains `request-begin` and `request-end`, its `recordJournalBody` comment naming both; the begin-mark replay passes `Request`. `stats/cmd/flow/stage.go` `runStageBegin`: `-request <n>`, a positive integer, absent by default, refused (exit 2) with any `-command` but `/flow-interactive`, sent as `BeginStageRequest.Request`; `stageUsage` names it. Update every fake implementing `RecordStore`.
  - [x] **Step 3: Verify.** Step 1's command passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/api/records.go`, `stats/internal/api/requests.go`, `stats/internal/api/requests_test.go`, `stats/internal/api/server.go`, `stats/internal/api/stages.go`, `stats/internal/api/stages_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/client/client.go`, `stats/internal/reconcile/reconcile.go`, `stats/internal/reconcile/record_test.go`, `stats/internal/reconcile/stage_test.go`, `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`
**Allowed-collateral:** `stats/internal/api/*_test.go`, `stats/internal/reconcile/*_test.go`, `stats/cmd/flowd/*_test.go`, `stats/internal/client/client_test.go`, `stats/internal/web/embed_test.go`
**Tests:** `TestRequestRoutes`, `TestApplyBeginStageMarkRefusesRequestOffInteractive`, `TestListStageRunsServesRequest`, `TestReplayRequestEntries`, `TestReplayBeginMarkKeepsRequest`, `TestStageBeginRequestFlag`
**Regression:** `TestReplayRequestEntries` fails — a request written while the store was down is never recorded.
**Commit:** feat(stats): serve requests and request-tagged stage runs through the API and journal
**After:** Task 1, 2

**Decision:** session-change-request-record

- [x] 4. Record requests with `flow record request` and guard a landed end
  - [x] **Step 1: Write the tests.** `stats/cmd/flow/request_test.go`, on `stage_test.go`'s `gitRepo`, `isolatedStateRoot` and `genuineDaemon` helpers and `TestStageEndVisualVerifyRefusesAnUntaskedCommit`'s bare-origin fixture — a peer checkout carrying `.superpowers/sdd/<name>/`, branch `<name>`, `origin/HEAD` set, a `.flow/project.md` whose `## visual verification` declares `ui paths` `web/**`. `TestRecordRequestBeginEnd` — `begin -change c -n 1 -title t` posts `{n, title, receivedAt}` and prints `recorded: request 1`; `begin` with neither `-title` nor `-brainstorm-reason`, `-n 0`, and a missing `-change` each exit 2; `end -outcome answered` posts with no shas; `end -outcome answered -commit x` and `-outcome merged` exit 2; `end -outcome landed` without `-base` exits 2; with the daemon down `begin` exits 0, prints the journal warning and appends a `request-begin` body to the record journal. `TestRecordRequestEndLandedGuard` — subcases, each asserting exit code and that no write reached the daemon on a refusal: a commit not on `origin/<base>` → 1; `-base` not an ancestor of `-commit` → 1; a pushed commit touching `docs/a.md` → 0 and no stage-run read; a pushed commit touching `web/x.ts` with no `flow.visual-verify` run carrying the request → 1 naming the UI check; its latest such run ended `failed` while an earlier one ended `completed` → 1; its latest ended `completed` → 0; a run carrying another request only → 1; the daemon down → the git checks still refuse a commit not on the base, and a valid UI-touching end prints one `⚠` line naming the skipped read and journals (exit 0); `.flow/project.md` without the section → 0; `ui paths` empty → 1 quoting `VISUAL-TRIGGER-CANNOT-ANSWER`'s cause; no checkout carrying `.superpowers/sdd/<name>/` → 1. Run `cd stats && go test ./cmd/flow -run 'TestRecordRequest' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/cmd/flow/request.go`: `runRecordRequest` routing `begin|end`, each on `registerRecordConnFlags` with `-change`, `-n`, and `-title`/`-brainstorm-reason` (`begin`) or `-outcome`/`-commit`/`-base` (`end`, the shas through `validateCommitSHA`); the write through `callRecord` and `classifyRecordWrite` with kinds `request-begin`/`request-end`, the CLI stamping `ReceivedAt`/`EndedAt` to `time.Now()`. `gateLandedRequest(ctx, f, n, commit, base, stderr) int` runs design.md's **Request-end guard** checks 1–4 before the write, in-process: `changeCheckouts`, `guard.Registry["resolve-base-branch"]`, `git merge-base --is-ancestor` twice, `git diff --name-only <base>..<commit>` fed as `guard.Env.Stdin` to `guard.Registry["check-visual-trigger"]` with the worktree as root, and on its exit 0 `client.ListStageRuns(project, change, "/flow-interactive", "flow.visual-verify")` filtered to the request, the latest by `StartedAt` required `completed`; `client.ErrUnavailable` there is the one `⚠` line and no refusal. `stats/cmd/flow/record.go`: `runRecord` gains `case "request":`, `recordUsage` documents both verbs and their exit codes, and the `recordJournalBody` comment names the two kinds.
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` with the latest-run selection taking any `completed` run instead of the latest — the `failed`-latest subcase of `TestRecordRequestEndLandedGuard` fails; restore.
  - [x] **Step 4: Verify.** Step 1's command passes; lint. Commit.

**Build:** green
**Files:** `stats/cmd/flow/request.go`, `stats/cmd/flow/request_test.go`, `stats/cmd/flow/record.go`
**Allowed-collateral:** `stats/cmd/flow/*_test.go`
**Tests:** `TestRecordRequestBeginEnd`, `TestRecordRequestEndLandedGuard`
**Regression:** `TestRecordRequestEndLandedGuard` fails — a UI request lands with no passing live-page check, or a commit never pushed is recorded as landed.
**Commit:** feat(stats): record requests with flow record request and guard a landed end
**After:** Task 3

**Decision:** request-end-guard
**Decision:** request-base-first-landing
**Decision:** request-verb-own-file

- [x] 5. Serve an interactive session's command and requests on the state board
  - [x] **Step 1: Write the tests.** `stats/internal/store/aggregate_test.go`: `TestLiveStateBoardCarriesCommandAndRequests` — a change whose latest stage run is `/flow-interactive` reads `Command` `/flow-interactive` and its requests in `N` order with derived `State`; a `/flow` change reads `/flow` and no requests. `stats/internal/api/stats_test.go`: `TestStateBoardServesRequestsAsTasks` — a row with requests serves `command` and `tasks` from them: an ended request `state` `done` with its `outcome`, a landed one's `commit` cut to seven characters, `reason` when set, `in progress` and `pending` for the other two; a row without requests keeps the `tasks.md` reading. Run `cd stats && go test ./internal/store ./internal/api -run 'TestLiveStateBoardCarriesCommandAndRequests|TestStateBoardServesRequestsAsTasks' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/internal/store/aggregate.go`: `LiveStateRow` gains `Command *string` (the `last_run` lateral selects `sr.command`) and `Requests []records.Request`, read by a `jsonb_agg` lateral over the change's `requests` — on the subagents lateral's pattern — carrying whether any stage run has its `request_n`, its `State` set through `requestState`. `stats/internal/api/stats.go`: `stateBoardRowDTO.Command string` (`json:"command,omitempty"`); `taskDTO` gains `Outcome`, `Commit`, `Reason` (each `omitempty`); `toStateBoardDTOs` serves `tasks` from `r.Requests` when it has any, per design.md's **Store, CLI and API** board bullet.
  - [x] **Step 3: Verify.** Step 1's command passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/plantasks_test.go`
**Tests:** `TestLiveStateBoardCarriesCommandAndRequests`, `TestStateBoardServesRequestsAsTasks`
**Regression:** `TestStateBoardServesRequestsAsTasks` fails — an interactive session's card counts no requests.
**Commit:** feat(stats): serve an interactive session's command and requests on the state board
**After:** Task 1, 3

**Decision:** requests-served-as-tasks

- [x] 6. Render a session's requests in the self-review bundle
  - [x] **Step 1: Write the tests.** `stats/internal/selfreview/bundle_test.go`: `TestBundleRendersRequests` — on a git fixture with two commits pushed past a base, `Bundle` given requests (landed 1 with its `BaseSHA`/`CommitSHA`, answered 2) renders under `## git log --stat` `### Request 1: <title> — landed` followed by that range's `git log --stat` and `### Request 2: <title> — answered` with no log, in `N` order, and never calls `changeBranchLog`'s range; with no requests the section is unchanged. `stats/internal/api/selfreview_test.go`: `TestSelfReviewBundleReadsRequests` — the handler passes the store's `ListRequests` to the bundle; a store error on that read leaves the requests empty. Run `cd stats && go test ./internal/selfreview ./internal/api -run 'TestBundleRendersRequests|TestSelfReviewBundleReadsRequests' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/internal/selfreview/git.go`: `Bundle` takes `requests []records.Request`; with any, the git-log source is `requestsLog(g, mainRepo, requests)` — design.md's **Explicit stop** step 2 rendering, each range through `dropPNGStatLines(trimmedOutput(…))` — in place of `changeBranchLog`. `stats/internal/api/selfreview.go`: `selfreviewStore` gains `ListRequests`; the bundle handler reads it and passes it on.
  - [x] **Step 3: Verify.** Step 1's command passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/selfreview/git.go`, `stats/internal/selfreview/bundle_test.go`, `stats/internal/api/selfreview.go`, `stats/internal/api/selfreview_test.go`
**Allowed-collateral:** `stats/internal/api/*_test.go`
**Tests:** `TestBundleRendersRequests`, `TestSelfReviewBundleReadsRequests`
**Regression:** `TestBundleRendersRequests` fails — the stop's self-review pass reads an empty git log and judges nothing the session shipped.
**Commit:** feat(selfreview): render a session's requests in the self-review bundle
**After:** Task 1, 3

**Decision:** self-review-at-stop

- [x] 7. Draw an interactive session's card as the prototype
  - [x] **Step 1: Write the tests.** `stats/web/src/lib/interactive.test.ts`, `describe("interactive helpers")`: `isInteractive(row)` true only for `command` `/flow-interactive`; `currentRequest(row)` the `in progress` task; `lastEndedRequest(row)` the highest-numbered `done` task; `interactiveWaiting(row)` true when no task is `in progress` and no stage is open. `stats/web/src/views/views.test.tsx`, `it("interactive session card …")` on the file's state-board fetch-mock shape, with a running row (an open `flow.sdd-tdd`, eight requests as the prototype lists them, request 7 in progress) and a waiting row (three landed, no open stage, activity idle): pill text `Interactive`; running line `Request 7: Show the stage's model on the Agents row`; meta `requests 6/8 · $4.12`; the waiting card reads `Waiting for the next request · 6m`, `Last: Request 3 landed — Hide the empty streak row`, `requests 3/3`, carries no `needs` class, no activity badge and no `agent-count`; a `/flow` row still reads `tasks d/n` and its phase pill. Run `cd stats/web && npx vitest run src/lib/interactive.test.ts src/views/views.test.tsx -t 'interactive'` — fails.
  - [x] **Step 2: Implement.** `stats/web/src/api.ts`: `StateBoardRow.command?: string`; `PlanTask` gains `outcome?`, `commit?`, `reason?`; `StageRunDTO.request?: number`. `stats/web/src/lib/interactive.ts`: the four helpers. `stats/web/src/views/StateBoard.tsx` `RunningCard`, `StatePill` and `needsOperator`: the card of design.md's **Board** section, every glyph, text, colour, spacing and the hover and focus border reproduced from `.superpowers/sdd/planner/prototype-board.html` with `theme.css` tokens — the ring glyph and the muted last line added to `stats/web/src/styles.css`.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && cd web && npx tsc -b && npx vitest run src/lib/interactive.test.ts src/views/views.test.tsx -t 'interactive|state board'` passes. Commit.

**Build:** green
**Files:** `stats/web/src/api.ts`, `stats/web/src/lib/interactive.ts`, `stats/web/src/lib/interactive.test.ts`, `stats/web/src/views/StateBoard.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/styles.css`
**Allowed-collateral:** `stats/web/src/lib/runningTask.ts`, `stats/web/src/lib/runningTask.test.ts`, `stats/web/src/lib/taskCount.ts`
**Tests:** `interactive helpers`, `interactive session card`
**Regression:** the `interactive session card` case fails — a session reads `Paused after …` with a needs-you border between requests.
**Commit:** feat(web): draw an interactive session's card as the prototype
**After:** Task 5

**Decision:** board-interactive-aware
**Decision:** interactive-display-states

- [x] 8. Draw an interactive session's run view as the prototype
  - [x] **Step 1: Write the tests.** `stats/web/src/views/RunDetail.test.tsx`, `describe("interactive run view")` on the file's mock setup with an interactive board row and stage runs carrying `request` and `command` `/flow-interactive`: the pill reads `Interactive`; a `Requests` region (no `Tasks` region) with note `6/8` lists each request with its mark, its short commit, `answered`, `dropped` (struck through), `queued`, the current one `aria-current="step"`, and `brainstormed — new store fields and two panels` under request 6; the `Main` region's note reads `Request 7 · 4m` and its steps are `Loading context` done, `Implementing tasks` current, `Running lint and tests`, `Landing the change`, `Updating the main checkout`, `Checking the UI` todo; `Designing the change` appears only when request 7 has a `flow.brainstorm` run. `stats/web/src/components/StageRunTable.test.tsx`, `it("request column …")`: with the interactive switch on, the first header is `Request`, each row's first cell reads `R<n>` (blank without a request), rows run newest first and no `lb-group` row renders; off, the table is unchanged. Run `cd stats/web && npx vitest run src/views/RunDetail.test.tsx src/components/StageRunTable.test.tsx -t 'interactive run view|request column'` — fails.
  - [x] **Step 2: Implement.** `stats/web/src/views/RunDetail.tsx`: for an interactive change, the Requests panel in the Tasks slot, the Main panel's note and steps, and the timeline's upcoming lanes cut to the per-request stages, per design.md's **Board** run-view bullet. `stats/web/src/components/StageRunTable.tsx`: an optional `requests` prop switching on the `Request` column and the flat newest-first rows, every `colSpan` following the column count. Styles from the prototype into `stats/web/src/styles.css`; helpers into `stats/web/src/lib/interactive.ts`.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && cd web && npx tsc -b && npx vitest run src/views/RunDetail.test.tsx src/components/StageRunTable.test.tsx -t 'interactive run view|request column|run detail'` passes. Commit.

**Build:** green
**Files:** `stats/web/src/views/RunDetail.tsx`, `stats/web/src/views/RunDetail.test.tsx`, `stats/web/src/components/StageRunTable.tsx`, `stats/web/src/components/StageRunTable.test.tsx`, `stats/web/src/styles.css`, `stats/web/src/lib/interactive.ts`
**Allowed-collateral:** `stats/web/src/lib/interactive.test.ts`, `stats/web/src/lib/mainNow.ts`, `stats/web/src/lib/mainNow.test.ts`
**Tests:** `interactive run view`, `request column`
**Regression:** the `interactive run view` case fails — the run view lists requests as plan tasks under `/flow`'s phase steps.
**Commit:** feat(web): draw an interactive session's run view as the prototype
**After:** Task 7

**Decision:** board-interactive-aware
**Decision:** interactive-display-states

- [x] 9. Capture the interactive board and run view against the prototype
  - [x] **Step 1: Seed.** `stats/cmd/uitest-seed/seed.go` gains `seedInteractive`, called from `seedFixture`: in `uitest-alpha`, inside the pinned board period, the prototype's two sessions under its change names — `kan-1072-interactive-session-2026-10-30`, running: the prototype's eight requests with their titles, outcomes, commits and request 6's brainstorm reason, stage runs carrying their requests (`command` `/flow-interactive`) so the run view's stage-run count, the visible `R7`/`R6` rows, their times, durations, tokens, model and effort and costs, the cost-by-stage bars and the totals match the prototype, an open `flow.sdd-tdd` run for request 7 begun three minutes before `fixtureNow` with activity `working` and detail `Edit: stats/web/src/lib/phase.ts`; `kan-1073-interactive-session-2026-10-30`, waiting: three landed requests, the last `Hide the empty streak row`, its last stage run ended six minutes before `fixtureNow`. `stats/cmd/uitest-seed/seed_test.go`: `TestSeedInteractiveSessions` — on the throwaway database, both sessions list their requests with the seeded states and the running session's open run carries request 7. Run `cd stats && go test ./cmd/uitest-seed -run 'TestSeedInteractiveSessions|TestSeedPopulatesFixture' -count=1` — passes.
  - [x] **Step 2: Capture.** `stats/web/tests/visual/interactive.spec.ts` on `experiments.spec.ts`'s shape, the period pinned by `PINNED_QUERY` and the clock by `FIXTURE_NOW`, every locator by role, label or test id: the running card, the waiting card, the running card hovered, the running card focused from the keyboard, and the running session's run view (`#/run/uitest-alpha/kan-1072-interactive-session-2026-10-30`), each `toHaveScreenshot`. Its `interactive.spec.ts.mockups` sidecar pairs each capture with a frame id of `.superpowers/sdd/planner/prototype-board.html`, rendered at the capture's viewport, dark, in the same state into `.superpowers/sdd/interactive-frames/`. Run the project's `capture` command for this spec, then `scripts/compose-mockup-frames.sh` on the sidecar and `scripts/measure-visual-properties.sh` band, seam and matrix reads on every pair; fix the SPA or the seed until clean, or name each leftover with its cause in the hand-back — the fixture's project labels in place of the prototype's included.
  - [x] **Step 3: Recapture what the seed moved.** Run the project's `verify` command; every capture the seed moved — the board in `baseline.spec.ts` and `full-app-suite.spec.ts`, and any aggregate view whose figures now count the seeded runs — is recaptured with `capture` and each changed PNG read before it is kept; then `scripts/rebuild-full-app-suite-zip.sh <wt> <wt>/stats/web/tests/visual`.
  - [x] **Step 4: Verify.** The project's `verify` command exits 0; `cd stats && gofmt -l . && make -s web-dist && go vet ./...` clean. Commit.

**Build:** green
**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/cmd/uitest-seed/seed_test.go`, `stats/web/tests/visual/interactive.spec.ts`, `stats/web/tests/visual/interactive.spec.ts.mockups`
**Allowed-collateral:** `stats/web/tests/visual/*-snapshots/**`, `stats/web/tests/visual/full-app-suite.zip`, `stats/web/src/views/*.tsx`, `stats/web/src/components/*.tsx`, `stats/web/src/styles.css`, `stats/web/src/lib/interactive.ts`, `stats/web/tests/visual/controls.spec.ts`, `stats/web/tests/visual/views.spec.ts`, `stats/web/tests/visual/weekly-costs.spec.ts`
**Tests:** `interactive.spec.ts`, `TestSeedInteractiveSessions`
**Regression:** `interactive.spec.ts` fails — the interactive card or run view drifted from the approved prototype.
**Commit:** test(web): capture the interactive board and run view against the prototype
**After:** Task 1, 3, 8

**Decision:** seed-interactive-in-fixture

Correction (2026-10-30): the plan expected only captures to move; the new seed also moved fixture rows that `controls.spec.ts`, `views.spec.ts` and `weekly-costs.spec.ts` assert (leaderboard top ten, trend cost order, weekly-cost counts), so those specs were updated and about thirty existing snapshot PNGs plus `full-app-suite.zip` re-baselined, each diff read as fixture data.

- [x] 10. Read the session's latest stage run of any command in the agent-baseline hook
  - [x] **Step 1: Write the tests.** `stats/internal/repotest/enforce_agent_baseline_hook_test.go`: `TestEnforceAgentBaselineHookReadsAnyCommand` — a fixture server answering the session's row (`command` `/flow-fast`) only to a query carrying no `command` parameter, a transcript whose user turn typed `/flow-fast`, no open ledger row: a row-described dispatch is denied; on today's hook it is allowed. `TestEnforceAgentBaselineHookSeesFlowInteractive` — the same with a typed `/flow-interactive` and a `/flow-interactive` row: denied. Run `cd stats && go test ./internal/repotest -run 'TestEnforceAgentBaselineHook' -count=1` — the two new tests fail.
  - [x] **Step 2: Implement.** `hooks/enforce-agent-baseline.py`: `FLOW_COMMAND` admits `-interactive`; `session_change` drops the `command` parameter, so it reads the session's latest stage run of any command.
  - [x] **Step 3: Verify.** Step 1's command passes; `scripts/check-python-suppressions.sh` exits 0. Commit.

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`
**Tests:** `TestEnforceAgentBaselineHookReadsAnyCommand`, `TestEnforceAgentBaselineHookSeesFlowInteractive`
**Regression:** `TestEnforceAgentBaselineHookReadsAnyCommand` fails — a `/flow-fast` or `/flow-interactive` session's dispatches skip the begun-row check.
**Commit:** fix(hooks): read the session's latest stage run of any command in the agent-baseline hook
**After:** none

**Decision:** session-change-any-command

Correction (2026-10-30): the plan named only the two new tests; commit 9ad3fb14 also sets `FLOW_ADDR=http://127.0.0.1:1` in the existing table test's `runHook`, because case 13o read the live dev store, whose open allowed-pairs experiment refused its `flow-planner-medium` dispatch — reproduced failing with the line reverted, passing with it.

- [x] 11. Name an unfinished `/flow-interactive` session on every plain prompt
  - [x] **Step 1: Write the tests.** `stats/internal/repotest/flow_active_change_hook_test.go`: the subtest `case 1: command=/flow filter in the query` becomes `case 1: no command filter in the query`, asserting the query carries no `command`. `TestFlowActiveChangeHookInteractive` — a latest row `command` `/flow-interactive`, stage `flow.sdd-tdd`, injects design.md's **Hooks** interactive line naming the change; stage `flow.write-finished` injects nothing; `command` `/flow-fast` injects nothing; `/flow` keeps today's line. Run `cd stats && go test ./internal/repotest -run 'TestFlowActiveChangeHook' -count=1` — fails.
  - [x] **Step 2: Implement.** `hooks/flow-active-change.py`: `last_flow_change` becomes one GET of the session's latest stage run of any command returning its change, command and stage; `main` branches on the command as design.md's **Hooks** bullet states; the module docstring follows.
  - [x] **Step 3: Verify.** Step 1's command passes; `scripts/check-python-suppressions.sh` exits 0. Commit.

**Build:** green
**Files:** `hooks/flow-active-change.py`, `stats/internal/repotest/flow_active_change_hook_test.go`
**Tests:** `TestFlowActiveChangeHookInteractive`
**Regression:** `TestFlowActiveChangeHookInteractive` fails — after a compaction a plain message is no longer pointed at the request protocol.
**Commit:** feat(hooks): name an unfinished /flow-interactive session on every plain prompt
**After:** none

**Decision:** prompt-names-interactive-session

- [x] 12. Add the `/flow-interactive` skill and command
  - [x] **Step 1: Write the pin test.** `stats/internal/stages/names_test.go`: `extractFlowFastSkillKeys` takes the skill path; `TestStageKeysMatchFlowInteractiveSkillTable` pins `stages.FlowInteractive`'s keys to `skills/flow-interactive/SKILL.md`'s `## Stage keys` table both ways, as `TestStageKeysMatchFlowFastSkillTable` does for `/flow-fast`. Run `cd stats && go test ./internal/stages -run 'TestStageKeysMatch' -count=1` — fails: no such file.
  - [x] **Step 2: Write the skill.** `skills/flow-interactive/SKILL.md`, frontmatter `name: flow-interactive`, a `description` ending "Use for /flow-interactive.", `allowed-tools: Bash(flow:*), Bash(land-self-review-report.sh:*)`, `license: MIT`. Its body states design.md's **Shape**, **Session start**, **One request**, **Request-end guard** (what `flow record request end` refuses and its exit codes), **Explicit stop** and **Resume** as the session's instructions, citing `/flow-fast`'s sections it reuses (`skills/flow-fast/SKILL.md`) rather than copying them; an announce line; a guardrails paragraph (no subagent but the self-review pass's and the in-run fix loop's, no `spectre/` write save those passes' own changes, no push but `<name>` and `<name>:<base>`); a `## Stage keys` table in `/flow-fast`'s `| Step | Keys |` shape naming exactly the eleven keys; every mark a literal call — `flow stage mark -command '/flow-interactive' -stages flow.kickoff -harness <harness> -session-token fi-<literal-token> <name>`, each per-request stage a `flow stage begin … -request <n> -harness <harness> -session-token fi-<literal-token> <name>` and `flow stage end … -outcome completed|failed <name>` pair, the stop's `flow.self-review` and `flow.cleanup` pairs with no `-request`, `flow.write-finished` by `flow stage mark`; and every `flow record request` call. The brainstorm step's **AskUserQuestion** cites **Unanswered mid-run asks** (`skills/flow-contracts/pipeline.md`); per-request reports carry neither `**Learning:**` nor `**Prompt tuning:**`, the stop report both.
  - [x] **Step 3: Write the command and the symlinks.** `commands-claude/flow-interactive.md` on `commands-claude/flow-fast.md`'s shape: `model: opus`, a one-line description, a body naming the skill, the input (empty, a first request, a Jira key or an open session's name; no flags) and what the operator does next. `skills/flow-interactive/scripts/` symlinks onto `../../../scripts/<name>` for every guard the skill invokes — the set `scripts/check-guard-symlinks.sh` requires, `lib` included; it exits 0.
  - [x] **Step 4: Verify.** Step 1's command passes; `scripts/run-verify-list.sh <wt> lint` exits 0 save the `scripts/check-verbatim-moves.sh` FAIL lines this task's new sentences raise, reported per **Global constraints**; `SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global` installs `flow-interactive`. Commit.

**Build:** green
**Files:** `skills/flow-interactive/SKILL.md`, `commands-claude/flow-interactive.md`, `stats/internal/stages/names_test.go`, `skills/flow-interactive/scripts/lib`, `skills/flow-interactive/scripts/project-get.sh`, `skills/flow-interactive/scripts/resolve-base-branch.sh`, `skills/flow-interactive/scripts/check-fast-route-record.sh`, `skills/flow-interactive/scripts/check-visual-trigger.sh`, `skills/flow-interactive/scripts/land-self-review-report.sh`
**Allowed-collateral:** `skills/flow-interactive/scripts/*`, `stats/internal/guard/guardsymlinks.go`, `stats/internal/guard/check_guard_symlinks_test.go`, `stats/internal/guard/references.go`
**Tests:** `TestStageKeysMatchFlowInteractiveSkillTable`
**Regression:** `TestStageKeysMatchFlowInteractiveSkillTable` fails — the skill marks a key the vocabulary refuses for `/flow-interactive`, or the vocabulary admits one the skill never marks.
**Commit:** feat(flow-interactive): add the /flow-interactive skill and command
**After:** Task 2, 4

**Decision:** inline-implementation
**Decision:** merge-and-push-only
**Decision:** session-worktree
**Decision:** no-learning-note
**Decision:** hook-registration-notice
**Decision:** live-ui-check-after-landing
**Decision:** resume-by-key-or-name
**Decision:** brainstorm-in-session
**Decision:** no-base-flag

- [x] 13. Register `/flow-interactive` in the indexes and contracts
  - [x] **Step 1: Edit the indexes.** After each `skills/flow-fast/` row of the skill index in `CLAUDE.md`, `AGENTS.md`, `templates/CLAUDE.md` and `templates/AGENTS.md`, the same new row: `` | `skills/flow-interactive/` | `/flow-interactive` | Conversational session: one invocation is one change, one Jira issue and one worktree, and a stream of requests the main session works inline — each looked at live, implemented test-first, verified with the project's lint and targeted tests, landed by merge and push, reinstalled and, when it touches `ui paths`, checked on the live page — until an explicit stop runs one self-review pass, cleanup and `FINISHED`. Every request is recorded in flowd; no spectre artifacts or state file | ``. `README.md`'s command table, after `/flow-fast`: `` | `/flow-interactive [key or name]` | Conversational session: each request the operator types is implemented inline, verified, landed by merge and push and, for UI, checked on the live page; an explicit stop self-reviews, cleans up and finishes. No spectre artifacts or state file | ``. `skills/README.md`: after `/flow-fast`'s Command map row, a row whose cells are `` `/flow-interactive [key or name]` ``, `` `flow-interactive` `` and the index row's description, and after `flow-fast/`'s Skills tree line, the line `├── flow-interactive/   ← /flow-interactive (a conversational session landing each request)`.
  - [x] **Step 2: Edit the reinstall rule.** In `CLAUDE.md` and `AGENTS.md`'s **Reinstall after a change to this repository is fully done**, after "on `/flow-fast`, after its cleanup;" insert " on `/flow-interactive`, after each request lands;". Wrap at each file's width.
  - [x] **Step 3: Edit the Jira contract.** `skills/flow-contracts/jira-integration.md`: **Resolution**'s opening sentence names "a `/flow-interactive` session start" beside "a `/flow-fast` run"; after the "An argument asking for a new issue" bullet, a bullet: "**A `/flow-interactive` session start with no key in its arguments** files one the same way, unasked, its summary `Interactive session <YYYY-MM-DD>` — unless the argument declines an issue (\"no jira\"), which resolves `jiraIssue: null` with no ask."; **Transitions** gains the rows `` | `/flow-interactive` | session start, immediately after the key resolves | **In Progress** | `` and `` | `/flow-interactive` | the explicit stop, after cleanup | **Done** | ``; **Description sync**'s opening sentence gains, after "are the only commands that write the issue description", " — save a `/flow-interactive` stop's session summary, below —", and after the **Append only.** paragraph a paragraph: "**A `/flow-interactive` stop appends a `## Session summary` section** at the end of the description — one line per request: its number, title, outcome and landed commit — under the pre-write assertion below."
  - [x] **Step 4: Edit the pipeline contract.** `skills/flow-contracts/pipeline.md` **Command surface**: after "one minimal-ceremony variant that writes no state file (`/flow-fast`)" insert ", one conversational session that lands each request as it goes and writes no state file (`/flow-interactive`)", and after "on `/flow` and `/flow-fast`, a description or Jira key seeding a new change;" insert " on `/flow-interactive`, a first request, a Jira key or an open session's name;". Wrap at the file's width.
  - [x] **Step 5: Verify.** `scripts/run-verify-list.sh <wt> lint` exits 0 save the `scripts/check-verbatim-moves.sh` FAIL lines this task's new or reworded sentences raise, reported per **Global constraints**. Commit.

**Build:** green
**Files:** `CLAUDE.md`, `AGENTS.md`, `templates/CLAUDE.md`, `templates/AGENTS.md`, `README.md`, `skills/README.md`, `skills/flow-contracts/jira-integration.md`, `skills/flow-contracts/pipeline.md`
**Tests:** **none** — registration prose; Step 5's lint is the check.
**Regression:** none — prose.
**Commit:** docs(flow): register /flow-interactive in the indexes and contracts
**After:** Task 2, 12

- [ ] 14. In-run fix 1 — flow.verify: root the `.superpowers/` citation in the skill's Resume section
  - [ ] **Step 1: Reproduce.** `scripts/check-installed-citations.sh` exits 1 naming `skills/flow-interactive/SKILL.md:275: citation `.superpowers/` names no root`.
  - [ ] **Step 2: Fix.** In `skills/flow-interactive/SKILL.md`'s `## Resume`, cite the directory as `<project>/.superpowers/`, and report any `scripts/check-verbatim-moves.sh` FAIL line the reworded sentence raises per **Global constraints**.
  - [ ] **Step 3: Verify.** `scripts/check-installed-citations.sh`, `scripts/check-references.sh` and `scripts/check-verbatim-moves.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow-interactive/SKILL.md`
**Tests:** **none** — a citation in prose; Step 3's guards are the check.
**Regression:** `scripts/check-installed-citations.sh` exits 1 — the skill cites a path with no root.
**Commit:** docs(flow-interactive): root the state-dir citation in Resume
**After:** Task 12
