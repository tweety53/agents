# self-review-kan-1038-stats-weekly-costs-one-chart

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the thirteen non-`big` findings of `kan-1038-stats-weekly-costs-one-chart`'s self-review pass.

**Architecture:** small edits to flow-guard (`stats/internal/guard`), the harvester, the `flow` CLI,
`.flow/project.md`, two harness scripts, the SPA's tests and the run-loaded prompts, each behaviour
change with a test that fails without it.

**Spec:** `design.md` beside this file.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1038-stats-weekly-costs-one-chart`, branch `spectre/kan-1038-stats-weekly-costs-one-chart`; `**Files:**` are relative to its root. Never touch the main checkout; never create a worktree; never write `spectre/changes/`.
- One commit per task with its `**Commit:**` subject, no `Task-Id:` trailer. Never merge or push.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container. `make -C stats ui-test-up` (the disposable UI-test stack) is allowed.
- Go test commands run as `cd stats && make -s web-dist && go test <pkg> -run '<name>' -count=1`, Bash timeout 600000; Go lint is `cd stats && gofmt -l . && go vet <pkg>`.
- Every prompt edit is copied exactly as the task's fenced block states it: `verbatim-moves.txt` beside this file acknowledges exactly those sentences, and `scripts/check-verbatim-moves.sh` fails on any other wording.
- Committed text cites `kan-1038`, never a commit sha of this branch.

## Review Focus

- Task 3: `TestFlowAddrDeclaration` still holds exactly one records-resolver registration, and each new test fails with its verb back on `resolveDefaultAddr()`.
- Task 5: no timeout, threshold or retry is widened; the commit body shows each check failing with its pin removed.
- Task 6: a reproducer that restores what it changed still reaches its verdict.
- Task 8: the restamped figures match the turns inside each row's window.

- [x] 1. Run the guard package's script audits for a touched script
  - [x] **Step 1: Write the failing test** `TestPipelineFixChecksScriptAudits` in `stats/internal/guard/pipeline_fix_test.go`, on `TestPipelineFixChecksGoAudits`'s fixture: with `stats/internal/guard/doc.go` written into the worktree, touched `scripts/pipeline-fix.sh`, `scripts/lib/flow-guard.sh` and `setup.sh` each select `pfScriptAudits`; a worktree without `stats/internal/guard` and a touched Go file alone do not. Run it; it fails.
  - [x] **Step 2: Fix** `stats/internal/guard/pipelinefix.go`: add, beside `pfGoAudits`, the constant below with a comment naming the four tests and kan-1038; in `pfChecks`, `add(pfScriptAudits)` when `isDir(filepath.Join(wt, "stats/internal/guard"))` and a touched path is `setup.sh` or starts with `scripts/`. Extend `pfChecks`'s doc comment by that clause.

```go verified:the four test names grepped in stats/internal/guard/*_test.go at plan time
const pfScriptAudits = "cd stats && go test ./internal/guard/ -run '^Test(NoPipeIntoEarlyExitGrep|ShimMissingLibCannotAnswer|EveryFlowGuardShimRequiresLib|ShimSiblingsDeclared)$' -count=1"
```
  - [x] **Step 3: Header.** In `scripts/pipeline-fix.sh`'s `verify` entry, after "the setuptest package for setup.sh,", add "the internal/guard tests that audit every script (TestNoPipeIntoEarlyExitGrep and its kin) for a touched script or setup.sh,".
  - [x] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestPipelineFix' -count=1` passes; `cd stats && gofmt -l . && go vet ./internal/guard/` clean; `scripts/check-references.sh` exits 0.
  - [x] **Step 5: Commit** the three files.

**Build:** green
**Files:** `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`, `scripts/pipeline-fix.sh`
**Tests:** `TestPipelineFixChecksScriptAudits`
**Regression:** `TestPipelineFixChecksScriptAudits` — reverting drops the audits from a touched script's checks.
**Baseline:** before=12 after=13
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/pipeline_fix_test.go @ branch spectre/kan-1038-stats-weekly-costs-one-chart -->
**Commit:** fix(guard): run the script audits for a touched script in pipeline-fix verify
**After:** none

**Decision:** f1-script-audits

- [x] 2. Record every begin before its launch
  - [x] **Step 1: Edit** `skills/flow/implement.md`, the both-calls paragraph — replace the first block below with the second:

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
Both calls are required. `begin` is recorded immediately before the dispatch — or, when several
launches go out together in one message, in the next Bash call, one call for all (step 3 of **The
next implementer overlaps the guard**, below) — carrying no
`-agent-id`: the daemon captures the agent's identifier.
```

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
Both calls are required. `begin` is recorded immediately before the dispatch — when several
launches go out together in one message, in the one Bash call immediately before that message, one
call for all (step 3 of **The next implementer overlaps the guard**, below) — carrying no
`-agent-id`: the daemon captures the agent's identifier. A launch whose row is not yet open is
refused by `hooks/enforce-agent-baseline.py` (kan-1038), and the harvester fills the row's empty
`agent_id` from the launch result afterwards.
```
  - [x] **Step 2: Edit** `skills/flow/implement.md`, step 3 of **The next implementer overlaps the guard** — replace the first block with the second:

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
3. **One message launches group N+2's implementer and, on `big` when any gate fired in group N+1,
   that group's one reviewer bundle (below). The next Bash call records every launch's `begin`**.
```

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
3. **One Bash call records the `begin` of group N+2's implementer and, on `big` when any gate fired
   in group N+1, of that group's one reviewer bundle (below). The next message launches them all**.
```
  - [x] **Step 3: Edit** `skills/flow/review-panel.md`: replace "Every dispatch of a round launches in one message; every `begin` is recorded in the next Bash\ncall, one call for all, as" with "Every `begin` of a round is recorded in the Bash call immediately before the one message that\nlaunches every dispatch of the round, one call for all, as" (`\n` a line break); and in the `-agent-id` paragraph replace "in the one Bash call after its launch message" with "in the one Bash call before its launch message".
  - [x] **Step 4: Edit** `skills/flow/SKILL-rationale.md`, the `implement.md — 4. Execute` section: replace "— or, when several launches go out together in one message, in the next Bash call, one call for all" with "— when several launches go out together in one message, in the one Bash call immediately before that message, one call for all".
  - [x] **Step 5: Verify.** `git grep -n -F -e 'in the next Bash call' -e 'one Bash call after its launch' -e 'The next Bash call records' -- skills/flow/implement.md skills/flow/review-panel.md skills/flow/SKILL-rationale.md` prints nothing; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0.
  - [x] **Step 6: Commit** the three files.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`, `skills/flow/SKILL-rationale.md`
**Tests:** **none** — text only; `hooks/enforce-agent-baseline.py`'s begin-row check already enforces the order.
**Regression:** none — no test is added.
**Commit:** docs(flow): record every dispatch begin before its launch
**After:** none

**Decision:** f2-begin-first

- [x] 3. Resolve the records address for flow stage, state, settings and jira
  - [x] **Step 1: Write the failing tests**, each modelled on `TestSuiteResolvesTheRecordsAddress` (`stats/cmd/flow/suite_test.go`: `FLOW_ADDR` at `deadPortURL(t)`, `FLOW_RECORDS_ADDR` at a fake store, asserting the request reached it): `TestStageResolvesTheRecordsAddress` in `stats/cmd/flow/stage_test.go` (a `flow stage begin`), `TestStateResolvesTheRecordsAddress` in `stats/cmd/flow/state_test.go` (a `flow state get`), `TestSettingsResolvesTheRecordsAddress` in `stats/cmd/flow/settings_test.go`, `TestJiraResolvesTheRecordsAddress` in `stats/cmd/flow/jira_test.go` (a `flow jira transition` against its `fake.handler()`). Run them; all four fail.
  - [x] **Step 2: Fix.** In `stats/cmd/flow/record.go` add `func registerRecordsAddrFlag(fset *flag.FlagSet, p *string) { fset.StringVar(p, "addr", resolveRecordsAddr(), "flowd base URL") }`, the one records-resolver registration; `registerRecordConnFlags` calls it, and its comment names the verbs below as record-family members. Replace the `-addr` registration in `registerStageIdentityFlags` and `runStageSessions` (`stage.go`; the latter registers into a local `string`), `parseStateFlagsN`, `parseStateListFlags` and `parseStateFindFlags` (`state.go`), `registerSettingsConnFlags` (`settings.go`) and `registerJiraConnFlags` (`jira.go`) with a `registerRecordsAddrFlag` call.
  - [x] **Step 3: Declare.** In `.flow/project.md` `## workspace isolation`, the record-family sentence's parenthesis gains `` `flow stage`, `flow state`, `flow settings`, `flow jira` `` after `` `flow record` ``, on the sentence's one physical line through "resolves", and the sentence adds that stage marks, state rows, settings and Jira transitions must outlive the worktree's daemon as the records do (kan-1038).
  - [x] **Step 4: Follow the seam in the audit** `stats/internal/repotest/flow_addr_declaration_test.go`: assertion 2 requires the one `"addr", resolveRecordsAddr()` registration inside `registerRecordsAddrFlag`; assertion 3 counts a file whose line contains `registerRecordConnFlags(` or `registerRecordsAddrFlag(` (a call, not the definitions in `record.go` — record.go is in the family already); the `mutate-record-seam-default` case mutates `registerRecordsAddrFlag`'s registration, and `mutate-literal-addr-registration` moves to `record.go`'s `registerConnFlags` registration.
  - [x] **Step 5: Verify.** `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'ResolvesTheRecordsAddress|ResolveDefaultAddr|ResolveRecordsAddr|RiderVerbs' -count=1` and `cd stats && go test ./internal/repotest/ -run '^TestFlowAddrDeclaration$' -count=1` pass; `cd stats && gofmt -l . && go vet ./cmd/flow/ ./internal/repotest/` clean.
  - [x] **Step 6: Commit** the files below.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/stage.go`, `stats/cmd/flow/state.go`, `stats/cmd/flow/settings.go`, `stats/cmd/flow/jira.go`, `stats/cmd/flow/stage_test.go`, `stats/cmd/flow/state_test.go`, `stats/cmd/flow/settings_test.go`, `stats/cmd/flow/jira_test.go`, `stats/internal/repotest/flow_addr_declaration_test.go`, `.flow/project.md`
**Tests:** `TestStageResolvesTheRecordsAddress`, `TestStateResolvesTheRecordsAddress`, `TestSettingsResolvesTheRecordsAddress`, `TestJiraResolvesTheRecordsAddress`
**Regression:** each — reverting sends its verb to `FLOW_ADDR`'s dead port and the fake store sees no request.
**Baseline:** before=336 after=340
<!-- measured: cat stats/cmd/flow/*_test.go | grep -cE '^func Test[A-Z]' @ branch spectre/kan-1038-stats-weekly-costs-one-chart -->
**Commit:** fix(flow-cli): resolve the records address for stage, state, settings and jira
**After:** none

**Decision:** f3-records-addr

- [x] 4. Never count an inline verifier row in the visual-verify gate
  - [x] **Step 1: Write the failing test** `TestCheckVisualVerifyDispatchedIgnoresAnInlineVerifier` in `stats/internal/guard/check_visual_verify_dispatched_test.go`, on `vvdRun`: a UI change whose only completed `visual-verify-fix-run-1` verifier row carries `"agentId":"inline"` reads `VISUAL-VERIFY-MISSING`; the same row with a launched agent's id reads `VISUAL-VERIFY-OK`. Run it; it fails.
  - [x] **Step 2: Fix** `vvdCompletedVerifiers` in `stats/internal/guard/visualverifydispatched.go`: `continue` on `row["agentId"] == "inline"` beside the role check, with a comment: a row the parent backfilled inline proves no verifier ran (kan-1038).
  - [x] **Step 3: Header.** `scripts/check-visual-verify-dispatched.sh`: after "A stage mark is not evidence the stage ran; only a recorded dispatch is.", add "A row recorded `-agent-id inline` is the parent's own, not a verifier's, and never counts (kan-1038)."
  - [x] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatched' -count=1` passes; `cd stats && gofmt -l . && go vet ./internal/guard/` clean.
  - [x] **Step 5: Commit** the three files.

**Build:** green
**Files:** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`
**Tests:** `TestCheckVisualVerifyDispatchedIgnoresAnInlineVerifier`
**Regression:** `TestCheckVisualVerifyDispatchedIgnoresAnInlineVerifier` — reverting lets the inline row read OK.
**Baseline:** before=10 after=11
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/check_visual_verify_dispatched_test.go @ branch spectre/kan-1038-stats-weekly-costs-one-chart -->
**Commit:** fix(guard): never count an inline verifier row in the visual-verify gate
**After:** none

**Decision:** f4-inline-verifier

- [x] 5. Reproduce the two load flakes, then pin them at their source
  - [x] **Step 1: Reproduce** on Task 9's commit, from `stats/web` with the dev stack's UI-test instance fresh: run `scripts/run-verify-list.sh <worktree> lint` and `scripts/run-verify-list.sh <worktree> test` (no recording arguments) at once beside `npx playwright test tests/visual --reporter=list`, keeping every output whole — never grepped. Repeat to the measured rate per the PROOF RUNS rule; with no failure after the default run count, add one `yes > /dev/null` per core during the runs. Record the vitest test's timing and the two failing screenshot tests' names, error text and `test-results/*-diff.png`. unverified:that the flakes reproduce at all after Task 9 halves the guard load — if neither reproduces under synthetic load either, report it with the run counts and stop; the task is then not done, not waived.
  - [x] **Step 2: Pin the vitest test** `state board shows the change's state and run totals` (`stats/web/src/views/views.test.tsx`) at its source: scope its screen-wide `getByRole` queries to the table or row they read with `within()`, so its work under load stays far below the default timeout. Never raise `testTimeout` or the test's own timeout.
  - [x] **Step 3: Pin each screenshot** by the cause its error names: a "waiting for stable screenshots" timeout → pin the visual suite's concurrency in `stats/web/playwright.config.ts` (`workers`) and correct its "no rendering pin is needed" comment; a real pixel difference → pin whatever moved, read off the `-actual` and `-diff` images. Never a `maxDiffPixels`, `threshold`, `timeout` or `retries` change.
  - [x] **Step 4: Prove each pin.** Under the Step 1 load, each check fails with its pin removed and passes with it, at the run counts PROOF RUNS sizes; the commit body names each pin, what it replaces, and both run counts.
  - [x] **Step 5: Verify.** `cd stats/web && npx vitest run src/views/views.test.tsx -t 'state board shows'` and `cd stats/web && npx playwright test <the two specs Step 1 names>` pass; `cd stats && make -s web-dist && cd web && npx tsc -b` clean.
  - [x] **Step 6: Commit** the pins.

**Build:** green
**Files:** `stats/web/src/views/views.test.tsx`
**Allowed-collateral:** `stats/web/tests/visual/**`, `stats/web/src/**/*.test.tsx`
**Tests:** **none** — repairs existing tests: the vitest test state board shows the change's state and run totals, and the two screenshot tests Step 1 names.
**Regression:** the named vitest test and the two screenshot tests fail under the Step 1 load with the pins reverted.
**Commit:** test(stats): pin the SPA test and screenshots that flaked under verify load
**After:** Task 9

**Decision:** f5-load-flakes

- [x] 6. Refuse a reproducer run that leaves the worktree changed
  - [x] **Step 1: Write the failing test** `TestRunReproducerRefusesAChangedTree` in `stats/internal/guard/runreproducer_test.go`: in a git worktree fixture, a reproducer that appends to a tracked file and exits 1 makes `run-reproducer` exit 5, stderr carrying both status listings; one that changes the file and restores it before exiting 1 exits 0. Run it; it fails.
  - [x] **Step 2: Fix** `stats/internal/guard/runreproducer.go`: factor `rrGitStatus(worktree) (string, bool)` out of `rrGitStatusNote`; snapshot before `cmd.Start()`; after the survivor, timeout and table-error branches, when the snapshot was taken and the after-listing differs, print `run-reproducer: the run left the worktree changed — before:` and `after:` with both listings and return 5.
  - [x] **Step 3: Callers.** `stats/internal/guard/panelexitcontract.go` gains `case 5:` reading "the reproducer left the tree it ran in changed, which is no verdict" as cannot-answer; `stats/internal/guard/provereproducer.go`'s leg message lists "5 tree changed".
  - [x] **Step 4: Header.** `scripts/run-reproducer.sh`'s exit-code list gains: "5  the run left the worktree changed — `git status --porcelain --untracked-files=normal` differs from the listing taken before the start; both listings go to stderr, and the exit status is no verdict (kan-1038)."
  - [x] **Step 5: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestRunReproducer|TestProveReproducer|TestCheckPanelExitContract' -count=1` passes; `cd stats && gofmt -l . && go vet ./internal/guard/` clean.
  - [x] **Step 6: Commit** the five files.

