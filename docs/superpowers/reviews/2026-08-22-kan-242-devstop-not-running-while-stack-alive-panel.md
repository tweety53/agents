# Review panel — kan-242-devstop-not-running-while-stack-alive

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | gymie-Principles | Minor | gradle/dev-lifecycle.gradle.kts:2156 | The Undetermined message-building block shares the listOfNotNull(...takeIf...) shape with the survivors block at :2129. The reviewer judges WET correct here (different knowledge, an extracted helper would be the wrong abstraction) and raises it only because the diff commentary did not call it out. No fix proposed. |
| F2 | agents-Principles | Minor | scripts/check-worktree-processes.sh:69 | set -uo pipefail omits -e where every sibling finish guard uses -euo pipefail. The omission is load-bearing: under set -e the shell aborts at LSOF_OUT="$(lsof ...)" before LSOF_RC=$? on the next line is read, so die() could never name the reason. Least Astonishment: the reason was nowhere stated. Fixed by a header comment above the set line. |
| F3 | gymie-CodeReviewLow | Important | gradle/dev-lifecycle.gradle.kts:2153 | The Undetermined branch message opens "this run signalled what it tracks", but stopOutcome returns Stopped whenever pidRecord is Recorded.Value or portHadListener, so reaching Undetermined guarantees trackedPid==null AND orphans.isEmpty() — killPidTree was never called. Nothing was signalled. The message also contradicts its own clauses, which say nothing was signalled. Phrasing copy-pasted from the survivors branch, which is only reachable after a real kill attempt. |

findings-total: 3
finding-status: F1 withdrawn the operator judged WET correct here: the shape shared with the survivors block at :2129 is coincidental similarity, not shared knowledge, and a helper for two call sites needing different wording and clause counts would be the wrong abstraction — the raising reviewer argued the same and assessed the diff principles-compliant
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 none — judgment call on a defensible duplication, not a rule violation
finding-reproducer: F2 none — the defect was an undocumented decision, and no command demonstrates a missing comment; the mechanism it concerns is shown in the finding note
finding-reproducer: F3 none — logically entailed by the guard at StopOutcome.kt:103; a live reproduction would require starting the dev stack, which review must not do
