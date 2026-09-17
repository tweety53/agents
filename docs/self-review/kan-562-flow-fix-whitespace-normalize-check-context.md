# Self-review context bundle for kan-562-flow-fix-whitespace-normalize-check

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-562-flow-fix-whitespace-normalize-check/tasks.md (absent)
skipped: spectre/changes/archive/kan-562-flow-fix-whitespace-normalize-check/design.md (absent)
skipped: spectre/changes/archive/kan-562-flow-fix-whitespace-normalize-check/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-562-flow-fix-whitespace-normalize-check.md

# SDD ledger — kan-562-flow-fix-whitespace-normalize-check

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: sonnet effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T22:55:03Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: haiku effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T23:21:17Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: haiku effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-09-17T23:21:17Z
- Tokens: not measured

## Dispatch 4 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: haiku effort=low
- Commit: cabedec
- Outcome: completed
- Started: 2026-09-17T23:28:22Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-562-flow-fix-whitespace-normalize-check-panel.md

# Review panel — kan-562-flow-fix-whitespace-normalize-check

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | important | scripts/check-task-commit-fields.py:1171 | the tree search swapped a mechanism that read every searchable blob for one that silently reads less than the tree: git archive honors export-ignore and export-subst, so on repos that use them the guard answers not-found-in-tree about content that IS in the tree, with no observable signal |   |
| F2 | primary | minor | scripts/check-task-commit-fields.py:1171 | the whole tree is buffered in memory per check where git grep streamed |   |
| F3 | primary+principles | minor | scripts/check-task-commit-fields.py:1191 | the fold rule 'every whitespace run becomes one space' is written twice, as the same regex in two places (:1191 and :1224), so a future change to the fold rule must land twice |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 deferred — the guard already buffers whole diffs in memory; tree-scale buffering matches its existing runtime profile and no repo-scale tree is in its corpus
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-0-3.sh
finding-reproducer: F2 none — resource profile
finding-reproducer: F3 none — structural nit

## Pass log

### Round 0

- x
- probe row superseded — real round-0 notes follow
- base moved: 3 commits on origin/main since merge base 95a2ef81c20929d50f88277f87399abb22e01725, no overlap with this change's paths — continued
- diff-size 168, under cap; docs-only: no (scripts/check-task-commit-fields.py); citation trigger: exit 1, no citation check
- roster: compact — 6; one bundle primary+principles, free grouping: compact roster is the floor bundle alone

### Round 1

- FIX_BASE 4a1abe2; inline panel-fix (execution inline, no subagent): rewrote _folded_tree_text onto git ls-tree -r -z + git cat-file --batch — neither applies export attributes, so the search sees every blob git grep saw; path filter moved into Python (git ls-tree supports no exclude magic); fold rule stated once in _fold_ws_bytes (F3); harness Case 136 pins the seen-every-blob property
- rerun cap from held sha 4a1abe2: 138 under cap; docs-only round 1: no (scripts/check-task-commit-fields.py); re-running slots: primary (F1 F2 F3) and principles (F1 F3), one dispatch each on the rerun pair, targeted at fix-round-1.diff
- re-runs clean: primary F1 fixed F3 fixed, principles F1 fixed F3 fixed, no new defect at the reviewed sites; one pre-existing observation (run_git strict text decode on non-locale-encoded filenames) recorded in both reports, fail-loud, not new
fix-mutation: scripts/check-task-commit-fields.py — blob contents dropped from the folded tree text (MUTANT marker), confirmed landed by grep — scripts/test-check-task-commit-fields.sh — cases 1/3/5/7 fail with not-found-in-the-tree
fix-mutation: scripts/check-task-commit-fields.py — none — no executable behaviour changed outside the tree search; the diff-check and Case-label sites are untouched by the fix diff
fix-mutations-total: 2

## Branch log

commit cabedecafe53ecdfac9b06ad6997f868e8963a08
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 02:20:25 2026 +0300

    fix(commit-fields): search tree blobs through ls-tree and cat-file

 scripts/check-task-commit-fields.py      | 111 ++++++++++++++++++++-----------
 scripts/test-check-task-commit-fields.sh |  27 ++++++++
 2 files changed, 98 insertions(+), 40 deletions(-)

commit 4a1abe2c5da08500f38777cac9d621bc6de5f1e4
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 01:51:57 2026 +0300

    fix(commit-fields): whitespace-normalize the Tests: match

 scripts/check-task-commit-fields.py | 114 +++++++++++++++++++++++++++---------
 1 file changed, 86 insertions(+), 28 deletions(-)

commit 852cb93cf2d35b2dc7d3a870ca5386532a63e23c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 18 01:48:05 2026 +0300

    test(commit-fields): cover the wrapped Tests: sentence shape

 scripts/test-check-task-commit-fields.sh | 54 ++++++++++++++++++++++++++++++++
 1 file changed, 54 insertions(+)

## Session narrative

This run implemented KAN-562 — whitespace-normalize check-task-commit-fields.sh's Tests: match — inline under the project's dynamic toggles, which resolved to class small, an inline execution, and a compact review panel (one primary+principles bundle). TDD held: harness Cases 134–135 were written and seen red first (Case 134 failing at both match sites), then the fix folded whitespace on both sides of both matches — a `\s+`-joined escaped-word needle over a marker-stripped diff for the diff check, and a folded-content search replacing line-based `git grep -F` for the tree check. The panel's one Important finding was real and mine: my first tree-check rewrite reached for `git archive`, which honors export-ignore/export-subst and would have silently narrowed the search on repos using those attributes — the reviewer's reproducer demonstrated it on a fixture before the fix round replaced the mechanism with `git ls-tree -r` + `git cat-file --batch` (neither applies export attributes), moved the path filter into Python after probing that `git ls-tree` supports no exclude pathspec magic, and added Case 136 to pin the seen-every-blob property; the fix was mutation-proved and confirmed by a clean targeted re-run from both raising roles. Where it struggled: two probes mis-fired before landing (a `flow stage end -harness` flag the end form does not take, and a junk `flow record pass` probe row that stays in the ledger), and the first relocate-comparison call passed four arguments to a three-argument script — all corrected in the same run. F2 (memory buffering) was deferred with reason; everything else the panel raised is fixed and re-verified.