**Build:** green
**Files:** `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`, `stats/internal/guard/panelexitcontract.go`, `stats/internal/guard/provereproducer.go`, `scripts/run-reproducer.sh`, `stats/internal/guard/prove_reproducer_test.go`
**Tests:** `TestRunReproducerRefusesAChangedTree`
**Regression:** `TestRunReproducerRefusesAChangedTree` — reverting returns the reproducer's own verdict over a changed tree.
**Baseline:** before=7 after=8
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/runreproducer_test.go @ branch spectre/kan-1038-stats-weekly-costs-one-chart -->
**Commit:** fix(guard): refuse a reproducer run that leaves the worktree changed
**After:** none
**Correction:** the commit also edits `stats/internal/guard/prove_reproducer_test.go`, whose expected leg message gains "5 tree changed" with Step 3's provereproducer.go change.

**Decision:** f6-tree-unchanged

- [x] 7. Restart a stale UI-test stack before capture and verify
  - [x] **Step 1: Edit** `.flow/project.md` `## visual verification`'s command table: `verify` becomes `` cd stats/web && { ../../scripts/check-dev-stack-fresh.sh ../.. \|\| make -C .. ui-test-up; } && npm run test:visual `` and `capture` becomes `` cd stats/web && { ../../scripts/check-dev-stack-fresh.sh ../.. \|\| make -C .. ui-test-up; } && npx playwright test <spec> --update-snapshots ``. After the `fingerprint` paragraph add: "**`capture` and `verify` check the stack first.** Each opens on `check-dev-stack-fresh.sh`, the declared `fingerprint` run, and restarts the disposable UI-test stack with `make ui-test-up` when it is stale, so an implementer's or fixer's capture never reads a pre-fix build as a failure (kan-1038); `make ui-test-up` touches only that stack, never the `flowd` on 4173."
  - [x] **Step 2: Verify.** `scripts/check-visual-verification.sh .` exits 0; `scripts/project-get.sh . 'visual verification'` prints both rows with `||` unescaped; `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckVisualVerification|TestCheckDevStackFresh' -count=1` passes.
  - [x] **Step 3: Commit** the file.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — a configuration row; `check-visual-verification.sh` validates its shape.
