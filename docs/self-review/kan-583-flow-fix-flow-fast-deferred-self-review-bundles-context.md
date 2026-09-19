# Self-review context bundle for kan-583-flow-fix-flow-fast-deferred-self-review-bundles

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-583-flow-fix-flow-fast-deferred-self-review-bundles/tasks.md (absent)
skipped: spectre/changes/archive/kan-583-flow-fix-flow-fast-deferred-self-review-bundles/design.md (absent)
skipped: spectre/changes/archive/kan-583-flow-fix-flow-fast-deferred-self-review-bundles/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-583-flow-fix-flow-fast-deferred-self-review-bundles.md

# SDD ledger — kan-583-flow-fix-flow-fast-deferred-self-review-bundles

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T18:05:36Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-583-flow-fix-flow-fast-deferred-self-review-bundles-panel.md

# Review panel — kan-583-flow-fix-flow-fast-deferred-self-review-bundles

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | stats/internal/selfreview/bundle_test.go:241 | the absent-git-log guard is t.Errorf (non-fatal), so execution reaches the slice at strings.Index -1 and the test panics, masking its diagnostic |   |
| F2 | primary+principles | Minor | stats/internal/selfreview/bundle_test.go:255 | the comment claims an unset origin/HEAD degrades the same way, but every test creates origin/HEAD — nothing pins the scenario the comment names |   |
| F3 | primary | Minor | stats/internal/selfreview/git.go:288 | firstReadableRepo stops at the first readable repo, never falling through to later repos that carry the change branch — dormant today, asymmetric with archiveRepo probing every repo |   |
| F4 | principles | Minor | stats/internal/selfreview/git.go:277 | the git readability probe now lives in two inverted loops; a probe change must be applied twice, and TestUnreadableRepoProbeIsGitDir pins only one of them |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 deferred the asymmetry is dormant — both call sites pass exactly one repository, the main checkout
finding-status: F4 deferred cosmetic duplication of a three-line probe; no behaviour either way

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-1.sh

## Pass log

### Round 0

- diff size 103, under cap — proceeding
- docs-only: no (stats/internal/selfreview/bundle_test.go) — resolved roster runs
- no addition this round — the resolved list ran alone
- rebase at panel entry: origin/main had moved 2 commits, no path overlap — rebased clean, base now 669b8960aee09cec19e9ceae687939045774f4d0
- panel bundle ran 1049s — past the 15-minute ceiling; completed with both report files, so the breach path (re-dispatch a non-returning slot) did not apply; elapsed recorded here

## Branch log

commit b3738881204ca9619ba5471b75b969281228fe5c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 21:26:14 2026 +0300

    test(selfreview): fail the fallback guard fatally and drop its unpinned scenario claim

 stats/internal/selfreview/bundle_test.go | 5 ++---
 1 file changed, 2 insertions(+), 3 deletions(-)

commit bb49547bbc61a1edb66115d812a11fd5846b8211
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 21:00:20 2026 +0300

    docs(flow-fast): state the three sources a flow-fast bundle can never have

 skills/flow-fast/SKILL.md | 11 ++++++-----
 1 file changed, 6 insertions(+), 5 deletions(-)

commit 4e3e7dff0f04ac306e7bb933b5f0213e74e35b1a
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 20:59:42 2026 +0300

    fix(selfreview): derive the bundle git-log source from the main checkout without an archive

 stats/internal/selfreview/bundle_test.go | 53 ++++++++++++++++++++++++++++++++
 stats/internal/selfreview/git.go         | 39 ++++++++++++++++++++++-
 2 files changed, 91 insertions(+), 1 deletion(-)

## Session narrative

This run implemented KAN-583 in the agents repository: the self-review bundle assembler now derives its git-log source from the main checkout — the change branch's commits against origin/HEAD — when no repository archives the change, and flow-fast's contract now names the three spectre sources (tasks.md, design.md, narrative.md) a flow-fast run can never have, so the four-of-six shortfall the kan-533 bundle reported is by design rather than silent. Where it struggled: an early test edit landed in the main checkout before being reverted and re-applied in the worktree; the panel entry found origin/main moved and rebased clean; the single bundle dispatch ran 1049s, past the 15-minute ceiling, but completed with both reports, so the breach path did not apply and the elapsed time is recorded in the pass log instead; and the saved bundle above still reports the git-log source skipped because the running flowd predates this branch — the fix serves only after this change lands and the daemon is rebuilt, which is also why this file keeps the Branch log section the change itself removes.
