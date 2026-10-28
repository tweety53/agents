# self-review-kan-1024-templates-warm-up-cool-down-any-workout

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1024's ten self-review findings — F-b (Task 1), F-i (Task 2), F-j (Task 3),
F-k (Task 4), F-l (Task 5), F-m (Task 6), F-n (Task 7), F-c (Task 8), F-d (Task 9), F-e (Task 10),
as `proposal.md` names them.

**Architecture:** the kickoff merge-order writer reads the state record; two begin-time checks in
`flow record dispatch`; one plan-shape rule; two renderer fixes; a bounded abandoned-stage sweep
with a repair command; an end-before-start refusal; the round stated in the slot prompt and a
cross-round reproducer check; two playwright flakes, one capture filter and one buildSrc
extraction in the kan-1024 worktrees; four live store repairs.

**Tech Stack:** Go (`stats/`), Bash, Markdown; TypeScript/Playwright (gymie-playwright); Kotlin and
Gradle Kotlin DSL (gymie).

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-28-self-review-kan-1024-templates-warm-up-cool-down-any-workout-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer, and no `Co-Authored-By:` trailer. Never write `spectre/changes/` in any repository. Never merge or push.
- Tasks 1–7 run in `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1024-templates-warm-up-cool-down-any-workout` (`<agents-wt>` below), branch `self-review-kan-1024-templates-warm-up-cool-down-any-workout`. Tasks 8 and 9 run in `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1024-templates-warm-up-cool-down-any-workout` (`<pw-wt>`), Task 10 in `/Users/tweety53/Projects/gymie-worktrees/kan-1024-templates-warm-up-cool-down-any-workout` (`<gymie-wt>`), each on branch `spectre/kan-1024-templates-warm-up-cool-down-any-workout`, committing after that branch's existing commits and rewriting none. Every `**Files:**` path is relative to its task's worktree root. Never touch a main checkout, and never create a worktree.
- The gymie-playwright and gymie commits (Tasks 8–10) are commits outside every task of kan-1024's change (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`): add `-m "Flow-Key: self-review-kan-1024-templates-warm-up-cool-down-any-workout-fix-g<g>"`, `<g>` the fixer group's number (`design.md`, `fixer-grouping`).
- Go: run `gofmt -w` on every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Mutation proofs restore with `break-and-prove.sh` or `mutate-and-verify.sh`, or a saved diff re-applied with `git apply` — never `git checkout --` on a file carrying uncommitted edits.
- Live store writes (Tasks 2's check, 5, 6 and 7) go through `flowd` subcommands, which open the database directly, and `flow record` against the running daemon; never stop or restart `flowd` or the `flow-postgres` container, and never run tests against the live database.
- Gymie: Gradle and Playwright runs take a Bash `timeout` of 600000; one Gradle run at a time per worktree. The kan-1024 stack is started by gymie's `.flow/project.md` `## visual verification` `verify` row, run from `<pw-wt>` with its `<specs>` filled in — it runs `devStackFresh` in `<gymie-wt>` first. Run `adb`, if ever needed, from `~` only.

## Review Focus

- Task 1: a change whose `link.md` already lists some repositories gets only the missing ones, in `spectre/peers` order; `<from>`'s own repository is never listed twice.
- Task 2: a `backfill`, a `-retry` key, a chunk key `panel-fix-<r>-2`, and a key reused only by `aborted` rows pass; an unreachable daemon never refuses.
- Task 5: a stage run with no later stage of its change still ends at the sweep's instant; rows with no `change_id` are untouched by the bound.
- Task 6: the same-second clamp never moves `ended_at` later than `started_at`; a `blocked` or `aborted` end before the start is refused like a `completed` one.
- Task 7: a same-round shared path (the joined-slot re-record) still passes.
- Tasks 8–9: no wider loop, longer timeout or extra retry anywhere in the diff.

- [x] 1. Kickoff's merge order names every worktree the state record holds
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `<agents-wt>`, add the test below and run `cd stats && go test ./internal/guard -run 'TestKickoffWorktreeFromListsEveryRecordedWorktree' -count=1` — it fails: the new `link.md` lists `.` and the last repository only.
  - [x] **Step 2: Write the test** `TestKickoffWorktreeFromListsEveryRecordedWorktree` in `stats/internal/guard/kickoff_worktree_test.go`, on `TestKickoffWorktreeFromWritesMergeOrder`'s fixtures: a canonical project and two peers; the first peer is added with `--from` before `spectre/changes/<name>/` exists (no `link.md`), then the directory is created and the second peer added with `--from`. `link.md` must read `## Merge order` with `.`, then both peers in `spectre/peers` order.
  - [x] **Step 3: Fix.** In `kickoffWorktree` (`stats/internal/guard/kickoffworktree.go`), once `flow state add-worktree` has run and `link != ""`, read `flow state get -C <from> <name>`, resolve each `worktrees` key's main checkout (`git rev-parse --path-format=absolute --git-common-dir`) against the `spectre/peers` paths as `kwPeerName` does, drop `<from>`'s own repository, and call `kwMergeOrder` once per resolved name in `spectre/peers` order; print `merge-order: <link>` once when any call wrote. A recorded worktree no peer names is skipped. Update the header comment of `scripts/kickoff-worktree.sh` (the `link.md` paragraph) to say the record's repositories are listed.
  - [x] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, make the loop call `kwMergeOrder` for the new peer alone — the test fails; restore.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestKickoffWorktree' -count=1` passes; `scripts/check-guard-symlinks.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/kickoffworktree.go`, `stats/internal/guard/kickoff_worktree_test.go`, `scripts/kickoff-worktree.sh`
**Tests:** `TestKickoffWorktreeFromListsEveryRecordedWorktree`
**Regression:** `TestKickoffWorktreeFromListsEveryRecordedWorktree` fails — a `link.md` written after kickoff names only the repository added last.
**Baseline:** before=4 after=5 <!-- measured: grep -c -E '^func Test' stats/internal/guard/kickoff_worktree_test.go @ c30fab36 -->
**Commit:** fix(stats): list every recorded worktree in kickoff's merge order
**After:** none

**Decision:** merge-order-from-state-record

- [x] 2. Refuse a panel-fix begin with no findings to fix or a reused key
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tc "select id, dispatch_key, session_token from dispatches where id in (8256, 8275)"` prints two `panel-fix-1` rows under different tokens, and `select min(f.id) from findings f join changes c on c.id=f.change_id where c.name='kan-1024-templates-warm-up-cool-down-any-workout' and f.round=0` is above every round-1 row's id — round 0's rows were written after its fix began.
  - [x] **Step 2: Write the tests** in `stats/cmd/flow/record_test.go`, on the panel-round-floor test's fake run-record server: `TestRecordDispatchRefusesPanelFixWithoutFindings` — `begin -key panel-fix-2` over a run record whose round 1 holds only Minor rows exits 2 naming round 1 and journals nothing; with an `important` (lowercase) round-1 row it passes; `backfill` of the same key passes. `TestRecordDispatchRefusesReusedPanelFixKey` — `begin -key panel-fix-1` when the record holds a `panel-fix-1` dispatch under another session token exits 2 naming that token; when that dispatch's outcome is `aborted`, or its token is this one, it passes.
  - [x] **Step 3: Fix.** In `stats/cmd/flow/record.go`, beside the panel-round-floor check and only when `*startedAt == ""`: for a `-key` matching `panelFixKeyRE`, read the run record once (a failed read skips both checks) and refuse, exit 2 before any store call, (a) when no finding has round `<r>-1` and severity Critical or Important, case-insensitive — the message names the round and says to record that round's findings with `flow record finding -from` first; (b) when a dispatch carries the same key under a different session token and an outcome other than `aborted` — the message names that token and the next free round.
  - [x] **Step 4: Prove the tests pin the fix.** With `break-and-prove.sh`, disable (a), then (b) — each test fails in turn; restore.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestRecordDispatch' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchRefusesPanelFixWithoutFindings`, `TestRecordDispatchRefusesReusedPanelFixKey`
**Regression:** both fail — a panel-fix begins with no Critical/Important row in its review round, or under a key another session of the change used.
**Baseline:** before=145 after=147 <!-- measured: grep -c -E '^func Test' stats/cmd/flow/record_test.go @ c30fab36 -->
**Commit:** fix(stats): refuse a panel-fix begin with no findings or a reused key
**After:** none

**Decision:** panel-fix-begin-needs-findings

- [x] 3. Plan-shape refuses a visual-verify fix task that omits visual-verification.md
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Write a scratch plan holding one task titled `In-run fix 1 — visual-verify: x` with `**Files:** `a.kt`` and run `scripts/check-plan-shape.sh <scratch>/tasks.md` — it exits 0.
  - [x] **Step 2: Write the test** `TestCheckPlanShapeVisualVerifyFixNamesTheRecord` in `stats/internal/guard/check_plan_shape_test.go`: that task is refused with a line naming the task and `visual-verification.md`; with `spectre/changes/x/visual-verification.md` added to `**Files:**` it passes; a task titled `In-run fix 1 — verify: x` without it passes.
  - [x] **Step 3: Fix.** Add rule F21 to `stats/internal/guard/planshape.go` (its rule table in the header comment, and the check in `checkTask` under `p.on("F21")`): a task whose title matches `^In-run fix [0-9]+ — visual-verify` and whose `**Files:**` holds no path ending `visual-verification.md` is refused at its title line, citing **The loop** (`skills/flow/verify-fix-loop.md`). Add F21 to the rule list in `scripts/check-plan-shape.sh`'s header.
  - [x] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, make the F21 condition false — the test fails; restore.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckPlanShape' -count=1` passes; `scripts/check-plan-shape.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`  <!-- corrected: the shim header carries no rule list; F21 is documented in planshape.go -->
**Tests:** `TestCheckPlanShapeVisualVerifyFixNamesTheRecord`
**Regression:** `TestCheckPlanShapeVisualVerifyFixNamesTheRecord` fails — an in-run visual-verify fix task that omits the change's record passes the guard again.
**Baseline:** before=3 after=4 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_plan_shape_test.go @ c30fab36 -->
**Commit:** fix(stats): refuse a visual-verify fix task that omits the record
**After:** none

**Decision:** plan-shape-visual-fix-record

- [x] 4. The panel verdict skips rows superseded within their round
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Add the test below; `cd stats && go test ./internal/records -run 'TestRenderPanelVerdictSkipsSupersededRows' -count=1` fails reading `2 Important`.
  - [x] **Step 2: Write the test** `TestRenderPanelVerdictSkipsSupersededRows` in `stats/internal/records/render_test.go`: round 1 holds Important `F1` and `F2`, `F2` superseding `F1` — the verdict reads `fix (0 Critical, 1 Important)`; a round-2 row superseding a round-1 row leaves round 1's count unchanged.
  - [x] **Step 3: Fix.** In `RenderPanel` (`stats/internal/records/render.go`), the per-round Critical/Important count skips a finding whose ref another finding of the same round names in `Supersedes`.
  - [x] **Step 4: Prove the test pins the fix.** With `break-and-prove.sh`, drop the skip — the test fails; restore.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l internal/records` prints nothing; `cd stats && go test ./internal/records -run 'TestRenderPanel' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/records/render.go`, `stats/internal/records/render_test.go`
**Tests:** `TestRenderPanelVerdictSkipsSupersededRows`
**Regression:** `TestRenderPanelVerdictSkipsSupersededRows` fails — a round's verdict counts a finding and the finding that supersedes it.
**Baseline:** before=38 after=39 <!-- measured: grep -c -E '^func Test' stats/internal/records/render_test.go @ c30fab36 -->
**Commit:** fix(stats): skip superseded findings in the panel verdict count
**After:** none

**Decision:** verdict-skips-superseded

- [x] 5. End an abandoned stage at the next stage's start, and clip the rows already swept
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tc "select count(*) from stage_runs s where outcome='abandoned' and exists (select 1 from stage_runs s2 where s2.change_id=s.change_id and s2.id<>s.id and s2.started_at > s.started_at and s2.started_at < s.ended_at)"` prints the overlap count (30 at plan time).
    <!-- measured: that query @ dev store 2026-10-28 → 30 of 110 abandoned rows -->
  - [x] **Step 2: Write the tests** in `stats/internal/store/stageruns_test.go`: `TestSweepAbandonedEndsAtTheNextStageStart` — an open run with a later run of the same change is closed at that later run's `started_at`; an open run with none is closed at the sweep's instant. `TestClipAbandonedStageRunsBoundsWrittenRows` — an `abandoned` row whose `ended_at` lies past a later stage's start is clipped to it, a `completed` row and an `abandoned` row with no later stage are untouched, and the count returned is the rows changed.
  - [x] **Step 3: Fix.** In `stats/internal/store/stageruns.go`, `SweepAbandoned` sets `ended_at = COALESCE((SELECT min(s2.started_at) FROM stage_runs s2 WHERE s2.change_id = stage_runs.change_id AND s2.id <> stage_runs.id AND s2.started_at > stage_runs.started_at), now())`; add `ClipAbandonedStageRuns(ctx) (int64, error)`, one `UPDATE` of `abandoned` rows to `least(ended_at, <that subquery>)` where the subquery's value is earlier; update the doc comment and the note in `stats/internal/harvest/attribute.go` that says the sweep ends at its own instant. Add `flowd clip-abandoned` in `stats/cmd/flowd/main.go`, shaped like `retoken-stage-run`, printing `clipped: <n>`.
  - [x] **Step 4: Prove the tests pin the fix.** With `break-and-prove.sh`, put `ended_at = now()` back — the sweep test fails; make the clip a no-op — the clip test fails; restore.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l .` prints nothing; `cd stats && go test ./internal/store -run 'TestSweepAbandoned|TestClipAbandonedStageRuns' -count=1` passes; `cd stats && make -s web-dist && go vet ./...` exits 0. Commit.
  - [x] **Step 6: Repair the rows** from `<agents-wt>/stats`, after the commit: `make -s web-dist && go run ./cmd/flowd clip-abandoned`; Step 1's query prints `0`. Report both outputs.

**Build:** green
**Files:** `stats/internal/store/stageruns.go`, `stats/internal/store/stageruns_test.go`, `stats/cmd/flowd/main.go`, `stats/internal/harvest/attribute.go`
**Tests:** `TestSweepAbandonedEndsAtTheNextStageStart`, `TestClipAbandonedStageRunsBoundsWrittenRows`
**Regression:** both fail — an abandoned stage ends at the sweep's instant again, overlapping the change's later stages, and no command bounds the rows already written.
**Baseline:** before=56 after=58 <!-- measured: grep -c -E '^func Test' stats/internal/store/stageruns_test.go @ c30fab36 -->
**Commit:** fix(stats): end an abandoned stage at the next stage's start
**After:** none

**Decision:** abandoned-ends-at-next-stage

- [x] 6. Clamp or refuse a dispatch end before its start, and flag the stored ones
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tc "select count(*) from dispatches where ended_at < started_at"` prints the count (20 at plan time), and row 8281 ends 0.816383 s before it starts.
    <!-- measured: that query @ dev store 2026-10-28 → 20; select ended_at - started_at from dispatches where id=8281 → -00:00:00.816383 -->
  - [x] **Step 2: Write the tests.** In `stats/internal/store/records_test.go`: `TestEndDispatchClampsSameSecondEnd` — a row begun at `…:25.8` ended at `…:25` is stored ending at its start; `TestEndDispatchRefusesEndBeforeStart` — an end a second or more earlier is `ErrDispatchEndsBeforeStart` for `completed`, `aborted` and `blocked`, and the row stays open. In `stats/internal/records/render_test.go`: `TestRenderDispatchEndingBeforeStart` — a dispatch whose stored end precedes its start renders `ends before its start`, never a negative duration.
  - [x] **Step 3: Fix.** In `EndDispatch` (`stats/internal/store/records.go`), read the row's `started_at` for every outcome before the update (the existing `completed`-only read folds into it); an `EndedAt` before it but not before `started_at` truncated to the second is replaced by `started_at`; an earlier one returns `ErrDispatchEndsBeforeStart`, a new error naming both instants. Add it to the refusal set in `stats/internal/api/server.go` and `stats/internal/reconcile/reconcile.go` beside `ErrDispatchWindowTooShort`. In `stats/internal/records/render.go`, the `- Ended:` line prints `(ends before its start)` in place of the duration when `EndedAt` precedes `StartedAt`.
  - [x] **Step 4: Prove the tests pin the fix.** With `break-and-prove.sh`, remove the clamp, then the refusal, then the render branch — the matching test fails each time; restore.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l .` prints nothing; `cd stats && go test ./internal/store -run 'TestEndDispatch' -count=1` and `cd stats && go test ./internal/records -run 'TestRenderDispatchEndingBeforeStart' -count=1` pass; `cd stats && make -s web-dist && go vet ./...` exits 0. Commit.
  - [x] **Step 6: Repair row 8281** from `<agents-wt>/stats`, after the commit: `make -s web-dist && go run ./cmd/flowd retime-dispatch 8281 2026-10-28T00:11:25.816383Z 2026-10-28T00:11:25.816383Z ''` (its `commit_sha` is empty); Step 1's count query then prints one fewer. Report both outputs.

**Build:** green
**Files:** `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/internal/records/render.go`, `stats/internal/records/render_test.go`, `stats/internal/api/server.go`, `stats/internal/reconcile/reconcile.go`
**Tests:** `TestEndDispatchClampsSameSecondEnd`, `TestEndDispatchRefusesEndBeforeStart`, `TestRenderDispatchEndingBeforeStart`
**Regression:** all three fail — an end before the start is stored again and rendered as a negative duration.
**Baseline:** before=147 after=150 <!-- measured: cat stats/internal/store/records_test.go stats/internal/records/render_test.go | grep -c -E '^func Test' @ c30fab36 -->
**Commit:** fix(stats): refuse a dispatch end before its start
**After:** Task 4

**Decision:** dispatch-end-before-start

- [x] 7. State the round in the slot prompt, refuse a cross-round reproducer path, repair F8 and F11
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -c -E '— .[0-9]+. here' /Users/tweety53/Projects/gymie-worktrees/kan-1024-templates-warm-up-cool-down-any-workout/.superpowers/sdd/slot-prompt-0-mutation+failure-modes.md` prints `0` — the rendered prompt never states the round; the kan-1024 findings rows F5 and F8 both record `.superpowers/sdd/reproducers/1-mutation-1.sh`, F4 and F11 both `1-test-sanity-1.sh`, at rounds 1 and 0.
  - [x] **Step 2: Write the tests.** `TestRenderSlotPromptStatesTheRound` in `stats/internal/guard/render_slot_prompt_test.go`: a round-`3` render's REPRODUCER paragraph carries `` — `3` here``. `TestCheckPanelReproducersRefusesCrossRoundPath` in `stats/internal/guard/check_panel_reproducers_test.go`: two `fixed` findings at rounds 0 and 1 recording one `reproducers/` path are refused, naming both refs; two findings of one round sharing a path pass.
  - [x] **Step 3: Fix.** In `stats/internal/guard/renderslotprompt.go`, the REPRODUCER block's "`R` this round's number" is replaced by that phrase plus `` — `<round>` here``, the round rendered; the phrase missing from the block is exit 2. In `stats/internal/guard/panelreproducers.go`, after the per-finding loop, refuse every runnable reproducer path under `reproducers/` that findings of two different rounds record, any status. Add the refusal to `scripts/check-panel-reproducers.sh`'s header.
  - [x] **Step 4: Prove the tests pin the fix.** With `break-and-prove.sh`, drop the replacement, then the cross-round check — the matching test fails each time; restore.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestRenderSlotPrompt|TestCheckPanelReproducers' -count=1` passes; `scripts/check-guard-symlinks.sh` and `scripts/check-dispatch-paragraphs.sh` exit 0. Commit.
  - [x] **Step 6: Repair F8 and F11.** Read both rows (`flow record findings -C /Users/tweety53/Projects/gymie-worktrees/kan-1024-templates-warm-up-cool-down-any-workout -change kan-1024-templates-warm-up-cool-down-any-workout`), then rewrite each through `flow record finding -C <that gymie worktree> -change kan-1024-templates-warm-up-cool-down-any-workout -ref F8|F11` with every field it holds unchanged save `-reproducer 'none — overwritten by round 1'"'"'s script of the same name'`. Report the two rows after.

**Build:** green
**Files:** `stats/internal/guard/renderslotprompt.go`, `stats/internal/guard/render_slot_prompt_test.go`, `stats/internal/guard/panelreproducers.go`, `stats/internal/guard/check_panel_reproducers_test.go`, `scripts/check-panel-reproducers.sh`
**Tests:** `TestRenderSlotPromptStatesTheRound`, `TestCheckPanelReproducersRefusesCrossRoundPath`
**Regression:** both fail — a slot is again left to guess its round, and a later round's script silently replaces an earlier round's under one path.
**Baseline:** before=8 after=10 <!-- measured: cat stats/internal/guard/check_panel_reproducers_test.go stats/internal/guard/render_slot_prompt_test.go | grep -c -E '^func Test' @ c30fab36 -->
**Commit:** fix(stats): state the round in slot prompts and refuse cross-round reproducer paths
**After:** none

**Decision:** reproducer-round-stated

- [x] 8. Trace and fix the templates `sq` and offline floor-notice flakes
  - [x] **Step 1: Trace before touching anything; do not guess.** In `<pw-wt>`, run the `verify` row with `<specs>` = `fidelity/templates.spec.ts tests/offline-floor-actions.spec.ts --repeat-each 5 --workers 4` and read each failure's trace (`npx playwright show-trace` or the trace's network and action log): for `fidelity/templates.spec.ts` (the `the list never caught up with `sq`` loop after `page.clock.pauseAt`) name what the `sq` list waits on — a request, a debounce on the paused frame clock, or a recomposition — and for `tests/offline-floor-actions.spec.ts` (`waitForTag(page, 'floor_notice', 30_000)`) name what the notice waits on — the queue's drain, the session read, or a timer. A flake that does not fail in five repeats is rerun at `--repeat-each 20`; one that still does not fail is reported with the runs, never closed unreproduced (**The loop**, `skills/flow/verify-fix-loop.md`).
  - [x] **Step 2: Fix each at its cause.** Make the spec wait on what Step 1 named — a `page.waitForResponse` registered before the key that issues it, a poll of the app state the notice follows (`storedQueueLength`, the session read), or a clock tick that the trace shows the debounce needs — and remove the tick loop. Never a wider loop, a longer timeout, a retry or a sleep.
  - [x] **Step 3: Prove the mechanism reaches the code.** Remove the new wait, run the Step 1 command — the flake reproduces; restore — the same command passes all repeats. Report both runs' tallies.
  - [x] **Step 4: Verify.** `npx tsc --noEmit -p .` in `<pw-wt>` exits 0. Commit.

