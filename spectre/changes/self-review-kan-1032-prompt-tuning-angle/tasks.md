# self-review-kan-1032-prompt-tuning-angle

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1032's self-review findings F-A1 (Task 1), F-A2 (Task 2), F-A3 (Task 3),
F-A4 (Task 4), F-A5 (Task 5), F-A6 (Task 6) and F-A7 (Task 7), as `proposal.md` names them.

**Architecture:** `run-verify-list` stops on a refused begin; one brief sentence; the baseline
hook asks for a loop's `Key:` line and lets planning agents through; `flow record dispatch
begin` refuses a loop key off its loop change; migration 0047 drops the stored twins;
`integrate-change commit` requires the in-run fix line; a case's round 0 passes; the commit
hook judges no message its raw-word fallback split.

**Tech Stack:** Go (`stats/`), Python (`hooks/`), SQL, Bash, Markdown.

**Spec:** `design.md` beside this file.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer
  and no `Co-Authored-By:` trailer. Never write `spectre/changes/` — this change's `patches/`
  and `verbatim-moves.txt` are committed with the plan; never stage
  `spectre/changes/archive/kan-1032-prompt-tuning-angle/narrative.md`, whose uncommitted edit is
  the coordinator's. Never merge or push.
- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1032-prompt-tuning-angle`
  (`<agents-wt>` below), branch `spectre/kan-1032-prompt-tuning-angle`. Every `**Files:**` path is
  relative to it. Never touch a main checkout, and never create a worktree.
- Tasks 1 to 4 apply `<agents-wt>/spectre/changes/self-review-kan-1032-prompt-tuning-angle/patches/task-<n>.diff`
  with `git apply <that path>` from `<agents-wt>` for their prose; `git apply --check` passed for
  all four at `21133166`. A patch that does not apply is reported, never hand-merged by guesswork.
  Everything else is written from the steps.
- No new inline suppressions; `gofmt -l .` in `stats/` prints nothing. Never stop flowd or
  flow-postgres, never drop the `flow` database, never write to it — a store test runs on its own
  scratch database.

## Review Focus

- A refused begin stops the list; an unreachable store and a failed end still never change the
  exit code (`verify-list-stops-on-refused-begin`).
- The hook never again asks for a loop row under the run's change, and planning agents pass
  (`loop-key-needs-its-key-line`, `planning-agents-pass-row-checks`).
- Migration 0047 deletes only an agentless row with a same-key twin under its own loop change
  (`migration-drops-loop-twins`).
- `commit` skips a row the self-review report names and a store it cannot read
  (`commit-requires-in-run-fix-line`).

