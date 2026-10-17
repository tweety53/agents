# self-review-kan-989-stats-ui-dashboard-review

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-17-self-review-kan-989-stats-ui-dashboard-review-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`. Never merge or push.
- Fixer dispatches run in series as `design.md`'s `first-review-effort` groups them: g1 Tasks 4, 3, 6; g2 Tasks 1, 9, 2, 7, 8; g3 Task 5 — each group's tasks in that order.
- Every guard and Go test run in this change exports `FLOW_GUARD_TELEMETRY=off` until Task 4 lands, so no run posts to the dev store on 127.0.0.1:4173.
- Never stop, restart or write to `flowd` on 127.0.0.1:4173 or its Postgres on 5433. Task 5's stack is the UI-test stack on 127.0.0.1:4174 alone.
- A step that applies a diff runs `git apply` from the worktree root; a diff that does not apply is reported, never hand-merged. Every prose sentence a diff adds is acknowledged by its exact text in this change's `verbatim-moves.txt`; any other wording fails `scripts/check-verbatim-moves.sh`.

## Review focus

- A guard call with no range in a repository without `origin` must behave exactly as today — no new exit 2 (Task 1).
- Commits touching only `spectre/changes/` after the last verify row must not make the change unfinished (Task 1).
- A `self-review-…-fix` or `pipeline-fix-…` implementer key must still be accepted by `dispatch begin` (Task 6).
- A `panel-fix-<n>` row of role `panel-fix` must still be accepted (Task 3).
- No re-captured baseline may show an empty table where its seeded view has rows (Task 5).

- [x] 1. The unfinished-work gate derives a missing range and sees skipped verification
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Add the cases of step 2 to `TestCheckUnfinishedWork` in `stats/internal/guard/check_unfinished_work_test.go` and run `cd stats && FLOW_GUARD_TELEMETRY=off go test ./internal/guard -run 'TestCheckUnfinishedWork' -count=1`: the derived-range case fails with exit 2 (`no <worktree>=<merge-base> range was passed`), the two `OUTSTANDING` verify and visual cases fail reading `CLEAR:`; the `CLEAR` cases pass.
  - [x] **Step 2: Write the cases.** Each fixture is `newUW` plus a real git repository at `f.wt` (as `uwCommits` builds one) whose commits touch the files named. `each(demo(…), …)` labels and assertions:
    - `a call with no range derives the worktree's own merge base` — `f.wt` has a bare `origin` whose `HEAD` is `main`, `origin/main` at the base commit, and one later commit `fix: the appended task` with trailer `Task-Id: 7` touching `src/a.go`; the plan's last task 7 is ticked and appended; dispatch rows name task 7 with a later `panel-1-primary` reviewer row and a pass row for round 1, and a `completed` `verify` row whose `startedAt` is after that commit; the guard runs with no range → `uwVerdict(…, "CLEAR:")`.
    - `a code commit after the last completed verify row is OUTSTANDING` — range passed; the newest commit touching `src/a.go` is authored after the only `completed` `verify` row's `startedAt` → `uwVerdict(…, "OUTSTANDING:")`.
    - `the breakdown names the code commit no verify row covers` — same scenario → `uwReason(…, "no completed inline verify row since")`.
    - `a completed verify row after the last code commit is CLEAR` — the row's `startedAt` after the commit → `uwVerdict(…, "CLEAR:")`.
    - `planning-only commits after the verify row are CLEAR` — the only commit after the row touches `spectre/changes/demo/tasks.md` alone → `uwVerdict(…, "CLEAR:")`.
    - `an archived change is not asked for a verify row` — the plan under `spectre/changes/archive/demo/`, a code commit after the last verify row → `uwVerdict(…, "CLEAR:")`.
    - `a ui-path commit no visual verifier saw is OUTSTANDING` — `f.wt/.flow/project.md` declares `## visual verification` with `ui paths` `src/web/`, a commit touching `src/web/App.tsx`, a later `completed` `verify` row, no `visual-verify` row → `uwVerdict(…, "OUTSTANDING:")`.
    - `the breakdown carries the visual guard's VISUAL-VERIFY-MISSING line` — same scenario → `uwReason(…, "VISUAL-VERIFY-MISSING")`.
  - [x] **Step 3: Implement in `stats/internal/guard/unfinishedwork.go`.** `func uwDerivedRange(env Env, worktree string) ([2]string, bool)` — `resolveBaseBranch([]string{worktree}, env, &out, io.Discard)` then `git -C <worktree> merge-base HEAD origin/<base>`; false on any failure. Called once, right after argument parsing, only when no range argument was passed; false leaves `ranges` empty (decision `uw-derive-missing-range`). The `dispatches` struct gains `StartedAt time.Time` (`json:"startedAt"`). Inside `if anchor == worktree`, after signal six: when `ranges` is non-empty and the plan dir is not under `<spec-root>/changes/archive/`, `git -C <r[0]> log --no-merges --format=%H %at <r[1]>..HEAD -- . ':(exclude)<spec-root>/changes/'` over every range gives the newest code commit; when its author time is after the newest `startedAt` among rows with role `verifier`, outcome `completed`, key `verify` or prefix `verify-` (none counts as never), append `no completed inline verify row since <short sha> — run Inline verify (skills/flow/verify-and-handoff.md) and record its row before landing` (decision `uw-verify-row-after-code`). Then, for the range whose worktree is this guard's own, run `checkVisualVerifyDispatched([]string{worktree, name, <its merge-base>, canonical}, env, &vOut, &vErr)`: rc 1 appends its stdout line, trimmed, to `reasons`; rc 2 writes its stderr behind `uwPrefix` and returns 2 (decision `uw-fold-visual-verdict`). Update the header comment of `scripts/check-unfinished-work.sh` to name the derived range, the inline-verify signal and the folded visual verdict.
  - [x] **Step 4: Apply the prose.**

