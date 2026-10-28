## Context

KAN-1041 extends the Weekly costs view KAN-1037 built and KAN-1038 reduced to one chart (`stats/web/src/views/WeeklyCosts.tsx`, `stats/internal/store/aggregate.go` `WeeklyCosts`).

**The approved prototype is the spec** (`plan-visual-prototype`): `/Users/tweety53/Projects/agents-worktrees/kan-1041-weekly-dashboard-model-effort-correlation/.superpowers/sdd/weekly-model-effort-preview.html`, published at https://claude.ai/artifact/7TZTUL4b6gJqu1nHzX7i5j (version 2). The implementation and `flow.visual-verify` compare against it.

What the store holds, measured against the dev store on 2026-10-28 over the last 56 days:

- Effort is recorded only per dispatch: `stage_runs.metrics->'dispatches'-><agentId>->'signals'->'served_efforts'` (a map effort → turns) and `dispatches.effort`. The parent session's own buckets (`metrics->'models'-><model>`) carry no effort.
- None of the 4,700-odd dispatch buckets in the window carries `cost_usd`. Their `model` is the alias the dispatch was given (`opus`, `sonnet`, `inherit`, empty), and the `pricing` table holds only full ids (`claude-opus-5-5`, …), so `Store.Price`'s dispatch loop resolves no rate and skips each one. The real id is in `signals.served_models` (a map model id → turns).
<!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "with d as (select e.value v from stage_runs sr, jsonb_each(sr.metrics->'dispatches') e where sr.started_at > now()-interval '56 days') select v ? 'cost_usd', count(*) from d group by 1" @ 2026-10-28 dev store -->

## Decisions

### Price a dispatch by its served model

**ID:** dispatch-served-model
**Status:** active
**Chosen:** `Store.Price` resolves a dispatch's rate from the `signals.served_models` key with the most turns (ties broken by name, for determinism), falling back to the bucket's own `model` when the map is absent or empty — the served id is what was billed.
**Considered:** map aliases to ids in a table (`opus` → `claude-opus-5-5`) — goes stale on every model release and is wrong for `inherit`; price every served model by its own turn share — tokens are not recorded per served model, so the split would be invented.

### A dispatch's effort is its most-served effort

**ID:** dispatch-served-effort
**Status:** active
**Chosen:** the effort of a dispatch line is the `signals.served_efforts` key with the most turns (ties by name), or `default` when the map is absent or empty — `default` is the vocabulary `dispatches.effort` already uses for "none set" (migration 0020).
**Considered:** `dispatches.effort` from the records table — joining a bucket's agent id to a record row has no key today; the served signal is in the same bucket as the cost.

### The parent session is its own line per model

**ID:** session-line
**Status:** active
**Chosen:** for each stage run and model bucket that carries `cost_usd`, the session cost is that bucket's `cost_usd` minus the priced dispatch costs whose served model is that model, floored at 0, charted as `<model> · session`. The chart's lines then add up to Overall wherever every dispatch priced. Operator's answer at planning, 2026-10-28.
**Considered:** dispatches only — the lines would not add up to Overall, and the parent's cost (usually the largest share) would be missing from the correlation.

### Model shown without the `claude-` prefix

**ID:** model-label
**Status:** active
**Chosen:** the line label is the model id with a leading `claude-` removed, then ` · `, then the effort or `session` — `opus-5-5 · low`. Display-only, decided rather than asked.
**Considered:** the full id — longer labels in the toggle row for no information; the alias — two releases of one family would merge into one line.

### A separate view for the model·effort series

**ID:** model-effort-view
**Status:** active
**Chosen:** a new stats view `weekly-model-effort-costs` returning long rows `{week, model, effort, costUsd}`; the SPA pivots them into lines. The phase view's row shape stays as it is, apart from `visualVerifyUsd`.
**Considered:** add the series into `weekly-costs` rows — a variable set of columns per response, which the fixed DTO cannot carry.

### Visual verify is a sixth phase category

**ID:** visual-verify-category
**Status:** active
**Chosen:** `flow.visual-verify` counts as category `visual-verify` in `WeeklyCosts`, a `visualVerifyUsd` field, and a `Visual verify` series after Self-review. It is classified after self-review, like the other phases. Operator's request mid-planning, 2026-10-28.
**Considered:** none — a direct request.

### Every dot carries a permanent cost label

**ID:** permanent-cost-label
**Status:** active
**Chosen:** each dot's pill reads `$<cost> ▲<n>%` (or `▼`, or no arrow at 0%), the first week and the week after a gap `$<cost>` alone, stacked per week column by the existing `stackLabels` (now including week 0), on both charts. The tooltip stays.
**Considered:** drop the tooltip now that the label carries the figures — it still names the series and week, which the pill does not.