- [ ] 1. Stop the verify list when the store refuses its begin
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In
    `stats/internal/guard/run_verify_list_test.go`, the recording test's fake `flow` logs and exits
    0; read `record` in `stats/internal/guard/runverifylist.go` and confirm `record("begin")`'s
    result is never read before the list's commands run.
  - [ ] **Step 2: Write the failing test** `TestRunVerifyListStopsOnARefusedBegin` in
    `stats/internal/guard/run_verify_list_test.go`, on the recording test's fixture: a fake `flow`
    that prints `flow: record dispatch refused: dirty` and exits 2 on `record dispatch begin` →
    `run-verify-list <wt> test <five recording args>` exits 4, the list's command (`echo ran`
    appended to the log) never ran, and stderr carries
    `run-verify-list: flow record dispatch begin refused — no command ran: flow: record dispatch refused: dirty`;
    a fake `flow` that exits 0 on `begin` and 1 on `end` → the list exits 0. Run
    `cd stats && go test ./internal/guard -run 'TestRunVerifyListStopsOnARefusedBegin' -count=1` — it fails.
  - [ ] **Step 3: Implement** in `runverifylist.go`: `record` returns whether the call succeeded;
    when `begin` fails, print the line above (`strings.TrimSpace` of its output) and return 4 before
    any command. `end`'s failure keeps today's warning. Rewrite the RECORDING paragraph of the file
    header to say so, and add `#   4  the store refused the recording begin (stderr names why): no command ran`
    to the Exit codes block of `scripts/run-verify-list.sh`.
  - [ ] **Step 4: Apply `patches/task-1.diff`** from `<agents-wt>`: the **Recording** sentence
    (`skills/flow/verify-and-handoff.md`).
  - [ ] **Step 5: Prove the test pins the fix.** Save a copy of `runverifylist.go`, make the
    refused `begin` fall through again — the Step 2 test fails; restore the copy.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l internal/guard` prints nothing;
    `go test ./internal/guard -run 'TestRunVerifyList' -count=1` passes; from `<agents-wt>`,
    `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh` and
    `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `scripts/run-verify-list.sh`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestRunVerifyListStopsOnARefusedBegin`
**Regression:** `TestRunVerifyListStopsOnARefusedBegin` fails — a refused begin runs the whole list again and leaves its row unrecorded.
**Baseline:** before=2 after=3 <!-- measured: grep -c -E '^func Test' stats/internal/guard/run_verify_list_test.go @ 21133166 → 2 -->
**Commit:** fix(guard): stop the verify list when the store refuses its begin
**After:** none

**Decision:** verify-list-stops-on-refused-begin

- [ ] 2. State how a mutation reproducer selects its tests
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'every test of the package, file or module' skills/flow/review-panel.md` prints `0`.
  - [ ] **Step 2: Apply `patches/task-2.diff`** from `<agents-wt>`: the sentence closing the
    reproducer paragraph of **The mutation-testing brief** (`skills/flow/review-panel.md`).
  - [ ] **Step 3: Verify.** The Step 1 grep prints `1`; from `<agents-wt>`,
    `scripts/check-verbatim-moves.sh`, `scripts/check-mutation-reproducer-pin.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-markdown-integrity.sh` and
    `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none — one brief sentence; no test reads the brief's wording.
**Regression:** none — no test; reverting drops the rule.
**Commit:** docs(flow): run a mutation reproducer over the whole package
**After:** none

**Decision:** mutation-reproducer-runs-whole-package

- [ ] 3. Record loop dispatches once, under the loop's change
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Pipe an `openRowPayload`
    (`stats/internal/repotest/enforce_agent_baseline_hook_test.go`) with description
    `pipeline-fix-7-plan x` and no `Key:` line, under `openRowFixture` rows `[]`, into
    `hooks/enforce-agent-baseline.py` (or read `begin_row_reason`): it denies with
    `no open ledger row for 'pipeline-fix-7-plan' in c1`, asking for `flow record dispatch begin -change c1 -key pipeline-fix-7-plan`.
  - [ ] **Step 2: Write the failing tests.**
    `TestEnforceAgentBaselineHookAsksForTheLoopKeyLine` in the hook test file: description
    `pipeline-fix-7-plan x` with no `Key:` line → denied, the reason carries `Key: pipeline-fix-7-plan`
    and `enforce-loop-dispatch.py` and never `flow record dispatch begin`, and the fake `flow` was
    never called; the same with `self-review-x-plan y`; the same description with the prompt line
    `Key: pipeline-fix-7-plan` → allowed.
    `TestRecordDispatchBeginRefusesALoopKeyOffItsLoop` in `stats/cmd/flow/record_test.go`, on the
    shape of its other begin-validation cases: `-change kan-1 -key pipeline-fix-7-plan` exits 2,
    stderr naming `enforce-loop-dispatch.py`; `-change pipeline-fix-kan-1-7 -key pipeline-fix-7-plan`
    is accepted; `-change kan-1 -key self-review-kan-1-plan` is accepted.
    `TestMigration0047DropsLoopDispatchTwins` in `stats/internal/store/legacyshapes_test.go`, on a
    `newTestDatabase`: seed changes `kan-1` and `pipeline-fix-kan-1-7` in one project and
    `kan-2` with no loop change; under `kan-1` an agentless `pipeline-fix-7-plan` (with a
    `findings` row pointing at it), an agentless `pipeline-fix-8-plan` (no twin) and a stamped
    `pipeline-fix-7-fix`; under `pipeline-fix-kan-1-7` stamped `pipeline-fix-7-plan` and
    `pipeline-fix-7-fix`; under `kan-2` an agentless `pipeline-fix-7-plan`. Execute the
    migration file's SQL (`os.ReadFile` of `migrations/0047_drop_loop_dispatch_twins.sql`) twice:
    only `kan-1`'s agentless `pipeline-fix-7-plan` is gone, its finding's `dispatch_id` is NULL,
    every other row stays. Run
    `cd stats && go test ./internal/repotest -run 'TestEnforceAgentBaselineHookAsksForTheLoopKeyLine' -count=1`,
    `go test ./cmd/flow -run 'TestRecordDispatchBeginRefusesALoopKeyOffItsLoop' -count=1` and
    `go test ./internal/store -run 'TestMigration0047DropsLoopDispatchTwins' -count=1` — each fails.
  - [ ] **Step 3: Implement.** In `hooks/enforce-agent-baseline.py`, a `loop_key_line_reason`
    checked after `description_reason` and before `begin_row_reason`: in a flow session, a
    description whose first word opens `pipeline-fix-` or `self-review-` and a prompt with no
    `LOOP_KEY_LINE` is denied with "Dispatch blocked: `<word>` is a loop key — its prompt carries
    `Key: <word>` as a line of its own, and hooks/enforce-loop-dispatch.py records its row under
    the loop's change off that line. Add the line and re-send; never `flow record dispatch begin`
    it under the run's change."; `begin_row_reason` returns `None` for such a word. Update the
    module docstring's begun-row paragraph. In `stats/cmd/flow/record.go`'s
    `runRecordDispatchBegin`, next to the other key checks: a `-key` opening `pipeline-fix-` whose
    change does not open `pipeline-fix-` prints `flow: -key %q is a pipeline-fix loop key:
    hooks/enforce-loop-dispatch.py records it under the loop's change, never under %q
    (skills/flow-contracts/pipeline.md)` and exits 2. Write
    `stats/internal/store/migrations/0047_drop_loop_dispatch_twins.sql` with a header comment
    citing kan-1032's self-review: in one statement chain, the twins are the dispatches `d` with
    `d.agent_id IS NULL`, `d.dispatch_key LIKE 'pipeline-fix-%'`, `d`'s change name `NOT LIKE
    'pipeline-fix-%'`, and a dispatch `t` with the same `dispatch_key` under a change of the same
    `project_key` named `LIKE 'pipeline-fix-' || <d's change name> || '-%'`; set
    `findings.dispatch_id` to NULL where it names a twin, then delete the twins.
  - [ ] **Step 4: Apply `patches/task-3.diff`** from `<agents-wt>`: the sentence under the closed
    list (`skills/flow/implement.md`) and item 1's sentence in **Pipeline defects found mid-run**
    (`skills/flow-contracts/pipeline.md`).
  - [ ] **Step 5: Prove the tests pin the fix.** One at a time, on a saved copy: drop the new hook
    reason (the hook test fails), drop the `pipeline-fix-` begin check (the record test fails),
    drop the twin `EXISTS` clause from the migration (the migration test fails); restore each.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l . ` prints nothing; `go vet ./cmd/flow ./internal/store`
    is clean; `go test ./internal/repotest -run 'TestEnforceAgentBaselineHook' -count=1`,
    `go test ./cmd/flow -run 'TestRecordDispatch' -count=1` and
    `go test ./internal/store -run 'TestMigration' -count=1` pass; from `<agents-wt>`,
    `scripts/check-python-suppressions.sh`, `scripts/check-verbatim-moves.sh`,
    `scripts/check-markdown-integrity.sh`, `scripts/check-dispatch-paragraphs.sh` and
    `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`, `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/store/migrations/0047_drop_loop_dispatch_twins.sql`, `stats/internal/store/legacyshapes_test.go`, `skills/flow/implement.md`, `skills/flow-contracts/pipeline.md`
**Tests:** `TestEnforceAgentBaselineHookAsksForTheLoopKeyLine`, `TestRecordDispatchBeginRefusesALoopKeyOffItsLoop`, `TestMigration0047DropsLoopDispatchTwins`
**Regression:** `TestEnforceAgentBaselineHookAsksForTheLoopKeyLine` fails — a loop launch without its Key line is told to begin a row under the run's change again; `TestRecordDispatchBeginRefusesALoopKeyOffItsLoop` fails — the parent's hand-recorded twin is accepted; `TestMigration0047DropsLoopDispatchTwins` fails — the stored twins stay, unmeasured.
**Baseline:** before=137 after=140 <!-- measured: cat stats/internal/repotest/enforce_agent_baseline_hook_test.go stats/cmd/flow/record_test.go stats/internal/store/legacyshapes_test.go | grep -c -E '^func Test' @ 21133166 → 137 -->
**Commit:** fix(flow): record loop dispatches once, under the loop's change
**After:** none

**Decision:** loop-key-needs-its-key-line
**Decision:** begin-refuses-loop-key-off-loop-change
**Decision:** migration-drops-loop-twins

- [ ] 4. Require the in-run fix line for every fixed in-run row
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read `commit` in
    `stats/internal/guard/integratechange_commit.go`: only lines matching `icFixLine` are checked,
    and no store row is read.
  - [ ] **Step 2: Write the failing test** `TestIntegrateChangeRequiresInRunFixLine` in
    `stats/internal/guard/integrate_change_prepare_test.go`, through `icRunCases`. First make the
    fake `flow` in `stats/internal/guard/integrate_change_test.go` answer `self-review findings`
    with `cat "$WORK/sr-findings.json" 2>/dev/null || echo '[]'`. Cases: a findings file with
    `[{"angle":"flow-fix","disposition":"fixed","ref":"1a2b3c4d5e6f"}]` and a narrative with no
    `In-run pipeline fix:` line → `commit` exits 1 and its last line opens `STOP: narrative` and
    names `1a2b3c4`; the same with the narrative line
    the narrative line in the format `narrative_format_in_format` uses → `commit` proceeds as
    `narrative_format_in_format` does; a `flow-cost` `fixed` row and a `flow-fix` `noted` row
    with no line → `commit` proceeds. Run
    `cd stats && go test ./internal/guard -run 'TestIntegrateChangeRequiresInRunFixLine' -count=1` — it fails.
  - [ ] **Step 3: Implement** in `integratechange_commit.go`, after the format loop: run
    `flow self-review findings -C <main> -change <name>` (as `recordFixedRows` does), parse its JSON
    array; for each row with `angle` `flow-fix` and `disposition` `fixed` whose `ref` no
    `icFixFormat` sha group prefixes (compare the shorter against the longer's prefix) and whose
    `ref` the self-review report (`c.report()`, when it exists) does not contain, collect it, then
    stop `narrative` naming each missing sha and the line's format. A failing or unparseable read
    prints `integrate-change: warning: flow self-review findings unreadable — the in-run fix line
    check is skipped` on stderr and continues. Extend the `commit` paragraph of the
    `integratechange.go` header with the new stop.
  - [ ] **Step 4: Apply `patches/task-4.diff`** from `<agents-wt>`: the second
    **`STOP: narrative`** line (`skills/flow/integrate.md`).
  - [ ] **Step 5: Prove the test pins the fix.** On a saved copy, drop the new check — the first
    Step 2 case fails; restore the copy.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l internal/guard` prints nothing;
    `go test ./internal/guard -run 'TestIntegrateChange' -count=1` passes; from `<agents-wt>`,
    `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh` and
    `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/integratechange_commit.go`, `stats/internal/guard/integratechange.go`, `stats/internal/guard/integrate_change_prepare_test.go`, `stats/internal/guard/integrate_change_test.go`, `skills/flow/integrate.md`
