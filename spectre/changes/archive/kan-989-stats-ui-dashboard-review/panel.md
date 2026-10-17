# Review panel — kan-989-stats-ui-dashboard-review

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/store/health.go:113 | Mean cost per finished change averages in every finished change with no stage run as $0. |   |
| F2 | primary | Important | spectre/changes/kan-989-stats-ui-dashboard-review/tasks.md:100 | four tasks' Baseline fields do not match what their own recorded counter measures at their commits. |   |
| F3 | primary | Minor | stats/web/src/views/StageLeaderboard.tsx:28 | task 6 step 2's default sort Total cost descending was not implemented, and no Correction records it. |   |
| F4 | primary | Minor | stats/internal/store/health.go:123 | a period whose stage runs carry no cost shows Total cost $0.00, never an unavailable value. |   |
| F5 | primary | Minor | stats/README.md:62 | the README still says the Flow health view shows which guards ever fire. |   |
| F6 | primary | Minor | stats/internal/store/aggregate.go:73 | the board's whole-life run-totals LATERAL makes LiveStateBoard 4–8x slower. |   |
| F7 | principles | Minor | stats/internal/store/aggregate.go:79 | DRY: the stage run cost SQL expression is written out twice. |   |
| F8 | principles | Minor | stats/web/src/views/Decisions.tsx:106 | summarizeByImplementerModel repeats summarizeByClassAndExecution's grouping loop and the findings rule. |   |
| F9 | principles | Minor | stats/web/src/views/FlowHealth.tsx:18 | DAY_MS re-declares PeriodPicker's constant, and formatLeadTime is a formatter defined in a view. |   |

findings-total: 9
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed

reproducers-total: 9
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 none — plan drift, rendered order is correct
finding-reproducer: F4 none — follows by construction from the COALESCE
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F6 none — timing-based
finding-reproducer: F7 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F8 none — structural
finding-reproducer: F9 none — structural

## Pass log

### Round 0

- diff size: 5020 lines (over cap) — proceeded automatically, the panel reads the whole diff
- docs-only: exit 1 (first non-doc path stats/cmd/flow/record_test.go) — resolved roster runs unchanged: primary+principles, plus test-sanity+call-efficiency
- roster: full
- no addition this round — the resolved list ran alone.
- standards passed to principles: CLAUDE.md, AGENTS.md

### Round 1

- primary+principles re-run on slot deltas: clean

### Round 2

- fix round 1 (tasks 11-12): primary on opus/medium over 0a042b64..b433ccb7 — clean, no findings
