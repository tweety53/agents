# self-review-kan-1015-templates-step-1-fe-templates-room-search-gym

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close every non-`big` finding of the self-review pass of
`kan-1015-templates-step-1-fe-templates-room-search-gym` — fifteen pipeline defects, the kan-466
product defect the sweep found, and the archived narrative's wrong cause — and repair what the
defects already wrote wrong.

**Architecture:** prose rules in `skills/flow/`; Go guards in `stats/internal/guard/`, the records
renderer and one store column; two Python hooks with Go harnesses in `stats/internal/repotest/`;
store repairs through the `flow` CLI; commits in the change's own gymie worktrees for the archived
plan, the narrative and the product fix.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-24-self-review-kan-1015-templates-step-1-fe-templates-room-search-gym-design.md`
in this worktree.

## Global constraints

- Agents tasks: work in `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1015-templates-step-1-fe-templates-room-search-gym`, branch `self-review-kan-1015-templates-step-1-fe-templates-room-search-gym`; every agents `**Files:**` path is relative to its root. One commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` there — this change's `verbatim-moves.txt` is committed with the plan. Never merge or push.
- Gymie tasks and repairs work in the change's own worktrees on `spectre/kan-1015-templates-step-1-fe-templates-room-search-gym`: backend `/Users/tweety53/Projects/gymie-worktrees/kan-1015-templates-step-1-fe-templates-room-search-gym` (`<backend wt>`), frontend `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-1015-templates-step-1-fe-templates-room-search-gym` (`<frontend wt>`), playwright `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1015-templates-step-1-fe-templates-room-search-gym` (`<playwright wt>`). Every gymie commit carries `-m "Flow-Key: self-review-kan-1015-templates-step-1-fe-templates-room-search-gym-fix"`; never push. Never touch a main checkout, and never create a worktree outside the merge-base worktrees Task 16's trace names.
- `<name>` is `kan-1015-templates-step-1-fe-templates-room-search-gym`. Store writes for `<name>` run with `-C <backend wt>`.
- Run every guard from this worktree's own `scripts/`, never the main checkout's: `flow-guard` builds from the checkout the shim sits in.
- Run `gofmt -w` on every Go file you change; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container. Gradle and Playwright runs take `timeout 600000`.
- A doc task's prose lands character for character as its fenced replacement states: `scripts/check-verbatim-moves.sh` matches every sentence against `verbatim-moves.txt` exactly. A replacement whose old text is not found once is reported, never hand-merged.

## Review Focus

- Task 10: a dispatch the store cannot answer for — flowd down, no `/flow` stage-begin for the session, the `flow` CLI absent — passes; only an answered store with no open matching row denies, and a loop-key dispatch is never checked.
- Task 11: an implementer's `git commit -m "<subject>" -m "Task-Id: 3"` and a `chore(spectre): …` commit on `spectre/<name>` pass; a commit on any other branch is never judged; `-F`/`--amend` are never judged.
- Task 6: an unticked append on a plan the guards find green exits 0; a tick of an existing task alone runs neither guard.
- Task 9: a verifier row begun with no `-round-base` reads `roundBase` empty, and older rows keep passing the guard.
- Task 13: a report line that says "not pre-existing" or names `merge-base capture failed — <reason>` beside a this-change attribution is not refused.

- [ ] 1. Hold the PIXEL PROBE to rendered pixels
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'rendered pixels or geometry' skills/flow/review-panel.md skills/flow/review-panel-fix-round.md stats/internal/guard/dispatchparagraphs.go` prints one line in each file.
    <!-- measured: git grep -n 'rendered pixels or geometry' @ 0f456593 → review-panel.md:1028, review-panel-fix-round.md:378, dispatchparagraphs.go:53 -->
  - [ ] **Step 2: Apply the prose.** In `skills/flow/review-panel.md` replace the first block with the second:

````text verified:applied by str.replace in this worktree @ 0f456593, old text found once
> against the actual rendered pixels or geometry, in the style the reviewers' reproducers
````

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed this sentence, listed in verbatim-moves.txt
> against the actual rendered pixels, in the style the reviewers' reproducers
````

    In `skills/flow/review-panel-fix-round.md` replace `actual rendered pixels or geometry, and a fix commit for this class of bug that landed no such` with `actual rendered pixels, and a fix commit for this class of bug that landed no such`.
  - [ ] **Step 3: Move the guard's phrase.** In `stats/internal/guard/dispatchparagraphs.go` the `pixel` entry's required phrase `"actual rendered pixels or geometry"` becomes `"actual rendered pixels, in the style"`; in `scripts/check-dispatch-paragraphs.sh` the header comment's `assertion against the actual rendered pixels or geometry, and rejects a` becomes `assertion against the actual rendered pixels, and rejects a`; in `stats/internal/guard/check_dispatch_paragraphs_test.go` every fixture line `> against the actual rendered pixels or geometry, in the style the reviewers' reproducers` becomes `> against the actual rendered pixels, in the style the reviewers' reproducers`, and case 58's `dpHas` phrase becomes `"actual rendered pixels, in the style"`.
  - [ ] **Step 4: Verify.** `scripts/check-dispatch-paragraphs.sh` exits 0; `cd stats && go test ./internal/guard/ -run 'TestCheckDispatchParagraphs' -count=1` passes; `scripts/check-verbatim-moves.sh` prints no FAIL for these two sentences; `cd stats && gofmt -l internal/guard && go vet ./internal/guard` prints nothing.
  - [ ] **Step 5: Commit** the five files with the subject below.

**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`, `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`, `scripts/check-dispatch-paragraphs.sh`
**Tests:** none — repairs existing tests: case 58 of `TestCheckDispatchParagraphs` and its fixtures move to the new phrase and still fail a block that lacks it.
**Regression:** reverting restores "or geometry", and a fix round may again land a bounds-only probe for a fill only the render shows.
**Commit:** fix(review-panel): hold the PIXEL PROBE to rendered pixels
**After:** none