**Tests:** `TestIntegrateChangeRequiresInRunFixLine`
**Regression:** `TestIntegrateChangeRequiresInRunFixLine` fails — a landed in-run fix narrated without its line passes `commit` again.
**Baseline:** before=1 after=2 <!-- measured: grep -c -E '^func Test' stats/internal/guard/integrate_change_prepare_test.go @ 21133166 → 1 -->
**Commit:** fix(guard): require the in-run fix line for every fixed in-run row
**After:** none

**Decision:** commit-requires-in-run-fix-line

- [ ] 5. Accept round 0 in a prompt-tuning case
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In
    `stats/internal/guard/check_prompt_tuning_cases_test.go`, the row "an unset source round is
    caught" sets `Round = 0` and expects exit 1 — round 0, the initial panel's round, is refused.
  - [ ] **Step 2: Write the failing test** `TestCheckPromptTuningCasesSourceRound` in that file, on
    the fixture the table test uses: `round` 0 → exit 0; `round` absent from `case.json` → exit 1,
    output `source.round is not set`; `round` -1 → exit 1, output `source.round`. Delete the old
    row. Run `cd stats && go test ./internal/guard -run 'TestCheckPromptTuningCasesSourceRound' -count=1` — it fails.
  - [ ] **Step 3: Implement.** `Source.Round` in `stats/internal/prompttuning/prompttuning.go`
    becomes `*int` (`json:"round"`); `prompttuningcases.go` reports `source.round is not set` for
    `nil` and `source.round %d is negative` below 0. Update every literal (`Round: new(2)`) in
    `check_prompt_tuning_cases_test.go` and `stats/internal/prompttuning/prompttuning_test.go`, and
    any reader `go build ./...` names.
  - [ ] **Step 4: Prove the test pins the fix.** On a saved copy, restore `< 1` semantics — the
    round-0 case fails; restore the copy.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . ` prints nothing; `go vet ./internal/... ./cmd/...`
    is clean; `go test ./internal/guard ./internal/prompttuning ./cmd/flow -run 'PromptTuning' -count=1`
    passes; from `<agents-wt>`, `scripts/check-prompt-tuning-cases.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/prompttuning/prompttuning.go`, `stats/internal/prompttuning/prompttuning_test.go`, `stats/internal/guard/prompttuningcases.go`, `stats/internal/guard/check_prompt_tuning_cases_test.go`
**Tests:** `TestCheckPromptTuningCasesSourceRound`
**Regression:** `TestCheckPromptTuningCasesSourceRound` fails — a case from the initial panel round is refused again, or an absent round passes.
**Baseline:** before=1 after=2 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_prompt_tuning_cases_test.go @ 21133166 → 1 -->
**Commit:** fix(guard): accept round 0 in a prompt-tuning case
**After:** none

**Decision:** round-zero-is-a-round

- [ ] 6. Let planning agents through the hook's row checks
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** An `openRowPayload` with
    `subagent_type` `Explore` and description `self-review plan — trace` is denied with
    `DESCRIPTION_REASON` (`hooks/enforce-agent-baseline.py`).
  - [ ] **Step 2: Write the failing test** `TestEnforceAgentBaselineHookAdmitsPlanningAgents` in
    `stats/internal/repotest/enforce_agent_baseline_hook_test.go` (give `openRowPayload` a
    `subagent_type` parameter, or build the payload beside it): `Explore` and `Plan` with
    description `self-review plan — trace`, rows `[]` → allowed; `flow-low` with the same
    description → denied. Run
    `cd stats && go test ./internal/repotest -run 'TestEnforceAgentBaselineHookAdmitsPlanningAgents' -count=1` — it fails.
  - [ ] **Step 3: Implement.** A `PLANNING_AGENT_TYPES = {"Explore", "Plan"}` set beside
    `EXEMPT_SUBAGENT_TYPES`, with a comment citing kan-1032's self-review; `description_reason`,
    `loop_key_line_reason` and `begin_row_reason` return `None` for those types. The
    baseline-pointer, pinned-role and flow-high checks still apply. Add one sentence to the module
    docstring.
  - [ ] **Step 4: Prove the test pins the fix.** On a saved copy, drop the exemption — the test
    fails; restore the copy.
  - [ ] **Step 5: Verify.** `cd stats && go test ./internal/repotest -run 'TestEnforceAgentBaselineHook' -count=1`
    passes; from `<agents-wt>`, `scripts/check-python-suppressions.sh` exits 0. Commit.

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`
**Tests:** `TestEnforceAgentBaselineHookAdmitsPlanningAgents`
**Regression:** `TestEnforceAgentBaselineHookAdmitsPlanningAgents` fails — a planner's Explore or Plan dispatch is denied in a flow session again.
**Baseline:** before=5 after=6 <!-- measured: grep -c -E '^func Test' stats/internal/repotest/enforce_agent_baseline_hook_test.go @ 21133166 → 4, plus Task 3's one -->
**Commit:** fix(hooks): let planning agents through the row checks
**After:** Task 3

