# self-review-kan-1061-flow-add-flow-interactive-fast-conversational

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1061's four non-big self-review findings and write its two prompt-tuning
cases as one self-review fix loop, and repair the ledger rows the slot defect already wrote wrong.

**Architecture:** a slot check and two printed launch lines in `flow record dispatch begin`, a
`flowd reslot-dispatch` repair command, one planner sentence, an incident write in the two dispatch
hooks' deny path, and two rendered prompt-tuning case folders.

**Tech Stack:** Go (`stats/cmd/flow`, `stats/cmd/flowd`, `stats/internal/records`,
`stats/internal/store`, `stats/internal/repotest`), Python hooks, Markdown skills.

**Spec:** `design.md` beside this file; the brainstorm trail is
`.superpowers/sdd/2026-10-30-self-review-kan-1061-flow-add-flow-interactive-fast-conversational-design.md`
in this worktree.

## Global Constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational`, branch `spectre/kan-1061-flow-add-flow-interactive-fast-conversational`; every `**Files:**` path is relative to its root. Never touch a main checkout, never create a worktree, never merge or push.
- One commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`.
- Run `cd stats && gofmt -w .` on Go changes before committing; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` stay clean.
- Never start, stop or restart the dev `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container. Store tests create and drop their own databases in it, as they already do. Task 1's repair step writes five rows of the dev `flow` database through `flowd reslot-dispatch`, and nothing else in it.
- A task whose `**Files:**` holds a Markdown file under `skills/` runs `scripts/run-verify-list.sh <worktree> lint` in its verify step.

## Review Focus

- Task 1: a contract key with a worktree or attempt suffix (`panel-9-primary-<repo>`, `panel-6-mutation+failure-modes-rerun`, `-retry`, `-cite`) under its own slot is admitted; a non-reviewer row and a non-panel key are never checked.
- Task 3: `recorded: dispatch <seq>` stays the first stdout line; a back-fill and an inline begin print nothing more; no Role line is printed for a panel reviewer row.
- Task 4: the deny JSON is the hook's whole stdout and is written whatever the incident call does; no test writes the dev store.

- [ ] 1. Refuse a panel row whose slot is not its key's, and repair kan-1061's five rows
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** From `stats/`, `go build -o /tmp/flow-sr1061 ./cmd/flow`, then against a fake daemon (the `genuineDaemon` helper of `stats/cmd/flow/record_test.go`) or by reading `runRecordDispatchBegin`, confirm a `-role reviewer -key panel-3-test-sanity+call-efficiency -slot test-sanity` begin is admitted today.
  - [ ] **Step 2: Write the failing tests.** `TestRecordDispatchBeginRefusesASlotOffItsPanelKey` in `stats/cmd/flow/record_test.go`, modelled on the existing panel-key refusal tests there: refused (exit 2, stderr naming the key and `skills/flow/review-panel.md`) for `panel-3-test-sanity+call-efficiency` with `-slot test-sanity`, `panel-4-mutation+failure-modes` with `-slot mutation`, `panel-3-extra` with `-slot test-sanity+call-efficiency`; admitted for `panel-3-test-sanity+call-efficiency` with its joined slot, `panel-2-primary` with `primary`, `panel-1-primary+principles-cite` with `primary+principles`, `panel-0-primary-retry` with `primary`, `panel-9-primary-agents` with `primary`; a `-role merger` `panel-3-merger` row is never checked. `TestSetDispatchSlot` in `stats/internal/store/records_test.go`: sets a reviewer row's slot, `ErrDispatchNotFound` on an unknown id and on a non-reviewer row. Run both: they fail.
  - [ ] **Step 3: Implement** (decision `sr1-slot-matches-key`) `PanelKeyNamesSlot(key, slot string) bool` in `stats/internal/records/render.go` beside `PanelDispatchRound`, and the refusal in `runRecordDispatchBegin` (`stats/cmd/flow/record.go`) right after the empty-slot refusal.
  - [ ] **Step 4: Implement** (decision `sr1-reslot-dispatch`) `Store.SetDispatchSlot` in `stats/internal/store/records.go` and `flowd reslot-dispatch <dispatch-id> <slot>` in `stats/cmd/flowd/main.go`, dispatched beside `reoutcome-dispatch`, opening the store from `config.FromEnv`, reading the row's key and refusing a slot `records.PanelKeyNamesSlot` rejects before the update.
  - [ ] **Step 5: Verify.** `cd stats && go test ./cmd/flow/ -run 'TestRecordDispatchBegin' -count=1` and `cd stats && go test ./internal/store/ -run 'TestSetDispatchSlot' -count=1` green, `cd stats && go build ./cmd/flowd/` clean; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean.
  - [ ] **Step 6: Commit** the files with the subject below.
  - [ ] **Step 7: Repair the five rows** of change `kan-1061-flow-add-flow-interactive-fast-conversational` once the commit exists, from this worktree's `stats/` with the dev store's default DSN:

```bash verified:ids, keys and slots read from the dev store with read-only psql in flow-postgres at 2026-10-30; reslot-dispatch is this task's
go run ./cmd/flowd reslot-dispatch 8830 test-sanity+call-efficiency
go run ./cmd/flowd reslot-dispatch 8831 mutation+failure-modes
go run ./cmd/flowd reslot-dispatch 8834 mutation+failure-modes
go run ./cmd/flowd reslot-dispatch 8847 test-sanity+call-efficiency
go run ./cmd/flowd reslot-dispatch 8848 mutation+failure-modes
```

    Then `flow record dispatches -change kan-1061-flow-add-flow-interactive-fast-conversational -C /Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational` shows the joined slot on every `panel-<r>-test-sanity+call-efficiency` and `panel-<r>-mutation+failure-modes` row; report the read-back.

**Build:** green
**Files:** `stats/internal/records/render.go`, `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flowd/main.go`
**Allowed-collateral:** `stats/**/*_test.go`
**Tests:** `TestRecordDispatchBeginRefusesASlotOffItsPanelKey`, `TestSetDispatchSlot`
**Regression:** `TestRecordDispatchBeginRefusesASlotOffItsPanelKey` — reverting admits `panel-3-test-sanity+call-efficiency` under `-slot test-sanity`; `TestSetDispatchSlot` — reverting removes the method.
**Baseline:** before=277 after=279
<!-- measured: cat stats/cmd/flow/record_test.go stats/internal/store/records_test.go | grep -c '^func Test' @ 1b6e694b -->
**Commit:** fix(stats): refuse a panel row whose slot is not its key's and repair mis-slotted rows
**After:** none

**Decision:** sr1-slot-matches-key, sr1-reslot-dispatch

- [ ] 2. Carry the capture shape into the planner's capture-task paragraph
  - [ ] **Step 1: Edit** (decision `sr2-capture-shape-at-plan`) `skills/flow/brainstorm-planner.md`: at the end of the paragraph opening **Write a capture task that baselines against a mockup to measure before it commits.**, after `(kan-1046 self-review, 8c).`, append, inside the same `>` quote, this sentence:

```markdown verified:authored for this change; the step numbers read from skills/flow/visual-verify-verifier.md at branch spectre/kan-1061-flow-add-flow-interactive-fast-conversational (step 8 full page or viewport, step 9 check-identical-baselines)
> Every screenshot such a task's spec takes is the full page or viewport, never an element or a
> clipped region, and each captures a state no sibling capture shows — steps 8 and 9 of
> `skills/flow/visual-verify-verifier.md` refuse a clip and a byte-identical pair (kan-1061
> self-review).
```

    Rewrap the paragraph's lines near 100 columns; change no other text.
  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational lint` exits 0.
  - [ ] **Step 3: Commit** with the subject below.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none** — prose; the lint list's guards cover it.
