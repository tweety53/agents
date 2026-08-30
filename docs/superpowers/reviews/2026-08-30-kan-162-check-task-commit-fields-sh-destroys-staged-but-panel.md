# Review panel — kan-162-check-task-commit-fields-sh-destroys-staged-but

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Major | scripts/check-task-commit-fields.py:1102-1112 | the stash-pop-conflict finally re-raise discards any exception already propagating from the protected body (e.g. a revert/reset failure inside check_regression/check_baseline), so main() prints only the stash-list recovery message and the operator never sees the actual root-cause failure |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 none — requires importing the module and injecting a synthetic in-flight exception plus a real stash-pop conflict; see report
