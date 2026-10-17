# self-review-kan-940-step-5-frontend-cardio-row-on-the-kan-928

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-17-self-review-kan-940-step-5-frontend-cardio-row-on-the-kan-928-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`. Never merge or push.
- Fixer groups run in series, each task in its group in the order listed (`design.md`, decision `fixer-groups`): `fix-g1` Tasks 4, 5, 6, 8, 10; `fix-g2` Tasks 1, 2; `fix-g3` Tasks 3, 7, 9, 11. Baselines assume that chain.
- Every skills diff below is applied with `git apply` (Tasks 5 and 9: `git apply --unidiff-zero`) from the repository root, exactly as given, in the chain order the groups give; every added sentence is acknowledged verbatim in this change's `verbatim-moves.txt`, so any other wording fails `scripts/check-verbatim-moves.sh`. A diff that does not apply is reported, never hand-merged.
- Go in `stats/`: run `cd stats && gofmt -w .` then `gofmt -l .` (prints nothing) and `go vet` of each touched package. Each verify step tests only the packages and `-run` names it lists; the parent's `pipeline-fix.sh verify` runs the rest.
- A test that runs the `flow` CLI inside this Claude Code session inherits `CLAUDE_CODE_SESSION_ID`: after Task 8, any test passing `-tokens` sets `t.Setenv("CLAUDE_CODE_SESSION_ID", "")` (no `t.Parallel()` in such a test).

## Review focus

- A landing that moved no `stats/` path must not restart flowd, and a restart must run only after the push and a successful fast-forward (Task 1).
- `close-task` exit 3 only where every tick and push succeeded; exits 1 and 2 keep their meaning (Task 2).
- A positional that follows a flag-looking value — `-C <dir> <name>` — still parses as before (Task 4).
- An explicit `-effort default` is still accepted (Task 5); an empty `-location` is still accepted (Task 6).
- The ` — HEAD <sha>` suffix stays the last part of every `VISUAL-VERIFY-MISSING:` line (Task 11).

- [ ] 1. A landing that moved stats/ restarts flowd
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -rn 'restart' stats/internal/guard/pipelinefix.go stats/internal/guard/refreshmaincheckout.go` and confirm no hit — the defect: after a landing that changed `stats/`, the live flowd and `~/.local/bin/flow` keep the old build.
  - [ ] **Step 2: Write the failing tests.** In `stats/internal/guard/pipeline_fix_test.go`, `TestPipelineFixLandRestartsFlowdWhenStatsMoved`: a `pfNew` fixture whose `bin/` also holds a `make` stub appending `make $*` to `fx.log`; the landed commit adds `stats/Makefile` (a non-Go `stats/` path, so `pfChecks` adds no `cd stats` line — read at `pipelinefix.go:320-331`); assert rc 0, output contains `FLOWD-RESTARTED: ` + `fx.repo` + `/stats`, and `fx.calls(t)` contains `make -C ` + `fx.repo` + `/stats restart`; a second sub-case landing `docs/fix.md` only asserts no `make` call and no `FLOWD-` line. In `stats/internal/guard/refresh_main_checkout_test.go`, `TestRefreshMainCheckoutRestartsFlowdWhenStatsMoved`: the same `make` stub on the env's `PATH`; origin gains a commit adding `stats/Makefile`; after `REFRESH-DONE:` the stub log names `-C <main>/stats restart`; with the stub exiting 1 the output carries `FLOWD-RESTART-FAILED:` and rc is 1. Run `cd stats && go test ./internal/guard/ -run 'TestPipelineFixLandRestartsFlowdWhenStatsMoved|TestRefreshMainCheckoutRestartsFlowdWhenStatsMoved' -count=1`: both FAIL (no `make` call).
  - [ ] **Step 3: Implement** `restartFlowdIfStatsMoved(env Env, main, old, new string, stdout, stderr io.Writer) int` in `stats/internal/guard/pipelinefix.go` per decision `flowd-restart-on-landing`: `git -C <main> diff --name-only <old> <new>`; no `stats/` path, or no `<main>/stats/Makefile`, returns 0 silently; else `make` resolved by `lookPath(env, "make")` runs `-C <main>/stats restart` with its output relayed. In `pfLand`, read `git -C <main> rev-parse HEAD` before the `pull --ff-only` and, only when the pull succeeded, call the helper with that and the new HEAD after printing `landed:`; a non-zero helper return makes `pfLand` return 1. In `refreshMainCheckout`, call it after the `REFRESH-DONE:` line with `<old>`/`<new>` that line prints; a non-zero return makes the exit 2. Add to `scripts/pipeline-fix.sh`'s `land` usage and to `scripts/refresh-main-checkout.sh`'s exit-code block the `FLOWD-RESTARTED:`/`FLOWD-RESTART-FAILED:` lines and their exits. In `.flow/project.md`, under the protected-services table, add one paragraph: the restart `pipeline-fix.sh land` and `refresh-main-checkout.sh` run after a landing that moved `stats/` is the one sanctioned restart — after the push, from the landed code, at the end of the run that landed it.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./internal/guard` prints nothing; `go vet ./internal/guard/`; `go test ./internal/guard/ -run 'TestPipelineFix|TestRefreshMainCheckout' -count=1` passes; from the root `scripts/test-protect-main-checkout.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`, `stats/internal/guard/refreshmaincheckout.go`, `stats/internal/guard/refresh_main_checkout_test.go`, `scripts/pipeline-fix.sh`, `scripts/refresh-main-checkout.sh`, `.flow/project.md`
**Tests:** `TestPipelineFixLandRestartsFlowdWhenStatsMoved`, `TestRefreshMainCheckoutRestartsFlowdWhenStatsMoved`
**Regression:** reverting the commit leaves the `make` stub uncalled after a `stats/` landing, and both tests fail on the missing `make -C … restart` call.
**Baseline:** before=9 after=11
<!-- measured: cat stats/internal/guard/pipeline_fix_test.go stats/internal/guard/refresh_main_checkout_test.go | grep -c '^func Test' @ 04a3a161 (9 before; after adds this task's two) -->
**Commit:** feat(guard): restart flowd after a landing that moved stats/
**After:** none

**Decision:** flowd-restart-on-landing

- [ ] 2. close-task surfaces an unrecorded dispatch end
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `cd stats && go test ./internal/guard/ -run 'TestCloseTaskExitContract/a_failed_record_never_blocks' -count=1 -v` and confirm it passes with exit 0 — the defect: the failed end's only trace is the first stderr line.
  - [ ] **Step 2: Write the failing test.** `TestCloseTaskSurfacesAFailedDispatchEnd` in `stats/internal/guard/close_task_test.go`: `ctFixture(t, true, "0", "1")`, one committed task, run with `-end-key k -session-token mf-t`; assert `r.rc == 3`, the tick and push calls still logged, and the last non-empty line of stderr starts `close-task: UNRECORDED — the dispatch end for k did not reach the store; run: flow record dispatch end -change demo -C ` and ends `-outcome completed`. Change the "a failed record never blocks" case's expected rc from 0 to 3. Run `cd stats && go test ./internal/guard/ -run 'TestCloseTask' -count=1`: FAILS (rc 0).
  - [ ] **Step 3: Implement** per decision `close-task-unrecorded-exit-3` in `stats/internal/guard/closetask.go`: remember the failed end; on every return after it, print the UNRECORDED line last (with ` -commit <sha>` when `-end-commit` was given); return 3 where the function would return 0. Update `scripts/close-task.sh`'s step 1 and exit-code paragraph to name exit 3. Apply this diff with `git apply --unidiff-zero`:

````diff verified:applied at 04a3a161 in chain order 5, 2, 3, 7, 9, 11 (5 and 9 with --unidiff-zero); together they pass check-verbatim-moves.sh with verbatim-moves.txt and the doc guards design.md's Context names
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index f920afa8..09f22a9d 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -693,5 +693,7 @@ entry, one or more bundles `plan-dispatch-bundles.sh` emits. At each boundary, i
    cannot resolve, a commit range git cannot resolve, a usage error), or a failed tick or push —
    stops the run**: re-committing cannot repair an inability, so it is never folded into
-   the re-commit loop. Exit 0 prints every task's `FIRE:` or `QUIET:` gate line.
+   the re-commit loop. Exit 0 prints every task's `FIRE:` or `QUIET:` gate line. **Exit 3 with `close-task: UNRECORDED —`** closed every task — ticked and
+   pushed — but the implementer dispatch's end never reached the store: run the `flow record
+   dispatch end` command that line prints before the next dispatch goes out.
 
    **A guard call that times out is inspected before it is retried.** Run
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./internal/guard` prints nothing; `go vet ./internal/guard/`; `go test ./internal/guard/ -run 'TestCloseTask' -count=1` passes; from the root `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `scripts/close-task.sh`, `skills/flow/implement.md`
**Tests:** `TestCloseTaskSurfacesAFailedDispatchEnd`
**Regression:** reverting the commit returns exit 0 with no UNRECORDED line, failing the new test and the amended exit-contract case.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/close_task_test.go @ 04a3a161 (4; Task 10 adds one before this task) -->
**Commit:** fix(guard): close-task surfaces an unrecorded dispatch end as exit 3
**After:** Task 5, 10

**Decision:** close-task-unrecorded-exit-3

- [ ] 3. Spell the inline verify's dispatch pair in full
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `sed -n 114,121p skills/flow/verify-and-handoff.md` and confirm the **Recording** paragraph names no `-change` and no `-session-token`.
  - [ ] **Step 2: Apply this diff** from the repository root:

````diff verified:applied at 04a3a161 in chain order 5, 2, 3, 7, 9, 11 (5 and 9 with --unidiff-zero); together they pass check-verbatim-moves.sh with verbatim-moves.txt and the doc guards design.md's Context names
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index d58b3309..66cb4414 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -121,4 +121,11 @@ recorded before the first command in the list; `end` after the `## Report` is wr
 environment itself failed twice (a missing build prerequisite).
 
