# Live verification — kan-1041-weekly-dashboard-model-effort-correlation

Stack: a scratch `flowd` built from the branch HEAD on 127.0.0.1:6803, against `flow_kan1041_live`, a `pg_dump` copy of the dev `flow` database taken 2026-10-28 (the dev stack and its database were not touched). Window: 2026-09-02 to 2026-10-28, the dashboard's 8 weeks.

| Check | Before | After |
|-------|--------|-------|
| Dispatch buckets carrying `cost_usd` (last 56 days) | 0 of 4705 — merge-base code prices no dispatch, since no pricing row matches a dispatch's alias model | 2481 of 4711 |
| `flow pricing reprice -since 2026-09-02` | command absent at the merge base | `priced 4899, skipped 10071` (skipped runs carry no tokens or a model with no rate, e.g. `claude-sonnet-5-5`) |
| Per-week sum of `weekly-model-effort-costs` against `weekly-costs` `overallUsd` | no model·effort view at the merge base | equal in all 9 weeks (2026-09-02 $2253.34 … 2026-10-28 $10.60) |

The failure look the design names — the count staying 0, `priced` 0, or the per-week sum falling short of Overall — did not appear.

The first live run exposed two defects that In-run fixes 1 and 2 fixed: reprice aborted on an unattached plan session (NULL `change_id`); and the model·effort sums exceeded Overall (2026-09-30: $1551.51 against $939.78), because runs priced before an unpriceable model's bucket arrived kept a stale top-level total.
