# Review panel — kan-312-myflow-per-task-review-and-the-panel-duplicate

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Major | scripts/check-panel-docs-only.sh:27-70 | check-panel-docs-only.sh:27-70 is a line-for-line duplicate of check-panel-citation-trigger.sh:19-68 — GIT_BIN resolution, arg/worktree/merge-base validation, and COMMITTED/STAGED/UNSTAGED collection are copy-pasted rather than factored into scripts/lib/, which this repo already uses for exactly this purpose |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 diff <(sed -n '19,68p' scripts/check-panel-citation-trigger.sh | sed 's/citation-trigger//g') <(sed -n '27,70p' scripts/check-panel-docs-only.sh | sed 's/docs-only//g')
