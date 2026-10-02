# Self-review context bundle for kan-817-flow-fix-the-done-when-named-gitignored

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-817-flow-fix-the-done-when-named-gitignored/tasks.md (absent)
skipped: spectre/changes/archive/kan-817-flow-fix-the-done-when-named-gitignored/design.md (absent)
skipped: spectre/changes/archive/kan-817-flow-fix-the-done-when-named-gitignored/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-817-flow-fix-the-done-when-named-gitignored.md

# SDD ledger — kan-817-flow-fix-the-done-when-named-gitignored

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-0-primary-timeout
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: timed-out
- Started: 2026-10-02T19:13:40Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: ba9797c78cb017eaf3687f67a9d5862f2b695d70
- Outcome: completed
- Started: 2026-10-02T19:13:40Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-0-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-02T19:15:58Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-0-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-02T19:15:58Z
- Tokens: not measured

## Dispatch 5 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-0-primary-r2
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-02T19:15:58Z
- Tokens: not measured

## Dispatch 6 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 1f7fdc6a5a1a4dd7ee651a76ac1a5db22ffddc85
- Outcome: completed
- Started: 2026-10-02T19:15:58Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-817-flow-fix-the-done-when-named-gitignored-panel.md

# Review panel — kan-817-flow-fix-the-done-when-named-gitignored

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | stats/internal/guard/donewhenpaths.go:118 | A backticked path wrapped in punctuation is refused even when tracked — dwPaths strips backticks before the edge-punctuation trim, so (`shots/27.png` re-baselined) is judged as a name with stray backticks and a fully compliant tree is refused |   |
| F2 | primary | important | scripts/commit-archive.sh:7 | scripts/commit-archive.sh — the declared contract of commitArchive — was not updated for the Done-when cross-check: no Done-when step in the In-order sequence, no DONE-WHEN-PATH verdict form, no new exit causes |   |
| F3 | primary | minor | stats/internal/guard/donewhenpaths.go:135 | A named dotless file (src/Makefile, docs/LICENSE) is never judged — the last-segment-dot rule answers DONE-WHEN-OK for an untracked gitignored dotless name, the kan-743 shape this guard exists to catch |   |
| F4 | primary | minor | stats/internal/guard/donewhenpaths.go:30 | Fenced code blocks are not honored by the section scanner — a #-line inside a fence closes a live Done-when section early so paths after it go unjudged, and a quoted Done-when template in a fence opens a phantom section |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh

## Pass log

### Round 0

- roster: compact — 76
- diff size: 427/under cap — proceed
- docs-only: exit 1 — first non-doc path scripts/check-done-when-paths.sh; resolved roster ran unchanged
- base moved 6 commits, no overlap — auto-rebased clean; merge base now 8417e830f2fdd7a3e9d03ab5c4521e73e57f79df; planning commit: none — flow-fast keeps the plan under gitignored .superpowers
- primary breached the 15-minute ceiling (16.2 min) — stopped, row closed timed-out, re-dispatching once

### Round 1

- fix round 1 opens inline — F1-F4 to the parent; F4 reproducer bounced once to primary: premise cites scripts/check-done-when-paths.sh:19, content sits at :14
- fix-mutations-total꞉ 4 — three flips each caught by the fix own strengthened cases, one comment-only exemption
- clean re-run close: primary re-run confirmed F1-F4 fixed, no new finding — stage closes
fix-mutation: stats/internal/guard/donewhenpaths.go — the backtick re-trim after the edge trim removed — guard observable 1→0: compliant tree refused pre-fix, DONE-WHEN-OK post-fix; cases 14+15 failed under the flip
fix-mutation: stats/internal/guard/donewhenpaths.go — the last-segment rule reverted to the dot rule — guard observable 0→1: untracked dotless docs/LICENSE unjudged pre-fix, refused post-fix; case 16 failed under the flip
fix-mutation: stats/internal/guard/donewhenpaths.go — the fence skip-guard removed — guard observable 0→1: path after a fence line unjudged pre-fix, judged post-fix; cases 17+18 failed under the flip
fix-mutation: scripts/commit-archive.sh — none — comment-only contract header — no executable behaviour changed
fix-mutations-total: 4
## git log --stat

