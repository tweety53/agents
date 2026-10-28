# kan-1046-review-dynamics-dashboard

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a Review dynamics view trending review-panel, implementer and visual-verify outcomes per day or week, with verdict tiles against the previous range, a slot × model · effort heatmap, hover tooltips, prompt-change markers, view-owned range presets, and a `prompt_changes` store fed by `flow prompt-tuning record`/`import`.

**Spec:** `design.md` beside this file; the approved prototype `/Users/tweety53/Projects/agents-worktrees/kan-1046-review-dynamics-dashboard/.superpowers/sdd/review-dynamics-prototype.html` (https://claude.ai/artifact/QCFtCagfSawyAdVUepj73C, version 2) is the layout, copy and behaviour reference.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1046-review-dynamics-dashboard`. Every `**Files:**` path is relative to its root. Never touch the main checkout.
- Never stop the dev workspace's `flowd` or `flow-postgres`, and never write to the dev `flow` database; store tests use the test database their harness already creates.
- No new dependency; the charts stay inline SVG.
- Go lint per task: `cd stats && make -s web-dist && go vet ./... && gofmt -l .` must print nothing. SPA lint: `cd stats/web && npx tsc -b`.

- [ ] 1. Store: canonical_severity and the prompt_changes table
  - [ ] **Step 1: Write the failing tests** in `stats/internal/store/promptchanges_test.go`, on the package's existing test-database harness:
    - `TestCanonicalSeverityFolds` — `SELECT canonical_severity(x)` for `Critical`, `blocking`, `Severe` → `critical`; `Important`, `MAJOR`, `high`, `Medium`, `moderate` → `important`; `Minor`, `low`, `Informational`, `closed` → `minor`.
    - `TestUpsertPromptChangeTuningOverEdit` — `UpsertPromptChange` an `edit` row (sha `abc:skills/flow/primary-reviewer-prompt.md`, area `review`, slot `primary`), then the same sha as `tuning` with case id `kan-1-x-1`: one row, kind `tuning`, case id set; a second `edit` upsert of that sha leaves it `tuning`. `ListPromptChanges(ctx, from, to)` returns rows ordered by `at` within `[from, to)`; `LastPromptChangeAt(ctx)` returns the newest `at`, nil on an empty table. An area outside `review|impl|vv|fix` is refused.
    Run `cd stats && go test ./internal/store/ -run 'TestCanonicalSeverityFolds|TestUpsertPromptChange'` — RED.
  - [ ] **Step 2: Implement** `stats/internal/store/migrations/0049_review_dynamics.sql` (header comment as `0028_canonical_slots.sql` writes one): `canonical_severity(raw text) RETURNS text LANGUAGE sql IMMUTABLE STRICT` per `canonical-severity`, and `prompt_changes` per `prompt-changes-table` with an index on `at`. Then `stats/internal/store/promptchanges.go`: `type PromptChange struct{ SHA string; At time.Time; Kind, Area string; Slot *string; Target, Summary string; CaseID *string }`, `UpsertPromptChange` (`ON CONFLICT (sha) DO UPDATE` keeping `tuning` over `edit`: `kind = CASE WHEN prompt_changes.kind = 'tuning' THEN 'tuning' ELSE EXCLUDED.kind END`, `case_id = COALESCE(EXCLUDED.case_id, prompt_changes.case_id)`, the other fields from the tuning row when the incoming kind is `tuning`), `ListPromptChanges`, `LastPromptChangeAt`. Doc comments cite the decisions. Register the migration wherever `migrations.go` requires it. <!-- unverified:confirm migrations.go embeds the directory, so no list edit is needed -->
  - [ ] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/store/ -run 'TestCanonicalSeverityFolds|TestUpsertPromptChange|TestMigrat'`.

**Files:** `stats/internal/store/migrations/0049_review_dynamics.sql`, `stats/internal/store/promptchanges.go`, `stats/internal/store/promptchanges_test.go`
**Allowed-collateral:** `stats/internal/store/migrations.go`
**Tests:** `TestCanonicalSeverityFolds`, `TestUpsertPromptChangeTuningOverEdit`
**Regression:** `TestCanonicalSeverityFolds` fails, because the function does not exist; `TestUpsertPromptChangeTuningOverEdit` fails, because there is no `prompt_changes` table.
**Baseline:** before=332 after=334 <!-- measured: grep -rhE '^func Test' stats/internal/store | wc -l @ c4202f60 -->
**Build:** green
**After:** none
**Commit:** `feat(store): fold severities and record prompt changes`

**Decision:** canonical-severity
**Decision:** prompt-changes-table

- [ ] 2. Store: the ReviewDynamics aggregate
  - [ ] **Step 1: Write the failing tests** in `stats/internal/store/reviewdynamics_test.go`, fixtures built with the package's existing `BeginStage`/`RecordDispatch`/`UpsertFinding`/`RecordDecision` helpers:
    - `TestReviewDynamicsSeries` — two weeks; week 1: change A (first reviewer dispatch Mon, `primary+principles` on `opus`/`medium`, findings round 0 `Critical` primary, `major` principles, round 2 `minor` primary; two implementer dispatches `opus`/`low`, one `blocked`; verifier dispatches `visual-verify`, `visual-verify-fix-1`, `visual-verify-fix-2`), change B (one `test-sanity` dispatch `sonnet`/`low`, no findings, no verifier dispatch); week 2: change C (one `primary` dispatch `sonnet`/`high`, one `Minor` finding with no dispatch seq; verifier `visual-verify` only). Period covers both weeks, 15 days, so `bucket` is `week`. Expect: `findings-per-change` week 1 critical 0.5, important 0.5, minor 0.5, week 2 minor 1; `fix-rounds` week 1 mean 1 (A 2, B 0), max 2; `changes` 2 then 1; `per-dispatch-slot` week 1 primary 2 (two findings ÷ one dispatch), principles 1, test-sanity 0; `per-dispatch-model` week 1 opus 3 (three findings ÷ one dispatch row — model and effort rates count dispatch rows, never unnested slots), sonnet 0; `per-dispatch-effort` week 1 medium 3, low 0; `impl-findings-per-task` week 1 `opus · low` 1.5 (3 findings ÷ 2 dispatches); `impl-blocked-share` week 1 0.5; `vv-rounds` week 1 mean 2 max 2, week 2 mean 0; `vv-first-pass` week 1 0, week 2 1. Every row carries its `basis` (changes or dispatches counted). A period of 14 days or less returns `bucket` `day`. A `project` restricts every series.
    - `TestReviewDynamicsHeatSummaryMarkers` — heat rows `{slot, model, effort, dispatches, critical, important, minor}` for the range and for the previous range of the same length; `summary` and `previousSummary` carry `findingsPerChange`, `criticalImportantPerChange`, `meanFixRounds`, `implFindingsPerTask`, `vvFirstPassShare`, `changes`; `markers` are the `prompt_changes` rows inside the range; `lastPromptChangeAt` is the newest row whatever the range.
    Run `cd stats && go test ./internal/store/ -run 'TestReviewDynamics'` — RED.
  - [ ] **Step 2: Implement** `stats/internal/store/reviewdynamics.go`: `type ReviewDynamics struct{ Bucket string; Series []DynamicsPoint; Heat, PreviousHeat []HeatCell; Summary, PreviousSummary DynamicsSummary; Markers []PromptChange; LastPromptChangeAt *time.Time }`, `DynamicsPoint{Bucket time.Time; Metric, Series string; Value float64; Basis int}`, and `func (s *Store) ReviewDynamics(ctx context.Context, period Period, project *string) (ReviewDynamics, error)`. Bucket per `bucket-from-range` with `date_trunc('day'|'week', … AT TIME ZONE 'UTC')`. Per-change cohorts per `cohort-by-change`, per-dispatch rates by unnesting the slot as `Store.Reviewers` does and joining findings on `dispatch_id` and slot; severities through `canonical_severity`; implementer and visual-verify per `impl-vv-metrics`. Metric ids exactly: `findings-per-change`, `per-dispatch-slot`, `fix-rounds` (series `mean`, `max`), `per-dispatch-model`, `per-dispatch-effort`, `changes`, `impl-findings-per-task`, `impl-blocked-share`, `vv-rounds` (`mean`, `max`), `vv-first-pass`. The summary is the same formulas over the whole range; the previous range per `previous-range`. One SQL statement per metric family is fine; no per-bucket loop of queries. Doc comment states every formula.
  - [ ] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/store/ -run 'TestReviewDynamics|TestReviewers'`.

**Files:** `stats/internal/store/reviewdynamics.go`, `stats/internal/store/reviewdynamics_test.go`
**Tests:** `TestReviewDynamicsSeries`, `TestReviewDynamicsHeatSummaryMarkers`
**Regression:** `TestReviewDynamicsSeries` fails, because no aggregate trends findings over time; `TestReviewDynamicsHeatSummaryMarkers` fails, because no previous-range comparison or markers exist.
**Baseline:** before=334 after=336 <!-- measured: grep -rhE '^func Test' stats/internal/store | wc -l @ c4202f60 -->
**Build:** green
**After:** Task 1
**Commit:** `feat(store): review dynamics series, heatmap, summary and markers`

**Decision:** cohort-by-change
**Decision:** impl-vv-metrics
**Decision:** bucket-from-range
**Decision:** previous-range
**Decision:** markers-by-area

- [ ] 3. API: the review-dynamics view and POST prompt-changes
  - [ ] **Step 1: Write the failing tests:** in `stats/internal/api/stats_test.go`, `TestStatsReviewDynamics` — `GET /api/v1/stats/review-dynamics?from=…&to=…&project=p` answers `rows` as one object `{bucket, series: [{bucket: "YYYY-MM-DD", metric, series, value, basis}], heat, previousHeat, summary, previousSummary, markers: [{sha, at, kind, area, slot, target, summary, caseId}], lastPromptChangeAt}` from the fake store, passes the project through, and refuses a `model` parameter as the state board does. In `stats/internal/api/promptchanges_test.go`, `TestPostPromptChange` — `POST /api/v1/prompt-changes` with a JSON row calls `UpsertPromptChange` and answers 204; a body with an unknown area or a missing sha answers 400. Run `cd stats && go test ./internal/api/ -run 'TestStatsReviewDynamics|TestPostPromptChange'` — RED.
  - [ ] **Step 2: Implement:** in `stats/internal/api/stats.go`, `viewReviewDynamics viewName = "review-dynamics"` in the view list, the model-restriction refusal and `rowsFor`, with DTOs and converter; `ReviewDynamics` and `UpsertPromptChange` on the store interface. `stats/internal/api/promptchanges.go`: the POST handler; route it in `stats/internal/api/server.go`. Add the methods to every fake store the package's tests use.
  - [ ] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/api/ -run 'TestStatsReviewDynamics|TestPostPromptChange|TestStats'`.

**Files:** `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/promptchanges.go`, `stats/internal/api/promptchanges_test.go`, `stats/internal/api/server.go`
**Allowed-collateral:** `stats/internal/api/*_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`
**Tests:** `TestStatsReviewDynamics`, `TestPostPromptChange`
**Regression:** `TestStatsReviewDynamics` fails, because the view name is unknown and answers 400; `TestPostPromptChange` fails, because the route does not exist.
**Baseline:** before=182 after=184 <!-- measured: grep -rhE '^func Test' stats/internal/api | wc -l @ c4202f60 -->
**Build:** green
**After:** Task 2
**Commit:** `feat(api): review-dynamics view and prompt-changes endpoint`

**Decision:** bucket-from-range

- [ ] 4. CLI: flow prompt-tuning record and import
  - [ ] **Step 1: Write the failing tests:**
    - `stats/internal/prompttuning/prompttuning_test.go`: `TestAreaFor` — each mapped path of `prompt-area-map` returns its area and slot; an unmapped path returns not-found.
    - `stats/internal/client/promptchanges_test.go`: `TestRecordPromptChange` — posts the row as JSON to `/api/v1/prompt-changes` on an `httptest` server and surfaces a 400 body as an error.
    - `stats/cmd/flow/prompttuning_test.go`: `TestPromptTuningRecord` — `flow prompt-tuning record <case-dir> -sha <sha> -at 2026-10-27T10:00:00Z` with a `case.json` targeting `skills/flow/test-sanity-reviewer-prompt.md` posts kind `tuning`, area `review`, slot `test-sanity`, `case_id` the case id, summary the defect class, `sha` `<sha>:<target>`; a case targeting `rules/x.mdc` of type `implementer` posts area `impl`, slot from `source.slot` empty; a missing `-sha` exits 2. `TestPromptTuningImport` — in a temp git repo with two commits on `skills/flow/primary-reviewer-prompt.md` (one also touching `skills/flow/visual-verify-verifier.md`) and one on an unmapped file, `flow prompt-tuning import -C <repo>` posts three `edit` rows keyed `<sha>:<target>` with the commit time and subject; `-since` drops older commits; prints `imported: 3`.
    Run `cd stats && go test ./internal/prompttuning/ ./internal/client/ ./cmd/flow/ -run 'TestAreaFor|TestRecordPromptChange|TestPromptTuning'` — RED.
  - [ ] **Step 2: Implement:** `AreaFor(target string) (area, slot string, ok bool)` and the map in `stats/internal/prompttuning/prompttuning.go`, plus `MappedTargets()` for import; `func (c *Client) RecordPromptChange(ctx, row) error` in `stats/internal/client/promptchanges.go`; in `stats/cmd/flow/prompttuning.go`, `record` and `import` subcommands beside `render` (usage text updated), `record` reading `case.json` through `prompttuning.Load` and defaulting `-at` to the sha's commit time from `git show -s --format=%cI <sha>` run in the case dir's repository, `import` per `git-import-backfill`.
  - [ ] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/prompttuning/ ./internal/client/ ./cmd/flow/ -run 'TestAreaFor|TestRecordPromptChange|TestPromptTuning'`.

**Files:** `stats/internal/prompttuning/prompttuning.go`, `stats/internal/prompttuning/prompttuning_test.go`, `stats/internal/client/promptchanges.go`, `stats/internal/client/promptchanges_test.go`, `stats/cmd/flow/prompttuning.go`, `stats/cmd/flow/prompttuning_test.go`
**Tests:** `TestAreaFor`, `TestRecordPromptChange`, `TestPromptTuningRecord`, `TestPromptTuningImport`
**Regression:** `TestPromptTuningRecord` fails, because `record` is an unknown subcommand and exits 2; `TestPromptTuningImport` fails likewise; `TestAreaFor` and `TestRecordPromptChange` fail to compile.
**Baseline:** before=354 after=356 <!-- measured: grep -rhE '^func Test' stats/cmd/flow | wc -l @ c4202f60 -->
**Build:** green
**After:** Task 3
**Commit:** `feat(stats): flow prompt-tuning record and import`

**Decision:** prompt-area-map
**Decision:** git-import-backfill

- [ ] 5. Skill: /flow-prompt-tuning records each applied case
  - [ ] **Step 1:** in `skills/flow-prompt-tuning/SKILL.md` step **6. Ledger and cleanup**, before the case folder's `git rm`: for every applied case, `flow prompt-tuning record <case-dir> -sha <the applying commit's sha>`; a non-zero exit is reported with its stderr and the step continues (the store is never a gate). State that the record feeds the Review dynamics view's markers. Keep every existing sentence.
  - [ ] **Step 2: Verify:** `scripts/check-references.sh` and every other `scripts/check-*.sh` guard `.flow/project.md` `## lint` names for skill files.

**Files:** `skills/flow-prompt-tuning/SKILL.md`
**Tests:** **none** — a skill text change; the guards cover it.
**Build:** green
**After:** Task 4
**Commit:** `feat(flow-prompt-tuning): record each applied case in the stats store`

**Decision:** prompt-changes-table

- [ ] 6. SPA: the Review dynamics view
  - [ ] **Step 1: Write the failing tests** in `stats/web/src/views/views.test.tsx` (fixture for the new slug, as the other views have):
    - `review dynamics draws every section with tiles against the previous range` — headings `Review panel`, `Implementer`, `Visual verify`, `Where findings come from`; the ten chart panels by their prototype titles; tile `Findings per change` reads the summary value and `▼ 25% · better` for a 0.75 vs 1.0 fixture; `Visual verify passed first time` with a higher value reads `better`; `Changes reviewed` carries no better/worse word; a ±2 % change reads `≈ unchanged`.
    - `review dynamics range presets set the period` — clicking `Last 2 weeks` calls `onPeriodChange` with a 14-day range ending now; `Since last prompt change` uses `lastPromptChangeAt` and is disabled when it is null; with no `from`/`to` in the URL the active preset is `Last 8 weeks`.
    - `review dynamics tooltip lists values, basis and marker` — hovering the severity chart's overlay at a marker bucket shows each series' value, `based on N changes` and the marker text; hovering a heatmap cell's `title` carries dispatches and counts.
    Run `cd stats/web && npx vitest run -t 'review dynamics'` — RED.
  - [ ] **Step 2: Implement:**
    - `stats/web/src/api.ts`: `review-dynamics` in `VIEW_NAMES` and the response types.
    - `stats/web/src/components/DynamicsChart.tsx`: per `dynamics-chart-component`, drawn exactly as the prototype's `chart()` (480×170 viewBox, padding l34 r10 t12 b22, three y ticks, ≤6 date labels, dashed marker line + dot, crosshair, tooltip `Week of MM-DD` or `MM-DD`, `based on N <basis>`, `◆ <what>: <text>` marker lines), series colours `--series-n`, marker colour a new `--marker` token in `stats/web/src/theme.css` (dark `#b877d9`, light `#8250df`).
    - `stats/web/src/views/ReviewDynamics.tsx`: the prototype's range row, tiles, three sections, heatmap (rows = slots, columns = model · effort seen in range, cell = findings per dispatch shaded by `--heat` alpha `0.12 + 0.78 × rate/max`, arrow vs `previousHeat`, `title` with dispatches and counts, `—` for an empty cell), the footnote copy. Copy and section order exactly as the prototype; the header line drops the prototype's "Prototype with sample data." sentence only.
    - `stats/web/src/App.tsx`: label `Review dynamics`, component, model filter disabled with a reason; `stats/web/src/components/DashboardBar.tsx`: no `PeriodPicker` on this view per `view-owned-range`.
    - `stats/web/src/styles.css`: the prototype's panel, tile, tooltip and heatmap rules on existing tokens.
  - [ ] **Step 3: Verify:** SPA lint, `cd stats/web && npx vitest run -t 'review dynamics'`.

**Files:** `stats/web/src/api.ts`, `stats/web/src/App.tsx`, `stats/web/src/components/DashboardBar.tsx`, `stats/web/src/components/DynamicsChart.tsx`, `stats/web/src/views/ReviewDynamics.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/styles.css`, `stats/web/src/theme.css`
**Allowed-collateral:** `stats/web/src/**/*.test.tsx`, `stats/web/src/api.test.ts`
**Tests:** `review dynamics draws every section with tiles against the previous range`, `review dynamics range presets set the period`, `review dynamics tooltip lists values, basis and marker`
**Regression:** each fails, because no `review-dynamics` view is registered.
**Baseline:** before=175 after=178 <!-- measured: grep -rhE '^\s*(it|test)\(' stats/web/src | wc -l @ c4202f60 -->
**Build:** green
**After:** Task 3
**Commit:** `feat(stats): Review dynamics view`

**Decision:** view-owned-range
**Decision:** dynamics-chart-component
**Decision:** markers-by-area
**Decision:** previous-range

- [ ] 7. Seed and baseline the Review dynamics visual spec
  - [ ] **Step 1:** in `stats/cmd/uitest-seed/seed.go`, seed inside `PINNED_QUERY`'s window (2026-08-01..08-16) three or more changes with reviewer dispatches on two models and efforts, findings across rounds and severities, implementer dispatches (one blocked), verifier `visual-verify`/`visual-verify-fix-1` dispatches, and two `prompt_changes` rows (one `review`, one `vv`), so every chart and the heatmap draw.
  - [ ] **Step 2:** `stats/web/tests/visual/review-dynamics.spec.ts`: `review dynamics full page` (goto `/#/review-dynamics?${PINNED_QUERY}`, heading, `bodyBackground` dark, full-page screenshot `review-dynamics.png`), `review dynamics chart hover` (hover the severity chart overlay at a marker bucket, screenshot `review-dynamics-hover.png`), `review dynamics light` (light scheme). Add the view to `full-app-suite.spec.ts`'s list if it enumerates views. Capture per `.flow/project.md` `## visual verification` (`--update-snapshots`), then run the verify row clean; open each PNG and compare with the prototype; fix overlap or clipping before committing. Rebuild `stats/web/tests/visual/full-app-suite.zip` if the suite's snapshots changed.

**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/review-dynamics.spec.ts`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-darwin.png`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-hover-darwin.png`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-light-darwin.png`
**Allowed-collateral:** `stats/web/tests/visual/*.spec.ts`, `stats/web/tests/visual/*.spec.ts-snapshots/*.png`, `stats/web/tests/visual/full-app-suite.zip`
**Tests:** `review dynamics full page`, `review dynamics chart hover`, `review dynamics light`
**Regression:** `review dynamics full page` fails, because the route renders no view.
**Baseline:** before=53 after=56 <!-- measured: grep -rhE '^\s*(it|test)\(' stats/web/tests/visual | wc -l @ c4202f60 -->
**Build:** green
**After:** Task 6
**Commit:** `test(stats): baseline the Review dynamics visual spec`
