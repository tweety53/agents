# Review panel — kan-434-flow-wave-parallel-bundles-via-an-after-plan

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | Minor | scripts/plan-dispatch-bundles.py:571 | the lexical-sort mutant of the after-line ordering survives the whole suite: no case pins numeric ordering of after <k>: ids, so a regression to sorted(acc) emits after 1: 1 10 2 where the contract says sorted numerically |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 none — the defect is a harness gap; demonstrating it requires mutating scripts/plan-dispatch-bundles.py (a sed-class edit) before re-running the harness, which the reproducer shape rules refuse
