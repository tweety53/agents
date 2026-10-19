# self-review-kan-970-fix-the-remaining-known-bugs-md-entries-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the nine non-`big` findings of `kan-970-fix-the-remaining-known-bugs-md-entries-and`'s
self-review pass — two gate guards that misjudged a change, a verifier record outside its change
root, a hook that scanned data as commands and commands as data, two guards misreporting an unsplit
zsh list, three planner and verify doc gaps, and gymie's restart dropping OAuth-port isolation.

**Architecture:** in-process fixes to `check-visual-verify-dispatched` and `check-unfinished-work`
(`stats/internal/guard/`), a `gather-dispatch-context` refusal message, the main-checkout hook's
heredoc and tokenizer scan, three prose edits, one gymie record move, and one buildSrc rule gymie's
`isolateOAuthPorts()` consults. Each task applies its committed patch from `patches/` beside this
file.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-19-self-review-kan-970-fix-the-remaining-known-bugs-md-entries-and-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer — Task 3 makes one commit in each repository, as its steps say. Never write `spectre/changes/` in this agents worktree: this change's `patches/` and `verbatim-moves.txt` are already committed with the plan. Never merge or push.
- Each task names the repository it runs in, by its worktree; every `**Files:**` path is relative to that worktree's root. Tasks 1–6, 8 and 9 run in this agents worktree `/Users/tweety53/Projects/agents-worktrees/kan-970-fix-the-remaining-known-bugs-md-entries-and`; Task 7 and Task 3's record half run in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-970-fix-the-remaining-known-bugs-md-entries-and`. Both are on branch `spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and`; never touch a main checkout, and never create a worktree.
- `P` below is this change root, `/Users/tweety53/Projects/agents-worktrees/kan-970-fix-the-remaining-known-bugs-md-entries-and/spectre/changes/self-review-kan-970-fix-the-remaining-known-bugs-md-entries-and`. Each task's diff is `$P/patches/task-<n>.diff`, applied with `git apply` from its worktree's root, in task order: Task 3's applies on top of Task 1's, Task 5's on top of Task 2's, Task 8's on top of Task 6's. A patch that does not apply is reported, never hand-merged by guesswork.
- Run `cd stats && gofmt -w .` on Go changes before committing; `gofmt -l` and `go vet` stay clean. In gymie run `./gradlew ktlintFormat` first. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 1: a satellite's real UI commit after the verified HEAD is still MISSING — only `<spec root>/changes/` paths are dropped (`TestCheckVisualVerifyDispatchedMissingNamesPathsSinceVerifiedHead` stays green).
- Task 2: a change with a `panel-<round>-…` reviewer row is judged by signal five alone, and an archived change's re-run is not asked (the signal sits inside signal seven's guard).
- Task 3: a Decide skip still answers OK with no record, and the record is looked for under the canonical worktree when one is passed.
- Task 4: every existing deny case of `scripts/test-protect-main-checkout.sh` still denies; a quoted heredoc body is still dropped whole.
- Task 7: a main-checkout run never inherits isolation, and a run after `devStop` binds the declared OAuth ports.

- [x] 1. Drop a change's own records from the visual-verify diff
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, apply the test hunk alone — `git apply --include='stats/internal/guard/check_visual_verify_dispatched_test.go' $P/patches/task-1.diff` — and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatchedIgnoresAChangeRecordMove' -count=1`: it fails, the guard answering VISUAL-VERIFY-MISSING with `spectre/changes/archive/demo/link.md` as the path changed since the verified HEAD.
    <!-- measured: the test hunk alone, go test -run TestCheckVisualVerifyDispatchedIgnoresAChangeRecordMove @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and → FAIL, exit 1 MISSING naming spectre/changes/archive/demo/link.md -->
  - [x] **Step 2: Apply the fix** — the rest: `git apply --exclude='stats/internal/guard/check_visual_verify_dispatched_test.go' $P/patches/task-1.diff` (the guard drops `<spec root>/changes/` paths before the trigger question and the freshness loop; the script header states it).
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatched' -count=1` passes; `scripts/check-references.sh` exits 0.
  - [x] **Step 4: Commit** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go` and `scripts/check-visual-verify-dispatched.sh` with the subject below.

**Build:** green
**Files:** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`
**Tests:** `TestCheckVisualVerifyDispatchedIgnoresAChangeRecordMove`
**Regression:** `TestCheckVisualVerifyDispatchedIgnoresAChangeRecordMove` — reverting the commit reads the archive move of a satellite's `link.md` as a UI path changed since the verified HEAD, and the test fails on MISSING.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/check_visual_verify_dispatched_test.go @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and, before and after patches/task-1.diff -->
**Commit:** fix(guard): drop a change's own records from the visual-verify diff
**After:** none

