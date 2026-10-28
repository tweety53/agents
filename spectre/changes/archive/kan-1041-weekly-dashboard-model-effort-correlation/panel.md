# Review panel — kan-1041-weekly-dashboard-model-effort-correlation

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | primary | Important | stats/cmd/flow/pricing.go:53 | flow pricing reprice cannot finish over the window it is meant to backfill: RepriceSince prices ~15k runs serially inside one request capped by the CLI's 2s default timeout and flowd's 30s write timeout. |   |
| F2 | 0 | primary | Minor | stats/web/src/views/WeeklyCosts.tsx:150 | WeeklyCostChart's shown state is initialised once; a refetch that changes the series set misaligns toggles with lines. |   |
| F3 | 0 | primary | Minor | stats/internal/store/pricing.go:702 | A run whose dispatches priced but whose models returned ErrPricingNotFound is counted skipped, understating priced. |   |
| F4 | 0 | principles | Minor | stats/internal/store/aggregate.go:493 | The most-turns-ties-by-name rule is written twice (servedModel in Go and the SQL in WeeklyModelEffortCosts) with no test that they agree on a tie. |   |
| F5 | 0 | principles | Minor | stats/cmd/flow/pricing.go:36 | The pricing command reuses settingsConnFlags, a name that no longer fits. |   |
| F6 | 0 | call-efficiency | Minor | stats/internal/store/pricing.go:698 | RepriceSince calls Price once per stage run in the window. |   |
| F7 | 0 | failure-modes | Minor | stats/internal/store/aggregate.go:493 | v::numeric on a non-numeric served_models value fails the whole view query. |   |
| F8 | 0 | mutation | Important | stats/internal/store/pricing.go:278 | Tie-break m < best -> m > best survives; no test pins most-turns tie by name. |   |
| F9 | 0 | mutation | Important | stats/internal/store/pricing.go:519 | Guard model == "" -> db.Model == "" survives; no test of a dispatch with served_models but empty recorded model. |   |
| F10 | 0 | mutation | Important | stats/internal/store/pricing.go:702 | Dropping ErrPricingNotFound from the skipped case survives; no test of a skipped-by-no-rate run. |   |
| F11 | 0 | mutation | Important | stats/internal/store/pricing.go:690 | started_at >= $1 -> > survives; since boundary untested. |   |
| F12 | 0 | mutation | Important | stats/internal/store/aggregate.go:503 | GREATEST(0, floor -> -1000 survives; session floor untested. |   |
| F13 | 0 | mutation | Important | stats/internal/store/aggregate.go:493 | SQL served-model order DESC->ASC survives; SQL most-turns pick untested. |   |
| F14 | 0 | mutation | Important | stats/cmd/flow/pricing.go:47 | fset.NArg() != 0 dropped survives; stray-arg rejection untested. |   |
| F15 | 1 | mutation | Important | stats/internal/store/aggregate.go:493 | served-model tie-break k -> k DESC survives: the added test ties only on efforts; model counts are 1 vs 4. |   |
| F16 | 1 | mutation | Important | stats/internal/store/aggregate.go:497 | served-effort order v DESC -> v ASC survives: the only effort test data is a 2/2 tie. |   |
| F17 | 5 | mutation | Important | stats/web/src/views/WeeklyCosts.tsx:241 | Mutant pillWidth(cost + delta) -> pillWidth(cost) survives: fixture figures fit the 40/84 floors, so no render-path test exercises a widened pill. |   |
| F18 | 5 | test-sanity | Minor | stats/web/src/views/views.test.tsx:584 | The two direct pillWidth assertions compare the function against its own formula and prove no rendered pill fits its text. |   |
| F19 | 7 | primary | Minor | stats/internal/guard/check_base_moved_test.go:111 | New lockedWriter duplicates the packages existing lsrrLocked. |   |
| F20 | 7 | primary | Minor | stats/internal/guard/check_base_moved_test.go:622 | Failure message counts newlines while the assertion counts out/err tokens. |   |

findings-total: 20
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 withdrawn decision reprice-command counts an ErrPricingNotFound run as skipped by design
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 withdrawn names no defect: one Price per run is the reprice mechanism on a manual path
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed

reproducers-total: 20
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 none — needs a live refetch that changes the series set
finding-reproducer: F3 none — counting semantics, no observable crash
finding-reproducer: F4 none — DRY between Go and SQL; covered by tie tests
finding-reproducer: F5 none — naming
finding-reproducer: F6 none — cost scales with the window
finding-reproducer: F7 none — needs malformed transcript data
finding-reproducer: F8 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/0-mutation-4.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/0-mutation-5.sh
finding-reproducer: F13 .superpowers/sdd/reproducers/0-mutation-6.sh
finding-reproducer: F14 .superpowers/sdd/reproducers/0-mutation-7.sh
finding-reproducer: F15 .superpowers/sdd/reproducers/1-mutation-1.sh
finding-reproducer: F16 .superpowers/sdd/reproducers/1-mutation-2.sh
finding-reproducer: F17 .superpowers/sdd/reproducers/5-mutation-1.sh
finding-reproducer: F18 none — Minor
finding-reproducer: F19 none — no wrong behaviour
finding-reproducer: F20 none — failure message only

## Pass log

### Round 0

- roster: compact — 21; docs-only exit 1 (stats/cmd/flow/main.go) — resolved roster runs; diff 1798 lines under cap; standards passed: CLAUDE.md, AGENTS.md; no addition this round — the resolved list ran alone; GATE FOCUS: tasks 1-9 fired (>40 lines)
- dispatched: panel-0-primary+principles on opus/low; panel-0-test-sanity+call-efficiency on opus/low
- dispatched: panel-0-mutation+failure-modes on sonnet/low
- verdict: fix (0 Critical, 8 Important)

### Round 1

- panel-fix-1 inline; F14 reproducer re-authored (-run ^TestPricing: the whole-package run left a detached child) and re-proved both legs; auto-decided F3: withdrawn — decision reprice-command counts ErrPricingNotFound as skipped by design; auto-decided F6: withdrawn — names no defect, a manual one-off path
- dispatched: panel-1-primary+principles+failure-modes on sonnet/low over 3faafa95eb601a37fc98eb5f07b06f745d039ac2; panel-1-mutation on sonnet/low
- verdict: fix (0 Critical, 2 Important)
fix-mutation: stats/internal/store/pricing.go — tie m < best -> m > best — TestPriceDispatchUsesServedModel
fix-mutation: stats/internal/store/pricing.go — guard model == "" -> db.Model == "" — TestPriceDispatchUsesServedModel
fix-mutation: stats/internal/store/pricing.go — ErrPricingNotFound dropped from skipped — TestRepriceSinceCountsPricedAndSkipped
fix-mutation: stats/internal/store/pricing.go — started_at >= $1 -> > — TestRepriceSinceCountsPricedAndSkipped
fix-mutation: stats/internal/store/aggregate.go — GREATEST(0 -> GREATEST(-1000 — TestWeeklyModelEffortCostsSplitsDispatchAndSession
fix-mutation: stats/internal/store/aggregate.go — ORDER BY v DESC -> ASC — TestWeeklyModelEffortCostsSplitsDispatchAndSession
fix-mutation: stats/internal/store/aggregate.go — tie ORDER BY k -> k DESC — TestWeeklyModelEffortCostsSplitsDispatchAndSession
fix-mutation: stats/cmd/flow/pricing.go — NArg check dropped — TestPricingRepricePrintsCounts
fix-mutation: stats/cmd/flow/pricing.go — reprice timeout -> 2s — TestPricingRepricePrintsCounts
fix-mutation: stats/internal/api/pricing.go — none — the write-deadline lift needs a >30s request; covered by the live check
fix-mutations-total: 10

### Round 2

- dispatched: panel-2-mutation on sonnet/low
- verdict: clean
fix-mutation: stats/internal/store/aggregate.go — ORDER BY v DESC, k -> k DESC — TestWeeklyModelEffortCostsSplitsDispatchAndSession
fix-mutation: stats/internal/store/aggregate.go — ORDER BY v DESC -> ASC — TestWeeklyModelEffortCostsSplitsDispatchAndSession
fix-mutations-total: 2

### Round 3

- dispatched: panel-3-primary on opus/low; panel-3-test-sanity+call-efficiency on opus/low; panel-3-mutation+failure-modes on sonnet/low
- late-fix re-review of In-run fix 1 (task 10): primary + always-on extras over late-fix.diff
- Round 3 (late-fix re-review of In-run fix 1, task 10): primary, test-sanity+call-efficiency, mutation+failure-modes — all clean, 0 findings; 3/3 mutants killed.
- verdict: clean

### Round 4

- dispatched: panel-4-primary on opus/low; panel-4-test-sanity+call-efficiency on opus/low; panel-4-mutation+failure-modes on sonnet/low
- late-fix re-review of In-run fix 2 (task 11): primary + always-on extras over late-fix.diff
- Round 4 (late-fix re-review of In-run fix 2, task 11): primary, test-sanity+call-efficiency, mutation+failure-modes — all clean, 0 findings; 4/4 mutants killed.
- verdict: clean
fix-mutation: stats/internal/store/pricing.go — clear-partial-total removed — TestPriceOneUnpriceableBucketOmitsTopLevelTotal
fix-mutation: stats/internal/store/aggregate.go — AND metrics ? 'cost_usd' removed — TestWeeklyModelEffortCostsSplitsDispatchAndSession
fix-mutations-total: 2

### Round 5

- dispatched: panel-5-primary on opus/low; panel-5-test-sanity+call-efficiency on opus/low; panel-5-mutation+failure-modes on sonnet/low
- late-fix re-review of In-run fix 3 (task 12): primary + always-on extras over late-fix.diff
- verdict: fix (0 Critical, 1 Important)
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — pillWidth base widths reverted to 12+chars*6 — weekly costs matches the prototype header and pill widths
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — pillWidth(cost + delta) -> pillWidth(cost) — weekly costs widens a pill whose figure overflows the prototype's width
fix-mutations-total: 2

### Round 6

- dispatched: panel-6-mutation on sonnet/low
- fix round re-run of mutation over panel-fix-6
- verdict: clean

### Round 7

- dispatched: panel-7-primary on opus/low; panel-7-test-sanity+call-efficiency on opus/low; panel-7-mutation+failure-modes on sonnet/low
- late-fix re-review of In-run fix 4 (task 13): primary + always-on extras over late-fix.diff
- verdict: clean
fix-mutation: stats/internal/guard/check_base_moved_test.go — merged lock removed (merged := io.Writer(&all)) — TestRunGuardMergesAChildsTwoStreams
fix-mutations-total: 1