````diff verified:git apply --check at 2c0befa3; with all five prose diffs applied in a scratch clone check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index 6d4d05ca..bbc03ab5 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -146,6 +146,9 @@ above does not read it raw. Pass `[canonical-worktree]` on every call in the run
 the resolved set whose own change directory (**A change's directory**,
 `skills/flow-contracts/pipeline.md`) holds `tasks.md`.
 Pass every worktree of the resolved set as `<worktree>=<merge-base>`, its merge base `git -C <worktree> merge-base HEAD origin/<base>`, on every call.
+A call passing none gets its own worktree's range derived the same way, from the base
+`resolve-base-branch.sh` names; the guard's header (`<agents repo>/scripts/check-unfinished-work.sh`)
+is canonical for the visual-verify and inline-verify signals it folds in.
 
 | Verdict | Meaning |
 |---------|---------|
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index b3f03676..64c97b12 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -409,6 +409,9 @@ this run's own unfinished work: finish what its breakdown names — a missing di
 run and recorded, never recorded after the fact for work that did not run — and run the guard
 again; the state is written only once every worktree prints `CLEAR:`. A missing verdict line or an
 exit 2 is reported and stops the run at the state it read.
+The guard folds in `check-visual-verify-dispatched.sh`'s verdict for its own worktree and finds a
+code commit no later completed `verify` row covers unfinished, so a skipped visual verification or
+an unrecorded inline verify stops this run too.
 
 **Every run then runs the visual guard on itself** — creating, resuming and fix runs alike — once
 per worktree of the resolved set, with the arguments integrate's gate passes it (**Run 1 — the
````
  - [x] **Step 5: Verify.** `cd stats && gofmt -l internal/guard && FLOW_GUARD_TELEMETRY=off go vet ./internal/guard && FLOW_GUARD_TELEMETRY=off go test ./internal/guard -run 'TestCheckUnfinishedWork|TestCheckVisualVerifyDispatched' -count=1` passes with every new case; `FLOW_GUARD_TELEMETRY=off scripts/check-verbatim-moves.sh`, `FLOW_GUARD_TELEMETRY=off scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`, `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/finish-contract-run1.md`, `scripts/test-integrate-change.sh`
