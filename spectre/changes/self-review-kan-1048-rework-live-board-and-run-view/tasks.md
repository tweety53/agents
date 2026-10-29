# self-review-kan-1048-rework-live-board-and-run-view

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1048's seven self-review findings — 1-A (Task 1), 1-B (Task 2), 1-D (Task 3),
1-E (Task 4), 1-F (Task 5), 1-J (Task 6), 8-A (Task 7), as `proposal.md` names them.

**Architecture:** a regex anchor in `check-verify-green`; one repotest sweep over guard shims; an
errexit refusal in the reproducer runner; a planning-path row in `check-plan-shape`; a bounded
startup restamp in `flowd` plus a one-off live repair; an `agent_id` filter in the activity hook; one
verifier paragraph.

**Tech Stack:** Go (`stats/`), Python (hook), Bash, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-29-self-review-kan-1048-rework-live-board-and-run-view-design.md` in this
worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/`. Never merge or push.
- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1048-rework-live-board-and-run-view` (`<wt>` below), branch `spectre/kan-1048-rework-live-board-and-run-view`. Every `**Files:**` path is relative to `<wt>`. Never touch the main checkout, and never create a worktree.
- Go: `gofmt -w` every Go file you change; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Mutation proofs restore with `scripts/break-and-prove.sh` or `scripts/mutate-and-verify.sh`, or a saved diff re-applied with `git apply` — never `git checkout --` on a file carrying uncommitted edits.
- Task 5's live repair runs `cd <wt>/stats && make -s web-dist && go run ./cmd/flowd restamp-ends …`, which reaches the dev `flow` database — dev-workspace storage, not production. Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database, and never run tests against the live database.
- Go test runs set a Bash `timeout` of 600000.

## Review Focus

- Task 1: `verify-2-fe`, `verify-fix-1-fe`, `verify-2`, `verify-fix-1` still group as before.
- Task 2: the test fails when one shim drops a sibling line, and when a source with sibling execs maps to no shim.
- Task 3: a reproducer without the marker is unaffected; the refusal happens before execution.
- Task 4: `visual-verification.md` stays allowed (F21 still requires it on a visual-verify fix).
- Task 5: a restamp error never stops startup; the window is relative to `now`.
- Task 6: `Notification`, `UserPromptSubmit` and `Stop` carrying `agent_id` behave as before.

- [ ] 1. Anchor the verify re-run number
  - [ ] **Step 1: Write the test.** In `TestCheckVerifyGreen` (`stats/internal/guard/check_verify_green_test.go`), add the row `a re-run of a digit-led repository's blocked verify is green`: rows `verify-3d-viewer` blocked, then `verify-2-3d-viewer` completed, both `mf-tok` — want exit 0 and `1 inline verify row(s) green`. Run `cd stats && go test ./internal/guard -run 'TestCheckVerifyGreen' -count=1` — the row fails (two groups, `verifyd-viewer` blocked).
  - [ ] **Step 2: Fix.** In `stats/internal/guard/verifygreen.go`, set `vgFixSuffix` to `^verify-(fix-)?[0-9]+(-|$)` and replace with `"verify$2"` at its one use, per decision `verify-rerun-anchor`; the comment above it says the number ends at a dash or the key's end, so a repository name starting with a digit is never a re-run number (kan-1048 self-review, 1-A).
  - [ ] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, drop `(-|$)` — the new row fails; restore.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckVerifyGreen' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/guard/verifygreen.go`, `stats/internal/guard/check_verify_green_test.go`
**Tests:** `a re-run of a digit-led repository's blocked verify is green`
**Regression:** `a re-run of a digit-led repository's blocked verify is green` fails — a blocked `verify-3d-viewer` stays open after its green re-run.
**Commit:** fix(guard): end a verify re-run number at a dash or the key's end
**After:** none

**Decision:** verify-rerun-anchor

