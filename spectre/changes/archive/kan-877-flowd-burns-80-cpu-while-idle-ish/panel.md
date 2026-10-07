# Review panel — kan-877-flowd-burns-80-cpu-while-idle-ish

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/harvest/watcher.go:895 | the "scanned once" flag is set when the files have been read, not when the token has been bound: a failed BindSession (or a failed ambiguity give-up record) is never retried from the already-read bytes, so the token later gives up as session-never-bound |   |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh

## Pass log

### Round 0

- roster: compact — 40
- diff size: 578 lines, under cap; docs-only: exit 1 (stats/cmd/flow/stage.go); roster dispatched: primary+principles; standards: CLAUDE.md, AGENTS.md; no addition this round — the resolved list ran alone

### Round 1

- fix round 1: F1 (primary, Important) fixed inline by the parent — decision fixer is inline; reads fix-round-1.diff
- re-run: primary (raised F1) on rerun pair sonnet/low, reads fix-round-1.diff; principles raised nothing — not re-run; no addition this round — the resolved list ran alone
fix-mutation: stats/internal/harvest/watcher.go — rescanIfRetried body: w.retriedScanDone = false replaced by a no-op — TestRetriedTokenScanRerunsAfterFailedResolve
fix-mutations-total: 1
