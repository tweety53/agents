# self-review-kan-1047-review-dynamics-heatmap-by-project

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-28-self-review-kan-1047-review-dynamics-heatmap-by-project-design.md` in this worktree.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1047-review-dynamics-heatmap-by-project`, on branch `spectre/kan-1047-review-dynamics-heatmap-by-project`; every `**Files:**` path is relative to it. Never touch the main checkout.
- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. No new lint suppressions.
- Each Go test added must fail with its task's production change reverted — show that once before committing.
- Doc edits add the sentence named and cut nothing else, save the exact replacements a step names.

## Review focus

- A reproducer run on a clean tree must leave it clean whatever the reproducer did, and a dirty path's uncommitted bytes must still come back (Task 1).
- The brief sentences must render verbatim through `render-slot-prompt.sh` (Task 2).
- F22 must not refuse a plan with no prototype citation, no `ui paths` task, or a `regression checkout` (Task 4).
- The visual `verify` must never run concurrently with the rest of a `test` list, never on a `lint` list, and never when `decision.json` records `visual.verify` `skipped` (Task 5).

- [ ] 1. Restore every tracked path a reproducer run changed
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Add `TestRunReproducerRestoresACleanPathTheRunChanged` to `stats/internal/guard/runreproducer_test.go`, built like `TestRunReproducerRestoresAnUncommittedEdit` (a committed `f.txt` holding `x`, a clean tree), with two cases: (a) a reproducer `printf 'mutant\n' > f.txt; exit 0` — expect exit 5, stderr carrying `run-reproducer: restored f.txt`, and `f.txt` back to `x`; (b) a reproducer `printf 'mutant\n' > f.txt; sleep 30` run under a short bound (`RUN_REPRODUCER_BOUND_SECONDS`, as the existing bound tests set it) — expect exit 3 and `f.txt` back to `x`. Run `cd stats && go test ./internal/guard/ -run 'TestRunReproducerRestoresACleanPathTheRunChanged' -count=1`: both cases fail on `f.txt` still `mutant` — the defect (kan-1047's live mutant in `stats/web/src/components/DynamicsChart.tsx`).
  - [ ] **Step 2: Fix.** In `stats/internal/guard/runreproducer.go`, extend `rrRestoreDirty` (decision `reproducer-restores-clean-paths`): after its loop over the snapshot, list the tracked paths differing from HEAD now (`git -C <worktree> diff --no-renames --name-only -z HEAD --relative`, as `rrSnapshotDirty` does), and for each path not in the snapshot run `git -C <worktree> checkout -q HEAD -- <path>`; append each one that succeeded to `restored` before the sort, so it is named on stderr like the rest. Update `rrRestoreDirty`'s doc comment and the comment above `rrSnapshotDirty`'s call to say a path clean before the run is put back from HEAD, and the `restored <path>` paragraph of `scripts/run-reproducer.sh`'s header to say the same. Leave every verdict branch unchanged.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && make -s web-dist && go vet ./internal/guard/` exits 0; `cd stats && go test ./internal/guard/ -run 'TestRunReproducer' -count=1` passes. Revert Step 2's Go hunk once and see the new test fail, then restore it. Commit.

**Build:** green
**Files:** `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`, `scripts/run-reproducer.sh`
**Tests:** `TestRunReproducerRestoresACleanPathTheRunChanged`
**Regression:** reverting the new restore loop leaves the reproducer's edit of the clean `f.txt` in the tree, failing both cases.
**Baseline:** before=9 after=10
<!-- measured: cat stats/internal/guard/runreproducer_test.go | grep -c '^func Test' @ branch spectre/kan-1047-review-dynamics-heatmap-by-project (before; after adds this task's one test) -->
**Commit:** fix(guard): restore every tracked path a reproducer run changed
**After:** none

**Decision:** reproducer-restores-clean-paths

- [ ] 2. Reproducers never cd and never nest shell quoting
  - [ ] **Step 1: Edit.** In `skills/flow/review-panel.md`, **The mutation-testing brief**, insert after the sentence ending "the reverse of the generic exit-code contract." this sentence: "A reproducer never `cd`s — it runs each tool against its directory with `go -C`, `make -C` or `npm --prefix` and names every file as `$PWD/<path>` — and makes its edit through a patch file or a fixed-string replacement, never through shell quoting nested inside another quoted string (kan-1047: an EXIT trap restoring a relative path after `cd stats` left a live mutant, and nested single quotes in a `sed` never applied the edit)." (decision `reproducer-no-cd-no-nested-quotes`).
  - [ ] **Step 2: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-mutation-reproducer-pin.sh`, `scripts/check-normative-inventory.sh` and `scripts/check-installed-citations.sh` each exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — prose only; the doc guards in Step 2 cover it.
**Regression:** reverting the commit drops the no-`cd`, no-nested-quoting rule from every rendered mutation brief.
**Commit:** docs(flow): forbid cd and nested shell quoting in a mutation reproducer
**After:** none

**Decision:** reproducer-no-cd-no-nested-quotes

- [ ] 3. The prototype file draws the page alone
  - [ ] **Step 1: Edit.** In `skills/flow/brainstorm-planner.md`, **The checklist**, the bullet opening "**A visible change is prototyped before its design is approved**", insert after the sentence ending "iterate it on the operator's feedback each round until they approve; that approval is the convergence confirm's **approve the design and move on** (**Convergence**, below)." this sentence: "The file draws the proposed page alone — no caption, note or screenshot beside it — since `flow.visual-verify` renders the whole file as the capture's frame; a note to the operator goes in the message (kan-1047: a caption over a JPEG of the live page never paired with a capture)." (decision `prototype-draws-page-alone`).
  - [ ] **Step 2: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-normative-inventory.sh` and `scripts/check-installed-citations.sh` each exit 0. Commit.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none** — prose only; the doc guards in Step 2 cover it.
**Regression:** reverting the commit drops the rule, and a prototype may again carry a caption the verifier renders as part of the frame.
**Commit:** docs(flow): draw the prototype page alone
**After:** none

**Decision:** prototype-draws-page-alone

- [ ] 4. Refuse a prototype-specified UI plan with no capture task
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Add `TestCheckPlanShapePrototypeNeedsACapture` to `stats/internal/guard/check_plan_shape_test.go`: a temp git repository whose `.flow/project.md` declares `## visual verification` with `ui paths` `web/src/**` and `screenshots` `web/tests/visual` (the two-table shape `.flow/project.md` of this repository uses, `verify`/`capture` rows included), and `spectre/changes/x/` holding a `design.md` that cites `.superpowers/sdd/x-prototype.html` and a well-shaped `tasks.md`. Cases: (a) one task with `**Files:**` `web/src/App.tsx` — refused, exit 1, a line carrying `F22`; (b) as (a) plus a second task with `**Files:**` `web/tests/visual/x.spec.ts` — passes; (c) as (a) with the design citing no `.html` — passes; (d) as (a) with a `regression checkout` row — passes. Build the plan text the way the file's existing tests do. Run `cd stats && go test ./internal/guard/ -run 'TestCheckPlanShapePrototypeNeedsACapture' -count=1`: case (a) fails — the defect (kan-1047's committed plan touched `stats/web/src/**` in three tasks and named no capture).
  - [ ] **Step 2: Fix.** In `stats/internal/guard/planshape.go` add F22 (decision `plan-shape-prototype-capture`): a `prototypeCapture()` method appended to `check()`'s list, guarded by `p.on("F22")`. It reads the sibling `design.md` (absent or unreadable: no finding — F19 owns that), matches it against ``\.superpowers/sdd/[^\s`)]+\.html``, resolves the root with `git -C <dir of tasks.md> rev-parse --show-toplevel` (failure: no finding), reads that root's `.flow/project.md` `## visual verification` section with the existing section helpers (`vvSectionLines`, `vtSplitCells`, `vtFoldCell` — read `ui paths` via `visualTrigger` and the `screenshots` and `regression checkout` cells as `specreach.go` reads its cell), and returns nothing when the section, `screenshots` or a `ui paths` match is missing or a `regression checkout` is declared. Otherwise, over every task's parsed `**Files:**` (as F21 reads `fields.files`): when some path makes `visualTrigger(p.env, root, …, io.Discard, io.Discard)` exit 0 and no path has the prefix `<screenshots>/`, report at that first task's title line: "F22 design.md cites a prototype and this task touches `ui paths` (`<path>`), but no task's **Files:** names a capture under `<screenshots>/` — plan the capture spec driving every interactive state the prototype draws (skills/flow/brainstorm-planner.md)". Add the F22 row to the header comment's findings list and change "Twenty-one findings, F1-F21" and "F7-F21's are the rows below" to F22.
  - [ ] **Step 3: Cite the guard.** In `skills/flow/brainstorm-planner.md`, the same prototype bullet, after the sentence ending "drives each interactive state it shows — hover, focus, pressed — so `flow.visual-verify` renders and compares that state too (kan-1038)." insert: "`check-plan-shape.sh` F22 refuses a plan whose design cites the prototype and whose tasks touch a `ui paths` glob while no task's `**Files:**` names a file under `screenshots` (kan-1047)."
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && make -s web-dist && go vet ./internal/guard/` exits 0; `cd stats && go test ./internal/guard/ -run 'TestCheckPlanShape' -count=1` passes; `scripts/check-plan-shape.sh` with no argument exits 0 over every live plan; `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-normative-inventory.sh` exit 0. Revert Step 2's Go hunk once and see case (a) fail, then restore it. Commit.

**Build:** green
**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`, `skills/flow/brainstorm-planner.md`
**Tests:** `TestCheckPlanShapePrototypeNeedsACapture`
**Regression:** reverting F22 lets case (a)'s plan pass with no capture task, failing the test.
**Baseline:** before=4 after=5
<!-- measured: cat stats/internal/guard/check_plan_shape_test.go | grep -c '^func Test' @ branch spectre/kan-1047-review-dynamics-heatmap-by-project (before; after adds this task's one test) -->
**Commit:** fix(guard): refuse a prototype-specified UI plan with no capture task
**After:** Task 3

**Decision:** plan-shape-prototype-capture

- [ ] 5. End a test verify list with the visual verify
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Add `TestRunVerifyListRunsTheVisualVerify` to `stats/internal/guard/run_verify_list_test.go`, built on the file's existing fixture: a git worktree whose `.flow/project.md` declares a `## test` list (`echo LIST-RAN`) and a `## visual verification` section with `ui paths` `web/src/**`, `screenshots` `web/tests/visual` and a `verify` row `echo VISUAL-RAN <specs>`, and a bare clone set as its `origin` so the base branch resolves. Cases: (a) a commit on the branch touching `web/src/a.tsx` — the `test` run prints a `VISUAL-RAN` block after the `LIST-RAN` block, exit 0; (b) as (a) with the verify row `exit 1` — exit 1; (c) as (a) with `.superpowers/sdd/decision.json` holding `{"visual":{"verify":"skipped","reason":"r"}}` — no `VISUAL-RAN`; (d) the commit touching only `docs/a.md` — no `VISUAL-RAN`; (e) as (a) with no `origin` — `VISUAL-RAN` (the base does not resolve, so it runs); (f) as (a) on the `lint` list — no `VISUAL-RAN`. Run `cd stats && go test ./internal/guard/ -run 'TestRunVerifyListRunsTheVisualVerify' -count=1`: (a), (b) and (e) fail — the defect.
  - [ ] **Step 2: Fix.** In `stats/internal/guard/runverifylist.go` (decision `verify-list-runs-visual`), on a `test` list only, after the concurrent results are reported and before the green-sha write: read the `## visual verification` section of `<worktree>/.flow/project.md` (the section helpers `visualsection.go` and `specreach.go` use); with a `verify` row declared, `<worktree>/.superpowers/sdd/decision.json` not recording `visual.verify` `skipped`, and `unchanged("")` false, decide the trigger — `Registry["resolve-base-branch"]` in-process for the base, `resolveRemoteBase` on it, `git merge-base HEAD <ref>`, `git diff --name-only <base>..HEAD` into `visualTrigger(env, wt, …, io.Discard, io.Discard)`; any failure before the trigger answers counts as triggered. When triggered, run the row with `<specs>` replaced by the empty string through the existing `run` closure (foreground, after every other command), its working directory the `regression checkout` cell when declared and `<worktree>` otherwise, key `visual`. It counts in `failed` like any command. Extend the header comment (THE LIST paragraph and OUTPUT) to state this, and mirror it in `scripts/run-verify-list.sh`'s header.
  - [ ] **Step 3: Update the parent's prose.** In `skills/flow/implement.md`, the paragraph opening "When the change's diff touches a `## visual verification` `ui paths` glob", replace the text from that opening through "so a capture a shared-component change moved fails here rather than at `flow.verify`." with: "`run-verify-list.sh`'s `test` list ends with that section's `verify` run whole — `<specs>` empty, every capture — when its own worktree's diff from its base branch touches a `ui paths` glob and `decision.json` does not record `visual.verify` `skipped`, its header canonical for the rest, so a capture a shared-component change moved fails here rather than at `flow.verify`. When only the `--repo` run step 2 of **Visual verification** (`skills/flow/verify-and-handoff.md`) adds exits 0, for another worktree of the resolved set, its merge base from this run's working notes, the parent runs that `verify` itself after the list, against the canonical worktree's stack, under the stack rules of **One stack for the verification tasks** below." Then change "Before `verify`, `check-dev-stack-fresh.sh <canonical-worktree>` exit 1 restarts the stack" to "Before the list, on a change whose diff touches a `ui paths` glob, `check-dev-stack-fresh.sh <canonical-worktree>` exit 1 restarts the stack", and "`verify` runs from the resolved regression checkout (the project root when none is declared), after the section's `setup` row when the toolchain is missing." to "The section's `setup` row runs before the list when the toolchain is missing." Leave the rest of the paragraph unchanged.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && make -s web-dist && go vet ./internal/guard/` exits 0; `cd stats && go test ./internal/guard/ -run 'TestRunVerifyList' -count=1` passes; `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-normative-inventory.sh` and `scripts/check-installed-citations.sh` exit 0. Revert Step 2's Go hunk once and see (a) fail, then restore it. Commit.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `scripts/run-verify-list.sh`, `skills/flow/implement.md`
**Tests:** `TestRunVerifyListRunsTheVisualVerify`
**Regression:** reverting the visual step leaves a `test` list ending without the `verify` run, failing cases (a), (b) and (e).
**Baseline:** before=9 after=10
<!-- measured: cat stats/internal/guard/run_verify_list_test.go | grep -c '^func Test' @ branch spectre/kan-1047-review-dynamics-heatmap-by-project (before; after adds this task's one test) -->
**Commit:** fix(guard): end a test verify list with the visual verify
**After:** none

**Decision:** verify-list-runs-visual
