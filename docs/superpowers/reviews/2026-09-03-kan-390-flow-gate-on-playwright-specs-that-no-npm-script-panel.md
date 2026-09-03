# Review panel — kan-390-flow-gate-on-playwright-specs-that-no-npm-script

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Minor | scripts/check-spec-reach.sh:71 | dropping the explicit [ -d "$ROOT" ] guard is not caught by test-check-spec-reach.sh — without it, cd would fail under set -e with a raw shell exit 1 and no clean message, instead of the documented exit 2 |
| F2 | Bugbot | Minor | scripts/check-spec-reach.sh:192 | grep -qxF's exact-line match (vs -qF's substring match) has no test asserting it is necessary |
| F3 | Bugbot | Minor | scripts/check-spec-reach.sh:147 | dropping the NODE_RC -ne 0 diagnostic check is not caught by the harness — the exit code stays 2 via the downstream LIST_RC check, but the error message degrades to nonsense (node's error text is treated as a script name) |
| F4 | Bugbot | Minor | scripts/check-spec-reach.sh:106 | the awk section-boundary test (hlevel > SEC_LEVEL) has zero coverage for a sibling ## section following ## visual verification — a real behavioral divergence if that boundary regresses, uncaught today |
| F5 | Bugbot | Minor | scripts/check-spec-reach.sh:114 | a duplicate regression checkout row inside one ## visual verification section is not rejected as ambiguous, unlike a duplicate heading — asymmetric with the guard's own cannot-answer policy, and untested |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/repro/repro-m3-root-guard.sh
finding-reproducer: F2 none — no practical exploit: two distinct absolute paths under one checkout root cannot substring-collide under grep -qF vs -qxF, and Playwright only ever reports real absolute paths, never attacker-crafted strings
finding-reproducer: F3 .superpowers/sdd/repro/repro-m9-node-rc.sh
finding-reproducer: F4 .superpowers/sdd/repro/repro-m13-awk-boundary.sh
finding-reproducer: F5 .superpowers/sdd/repro/repro-m14-duplicate-row.sh
