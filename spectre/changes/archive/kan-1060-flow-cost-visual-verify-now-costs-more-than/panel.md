# Review panel — kan-1060-flow-cost-visual-verify-now-costs-more-than

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | failure-modes | Important | stats/internal/guard/visualverifypass.go:80 | `die` exits 2 without ending a stage `markBegin` opened |   |
| F2 | 0 | failure-modes | Minor | stats/internal/guard/visualverifypass_prepare.go:406 | With several rows, a refused `dispatch begin` for row 2 `stop`s the stage but leaves row 1's dispatch begun and never ended. |   |
| F3 | 0 | mutation | Important | stats/internal/guard/visualverifypass.go:179 | `--started` without `--continue` usage check deleted (`if false {`): no test fails. |   |
| F4 | 0 | mutation | Important | stats/internal/guard/visualverifypass_close.go:171 | `exits:` non-zero detection narrowed to exit 1 only (`n == "1"`): no test uses an exit other than 0/1 |   |
| F5 | 0 | mutation | Important | stats/internal/guard/visualverifypass_close.go:90 | the `cause` default to `environment` deleted: no test closes a blocked report lacking `cause:` |   |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-failure-modes-1.sh
finding-reproducer: F2 none — needs a stub flow that refuses only the second dispatch begin
finding-reproducer: F3 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-mutation-3.sh

## Pass log

### Round 0

- base: MOVED 8 commits with overlap (stats/internal/api, store/aggregate, web api.ts, views.test.tsx) — exit 3 auto-resolved Continue per Auto-resolution; integrate sync settles it
- diff size: 7205 lines over cap — proceeded automatically
- docs-only: exit 1 (scripts/check-frame-list.sh) — roster unchanged: primary+principles, plus test-sanity+call-efficiency and mutation+failure-modes
- roster: compact — 11
- no addition this round — the resolved list ran alone.
- standards: CLAUDE.md
- dispatched: panel-0-primary+principles on opus/low; panel-0-test-sanity+call-efficiency on sonnet/low
- dispatched: panel-0-mutation+failure-modes on opus/low
- verdict: fix (0 Critical, 4 Important)

### Round 1

- panel-fix-1: one chunk, F1–F5 (4 Important + 1 Minor), fixer opus/low, reads final-review.diff findings
- dispatched: panel-1-mutation+failure-modes on opus/low over 380d9bb313e95cb9c1f8845f17b9d4d791f29ca8
- verdict: clean
fix-mutation: stats/internal/guard/visualverifypass.go — die's markEnd("stopped") replaced by no-op — TestVisualVerifyPassCloseDieEndsStage
fix-mutation: stats/internal/guard/visualverifypass_prepare.go — begun-row end loop iterates nothing — TestVisualVerifyPassPrepareBeginsRows
fix-mutation: stats/internal/guard/visualverifypass.go — --started without --continue check to if false — TestVisualVerifyPassUsage
fix-mutation: stats/internal/guard/visualverifypass_close.go — n != "0" to n == "1" — TestVisualVerifyPassCloseBlockedDefaults
fix-mutation: stats/internal/guard/visualverifypass_close.go — cause default to if false — TestVisualVerifyPassCloseBlockedDefaults
fix-mutations-total: 5