**Tests:** `a call with no range derives the worktree's own merge base`, `a code commit after the last completed verify row is OUTSTANDING`, `the breakdown names the code commit no verify row covers`, `a completed verify row after the last code commit is CLEAR`, `planning-only commits after the verify row are CLEAR`, `an archived change is not asked for a verify row`, `a ui-path commit no visual verifier saw is OUTSTANDING`, `the breakdown carries the visual guard's VISUAL-VERIFY-MISSING line`
**Regression:** reverting the commit makes the first case exit 2 again, and the second, third, seventh and eighth read `CLEAR:` for a change that skipped inline or visual verification; the fourth, fifth and sixth pin that the new signals stay quiet where nothing is missing.
**Baseline:** before=101 after=109
<!-- measured: grep -c 'uwRC(\|uwVerdict(\|uwReason(' stats/internal/guard/check_unfinished_work_test.go @ 2c0befa3 (before; after adds this task's eight cases) -->
**Commit:** fix(guard): find skipped inline and visual verification in the unfinished-work gate
**After:** none

**Decision:** uw-derive-missing-range
**Decision:** uw-fold-visual-verdict
**Decision:** uw-verify-row-after-code

- [x] 2. Fix-run appends pass the plan guards before their planning commit
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `git show d25bd8fb:spectre/changes/kan-989-stats-ui-dashboard-review/tasks.md > "$TMPDIR/t.md" && FLOW_GUARD_TELEMETRY=off scripts/check-plan-shape.sh "$TMPDIR/t.md"; echo $?` prints the task 11 and 12 suite hits and `1`; `grep -c 'check-plan-shape' skills/flow/document-fix.md` prints `0`.
  - [x] **Step 2: Apply the edit.**

````diff verified:git apply --check at 2c0befa3; with all five prose diffs applied in a scratch clone check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt
diff --git a/skills/flow/document-fix.md b/skills/flow/document-fix.md
index 486fa880..6301b424 100644
--- a/skills/flow/document-fix.md
+++ b/skills/flow/document-fix.md
@@ -123,2 +123,7 @@ flow tasks count -C <worktree> <name>
 
+**The appended plan passes the plan guards before it is committed.** Run `check-plan-shape.sh`
+and `check-plan-provenance.sh` over the appended `tasks.md` and handle each exit as **D. Basic
+Workflow #3 — Writing plans** (`skills/flow/brainstorm-planner.md`) handles it, before the
+planning commit below.
+
 Then make the fix-run planning commit over the appended plan (**Planning commits**,
````
  - [x] **Step 3: Verify.** `FLOW_GUARD_TELEMETRY=off scripts/check-verbatim-moves.sh`, `FLOW_GUARD_TELEMETRY=off scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/document-fix.md`
**Tests:** **none** — one prose paragraph; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): run the plan guards over a fix run's appended plan before committing it
**After:** none

**Decision:** fix-append-runs-plan-guards

- [x] 3. Dispatch begin refuses a panel key with no numeric round
  - [x] **Step 1: Write the failing test.** `TestRecordDispatchBeginRefusesANonNumericPanelRound` in `stats/cmd/flow/record_test.go`: `flow record dispatch begin -role reviewer -key panel-fix1-primary …` exits 2 with stderr naming `panel-<round>-<slot>` and sends nothing to the fake daemon; `-role reviewer -key panel-2-primary` and `-role panel-fix -key panel-fix-2` are accepted.
  - [x] **Step 2: Run it to see it fail.** `cd stats && FLOW_GUARD_TELEMETRY=off go test ./cmd/flow -run 'TestRecordDispatchBeginRefusesANonNumericPanelRound' -count=1` fails: the bad key is recorded.
  - [x] **Step 3: Implement.** In `runRecordDispatchBegin` (`stats/cmd/flow/record.go`), after the role check: a `reviewer` row whose key starts `panel-` and for which `records.PanelDispatchRound(records.Dispatch{Key: *key, Role: *role})` is false exits 2 with `flow: -key %q is not panel-<round>-<slot>: a reviewer's panel key carries its numeric round`.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l cmd/flow && go vet ./cmd/flow && FLOW_GUARD_TELEMETRY=off go test ./cmd/flow -run 'TestRecordDispatch' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBeginRefusesANonNumericPanelRound`
**Regression:** reverting `record.go` records `panel-fix1-primary`, and the test fails on the exit code.
**Baseline:** before=108 after=109
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 2c0befa3 (before; after adds this task's one test) -->
**Commit:** fix(flow): refuse a reviewer panel key with no numeric round
**After:** none

**Decision:** dispatch-key-shapes

- [x] 4. Test packages turn guard telemetry off
  - [x] **Step 1: Write the failing test.** `TestChildEnvTurnsGuardTelemetryOff` in `stats/internal/guard/helpers_test.go`: `os.Getenv("FLOW_GUARD_TELEMETRY")` is `off`, and `exec.Command("sh", "-c", "printf %s \"$FLOW_GUARD_TELEMETRY\"")` with `Env = os.Environ()` prints `off`.
  - [x] **Step 2: Run it to see it fail.** `cd stats && go test ./internal/guard -run 'TestChildEnvTurnsGuardTelemetryOff' -count=1` fails reading an empty value.
  - [x] **Step 3: Implement.** `os.Setenv("FLOW_GUARD_TELEMETRY", "off")` as the first statement of `TestMain` in `stats/internal/guard/helpers_test.go` and in `stats/cmd/flow/journal_test.go`; add `"FLOW_GUARD_TELEMETRY=off"` to the hand-built `cmd.Env` lists in `stats/internal/guard/check_task_commit_fields_test.go`, `stats/internal/guard/mutate_and_verify_test.go` and `stats/internal/guard/runreproducer_test.go`. No stored row is deleted (decision `test-telemetry-off`).
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/guard cmd/flow && go vet ./internal/guard ./cmd/flow && go test ./internal/guard -run 'TestChildEnvTurnsGuardTelemetryOff' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/guard/helpers_test.go`, `stats/cmd/flow/journal_test.go`, `stats/internal/guard/check_task_commit_fields_test.go`, `stats/internal/guard/mutate_and_verify_test.go`, `stats/internal/guard/runreproducer_test.go`, `scripts/run-guard-tests.sh`, `scripts/test-run-guard-tests.sh`
**Tests:** `TestChildEnvTurnsGuardTelemetryOff`
**Regression:** reverting `helpers_test.go`'s `Setenv` makes the test read an empty value, and every guard the package execs posts a `guard_runs` row to 127.0.0.1:4173 again.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/helpers_test.go @ 2c0befa3 (before; after adds this task's one test) -->
**Commit:** test(guard): keep package tests from posting guard telemetry to the dev store
**After:** none

