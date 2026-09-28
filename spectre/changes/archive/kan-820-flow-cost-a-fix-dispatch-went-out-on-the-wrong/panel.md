# Review panel — kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Important | skills/flow/SKILL.md:56 | the resolver's tr -d backtick + xargs normalization mangles mismatched bodies into valid models instead of reporting-and-dropping, and is a second drifted implementation of the body rule the Go guard implements |   |
| F2 | primary+principles | Important | skills/flow/SKILL.md:82 | the kept outage paragraph instructs falling back to the literal opus, contradicting the implemented project-wins-on-store-outage property, and the bare settings-get read aborts under set -e before the project key is read |   |
| F3 | primary | Minor | spectre/changes/kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong/tasks.md:132 | task 3's checked verify step names scripts/check-contract-budget.sh, which no longer exists on main or at the merge base |   |
| F4 | primary | Minor | spectre/changes/kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong/tasks.md:185 | task 5's commit fence still stages verify-and-handoff.md, contradicting the task's own recorded correction and its actual commit |   |
| F5 | primary+principles | Minor | skills/flow/SKILL.md:52 | the outage-safe read's 2>/dev/null swallows the CLI's stderr, so a store outage resolves silently and the rewritten paragraph dropped the report-the-stderr duty |   |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 deferred the fix diff introduced it; one-token fix deferred with the sibling archive.md stderr-visible pattern named

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/1-primary-1.sh

## Pass log

### Round 0

- roster: compact — 22
- diff size: 523 changed lines, cap 600 — under cap, proceed
- docs-only reduction: not applied — exit 1, first non-doc path scripts/check-model-keys.sh
- no addition this round — the resolved list ran alone
