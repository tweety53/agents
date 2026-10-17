# kan-989-stats-ui-dashboard-review

## Why

The operator reviewed every stats UI dashboard one by one (KAN-989) and found noise, dead weight and a slow view: finished pipeline-fix changes crowd the live board, two dashboards (Cache efficiency, Runs) are not worth a nav slot, tables carry columns nobody reads, Flow health shows tables where totals would answer the question, and Reviewers loads ~15× slower than the other views.

## What changes

- **Live state board** — hides `pipeline-fix-*` changes; drops the `By` and `Next command` columns; gains per-change whole-life `Runs`, `Cost`, `Wall clock`, `Errors` (the Runs view merged in).
- **Stage leaderboard** — new `Total cost` column; only the top 10 stages by total cost.
- **Trend over time** — unchanged.
- **Cache efficiency** — removed.
- **Reviewers** — loads fast; drops `Findings/dispatch` and `Withdrawn`; gains a sortable `Efficiency` score.
- **Decisions** — drops the `Runs` tile and the `Every decision` table; adds a `By implementer model` summary.
- **Runs** — removed (folded into the live board; run detail stays).
- **Flow health** — every table replaced by eight period totals tiles.
- **Self-review fixes** — drops the `Per change` block.
- No stored data, table, column or migration is deleted.

## Fix round 1

- Dollar amounts render at two decimals everywhere; a live-board change with no stage runs shows `—` for cost and wall clock.
- The UI-test seed covers decisions, findings, errored runs and guard fires so every reworked view is verified populated.
