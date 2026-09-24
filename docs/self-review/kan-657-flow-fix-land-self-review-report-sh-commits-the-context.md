# Self-review context bundle for kan-657-flow-fix-land-self-review-report-sh-commits-the

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-657-flow-fix-land-self-review-report-sh-commits-the/tasks.md (absent)
skipped: spectre/changes/archive/kan-657-flow-fix-land-self-review-report-sh-commits-the/design.md (absent)
skipped: spectre/changes/archive/kan-657-flow-fix-land-self-review-report-sh-commits-the/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-657-flow-fix-land-self-review-report-sh-commits-the.md

# SDD ledger — kan-657-flow-fix-land-self-review-report-sh-commits-the

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-24T22:13:29Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 92d17a8
- Outcome: completed
- Started: 2026-09-24T22:52:19Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 7ca71cb24fab365c4500c360d336120af585e9f6
- Outcome: completed
- Started: 2026-09-24T22:52:29Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-657-flow-fix-land-self-review-report-sh-commits-the-panel.md

# Review panel — kan-657-flow-fix-land-self-review-report-sh-commits-the

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | scripts/land-self-review-report.sh:94,108 | the staged-set check and the commit are two separate git calls, so a foreign path staged in between (the same concurrent-session adversary the change names) passes the check and is swept into the commit anyway |   |
| F2 | primary+principles | minor | scripts/land-self-review-report.sh:100-103 | when the mismatch is a vanished own path rather than extra foreign paths, the refusal fires LAND-FOREIGN-STAGED with an empty foreign list and names nothing to investigate |   |
| F3 | primary | minor | scripts/land-self-review-report.sh:94 | the load-bearing --no-renames flag (default rename detection hides a rename-paired foreign deletion from --name-only, reopening a silent sweep) is pinned by no harness case, so dropping it would keep the suite green |   |
| F4 | primary | minor | skills/flow-fast/SKILL.md:279 | the commit subject claims the refusal is named at its call sites but this third invocation site names none of the chain refusals (pre-existing silence; plan scoped two sites) |   |
| F5 | principles | minor | scripts/land-self-review-report.sh:80-83,95-99 | the chain own-paths knowledge is stated twice (add/rm calls and the EXPECTED construction) with no note of why the duplication is kept; WET justifies it, the missing annotation is the violation cue |   |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 grep -q LAND-FOREIGN-STAGED skills/flow-fast/SKILL.md
finding-reproducer: F5 none — duplication judgment, not runtime behavior

## Pass log

### Round 0

- roster: full (compact roll 90 not < 90); experimental 82 — no slot
- diff size 177 under cap — proceeded; docs-only exit 1 (first non-doc path scripts/land-self-review-report.sh) — resolved roster ran
- no addition this round — the resolved list ran alone
- base moved: 4 commits on origin/main, no overlap — auto-rebase clean, merge base now d3617cf4; context bundle rebuilt; dispatch panel-0-primary+principles read .superpowers/sdd/final-review.diff

### Round 1

- panel-fix ran inline — the parent applied it (execution inline), no subagent; read .superpowers/sdd/fix-round-1.diff; F4 reproducer bounced once (inline command unauditable) and re-authored by the raising slot as .superpowers/sdd/reproducers/0-primary-4.sh, fresh dispatch-time verdict demonstrated sha b3909714
- fix-mutations: 3 flips recorded, one per fixed executable behaviour; F4/F5 hunks are prose/comment-only, not executable behaviours; all four runnable reproducers re-run post-fix: not demonstrated, shas pinned to dispatch-time runs
- delta re-run panel-1-primary (rerun pair via harness mapping glm-5.3-flash/high): all four of its findings fixed, no new defects at the named sites; principles did not re-run — it raised no Critical or Important
fix-mutation: scripts/land-self-review-report.sh — the commit dropped its pathspec (bare g commit -m) — test_land_commit_pathspec_limited
fix-mutation: scripts/land-self-review-report.sh — the refusal's missing-own-path clause removed — test_land_missing_own_path_named
fix-mutation: scripts/land-self-review-report.sh — --no-renames dropped from the staged-set capture — test_land_rename_pair_refuses
fix-mutations-total: 3
## git log --stat

commit 970afc3cecddfdb313a1106aa25e89c66ec596df
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 01:46:05 2026 +0300

    docs(flow): name the refusal at the flow-fast call site too

 skills/flow-fast/SKILL.md | 4 +++-
 1 file changed, 3 insertions(+), 1 deletion(-)

commit 92d17a86cefc790bc95eca7f4ac2868655af0b38
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 01:46:05 2026 +0300

    fix(scripts): close the landing chain's check-to-commit race and name what the refusal refuses

 scripts/land-self-review-report.sh      |   6 +-
 scripts/test-land-self-review-report.sh | 100 +++++++++++++++++++++++++++++++-
 2 files changed, 101 insertions(+), 5 deletions(-)

commit 7ca71cb24fab365c4500c360d336120af585e9f6
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 01:10:07 2026 +0300

    docs(flow): name the landing chain's foreign-staged refusal at its call sites

 skills/flow-self-review/SKILL.md | 4 +++-
 skills/flow/archive.md           | 5 +++--
 2 files changed, 6 insertions(+), 3 deletions(-)

commit f0d77035465611301c71d906212cd562af36329a
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 01:09:23 2026 +0300

    fix(scripts): refuse to commit a staged index the landing chain does not own

 scripts/land-self-review-report.sh      |  58 +++++++++++++----
 scripts/test-land-self-review-report.sh | 110 +++++++++++++++++++++++++++++++-
 2 files changed, 156 insertions(+), 12 deletions(-)

## Session narrative

A /flow-fast run over KAN-657: the shared landing-chain script committed the whole index of whatever checkout it ran in, and on 2026-09-22 that swept 121 foreign staged paths into a push that reverted a just-merged change. This run verified the defect still stood at the base, then made the chain index-safe in three moves: a loud exact-set assertion on the staged index (LAND-FOREIGN-STAGED, exit 3) after the chain stages its own paths, a branch re-assert immediately before the commit and the pull/push pair, and a pathspec-limited commit that closes the check-to-commit race the panel Important (F1) caught — the assertion alone left that window open. Four minors landed alongside: the refusal now names a vanished own path instead of firing with an empty foreign list, a harness case pins the load-bearing --no-renames, the third call site (flow-fast SKILL.md) names the refusal like the other two, and a comment records why the own-paths set is deliberately stated twice. The panel ran as one primary+principles bundle; its fix round ran inline in this session with three mutation flips, each proven to fail exactly the harness case that pins it, and the delta re-run confirmed every finding fixed with no new defects. Where the run struggled: the reproducer exit-contract guard bounced F4 once (an inline command cannot carry the demonstrates declaration a script can), and two of the three new harness cases needed a second pass for setup bugs of the run own making — a whole-index setup commit swallowing its own fixture, and a git rm emptying the directory a later printf wrote into.
