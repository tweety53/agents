## Context

Self-review fixes for `kan-1038-stats-weekly-costs-one-chart`, on the change's own branch
`spectre/kan-1038-stats-weekly-costs-one-chart` in
`/Users/tweety53/Projects/agents-worktrees/kan-1038-stats-weekly-costs-one-chart`
(`fixes-ride-the-change`). The change is archived; this plan never touches `spectre/changes/archive/`.
The pass's evidence is `.superpowers/sdd/self-review-pass.md`; what planning measured on top of it:

**F1.** `pfChecks` (`stats/internal/guard/pipelinefix.go`) adds `pfGoAudits` (repotest) for touched Go
and `go test ./internal/repotest/` for a touched hook, but nothing of `internal/guard` for a touched
script. The guard tests that read this checkout's real scripts are `TestNoPipeIntoEarlyExitGrep`
(`grep_pipe_test.go`, globs `setup.sh`, `scripts/*.sh`, `scripts/lib/*.sh`),
`TestShimMissingLibCannotAnswer` (`shim_template_test.go`, `scripts/*.sh`),
`TestEveryFlowGuardShimRequiresLib` and `TestShimSiblingsDeclared` (`check_guard_symlinks_test.go`,
`scripts/`). No `internal/guard` test reads `hooks/`: `git grep -n '\.\./\.\./\.\./hooks'
stats/internal/guard` finds nothing, and a touched hook already selects repotest, where its tests live.

**F2.** Five passages order launch before `begin`: `skills/flow/review-panel.md` (the round's
`panel-round-begins.sh` sentence and the `-agent-id` sentence), `skills/flow/implement.md` (the
both-calls paragraph and step 3 of **The next implementer overlaps the guard**) and
`skills/flow/SKILL-rationale.md`. `begin_row_reason` (`hooks/enforce-agent-baseline.py`) denies any
launch with no open row.

**F3.** `resolveRecordsAddr()` (`stats/cmd/flow/state.go`) is wired only inside
`registerRecordConnFlags` (`record.go`); `registerStageIdentityFlags` and `runStageSessions`
(`stage.go`), `parseStateFlagsN`, `parseStateListFlags` and `parseStateFindFlags` (`state.go`),
`registerSettingsConnFlags` (`settings.go`) and `registerJiraConnFlags` (`jira.go`) register `-addr`
from `resolveDefaultAddr()`. `TestFlowAddrDeclaration`
(`stats/internal/repotest/flow_addr_declaration_test.go`) asserts exactly one `"addr",
resolveRecordsAddr()` registration, inside `registerRecordConnFlags`, and that the files calling the
record seam are exactly those of `.flow/project.md`'s record-family sentence. Repair measured at plan
time: `flow record journal-count -C . -change kan-1038-stats-weekly-costs-one-chart` prints `0`, the
stage journal `~/Agents/flow/state/agents-a740d89c/kan-1038-stats-weekly-costs-one-chart.journal.stage`
is empty, and `flow stage sessions -C . -stage flow.verify kan-1038-stats-weekly-costs-one-chart`
prints `1` — the journalled marks already replayed; no repair is outstanding.

**F4.** `vvdCompletedVerifiers` (`stats/internal/guard/visualverifydispatched.go`) reads role, key,
outcome, diffBase, startedAt and roundBase and never `agentId`; `panelfixsingledispatch.go` already
reads `row["agentId"] == "inline"`.

**F5.** The vitest test is `views render their fixture response's actual values > state board shows
the change's state and run totals` (`stats/web/src/views/views.test.tsx`; the run's log
`.git/worktrees/kan-1038-stats-weekly-costs-one-chart/flow-verify-failed/test-2.log`: `Test timed out
in 5000ms`, 5721ms). It is the file's first test, its fetch mock resolves at once and no timer is
involved; it runs about twelve screen-wide `getByRole` queries, whose visibility walk is costly in
jsdom. The two screenshot specs are not recoverable: the run grepped the reporter's output down to its
totals and `stats/web/test-results` was overwritten. `stats/web` carries no CSS animation,
transition or web font; `playwright.config.ts` leaves `workers` unset and has no
`expect.toHaveScreenshot` block.

