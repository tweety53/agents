# Review panel — kan-843-agents-remove-cursor-codex-harnesses-myflow-era

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/store/migrations/0031_drop_legacy_shapes.sql:41 | The pricing backfill treats collapsed = 5m as published-flat, but 0007 copied the collapsed column into the 5m rate on every older row, so 0031 invents a 1h rate equal to 5m on every pre-0007 row: 1h cache writes that used to refuse now price at the 5m rate, understating cost. |   |
| F2 | primary | Important | skills/flow/brainstorm-planner.md:18 | The verify-the-defect-still-exists step now fires only on flow-fix/flow-cost labels, while Jira issues filed under myflow- labels were never relabelled, so a /flow run on one skips the check. |   |
| F3 | principles | Minor | stats/internal/store/migrations/0031_drop_legacy_shapes.sql:41 | glm-5.3-flash's 1h rate has two writers (the startup seed upsert and 0031's backfill), and the backfill's collapsed = 5m rule is a second definition of flat beside the pricing code's 1h = 5m. |   |
| F4 | principles | Minor | scripts/test-setup.sh:945 | The nothing-to-render group now seeds AGENTS.md and lost the no-AGENTS.md-is-created assertion, so a mutant that creates AGENTS.md in a project with no config passes the harness. |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 withdrawn premise unsupported: Jira holds no open myflow-fix/myflow-cost issue (only KAN-264, myflow-automation, which the check never covered); new issues carry flow- labels
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-2.sh

## Pass log

### Round 0

- roster: compact — 33
- diff size: 5528 changed lines, over cap — proceeded unasked; docs-only: exit 1 (.gitignore) — full roster primary+principles; no operator-added slot this round — the resolved list ran alone; base rebased onto 4bb54bd5 (operator chose Rebase; KNOWN-BUGS.md append conflict resolved keeping both sides on operator instruction); standards: CLAUDE.md, AGENTS.md
- F2 withdrawn on primary-source check: JQL labels in (myflow-*) AND statusCategory != Done returns only KAN-264 (myflow-automation). F1: operator chose "Drop the backfill" — design decision stats-legacy-migrate-then-remove superseded.
- F4 bounced once to principles: malformed premise (4bb54bd5: prefix); repaired, now demonstrated

### Round 1

- panel-fix-1 (opus/medium) fixed F1,F3,F4 in f9e8f326+9e208a61; F3,F4 reproducers flipped (pinned); F1 pinned re-run refused ambiguous (premise names the deleted backfill line) — routed to primary for re-authoring; plan-unchanged flagged the fixer's PLAN FIELDS Tests: edit to task 2, hand-verified
- primary re-run (opus/low, fix-round-1.diff): F1 fixed, no new finding; reproducer re-authored and proven both legs (sha 57b44888…); principles not re-run — raised only Minors, both fixed
fix-mutation: stats/internal/store/migrations/0031_drop_legacy_shapes.sql — re-inserted the 1h backfill UPDATE — TestMigration0031RewritesLegacyRows
fix-mutation: scripts/test-setup.sh — setup.sh install_project_standards creates AGENTS.md when absent — no AGENTS.md is created for a project with no config
fix-mutations-total: 2
