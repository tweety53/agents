# Self-review context bundle for kan-647-flow-fix-preflight-sees-staged-work-but-not-a

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-647-flow-fix-preflight-sees-staged-work-but-not-a/tasks.md (absent)
skipped: spectre/changes/archive/kan-647-flow-fix-preflight-sees-staged-work-but-not-a/design.md (absent)
skipped: spectre/changes/archive/kan-647-flow-fix-preflight-sees-staged-work-but-not-a/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-647-flow-fix-preflight-sees-staged-work-but-not-a.md

# SDD ledger — kan-647-flow-fix-preflight-sees-staged-work-but-not-a

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 4e92e5476ac4e11299e50d114fd19735f97a2b8c
- Outcome: completed
- Started: 2026-09-24T21:23:11Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: abb7cde7797166300502b0d495bd6169f044eb22
- Outcome: completed
- Started: 2026-09-24T21:53:54Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: abb7cde7797166300502b0d495bd6169f044eb22
- Outcome: completed
- Started: 2026-09-24T21:53:54Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-647-flow-fix-preflight-sees-staged-work-but-not-a-panel.md

# Review panel — kan-647-flow-fix-preflight-sees-staged-work-but-not-a

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | important | skills/flow-contracts/finish-contract-run1.md:115 | the diff added a second operator-question block instead of rewording the original, leaving two differently-worded stop-and-ask prompts for what the prose calls one ask |   |
| F2 | primary+principles | important | scripts/check-contract-budget.sh:167 | the budgets row was raised to exactly the current size — headroom ratio 1.00 — where the guard rule is landing size plus 25 percent, i.e. 38565 |   |
| F3 | primary | minor | skills/flow/scripts/check-main-checkout-drift.sh:1 | the carried symlink is in no task Files list, so a planned-pieces review cannot see it was intentional |   |
| F4 | primary | minor | .superpowers/sdd/kan-647-flow-fix-preflight-sees-staged-work-but-not-a/tasks.md:75 | task 4 Commit field names chore(scripts): verify the drift-guard change, a commit the branch does not carry |   |
| F5 | primary | minor | scripts/test-check-main-checkout-drift.sh:1 | no harness case passes an existing regular file as the argument, so the exit-2 refusal it must give is unpinned |   |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 none — plan-metadata deviation; no runtime behavior demonstrates it
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F5 none — absence of a test case; the correct behavior was demonstrated by hand
## git log --stat

commit d4e54ba281e6b51b68cee66719afb357db6b157e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 00:53:33 2026 +0300

    fix(finish-contract): close the pass-1 panel findings

 scripts/check-contract-budget.sh              |  2 +-
 scripts/test-check-main-checkout-drift.sh     | 10 ++++++++++
 skills/flow-contracts/finish-contract-run1.md |  4 ----
 3 files changed, 11 insertions(+), 5 deletions(-)

commit abb7cde7797166300502b0d495bd6169f044eb22
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 00:10:21 2026 +0300

    feat(finish-contract): surface main-checkout drift before the run preflight

 scripts/check-contract-budget.sh                 |  2 +-
 skills/flow-contracts/finish-contract-run1.md    | 26 ++++++++++++++++++++++++
 skills/flow/integrate.md                         | 13 ++++++------
 skills/flow/scripts/check-main-checkout-drift.sh |  1 +
 4 files changed, 35 insertions(+), 7 deletions(-)

commit 8832f4fe6dff3073bd680c6ce48b752d0fafbc82
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 00:08:24 2026 +0300

    feat(scripts): add the main-checkout drift guard

 scripts/check-main-checkout-drift.sh      | 108 ++++++++++++++
 scripts/test-check-main-checkout-drift.sh | 239 ++++++++++++++++++++++++++++++
 2 files changed, 347 insertions(+)

## Session narrative

Implemented KAN-647 inline: a new guard `scripts/check-main-checkout-drift.sh` that names a main checkout's drift from its expected post-merge state (a foreign or detached branch, tracked-content drift, with clean-but-behind and untracked-only states explicitly not findings), a 14-case assertion harness beside it, and the pre-run surface extension in `finish-contract-run1.md` and `integrate.md` that runs the drift guard once per distinct main checkout in the same position as the KAN-546 staged-work surface and folds its verdicts into the same one-shot stop-and-ask; `check-foreign-staged.sh` and `check-finish-preflight.sh` keep their recorded contracts untouched. Where the run struggled was entirely in the repository's own guard mesh, and each hit was a real lesson: the plan-shape guard dictated the `**After:**` ordering grammar; `check-guard-symlinks.sh` demanded both a carried symlink beside `skills/flow/` and a line-local citation shape in `integrate.md` (a wrapped line silently dropped two guard names from the required set); the contract-budget ratchet tripped on the contract growth; and the review panel found the budget row raised with zero headroom against the guard's own +25% rule plus the duplicated operator ask my contract edit left behind — both fixed and closed by reproducer flip, with the F4 instrument re-authored when the fix renamed the defect's subject and one instrument bounced once on an unresolvable `# demonstrates:` citation. The panel's pass notes could not be written to the store because this flow CLI build has no `flow record pass` verb; they live in this narrative instead.