**F6.** `rrGitStatusNote` (`stats/internal/guard/runreproducer.go`) prints the tree's status only on
the kill, timeout and table-error branches. Exit codes 0–4 are taken (`scripts/run-reproducer.sh`).
`panelexitcontract.go` reads any other code as cannot-answer; `provereproducer.go` spends the proof on
any leg other than 0 or 1.

**F7.** `.flow/project.md` `## visual verification`: `capture` is `cd stats/web && npx playwright test
<spec> --update-snapshots` and `verify` is `cd stats/web && npm run test:visual`; only
`visual-verify.md` step 6 runs `fingerprint`. `scripts/check-dev-stack-fresh.sh <worktree>` runs the
declared `fingerprint` row and exits 0 fresh, 1 stale, 2 cannot answer. `make ui-test-up`
(`stats/Makefile`) stops the previous UI-test daemon on its own port and rebuilds.

**F8.** Traced: no launch pairing is involved. Every verify row is begun and ended by
`runVerifyList` (`stats/internal/guard/runverifylist.go`), so the parent transcript holds only the
`run-verify-list.sh <worktree> lint|test <change> <key> <session-token> <model> <effort>` call, never
a `flow record dispatch begin … -agent-id inline`. `settleClosedRows`
(`stats/internal/harvest/watcher.go`) re-measures closed inline rows over the whole transcript, but
takes its candidates only from `inlineBeginFromInvocation`, which needs that literal begin. Rows 13
(`verify`) and 20 (`verify-fix-1`) carry exactly the one turn that ran the parent's later explicit
`dispatch end` (13: input 2, output 370, cache read 356706 — the turn issuing the end at
10:48:09Z); 28, 34, 35, 38, 39 and 40 had no explicit end and read `not measured`. Both outcomes are
wrong, from one cause. `flowd restamp-dispatch <dispatch-id> <parent transcript>`
(`stats/cmd/flowd/main.go`) re-measures one inline row over a transcript and replaces its figures.

**F9.** `.flow/project.md` `## test` runs `scripts/run-guard-tests.sh`, which runs
`scripts/test-go-guards.sh` (`go test ./internal/guard/... -count=1`), beside `go test ./... -race
-count=1`, which runs `internal/guard` again. `go list ./internal/guard/...` lists the one package.
`runguardtests.go` accepts a ported guard's `_test.go` as its companion because `test-go-guards.sh`
runs those tests, so that harness stays.

**F10.** The FIX-ROUND SCOPE block (`skills/flow/review-panel-fix-round.md`) is rendered verbatim into
every re-run (`renderslotprompt.go`); the introduced-defect clause sits only in the parent-facing
paragraph above it. `check-dispatch-paragraphs.sh`'s `fixround` entry (`dispatchparagraphs.go`) pins
three phrases of the block, matched on the block's lines joined by spaces.

**F11.** `compose-mockup-frames.sh` takes `<map> <mockups root> <out dir> [frame]`; a map line is
`<screenshot name> <frame id>`, the frame read as `<root>/<frame id>.png`, and a `frame none` first
line crops and size-checks nothing. Any directory serves as the root, so a rendered prototype composes
with no code change. No tool renders a local HTML file in a hover or focus state;
`stats/web` pins `@playwright/test` 1.62.1. `check-verify-report.sh --design`
(`verifyreport.go`, sr12) already reads `design.md`'s backticked HTML/CSS sources, and allows step 10
`n/a` whenever its reason names `mockups`.

**F12.** `skills/flow/verify-fix-loop.md` step 1 states the appended task's field family by citation
only; DRIFT CHECK (`skills/flow/implement.md`) wants the record updated in the same commit, and
`check-task-commit-fields.sh` holds the commit to **Files:**.

