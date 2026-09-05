# Review panel — kan-384-flow-dispatch-cost-still-unmeasured-for-most

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | Major | stats/cmd/flow/stage.go:319-321 | Dropping the != "" guard on CLAUDE_CODE_SESSION_ID (always binding, even empty) survives the full TestStageBegin* suite — no test covers -session absent and CLAUDE_CODE_SESSION_ID unset/empty, so a mark made with the var unset would send an empty-string sessionId instead of omitting the field. |
| F2 | principles | Minor | stats/cmd/flow/stage.go:321-324 | os.Getenv("CLAUDE_CODE_SESSION_ID") is called twice (once in the case condition, once again in the body) where the sibling precedence helper resolveHarness reads its env var once via if v := os.Getenv(...); v != "". |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 .superpowers/sdd/reproduce-f1.sh
finding-reproducer: F2 none — a style/efficiency nit, not a behavioral defect; grep confirms the double call: grep -n 'CLAUDE_CODE_SESSION_ID' stats/cmd/flow/stage.go
