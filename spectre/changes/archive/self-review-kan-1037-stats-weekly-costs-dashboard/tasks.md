# self-review-kan-1037-stats-weekly-costs-dashboard

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the eight non-`big` findings of `kan-1037-stats-weekly-costs-dashboard`'s self-review pass.

**Architecture:** small edits to flow-guard (`stats/internal/guard`), the `flow` CLI, the store migrations, `stats/Makefile` and two skill files, each with a test that fails without it.

**Spec:** `design.md` beside this file.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1037-stats-weekly-costs-dashboard`, branch `spectre/kan-1037-stats-weekly-costs-dashboard`; `**Files:**` are relative to its root. Never touch the main checkout; never create a worktree; never write `spectre/changes/archive/`.
- One commit per task with its `**Commit:**` subject, no `Task-Id:` trailer. Never merge or push.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- Go test commands run as `cd stats && make -s web-dist && go test <pkg> -run '<name>' -count=1`, Bash timeout 600000; Go lint is `cd stats && gofmt -l . && go vet <pkg>`.
- Committed text cites `kan-1037`, never a commit sha of this branch.

## Review Focus

- Task 1: only verifier rows with outcome `completed` or `skipped` count; `blocked` still does not.
- Task 3: the new test hangs (times out) with the blocking open swapped back in.
- Task 6: the test proves the row's `end` is recorded after stdout's reader closed.

- [x] 1. Let signal seven accept a skipped verify row and correct the Recording paragraph
  - [x] **Step 1: Write the failing test** in `stats/internal/guard/check_unfinished_work_test.go`, `TestCheckUnfinishedWorkSkippedVerifyRowCovers`: a code commit authored before a verifier row `verify` whose outcome is `skipped` gives `CLEAR:`; the same with outcome `blocked` gives `OUTSTANDING:`. Reuse `TestCheckUnfinishedWork`'s fixture shape (`newUWBare`, the dispatches file). Run it; the skipped case fails.
  - [x] **Step 2: Fix** `stats/internal/guard/unfinishedwork.go` signal seven: the outcome test becomes `(d.Outcome == "completed" || d.Outcome == "skipped")`; extend the comment above by one sentence citing decision `f1-skipped-is-verified` (kan-1037).
  - [x] **Step 3: Edit** `skills/flow/verify-and-handoff.md` **Recording**: replace the sentences from "each records `begin` before its first command and `end -outcome\ncompleted` after its last when its list exited 0" through "(a missing build\nprerequisite)." so they state: each group records `begin` before its first command and ends the row itself — `completed` when its list exited 0, `skipped` when every command of a test group was cache-skipped, `blocked -cause test-failure` when the list exited non-zero; once the one inline re-run is decided, a passing re-run ends a fresh row `completed`, and a command that failed twice for a missing build prerequisite is ended `-outcome blocked -cause environment` by the parent. Keep the `begin`-refused (exit 4) and failed-`end` sentences unchanged.
  - [x] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1` passes; `cd stats && gofmt -l . && go vet ./internal/guard/` clean; `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0.
  - [x] **Step 5: Commit** the three files.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestCheckUnfinishedWorkSkippedVerifyRowCovers`
**Regression:** `TestCheckUnfinishedWorkSkippedVerifyRowCovers` — reverting makes the skipped row's case OUTSTANDING.
**Baseline:** before=11 after=12
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/check_unfinished_work_test.go @ spectre/kan-1037-stats-weekly-costs-dashboard -->
**Commit:** fix(guard): let signal seven accept a skipped verify row
**After:** none

**Decision:** f1-skipped-is-verified

