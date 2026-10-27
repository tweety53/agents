# kan-1037-stats-weekly-costs-dashboard — session narrative

## 2026-10-27 — creating run

- Mid-run the operator stopped picking "(Recommended)" options one by one after the asks cost about 380k uncached tokens; that became pipeline-fix loop 34, which states the non-stop run shape past the plan gate and that a question carrying a recommended option is never asked there. Every later ask in this run was auto-taken.
- Operator answers taken during visual verification: the weekly table always lists 8 rows, a single chart point is centred, and the UI-test seed is extended. These became in-run fix task 3.
- Task 3's first seed extension put the planning, implementation and self-review stages into the pinned query's week. That changed the pinned captures' look and left weekly-costs.spec.ts asserting the old two-point chart, which panel round 3 caught. Abandoned: re-counting the pinned spec to five points, because every other pinned capture would shift too. Kept: a separate change in the week of 2026-07-22, outside the pinned window, asserted by its own table row.
- The verify test harness hung once on TestProveReproducerSignals. Its SIGTERM case opens the release FIFO O_WRONLY after the reader has gone. The re-run passed, and the root cause is left for a follow-up pipeline-fix.
- The re-verify's tsc failed because stats/web/node_modules had been emptied during the panel's throwaway-copy mutation work. `npm ci` restored it, and lint went green.