**Regression:** none — no test is added.
**Commit:** fix(flow-config): restart a stale UI-test stack before capture and verify
**After:** Task 3

**Decision:** f7-fresh-stack

- [x] 8. Settle the verify rows run-verify-list.sh begins, and restamp this change's
  - [x] **Step 1: Write the failing tests** in `stats/internal/harvest/watcher_test.go`, beside the inline settle tests and their fake deps: `TestSettleClosedRowsMeasuresARunVerifyListRow` — a parent transcript whose only begin for key `verify` is a Bash `scripts/run-verify-list.sh <wt> test <change> verify <token> opus medium` call, followed by several parent turns inside the closed row's window, restamps the row with the sum of those turns; `TestSettleClosedRowsZeroesAnEmptyRunVerifyListRow` — the same with no turn inside, a zero restamp. Run them; both fail.
  - [x] **Step 2: Fix** `stats/internal/harvest/watcher.go`: beside `inlineBeginFromInvocation`, `runVerifyListBeginFromInvocation(invocation string, vars map[string]string) (dispatchBeginEvent, bool)` recognises `run-verify-list.sh` with `lint` or `test` and all five recording arguments, returning `dispatchBeginEvent{SessionToken: <session-token>, Key: <key>}` (refusing a `$` or backtick in either, as `inlineBeginFromInvocation` does); the settle pass takes candidates from both extractors.
  - [x] **Step 3: Repair.** Look up the store ids of this change's verify rows — keys `verify`, `verify-fix-1` through `verify-fix-4`, `verify-6`, `verify-7`, `visual-verify-fix-run-1` (ledger dispatches 13, 20, 28, 34, 35, 38, 39, 40) — and run `cd stats && go run ./cmd/flowd restamp-dispatch <id> ~/.claude/projects/-Users-tweety53-Projects-agents/c96e0ccd-56f2-4b3d-94c9-3299350a571f.jsonl` for each; record each row's before/after figures in the commit body. unverified:that the ledger's dispatch number is the store id — resolve each id by key first.
  - [x] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/harvest/ -run 'TestSettleClosedRows' -count=1` passes; `cd stats && gofmt -l . && go vet ./internal/harvest/` clean.
  - [x] **Step 5: Commit** the two files.

**Build:** green
**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`
**Tests:** `TestSettleClosedRowsMeasuresARunVerifyListRow`, `TestSettleClosedRowsZeroesAnEmptyRunVerifyListRow`
**Regression:** both — reverting leaves the script-begun row unsettled, its figures one turn or none.
**Baseline:** before=73 after=75
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/harvest/watcher_test.go @ branch spectre/kan-1038-stats-weekly-costs-one-chart -->
**Commit:** fix(harvest): settle the verify rows run-verify-list.sh begins
**After:** none

