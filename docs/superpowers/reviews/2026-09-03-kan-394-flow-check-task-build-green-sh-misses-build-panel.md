# Review panel — kan-394-flow-check-task-build-green-sh-misses-build

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Minor | scripts/lib/plan_grammar.py:163 | no test pins the required whitespace between **Build:** and the keyword; widening \s+ to \s* silently accepts **Build:**green as well-formed |
| F2 | Bugbot | Minor | scripts/check-task-build-green.py:265 | no test pins that the malformed-tag continue prevents a duplicate "no **Build:** tag" violation on the same line |
| F3 | Bugbot | Minor | scripts/check-task-commit-fields.py:588 | no test in check-task-commit-fields.py suite covers a Squash-with partner whose own **Build:** line is malformed |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 none — requires source mutation, see panel-report-0-bugbot.md Finding 1
finding-reproducer: F2 none — requires source mutation, see panel-report-0-bugbot.md Finding 2
finding-reproducer: F3 none — requires source mutation, see panel-report-0-bugbot.md Finding 3
