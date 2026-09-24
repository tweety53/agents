# Self-review context bundle for kan-650-flow-improvement-the-known-bugs-md-sweep-turns

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-650-flow-improvement-the-known-bugs-md-sweep-turns/tasks.md (absent)
skipped: spectre/changes/archive/kan-650-flow-improvement-the-known-bugs-md-sweep-turns/design.md (absent)
skipped: spectre/changes/archive/kan-650-flow-improvement-the-known-bugs-md-sweep-turns/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-650-flow-improvement-the-known-bugs-md-sweep-turns.md

# SDD ledger — kan-650-flow-improvement-the-known-bugs-md-sweep-turns

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-24T21:14:20Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-24T21:33:20Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-650-flow-improvement-the-known-bugs-md-sweep-turns-panel.md

# Review panel — kan-650-flow-improvement-the-known-bugs-md-sweep-turns

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | skills/flow-contracts/known-bugs.md:56 | Contract claims the full-suite runs a baseline regeneration forces consult the sweep, but only step 7 is routed; step 8 full-app-suite runs keep unconditional block language, so a pre-existing visual defect surfacing there blocks despite the coverage claim. |   |
| P1 | principles | minor | skills/flow/verify-and-handoff.md:102 | DRY/SSOT: the classification test (introducing commit an ancestor of the default branch) now has two representations; defensible WET tradeoff but undeclared against tasks.md routing-only rule. |   |
| F2 | primary | minor | scripts/check-contract-budget.sh:176 | Budgets row 4061 is one byte above the plan-required figure (landed size 3248 + 25% = 4060). |   |
| F3 | primary | minor | skills/flow-contracts/known-bugs.md:24 | The sweep-mandated root KNOWN-BUGS.md trips check-contract-budget.sh (no budget declared, demonstrated in a sandbox copy); the contract never states that budgets() row obligation. |   |
| F4 | primary | minor | skills/flow/verify-and-handoff.md:102 | Citation sentences restate sweep rule fragments where tasks.md required routing only. |   |

findings-total: 5
finding-status: F1 fixed
finding-status: P1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: P1 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh

## Pass log

### Round 0

- roster: compact — 70
- no addition this round — the resolved list ran alone.
- diff size 74 — under cap; proceed automatic (exit 0)
- docs-only: exit 1 — first non-documentation path scripts/check-contract-budget.sh; resolved roster runs: primary+principles
- base moved (3 commits on origin/main, overlaps skills/flow/verify-and-handoff.md) — /flow-fast asks no base-movement question; continue — review as is; regions disjoint by inspection; landing rebase will reconcile
## git log --stat

commit 8c91c708e3b05a53b239e028671fab53f3b1b0a7
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 00:37:25 2026 +0300

    fix(flow-contracts): spell the KNOWN-BUGS.md location in the repo's root vocabulary

 skills/flow-contracts/known-bugs.md | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)

commit 2c846308559cf01e5c7fc9e58cdd0611d9b41e9d
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 00:32:07 2026 +0300

    fix(flow): route the full-suite capture through the sweep, trim citations to routing

 scripts/check-contract-budget.sh    | 2 +-
 skills/flow-contracts/known-bugs.md | 5 ++++-
 skills/flow-fast/SKILL.md           | 3 +--
 skills/flow/verify-and-handoff.md   | 6 ++----
 skills/flow/visual-verify.md        | 8 +++++---
 5 files changed, 13 insertions(+), 11 deletions(-)

commit 6629c9f7efebe5105d0309677b7c90485b59907f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 00:09:32 2026 +0300

    docs(flow): consult the known-bugs sweep when a verify surfaces pre-existing failures

 skills/flow-fast/SKILL.md         | 4 +++-
 skills/flow/verify-and-handoff.md | 6 +++++-
 skills/flow/visual-verify.md      | 4 +++-
 3 files changed, 11 insertions(+), 3 deletions(-)

commit 2ce31b309f5cd69cf28c0ec13c6ef36e3fd0aaa4
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 00:08:18 2026 +0300

    feat(flow-contracts): canonical known-bugs sweep contract

 scripts/check-contract-budget.sh    |  1 +
 skills/flow-contracts/known-bugs.md | 59 +++++++++++++++++++++++++++++++++++++
 2 files changed, 60 insertions(+)

## Session narrative

The run implemented KAN-650 inline on a small-class decision (compact panel, primary+principles
bundled): a new canonical contract, `skills/flow-contracts/known-bugs.md`, defines the
KNOWN-BUGS.md sweep — prove a surfaced failing test pre-existing by naming its introducing
commit and checking it an ancestor of the default branch, record it durably (spec, root cause,
introducing commit), never repair it, feed `## known failures` where declared — cited from the
three verify readers (`verify-and-handoff.md`, `visual-verify.md`, `flow-fast/SKILL.md`).
Friction, all resolved: origin/main moved mid-run (3 commits overlapping
`verify-and-handoff.md`), and since `/flow-fast` asks no base-movement question the run took the
panel contract's continue-as-is course after confirming the regions disjoint; the panel raised
one Important (the contract claimed full-suite capture coverage the citations did not wire) plus
three Minors and one principles Minor, all fixed in one pathspec-scoped commit and confirmed by
reproducer re-runs and a targeted primary re-run; `check-installed-citations.sh` caught the
contract's `<project-root>/` placeholder spelling, fixed to the repo's `<project>/` vocabulary.
Time sink: the fresh worktree's SPA build for `go vet`/`tsc`. Nothing was deliberately left out;
`/flow-self-review` was not modified (the sweep is stage behavior, and no skill-file anchor
exists in that skill's bundle-only pass).
