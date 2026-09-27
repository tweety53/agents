# Review panel — kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Minor | stats/internal/guard/panelexitcontract.go:386 | a declared-but-unasserted premise escapes both enforcement points — pcPremiseAudit only checks declarations resolve while the body assertion is prose-only — so a fix renaming the target still produces the vacuous green the proposal says can no longer happen |   |
| F2 | primary | Minor | skills/flow/review-panel.md:977 | the exit-1 one-disposition-per-class enumeration was not extended with the declared-but-unresolvable-premise class this same diff adds — the parent reading exit 1 has no stated disposition for it |   |

findings-total: 2
finding-status: F1 deferred — the design accepted this residual: the guard audits declarations, the body assertion stays an authoring rule
finding-status: F2 deferred — prose enumeration in the guard-section description lacks the new violation class

reproducers-total: 2
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh

## Pass log

### Round 0

- base moved 5 commits with overlap on skills/flow/review-panel.md; operator chose rebase; clean rebase onto 675a55a9, re-check CLEAR
- roster: compact — 12
- diff-size 594 under cap (single worktree)
- docs-only exit 1 — first non-doc path scripts/check-panel-reproducer-exit-contract.sh; resolved roster runs
- no addition this round — the resolved list ran alone
- agents ran: primary+principles one bundle on glm-5.3-flash/high; findings F1 (primary+principles, deduped), F2 (primary) — all Minor, none Critical/Important; deferral default applied, no fix round, no re-run
