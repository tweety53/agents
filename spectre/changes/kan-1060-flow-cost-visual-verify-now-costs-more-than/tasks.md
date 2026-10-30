# kan-1060-flow-cost-visual-verify-now-costs-more-than

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — the verifier's polled-verify recipe moves from `skills/flow/visual-verify-verifier.md` step 7 into `skills/flow/verify-and-handoff.md`'s **Inline verify**
> **Design shape:** plan-mode

**Goal:** Cut `flow.visual-verify`'s cost (KAN-1060): a thin parent behind `visual-verify-pass.sh prepare|close`, a verifier that no longer re-runs `verify`, one `measure` call for the measuring chain, every flagged element explained with one report per round, the by-hand frame/matrix check made automatic, and Main session / Main share columns on the Stage leaderboard.

**Architecture:** one new `flow-guard` entry, `visual-verify-pass` (`prepare`, `close`, `measure`), running the existing guards in-process; the inline verify row records its HEAD and `check-verify-green` judges by it; `measure-visual-properties` gains an `over` list that `check-verify-report` reconciles against the report; `check-plan-shape` gains F24; the contract prose lands as two patches planned with the change; the leaderboard gains two columns from the runs' own dispatch buckets.

**Tech Stack:** Go (`stats/internal/guard`, `stats/internal/store`, `stats/internal/api`), Bash shims, React/TypeScript (`stats/web`), Playwright, Markdown.

