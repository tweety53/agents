# Review panel — planner-chooses-models-drop-default-model

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow/implement.md:925 | The gated per-task reviewer — a first-pass review — takes its model from the group's or implementer pair, which Model and effort lets be sonnet; no bound pins it to opus, contradicting the operator's first-pass-review-on-opus rule. |   |
| F2 | primary | Minor | stats/internal/api/records.go:252 | encoding/json matches keys case-insensitively, last key wins, so {"implementer":{"model":"haiku"},"Implementer":"skipped — inline"} passes the refusal while jsonb readers still see haiku. |   |
| F3 | primary | Minor | skills/flow/brainstorm.md:244 | The record-time refusal does not stop the run: the Decide call site states no rule for a refused flow record decision, and dispatches read the on-disk decision.json. |   |
| F4 | primary | Minor | skills/flow-contracts/model-policy-rationale.md:24 | The rationale still says implementers run on Opus and sit at the ceiling from round 1, while the new bounds allow sonnet for implementers. |   |
| F5 | principles | Minor | stats/internal/api/records.go:311 | The refusal message hardcodes "opus, sonnet", a second copy of the set store.ValidModels owns, breaking valid-models-opus-sonnet's one-place bound. |   |
| F6 | principles | Minor | skills/flow-contracts/model-policy.md:38 | The no-recorded-pair-runs-on-opus rule is stated in two files whose lists already differ, readers cite different sources, and brainstorm-planner's canonical-for-which-model claim is too broad. |   |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-principles-2.sh

## Pass log

### Round 0

- roster: compact — 48
- diff size: 3226 changed lines, under cap; docs-only: exit 1 (first non-doc path scripts/check-model-keys.sh) — resolved roster primary+principles dispatched; no addition this round — the resolved list ran alone; standards passed: CLAUDE.md, AGENTS.md

### Round 1

- fix round 1: F1-F6 fixed inline by the parent (FIX_BASE 16510047, fix-round-1.diff); reproducers flipped demonstrated->not-demonstrated on pinned shas; records.go behaviours mutation-proved (F5 test strengthened to an exact suffix after its first mutant survived)
- base CLEAR; re-run: primary alone (raised Important F1) on the rerun pair opus/low, targeted at F1-F4; principles not re-run (Minors only); no addition this round — the resolved list ran alone
- base MOVED (2 commits, no overlap): operator chose Continue, no rebase — the auto-rebase would need a force-push Branch backup reserves for integrate; integrate syncs the branch
fix-mutation: stats/internal/api/records.go — exact-key map decode reverted to the pre-fix case-insensitive struct decode — TestRecordDecisionRejectsCaseVariantKeys
fix-mutation: stats/internal/api/records.go — refusal list no longer derived from ValidModels (extra haiku entry) — TestRecordDecisionRefusalNamesValidModels
fix-mutation: skills/ — none — prose-only hunks (implement, brainstorm, brainstorm-planner, model-policy, rationale): no executable behaviour
fix-mutations-total: 3