+```bash
+flow record dispatch begin -change <name> -role verifier -model <parent model> -effort <parent effort> \
+  -agent-id inline -key verify[-<worktree basename>] -session-token mf-<literal-token>
+flow record dispatch end -change <name> -key verify[-<worktree basename>] \
+  -session-token mf-<literal-token> -outcome completed
+```
+
 ### Live check
 
````

  - [ ] **Step 3: Verify.** From the root `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` pins its exact text.
**Regression:** reverting the commit leaves the inline verify's begin without its two required flags.
**Commit:** docs(flow): spell the inline verify's dispatch begin and end in full
**After:** none

**Decision:** verify-recording-full-pair

- [ ] 4. flow stage and flow tasks take flags after the change name
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `cd stats && go build -o /tmp/flow-sr ./cmd/flow && /tmp/flow-sr stage begin -command /flow -stage flow.verify -session-token mf-x demo -C /tmp; echo $?` prints `expected exactly one argument` and `2`.
  - [ ] **Step 2: Write the failing tests.** `TestStageBeginTakesFlagsAfterTheChangeName` in `stats/cmd/flow/stage_test.go` (the `TestStageBeginRecordsIdentityAndInstant` fake-daemon pattern): `stage begin -command /flow -stage flow.verify -harness h -session-token mf-x demo -C <repo>` exits 0 and the request names change `demo`; `… demo extra` still exits 2. `TestTasksCountTakesFlagsAfterTheChangeName` in `stats/cmd/flow/tasks_test.go`: `tasks count <name> -C <repo>` exits 0 and prints the count. Run `cd stats && go test ./cmd/flow/ -run 'TakesFlagsAfterTheChangeName' -count=1`: both FAIL with exit 2.
  - [ ] **Step 3: Implement** per decision `flags-after-positionals`: `parseInterspersed(fset *flag.FlagSet, args []string) ([]string, error)` in `stats/cmd/flow/state.go`, the loop `parseStateFlagsN` holds at `state.go:262-275`, which then calls it; `finishStageIdentityFlags(fset *flag.FlagSet, pos []string, f *stageIdentityFlags) error` counts `pos`, its three callers in `stage.go` passing what `parseInterspersed` returned (the `args[:sep]` caller keeps its slice); `runTasksCount` and `runTasksTick` in `tasks.go` count the returned positionals.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./cmd/flow` prints nothing; `go vet ./cmd/flow/`; `go test ./cmd/flow/ -run 'TestStage|TestTasks|TestState' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/state.go`, `stats/cmd/flow/stage.go`, `stats/cmd/flow/tasks.go`, `stats/cmd/flow/stage_test.go`, `stats/cmd/flow/tasks_test.go`
**Tests:** `TestStageBeginTakesFlagsAfterTheChangeName`, `TestTasksCountTakesFlagsAfterTheChangeName`
**Regression:** reverting the commit exits 2 on `<name> -C <dir>`, failing both tests.
**Baseline:** before=54 after=56
<!-- measured: cat stats/cmd/flow/stage_test.go stats/cmd/flow/tasks_test.go | grep -c '^func Test' @ 04a3a161 (54 before; after adds this task's two) -->
**Commit:** fix(flow-cli): take flags after the change name in flow stage and flow tasks
**After:** none

**Decision:** flags-after-positionals

- [ ] 5. dispatch begin requires -effort
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'fset.String("effort", "default"' stats/cmd/flow/record.go` hits — the defect: an omitted `-effort` is recorded as `default` with no error.
  - [ ] **Step 2: Write the failing test.** `TestRecordDispatchBeginRequiresEffort` in `stats/cmd/flow/record_test.go` (the `TestRecordFindingRejectsUnknownStatus` pattern): `record dispatch begin` with every required flag but `-effort` exits 2, stderr names `-effort`, the daemon is not contacted, no journal written; the same call with `-effort default` exits 0. Run `cd stats && go test ./cmd/flow/ -run TestRecordDispatchBeginRequiresEffort -count=1`: FAILS (exit 0).
  - [ ] **Step 3: Implement** per decision `effort-required-on-begin`: `-effort`'s default becomes `""`, `[2]string{"-effort", *effort}` joins `requireRecordFlags` in `runRecordDispatchBegin`, the usage line drops the brackets around `-effort level`; every `dispatch begin` call in `stats/cmd/flow/*_test.go` lacking `-effort` gains `"-effort", "default"`. Apply this diff with `git apply --unidiff-zero`:

````diff verified:applied at 04a3a161 in chain order 5, 2, 3, 7, 9, 11 (5 and 9 with --unidiff-zero); together they pass check-verbatim-moves.sh with verbatim-moves.txt and the doc guards design.md's Context names
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index f920afa8..75236867 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -125 +125 @@ flow record dispatch begin -change <name> -role <role> -model <the model origina
-  -key <key>-retry -session-token mf-<literal-token>
+  -effort <the effort originally requested> -key <key>-retry -session-token mf-<literal-token>
@@ -352 +352 @@ Immediately before dispatching:
-flow record dispatch begin -change <name> -task <n> -role implementer -model <m> \
+flow record dispatch begin -change <name> -task <n> -role implementer -model <m> -effort <e> \
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./cmd/flow` prints nothing; `go vet ./cmd/flow/`; `go test ./cmd/flow/ -run 'TestRecordDispatch' -count=1` passes; from the root `scripts/check-verbatim-moves.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/implement.md`
**Tests:** `TestRecordDispatchBeginRequiresEffort`
**Regression:** reverting the commit records an omitted effort as `default`, and the test fails on exit 0.
**Baseline:** before=108 after=109
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 04a3a161 (108 before; after adds this task's one) -->
**Commit:** fix(flow-cli): require -effort on dispatch begin
**After:** none

**Decision:** effort-required-on-begin

- [ ] 6. record finding refuses a -location with no line
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n '"location"' stats/cmd/flow/record.go` shows the flag read with no check after it.
  - [ ] **Step 2: Write the failing test.** `TestRecordFindingRefusesALocationWithNoLine` (same pattern): `-location 'EditorScreen delete confirm'` exits 2, stderr names `file:line`, daemon not contacted; `-location stats/cmd/flow/record.go:12-20` and an omitted `-location` both exit 0. Run `cd stats && go test ./cmd/flow/ -run TestRecordFindingRefusesALocationWithNoLine -count=1`: FAILS.
  - [ ] **Step 3: Implement** per decision `location-needs-line` in `runRecordFinding`: a non-empty `*location` not matching `regexp.MustCompile(`:[0-9]+`)` prints `flow: -location %q must name a file:line (e.g. stats/cmd/flow/record.go:1360)` and returns 2, before the store is contacted.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./cmd/flow` prints nothing; `go vet ./cmd/flow/`; `go test ./cmd/flow/ -run 'TestRecordFinding' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordFindingRefusesALocationWithNoLine`
**Regression:** reverting the commit records prose as a location, and the test fails on exit 0.
**Baseline:** before=109 after=110
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 04a3a161 (108; Task 5 adds one before this task) -->
**Commit:** fix(flow-cli): refuse a finding location that names no line
**After:** Task 5

**Decision:** location-needs-line

- [ ] 7. The self-review bundle drops matrix rows and PNG stat lines
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Confirm `archivedFiles` in `stats/internal/selfreview/git.go` copies `visual-verification.md` whole and `gitLogSection`/`changeBranchLog` return `git log --stat` unfiltered.
  - [ ] **Step 2: Write the failing test.** `TestBundleElidesMatrixRowsAndPNGStatLines` in `stats/internal/selfreview/bundle_test.go` (the `TestBundleServesArchivedVisualVerification` fixture): an archived `visual-verification.md` with a `### Departures` table, a `#### cardio-R1 matrix` heading followed by a header row, a separator and two rows, then `## Motions`; a commit touching `a.go` and `shot.png`. Assert the bundle keeps the departures table and the `####` heading, carries `_(4 matrix rows elided)_`, holds none of the matrix rows, and its `git log --stat` section names `a.go` and not `shot.png`. Run `cd stats && go test ./internal/selfreview/ -run TestBundleElidesMatrixRowsAndPNGStatLines -count=1`: FAILS.
  - [ ] **Step 3: Implement** per decision `bundle-text-filters`: `elideMatrixRows(md string) string` applied to `visual-verification.md`'s content in `Bundle`, and `dropPNGStatLines(s string) string` applied to `gitLogSection`'s and `changeBranchLog`'s output, in `git.go`. Apply this diff with `git apply --unidiff-zero`:

````diff verified:applied at 04a3a161 in chain order 5, 2, 3, 7, 9, 11 (5 and 9 with --unidiff-zero); together they pass check-verbatim-moves.sh with verbatim-moves.txt and the doc guards design.md's Context names
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index 7dacb2a1..5a67597a 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -304,5 +304,7 @@ roots for the pipeline's changes), and the `git log --stat` of every implementat
 archive commit of the change, oldest first, followed by each peer the archived `link.md`'s
 `## Merge order` names, its `spectre/<name>` commits against its base under a `### <peer>`
-heading — and prints it as one Markdown document. A source that is absent is reported
+heading — and prints it as one Markdown document. Its `visual-verification.md` keeps every heading and
+drops the table rows under each `####` element × property matrix heading, and its `git log --stat`
+lines leave out every `*.png` path. A source that is absent is reported
 `skipped: <source> (absent)` inside the bundle, never fatal.
 The session appends `## Session narrative` (one paragraph it writes for run 1 itself; the
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./internal/selfreview` prints nothing; `go vet ./internal/selfreview/`; `go test ./internal/selfreview/ -run 'TestBundle' -count=1` passes; from the root `scripts/check-verbatim-moves.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/selfreview/git.go`, `stats/internal/selfreview/bundle_test.go`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** `TestBundleElidesMatrixRowsAndPNGStatLines`
**Regression:** reverting the commit puts the matrix rows and the PNG stat line back in the bundle, and the test fails.
**Baseline:** before=27 after=28
<!-- measured: grep -c '^func Test' stats/internal/selfreview/bundle_test.go @ 04a3a161 (27 before; after adds this task's one) -->
**Commit:** fix(selfreview): drop visual matrix rows and PNG stat lines from the bundle
**After:** none

**Decision:** bundle-text-filters

- [ ] 8. -tokens is refused under Claude Code
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read `checkDispatchEndFields` in `stats/cmd/flow/record.go` and confirm it accepts any well-formed `-tokens` regardless of harness.
  - [ ] **Step 2: Write the failing test.** `TestRecordDispatchEndRefusesTokensUnderClaudeCode` in `stats/cmd/flow/record_test.go`: `t.Setenv("CLAUDE_CODE_SESSION_ID", "s")`; `record dispatch end … -tokens input=1,output=2` exits 2, stderr names `-tokens` and `CLAUDE_CODE_SESSION_ID`, daemon not contacted; the same for `dispatch backfill`; with `t.Setenv("CLAUDE_CODE_SESSION_ID", "")` the end exits 0. Run `cd stats && go test ./cmd/flow/ -run TestRecordDispatchEndRefusesTokensUnderClaudeCode -count=1`: FAILS.
  - [ ] **Step 3: Implement** per decision `tokens-refused-with-transcript` in `checkDispatchEndFields`; extend the `-tokens` help paragraph in `record.go` with one sentence saying so; set `CLAUDE_CODE_SESSION_ID` empty in every existing test that passes `-tokens`.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./cmd/flow` prints nothing; `go vet ./cmd/flow/`; `go test ./cmd/flow/ -run 'TestRecordDispatch' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchEndRefusesTokensUnderClaudeCode`
**Regression:** reverting the commit lets a Claude Code inline row overwrite its harvested cost, and the test fails on exit 0.
**Baseline:** before=110 after=111
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 04a3a161 (108; Tasks 5 and 6 add one each before this task) -->
**Commit:** fix(flow-cli): refuse -tokens where the harvester reads a transcript
**After:** Task 5, 6

**Decision:** tokens-refused-with-transcript

- [ ] 9. The unfinished-work close command carries -ended-at
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'outcome <outcome> -C' stats/internal/guard/unfinishedwork.go` shows the printed command ending at `-C <anchor>`.
  - [ ] **Step 2: Write the failing test.** `TestCheckUnfinishedWorkCloseCommandCarriesEndedAt` in `stats/internal/guard/check_unfinished_work_test.go`, reusing `TestCheckUnfinishedWork`'s fixture helpers: one open `reviewer` row and one open `implementer` row; the reviewer's command ends `-C <anchor> -ended-at <when its work ended>`, the implementer's `-C <anchor> -ended-at <when its work ended> -commit <its task commit>`. Run `cd stats && go test ./internal/guard/ -run TestCheckUnfinishedWorkCloseCommandCarriesEndedAt -count=1`: FAILS.
  - [ ] **Step 3: Implement** per decision `close-command-ended-at` in `unfinishedwork.go`, reading the row's `Role` the struct already decodes. Apply this diff with `git apply --unidiff-zero`:

````diff verified:applied at 04a3a161 in chain order 5, 2, 3, 7, 9, 11 (5 and 9 with --unidiff-zero); together they pass check-verbatim-moves.sh with verbatim-moves.txt and the doc guards design.md's Context names
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -405 +405,2 @@
-flow record dispatch end -change <name> -C <canonical-worktree> -key <key> -session-token <its token> -outcome <completed|aborted>
+flow record dispatch end -change <name> -C <canonical-worktree> -key <key> -session-token <its token> -outcome <completed|aborted> \
+  -ended-at <when its work ended> [-commit <its task commit>]
@@ -407,0 +409,4 @@
+`-ended-at` is the instant the row's work ended — its task commit's committer date
+(`git log -1 --format=%cI <sha>`), else its report file's mtime — never the moment it is closed,
+and an implementer row carries `-commit` with its task commit's full sha.
+
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./internal/guard` prints nothing; `go vet ./internal/guard/`; `go test ./internal/guard/ -run TestCheckUnfinishedWork -count=1` passes; from the root `scripts/check-verbatim-moves.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestCheckUnfinishedWorkCloseCommandCarriesEndedAt`
**Regression:** reverting the commit prints a close command that stamps "now", and the test fails on the missing `-ended-at`.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/check_unfinished_work_test.go @ 04a3a161 (1 before; after adds this task's one) -->
**Commit:** fix(guard): the unfinished-work close command carries when the work ended
**After:** Task 3

**Decision:** close-command-ended-at

- [ ] 10. -commit is a full sha
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n '"-commit", "abc1234"' stats/cmd/flow/record_test.go` hits four passing tests — the defect: an abbreviated sha is stored.
  - [ ] **Step 2: Write the failing tests.** `TestRecordDispatchRefusesAnAbbreviatedCommit` in `stats/cmd/flow/record_test.go`: `dispatch end … -commit 99153ab4` and `dispatch backfill … -commit 99153ab4` exit 2 naming `40`, daemon not contacted; a 40-hex `-commit` exits 0. `TestCloseTaskExpandsAnAbbreviatedEndCommit` in `stats/internal/guard/close_task_test.go`: `-end-commit <the first 8 chars of a fixture commit>`; the logged `flow record dispatch end` call carries the full 40-char sha. Run `cd stats && go test ./cmd/flow/ -run TestRecordDispatchRefusesAnAbbreviatedCommit -count=1` and `go test ./internal/guard/ -run TestCloseTaskExpandsAnAbbreviatedEndCommit -count=1`: both FAIL.
  - [ ] **Step 3: Implement** per decision `commit-full-sha`: `validateCommitSHA(flagName, sha string) error` in `record.go` (`^[0-9a-f]{40}$`, empty allowed), called by `runRecordDispatchEnd` and `runRecordDispatchBackfill` before the store; replace the four `abc1234` test values with a 40-hex literal; in `closetask.go` expand `-end-commit` through `git -C <canonical> rev-parse --verify --quiet <sha>^{commit}` (`envGit(env)`), keeping the given value when that fails.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./cmd/flow ./internal/guard` prints nothing; `go vet ./cmd/flow/ ./internal/guard/`; `go test ./cmd/flow/ -run 'TestRecordDispatch' -count=1` and `go test ./internal/guard/ -run 'TestCloseTask' -count=1` pass. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`
**Tests:** `TestRecordDispatchRefusesAnAbbreviatedCommit`, `TestCloseTaskExpandsAnAbbreviatedEndCommit`
**Regression:** reverting the commit stores `99153ab4` as a commit and forwards an abbreviated `-end-commit`, failing both tests.
**Baseline:** before=115 after=117
<!-- measured: cat stats/cmd/flow/record_test.go stats/internal/guard/close_task_test.go | grep -c '^func Test' @ 04a3a161 (112; Tasks 5, 6 and 8 add one each before this task) -->
**Commit:** fix(flow-cli): require a full sha for a dispatch's commit
**After:** Task 5, 6, 8

**Decision:** commit-full-sha

- [ ] 11. A VISUAL-VERIFY-MISSING re-run is scoped to the verified HEAD
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read `visualverifydispatched.go`'s heads loop and the `len(heads) > 0` verdict and confirm the verdict names no path; read `verify-and-handoff.md`'s `VISUAL-VERIFY-MISSING:` paragraph and confirm the re-run is unscoped.
  - [ ] **Step 2: Write the failing test.** `TestCheckVisualVerifyDispatchedMissingNamesPathsSinceVerifiedHead` in `stats/internal/guard/check_visual_verify_dispatched_test.go`, on `TestCheckVisualVerifyDispatched`'s fixture: a verifier row whose `diffBase` is commit A, then a UI-path commit B touching `web/src/App.tsx`; the MISSING line contains ` — paths changed since verified HEAD <A>: web/src/App.tsx` and still ends ` — HEAD <B>`. Run `cd stats && go test ./internal/guard/ -run TestCheckVisualVerifyDispatchedMissingNamesPathsSinceVerifiedHead -count=1`: FAILS.
  - [ ] **Step 3: Implement** per decision `missing-names-changed-paths`: in the heads loop keep the last head that resolved and its `changedSince` (unverified:confirm `vvdCompletedVerifiers` returns heads in store order, oldest first; if not, pick the newest by row seq); append the clause in the `len(heads) > 0` branch before ` — HEAD `. Add the clause to the MISSING line in `scripts/check-visual-verify-dispatched.sh`'s header. Apply this diff with `git apply --unidiff-zero`:

````diff verified:applied at 04a3a161 in chain order 5, 2, 3, 7, 9, 11 (5 and 9 with --unidiff-zero); together they pass check-verbatim-moves.sh with verbatim-moves.txt and the doc guards design.md's Context names
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index d58b3309..66cb4414 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -423,5 +435,9 @@ a visual-fix task's, a panel fix's, a full-suite fix's or a fix run's own. Run *
 verification** above again from its `begin` mark on the current HEAD, take every departure through
 **The loop** (`skills/flow/verify-fix-loop.md`), and run the guard again; the state is written,
-and the handoff printed, only once every worktree prints `VISUAL-VERIFY-OK:`. Integrate stops on
+and the handoff printed, only once every worktree prints `VISUAL-VERIFY-OK:`. When the line names a
+verified HEAD and the paths changed since it, that re-run is scoped as a round of **The loop** is
+(`skills/flow/verify-fix-loop.md`): the named HEAD is its round base, its `motions:` names only the
+motions `git diff <that HEAD>..HEAD` adds, changes or removes, and a clean scoped report closes it,
+since that HEAD's full run stands. Integrate stops on
 the same verdict and no operator course there passes it. An exit 2 is reported and stops the run
 at the state it read.
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./internal/guard` prints nothing; `go vet ./internal/guard/`; `go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatched' -count=1` passes; from the root `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestCheckVisualVerifyDispatchedMissingNamesPathsSinceVerifiedHead`
**Regression:** reverting the commit drops the changed-paths clause, and the test fails on its absence.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^func Test' stats/internal/guard/check_visual_verify_dispatched_test.go @ 04a3a161 (3 before; after adds this task's one) -->
**Commit:** feat(guard): a VISUAL-VERIFY-MISSING line names the paths since its verified HEAD
**After:** Task 3, 9

**Decision:** missing-names-changed-paths
