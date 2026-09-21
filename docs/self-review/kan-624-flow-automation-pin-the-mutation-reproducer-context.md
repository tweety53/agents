# Self-review context bundle for kan-624-flow-automation-pin-the-mutation-reproducer

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-624-flow-automation-pin-the-mutation-reproducer/tasks.md (absent)
skipped: spectre/changes/archive/kan-624-flow-automation-pin-the-mutation-reproducer/design.md (absent)
skipped: spectre/changes/archive/kan-624-flow-automation-pin-the-mutation-reproducer/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-624-flow-automation-pin-the-mutation-reproducer.md

# SDD ledger — kan-624-flow-automation-pin-the-mutation-reproducer

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 91a836e53577a6858a91ad915149919cd809cb47
- Outcome: completed
- Started: 2026-09-21T19:33:13Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 9897f3d
- Outcome: completed
- Started: 2026-09-21T20:10:36Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 9897f3d
- Outcome: completed
- Started: 2026-09-21T20:10:36Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-624-flow-automation-pin-the-mutation-reproducer-panel.md

# Review panel — kan-624-flow-automation-pin-the-mutation-reproducer

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Important | scripts/check-mutation-reproducer-pin.sh:146 | The four checks require only at-least-one occurrence of each needle per file, so a single stated site of the four the header names can drift — e.g. the real runner's header comment moved to 'first 12 lines' while code and the second comment still say 10 — and the guard answers MUTATION-REPRODUCER-PIN-OK, exit 0; the panel's mutation-testing brief window statement is split across a line break and is never machine-visible at all. |   |
| F2 | primary+principles | Important | scripts/check-mutation-reproducer-pin.sh:116 | The 'marker not extractable → exit 2' contract cannot fire because sed's no-match answer is the line itself, so a canonical line whose marker is not single-quoted (a realistic re-quote refactor) is misreported as exit-1 drift with the whole line pasted in as the 'marker'. |   |
| F3 | primary | Minor | scripts/test-check-mutation-reproducer-pin.sh:59 | No cleanup trap, so every harness run leaks seven fixture trees under TMPDIR, unlike ten sibling harnesses. |   |
| F4 | primary+principles | Minor | scripts/check-mutation-reproducer-pin.sh:137 | Drift violations are reported as file:0:, a line number that names no editable line, against the project's own 'every guard reports file:line' prose. |   |
| F5 | primary | Minor | scripts/check-mutation-reproducer-pin.sh:215 | The joined-line scan re-reports every single-line near-miss at the previous line's number, so one drift yields a phantom extra row naming a line that carries no statement; report in the joined scan only when the match spans the join boundary. |   |
| F6 | principles | Minor | scripts/check-mutation-reproducer-pin.sh:223 | The near-miss marker regex is lowercase-only, so a case-variant backticked span escapes both presence and near-miss — a narrow residual gap, not the round-0 hole. |   |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 deferred — a case-variant backticked span is prose emphasis, not a convention statement; the pin binds the lowercase statements the canonical line and the prose carry

reproducers-total: 6
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 none — cosmetically duplicated near-miss row at line N-1, demonstrated by the reporter's joined-scan probe; the fix is the scan's boundary condition
finding-reproducer: F6 none — a case-variant backticked span is prose emphasis, not a convention statement; the pin binds the literal lowercase statements both files carry

## Pass log

### Round 0

- roster: compact — 17
- diff-size: 335 under cap — proceed
- not docs-only: scripts/check-mutation-reproducer-pin.sh — resolved roster runs
- no addition this round — the resolved list ran alone

### Round 1

- diff-size re-check at held sha 9897f3d — under cap, proceed; docs-only re-run: still not docs-only
- fix round 1: agents = primary + principles re-runs, targeted (dynamic — one dispatch per role, rerun pair haiku/low, zcode-mapped glm-5.3-flash/high); diff path .superpowers/sdd/fix-round-1.diff
## git log --stat

commit 0e7d067a12b8b86c84c42a5cb12657f7c18b5680
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 23:18:29 2026 +0300

    fix(guards): report a joined near-miss only when it spans the join

 scripts/check-mutation-reproducer-pin.sh | 29 +++++++++++++++++++++--------
 1 file changed, 21 insertions(+), 8 deletions(-)

commit f81c2d699dcc6a004353a2a2e3b3a31eba46e7f9
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 23:10:05 2026 +0300

    fix(guards): pin every mutation-reproducer statement site, not one per file

 scripts/check-mutation-reproducer-pin.sh      | 119 +++++++++++++++++++++-----
 scripts/test-check-mutation-reproducer-pin.sh | 112 +++++++++++++++++++++++-
 2 files changed, 207 insertions(+), 24 deletions(-)

commit 9897f3d3560baabee30a1efdfe242d746c308818
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 22:29:55 2026 +0300

    chore(project-config): run the mutation-reproducer pin guard in lint

 .flow/project.md | 1 +
 1 file changed, 1 insertion(+)

commit e7c88d9337c5458c658a0cf5892bd88b50b9758b
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 22:29:55 2026 +0300

    test(guards): pin the mutation-reproducer marker and window across prose and runner

 scripts/check-mutation-reproducer-pin.sh      | 158 +++++++++++++++++++++++
 scripts/test-check-mutation-reproducer-pin.sh | 176 ++++++++++++++++++++++++++
 2 files changed, 334 insertions(+)

## Session narrative

A `/flow-fast` creating run for KAN-624 (pin the mutation-reproducer marker literal and 10-line window mechanically across prose and runner). The design judgment call: a dedicated guard (`scripts/check-mutation-reproducer-pin.sh`) rather than extending check-dispatch-paragraphs, because the pin binds a code line plus prose statements, not blockquote paragraphs; its source of truth is the runner's canonical declaration line, from which the guard extracts the marker and window and requires every stated site in both files to agree. Implemented test-first (harness RED before the guard existed), committed in the plan's two tasks, then reviewed by the decided panel (primary+principles floor bundle, one dispatch). Round 0 raised 2 Important + 2 Minor; both Importants (presence-only per-file checks blind to single-site drift; sed extraction that could never fail) were fixed inline with a near-miss absence scan over single and line-joined text, grep -oE extraction, anchor-line drift rows and a cleanup trap; both slots re-run targeted on the fix delta and verified all four fixed from their own reproducer runs. The re-runs raised two Minors of their own; the phantom duplicate near-miss row was fixed inline (join-boundary condition), the case-variant marker gap deferred as out-of-scope. Where it struggled: the fix-round reproducer verification hit the runner's 20s bound twice — machine load from concurrent agent sessions, diagnosed by process snapshots and resolved with the runner's own documented RUN_REPRODUCER_BOUND_SECONDS override, never by touching the code under test; and the fresh worktree needed its `## worktree setup` build before go vet/tsc could run. Lint list and the 78-harness guard suite both green at close.