- [ ] 2. Check every guard shim names its siblings
  - [ ] **Step 1: Write the test.** `git mv stats/internal/repotest/run_verify_list_siblings_test.go stats/internal/repotest/guard_shim_siblings_test.go`; replace `TestRunVerifyListShimNamesItsSiblings` with `TestGuardShimsNameTheirSiblings` per decision `shim-siblings-every-guard`, its doc comment citing kan-1048 (run-capped.sh) and kan-1048 self-review 1-B. Run `cd stats && go test ./internal/repotest -run 'TestGuardShimsNameTheirSiblings' -count=1` — it passes on arrival. <!-- measured: a python prototype of the same four patterns and mapping over stats/internal/guard/*.go found no missing sibling @ 46092b9f -->
  - [ ] **Step 2: Prove the test pins the invariant.** With `scripts/break-and-prove.sh`, delete `"$SCRIPT_DIR/resolve-visual-screenshots.sh"` from `scripts/rebuild-full-app-suite-zip.sh`'s `:` line — the test fails naming it; restore. Then delete `"$SCRIPT_DIR/check-narrative-runs.sh"` from `scripts/integrate-change.sh` — the test fails naming `integratechange_commit.go`'s sibling; restore.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l internal/repotest` prints nothing; `cd stats && go test ./internal/repotest -run 'TestGuardShimsNameTheirSiblings' -count=1` passes; `git grep -n -F -e 'TestRunVerifyListShimNamesItsSiblings' -- stats scripts skills` prints nothing. Commit.

**Build:** green
**Files:** `stats/internal/repotest/run_verify_list_siblings_test.go`, `stats/internal/repotest/guard_shim_siblings_test.go`
**Tests:** `TestGuardShimsNameTheirSiblings`
**Regression:** `TestGuardShimsNameTheirSiblings` fails — a guard shim that stops spelling a sibling the Go execs goes uninstalled by `setup.sh` and the guard fails to start.
**Commit:** test(repotest): check every guard shim names the siblings its Go execs
**After:** none

**Decision:** shim-siblings-every-guard

- [ ] 3. Refuse a mutation reproducer that does not run under errexit
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** A reproducer `.superpowers/sdd/reproducers/zz-spike.sh` holding `#!/bin/bash`, `# mutation-reproducer`, `python3 -c "assert False" 2>/dev/null`, `true` — `scripts/run-reproducer.sh "$PWD" .superpowers/sdd/reproducers/zz-spike.sh` prints `defect demonstrated` and exits 0 <!-- measured: that run @ 46092b9f -->. Delete the file.
  - [ ] **Step 2: Write the test** `TestRunReproducerRefusesAMutationWithoutErrexit` in `stats/internal/guard/runreproducer_test.go`, on the file's fixture helpers: a marker-declaring reproducer with no `set -e` is refused, exit 2, the message naming `set -e`, and it never ran; the same body with `set -e`, `set -eu`, `set -euo pipefail` or `set -o errexit` after the marker runs and answers by the mutation convention; a reproducer without the marker and without `set -e` is unaffected. Run `cd stats && go test ./internal/guard -run 'TestRunReproducerRefusesAMutationWithoutErrexit' -count=1` — it fails.
  - [ ] **Step 3: Fix.** In `runReproducer` (`stats/internal/guard/runreproducer.go`), add the refusal of decision `mutation-reproducer-errexit` right after the near-miss check, with its `ponytail:` comment. Add `set -e` after the marker in every fixture that declares it and expects a run — `runreproducer_test.go`, `check_panel_reproducer_exit_contract_test.go`, `prove_reproducer_test.go` (and `check_mutation_reproducer_pin_test.go` only if one of its fixtures runs the runner) — keeping each marker at its line (case 28.c's marker stays at line 10, 28.d's at 11). In `scripts/run-reproducer.sh`'s exit-2 description, name the refusal. In `skills/flow/review-panel.md`'s mutation-testing brief, after "it never buys a test.", add: a mutation reproducer runs under `set -e`, so an edit that fails exits non-zero, and `run-reproducer.sh` refuses one with no `set -e` line (kan-1048 self-review, 1-D: four of seven reproducers carried no landing check). In `skills/flow/test-sanity-reviewer-prompt.md`, the vacuous-test reproducer sentence adds "runs under `set -e`".
  - [ ] **Step 4: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, skip the refusal — the test fails; restore.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'RunReproducer|PanelReproducerExitContract|ProveReproducer|MutationReproducerPin' -count=1` passes; `scripts/check-mutation-reproducer-pin.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`, `scripts/run-reproducer.sh`, `skills/flow/review-panel.md`, `skills/flow/test-sanity-reviewer-prompt.md`
**Allowed-collateral:** `stats/internal/guard/check_panel_reproducer_exit_contract_test.go`, `stats/internal/guard/prove_reproducer_test.go`, `stats/internal/guard/check_mutation_reproducer_pin_test.go`
**Tests:** `TestRunReproducerRefusesAMutationWithoutErrexit`
**Regression:** `TestRunReproducerRefusesAMutationWithoutErrexit` fails — a mutation whose edit never landed reads as a surviving mutant.
**Baseline:** before=10 after=11 <!-- measured: grep -c -E '^func Test[A-Z][A-Za-z0-9]*\(t \*testing\.T\)' stats/internal/guard/runreproducer_test.go @ 46092b9f -->
**Commit:** fix(guard): refuse a mutation reproducer that does not run under set -e
**After:** none

**Decision:** mutation-reproducer-errexit

- [ ] 4. Refuse a planning path in a task's Files or Allowed-collateral
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `scripts/check-plan-shape.sh spectre/changes/archive/kan-1048-rework-live-board-and-run-view/tasks.md` exits 0 <!-- measured: that run @ 46092b9f -->, although its task 6's `**Allowed-collateral:**` names `spectre/changes/kan-1048-rework-live-board-and-run-view/visual-verification/**`.
  - [ ] **Step 2: Write the test** `TestCheckPlanShapeRefusesAPlanningPathEntry` in `stats/internal/guard/check_plan_shape_test.go`, on `psFixture` and `runGuard`: a task whose Allowed-collateral names `spectre/changes/x/visual-verification/**` exits 1 with `F23` and the entry; one whose Files names `docs/superpowers/notes.md` beside `a.go` exits 1; one naming `spectre/changes/x/visual-verification.md` beside `a.go` exits 0; `psMutant(p, "F23")` lets the first through. Run `cd stats && go test ./internal/guard -run 'TestCheckPlanShapeRefusesAPlanningPathEntry' -count=1` — it fails.
  - [ ] **Step 3: Fix.** In `stats/internal/guard/planshape.go`, add row F23 per decision `plan-shape-planning-entry` beside F14, reading `tcfBodyFields`' `Files` and `Allowed-collateral` and `psPlanning`, reported at the field's line; add its row to the header comment and make the count `F1-F23`. In `skills/flow/verify-fix-loop.md` step 1, after the sentence on `<changeRoot>/visual-verification.md` being untracked until the task's commit, add: the round's evidence directory, `<changeRoot>/visual-verification/`, is never named in the task's `**Files:**` or `**Allowed-collateral:**` — `check-plan-shape` F23 refuses it and `check-task-commit-planning-paths` sweeps it — and rides the stage's own step-12 commit (`skills/flow/visual-verify.md`) (kan-1048 self-review, 1-E).
  - [ ] **Step 4: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, drop the `Allowed-collateral` field from F23's read — the first case fails; restore.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckPlanShape' -count=1` passes; `scripts/check-plan-shape.sh spectre/changes/self-review-kan-1048-rework-live-board-and-run-view/tasks.md` exits 0; `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`, `skills/flow/verify-fix-loop.md`
**Tests:** `TestCheckPlanShapeRefusesAPlanningPathEntry`
**Regression:** `TestCheckPlanShapeRefusesAPlanningPathEntry` fails — a plan declaring a planning path for a task commit passes the gate and fails at the task's commit.
**Baseline:** before=5 after=6 <!-- measured: grep -c -E '^func Test[A-Z][A-Za-z0-9]*\(t \*testing\.T\)' stats/internal/guard/check_plan_shape_test.go @ 46092b9f -->
**Commit:** fix(guard): refuse a planning path in a task's declared files
**After:** none

**Decision:** plan-shape-planning-entry

- [ ] 5. flowd restamps recent stamped ends at startup
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tAc "select count(*) from dispatches where ended_at >= '2026-10-28T00:00:00Z' and agent_id is not null and agent_id <> 'inline' and ended_at <> date_trunc('milliseconds', ended_at)"` prints 19 <!-- measured: that query @ 2026-10-29 -->, and kan-1048's `panel-0-*` rows end at 22:17:12.5–.7Z.
  - [ ] **Step 2: Write the test** `TestStartupRestampsRecentStampedEnds` in `stats/cmd/flowd/wiring_test.go`, adding `func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }` and opening a store with `storetest.NewStore`: two non-inline agent rows with sub-millisecond ends, one ended an hour before `now`, one eight days before, each with an agent end inside its window; `restampRecentEnds` moves the first only and returns 1. <!-- unverified:seeding a change and a dispatch row through the store's exported API from outside package store — reuse what internal/store's seedChange and baseDispatch call --> Run `cd stats && make -s web-dist && go test ./cmd/flowd -run 'TestStartupRestampsRecentStampedEnds' -count=1` — it fails to compile.
  - [ ] **Step 3: Fix.** In `stats/cmd/flowd/main.go`, add `startupRestampWindow` and `restampRecentEnds` per decision `startup-restamp-window` and call it in `run` once `agentEnd` is built, logging its count and `since`, an error a warning (kan-1048 self-review, 1-F).
  - [ ] **Step 4: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, widen the window to `30 * 24 * time.Hour` — the test fails (two rows move); restore.
  - [ ] **Step 5: Repair.** `cd stats && go run ./cmd/flowd restamp-ends 2026-10-28T00:00:00Z`; re-run Step 1's query — 19 less the printed `restamped:` count, and the `panel-0-*` rows no longer share one end. Record both figures in the commit body.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l cmd` prints nothing; `cd stats && make -s web-dist && go vet ./cmd/flowd` exits 0 and `cd stats && go test ./cmd/flowd -run 'TestStartupRestampsRecentStampedEnds|TestDaemonWiresTheRealStore' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flowd/main.go`, `stats/cmd/flowd/wiring_test.go`
**Tests:** `TestStartupRestampsRecentStampedEnds`
**Regression:** `TestStartupRestampsRecentStampedEnds` fails — a daemon restarted after running an older build leaves its stamped ends wrong.
**Baseline:** before=10 after=11 <!-- measured: grep -c -E '^func Test[A-Z][A-Za-z0-9]*\(t \*testing\.T\)' stats/cmd/flowd/wiring_test.go @ 46092b9f -->
**Commit:** fix(stats): restamp recent stamped dispatch ends when flowd starts
**After:** none

**Decision:** startup-restamp-window

- [ ] 6. The activity hook ignores subagent tool events
  - [ ] **Step 1: Write the test.** In `TestFlowSessionActivityHook` (`stats/internal/repotest/flow_session_activity_hook_test.go`), add rows `a subagent's post tool use sends nothing` (`PostToolUse`, `"agent_id":"a1"`, want "") and `a subagent's ask user question sends nothing` (`PreToolUse` `AskUserQuestion` with `agent_id`, want ""), and `a subagent's permission notification still sends` (`Notification` with `agent_id`, want the `permission` PUT). Run `cd stats && go test ./internal/repotest -run 'TestFlowSessionActivityHook' -count=1` — the first two fail.
  - [ ] **Step 2: Fix.** In `activity()` (`hooks/flow-session-activity.py`), return `None` first for a `PreToolUse` or `PostToolUse` payload whose `agent_id` is non-empty, with a comment citing the Claude Code hooks docs' `agent_id` field and kan-1048 self-review 1-J, per decision `activity-ignores-subagents`.
  - [ ] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, drop the `agent_id` check — the first row fails; restore.
  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/repotest -run 'TestFlowSessionActivityHook' -count=1` passes; `python3 -m py_compile hooks/flow-session-activity.py` exits 0. Commit.

**Build:** green
**Files:** `hooks/flow-session-activity.py`, `stats/internal/repotest/flow_session_activity_hook_test.go`
**Tests:** `a subagent's post tool use sends nothing`, `a subagent's ask user question sends nothing`, `a subagent's permission notification still sends`
**Regression:** `a subagent's post tool use sends nothing` fails — a background subagent's tool call turns an open question back to working on the live board.
**Commit:** fix(hooks): ignore a subagent's tool events in the session activity hook
**After:** none

**Decision:** activity-ignores-subagents

- [ ] 7. Confirm a one-frame motion departure before reporting it
  - [ ] **Step 1: Edit.** In `skills/flow/visual-verify-verifier.md`'s motion step, after "A frame that breaks one is a departure, named by its index and what it shows.", add the paragraph of decision `one-frame-departure-confirmed`, citing kan-1048 self-review 8-A: a single screencast frame showed the tooltip lingering, React flushes the discrete `mouseleave` before paint and a `requestAnimationFrame` DOM read showed no lag, and the report cost a whole fix round.
  - [ ] **Step 2: Verify.** `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** none — a prose paragraph.
**Regression:** none — prose.
**Commit:** docs(flow): confirm a one-frame motion departure before reporting it
**After:** none

**Decision:** one-frame-departure-confirmed