### Reprice through the CLI

**ID:** reprice-command
**Status:** active
**Chosen:** `flow pricing reprice -since <YYYY-MM-DD>` posts to `POST /api/v1/pricing/reprice`, which calls `Store.Price` for every stage run started on or after that date and answers `{priced, skipped}`; a run Price refuses (`ErrTokensUnavailable`, `ErrPricingNotFound`) counts as skipped, any other error fails the call. Operator's answer at planning, 2026-10-28.
**Considered:** reprice at flowd start-up — runs on every restart, and does work nobody asked for once the window is filled; new runs only — the chart would stay empty for 8 weeks.

## Open questions

## Recommended approach

Fix the store first (`dispatch-served-model`), then add the two read paths (`visual-verify-category`, `model-effort-view`), then the CLI that backfills (`reprice-command`), then the SPA (`permanent-cost-label`, `model-label`) and its visual spec. Each step is testable on its own against the test Postgres the store tests already use.

## Critical files

- `stats/internal/store/pricing.go` — `Store.Price`'s dispatch loop; gains the served-model resolution and `RepriceSince`.
- `stats/internal/store/aggregate.go` — `WeeklyCosts` (visual-verify category) and the new `WeeklyModelEffortCosts`.
- `stats/internal/store/aggregate_test.go`, `stats/internal/store/stageruns_test.go` — store tests.
- `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go` — the two views' DTOs.
- `stats/internal/api/server.go`, `stats/internal/api/pricing.go`, `stats/internal/api/pricing_test.go` — the reprice route.
- `stats/internal/client/pricing.go`, `stats/internal/client/client_test.go` — client call.
- `stats/cmd/flow/main.go`, `stats/cmd/flow/pricing.go`, `stats/cmd/flow/pricing_test.go` — the CLI command.
- `stats/web/src/api.ts`, `stats/web/src/views/WeeklyCosts.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/styles.css`, `stats/web/src/theme.css` — the SPA.
- `stats/web/tests/visual/weekly-costs.spec.ts` with its snapshots; the SPA fixtures live in `stats/web/src/views/views.test.tsx`.
- `stats/cmd/uitest-seed/seed.go` — the UI-test stack's seed, which needs priced dispatches for the second chart to draw.

## Existing code to reuse

- `Store.Price`, `priceCandidate`, `chargeableTokens.cost` (`stats/internal/store/pricing.go`) — the dispatch pricing stays this path; only the model it resolves changes.
- The `WeeklyCosts` week bucketing `date_trunc('week', started_at - interval '2 days', 'UTC') + interval '2 days'` (`stats/internal/store/aggregate.go`).
- `WeeklyCostChart`, `stackLabels`, `tickStep`, `everyWeek`, `lastEightWeeks`, `change`, `deltaText` (`stats/web/src/views/WeeklyCosts.tsx`) — the second chart is the same component fed different series.
- `useStatsView` (`stats/web/src/hooks/useStatsView.ts`) and the `viewName` switch in `stats/internal/api/stats.go`.
- `runSettings`'s flag and client pattern (`stats/cmd/flow/settings.go`).

## Verification

- `cd stats && make -s web-dist && go vet ./... && gofmt -l . && go test ./internal/store/... ./internal/api/... ./internal/client/... ./cmd/flow/...`
- `cd stats/web && npx tsc -b && npx vitest run`
- The visual spec per `.flow/project.md` `## visual verification`, compared with the prototype.

## Live check

Against a scratch copy of the dev store, never the dev `flow` database itself: `docker exec flow-postgres createdb -U flow -T flow flow_kan1041_live`, start a workspace `flowd` from the change's build against it (`scripts/workspace.sh` ports), then:

1. Before: `select count(*) from stage_runs sr, jsonb_each(sr.metrics->'dispatches') e where e.value ? 'cost_usd' and sr.started_at > now()-interval '56 days'` — expect 0.
2. `flow pricing reprice -since <today − 56 days>` against that flowd — record `priced` and `skipped`.
3. After: the same count — expect most of the ~4,700 dispatch buckets priced (any left are `inherit`/empty with no served model).
4. `GET /api/v1/stats/weekly-model-effort-costs` over the 8 weeks: per week, the sum of all rows' `costUsd` is within $0.01 of that week's `overallUsd` from `weekly-costs` wherever every dispatch priced.

Not working looks like: the count stays 0, `priced` is 0, or the per-week sum falls short of Overall by the dispatch share. Drop the scratch database afterwards: `docker exec flow-postgres dropdb -U flow flow_kan1041_live`.