**F13.** `skills/flow-contracts/operator-prompts-auto-resolution.md` scopes auto-resolution to
implementation and fix runs; its last **What still stops** bullet keeps "every integrate and cleanup
prompt" asked.

**Doc edits and the verbatim guard.** Every rewritten run-loaded sentence is acknowledged in
`verbatim-moves.txt` beside this file, generated by running `scripts/check-verbatim-moves.sh` over the
exact edits `tasks.md` states, applied in a scratch copy of this branch at plan time (0 FAIL with the
file present). The tasks' text is therefore exact: a reworded edit fails that guard.

## Decisions

### Script audits for a touched script

**ID:** f1-script-audits
**Status:** active
**Chosen:** a constant `pfScriptAudits = "cd stats && go test ./internal/guard/ -run
'^Test(NoPipeIntoEarlyExitGrep|ShimMissingLibCannotAnswer|EveryFlowGuardShimRequiresLib|ShimSiblingsDeclared)$'
-count=1"`, added by `pfChecks` when a touched path is `setup.sh` or starts with `scripts/` and the
worktree has `stats/internal/guard`. Test `TestPipelineFixChecksScriptAudits`.
**Considered:** the same for `hooks/*` — disproved for this package: no `internal/guard` test reads
`hooks/` (Context, F1), and repotest already runs for a touched hook. The whole guard package for any
touched script — about 500 s per loop verify where four tests suffice.

### Begin before launch

**ID:** f2-begin-first
**Status:** active
**Chosen:** rewrite the five passages so every `begin` — a panel round's `panel-round-begins.sh`
call — is recorded in the Bash call immediately before the launch message, and the harvester still
fills `agent_id` from the launch result afterwards. Text only; the hook is the guard that already
enforces it.

### The records address for stage, state, settings and jira

**ID:** f3-records-addr
**Status:** active
**Chosen:** one helper `registerRecordsAddrFlag(fset *flag.FlagSet, p *string)` in `record.go`
holding the one `"addr", resolveRecordsAddr()` registration; `registerRecordConnFlags` and the seven
registrations of Context F3 call it (`runStageSessions` registers into a local `string`). The
record-family sentence of `.flow/project.md` adds `flow stage`, `flow state`, `flow settings` and
`flow jira`; `TestFlowAddrDeclaration` follows the seam: its assertion 2 locates the registration in
`registerRecordsAddrFlag`, its assertion 3 counts a file calling `registerRecordConnFlags(` or
`registerRecordsAddrFlag(` anywhere in a line, and its `mutate-literal-addr-registration` case moves
off `jira.go` to `registerConnFlags`'s own registration in `record.go`. One test per verb, each with
`FLOW_ADDR` at a dead port and `FLOW_RECORDS_ADDR` at a fake store, modelled on
`TestSuiteResolvesTheRecordsAddress`. No repair step: Context F3 measured none outstanding.
**Considered:** swapping each site to `resolveRecordsAddr()` inline — `TestFlowAddrDeclaration`
refuses a second registration of the records resolver. Moving `registerConnFlags` (hazard, spec,
tasks) too — their telemetry lives and dies with the run, as `record.go` states.

### An inline verifier row does not count

**ID:** f4-inline-verifier
**Status:** active
**Chosen:** `vvdCompletedVerifiers` skips a row whose `agentId` is `inline`. Test
`TestCheckVisualVerifyDispatchedIgnoresAnInlineVerifier`; the `check-visual-verify-dispatched.sh`
header says only a launched verifier counts.

### Load flakes reproduced, then pinned

