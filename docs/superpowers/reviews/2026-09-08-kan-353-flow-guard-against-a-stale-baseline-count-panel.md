# Review panel — kan-353-flow-guard-against-a-stale-baseline-count

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | minor | scripts/test-check-baseline-fresh.sh:688 | the freshness comparison's boundary is unpinned: flipping < to <= survives the whole harness because no case sets a declared directory's mtime exactly equal to the commit epoch |
| F2 | principles | minor | scripts/check-baseline-fresh.py:178 | a typo'd or prose line in the baseline results dirs body is silently dropped when a valid path line also exists, so a declared directory can go unchecked with no signal |
| F3 | code-review-low | minor | scripts/test-check-baseline-fresh.sh:430 | the harness header still describes the mutation cases as asserting the fixture passes, while the corrected helper asserts the tagged check's message vanishes |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F3 none — comment-only drift; no behavior to reproduce
