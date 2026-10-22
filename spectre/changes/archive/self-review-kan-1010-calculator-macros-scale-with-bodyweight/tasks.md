# self-review-kan-1010-calculator-macros-scale-with-bodyweight

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1010's eight self-review findings — 1 (Task 1), 2 (Task 2), 3 (Task 3), 4
(Task 4), 5 (Task 5), 6 (Tasks 6 and 7, split per decision `finding-6-two-tasks`), 8 (Task 8),
9 (Task 9), as `proposal.md` numbers them.

**Architecture:** two Go guard verdicts and one report check, one landing-path script fix, one
prompt-list item, the harvester's stamp pairing with a live store repair, two gymie Playwright rows
with their harness, one gymie-playwright script, and one traced flake fix.

**Tech Stack:** Go (`stats/`), Bash, Markdown, Node (gymie-playwright `scripts/`), Playwright.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-22-self-review-kan-1010-calculator-macros-scale-with-bodyweight-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `patches/` and `verbatim-moves.txt` are already committed with the plan. Never merge or push.
- Each task names the repository it runs in; every `**Files:**` path is relative to that worktree's root. Tasks 1–4 and 9 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1010-calculator-macros-scale-with-bodyweight` (`<agents-wt>` below); Tasks 5 and 7 in gymie's `/Users/tweety53/Projects/gymie-worktrees/kan-1010-calculator-macros-scale-with-bodyweight`; Tasks 6 and 8 in gymie-playwright's `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1010-calculator-macros-scale-with-bodyweight` — both on branch `spectre/kan-1010-calculator-macros-scale-with-bodyweight`. Never touch a main checkout, and never create a worktree.
- A task's diff is `<agents-wt>/spectre/changes/self-review-kan-1010-calculator-macros-scale-with-bodyweight/patches/task-<n>.diff`, applied with `git apply <that path>` from the task's worktree root, in task order. Every patch was applied and checked in a scratch clone (agents `b1b0d601`, gymie `1e060db3`, gymie-playwright `bec3d0c`) and `git apply --check` passed on each real worktree. A patch that does not apply is reported, never hand-merged by guesswork.
- Prose lands character for character as its patch carries it: `scripts/check-verbatim-moves.sh` matches every added or reworded sentence against this change's `verbatim-moves.txt` exactly; with all of Tasks 1, 3 and 4 applied in the scratch clone, `check-verbatim-moves.sh`, `check-markdown-integrity.py`, `check-references.sh`, `check-dispatch-paragraphs.sh` and `check-installed-citations.sh` exited 0.
- Go: run `gofmt -w` on every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- A commit in gymie or gymie-playwright is a commit outside every task of kan-1010's change (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`): add `-m "Flow-Key: self-review-kan-1010-calculator-macros-scale-with-bodyweight-fix-g<g>"`, `<g>` the fixer group of `design.md`'s `fixer-groups`; never touch that repository's `spectre/`.
- The kan-1010 dev stack serves FRONTEND_URL=http://localhost:4900 and BACKEND_URL=http://localhost:9980 (gateway); `./gradlew -q devStackFresh` in the gymie worktree answers whether it serves the worktree's build. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container.

## Review Focus

- Task 1: a tick of a task HEAD already carries, an unticked append, a brand-new plan and another change's tasks.md all pass; `reshape-branch.sh` and `commit-split.sh` never see the new verdict.
- Task 2: a list that does rewrite a tracked file still stops `rebase-lint`, with the tree dirty before it or not.
- Task 4: another change's verdict, a longer line number (`:520`), a longer path (`fidelity/x.spec.ts:52` for `x.spec.ts:52`) and a store outage clear nothing.
- Task 5: the rows still run from the main regression checkout (`/Users/tweety53/Projects/gymie-playwright`, falling back to `../gymie`), and a fresh stack's run is unchanged.
- Task 9: a begin whose row never appears is evicted by the existing bound and never claims a launch belonging to another begin.

