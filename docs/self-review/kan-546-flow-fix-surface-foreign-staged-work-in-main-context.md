# Self-review context bundle for kan-546-flow-fix-surface-foreign-staged-work-in-main

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-546-flow-fix-surface-foreign-staged-work-in-main/tasks.md (absent)
skipped: spectre/changes/archive/kan-546-flow-fix-surface-foreign-staged-work-in-main/design.md (absent)
skipped: spectre/changes/archive/kan-546-flow-fix-surface-foreign-staged-work-in-main/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-546-flow-fix-surface-foreign-staged-work-in-main.md

# SDD ledger — kan-546-flow-fix-surface-foreign-staged-work-in-main

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T19:14:06Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: exp-failure-modes
- Key: panel-0-exp-failure-modes
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T19:14:06Z
- Tokens: not measured

## Dispatch 3 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 053bef0
- Outcome: completed
- Started: 2026-09-17T19:39:41Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Slot: exp-failure-modes
- Key: panel-1-exp-failure-modes
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T19:44:11Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-546-flow-fix-surface-foreign-staged-work-in-main-panel.md

# Review panel — kan-546-flow-fix-surface-foreign-staged-work-in-main

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | skills/flow/SKILL.md:192 | check-foreign-staged.sh was inserted before check-finish-preflight.sh, but fi sorts before fo — the roster is not alphabetical as the task requires |   |
| F2 | primary+principles | Minor | scripts/check-foreign-staged.sh:73 | a git add -N intent-to-add entry prints with a leading-space code, so the guard says STAGED-CLEAN while check-finish-preflight.sh's identical read is non-empty and REFUSEs — the surface stays silent on a state the next gate stops for |   |
| F3 | exp-failure-modes | Important | skills/flow-contracts/finish-contract-run1.md:79 | the surfacing step's exit-2 cannot-answer outcome is unhandled in both wiring files — an unreadable main checkout falls through to the preflight with the surfacing never performed, reading an inability as clean |   |
| F4 | exp-failure-modes | Minor | skills/flow-contracts/finish-contract-run1.md:82 | the exit-2 clause says it stops exactly as a STAGED-FOREIGN does, but the Continue branch carries a listing an exit-2 never has |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-exp-failure-modes-1.sh
finding-reproducer: F4 none — prose nuance, no behavioral difference

## Pass log

### Round 0

- base movement: MOVED (1 commit on origin/main), no overlap with this change paths — continuing unprompted
- diff-size 346 under cap; docs-only exit 1 (first non-doc path scripts/check-foreign-staged.sh) — resolved roster runs unchanged; no operator-named addition this round — the resolved list ran alone; roster: full

### Round 1

- fix round 1 inline (panel-fix -agent-id inline): F1 roster swap, F2 intent-to-add listed with harness case + mutation proof, F3 exit-2 outcome stated in both wiring files; all three reproducers flipped to defect-gone; primary+principles not re-run — their round-0 findings were Minors fixed inline, covered by the harness and the flipped reproducers
- re-run verdict: F3 fixed, reproducer exits 0, real exit-2 path exercised; one new Minor F4 fixed inline (wording only, no re-run under the trivial bar)
fix-mutation: .superpowers/sdd/fix-round-1.diff — none — the fix changed prose and a two-token roster swap only; no executable behaviour changed
fix-mutation: scripts/check-foreign-staged.sh — the second-column-A filter (dropped the OR branch) — test-check-foreign-staged.sh intent-to-add case
fix-mutations-total: 2

## Branch log

commit 769cdf1c3d05ca2368c8741f0dda1521972a806f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Sep 17 22:47:09 2026 +0300

    docs(finish-contract): the exit-2 ask has nothing to list

 skills/flow-contracts/finish-contract-run1.md | 5 +++--
 1 file changed, 3 insertions(+), 2 deletions(-)

