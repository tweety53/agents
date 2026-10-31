## Context

KAN-1052 reworks the stats SPA's Stage leaderboard (`stats/web/src/views/StageLeaderboard.tsx`) so it
answers **"Which stages take the most of the spend, and which are growing?"** — asked before deciding
which stage of the pipeline to make cheaper next.

**The spec is the approved prototype**, version 3:
`.superpowers/sdd/stage-leaderboard-prototype.html` in this change's worktree
(`/Users/tweety53/Projects/agents-worktrees/kan-1052-rework-the-stage-leaderboard-dashboard`),
published at https://claude.ai/artifact/MnKPxwCnnSsspu4ULnePfs. Its markup, its `sl-*` rules, its
copy (headings, column headers, the tile notes, "Other N stages", the empty-state line) and its
number formats are built as drawn. Its `DATA` array is illustrative dev-store data; the built view
reads the endpoint.

What constrains it:

- **It is one change** because every part serves the same question on the same page: the endpoint's
  uncapped rows and per-project spend exist only for this view's Share column, Other row and Project
  chips, and the chip bar, tiles and table are one layout.
- **The view sits inside the real app shell** (`stats/web/src/App.tsx`'s header with the "flow
  stats" title and the horizontal view links) and uses the app's existing theme tokens
  (`stats/web/src/theme.css`) and `lb-*` styles (`stats/web/src/styles.css`); the app's light/dark
  behaviour is unchanged.
- **Prior decisions this change re-decides**, all from archived changes: `leaderboard-top10-total`
  (the endpoint's `LIMIT 10`), `leaderboard-main-columns` (only its Main session $ column; Main share
  stays) and the shared dashboard bar on this view. It follows KAN-1048's
  `board-chip-filters-run-none` chip pattern, `grouped-tables-own-markup` plain-table precedent and
  `hide-dashboard-bar` header rule.
- **The visual specs pin the browser clock** to the UI-test fixture's `FIXTURE_NOW`
  (`stats/web/tests/visual/support.ts`) because the fixture's timestamps are fixed at seed time; a
  rolling "ending now" window needs the same pin.

## Decisions

### The question the view answers

**ID:** leaderboard-spend-and-growth
**Status:** active
**Chosen:** "Which stages take the most of the spend, and which are growing?" — spend ranking plus each stage's change against the previous equal window; it is the question asked before choosing which stage to make cheaper, and the growth half is what today's view hides.
**Considered:** Spend only, as today — leaves the growth that motivated the ticket invisible. Per-run cost ranked by median/P90 — answers "which run is dear", not "where the money goes". Reliability and time (duration, failed outcomes) — a different question needing a new aggregate.

### The endpoint returns every costed stage

**ID:** leaderboard-transport-uncapped
**Status:** active
**Chosen:** `StageLeaderboard` drops its `LIMIT 10` and returns every costed command/stage of the window, still ordered by spend; the view keeps ten ranked rows and sums the rest into an "Other N stages" row. Re-decides archived `leaderboard-top10-total` for the transport only — the UI still caps the ranked rows at ten. The window's spend, every Share and the Command chip spend derive from the same rows.
**Considered:** Keep `LIMIT 10` and take the period's spend from a third call to `trend` — a second source for the denominator that can disagree with the rows, and no Other row. Keep `LIMIT 10` with no Share and no Other row — drops the first half of the question.

### Chip filters replace the dashboard bar

**ID:** leaderboard-chip-filters
**Status:** active
**Chosen:** A KAN-1048 chip bar owned by the view: Period `7d · 30d · 90d` (default `30d`, a rolling window ending at the moment the chip is chosen), Project, Command, and Clear filters when any group is off its default. The model filter is dropped; App.tsx skips `DashboardBar` for `stage-leaderboard` as it does for the state board, and the view ignores App's period and project props. The selected Project chip is the endpoint's `project` parameter (percentiles cannot be recombined client-side); Command filters the returned rows client-side.
**Considered:** Adding Model chips — not this view's question, and Weekly costs already splits cost by model–effort. Period and Project only — leaves the Command dropdown's job undone. An "All" period chip — has no previous window, so the growth column and riser tile would have to hide.

### Project chips list only projects with spend

**ID:** leaderboard-project-chips-with-spend
**Status:** active
**Chosen:** One Project chip per project with costed runs in the window, labelled `projectLabel(key)` with its spend and the full key as the hover title, ordered by spend. The `stage-leaderboard` response gains `projects: [{projectKey, totalCostUsd}]`, computed over the window ignoring the `project` parameter so every chip stays while one is selected.
**Considered:** Every registered project from `/api/v1/projects`, names only — fourteen chips in the dev store, nine of them throwaway or test checkouts that never spend. Projects with spend, names only — the same new data with less on the chip.

### Per-project spend is its own store method

**ID:** leaderboard-project-spend-method
**Status:** active
**Chosen:** A new `Store.StageLeaderboardProjects(ctx, period)` on `StatsStore`, called by the stats handler only for `stage-leaderboard`, whose result fills a `projects` field on `statsResponse` (`omitempty`, so no other view's response changes). `StageLeaderboard`'s signature stays.
**Considered:** Changing `StageLeaderboard` to return rows and projects together — touches the same four test fakes plus every existing store-test caller, and mixes a project-ignoring query into a project-filtered one. Computing project totals client-side from per-project calls — one request per project per window.

### Plain ranked table

**ID:** leaderboard-plain-ranked-table
**Status:** active
**Chosen:** A plain `lb-table` ranked by spend with a `#` rank column, outside `DataTable`, with no sorting and no search box; the top ten rows plus the Other row.
**Considered:** Keeping `DataTable` with sortable columns — the ranking is the question, and ten rows need no search.

### Columns kept, added and dropped

**ID:** leaderboard-columns
**Status:** active
**Chosen:** Columns in the prototype's order: `#`, Stage, Command, Spend (with the `lb-costbar` bar scaled to the top row), Share, "vs previous <period>", Runs, Median run, P90 run, Main share. Mean cost and Main session $ are dropped — Spend ÷ Runs and Spend × Main share respectively. Re-decides archived `leaderboard-main-columns` for the dollar column only. `meanCostUsd` and `mainCostUsd` stay in the response; nothing reads them on this page.
**Considered:** Dropping Mean cost only — keeps a derivable dollar column. Keeping both — two columns that add nothing to what stays. Removing the two fields from the response — a wire change with no reader to benefit.

### P95 and P99 are not added

**ID:** leaderboard-no-p95-p99
**Status:** active
**Chosen:** Median and P90 stay the only percentiles. Measured on the dev store for the top ten rows by spend: P99 sits above fewer than three runs on the largest row of the whole store and above fewer than one on every project-filtered or 7d view, so it reports the one or two outliers the run view already shows; under a project chip or the 7d chip most rows hold 7–40 runs, leaving P95 on the same outliers, so its meaning changes with the chip selection; and on the same rows P95 moves no stage between Median vs P90's two classes ("every run is dear" against "a few runs blow up").
**Considered:** Adding P95 — sound only on the default all-projects 30d view, misleading under the narrowing chips. Adding P95 and P99 — P99 is the maximum in practice.

### Tile layout

**ID:** leaderboard-tile-layout
**Status:** active
**Chosen:** Spend stacked over Biggest riser in a narrower left column, Top 3 spanning both rows in a wider right column (2fr : 3fr) from 760px of viewport width; one column below that, ordered Spend, Top 3, Biggest riser. The two short tiles together are as tall as the list, so no tile stands part-empty.
**Considered:** An equal-width three-column grid (prototype v2) — the list sets the row height and the two figure tiles stand mostly empty, the operator's complaint. Content-sized short tiles beside a wide list — the same vertical emptiness, narrower. A horizontal Top 3 list — leaves Spend and Biggest riser unequal and splits each stage across a narrow card.

### Plain-English stage names, readable project names, app theme

**ID:** leaderboard-readable-names
**Status:** active
**Chosen:** Every stage on the page — Top 3, Biggest riser, table — shows `stageLabel(id)` as its primary label with the id beneath in small monospace; an id `stageLabel` does not map shows the id alone, once. Project chips show `projectLabel(key)`. The page uses the app's own shell, light/dark tokens and `lb-*` rules, adding only `sl-*` rules that consume existing tokens.
**Considered:** The id as primary label — what the operator rejected in round 2. Name in the tiles but id in the table — one stage shown two ways on one page reads as two things. A dark standalone page with a sidebar (prototype v1) — unlike every other page of the app.

### The previous window is a second call

**ID:** leaderboard-previous-window-call
**Status:** active
**Chosen:** The view calls `stage-leaderboard` twice through `useStatsView`: the chosen window, and the same length shifted back one window, joined on command/stage. The window bounds are computed once when a Period chip is chosen, never per render.
**Considered:** A new endpoint returning both windows — server code for a join the view does in a few lines.

### Chip becomes a shared component

**ID:** leaderboard-shared-chip
**Status:** active
**Chosen:** StateBoard's local `Chip` moves to `stats/web/src/components/Chip.tsx`, used by StateBoard and the leaderboard — the second caller.
**Considered:** A second local copy in the leaderboard — the duplicate the second caller exists to remove.

### Leaderboard captures move to their own spec

**ID:** leaderboard-capture-spec
**Status:** active
**Chosen:** A new `stats/web/tests/visual/stage-leaderboard.spec.ts` carries the view's captures (default view at 1280px, the same at 390px, a Project chip hover) with the clock pinned to `FIXTURE_NOW`; the stage-leaderboard block of `views.spec.ts`, its snapshots and the `views.spec.ts.mockups` sidecar (whose every line maps those captures to the retired v2 prototype frames) are removed.
**Considered:** Rewriting the block inside `views.spec.ts` — that file's sort-and-filter tests and its PINNED_QUERY convention belong to the old view, and a separate file keeps the trend tests untouched.

### Dashboard-bar control tests move to Reviewers

**ID:** controls-spec-on-reviewers
**Status:** active
**Chosen:** `controls.spec.ts`'s PeriodPicker, SearchBox, ProjectFilter, FilterBar, DataTable sort, ChangeVariable and "ModelVariable enabled" tests retarget from the leaderboard, which no longer carries those controls, to the Reviewers view, which keeps the dashboard bar and a `DataTable` with a filterable column; their expected rows are read from the live UI-test stack.
**Considered:** Self-review fixes — also qualifies, but Reviewers is enabled for the model variable with real models from the same fixture dispatches. Deleting the tests — the controls still exist on every other view.

## Open questions

## Recommended approach

Server first, then the view, then its Playwright captures, with the control tests retargeted
alongside.

1. **Store and API** (`leaderboard-transport-uncapped`, `leaderboard-project-spend-method`): drop
   the `LIMIT 10`, add `StageLeaderboardProjects` — a `GROUP BY c.project_key` over the same window
   and the same costed-run filter as `StageLeaderboard` with no model and no project restriction —
   and put its rows on the stage-leaderboard response as `projects`.
2. **The view** (`leaderboard-chip-filters`, `leaderboard-project-chips-with-spend`,
   `leaderboard-plain-ranked-table`, `leaderboard-columns`, `leaderboard-tile-layout`,
   `leaderboard-readable-names`, `leaderboard-previous-window-call`, `leaderboard-shared-chip`):
   rewrite `StageLeaderboard.tsx` against the prototype's markup — chip bar, `sl-tiles`, the
   ranked `lb-table` — from two `useStatsView` calls; App.tsx stops mounting `DashboardBar` on it;
   the `sl-*` rules, `.up`/`.down` and `tr.other` copy into `styles.css` scoped under the view.
3. **Captures** (`leaderboard-capture-spec`) and **control tests** (`controls-spec-on-reviewers`),
   file-disjoint, after the view.

Rejected alternatives are recorded under each decision's `**Considered:**`.

## Critical files

- `stats/internal/store/aggregate.go` — `StageLeaderboard` (drops `LIMIT 10`) and the new `StageLeaderboardProjects`.
- `stats/internal/store/aggregate_test.go` — store tests over a seeded test database.
- `stats/internal/api/stats.go` — `StatsStore`, `statsResponse` (`projects`), the handler filling it for `stage-leaderboard`.
- `stats/internal/api/stats_test.go` — the `statsFake` and `fakeStore` fakes and the route tests.
- `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go` — two more `StatsStore` fakes that gain the method.
- `stats/web/src/api.ts` — `StatsResponse` gains `projects?`; a `ProjectSpend` type.
- `stats/web/src/components/Chip.tsx` — new, the shared chip.
- `stats/web/src/views/StateBoard.tsx` — imports the shared chip.
- `stats/web/src/views/StageLeaderboard.tsx` — the view, rewritten.
- `stats/web/src/App.tsx` — skips `DashboardBar` for `stage-leaderboard`.
- `stats/web/src/styles.css` — the view's `sl-*`, `.up`/`.down` and `tr.other` rules.
- `stats/web/src/views/views.test.tsx` — the view's vitest cases.
- `stats/web/tests/visual/stage-leaderboard.spec.ts` and its `-snapshots/` — new captures.
- `stats/web/tests/visual/views.spec.ts`, `views.spec.ts-snapshots/`, `views.spec.ts.mockups` — the old leaderboard block removed.
- `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-stage-leaderboard-darwin.png`, `full-app-suite.zip` — recaptured.
- `stats/web/tests/visual/controls.spec.ts` — control tests retargeted to Reviewers.

## Existing code to reuse

- `stats/web/src/lib/stageLabel.ts` (`stageLabel`) and `stats/web/src/lib/projectLabel.ts` (`projectLabel`) — names, unchanged.
- `stats/web/src/hooks/useStatsView.ts` (`useStatsView`) — both window fetches.
- `stats/web/src/components/Panel.tsx` / `ViewFrame` conventions and `stats/web/src/format.ts` (`formatUsd`, `formatInt`) where their output matches the prototype's formats.
- StateBoard's `Chip` and its `lb-fbar` / `lb-fgroup` / `lb-fl` / `lb-fchip` / `lb-c` / `lb-fclear` markup (`stats/web/src/views/StateBoard.tsx`).
- The `lb-*` rules in `stats/web/src/styles.css`: `lb`, `lb-tile`, `lb-lbl`, `lb-val`, `lb-panel`, `lb-panel-h`, `lb-panel-note`, `lb-scroll`, `lb-table`, `lb-costcell`, `lb-costbar`, `lb-muted`, `lb-sub`.
- App.tsx's existing header condition that skips `DashboardBar` (`hide-dashboard-bar`).
- `StageLeaderboard`'s `costed` CTE filter in `aggregate.go` for the project query's run filter; the store tests' `newTestStore`, `seedChange`, `baseBeginInput`, `costRun` helpers.
- `stats/web/tests/visual/support.ts` (`FIXTURE_NOW`) and `page.clock.setFixedTime`, as `baseline.spec.ts` and `controls.spec.ts`'s state-board block use them.
- `scripts/compose-mockup-frames.sh` and `scripts/measure-visual-properties.sh` for the plan-time capture comparison against the prototype.

## Verification

- `cd stats && make -s web-dist && go test -race -count=1 ./internal/store ./internal/api ./internal/web ./internal/client` passes.
- `cd stats/web && npm test` passes and `npx tsc -b` exits 0.
- `flow.visual-verify` renders the approved prototype as each capture's frame (step 10 of `skills/flow/visual-verify-verifier.md`) and compares the 1280px, 390px and chip-hover captures; the project's `verify` row passes against the committed baselines, the full app suite included.
- By eye on the UI-test stack (`make -C stats ui-test-up`, `http://127.0.0.1:4174/#/stage-leaderboard`): no dashboard bar under the header; Period, Project and Command chips; the three tiles in the 2fr : 3fr arrangement; the ranked table with the Other row when more than ten stages ran; switching to `7d`, a project and a command changes the figures and shows Clear filters.

## Live check

The change touches `flowd`'s `stage-leaderboard` endpoint, so it is exercised against a copy of the
real dev store with `scripts/live-store-copy.sh up kan-1052` (a `flowd` built from this worktree,
default port 4175), then `scripts/live-store-copy.sh down kan-1052`.

1. **Before**, on the protected daemon with a public read-only GET (no change, no credentials):
   `curl -s "http://127.0.0.1:4173/api/v1/stats/stage-leaderboard?from=<now-30d>&to=<now>" | jq '{rows: (.rows|length), projects: (.projects|length)}'`
   — record both figures; the current build answers `rows` of exactly ten and `projects` of `0`
   (`null|length`).
2. **After**, the same request against the copy's port: record `rows` and `projects`. It works when
   `rows` exceeds ten (the dev store ran dozens of costed stages in a month) and `projects` lists at
   least gymie and agents with a positive `totalCostUsd`, ordered by spend, with no zero-spend
   project.
3. Repeat step 2 with `&project=<gymie's full key>`: `rows` narrows, `projects` stays the same list
   and totals as step 2.
4. Open `http://127.0.0.1:4175/#/stage-leaderboard` and record the Spend tile's figure, the Top 3
   names and the Biggest riser; the Spend figure equals the sum of `totalCostUsd` over step 2's rows.

**Not working looks like:** `rows` still exactly ten, `projects` absent or empty, step 3's
`projects` narrowed to the one project, or a Spend tile that equals only the top ten's sum.
