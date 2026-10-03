# Review panel — kan-873-agents-port-the-next-five-slowest-bash-guards-to

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/guard/breakandprove.go:347 | a --clean command that backgrounds a child stalls each leg until that child exits. |   |
| F2 | primary | Minor | stats/internal/guard/prepareworkspace.go:55 | the sibling guard is looked up in $FLOW_GUARD_REPO_ROOT/scripts, not beside the shim. |   |
| F3 | primary | Minor | docs/prompt-audit-2026-09-29/audit-finish.md:573 | cites scripts/test-check-self-review-report.sh, which this change deletes. |   |
| F4 | principles | Minor | stats/internal/guard/selfreviewreport.go:425 | DRY: re-states crSort's exec'd-sort setup (references.go:280-306). |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-P1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-P2-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-P3-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-R1-1.sh

## Pass log

### Round 0

- auto-resolved: convergence confirm → approve the design and move on
- auto-resolved: which five guards → the five slowest by 3-run median harness time (KAN-842 exclusions)
- auto-resolved: Proceed to implementation? → Yes
- roster: compact — 16
- diff size: 8345 lines, over cap — proceeded automatically
- docs-only: exit 1 — first non-doc path scripts/aside-planning-artifacts.sh; roster primary+principles dispatched
- no addition this round — the resolved list ran alone.
- standards: CLAUDE.md, AGENTS.md
- panel-fix-0 (opus/medium) fixed F1-F4 in 509d807c; reproducers re-authored (premise lines pinned rewritten lines; P3 body could never pass) and proved both directions against 2c37d289 via prove-reproducer.sh; fix-round-0 diff 2c37d289..509d807c
fix-mutation: stats/internal/guard/breakandprove.go — clean.Stdout set back to the EPIPE wrapper stdout — TestBreakAndProveCleanBackgroundChild
fix-mutation: stats/internal/guard/prepareworkspace.go — reverted to the FLOW_GUARD_REPO_ROOT/scripts lookup and REPO_ROOT export — TestPrepareWorkspaceShimSibling
fix-mutation: stats/internal/guard/selfreviewreport.go — crSortSep NUL separator flipped to newline — TestCheckSelfReviewReport/a_report_name_carrying_a_newline_stays_one_path
fix-mutation: stats/internal/guard/references.go — crSort's unique branch dropped -u — TestPanelTouchedPathsParity/touched:_en_US.UTF-8_collation
fix-mutation: docs/prompt-audit-2026-09-29/audit-finish.md — none — a citation edit with no executable behaviour; check-references.sh exits 0
fix-mutations-total: 5

### Round 1

- primary re-ran alone on sonnet/low (rerun pair) on fix-round-0.diff + F1 site: F1 fixed, nothing new; principles raised only a Minor so not re-run
- primary re-dispatched under panel-1-primary-retry with -slot primary: the first round-1 dispatch row lacked its slot tag, so the closed guard could not see it; verdict F1 fixed, nothing new
