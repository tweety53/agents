# self-review-kan-1008-list-mode-step-2-frontend-list-runner-folding

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1008's ten self-review findings, one task each: F1 (Task 1), F2 (Task 2), F3
(Task 3), F4 (Task 4), F5 (Task 5), F6 (Task 6), F7 (Task 7), F8 (Task 8), F9 (Task 9), F10
(Task 10), as `proposal.md` numbers them.

**Architecture:** a directory form of one guard chained before a stage end, an exit-2 path in
another, a slash-joined id pattern in a third, a CLI flag with its store column and a repair
command, close-task's join of the recorded worktrees, two verifier-contract sentences, one
gymie-frontend project-file line, and two traced gymie-playwright flake fixes.

**Tech Stack:** Go (`stats/`), Bash, Markdown, Gradle (gymie-frontend), Playwright (gymie-playwright).

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-23-self-review-kan-1008-list-mode-step-2-frontend-list-runner-folding-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `patches/` and `verbatim-moves.txt` are already committed with the plan. Never merge or push.
- Each task names the repository it runs in; every `**Files:**` path is relative to that worktree's root. Tasks 1, 2 and 4–8 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1008-list-mode-step-2-frontend-list-runner-folding` (`<agents-wt>` below); Task 3 in gymie-frontend's `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding`; Tasks 9 and 10 in gymie-playwright's `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding` — both on branch `spectre/kan-1008-list-mode-step-2-frontend-list-runner-folding`. Never touch a main checkout, and never create a worktree.
- A task's diff is `<agents-wt>/spectre/changes/self-review-kan-1008-list-mode-step-2-frontend-list-runner-folding/patches/task-<n>.diff`, applied with `git apply <that path>` from the task's worktree root, in task order. Every patch was applied and checked in its worktree (agents `50045985`, gymie-frontend `0276408a`) and reverted again; Tasks 1, 2 and 4–8's patches apply together in order. A patch that does not apply is reported, never hand-merged by guesswork.
- Prose lands character for character as its patch carries it: `scripts/check-verbatim-moves.sh` matches every added or reworded sentence against this change's `verbatim-moves.txt` exactly.
- Restore a file a proof step flips with `scripts/break-and-prove.sh` (a pre-mutation byte snapshot), never `git checkout --` over uncommitted edits and never `git stash`: the stash list is shared with every worktree of the repository.
- Go: run `gofmt -w` on every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container.
- A commit in gymie-frontend or gymie-playwright is a commit outside every task of kan-1008's change (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`): add `-m "Flow-Key: self-review-kan-1008-list-mode-step-2-frontend-list-runner-folding-fix-g<g>"`, `<g>` the fixer group of `design.md`'s `fixer-groups`; never touch that repository's `spectre/`.
- The kan-1008 dev stack serves FRONTEND_URL=http://localhost:6740 and the gateway on port 11820; `./gradlew -q devStatus` in `/Users/tweety53/Projects/gymie-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding` reports it, and `./gradlew -q devStackFresh` there answers whether it serves the worktrees' build. Tasks 9 and 10 run against it and never start, stop or rebuild it: a stack that is down or stale is reported to the parent.

## Review Focus

- Task 1: a design naming no check, or a `None`/`No `/`Not applicable` section, passes; integrate's `liveCheckGate` is untouched.
- Task 7: `fold-dispatch` refuses rows of another change or session token and a drop row with an agent id, changing nothing; an `end` without `-diff-base` keeps begin's value.
- Task 8: a single-repository change (recorded set = the canonical worktree) closes exactly as before; a recorded worktree with no commit of the task adds no pair and no push.
- Tasks 9 and 10: no timeout, retry count, sleep or poll interval is raised anywhere.

