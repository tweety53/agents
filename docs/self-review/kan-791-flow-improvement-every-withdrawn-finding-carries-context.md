# Self-review context bundle for kan-791-flow-improvement-every-withdrawn-finding-carries

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-791-flow-improvement-every-withdrawn-finding-carries/tasks.md (absent)
skipped: spectre/changes/archive/kan-791-flow-improvement-every-withdrawn-finding-carries/design.md (absent)
skipped: spectre/changes/archive/kan-791-flow-improvement-every-withdrawn-finding-carries/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-791-flow-improvement-every-withdrawn-finding-carries.md

# SDD ledger — kan-791-flow-improvement-every-withdrawn-finding-carries

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-04T22:26:30Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-791-flow-improvement-every-withdrawn-finding-carries-panel.md

# Review panel — kan-791-flow-improvement-every-withdrawn-finding-carries

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | stats/internal/store/records.go:775 | SetFindingStatus's new doc paragraph is concatenated onto the previous one — no blank // line before it, unlike the matching UpsertFinding paragraph |   |
| F2 | primary | Minor | stats/internal/guard/unfinishedwork.go:205 | the rewritten comment says the four-shape status vocabulary is enforced at the store itself since KAN-791, but the store refuses only the bare-withdrawn shape — an HTTP write of bogus or withdrawnfoo still lands |   |
| F3 | primary | Minor | stats/internal/api/records_test.go:126 | the fake's UpsertFinding does not mirror the store's new bare-withdrawn refusal while its SetFindingStatus does — a POST of a bare withdrawn would answer 201 from the fake but 409 in production |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 none — a comment-paragraph asymmetry; the joined paragraphs are records.go:773-776, visible by sed
finding-reproducer: F2 none — an overstated comment claim; the contradicting facts are the absent vocabulary refusal in store.SetFindingStatus and ApplyFindingStatus's non-empty-only check
finding-reproducer: F3 none — demonstrating it requires adding a test to the package, which the read-only review does not do; the divergence is visible by comparing the fake with the store

## Pass log

### Round 0

- no addition this round — the resolved list ran alone
- roster: compact — 29
- base moved with no path overlap — auto-rebased clean; working-notes merge base now 4826fcd3fb7578c28fd27f96c5f7f8e3ea08bb83
- diff size 201 of cap 2000 — under cap; proceed
- docs-only reduction: no — first non-documentation path scripts/check-panel-findings-closed.sh; roster dispatched as resolved: primary+principles (opus/medium decided; glm-5.3-flash/high mapped)
- planning commit: skipped — no spectre plan tree exists on a flow-fast run (planning artifacts are git-excluded .superpowers/), nothing to commit; plan-unchanged and tree-markers brackets recorded
- round 0 closed: bundled dispatch panel-0-primary+principles (glm-5.3-flash/high mapped); principles clean, primary raised F1-F3, all Minor — fixed by the parent inline, no fix round
## git log --stat

commit 2204025da540c390edb1f9071c17b4a094632a2a
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Oct 5 01:29:51 2026 +0300

    fix(stats): review Minors

 stats/internal/api/records_test.go     |  7 +++++++
 stats/internal/guard/unfinishedwork.go | 15 ++++++++-------
 stats/internal/store/records.go        |  1 +
 3 files changed, 16 insertions(+), 7 deletions(-)

commit 088935498086dc478a164d923720f8b91836c037
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Oct 5 01:08:14 2026 +0300

    fix(guard): count a bare withdrawn finding as open at the close gates
    
    KAN-791: both close guards' open-finding predicates drew the closed
    line at the bare word, so a reasonless withdrawal -- the silent
    drop the contract forbids -- read as closed. The predicate now requires
    the reason (the same line the store's withdrawnWithoutReason draws),
    each guard keeping its own copy per the duplicate-the-predicate
    decision; the shim contract headers state the rule. Withdrawn findings
    still claim no verification in the fixed-without-clean-rerun
    correlation.

 scripts/check-panel-findings-closed.sh             |  6 ++++++
 scripts/check-unfinished-work.sh                   |  5 +++++
 .../guard/check_panel_findings_closed_test.go      |  2 ++
 stats/internal/guard/check_unfinished_work_test.go |  6 ++++++
 stats/internal/guard/panelfindingsclosed.go        | 12 ++++++++---
 stats/internal/guard/unfinishedwork.go             | 25 +++++++++++++++++-----
 6 files changed, 48 insertions(+), 8 deletions(-)

commit cf234381207d1b13a99c171a499ec940705df29a
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Oct 5 01:04:18 2026 +0300

    feat(api): answer 409 for a bare withdrawn finding status
    
    KAN-791: ErrWithdrawnReasonMissing is the third member of the
    reached-and-correctly-refused family beside ErrDeferredNotMinor and
    ErrCategoryNotDeferred -- mapping it anything but 409 would have
    internal/client journal a replay the store refuses identically every
    time. The route test's fake mirrors the real store's refusal order:
    category rule first, then the bare-withdrawn rule, before existence.

 stats/internal/api/records_test.go | 26 ++++++++++++++++++++++++++
 stats/internal/api/server.go       | 10 +++++++++-
 2 files changed, 35 insertions(+), 1 deletion(-)

commit 57cbfce09bdb4149f86d13e745acf773b352cf1e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Oct 5 01:02:32 2026 +0300

    feat(store): refuse a bare withdrawn finding status
    
    KAN-791: a finding leaves the board only into fixed or withdrawn
    <reason>. validateFindingStatus already refuses the bare word at the
    CLI, but an HTTP write or a replayed journal entry crosses the store,
    not the validator -- so UpsertFinding and SetFindingStatus now refuse a
    status whose withdrawn prefix carries no reason text, with a typed
    ErrWithdrawnReasonMissing beside the deferred sentinels. The predicate
    (drawn at the missing reason, not at the CLI's spacing rule) is the one
    the close guards' open-finding tests redraw in the same place.

 stats/internal/store/records.go      | 45 +++++++++++++++++++++++++
 stats/internal/store/records_test.go | 64 ++++++++++++++++++++++++++++++++++++
 2 files changed, 109 insertions(+)

## Session narrative

Implemented KAN-791 in three planned commits: the store now refuses a bare `withdrawn` finding status on both write paths (new ErrWithdrawnReasonMissing, the predicate drawn at the missing reason rather than the CLI's spacing rule), the API maps that sentinel to 409 beside its contradiction-class siblings, and both close guards (check-panel-findings-closed, check-unfinished-work) count a reasonless withdrawal as an open finding, their shim headers stating the rule. TDD throughout — the store test was written first and failed on the undefined sentinel; the API route test failed answering 500 before the mapping landed. The round-0 panel (bundled primary+principles) raised three Minors — a joined doc-comment paragraph, an overstated guarantee clause in the unfinished-work comment, and the api test fake not mirroring the raise-path refusal — all fixed by the parent in one review-Minors commit and recorded fixed; the panel found no Critical or Important. Where the run struggled: the first Edit landed on the main checkout's identical copy of records_test.go instead of the worktree's (reverted, redone in the worktree); a backtick in a commit message ran as shell substitution (amended); the base moved under the branch mid-run and the panel preflight was re-run against the new merge base after the automatic clean rebase. The api test fake now mirrors both store refusal rules in the store's own order; the deliberate looser-store-predicate line (withdrawnfoo is the CLI's spacing rule to refuse) is stated in tasks.md and in the guard comments.