**Spec:** `design.md` beside this file; the approved prototype `.superpowers/sdd/2026-10-30-stage-leaderboard-main-share-prototype.html` (no Artifact URL recorded); the contract text in `contract-edits/thin-parent-and-retired-verify.patch` and `contract-edits/measure-and-explain.patch` beside this file — canonical for every closing line, argument and report line the Go tasks implement, and read-only to every task.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1060-flow-cost-visual-verify-now-costs-more-than` (`<wt>`), branch `spectre/kan-1060-flow-cost-visual-verify-now-costs-more-than`; `**Files:**` paths are relative to `<wt>`. Never touch the main checkout, never create a worktree, never write `spectre/changes/`.
- Go: `gofmt -w` every Go file changed; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean; `scripts/test-go-guards.sh` passes for every guard task. SPA: `cd stats && make -s web-dist && cd web && npx tsc -b` exits 0.
- Guard tests build fixture repositories under `t.TempDir()` and a fake `flow` on the `Env`'s PATH (or `Env.Dispatches`), as `check_verify_green_test.go` and `integrate_change_test.go` do — never the live store. Store tests use the package's existing test-database helper, never the live `flow` database. Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database.
- Every Go test run and every Playwright run sets a Bash `timeout` of 600000.
- Mutation proofs restore with `scripts/break-and-prove.sh` — never `git checkout --` over uncommitted edits.
- New script logic lives in Go behind a `flow_guard_exec` shim; a changed guard's `scripts/<name>.sh` header is its contract and changes in the same task.

## Review Focus

- Task 1: a green row counts only while every commit since its HEAD sits under `<spec root>/changes/`; a row with no `-diff-base`, or one resolving in no worktree of the set, is never green.
- Tasks 3–5: every failing check ends the stage `stopped` before any verifier row begins; `prepare --continue` never begins the stage again; `close` never ends the stage while a surviving worktree's newest row is open, and never commits after `rest`, `redispatch` or `cure-environment`.
- Task 10: `over` is read from this dispatch's manifest, never the committed matrix JSON; a composed frame with no manifest leaves step 10 undone.
- Task 13: a run with any dispatch bucket lacking `cost_usd` is out of both main figures, never priced as zero; each run's main figure is floored at zero before summing.

- [x] 1. Record the inline verify row's HEAD and judge green by it
  - [x] **Step 1: Write the tests.** `stats/internal/guard/run_verify_list_test.go`: `TestRunVerifyListRecordsHead` — with the recording arguments, the fake `flow`'s logged `record dispatch begin` argv carries `-diff-base <sha>`, the sha `git -C <worktree> rev-parse HEAD` prints for the verified `<worktree>` (the first argument, never the canonical one). `stats/internal/guard/check_verify_green_test.go`: `TestCheckVerifyGreenNeedsAHead` — a completed `verify` row with an empty `diffBase` exits 1, its `VERIFY-NOT-GREEN:` line naming the key and `no HEAD recorded`; `TestCheckVerifyGreenSpansPlanningCommits` — a row at commit A, then one commit touching only `spectre/changes/demo/tasks.md`, exits 0 and its `VERIFY-GREEN:` line names `verify@<A's short sha>`; `TestCheckVerifyGreenRefusesACodeCommitSinceTheRow` — a commit touching `src/a.go` after A exits 1 naming that commit; `TestCheckVerifyGreenJudgesEachWorktree` — called `<wt-a> demo <token> <wt-b>`, rows `verify-a` (diffBase in a) and `verify-b` (diffBase in b) each judged in the worktree their sha resolves in, exit 0; a diffBase resolving in neither worktree exits 1. Run `cd stats && go test ./internal/guard -run 'TestRunVerifyListRecordsHead|TestCheckVerifyGreen' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/internal/guard/runverifylist.go`: the `record` closure's `begin` appends `-diff-base` with the HEAD of `args[0]`, resolved through `envGit`. `stats/internal/guard/verifygreen.go`: accept `<worktree> <change-name> <session-token> [<worktree>...]`; read `diffBase` from each row; keep the latest-row-per-group rule; a latest row is green only when its outcome is `completed` or `skipped`, its `diffBase` is non-empty and resolves (`cat-file -e <sha>^{commit}`) in a worktree of the arguments, and `git -C <that worktree> log --format=%h <sha>..HEAD -- . ':(exclude)<leaf>/changes/'` prints nothing, `<leaf>` from `specRootLeaf` — signal seven's exclusion in `unfinishedwork.go`. The `VERIFY-GREEN:` line lists each green row as `<key>@<short sha>`; a `VERIFY-NOT-GREEN:` line names each refused row and why (`no HEAD recorded`, `HEAD resolves in no worktree`, `<commit> since <short sha>`). Update the headers of `scripts/check-verify-green.sh` (usage, the HEAD rule, decision `green-verify-spans-planning-commits`) and `scripts/run-verify-list.sh` (the begin's `-diff-base`).
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` with the `log` emptiness check removed — `TestCheckVerifyGreenRefusesACodeCommitSinceTheRow` fails; restore.
  - [x] **Step 4: Verify.** `cd stats && go test ./internal/guard -run 'RunVerifyList|CheckVerifyGreen' -count=1` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `stats/internal/guard/verifygreen.go`, `stats/internal/guard/check_verify_green_test.go`, `scripts/check-verify-green.sh`, `scripts/run-verify-list.sh`
**Tests:** `TestRunVerifyListRecordsHead`, `TestCheckVerifyGreenNeedsAHead`, `TestCheckVerifyGreenSpansPlanningCommits`, `TestCheckVerifyGreenRefusesACodeCommitSinceTheRow`, `TestCheckVerifyGreenJudgesEachWorktree`
**Regression:** `TestCheckVerifyGreenRefusesACodeCommitSinceTheRow` fails — a verifier is dispatched on code `flow.verify` never ran green.
**Commit:** feat(guard): record the inline verify row's HEAD and count it green only while planning paths changed since
**After:** none

**Decision:** verify-row-records-head
**Decision:** green-verify-spans-planning-commits

- [x] 2. Retire the verifier's `verify` step and read a report's `## Report` alone
  - [x] **Step 1: Write the tests.** `stats/internal/guard/check_verify_report_test.go`: `TestCheckVerifyReportTakesNoScopedSpecs` — a third positional argument exits 2 with the usage line `check-verify-report.sh <report file> <motions named> [--design <design.md>]`; a report whose `steps:` line reads `4 done | 8 done | 9 done | 10 done | 11 done | motion n/a — no motions named` is complete, and one still carrying a `7` item is too (the item is ignored). Delete `TestCheckVerifyReportScopedSpecCount`. `stats/internal/guard/check_visual_verify_dispatched_test.go`: `TestCheckVisualVerifyDispatchedIgnoresRerunWording` — a completed verifier row at a later HEAD whose report's `- verify:` line reads `rerun of 2 failed specs` is counted as any other row; delete `TestCheckVisualVerifyDispatchedRefusesARerunAcrossHeads`. `TestCheckVerifyReportReadsReportAlone` — a `## Verdict` block above the `## Report` whose `- steps:` line reads every step `done` and whose `- verify:` line differs, over a `## Report` leaving step 10 undone, exits 1 on step 10. `stats/internal/guard/check_frame_list_test.go`: `TestCheckFrameListReadsReportAlone` — a `## Verdict` departure line `- <frame id>: …` does not stand in for that frame's missing per-frame line under `## Report`; the gap `no report line: <frame id>` is reported. Run `cd stats && go test ./internal/guard -run 'TestCheckVerifyReportTakesNoScopedSpecs|TestCheckVerifyReportReadsReportAlone|TestCheckFrameListReadsReportAlone|TestCheckVisualVerifyDispatchedIgnoresRerunWording' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/internal/guard/verifyreport.go`: drop `"7"` from `vrSteps`, the `scoped` argument, `vrMappedSpecs` and the step-7 status block; the usage line as above; every line it reads (`- verify:`, `- steps:` included) is read from the `## Report` heading to the next `## ` heading. `stats/internal/guard/framelist.go`: its `- spec:`, `- frames:` and per-frame lines are read from `## Report` alone the same way (decision `verdict-only-handback`). `stats/internal/guard/visualverifydispatched.go`: drop the rerun rule (the block opening `// A RERUN OF FAILED SPECS STANDS ONLY AT THE HEAD THE ROW BEFORE IT`), `vvdRerun`, and `vvdHeadBefore` when nothing else calls it. Headers: `scripts/check-verify-report.sh` (usage, the `<scoped specs>` paragraph and step 7 gone, decision `verifier-runs-no-verify`), `scripts/check-visual-verify-dispatched.sh` (the `rerun of` paragraph gone).
  - [x] **Step 3: Verify.** `git grep -n -F -e 'vvdRerun' -e 'scoped specs' -e 'mapped specs' -- stats/internal/guard scripts/check-verify-report.sh scripts/check-visual-verify-dispatched.sh` prints nothing; `cd stats && go test ./internal/guard -run 'CheckVerifyReport|CheckFrameList|CheckVisualVerifyDispatched' -count=1` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `stats/internal/guard/framelist.go`, `stats/internal/guard/check_frame_list_test.go`, `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-verify-report.sh`, `scripts/check-visual-verify-dispatched.sh`
**Allowed-collateral:** `scripts/check-frame-list.sh`
**Tests:** `TestCheckVerifyReportTakesNoScopedSpecs`, `TestCheckVerifyReportReadsReportAlone`, `TestCheckFrameListReadsReportAlone`, `TestCheckVisualVerifyDispatchedIgnoresRerunWording`
**Regression:** `TestCheckVerifyReportTakesNoScopedSpecs` fails — the report guard still demands a step the verifier no longer runs.
**Commit:** feat(guard): retire the verifier's verify step and parse only a report's Report section
**After:** none

**Decision:** verifier-runs-no-verify
**Decision:** verdict-only-handback

