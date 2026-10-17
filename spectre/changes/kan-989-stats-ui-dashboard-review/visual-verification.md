# Visual verification — kan-989-stats-ui-dashboard-review

Capture: `npx playwright test tests/visual/full-app-suite.spec.ts --update-snapshots -g '<six touched screens>'`, exit 0, at HEAD f43cb050 against the UI-test stack on 127.0.0.1:4174. Every test asserts its nav link carries `aria-current="page"` before the full-page screenshot. No mockups declared, so no composite, band, seam or matrix reading applies (n/a — no frame). No motions named.

## state-board
Evidence: `visual-verification/evidence/full-state-board-darwin.png`. Columns Runs / Cost / Wall clock / Errors present. kan-101-add-widget (0 runs) shows `—` for Cost and Wall clock; kan-102 $0.59 / 35.0 min / 2 errors; kan-201 $2.78 / 10.0 min / 0. Costs at two decimals. No defect.

## stage-leaderboard
Evidence: `visual-verification/evidence/full-stage-leaderboard-darwin.png`. Total cost column present, rows ordered by total ($2.93, $0.41, $0.03); mean/median/p90 at two decimals. Top-ten cap not exercised: the seed holds three stages.

## reviewers
Evidence: `visual-verification/evidence/full-reviewers-darwin.png`. Efficiency column present: primary 11.00, principles 4.00. No defect.

## decisions
Evidence: `visual-verification/evidence/full-decisions-darwin.png`. "By implementer model" block present: opus 1 run $2.75 5.0 min, sonnet 1 run $0.59 35.0 min; costs at two decimals. No defect.

## flow-health
Evidence: `visual-verification/evidence/full-flow-health-darwin.png`. Eight tiles: Changes started 2, Changes finished 1, Total cost $3.37, Mean cost per finished change $2.78, Median lead time 1.1 d, Mean panel rounds 1.00, Findings (C/I/Mi) 1/1/1, Guard fires 1. Total cost re-derived: state-board 0.59 + 2.78 = 3.37 and leaderboard 2.93 + 0.41 + 0.03 = 3.37. No defect.

## self-review
Evidence: `visual-verification/evidence/full-self-review-darwin.png`. Empty state ("No self-review findings in this period." with Reset period to default); no Per change block. Populated path not captured: the seed has no findings.

## nav
Cache efficiency and Runs are absent from the nav on every screen.
