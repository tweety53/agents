# Review panel — kan-263-myflow-forbid-backgrounded-builds-in-the-shared

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Major | scripts/check-dispatch-paragraphs.sh:141 | min_blocks=2 for the two new foreground sites is not exercised at threshold — no test case covers exactly one correct block present with the second missing, so a dispatcher dropping the second required occurrence goes undetected |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 none — needs a new fixture test case, not a single runnable one-liner against the real tree
