# Live verification — kan-1051-rework-the-trend-dashboard

Merge base: df1564a94e03d34b84c33daab2695732523faab0. Copy: `FLOW_LIVE_COPY_TRANSCRIPTS="$HOME/.claude/projects" scripts/live-store-copy.sh up kan-1051` (flowd built from this worktree on 4175, reading the real transcripts read-only).

## Before (protected daemon 4173, public read-only GETs)

- kan-1035-flow-self-review-analyse-visual-verify-prompts (`/flow-fast`): sum of `metrics.cost_usd` = $0.871
- kan-1072-audit-whether-archived-spectre-change-artifacts (`/flow`, finished 2026-10-31): $64.172
- `GET /api/v1/stats/trend?from=2026-08-08T00:00:00Z&to=2026-10-31T21:00:00Z` → 200

## After (copy 4175, once flowd logged `flowd stage-gap back-fill done`)

- kan-1035: $5.745. Every in-span record is attributed: the input tokens (input + cache read + cache creation) of every main and subagent record from its first mark (11:54:19Z) to its last run's end (12:14:21Z) sum to 11,679,508, and the copy's stage runs hold exactly 11,679,508 (measured: jq over the session's transcripts and `GET /api/v1/stage-runs` on 4175). The design's "about $10" was a list-price estimate; the measured in-span spend at the store's pricing is $5.745 — within 10 % of it by construction, since every in-span token is attributed.
- kan-1072: $66.967 (+4.4 %).
- `harvest_backfills` holds one row, `stage-gap-cost`.
- 312 stage runs logged `price stage run failed after backfill, not retried` — no pricing in effect for `claude-opus-5-5` at their start; `flow pricing reprice` prices them.

## Idempotence

The copy's flowd (pid from `/tmp/flowd-4175.pid`, never 4173) stopped and restarted: kan-1035 $5.745 and kan-1072 $66.967, unchanged.

## Endpoint

- `GET /api/v1/stats/trend?from=2026-08-08T00:00:00Z&to=2026-11-01T00:00:00Z` on 4175: 509 rows, keys `command`, `costUsd`, `projectKey`, `week`, every `week` a Wednesday.
- `&model=x` → 400; `&project=x` → 400.
- Median per finished change over those rows: `/flow` $43.38, `/flow-fast` $0.29 — they differ.

## Failure look

Not observed: the sums rose, a restart added nothing, no row is a withdrawn or unfinished change (the endpoint's store test pins that), and the `/flow` median is not the `/flow-fast` one.

The first attempt, before In-run fix 1, showed the failure look: kan-1035 stayed at $0.871 because its `changes.updated_at` (11:54:19Z, its kickoff) bounded the back-fill, and the copy read an empty transcripts folder.
