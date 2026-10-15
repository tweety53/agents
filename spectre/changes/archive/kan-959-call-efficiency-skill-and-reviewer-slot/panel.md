# Review panel — kan-959-call-efficiency-skill-and-reviewer-slot

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | skills/flow-fast/SKILL.md:90 | stale Decide step citations after renumbering — flow-fast steps 1–4, verify-and-handoff step 5 |   |
| F2 | primary | important | skills/flow/review-panel.md:432 | dispatch recording rules contradict the extra dispatch's pair (-model opus no exception; -slot from The roster only; late-fix one dispatch) |   |
| F3 | primary | minor | stats/internal/guard/plan_class_test.go:274 | micro end-to-end extra line not pinned |   |
| F4 | primary | minor | skills/flow/call-efficiency-reviewer-prompt.md:56 | call-efficiency Do Not bullets carry failure-modes wording |   |
| F5 | primary | minor | README.md:115 | README indexes stale: rolls, dispatch count, /call-efficiency-audit missing |   |
| F6 | primary | minor | skills/flow/review-panel.md:170 | placeholder table omits the new templates |   |
| F7 | primary | minor | skills/flow-fast/SKILL.md:260 | flow-fast micro wording reads as skipping the panel |   |
| F8 | principles | minor | skills/flow/call-efficiency-reviewer-prompt.md:56 | Least Astonishment: copied Do Not bullets |   |
| F9 | principles | minor | skills/flow/review-panel.md:432 | SSoT: default-panel model opus no exception vs extra rolled pair |   |

findings-total: 9
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed

reproducers-total: 9
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-5.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-primary-6.sh
finding-reproducer: F7 none — wording ambiguity, no mechanical check
finding-reproducer: F8 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/0-primary-2.sh

## Pass log

### Round 0

- diff size 1434, under cap; docs-only exit 1 (scripts/plan-class.sh); roster: compact — 54; decided roster primary+principles opus/medium plus the always-on extra dispatch test-sanity+call-efficiency opus/low (extra roll 53); no operator-added slot; templates rendered from the change worktree skill dir, which carries the new test-sanity/call-efficiency templates

### Round 1

- panel-fix-1 inline row: begin was recorded after the work, so its end was refused under the 10s window; closed aborted — the fix commit is on the branch
