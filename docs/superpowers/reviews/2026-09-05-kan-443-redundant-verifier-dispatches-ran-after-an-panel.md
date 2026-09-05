# Review panel — kan-443-redundant-verifier-dispatches-ran-after-an

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | Major | stats/internal/records/render.go:340 | TestRenderLedgerNamesDispatchKey uses a key with no marker characters, so removing the neutraliseMarkers wrap on the new Key line still passes — a dispatch key containing #, backtick, or other markers would render un-neutralised and break the ledger's markdown structure |
| F2 | mutation | Minor | stats/internal/records/render.go:339-341 | TestRenderLedgerNamesDispatchKey only checks the Key line exists via strings.Contains, not its position relative to Model — moving the block after Model still passes, so field ordering has no regression coverage |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 none — patch-based reproducer, see panel-report-0-mutation.md
finding-reproducer: F2 none — patch-based reproducer, see panel-report-0-mutation.md