- [ ] 1. Gate flow.verify's stage end on the live check
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** From `<agents-wt>`, `grep -n -B2 "stage flow.verify -outcome completed" skills/flow/verify-and-handoff.md` shows the stage end with no command before it in its block, and `grep -c "liveCheckGate" stats/internal/guard/integratechange_prepare.go` prints `4` and `grep -rln "icLiveCheckNamed" stats/internal/guard --include='*.go'` names only that file — integrate is the only gate.
    <!-- measured: grep -c "liveCheckGate" stats/internal/guard/integratechange_prepare.go @ 50045985 → 4 -->
  - [ ] **Step 2: Apply `patches/task-1.diff`** from `<agents-wt>`: `stats/internal/guard/liveverification.go` takes a change directory (`LIVE-CHECK-NONE`, a missing record exit 1, no `design.md` exit 2); `TestCheckLiveVerificationChangeRoot` in `stats/internal/guard/check_live_verification_test.go`, whose exit-2 directory case becomes "a directory with no design.md"; the directory form in `scripts/check-live-verification.sh`'s header; **The live check gates the stage end** and the chained block in **Close the stage** of `skills/flow/verify-and-handoff.md`.
  - [ ] **Step 3: Prove the test pins the guard.** Save `git diff -R -- stats/internal/guard/liveverification.go` to a scratch file outside the worktree and run `scripts/break-and-prove.sh stats/internal/guard/liveverification.go --patch <that file> -- sh -c 'cd stats && go test ./internal/guard -run TestCheckLiveVerification -count=1'`: run 1 fails on every `TestCheckLiveVerificationChangeRoot` case, run 2 passes.
    <!-- measured: that break-and-prove run @ 50045985 with the patch → run 1 exited 1 (five ChangeRoot cases FAIL), run 2 exited 0 -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckLiveVerification' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh` and `scripts/check-stage-mark-calls.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/liveverification.go`, `stats/internal/guard/check_live_verification_test.go`, `scripts/check-live-verification.sh`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestCheckLiveVerificationChangeRoot`
**Regression:** `TestCheckLiveVerificationChangeRoot` fails — the change-directory form is read as a record path again and nothing gates the verify stage end.
**Baseline:** before=1 after=2 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_live_verification_test.go @ 50045985 → 1, and 2 with the patch -->
**Commit:** fix(flow-guard): gate flow.verify's stage end on the live check
**After:** none

**Decision:** live-check-gate-at-verify-end

- [ ] 2. Answer exit 2 when the verify-green read finds no row of the change
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** From `/Users/tweety53/Projects/gymie`, `flow record dispatches -change kan-1008-list-mode-step-2-frontend-list-runner-folding -C /Users/tweety53/Projects/gymie-frontend-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding` prints `[]`, while `-C /Users/tweety53/Projects/gymie-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding` prints the rows.
  - [ ] **Step 2: Apply `patches/task-2.diff`** from `<agents-wt>`: `stats/internal/guard/verifygreen.go` exits 2 on a read with no row, naming the canonical worktree; one case in `stats/internal/guard/check_verify_green_test.go`; the exit-2 line of `scripts/check-verify-green.sh`; `skills/flow/visual-verify.md` passes `<canonical-worktree>` and says which one it is.
  - [ ] **Step 3: Prove the case pins the guard.** Save `git diff -R -- stats/internal/guard/verifygreen.go` outside the worktree and run `scripts/break-and-prove.sh stats/internal/guard/verifygreen.go --patch <that file> -- sh -c 'cd stats && go test ./internal/guard -run TestCheckVerifyGreen -count=1'`: run 1 fails on "no row for the change names the canonical worktree" (exit 1, want 2), run 2 passes.
    <!-- measured: that break-and-prove run @ 50045985 with the patch → run 1 exited 1, run 2 exited 0 -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckVerifyGreen' -count=1` passes; `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/verifygreen.go`, `stats/internal/guard/check_verify_green_test.go`, `scripts/check-verify-green.sh`, `skills/flow/visual-verify.md`
**Tests:** `no row for the change names the canonical worktree`
**Regression:** `no row for the change names the canonical worktree` fails — a sibling worktree's empty read reports "recorded no inline verify" again.
**Baseline:** before=8 after=9 <!-- measured: grep -c -E '^[[:space:]]+\{"' stats/internal/guard/check_verify_green_test.go @ 50045985 → 8, and 9 with the patch -->
**Commit:** fix(flow-guard): answer exit 2 when the verify-green read finds no row of the change
**After:** none

**Decision:** verify-green-empty-read-exit-2

