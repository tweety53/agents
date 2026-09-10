# Review panel — kan-482-flow-conductor-dispatches-one-fix-subagent-per

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | Major | spectre/changes/kan-482-flow-conductor-dispatches-one-fix-subagent-per/proposal.md:21-22 | proposal declares -session-token/-role optional filters on flow record dispatches; the implemented verb rejects both, and no amendment records the drop — amend the proposal to the shipped -change-only surface or add the flags |
| F2 | primary | Minor | skills/flow/review-panel.md:959 | task 3's steps name only an addition beside the findings-closed call, but the diff also shortens the existing 'Exit 0 proceeds to the stage close below.' sentence — diff content no step describes |
| F3 | mutation | Major | scripts/check-panel-fix-single-dispatch.sh:102-103 | surviving mutant M15: deleting the session-token condition from the guard's jq filter passes the whole harness — case 6's foreign-token row is itself a clean round, so the other-tokens-never-count contract is unpinned |
| F4 | mutation | Major | stats/cmd/flow/record.go:1286-1288 | surviving mutant M3: no TestRunRecordDispatches test drives an existing change whose dispatches is null — dropping the nil-guard passes the suite and the verb would print null instead of the contracted [] |
| F5 | mutation | Minor | scripts/check-panel-fix-single-dispatch.sh:169 | surviving mutant M7: weakening the per-round original-count rule to -gt 2 passes the harness — the sibling total/composition rule catches the same shape, so the rule's presence alone is unpinned |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 deferred — accepted equivalent mutant: the per-round original-count rule's exit behavior is fully covered by the sibling total/composition rule (case 3's double-sabotage kills the pair), so no test can pin it alone; the branch stays for its violation message naming the one-per-reviewer anti-pattern

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-0-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-0-1.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-0-2.sh