- [x] 1. Refuse a planning commit that appends an already-ticked task
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `git -C /Users/tweety53/Projects/gymie show 76e18959 -- '*tasks.md' | grep -c '^+- \[x\] 3\.'` prints `1`: kan-1010's fix-run planning commit appended Task 3 ticked, and `check-planning-commit-location.sh` has no check that could refuse it.
    <!-- measured: git -C /Users/tweety53/Projects/gymie show 76e18959 -- '*tasks.md' | grep -c '^+- \[x\] 3\.' @ gymie 76e18959 → 1 -->
  - [x] **Step 2: Apply `patches/task-1.diff`** from `<agents-wt>`: it adds `checkPlanningCommit` (registered as the script; `checkPlanningCommitLocation` stays the in-process call of `reshapebranch.go` and `commitsplit.go`) to `stats/internal/guard/planningcommitlocation.go`, `TestCheckPlanningCommitTickedAppend` to `stats/internal/guard/check_planning_commit_location_test.go`, the `PLANNING-COMMIT-TICKED-APPEND` line to `scripts/check-planning-commit-location.sh`'s header, and the **Every appended task is committed unticked** paragraph to `skills/flow/document-fix.md`.
    <!-- measured: git apply --check patches/task-1.diff @ b1b0d601 → exit 0 -->
  - [x] **Step 3: Prove the test pins the guard.** Save `git diff -- stats/internal/guard/planningcommitlocation.go` to a scratch file outside the worktree, `git checkout -- stats/internal/guard/planningcommitlocation.go`, run `cd stats && go test ./internal/guard -run 'TestCheckPlanningCommitTickedAppend' -count=1` — it fails on the two refusal subtests — then `git apply` the saved diff.
    <!-- measured: that run in a scratch clone at b1b0d601 with the test half alone → FAIL on "an appended ticked task is refused" and "an archived plan's ticked append is refused" -->
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckPlanningCommit' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-references.sh` exit 0. Commit.
    <!-- measured: those commands in a scratch clone at b1b0d601 with the patch → ok; go test ./internal/guard (whole package) ok in 119 s -->

**Build:** green
**Files:** `stats/internal/guard/planningcommitlocation.go`, `stats/internal/guard/check_planning_commit_location_test.go`, `scripts/check-planning-commit-location.sh`, `skills/flow/document-fix.md`
**Tests:** `TestCheckPlanningCommitTickedAppend`
**Regression:** `TestCheckPlanningCommitTickedAppend` fails — a planning commit appending `- [x] 3.` passes the guard again.
**Baseline:** before=1 after=2 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_planning_commit_location_test.go @ b1b0d601 → 1, and 2 with the patch in a scratch clone -->
**Commit:** fix(guard): refuse a planning commit appending an already-ticked task
**After:** none

**Decision:** ticked-append-refused

- [x] 2. Stop rebase-lint only on what the lint list rewrote
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `sed -n '/^rebase_lint() {/,/^}/p' scripts/integrate-change.sh | grep -c 'status --porcelain --untracked-files=no'` prints `1`: a passing list stops whenever the tree is dirty after it.
    <!-- measured: that command @ b1b0d601 → 1 -->
  - [x] **Step 2: Apply `patches/task-2.diff`** from `<agents-wt>`: `rebase_lint` hashes `git diff HEAD` before `prepare-workspace.sh` and stops only when the hash after a passing list differs; `scripts/test-integrate-change.sh` gains case `rebase_lint_dirty` (log 42) and its header line.
    <!-- measured: git apply --check patches/task-2.diff @ b1b0d601 → exit 0 -->
  - [x] **Step 3: Prove the case pins the fix.** Save `git diff -- scripts/integrate-change.sh` outside the worktree, `git checkout -- scripts/integrate-change.sh`, run `bash scripts/test-integrate-change.sh` — it ends `197 passed, 2 failed`, the two `rebase_lint_dirty` checks "the list passed" and "no rebase-lint stop" — then `git apply` the saved diff.
    <!-- measured: that run in a scratch clone at b1b0d601 with the harness half alone → 197 passed, 2 failed -->
  - [x] **Step 4: Verify.** `bash scripts/test-integrate-change.sh` ends `199 passed, 0 failed` — `rebase_lint_rewrite` still stops a list that rewrites a file. Commit.
    <!-- measured: bash scripts/test-integrate-change.sh in a scratch clone at b1b0d601 with the patch → 199 passed, 0 failed -->

