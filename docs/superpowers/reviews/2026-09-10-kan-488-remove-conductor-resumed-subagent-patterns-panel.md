# Review panel — kan-488-remove-conductor-resumed-subagent-patterns

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | code-review-low | Critical | skills/flow/brainstorm.md:174 | mv command missing / between <name> and .superpowers-sdd-decision.json |
| F2 | primary | Critical | commands/flow-plan.md:8 | stale reference to deleted researcher-subagent mechanism and removed section name |
| F3 | primary | Important | skills/flow-contracts/model-policy.md:22 | still names conductor and planner as active dispatch roles |
| F4 | primary | Minor | spectre/changes/kan-488-remove-conductor-resumed-subagent-patterns/tasks.md:213 | corrupted orphaned fragment left over from an earlier edit |
| F5 | primary | Minor | scripts/check-contract-budget.sh:117 | rename swept unrelated historical myflow-research into fabricated myflow-plan |
| F6 | mutation | Minor | scripts/test-check-dispatch-paragraphs.sh:257 | surviving mutant: no fixture pinned implement.md TOOLS/HANDSHAKE min-blocks at exactly 1 |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 .superpowers/sdd/reproducers/0-code-review-low-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-5.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-mutation-1.sh