**Regression:** none — a reverted sentence fails no test.
**Commit:** docs(flow): state the capture shape in the planner's capture-task paragraph
**After:** none

**Decision:** sr2-capture-shape-at-plan

- [ ] 3. Print the launch lines on a live dispatch begin
  - [ ] **Step 1: Write the failing test** `TestRecordDispatchBeginPrintsTheLaunchLines` in `stats/cmd/flow/record_test.go`, on a fake daemon answering the write and a scratch change whose `tasks.md` holds three column-0 tasks: `task-2-implementer` prints `recorded: dispatch <seq>`, `description: Task 2/3`, `Role: implementer` in that order; `task-1+3-reviewer` prints `description: Tasks 1+3/3` and `Role: gated-reviewer`; `task-1-reviewer-fix-1` prints `Role: gated-re-reviewer`; `panel-2-primary` (reviewer) prints `description: panel-2-primary` and no `Role:` line; `panel-2-merger` (merger) `Role: panel-merger`; `panel-fix-3` `Role: panel-fix`; `visual-verify-2` (verifier) `Role: visual-verifier`; `tooling-analysis-1` (planner) `Role: tooling-analyst`; a back-filled row (`-started-at`) and an `-agent-id inline` row print only the `recorded:` line; an unreadable plan omits the `description:` line of a task key. Run it: fails.
  - [ ] **Step 2: Implement** (decision `sr3-begin-prints-launch-lines`) in `runRecordDispatchBegin` (`stats/cmd/flow/record.go`): once `classifyRecordWrite` returns 0 on a live, non-inline begin, print the lines, the plan's count from `countPlanLines(plan, taskLinePattern)`.
  - [ ] **Step 3: Edit** `skills/flow/implement.md`: at the end of the paragraph opening **Every dispatch's Agent-tool `description` names what it carries**, after `launches recorded no row.`, append:

