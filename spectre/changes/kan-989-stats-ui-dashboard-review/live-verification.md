# kan-989-stats-ui-dashboard-review — live verification (2026-10-17)

The dev daemon on 127.0.0.1:4173 is protected (never rebuilt or restarted by a run), so it still serves the base build; the after figures come from this worktree's UI-test stack on 127.0.0.1:4174 and from the Reviewers SQL run directly against the dev store.

- **Reviewers, before** — dev daemon (base build), all-time period: 1.75 s, 0.75 s; base SQL on the dev store: 325 ms execution.
- **Reviewers, after** — this branch's SQL on the dev store: 7.7 ms execution, rows identical to the base query (12 rows, byte-equal); UI-test stack endpoint: 0.04 / 0.013 / 0.024 s. Under the 0.1 s bar — matched.
- **State board** — rows carry `runCount`/`costUsd`/`wallClockMs`/`errors` (kan-201: 2 runs, $2.778, 600000 ms) beside `updatedBy`; the SPA hides `pipeline-fix-*` (vitest `state board hides pipeline-fix changes`).
- **Removed views** — `runs`, `cache-efficiency`, `guards` answer 400 unrecognised view; `#/runs` and `#/cache-efficiency` fall back to the default route.
- **Flow health** — `flow-summary` answers 200; the full-app capture shows eight tiles.
