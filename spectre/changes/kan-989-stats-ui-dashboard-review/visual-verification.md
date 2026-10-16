# Visual verification — kan-989-stats-ui-dashboard-review

No `mockups` root is declared: no frame composed; band, seam, ink, rows and matrix readings are n/a — no frame. Captures are the committed full-app-suite baselines (capture exit 0, regenerated identical), period pinned by `PINNED_QUERY`, against the UI-test stack on 127.0.0.1:4174.

## Live state board
- Screenshot: `../../../stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-state-board-darwin.png`
- Seen: Changes tile 3; table Project, Change, State, Plan growth, Updated, Runs, Cost, Wall clock, Errors — no By / Next command columns. Rows: kan-101 STARTED 0 runs $0.00 "0 ms"; kan-102 IN_PROGRESS 1 $0.179 5.0 min; kan-201 FINISHED 2 $2.778 10.0 min. No pipeline-fix row. Model picker disabled with its reason. Nav has no Runs or Cache efficiency entry.
- Notes: a change with no runs shows Wall clock "0 ms" and Cost "$0.00" where Plan growth shows "—"; cost shown at 3 decimals except "$0.00".

## Stage leaderboard
- Screenshot: `../../../stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-stage-leaderboard-darwin.png`
- Seen: subtitle "The ten stages costing the most in total"; tiles Stages 2, Total runs 3; columns Command, Stage, Runs, Total cost, Mean cost, Median cost, P90 cost; /flow SDD + TDD per task 2 $2.929 $1.4645 $1.4645 $2.4929; /flow-fast integrate 1 $0.028 ×4.
- Notes: Total cost at 3 decimals, Mean/Median/P90 at 4 decimals in the same table. Only 2 stages seeded — the top-ten cut is not exercised by any capture.

## Reviewers
- Screenshot: `../../../stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-reviewers-darwin.png`
- Seen: empty state only — Slots 0, Total dispatches 0, "No reviewer dispatches in this period." with Reset period to default. The new Efficiency column and the dropped Findings/dispatch and Withdrawn columns are not visible in any capture.

## Decisions
- Screenshot: `../../../stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-decisions-darwin.png`
- Seen: no Runs tile, no Every decision table; two cards "By class and execution" and "By implementer model", both "No decisions in this period." The populated By implementer model table is not visible in any capture.

## Flow health
- Screenshot: `../../../stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-flow-health-darwin.png`
- Seen: eight tiles only — Changes started 2, Changes finished 1, Total cost $2.957, Mean cost per finished change $2.778, Median lead time 1.1 d, Mean panel rounds "unavailable" (italic), Findings (C/I/Mi) 0/0/0, Guard fires 0. No tables. Model picker disabled with its reason.

## Self-review fixes
- Screenshot: `../../../stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-self-review-darwin.png`
- Seen: Fixed 0, Filed 0, Every finding empty state with filters Change/Angle/Outcome; no Per change block.

## Motions
None named.