- [x] 3. `visual-verify-pass prepare`: the stage begin and steps 1–3
  - [x] **Step 1: Write the tests.** `stats/internal/guard/visual_verify_pass_prepare_test.go`, on fixture worktrees carrying a `.flow/project.md` `## visual verification` section and a fake `flow` logging its argv: `TestVisualVerifyPassUsage` — no subcommand, an unknown one, a missing `--session-token`, a token not starting `mf-`, a missing `--harness`, a `--merge-base` not shaped `<worktree>=<sha>` and a `--app` not shaped `<worktree>:<app-root>=<url>` each exit 2 with a usage line, and nothing is recorded. `TestVisualVerifyPassPrepareKickoff` — a declared `regression checkout` whose repository the set holds no worktree in prints `STOP: kickoff <repo>` and exits 1 before any `flow` call. `TestVisualVerifyPassPrepareSkips` — `decision.json` recording `visual.verify` `skipped` ends the stage `skipped` and prints `NEXT: skip — skipped at Decide — <reason>`; no worktree surviving step 2 ends it `completed` with `NEXT: skip — <why>`; each exits 0 and begins no row. `TestVisualVerifyPassPrepareComposeOnly` — a `--round-base` diff touching only the mockups root prints `NEXT: compose-only`. `TestVisualVerifyPassPrepareStopsOnVerifyNotGreen` — `check-verify-green` exit 1 ends the stage `stopped` and prints `STOP: verify-not-green`, exit 1. `TestVisualVerifyPassPrepareStopsOnPreflight` — a failing pre-flight and a failing `check-sidecar-reach` end the stage `stopped`, one `STOP:` line naming both checks with their evidence, exit 1, no row begun. `TestVisualVerifyPassPrepareMissStopsForAnalyst` — `--miss` with every check passing prints `NEXT: analyst`, exit 0, no row begun. `TestVisualVerifyPassPrepareWritesStageFile` — the stage file holds the token, the begin instant, the set's size, and one line per surviving worktree with its repository name, merge base and round base. Run `cd stats && go test ./internal/guard -run 'TestVisualVerifyPass' -count=1` — fails to compile.
  - [x] **Step 2: Implement the entry.** `stats/internal/guard/visualverifypass.go`: `Registry["visual-verify-pass"]`, dispatching `prepare` (and later `close`, `measure`) on `integratechange.go`'s pattern — a state struct with `say`, `stop`, `markBegin`, `markEnd` and one closing `NEXT:` or `STOP:` line; exit 0 on `NEXT:`, 1 on `STOP:`, 2 on usage or an unreadable input. `scripts/visual-verify-pass.sh`: the `flow_guard_exec` shim exporting `FLOW_GUARD_REPO_ROOT` from `flow_guard_root` and `FLOW_GUARD_WORKSPACE_ISOLATION="$SCRIPT_DIR/check-workspace-isolation.sh"` (as `scripts/prepare-workspace.sh` does), its header the canonical contract for every argument, step and closing line of all three subcommands, written for `prepare` here and extended by Tasks 4, 5 and 9.
  - [x] **Step 3: Implement `prepare` through step 3.** `stats/internal/guard/visualverifypass_prepare.go`: `prepare <canonical-worktree> <name> <key> --harness <h> --session-token mf-<t> --merge-base <wt>=<sha>... [--worktree <wt>]... [--since <wt>=<sha>]... [--round-base <wt>=<sha>]... [--app <wt>:<root>=<url>]... [--miss] [--continue [--started]]`. In order: the kickoff check (the regression-checkout and flowBase cases of step 3 in the patch's `skills/flow/visual-verify.md`; `STOP: kickoff <repo>` before anything is begun); `flow stage begin -command '/flow' -stage flow.visual-verify -harness <h> -session-token <t> <name>`; for each worktree, its `git diff --name-only <since or merge-base>...HEAD` piped through `checkVisualTrigger` in-process via `Env.Stdin` (`--repo <repository name>` for a non-canonical worktree, its paths prefixed with that name, as the patch's step 1–2 text states); the `decision.json` skipped check; `--round-base` diffs: all empty → `NEXT: skip — round diff empty`, only the mockups root → `NEXT: compose-only`; then, once per stage, `checkVerifyGreen` with the canonical worktree, name, token and the set's other worktrees; per survivor `prepareWorkspace` piped into `checkVisualPreflight` with that worktree's `--app` pairs; with `mockups` declared `checkSidecarReach`. Every failing check (exit 1, or exit 2 with its stderr) is collected, then the stage ends `stopped` and one `STOP:` line names them; `check-verify-green`'s exit 1 alone prints `STOP: verify-not-green`. `--miss` then prints `NEXT: analyst`. The stage file `<canonical-worktree>/.superpowers/sdd/visual-verify-stage.txt` is written once the stage has begun, and is the only state a later call reads.
  - [x] **Step 4: Prove.** `scripts/break-and-prove.sh` with the stage `end` on a failing check removed — `TestVisualVerifyPassPrepareStopsOnPreflight` fails; restore.
  - [x] **Step 5: Verify.** `cd stats && go test ./internal/guard -run 'VisualVerifyPass|CheckVisualTrigger|CheckVerifyGreen|CheckVisualPreflight' -count=1` passes; `scripts/test-go-guards.sh` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/visualverifypass.go`, `stats/internal/guard/visualverifypass_prepare.go`, `stats/internal/guard/visual_verify_pass_prepare_test.go`, `scripts/visual-verify-pass.sh`
**Tests:** `TestVisualVerifyPassUsage`, `TestVisualVerifyPassPrepareKickoff`, `TestVisualVerifyPassPrepareSkips`, `TestVisualVerifyPassPrepareComposeOnly`, `TestVisualVerifyPassPrepareStopsOnVerifyNotGreen`, `TestVisualVerifyPassPrepareStopsOnPreflight`, `TestVisualVerifyPassPrepareMissStopsForAnalyst`, `TestVisualVerifyPassPrepareWritesStageFile`
**Regression:** `TestVisualVerifyPassPrepareStopsOnPreflight` fails — a verifier is dispatched into an environment the pre-flight proved cannot pass.
**Commit:** feat(guard): add visual-verify-pass prepare through the pre-flight
**After:** Task 1

**Decision:** thin-parent-prepare-close
**Decision:** parent-passes-apps
**Decision:** stage-file-names-its-rows
**Decision:** analyst-after-preflight

- [x] 4. `visual-verify-pass prepare`: probe, fingerprint, the verifier's bundle and its row
  - [x] **Step 1: Write the tests.** In `stats/internal/guard/visual_verify_pass_prepare_test.go`, with an `httptest` server standing in for each app: `TestVisualVerifyPassPrepareStartStack` — no `--app` URL answering prints `NEXT: start-stack`, exit 0, no row begun. `TestVisualVerifyPassPrepareRestartStack` — the fingerprint's exit 1 prints `NEXT: restart-stack`; the same under `--continue --started` ends the stage `stopped` with a `STOP: fingerprint — <the guard's output>` line, exit 1; the fingerprint's exit 2 continues, the bundle's `fingerprint:` line reading `not declared — <its stderr reason>`. `TestVisualVerifyPassPrepareBundle` — `<worktree>/.superpowers/sdd/verify-prompt-<key>.md` states the absolute worktree, the `KEY=value` lines `prepare-workspace` printed, the resolved `setup`, `capture` and `specs` commands and `screenshots` root, the `mockups` root and `mockup frame` when declared, each `--app` URL, the `## run` commands, the `stack:` and `fingerprint:` results, the green inline verify row by key and HEAD, `<changeRoot>`, the relay contract with this key's report path resolved, the path of `skills/flow/visual-verify-verifier.md`, and `mockup sources:` with every HTML or CSS path `design.md` cites that exists (or `none`). `TestVisualVerifyPassPrepareDecodesBundledPage` — a cited page carrying `<script type="__bundler/template">` is decoded into `<worktree>/.superpowers/sdd/verify-mockup-<key>/` (`index.html` plus each manifest entry, gzip-decompressed when `compressed`), the bundle naming that directory. `TestVisualVerifyPassPrepareBeginsRows` — one `flow record dispatch begin -role verifier -model opus -effort low -key <key> -diff-base <HEAD>[,<sibling HEAD>…]` per survivor, `-round-base <sha>` when passed, keys suffixed `-<repository name>` when the set holds more than one worktree, the stage file gaining `row <key> <worktree>` per row; stdout one `DISPATCH: <key> <bundle>` line per row, then `NEXT: dispatch <n>`. `TestVisualVerifyPassPrepareContinue` — `--continue` begins no stage and runs no step-3 check; a `-rest` key's bundle carries `## Incomplete attempt` with the incomplete report's path and the guard verdict lines the stage file holds for it (`verdict <key> <line>`, which Task 5's `close` writes), a `visual-verify-<n>` key's bundle `## Previous attempt` with every earlier report's path of that worktree; a key ending in a survivor's `-<repository name>` runs that worktree alone. Run — fails.
  - [x] **Step 2: Implement.** In `stats/internal/guard/visualverifypass_prepare.go`: step 5's probe of each surviving worktree's `--app` URLs (a short-timeout `GET`; any answer counts); step 6 through `checkDevStackFresh`; the row resolution through the `## visual verification` readers in `visualsection.go` and `runverifylist.go`'s `<frontend-root>`/`<merge-base>` substitution; mockup sources through `vrMockupSource` over `design.md`; the bundled-page decode in Go (the patch's `skills/flow/visual-verify.md` removes the inline Python it replaces); the relay contract read from `$FLOW_GUARD_REPO_ROOT/skills/flow/visual-verify.md`, the paragraph opening `**The relay contract.**`, with `<abs-worktree>/.superpowers/sdd/verify-report-<key>.md` resolved — exit 2 when the paragraph is missing; the green row from `flow record dispatches` (key and `diffBase`); each row begun through `pcFlow`. `--continue` reads the survivors from the stage file and its own `--app` and `--round-base` flags. Extend the shim's header.
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` with `-effort low` removed from the begin — `TestVisualVerifyPassPrepareBeginsRows` fails; restore.
  - [x] **Step 4: Verify.** `cd stats && go test ./internal/guard -run 'VisualVerifyPass' -count=1` passes; `scripts/test-go-guards.sh` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/visualverifypass_prepare.go`, `stats/internal/guard/visual_verify_pass_prepare_test.go`, `scripts/visual-verify-pass.sh`, `stats/internal/guard/visualverifypass.go`
