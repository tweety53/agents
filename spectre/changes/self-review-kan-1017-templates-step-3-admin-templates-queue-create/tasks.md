# self-review-kan-1017-templates-step-3-admin-templates-queue-create

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1017's fifteen non-big self-review findings as one self-review fix loop, and
repair the records those defects already wrote wrong.

**Architecture:** mechanical refusals and tools at the points a run already passes through — the
panel stage's close, `run-verify-list`, `flow record dispatch begin`, the tree-markers, task-records,
preflight, compose, measure, verify-report and kickoff guards, and a `flow tasks files` verb — each
with one added skill sentence; three ledger rows and one archived plan repaired; one gymie-playwright
sidecar line; one gymie test-harness flake bisected and fixed at its source.

**Tech Stack:** Go (`stats/cmd/flow`, `stats/cmd/flowd`, `stats/internal/guard`, `stats/internal/store`),
Markdown skills; gymie Kotlin/Gradle tests (Task 15).

**Spec:** `design.md` beside this file; the brainstorm trail is
`.superpowers/sdd/2026-10-26-self-review-kan-1017-templates-step-3-admin-templates-queue-create-design.md`
in this worktree.

## Global Constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` of the agents repo — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push, in any repository.
- Agents tasks (1–14) work in `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1017-templates-step-3-admin-templates-queue-create`, branch `self-review-kan-1017-templates-step-3-admin-templates-queue-create`; every agents `**Files:**` path is relative to its root. Task 10's sidecar step works in `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1017-templates-step-3-admin-templates-queue-create`, and Task 15 in `/Users/tweety53/Projects/gymie-worktrees/kan-1017-templates-step-3-admin-templates-queue-create`, both on branch `spectre/kan-1017-templates-step-3-admin-templates-queue-create`; Task 15's `**Files:**` paths are relative to that gymie worktree. Never touch a main checkout, and never create a worktree.
- A prose diff given in a step is applied with `git apply` from the worktree's root after copying the lines between its fence markers into a scratch file outside the worktree; each was applied and checked at `5e45b79a`. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly. No other run-loaded prose (`skills/`, `rules/`, `commands-claude/`) is edited.
- Run `cd stats && gofmt -w .` on Go changes before committing; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container; store tests create and drop their own databases in it, as they already do.
- A step marked **Parent step** is run by the self-review parent session, never the fixer; the fixer skips it and reports nothing for it.
- Gradle (Task 15) runs with a Bash `timeout: 600000`, never in the background, and never concurrently with another Gradle or mutator in that worktree.

## Review Focus

- Task 1: a `flow.review-panel` end with `-outcome stopped`, and every other stage's end, is never gated.
- Task 3: a list that both skipped and ran entries ends `completed` (or `blocked` on a failure), never `skipped`; a non-inline row's floor is unchanged.
- Task 4: a `verify-*` (non-visual) verifier row and a `reviewer` row are never refused; a store read failure never refuses.
- Task 5: before any in-run fix is ticked, and after its `-fix-<k>-full` row completed, an unscoped `visual-verify` key is admitted.
- Task 6: a widening never touches another task's `**Files:**` line, a fenced block, or a path already present.
- Task 8: paths under `.superpowers/` and a snapshot written before this change (no `#porcelain` section) never fail `verify`.

- [ ] 1. Gate the panel stage's close on the findings-closed guard
  - [ ] **Step 1: Write the failing test** `TestStageEndReviewPanelRunsFindingsClosed` in
    `stats/cmd/flow/stage_test.go`, modelled on `TestStageEndFlowVerifyRunsLiveCheck`: a
    `/flow` `flow.review-panel` end `-outcome completed` is refused (exit 1, stderr naming
    `check-panel-findings-closed`) when the guard sees an `open` finding, admitted when every
    finding is closed, and an end with `-outcome stopped` is admitted with the open finding.
    Run it: fails on the first case.
  - [ ] **Step 2: Implement** (decision `sr1-gate-stage-end`) `gatePanelFindingsClosed(f stageIdentityFlags, outcome string, stderr io.Writer) int` in `stats/cmd/flow/stage.go`, called beside `gateLiveVerification`: for `-command '/flow'`, `-stage flow.review-panel`, `-outcome completed` only, it runs `guard.Registry["check-panel-findings-closed"]([]string{<-C dir>, <change>}, guard.Env{Getenv: os.Getenv, Dir: f.dir}, stderr, stderr)` and returns 1 on any non-zero exit with `flow: stage end refused: check-panel-findings-closed exited <n>`.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 22cd06a0..001b2b3a 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -1078,2 +1078,6 @@ the round's `Flow-Key:` trailer.
 
