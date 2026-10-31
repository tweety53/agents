## Context

The Trend over time view (`stats/web/src/views/Trend.tsx`, backed by `TrendOverTime` in
`stats/internal/store/aggregate.go`) shows daily total spend and stage-run counts, which move with
how many changes ran, not with what one costs. This change makes it answer **"Is a finished change
getting cheaper as flow changes?"**, and fixes the cost capture whose figures it plots, because
those figures undercount every command (`proposal.md`, `## Why`).

**The spec is the approved prototype**:
`.superpowers/sdd/trend-prototype.html` in the worktree, published at
https://claude.ai/artifact/DD57gkG9bSht6o2EuD3kLj. It is canonical for markup, copy, colours, the
`tr-*` rules, number formats, chart geometry and the tip's text; the implementation and
`flow.visual-verify` compare against it. It is read-only to every task.

What it draws, top to bottom, inside the unchanged app shell:

1. `h2` "Trend over time", sub "Is a finished change getting cheaper as flow changes?"
2. A KAN-1048 chip bar (`lb-fbar`, the shared `stats/web/src/components/Chip.tsx`, as
   `StageLeaderboard.tsx` builds it): **Period** `8w · 12w · 26w` (default `12w`); **Command** one
   chip per command that finished a change in the window, with its count, exactly one always
   selected, no "All" (default `/flow`); **Project** `All` plus one chip per project with finished
   changes of the selected command, readable label (`projectLabel`) with count, full key as title;
   **Clear filters** when off defaults. Choosing a Command resets Project to All.
3. Three static tiles: **Median per change** (the window's latest half, with its change against the
   first half — e.g. last 6 weeks vs the 6 before in 12w — green down / red up), **Changes
   finished** (count, with "$X spent on them"), **Flow changes landed** (changes finished in the
   `agents` project, every command — independent of the Command and Project chips).
4. Panel "Cost per finished change, by week": median line (solid, `--series-1`), mean line (dashed,
   `--series-4`), flow-changes-landed bars beneath on their own scale (`--marker`). A week with no
   finished change is a gap, never $0. Hovering a week shows a guide line and the tip "Week of Oct
   14 / 34 changes · median $19.20 · mean $21.06 / 50 flow changes landed".
5. Panel "By week" (`lb-table`, newest first, note "weeks start Wednesday 00:00 UTC"): Week (the
   current one marked "so far"), Changes, Median, vs previous week, Mean, Spend, Flow changes.
   Empty: "No finished changes in this period."

**Cost capture.** Harvest attributes a transcript record to a stage run only when its timestamp
falls inside one of its session's windows `[started_at, ended_at)`
(`stats/internal/harvest/attribute.go`, `Attributor.Attribute`, `bestWindow`); any other record is
dropped while its bytes are consumed — `CommitHarvestBatch` advances the offset with an empty delta
(`stats/internal/store/harvest.go`), so the record is never read again. `/flow-fast` marks stages
in bursts after the work, so its windows last milliseconds: on
`kan-1035-flow-self-review-analyse-visual-verify-prompts` about 85 % of the run's spend fell in gaps
(recorded whole-life cost $0.87). Every command loses some — over the last 45 days' finished
changes with transcripts on disk, roughly `/flow` 11 %, `/flow-plan` 13 %, `/flow-interactive`
16 %, `/flow-fast` 41 %. Session binding and the dispatch grain are not the cause; the older
never-bound sessions (39 `/flow-fast`, 19 `/flow`, almost all September) have no transcript left
and are out of scope.

**Scope against the ticket.** KAN-1051's own text asks for the view alone; answer 7 of design
round 1 brought the cost-capture fix into this change rather than a Jira follow-up (decision
`cost-capture-in-scope`). It is one change because the view is only correct once the figures it
plots are.

## Decisions

### The question the view answers

**ID:** trend-question-cost-per-change
**Status:** active
**Chosen:** "Is a finished change getting cheaper as flow changes?" — whole-life cost per finished change, bucketed by the Wednesday week it finished; it is the question the view already claims and does not answer.
**Considered:** "Is total spend going down?" (keep daily spend, restyled) — duplicates Weekly costs. "Is a stage run getting cheaper?" (cost per stage run over time) — moves with stage mix, not with what a change costs.

### Today's figures leave the view