**Build:** green
**Files:** `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`
**Tests:** `rebase_lint_dirty`
**Regression:** `rebase_lint_dirty` fails — an uncommitted plan edit the rebase restored stops prepare at `STOP: rebase-lint` again.
**Baseline:** before=134 after=138 <!-- measured: grep -c -E '^check "' scripts/test-integrate-change.sh @ b1b0d601 → 134, and 138 with the patch in a scratch clone -->
**Commit:** fix(scripts): stop rebase-lint only on what the lint list rewrote
**After:** none

**Decision:** rebase-lint-diff-snapshot

- [x] 3. Pass the stack and fingerprint results to the visual verifier
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -h '^- fingerprint:' /Users/tweety53/Projects/gymie-worktrees/kan-1010-calculator-macros-scale-with-bodyweight/.superpowers/sdd/verify-report-*.md | sort -u` prints `- fingerprint: not declared`, and `grep -c "steps 5 and 6's results" skills/flow/visual-verify.md` prints `0`.
    <!-- measured: both commands @ b1b0d601 and the kan-1010 gymie worktree → "- fingerprint: not declared" and 0 -->
  - [x] **Step 2: Apply `patches/task-3.diff`** from `<agents-wt>`: the prompt list of `skills/flow/visual-verify.md` names steps 5 and 6's results; the report template line in `skills/flow/visual-verify-verifier.md` reads `not declared — <step 6's reason, from the prompt>`; `stats/internal/guard/verifyreport.go` counts a bare `- fingerprint: not declared` as undone, with `TestCheckVerifyReportFingerprintReason` and `scripts/check-verify-report.sh`'s header.
    <!-- measured: git apply --check patches/task-3.diff @ b1b0d601 → exit 0 -->
  - [x] **Step 3: Prove the test pins the check.** Save `git diff -- stats/internal/guard/verifyreport.go` outside the worktree, `git checkout -- stats/internal/guard/verifyreport.go`, run `cd stats && go test ./internal/guard -run 'TestCheckVerifyReportFingerprintReason' -count=1` — it fails on the bare line — then `git apply` the saved diff.
    <!-- measured: that run in a scratch clone at b1b0d601 with the test half alone → FAIL, rc=0 VERIFY-REPORT-COMPLETE for "- fingerprint: not declared" -->
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckVerifyReport' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-references.sh` and `scripts/check-dispatch-paragraphs.sh` exit 0. Commit.
    <!-- measured: those commands in a scratch clone at b1b0d601 with Tasks 1, 3 and 4's patches → ok -->

**Build:** green
**Files:** `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`, `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`
**Tests:** `TestCheckVerifyReportFingerprintReason`
**Regression:** `TestCheckVerifyReportFingerprintReason` fails — a report whose fingerprint line carries no step-6 reason is accepted as complete again.
**Baseline:** before=1 after=2 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_verify_report_test.go @ b1b0d601 → 1, and 2 with the patch in a scratch clone -->
**Commit:** fix(flow): pass the stack and fingerprint results to the visual verifier
**After:** none

**Decision:** verifier-gets-stack-results

