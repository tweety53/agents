## Context

KAN-1046 adds a trend view of review-panel, implementer and visual-verify outcomes to the flow stats SPA (`stats/web`), backed by one new aggregate in `stats/internal/store`.

**The approved prototype is the spec** (`plan-visual-prototype`): `/Users/tweety53/Projects/agents-worktrees/kan-1046-review-dynamics-dashboard/.superpowers/sdd/review-dynamics-prototype.html`, published at https://claude.ai/artifact/QCFtCagfSawyAdVUepj73C (version 2, approved 2026-10-28). The implementation and `flow.visual-verify` compare against it, and the capture spec drives its hover state (chart crosshair tooltip, heatmap cell).

What the dev store holds, measured on 2026-10-28:

- `findings` has 3364 rows and no timestamp; 213 have no `dispatch_id`.
<!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select count(*), count(*) filter (where dispatch_id is null) from findings" @ 2026-10-28 dev store -->
- `findings.severity` is free text in 27 spellings (`Minor`, `minor`, `Important`, `major`, `Blocking`, `High`, `closed`, …).
<!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select count(distinct severity) from findings" @ 2026-10-28 dev store -->
- Slot, model and effort are on `dispatches` (`slot`, `model`, `effort`), reached from a finding through `dispatch_id`; a bundle's slot is `+`-joined. Fix rounds are `findings.round` (0 = first pass).
- Verifier dispatches of visual verify are keyed `visual-verify…`, a re-run `visual-verify-fix-<n>…`: 88 changes, 151 fix re-runs.
<!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select count(distinct change_id), count(*) filter (where dispatch_key ~ '^visual-verify-fix-[0-9]+') from dispatches where role='verifier' and dispatch_key like 'visual-verify%'" @ 2026-10-28 dev store -->
- Nothing records prompt tunings: no migration names them and `docs/prompt-tuning/LEDGER.md` holds no entry. The tunable prompts' git history since 2026-08-15 is: primary 19 commits, principles 22, failure-modes 8, test-sanity 2, call-efficiency 2, engineering-principles 8, sdd-dispatch 5, visual-verify-verifier 45, review-panel-fix-round 37.
<!-- measured: git log --since=2026-08-15 --oneline -- <file> | wc -l per file @ c4202f60 -->

## Decisions

### Cohort by change, rate by dispatch

**ID:** cohort-by-change
**Status:** active
**Chosen:** per-change series (severity per change, fix rounds, changes reviewed) put each change in the bucket of its first reviewer dispatch's `started_at`, and count every finding of that change whatever its `dispatch_id`. Per-dispatch series (slot, model, effort, heatmap) bucket by the reviewer dispatch's own `started_at` and count the findings joined to it by `dispatch_id` and slot, as `Store.Reviewers` does. Every rate is normalised by changes or dispatches, never a raw count, so volume does not read as quality.
**Considered:** add `findings.recorded_at` and date each finding — new data for no gain, since a change's findings belong to one review cohort and the 213 dispatch-less findings already join through the change; raw counts per bucket — a busy week would always look worse.

### Fold severities in SQL

**ID:** canonical-severity
**Status:** active
**Chosen:** a SQL function `canonical_severity(text)` (migration 0049), as `canonical_slot` is: `critical`, `blocking`, `severe` → `critical`; `important`, `major`, `high`, `medium`, `moderate` → `important`; every other spelling → `minor`. Case-insensitive. Only the new aggregate uses it; stored values stay as written.
**Considered:** rewrite stored values — loses what the reviewer wrote; fold in Go — a second place the rule would live beside the SQL that groups.

### One prompt_changes table for tunings and edits

**ID:** prompt-changes-table
**Status:** active
**Chosen:** `prompt_changes(id, sha text UNIQUE, at timestamptz, kind text CHECK (kind IN ('tuning','edit')), area text CHECK (area IN ('review','impl','vv','fix')), slot text NULL, target text, summary text, case_id text NULL)`. A tuning row upserts over an edit row of the same sha (a tuning is the more specific fact).
**Considered:** separate tables for tunings and edits — the charts read both as markers, so two tables only add a UNION; parse `LEDGER.md` — loses the slot, and it is empty.

### Area from the prompt file

**ID:** prompt-area-map
**Status:** active
**Chosen:** one Go map in `stats/internal/prompttuning` from target path to `{area, slot}`: `skills/flow/<slot>-reviewer-prompt.md` for primary, principles, failure-modes, test-sanity, call-efficiency → `review` + that slot; `skills/flow/engineering-principles.md`, `skills/flow/sdd-dispatch.md` → `impl`; `skills/flow/visual-verify-verifier.md` → `vv`; `skills/flow/review-panel-fix-round.md` → `fix`. `import` reads exactly these files; `record` takes the area from the map, falling back to the case's `type` (`implementer` → `impl`, `reviewer` → `review`) and its `source.slot` for a target off the map. Planner prompts are left out: nothing measures plan quality yet.
**Considered:** include `implement.md` and `review-panel.md` — 226 commits of mostly pipeline mechanics would bury the prompt changes; operator-entered area per row — every import would need a hand label.

