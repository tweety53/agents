# Review panel — kan-379-resolve-project-md-keys-through-flow-scripts

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | low | scripts/lib/project-section.sh:41 | test-project-get.sh's case 5 uses command substitution which itself strips trailing newlines, so removing project_section's own trailing-blank-line trim is a surviving mutant — the header comment's 'trailing blank lines removed' contract is unverified |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 none — reproducer rejected by check-panel-reproducers.sh (shell metacharacters); recorded unverifiable, operator directed to proceed to fix regardless