- [x] 4. Clear a verify failure an earlier false-positive verdict names
  - [x] **Step 1: Reproduce before touching anything; do not guess.** From `/Users/tweety53/Projects/gymie`, `flow record verdicts -guard check-unfinished-work -false-positive | python3 -c 'import json,sys; print(sorted(v["id"] for v in json.load(sys.stdin) if v["change"].startswith("kan-1010") and "screenshots.spec.ts:52" in v["falsePositiveReason"]))'` prints `[662, 666, 681]` — the same failure flagged three times.
    <!-- measured: that command from /Users/tweety53/Projects/gymie @ dev store 2026-10-22 → [662, 666, 681] -->
  - [x] **Step 2: Apply `patches/task-4.diff`** from `<agents-wt>`: signal nine of `stats/internal/guard/unfinishedwork.go` reads the change's false-positive reasons once (`uwFlaggedReasons`) and clears a failure one names as a whole token (`uwNamesFailure`); `TestCheckUnfinishedWorkVerifyFailureFlaggedEarlier` in `stats/internal/guard/check_unfinished_work_test.go`; signal nine's header in `scripts/check-unfinished-work.sh`; the **A flagged verify failure stays cleared** paragraph in `skills/flow/unfinished-work-gate.md`.
    <!-- measured: git apply --check patches/task-4.diff @ b1b0d601 → exit 0 -->
  - [x] **Step 3: Prove the test pins the guard.** Save `git diff -- stats/internal/guard/unfinishedwork.go` outside the worktree, `git checkout -- stats/internal/guard/unfinishedwork.go`, run `cd stats && go test ./internal/guard -run 'TestCheckUnfinishedWorkVerifyFailureFlaggedEarlier' -count=1` — it fails on "this change's verdict naming it clears it" — then `git apply` the saved diff.
    <!-- measured: that run in a scratch clone at b1b0d601 with the test half alone → FAIL on that subtest -->
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckUnfinishedWork' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-references.sh` exit 0. Commit.
    <!-- measured: those commands in a scratch clone at b1b0d601 with the patch → ok -->

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`, `skills/flow/unfinished-work-gate.md`
**Tests:** `TestCheckUnfinishedWorkVerifyFailureFlaggedEarlier`
**Regression:** `TestCheckUnfinishedWorkVerifyFailureFlaggedEarlier` fails — a verify failure already flagged a false positive is listed outstanding on every later run again.
**Baseline:** before=6 after=7 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_unfinished_work_test.go @ b1b0d601 → 6, and 7 with the patch in a scratch clone -->
**Commit:** fix(guard): clear a verify failure an earlier false-positive verdict names
**After:** none

**Decision:** earlier-false-positive-clears

- [x] 5. Run gymie's Playwright verify and capture rows only on a fresh dev stack
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In the gymie worktree, `grep -c '^| `verify` | `B=../gymie' .flow/project.md` prints `0`, and `grep -n 'devStackFresh' .flow/project.md` names line 39 (`## run`, "before any Playwright run"), the `fingerprint` row and its explanation, never the `verify` or `capture` row.
    <!-- measured: both greps in the gymie worktree @ 1e060db3 → 0; lines 39, 396 and 415 -->
  - [x] **Step 2: Apply `patches/task-5.diff`** from the gymie worktree root: the `verify` and `capture` rows of `.flow/project.md` open on `B=../gymie; [ -d ../../gymie-worktrees/"${PWD##*/}" ] && B=../../gymie-worktrees/"${PWD##*/}"; (cd "$B" && ./gradlew -q devStackFresh) && { … }`; `scripts/test-playwright-rows-check-stack.sh` (executable) is added; `.github/workflows/ci.yml`'s shell-guard step runs it.
    <!-- measured: git apply --check task-5.diff in the gymie worktree @ 1e060db3 → exit 0 -->
  - [x] **Step 3: Prove the harness pins the rows.** Save `git diff -- .flow/project.md` to a scratch file outside the worktree, `git checkout -- .flow/project.md`, run `bash scripts/test-playwright-rows-check-stack.sh` — it exits 1 with four `FAIL:` lines, each row's "a stale stack exited 0" and "a stale stack still ran npx" — then `git apply` the saved diff. Never `git stash`: the stash list is shared with every worktree of the repository.
    <!-- measured: the harness against the unpatched project.md in a scratch copy → 4 failure(s), exit 1 -->
  - [x] **Step 4: Verify.** `bash scripts/test-playwright-rows-check-stack.sh` prints `test-playwright-rows-check-stack: ok`; from the gymie-playwright worktree, the patched `verify` row with `<specs>` set to `fidelity/body-calculator.spec.ts`, run under `FRONTEND_PORT=4900 GATEWAY_PORT=9980`, passes its 3 tests; `/Users/tweety53/Projects/agents/scripts/check-visual-verification.sh .` prints `VISUAL-OK`. Commit in the gymie worktree with the `Flow-Key: …-fix-g3` trailer.
    <!-- measured: the harness → ok; the live verify row from the gymie-playwright worktree on the fresh kan-1010 stack → 3 passed (8.8s); check-visual-verification.sh on the patched copy → VISUAL-OK -->

