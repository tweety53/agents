# Live verification

Measured 2026-10-27 against the dev store (read-only).

- Before (merge base 0e97087973ac339a33f17b2389362ec9e6c3bf1c, the dev `flowd` on 127.0.0.1:4173 built from it): `GET /api/v1/stats/weekly-costs?from=2026-09-02T00:00:00Z&to=2026-10-27T03:30:00Z` → HTTP 400 `unrecognised view "weekly-costs"`.
- After (this worktree's `flowd`, `FLOWD_PORT=4199`): the same request → 8 rows, weeks 2026-09-02 … 2026-10-21, each a Wednesday; every category column non-null.
- Week 2026-10-14: `overallUsd` 889.3234428 = psql `sum((metrics->>'cost_usd')::numeric)` over `started_at` in [2026-10-14, 2026-10-21) → 889.323442800000022093. Matched.
- Failure look (a Monday-keyed week, an overall mismatch, or all categories null with overall set): not seen.
