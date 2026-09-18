# Self-review context bundle for kan-568-flow-fix-define-the-mutation-reproducer

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-568-flow-fix-define-the-mutation-reproducer/tasks.md (absent)
skipped: spectre/changes/archive/kan-568-flow-fix-define-the-mutation-reproducer/design.md (absent)
skipped: spectre/changes/archive/kan-568-flow-fix-define-the-mutation-reproducer/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-568-flow-fix-define-the-mutation-reproducer.md

# SDD ledger — kan-568-flow-fix-define-the-mutation-reproducer

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-18T18:39:20Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: not recorded
- Started: 2026-09-18T19:03:48Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-18T19:13:49Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-568-flow-fix-define-the-mutation-reproducer-panel.md

# Review panel — kan-568-flow-fix-define-the-mutation-reproducer

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | scripts/run-reproducer.sh:248 | a comment-only edit adding the mutation declaration to a dispatched reproducer file between the dispatch-time run and the fix-round re-run re-vocabularies the KAN-524 comparison: an unchanged reproducer (exit 5 both rounds) reads as fix verified (runner exit 1), because the refusal compares verdicts under whatever convention the file declares at re-run time |   |
| F2 | primary | minor | scripts/check-panel-reproducer-exit-contract.sh:185 | the guard's exit-1 message 'an open finding's reproducer must exit non-zero here' is backwards for a declared mutation reproducer, whose healthy not-demonstrated reading is a non-zero exit; guard behavior is correct (verdict-based), only the message inverts |   |
| F3 | primary | minor | skills/flow/review-panel.md:378 | the absolute fixed-and-one-directional exit-code contract carried verbatim on every slot dispatch now has an exception (the mutation-reproducer convention) stated only in the mutation-testing brief; the contract paragraph carries no cross-reference to it |   |
| F4 | primary | minor | scripts/test-run-reproducer.sh:633 | fixture() prepends shebang and RAN-marker lines, so case 28.a with 12 filler lines puts the marker at line 15 and any detection window of 1-14 lines passes; the documented 10-line edge of the window is unpinned — add marker-at-line-10 and marker-at-line-11 boundary cases |   |
| F5 | principles | minor | skills/flow/review-panel.md:543 | DRY/SSoT: the exact literal # mutation-reproducer and the 10-line window are each stated in skills/flow/review-panel.md and scripts/run-reproducer.sh with no mechanical pin keeping them together (check-dispatch-paragraphs.sh covers other required phrases, not this one) |   |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 deferred message-wording only — the guard verdict is already correct; the text fix would touch harness-pinned wording
finding-status: F3 fixed
finding-status: F4 deferred needs new boundary tests in the harness, which is not an inline-trivial edit
finding-status: F5 deferred needs a new mechanical pin inside a guard plus its own test, a judgment call beyond the inline bar

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 none — prose coherence: the contract paragraph is now qualified by an exception stated only in the mutation-testing brief, and a cross-reference is not runnable
finding-reproducer: F4 none — coverage gap: the fix is new boundary cases in the harness itself
finding-reproducer: F5 none — future-drift finding: the two copies of the marker literal and window currently agree (grepped), so no present defect is runnable

## Pass log

### Round 0

- roster: compact — compact_roll 74 < 90 (class small); grouping static (bundle_roll 3 < 30); experimental_roll 39 ≥ 30 — no experimental slot
- diff size 154 lines — under cap, proceed
- docs-only exit 1 — resolved roster runs unchanged; first non-documentation path scripts/run-reproducer.sh
- no addition this round — the resolved list ran alone

### Round 1

- primary re-runs — it raised F1 (fixed); principles not re-run — its only finding F5 was deferred, no fix of its own to re-review
fix-mutation: scripts/run-reproducer.sh — inverted the --reproducer-sha pin comparison (!= to =), so a matching pin refuses and a mismatched pin runs — test-run-reproducer cases 29 and 30 (both failed against the mutation)
fix-mutation: skills/flow/review-panel.md — none — docs-only fix — the contract and cross-reference sentences name no executable behaviour
fix-mutations-total: 2

## Branch log

commit 39567cb0f46bbea4e688deabad76dc25ec6c9b2f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 22:21:30 2026 +0300

    docs(review-panel): root the mutation-marker citation to the runner

 skills/flow/review-panel.md | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)

commit 51899c906334ba56eb5e0c77e474661ea3f821a2
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 22:13:24 2026 +0300

    fix(run-reproducer): pin the reproducer identity across fix-round re-runs

 scripts/lib/sha256-hex.sh                          | 25 +++++++
 scripts/run-reproducer.sh                          | 84 +++++++++++++++++++---
 .../test-check-panel-reproducer-exit-contract.sh   |  2 +
 scripts/test-run-reproducer.sh                     | 43 +++++++++++
 skills/flow/review-panel.md                        | 14 +++-
 5 files changed, 154 insertions(+), 14 deletions(-)

commit 40459cc349cd015ac9b627bf90956da9d41a4c8a
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 21:34:44 2026 +0300

    docs(review-panel): declare the mutation-reproducer marker in the mutation brief

 skills/flow/review-panel.md | 5 ++++-
 1 file changed, 4 insertions(+), 1 deletion(-)

commit 2b5635be062422bcbb4f074f79be18d6b1a19956
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 21:34:27 2026 +0300

    feat(run-reproducer): define the mutation-reproducer convention

 scripts/run-reproducer.sh      | 69 ++++++++++++++++++++++++++++++++----
 scripts/test-run-reproducer.sh | 80 ++++++++++++++++++++++++++++++++++++++++++
 2 files changed, 142 insertions(+), 7 deletions(-)

## Session narrative

This run implemented KAN-568 inline: run-reproducer.sh gained the mutation-reproducer convention — a reproducer declaring the exact line `# mutation-reproducer` within its first 10 lines is read inverted (exit 0 = defect present), closing the hand-verified-substitution gap kan-485 exposed — with TDD harness cases written red first, and one sentence in review-panel.md's mutation-testing brief declaring the marker. The whole-branch panel (primary+principles, one bundled dispatch) raised five findings; four Minors were deferred with categories by the default rule, and the one Important finding — that a comment-only re-declaration of a dispatched reproducer between dispatch and fix-round re-run silently re-vocabularies the KAN-524 ambiguity comparison — drove a mechanical fix: the runner now prints a reproducer sha with every verdict and `--reproducer-sha` pins the re-run to the dispatch-time file, refusing a mismatch before execution; the fix-round contract in review-panel.md requires the pin and names the re-authoring path a refusal puts a reproducer on, which this round then exercised on its own finding. The pin was mutation-proved (inverted comparison caught by both new harness cases), the re-authored reproducer flips to defect-absent on the fixed tree, and the rerun reviewer confirmed the fix. Where it struggled: the verification contract for the finding initially looked unfillable by the prose fix the reviewer suggested, because a stateless runner cannot detect a re-authored file — the sha pin was the smallest mechanism that makes the refusal fire, and the citation-lint guard then caught a bare-path citation in the brief sentence, fixed in a follow-up commit.