+`flow stage end -stage flow.review-panel -outcome completed` runs `check-panel-findings-closed` in
+process and refuses the close on its exit 1 or 2, so a round closed without the guard never reaches
+integrate (kan-1017: rounds 2–7 closed unguarded, and nine findings went unrecorded).
+
 **Before closing the stage**, the parent runs both close guards:
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./cmd/flow/ -run 'TestStageEnd' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`, `skills/flow/review-panel.md`
**Tests:** `TestStageEndReviewPanelRunsFindingsClosed`
**Regression:** `TestStageEndReviewPanelRunsFindingsClosed` — reverting admits a completed panel close over an open finding.
**Baseline:** before=38 after=39
<!-- measured: grep -c '^func Test' stats/cmd/flow/stage_test.go @ 5e45b79a -->
**Commit:** fix(flow): refuse a panel stage close the findings-closed guard refuses
**After:** none

**Decision:** sr1-gate-stage-end

- [ ] 2. Name the open findings in the late-fix trigger's condition-1 reason
  - [ ] **Step 1: Write the failing test** `TestCheckLateFixTriggerNamesTheOpenFindings` in
    `stats/internal/guard/check_late_fix_trigger_test.go`, using `lftNewFx`, `lftEnv` and the
    `open` fixture: exit 1, and the condition-1 line contains `F1`. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr2-append-stderr`) in `checkLateFixTrigger`
    (`stats/internal/guard/latefixtrigger.go`): case 1 appends `": " + strings.Join(<non-empty trimmed closedErr lines>, "; ")` to the reason when `closedErr` is non-empty.
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckLateFixTrigger' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/latefixtrigger.go`, `stats/internal/guard/check_late_fix_trigger_test.go`
**Tests:** `TestCheckLateFixTriggerNamesTheOpenFindings`
**Regression:** `TestCheckLateFixTriggerNamesTheOpenFindings` — reverting prints condition 1 with no ref.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/check_late_fix_trigger_test.go @ 5e45b79a -->
**Commit:** fix(guard): name the open findings in the late-fix trigger's condition 1
**After:** none

**Decision:** sr2-append-stderr

- [ ] 3. End every verify-list row, admit a skipped one, and repair kan-1017's three rows
  - [ ] **Step 1: Write the failing tests.** `TestRunVerifyListEndsEveryRun` in
    `stats/internal/guard/run_verify_list_test.go`, on the argv-logging fake `flow` of
    `TestRunVerifyListRecordsTheVerifyRow`: a failing list logs `end … -outcome blocked -cause test-failure`;
    an all-skipped list (inputs unchanged since its recorded run) logs `end … -outcome skipped`; a
    green list still logs `-outcome completed`; no `end` carries `-tokens`.
    `TestEndDispatchAdmitsAShortSkippedInlineRow` in `stats/internal/store/records_test.go`: an
    inline row ended 1 s after begin with outcome `skipped` is admitted, and with `completed` still
    fails `ErrInlineWindowTooShort`. `TestSetDispatchOutcome` there: sets outcome and cause on a
    row, clears the cause for a non-`blocked` outcome, `ErrDispatchNotFound` on an unknown id. Run
    them: fail.
  - [ ] **Step 2: Implement** (decision `sr4-skipped-outcome`): in `runverifylist.go`, every
    return after a successful `begin` calls `record("end", …)` with the outcome above (a list that
    skipped some and ran others is `completed` or `blocked`); `skipped` joins `recordOutcomes`
    (`stats/cmd/flow/record.go`); the inline-floor check in `EndDispatch`
    (`stats/internal/store/records.go`) is skipped when the end's outcome is `skipped`; the header
    of `scripts/run-verify-list.sh` states the three outcomes.
  - [ ] **Step 3: Implement** (decision `sr4-repair-rows`) `func (s *Store) SetDispatchOutcome(ctx context.Context, id int64, outcome, cause string) error` in `stats/internal/store/records.go` (`UPDATE dispatches SET outcome = $2, cause = $3`, `nullIfEmpty(cause)`, `ErrDispatchNotFound` on no row), and `flowd reoutcome-dispatch <dispatch-id> <outcome> [<cause>]` in `stats/cmd/flowd/main.go`, beside `retime-dispatch`, refusing an outcome outside the closed set and a cause without `blocked`.
  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestRunVerifyList' -count=1`, `cd stats && go test ./internal/store/ -run 'TestEndDispatch|TestSetDispatchOutcome' -count=1` and `cd stats && go test ./cmd/flow/ -run 'TestRecordDispatchEnd' -count=1` green, `cd stats && go build ./cmd/flowd/` clean; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean.
  - [ ] **Step 5: Commit** the files with the subject below.
  - [ ] **Step 6: Parent step — repair rows 47, 65 and 76** of change `kan-1017-templates-step-3-admin-templates-queue-create` once this task's commit exists, from this worktree's `stats/` with the environment the live `flowd` runs under (`FLOWD_DSN` when set there); `P` is the parent transcript `~/.claude/projects/-Users-tweety53-Projects-gymie/dee709bf-cf6a-4f4e-93d1-c35f55fc18e0.jsonl`:

