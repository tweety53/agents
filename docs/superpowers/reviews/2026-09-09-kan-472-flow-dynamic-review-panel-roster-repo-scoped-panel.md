# Review panel — kan-472-flow-dynamic-review-panel-roster-repo-scoped

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | Major | skills/flow/implement.md:133-135 | implement.md:133-135 still cites the "An unspawnable id is substituted, not skipped" section, which task 18 removed whole from review-panel.md — the citation now points at nothing. |
| F2 | primary | Major | skills/flow/SKILL.md:148-152,skills/flow/verify-and-handoff.md:513-515,README.md:238-243,README.md:363-368,skills/flow-contracts/model-policy.md:72,skills/flow-contracts/model-policy.md:81,skills/flow-contracts/model-policy-rationale.md:73 | Five files outside task 18's edited set still describe Bugbot/Security as subagent_type-dispatched with unknown (agent-defined) model, and README.md still says two slots are never merged into one — both contradicted by design.md's bugbot-security-prompt-roles decision, which retires that mechanism globally. |
| F3 | principles | Major | skills/flow/implement.md:133-135 | Single Source of Truth violation: one behavioural fact (how Bugbot/Security are dispatched, and whether panel slots ever merge into one dispatch) has two contradictory descriptions live in the repository — review-panel.md states the retired, bundled reality; seven other locations across five files still assert the old subagent_type/unmerged reality. |
| F4 | code-review-low | Major | skills/flow/implement.md:133-135 | High-confidence broken/contradictory documentation: implement.md cites a review-panel.md section task 18 deleted outright, and five further files describe the retired subagent_type Bugbot/Security mechanism as current — an operationally wrong answer for two of the seven roster slots. |
| F5 | primary | Major | skills/flow/verify-and-handoff.md:478 | /flow's own live-printed Panel: handoff template was never updated for the dynamic panel — still carries dead 'substituted: ... in place of their own agent type' vocabulary from the retired subagent_type mechanism, and none of the dynamic-roster fields (class/compact/rerun/dispatches) the parallel handoff-blocks.md template gained. |
| F6 | principles | Minor | skills/flow/verify-and-handoff.md:478 | Single Source of Truth violation: verify-and-handoff.md and handoff-blocks.md render the same Panel: fact from the same decision state, but only handoff-blocks.md was updated for the dynamic roster — the two now diverge. |
| F7 | code-review-low | Major | skills/flow/verify-and-handoff.md:478 | Two concrete bugs in what this template prints on every run: a dead branch (substituted: is permanently 'none' forever, since nothing is ever subagent_type-substituted any more) and missing data (no field for class/compact/model/effort/rerun/dispatch grouping, the exact information this change exists to surface). |
| F8 | code-review-low | Critical | stats/internal/store/aggregate.go:528,708,skills/flow/review-panel.md:660-661,stats/internal/store/records.go:26,436 | Severity vocabulary mismatch: every reviewer prompt in the pipeline (bugbot/security/simple-reviewer/principles/failure-modes) reports findings as Critical/Important/Minor, but this diff's new Reviewers/Decisions SQL filters severity ILIKE 'major' — a value no real finding is ever recorded with — so that column is permanently zero on real data. |
| F9 | principles | Minor | stats/internal/store/aggregate.go:733-753 | DRY: Decisions()'s SQL repeats the jsonb_typeof(sd.decision->'panel') = 'object' test six times instead of computing it once (e.g. a CTE); a future third panel shape would need six hand-applied edits to stay correct. |
| F10 | primary | Minor | skills/flow/review-panel.md:124 | The roster table's primary row still credits code-quality ownership to code-review-low's and Bugbot's, never updated to name simple-reviewer per design.md's compact-roster-always-code-quality decision record; cosmetic, no dispatch-shape defect. |
| F11 | principles | Critical | stats/internal/store/records.go:451 | SetFindingStatus's deferred-Minor guard compares severity case-sensitively (AND f.severity = 'Minor') against a column the rest of the store (Reviewers/Decisions' ILIKE queries, its own doc comment, records_test.go's lowercase fixtures) treats as case-insensitive free text — a finding recorded as severity minor is wrongly refused deferral. |
| F12 | principles | Minor | stats/internal/store/aggregate.go:7015-7044 | Decisions' array-flatten LATERAL join (flatten an array of string-arrays into a middot-joined string of plus-joined groups) is duplicated verbatim for panel.dispatches and top-level groups — the same duplicated-knowledge shape round 2's panel_is_object finding already fixed a few lines away. |
| F13 | code-review-low | Important | stats/internal/store/records.go:451 | Same defect as F11, raised independently: SetFindingStatus refuses to defer a Minor finding recorded with severity="minor" (lowercase), confirmed live against Postgres. |
| F14 | primary | Important | stats/internal/store/aggregate.go,stats/internal/store/aggregate_test.go (task 17 commit) | Task 17's commit does not build/vet cleanly on its own: its SELECT already returns grouping/dispatches/groups columns and its own test already references DecisionRow.Grouping/.Dispatches, but the struct fields and Scan() call for those columns are only added by task 23's later commit — a Build: green contract violation in the commit history's bisectability, though the final HEAD state builds and passes. |
| F15 | primary | Important | stats/web/src/api.ts,stats/web/src/views/Decisions.tsx (task 17 commit) | Task 17's commit still fails to build standalone on the frontend side: Decisions.tsx already reads DecisionRow.grouping/.dispatches/.implementerGroups but api.ts's DecisionRow interface doesn't declare those fields until task 23's commit. Same bisectability class as F14, on the TS side of the same task's stack. |
| F16 | primary | Important | stats/internal/store/aggregate_test.go (task 16 commit) | Task 16's commit doesn't build standalone: TestDecisionsJoinsRunTotals (calling store.Decisions, added by task 17) landed in task 16's commit tree, one task before the method it tests exists. Fourth instance of the hunk-misplacement bisectability bug this stack has hit tonight (F12/F14 Go-side, F15 TS-side). |
| F17 | primary | Minor | scripts/test-check-dispatch-paragraphs.sh:1266,1679 | Two distinct test cases both labeled "case 42" (a MUTATION PROOF phrase case task 10 renumbered from 35, and task 20s new INDEPENDENT PASSES case, which should have been 45). Harness runs correctly regardless (label is a string, not a key), but a failure in either prints an ambiguous, untraceable case-42 line. |

findings-total: 17
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed

reproducers-total: 17
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/1-primary-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/1-primary-1.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/1-primary-1.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/2-code-review-low-1.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/2-principles-1.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/3-primary-1.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/3-principles-1.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/3-principles-2.sh
finding-reproducer: F13 .superpowers/sdd/reproducers/3-code-review-low-1.sh
finding-reproducer: F14 .superpowers/sdd/reproducers/4-primary-1.sh
finding-reproducer: F15 .superpowers/sdd/reproducers/5-primary-1.sh
finding-reproducer: F16 .superpowers/sdd/reproducers/6-primary-1.sh
finding-reproducer: F17 .superpowers/sdd/reproducers/7-primary-1.sh