**ID:** trend-drops-daily-views
**Status:** active
**Chosen:** Drop the Days and Total cost tiles, the Daily cost chart, the By-day table, the model filter and the shared `DashboardBar` on this view (`App.tsx` skips it for `trend` as for `state-board` and `stage-leaderboard`) — none answers per-change cost, spend over time stays on Weekly costs, and a change's whole-life cost spans models (Weekly costs splits by model·effort).
**Considered:** Keep the daily panels beneath the new ones — two questions on one view, and the daily figures duplicate Weekly costs.

### One Command chip, always selected

**ID:** trend-command-chip-single
**Status:** active
**Chosen:** One Command chip always selected, default `/flow`, no "All" — `/flow` (~$40 a change) and `/flow-fast` (~$0.30) blended give a median that tracks the command mix, not cost.
**Considered:** An "All" chip as default plus one per command — matches the leaderboard, but the default view would show the misleading blend. No Command filter, `/flow` only — hides the other commands' trends entirely.

### "As flow changes" is the agents project's finished changes

**ID:** trend-flow-changes-agents-project
**Status:** active
**Chosen:** Changes finished in the `agents` project per week (matched by project label `agents`, every command): bars under the line, a table column and a tile — derived from existing rows, and they cover guards, scripts and skills alike.
**Considered:** `prompt_changes` rows (Review dynamics' markers) — already recorded, but cover prompt edits only. No marker — the cost line alone cannot be read against what changed.

### The headline compares the window's halves

**ID:** trend-headline-halves
**Status:** active
**Chosen:** The window's latest half against its first half (in 12w, the last 6 weeks vs the 6 before) — always has data inside the chosen window and needs no second request.
**Considered:** The whole window against the previous equal window (the leaderboard's pattern) — a second request, and empty before the store's first weeks. The latest full week against the week before — one week of about 30 changes is noisy.

### Period chips in whole weeks

**ID:** trend-period-weeks
**Status:** active
**Chosen:** `8w · 12w · 26w`, default `12w` — whole weeks for weekly buckets; a window of N weeks is the N Wednesday weeks ending with the current one, from that first Wednesday 00:00 UTC to now.
**Considered:** `7d · 30d · 90d` as the leaderboard — 7d is one bucket, no trend. `12w · All` — an unbounded window grows without limit.

### The endpoint returns one row per finished change

**ID:** trend-endpoint-per-change-rows
**Status:** active
**Chosen:** The `trend` view key, reworked in place, returns one row per finished change in the window — finished week, command, project key, life cost — and every filter and aggregate (medians of any slice, the half-window comparison, flow changes landed) is computed client-side from one request; about 2,000 rows at 26w.
**Considered:** Server-side weekly aggregates per filter (`project` and `command` parameters, median and mean per week) — the half-window median then needs another aggregate or request, and every chip a round trip.

### The trend endpoint refuses a model or project restriction

**ID:** trend-endpoint-refuses-filters
**Status:** active
**Chosen:** A `model` or `project` parameter on `trend` is refused with 400, as the state board refuses `model` — the rows are every project's finished changes, which Flow changes landed needs, and a model restriction has no meaning for a whole-life cost; refusing beats silently ignoring a parameter the old view sent.
**Considered:** Accept and ignore `project` — a caller passing it would read every project's rows as its own. Apply `project` server-side — Flow changes landed would then silently count only that project.

### Data derived from existing rows

**ID:** trend-derived-no-column
**Status:** active
**Chosen:** No column or migration on `changes`: a finished change is `state = 'FINISHED' AND NOT withdrawn` with finish instant `updated_at` (Flow health's `FlowSummary` definition, `stats/internal/store/health.go`); its life cost is the sum of `stageRunCostExpr` over its stage runs, and a change with no costed run is left out of every figure and count (absence is never $0); its week is the Wednesday 00:00 UTC bucket `WeeklyCosts` uses; its command is the command of its first stage run by `started_at`.
**Considered:** A stored per-change cost or command column — state that derives from rows already kept, and a second copy to drift.

### Cost capture is fixed in this change

**ID:** cost-capture-in-scope
**Status:** active
**Chosen:** Investigate and fix the missing cost here (operator's answer 7, round 1) — the figures the view plots are otherwise wrong for every command.
**Considered:** Out of scope, with a Jira follow-up filed at integrate — the default `/flow` chip would still plot an undercount of about 11 %, `/flow-fast` most of its spend missing.

### Spend between two stage windows goes to the stage that began last

**ID:** gap-spend-to-last-begun-stage
**Status:** active
**Chosen:** `bestWindow` keeps its rule; when no window contains a record, a fallback picks the session's window with the latest `started_at` at or before the record (the higher attempt on a tie), provided the record falls before that window's change finished (`updated_at` once `state = 'FINISHED'`; unbounded while not). `harvest.Window` gains the change's finish instant. Every source the watcher feeds to `Attributor` — Claude transcripts, subagent files, rollout files — goes through it, so every command is fixed at once and no skill's marking changes; `DispatchAttributor` is untouched. Per-stage costs then include the session's time until the next stage begins, as `SweepAbandoned` already treats an abandoned run as ending at the next stage's start.
**Considered:** One synthetic per-change stage run `flow.between-stages` — keeps per-stage figures to their windows, but adds a stage key every stage vocabulary, leaderboard and guard has to learn. Changing `/flow-fast` to keep each window open while its stage works — skill prose only, the model already ignores the begin-work-end shape, the other commands keep their gaps and nothing is back-filled.

### Spend before a session's first mark is not attributed

**ID:** pre-first-mark-unattributed
**Status:** active
**Chosen:** Not attributed — it is whatever the session did before the run (kan-1035's session spent $2.16 answering an unrelated question first).
**Considered:** Attributed to the first stage from the message that invoked the command — needs the invocation located in each harness's transcript and adds the run's own pre-kickoff reading.

### The finish instant comes from a store lookup by change

**ID:** finish-instant-store-lookup
**Status:** superseded by finish-instant-latest-of-state-and-runs
**Chosen:** A store method `FinishedAtByChange(ctx, changeIDs []int64) (map[int64]time.Time, error)` returns `updated_at` for each listed change whose state is `FINISHED`; `storeWindowSource` (`stats/cmd/flowd/main.go`) and its test twin `e2eWindowSource` (`stats/internal/harvest/endtoend_test.go`) call it once per session after `QueryStageRuns` and fill each window's finish instant from its run's `ChangeID` — one extra query per session, as `WindowsForSession` is already consulted once per session.
**Considered:** Widen `QueryStageRuns`' row with the change's state and `updated_at` — every caller of the generic query pays for two fields only attribution reads.

**Superseded because:** the live check's back-fill added nothing to kan-1035 — its `updated_at` precedes the end of its last stage run, the same order other finished changes show on the dev store.
### The finish instant is the later of the state write and the last stage run

**ID:** finish-instant-latest-of-state-and-runs
**Status:** active
**Chosen:** `FinishedAtByChange` returns `GREATEST(changes.updated_at, the latest COALESCE(ended_at, started_at) of the change's stage runs)` for a FINISHED change.
**Considered:** Keep `updated_at` alone — the live check showed it is the kickoff instant for changes whose state was written FINISHED before their last marks (kan-1035: 11:54:19Z against a last run ending 12:14:21Z), so the bound dropped every gap record. Add a `finished_at` column — a migration and a writer for a value the stage runs already carry.

### Already-dropped spend is back-filled once at flowd startup

**ID:** gap-backfill-once-at-startup
**Status:** active
**Chosen:** Before the watcher's first cycle, when its marker is unset, `flowd` reads `[0, stored offset)` of every transcript on disk whose path has a stored offset, keeps only the records `bestWindow` places nowhere and the fallback places somewhere, and adds their deltas through `jsonb_deep_add`; the watcher then resumes from the stored offsets under the new rule, so no record is counted twice. A migration adds the one-row marker table `harvest_backfills` (`name` primary key, `ran_at`), name `stage-gap-cost`. It reaches transcripts still on disk (from about 2026-10-03); earlier changes keep their recorded cost. It runs on the `make -C stats restart` after landing.
**Considered:** No back-fill — the view's recent weeks then mix fixed and unfixed changes for 12 weeks. Re-harvest every transcript from zero — rewrites cost already recorded for sessions whose transcripts have since been deleted, losing it.

### The back-fill commits in one transaction with its marker

**ID:** backfill-one-transaction
**Status:** active
**Chosen:** The back-fill sums its deltas over every transcript in memory and commits them in one transaction that first inserts the `harvest_backfills` row (`ON CONFLICT DO NOTHING`) and applies nothing when the row already existed — exactly once even if `flowd` dies mid-run or two start together; then prices every stage run it touched, as the watcher does after a commit. A failure is logged and `flowd` starts anyway; the marker stays unset and the next startup retries.
**Considered:** One commit per transcript with the marker in the last — a crash between two commits re-adds the earlier transcripts' gap spend on the next startup. A per-transcript marker — a table of rows nothing else needs.

### Weeks not back-filled are drawn as they are

**ID:** unbackfilled-weeks-unmarked
**Status:** active
**Chosen:** No marker — the prototype is unchanged; before about 2026-10-03 costs stay low, mostly on the `/flow-fast` chip, which is not the default.
**Considered:** A note under the chart naming the first fully captured week — a new line the prototype would gain, and one more state to test.

### The UI-test fixture gains trend changes without moving other baselines

**ID:** trend-seed-isolated
**Status:** superseded by trend-seed-recapture
**Chosen:** A `seedTrend` function in `stats/cmd/uitest-seed/seed.go` adds finished changes across several weeks before the fixture clock, two commands and an `agents` project, with instants chosen so no other view's committed baseline changes — finish instants before the state board's and Flow health's capture windows, stage runs at one early instant before every other view's window — proved by the project's `verify` row over every spec but the trend's.
**Considered:** Give the existing `uitest-dynamics` finished changes costed stage runs — no `agents` project, and their weeks fix the shape of the chart. Seed at realistic instants and recapture every baseline that moves — a dozen unrelated captures for visual-verify to re-judge.
**Superseded because:** Task 5's verify run failed 9 captures: the state board draws a chip per registered project, and Experiments counts the seeded finished changes.

### The trend seed's moved baselines are re-recorded

**ID:** trend-seed-recapture
**Status:** active
**Chosen:** Keep `seedTrend` as Task 5 landed it and re-record the 9 baselines it moves — state board, Experiments (2), full-app state board and Experiments, interactive (4) — in an appended task, each capture looked at before it is committed.
**Considered:** Move the seed out of every other view's reach — impossible for the state board, which gives every registered project a chip whatever its instants, so the `agents` project Flow changes landed needs always moves it.

### Trend captures move to their own spec

**ID:** trend-capture-spec
**Status:** active
**Chosen:** A new `stats/web/tests/visual/trend.spec.ts` captures the view at 1280px, at 390px and with a week hovered, clock pinned to `FIXTURE_NOW`, full page; `views.spec.ts`, which holds only the old trend block since KAN-1052 moved the leaderboard out, is deleted with its snapshot.
**Considered:** Rewrite the block inside `views.spec.ts` — a file named for several views holding one, against the leaderboard's own-spec pattern.

### App period tests move to the Reviewers view

**ID:** app-period-tests-on-reviewers
**Status:** active
**Chosen:** The `App.test.tsx` tests that read the URL period through the first `fetchStatsView` call, and the filter-bar test's "a view with the bar" half, point at `#/reviewers`, which keeps `DashboardBar`; the route-resolution tests that only check the heading stay on `#/trend`.
**Considered:** Keep them on `#/trend` — its first request now carries its own chip window, not the URL period, so they would fail for a reason unrelated to what they test.

## Open questions

## Recommended approach

Cost capture first, then the endpoint, the view, the fixture and the captures.

1. **Attribution** (`gap-spend-to-last-begun-stage`, `pre-first-mark-unattributed`,
   `finish-instant-store-lookup`): `harvest.Window` gains `ChangeFinishedAt *time.Time`; a
   `fallbackWindow` beside `bestWindow`; `Attribute` uses it when `bestWindow` misses; the store
   lookup fills the finish instant in both window-source adapters.
2. **Back-fill** (`gap-backfill-once-at-startup`, `backfill-one-transaction`): migration `0059`,
   two store methods, an `Attributor.AttributeGaps` that returns only fallback-placed records'
   deltas, a `Watcher.BackfillStageGaps` that reads each transcript's already-harvested bytes, and
   the startup call in `flowd` before `watcher.Run`.
3. **Endpoint** (`trend-endpoint-per-change-rows`, `trend-endpoint-refuses-filters`,
   `trend-derived-no-column`): `TrendOverTime` becomes `TrendFinishedChanges`, one row per finished
   change; the API maps it to `{week, command, projectKey, costUsd}` and refuses `model` and
   `project`.
4. **View** (`trend-question-cost-per-change`, `trend-drops-daily-views`,
   `trend-command-chip-single`, `trend-flow-changes-agents-project`, `trend-headline-halves`,
   `trend-period-weeks`, `app-period-tests-on-reviewers`): `Trend.tsx` rewritten to the
   prototype's markup and `render()`; the `tr-*` rules copied into `styles.css`; App.tsx skips the
   dashboard bar for `trend`.
5. **Fixture** (`trend-seed-isolated`) and **captures** (`trend-capture-spec`).

Rejected alternatives are recorded under each decision's `**Considered:**`.

## Critical files

- `stats/internal/harvest/attribute.go` — `Window`, `bestWindow`, the new fallback, `Attribute`, `AttributeGaps`.
- `stats/internal/harvest/attribute_test.go` — attribution unit tests.
- `stats/internal/harvest/watcher.go` — `Watcher`, `Source`, `RunOnce`; the new `BackfillStageGaps`.
- `stats/internal/harvest/endtoend_test.go` — `e2eWindowSource`, the burst-marked and back-fill end-to-end tests.
- `stats/cmd/flowd/main.go` — `storeWindowSource`, the watcher wiring and the startup back-fill call.
- `stats/internal/store/changes.go` — `FinishedAtByChange`.
- `stats/internal/store/harvest.go` — `CommitHarvestBatch`; the new `HarvestBackfillDone` and `CommitHarvestBackfill`.
- `stats/internal/store/migrations/0059_harvest_backfills.sql` — the marker table.
- `stats/internal/store/aggregate.go` — `TrendOverTime` replaced by `TrendFinishedChanges`.
- `stats/internal/store/aggregate_test.go`, `stats/internal/store/harvest_test.go`, `stats/internal/store/changes_test.go` — store tests.
- `stats/internal/api/stats.go` — `StatsStore`, the `trend` case, `toTrendDTOs`, the parameter refusal.
- `stats/internal/api/stats_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go` — `StatsStore` fakes and route tests.
- `stats/web/src/views/Trend.tsx` — the view, rewritten.
- `stats/web/src/api.ts` — the `TrendChange` row type replacing `TrendPoint`.
- `stats/web/src/App.tsx` — skips `DashboardBar` for `trend`.
- `stats/web/src/styles.css` — the prototype's `tr-*` rules.
- `stats/web/src/views/views.test.tsx`, `stats/web/src/App.test.tsx` — vitest cases.
- `stats/cmd/uitest-seed/seed.go`, `stats/cmd/uitest-seed/seed_test.go` — the trend fixture.
- `stats/web/tests/visual/trend.spec.ts` and its `-snapshots/` — new captures; `stats/web/tests/visual/views.spec.ts` and `views.spec.ts-snapshots/trend-darwin.png` — deleted; `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-trend-darwin.png`, `stats/web/tests/visual/full-app-suite.zip` — recaptured.
- `stats/README.md` — the attribution rule and the one-shot back-fill, in its Pricing section.

## Existing code to reuse

- `bestWindow`'s ordering (latest `StartedAt`, higher `Attempt` on a tie) and `intervalContains` in `stats/internal/harvest/attribute.go` for the fallback's tie-break.
- `CommitHarvestBatch`'s `jsonb_deep_add` merge (`stats/internal/store/harvest.go`, `stats/internal/store/migrations/0005_jsonb_deep_add.sql`) and `encodePatches` (`stats/internal/harvest/watcher.go`) for the back-fill's patches; the watcher's post-commit `Price` call for pricing the touched runs.
- `ParseAssistantRecords` / `ParseSignalRecords` (`stats/internal/harvest/transcript.go`) and `ParseRolloutRecords` (`stats/internal/harvest/rollout.go`) — the byte-level parsers each source's `ReadNew` already runs.
- `FlowSummary`'s `costed` / `firsts` / `finished` CTEs (`stats/internal/store/health.go`) and `stageRunCostExpr` (`stats/internal/store/query.go`) for the per-change query; `WeeklyCosts`' Wednesday bucket expression (`stats/internal/store/aggregate.go`).
- The state board's `model` refusal in `stats/internal/api/stats.go`'s `view()`.
- `StageLeaderboard.tsx`'s chip bar (`Group`, `Chip`, `lb-fclear`), its window-on-chip-choice pattern (`windowFor`) and `Panel` absence handling; `projectLabel` (`stats/web/src/lib/projectLabel.ts`); `useStatsView` (`stats/web/src/hooks/useStatsView.ts`).
- The `lb-*` rules in `stats/web/src/styles.css`.
- `storetest.NewStore` (`stats/internal/storetest/storetest.go`) and the end-to-end helpers `newEndToEndStore`, `writeLines`, `marshalLine` in `stats/internal/harvest/endtoend_test.go`.
- `seedReviewDynamics`' shape in `stats/cmd/uitest-seed/seed.go` (`PutChange`, `BeginStage`, `MergeMetrics`, `EndStage`).
- `stats/web/tests/visual/stage-leaderboard.spec.ts` and `support.ts` (`FIXTURE_NOW`, `page.clock.setFixedTime`) for the capture spec; `scripts/compose-mockup-frames.sh` and `scripts/measure-visual-properties.sh` for the capture comparison against the prototype.

## Verification

- `cd stats && make -s web-dist && go test -race -count=1 ./internal/harvest ./internal/store ./internal/api ./internal/web ./internal/client ./cmd/...` passes.
- `cd stats/web && npm test` passes and `npx tsc -b` exits 0.
- `flow.visual-verify` renders the approved prototype as each capture's frame (step 10 of `skills/flow/visual-verify-verifier.md`) and compares the 1280px, 390px and week-hover captures; the project's `verify` row passes against the committed baselines, the full app suite included, and no other view's baseline changed.
- By eye on the UI-test stack (`make -C stats ui-test-up`, `http://127.0.0.1:4174/#/trend`): no dashboard bar; Period, Command and Project chips with `12w` and `/flow` pressed; the three tiles; the median and mean lines with a gap at an empty week and bars beneath; hovering a week shows the guide and the tip; choosing `/flow-fast` resets Project to All and shows Clear filters.

## Live check

The change touches `flowd`'s harvest attribution, its startup and the stored metrics of real stage
runs, so it is exercised against a copy of the real dev store with
`FLOW_LIVE_COPY_TRANSCRIPTS="$HOME/.claude/projects" scripts/live-store-copy.sh up kan-1051` (a `flowd` built from this worktree, default port 4175,
whose startup runs the back-fill over the copy), then `scripts/live-store-copy.sh down kan-1051`.

1. **Before**, on the protected daemon with public read-only GETs (no change, no credentials):
   record the whole-life cost of `kan-1035-flow-self-review-analyse-visual-verify-prompts` —
   the sum of `metrics.cost_usd` over
   `curl -s 'http://127.0.0.1:4173/api/v1/stage-runs?name=kan-1035-flow-self-review-analyse-visual-verify-prompts&limit=200'`
   (recorded $0.87 at design time) — and the same sum for one `/flow` change finished after
   2026-10-03 (name it in the record). Record `curl -s -o /dev/null -w '%{http_code}' 'http://127.0.0.1:4173/api/v1/stats/trend?from=<12 weeks ago>&to=<now>'`.
2. **After**, the same requests against port 4175 once the copy's `flowd` logs the back-fill done:
   kan-1035's sum rises to within 10 % of its in-span spend (about $10 at list prices: main
   session $3.87 + subagents $4.46 in gaps plus $0.99 + $0.69 in windows), and the `/flow` change's
   sum rises by a few to about fifteen per cent; the copy's `harvest_backfills` holds one
   `stage-gap-cost` row. If kan-1035's transcript is no longer on disk, pick the `/flow-fast`
   change finished latest after 2026-10-03 and record its before/after instead.
3. **Idempotence**: restart only the copy's `flowd` (the process `live-store-copy.sh up` started on
   4175, never 4173) and repeat step 2's sums — unchanged.
4. **Endpoint**: `GET /api/v1/stats/trend?from=<12 weeks ago>&to=<now>` on 4175 returns rows with
   `week` a Wednesday, `command`, `projectKey`, `costUsd`; adding `&model=x` or `&project=x`
   answers 400. Open `http://127.0.0.1:4175/#/trend` and record the Median per change tile, the
   Changes finished count and Flow changes landed.

**Not working looks like:** step 2's sums equal step 1's (the back-fill did nothing or the
fallback never fires), step 3's sums grow again (the back-fill is not once-only), a `trend` row
with a cost for a withdrawn or unfinished change, or a view whose `/flow` median equals its
`/flow-fast` one.