```markdown verified:authored for this change; the paragraph and the Role line column read from skills/flow/implement.md at branch spectre/kan-1061-flow-add-flow-interactive-fast-conversational
A live `flow record dispatch begin` prints both for its row after its `recorded:` line —
`description: <opening>`, and `Role: <role id>` wherever the row decides it — and the parent
pastes that opening into the Agent call's `description` and that line into its prompt, never
composing either; where no `Role:` line is printed, as for a panel reviewer row, the id is the
**Role line** column's above.
```

  - [ ] **Step 4: Verify.** `cd stats && go test ./cmd/flow/ -run 'TestRecordDispatchBegin' -count=1` and `cd stats && go test ./internal/guard/ -run 'TestPanelRoundBegins|TestCheckLoopDispatch' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational lint` exits 0.
  - [ ] **Step 5: Commit** with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/implement.md`
**Allowed-collateral:** `stats/**/*_test.go`
**Tests:** `TestRecordDispatchBeginPrintsTheLaunchLines`
**Regression:** `TestRecordDispatchBeginPrintsTheLaunchLines` — reverting prints only `recorded: dispatch <seq>`.
**Baseline:** before=162 after=163
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 1b6e694b -->
**Commit:** feat(stats): print a dispatch's description opening and Role line on begin
**After:** Task 1

**Decision:** sr3-begin-prints-launch-lines

- [ ] 4. Record each dispatch-hook denial as an incident
  - [ ] **Step 1: Write the failing tests.** In `stats/internal/repotest/helpers_test.go`, `runHook` appends `FLOW_HOOK_INCIDENT_CMD=true` before the caller's env, so no hook run writes the dev store. `TestEnforceAgentBaselineHookRecordsADenial` in `stats/internal/repotest/enforce_agent_baseline_hook_test.go`, with `FLOW_HOOK_INCIDENT_CMD` an argv-logging stub: a prompt with no baseline pointer is denied and logs one `record incident` call carrying `-guard enforce-agent-baseline`, `-minutes-lost 0` and a `-symptom` opening with the deny reason's first words; an `openRowFixture` deny (no open row for `task-2-`) also carries `-change c1`; an allowed dispatch logs nothing; a stub exiting 1 still yields the deny JSON as the whole stdout and exit 0. `TestEnforceLoopDispatchHookRecordsADenial` in `stats/internal/repotest/enforce_loop_dispatch_hook_test.go`, on its `newLoopStub` checker refusing: one call carrying `-guard enforce-loop-dispatch`; an admitted loop dispatch logs nothing. Run them: fail.
  - [ ] **Step 2: Implement** (decision `sr4-hook-denial-incident`) in `hooks/enforce-agent-baseline.py` and `hooks/enforce-loop-dispatch.py`; update each module docstring's description of a deny by one sentence naming the incident write and its fail-open.
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/repotest/ -run 'TestEnforceAgentBaselineHook|TestEnforceLoopDispatchHook' -count=1` green; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` clean; `scripts/check-python-suppressions.sh` exits 0.
  - [ ] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `hooks/enforce-loop-dispatch.py`, `stats/internal/repotest/helpers_test.go`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`, `stats/internal/repotest/enforce_loop_dispatch_hook_test.go`
