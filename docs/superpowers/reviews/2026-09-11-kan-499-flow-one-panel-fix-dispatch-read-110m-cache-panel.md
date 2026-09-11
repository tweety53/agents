# Review panel — kan-499-flow-one-panel-fix-dispatch-read-110m-cache

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | simple-reviewer | Minor | scripts/check-panel-fix-single-dispatch.sh:166 | panel-fix-01 and panel-fix-1 are two distinct base keys that parse to the same round, so a round can carry two originals and skip the chunk bound; pre-existing at the merge base, inherited by the new round loop |

findings-total: 1
finding-status: F1 deferred — pre-existing at the merge base (KAN-499 scope is the chunked contract, not key canonicalisation); follow-up-sized

reproducers-total: 1
finding-reproducer: F1 .superpowers/sdd/reproducers/0-simple-reviewer-1.sh