**Build:** green
**Files:** `.flow/project.md`, `scripts/test-playwright-rows-check-stack.sh`, `.github/workflows/ci.yml`
**Tests:** `test-playwright-rows-check-stack.sh`
**Regression:** `test-playwright-rows-check-stack.sh` fails — the `verify` and `capture` rows run Playwright on a stale stack again.
**Baseline:** before=8 after=9 <!-- measured: grep -c -E '^ +bash scripts/test-' .github/workflows/ci.yml @ gymie 1e060db3 → 8, and 9 with the patch in a scratch copy -->
**Commit:** fix(dev): run the Playwright verify and capture rows only on a fresh dev stack
**After:** none

**Decision:** rows-gate-on-fresh-stack

- [x] 6. Re-baseline every capture the specs take
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In the gymie-playwright worktree, `grep -c "update-snapshots=changed" scripts/rebaseline.mjs` prints `1`.
    <!-- measured: that grep @ gymie-playwright bec3d0c → 1 -->
  - [x] **Step 2: Apply `patches/task-6.diff`** from the gymie-playwright worktree root: `scripts/rebaseline.mjs` exports `playwrightArgs`, which ends `--update-snapshots=all`, and runs it; `scripts/rebaseline.test.mjs` gains the test pinning it.
    <!-- measured: git apply --check task-6.diff in the gymie-playwright worktree @ bec3d0c → exit 0 -->
  - [x] **Step 3: Verify.** `node --test scripts/rebaseline.test.mjs` reports `pass 4`, `fail 0`; against the unpatched `scripts/rebaseline.mjs` (save `git diff -- scripts/rebaseline.mjs` outside the worktree, `git checkout -- scripts/rebaseline.mjs`, run, then `git apply` the saved diff) it fails with "does not provide an export named 'playwrightArgs'". Commit in the gymie-playwright worktree with the `Flow-Key: …-fix-g3` trailer.
    <!-- measured: node --test in a scratch copy → pass 4 fail 0 patched; SyntaxError, pass 0 fail 1 unpatched. An --update-snapshots=all run of fidelity/body-calculator.spec.ts on the fresh kan-1010 stack left git status clean (3 passed) -->

**Build:** green
**Files:** `scripts/rebaseline.mjs`, `scripts/rebaseline.test.mjs`
**Tests:** `every capture the specs take is rewritten, an in-tolerance move included`
**Regression:** `every capture the specs take is rewritten, an in-tolerance move included` fails — the rebaseline run leaves in-tolerance moves on their old PNGs again.
**Baseline:** before=3 after=4 <!-- measured: grep -c -E '^test\(' scripts/rebaseline.test.mjs @ gymie-playwright bec3d0c → 3, and 4 with the patch in a scratch copy -->
**Commit:** fix(scripts): re-baseline every capture the specs take
**After:** none

**Decision:** rebaseline-all

