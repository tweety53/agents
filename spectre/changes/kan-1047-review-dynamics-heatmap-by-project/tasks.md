# kan-1047-review-dynamics-heatmap-by-project

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, split by source
> set, counted by a word-boundary-safe counter, with any `measured:` command naming no ref of its own.

**Goal:** Projects, Task size and Reviewers bars on Review dynamics, four-week default (`proposal.md`).

**Spec:** `design.md`; prototype `.superpowers/sdd/heatmap-by-project-prototype.html`.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1047-review-dynamics-heatmap-by-project`; never the main checkout.
- `cd stats && gofmt -l .` prints nothing; `cd stats && make -s web-dist && go vet ./...` and `cd stats/web && npx tsc -b` exit 0.

- [x] 1. Review dynamics filters by project, task size and current reviewers
  - [x] **Step 1: Store.** `ReviewDynamics(ctx, period, project, class *string)`: `Projects` (distinct reviewed project keys in the period), `classFilterSQL` on every query as `$5`, implementer outcomes moved to `$6`.
  - [x] **Step 2: API.** `class` in `statsQueryParams`; 400 off review-dynamics or outside micro/small/regular/big; `projects` on the DTO.
  - [x] **Step 3: SPA.** `ViewProps.onProjectChange` wired from App; `StatsViewParams.class` sent and in `useStatsView`'s deps; Projects, Task size and Reviewers bars; `CURRENT_SLOTS` filter; default preset `"4"`.
  - [x] **Step 4: Tests.** `TestReviewDynamicsClassAndProjects`, the class cases in `TestStatsReviewDynamics`, the vitest `review dynamics filters by project, task size and retired reviewers`; existing fixtures move to current slot names and the 28-day default.
  - [x] **Step 5: Verify.** `cd stats && go test ./internal/store ./internal/api -run 'TestReviewDynamics|TestStatsReviewDynamics' -count=1`; `cd stats/web && npx vitest run -t 'review dynamics'`; the global-constraint lint lines. Commit.

**Build:** green
**Files:** `stats/internal/store/reviewdynamics.go`, `stats/internal/store/reviewdynamics_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/promptchanges.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`, `stats/web/src/App.tsx`, `stats/web/src/api.ts`, `stats/web/src/viewTypes.ts`, `stats/web/src/hooks/useStatsView.ts`, `stats/web/src/views/ReviewDynamics.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `TestReviewDynamicsClassAndProjects`
**Regression:** `TestReviewDynamicsClassAndProjects` fails — a class no longer narrows the changes, or `Projects` is empty.
**Baseline:** before=3 after=4 <!-- measured: grep -c -E '^func Test' stats/internal/store/reviewdynamics_test.go @ origin/main -->
**Commit:** feat(stats): filter review dynamics by project, task size and current reviewers
**After:** none

- [x] 2. Two charts per row and a Prompt changes marker filter
  - [x] **Step 1: Layout.** `.dynamics-grid` is always two columns (`repeat(2, minmax(0, 1fr))`).
  - [x] **Step 2: Prompt changes bar.** `flow-prompt-tuning` (default, markers of kind `tuning`) · `All` · `auto prompt tuning` (markers of kind `edit`); filtered client side in `markersFor`.
  - [x] **Step 3: Test.** vitest `review dynamics filters prompt-change markers by kind`.
  - [x] **Step 4: Verify.** `cd stats/web && npx vitest run -t 'review dynamics'`; `npx tsc -b`. Commit.

**Build:** green
**Files:** `stats/web/src/styles.css`, `stats/web/src/views/ReviewDynamics.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `review dynamics filters prompt-change markers by kind`
**Regression:** the vitest fails — an `edit` marker draws under `flow-prompt-tuning`.
**Commit:** feat(stats): two-column review dynamics and a prompt-change marker filter
**After:** Task 1

- [x] 3. Drop the "Since last prompt change" range preset
  - [x] **Step 1: Test.** vitest `review dynamics range presets set the period` asserts the Range bar has no "Since last prompt change" button.
  - [x] **Step 2: SPA.** Remove the `tune` preset from `PRESETS`, `Preset` and `presetPeriod`.
  - [x] **Step 3: Verify.** `cd stats/web && npx vitest run -t 'review dynamics'`; `npx tsc -b`. Commit.

**Build:** green
**Files:** `stats/web/src/views/ReviewDynamics.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `review dynamics range presets set the period`
**Regression:** the vitest fails — the Range bar still offers "Since last prompt change".
**Commit:** feat(stats): drop the since-last-prompt-change range preset
**After:** Task 2