**Decision:** vvd-drop-change-records

- [x] 2. Report a change with code commits and no review-panel pass
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, apply the test hunk alone — `git apply --include='stats/internal/guard/check_unfinished_work_test.go' $P/patches/task-2.diff` — and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckUnfinishedWorkNoPanelPassRecorded' -count=1`: the code-commit-with-no-pass-row subtest fails CLEAR.
    <!-- measured: the test hunk alone, go test -run TestCheckUnfinishedWorkNoPanelPassRecorded @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and → FAIL on a_code_commit_with_no_pass_row_is_OUTSTANDING -->
  - [x] **Step 2: Apply the fix** — the rest: `git apply --exclude='stats/internal/guard/check_unfinished_work_test.go' $P/patches/task-2.diff` (signal eight in `unfinishedwork.go`, its header paragraph in the script, and a `record passes` answer in `scripts/test-integrate-change.sh`'s stub `flow`).
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1` passes; `bash scripts/test-integrate-change.sh` ends `0 failed`; `bash scripts/test-lib-change-plan.sh` passes.
  - [x] **Step 4: Commit** the four files below with the subject below.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`, `scripts/test-integrate-change.sh`
**Tests:** `TestCheckUnfinishedWorkNoPanelPassRecorded`
**Regression:** `TestCheckUnfinishedWorkNoPanelPassRecorded` — reverting the commit clears a change with a code commit and an empty pass-row read, and the OUTSTANDING subtest fails.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/check_unfinished_work_test.go @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and, before and after patches/task-2.diff -->
**Commit:** fix(guard): report a change with code commits and no review-panel pass
**After:** none

**Decision:** uw-no-pass-signal

- [x] 3. Require the visual verifier's record in the change root
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, with Task 1 committed, apply the test hunks alone — `git apply --include='stats/internal/guard/check_visual_verify_dispatched_test.go' $P/patches/task-3.diff` — and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatchedRequiresTheVerifierRecord' -count=1`: it fails, a verifier row of HEAD answering OK with no record in the change root.
    <!-- measured: the test hunks alone on Task 1's patch, go test -run TestCheckVisualVerifyDispatchedRequiresTheVerifierRecord @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and → FAIL -->
  - [x] **Step 2: Apply the fix** — the rest: `git apply --exclude='stats/internal/guard/check_visual_verify_dispatched_test.go' $P/patches/task-3.diff` (the record check in `visualverifydispatched.go`, its header paragraph, and a committed record in `scripts/test-integrate-change.sh`'s visual fixture).
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatched|TestCheckUnfinishedWork' -count=1` passes; `bash scripts/test-integrate-change.sh` ends `0 failed`.
  - [x] **Step 4: Commit** the four agents files below with the subject below.
  - [x] **Step 5: Move kan-970's record, in the gymie worktree.** From `/Users/tweety53/Projects/gymie-worktrees/kan-970-fix-the-remaining-known-bugs-md-entries-and`, run `bash $P/patches/task-3-record.sh`: it copies every scratchpad file the record cites into `spectre/changes/archive/kan-970-fix-the-remaining-known-bugs-md-entries-and/visual-verification/evidence/`, writes the record there with its citations rewritten to the copies and a provenance note under its title, and removes `.superpowers/sdd/visual-verification.md`. It exits non-zero, writing nothing past the failing copy, when a cited scratchpad file is gone — report that, never re-create the evidence.
    <!-- measured: bash patches/task-3-record.sh on a scratch copy of the gymie record @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and → 26 evidence files copied, no scratchpad path left in the record -->
  - [x] **Step 6: Verify the record half.** `grep -c '/private/tmp' spectre/changes/archive/kan-970-fix-the-remaining-known-bugs-md-entries-and/visual-verification.md` prints `0`; then, from this agents worktree, `scripts/check-visual-verify-dispatched.sh /Users/tweety53/Projects/gymie-frontend-worktrees/kan-970-fix-the-remaining-known-bugs-md-entries-and kan-970-fix-the-remaining-known-bugs-md-entries-and 0a539237 /Users/tweety53/Projects/gymie-worktrees/kan-970-fix-the-remaining-known-bugs-md-entries-and` answers VISUAL-VERIFY-OK.
  - [x] **Step 7: Commit in gymie** `git add spectre/changes/archive/kan-970-fix-the-remaining-known-bugs-md-entries-and/visual-verification.md spectre/changes/archive/kan-970-fix-the-remaining-known-bugs-md-entries-and/visual-verification/evidence` and commit with the subject `chore(spectre): move kan-970's visual record into its change root` (decision `f3-record-half-own-commit`) — that commit touches nothing outside those two paths.

**Build:** green
**Files:** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`, `scripts/test-integrate-change.sh`
**Tests:** `TestCheckVisualVerifyDispatchedRequiresTheVerifierRecord`
**Regression:** `TestCheckVisualVerifyDispatchedRequiresTheVerifierRecord` — reverting the commit answers OK on a verifier row whose change root holds no record, and the test fails on its MISSING line.
**Baseline:** before=6 after=7
<!-- measured: grep -c '^func Test' stats/internal/guard/check_visual_verify_dispatched_test.go @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and, after patches/task-1.diff and after patches/task-3.diff -->
**Commit:** fix(guard): require the visual verifier's record in the change root
**After:** Task 1, 2

**Decision:** vvd-record-required
**Decision:** f3-evidence-in-change-root
**Decision:** f3-record-half-own-commit

- [x] 4. Scan what bash runs, and only that, in the main-checkout hook
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, apply the harness hunks alone — `git apply --include='scripts/test-protect-main-checkout.sh' $P/patches/task-4.diff` — and run `bash scripts/test-protect-main-checkout.sh`: cases 110–113 fail, 110 denied and 111–113 allowed.
    <!-- measured: the harness hunks alone, bash scripts/test-protect-main-checkout.sh @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and → 4 failure(s), cases 110-113 -->
  - [x] **Step 2: Apply the fix** — the rest: `git apply --include='hooks/protect-main-checkout.py' $P/patches/task-4.diff` (`SUBST_SPAN_RE`, the span-only body in `strip_data_heredocs`, and `tokenize`'s appended span tokens).
  - [x] **Step 3: Verify.** `bash scripts/test-protect-main-checkout.sh` prints `all protect-main-checkout cases pass`; `scripts/check-python-suppressions.sh` exits 0.
  - [x] **Step 4: Commit** the two files below with the subject below.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`
**Tests:** `110 an unquoted data heredoc's literal redirect beside a backtick span is data`, `111 a backtick span in an unquoted data heredoc body is still scanned`, `112 a backtick span is scanned as a command`, `113 a command substitution inside a double-quoted word is scanned`
**Regression:** case 110 — reverting the commit keeps the whole unquoted body scanned and denies the write; cases 111–113 — reverting it never scans a backtick span or a quoted `$(…)`, and each is allowed.
**Baseline:** before=97 after=101
<!-- measured: grep -c '^expect ' scripts/test-protect-main-checkout.sh @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and, before and after patches/task-4.diff -->
**Commit:** fix(hooks): scan only what bash runs in a data heredoc, and every substitution elsewhere
**After:** none

**Decision:** hook-subst-spans

- [x] 5. Name an unsplit zsh list when a guard argument holds whitespace
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, with Task 2 committed, apply the test hunks alone — `git apply --include='stats/internal/guard/*_test.go' $P/patches/task-5.diff` — and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckUnfinishedWorkRefusesAnUnsplitRangeList|TestGatherDispatchContextRefusesAnUnsplitTaskIDList' -count=1`: both fail, neither refusal naming zsh.
    <!-- measured: the test hunks alone on Task 2's patch, go test -run of both tests @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and → both FAIL -->
  - [x] **Step 2: Apply the fix** — the rest: `git apply --exclude='stats/internal/guard/*_test.go' $P/patches/task-5.diff` (`unsplitListHint`, the merge-base whitespace refusal, the task-id hint).
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckUnfinishedWork|TestGatherDispatchContext' -count=1` passes.
  - [x] **Step 4: Commit** the four files below with the subject below.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/gatherdispatch.go`, `stats/internal/guard/check_unfinished_work_test.go`, `stats/internal/guard/gatherdispatch_test.go`
**Tests:** `TestCheckUnfinishedWorkRefusesAnUnsplitRangeList`, `TestGatherDispatchContextRefusesAnUnsplitTaskIDList`
**Regression:** `TestCheckUnfinishedWorkRefusesAnUnsplitRangeList` — reverting the commit accepts two ranges as one argument and the refusal names no zsh cause; `TestGatherDispatchContextRefusesAnUnsplitTaskIDList` — reverting it refuses `5 6` without the hint.
**Baseline:** before=6 after=8
<!-- measured: cat stats/internal/guard/check_unfinished_work_test.go stats/internal/guard/gatherdispatch_test.go | grep -c '^func Test' @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and, after patches/task-2.diff and after patches/task-5.diff -->
**Commit:** fix(guard): name an unsplit zsh list when a guard argument holds whitespace
**After:** Task 2

**Decision:** unsplit-list-hint

- [x] 6. State the After list separator
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'Task <ids>' skills/flow/brainstorm-planner.md` shows the value written `Task <ids>` with no separator or example.
  - [x] **Step 2: Apply the edit.** `git apply $P/patches/task-6.diff`.
  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `python3 scripts/check-markdown-integrity.py` exit 0.
  - [x] **Step 4: Commit** the file below with the subject below.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): state the After list separator
**After:** none

**Decision:** after-separator-doc

- [x] 7. Keep a running isolated gymie stack isolated across a restart
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-970-fix-the-remaining-known-bugs-md-entries-and`, run `./gradlew devStatus`: with the isolated stack running it reports "this workspace's recorded block, not its deterministic one" — a run without `-PisolateOAuthPorts` resolves the declared frontend and gateway ports. If no frontend or gateway is running, record that and go on: the buildSrc tests below carry the regression.
  - [x] **Step 2: Apply the fix.** From the gymie worktree root, `git apply $P/patches/task-7.diff` (`inheritsOAuthIsolation` in `FrontendPortContract.kt` with two tests, `isolateOAuthPorts()` consulting it once per invocation, and the false `## run` comment in `.flow/project.md` cut).
  - [x] **Step 3: Verify.** `./gradlew ktlintFormat`, then `./gradlew -p buildSrc test --tests 'com.gymie.gradle.FrontendPortContractTest'` passes; `./gradlew devStatus` prints the "Keeping -PisolateOAuthPorts" line while the isolated stack runs; `./gradlew verifyChange` passes.
  - [x] **Step 4: Commit in gymie** the four files below with the subject below.

**Build:** green
**Files:** `gradle/dev-lifecycle.gradle.kts`, `buildSrc/src/main/kotlin/com/gymie/gradle/FrontendPortContract.kt`, `buildSrc/src/test/kotlin/com/gymie/gradle/FrontendPortContractTest.kt`, `.flow/project.md` (gymie repository, worktree `/Users/tweety53/Projects/gymie-worktrees/kan-970-fix-the-remaining-known-bugs-md-entries-and`)
**Tests:** `a running frontend on a moved port keeps a restart without the flag isolated`, `a stopped or unmoved frontend and gateway leave a run without the flag unisolated`
**Regression:** both tests — reverting the commit removes `inheritsOAuthIsolation`, and `FrontendPortContractTest` no longer compiles; behaviourally, a flagless `devRestart` binds the declared OAuth ports again.
**Baseline:** before=21 after=23
<!-- measured: grep -cE '^\s*@Test$' buildSrc/src/test/kotlin/com/gymie/gradle/FrontendPortContractTest.kt @ branch spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and (gymie), before and after patches/task-7.diff -->
**Commit:** fix(gradle): keep a running isolated stack isolated across a restart
**After:** none

**Decision:** oauth-isolation-inherits

- [x] 8. Check a reported defect against history before designing its fix
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'git log --since' skills/flow/brainstorm-planner.md skills/flow/brainstorm.md` finds nothing: no planner step checks a reported defect against history.
  - [x] **Step 2: Apply the edit.** With Task 6 committed, `git apply $P/patches/task-8.diff`.
  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `python3 scripts/check-markdown-integrity.py` exit 0.
  - [x] **Step 4: Commit** the file below with the subject below.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): check a reported defect against history before designing its fix
**After:** Task 6

**Decision:** history-check-in-section-b

- [x] 9. Keep inline verify's begin and end apart in the recording example
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `skills/flow/verify-and-handoff.md`, the **Recording.** example block's `flow record dispatch begin` and `flow record dispatch end` lines sit adjacent with nothing between them.
  - [x] **Step 2: Apply the edit.** `git apply $P/patches/task-9.diff`.
  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `python3 scripts/check-markdown-integrity.py` exit 0.
  - [x] **Step 4: Commit** the file below with the subject below.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — a doc example; `check-verbatim-moves.sh` covers its acknowledged line.
**Regression:** none — no test is declared.
**Commit:** docs(flow): keep inline verify's begin and end apart in the recording example
**After:** none

**Decision:** inline-verify-example-gap