**Decision:** test-telemetry-off

- [x] 5. The visual suite renders deterministically and compares at zero tolerance
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `cd stats && make ui-test-up` (the stack on 127.0.0.1:4174 — never 4173). Delete the `expect` block of `stats/web/playwright.config.ts` and run `cd stats/web && npm run test:visual` three times on the unchanged tree; record each run's failing screenshots and their differing-pixel counts. That is the flake's measured rate at zero tolerance.
  - [x] **Step 2: Pin the rendering.** Add `launchOptions: { args: ["--font-render-hinting=none", "--disable-font-subpixel-positioning", "--disable-lcd-text", "--disable-gpu", "--force-color-profile=srgb"] }` to `use` and replace the deleted block's comment with one naming the flags as the determinism pin. <!-- unverified:that Playwright 1.62.1 passes use.launchOptions.args to Chromium — confirm through Context7's @playwright/test docs before relying on it -->
  - [x] **Step 3: Re-capture.** `cd stats/web && npx playwright test --update-snapshots`, then Read every changed PNG under `stats/web/tests/visual/*-snapshots/`; each seeded view shows its seeded rows — `reviewers-darwin.png` and `full-reviewers-darwin.png` show slot rows, not an empty table.
  - [x] **Step 4: Prove the pin.** Run `npm run test:visual` three times: all pass. Remove the `args` and run it three times more: a step 1 failure reappears. Restore the `args`. If step 1 saw no failure in its three runs, drop the `args` instead — the commit carries the tolerance removal and the re-captured baselines, its body stating the measured runs and failures (decision `playwright-deterministic-zero-tolerance`).
  - [x] **Step 5: Verify.** `cd stats/web && npx tsc -b` exits 0 and `npm run test:visual` passes; `cd stats && make ui-test-down`. Commit, the body carrying steps 1 and 4's measured figures.