- [x] 2. State that a verify re-run after a new code commit takes a fresh key
  - [x] **Step 1: Write the failing test** `TestCheckUnfinishedWorkReasonNamesFreshVerifyKey` in `stats/internal/guard/check_unfinished_work_test.go`: a code commit after the last completed verify row gives a reason line containing "a re-run after a new commit takes a fresh `verify-<k>` key". Run it; it fails.
  - [x] **Step 2: Fix** the signal-seven reason string in `stats/internal/guard/unfinishedwork.go`: append " — a re-run after a new commit takes a fresh `verify-<k>` key (the store keeps a reused key's first start)".
  - [x] **Step 3: Edit** `skills/flow/verify-and-handoff.md` **Recording**: after the paragraph's sentence on the first `begin` stamping the start, add: "A verify re-run after a new code commit passes a key no earlier verify row of the change used — `verify-<k>`, `<k>` counting from 2, suffixed as above — because the store keeps a reused key's first `startedAt`, and a row started before the commit never covers it (kan-1037)."
  - [x] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1` passes; `cd stats && gofmt -l . && go vet ./internal/guard/` clean; `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0.
  - [x] **Step 5: Commit** the three files.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestCheckUnfinishedWorkReasonNamesFreshVerifyKey`
**Regression:** `TestCheckUnfinishedWorkReasonNamesFreshVerifyKey` — reverting drops the fresh-key clause from the reason.
**Baseline:** before=12 after=13
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/check_unfinished_work_test.go @ spectre/kan-1037-stats-weekly-costs-dashboard -->
**Commit:** fix(guard): name the fresh verify key in signal seven's reason
**After:** Task 1

**Decision:** f2-fresh-verify-key

- [x] 3. Release the prove-reproducer FIFO without blocking
  - [x] **Step 1: Write the failing test** `TestPRReleaseFIFOWithNoReaderReturns` in `stats/internal/guard/prove_reproducer_test.go`: `syscall.Mkfifo` in `t.TempDir()`, no reader, call `prReleaseFIFO(path)` in a goroutine, fail if it has not returned within 5 s. Write `prReleaseFIFO` first with the current blocking `os.O_WRONLY` open and run the test: it fails on the 5 s bound (the hang, proven).
  - [x] **Step 2: Fix** `prReleaseFIFO`: open with `os.O_WRONLY|syscall.O_NONBLOCK`, close on success, ignore the error (ENXIO = no reader). Replace both `os.OpenFile(release, os.O_WRONLY, 0)` sites in `TestProveReproducerSignals` with `prReleaseFIFO(release)`. For the ignored-signal site, which must wait for the leg's reader, keep the release in a loop: `for !prReleaseFIFO(release) { select { case <-done: break loop; case <-time.After(10*time.Millisecond): } }` — `prReleaseFIFO` returns whether it opened. Comment cites decision `f3-nonblocking-release` (kan-1037).
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestPRReleaseFIFOWithNoReaderReturns|TestProveReproducerSignals' -count=3` passes; `cd stats && gofmt -l . && go vet ./internal/guard/` clean.
  - [x] **Step 4: Commit** the file.

**Build:** green
**Files:** `stats/internal/guard/prove_reproducer_test.go`
**Tests:** `TestPRReleaseFIFOWithNoReaderReturns`
**Regression:** `TestPRReleaseFIFOWithNoReaderReturns` — reverting to the blocking open hangs past the 5 s bound.
**Baseline:** before=7 after=8
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/prove_reproducer_test.go @ spectre/kan-1037-stats-weekly-costs-dashboard -->
**Commit:** test(guard): release the prove-reproducer FIFO without blocking
**After:** none

**Decision:** f3-nonblocking-release

