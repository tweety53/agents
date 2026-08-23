# Review panel — kan-302-panel-code-review-slot-hangs-on-fork

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Task 3 review | Major | skills/myflow-do/SKILL.md | Second-breach prompt cited skills/myflow-contracts/operator-prompts.md but rendered no prompt, no (recommended) option and no silent default, and closed with a sentence contradicting the contract outright. Fixed: prompt rendered with Stop the run as the recommended default; plan, design and delta spec corrected alongside. |
| F2 | Task 4 review | Critical | openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/tasks.md | The plan asserted openspec/changes/ is inside check-normative-inventory.sh's owned corpus; it is not. owned-corpus.sh scopes to skills rules openspec/specs commands commands-claude .myflow plus root-level .md, and excludes openspec/changes/ deliberately so a during-a-change inventory comparison stays valid. Both tasks.md and verification.md corrected. |
| F3 | Task 4 review | Major | openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/verification.md | verification.md claimed the budget-estimate correction was still pending in proposal.md and design.md; it had already landed. Section rewritten to record the correction as made. |
| F4 | Task 4 review | Minor | openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/verification.md | Cited a non-existent stats/.gitignore; the entry is stats/internal/web/dist/ at .gitignore line 23. Citation corrected. |
| F5 | Task 4 re-review | Minor | openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/tasks.md | F2's fix left the next bullet saying the skill's sentences appear in the inventory 'exactly as the delta specs do', contradicting the corrected bullet in the same file. Reworded to state skills/ is a corpus scope root. |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 grep -n -A5 "shaped per \*\*Operator prompts\*\*" skills/myflow-do/SKILL.md
finding-reproducer: F2 grep -n OWNED_CORPUS_SCOPE_DIRS scripts/lib/owned-corpus.sh
finding-reproducer: F3 sed -n '69,77p' openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/proposal.md
finding-reproducer: F4 ls stats/.gitignore; grep -n internal/web/dist .gitignore
finding-reproducer: F5 sed -n '318,340p' openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/tasks.md