**Tests:** `TestVisualVerifyPassPrepareStartStack`, `TestVisualVerifyPassPrepareRestartStack`, `TestVisualVerifyPassPrepareBundle`, `TestVisualVerifyPassPrepareDecodesBundledPage`, `TestVisualVerifyPassPrepareBeginsRows`, `TestVisualVerifyPassPrepareContinue`
**Regression:** `TestVisualVerifyPassPrepareBeginsRows` fails — a verifier runs with no row, or off its opus/low pair.
**Commit:** feat(guard): resolve the verifier's bundle and begin its row in visual-verify-pass prepare
**After:** Task 3

**Decision:** parent-starts-the-stack
**Decision:** prepare-continue-flag
**Decision:** redispatch-bundle-by-path
**Decision:** verifier-stays-opus

- [x] 5. `visual-verify-pass close`: report guards, row end, step 12 and the stage end
  - [x] **Step 1: Write the tests.** `stats/internal/guard/visual_verify_pass_close_test.go`, each on a stage file from a `prepare` fixture and a report written beside it: `TestVisualVerifyPassCloseDone` — a complete report, every exit 0, no departure: row ended `completed`, step 12's commit made, `ROUND-BASE: <worktree>=<sha>`, the stage ended `completed` once every survivor's newest row is closed, `flow record render -change <name> -kind ledger -repo <canonical>` run, `STAGE: completed`, `NEXT: done`, exit 0. `TestVisualVerifyPassCloseRest` — `check-verify-report` exit 1 (or `check-frame-list` exit 1 other than a `no sidecar line` gap) ends the row `aborted`, appends the guards' verdict lines to the stage file as `verdict <key> <line>` for the remainder's bundle, prints `NEXT: rest <key>` with `-rest` then `-rest-<n>` placed before any `-<repository name>` suffix, makes no commit and leaves the stage open. `TestVisualVerifyPassCloseRedispatch` — the stage's first complete report carrying a non-zero exit ends the row `blocked -cause <its cause>` and prints `NEXT: redispatch visual-verify-2[-<repo>]`, no commit. `TestVisualVerifyPassCloseLaterBlocks` — from the second such report, cause `environment` prints `NEXT: cure-environment <next key>` with no commit; `test-failure` or `missing-fixture` prints `NEXT: loop`, commits, and ends the stage `stopped`. `TestVisualVerifyPassCloseJudge` — a complete report naming departures, or a `no sidecar line` gap, prints its `## Verdict` and `NEXT: judge`, commits, and leaves the row and the stage open. `TestVisualVerifyPassCloseNoReport` — no report file: row `aborted`, stage ended `stopped`, `STOP: no report`, exit 1. `TestVisualVerifyPassCloseCommitsPathspec` — the commit carries exactly the report's `spec:` path, every PNG and the full-app-suite spec and zip its lines name, and `<changeRoot>/visual-verification/`, with subject `test: record the <key> visual verification` and trailer `Flow-Key: flow.visual-verify`, leaving a pre-staged foreign file staged and uncommitted; with a declared `regression checkout` the capture paths commit in its worktree on `spectre/<name>` (or its main checkout, printing `PUSH: git -C <checkout> push`) and `visual-verification/` in the canonical worktree; nothing to commit makes no commit and still prints `ROUND-BASE:`. Run — fails.
  - [x] **Step 2: Implement.** `stats/internal/guard/visualverifypass_close.go`: `close <canonical-worktree> <name> <key> <motions named> --session-token mf-<t>`; the report `<worktree>/.superpowers/sdd/verify-report-<key>.md`, its worktree from the stage file's `row` line; `checkVerifyReport` with `--design <changeRoot>/design.md` and, with `mockups` declared, `checkFrameList`, in-process; the `## Verdict` block read for exits, cause and departures; outcomes, commits and the stage end exactly as Step 1 states; `flow record dispatch end` and `flow stage end -command '/flow' -stage flow.visual-verify` through `pcFlow`. Extend the shim's header with `close`'s outcomes and decision `close-commits-final-reports`.
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` with the `rest`/`redispatch`/`cure-environment` commit exclusion removed — `TestVisualVerifyPassCloseRest` fails; restore.
  - [x] **Step 4: Verify.** `cd stats && go test ./internal/guard -run 'VisualVerifyPass' -count=1` passes; `scripts/test-go-guards.sh` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/visualverifypass.go`, `stats/internal/guard/visualverifypass_close.go`, `stats/internal/guard/visual_verify_pass_close_test.go`, `scripts/visual-verify-pass.sh`