**ID:** f5-load-flakes
**Status:** active
**Chosen:** the task opens on the trace. After F9's task, reproduce under verify's own concurrency —
both `run-verify-list.sh` groups plus `npm run test:visual` at once, with the full Playwright
`--reporter=list` output kept — until the timing-out test and both screenshots are named with their
error text; add synthetic CPU load (`yes > /dev/null`, one per core) when 15 runs reproduce nothing.
Then pin each at its source. For the vitest test: cut its own work — scope its screen-wide role
queries to the table they read with `within()`. For a screenshot: from its `-actual`/`-diff` images —
a stability timeout is pinned by the visual suite's own `workers`, a pixel difference by what moved.
Never a longer timeout, a threshold or a retry (`fix-determinism-at-the-source`); the commit body
names the pins and shows each check failing with its pin removed.
**Considered:** `configure({ defaultHidden: true })` in `setupTests.ts` — changes every role query's
semantics across the suite to speed one test.

### run-reproducer refuses a changed tree

**ID:** f6-tree-unchanged
**Status:** active
**Chosen:** snapshot `git -C <worktree> status --porcelain --untracked-files=normal` before the start;
after a run that reached a verdict, a different listing prints both listings on stderr and exits 5, a
new code in the header's contract: "the run left the worktree changed; its exit status is no
verdict". A tree git cannot read before the start takes no snapshot and skips the check, as
`rrGitStatusNote` is silent. `provereproducer.go`'s leg message and `panelexitcontract.go` name exit
5. Test `TestRunReproducerRefusesAChangedTree`.
**Considered:** folding it into exit 4 — a changed tree is the reproducer's fault, not the
environment's, and the caller's remedy differs.

### Fresh stack before capture and verify

**ID:** f7-fresh-stack
**Status:** active
**Chosen:** both rows open on `{ ../../scripts/check-dev-stack-fresh.sh ../.. || make -C ..
ui-test-up; } &&`, reusing the declared `fingerprint` through the existing guard rather than a second
copy of it; one sentence in the section says why.

### Settle script-begun verify rows

**ID:** f8-settle-verify-rows
**Status:** active
**Chosen:** a second candidate extractor beside `inlineBeginFromInvocation`, fed to the same settle
pass: an invocation of `run-verify-list.sh` with all five recording arguments yields
`dispatchBeginEvent{SessionToken: <session-token>, Key: <key>}` as an inline begin. Tests
`TestSettleClosedRowsMeasuresARunVerifyListRow` (every parent turn inside the window summed) and
`TestSettleClosedRowsZeroesAnEmptyRunVerifyListRow` (a zero restamp). Repair: `flowd
restamp-dispatch` over the parent transcript for this change's eight verify rows.
**Considered:** never crediting an inline row — the inline path is the design; the defect is the
missing candidate.

### The guard package once per verify

**ID:** f9-guard-once
**Status:** active
**Chosen:** `scripts/test-go-guards.sh` runs `go test ./internal/guard/... -race -count=1`, and the
`## test` Go line becomes `cd stats && make -s web-dist && go test -race -count=1 $(go list ./... |
grep -v '/internal/guard$')`. Test `TestTestListRunsTheGuardPackageOnce` (repotest) pins both.
**Considered:** dropping `test-go-guards.sh` — `runguardtests.go`'s companion rule rests on it.

### Introduced defects in every re-run's scope

**ID:** f10-introduced-defects
**Status:** active
**Chosen:** the FIX-ROUND SCOPE block gains the sentence `tasks.md` Task 10 states, and the
`fixround` entry pins `for a defect the fix itself introduces`. Test
`TestCheckDispatchParagraphsFixRoundNamesIntroducedDefects`.

### The approved prototype is the frame

**ID:** f11-prototype-frame
**Status:** active
**Chosen:** the verifier renders the prototype with the project's own Playwright, one PNG per
capture at its viewport and state, into `<changeRoot>/visual-verification/prototype/`, maps them in a
`frame none` map, and composes with the existing script — no new script. `visual-verify.md` and
`brainstorm-planner.md` state the parent's and planner's halves. `check-verify-report.sh --design`
refuses step 10 `n/a` while `design.md` cites a backticked `.html`/`.htm` source that exists, unless a
wire contract names no view. Test `TestCheckVerifyReportStep10WithACitedPrototype`.
**Considered:** a Playwright spec in `stats/web` — serves this repository only, where the verifier
runs in every project. `npx playwright screenshot` — cannot hover or focus.