```bash verified:ids, begins and current ends read from flow record dispatches -change kan-1017-templates-step-3-admin-templates-queue-create at 2026-10-26; retime-dispatch and restamp-dispatch usage read from stats/cmd/flowd/main.go at 5e45b79a; reoutcome-dispatch is this task's
go run ./cmd/flowd retime-dispatch 7820 2026-10-26T12:14:46.949903Z 2026-10-26T12:16:04Z ""
go run ./cmd/flowd reoutcome-dispatch 7820 blocked test-failure
go run ./cmd/flowd restamp-dispatch 7820 "$P"
go run ./cmd/flowd retime-dispatch 7846 2026-10-26T13:53:37.04146Z 2026-10-26T13:53:38Z ""
go run ./cmd/flowd reoutcome-dispatch 7846 skipped
go run ./cmd/flowd restamp-dispatch 7846 "$P"
go run ./cmd/flowd retime-dispatch 7863 2026-10-26T15:00:01.436221Z 2026-10-26T15:00:02Z ""
go run ./cmd/flowd reoutcome-dispatch 7863 skipped
go run ./cmd/flowd restamp-dispatch 7863 "$P"
```

    Then `flow record dispatches -change kan-1017-templates-step-3-admin-templates-queue-create`
    (from the gymie change worktree) shows the three new ends and outcomes. unverified:confirm `restamp-dispatch` accepts a row already marked inline (`MarkDispatchInline` on an inline row) before running it; if it refuses, the retime and reoutcome stand and the row's tokens are recomputed by the harvest.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `stats/cmd/flow/record.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flowd/main.go`, `scripts/run-verify-list.sh`
**Allowed-collateral:** `stats/**/*_test.go`
**Tests:** `TestRunVerifyListEndsEveryRun`, `TestEndDispatchAdmitsAShortSkippedInlineRow`, `TestSetDispatchOutcome`
**Regression:** `TestRunVerifyListEndsEveryRun` — reverting leaves a failed list's row open and ends a skipped one `completed`; `TestEndDispatchAdmitsAShortSkippedInlineRow` — reverting refuses the 1 s skipped row; `TestSetDispatchOutcome` — reverting removes the method.
**Baseline:** before=108 after=111
<!-- measured: cat stats/internal/guard/run_verify_list_test.go stats/internal/store/records_test.go | grep -c '^func Test' @ 5e45b79a -->
**Commit:** fix(stats): end every verify-list row and repair hand-closed ones
**After:** none

**Decision:** sr4-skipped-outcome, sr4-repair-rows

- [ ] 4. Refuse a visual verifier beside an open mutator, and the reverse
  - [ ] **Step 1: Write the failing test** `TestRecordDispatchBeginRefusesAVerifierBesideAnOpenMutator` in `stats/cmd/flow/record_test.go`, modelled on `TestRecordDispatchBeginRefusesAVisualVerifyVerifierWithoutDiffBase` with `genuineDaemon` answering the run-record GET with `{"dispatches":[…]}`: `visual-verify-2` begin refused (exit 2, naming `panel-fix-7`) while `panel-fix-7` (role `panel-fix`) is open; `panel-fix-8` refused while `visual-verify-2` is open; admitted when the open row is `verify-test-fix-1` (role `verifier`) or a `reviewer`, when the rows are ended, and when the run-record read fails. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr5-change-wide`) in `runRecordDispatchBegin`'s refusal block (`stats/cmd/flow/record.go`), reading the run record the way the panel-round-floor block does.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index e8dad51b..0094c3a7 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -407,3 +407,7 @@ agent of any name — may be in flight against a given worktree at any moment**
 one: it reads an immutable commit range, and any number of them may run beside the one mutator.
-Dispatches into different worktrees remain free to run concurrently. This explicitly overrides
+Dispatches into different worktrees remain free to run concurrently. A visual verifier is a
+mutator here too — it builds the stack from the tree and writes `<changeRoot>/visual-verification/`
+— and `flow record dispatch begin` refuses a `visual-verify` row while an `implementer`,
+`panel-fix` or `red-partner` row of the change is open, and any of those while a `visual-verify`
+row is open. This explicitly overrides
 `superpowers:subagent-driven-development`'s parallel dispatch guidance and
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./cmd/flow/ -run 'TestRecordDispatchBegin' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/implement.md`
**Allowed-collateral:** `stats/cmd/flow/*_test.go`
**Tests:** `TestRecordDispatchBeginRefusesAVerifierBesideAnOpenMutator`
**Regression:** `TestRecordDispatchBeginRefusesAVerifierBesideAnOpenMutator` — reverting admits a visual verifier while a panel-fix is open.
**Baseline:** before=136 after=137
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 5e45b79a -->
**Commit:** fix(flow): refuse a visual verifier beside an open mutator
**After:** Task 3

**Decision:** sr5-change-wide

