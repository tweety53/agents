# Review panel — kan-382-flow-review-panel-reproducer-shape-rejected

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | Major | stats/cmd/flow/record_test.go | Mutation 8 — removing the exemption path early return survives the suite: no test pins that a none — <reason> reason may carry shell metacharacters, which the read-time guard accepts, so an equivalent refactor could bounce a legal exemption at write time |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 .superpowers/sdd/reproducers/0-mutation-1.sh
