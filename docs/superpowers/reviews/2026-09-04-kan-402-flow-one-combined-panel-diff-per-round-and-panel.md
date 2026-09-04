# Review panel — kan-402-flow-one-combined-panel-diff-per-round-and

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | Minor | skills/flow/review-panel.md:275-277 | The CONTEXT BUNDLE paragraph still names a single <abs-worktree>/.superpowers/sdd/dispatch-context.md, while the sibling relocation-comparison and CITATION CHECK paragraphs right next to it were correctly rewritten to enumerate 'for every worktree whose ... exists, one path each' — dispatch-context.md is per-worktree too (the rebuild step already sits inside the once-per-worktree loop), so it should be named the same way. |
| F2 | bugbot | Medium | scripts/gather-dispatch-context.sh:381-388 | The guard refusing scoping when tasks.md is absent/refused checks only "${FOUND_LABELS[$last]}" != "tasks.md"; no test exercises task-ids given while tasks.md is missing/refused, so this branch is a surviving mutant (dropping the check left all cases passing) rather than a proven-covered guard. |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 none — a prose-consistency defect, not independently executable
finding-reproducer: F2 none — this is a coverage gap, not a live defect; the current code is still correct
