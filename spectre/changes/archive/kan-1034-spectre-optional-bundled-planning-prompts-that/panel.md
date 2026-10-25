# Review panel — kan-1034-spectre-optional-bundled-planning-prompts-that

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | spectre:.claude/skills/spectre-planning/SKILL.md:344 | The vendored skill can't be used without flow: it cites 9 flow files, two as required Load targets, and orders flow-only guards and stages. |   |
| F2 | primary | Important | skills/flow/verify-fix-loop.md:55 | Two citations point at paragraphs no longer in the cited file (verify-fix-loop.md:55, document-fix.md:126 cite brainstorm-planner.md section D). |   |
| F3 | primary | Minor | skills/flow/brainstorm-planner.md:69 | 'above' now points at nothing: silence default above, HARD GATE above, The checklist above — that text moved to the new skill. |   |
| F4 | primary | Minor | skills/spectre-planning/SKILL.md:39 | HARD GATE makes ExitPlanMode the approval only for /flow and /flow-plan; the Plan mode section makes it the approval on any Claude Code session. |   |
| F5 | principles | Important | skills/spectre-planning/SKILL.md:433 | Separation of concerns: the generic skill depends on flow — 24 lines citing /flow, skills/flow* paths or flow.verify. | supersedes F1 |
| F6 | principles | Minor | skills/spectre-planning/SKILL.md:10 | ### Plan mode is an h3 directly under the h1, with no parent ## section. |   |
| F7 | mutation | Important | stats/internal/guard/loadsets.go:74 | Surviving mutant: dropping spectre-planning/SKILL.md from the planning load set leaves the guard package green — the fixture file is 0 bytes. |   |
| F8 | primary | Minor | skills/spectre-planning/SKILL.md:271 | The fix round dropped the file's final newline, in both copies. |   |
| F9 | primary | Minor | skills/flow/brainstorm-planner.md:290 | section **C**'s four plan sections now points to the planner's own section C, which no longer holds them. |   |
| F10 | primary | Minor | skills/flow/brainstorm-planner.md:548 | Two blank lines were left before the panel.dispatches paragraph. |   |

findings-total: 10
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed

reproducers-total: 10
finding-reproducer: F1 .superpowers/sdd/reproducers/r1-P1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/r1-P2-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/r1-P3-1.sh
finding-reproducer: F4 none — a contradiction between two prose sentences
finding-reproducer: F5 .superpowers/sdd/reproducers/r1-S1-1.sh
finding-reproducer: F6 none — heading style
finding-reproducer: F7 .superpowers/sdd/reproducers/r1-F1-1.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/r1-PR1-1.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/r1-PR2-1.sh
finding-reproducer: F10 none — cosmetic blank line

## Pass log

### Round 0

- dispatched: panel-0-primary+principles on opus/medium; panel-0-test-sanity+call-efficiency on sonnet/low; panel-0-mutation+failure-modes on sonnet/low
- verdict: fix (0 Critical, 4 Important)

### Round 1

- dispatched: panel-1-primary+principles+mutation on opus/low
- verdict: clean