### Markers by area

**ID:** markers-by-area
**Status:** active
**Chosen:** each chart shows the markers of its own area — severity, slot, model and effort charts `review`; fix rounds `fix`; implementer charts `impl`; visual-verify charts `vv`; changes reviewed none. A marker is a dashed vertical line with a dot; its text appears in the crosshair tooltip.
**Considered:** every marker on every chart — with ~150 edits the charts drown; labels always visible — overlap at this density.

### Git import as the backfill

**ID:** git-import-backfill
**Status:** active
**Chosen:** `flow prompt-tuning import -C <repo> [-since YYYY-MM-DD]` runs `git log --format=%H%x09%cI%x09%s -- <file>` per mapped file and posts one `edit` row per (sha, file); a commit touching two mapped files yields one row per file, keyed `<sha>:<target>`, so `sha` stores that key. Idempotent: a re-run upserts. It runs once on the dev store after landing, and any time later.
**Considered:** a migration that reads git — migrations run inside flowd with no repository; key by sha alone — a commit editing two prompts would lose one area.

### Implementer and visual-verify metrics

**ID:** impl-vv-metrics
**Status:** active
**Chosen:** implementer — per change (cohort: first implementer dispatch), panel findings of the change ÷ its implementer dispatches, series keyed by the change's most frequent implementer `model · effort` (ties by name); blocked share — implementer dispatches started in the bucket whose `outcome` is `blocked`, `aborted`, `failed`, `stopped` or `cancelled`, ÷ all. Visual verify — per change (cohort: first verifier dispatch keyed `visual-verify%`), rounds = the highest `<n>` of a `visual-verify-fix-<n>` key, 0 when none; passed first time = share of changes with 0 rounds.
**Considered:** implementer findings per task id — `findings.task_id` is mostly empty; visual-verify rounds from `stage_runs` — a re-run is one stage run.

### Bucket size from the range

**ID:** bucket-from-range
**Status:** active
**Chosen:** the API buckets by day when `to − from ≤ 14 days`, else by ISO week (Monday, UTC); the response states `bucket`. Not a query parameter.
**Considered:** a day/week toggle — not asked for; always weekly — "Last week" would draw one point.

### The view owns its range bar

**ID:** view-owned-range
**Status:** active
**Chosen:** on `review-dynamics` the global `DashboardBar` hides its `PeriodPicker` and the view renders the prototype's row: Last week, Last 2 weeks, Last 4 weeks, Last 8 weeks (the view's default when the URL names no range), Last 12 weeks, Since last prompt change, All time. Each sets the period through `onPeriodChange`, so the URL keeps `from`/`to`. "Since last prompt change" uses `lastPromptChangeAt` from the response (period-independent); it is disabled when that is null. Project filter stays.
**Considered:** change the global presets — every other view's snapshot changes for an ask about this view; a `presets` prop on `PeriodPicker` — "Since last prompt change" needs this view's data, so the view must render it anyway.

### Previous range of the same length

**ID:** previous-range
**Status:** active
**Chosen:** the API also computes the summary and heatmap over `[from − (to − from), from)`; tiles and heatmap arrows compare against it. A change within ±3 % reads `≈ unchanged`; "no earlier range" when the previous summary has no changes. Lower is better on every tile except "Visual verify passed first time" (higher is better) and "Changes reviewed" (neutral, no colour).
**Considered:** first half vs second half of the range — halves a short range's data.

### A new chart component

**ID:** dynamics-chart-component
**Status:** active
**Chosen:** `stats/web/src/components/DynamicsChart.tsx`: one shared y-scale per chart with three ticks, date labels, markers, a hover crosshair and tooltip, as the prototype draws. `TimeSeriesPanel` stays as it is.
**Considered:** extend `TimeSeriesPanel` — it scales each series separately, spaces points by index without labels, and Trend's snapshots would change.

### Model family and model version, both

**ID:** model-version-split
**Status:** active
**Chosen:** (operator, 2026-10-28, mid-implementation) two versions of one model family are never merged without saying so. Every per-model series — `per-dispatch-model`, `impl-findings-per-task` — and the heatmap exist twice: keyed by family (`opus`, `sonnet`, `haiku`, `fable`, `glm`; versions merged) and keyed by version (`claude-opus-5-5`, …). The view's **Family / Version** toggle picks one; Family is the default. Effort pairs with either: `model · effort` and `model-version · effort`. A dispatch's version is `metrics->>'modelVersion'`, else `dispatches.model` when it is a full ID (`claude-…`, `glm-…`), else `<family> (version unrecorded)`. Family is a SQL function `model_family(text)` folding both spellings. The harvester writes `modelVersion` into a dispatch's metrics delta from the newest attributed assistant message's `message.model`; `flowd backfill-models` (a direct-database repair subcommand, as `flowd restamp-dispatch` is) stamps every dispatch with an `agent_id` and no `modelVersion` from its `agent-<id>.jsonl`, idempotently.
**Considered:** version only, no toggle — operator chose both views; record from now on without backfill — operator chose the backfill (AskUserQuestion, 2026-10-28); a new `dispatches.model_version` column — `metrics` already takes harvested facts through `CommitHarvestBatch`, so no migration or second write path.
**Measured:** <!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select model, count(*) from dispatches group by 1" @ 2026-10-28 dev store --> `opus` 3815, `sonnet` 3525, `claude-opus-5-5` 327, `claude-sonnet-5-5` 20; no `metrics` key carries a model.