- [x] 7. Name the rebaseline run's update mode in gymie's project file
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In the gymie worktree, `grep -c 'one Playwright run with `--update-snapshots=changed`' .flow/project.md` prints `1`.
    <!-- measured: that grep in the gymie worktree @ 1e060db3 → 1 -->
  - [x] **Step 2: Apply `patches/task-7.diff`** from the gymie worktree root: the rebaseline sentence names `--update-snapshots=all`, why not `changed`, and that the listing names only PNGs whose pixels differ from `HEAD`.
    <!-- measured: git apply --check task-7.diff in the gymie worktree @ 1e060db3 → exit 0, independently of task-5.diff -->
  - [x] **Step 3: Verify.** `git diff --stat` names only `.flow/project.md`; `grep -c 'update-snapshots=changed' .flow/project.md` prints `0`. Commit in the gymie worktree with the `Flow-Key: …-fix-g3` trailer.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** none — a documentation sentence; Task 6's test pins the behaviour it describes
**Regression:** none — documentation only
**Commit:** docs(dev): name the rebaseline run's update-snapshots mode
**After:** Task 5

**Decision:** finding-6-two-tasks

- [ ] 8. Make the reachable-screen sweep settle deterministically under load
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Confirm `./gradlew -q devStackFresh` exits 0 in the gymie worktree (else stop and report: the stack is the parent's to restart). In the gymie-playwright worktree, run the sweep under the load it failed in — `FRONTEND_URL=http://localhost:4900 BACKEND_URL=http://localhost:9980 GATEWAY_PORT=9980 npx playwright test screenshots.spec.ts --repeat-each=10 --workers=5`, and if that stays green, the sweep beside the suite's other specs at the configured 5 workers, narrowing per `docs/flake-bisection.md` (gymie) to the smallest set that reproduces. A full-suite run is the bounded exception that brief allows, its reason stated first. Record the reproducing command and its failure rate. The verify report it failed in reads "test timeout 120000ms in waitForStableCanvas; passes on immediate re-run".
    <!-- unverified:whether the repeated run alone reproduces it; kan-1010's own repeated and stressed runs never did -->
  - [ ] **Step 2: Trace it.** With a temporary, uncommitted timing log around each `waitForStableCanvas` call the sweep reaches (`fixtures.ts`: `expectRoom` and the direct calls), rerun the reproducing command and name the call that spends the budget: one canvas that never stops changing (what keeps drawing — a caret, a spinner, an animation on the live clock), or every call slow under load with none stuck. Remove the timing log.
  - [ ] **Step 3: Fix at the source** the trace named, per `~/.claude/rules/fix-determinism-at-the-source.md`: freeze what keeps drawing (`pinClock`/`pauseClock`/`onPausedClock` in `fixtures.ts`, as kan-1010's wr-E3b fix froze its caret), or gate the capture on the app's own drawn-and-idle state rather than on screenshot polling. Never raise the test timeout, `waitForStableCanvas`'s `timeoutMs`, `pollMs` or `minMs`, a retry count or a sleep. Should the trace place the nondeterminism in gymie-frontend's code, stop and report the trace rather than editing that repository.
  - [ ] **Step 4: Prove the gate reaches the code under test.** Run the reproducing command often enough that Step 1's failure rate predicts at least ten failures, and see none; then remove the gate, rerun it until it fails, and restore the gate. Record both figures in the commit body.
  - [ ] **Step 5: Verify.** `FRONTEND_URL=http://localhost:4900 BACKEND_URL=http://localhost:9980 GATEWAY_PORT=9980 npx playwright test screenshots.spec.ts` passes alone, and `npx playwright test --list` exits 0 (the checkout carries no tsconfig or type-check script; Playwright's own load is the compile check); a capture the change touches is byte-identical to `HEAD` or named in the commit body. Commit in the gymie-playwright worktree with the `Flow-Key: …-fix-g4` trailer.

**Build:** green
**Files:** `fixtures.ts`, `screenshots.spec.ts`
**Tests:** none — repairs existing tests: `screenshots.spec.ts`'s "sweep every reachable screen", which fails under Step 1's load without the fix
**Regression:** none — the sweep's flake under load returns; Step 4's removal run is the proof
**Commit:** fix(screenshots): settle the reachable-screen sweep deterministically under load
**After:** none

**Decision:** sweep-flake-after-trace

- [x] 9. Retry a dispatch stamp whose row is not written yet, and stamp kan-1010's panel-0 rows
  - [x] **Step 1: Reproduce before touching anything; do not guess.** From `/Users/tweety53/Projects/gymie`, `flow record dispatches -change kan-1010-calculator-macros-scale-with-bodyweight | python3 -c 'import json,sys; print([(d["id"], d["agentId"]) for d in json.load(sys.stdin) if d["key"].startswith("panel-0-")])'` prints `[(6922, None), (6923, None)]`.
    <!-- measured: that command @ dev store 2026-10-22 → [(6922, None), (6923, None)] -->
  - [x] **Step 2: Apply `patches/task-9.diff`** from `<agents-wt>`: `stampDispatchAgents` (`stats/internal/harvest/watcher.go`) collects each time-paired stamp the store reports as not stamped and returns its begin and launch to the pending lists after the pass; `TestWatcherRetriesAPairWhoseRowIsNotWrittenYet` and its `rowStampDeps` in `stats/internal/harvest/watcher_test.go` replay kan-1010's panel round 0 — two keyed launches, then the Bash line recording both begins read before the rows exist, then the session's next Bash call.
    <!-- measured: git apply --check patches/task-9.diff @ b1b0d601 → exit 0 -->
  - [x] **Step 3: Prove the replay pins the fix.** Save `git diff -- stats/internal/harvest/watcher.go` outside the worktree, `git checkout -- stats/internal/harvest/watcher.go`, run `cd stats && go test ./internal/harvest -run 'TestWatcherRetriesAPairWhoseRowIsNotWrittenYet' -count=1` — it fails with `stamps = []` — then `git apply` the saved diff.
    <!-- measured: that run in a scratch clone at b1b0d601 with the test half alone → FAIL, stamps = [] -->
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/harvest` prints nothing; `cd stats && go test ./internal/harvest -run 'TestWatcher' -count=1` passes. Commit.
    <!-- measured: go test ./internal/harvest (whole package) in a scratch clone at b1b0d601 with the patch → ok -->
  - [x] **Step 5: Repair the two rows**, after the commit, from `<agents-wt>/stats`: `make -s web-dist`, then for each pair run `FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd restamp-dispatch <id> <transcript>` — `6922` with `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/2a92ac3e-0c5c-4624-ab01-ea0146376e6b/subagents/agent-a198d990a8e9305e5.jsonl`, `6923` with `…/subagents/agent-ad9089f7f5a39abe2.jsonl` in the same directory. Step 1's command then prints `[(6922, 'a198d990a8e9305e5'), (6923, 'ad9089f7f5a39abe2')]`; report both outputs.
    <!-- measured: both agent transcripts exist under that subagents/ directory; the launch results at 20:22:05.589Z and 20:22:07.203Z name those agent ids for those keys -->

**Build:** green
**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`
**Tests:** `TestWatcherRetriesAPairWhoseRowIsNotWrittenYet`
**Regression:** `TestWatcherRetriesAPairWhoseRowIsNotWrittenYet` fails — a panel round's rows, begun after their launches, are left unstamped again.
**Baseline:** before=66 after=67 <!-- measured: grep -c -E '^func Test' stats/internal/harvest/watcher_test.go @ b1b0d601 → 66, and 67 with the patch in a scratch clone -->
**Commit:** fix(harvest): retry a stamp whose dispatch row is not written yet
**After:** none

**Decision:** stamp-retry-unwritten-row
