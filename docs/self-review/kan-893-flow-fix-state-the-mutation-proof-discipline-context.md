# Self-review context bundle for kan-893-flow-fix-state-the-mutation-proof-discipline

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-893-flow-fix-state-the-mutation-proof-discipline/tasks.md (absent)
skipped: spectre/changes/archive/kan-893-flow-fix-state-the-mutation-proof-discipline/design.md (absent)
skipped: spectre/changes/archive/kan-893-flow-fix-state-the-mutation-proof-discipline/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-893-flow-fix-state-the-mutation-proof-discipline.md

# SDD ledger — kan-893-flow-fix-state-the-mutation-proof-discipline

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-06T11:56:20Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-893-flow-fix-state-the-mutation-proof-discipline-panel.md

# Review panel — kan-893-flow-fix-state-the-mutation-proof-discipline

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | skills/flow/review-panel.md:836 | The restore-semantics claim is overstated in both files: git checkout -- restores from the index, so a staged-but-uncommitted fix survives a flip restore, and through mutate-and-verify.sh an uncommitted fix is refused (exit 2) before anything mutates — the death scenario is the unstaged manual flip only |   |
| F2 | principles | Minor | skills/flow/review-panel.md:836 | Least Astonishment: the prose mechanism claim contradicts actual behavior — a staged uncommitted fix survives git checkout -- and the script itself refuses a dirty touched file, so an uncommitted fix cannot die by restore through the guard; doc precision, reproducer and fix wording in panel-report-0-primary.md |   |
| F3 | primary | Minor | .superpowers/sdd/kan-893-flow-fix-state-the-mutation-proof-discipline/tasks.md:6 | tasks.md's header misstates the overlap — it says kan-891 and kan-903 both edited review-panel.md; kan-903 edited only .flow/project.md |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F3 none — the defect is a false sentence in the fix-run plan header, evidenced by git show 54d14441 --stat (one file, .flow/project.md)

## Pass log

### Round 0

- auto-resolved: the base branch has moved and touches paths this change also touched (origin/main +2: kan-891, kan-903, both editing skills/flow/review-panel.md) → Stop
- roster: full
- diff size: 12 lines, under cap — proceed
- docs-only: exit 1 — first non-documentation path scripts/mutate-and-verify.sh; resolved roster runs unchanged
- no addition this round — the resolved list ran alone
- decided under: rebase onto main per explicit operator instruction; working-notes merge base now 740712e6
## git log --stat

commit 114911176f3385a565ca9e968d251e603734e664
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 15:17:44 2026 +0300

    fix(flow): review Minors
    
    Panel round 0 (F1, F2): the restore-semantics claim stated a mechanism
    git checkout -- does not have — it reverts the file to its index state, so
    an unstaged fix dies with the restore, a staged one survives, and
    mutate-and-verify.sh itself refuses a dirty touched file (exit 2) before
    anything mutates. Both sentences now state the enforced behavior.
    verbatim-moves.txt re-lists the replacement sentence.

 scripts/mutate-and-verify.sh | 12 +++++++-----
 skills/flow/review-panel.md  |  4 ++--
 2 files changed, 9 insertions(+), 7 deletions(-)

commit fadf906c182a39c4fdaf946267ec359d19a1292d
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 01:46:30 2026 +0300

    docs(flow): state the commit-before-flip mutation-proof discipline
    
    kan-795's fix round restored a mutation flip with git checkout -- before
    the fix was committed, discarding the uncommitted fix so one flip passed
    against unmutated code. State the recovery discipline it adopted in both
    places the mutation-proof guidance lives: the fix commit lands before the
    first flip, and every flip is asserted landed - the mutated bytes read
    back from disk - before its killing test's failure is trusted as proof.
    
    Beside the MUTATION PROOF paragraph (skills/flow/review-panel.md), which
    every fix dispatch carries, and in the mutate-and-verify contract
    (scripts/mutate-and-verify.sh's header, which the Go guard names as the
    contract). The new run-loaded sentence is listed in verbatim-moves.txt.
    Prose only; no guard logic changes.

 scripts/mutate-and-verify.sh | 6 +++++-
 skills/flow/review-panel.md  | 6 +++++-
 2 files changed, 10 insertions(+), 2 deletions(-)

## Session narrative

The first run implemented the change (commit 10ee3cfe) and stopped at the review panel: the base-movement check found origin/main had gained kan-891 and kan-903 mid-run, both — so the stop handoff then believed — editing skills/flow/review-panel.md, which this branch also edits; the overlap prompt auto-resolved on its recommended Stop. This fix run rebased under an explicit operator instruction (clean, replayed as fadf906c), and the panel round-0 reviewers then established from the commits themselves that only kan-891 touched review-panel.md (kan-903 touched .flow/project.md only) — the stop was grounded on an overlap claim one commit wider than the truth, a fact worth carrying to any future base-movement messaging. Two approaches were tried and abandoned: running the verbatim-moves guard via the main checkout's scripts/check-verbatim-moves.sh scored the main checkout's corpus (FLOW_GUARD_REPO_ROOT resolves from the script's own path, not the cwd) and reported vacuous 0-violation passes — including against a deliberately removed listing — until the run switched to the worktree's own copy, which FAILed the new sentence exactly as designed; and writing the MUTATION PROOF discipline as stated was twice revised by review — the panel Minors corrected the restore semantics (git checkout -- reverts to the index, not the last commit; the guard itself refuses a dirty touched file), so the landed sentences state the enforced behavior rather than the first draft's overstatement. Verification: full ## lint (27 commands) green in the worktree, guard-test suite 40/40, the branch record guard clean, findings F1-F3 fixed inline and recorded fixed.
