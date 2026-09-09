## Context

KAN-450 asks for two things the store already half-has: dispatch end instants (stored, never
rendered) and a cost-per-change figure (per-dispatch dollars exist only inside the stage-run
metrics bag, joinable but never joined). One change because both land in the same artifact — the
rendered SDD ledger — through the same read path (`RunRecord`/`readDispatches`). Constraints:

- `internal/records` imports only `encoding/json` and `time`; the aggregation must stay pure Go
  over `records.Run`, no store dependency.
- Dollars are written by exactly one path — `store.Price`'s nested `dispatches.<agentId>.cost_usd`
  (KAN-201's requirement that per-dispatch cost derive through the same pricing path as every
  other figure). This change reads that figure; it never computes one.
- Absence is not a value, everywhere: no `tokens` key → unmeasured; no joinable priced bucket →
  unpriced. Neither renders or sums as zero.

## Decisions

### The ledger is the only surface

**ID:** ledger-surface
**Status:** active
**Chosen:** render Ended/Duration and a cost-by-role-and-task totals section into the SDD ledger — no new API route, CLI command or SPA view.
**Considered:** a store aggregate + `GET /records/.../cost` + `flow record cost` CLI (operator's option B) — four layers of machinery that surface nothing unless someone runs the command; both A and B (option C) — over-build for a figure the archived artifact can carry itself.

### Per-dispatch cost arrives by join, not by a second pricing pass

**ID:** join-derived-cost
**Status:** active
**Chosen:** `readDispatches` LEFT JOINs `stage_runs` on `id = stage_run_id` and lifts `metrics->'dispatches'-><agent_id>->>'cost_usd'` into a new optional `CostUSD *float64` on `records.Dispatch`.
**Considered:** extending `Price` (or adding a sibling) to price dispatch rows directly — a second writer of the same derived figure, with its own model-resolution semantics (the row's recorded model may be `unknown (agent-defined)`) and a double-pricing drift risk; rejected without a need that justifies it.

### Totals are computed in the records package

**ID:** totals-in-records
**Status:** active
**Chosen:** a pure function over `records.Run` groups rows by (role, task), sums token buckets and `CostUSD`, and counts unmeasured/unpriced rows; `RenderLedger` renders it.
**Considered:** a SQL aggregate beside `CostPerChange` — duplicates the grouping the renderer needs anyway and puts ledger-shaped output in the store layer.

## Open questions

*(none — the design gate closed with nothing deferred)*