- [ ] 5. Refuse an unscoped or misnamed visual re-run after an in-run fix
  - [ ] **Step 1: Write the failing test** `TestRecordDispatchBeginRefusesAnUnscopedVisualVerifyAfterAnInRunFix` in `stats/cmd/flow/record_test.go`, with a `tasks.md` holding a ticked `- [x] 3. In-run fix 2 — flow.visual-verify: …` and a completed `visual-verify-gymie` row: `visual-verify-3-gymie` refused (no `-fix-2`); `visual-verify-fix-2-gymie` without `-round-base` refused; with `-round-base abc1234` admitted; `visual-verify-fix-2-full-gymie` admitted; `visual-verify-fix-2-frontend` admitted with no completed row ending `-frontend`; and, once a completed `visual-verify-fix-2-full-gymie` row exists, `visual-verify-4-gymie` admitted. With no ticked in-run fix, `visual-verify-2-gymie` admitted. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr17-tasks-md-read`) in `runRecordDispatchBegin`, reading the plan through `planPath` and the title pattern `^- \[x\] [0-9]+\. In-run fix ([0-9]+) — ` outside fences.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index 754dcebc..f4d02e26 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -165,2 +165,6 @@ names, and `check-visual-verify-dispatched.sh` refuses a worktree whose newest c
 row carries one — the closing unscoped run is not recorded.
+After an `In-run fix <k>` task is ticked, and until a `-fix-<k>-full` verifier row has completed,
+`flow record dispatch begin` refuses a `visual-verify` key that does not carry `-fix-<k>`, and one
+that is not the `-full` run, carries no `-round-base` and shares its repository suffix with a
+completed verifier row.
 Its clean report is the one the handoff reports.
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./cmd/flow/ -run 'TestRecordDispatchBegin' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/verify-fix-loop.md`
**Allowed-collateral:** `stats/cmd/flow/*_test.go`
**Tests:** `TestRecordDispatchBeginRefusesAnUnscopedVisualVerifyAfterAnInRunFix`
**Regression:** `TestRecordDispatchBeginRefusesAnUnscopedVisualVerifyAfterAnInRunFix` — reverting admits `visual-verify-3-gymie` after In-run fix 2.
**Baseline:** before=137 after=138
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 5e45b79a (136, plus Task 4's one test) -->
**Commit:** fix(flow): refuse an unscoped visual re-run after an in-run fix
**After:** Task 4

**Decision:** sr17-tasks-md-read

- [ ] 6. Widen one task's Files with `flow tasks files`
  - [ ] **Step 1: Write the failing test** `TestTasksFilesAddWidensOneTask` in `stats/cmd/flow/tasks_test.go`: on a three-task plan whose tasks 1 and 3 carry the same `**Files:**` text, `flow tasks files -C <dir> <change> 1 add a.go b.go` adds both to task 1 only, a second run adds nothing (dedupe), `**Files:** none — x` becomes `**Files:** \`c.go\``, a `**Files:**`-looking line inside a fence is untouched, every other byte is unchanged, and an unknown id exits 1. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr16-files-verb`) `filesAdd(plan []byte, id string, paths []string) ([]byte, error)` in `stats/cmd/flow/tick.go`, sharing the task-body scan with `tickTask` through one helper, and `case "files"` in `runTasks` (`stats/cmd/flow/tasks.go`) with `tasksUsage` extended.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index e8dad51b..ddf560e0 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -849,3 +849,5 @@ against the record as it stood, refuses the commit, and names every undeclared p
 judges the deviation legitimate or not on exactly that refusal, and only a legitimate one is
-transcribed and the guard re-run green. The review gate's undeclared-path arm reads the paths the
+transcribed and the guard re-run green. The parent transcribes a widening with
+`flow tasks files -C <worktree> <change> <id> add <path>…`, which edits that one task's
+`**Files:**` field, never with a hand `sed` over `tasks.md`. The review gate's undeclared-path arm reads the paths the
 refusal named, so the disclosure cannot disarm the gate. This is expected practice
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./cmd/flow/ -run 'TestTasks|TestTick' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/tasks.go`, `stats/cmd/flow/tick.go`, `stats/cmd/flow/tasks_test.go`, `skills/flow/implement.md`
**Tests:** `TestTasksFilesAddWidensOneTask`
**Regression:** `TestTasksFilesAddWidensOneTask` — reverting makes `flow tasks files` an unknown verb.
**Baseline:** before=21 after=22
<!-- measured: grep -c '^func Test' stats/cmd/flow/tasks_test.go @ 5e45b79a -->
**Commit:** feat(flow): widen one task's Files with flow tasks files
**After:** Task 4

**Decision:** sr16-files-verb

- [ ] 7. Refuse a ticked task with an unticked step, and tick kan-1017's steps
  - [ ] **Step 1: Write the failing test** `TestCheckTaskRecordsRefusesATickedTaskWithAnUntickedStep` in `stats/internal/guard/check_task_records_test.go`, on `trRepo` and `fx.task(1, "x", …, "  - [ ] **Step 1: a**")`: exit 1 naming task 1 and `unticked step`; with `"  - [x] **Step 1: a**"` the violation is absent; a `  - [ ] ` line inside a fence is ignored. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr7-guard-and-repair`) in `trCheck` (`stats/internal/guard/taskrecords.go`), per ticked task via `tcfSelectTask` and `tcfUnfenced`.
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckTaskRecords' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-task-records.sh` exits 0 on this worktree.
  - [ ] **Step 4: Commit** the two files with the subject below.
  - [ ] **Step 5: Parent step — tick kan-1017's 27 step lines** in the gymie change worktree `/Users/tweety53/Projects/gymie-worktrees/kan-1017-templates-step-3-admin-templates-queue-create`, branch `spectre/kan-1017-templates-step-3-admin-templates-queue-create`, while no fixer works in that worktree: in its `spectre/changes/archive/kan-1017-templates-step-3-admin-templates-queue-create/tasks.md`, run `sed -i '' '231,322s/^  - \[ \] \*\*Step/  - [x] **Step/'` on that file, confirm `git diff --numstat` reports 27 changed lines, and commit that one file there as `chore(spectre): tick the steps of kan-1017 tasks 11-16`. Then this task's build of `scripts/check-task-records.sh` on that worktree exits 0.
    <!-- measured: awk count of '  - [ ] ' lines under ticked tasks in that archived tasks.md = 27, at lines 231-322 @ gymie kan-1017 worktree b18bf74d -->

**Build:** green
**Files:** `stats/internal/guard/taskrecords.go`, `stats/internal/guard/check_task_records_test.go`
**Tests:** `TestCheckTaskRecordsRefusesATickedTaskWithAnUntickedStep`
**Regression:** `TestCheckTaskRecordsRefusesATickedTaskWithAnUntickedStep` — reverting passes a ticked task over an unticked step.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/check_task_records_test.go @ 5e45b79a -->
**Commit:** fix(guard): refuse a ticked task whose steps are unticked
**After:** none

**Decision:** sr7-guard-and-repair

- [ ] 8. Fail tree-markers verify on a path a slot left behind
  - [ ] **Step 1: Write the failing test** `TestCheckTreeMarkersFailsOnAStrayPath` in `stats/internal/guard/check_tree_markers_test.go`: `snapshot`, then an untracked `components/X.tsx.x` → `verify` exit 1 naming it; a modified tracked file → exit 1 naming it; a new file under `.superpowers/sdd/` → exit 0; a snapshot file with no `#porcelain` section → exit 0 with the stray present. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr6-porcelain-in-tree-markers`) in `checkTreeMarkers` (`stats/internal/guard/treemarkers.go`) and the header of `scripts/check-tree-markers.sh`.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 22cd06a0..c9ff61f2 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -475,3 +475,5 @@ artifacts and working notes the round reads, `check-tree-markers.sh snapshot <wo
 `check-tree-markers.sh verify <worktree> <markers-file> <snapshot-file>` runs before any report
-is read or finding recorded.
+is read or finding recorded. `verify` also compares the worktree's `git status --porcelain` with the one `snapshot`
+recorded, paths under `.superpowers/` aside, and fails naming every path a slot left changed or
+untracked — a reproducer's scratch copy included (kan-1017).
 
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckTreeMarkers' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/treemarkers.go`, `stats/internal/guard/check_tree_markers_test.go`, `scripts/check-tree-markers.sh`, `skills/flow/review-panel.md`
**Tests:** `TestCheckTreeMarkersFailsOnAStrayPath`
**Regression:** `TestCheckTreeMarkersFailsOnAStrayPath` — reverting passes `verify` with a stray untracked file.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/check_tree_markers_test.go @ 5e45b79a -->
**Commit:** fix(guard): fail tree-markers verify on a path a slot left behind
**After:** Task 1

**Decision:** sr6-porcelain-in-tree-markers

- [ ] 9. Fail visual preflight for a flowBase change missing an apps worktree
  - [ ] **Step 1: Write the failing test** `TestCheckVisualPreflightFlowBaseSiblings` in `stats/internal/guard/check_visual_preflight_test.go`, modelled on `TestCheckVisualPreflightCrossRepo`: a worktree on `spectre/x` with `branch.spectre/x.flowBase` set and a second `## apps` repository with no `<repo>-worktrees/x` → exit 1 with `FAIL: siblings`; the sibling present → no such line; no flowBase recorded → no such line. Run it: fails.
  - [ ] **Step 2: Implement** a fifth check in `checkVisualPreflight` (`stats/internal/guard/visualpreflight.go`) after the bases loop, reading `## apps` with `tcfAppRoots(projectSection(…, "apps"))` and the flowBase with `git config --get branch.<branch>.flowBase`.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index fec28777..09b37af6 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -219,3 +219,5 @@ verifier reports on a `wire contract` line blocks as a defect in a captured scre
    default branch: kan-1022's stack served the `develop` frontend to a change based on
-   feature/templates, and 28 specs failed on it. With the
+   feature/templates, and 28 specs failed on it. `check-visual-preflight.sh` fails a
+   worktree whose branch records a flowBase while an `## apps` repository holds no
+   `<repo>-worktrees/<name>` worktree. With the
    parent's own Bash call, before the
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckVisualPreflight' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/visualpreflight.go`, `stats/internal/guard/check_visual_preflight_test.go`, `skills/flow/visual-verify.md`
**Allowed-collateral:** `scripts/check-visual-preflight.sh`
**Tests:** `TestCheckVisualPreflightFlowBaseSiblings`
**Regression:** `TestCheckVisualPreflightFlowBaseSiblings` — reverting passes preflight with the apps sibling missing.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/check_visual_preflight_test.go @ 5e45b79a -->
**Commit:** fix(guard): fail visual preflight for a flowBase change missing an apps worktree
**After:** none

**Decision:** none

- [ ] 10. Let a mockups sidecar override the frame geometry
  - [ ] **Step 1: Write the failing test** `TestComposeMockupFramesSidecarFrameLine` in `stats/internal/guard/compose_mockup_frames_test.go`, on `cmfRun` and `cmfSize`: a sidecar opening `frame none` composes a frame at native size although the fourth argument is `scale=2 status=26 border=1`; `frame scale=1 status=0 border=0` overrides the argument; `frame bogus` exits 2; a `frame` line after a map line is a finding (exit 1). Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr8-sidecar-frame`) in `composeMockupFrames` (`stats/internal/guard/composemockupframes.go`) and the header of `scripts/compose-mockup-frames.sh`.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow-contracts/project-configuration-visual.md b/skills/flow-contracts/project-configuration-visual.md
index f334150b..eb43c82a 100644
--- a/skills/flow-contracts/project-configuration-visual.md
+++ b/skills/flow-contracts/project-configuration-visual.md
@@ -24,3 +24,3 @@ rows below is reported and its row dropped, never silently ignored.
 | `mockups` | no | Directory holding one `<frame id>.png` per drawn frame, relative to the project root. Declared, it enables the compose step of `flow.visual-verify` for any capture spec carrying a `<spec>.mockups` sidecar; absent, the stage composes nothing and reports `mockups: not declared`. |
-| `mockup frame` | no | `scale=<int> status=<px> border=<px>`: the mockup PNGs' scale factor, the status-line height and the border width, the last two in logical px. Declared, the compose step crops each frame to its content area — the sides and the top from these values, the bottom found per frame as the first row below the status line every one of whose pixels between the borders equals the page colour at (0, 0) — and reports a capture whose size differs from the cropped frame as a finding rather than composing it. Absent, frames compose at native size, uncropped, with no size check. |
+| `mockup frame` | no | `scale=<int> status=<px> border=<px>`: the mockup PNGs' scale factor, the status-line height and the border width, the last two in logical px. Declared, the compose step crops each frame to its content area — the sides and the top from these values, the bottom found per frame as the first row below the status line every one of whose pixels between the borders equals the page colour at (0, 0) — and reports a capture whose size differs from the cropped frame as a finding rather than composing it. Absent, frames compose at native size, uncropped, with no size check. A `<spec>.mockups` sidecar whose first non-comment line is `frame <geometry>` or `frame none` overrides this value for its own frames, `none` composing them at native size, uncropped. |
 
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index aa1d2334..7d29ac30 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -152,3 +152,3 @@ never an exit code the command did not return.
    The fourth argument is this project's resolved `mockup frame` value, quoted; omit it entirely
-   when the row is absent.
+   when the row is absent. A sidecar's own first `frame` line overrides it.
 
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestComposeMockupFrames' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the five files with the subject below.
  - [ ] **Step 6: Sidecar line in gymie-playwright.** In `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1017-templates-step-3-admin-templates-queue-create` (branch `spectre/kan-1017-templates-step-3-admin-templates-queue-create`), insert `frame none` as the first non-comment line of `fidelity/admin-templates.spec.ts.mockups`; run this worktree's `scripts/resolve-visual-screenshots.sh <that worktree> admin-templates.spec.ts | scripts/compose-mockup-frames.sh <sidecar> <gymie change worktree>/docs/design/mockups <scratchpad dir> 'scale=2 status=26 border=1'` and confirm exit 0 with every AD frame composed at native size; then commit that one file there as `chore(fidelity): compose the admin template frames at native size`. unverified:confirm the sidecar's map lines and the gymie `mockups` root match the command above before running it.

**Build:** green
**Files:** `stats/internal/guard/composemockupframes.go`, `stats/internal/guard/compose_mockup_frames_test.go`, `scripts/compose-mockup-frames.sh`, `skills/flow-contracts/project-configuration-visual.md`, `skills/flow/visual-verify-verifier.md`
**Tests:** `TestComposeMockupFramesSidecarFrameLine`
**Regression:** `TestComposeMockupFramesSidecarFrameLine` — reverting reads `frame none` as a malformed map line.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/compose_mockup_frames_test.go @ 5e45b79a -->
**Commit:** feat(guard): let a mockups sidecar override the frame geometry
**After:** none

**Decision:** sr8-sidecar-frame

- [ ] 11. Fill the element matrix in one call and clip regions to the image
  - [ ] **Step 1: Write the failing test** `TestMeasureVisualPropertiesMatrix` in `stats/internal/guard/measure_visual_properties_test.go`, on `mvpPNG`, `mvpRun`, `mvpJSON`: two synthetic images with a heading, a bordered box and a rule → `--props matrix` returns one row per paired element with that kind's columns filled; a region `0,795,100,7` on an 800 px image returns a clipped read (exit 0, clipped side reported) instead of exit 2; a region starting at y=800 still exits 2. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr11-matrix-and-clip`) in `stats/internal/guard/measurevisualproperties.go` (`mvpOptIn`, `mvpPageProps`, `measure`, `mvpOpen`) and the header of `scripts/measure-visual-properties.sh`.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index aa1d2334..e4f2090b 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -414,3 +414,5 @@ never an exit code the command did not return.
       (kan-1015: the band read took each scrim-dimmed frame as one band, and every matrix was
-      measured with a hand-written script).
+      measured with a hand-written script). `measure-visual-properties.sh <frame> <capture> --scale <f> --props matrix` then
+      fills every column below for every paired element in one call, each region clipped to its
+      image; a row it does not pair takes the per-element calls.
    2. **Fill every column for every row from the script's own output**, one call per element:
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestMeasureVisualProperties' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/measurevisualproperties.go`, `stats/internal/guard/measure_visual_properties_test.go`, `scripts/measure-visual-properties.sh`, `skills/flow/visual-verify-verifier.md`
**Tests:** `TestMeasureVisualPropertiesMatrix`
**Regression:** `TestMeasureVisualPropertiesMatrix` — reverting refuses `--props matrix` and exits 2 on the edge band.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^func Test' stats/internal/guard/measure_visual_properties_test.go @ 5e45b79a -->
**Commit:** feat(guard): fill the element matrix in one measure call
**After:** Task 10