commit 053bef0b07f76c57a336b392364b62e99ca5a2e9
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Sep 17 22:43:19 2026 +0300

    fix(guards): list intent-to-add entries as staged work

 scripts/check-foreign-staged.sh               | 16 ++++++++++------
 scripts/test-check-foreign-staged.sh          | 18 ++++++++++++++++++
 skills/flow-contracts/finish-contract-run1.md |  1 +
 3 files changed, 29 insertions(+), 6 deletions(-)

commit 276e625c4b433a80f531f1a21e2ebd7c9f86f6fb
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Sep 17 22:40:20 2026 +0300

    fix(guards): close the round-1 panel findings

 scripts/check-foreign-staged.sh               | 5 +++++
 skills/flow-contracts/finish-contract-run1.md | 4 +++-
 skills/flow/SKILL.md                          | 2 +-
 skills/flow/integrate.md                      | 3 ++-
 4 files changed, 11 insertions(+), 3 deletions(-)

commit b1ecd31c7d9a013a5ea7b5782ccf78e825d2b50f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Sep 17 22:10:58 2026 +0300

    docs(project-config): exclude the staged-work guard from lint

 .flow/project.md | 11 ++++++-----
 1 file changed, 6 insertions(+), 5 deletions(-)

commit c8c5f6b8bd82b917386a5e53026644abe6f704bd
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Sep 17 22:10:58 2026 +0300

    docs(guards): roster the staged-work surface guard

 skills/flow-contracts/pipeline.md | 3 ++-
 skills/flow/SKILL.md              | 2 +-
 2 files changed, 3 insertions(+), 2 deletions(-)

commit dac646531d589eb476fdcda545b3bfd14564a177
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Sep 17 22:10:11 2026 +0300

    docs(flow): run the staged-work surface before the preflight

 skills/flow/integrate.md | 6 ++++++
 1 file changed, 6 insertions(+)

commit 2bf7bf9ec81f152124a521c22f18a27ea27edbbb
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Sep 17 22:10:11 2026 +0300

    docs(finish-contract): surface foreign staged work before the preflight

 skills/flow-contracts/finish-contract-run1.md | 33 +++++++++++++++++++++++++++
 1 file changed, 33 insertions(+)

commit bc8fc08c121a0c4c320faa54132a01ae847d0d33
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Sep 17 22:08:48 2026 +0300

    feat(guards): list foreign staged work in a main checkout

 scripts/check-foreign-staged.sh             |  80 +++++++++++
 scripts/test-check-foreign-staged.sh        | 210 ++++++++++++++++++++++++++++
 skills/flow/scripts/check-foreign-staged.sh |   1 +
 3 files changed, 291 insertions(+)

## Session narrative

This /flow-fast run implemented KAN-546: a pre-run guard, `scripts/check-foreign-staged.sh`, that surfaces foreign staged work in a main checkout before a resumed /flow run's preflight reaches it, with the wiring to make it a step of bare /flow's run decision. The brainstorm read the finish contracts, the preflight guard and the kan-546 ticket, and settled the design in one pass: the guard is a disclosure-and-stop (list, ask, two courses), never a gate that mutates anything, because the defect being prevented was the run itself improvising reset/stash surgery (kan-437's run 2). The decide roll came back class small, execution inline, full roster with one experimental slot (failure-modes), free grouping — two pass-1 dispatches on the recorded pairs, replaced per the zcode harness mapping. The panel earned its keep: exp-failure-modes caught a real Important (the exit-2 cannot-answer outcome unhandled in both wiring files — an inability readable as clean), primary caught the roster misorder and, with principles, the intent-to-add divergence; the divergence's own reproducer pushed the fix past the reviewer's suggested naming sentence to a small behavior change (second-column-A entries are listed), which the harness now pins with a mutation proof. The run struggled twice in small ways: the first three files were written into the main checkout before being caught and moved into the worktree, and the reproducer re-runs initially reported refusals because the parent passed the raw pre-fix reproducer exit rather than the verdict run-reproducer.sh printed — both corrected inside the round. No scope was added beyond the ticket; the hook (`hooks/protect-main-checkout.py`) was deliberately left untouched — it prevents creating residue, this guard surfaces existing residue.