**Decision:** f8-settle-verify-rows

- [x] 9. Run the guard package once per verify
  - [x] **Step 1: Write the failing test** `TestTestListRunsTheGuardPackageOnce` in a new `stats/internal/repotest/test_list_test.go`, using `repoRoot(t)`: `scripts/test-go-guards.sh`'s Go test line carries `-race`, and the Go line of `.flow/project.md`'s `## test` fence draws its packages from the module's package list with `/internal/guard` filtered out, never the bare module pattern. Run it; it fails.
  - [x] **Step 2: Fix** both lines exactly as decision `f9-guard-once` (`design.md`) states them: `scripts/test-go-guards.sh` gains `-race`, its header saying the `## test` Go line excludes this package so it runs once (kan-1038); `.flow/project.md` `## test`'s Go line filters `/internal/guard` out; the paragraph naming `scripts/test-go-guards.sh` adds that the Go line excludes `internal/guard`, which that harness runs under `-race`; the `stats-go` suite's command in the runtime paragraph is updated to match.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/repotest/ -run '^TestTestListRunsTheGuardPackageOnce$' -count=1` passes; `cd stats && go list ./... | grep -c '/internal/guard$'` prints `1` and the filtered list omits it; `cd stats && gofmt -l . && go vet ./internal/repotest/` clean.
  - [x] **Step 4: Commit** the three files.

**Build:** green
**Files:** `scripts/test-go-guards.sh`, `.flow/project.md`, `stats/internal/repotest/test_list_test.go`
**Tests:** `TestTestListRunsTheGuardPackageOnce`
**Regression:** `TestTestListRunsTheGuardPackageOnce` — reverting runs the guard package in both lines again.
**Baseline:** before=21 after=22
<!-- measured: cat stats/internal/repotest/*_test.go | grep -cE '^func Test[A-Z]' @ branch spectre/kan-1038-stats-weekly-costs-one-chart -->
**Commit:** perf(flow-config): run the guard package once per verify
**After:** Task 7

**Decision:** f9-guard-once

- [x] 10. Ask every fix-round re-run for defects the fix introduces
  - [x] **Step 1: Write the failing test** `TestCheckDispatchParagraphsFixRoundNamesIntroducedDefects` in `stats/internal/guard/check_dispatch_paragraphs_test.go`, on `dpDoc`: a `review-panel-fix-round.md` whose FIX-ROUND SCOPE block lacks "for a defect the fix itself introduces" exits 1 naming that phrase; the block carrying it passes. Run it; it fails.
  - [x] **Step 2: Fix** `stats/internal/guard/dispatchparagraphs.go`: the `fixround` entry's phrases gain `"for a defect the fix itself introduces"`; the `dpFixRoundScopeBlock*` fixtures carry the new sentence so every other case still reads as before.
  - [x] **Step 3: Edit** `skills/flow/review-panel-fix-round.md`'s FIX-ROUND SCOPE block — replace the first line below with the four after it:

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
> claims — and say which of those it was. Given a diff, your report file carries one verdict line
```

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
> claims — and say which of those it was. Read every hunk of the fix diff, a fix made for another
> slot's finding included, for a defect the fix itself introduces — an evaluation moved out of the
> guard that kept it safe, a branch the fix changed — and report it as a new finding at its line.
> Given a diff, your report file carries one verdict line
```
  - [x] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckDispatchParagraphs|TestRenderSlotPrompt' -count=1` passes; `scripts/check-dispatch-paragraphs.sh`, `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh` exit 0; `cd stats && gofmt -l . && go vet ./internal/guard/` clean.
  - [x] **Step 5: Commit** the three files.