**Decision:** planning-agents-pass-row-checks

- [ ] 7. Read no message the hook's raw-word fallback split
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** A Bash payload whose command
    is a quoted heredoc whose body holds one apostrophe (`the coordinator's note`) followed by
    `git -C {SWT} commit -q -m "chore(spectre): plan c1"` is denied by `hooks/protect-main-checkout.py`
    with `no \`chore\` subject`; the same command without the apostrophe is allowed — `tokenize`'s
    unbalanced-quote fallback keeps the quote on `"chore(spectre):`, so `CHORE_SUBJECT_RE` misses.
  - [ ] **Step 2: Write the failing test**
    `TestProtectMainCheckoutSkipsAMessageItCannotRead` in
    `stats/internal/repotest/protect_main_checkout_test.go`, on the fixture of
    `TestProtectMainCheckoutDeniesAnUntraileredChangeBranchCommit`: the Step 1 command on
    `spectre/c1` → allowed. Run
    `cd stats && go test ./internal/repotest -run 'TestProtectMainCheckoutSkipsAMessageItCannotRead' -count=1` — it fails.
  - [ ] **Step 3: Implement** in `commit_lacks_trailer`: a message value opening with `"` or `'`
    is a raw word the fallback split, so the commit is not judged (`return False`), as the
    docstring already says of a message the command text does not state; add that case to the
    docstring.
  - [ ] **Step 4: Prove the test pins the fix.** On a saved copy, drop the new return — the test
    fails; restore the copy.
  - [ ] **Step 5: Verify.** `cd stats && go test ./internal/repotest -run 'TestProtectMainCheckout' -count=1`
    passes; from `<agents-wt>`, `scripts/check-python-suppressions.sh` exits 0. Commit.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go`
**Tests:** `TestProtectMainCheckoutSkipsAMessageItCannotRead`
**Regression:** `TestProtectMainCheckoutSkipsAMessageItCannotRead` fails — a chore commit after an apostrophe in a heredoc is denied again.
**Baseline:** before=3 after=4 <!-- measured: grep -c -E '^func Test' stats/internal/repotest/protect_main_checkout_test.go @ 21133166 → 3 -->
**Commit:** fix(hooks): judge no commit message the raw-word fallback split
**After:** none

**Decision:** unread-message-is-not-judged