- [x] 4. In-run fix 1 — flow.visual-verify: stale default-preset and chart-hover assertions; new bar states uncaptured
  - [x] **Step 1: Spec.** `review dynamics default preset` asserts `Last 4 weeks`; `review dynamics chart hover` picks Prompt changes `All` before hovering; a new `review dynamics filters` capture picks the fixture project, Reviewers `All` and Prompt changes `auto prompt tuning`, and a Task size button, asserting each pressed.
  - [x] **Step 2: Capture.** The project's `capture` with `review-dynamics.spec.ts` and `--update-snapshots=all`; every changed PNG read; `rebuild-full-app-suite-zip.sh`; `visual-verification.md` rewritten.
  - [x] **Step 3: Verify.** `capture` exits 0. Commit.

**Build:** green
**Files:** `stats/web/tests/visual/review-dynamics.spec.ts`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-darwin.png`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-version-darwin.png`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-default-preset-darwin.png`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-hover-darwin.png`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-heat-hover-darwin.png`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-light-darwin.png`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-filters-darwin.png`, `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-review-dynamics-darwin.png`, `stats/web/tests/visual/full-app-suite.zip`, `spectre/changes/kan-1047-review-dynamics-heatmap-by-project/visual-verification.md`
**Tests:** `review dynamics default preset`, `review dynamics chart hover`, `review dynamics filters`
**Regression:** `review dynamics filters` fails — a bar's pick no longer presses its button or the page no longer matches its baseline.
**Commit:** test(stats): capture the review dynamics bars and refresh their baselines
**After:** Task 3

- [x] 5. In-run fix 2 — flow.visual-verify: no capture picks a project or a task size
  - [x] **Step 1: Seed.** `seedReviewDynamics` records one decision per `uitest-dynamics` change, class `big` for the first change of each week and `small` for the second, so a Task size pick has data to draw.
  - [x] **Step 2: Spec.** `review dynamics filters` also captures `review-dynamics-picked.png` after picking project `uitest-dynamics` and Task size `Big`, each asserted pressed.
  - [x] **Step 3: Capture.** Restart the UI-test stack (re-seed), the project's `capture` for `review-dynamics.spec.ts` with `--update-snapshots=all` and for the full app suite, every changed PNG read, `rebuild-full-app-suite-zip.sh`, `visual-verification.md` updated; `npm run test:visual` exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/review-dynamics.spec.ts`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-picked-darwin.png`, `spectre/changes/kan-1047-review-dynamics-heatmap-by-project/visual-verification.md`
**Tests:** `review dynamics filters`
**Regression:** `review dynamics filters` fails — the project or Task size pick no longer presses its button or no longer matches its baseline.
**Commit:** test(stats): seed plan classes and capture a project and task-size pick
**After:** Task 4

Correction (2026-10-28): the plan declared `stats/web/tests/visual/full-app-suite.zip`; the rebuilt zip came out byte-identical (the review-dynamics screen of the full app suite did not move), so the commit does not carry it.

- [ ] 6. In-run fix 3 — flow.visual-verify: duplicate `1` tick on a small integer axis; pointer left on a button in the filters captures
  - [ ] **Step 1: Test.** vitest `dynamics chart drops a tick whose label repeats the one below` renders `DynamicsChart` with an integer format over values peaking at 1 and asserts the y labels read `0`, `1`.
  - [ ] **Step 2: Chart.** `DynamicsChart` skips a y tick whose formatted label equals the previous tick's (the shared source: every chart with an integer format).
  - [ ] **Step 3: Spec.** `review dynamics filters` moves the pointer to (0, 0) before each screenshot.
  - [ ] **Step 4: Capture.** `capture` for `review-dynamics.spec.ts` and the full app suite with `--update-snapshots=all`, changed PNGs read, zip rebuilt when it moves, `visual-verification.md` updated; `npm run test:visual` and the vitest exit 0. Commit.

**Build:** green
**Files:** `stats/web/src/components/DynamicsChart.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/tests/visual/review-dynamics.spec.ts`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-picked-darwin.png`, `spectre/changes/kan-1047-review-dynamics-heatmap-by-project/visual-verification.md`
**Tests:** `dynamics chart drops a tick whose label repeats the one below`, `review dynamics filters`
**Regression:** the vitest fails — an integer axis peaking at 1 labels two ticks `1`.
**Commit:** fix(stats): drop a repeated axis label on small integer review dynamics charts
**After:** Task 5