**Build:** green
**Files:** `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`, `skills/flow/review-panel-fix-round.md`
**Allowed-collateral:** `stats/internal/guard/render_slot_prompt_test.go`
**Tests:** `TestCheckDispatchParagraphsFixRoundNamesIntroducedDefects`
**Regression:** `TestCheckDispatchParagraphsFixRoundNamesIntroducedDefects` — reverting lets the block drop the clause unnoticed.
**Baseline:** before=2 after=3
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/check_dispatch_paragraphs_test.go @ branch spectre/kan-1038-stats-weekly-costs-one-chart -->
**Commit:** fix(flow): ask every fix-round re-run for defects the fix introduces
**After:** none

**Decision:** f10-introduced-defects

- [x] 11. Render an approved prototype as each capture's frame
  - [x] **Step 1: Write the failing test** `TestCheckVerifyReportStep10WithACitedPrototype` in `stats/internal/guard/check_verify_report_test.go`, on `TestCheckVerifyReportNoMockupSourceWithCitedSource`'s fixture: `design.md` citing `` `.superpowers/sdd/proto.html` `` that exists and a report reading `10 n/a — mockups not declared` exits 1 naming the prototype; the cited file absent, no `--design`, a `.css`-only citation, or a wire contract naming no view each exit 0. Run it; it fails.
  - [x] **Step 2: Fix** `stats/internal/guard/verifyreport.go`: in the sr12 block, when step 10 reads `n/a` and no wire contract names no view, the first backticked `.html`/`.htm` source `design.md` cites that exists sets `status["10"]` to "n/a while design.md cites the prototype <path> — render it as each capture's frame (visual-verify-verifier.md step 10)". `scripts/check-verify-report.sh`'s header names the refusal (kan-1038).
  - [x] **Step 3: Edit** `skills/flow/visual-verify-verifier.md` step 10 — insert this paragraph before the `> **The \`mockups\` declaration gates the frame comparison, never the sweeps.**` blockquote, and in that blockquote replace "project declaring no `mockups` root composes nothing and reports each of them" with "project declaring no `mockups` root, with no approved prototype to render (above), composes\n   > nothing and reports each of them":

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
   **An approved prototype is the frame where no mockup is** (kan-1038). When no sidecar maps a
   capture — `mockups` undeclared included — and `design.md`'s `## Context` cites an approved
   prototype HTML file that exists in the worktree (**The checklist**,
   `skills/flow/brainstorm-planner.md`), render that file as each capture test's frame first. With
   the project's own installed Playwright, a Node script you write under
   `<changeRoot>/visual-verification/prototype/` opens `file://<prototype>` at the capture test's
   own viewport, device scale factor and colour scheme, drives the state that test captures —
   every hover, focus or pressed state the prototype draws included — and writes
   `<changeRoot>/visual-verification/prototype/<screenshot stem>.png`. Then write
   `<changeRoot>/visual-verification/prototype.mockups`, its first line `frame none` and one
   `<screenshot name> <screenshot stem>` line per capture, and run the compose command above with
   that map in place of `<spec>.mockups` and `<changeRoot>/visual-verification/prototype` in place
   of `<worktree>/<mockups>`. Every frame-bound reading below then runs on those pairs as on a
   mockup's. A state the prototype draws that no capture test drives is a missing capture, and
   blocks.