- [ ] 2. Check the base against its tracked branch before task 1
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'check-base-behind' skills/flow/implement.md; echo $?` prints only `1`.
    <!-- measured: grep -n check-base-behind skills/flow/implement.md @ 0f456593 → exit 1 -->
  - [ ] **Step 2: Apply the prose.** In `skills/flow/implement.md` §1, insert this paragraph, followed by one blank line, immediately before the line `**The plan is refreshed against the base before task 1 runs.** Before the`:

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed its sentences, listed in verbatim-moves.txt
**Every recorded base is compared with the branch it tracks before task 1 runs.** Run
`check-base-behind.sh <worktree>...` over the run's resolved worktree set; its header is canonical
for its lines. Exit 0 → continue. Exit 1 → relay every `BASE-BEHIND:` block, run its two command
lines in order as **Verify** (`skills/flow/verify-and-handoff.md`) runs them, and run the check
again. Exit 2 → name its stderr in the handoff and continue (kan-1015: feature/templates sat 21
commits behind develop through every task, and the base sync at verify exposed a Flyway clash only
after implement).
````
    <!-- measured: the verify report's base check, 21 commits, read from the archived narrative.md of <name> @ backend fe872cce -->
  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` print no FAIL for this paragraph.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** **none** — prose.
**Regression:** reverting leaves a stale base undetected until verify.
**Commit:** fix(implement): check the base against its tracked branch before task 1
**After:** none

- [ ] 3. Run a visual-fix task's capture spec before it closes
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'runs its capture spec before it closes' skills/flow/verify-fix-loop.md; echo $?` prints only `1`.
  - [ ] **Step 2: Apply the prose.** In `skills/flow/verify-fix-loop.md` step 2, insert these lines immediately before the line `   **Every recapture the round runs rebuilds the full app suite's zip after it** —`:

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed its sentences, listed in verbatim-moves.txt
   **A task fixing a visual departure runs its capture spec before it closes.** Its verify step
   runs the project's `capture` command with the round's capture spec for `<spec>`, against the
   stack rebuilt from its fix, and the task closes through `close-task.sh` only on that run's exit
   0 — a red capture is fixed in the same task and folded into its one commit (kan-1015: Tasks 21–24
   closed on desktop geometry tests, and each departure surfaced only at the next verifier).
````
  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` print no FAIL for these lines.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — prose.
**Regression:** reverting lets an appended visual-fix task close on unit tests alone.
**Commit:** fix(verify-fix-loop): run a visual-fix task's capture spec before it closes
**After:** none

- [ ] 4. Measure fix rounds on the 2× pair, never a composite
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'never on a `compose-mockup-frames.sh`' skills/flow/verify-fix-loop.md; echo $?` prints only `1`.
  - [ ] **Step 2: Apply the prose.** In `skills/flow/verify-fix-loop.md` replace the first block with the second:

````text verified:applied by str.replace in this worktree @ 0f456593, old text found once
   a new 1–3px offset the previous fix had judged by eye, 45–70 minutes a round). A leftover of any size
````

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed the new sentence, listed in verbatim-moves.txt
   a new 1–3px offset the previous fix had judged by eye, 45–70 minutes a round). It measures each
   pair on the 2× capture and the 2× mockup frame themselves, never on a `compose-mockup-frames.sh`
   composite (kan-1015 self-review). A leftover of any size
````
  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh` and `scripts/check-markdown-integrity.sh` print no FAIL for this sentence.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — prose.
**Regression:** reverting leaves the implementer free to measure a composite.
**Commit:** fix(verify-fix-loop): measure fix rounds on the 2x pair, never a composite
**After:** Task 3

- [ ] 5. Order F11 pairs through the After graph
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Write `TestCheckPlanShapeF11Reachability` in `stats/internal/guard/check_plan_shape_test.go`: a plan of three tasks, task 1 `**Files:** \`a.go\``, task 2 `**Files:** \`b.go\`` `**After:** Task 1`, task 3 `**Files:** \`a.go\`` `**After:** Task 2`, each with the fields `check-plan-shape` reads; `checkPlanShape` exits 0 and prints no F11 line. A second case drops task 3's `**After:**` to `none` and expects the F11 line naming task 3 and task 1. Run `cd stats && go test ./internal/guard/ -run 'TestCheckPlanShapeF11Reachability' -count=1`: the first case fails.
  - [ ] **Step 2: Fix.** In `fileOwnership` (`stats/internal/guard/planshape.go`) replace the direct test `!slices.Contains(declared[earlier], later) && !slices.Contains(declared[later], earlier)` with `!psReaches(declared, later, earlier) && !psReaches(declared, earlier, later)`, where `psReaches(declared map[string][]string, from, to string) bool` walks `declared` depth-first from `from` with a visited set and answers whether `to` is reached. Update the F11 line of the file's header comment: "no GATING declared After edge between a co-owning pair, in either direction" becomes "no path of GATING declared After edges between a co-owning pair, in either direction".
  - [ ] **Step 3: Verify.** The test passes; `cd stats && go test ./internal/guard/ -run 'TestCheckPlanShape' -count=1` passes; `gofmt -l internal/guard` and `go vet ./internal/guard` print nothing.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`
**Tests:** `TestCheckPlanShapeF11Reachability`
**Regression:** `TestCheckPlanShapeF11Reachability` — reverting flags task 3 against task 1 though task 2 serialises them.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/check_plan_shape_test.go -->
**Commit:** fix(guard): order F11 pairs through the After graph
**After:** none

**Decision:** f11-reachability

- [ ] 6. Run the plan guards on a planning commit that appends a task, and repair the archived change record
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run this worktree's `scripts/check-plan-shape.sh` and `scripts/check-task-build-green.sh` on `<backend wt>/spectre/changes/archive/<name>/tasks.md`: both exit 1 — with Task 5 landed, plan-shape's After hits are task 15 against task 9 alone, beside its `**Tests:**` hits on tasks 21–23; build-green names tasks 21–24 `pending`.
    <!-- measured: both guards on the archived tasks.md @ backend fe872cce → exit 1, 42 plan-shape lines, 4 build-green lines (before Task 5) -->
  - [ ] **Step 2: Write the failing test.** `TestCheckPlanningCommitAppendRunsPlanGuards` in `stats/internal/guard/check_planning_commit_location_test.go`, built like `TestCheckPlanningCommitTickedAppend`: HEAD's `spectre/changes/<name>/tasks.md` holds one green task; the working tree appends an unticked task 2 whose `**Build:**` reads `pending`. `checkPlanningCommit` exits 1 and its output carries build-green's `task 2 has a **Build:** line reading "pending"` line. A second case appends a well-formed task 2 and expects exit 0. Run it: the first case fails.
  - [ ] **Step 3: Fix.** In `checkPlanningCommit` (`stats/internal/guard/planningcommitlocation.go`), collect every task number the diff adds (`m[1] == "+"`, ticked or not) that `had` lacks; when one or more exist for a file, run `checkPlanShape([]string{<wt>/<path>}, env, stdout, stderr)` and `checkTaskBuildGreen([]string{<wt>/<path>}, env, stdout, stderr)` on the working tree's file; an exit 2 from either returns 2, an exit 1 sets `rc = 1`. Add to `scripts/check-planning-commit-location.sh`'s header, after the ticked-append paragraph: "A tasks.md the working tree appends a task to — any task number HEAD's file did not carry — is also run through check-plan-shape.sh and check-task-build-green.sh, their lines printed as they print them; either one's exit 1 is this guard's exit 1 and its exit 2 this guard's exit 2 (kan-1015: tasks 15–24 were appended with `**Build:** pending` and no declared tests)."
  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckPlanningCommit' -count=1` passes; `gofmt -l internal/guard` and `go vet ./internal/guard` print nothing. Commit the three agents files with the subject below.
  - [ ] **Step 5: Repair the archived plan.** In `<backend wt>/spectre/changes/archive/<name>/tasks.md`: task 15's `**After:**` becomes `Task 3, 6, 7, 9, 11, 14`; tasks 21–24's `**Build:**` becomes `green` (each closed with its commit green — confirm from `implementer-report-18`–`21` in `<backend wt>/.superpowers/sdd/`); tasks 17, 18, 19, 21, 22, 23 and 24's `**Tests:**` name, backticked, the test functions their commits added — read from each report and from `git -C <frontend wt> show <its commit>` — keeping the prose after them. Then run this worktree's `scripts/check-plan-shape.sh` and `scripts/check-task-build-green.sh` on that file: both exit 0. Do not commit yet.
  - [ ] **Step 6: Correct the narrative's cause.** Confirm in the Material3 sources the frontend builds against (the `org.jetbrains.compose.material3` sources jar under `~/.gradle/caches/modules-2/files-2.1/`) that `LocalMinimumInteractiveComponentSize` is declared in commonMain with a 48dp default, and read Task 24's test in `<frontend wt>` before its fix commit to confirm it measured the button's layout bounds; a source that says otherwise stops this step with that quote. Then in `<backend wt>/spectre/changes/archive/<name>/narrative.md` replace `and Material3's 48dp minimumInteractiveComponentSize is enforced on web but not in desktop tests, so desktop geometry tests passed on a defect only the web render showed.` with `and the desktop geometry test asserted the Use button's layout bounds, never its drawn fill, so it passed on a defect the render showed — Material3's \`LocalMinimumInteractiveComponentSize\` defaults to 48dp in commonMain, desktop tests included (corrected by the self-review pass).`
  - [ ] **Step 7: Commit both repairs** in `<backend wt>` as one commit, `chore(spectre): repair the kan-1015 plan's appended tasks and narrative`, with the Flow-Key trailer.

