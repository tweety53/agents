# Self-review context bundle for kan-800-flow-improvement-defects-become-tasks-the

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-800-flow-improvement-defects-become-tasks-the/tasks.md (absent)
skipped: spectre/changes/archive/kan-800-flow-improvement-defects-become-tasks-the/design.md (absent)
skipped: spectre/changes/archive/kan-800-flow-improvement-defects-become-tasks-the/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-800-flow-improvement-defects-become-tasks-the.md

# SDD ledger — kan-800-flow-improvement-defects-become-tasks-the

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-0-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: bb24ea1900e8097942f66125541fbb48b8cca2ed
- Outcome: not recorded
- Started: 2026-10-03T21:33:29Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 72df60631db0b236ca4e3e4e88c32ee2b9b79624
- Outcome: completed
- Started: 2026-10-03T21:49:28Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-03T22:02:09Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-03T22:02:09Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-800-flow-improvement-defects-become-tasks-the-panel.md

# Review panel — kan-800-flow-improvement-defects-become-tasks-the

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow-contracts/known-bugs.md:34 | check-verbatim-moves.sh exits 1 — three new run-loaded sentences fail (splitter cuts at the "D." boundary and at the em-dash before "consults", stripping the citation shape) and no verbatim-moves.txt exists for this change; the documented remedy is committing spectre/changes/<change>/verbatim-moves.txt with the three FAIL lines verbatim |   |
| F2 | primary | Important | skills/flow-contracts/known-bugs.md:34 | Task 2 Step 1 promised its own worked entry line but added no entry template, and rule 1's only template hard-codes introduced-by <sha>, a field an owned finding has no value for |   |
| F3 | primary | Minor | skills/flow-contracts/known-bugs.md:34 | the new class silently excepts itself from two absolute sentences it does not amend — exactly-when (:13) and the sweep never absorbs it (:20-21) — leaving two opposite courses for the overlap case |   |
| F4 | primary | Minor | skills/flow/brainstorm-planner.md:286 | wording deviations from the plan spelled step text, content-equivalent, plus one clause restating the Allowed-collateral limit the paragraph above owns — confirm intentional |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 sed -n '13p;20,21p;34p' skills/flow-contracts/known-bugs.md
finding-reproducer: F4 none — plan-text comparison, not runnable

## Pass log

### Round 0

- roster: compact — 50
- diff-size: 8 lines, under cap — proceeded
- docs-only: exit 0 — every touched path ends .md; pass 1 reduced to primary alone
- not dispatched — docs-only reduction: principles

### Round 1

- F1 fix path note: the finding's note names the remedy path spectre/changes/<change>/verbatim-moves.txt — the fix commit touches it; the recorded -location is the symptom site
- docs-only: exit 1 where pass 1 saw exit 0 — spectre/changes/<change>/verbatim-moves.txt joined the touched set; principles dispatched this round on the whole final-review.diff
- diff-size: under cap; agents ran: primary re-run (fix round, targeted) + principles first dispatch (whole-branch, docs-only loss)
fix-mutation: spectre/changes/kan-800-flow-improvement-defects-become-tasks-the/verbatim-moves.txt — ack file added — guard run against the defect state with it absent reported 3+ FAILs exit 1, with it present 0 violations exit 0 — check-verbatim-moves 3→0 FAIL sentences
fix-mutation: skills/flow-contracts/known-bugs.md — owned-finding entry template added — F2 reproducer demonstrated (exit 1) → not demonstrated (exit 0)
fix-mutation: skills/flow-contracts/known-bugs.md — none — F3 exception clause is prose; no executable behaviour the round changed
fix-mutation: skills/flow/brainstorm-planner.md — none — F4 clause trim is prose; no executable behaviour the round changed
fix-mutations-total: 4
## git log --stat

commit 72df60631db0b236ca4e3e4e88c32ee2b9b79624
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 00:52:07 2026 +0300

    fix(flow): panel round 0 — ack new sentences, owned-finding entry template, exception clause

 skills/flow-contracts/known-bugs.md                              | 9 +++++++--
 skills/flow/brainstorm-planner.md                                | 2 +-
 .../verbatim-moves.txt                                           | 9 +++++++++
 3 files changed, 17 insertions(+), 3 deletions(-)

commit 097f1ba7007a05002e4bbbd616d9b6eb6e90b4e2
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 00:30:42 2026 +0300

    docs(rationale): record the kan-741 defects-become-tasks precedent

 skills/flow/SKILL-rationale.md | 1 +
 1 file changed, 1 insertion(+)

commit 3639506e85593166bac1ed15355fafd5d0a03941
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 00:30:22 2026 +0300

    feat(known-bugs): record a sweep finding another change owns, never repair it

 skills/flow-contracts/known-bugs.md | 5 +++++
 1 file changed, 5 insertions(+)

commit 0048a6bd2692b6c54b25ad807226cf487e604ef9
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 00:29:50 2026 +0300

    feat(flow): triage a verification sweep's findings before any fix

 skills/flow/brainstorm-planner.md | 2 ++
 1 file changed, 2 insertions(+)

## Session narrative

This run encoded kan-741's defects-become-tasks discipline into the flow corpus: a triage paragraph under the planner's verification-change rule (in-scope defect becomes its own appended task, never a fix inside the verification task itself; anything pre-existing or owned by another change takes the sweep's recorded course), the matching ownership class and worked entry template in the sweep contract, and the rationale precedent. Three task commits, then an inline panel-fix round. Where it struggled: the verbatim-moves guard's sentence splitter — the drafting anticipated it and still missed the "D." boundary cut, and the guard's own documented remedy (the change's verbatim-moves.txt ack) collides with flow-fast's never-write-spectre guardrail; the run resolved that in favour of lint-fix-priority, reading the ack as guard-mandated content rather than a pipeline artifact. The same ack flipped the docs-only panel reduction back open at round 1, expanding the round to include principles on the whole-branch diff — contract-correct, unanticipated at pass 1.