**Tests:** `TestVisualVerifyPassCloseDone`, `TestVisualVerifyPassCloseRest`, `TestVisualVerifyPassCloseRedispatch`, `TestVisualVerifyPassCloseLaterBlocks`, `TestVisualVerifyPassCloseJudge`, `TestVisualVerifyPassCloseNoReport`, `TestVisualVerifyPassCloseCommitsPathspec`
**Regression:** `TestVisualVerifyPassCloseRest` fails — an incomplete report is accepted or committed before its remainder completes it.
**Commit:** feat(guard): close a visual-verify pass from its report with visual-verify-pass close
**After:** Task 4

**Decision:** close-commits-final-reports
**Decision:** verdict-only-handback
**Decision:** stage-file-names-its-rows

- [x] 6. Thin the parent's visual-verify contract and retire the verifier's `verify`
  - [x] **Step 1: Apply.** Save `scripts/check-normative-inventory.sh`'s output to a file outside the tree; then `git apply spectre/changes/kan-1060-flow-cost-visual-verify-now-costs-more-than/contract-edits/thin-parent-and-retired-verify.patch` (on a conflict, `git apply --3way` and resolve to the patch's text); `ln -s ../../../scripts/visual-verify-pass.sh skills/flow/scripts/visual-verify-pass.sh`; `git rm skills/flow/scripts/check-visual-preflight.sh` — the patched contracts name `check-visual-preflight.sh` only inside `prepare`'s description, so `check-guard-symlinks.sh` no longer requires the link.
  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-installed-citations.sh` and `scripts/check-ask-silence.sh` exit 0; `scripts/check-normative-inventory.sh` prints the inventory saved in Step 1; `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`, `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`, `skills/flow/verify-fix-loop.md`, `skills/flow/SKILL-rationale.md`, `skills/flow/scripts/visual-verify-pass.sh`, `skills/flow/scripts/check-visual-preflight.sh`
**Tests:** **none** — contract prose; Step 2's guards are the check.
**Regression:** none — prose.
**Commit:** docs(flow): run visual-verify's parent steps through visual-verify-pass and retire the verifier's verify
**After:** Task 1, 2, 5

**Decision:** thin-parent-prepare-close
**Decision:** verdict-only-handback
**Decision:** verifier-runs-no-verify

- [x] 7. Hold `visual-verify-pass` calls to the stage-mark token and harness rules
  - [x] **Step 1: Write the test.** `stats/internal/guard/check_stage_mark_calls_test.go`: `TestCheckStageMarkCallsVisualVerifyPass` — in a scanned skill file, a `visual-verify-pass.sh prepare` line with no `--session-token`, with `--session-token "mf-$(date +%s)"`, with no `--harness`, or with `--harness claude-code` is one finding each; `--harness <harness> --session-token mf-<literal-token>` is clean; a `visual-verify-pass.sh close` line needs the literal token and no harness; a `visual-verify-pass.sh measure` line is not a call. Run — fails.
  - [x] **Step 2: Implement.** `stats/internal/guard/stagemarkcalls.go`: a `visual-verify-pass.sh prepare` line takes `integrate-change.sh`'s token and harness rules (two-dash flags folded as `smcAssemble` folds them), a `close` line the token rules alone, as `flow record dispatch`'s; `visual-verify.md` joins `smcCandidates`, its one call being `close`. Update `scripts/check-stage-mark-calls.sh`'s header.
  - [x] **Step 3: Verify.** `cd stats && go test ./internal/guard -run 'CheckStageMarkCalls' -count=1` passes; `scripts/check-stage-mark-calls.sh` exits 0 on the tree; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/stagemarkcalls.go`, `stats/internal/guard/check_stage_mark_calls_test.go`, `scripts/check-stage-mark-calls.sh`
