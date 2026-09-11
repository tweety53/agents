# Review panel — kan-512-flow-defer-self-review-to-a-stronger-model-via

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | Important | skills/flow-contracts/git-boundaries.md:23 | git-boundaries.md, archive.md step 10 and finish-contract-run2.md step 10 still assert step 9 always produces/commits the self-review report, which is false on the new ## self review: defer literal |
| F2 | primary | Important | skills/flow-settings/SKILL.md:15-16 | flow-settings/SKILL.md still says selfReviewModel is the model the self-review pass runs on/dispatches on, contradicting this same change's own correction in skills/flow/SKILL.md (governs no dispatch) |
| F3 | simple-reviewer | Minor | skills/flow-self-review/SKILL.md:85-93 | flow-self-review/SKILL.md step 5's commit shell has no branch re-assert immediately before git add/commit, unlike both sibling commit shells in archive.md |
| F4 | primary | Important | skills/flow-self-review/SKILL.md:89-95 | the F3 fix's branch guard covers only add/rm/commit, not the unconditional git pull --rebase / git push that follow in the same block, and carries no stop-on-mismatch prose unlike archive.md's sibling shells |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-simple-reviewer-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/1-primary-1.sh