**Build:** green
**Files:** `fidelity/templates.spec.ts`, `tests/offline-floor-actions.spec.ts`
**Tests:** none — repairs existing tests: `captures the templates step 1 frames` (fidelity/templates.spec.ts), `an offline start into a started session shows the notice and the live clock` (tests/offline-floor-actions.spec.ts)
**Regression:** the two specs fail intermittently again under parallel load.
**Commit:** test: wait on the state the sq list and the floor notice follow
**After:** none

- [x] 9. Capture AD3 over the frame's public rows only
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `<pw-wt>`, publish one stray template the spec does not seed (seed it and move it to `PUBLIC` with the spec file's own `psqlAll` helper, from a scratch spec or `node`), run the `verify` row with `<specs>` = `fidelity/admin-templates.spec.ts` — the `ad3-unpublish-sheet.png` comparison fails; put the stray back to `PRIVATE`.
  - [x] **Step 2: Fix.** In `fidelity/admin-templates.spec.ts`, before the AD3 snap, remove from the DOM every Public-list row whose template name is not one the `ad3-unpublish-sheet` baseline draws (the names the spec seeds and the migration-seeded Gymie templates the frame shows), and assert every name the frame draws is still present. Leave the baseline PNG as it is.
  - [x] **Step 3: Prove it.** Repeat Step 1 with the fix — the comparison passes with the stray present; then without the stray — it passes. Remove the stray.
  - [x] **Step 4: Verify.** `npx tsc --noEmit -p .` in `<pw-wt>` exits 0. Commit.

**Build:** green
**Files:** `fidelity/admin-templates.spec.ts`
**Tests:** none — repairs existing tests: `admin templates queue and editors match AD1–AD7` (fidelity/admin-templates.spec.ts)
**Regression:** any public template in the worktree database lands in the AD3 capture again.
**Commit:** test(fidelity): capture AD3 over the frame's public rows only
**After:** none

- [x] 10. Test the admin-frontend launch environment
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `<gymie-wt>`, `grep -c 'ADMIN_API_URL' buildSrc/src/test/kotlin/com/gymie/gradle/AdminFrontendEnvTest.kt` — the test asserts only the env file; `git show 86d7ebb3 -- gradle/dev-lifecycle.gradle.kts` shows the untested process-env block in `startService`.
  - [x] **Step 2: Write the test** `` `the admin-frontend launch env names the gateway twice` `` in `buildSrc/src/test/kotlin/com/gymie/gradle/AdminFrontendEnvTest.kt`: `adminFrontendLaunchEnv("http://localhost:18080")` returns `ADMIN_API_URL` and `NEXT_PUBLIC_ADMIN_API_URL` both set to that URL and nothing else.
  - [x] **Step 3: Fix.** Add `fun adminFrontendLaunchEnv(gatewayUrl: String): Map<String, String>` to `buildSrc/src/main/kotlin/com/gymie/gradle/AdminFrontendEnvOutcome.kt`, beside `ensureAdminFrontendEnvFile`, carrying the comment from `startService`; `startService` in `gradle/dev-lifecycle.gradle.kts` does `env.putAll(adminFrontendLaunchEnv(context.runtime.urlFor("gateway")))` for `admin-frontend`.
  - [x] **Step 4: Prove the test pins the fix.** Drop one of the two keys from the map — the test fails; restore.
  - [x] **Step 5: Verify.** `./gradlew ktlintFormat`, then `./gradlew verifyChange` and `./gradlew -p buildSrc test --tests 'com.gymie.gradle.AdminFrontendEnvTest'` pass; `./gradlew -q help` exits 0 (the script plugin compiles). Commit.

**Build:** green
**Files:** `buildSrc/src/main/kotlin/com/gymie/gradle/AdminFrontendEnvOutcome.kt`, `buildSrc/src/test/kotlin/com/gymie/gradle/AdminFrontendEnvTest.kt`, `gradle/dev-lifecycle.gradle.kts`
**Tests:** `the admin-frontend launch env names the gateway twice`
**Regression:** the test fails — the panel's process loses a gateway URL and a worktree's admin login posts to the main checkout's port.
**Baseline:** before=5 after=6 <!-- measured: grep -c -w '@Test' buildSrc/src/test/kotlin/com/gymie/gradle/AdminFrontendEnvTest.kt @ e923c6c8 -->
**Commit:** test(dev-lifecycle): cover the admin-frontend launch environment
**After:** none
