# Review panel — kan-431-flow-batch-conductor-tool-calls-replace-idle

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | mutation | minor | scripts/check-dispatch-paragraphs.sh | check-dispatch-paragraphs.sh has no REPORT FILE row; deleting either new REPORT FILE blockquote (implement.md's implementer dispatch, review-panel.md's fix-subagent dispatch) still reports 5 site(s) validated, exit 0 — a surviving mutant |
| F2 | mutation | minor | scripts/check-references.sh | check-references.sh does not recognise the (**Turn discipline**, `path`) citation shape at all — corrupting the citation to a nonexistent file still reports all referenced sections resolve, exit 0; the four new Turn discipline citations have zero guard coverage |

findings-total: 2
finding-status: F1 withdrawn pre-existing guard gap, out of this change's scope (proposal.md lists check-dispatch-paragraphs.sh as unchanged); operator directed a follow-up issue instead
finding-status: F2 withdrawn pre-existing guard gap, out of this change's scope (proposal.md lists check-references.sh as unchanged); operator directed a follow-up issue instead

reproducers-total: 2
finding-reproducer: F1 none — fix requires extending scripts/check-dispatch-paragraphs.sh's paragraph table, which is outside this task's declared Files and this change's proposed unchanged-scripts list; operator decision pending
finding-reproducer: F2 none — fix requires extending scripts/check-references.sh's is_associated shape recognition, outside this task's declared Files; operator decision pending
