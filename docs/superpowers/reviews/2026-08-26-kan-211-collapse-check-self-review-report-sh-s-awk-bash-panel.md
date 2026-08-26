# Review panel — kan-211-collapse-check-self-review-report-sh-s-awk-bash

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review (low) | Important | scripts/check-self-review-report.sh:404 | An unreadable report file is misclassified as a content violation. The old code checked awk exit status and died with exit 2; the new "done < \$f" redirection fails silently under set -uo pipefail, the loop body never runs, and the file is recorded as coverage 0 — surfacing as an undeclared-zero violation at exit 1 instead of the documented exit 2 "cannot answer". |
| F2 | Primary | Minor | scripts/check-self-review-report.sh:304 | idx=-1 is initialized before the per-file loop but always overwritten alongside cur before its first read. Dead initialization left over from mirroring LAST_IDX sentinel. |
| F3 | Primary | Important | scripts/check-self-review-report.sh:318 | The bash-3.2 justification for [[ -r "$f" ]] over a real open-test is incomplete. Ruling out exec {fd}< "$f" only excludes the dynamic fd-allocation form (bash 4.1+). The fixed-descriptor form exec 3<"$f" \|\| die is bash-3.2-safe and performs a genuine open() test rather than [[ -r ]]$(printf s) access()-based permission-bit check, removing a TOCTOU window and covering ACL/immutable-attribute cases the permission bit misses. |
| F4 | Primary | Minor | scripts/check-self-review-report.sh:318 | The precheck restores the failed-open case but not the old awk-exit-status check coverage of a mid-read I/O failure. A file that opens but fails partway through reading no longer surfaces as exit 2; it classifies whatever partial content was read. Not recorded in design.md Known limits, which covers NUL bytes only. |
| F5 | Code review (low) | Minor | scripts/check-self-review-report.sh:318 | The comment claiming the read loop $(printf s) exit status cannot discriminate a failed redirection from a normal EOF read is factually wrong, verified on bash 3.2.57 and 5.3.15 and by instrumenting the real loop against seven fixtures. A failed redirection aborts the compound command before the body runs, giving 1; every normal completion gives 0. The precheck therefore introduces a TOCTOU window the pre-collapse guard did not have, and checks a proxy rather than the operation, against the principle the header itself states. |
| F6 | Principles | Important | scripts/test-check-self-review-report.sh:844 | Single Source of Truth. Case 24 comment restates design.md Known limits nearly point-for-point: read I/O-error-versus-EOF indistinguishability, the find -maxdepth 1 -type f exclusions, and the hardware-error residue. design.md is canonical for that reasoning; the same file header cites the limit in one sentence instead of restating it. Replace with a one-sentence pointer in the header style. |
| F7 | Principles | Minor | scripts/test-check-self-review-report.sh:26 | Least Astonishment. Every other mutation-matrix row keeps its name on one line and wraps the case-number value at column 47; the new per-report read status row wraps the name instead, landing its value at column 46 and breaking the table alignment. |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed

reproducers-total: 7
finding-reproducer: F1 none — the fixture needs chmod 000 on a file, which no single command free of shell metacharacters expresses; the fix adds a harness case that makes it runnable
finding-reproducer: F2 none — cosmetic only; idx is never dereferenced while cur is empty, so no command distinguishes the two states
finding-reproducer: F3 none — a design/reasoning gap (dynamic {fd} versus fixed-fd), not a currently-triggerable behavioural bug; case 24 passes under the chosen implementation
finding-reproducer: F4 none — a mid-read I/O error, as opposed to a failed open, cannot be simulated with chmod or ordinary fixture files; no runnable reproducer exists in this environment
finding-reproducer: F5 none — demonstrating this needs an ad hoc bash probe and an instrumented copy of the script outside the worktree; not expressible as one worktree-relative command without leaving added files in the tree
finding-reproducer: F6 sed -n 844,852p scripts/test-check-self-review-report.sh
finding-reproducer: F7 sed -n 24,30p scripts/test-check-self-review-report.sh
