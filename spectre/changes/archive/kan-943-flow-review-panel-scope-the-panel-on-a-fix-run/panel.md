# Review panel — kan-943-flow-review-panel-scope-the-panel-on-a-fix-run

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow/review-panel.md:346 | Shared INDEPENDENT PASSES / ENTRY CONTEXT / mutation entry-context paragraphs still point every slot at final-review.diff, contradicting the append scope's late-fix.diff read. |   |
| F2 | primary | Important | skills/flow/review-panel-late-fix.md:73 | 'the docs-only guard on the same delta' can reduce pass 1 to primary on a docs-only delta of a code branch, an unrecorded reduction the next re-run (merge-base docs-only check) then contradicts. |   |
| F3 | primary | Minor | skills/flow/review-panel-late-fix.md:75 | the late-fix.md sentence '...whose reduction to primary still applies. The append scope is recorded...' is now one over-long line |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/1-append-scope-prompt-names-final-diff-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/1-append-scope-docs-only-delta-2.sh
finding-reproducer: F3 none — line wrapping only

## Pass log

### Round 0

- auto-resolved: panel scope on a fix-run append → since-close delta, full decided roster
- auto-resolved: applies to any fix run failing only late-fix conditions 3/4 → yes
- auto-resolved: verdict mechanism → exit 3 of check-late-fix-trigger
- auto-resolved: void on Critical/Important under append scope → no
- auto-resolved: convergence confirm → approve the design and move on
- auto-resolved: Proceed to implementation? → Yes

### Round 1

- diff size: 222 lines, under cap
- docs-only: exit 1 — scripts/check-late-fix-trigger.sh; roster dispatched: primary+principles
- roster: compact — 67
- no addition this round — the resolved list ran alone.
- standards passed: CLAUDE.md, AGENTS.md
- panel-fix-1 inline (parent): F1, F2 fixed; diff fix-round-2.diff
fix-mutation: stats/internal/guard/renderslotprompt.go — named() returns b unchanged for every kind — TestRenderSlotPrompt/-diff_late-fix, -diff_delta, -diff_fix-round
fix-mutation: skills/flow/review-panel-late-fix.md — none — prose contract — reproducer 1-append-scope-docs-only-delta-2.sh flips 1→0 on the reworded entry-check sentence
fix-mutation: skills/flow/review-panel.md — none — prose lead-in, no executable behaviour
fix-mutation: scripts/render-slot-prompt.sh — none — header comment, no executable behaviour
fix-mutations-total: 4
