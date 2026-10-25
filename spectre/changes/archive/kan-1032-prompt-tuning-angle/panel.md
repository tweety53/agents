# Review panel — kan-1032-prompt-tuning-angle

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow-prompt-tuning/SKILL.md:50 | Step 1 leaves a red case untouched, but step 6 requires check-prompt-tuning-cases.sh to exit 0 over the same root — one red case blocks the ledger commit and the landing forever. |   |
| F2 | primary | Important | skills/flow-self-review/SKILL.md:172 | Step 2a writes cases to <agents repo>/docs/prompt-tuning/ (the read-only main checkout) before the fix-branch worktree exists, so they are committed from a worktree that never had them. |   |
| F3 | primary | Minor | stats/internal/guard/prompttuningcases.go:155 | Any git show failure reads as target absent, so a scoped case with empty before and a target_sha naming no commit passes. |   |
| F4 | test-sanity | Minor | stats/internal/prompttuning/prompttuning_test.go:72 | TestRenderIsDeterministic renders the same input twice; html/template has no nondeterminism source, so the test guards almost nothing. |   |
| F5 | failure-modes | Minor | stats/internal/guard/prompttuningcases.go:155 | Any git show failure is read as target absent. | supersedes F3 |
| F6 | failure-modes | Minor | stats/cmd/flow/prompttuning.go:36 | os.WriteFile is not atomic; a truncated index.html is possible. |   |
| F7 | failure-modes | Minor | stats/internal/guard/looplanding.go:107 | An unreadable origin/HEAD fails closed to the row checks. |   |
| F8 | mutation | Important | stats/internal/guard/looplanding.go:115 | Dropping \|\| len(paths)==0 survives: an empty diff would be admitted as prompt-tuning-only. |   |
| F9 | mutation | Important | stats/internal/guard/looplanding.go:146 | Dropping the self-review- prefix check survives: any branch would be admitted. |   |
| F10 | mutation | Important | stats/internal/guard/selfreviewreport.go:63 | Changing the seven-angle list's count 7 to 6 survives TestCheckSelfReviewReport. |   |
| F11 | mutation | Important | stats/internal/guard/prompttuningcases.go:136 | Dropping acceptance of reviewer_kind false-positive survives. |   |
| F12 | mutation | Minor | stats/internal/guard/prompttuningcases.go:130 | Minor survivors: Round<1 to <0, empty-before-with-existing-target, scoped non-empty before with absent target, args[0]!=render, 0o644. |   |

findings-total: 12
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 withdrawn the test pins the byte-stability contract the guard compares index.html against
finding-status: F5 fixed
finding-status: F6 withdrawn a truncated write is detected by the guard byte-compare and repaired by re-rendering, losing no data
finding-status: F7 withdrawn an unreadable origin/HEAD fails closed to the row checks, which is the intended behaviour
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed

reproducers-total: 12
finding-reproducer: F1 .superpowers/sdd/reproducers/r0-P2-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/r0-P2-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/r0-P1-1.sh
finding-reproducer: F4 none — Minor, ceremony test with no mechanism to mutate
finding-reproducer: F5 .superpowers/sdd/reproducers/r0-P1-1.sh
finding-reproducer: F6 none — Minor, a truncated write is caught by the guard's byte-compare
finding-reproducer: F7 none — Minor, fails closed
finding-reproducer: F8 .superpowers/sdd/reproducers/r0-M1-1.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/r0-M2-1.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/r0-M3-1.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/r0-M4-1.sh
finding-reproducer: F12 none — Minor survivors listed in the report

## Pass log

### Round 0

- dispatched: panel-0-primary+principles on opus/low; panel-0-test-sanity+call-efficiency on sonnet/low; panel-0-failure-modes+mutation on sonnet/low
- verdict: fix (0 Critical, 6 Important)

### Round 1

- dispatched: panel-1-primary+mutation on sonnet/low over 0db772676aad1e7fa5f2e4c4ff4068ff3fb25015
- verdict: clean