```
  - [x] **Step 4: Edit** `skills/flow/visual-verify.md`: after "`design.md` cites a source that exists (kan-1017: every verifier left 7 of 11 matrix columns n/a)." add, on new lines, "It also refuses step 10 reported n/a while `design.md` cites an HTML source that exists: an\napproved prototype is rendered as each capture's frame (step 10,\n`skills/flow/visual-verify-verifier.md`; kan-1038)."
  - [x] **Step 5: Edit** `skills/flow/brainstorm-planner.md`'s prototype bullet: replace "`design.md`'s `## Context` as the spec the implementation and `flow.visual-verify` compare\n  against." with "`design.md`'s `## Context` as the spec the implementation and `flow.visual-verify` compare\n  against, and every capture test of a view it draws drives each interactive state it shows —\n  hover, focus, pressed — so `flow.visual-verify` renders and compares that state too\n  (kan-1038)."
  - [x] **Step 6: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0; `cd stats && gofmt -l . && go vet ./internal/guard/` clean.
  - [x] **Step 7: Commit** the six files.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`, `skills/flow/visual-verify-verifier.md`, `skills/flow/visual-verify.md`, `skills/flow/brainstorm-planner.md`, `spectre/changes/self-review-kan-1038-stats-weekly-costs-one-chart/verbatim-moves.txt`
**Tests:** `TestCheckVerifyReportStep10WithACitedPrototype`
**Regression:** `TestCheckVerifyReportStep10WithACitedPrototype` — reverting accepts step 10 `n/a` beside a cited prototype.
**Baseline:** before=5 after=6
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/check_verify_report_test.go @ branch spectre/kan-1038-stats-weekly-costs-one-chart -->
**Commit:** feat(flow): render an approved prototype as each capture's frame
**After:** none

**Decision:** f11-prototype-frame

- [x] 12. List the visual record in a visual-fix task's Files
  - [x] **Step 1: Edit** `skills/flow/verify-fix-loop.md` step 1 — replace the first line below with the block after it:

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
   this stage's round in this run. Its planning commit and the re-decide run as that file states.
```

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
   this stage's round in this run. On a `flow.visual-verify` round its `**Files:**` lists the
   change's `<changeRoot>/visual-verification.md` beside the fix's own paths, and the task rewrites
   the record's sections its fix changes in the same commit, as DRIFT CHECK
   (`skills/flow/implement.md`) requires (kan-1038). Its planning commit and the re-decide run as
   that file states.
