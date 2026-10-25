# Review panel — kan-1033-plan-mode-planning

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow-contracts/pipeline.md:150 | contract texts claim every dispatch lacks Agent; planner rows now carry it |   |
| F2 | primary | Minor | stats/internal/guard/check_dispatch_paragraphs_test.go:55 | dpFixture doc comment displaced by dpCleanPlanning |   |
| F3 | primary | Minor | spectre/changes/kan-1033-plan-mode-planning/tasks.md:1 | own tasks.md lacks the Design shape line |   |
| F4 | principles | Minor | skills/flow-self-review/SKILL.md:186 | flow-planner-medium called a flow-<effort> definition |   |
| F5 | mutation | Important | stats/internal/guard/dispatchparagraphs.go:83 | PLANNING DELEGATION phrases unpinned |   |
| F6 | mutation | Minor | stats/internal/guard/planshape.go:389 | stop-at-first-task untested |   |
| F7 | mutation | Minor | stats/internal/guard/planshape.go:407 | heading trim untested |   |
| F8 | mutation | Minor | stats/internal/guard/planshape.go:379 | regex end anchor untested |   |
| F9 | failure-modes | Minor | stats/internal/guard/planshape.go:400 | unreadable design.md reported as missing |   |
| F10 | mutation | Important | stats/internal/guard/planshape.go:406 | unreadable design.md branch untested |   |
| F11 | primary | Important | hooks/enforce-agent-baseline.py:158 | flake fix lacked measured rate and run count; recorded in design.md: 2/2 failed before, 4/4 passed after |   |
| F12 | primary | Minor | hooks/enforce-agent-baseline.py:158 | FLOW_HOOK_UNBOUNDED lifts the fail-open bound if set in production |   |

findings-total: 12
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
finding-status: F12 withdrawn accepted test seam; production never sets FLOW_HOOK_UNBOUNDED

reproducers-total: 12
finding-reproducer: F1 .superpowers/sdd/reproducers/0-I1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-M1-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-M2-1.sh
finding-reproducer: F4 none — wording judgment
finding-reproducer: F5 .superpowers/sdd/reproducers/0-S1-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-S2-1.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-S3-1.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/0-S4-1.sh
finding-reproducer: F9 none — error-cause wording
finding-reproducer: F10 .superpowers/sdd/reproducers/1-M1-1.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/3-F1-1.sh
finding-reproducer: F12 none — operator-set only

## Pass log

### Round 0

- diff size 1028 lines, under cap; docs-only verdict 1 (first non-doc path scripts/plan-agent-pair.sh) — roster unchanged: primary+principles, plus always-on test-sanity+call-efficiency and mutation+failure-modes; roster: compact — 85; no addition this round — the resolved list ran alone
- verdict: fix (0 Critical, 2 Important)

### Round 1

- no pass row recorded — dispatched: panel-1-primary+principles, panel-1-test-sanity+call-efficiency, panel-1-mutation+failure-modes
- verdict: fix (0 Critical, 1 Important)

### Round 2

- no pass row recorded — dispatched: panel-2-mutation
- verdict: clean

### Round 3

- no pass row recorded — dispatched: panel-3-primary+principles
- verdict: fix (0 Critical, 1 Important)

### Round 4

- no pass row recorded — dispatched: panel-4-primary
- verdict: clean
