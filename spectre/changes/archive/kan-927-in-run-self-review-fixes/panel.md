# Review panel — kan-927-in-run-self-review-fixes

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow-contracts/finish-contract-run1.md:306 | an archived re-run with new work runs the whole self-review pass again, recording every filed and declined finding a second time and re-asking the filing question; /flow-fast skips the pass instead |   |
| F2 | primary | Minor | skills/flow-contracts/finish-contract-run1.md:118 | the archived re-run base no longer recognises the old self-review context bundle subject, which git.go reservedShapes still honours |   |
| F3 | primary | Minor | spectre/changes/kan-927-in-run-self-review-fixes/design.md:78 | the live-check curl uses date-only from/to, which the API rejects as not RFC 3339 |   |
| F4 | principles | Minor | stats/internal/api/stats.go:430 | the no-token-measurement, no-model rule lives in isHealthView and again as an inline self-review special case |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 none — behaviour correct and tested; DRY only

## Pass log

### Round 0

- reachability: still reproduces — finish-contract-run1.md "Self-review is always deferred: no reasoning pass" (integrate saves a bundle only); stats/web has no self_review reference; pipeline.md in-run fixes recorded in narrative only; store: 1 fixed row of 92
- auto-resolved: where does the in-run pass run? → integrate run 1, after the archive commit, in place of saving the bundle (recommended)
- auto-resolved: where do self-review fixes land? → on the change branch when the change is in the agents repo, else on an agents worktree branch landed by its default route (recommended)
- auto-resolved: dispatch pairs for fixes? → fixer flow-medium opus (sonnet for trivial), first review flow-medium opus, re-reviews flow-medium sonnet (recommended; global rule keeps first-pass review on opus)
- auto-resolved: verify scope? → the ## lint lines the touched files need plus targeted tests; no full suite (recommended)
- auto-resolved: /flow-fast too? → yes, same in-run pass at its 5. Verify (recommended)
- auto-resolved: UI shape? → new stats view self-review: one row per finding, outcome filter, commit link from the daemon checkout origin, per-change counts derived client-side (recommended)
- auto-resolved: convergence confirm → approve the design and move on (decisions: recommended)
- auto-resolved: plan review gate → Yes (decisions: recommended)
- roster: compact — 23
- diff size 1212 lines, under cap; docs-only exit 1 (scripts/check-cleanup-complete.sh); roster dispatched: primary+principles (opus/medium); no operator-named addition this round — the resolved list ran alone; standards: CLAUDE.md, AGENTS.md
- F1-F3 reproducers re-authored (premises cited lines the fix changed); prove-reproducer.sh held both legs for each
fix-mutation: stats/internal/api/health.go — dropped viewSelfReview from isHealthView — TestHealthViewsAreNeverUnmeasured/self-review
fix-mutation: skills/flow-contracts/finish-contract-run1.md — none — prose rule, checked by the reproducers and guards
fix-mutations-total: 2

### Round 1

- primary+principles re-run on fix-round-1.diff (rerun pair sonnet/low): F1-F4 verified fixed, no new finding