**Allowed-collateral:** `stats/**/*_test.go`
**Tests:** `TestEnforceAgentBaselineHookRecordsADenial`, `TestEnforceLoopDispatchHookRecordsADenial`
**Regression:** `TestEnforceAgentBaselineHookRecordsADenial` — reverting logs no `record incident` call on a deny; `TestEnforceLoopDispatchHookRecordsADenial` — the same for the loop hook.
**Baseline:** before=12 after=14
<!-- measured: cat stats/internal/repotest/enforce_agent_baseline_hook_test.go stats/internal/repotest/enforce_loop_dispatch_hook_test.go | grep -c '^func Test' @ 1b6e694b -->
**Commit:** feat(hooks): record each dispatch-hook denial as a flow incident
**After:** none

**Decision:** sr4-hook-denial-incident

- [ ] 5. Write kan-1061's two prompt-tuning cases
  - [ ] **Step 1: Write the case folders** (decision `sr5-cases-task`) `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/` and `-2/`, each with the files of step 2a's table in `skills/flow-self-review/SKILL.md`, from section 8 of `.superpowers/sdd/self-review-pass.md` in this worktree:
    - `case.json`: the pass's **Case 1** and **Case 2** JSON bodies, verbatim.
    - `finding.md`: in the shape of `docs/prompt-tuning/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd-1/finding.md` — Case 1: severity Important, location `stats/web/src/views/RunDetail.tsx` (Stage runs panel), the verifier's sentence from the first report in `spectre/changes/archive/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification.md` naming the 5 rows and the "Show all 31" chip, and the implementer's leftover line from `.superpowers/sdd/implementer-report-3.md` (the `Stage runs panel shows latest 5` line); Case 2: severity Important, location `stats/cmd/flow/request.go:222`, F1's sentence as the pass quotes it.
    - `bad.diff` / `good.diff`: Case 1 — the Stage runs hunks of `git show ffeaf7eb -- stats/web/src/views/RunDetail.tsx` as before and after (the `LATEST_RUNS` slice and "Show all" chip; the `INTERACTIVE_RUNS` hunk and the chip guard); Case 2 — the `ListStageRuns` / `errors.Is(err, client.ErrUnavailable)` hunks of `git show a9539509 -- stats/cmd/flow/request.go stats/internal/client/client.go` as before and after.
    - `task.md`: Case 1 — Task 9 of `spectre/changes/archive/kan-1061-flow-add-flow-interactive-fast-conversational/tasks.md`, verbatim with its fields; Case 2 — Task 4 of that file, the same way.
  - [ ] **Step 2: Render and check.** `flow prompt-tuning render docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1` and `-2`, then `scripts/check-prompt-tuning-cases.sh` exits 0. A red line is fixed by rewriting the field it names; a case no generic rewrite passes is dropped and reported.
  - [ ] **Step 3: Commit** with the subject below.

**Build:** green
**Files:** `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/case.json`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/finding.md`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/bad.diff`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/good.diff`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/task.md`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/index.html`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-2/case.json`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-2/finding.md`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-2/bad.diff`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-2/good.diff`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-2/task.md`, `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-2/index.html`
**Tests:** **none** — case files; `scripts/check-prompt-tuning-cases.sh` checks them.
**Regression:** none — data files.
**Commit:** docs(prompt-tuning): kan-1061-flow-add-flow-interactive-fast-conversational cases
**After:** none

**Decision:** sr5-cases-task
