# Review panel — kan-395-flow-fingerprint-the-served-bundle-before-a

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | Minor | scripts/check-visual-verification.sh:359 | the error message's vocabulary list can drop 'fingerprint' with no test failing — no case asserts the message names all four vocabulary words |
| F2 | primary | Important | skills/flow/verify-and-handoff.md:215 | step 10's continuation lines and its push-command fence are still indented 3 spaces (correct width for the old single-digit '9.' marker) instead of 4 (required for the new two-digit '10.' marker), breaking CommonMark list nesting for that fence — item 11 got the equivalent re-indent right |
| F3 | primary | Minor | skills/flow/verify-and-handoff.md:157 | the verifier-dispatch paragraph's second line runs 108 characters against ~96-99 for its neighbors — a minor prose-wrap inconsistency from splicing 'fingerprint' into an existing wrapped sentence |
| F4 | mutation | Minor | scripts/test-check-visual-verification.sh:797 | case 29's assert_out_contains for backtick-verify and backtick-capture pass via the separate required-command-absent violation line, not the vocabulary-closed message itself — only setup and fingerprint are genuinely pinned by that assertion, since neither is a required command |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 none — mutation only, no single runnable command isolates the message-text drift
finding-reproducer: F2 none — markdown list-nesting defect, no single command isolates it; visually confirm 4-space continuation indent under item 10, matching item 11's
finding-reproducer: F3 none — cosmetic wrap width, not a correctness or lint issue (all check-*.sh guards pass)
finding-reproducer: F4 none — confirmed by reading scripts/check-visual-verification.sh:435-436 (require_nonempty '#CMD' 'verify'/'capture' fire independent 'is absent — this setting is required' violations whose text also contains the word), not a single isolating command