- [x] 4. Refuse npm ci over a symlinked node_modules
  - [x] **Step 1: Write the failing test** `TestMakeWebBuildRefusesSymlinkedNodeModules` in `stats/internal/repotest/make_build_test.go`: copy `stats/Makefile` into a temp dir with a `web/` dir whose `node_modules` is a symlink to a temp dir holding a marker file and a stub `npm` on `PATH` that exits 0; `make web-build` and `make web-test` each exit non-zero with "symlinked web/node_modules" on stderr, the stub never invoked, the marker still present. Run it; it fails.
  - [x] **Step 2: Fix** `stats/Makefile`: `web-test` and `web-build` recipes become `cd web && { test ! -L node_modules || { echo "refusing npm ci over a symlinked web/node_modules -- it empties the shared directory (kan-1037)" >&2; exit 1; }; } && npm ci && …`, keeping each target's existing tail.
  - [x] **Step 3: Edit** `skills/flow/review-panel-optional-slots.md` **The throwaway worktree**: after the paragraph ending "that slot is\nnot dispatched against that copy.", add: "Never symlink a shared `node_modules` into a copy: the copy's own `npm ci` installs one, and over a symlink it empties the shared directory (kan-1037) — `stats/Makefile` refuses to run it there."
  - [x] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/repotest/ -run 'TestMakeWebBuildRefusesSymlinkedNodeModules' -count=1` passes; `cd stats && gofmt -l . && go vet ./internal/repotest/` clean; `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0.
  - [x] **Step 5: Commit** the three files.

**Build:** green
**Files:** `stats/Makefile`, `stats/internal/repotest/make_build_test.go`, `skills/flow/review-panel-optional-slots.md`
**Tests:** `TestMakeWebBuildRefusesSymlinkedNodeModules`
**Regression:** `TestMakeWebBuildRefusesSymlinkedNodeModules` — reverting runs `npm ci` through the symlink.
**Baseline:** before=1 after=2
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/repotest/make_build_test.go @ spectre/kan-1037-stats-weekly-costs-dashboard -->
**Commit:** fix(stats): refuse npm ci over a symlinked web/node_modules
**After:** none

**Decision:** f4-no-symlinked-node-modules

- [x] 5. Match the mutation throwaway copy in the cleanup guard
  - [x] **Step 1: Write the failing test** `TestCCThrowawayCopyMatchesMutation` in `stats/internal/guard/check_cleanup_complete_test.go`: table over `ccWaveGroupCopy` — `/x/demo-mutation-1` true, `/x/demo-wave-group-2` true, `/x/demo-mutation-` false, `/x/demo-mutation-1a` false, `/x/other-mutation-1` false. Run it; the first case fails.
  - [x] **Step 2: Fix** `ccWaveGroupCopy` in `stats/internal/guard/cleanupcomplete.go`: loop over the suffixes `-wave-group-` and `-mutation-`, same digits-only rule; doc comment names `review-panel-optional-slots.md`'s `<worktree>-mutation-<round>` (kan-1037).
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCCThrowawayCopyMatchesMutation|TestCheckCleanupComplete' -count=1` passes; `cd stats && gofmt -l . && go vet ./internal/guard/` clean.
  - [x] **Step 4: Commit** the two files.

**Build:** green
**Files:** `stats/internal/guard/cleanupcomplete.go`, `stats/internal/guard/check_cleanup_complete_test.go`
**Tests:** `TestCCThrowawayCopyMatchesMutation`
**Regression:** `TestCCThrowawayCopyMatchesMutation` — reverting leaves `demo-mutation-1` unmatched.
**Baseline:** before=2 after=3
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/check_cleanup_complete_test.go @ spectre/kan-1037-stats-weekly-costs-dashboard -->
**Commit:** fix(guard): match the mutation throwaway copy in the cleanup guard
**After:** none

**Decision:** f5-mutation-copy

- [x] 6. Ignore SIGPIPE in run-verify-list
  - [x] **Step 1: Write the failing test** `TestRunVerifyListSurvivesAClosedStdout` in `stats/internal/guard/run_verify_list_test.go`: run the flow-guard `run-verify-list` subcommand as a subprocess the way the file's other subprocess cases do (re-exec of the test binary or the built binary, whichever the file already uses), with a two-command list whose first command prints enough to fill a pipe, stdout an `os.Pipe` whose read end is closed before start; assert the process was not signalled and the fake store recorded the row's `end`. Run it; it fails (death by SIGPIPE).
  - [x] **Step 2: Fix** `stats/internal/guard/runverifylist.go`: first statement of `runVerifyList` is `signal.Ignore(syscall.SIGPIPE)`, with a comment: a closed reader (`| head`) must not kill the list before its `end` (kan-1037).
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestRunVerifyList' -count=1` passes; `cd stats && gofmt -l . && go vet ./internal/guard/` clean.
  - [x] **Step 4: Commit** the two files.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `spectre/changes/self-review-kan-1037-stats-weekly-costs-dashboard/design.md`
