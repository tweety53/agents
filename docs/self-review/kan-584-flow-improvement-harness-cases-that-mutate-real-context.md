# Self-review context bundle for kan-584-flow-improvement-harness-cases-that-mutate-real

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-584-flow-improvement-harness-cases-that-mutate-real/tasks.md (absent)
skipped: spectre/changes/archive/kan-584-flow-improvement-harness-cases-that-mutate-real/design.md (absent)
skipped: spectre/changes/archive/kan-584-flow-improvement-harness-cases-that-mutate-real/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-584-flow-improvement-harness-cases-that-mutate-real.md

# SDD ledger — kan-584-flow-improvement-harness-cases-that-mutate-real

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T18:03:59Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 8558573
- Outcome: completed
- Started: 2026-09-19T18:29:00Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 8558573
- Outcome: completed
- Started: 2026-09-19T18:29:00Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-584-flow-improvement-harness-cases-that-mutate-real-panel.md

# Review panel — kan-584-flow-improvement-harness-cases-that-mutate-real

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | scripts/run-guard-tests.sh:86 | bare-repo GUARD_TESTS_REPO_ROOT passes validation (rev-parse --is-inside-work-tree exits 0 printing false) and the before-snapshot git status dies under set -e — undocumented exit 128, raw git fatal, leaked TIME_DIR |   |
| P1 | principles | Important | scripts/run-guard-tests.sh:86 | same root as F1, principles angle: a bare-repo environmental input crashes via set -e instead of taking the documented graceful skip — failures must degrade gracefully and be observable |   |
| F2 | primary | Minor | scripts/run-guard-tests.sh:49-57 | gate predicate is porcelain-output-unchanged, so mutating the content of an already-dirty tracked file escapes with exit 0 while the header claims a green run proves nothing touched the live tree — overclaim |   |

findings-total: 3
finding-status: F1 fixed
finding-status: P1 fixed
finding-status: F2 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: P1 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh

## Pass log

### Round 0

- base moved on origin/main, no overlap — rebased clean onto 669b896, merge base updated
- roster: compact — 7
- diff-size 211 lines, under cap — proceeded
- docs-only: no — first non-doc path scripts/run-guard-tests.sh; resolved roster dispatched
- no addition this round — the resolved list ran alone

### Round 1

- base moved, no overlap — rebased clean onto origin/main; FIX_BASE was 8558573, fix commit 350349f
fix-mutation: scripts/run-guard-tests.sh — work-tree validation reads rev-parse verdict; before-snapshot guarded — 1→0 bare-repo watched-root crashes (pre: fatal + leaked run-guard-tests-time dir; post: announced skip, exit 0, no leak); harness case 7e pins it
fix-mutation: scripts/run-guard-tests.sh — after-snapshot guarded like the before-side — none — the failure needs the watched repo to vanish mid-run, unbuildable in a harness; pattern verified by inspection against the measured before-side skip
fix-mutation: scripts/run-guard-tests.sh — none — F2 hunks are comment-only — no executable behaviour changed
fix-mutations-total: 3

## Branch log

```text
commit cc4ebe326978014736119bc954778c595fe7840b
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 21:28:13 2026 +0300

    fix(scripts): gate skips a bare watched repo instead of crashing

 scripts/run-guard-tests.sh      | 43 ++++++++++++++++++++++++++++++-----------
 scripts/test-run-guard-tests.sh | 23 ++++++++++++++++++++++
 2 files changed, 55 insertions(+), 11 deletions(-)

commit 3dec4c7e39af5c4434e4bc8575d718ff5a2164c7
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 21:00:57 2026 +0300

    test(scripts): cover the runner's tree-integrity gate

 scripts/test-run-guard-tests.sh | 118 ++++++++++++++++++++++++++++++++++++++++
 1 file changed, 118 insertions(+)

commit 8ac8b140db21ad9cd733dc010ea8ed050450a09d
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 21:00:54 2026 +0300

    feat(scripts): gate the guard suite on the watched repo staying clean

 scripts/run-guard-tests.sh | 80 ++++++++++++++++++++++++++++++++++++++++------
 1 file changed, 71 insertions(+), 9 deletions(-)

commit 3c510ebf210a072da1436e5f441f280181eef065
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 20:55:52 2026 +0300

    fix(scripts): sandbox test-setup-agents.sh output capture

 scripts/test-setup-agents.sh | 13 +++++++++----
 1 file changed, 9 insertions(+), 4 deletions(-)
```

## Session narrative

KAN-584 asked that harness cases mutating real shared files run against sandbox fixtures instead. The session audited all 75 guard-test harnesses for mutations of the live repo or real HOME and found the suite overwhelmingly compliant — the one concrete offender was test-setup-agents.sh, whose installer-output capture went to the fixed path /tmp/test-setup-agents.out (shared across concurrent runs, never cleaned up); a fixed-path case and a stale leftover file were both removed. Because nothing enforced the discipline for future cases, the change also added a tree-integrity gate to run-guard-tests.sh: the runner snapshots git status --porcelain of the watched repository before the suite and fails, naming the delta, if the tree changed — GUARD_TESTS_REPO_ROOT overrides the watched root for the harness, and a non-git watched root takes an announced skip. The struggle was concentrated in three places: the first scratch mutation-proof of the new harness cases proved nothing because the copied runner lacked lib/parallel.sh and died at source time, so the proof was redone with the library present before the claims were allowed to stand; a header edit to test-setup-agents.sh briefly clipped a comment continuation line and broke the harness under set -e; and the panel itself found a real hole in the gate — rev-parse --is-inside-work-tree exits 0 printing false in a bare repository, so validation had to read the verdict rather than the exit code — plus an overclaiming header sentence, both fixed and mutation-proved in round 1 with both Important reproducers re-run green by the re-reviewing slots.
