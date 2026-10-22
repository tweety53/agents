# Review panel — kan-921-flow-improvement-port-the-guard-test-runner-run

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | stats/internal/guard/runguardtests.go:92 | filepath.Glob sorts the harness list byte by byte, where the bash runner's glob sorted by LC_COLLATE, so test-a.sh/test-B.sh run in a different order when no durations apply. |   |
| F2 | principles | Minor | stats/internal/guard/coverage_test.go:40 | The row 'a declared zero without a reason' now has exactly the same calls as 'with an empty reason' — a duplicate whose name claims a different case. |   |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 .superpowers/sdd/reproducers/0-P1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-PR1-1.sh

## Pass log

### Round 0

- roster: compact — 83
- diff size: 3394 lines, under cap; docs-only: exit 1 (first non-doc path scripts/check-guard-symlinks.sh) — roster primary+principles plus test-sanity+call-efficiency dispatched; no operator-named addition this round — the resolved list ran alone; standards passed: CLAUDE.md, AGENTS.md
