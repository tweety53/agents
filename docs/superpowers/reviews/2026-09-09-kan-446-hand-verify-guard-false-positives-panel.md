# Review panel — kan-446-hand-verify-guard-false-positives

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | Minor | scripts/check-base-moved.sh:42 | the hand-verification paragraph says to recount with the bare <base-ref> argument, but the guard counts against its substituted EFFECTIVE_REF, so a hand-verifier can recount a different ref than the verdict was computed against |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 none — the defect is imprecise prose, not observable behavior