**Build:** green
**Files:** `stats/web/playwright.config.ts`
**Allowed-collateral:** `stats/web/tests/visual/**/*.png` — the baselines step 3 re-captures
**Tests:** **none** — the existing visual specs are the tests; step 4's runs prove the pin.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** test(stats): pin Chromium rendering and compare visual baselines at zero tolerance
**After:** none

**Decision:** playwright-deterministic-zero-tolerance

- [x] 6. Dispatch begin refuses an off-convention implementer key
  - [x] **Step 1: Write the failing test.** `TestRecordDispatchBeginRefusesAnOffConventionImplementerKey` in `stats/cmd/flow/record_test.go`: `-role implementer` with key `task-11-12` or `task-12` exits 2 with stderr naming `task-<ids>-implementer` and sends nothing; `task-12-implementer`, `task-11+12-implementer`, `task-3+4-implementer-fix-1`, `self-review-demo-fix-g1` and `pipeline-fix-7-demo-fix` are accepted.
  - [x] **Step 2: Run it to see it fail.** `cd stats && FLOW_GUARD_TELEMETRY=off go test ./cmd/flow -run 'TestRecordDispatchBeginRefusesAnOffConventionImplementerKey' -count=1` fails: `task-11-12` is recorded.
  - [x] **Step 3: Implement.** In `runRecordDispatchBegin`, beside Task 3's check: an `implementer` key matching neither `^task-[0-9]+(\+[0-9]+)*-implementer(-fix-[0-9]+)?$` nor starting `self-review-` or `pipeline-fix-` exits 2 with `flow: -key %q is not task-<ids>-implementer[-fix-<k>]: ids are +-joined`.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l cmd/flow && go vet ./cmd/flow && FLOW_GUARD_TELEMETRY=off go test ./cmd/flow -run 'TestRecordDispatch' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBeginRefusesAnOffConventionImplementerKey`
**Regression:** reverting the check records `task-11-12`, which `uwTaskKey` reads as task 11 alone, and the test fails on the exit code.
**Baseline:** before=109 after=110
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 2c0befa3 (108; before counts Task 3's one test, after adds this task's one) -->
**Commit:** fix(flow): refuse an implementer dispatch key off the task-ids convention
**After:** Task 3

**Decision:** dispatch-key-shapes

- [x] 7. The in-run fix loop closes the failing verify row first
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `flow record dispatches -change kan-989-stats-ui-dashboard-review -C .` shows row 21 `verify` begun 11:56 and closed `blocked`, row 25 `verify-fix-1` begun 12:14; `grep -n 'outcome blocked' skills/flow/verify-fix-loop.md` prints nothing.
  - [x] **Step 2: Apply the edit.**

````diff verified:git apply --check at 2c0befa3; with all five prose diffs applied in a scratch clone check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index 0ed54c72..d3433521 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -26,7 +26,9 @@ One round runs these steps in order:
    states for a found defect, and its title reads `In-run fix <k> — <stage>: <the defects>`, `<k>`
    this stage's round in this run. Its planning commit and the re-decide run as that file states.
    **One round appends one task**, carrying every defect the report names — never a task or a
-   round per defect.
+   round per defect. Before the append, the stage's failing row is closed with `flow record dispatch
+   end` at `-outcome blocked -cause test-failure`, so no row of the stage stays open while this
+   round's rows open.
 2. **Implement it.** **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) executes the appended
    task exactly as a plan-time task: the decision's implementer pair, `close-task.sh` with its
    guard and its review gate. The implementer's dispatch also carries the MUTATION PROOF
````
  - [x] **Step 3: Verify.** `FLOW_GUARD_TELEMETRY=off scripts/check-verbatim-moves.sh`, `FLOW_GUARD_TELEMETRY=off scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — one prose sentence; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): close the failing verify row before the in-run fix loop appends