## Open questions

## Recommended approach

One aggregate, `Store.ReviewDynamics(ctx, period, project)`, returns series rows `{bucket, metric, series, value, basis}`, the heatmap rows for the range and the previous range, the summary for both, the markers in range and `lastPromptChangeAt`; the API serves it as view `review-dynamics` and the SPA draws it with `DynamicsChart` and a heatmap table. Prompt changes enter through `flow prompt-tuning record`/`import` over a new POST endpoint. Rejected alternatives: per-finding timestamps (`cohort-by-change`), stored severity rewrite (`canonical-severity`), two tables (`prompt-changes-table`), global presets (`view-owned-range`), extending `TimeSeriesPanel` (`dynamics-chart-component`).

## Critical files

- `stats/internal/store/migrations/0049_review_dynamics.sql` — `canonical_severity`, `prompt_changes`.
- `stats/internal/store/promptchanges.go` (+ test) — upsert and list prompt changes.
- `stats/internal/store/reviewdynamics.go` (+ test) — the aggregate.
- `stats/internal/api/stats.go`, `stats/internal/api/promptchanges.go`, `stats/internal/api/server.go` (+ tests) — view and POST endpoint.
- `stats/internal/client/promptchanges.go` (+ test) — client call.
- `stats/internal/prompttuning/prompttuning.go` (+ test) — area map.
- `stats/cmd/flow/prompttuning.go` (+ test) — `record`, `import`.
- `skills/flow-prompt-tuning/SKILL.md` — step 6 records each applied case.
- `stats/web/src/api.ts`, `App.tsx`, `components/DashboardBar.tsx`, `components/DynamicsChart.tsx`, `views/ReviewDynamics.tsx`, `views/views.test.tsx`, `styles.css`.
- `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/review-dynamics.spec.ts` (+ snapshots), `full-app-suite.spec.ts`.

## Existing code to reuse

- `canonical_slot` (`0028_canonical_slots.sql`) — the pattern for `canonical_severity`; applied to slots in the new query.
- `Store.Reviewers` (`stats/internal/store/aggregate.go`) — bundle unnesting and the dispatch-id + slot join.
- `Store.WeeklyCosts` (`aggregate.go`) — week bucketing in SQL.
- `statsHandler.rowsFor` and the view allowlist (`stats/internal/api/stats.go`) — registering a view.
- `useStatsView`, `Panel`, `ViewFrame`, `Unavailable` (`stats/web/src`) — data, loading and empty states.
- `prompttuning.Load` (`stats/internal/prompttuning/prompttuning.go`) — reads `case.json` for `record`.
- `tests/visual/support.ts` `PINNED_QUERY`, `bodyBackground` — capture pattern.

## Verification

- `cd stats && go test ./internal/store/ ./internal/api/ ./internal/client/ ./internal/prompttuning/ ./cmd/flow/` and `cd stats/web && npx vitest run`.
- `make ui-test-up`, open `http://127.0.0.1:4174/#/review-dynamics`, click each preset, hover a chart and a heatmap cell, compare with the prototype; `npm run test:visual`.
- The live check below.

## Live check

Against a scratch copy of the dev store, never the dev `flow` database itself: `docker exec flow-postgres createdb -U flow -T flow flow_kan1046_live`, start a workspace `flowd` from the change's build against it (`scripts/workspace.sh` ports), then:

1. Before: `select count(*) from prompt_changes` fails (no table) on the copy before the new flowd migrates it; after start it is 0.
2. `flow prompt-tuning import -C /Users/tweety53/Projects/agents-worktrees/kan-1046-review-dynamics-dashboard -since 2026-08-15` against that flowd — expect about 148 rows, one per (commit, mapped file), matching the per-file counts in Context.
3. `GET /api/v1/stats/review-dynamics?from=<today−56d>&to=<now>`: `bucket` is `week`; series `findings-per-change` carries `critical`/`important`/`minor` for each week with reviewed changes; `slot` series include `primary` and `principles`; `markers` non-empty; `lastPromptChangeAt` set. With `from=<today−14d>` the `bucket` is `day`.
4. Cross-check one week: findings per change for `minor` equals `count(findings with canonical_severity = minor of changes whose first reviewer dispatch falls in the week) ÷ count(those changes)` run by hand in psql.

Not working looks like: import posts 0 rows, `markers` empty, every series empty, or the hand count disagreeing. Drop the scratch database afterwards: `docker exec flow-postgres dropdb -U flow flow_kan1046_live`. After landing, `flow prompt-tuning import -C /Users/tweety53/Projects/agents -since 2026-08-15` runs once against the dev flowd so the live dashboard carries the markers.