**Build:** green
**Files:** `stats/internal/guard/planningcommitlocation.go`, `stats/internal/guard/check_planning_commit_location_test.go`, `scripts/check-planning-commit-location.sh`, `spectre/changes/archive/kan-1015-templates-step-1-fe-templates-room-search-gym/tasks.md`, `spectre/changes/archive/kan-1015-templates-step-1-fe-templates-room-search-gym/narrative.md`
**Tests:** `TestCheckPlanningCommitAppendRunsPlanGuards`
**Regression:** `TestCheckPlanningCommitAppendRunsPlanGuards` — reverting lets an append with `**Build:** pending` pass the planning commit.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/check_planning_commit_location_test.go -->
**Commit:** fix(guard): run the plan guards on a planning commit that appends a task
**After:** Task 5

**Decision:** append-runs-plan-guards

- [ ] 7. Refuse a panel report the store under-counts at integrate, and record the missing rows
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `scripts/check-panel-findings-closed.sh <backend wt> <name>` exits 1 naming `panel-report-0-failure-modes.md (report 3, recorded 0)`, `panel-report-0-mutation.md (report 1, recorded 0)`, `panel-report-0-primary.md (report 2, recorded 0)`, `panel-report-0-principles.md (report 6, recorded 0)` and `panel-report-2-mutation.md (report 1, recorded 0)`; `scripts/check-unfinished-work.sh <backend wt> <name>` names none of them.
    <!-- measured: check-panel-findings-closed.sh <backend wt> <name> @ backend fe872cce → those five reports -->
  - [ ] **Step 2: Write the failing test.** `TestCheckUnfinishedWorkReportCount` in `stats/internal/guard/check_unfinished_work_test.go`, on the file's existing fake-store setup: the store answers `[]` findings and the worktree's `.superpowers/sdd/panel-report-0-primary.md` reads `Critical: 0` / `Important: 2`; the verdict is `OUTSTANDING:` naming `panel-report-0-primary.md (report 2, recorded 0)`. A second case with two matching Important rows of round 0, slot `primary`, reads no such reason. Run it: the first case fails.
  - [ ] **Step 3: Fix.** Move the report-count loop of `checkPanelFindingsClosed` (`stats/internal/guard/panelfindingsclosed.go`, from `reports, _ := filepath.Glob(` through the `uncounted` append) into `func cfcUncounted(worktree string, findings []cfcFindingRow) ([]string, error)`, returning the same `<file> (report <n>, recorded <m>)` strings and the read error; `checkPanelFindingsClosed` calls it and keeps its message and exit 2 on the error. In `checkUnfinishedWork` (`stats/internal/guard/unfinishedwork.go`), right after the open-findings reason and only when `anchor == worktree`, parse `findingsJSON` with `cfcParse`, call `cfcUncounted(anchor, rows)`, and on a non-empty answer append the reason `<n> panel report(s) count findings the store does not hold (<list>) — record each with \`flow record finding -change <name> -C <anchor>\` before landing`; a parse failure or a read error returns 2 as the other store reads do. Add the signal to `scripts/check-unfinished-work.sh`'s header list of signals in the same words.
  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckUnfinishedWork|TestCheckPanelFindingsClosed' -count=1` passes; `gofmt -l internal/guard` and `go vet ./internal/guard` print nothing. Commit the four agents files with the subject below.
  - [ ] **Step 5: Repair the store.** For each finding the five reports list under Critical or Important — 13 rows — run `flow record finding -change <name> -C <backend wt> -ref <the next ref flow record next-ref prints> -round <round> -slot <the report's slot> -severity <critical|important> -location <the report's location> -status <status> -note '<the reviewer's sentence>'`, its `-status` the finding's real close: `fixed` when a commit fixed it (the fix-report or the inline commit names it), `withdrawn <reason>` when the report or a later round withdrew it. Then run `scripts/check-panel-findings-closed.sh <backend wt> <name>` and resolve every line it prints per **Recording findings** (`skills/flow/review-panel.md`) — never by recording a dispatch row that did not run; a refusal no honest row can answer is reported with the guard's line. Re-render the panel record: `flow record render -change <name> -kind panel -repo <backend wt>`.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/panelfindingsclosed.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`
**Tests:** `TestCheckUnfinishedWorkReportCount`
**Regression:** `TestCheckUnfinishedWorkReportCount` — reverting clears a change whose panel reports count findings the store never received.
**Baseline:** before=7 after=8
<!-- measured: grep -c '^func Test' stats/internal/guard/check_unfinished_work_test.go -->
**Commit:** fix(guard): refuse a panel report the store under-counts at integrate
**After:** none

**Decision:** share-report-count

- [ ] 8. Derive the pass log's verdict from finding rows
  - [ ] **Step 1: Write the failing test.** `TestRenderPanelPassLogVerdictFromFindings` in `stats/internal/records/render_test.go`: a `Run` with pass rows for rounds 0 and 1, findings one Critical and two Important in round 0 and none in round 1; the round 0 section carries `- verdict: fix (1 Critical, 2 Important)` after its notes, and round 1 `- verdict: clean`. Run `cd stats && go test ./internal/records/ -run 'TestRenderPanelPassLogVerdictFromFindings' -count=1`: it fails.
  - [ ] **Step 2: Fix.** In `RenderPanel` (`stats/internal/records/render.go`), after the round's pass notes (and after the `no pass row recorded` line), count `r.Findings` whose `Round` is the round and whose severity is `critical` or `important`, case-insensitive, any status, and write `- verdict: clean` when both counts are 0, else `- verdict: fix (<c> Critical, <i> Important)`. Note it in the pass-log comment block above the loop: the notes stay verbatim; the verdict is derived (kan-1015 round 1's note claimed four Importants the store never held).
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/records/ -run 'TestRenderPanel' -count=1` passes; `gofmt -l internal/records` and `go vet ./internal/records` print nothing.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/internal/records/render.go`, `stats/internal/records/render_test.go`
**Tests:** `TestRenderPanelPassLogVerdictFromFindings`
**Regression:** `TestRenderPanelPassLogVerdictFromFindings` — reverting leaves the pass log's verdict to the note's free text.
**Baseline:** before=36 after=37
<!-- measured: grep -c '^func Test' stats/internal/records/render_test.go -->
**Commit:** fix(records): derive the pass log's verdict from finding rows
**After:** none

**Decision:** derived-verdict

- [ ] 9. Record a verifier's round base and refuse a scoped closing run
  - [ ] **Step 1: Write the failing tests.** `TestDispatchRoundBaseRoundTrips` in `stats/internal/store/records_test.go`: a dispatch begun with a round base `c75529d2` reads it back as `RoundBase`, and one begun without reads `""`. `TestCheckVisualVerifyDispatchedRefusesAScopedNewestVerifier` in `stats/internal/guard/check_visual_verify_dispatched_test.go`, on the file's existing fixtures: two completed `visual-verify` rows for the worktree, the newer carrying `"roundBase":"c75529d2"` → exit 1 and a `VISUAL-VERIFY-MISSING:` line naming the key and `c75529d2`; the newer carrying none → exit 0. Run both with `-run`: they fail (the first does not compile).
  - [ ] **Step 2: Fix.** Mirror the `diff_base` path end to end (`grep -rn 'diff_base\|DiffBase\|diff-base' stats/cmd stats/internal --include=*.go` lists every site): migration `stats/internal/store/migrations/0046_dispatch_round_base.sql` adding a nullable `round_base text` to `dispatches`; the store's begin insert and its dispatch reads; the `RoundBase string \`json:"roundBase,omitempty"\`` field on `records.Dispatch` (`stats/internal/records/types.go`); the daemon's begin route; and `flow record dispatch begin`'s `-round-base sha` flag and usage line in `stats/cmd/flow/record.go`. In `stats/internal/guard/visualverifydispatched.go`, among the qualifying verifier rows the guard already finds for the worktree, the newest by `startedAt` must carry an empty `roundBase`; otherwise the verdict is `VISUAL-VERIFY-MISSING: the newest verifier row '<key>' ran scoped to round base <sha> — the loop's closing unscoped run is not recorded (skills/flow/verify-fix-loop.md, **One full run closes the loop**)`. Add that verdict to `scripts/check-visual-verify-dispatched.sh`'s header.
  - [ ] **Step 3: Apply the prose.** In `skills/flow/verify-fix-loop.md` insert these lines immediately after the line `clean scoped re-run, never another round's fix (KAN-1000: three rounds in a row re-ran unscoped).`:

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed this sentence, listed in verbatim-moves.txt
A scoped re-run's verifier row is begun with `-round-base <sha>`, the round base its prompt
names, and `check-visual-verify-dispatched.sh` refuses a worktree whose newest completed verifier
row carries one — the closing unscoped run is not recorded.
````
  - [ ] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/store/ -run 'TestDispatchRoundBaseRoundTrips' -count=1` and `go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatched' -count=1` pass; `go build ./cmd/...`, `gofmt -l .` and `go vet ./internal/... ./cmd/...` are clean; `scripts/check-verbatim-moves.sh` prints no FAIL for the sentence.
  - [ ] **Step 5: Commit** every file with the subject below.

