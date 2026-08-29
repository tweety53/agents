# Review panel — kan-367-check-task-commit-fields-sh-refuses-to-run-while

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Important | scripts/check-task-commit-fields.sh:151-164 | The named-change fix-sibling resolution duplicates the existing glob-path's highest-numbered-fix-sibling algorithm instead of sharing it |
| F2 | Principles | Minor | scripts/check-task-commit-fields.sh:175-178 | A third near-identical copy of the PARENT_SHA-conditional exec dispatch snippet, added with no note that the duplication was a considered tradeoff |
| F3 | Bugbot | Medium | scripts/check-task-commit-fields.sh:159 | The named-change branch's fix-sibling loop has no test covering a numbered -fix-N sibling directory that lacks tasks.md; dropping its existence check is a surviving mutant |
| F4 | Bugbot | Low | scripts/check-task-commit-fields.sh:138 | No test asserts the change-name validator actually rejects a path-traversal-shaped or slash-containing value; widening the char-class survives the suite |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 none — a duplication finding has no failing-command reproducer; grep -n 'CHOSEN_N=-1' scripts/check-task-commit-fields.sh shows both occurrences
finding-reproducer: F2 none — a duplication finding has no failing-command reproducer; grep -n 'exec python3 "$PYTHON_GUARD"' scripts/check-task-commit-fields.sh returns 5 matches
finding-reproducer: F3 none — surviving mutant, not a failing command against unmutated code; construct <name>-fix-1/tasks.md plus <name>-fix-2/ (no tasks.md) and confirm no test asserts <name>-fix-1 is still chosen
finding-reproducer: F4 none — surviving mutant, not a failing command against unmutated code; add a test asserting check-task-commit-fields.sh ... "../x" and "a/b" as change-name both exit 2 naming 'invalid change name'
