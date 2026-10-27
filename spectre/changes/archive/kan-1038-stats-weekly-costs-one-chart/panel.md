# Review panel — kan-1038-stats-weekly-costs-one-chart

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | primary | Important | skills/flow/brainstorm-planner.md:41 | The prototype bullet demands an iterate-until-approved operator round that the recommended-decisions mode and the unattended self-review/in-run fix planners never ask. |   |
| F2 | 0 | primary | Minor | stats/web/src/views/WeeklyCosts.tsx:217 | Focusable chart points have no accessible name: tabIndex=0 but no aria-label or title, a regression from TimeSeriesPanel's titled points. |   |
| F3 | 0 | principles | Minor | stats/web/src/views/WeeklyCosts.tsx:103 | stackLabels' push-apart and shift-up branches are untested: no fixture puts two labels within 17px. |   |
| F4 | 0 | failure-modes | Minor | stats/web/src/views/WeeklyCosts.tsx:268 | lastEightWeeks is memoised once from new Date(), so a tab left open across a Wednesday keeps a stale window. |   |
| F5 | 0 | mutation | Important | stats/web/src/views/WeeklyCosts.tsx:56 | change() guard previous === 0 is untested; mutating it survives every test. |   |
| F6 | 0 | mutation | Important | stats/web/src/views/WeeklyCosts.tsx:66 | deltaText's pct === 0 neutral branch is untested. |   |
| F7 | 0 | mutation | Important | stats/web/src/views/WeeklyCosts.tsx:140 | week ? change mutated to week > 1 ? change survives: a dropped week-1 delta goes unnoticed. |   |
| F8 | 0 | mutation | Important | stats/web/src/views/WeeklyCosts.tsx:186 | The line-break week && line[week-1] ? L mutated to week ? L survives: a line bridging a gap goes unnoticed. |   |
| F9 | 2 | mutation | Important | stats/web/src/views/WeeklyCosts.tsx:187 | Adding \|\| first.series === 3 to the guard hides the Self-review line and every test passes: no test asserts a Self-review line is drawn. |   |
| F10 | 4 | primary | Important | spectre/changes/kan-1038-stats-weekly-costs-one-chart/visual-verification.md:23 | The visual-verification record still describes the spec as it was before fix 2: no pointer park, no hover capture, no tooltip evidence, and Overall 2026-08-12 read as 3.36. |   |
| F11 | 6 | primary | Minor | spectre/changes/kan-1038-stats-weekly-costs-one-chart/visual-verification.md:28 | visual record still reads 'vs previous week' and lists the three tooltip departures as open; tooltip evidence PNG stale |   |

findings-total: 11
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed

reproducers-total: 11
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F4 none — needs a wall-clock wait across a Wednesday
finding-reproducer: F5 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/0-mutation-4.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/2-mutation-1.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/4-primary-1.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/6-primary-1.sh

## Pass log

### Round 0

- dispatched: panel-0-primary+principles on opus/medium; panel-0-test-sanity+call-efficiency on sonnet/low; panel-0-failure-modes+mutation on sonnet/low
- verdict: fix (0 Critical, 5 Important)

### Round 1

- dispatched: panel-1-primary+principles+failure-modes on sonnet/low; panel-1-mutation on sonnet/low
- F8 citation: the -cite send-back reported line 397 but left :93 in the file; parent transcribed the slot-reported line 397 (inside the views.test.tsx hunk); F8 kill independently shown by break-and-prove
- verdict: clean
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — previous === 0 to === -1 — weekly costs labels zero changes, skips a zero previous week and stacks close labels apart
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — pct === 0 to === 999 — weekly costs labels zero changes, skips a zero previous week and stacks close labels apart
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — week ? change to week > 1 ? change — weekly costs labels zero changes, skips a zero previous week and stacks close labels apart
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — line-break pen-lift dropped — weekly costs breaks a line at an unmeasured week
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — aria-label series name dropped — weekly costs breaks a line at an unmeasured week
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — LABEL_GAP push-apart dropped — weekly costs labels zero changes, skips a zero previous week and stacks close labels apart
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — overflow shift disabled — weekly costs labels zero changes, skips a zero previous week and stacks close labels apart
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — none — the day-keyed memo depends on the wall clock crossing midnight
fix-mutation: skills/flow/brainstorm-planner.md — none — skill text; its reproducer flipped
fix-mutations-total: 9

### Round 2

- dispatched: panel-2-primary on opus/medium; panel-2-test-sanity+call-efficiency on sonnet/low; panel-2-failure-modes+mutation on sonnet/low
- verdict: fix (0 Critical, 1 Important)

### Round 3

- dispatched: panel-3-mutation on sonnet/low
- verdict: clean
fix-mutation: stats/web/src/views/WeeklyCosts.tsx — guard also hides series 3 — weekly costs breaks a line at an unmeasured week
fix-mutations-total: 1

### Round 4

- dispatched: panel-4-primary on opus/medium; panel-4-test-sanity+call-efficiency on sonnet/low; panel-4-failure-modes+mutation on sonnet/low
- verdict: fix (0 Critical, 1 Important)

### Round 5

- dispatched: panel-5-primary on sonnet/low
- verdict: clean
fix-mutation: spectre/changes/kan-1038-stats-weekly-costs-one-chart/visual-verification.md — none — a record, no executable behaviour
fix-mutations-total: 1

### Round 6

- dispatched: panel-6-primary on opus/medium; panel-6-test-sanity+call-efficiency on sonnet/low; panel-6-failure-modes+mutation on sonnet/low
- verdict: clean
