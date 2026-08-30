# Review panel — kan-370-self-review-model

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Primary | Critical | scripts/check-contract-budget.sh:185 | skills/flow-settings/SKILL.md exceeds its declared byte budget; neither the diff nor the plan updates the budget table |
| F2 | Primary | Important | skills/flow-settings/SKILL.md:3 | frontmatter description omits the new selfReviewModel field the body now documents |
| F3 | Primary | Minor | skills/flow-settings/SKILL.md:43-47 | Current flow settings example block's colon alignment is inconsistent |
| F4 | Bugbot | Important | stats/internal/client/client_test.go:906-951 | TestSettingsRoundTripsSelfReviewModel fakes transport by re-encoding/decoding the same client.Settings type it's proving, so it does not actually validate the selfReviewModel wire key |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 cd .worktrees/kan-370-self-review-model && ./scripts/check-contract-budget.sh
finding-reproducer: F2 none — visual inspection of the frontmatter block
finding-reproducer: F3 none — visual inspection
finding-reproducer: F4 change client.Settings.SelfReviewModel's json tag to a wrong value in stats/internal/client/client.go; TestSettingsRoundTripsSelfReviewModel still passes
