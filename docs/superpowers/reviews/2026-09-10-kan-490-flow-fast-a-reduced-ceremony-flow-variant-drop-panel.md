# Review panel — kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | major | skills/flow-fast/SKILL.md:108-112 | guard-resolution prose contradicted the shipped skills/flow-fast/scripts/ symlink layout — fixed |
| F2 | primary | major | spectre/changes/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop/tasks.md task 2 | Files: field didn't declare check-stage-mark-calls.sh + 9 symlinks task 2's commit actually touched — corrected in tasks.md |
| F3 | primary | major | spectre/changes/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop/tasks.md task 4 | Files: field didn't declare skills/flow-fast/scripts/project-get.sh — corrected in tasks.md |
| F4 | primary | major | spectre/changes/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop/tasks.md task 6 | Files: field didn't declare check-stage-mark-calls.sh — corrected in tasks.md |
| F5 | primary | minor | spectre/changes/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop/proposal.md | plan's own design work never called out the skill-local guard-symlink cost of per-skill-directory guard resolution |
| F6 | simple-reviewer | major | skills/flow-fast/brainstorm.md flow.create-artifacts | missing flow stage begin — fixed |
| F7 | simple-reviewer | major | skills/flow-fast/implement.md flow.isolate-workspace | mismarked/missing begin-end pairs — fixed |
| F8 | simple-reviewer | major | skills/flow-fast/review.md flow.review-panel | missing begin marks — fixed |
| F9 | simple-reviewer | major | skills/flow-fast/finish.md flow.preflight | key never marked at all — fixed |

findings-total: 9
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 deferred per /flow-fast's own Minor rule
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed

reproducers-total: 9
finding-reproducer: F1 grep -n 'carries its own symlink' skills/flow-fast/SKILL.md
finding-reproducer: F2 git show --stat 58604fd
finding-reproducer: F3 git show --stat 872f0c1
finding-reproducer: F4 git show --stat b537b68
finding-reproducer: F5 grep -i 'guard-symlink' spectre/changes/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop/proposal.md spectre/changes/kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop/design.md
finding-reproducer: F6 grep -n 'flow.create-artifacts' skills/flow-fast/brainstorm.md
finding-reproducer: F7 grep -n 'flow.isolate-workspace\|flow.load-context\|flow.document-fix\|flow.sdd-tdd' skills/flow-fast/implement.md
finding-reproducer: F8 grep -n 'flow.review-panel\|flow.verify\|flow.stage-diff\|flow.run-instructions\|flow.write-in-progress' skills/flow-fast/review.md
finding-reproducer: F9 grep -n 'flow.preflight' skills/flow-fast/finish.md