```
  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0.
  - [x] **Step 3: Commit** the file.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — text only.
**Regression:** none — no test is added.
**Commit:** docs(flow): list the visual record in a visual-fix task's Files
**After:** none

**Decision:** f12-record-in-files

- [x] 13. Never ask a scope or cost question with a recommended option at integrate or the self-review
  - [x] **Step 1: Edit** `skills/flow-contracts/operator-prompts-auto-resolution.md`: insert this paragraph, then a blank line, before "Every auto-resolution:"; and in the **Outside implementation and fix runs.** bullet replace "and every integrate and cleanup prompt." with "and every integrate and cleanup prompt but\n  the scope and cost questions above.":

```markdown verified:applied in a scratch copy at plan time; check-verbatim-moves 0 FAIL with verbatim-moves.txt
**A scope or cost question with a recommended option is never asked during integrate or the
self-review either** (operator, kan-1038). "Fix every finding (recommended), or only the
blocker?", whether to run one more fix or review round, whether a fix is worth its cost: the run
takes the recommended option, records it as below and goes on. A question written as prose in a
message, a final message's included, is an ask exactly as an `AskUserQuestion` call is. The
outward-facing and irreversible actions of **What still stops** below are still asked.
```
  - [x] **Step 2: Edit** `skills/flow-self-review/SKILL.md` step 3: after "**Every finding that is not `big` is fixed and landed without asking**, on one branch for the\nwhole pass." add " No scope or cost question with a recommended option is put to the operator anywhere in\nthis pass, as a prompt or as prose in a message (**Auto-resolution**,\n`skills/flow-contracts/operator-prompts-auto-resolution.md`).", the following sentence continuing on the same line.
  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0.
  - [x] **Step 4: Commit** the two files.

**Build:** green
**Files:** `skills/flow-contracts/operator-prompts-auto-resolution.md`, `skills/flow-self-review/SKILL.md`
**Tests:** **none** — text only.
**Regression:** none — no test is added.
**Commit:** docs(flow-contracts): auto-resolve scope and cost questions at integrate and self-review
**After:** none

**Decision:** f13-no-scope-asks