### The visual record rides the visual fix

**ID:** f12-record-in-files
**Status:** active
**Chosen:** one sentence in `verify-fix-loop.md` step 1, Task 12's text.

### No scope or cost asks at integrate or the self-review

**ID:** f13-no-scope-asks
**Status:** active
**Chosen:** a paragraph in **Auto-resolution** extending it to scope and cost questions with a
recommended option during integrate and the self-review, prose asks included; the integrate bullet of
**What still stops** excepts them; one sentence in the self-review's step 3. Outward-facing and
irreversible actions stay asked.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — thirteen tasks across the guard, the harvester, the `flow` CLI, the SPA
tests and run-loaded prompts.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** four fixer dispatches, in series: g1 Tasks 1, 4, 6, 10, 11 (`stats/internal/guard` and
their prompts); g2 Tasks 3, 7, 8, 9 (`stats/cmd/flow`, `.flow/project.md`, the harvester,
repotest); g3 Tasks 2, 12, 13 (prompts only); g4 Task 5 (the load trace, after Task 9).

## Open questions

None.

## Recommended approach

The decisions above, one task each, in `tasks.md`.

## Critical files

`stats/internal/guard/pipelinefix.go`, `stats/internal/guard/visualverifydispatched.go`,
`stats/internal/guard/runreproducer.go`, `stats/internal/guard/dispatchparagraphs.go`,
`stats/internal/guard/verifyreport.go`, `stats/internal/harvest/watcher.go`,
`stats/cmd/flow/record.go`, `.flow/project.md`, `scripts/test-go-guards.sh`,
`stats/web/src/views/views.test.tsx`, `stats/web/playwright.config.ts`,
`skills/flow/visual-verify-verifier.md`, `skills/flow-contracts/operator-prompts-auto-resolution.md`.

## Existing code to reuse

`TestPipelineFixChecksGoAudits`'s fixture, `vvdRun`/`vvdRows`, `TestSuiteResolvesTheRecordsAddress`
with `deadPortURL` and `genuineDaemon`, `rrGitStatusNote`'s git call, the inline settle tests'
fake deps in `watcher_test.go`, `dpDoc` and the `dpFixRoundScopeBlock` fixtures,
`TestCheckVerifyReportNoMockupSourceWithCitedSource`'s fixture, `check-dev-stack-fresh.sh`,
`compose-mockup-frames.sh`.

## Verification

Each task's targeted tests; `cd stats && make -s web-dist && go vet ./... && gofmt -l .`;
`scripts/check-references.sh`; `scripts/check-verbatim-moves.sh`.

## Live check

- F3: with the built `flow`, `FLOW_ADDR=http://127.0.0.1:1 FLOW_RECORDS_ADDR=http://127.0.0.1:4173
  flow state get -C <worktree> kan-1038-stats-weekly-costs-one-chart` prints the change's state and
  no `wrote local journal` line, and `flow record journal-count -C <worktree> -change
  kan-1038-stats-weekly-costs-one-chart` stays `0` before and after. Not working looks like
  `store unreachable` and a count above 0.
- F8: after Task 8's restamp, `flow record render -kind ledger` for the change shows token figures —
  a number, `0` included — on the eight verify rows, where before 28, 34, 35, 38, 39 and 40 read
  `not measured` and 13 read input 2 / output 370. Unchanged figures mean the restamp missed.
- F7: `cd stats/web && ../../scripts/check-dev-stack-fresh.sh ../..` before and after the edited
  `verify` row; a stale stack (exit 1) before reads fresh (exit 0) after, the 4173 `flowd` untouched.