**Build:** green
**Files:** `stats/internal/store/migrations/0046_dispatch_round_base.sql`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/internal/records/types.go`, `stats/cmd/flow/record.go`, `stats/cmd/flowd/main.go`, `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`, `skills/flow/verify-fix-loop.md`
**Allowed-collateral:** `stats/internal/store/*.go`, `stats/internal/server/**`
**Tests:** `TestDispatchRoundBaseRoundTrips`, `TestCheckVisualVerifyDispatchedRefusesAScopedNewestVerifier`
**Regression:** `TestCheckVisualVerifyDispatchedRefusesAScopedNewestVerifier` — reverting passes a change whose last verifier ran scoped; `TestDispatchRoundBaseRoundTrips` — reverting drops the column.
**Baseline:** before=106 after=108
<!-- measured: cat stats/internal/store/records_test.go stats/internal/guard/check_visual_verify_dispatched_test.go | grep -c '^func Test' -->
**Commit:** feat(store): record a verifier's round base and refuse a scoped closing run
**After:** Task 3, 4

**Decision:** round-base-column

- [ ] 10. Deny a flow dispatch whose row was never begun, and back-fill the missing rows
  - [ ] **Step 1: Write the failing test.** `TestEnforceAgentBaselineHookRequiresAnOpenRow` in `stats/internal/repotest/enforce_agent_baseline_hook_test.go`: an `httptest` server answering `/api/v1/stage-runs` with one row `{"changeName":"c1"}`, `FLOW_ADDR` pointed at it, and a fake `flow` first on `PATH` printing a dispatches array; a flow-session payload (transcript holding a `/flow` command) with the baseline pointer and `description: "panel-4-primary review"`. Cases: the array holds `panel-4-primary` with an empty `outcome` → allow; the array holds it with `outcome: "completed"` or lacks it → deny, the reason naming `flow record dispatch begin`; `description: "Task 3/9 (x)"` with an open `task-3-implementer` → allow; `description: "Tasks 1+3/9 (x)"` with an open `task-1+3-implementer` → allow; the fake `flow` exiting 1, or the server answering 500 → allow; a prompt carrying `Key: self-review-x-fix` → allow without calling `flow`. Run `cd stats && go test ./internal/repotest/ -run 'TestEnforceAgentBaselineHookRequiresAnOpenRow' -count=1`: the deny cases fail.
  - [ ] **Step 2: Fix.** In `hooks/enforce-agent-baseline.py` add `begin_row_reason(payload, tool_input, prompt)`, called after `description_reason` in `main`'s `or` chain: it returns `None` for a loop-key prompt, outside a flow session, or for a description `ROW_DESCRIPTION` does not match; resolves the change as `flow-active-change.py`'s `last_flow_change` does (`session_id` from the payload, `GET $FLOW_ADDR/api/v1/stage-runs?session_id=…&command=/flow&sort=-started_at&limit=1`, timeout 2 s); runs `flow record dispatches -change <change> -C <payload cwd>` (timeout 5 s); and denies only when that printed a JSON array holding no row with an empty `outcome` whose `key` is the description's first word, or — for `Task <ids>/<n>` / `Tasks <ids>/<n>` — starts with `task-<ids>-`. Every fault, timeout, non-zero exit or unparseable answer returns `None`. The reason: `Dispatch blocked: no open ledger row for '<word>' in <change> — record it first with \`flow record dispatch begin -change <change> -key <key> …\` (skills/flow/implement.md), then re-send this call unchanged.` Add one paragraph to the module docstring stating the check and that it fails open.
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/repotest/ -run 'TestEnforceAgentBaselineHook' -count=1` passes; `scripts/check-python-suppressions.sh` exits 0. Commit the two files with the subject below.
  - [ ] **Step 4: Back-fill the missing rows.** For each report below, find its Agent call in the run's transcript (`/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/c483d41b-ee52-4470-baed-b83e5f43d1f0.jsonl`: the `tool_use` whose prompt or description names the task, and its `tool_result`'s `agentId`), and run `flow record dispatch backfill -change <name> -C <backend wt> -task <n> -role <role> -model <the call's model> -effort <its subagent_type's effort> -agent-id <agentId> -started-at <tool_use timestamp> -ended-at <tool_result timestamp> -key <key> -session-token <the run's token, from the change's other rows> -commit <the report's sha> -outcome <completed|blocked>`. A report whose call already holds a row — `task-16-implementer`, seq 25 — is matched to that row and not back-filled.
    - gymie-playwright `implementer-report-9`, `-11`, `-13`, `-15` — Task 16, keys `task-16-implementer-r9`, `-r11`, `-r13`, `-r15`;
    - gymie-frontend `implementer-report-10`, `-12`, `-14` — Tasks 17, 18, 19, keys `task-17-implementer`, `task-18-implementer`, `task-19-implementer`;
    - gymie `implementer-report-16`, `-17` — Task 20, keys `task-20-implementer` (outcome `blocked`), `task-20-implementer-r17`; `-18` to `-21` — Tasks 21–24, keys `task-21-implementer` … `task-24-implementer`;
    - gymie `fix-report-1`, `fix-report-2` — role `panel-fix`, no `-task`, keys `panel-fix-1`, `panel-fix-2`; inline when the transcript holds no Agent call for it (`-agent-id inline` with `-commit`, no instants, per **A dispatch recorded only after it finished is a back-fill** in `skills/flow/implement.md`).
    Then `flow record dispatches -change <name> -C <backend wt>` lists all fifteen keys, each with an outcome.

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`
**Tests:** `TestEnforceAgentBaselineHookRequiresAnOpenRow`
**Regression:** `TestEnforceAgentBaselineHookRequiresAnOpenRow` — reverting admits a flow dispatch with no begun row.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/repotest/enforce_agent_baseline_hook_test.go -->
**Commit:** fix(hooks): deny a flow dispatch whose row was never begun
**After:** none

**Decision:** begin-row-via-cli

**Decision:** backfill-keys

- [ ] 11. Deny an untrailered commit on a change branch
  - [ ] **Step 1: Write the failing test.** `TestProtectMainCheckoutDeniesAnUntraileredChangeBranchCommit` in `stats/internal/repotest/protect_main_checkout_test.go`, on the file's existing repo fixtures: a linked worktree on `spectre/c1`; `git commit -m "fix(x): y"` → deny, the reason naming `Flow-Key:`; `git commit -m "fix(x): y" -m "Task-Id: 3"`, `git commit -m "chore(spectre): plan c1"`, `git commit -F msg.txt` and `git commit --amend --no-edit` → allow; the same `-m "fix(x): y"` in a linked worktree on `feature/c1` → allow. Run `cd stats && go test ./internal/repotest/ -run 'TestProtectMainCheckoutDeniesAnUntraileredChangeBranchCommit' -count=1`: the deny case fails.
  - [ ] **Step 2: Fix.** In `bash_hits` (`hooks/protect-main-checkout.py`), where a `commit` verb is collected, when `landings is not None` and the target worktree's current branch (`git(target, "symbolic-ref", "--short", "-q", "HEAD")`) starts with `spectre/` and `commit_lacks_trailer(cargs, target)` holds, append the reason `Blocked: this commit lands on <branch> with neither a \`Task-Id:\` nor a \`Flow-Key:\` trailer and no \`chore\` subject (**Commits outside every task**, skills/flow-contracts/git-boundaries.md). Add \`-m "Flow-Key: <key>"\`, or \`-m "Task-Id: <n>"\` on a task's own commit, then run it again.` A git failure reading the branch judges nothing. Add one paragraph to the module docstring beside the commit-then-push paragraph.
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/repotest/ -run 'TestProtectMainCheckout' -count=1` passes; `scripts/check-python-suppressions.sh` exits 0.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go`
**Tests:** `TestProtectMainCheckoutDeniesAnUntraileredChangeBranchCommit`
**Regression:** `TestProtectMainCheckoutDeniesAnUntraileredChangeBranchCommit` — reverting admits an untrailered non-chore commit on `spectre/<name>`.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/repotest/protect_main_checkout_test.go -->
**Commit:** fix(hooks): deny an untrailered commit on a change branch
**After:** none

**Decision:** spectre-branch-commit-trailer

- [ ] 12. Name the verifier's -rest keys on the closed list
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n -- '`-rest`' skills/flow/implement.md; echo $?` prints only `1`, while `skills/flow/visual-verify.md` defines `-rest` and `-rest-<n>`.
  - [ ] **Step 2: Apply the prose.** In `skills/flow/implement.md` replace `` `visual-verify` (`-2`, `-retry`, `-fix-<k>`, `-fix-<k>-full`) `` with `` `visual-verify` (`-2`, `-retry`, `-rest`, `-rest-<n>`, `-fix-<k>`, `-fix-<k>-full`) ``.
  - [ ] **Step 3: Pin the keys.** `TestEnforceAgentBaselineHookAdmitsRestKeys` in `stats/internal/repotest/enforce_agent_baseline_hook_test.go`, on Task 10's fixtures with an open row for each: descriptions `visual-verify-rest-gymie verifier remainder` and `visual-verify-rest-2-gymie verifier remainder` in a flow session → allow.
  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/repotest/ -run 'TestEnforceAgentBaselineHook' -count=1` passes; `scripts/check-verbatim-moves.sh` prints no FAIL for the table row.
  - [ ] **Step 5: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow/implement.md`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`
**Tests:** `TestEnforceAgentBaselineHookAdmitsRestKeys`
**Regression:** `TestEnforceAgentBaselineHookAdmitsRestKeys` — a later tightening of `ROW_KEY` that drops the `-rest` keys fails it.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/repotest/enforce_agent_baseline_hook_test.go -->
**Commit:** fix(implement): name the verifier's -rest keys on the closed list
**After:** Task 2, 10

**Decision:** rest-keys-on-closed-list

- [ ] 13. Allow the merge-base stack and refuse an unevidenced pre-existing attribution
  - [ ] **Step 1: Write the failing test.** `TestCheckVerifyReportPreExistingNeedsMergeBaseEvidence` in `stats/internal/guard/check_verify_report_test.go`, on a report otherwise complete (every step `done`): a `## Report` sub-bullet `  - kan-466 weight-tab: … Attributed as pre-existing from the code path only` → exit 1, `VERIFY-REPORT-INCOMPLETE:` naming `10` and `pre-existing attribution with no merge-base evidence`; the same line carrying `merge-base baseline: 4f8f8344:kan-466-weigh-in-history.spec.ts-snapshots/kan466-weight-tab-darwin.png` → exit 0; `present at merge base … merge-base capture: /tmp/x.png` → exit 0; `not pre-existing` → exit 0. Run it: the first case fails.
  - [ ] **Step 2: Fix.** In `checkVerifyReport` (`stats/internal/guard/verifyreport.go`), over the lines of the `## Report` section, a line matching `(?i)(?:^|[^-\w])pre-existing|present at (?:the )?merge base` and not `(?i)\bnot pre-existing`, carrying neither `merge-base baseline: ` nor `merge-base capture: `, makes step 10 undone with the status `pre-existing attribution with no merge-base evidence: <the line, trimmed, at most 120 characters>`. Add the rule to `scripts/check-verify-report.sh`'s header after the fingerprint paragraph: "A `## Report` line attributing a departure as `pre-existing` or `present at the merge base` that names neither `merge-base baseline: <sha>:<path>` nor `merge-base capture: <path>` leaves step 10 undone (kan-1015: a re-baseline over a defect attributed from the code path alone)."
  - [ ] **Step 3: Apply the prose.** In `skills/flow/visual-verify.md` replace the first block with the second; in `skills/flow/visual-verify-verifier.md` step 10 insert the third block immediately before the line `   Every line that reports the departure names its cause or its attribution`:

````text verified:applied by str.replace in this worktree @ 0f456593, old text found once
worktree's build, starting, stopping and restarting nothing, committing and pushing nothing.
````

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed the sentence, listed in verbatim-moves.txt
worktree's build, starting, stopping and restarting nothing but the merge-base stack step 10's
attribution starts, committing and pushing nothing.
````

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed the sentence, listed in verbatim-moves.txt
   A line attributing a departure as present at the merge base names its evidence beside it —
   `merge-base baseline: <merge-base sha>:<baseline png path>` or `merge-base capture: <absolute
   png path>` — and `check-verify-report.sh` refuses a `pre-existing` or `present at the merge
   base` attribution that names neither (kan-1015: the kan-466 weight-tab button was attributed
   pre-existing from the code path alone, and its baseline re-captured over the defect).
````
  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1` passes; `gofmt -l internal/guard`, `go vet ./internal/guard` clean; `scripts/check-verbatim-moves.sh` prints no FAIL for the two sentences.
  - [ ] **Step 5: Commit** the five files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`, `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`
**Tests:** `TestCheckVerifyReportPreExistingNeedsMergeBaseEvidence`
**Regression:** `TestCheckVerifyReportPreExistingNeedsMergeBaseEvidence` — reverting accepts a pre-existing attribution with no merge-base evidence.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/check_verify_report_test.go -->
**Commit:** fix(visual-verify): allow the merge-base stack and refuse an unevidenced pre-existing attribution
**After:** none

**Decision:** merge-base-stack-allowed

**Decision:** attribution-in-verify-report

- [ ] 14. Measure a frame's elements inside its app surface
  - [ ] **Step 1: Write the failing test.** `TestMeasureVisualPropertiesElements` in `stats/internal/guard/measure_visual_properties_test.go`: synthesise two PNGs 120×200 — rows 0–79 a dark scrim (`#555555`), rows 80–199 a white surface with a 1px full-width rule (`#cccccc`) at y=100 and a 40×10 dark text block (`#222222`) at x=20, y=120; the second identical but its text block at y=124. `--props elements --scale 1` prints `a.elements.surface_top` 80, two elements per image (`rule` then `text/control`), and `delta.elements` pairing them with `top` delta 0 then 4 and `gap` delta 4. A third image with no scrim reads `surface_top` 0. Run `cd stats && go test ./internal/guard/ -run 'TestMeasureVisualPropertiesElements' -count=1`: it fails (`elements` is refused).
  - [ ] **Step 2: Fix.** In `stats/internal/guard/measurevisualproperties.go` accept `elements` in `--props` without adding it to `mvpAllProps` (the default set is unchanged), and treat it as a page property beside `bands` and `seams`. Port the run's hand-written matrix script — its surface-top and segmentation functions are reproduced below — to Go: per image, `surface_top`, `background` (hex), and `elements`, each `top`, `height`, `left`, `width`, `right`, `colour` (hex), `kind` (`rule`, `fill band` or `text/control`) and `gap` (from the previous element's bottom, the first from `surface_top`). Pair the two images' elements in order through `mvpPairInOrder` with the `--scale` calibration and emit `delta.elements` and `delta.elements_summary` the way `bands` does; flag nothing. Document `elements` in `scripts/measure-visual-properties.sh`'s property list: "elements — the app surface's top (`surface_top`: the first row, walking column 3 up from the bottom, whose mean channel value is not above 200, plus one; 0 when that row is above row 60 — the whole image is the surface) and every ink run below it against the surface's modal colour (distance over 30), each run's box, modal colour, kind and gap above; paired in order across two images like `bands`. Not in the default set (kan-1015: scrim-dimmed sheet frames read as one band)."

```python verified:copied from /private/tmp/claude-501/-Users-tweety53-Projects-gymie/c483d41b-ee52-4470-baed-b83e5f43d1f0/scratchpad/vv/elems.py, the script kan-1015's verifier measured step 10's matrices with
def top(im):
    w, h = im.size
    y = h - 1
    while y > 0 and sum(im.getpixel((3, y))) / 3 > 200:
        y -= 1
    return 0 if y + 1 < 60 else y + 1

def elems(im, t):
    w, h = im.size
    px = im.load()
    x0, x1 = 8, w - 8
    bg = Counter(px[x, y] for y in range(t, h, 4) for x in range(x0, x1, 4)).most_common(1)[0][0]
    rows = [[x for x in range(x0, x1) if dist(px[x, y], bg) > 30] for y in range(t, h)]
    out = []
    y = 0
    while y < len(rows):
        if not rows[y]:
            y += 1
            continue
        s = y
        while y < len(rows) and rows[y]:
            y += 1
        L = min(min(r) for r in rows[s:y])
        R = max(max(r) for r in rows[s:y])
        mid = rows[s + (y - s) // 2]
        col = Counter(px[x, s + t + (y - s) // 2] for x in mid).most_common(1)[0][0]
        span = max(len(r) for r in rows[s:y])
        kind = 'rule' if (y - s) <= 3 and span > 0.5 * (x1 - x0) else ('fill band' if span > 0.8 * (x1 - x0) else 'text/control')
        out.append(dict(top=s + t, h=y - s, left=L, w=R - L + 1, right=w - 1 - R, col=col, kind=kind))
    prev = t
    for e in out:
        e['gap'] = e['top'] - prev
        prev = e['top'] + e['h']
    return bg, out
```
    <!-- measured: the thresholds 200, 60, 30, 8, 3, 0.5 and 0.8 read from elems.py itself (a session scratch file, not re-runnable after the session) -->
  - [ ] **Step 3: Apply the prose.** In `skills/flow/visual-verify-verifier.md` step 10, matrix item 1, replace `      cells in the new matrix, never its rows.` with:

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed the sentence, listed in verbatim-moves.txt
      cells in the new matrix, never its rows. Seed the enumeration with
      `measure-visual-properties.sh <frame> <capture> --scale <f> --props elements`: it finds each
      image's app surface — a sheet's top edge on a scrim frame — and lists every ink run inside it,
      paired in order with its deltas; every element it merges or misses is added by hand
      (kan-1015: the band read took each scrim-dimmed frame as one band, and every matrix was
      measured with a hand-written script).
````
  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestMeasureVisualProperties' -count=1` passes; `gofmt -l internal/guard`, `go vet ./internal/guard` clean; `scripts/check-verbatim-moves.sh` prints no FAIL for the sentence. Run the built tool on one archived scrim pair (`<backend wt>/spectre/changes/archive/<name>/visual-verification/evidence/templates-a4b-warm-up-sheet-darwin.png` against itself, `--scale 1`): `surface_top` is above 0 and every delta is 0.
  - [ ] **Step 5: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/measurevisualproperties.go`, `stats/internal/guard/measure_visual_properties_test.go`, `scripts/measure-visual-properties.sh`, `skills/flow/visual-verify-verifier.md`
**Tests:** `TestMeasureVisualPropertiesElements`
**Regression:** `TestMeasureVisualPropertiesElements` — reverting refuses `--props elements`.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/measure_visual_properties_test.go -->
**Commit:** feat(guard): measure a frame's elements inside its app surface
**After:** Task 13

**Decision:** elements-opt-in

- [ ] 15. Refuse a scoped re-run that ran another spec count
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `<playwright wt>`, `node <playwright wt>/scripts/specs-for-diff.mjs <frontend wt> c75529d2` prints 2 paths, and `grep '^- verify:' <backend wt>/.superpowers/sdd/verify-report-visual-verify-5-gymie.md` reads `95 mapped specs`.
    <!-- measured: both commands @ backend fe872cce, frontend a4e9f0e5 → 2 paths; "95 mapped specs" -->
  - [ ] **Step 2: Write the failing test.** `TestCheckVerifyReportScopedSpecCount` in `stats/internal/guard/check_verify_report_test.go`: a complete report whose `- verify:` line reads `exit 1 — 95 mapped specs, …` checked with a third argument `2` → exit 1, `VERIFY-REPORT-INCOMPLETE:` naming `7` and `ran 95 mapped specs where specs printed 2`; `exit 0 — 2 mapped specs` with `2` → exit 0; `not run — no spec since the round base` with `0` → exit 0; the 95 line with no third argument → exit 0; a third argument `x` → exit 2. Run it: the first case fails.
  - [ ] **Step 3: Fix.** `checkVerifyReport` accepts two or three arguments; the third, `<scoped specs>`, a non-negative integer, else exit 2 with the usage line updated to `check-verify-report.sh <report file> <motions named> [<scoped specs>]`. Given it, the report's `- verify:` line must read `not run — no spec since the round base` when it is 0, else name `<n> mapped specs` with `<n>` equal to it; otherwise step 7 is undone with the status `ran <k> mapped specs where specs printed <n>` (or `names no mapped-spec count`). Update `scripts/check-verify-report.sh`'s Usage and its header.
  - [ ] **Step 4: Apply the prose.** Three replacements:
    - `skills/flow/visual-verify.md`: the invocation line `check-verify-report.sh <abs-worktree>/.superpowers/sdd/verify-report-<key>.md <motions named>` gains ` [<scoped specs>]`, and the line after the fence, `` `<motions named>` is the count of motions this dispatch's prompt named, `0` for `motions: none`; ``, is followed by the two lines `` `<scoped specs>` is passed for a re-run under a round base alone, the count **The loop** `` and `` (`skills/flow/verify-fix-loop.md`) names in its prompt; ``.
    - `skills/flow/visual-verify-verifier.md` step 7: after the line `   and the report names `specs: not scoped — specs has no <merge-base>`.` insert the second block below.
    - `skills/flow/verify-fix-loop.md` step 4: replace `` `<frontend-root>` worktree has no round base, the prompt names none and `specs` runs unscoped. A worktree `` with the first block below.

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed the sentence, listed in verbatim-moves.txt
   `<frontend-root>` worktree has no round base, the prompt names none and `specs` runs unscoped.
   Before that dispatch the parent runs `specs` itself with the round base for `<merge-base>`, names
   the count of paths it prints in the prompt as `scoped specs: <n>`, and passes `<n>` to
   `check-verify-report.sh` as its third argument, which refuses a report whose `verify:` line
   names another count (kan-1015: a re-run under round base `c75529d2` printed 2 specs, and the
   verifier ran 95). A worktree
````

````text verified:applied by str.replace in this worktree @ 0f456593; check-verbatim-moves.sh then printed the sentence, listed in verbatim-moves.txt
   Under a round base the `verify:` line names the count of paths `specs` printed as `<n> mapped
   specs` beside its exit.
````
  - [ ] **Step 5: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1` passes; `gofmt -l internal/guard`, `go vet ./internal/guard` clean; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` print no FAIL for these lines.
  - [ ] **Step 6: Commit** the six files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`, `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`, `skills/flow/verify-fix-loop.md`
**Tests:** `TestCheckVerifyReportScopedSpecCount`
**Regression:** `TestCheckVerifyReportScopedSpecCount` — reverting accepts a scoped re-run that ran the whole suite.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^func Test' stats/internal/guard/check_verify_report_test.go -->
**Commit:** fix(visual-verify): refuse a scoped re-run that ran another spec count
**After:** Task 3, 4, 9, 13, 14

**Decision:** scoped-specs-count

- [ ] 16. Fix the kan-466 weight-tab button and restore its baseline
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read the pass's diff image `/private/tmp/claude-501/-Users-tweety53-Projects-gymie/c483d41b-ee52-4470-baed-b83e5f43d1f0/scratchpad/rerun/kan-466-weigh-in-history-K-1a91a-gainst-the-extended-backend-functional/kan466-weight-tab-diff.png`, `git -C <playwright wt> show 6554b4d^:kan-466-weigh-in-history.spec.ts-snapshots/kan466-weight-tab-darwin.png` and the committed one; measure the "Log a weigh-in" button on each (`measure-visual-properties.sh … --props fill,content`). Then capture the weight tab at the frontend merge base `4f8f8344` exactly as step 10 of `skills/flow/visual-verify-verifier.md` prescribes a merge-base capture (detached `<name>-merge-base` worktrees, its own derived workspace, the kan-466 spec), and at the change's HEAD on the change's stack. Record which renders draw fill `224,234,237` / ink `81,207,255`.
    <!-- measured: the fill and ink values quoted from the pass's finding, read off its diff image (a session scratch file) -->
  - [ ] **Step 2: Trace the cause** — `BroadsheetAccentOutlineButton` and `BodyEmptyHistory` (`shared/src/commonMain/kotlin/com/gymie/shared/ui/body/BodyScreen.kt`), what the spec does before its capture (a click, a focus, the parked pointer), and what the merge brought to `WeighInHistory`/`BodyState`; name the line that draws the tinted state. unverified:whether the tint is a hover, focus or disabled state, and whether its cause is in gymie-frontend or the kan-466 spec.
  - [ ] **Step 3: Fix the cause where it lives**, with a test that drives the real composable (or the spec's real path) and asserts the resting state's accent ink on a plain fill, failing without the fix; commit in that worktree as `fix(<the module the trace names>): <what it fixes>` with the Flow-Key trailer.
  - [ ] **Step 4: Restore the baseline.** Copy `6554b4d^`'s `kan466-weight-tab-darwin.png` back when the fixed render matches it, else re-capture it from the fixed build and read the PNG before staging; run the kan-466 spec against the change's rebuilt stack: it passes. Commit in `<playwright wt>` as `test(fidelity): restore the kan-466 weight-tab baseline` with the Flow-Key trailer, and rebuild the full app suite zip (`rebuild-full-app-suite-zip.sh`) when the baseline is in it.

**Build:** green
**Files:** `kan-466-weigh-in-history.spec.ts-snapshots/kan466-weight-tab-darwin.png`
**Allowed-collateral:** `shared/src/**`, `kan-466-weigh-in-history.spec.ts`, `full-app-suite.zip`
**Tests:** **none** — the cause's own test is named once the trace places it; the kan-466 capture against the restored baseline is the regression check.
**Regression:** reverting the fix fails the kan-466 capture against the restored baseline.
**Commit:** test(fidelity): restore the kan-466 weight-tab baseline
**After:** none

**Decision:** product-trace-first
