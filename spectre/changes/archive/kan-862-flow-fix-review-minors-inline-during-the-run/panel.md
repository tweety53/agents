# Review panel — kan-862-flow-fix-review-minors-inline-during-the-run

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | stats/cmd/flow/record.go:157 | three comments still justify keeping the deferred vocabulary by a dashboard feature this change removed (the deferred-Minor breakdown, numerator and rate) |   |
| F2 | primary | Minor | spectre/changes/kan-862-flow-fix-review-minors-inline-during-the-run/design.md:41 | design.md and Task 4 Step 1 still say 'the one source change', contradicting the delivered review-panel.md; Task 4 has no Correction note |   |
| F3 | principles | Minor | stats/internal/guard/unfinishedwork.go:347 | the two open-finding predicates kept identical on purpose now disagree: the panel guard reads deferred as open, the integrate gate still reads it as closed |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh

## Pass log

### Round 0

- diff size: 1765 lines, under cap; docs-only: exit 1 (first non-doc path scripts/check-panel-findings-closed.sh) — resolved roster unchanged: primary+principles; roster: compact — 6; standards: CLAUDE.md, AGENTS.md; no addition this round — the resolved list ran alone; citation check: ran scripts/check-references.sh (exit 0)
- decided under: KAN-862's own Panel re-runs rule (worktree review-panel.md) — a Minor-only round's Minors fixed inline by the parent in 977b3f4f, no slot re-run; the installed main-checkout guard, still on the Minor-deferral rule, exits 1 demanding re-runs for F1-F3, the worktree guard exits 0 FINDINGS-CLOSED
