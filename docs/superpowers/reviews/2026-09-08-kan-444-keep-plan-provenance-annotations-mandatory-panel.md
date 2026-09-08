# Review panel — kan-444-keep-plan-provenance-annotations-mandatory

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | Minor | scripts/check-guard-symlinks.sh:RULE3_AWK exemption map | restoring the rule 3 exemption for check-plan-provenance.sh is caught by no test: the harness passes with the exemption line re-inserted |
| F2 | mutation | Minor | scripts/check-guard-symlinks.sh:rules 2 and 3 | deleting the shipped provenance symlink leaves check-guard-symlinks.sh green: no rule requires an invoked guard to be shipped |
| F3 | primary | Minor | .flow/project.md:## lint — check-plan-shape.sh sits beside… paragraph | the lint commentary puts the provenance guard in the project-configured family, which shipping it makes false; the sentence belongs to the task that falsified it |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 withdrawn operator standing instruction for this run: fix only easy-to-fix minors and ignore all others — closing it means adding an invoked-implies-symlinked enforcement rule, new guard surface, not a repair; the run-start guard-presence check names a missing shipped guard as the backstop
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 none — the defect is the absence of a regression case; the mutation (re-inserting EXEMPT["check-plan-provenance.sh"]) leaves scripts/test-check-guard-symlinks.sh green today, and the fix adds the case that turns it red
finding-reproducer: F2 none — no single command demonstrates an absent enforcement; verified by mutation: removing skills/flow/scripts/check-plan-provenance.sh leaves check-guard-symlinks.sh exit 0
finding-reproducer: F3 none — prose drift; the paragraph names check-plan-provenance.sh among the project-configured family, which the change makes false
