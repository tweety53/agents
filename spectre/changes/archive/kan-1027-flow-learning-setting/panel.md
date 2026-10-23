# Review panel — kan-1027-flow-learning-setting

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow-contracts/learning.md:1 | new and reworded run-loaded prose fails the ## lint guard scripts/check-verbatim-moves.sh with 58 violations; the change carries no verbatim-moves.txt |   |
| F2 | primary | Minor | skills/flow-contracts/learning.md:19 | an unreachable store is described as a non-zero exit, but flow settings get \| jq .learning exits 0 with empty output when the store is down |   |
| F3 | primary | Minor | skills/flow-contracts/learning.md:51 | Exit 1 is the store refusing the topic — flow learning add also exits 1 when the store is unreachable, so a session would keep picking new topics against a dead store |   |
| F4 | primary | Minor | skills/flow-contracts/learning.md:45 | the topic is recorded when the note is composed, not when it is shown; a run that fails before the handoff burns one of the topic's three showings unseen |   |
| F5 | primary | Minor | skills/flow/SKILL-rationale.md:27 | SKILL-rationale.md still says an empty -reviewers is refused; the CLI now treats it as not given |   |
| F6 | primary | Minor | CLAUDE.md:65 | the flow-settings skill-index row in CLAUDE.md and AGENTS.md does not mention learning |   |
| F7 | primary | Minor | skills/flow-contracts/handoff-blocks.md:84 | the IN_PROGRESS template carries no Learning line while :110 says the run prints one, making the fewer-lines sentence false |   |
| F8 | principles | Critical | skills/flow-contracts/learning.md:1 | same defect as F1 against the Lint Fix Priority rule: check-verbatim-moves.sh fails on the change's prose | supersedes F1 |
| F9 | principles | Minor | stats/internal/api/settings.go:31 | SettingsStore widened with the two learning methods, so two fakes needed stubs to compile |   |

findings-total: 9
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 withdrawn the plan records widening SettingsStore to keep api.New unchanged; the two stubs are that choice, no defect

reproducers-total: 9
finding-reproducer: F1 .superpowers/sdd/reproducers/0-P1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-P2-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-P3-1.sh
finding-reproducer: F4 none — prose ordering
finding-reproducer: F5 .superpowers/sdd/reproducers/0-P5-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-P6-1.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-P7-1.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/0-R1-1.sh
finding-reproducer: F9 none — structural

## Pass log

### Round 0

- diff size: 1425 lines, under cap — proceed
- docs-only: exit 1 (stats/cmd/flow/learning_test.go) — resolved roster unchanged: primary+principles, plus test-sanity+call-efficiency
- roster: compact — 72; no addition this round — the resolved list ran alone; standards: CLAUDE.md, AGENTS.md

### Round 1

- F9 withdrawn: the plan records keeping api.New unchanged by widening SettingsStore (task 2 Step 2); stubs are its cost. Puzzle answer moved from <details> to an **Answer:** line — no HTML reaches the terminal (Handoff output).
