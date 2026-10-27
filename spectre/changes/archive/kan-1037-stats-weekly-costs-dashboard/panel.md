# Review panel — kan-1037-stats-weekly-costs-dashboard

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | primary | Critical | stats/internal/api/stats.go:50 | StatsStore gained WeeklyCosts but fakeStore (internal/web/embed_test.go) and stubStageStore (internal/client/client_test.go) did not; go vet/test ./... fail to compile |   |
| F2 | 0 | primary | Critical | stats/web/src/views/RunDetail.test.tsx:595 | RunDetail.test asserts VIEW_NAMES has length 7; it now has 8, npm test red |   |
| F3 | 0 | primary | Important | stats/web/src/components/TimeSeriesPanel.tsx:90 | each series is one polyline over its non-null points, so the line runs across a null week instead of breaking |   |
| F4 | 0 | primary | Important | spectre/changes/kan-1037-stats-weekly-costs-dashboard/tasks.md:120 | both tasks record Build: green, false while F1 and F2 stand |   |
| F5 | 0 | primary | Minor | stats/web/src/views/WeeklyCosts.tsx:54 | a picker starting mid-week makes the chart's first bucket partial with no label |   |
| F6 | 0 | principles | Important | stats/web/src/components/TimeSeriesPanel.tsx:90 | header comment promises a null breaks the line; the code draws across it | supersedes F3 |
| F7 | 0 | principles | Minor | stats/internal/store/aggregate.go:400 | category stage keys are SQL literals with no link to stats/internal/stages/names.go |   |
| F8 | 0 | test-sanity | Minor | stats/web/src/views/views.test.tsx:314 | the time-series-line >= 5 assertion counts series, not drawn data |   |
| F9 | 0 | call-efficiency | Minor | stats/web/src/views/WeeklyCosts.tsx:50 | table and chart fire the same request when the picker equals the 8-week range |   |
| F10 | 0 | mutation | Important | stats/internal/api/stats.go:762 | DTO category fields can be swapped and the api tests pass |   |
| F11 | 0 | mutation | Important | stats/internal/store/aggregate.go:399 | dropping the pipeline-fix-% match survives; no test seeds a pipeline-fix change |   |
| F12 | 0 | mutation | Important | stats/internal/store/aggregate.go:400 | removing any planning stage but flow.brainstorm survives |   |
| F13 | 0 | mutation | Important | stats/internal/store/aggregate.go:397 | model branch of the cost expression and the model filter are untested |   |
| F14 | 0 | mutation | Important | stats/internal/store/aggregate.go:407 | neutralising the project filter survives |   |
| F15 | 0 | mutation | Minor | stats/internal/store/aggregate.go:406 | < $2 -> <= $2 survives; no run on the period's upper bound |   |
| F16 | 0 | mutation | Important | stats/web/src/views/WeeklyCosts.tsx:52 | chart fetch can ignore the picker and drop model; table can drop model |   |
| F17 | 0 | mutation | Important | stats/web/src/views/WeeklyCosts.tsx:71 | every chart line can plot overallUsd and nothing fails |   |
| F18 | 0 | mutation | Important | stats/web/src/components/TimeSeriesPanel.tsx:49 | multi-series x placement is untested |   |
| F19 | 0 | mutation | Important | stats/web/src/components/TimeSeriesPanel.tsx:85 | per-series stroke can be dropped and the tests pass |   |
| F20 | 1 | mutation | Important | stats/internal/store/aggregate.go:432 | dropping the model filter sr.metrics->'models' ? $4 survives: a week whose runs all lack the model should yield no point |   |
| F21 | 1 | mutation | Important | stats/web/src/components/TimeSeriesPanel.tsx:68 | dropping the shared-days .sort() and lastIndex from the first series' length survive; the fixture's first series already holds every week in order | supersedes F18 |
| F22 | 1 | mutation | Minor | stats/internal/store/aggregate.go:392 | neutering the init panic survives; no test drives init with an undeclared key |   |
| F23 | 3 | primary | Important | stats/web/tests/visual/weekly-costs.spec.ts:33 | the seed extension left the weekly-costs visual spec and its baselines asserting the pre-fix two-point look |   |
| F24 | 3 | mutation | Important | stats/cmd/uitest-seed/seed.go:197 | Surviving mutants: the seed's new stage keys are guarded by no runnable test. |   |
| F25 | 3 | primary | Minor | stats/web/src/views/WeeklyCosts.tsx:82 | emptyMessage is now unreachable: everyWeek always yields 8 rows |   |
| F26 | 3 | principles | Minor | stats/web/src/views/WeeklyCosts.tsx:53 | the literal 8 in everyWeek duplicates the window length lastEightWeeks encodes as + 49 days |   |

