# Self-review context bundle for kan-585-flow-improvement-consume-served-vocabulary-from

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-585-flow-improvement-consume-served-vocabulary-from/tasks.md (absent)
skipped: spectre/changes/archive/kan-585-flow-improvement-consume-served-vocabulary-from/design.md (absent)
skipped: spectre/changes/archive/kan-585-flow-improvement-consume-served-vocabulary-from/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-585-flow-improvement-consume-served-vocabulary-from.md

# SDD ledger — kan-585-flow-improvement-consume-served-vocabulary-from

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T18:12:27Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T18:36:33Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T18:36:33Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-585-flow-improvement-consume-served-vocabulary-from-panel.md

# Review panel — kan-585-flow-improvement-consume-served-vocabulary-from

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | scripts/check-installed-rules.sh:205 | managed_files parse word-splits only the first matching line; a multi-line declaration is silently truncated, guard scans a subset and exits 0 over an unscanned file |   |
| F2 | primary | Minor | scripts/check-installed-rules.sh:205 | grep -m1 anchors the first line carrying the pattern anywhere, so a commented-out stale declaration is parsed as the served set |   |
| F3 | primary | Minor | scripts/check-self-review-report.sh:3 | header still hardcodes the count as five (now served by the table) and points above at a parse defined below it |   |
| F4 | principles | Important | scripts/check-installed-rules.sh:205 | the parse accepts a declaration line it cannot fully see and emits silent green, contradicting its own never-a-guess refusal contract |   |
| F5 | principles | Minor | scripts/check-installed-rules.sh:205 | first-match anchoring accepts a comment as the served source (silent stale serve) |   |
| F6 | principles | Minor | scripts/check-self-review-report.sh:3 | header re-states the served angle count as fixed five plus a wrong-direction parse pointer |   |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-principles-3.sh

## Pass log

### Round 0

- roster: compact — 29
- diff size 390 — under cap
- docs-only: no — first non-doc path scripts/check-installed-rules.sh; resolved roster runs unchanged
- no addition this round — the resolved list ran alone

### Round 1

- rerun: delta — F1/F4 Important to the fix; F2/F3/F5/F6 Minor fixed inline (trivial, same lines)

## Branch log

commit 4ac160e8a81128b0f59f0d55e928d497ad4e1d8e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 21:36:06 2026 +0300

    fix(scripts): refuse a truncated managed-files declaration and serve the guard header's count

 scripts/check-installed-rules.sh      | 15 +++++++++++++-
 scripts/check-self-review-report.sh   | 10 ++++++----
 scripts/test-check-installed-rules.sh | 37 +++++++++++++++++++++++++++++++++++
 3 files changed, 57 insertions(+), 5 deletions(-)

commit ec79b1c66e6a709c6ce129ab4b41bb824f3a723e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 21:09:08 2026 +0300

    fix(scripts): resolve the installed-rules guard's managed files from setup.sh

 scripts/check-installed-rules.sh      |  64 +++++++++++++++++---
 scripts/test-check-installed-rules.sh | 108 ++++++++++++++++++++++++++++++++++
 2 files changed, 164 insertions(+), 8 deletions(-)

commit 20edaa3e5054711b08fa40cbf8e1249a7172dabd
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 21:04:25 2026 +0300

    feat(scripts): resolve the report guard's angle labels from the canonical table

 scripts/check-self-review-report.sh      |  89 ++++++++++++++++-----
 scripts/test-check-self-review-report.sh | 129 +++++++++++++++++++++++++++++++
 2 files changed, 200 insertions(+), 18 deletions(-)

## Session narrative

This /flow-fast run implemented KAN-585 inline in the worktree: it audited every guard in scripts/ for a vocabulary checked against a copy that could lag its in-tree serving source, found two live instances, and rewired both. Task 1 made scripts/check-self-review-report.sh parse ANGLE_LABELS from the canonical angle table in skills/flow-contracts/finish-contract-run2.md (step 9) instead of hardcoding it, with a CHECK_SELF_REVIEW_ANGLES_CONTRACT override the harness alone sets, and made the per-report state reset derive its slot count from the parsed table; task 2 made scripts/check-installed-rules.sh parse its managed harness-file set from setup.sh's own `local managed_files=(...)` declaration, which closed a real drift — the guard's hardcoded pair predated `~/.zcode/AGENTS.md` and had never scanned it. Both harnesses gained cases proving source-following and both new refusal paths, mutation-measured. The round-0 panel (compact, primary+principles) raised six findings; the two Important ones (a multi-line managed_files declaration would have been silently truncated to a green subset scan) and four Minors were fixed in one pathspec-scoped commit, mutation-proved, and both slots' delta re-runs confirmed every finding fixed with reproducers exiting 0 and no new defects. Where the run struggled: the plan's first shape check failed on indented fields (column-0 grammar), the citation pre-check's exit 1 initially truncated a `set -e` compound command, and a transient store outage mid-panel resolved via the daemon's own journal reconcile — none reached the branch.
