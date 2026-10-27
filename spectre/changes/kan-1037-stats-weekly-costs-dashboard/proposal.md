# kan-1037-stats-weekly-costs-dashboard

## Why

The operator wants to see where flow spend goes week by week. Today the only cost-over-time view, Trend, gives a daily total with no split by pipeline phase.

## What changes

- A new stats dashboard, **Weekly costs**, with weeks running Wednesday 00:00 UTC to the next Wednesday.
- A table of the last 8 weeks, ending with the partial current week. It ignores the period picker and always shows five columns: overall, planning, implementation, review, and self-review (plan + fixes).
- A line chart of the same five totals in weekly buckets, driven by the dashboard's existing from–to picker.