- [ ] 3. Leave the iOS test task out of gymie-frontend's declared test line
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In the gymie-frontend worktree, `grep -c 'iosSimulatorArm64Test' .flow/project.md` prints `0`: the declared `## test` line runs `allTests` whole, and at plan time its dry run scheduled `:shared:…IosSimulatorArm64…` tasks.
    <!-- measured: grep -c iosSimulatorArm64Test .flow/project.md and the line's --dry-run piped to grep -i ios @ gymie-frontend 0276408a → 0; iOS tasks listed -->
  - [ ] **Step 2: Apply `patches/task-3.diff`** from the gymie-frontend worktree root: the `## test` line of `.flow/project.md` gains `-x iosSimulatorArm64Test` after `allTests`, and one paragraph under its fence says why.
  - [ ] **Step 3: Verify.** `grep -c -- '-x iosSimulatorArm64Test' .flow/project.md` prints `2` (the line and the paragraph's dry-run command); `git diff --stat` names only `.flow/project.md`. Commit in the gymie-frontend worktree with the `Flow-Key: …-fix-g1` trailer. The patched line's dry run scheduled no iOS task at plan time.
    <!-- measured: the patched line with --dry-run piped to grep -ci ios @ gymie-frontend 0276408a → 0; grep -c -- '-x iosSimulatorArm64Test' on a patched scratch copy → 2 -->

**Build:** green
**Files:** `.flow/project.md`
**Tests:** none — a declared command line; Step 3's dry run is its check
**Regression:** none — the declared `## test` line schedules the failing iOS link again
**Commit:** fix(flow-project): leave the iOS test task out of the declared test line
**After:** none

**Decision:** frontend-test-excludes-ios

- [ ] 4. Never let a found departure shorten the visual verifier's pass
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n "by choice" /Users/tweety53/Projects/gymie-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding/.superpowers/sdd/verify-report-visual-verify-rest-gymie.md` names the "left by choice, not blocked" line, and `grep -c "never shortens the pass" skills/flow/visual-verify-verifier.md` prints `0`.
    <!-- measured: both greps @ 50045985 and the kan-1008 gymie worktree → line 41; 0 -->
  - [ ] **Step 2: Apply `patches/task-4.diff`** from `<agents-wt>`: one bold sentence after "hand back an incomplete report only for a blocker you cannot cure, named." in `skills/flow/visual-verify-verifier.md`.
  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-dispatch-paragraphs.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** none — contract prose
**Regression:** none — documentation only
**Commit:** docs(flow): never let a found departure shorten the visual verifier's pass
**After:** none

**Decision:** departure-never-shortens-pass

- [ ] 5. Read every frame id of a slash-joined per-frame line
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'perFrame := ' stats/internal/guard/framelist.go` shows the pattern ``^\s*- ` + regexp.QuoteMeta(id) + `( [^:]*)?:``, which `- LM-L9 / LM-L11 …:` matches for LM-L9 only.
  - [ ] **Step 2: Apply `patches/task-5.diff`** from `<agents-wt>`: the per-frame pattern admits `([^\s:]+ / )*` before the id; one case in `stats/internal/guard/check_frame_list_test.go`; the `no report line` row of `scripts/check-frame-list.sh`'s header.
  - [ ] **Step 3: Prove the case pins the guard.** Save `git diff -R -- stats/internal/guard/framelist.go` outside the worktree and run `scripts/break-and-prove.sh stats/internal/guard/framelist.go --patch <that file> -- sh -c 'cd stats && go test ./internal/guard -run TestCheckFrameList -count=1'`: run 1 fails with "no report line: M1; no report line: Q2", run 2 passes.
    <!-- measured: that break-and-prove run @ 50045985 with the patch → run 1 exited 1, run 2 exited 0 -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckFrameList' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/internal/guard/framelist.go`, `stats/internal/guard/check_frame_list_test.go`, `scripts/check-frame-list.sh`
**Tests:** `a per-frame line naming frames joined by a slash`
**Regression:** `a per-frame line naming frames joined by a slash` fails — `- LM-L9 / LM-L11 …:` reports LM-L11 missing again.
**Baseline:** before=8 after=9 <!-- measured: grep -c -E '^[[:space:]]+\{"' stats/internal/guard/check_frame_list_test.go @ 50045985 → 8, and 9 with the patch -->
**Commit:** fix(flow-guard): read every frame id of a slash-joined per-frame line
**After:** none

**Decision:** frame-list-slash-ids

- [ ] 6. Record a height motion in a viewport tall enough for its settled height
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c "tall enough to hold" skills/flow/visual-verify-verifier.md` prints `0`: the motion step pins step 8's viewport for every motion.
    <!-- measured: that grep @ 50045985 → 0 -->
  - [ ] **Step 2: Apply `patches/task-6.diff`** from `<agents-wt>`: one bold sentence in the motion step of `skills/flow/visual-verify-verifier.md`, after the kan-995 sampling sentence.
  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-dispatch-paragraphs.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** none — contract prose
**Regression:** none — documentation only
**Commit:** docs(flow): record a height motion in a viewport tall enough for its settled height
**After:** Task 4

**Decision:** tall-viewport-for-height-motion

- [ ] 7. Correct a dispatch's diff base on its end, and fold kan-1008's back-filled row
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** From `/Users/tweety53/Projects/gymie`, `flow record dispatches -change kan-1008-list-mode-step-2-frontend-list-runner-folding | python3 -c 'import json,sys; print([(d["id"], d["agentId"], d.get("diffBase")) for d in json.load(sys.stdin) if d["key"].startswith("visual-verify-fix-3-full-gymie")])'` prints id 7195 with diffBase `aa758b6a` and id 7196 with agent `None` and the four-sha diffBase; `flow record dispatch end -h 2>&1 | grep -c diff-base` prints `0`.
    <!-- measured: that command from /Users/tweety53/Projects/gymie @ dev store 2026-10-23 → (7195, 'a9a8709ccc39e1f2f', 'aa758b6a'), (7196, None, 'aa758b6a,79c0541d…,980d3332…,aa758b6a41d4…') -->
  - [ ] **Step 2: Apply `patches/task-7.diff`** from `<agents-wt>`: `DispatchEnd.DiffBase` (`stats/internal/records/types.go`); `EndDispatch` sets `diff_base = COALESCE($11, d.diff_base)` and the new `FoldDispatch` (`stats/internal/store/records.go`) with `TestEndDispatchCorrectsDiffBase` and `TestFoldDispatchMovesTheDiffBaseAndDropsTheRow` (`stats/internal/store/records_test.go`); `-diff-base` on `dispatch end` (`stats/cmd/flow/record.go`) with `TestDispatchEndDiffBaseCarriedOnWire` (`stats/cmd/flow/record_test.go`); `flowd fold-dispatch` (`stats/cmd/flowd/main.go`); one sentence in **Recording.** of `skills/flow/visual-verify.md`.
  - [ ] **Step 3: Prove the tests pin the code.** With each mutation run through `scripts/break-and-prove.sh` (run 1 fails, run 2 passes): `stats/internal/store/records.go --sed 's/diff_base = COALESCE(\$11, d.diff_base)/diff_base = COALESCE(d.diff_base, $11)/' -- sh -c 'cd stats && go test ./internal/store -run TestEndDispatchCorrectsDiffBase -count=1'`; `stats/internal/store/records.go --sed 's/WHERE id = \$2 AND agent_id IS NULL AND diff_base IS NOT NULL/WHERE id = $2/' -- sh -c 'cd stats && go test ./internal/store -run TestFoldDispatch -count=1'`; and `stats/cmd/flow/record.go --patch <git diff -R -- stats/cmd/flow/record.go, saved outside the worktree> -- sh -c 'cd stats && go test ./cmd/flow -run TestDispatchEndDiffBaseCarriedOnWire -count=1'`.
    <!-- measured: the three break-and-prove runs @ 50045985 with the patch → each run 1 exited 1, run 2 exited 0 -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l .` prints nothing; `cd stats && make -s web-dist && go vet ./...` exits 0; `cd stats && go test ./internal/store -run 'TestEndDispatch|TestFoldDispatch' -count=1` and `cd stats && go test ./cmd/flow -run 'TestDispatchEnd' -count=1` pass; `scripts/check-verbatim-moves.sh` exits 0. Commit.
    <!-- measured: go test ./internal/store, ./cmd/flow, ./cmd/flowd, ./internal/records, ./internal/api, ./internal/reconcile, ./internal/client (whole packages) with the patch @ 50045985 → ok -->
  - [ ] **Step 5: Repair the rows**, after the commit, from `<agents-wt>/stats`: `FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd fold-dispatch 7195 7196` — it talks to the dev store directly and needs no daemon restart. Step 1's command then prints only id 7195, carrying the four-sha diffBase id 7196 had; report both outputs. A refusal (`dispatch not found`) changes nothing: report it, never retry with other ids.
    <!-- measured: grep -n DefaultDSN stats/internal/config/config.go @ 50045985 → postgres://flow:flow@localhost:5433/flow?sslmode=disable -->

**Build:** green
**Files:** `stats/internal/records/types.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/cmd/flowd/main.go`, `skills/flow/visual-verify.md`
**Tests:** `TestEndDispatchCorrectsDiffBase`, `TestFoldDispatchMovesTheDiffBaseAndDropsTheRow`, `TestDispatchEndDiffBaseCarriedOnWire`
**Regression:** `TestEndDispatchCorrectsDiffBase` and `TestDispatchEndDiffBaseCarriedOnWire` fail — `dispatch end` cannot correct a verifier's diff base again; `TestFoldDispatchMovesTheDiffBaseAndDropsTheRow` fails — the repair command is gone or folds rows of another session.
**Baseline:** before=221 after=224 <!-- measured: cat stats/internal/store/records_test.go stats/cmd/flow/record_test.go | grep -c -E '^func Test' @ 50045985 → 221, and 224 with the patch -->
**Commit:** feat(stats): correct a dispatch's diff base on its end, and fold a back-filled correction row
**After:** Task 2

**Decision:** end-diff-base-and-fold

- [ ] 8. Join every recorded worktree's task commit at close
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding`, `git log -1 --format='%h %(trailers:key=Task-Id,valueonly)' 5c19610e` prints `5c19610e 14`, and `grep -c '"state", "get"' <agents-wt>/stats/internal/guard/closetask.go` prints `0` — close-task never reads the recorded set.
    <!-- measured: both commands @ gymie kan-1008 worktree and 50045985 → "5c19610e 14"; 0 -->
  - [ ] **Step 2: Apply `patches/task-8.diff`** from `<agents-wt>`: `ctRecordedWorktrees` and `ctJoinRecorded` in `stats/internal/guard/closetask.go`; the fixture's `flow state get` stub (never logged) and `TestCloseTaskJoinsARecordedWorktreeTheMapLeftOut` in `stats/internal/guard/close_task_test.go`; step 2 of `scripts/close-task.sh`'s header.
  - [ ] **Step 3: Prove the test pins the join.** Save `git diff -R -- stats/internal/guard/closetask.go` outside the worktree and run `scripts/break-and-prove.sh stats/internal/guard/closetask.go --patch <that file> -- sh -c 'cd stats && go test ./internal/guard -run TestCloseTask -count=1'`: run 1 fails on "a peer commit outside Files is refused" and "a peer commit inside Files passes", run 2 passes.
    <!-- measured: that break-and-prove run @ 50045985 with the patch → run 1 exited 1, run 2 exited 0 -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCloseTask' -count=1` passes. Commit.
    <!-- measured: go test ./internal/guard (whole package) with Tasks 1, 2 and 4–8 applied @ 50045985 → ok in 159 s -->

**Build:** green
**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `scripts/close-task.sh`
**Tests:** `TestCloseTaskJoinsARecordedWorktreeTheMapLeftOut`
**Regression:** `TestCloseTaskJoinsARecordedWorktreeTheMapLeftOut` fails — a commit map omitting a recorded worktree leaves that worktree's task commits unjudged again.
**Baseline:** before=6 after=7 <!-- measured: grep -c -E '^func Test' stats/internal/guard/close_task_test.go @ 50045985 → 6, and 7 with the patch -->
**Commit:** fix(flow-guard): join every recorded worktree's task commit at close
**After:** none

**Decision:** close-task-joins-recorded-worktrees

- [ ] 9. Settle the floor's J5a deterministically and drop its widened retries
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Confirm `./gradlew -q devStackFresh` exits 0 in the gymie kan-1008 worktree (else stop and report). In the gymie-playwright worktree, remove `test.describe.configure({ retries: 6 });` from `fidelity/floor-states.spec.ts` (its describe at line 192) and run `FRONTEND_URL=http://localhost:6740 GATEWAY_PORT=11820 npx playwright test fidelity/floor-states.spec.ts -g 'J5a' --repeat-each=10 --workers=5`; if that stays green, run the whole file at the configured workers, then beside the suite's other specs, narrowing per `docs/flake-bisection.md` (gymie) to the smallest set that reproduces. Record the reproducing command, its failure rate and the failing assertion. kan-1008's verify reported J5a (`:553`) flaky, passing on retry.
    <!-- measured: grep -n "retries: 6" fidelity/floor-states.spec.ts @ gymie-playwright 4258134 → line 192 -->
    <!-- unverified:whether the repeated run alone reproduces it -->
  - [ ] **Step 2: Trace it.** With a temporary, uncommitted timing or state log around the step that fails, rerun the reproducing command and name what varies: an unpinned clock or identity, a poll the screenshot races, shared pinned accounts another spec touches (`kan-750-visual` shares them), or every step slow under load. Remove the log.
  - [ ] **Step 3: Fix at the source** the trace named, per `~/.claude/rules/fix-determinism-at-the-source.md` — pin what varies, gate on the app's own drawn state, or give J5a accounts nothing else touches. Never raise a timeout, a retry count, `waitForStableCanvas`'s settings or a sleep, and leave the describe's retries removed. Any other floor test the removal shows flaky is traced and fixed the same way in this task; one the trace cannot reach is reported `big`, uncommitted. Should the trace place the nondeterminism in gymie-frontend's code, stop and report the trace rather than editing that repository.
  - [ ] **Step 4: Prove the fix reaches the code under test.** Run the reproducing command often enough that Step 1's failure rate predicts at least ten failures, and see none; then remove the fix, rerun until it fails, and restore it. Record both figures in the commit body.
  - [ ] **Step 5: Verify.** `FRONTEND_URL=http://localhost:6740 GATEWAY_PORT=11820 npx playwright test fidelity/floor-states.spec.ts` passes with no retry, and `npx playwright test --list` exits 0; a capture the change touches is byte-identical to `HEAD` or named in the commit body. Commit in the gymie-playwright worktree with the `Flow-Key: …-fix-g2` trailer.

**Build:** green
**Files:** `fidelity/floor-states.spec.ts`
**Allowed-collateral:** `fidelity/fixtures.ts`, `fixtures.ts`, `fidelity/floor-states.spec.ts-snapshots/**` — wherever the trace puts the source of the nondeterminism
**Tests:** none — repairs existing tests: J5a, which fails under Step 1's load without the fix
**Regression:** none — J5a's flake returns, and the describe runs without the retries that hid it; Step 4's removal run is the proof
**Commit:** fix(fidelity): settle the floor's J5a deterministically and drop its widened retries
**After:** none

**Decision:** floor-flake-trace-first

- [ ] 10. Settle the kan-749 inset specs' boot under worker load
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** With the stack fresh (as Task 9 Step 1), in the gymie-playwright worktree run `FRONTEND_URL=http://localhost:6740 GATEWAY_PORT=11820 npx playwright test fidelity/kan-749-inset-scroll-runner.spec.ts --repeat-each=10 --workers=5`; if green, beside the suite's other specs at the configured workers, narrowing per `docs/flake-bisection.md` (gymie). kan-1008's verify saw the tests at `:364` and `:446` time out with the page still booting or Home missing the row after a reload; Task 16's implementer reproduced neither (`/Users/tweety53/Projects/gymie-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding/.superpowers/sdd/implementer-report-task-16.md`, "Left, with cause"). Record the reproducing command and its failure rate.
    <!-- unverified:whether any run reproduces it; Task 16's isolated and 24-spec runs did not -->
  - [ ] **Step 2: Trace it.** With a temporary, uncommitted timing log around each boot step the two tests take (`registerVerifiedUserViaApi`, the seeding loop, `page.goto`, the first `waitForTag`), rerun and name the step that spends the budget and why — the gateway's rate limit under seeding, a reload racing the seed, a boot waiting on a resource another worker holds. Remove the log.
  - [ ] **Step 3: Fix at the source** the trace named, per `~/.claude/rules/fix-determinism-at-the-source.md`; never raise a timeout, a retry count or a sleep. Should the trace place the cause in gymie-frontend's or gymie's code, stop and report the trace. Should no run reproduce it, report the runs and their counts, uncommitted, and the parent records the finding as not reproduced.
  - [ ] **Step 4: Prove the fix reaches the code under test** as Task 9 Step 4 does; record both figures in the commit body.
  - [ ] **Step 5: Verify.** `FRONTEND_URL=http://localhost:6740 GATEWAY_PORT=11820 npx playwright test fidelity/kan-749-inset-scroll-runner.spec.ts` passes, and `npx playwright test --list` exits 0. Commit in the gymie-playwright worktree with the `Flow-Key: …-fix-g2` trailer.

**Build:** green
**Files:** `fidelity/kan-749-inset-scroll-runner.spec.ts`
**Allowed-collateral:** `fidelity/fixtures.ts`, `fixtures.ts` — wherever the trace puts the source of the timeout
**Tests:** none — repairs existing tests: the two `kan-749` tests at `:364` and `:446`
**Regression:** none — the boot timeouts under worker load return; Step 4's removal run is the proof
**Commit:** fix(fidelity): settle the kan-749 inset specs' boot under worker load
**After:** Task 9

**Decision:** inset-timeout-trace-first
