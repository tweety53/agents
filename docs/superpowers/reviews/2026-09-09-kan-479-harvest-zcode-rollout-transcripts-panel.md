# Review panel — kan-479-harvest-zcode-rollout-transcripts

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | Minor | spectre/changes/kan-479-harvest-zcode-rollout-transcripts/design.md:unknown-split-flat-rate | design.md and tasks.md state the flat-rate rule as '1h nil or 1h == 5m', but the implementation (and its store-level pin TestPriceDispatchWithUnknownCacheSplitGetsNoCost) requires a nil-1h row to ALSO carry the collapsed column equal to the 5m rate — a pre-0007 row with only the collapsed column set must still refuse, since its zero 5m rate is an unset, not a published free. Amend the design decision and the task step to the refined rule. |
| F2 | primary | Minor | spectre/changes/kan-479-harvest-zcode-rollout-transcripts/tasks.md:Baseline-fields | The plan's Baseline after-values predict 1036 test cases; the measured suite counts 1035, and task 3's delta is +7 (the degenerate-row subtest added during implementation), not the planned +6. Amend the Baseline chain to the implemented deltas and tag the final count measured. |
| F3 | primary | Minor | spectre/changes/kan-479-harvest-zcode-rollout-transcripts/proposal.md | proposal's original What-changes bullet still keys the seed row GLM-5.3-Flash after fix 1 re-keyed it glm-5.3-flash — stale text its own Fix 1 section supersedes |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 none — plan-artifact divergence, not a code defect
finding-reproducer: F2 cd stats && go test ./... -count=1 -v | grep -c '^=== RUN'  # measured 1035, plan's chained prediction 1036
finding-reproducer: F3 none — planning-artifact wording, not a code defect