**Tests:** `TestCheckStageMarkCallsVisualVerifyPass`
**Regression:** `TestCheckStageMarkCallsVisualVerifyPass` fails — a substituted token in `prepare`'s call binds the stage to no session.
**Commit:** feat(guard): hold visual-verify-pass calls to the stage-mark token and harness rules
**After:** Task 6

**Decision:** thin-parent-prepare-close

- [x] 8. List every over-tolerance element in the matrix measurement
  - [x] **Step 1: Write the test.** `stats/internal/guard/measure_visual_properties_test.go`: `TestMeasureVisualPropertiesMatrixOver` — on two fixture images, `--props matrix`'s `summary.over` lists, by 1-based `row`, each matrix row that is `missing` or `extra`, carries `unresolved`, or is paired with a scaled numeric leaf of its pairing or `measured.delta` over `mvpOverLength` or a colour distance over `mvpOverColour`, each entry `{row, status, keys}` naming the leaves over; a row whose only deltas sit on `mvpUnscaled` keys, and a paired row within both tolerances, is absent. Run — fails.
  - [x] **Step 2: Implement.** `stats/internal/guard/measurevisualproperties.go`: constants `mvpOverLength = 2` and `mvpOverColour = 12` (step 10's band and seam tolerances, which the patched verifier cites); `mvpMatrix` builds `over` beside `paired`, `missing` and `extra`. Update `scripts/measure-visual-properties.sh`'s header (`summary.over`, decision `over-list`).
  - [x] **Step 3: Verify.** `cd stats && go test ./internal/guard -run 'MeasureVisualProperties' -count=1` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/measurevisualproperties.go`, `stats/internal/guard/measure_visual_properties_test.go`, `scripts/measure-visual-properties.sh`
**Tests:** `TestMeasureVisualPropertiesMatrixOver`
**Regression:** `TestMeasureVisualPropertiesMatrixOver` fails — a 3px offset the matrix measured never reaches the report.
**Commit:** feat(guard): list every over-tolerance element in the matrix measurement
**After:** none

**Decision:** over-list

- [x] 9. `visual-verify-pass measure`: one call for the measuring chain
  - [x] **Step 1: Write the tests.** `stats/internal/guard/visual_verify_pass_measure_test.go`, on a fixture worktree with a spec, its captures, a `<spec>.mockups` sidecar and frames: `TestVisualVerifyPassMeasureManifest` — `measure <worktree> <name> <key> <spec>` writes `<changeRoot>/visual-verification/matrix/<frame id>.json` per frame, the bands/seams JSON under `<worktree>/.superpowers/sdd/visual-measure-<key>/`, and `<worktree>/.superpowers/sdd/visual-manifest-<key>.json` holding `pairs[]` of `{frame, capture, mockup, composite, diff, exits{resolve, identical, compose, bands, matrix}, bands, seams, over}` — flagged bands and seams only, `over` copied from the matrix summary — and exits 0. `TestVisualVerifyPassMeasureCarriesOn` — a frame whose compose fails still leaves every other pair measured, the manifest records the non-zero exit, exit 1. `TestVisualVerifyPassMeasureUsage` — a missing argument or an unwritable output directory exits 2. `--prototype` composes against `<changeRoot>/visual-verification/prototype/`. Run — fails.
  - [x] **Step 2: Implement.** `stats/internal/guard/visualverifypass_measure.go`, its `case` in `visualverifypass.go`: `resolveVisualScreenshots`, `checkIdenticalBaselines`, `composeMockupFrames`, and `measureVisualProperties` with `--props bands,seams` and `--props matrix`, `--scale 1`, in-process on every pair, carrying on past a non-zero exit. Extend the shim's header with `measure` and decision `measure-one-call`.
  - [x] **Step 3: Verify.** `cd stats && go test ./internal/guard -run 'VisualVerifyPass' -count=1` passes; `scripts/test-go-guards.sh` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/visualverifypass.go`, `stats/internal/guard/visualverifypass_measure.go`, `stats/internal/guard/visual_verify_pass_measure_test.go`, `scripts/visual-verify-pass.sh`
**Tests:** `TestVisualVerifyPassMeasureManifest`, `TestVisualVerifyPassMeasureCarriesOn`, `TestVisualVerifyPassMeasureUsage`
**Regression:** `TestVisualVerifyPassMeasureCarriesOn` fails — one failed compose hides every other frame's measurement.
**Commit:** feat(guard): run the verifier's measuring chain as one visual-verify-pass measure call
**After:** Task 5, 8

**Decision:** measure-one-call

- [ ] 10. Check every flagged element and the frame/matrix completeness in the report guard
  - [ ] **Step 1: Write the tests.** `stats/internal/guard/check_verify_report_test.go`, each report beside a `visual-manifest-<key>.json`: `TestCheckVerifyReportOverToleranceLine` — per composed frame, `- <frame id> over-tolerance: <n> — #<row>: <cause>; …` must exist, `<n>` equal the manifest's `over` count for the frame and every row be named; a missing line, a wrong count or an unnamed row leaves step 10 undone; the manifest of a `-rest[-<n>]` report is its key's with that segment removed; a composed frame and no manifest leaves step 10 undone. `TestCheckVerifyReportOverToleranceCauses` — `departure`, `annotation`, `transcription`, `text width`, `decision <ID>` with an `**ID:**` of `design.md`, `present at the merge base` with its merge-base evidence, and `mockup source` on an `unresolved` row are admitted; an unknown cause, an ID `design.md` lacks and `mockup source` on a paired row are refused. `TestCheckVerifyReportComposedFrameLines` — a composed frame missing its `bands:`, `seams:` or `matrix:` line leaves step 10 undone. `TestCheckVerifyReportMatrixCells` — in `visual-verification.md`, a matrix table (header's first cell opening `element`) with an empty cell, or an n/a (or `—`) cell in a column that applies to its row's class, is refused unless it reads `no mockup source`; per class, n/a stands only in `border`, `fill`, `radius`, `shadow` for a `text run`, nowhere for a `control`, `row` or `card`, and in `content`, `alignment`, `border`, `radius`, `shadow` for a `rule`, `divider` or `fill band`; an unknown class admits only `no mockup source`. A `control` row whose `border` cell reads `n/a — measure failed` (kan-938's shape) is refused, naming its line. `TestCheckVerifyReportSweepItems` — a `sweeps:` item reading neither `done` nor `n/a — <why>` leaves step 10 undone. Run — fails.
  - [ ] **Step 2: Implement.** `stats/internal/guard/verifyreport.go`: the manifest read, the over-tolerance line, the closed causes, the composed-frame lines, the matrix cells by row class, and the sweep items, each as a step-10 status. Update `scripts/check-verify-report.sh`'s header (decisions `over-tolerance-line`, `by-hand-check-automated`, `matrix-kind-opens-with-class`).
  - [ ] **Step 3: Prove.** `scripts/break-and-prove.sh` with the manifest count comparison removed — `TestCheckVerifyReportOverToleranceLine` fails; restore.
  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard -run 'CheckVerifyReport|VisualVerifyPass' -count=1` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`, `stats/internal/guard/visual_verify_pass_close_test.go`
**Tests:** `TestCheckVerifyReportOverToleranceLine`, `TestCheckVerifyReportOverToleranceCauses`, `TestCheckVerifyReportComposedFrameLines`, `TestCheckVerifyReportMatrixCells`, `TestCheckVerifyReportSweepItems`
**Regression:** `TestCheckVerifyReportOverToleranceLine` fails — a frame the matrix flagged is called a match with nothing explained.
**Commit:** feat(guard): reconcile the report with the measure manifest and check matrix cells by row class
**After:** Task 2, 9

**Decision:** over-tolerance-line
**Decision:** by-hand-check-automated
**Decision:** matrix-kind-opens-with-class

- [x] 11. Refuse an in-run fix task with no opener or a shared one (F24)
  - [x] **Step 1: Write the test.** `stats/internal/guard/check_plan_shape_test.go`: `TestCheckPlanShapeF24` — a task titled `In-run fix 1 — visual-verify: …` with no `**Opener:**` field is one F24 finding at its title line; two `In-run fix` tasks whose `**Opener:**` names the same key are one finding at the second; distinct openers are clean; `--disable F24` silences it. Run — fails.
  - [x] **Step 2: Implement.** `stats/internal/guard/planshape.go`: F24 on the `In-run fix <k> —` title (`psVisualFixRE`'s shape, any stage) and the task's `Opener` field from `tcfBodyFields`, its finding text citing `skills/flow/verify-fix-loop.md` step 1; the findings list in the file's header comment gains F24. Update `scripts/check-plan-shape.sh`'s header if it enumerates findings.
  - [x] **Step 3: Verify.** `cd stats && go test ./internal/guard -run 'CheckPlanShape' -count=1` passes; `scripts/check-plan-shape.sh` exits 0 on the tree; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`, `stats/internal/guard/planparse.go`, `stats/internal/guard/close_task_test.go`
**Allowed-collateral:** `scripts/check-plan-shape.sh`
**Tests:** `TestCheckPlanShapeF24`
**Regression:** `TestCheckPlanShapeF24` fails — one verifier report splits into a round per defect again.
**Commit:** feat(guard): refuse an in-run fix task with no opener or one another task shares
**After:** none

**Decision:** one-report-one-round-opener

- [ ] 12. Run the measuring chain as one call and explain every flagged element in the contracts
  - [ ] **Step 1: Apply.** Save `scripts/check-normative-inventory.sh`'s output to a file outside the tree; then `git apply spectre/changes/kan-1060-flow-cost-visual-verify-now-costs-more-than/contract-edits/measure-and-explain.patch` (on a conflict, `git apply --3way` and resolve to the patch's text).
  - [ ] **Step 2: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-installed-citations.sh`, `scripts/check-ask-silence.sh` and `scripts/check-stage-mark-calls.sh` exit 0; `scripts/check-normative-inventory.sh` prints the inventory saved in Step 1; `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`, `skills/flow/verify-fix-loop.md`
**Tests:** **none** — contract prose; Step 2's guards are the check.
**Regression:** none — prose.
**Commit:** docs(flow): measure in one call and explain every over-tolerance element in the report
**After:** Task 6, 7, 9, 10, 11

**Decision:** measure-one-call
**Decision:** over-tolerance-line
**Decision:** by-hand-check-automated
**Decision:** one-report-one-round-opener

- [x] 13. Compute each stage's main-session cost and share
  - [x] **Step 1: Write the tests.** `stats/internal/store/aggregate_test.go`: `TestStageLeaderboardMainFigures` — three runs of one stage: one costing 10 with dispatch buckets 3 and 2 (main 5), one costing 4 with none (main 4), one costing 6 with a bucket lacking `cost_usd` (left out of both main figures); `MainCostUSD` reads 9 and `MainSharePct` 64.29 (9 of 14) to two places; a run whose buckets sum past its cost contributes 0; a stage whose every run is left out reads `nil` in both. `TestStageLeaderboardMainFiguresUnderModelFilter` — with the model filter, a run's main figure is that model's `models.<m>.cost_usd` minus the buckets that model served, a bucket's served model its top `signals.served_models` entry by turns (ties by name), else its `model`. `stats/internal/api/stats_test.go`: `TestStageLeaderboardMainFields` — the route's rows carry `mainCostUsd` and `mainSharePct`, `null` when absent. Run `cd stats && go test ./internal/store ./internal/api -run 'StageLeaderboard' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/internal/store/aggregate.go`: `StageLeaderboardRow` gains `MainCostUSD, MainSharePct *float64`; the `costed` CTE gains per run the dispatch sum (every `metrics.dispatches.*.cost_usd`, under a model filter only buckets whose served model — `dispatchBucket.servedModel()`'s rule in `stats/internal/store/pricing.go`, written in SQL — equals it) and whether any bucket lacks `cost_usd`; `main = GREATEST(cost - dispatch sum, 0)`, NULL for a run with a bucket lacking `cost_usd`; the select sums `main` and divides it by the summed cost of the runs with a main figure, times 100, NULL when none has one. `stats/internal/api/stats.go`: the leaderboard DTO gains `MainCostUSD *float64 json:"mainCostUsd"` and `MainSharePct *float64 json:"mainSharePct"`.
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` with the `GREATEST(…, 0)` floor removed — `TestStageLeaderboardMainFigures` fails; restore.
  - [x] **Step 4: Verify.** The tests pass; lint. Commit.

**Build:** green
**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`
**Tests:** `TestStageLeaderboardMainFigures`, `TestStageLeaderboardMainFiguresUnderModelFilter`, `TestStageLeaderboardMainFields`
**Regression:** `TestStageLeaderboardMainFigures` fails — main's share is priced from an incomplete run or goes negative.
**Commit:** feat(stats): compute each leaderboard stage's main-session cost and share
**After:** none

**Decision:** leaderboard-main-columns

- [x] 14. Draw Main session and Main share on the Stage leaderboard
  - [x] **Step 1: Write the test.** `stats/web/src/views/views.test.tsx`, on the file's `stage-leaderboard` fixture envelope: `stage leaderboard main columns` — the headers read, in order, `Command`, `Stage`, `Runs`, `Total cost`, `Main session`, `Main share`, `Mean cost`, `Median cost`, `P90 cost`; a row with `mainCostUsd: 3.2, mainSharePct: 45.6` draws `formatUsd(3.2)` and `46%`; a row with both `null` draws `—` in both cells; clicking `Main share` sorts ascending and its header reads `Main share ▲`. Scope each text query to its row, since `—` and figures repeat. Run `cd stats/web && npx vitest run src/views/views.test.tsx -t 'stage leaderboard main columns'` — fails.
  - [x] **Step 2: Implement.** `stats/web/src/api.ts`: `StageLeaderboardRow` gains `mainCostUsd: number | null` and `mainSharePct: number | null`. `stats/web/src/views/StageLeaderboard.tsx`: after `Total cost`, the sortable columns `Main session` (`formatUsd`) and `Main share` (`${Math.round(v)}%`), each accessor returning the value or `null`, each render drawing `—` for `null`; the header comment names the two columns and decision `leaderboard-main-columns`. Add fields to the test fixture's rows.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && cd web && npx tsc -b && npx vitest run src/views/views.test.tsx -t 'stage leaderboard main columns'` passes. Commit.

**Build:** green
**Files:** `stats/web/src/api.ts`, `stats/web/src/views/StageLeaderboard.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `stage leaderboard main columns`
**Regression:** `stage leaderboard main columns` fails — the board the Done-when measures on has no main-session columns.
**Commit:** feat(web): draw Main session and Main share on the Stage leaderboard
**After:** Task 13

**Decision:** leaderboard-main-columns

- [x] 15. Capture the Stage leaderboard against the prototype, at rest and sorted by Main share
  - [x] **Step 1: Seed a split.** `stats/cmd/uitest-seed/seed.go`: give kan-102's pinned-period `flow.review-panel` run two priced dispatch buckets inside its cost (the KAN-1041 `pricedDispatch` shape, served by `claude-sonnet-5`), so its Main share reads below 100% while every other row stays as seeded; update `stats/cmd/uitest-seed/seed_test.go` if it pins that run's metrics.
  - [x] **Step 2: Capture.** `stats/web/tests/visual/views.spec.ts`: the existing `stage-leaderboard.png` capture asserts the `Main session` and `Main share` headers before its screenshot; a new test `stage leaderboard sorted by Main share` clicks the header, asserts it reads `Main share ▲`, parks the pointer with `page.mouse.move(0, 0)` and screenshots `stage-leaderboard-main-share-sorted.png` full page; the describe block's header comment counts eight sortable columns and names the seeded split. `stats/web/tests/visual/views.spec.ts.mockups` names both captures beside frames `stage-leaderboard-prototype` and `stage-leaderboard-prototype-sorted`, with a comment in `experiments.spec.ts.mockups`'s shape naming `.superpowers/sdd/2026-10-30-stage-leaderboard-main-share-prototype.html` rendered at the capture's 1280px viewport, dark, at rest and after one `Main share` click, into `.superpowers/sdd/leaderboard-frames/`. Run the project's `capture` command for `views.spec.ts`, render the two frames, then `compose-mockup-frames.sh` on the sidecar and `measure-visual-properties.sh` band, seam and matrix reads on each pair; fix until clean or each leftover is named with its cause — the sorted header's `▲` is decision `sort-indicator-kept`'s expected difference. Recapture every other spec whose PNG the seed moved.
  - [x] **Step 3: Verify.** The project's `verify` command passes. Commit.

**Build:** green
**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/views.spec.ts`, `stats/web/tests/visual/views.spec.ts.mockups`
**Allowed-collateral:** `stats/cmd/uitest-seed/seed_test.go`, `stats/web/tests/visual/*.spec.ts-snapshots/**`, `stats/web/tests/visual/full-app-suite.zip`, `stats/web/tests/visual/weekly-costs.spec.ts`, `stats/web/tests/visual/*.spec.ts`, `stats/web/src/**`, `stats/Makefile`
**Tests:** `views.spec.ts`
**Regression:** `views.spec.ts` fails — the leaderboard drifted from the approved prototype.
**Commit:** test(web): capture the Stage leaderboard's main-session columns against the prototype
**After:** Task 14

**Decision:** sort-indicator-kept
**Decision:** leaderboard-main-columns

Correction (2026-10-30): the plan declared only the leaderboard specs as collateral. The seeded `claude-sonnet-5` buckets add a sixth `sonnet-5 · medium` series to the Weekly costs chart, so `weekly-costs.spec.ts`'s model·effort assertions move with it (decision taken: keep the plan's served model and widen the collateral). The base was already red under `verify` — baseline, combinations, full-app-suite and interactive specs, and `stats/Makefile`'s `ui-test-up` lacking its `policy-copy` prerequisite — and a red check is fixed in the run, so those fixes ride this task's dispatch as their own `Flow-Key: full-suite-fix-<n>` commits.