findings-total: 26
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 withdrawn the two fetches are the table-fixed-graph-on-picker decision; equal ranges coincide only when the picker is set to exactly the 8-week window
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
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed
finding-status: F25 fixed
finding-status: F26 fixed

reproducers-total: 26
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 none — presentation choice
finding-reproducer: F6 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F8 none — assertion strength only
finding-reproducer: F9 none — cold path, intended two-fetch design
finding-reproducer: F10 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F13 .superpowers/sdd/reproducers/0-mutation-4.sh
finding-reproducer: F14 .superpowers/sdd/reproducers/0-mutation-5.sh
finding-reproducer: F15 .superpowers/sdd/reproducers/0-mutation-6.sh
finding-reproducer: F16 .superpowers/sdd/reproducers/0-mutation-7.sh
finding-reproducer: F17 .superpowers/sdd/reproducers/0-mutation-8.sh
finding-reproducer: F18 .superpowers/sdd/reproducers/0-mutation-9.sh
finding-reproducer: F19 .superpowers/sdd/reproducers/0-mutation-10.sh
finding-reproducer: F20 .superpowers/sdd/reproducers/1-mutation-1.sh
finding-reproducer: F21 .superpowers/sdd/reproducers/1-mutation-2.sh
finding-reproducer: F22 none — init cannot be driven with a bad key without extracting the check
finding-reproducer: F23 .superpowers/sdd/reproducers/3-primary-1.sh
finding-reproducer: F24 .superpowers/sdd/reproducers/3-mutation-1.sh
finding-reproducer: F25 none — dead UI copy, no behavioural defect
finding-reproducer: F26 none — knowledge duplication, no current misbehaviour

## Pass log

### Round 0

- dispatched: panel-0-primary+principles on opus/medium; panel-0-test-sanity+call-efficiency on sonnet/low; panel-0-failure-modes+mutation on opus/low
- verdict: fix (2 Critical, 12 Important)

### Round 1

- dispatched: panel-1-primary+principles+test-sanity on opus/low; panel-1-call-efficiency+mutation on opus/low
- verdict: fix (0 Critical, 2 Important)
fix-mutation: stats/web/src/components/TimeSeriesPanel.tsx — segment adjacency === m.index - 1 -> <= m.index — weekly costs shows all five totals per week and a five-line chart
fix-mutation: stats/internal/api/stats_test.go — drop ImplementUSD from the fake row — TestEveryViewCarriesItsRealNumbersThrough/weekly-costs
fix-mutation: stats/internal/store/aggregate.go — planning key flow.writing-plans -> flow.writing-planz — TestWeeklyCostsBucketsByWednesdayAndCategory (init panic)
fix-mutations-total: 3

### Round 2

- dispatched: panel-2-mutation on opus/low
- verdict: clean
fix-mutation: stats/internal/store/aggregate.go — model filter ? $4 -> true — TestWeeklyCostsHonoursProjectModelAndPeriod
fix-mutation: stats/web/src/components/TimeSeriesPanel.tsx — drop shared-days sort — places every series on the shared, sorted day axis
fix-mutation: stats/internal/store/aggregate.go — flow.decide -> flow.decidz — TestWeeklyCostStagesAreDeclared
fix-mutations-total: 3

### Round 3

- dispatched: panel-3-primary+principles+test-sanity on opus/low; panel-3-failure-modes+mutation on opus/low
- verdict: fix (0 Critical, 2 Important)

### Round 4

- no pass row recorded — dispatched: panel-4-primary+principles+test-sanity, panel-4-failure-modes+mutation
- verdict: clean
fix-mutation: stats/cmd/uitest-seed/seed.go — "flow.sdd-tdd" → "SDD + TDD per task" — renders the 8-week table and the five-line chart in the dark palette
fix-mutations-total: 1