**After:** none

**Decision:** loop-closes-failing-row

- [x] 8. The visual verifier reports a denied command as blocked
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'denie' skills/flow/visual-verify-verifier.md` prints nothing: the template offers `exit <n>` alone.
  - [x] **Step 2: Apply the edit.**

````diff verified:git apply --check at 2c0befa3; with all five prose diffs applied in a scratch clone check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index a9e27b99..4291b768 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -4,6 +4,10 @@ Read by the verifier **Steps 3–13** (`skills/flow/visual-verify.md`) dispatche
 path its prompt carries — the parent never loads it whole. It carries step 4, steps 7–11, the
 motion step and the `## Report` template, run as written.
 
+**A command the harness denies has no exit code.** Its report line reads `blocked — environment:
+<the denial text, verbatim>` where `exit <n>` would stand, and the step's status is `blocked` —
+never an exit code the command did not return.
+
 4. **Run `setup`, if declared.** A non-zero exit blocks, printing the command verbatim.
 7. **Run `verify`.** When `specs` is declared, run it first from the `regression checkout`, with
    `<frontend-root>` substituted for the worktree and `<merge-base>` for its merge base — or for
````
  - [x] **Step 3: Verify.** `FLOW_GUARD_TELEMETRY=off scripts/check-verbatim-moves.sh`, `FLOW_GUARD_TELEMETRY=off scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — one prose paragraph; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): report a harness-denied verifier command as blocked with its denial text
**After:** none

**Decision:** denied-command-blocked

- [x] 9. Inline verify runs the lint and test lists as two concurrent groups
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `flow suite list | head -3` shows stats-go near two minutes; `scripts/project-get.sh . lint | wc -l` counts the lint list; `grep -n 'Run$' skills/flow/verify-and-handoff.md` finds the serial-list sentence.
  - [x] **Step 2: Apply the edit**, on top of Task 1's.

````diff verified:git apply --check at 2c0befa3 with Task 1's diff applied; with all five prose diffs applied in a scratch clone check-verbatim-moves.sh exits 0 against this change's verbatim-moves.txt
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 64c97b12..5048f1de 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -84,7 +84,11 @@ Export the `KEY=value` lines `prepare-workspace.sh` printed for that worktree, t
 commands, then the test commands, in the order printed, then `check-spec-reach.sh <worktree>` —
 one more command in the same list, whose exit 0 line `Spec reach: not configured` is the ordinary
 case for a project with no `regression checkout` (its header is canonical for its exit codes). Run
-every command in order and do not stop at the first failure. **Nothing runs them later** —
+every command in order and do not stop at the first failure.
+The lint commands with `check-spec-reach.sh` and the test commands run as two groups at once, each
+in the order printed, each started detached and polled to completion as step 7 of the verifier's
+steps (`skills/flow/visual-verify-verifier.md`) runs a `verify` that can outrun the shell tool's
+time limit; the `## Report` lists every command in the order printed. **Nothing runs them later** —
 `/flow`'s integrate phase has no verification gate — so a non-zero exit blocks this handoff.
 While a stack is up for this run, the test commands start only after `check-dev-stack-fresh.sh
 <worktree>` exits 0 or 2; exit 1 restarts the stack first, by the start rule of **Resolve
````
  - [x] **Step 3: Verify.** `FLOW_GUARD_TELEMETRY=off scripts/check-verbatim-moves.sh`, `FLOW_GUARD_TELEMETRY=off scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — one prose sentence; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): run the inline verify lint and test lists as two concurrent groups
**After:** Task 1

**Decision:** verify-two-groups

Correction (2026-10-17, review 1): Task 6's implementer key also accepts a trailing `-retry` (`task-<ids>-implementer[-fix-<k>][-retry]`), the handshake-mismatch retry key `skills/flow/implement.md` prescribes. Task 1's signal seven skips commits carrying a `Flow-Key: flow.commit-two` trailer, whose author date the reshape refreshes. Task 5 found no flake in 15 zero-tolerance runs, so no launch-arg pin was added and its subject reads `test(stats): compare visual baselines at zero tolerance`.
