# kan-450-record-dispatch-end-timestamps-duration-and

## Why

A dispatch's end instant and duration are already stored (`dispatches.ended_at`, required on
`flow record dispatch end`) but the rendered SDD ledger prints only `- Started:` — a reader
cannot see how long a dispatch ran without querying the store by hand. Cost per dispatch is the
same story one layer over: dollars exist only in the stage-run metrics bag
(`dispatches.<agentId>.cost_usd`, written by the existing `Price` path), so surfacing an outlier
like KAN-423's task-18 dispatch (145M cache-read tokens) meant manual summing across ledger
sections. KAN-450 (self-review finding from KAN-423, flow-stats-app angle).

## What changes

- The SDD ledger's per-dispatch sections gain `- Ended:` and `- Duration:` lines, rendered from
  the stored `ended_at`.
- Each dispatch row carries its derived cost: `readDispatches` LEFT JOINs `stage_runs` and lifts
  the nested `dispatches.<agentId>.cost_usd` onto the `records.Dispatch` wire shape as an
  optional field — absent stays absent, never zero.
- The ledger gains a closing `## Cost by role and task` section: per (role, task) totals for
  dispatch count, token sums and dollar sums, plus honest counts of unmeasured and unpriced
  dispatches — the cost-per-change figure, computed from the ledger rows themselves.

No API route, CLI command, SPA view or pricing-path change: the ledger is the surface, and every
dollar still comes from the one existing pricing path.
