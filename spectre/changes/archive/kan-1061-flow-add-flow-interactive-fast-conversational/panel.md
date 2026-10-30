# Review panel — kan-1061-flow-add-flow-interactive-fast-conversational

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | failure-modes | Important | stats/cmd/flow/request.go:222 | The landed-end gate treats every `ErrUnavailable` from `ListStageRuns` as "store unreachable" and passes with a ⚠ line. But `ListStageRuns` (stats/internal/client/client.go:623) also wraps a reachable daemon's non-200 answer in `ErrUnavailable` |   |
| F2 | 0 | primary | Minor | stats/internal/api/stats.go:762 | `requestTasks` maps state off `r.Outcome` and `r.State == "in progress"` rather than off `r.State` alone; two readings of the same derivation. |   |
| F3 | 0 | principles | Minor | stats/internal/store/aggregate.go:232 | DRY: the "started" fact (`EXISTS (SELECT 1 FROM stage_runs sr WHERE sr.change_id = ... AND sr.request_n = rq.n)`) is written twice, here and in `requestColumns` (`stats/internal/store/requests.go:30`). |   |
| F4 | 3 | mutation | Important | stats/web/src/views/RunDetail.tsx:913 | subagent rows' bare={interactive} can be dropped and no unit test fails |   |
| F5 | 3 | mutation | Important | stats/web/src/views/RunDetail.tsx:740 | removing the Timeline's interactive order: 5 survives |   |
| F6 | 3 | mutation | Important | stats/web/src/views/RunDetail.tsx:794 | removing Cost by stage's interactive order: 7 survives |   |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 .superpowers/sdd/reproducers/0-failure-modes-1.sh
finding-reproducer: F2 none — readability only
finding-reproducer: F3 none — structural duplication, no behaviour
finding-reproducer: F4 .superpowers/sdd/reproducers/2-3-1.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/2-3-2.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/2-3-3.sh

## Pass log

### Round 0

- roster: compact — 63
- diff size: 4885 lines, under cap — proceed
- docs-only: exit 1 — hooks/enforce-agent-baseline.py; roster dispatched: primary+principles, test-sanity+call-efficiency, mutation+failure-modes
- no addition this round — the resolved list ran alone.
- standards: CLAUDE.md, AGENTS.md
- dispatched: panel-0-primary+principles on opus/low; panel-0-test-sanity+call-efficiency on opus/low; panel-0-mutation+failure-modes on opus/low
- verdict: fix (0 Critical, 1 Important)

### Round 1

- panel-fix-1: F1 (Important), F2, F3 (Minor) to one fixer chunk on the decision fixer pair opus/medium
- auto-decided: a response without the Flow-Daemon header stays an unreachable store (skip with warning), per design Request-end guard point 4 and the raising report
- dispatched: panel-1-primary+principles+failure-modes on opus/low
- verdict: clean
fix-mutation: stats/cmd/flow/request.go — dropped && !errors.Is(err, client.ErrDaemonAnswered) — TestRecordRequestEndLandedGuard
fix-mutation: stats/internal/client/client.go — dropped ErrDaemonAnswered from non-200 wrap — TestRecordRequestEndLandedGuard
fix-mutation: stats/internal/client/client.go — dropped ErrDaemonAnswered from bad-JSON wrap — TestRecordRequestEndLandedGuard
fix-mutation: stats/internal/api/stats.go — case queued made unreachable — TestStateBoardServesRequestsAsTasks
fix-mutation: stats/internal/api/stats.go — default done changed to pending — TestStateBoardServesRequestsAsTasks
fix-mutation: stats/internal/store/requests.go — requestStarted forced false — TestListRequestsDerivesState
fix-mutations-total: 6

### Round 2

- in-run verify fix 1 re-review: primary + always-on extras over late-fix.diff since 96f9cf96
- dispatched: panel-2-primary on opus/low; panel-2-test-sanity+call-efficiency on opus/low; panel-2-mutation+failure-modes on opus/low
- verdict: clean

### Round 3

- in-run visual-verify fix 2 re-review: primary + always-on extras over late-fix-3.diff since 9825b70e
- dispatched: panel-3-primary on opus/low; panel-3-test-sanity+call-efficiency on opus/low; panel-3-mutation+failure-modes on opus/low
- verdict: fix (0 Critical, 3 Important)

### Round 4

- panel-fix-4 re-review of F4-F6: mutation+failure-modes over fix-round-4.diff since ffeaf7eb
- dispatched: panel-4-mutation+failure-modes on opus/low
- verdict: clean
fix-mutation: stats/web/src/views/RunDetail.tsx — subagent row `<StatusDot status="running" bare={interactive} />` → `<StatusDot status="running" />` — interactive run view > interactive run view draws the Agents status as the dot alone
fix-mutation: stats/web/src/views/RunDetail.tsx — Timeline's `style={interactive ? { order: 5 } : undefined}` removed — interactive run view > interactive run view lays Requests and Cost by stage in its right column
fix-mutation: stats/web/src/views/RunDetail.tsx — Cost by stage's `style={interactive ? { order: 7 } : undefined}` removed — interactive run view > interactive run view lays Requests and Cost by stage in its right column
fix-mutations-total: 3

### Round 5

- in-run visual-verify fix 3 (Task 16) re-review: primary + always-on extras over late-fix-5.diff since 14651cfd
- dispatched: panel-5-primary on opus/low; panel-5-test-sanity+call-efficiency on opus/low; panel-5-mutation+failure-modes on opus/low
- verdict: clean