**Decision:** sr11-matrix-and-clip

- [ ] 12. Report a band's darkest ink and its top-two share
  - [ ] **Step 1: Write the failing test** `TestMeasureVisualPropertiesBandColourDarkestInk` in `stats/internal/guard/measure_visual_properties_test.go`: a band whose pixels are 60% grey border `#e2e8f0` and 10% black text `#020817` reports `colour` `#020817` and the top-two `share`; a colour under 2% of the band (anti-aliasing) never wins. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr13-darkest-ink`) `mvpCounter.darkest()` and the `share` field in `measureBands`, `mvpBand.obj` and `mvpBandRecs` (`stats/internal/guard/measurevisualproperties.go`); the header of `scripts/measure-visual-properties.sh` states the field.
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard/ -run 'TestMeasureVisualProperties' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean.
  - [ ] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/measurevisualproperties.go`, `stats/internal/guard/measure_visual_properties_test.go`, `scripts/measure-visual-properties.sh`
**Tests:** `TestMeasureVisualPropertiesBandColourDarkestInk`
**Regression:** `TestMeasureVisualPropertiesBandColourDarkestInk` — reverting reports the grey border as the band colour.
**Baseline:** before=4 after=5
<!-- measured: grep -c '^func Test' stats/internal/guard/measure_visual_properties_test.go @ 5e45b79a (3, plus Task 11's one test) -->
**Commit:** fix(guard): report a band's darkest ink, not its modal colour
**After:** Task 11

**Decision:** sr13-darkest-ink

- [ ] 13. Name the mockup sources and refuse `no mockup source` when one exists
  - [ ] **Step 1: Write the failing test** `TestCheckVerifyReportNoMockupSourceWithCitedSource` in `stats/internal/guard/check_verify_report_test.go`, on `TestCheckVerifyReportFingerprintReason`'s temp-file setup with a worktree holding `<wt>/.superpowers/sdd/<report>`, `<changeRoot>/design.md` citing `` `handoff/Flow.html` `` and that file present: `--design <changeRoot>/design.md` with `no mockup source` in the report, or in `<changeRoot>/visual-verification.md`, exits 1 naming step 10 and the path; the cited file absent → that step is not reported; no `--design` → behaviour unchanged. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr12-design-flag`) the optional `--design` flag in `checkVerifyReport` (`stats/internal/guard/verifyreport.go`), keeping the three positionals, and the usage line of `scripts/check-verify-report.sh`.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; the embedded one-liner decoded the kan-1017 handoff page; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index fec28777..33545ea9 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -188,2 +188,19 @@ attribution starts, committing and pushing nothing.
 
+**The parent names the mockup sources in the prompt** — `mockup sources:`, then every HTML or CSS
+path `design.md` cites that exists in the tree, or `mockup sources: none`. A bundled page, one
+carrying `<script type="__bundler/template">`, is decoded into the verifier's scratchpad first:
+
+```bash
+python3 -I -c 'import sys,re,json,base64,gzip,os
+s=open(sys.argv[1],encoding="utf-8").read(); o=sys.argv[2]; os.makedirs(o,exist_ok=True)
+g=lambda t: json.loads(re.search(r"<script type=\"__bundler/"+t+r"\">\s*(.*?)\s*</script>",s,re.S).group(1))
+open(o+"/index.html","w").write(g("template"))
+for k,v in g("manifest").items():
+ b=base64.b64decode(v["data"]); open(o+"/"+k,"wb").write(gzip.decompress(b) if v["compressed"] else b)' "<page>" <scratchpad dir>
+```
+
+The verifier runs `check-verify-report.sh` with `--design <changeRoot>/design.md`, which refuses a
+report, or the `visual-verification.md` beside `design.md`, reading `no mockup source` while
+`design.md` cites a source that exists (kan-1017: every verifier left 7 of 11 matrix columns n/a).
+
 **The parent names the change's motions in the prompt** — `motions:`, then one `<motion id>:
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`, `skills/flow/visual-verify.md`
**Tests:** `TestCheckVerifyReportNoMockupSourceWithCitedSource`
**Regression:** `TestCheckVerifyReportNoMockupSourceWithCitedSource` — reverting accepts `no mockup source` beside a cited, present source.
**Baseline:** before=4 after=5
<!-- measured: grep -c '^func Test' stats/internal/guard/check_verify_report_test.go @ 5e45b79a -->
**Commit:** fix(guard): refuse no mockup source while design.md cites one
**After:** Task 9

**Decision:** sr12-design-flag

- [ ] 14. Write the merge-order record from kickoff `--from`
  - [ ] **Step 1: Write the failing test** `TestKickoffWorktreeFromWritesMergeOrder` in `stats/internal/guard/kickoff_worktree_test.go`, on the `--from` subtests' fixtures with the canonical worktree holding `spectre/peers` naming the peer and `spectre/changes/<name>/`: no `link.md` → one is written with `## Merge order`, `` 1. `.` ``, `` 2. `<peer>` ``; a `link.md` already naming `.` and another peer → the new peer is appended as the next number, nothing reordered; already named → file unchanged; the peer missing from `spectre/peers` → exit 2 and no worktree recorded; no change directory → no file written. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr15-kickoff-writes-merge-order`) in `stats/internal/guard/kickoffworktree.go`'s `--from` path and the header of `scripts/kickoff-worktree.sh`. unverified:confirm the peers file lives at `<worktree>/spectre/peers` with one name per line before relying on it.
  - [ ] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at 5e45b79a; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/cross-repo-worktrees.md b/skills/flow/cross-repo-worktrees.md
index 0dd5749e..afeaaef3 100644
--- a/skills/flow/cross-repo-worktrees.md
+++ b/skills/flow/cross-repo-worktrees.md
@@ -68,3 +68,5 @@ worktree holding the file, behind `check-planning-commit-location.sh <abs-worktr
 (**Planning commits**, `skills/flow-contracts/git-boundaries.md`), and what it wrote is recorded
-in the working notes beside the merge bases. A change with one repository runs nothing here.
+in the working notes beside the merge bases. `kickoff-worktree.sh --from <canonical worktree>` makes
+this record itself once that worktree holds `<changeRoot>` — writing the `link.md` in the shape
+below, or appending its repository's peers name — and leaves the planning commit to its caller. A change with one repository runs nothing here.
 
````

  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestKickoffWorktree' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/kickoffworktree.go`, `stats/internal/guard/kickoff_worktree_test.go`, `scripts/kickoff-worktree.sh`, `skills/flow/cross-repo-worktrees.md`
**Tests:** `TestKickoffWorktreeFromWritesMergeOrder`
**Regression:** `TestKickoffWorktreeFromWritesMergeOrder` — reverting writes no `link.md`.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^func Test' stats/internal/guard/kickoff_worktree_test.go @ 5e45b79a -->
**Commit:** feat(guard): write the merge-order record from kickoff --from
**After:** none

**Decision:** sr15-kickoff-writes-merge-order

- [ ] 15. Bisect and fix gymie's Redis handshake flake at its source
  - [ ] **Step 1: Read the procedure.** `docs/flake-bisection.md` of the gymie change worktree and `~/.claude/rules/fix-determinism-at-the-source.md`, in full. Work only in `/Users/tweety53/Projects/gymie-worktrees/kan-1017-templates-step-3-admin-templates-queue-create`, branch `spectre/kan-1017-templates-step-3-admin-templates-queue-create`; never run the full suite without stating the reason first.
  - [ ] **Step 2: Reproduce.** `./gradlew :app:test --tests 'com.gymie.AuthFlowIntegrationTest'` alone, repeated; then `--tests` pairs with each sibling class in the failing run's class order, repeated, until one reproduces `RedisConnectionException … RedisHandshakeHandler` from `support.cleanup()`. Record each command and its result.
  - [ ] **Step 3: Find the source.** With the reproducing unit, test the visible candidates one at a time — `spring.data.redis.password ""` sending an `AUTH` the passwordless container refuses; the shared container's default port wait admitting a connection before `Ready to accept connections`; cached Spring contexts' Lettuce pools against the one container — by the Lettuce debug log or by removing one candidate and re-running the unit.
  - [ ] **Step 4: Fix at the source** in `src/app/src/test/kotlin/com/gymie/AppIntegrationTest.kt` (or the support file the source lives in) — never a wider timeout, retry or sleep.
  - [ ] **Step 5: Prove it.** The reproducing unit passes on repeated runs with the fix and fails with the fix reverted; record both. `./gradlew ktlintFormat` then `./gradlew verifyChange` green.
  - [ ] **Step 6: Commit** in that gymie worktree with the subject below and `-m "Flow-Key: self-review-kan-1017-templates-step-3-admin-templates-queue-create-fix-g3"`. unverified:confirm gymie's untasked-commit gate admits that trailer; if it refuses, report it rather than changing the subject's type. When no unit reproduces within ten repeated runs of the class alone and of each pair, commit nothing and report every command and result as the evidence.

**Build:** green
**Files:** `src/app/src/test/kotlin/com/gymie/AppIntegrationTest.kt`
**Allowed-collateral:** `src/app/src/test/kotlin/com/gymie/**`, `src/app/src/test/resources/**`
**Tests:** **none** — the proof is the reproducing `--tests` unit failing with the fix reverted and passing with it, recorded in the report.
**Regression:** none — no test is declared.
**Commit:** fix(app): stop the shared test Redis refusing a fresh handshake
**After:** none

**Decision:** sr14-bisect-first