**Tests:** `TestRunVerifyListSurvivesAClosedStdout`
**Regression:** `TestRunVerifyListSurvivesAClosedStdout` — reverting lets SIGPIPE kill the list before its `end`.
**Baseline:** before=6 after=7
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/run_verify_list_test.go @ spectre/kan-1037-stats-weekly-costs-dashboard -->
**Commit:** fix(guard): ignore SIGPIPE in run-verify-list
**After:** none

**Decision:** f6-ignore-sigpipe

- [x] 7. Accept -change on flow stage
  - [x] **Step 1: Write the failing test** `TestStageBeginAcceptsChangeFlag` in `stats/cmd/flow/stage_test.go`, on the fake server `TestStageBeginRecordsIdentityAndInstant` uses: `flow stage begin -command /flow -stage flow.brainstorm -session-token mf-x -change demo` records against `demo`; `-change demo demo` is a usage error. Run it; it fails.
  - [x] **Step 2: Fix** `stats/cmd/flow/stage.go`: register `-change` beside `-jira-key` on the shared flag set; when set it is the change and no positional is accepted; update `stageUsage`'s `(<change> | -jira-key KEY)` forms to `(<change> | -change name | -jira-key KEY)` and `stage sessions`' `<change>` likewise.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestStage' -count=1` passes; `cd stats && gofmt -l . && go vet ./cmd/flow/` clean.
  - [x] **Step 4: Commit** the two files.

**Build:** green
**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`
**Tests:** `TestStageBeginAcceptsChangeFlag`
**Regression:** `TestStageBeginAcceptsChangeFlag` — reverting rejects `-change` as an unknown flag.
**Baseline:** before=40 after=41
<!-- measured: grep -cE '^func Test[A-Z]' stats/cmd/flow/stage_test.go @ spectre/kan-1037-stats-weekly-costs-dashboard -->
**Commit:** feat(flow): accept -change on flow stage
**After:** none

**Decision:** f7-stage-change-flag

- [x] 8. Drop the unused stage_runs.stage_key column
  - [x] **Step 1: Write the failing test** `TestStageRunsCarriesNoStageKey` in a new `stats/internal/store/drop_stage_key_test.go`, on the package's existing migrated test-database helper: `information_schema.columns` holds no row for `stage_runs.stage_key` and `pg_indexes` none for `stage_runs_stage_key`. Run it; it fails.
  - [x] **Step 2: Fix** — add `stats/internal/store/migrations/0048_drop_stage_runs_stage_key.sql`: a header comment (migration 0008 added the column for a vocabulary task that never wrote or read it; it misled kan-1037's design) then `DROP INDEX IF EXISTS stage_runs_stage_key;` and `ALTER TABLE stage_runs DROP COLUMN IF EXISTS stage_key;`. If a test lists the migration count or `stage_runs` columns, update it in this commit.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/store/ -count=1 -run 'TestStageRunsCarriesNoStageKey|Legacy'` passes; `cd stats && gofmt -l . && go vet ./internal/store/` clean.
  - [x] **Step 4: Commit** the files.

**Build:** green
**Files:** `stats/internal/store/migrations/0048_drop_stage_runs_stage_key.sql`, `stats/internal/store/drop_stage_key_test.go`
**Allowed-collateral:** `stats/internal/store/*_test.go`
**Tests:** `TestStageRunsCarriesNoStageKey`
**Regression:** `TestStageRunsCarriesNoStageKey` — reverting leaves the column and index in place.
**Commit:** fix(store): drop the unused stage_runs.stage_key column
**After:** none

**Decision:** f8-drop-stage-key