commit 278cc40cc688aebf72dfe5e9bbd521a839bff12d
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Oct 2 22:19:55 2026 +0300

    test(scripts): companion harness for check-done-when-paths

 scripts/test-check-done-when-paths.sh | 167 ++++++++++++++++++++++++++++++++++
 1 file changed, 167 insertions(+)

commit 1f7fdc6a1818c3a3de3b669d3013127a69cc2e6f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Oct 2 21:59:28 2026 +0300

    fix(guard): judge wrapped backticked and dotless paths, honor fences, state the Done-when step in the archive header

 scripts/check-done-when-paths.sh           | 12 +++++++----
 scripts/commit-archive.sh                  | 18 ++++++++++-------
 stats/internal/guard/donewhenpaths.go      | 23 ++++++++++++++++++---
 stats/internal/guard/donewhenpaths_test.go | 32 ++++++++++++++++++++++++++++++
 4 files changed, 71 insertions(+), 14 deletions(-)

commit ba9797c78cb017eaf3687f67a9d5862f2b695d70
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Oct 2 21:15:49 2026 +0300

    docs(flow): a Done-when names only committed, tracked paths

 skills/flow-contracts/finish-contract-run2.md | 9 +++++++++
 skills/flow/archive.md                        | 4 +++-
 skills/flow/brainstorm-planner.md             | 5 +++++
 3 files changed, 17 insertions(+), 1 deletion(-)

commit ed0fcf4c718c21f5b5468147255ac0fdb82c2806
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Oct 2 21:14:49 2026 +0300

    feat(guard): the archive commit refuses a Done-when naming an untracked path

 stats/internal/guard/commit_archive_test.go | 20 ++++++++++++++++++++
 stats/internal/guard/commitarchive.go       | 23 +++++++++++++++++++----
 2 files changed, 39 insertions(+), 4 deletions(-)

commit 9bb1ecca5cf2824ee858f63dbdadf5f6f6dad2f9
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Oct 2 21:14:49 2026 +0300

    feat(scripts): shim check-done-when-paths over the flow-guard registry

 scripts/check-done-when-paths.sh | 40 ++++++++++++++++++++++++++++++++++++++++
 1 file changed, 40 insertions(+)

commit 5f2787bb2df2e491e1b9172fc05c460a60e817c7
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Oct 2 21:14:49 2026 +0300

    feat(guard): check-done-when-paths refuses a Done-when naming an untracked path

 stats/internal/guard/donewhenpaths.go      | 141 ++++++++++++++++++++++
 stats/internal/guard/donewhenpaths_test.go | 185 +++++++++++++++++++++++++++++
 2 files changed, 326 insertions(+)

## Session narrative

The run implemented KAN-817 in one inline pass: a Go guard `check-done-when-paths` (scan every tracked markdown file's `## Done when` sections, extract path-like tokens, refuse any path the index does not track), a shim over the flow-guard registry, in-process wiring into `commitArchive` before the ledger copy, and the rule's prose at three homes (authoring in brainstorm-planner, enforcement canonical in finish-contract-run2, one citing line in archive.md), every new sentence carrying a `(`skills/….md`)` citation so check-verbatim-moves reads it as allowed new text. It struggled most with instrument drift: the store rejected every dispatch row recorded with the decision pair (opus/medium, opus/low) because harness zcode enforces glm-5.3-flash/high, leaving begins journalled and the findings-closed guard blind until the rows were re-recorded under the mapped pair and eight orphaned journal ends cleared; and the panel's own reproducers went ambiguous twice — once from a drifted premise line (bounced to the raising slot, repaired), once because the fix moved the lines their declarations cited (re-pointed: demonstrates to pre-fix sites, premises to lines identical in both trees). The first primary dispatch breached the 15-minute ceiling and was re-dispatched; the second returned clean in 11.5 minutes, raised F1-F4 (two Important, two Minor), all fixed inline in round 1 with each mutation flip caught by the strengthened table cases, and the round-1 re-run confirmed every fix with no new finding.
