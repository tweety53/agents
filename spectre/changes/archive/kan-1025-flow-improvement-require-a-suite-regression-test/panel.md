# Review panel — kan-1025-flow-improvement-require-a-suite-regression-test

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | primary | Important | stats/internal/guard/panelfindingsclosed.go:336 | git ls-files --error-unmatch reads the path as a pathspec, so a directory, `.` or a glob is accepted whenever anything under it is tracked. |   |
| F2 | 0 | primary | Important | stats/internal/guard/check_panel_findings_closed_test.go:579 | No test covers a path tracked only in a recorded worktree, nor the unreadable-state-record exit 2; mutating either branch leaves the suite green. |   |
| F3 | 0 | mutation | Important | stats/internal/guard/panelfindingsclosed.go:341 | A fixed Critical finding is never exercised: dropping `critical` from the severity filter survives. |   |
| F4 | 0 | mutation | Important | stats/internal/guard/panelfindingsclosed.go:351 | The .superpowers/ refusal is untested in effect: the case uses an untracked path, so removing the prefix check still exits 1 via 'is not tracked'. |   |
| F5 | 0 | failure-modes | Minor | stats/internal/guard/panelfindingsclosed.go:475 | The regex accepts `regression-test: F1 — none` with no reason, so the none form's required reason is not enforced. |   |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F5 none — Minor; demonstrating needs a new Go test case, which is the fix itself

## Pass log

### Round 0

- dispatched: panel-0-primary+principles on opus/medium; panel-0-test-sanity+call-efficiency on sonnet/low; panel-0-failure-modes+mutation on opus/low
- diff size 326 under cap; docs-only exit 1 (scripts/check-panel-findings-closed.sh); roster: primary+principles opus/medium, test-sanity+call-efficiency sonnet/low, failure-modes+mutation opus/low; roster: compact — 37; standards: CLAUDE.md, AGENTS.md; no addition this round — the resolved list ran alone
- ceiling breach: failure-modes+mutation ran ~45m with no report, stopped, closed timed-out; re-dispatched once
- parent corrections: F1/F2 premise line 79→80 (line number off, content verbatim on 80); F3/F4 mutation reproducers had generic exits under the # mutation-reproducer declaration — exits swapped to the mutation convention and test command narrowed to -run TestCheckPanelFindingsClosed (whole-package run outlived the 300s bound); defect logic unchanged
- verdict: fix (0 Critical, 4 Important)

### Round 1

- panel-fix-1 inline by parent over F1-F5; diff fix-round-1.diff; reproducers flipped demonstrated→not demonstrated; re-run: primary (F1,F2) + mutation+failure-modes (F3,F4,F5) on rerun pair sonnet/low
- dispatched: panel-1-primary+failure-modes+mutation on sonnet/low
- verdict: clean
fix-mutation: stats/internal/guard/panelfindingsclosed.go — tracked() exact-match reverted to any output — TestCheckPanelFindingsClosedFixedNeedsARegressionTest
fix-mutation: stats/internal/guard/panelfindingsclosed.go — recorded-worktree lookup forced false — TestCheckPanelFindingsClosedFixedNeedsARegressionTest
fix-mutation: stats/internal/guard/panelfindingsclosed.go — state-read failure ignored — TestCheckPanelFindingsClosedFixedNeedsARegressionTest
fix-mutation: stats/internal/guard/panelfindingsclosed.go — critical dropped from severity filter — TestCheckPanelFindingsClosedFixedNeedsARegressionTest
fix-mutation: stats/internal/guard/panelfindingsclosed.go — .superpowers/ refusal disabled — TestCheckPanelFindingsClosedFixedNeedsARegressionTest
fix-mutation: stats/internal/guard/panelfindingsclosed.go — none-without-reason check disabled — TestCheckPanelFindingsClosedFixedNeedsARegressionTest
fix-mutations-total: 6
