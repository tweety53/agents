# kan-1028-mutation-failure-modes-extra-dispatch

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a second always-on pass-1 bundle, `mutation+failure-modes`, on every panel and class, its pair its own 50/50 `sonnet`/`opus` roll at `low`; `failure-modes` and `mutation` leave the Decide tree's roster.

**Spec:** `spectre/changes/kan-1028-mutation-failure-modes-extra-dispatch/design.md` — read it before any task; each task's `**Decision:**` names the entry it implements.

## Global constraints

- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`.
- No test contacts the dev workspace's `flowd`.
- The bundle is written `mutation+failure-modes`, in that order, everywhere (decision `mf-bundle-order`); its decision key is `extra_dispatch_mutation`, its roll key `<name>extramf`.
- Every new or reworded sentence in a run-loaded file (`skills/**/*.md`, `rules/*`) is listed in `spectre/changes/kan-1028-mutation-failure-modes-extra-dispatch/verbatim-moves.txt`, exactly as `scripts/check-verbatim-moves.sh`'s FAIL line prints it after `::`, one per line, under a `#` comment naming the task. That file is never staged by a task's commit; a task appends to it and leaves it unstaged.

## Review Focus

- A decision with no `extra_dispatch_mutation` (every decision before this change) still renders and records. Pinned in Tasks 2 and 3.
- `micro` prints both `extra:` lines. Pinned in Task 1.
- No roster the tree prints carries `failure-modes` or `mutation`. Pinned in Task 1.

- [x] 1. `plan-class.sh` prints the second extra line and drops failure-modes/mutation from the tree
  - [x] **Step 1: Write the failing test.** Append `TestPlanClassExtraBundles` to `stats/internal/guard/plan_class_test.go`: for rolls 0, 49, 50, 99 it asserts `pcExtra("mutation+failure-modes", roll)` returns `extra: mutation+failure-modes · roll <r> · sonnet low\n` below 50 and `… opus low\n` from 50, and `pcExtra("test-sanity+call-efficiency", roll)` the existing line. Run `cd stats && go test ./internal/guard/ -run 'TestPlanClassExtraBundles' -count=1` — FAIL (`pcExtra` takes one argument).
  - [x] **Step 2: Implement.** In `stats/internal/guard/planclass.go`: `pcExtra(bundle string, roll uint64)` formats `extra: <bundle> · roll …`; its doc comment names both always-on bundles. After the existing `fmt.Fprint(stdout, pcExtra(…))` call (now passing `"test-sanity+call-efficiency"`), add `fmt.Fprint(stdout, pcExtra("mutation+failure-modes", roll(name+"extramf")))`. In `pcTree`, the full shape no longer appends `; failure-modes; mutation` to `roster` or `+failure-modes · mutation` to `dispatches`; `room` stays `compact >= 90 && class != "small"`; an admitted experimental slot appends `; exp-<name>` to `roster` and `+exp-<name>` to `dispatches`, so it joins the first dispatch. Update `pcTree`'s doc comment to match (decision `exp-slot-first-dispatch`).
  - [x] **Step 3: Update every pin.** `TestPlanClassExtra` calls `pcExtra("test-sanity+call-efficiency", roll)`. Every end-to-end expected stdout in `plan_class_test.go` gains the ninth line its guard run prints — run the guard, copy the `extra: mutation+failure-modes …` line literally, never recompute it. Every `pcTree` expectation drops `; failure-modes; mutation` / `+failure-modes · mutation`; the `tree-192` case reads `roster primary; principles; exp-a` and `dispatches primary+principles+exp-a`, and its label and the comment above it say the slot joins the first dispatch.
  - [x] **Step 4: Header.** In `scripts/plan-class.sh`'s header: `Prints exactly eight lines` → `nine lines`; the `grouping:` usage line reads `dispatches primary+principles[+exp-<name>]`; add `#   extra: mutation+failure-modes · roll N · sonnet|opus low` after the existing `extra:` usage line; the sentence about the slot joining "a full regular/big roster's overflow dispatch" says it joins the first dispatch of a full regular/big panel; append to the extra-line sentence: `# The second extra line is the always-on mutation+failure-modes dispatch's pair, rolled the same way on roll(<name>extramf) (kan-1028).`
  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestPlanClass' -count=1` — all exit 0, `gofmt -l` prints nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/planclass.go`, `stats/internal/guard/plan_class_test.go`, `scripts/plan-class.sh`
**Tests:** `TestPlanClassExtraBundles`
**Regression:** reverting the commit: `pcExtra` takes one argument again and the test fails to compile; the guard prints no `mutation+failure-modes` line.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/plan_class_test.go @ 9e91b85c -->
**Commit:** feat(guard): plan-class rolls the always-on mutation+failure-modes dispatch
**After:** none

**Decision:** mf-own-roll
**Decision:** mf-out-of-roster
**Decision:** exp-slot-first-dispatch

- [x] 2. `flow decision render` prints `extra_dispatch_mutation`
  - [x] **Step 1: Write the failing test.** Append `TestDecisionRenderExtraDispatchMutation` to `stats/cmd/flow/decision_test.go`, shaped like `TestDecisionRenderExtraDispatch`: a micro decision carrying both `extra_dispatch` and `"extra_dispatch_mutation":{"slots":["mutation","failure-modes"],"model":"opus","effort":"low","reason":"extramf roll 70"}` renders `- **Extra dispatch:** test-sanity+call-efficiency …` and then `- **Extra dispatch:** mutation+failure-modes opus/low\n`; the same decision without the key exits 0 and prints no `mutation+failure-modes`. Run `cd stats && go test ./cmd/flow/ -run 'TestDecisionRenderExtraDispatchMutation' -count=1` — FAIL.
  - [x] **Step 2: Implement.** In `stats/cmd/flow/decision.go`, add `ExtraDispatchMutation` to `dDecision` with the same anonymous struct type as `ExtraDispatch` and tag `json:"extra_dispatch_mutation"`; in the render, after the `ExtraDispatch` line, print the same `line("Extra dispatch", …)` for it when non-nil. It stays out of `dRequired`.
  - [x] **Step 3: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestDecisionRender' -count=1` — all exit 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/decision.go`, `stats/cmd/flow/decision_test.go`
**Tests:** `TestDecisionRenderExtraDispatchMutation`
**Regression:** reverting the commit drops the second Extra dispatch line; the test fails on it.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^func Test' stats/cmd/flow/decision_test.go @ 9e91b85c -->
**Commit:** feat(flow): render the decision's mutation+failure-modes extra dispatch
**After:** none

**Decision:** extra-dispatch-mutation-key

- [x] 3. `flow record decision` checks `extra_dispatch_mutation`'s model
  - [x] **Step 1: Write the failing test.** Append `TestRecordDecisionChecksExtraDispatchMutationModel` to `stats/internal/api/records_test.go`, shaped like `TestRecordDecisionChecksExtraDispatchModel` with the body's key `extra_dispatch_mutation` and slots `["mutation","failure-modes"]`: `haiku`, `fable` and `""` are refused with `ErrInvalidRecord` naming `extra_dispatch_mutation.model "<m>"` and never reach the store; `sonnet` and `opus` record. Run `cd stats && go test ./internal/api/ -run 'TestRecordDecisionChecksExtraDispatchMutationModel' -count=1` — the `refuses` subtests FAIL.
  - [x] **Step 2: Implement.** In `checkDecisionModels` (`stats/internal/api/records.go`), after the `extra_dispatch` check add the same call for `"extra_dispatch_mutation"`, and name `extra_dispatch_mutation.model` in the doc comment's list.
  - [x] **Step 3: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/api/ && go test ./internal/api/ -run 'TestRecordDecision' -count=1` — all exit 0. Commit.

**Build:** green
**Files:** `stats/internal/api/records.go`, `stats/internal/api/records_test.go`
**Tests:** `TestRecordDecisionChecksExtraDispatchMutationModel`
**Regression:** reverting the commit lets an off-policy `extra_dispatch_mutation` model record; the `refuses` subtests fail.
**Baseline:** before=55 after=56
<!-- measured: grep -c '^func Test' stats/internal/api/records_test.go @ 9e91b85c -->
**Commit:** feat(api): refuse an off-policy model in extra_dispatch_mutation
**After:** none

**Decision:** extra-dispatch-mutation-key

- [x] 4. The reviewers view counts the mutation+failure-modes bundle's roles
  - [x] **Step 1: Write the failing test.** Append `TestReviewersCountsMutationExtraDispatch` to `stats/internal/store/aggregate_test.go`, shaped like the existing test at the `test-sanity+call-efficiency` dispatch (around line 1049): a dispatch `-slot mutation+failure-modes`, a finding under `failure-modes`, and a decision `{"panel":"default","extra_dispatch_mutation":{"slots":["mutation","failure-modes"],"model":"opus","effort":"low"}}` with no roster naming either role; `Reviewers` returns rows for both, `failure-modes` with its finding counted and each row's description the one in `extraDispatchDescriptions`. Run it — FAIL (no rows: neither role is a roster member).
  - [x] **Step 2: Implement.** In `stats/internal/store/aggregate.go`'s `roster_slots`, add `UNION SELECT jsonb_path_query(decision, '$.extra_dispatch_mutation.slots[*]') #>> '{}' FROM decisions`; add `failure-modes` and `mutation` to `extraDispatchDescriptions`, copied from Task 5's table rows in `skills/flow/review-panel.md` (`Failure-modes — error return, timeout, partial write, concurrent re-entry`, `Mutation — sabotage-proofing`); the doc comments name both extra keys. `TestExtraDispatchDescriptionsMatchReviewPanel` already holds the map to the doc — it passes because both rows stay in review-panel.md in Task 5.
  - [x] **Step 3: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/store/ && go test ./internal/store/ -run 'TestReviewers|TestExtraDispatchDescriptions' -count=1` — all exit 0. Commit.

**Build:** green
**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`
**Tests:** `TestReviewersCountsMutationExtraDispatch`
**Regression:** reverting the commit drops both roles from the view when only the extra dispatch carried them; the test fails.
**Baseline:** before=25 after=26
<!-- measured: grep -c '^func Test' stats/internal/store/aggregate_test.go @ 9e91b85c -->
**Commit:** feat(store): count the mutation+failure-modes extra dispatch's roles
**After:** none

**Decision:** extra-dispatch-mutation-key

- [x] 5. Skill text: two always-on extra dispatches, failure-modes and mutation out of the roster
  - [x] **Step 1: review-panel.md.** Opening paragraph: "Every pass 1 also dispatches **The always-on extra dispatches** below." Rename the section **The always-on extra dispatch** → **The always-on extra dispatches** and update every citation of it repo-wide (`git grep -n 'always-on extra dispatch\*\*'`). The section states two bundles, `test-sanity+call-efficiency` on `extra_dispatch` and `mutation+failure-modes` on `extra_dispatch_mutation`, each on its own pair, a decision with no such key dispatching that bundle on `opus` at `low`; the `failure-modes` and `mutation` rows move from **The roster** table into this section's table unchanged (mutation's throwaway-worktree pointer kept). **The roster** names only `primary` and `principles`, and states that a `default` panel's resolved roster drops a role an always-on extra dispatch carries, recorded with `flow record pass -round 0 -note 'not dispatched — carried by an always-on extra dispatch: <slot>'`. Fix the Bundled dispatch and Grouping paragraphs' overflow text (the mutating role no longer in a roster) and the `-model`/`-effort` exception sentence (both extra pairs). The placeholder table's `[CONTEXT_BUNDLE_PATHS]` row is unchanged.
  - [x] **Step 2: brainstorm-planner.md.** Decide step 4 records both `extra:` lines: `extra_dispatch` from the first, `extra_dispatch_mutation` `{slots ["mutation","failure-modes"], …}` from the second. **Model and effort**'s extra bullet covers both bundles. The decision JSON paragraph lists `extra_dispatch_mutation`. The micro row's "the extra dispatch" → "the extra dispatches".
  - [x] **Step 3: the rest.** `skills/flow/review-panel-late-fix.md` (both reductions run beside both extras; drop "Mutation and Failure-modes are not dispatched"), `skills/flow/verify-fix-loop.md` step 3 (both bundles, `panel-<round>-mutation+failure-modes` beside the existing name), `skills/flow/review-panel-fix-round.md` (the re-run sentence names all four extra roles), `skills/flow-fast/SKILL.md`, `skills/flow/failure-modes-reviewer-prompt.md` and `skills/flow/review-panel-optional-slots.md` where they call either role a roster or optional slot, `skills/flow/test-sanity-reviewer-prompt.md` and `skills/flow/call-efficiency-reviewer-prompt.md` (the section-title citation), and `README.md`'s class table (`regular` no longer "adds `failure-modes`+`mutation`").
  - [x] **Step 4: Absence sweep.** `git grep -n -F -e 'failure-modes; mutation' -e '+failure-modes · mutation' -e 'adds `failure-modes`+`mutation`' -e 'Mutation and Failure-modes are not dispatched' -- skills README.md scripts` prints nothing.
  - [x] **Step 5: Verify.** Run `scripts/check-verbatim-moves.sh`, list every FAIL sentence in `verbatim-moves.txt` under `# Task 5`, re-run until exit 0; then every `scripts/check-*.sh` guard `.flow/project.md`'s `## lint` names, and `cd stats && go test ./internal/store/ -run TestExtraDispatchDescriptions -count=1` — all exit 0. Commit (not `verbatim-moves.txt`).

**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/review-panel-late-fix.md`, `skills/flow/verify-fix-loop.md`, `skills/flow/review-panel-fix-round.md`, `skills/flow-fast/SKILL.md`, `skills/flow/review-panel-optional-slots.md`, `skills/flow/test-sanity-reviewer-prompt.md`, `skills/flow/call-efficiency-reviewer-prompt.md`, `README.md`
**Allowed-collateral:** `skills/**/*.md`
**Tests:** none — repairs existing tests: none (test-sanity-reviewer-prompt.md is a reviewer prompt, not a test)
**Regression:** none — text; the absence sweep and `TestExtraDispatchDescriptionsMatchReviewPanel` guard it.
**Commit:** docs(flow): run mutation+failure-modes as an always-on extra dispatch
**After:** Task 1

**Decision:** mf-own-roll
**Decision:** mf-out-of-roster

Correction (2026-10-23): the plan named `skills/flow/failure-modes-reviewer-prompt.md` in `**Files:**`; it never calls its role a roster or optional slot, so the commit leaves it untouched and the field drops it.

- [x] 6. The learning note is a file beside the repository, its path printed at implementation start
  - [x] **Step 1: learning.md.** **When a note runs** gains the file: `<project>-learning/<name>.md`, `<project>` the main checkout's absolute path, created with `mkdir -p`, never inside a git repository. **Recording it** records when the file is written, not when the final message prints. **The printed section** becomes the file's content, and the run prints `Learning: <absolute path>` as soon as the file is written; the handoff carries `**Learning:** <absolute path>`.
  - [x] **Step 2: callers.** `skills/flow/implement.md` (compose, write and print the path at `flow.load-context`), `skills/flow/verify-and-handoff.md` (the handoff's `**Learning:**` line replaces the section), `skills/flow-contracts/handoff-blocks.md` (names `Learning` as a line), `skills/flow-fast/SKILL.md` (write and print the path at load-context; the final report repeats it).
  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh` (FAIL sentences listed under `# Task 6`), then every `scripts/check-*.sh` guard `.flow/project.md`'s `## lint` names — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow-contracts/learning.md`, `skills/flow/implement.md`, `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/handoff-blocks.md`, `skills/flow-fast/SKILL.md`
**Tests:** none — repairs existing tests: none (skill text)
**Regression:** none — text; read by the run.
**Commit:** docs(flow): write the learning note to a file beside the repository
**After:** Task 5

**Decision:** learning-note-file
