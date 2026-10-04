# Self-review context bundle for kan-788-flow-cost-start-runs-on-the-cheap-panel-shape

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-788-flow-cost-start-runs-on-the-cheap-panel-shape/tasks.md (absent)
skipped: spectre/changes/archive/kan-788-flow-cost-start-runs-on-the-cheap-panel-shape/design.md (absent)
skipped: spectre/changes/archive/kan-788-flow-cost-start-runs-on-the-cheap-panel-shape/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-788-flow-cost-start-runs-on-the-cheap-panel-shape.md

# SDD ledger — kan-788-flow-cost-start-runs-on-the-cheap-panel-shape

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-04T20:45:19Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-04T20:46:28Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-1-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: fc2d935d
- Outcome: completed
- Started: 2026-10-04T21:14:52Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-788-flow-cost-start-runs-on-the-cheap-panel-shape-panel.md

# Review panel — kan-788-flow-cost-start-runs-on-the-cheap-panel-shape

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | critical | skills/flow/review-panel.md:282 | check-verbatim-moves.sh exits 1 with 16 violations — the diff rewords run-loaded sentences and adds new ones, no verbatim-moves.txt exists for this change, and tasks 1/2 verify steps plus their Build: green fields are therefore false |   |
| N1 | primary | minor | skills/flow/review-panel.md:324 | the renderer paragraph still describes a rendered path only where a dispatch bundles it; silent on the mutation-alone render the fix introduced |   |
| F2 | primary | important | skills/flow/brainstorm-planner.md:447 | Decide still states one dispatch per re-running role — the dropped shape, contradicting the amended bundled re-run contract a live run reads beside it |   |
| N2 | primary | minor | verbatim-moves.txt:19 | the guard stderr summary line captured into the audit record — capture debris, and the source of the report off-by-one count |   |
| F3 | primary | important | stats/internal/guard/renderslotprompt.go:104 | the new static overflow dispatch is mutation alone (the tree prints it), but render-slot-prompt.sh refuses mutation alone (exit 2, verified against the real shim) while the contract mandates every dispatch is rendered and stops the round on refusal — the standard full panel cannot be dispatched as written |   |
| N3 | principles | minor | skills/flow/review-panel.md:324 | Least Astonishment — the contract text does not state the mutation-alone render behavior the fix introduced |   |
| F4 | primary | minor | skills/flow/review-panel.md:236 | one dispatch, never bundled. keeps the deleted exception vocabulary 46 lines above the new bundled-dispatch starting shape |   |
| F5 | principles | critical | .flow/project.md:136 | a mandatory ## lint guard (check-verbatim-moves.sh, expected to exit 0) is red on the delivered tree; the lint-fix-priority invariant forbids claiming done over it |   |
| F6 | principles | important | skills/flow/brainstorm-planner.md:447 | DRY / single source of truth — the re-run shape is restated in three run-loaded files; this diff amended two and the third drifted to the dropped rule |   |
| F7 | principles | minor | skills/flow/review-panel.md:236 | self-contradicting neighbours in the same amended file — Least Astonishment |   |

findings-total: 10
finding-status: F1 fixed
finding-status: N1 fixed
finding-status: F2 fixed
finding-status: N2 fixed
finding-status: F3 fixed
finding-status: N3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed

reproducers-total: 10
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary+principles-1.sh
finding-reproducer: N1 .superpowers/sdd/reproducers/1-primary+principles-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary+principles-2.sh
finding-reproducer: N2 .superpowers/sdd/reproducers/1-primary+principles-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary+principles-3.sh
finding-reproducer: N3 .superpowers/sdd/reproducers/1-primary+principles-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary+principles-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary+principles-5.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-primary+principles-6.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-primary+principles-7.sh

## Pass log

### Round 0

- diff size 56 lines, under cap; docs-only exit 1 — first non-doc path scripts/plan-class.sh — resolved compact roster runs; no addition this round — the resolved list ran alone

### Round 1

- fix round inline (execution inline, no subagent): F1-F7 addressed in 8d15d79e; re-run roster primary+principles bundled in one dispatch (the amended rule), each on its slot-delta-1 diff (held sha fc2d935d), 58 lines under cap; docs-only exit 1 — roster unchanged; FIX_BASE fc2d935d
fix-mutation: stats/internal/guard/renderslotprompt.go — re-added the mutation-alone refusal — TestRenderSlotPromptUnresolvedPlaceholder
fix-mutation: verbatim-moves.txt — guard observable measured on both sides — 16→0 check-verbatim-moves.sh violations
fix-mutation: skills/flow/brainstorm-planner.md — none — prose reword; no executable behaviour
fix-mutation: skills/flow/review-panel.md — none — prose reword; no executable behaviour
fix-mutations-total: 4
## git log --stat

commit 32e7f20f6d427521ea1efe6af53a8cf9035bb637
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Oct 5 00:15:43 2026 +0300

    fix(flow): review Minors

 skills/flow/review-panel.md | 3 ++-
 1 file changed, 2 insertions(+), 1 deletion(-)

commit 8d15d79e115ddfa93029e261a2a7ab7f1b4227c5
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 23:51:54 2026 +0300

    fix(flow): close the bundled-dispatch round's findings

 scripts/render-slot-prompt.sh                   |  6 ++++--
 skills/flow/brainstorm-planner.md               |  5 +++--
 skills/flow/review-panel.md                     |  2 +-
 stats/internal/guard/render_slot_prompt_test.go | 22 ++++++++++++++++++++--
 stats/internal/guard/renderslotprompt.go        | 23 ++++++++++++++---------
 5 files changed, 42 insertions(+), 16 deletions(-)

commit fc2d935ddea318c4b94ab27e7682109997a187cc
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 23:21:14 2026 +0300

    feat(flow): fill the static dispatch to its cap before opening a second

 scripts/plan-class.sh                   | 11 ++++++-----
 stats/internal/guard/plan_class_test.go |  4 ++--
 stats/internal/guard/planclass.go       |  9 +++++----
 3 files changed, 13 insertions(+), 11 deletions(-)

commit d8fb9af8202dc968a1ea4b34490fe6a501afb0e5
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 23:18:59 2026 +0300

    docs(flow): re-run a decided panel's roles in one bundled dispatch

 skills/flow/review-panel-fix-round.md | 10 ++++++----
 1 file changed, 6 insertions(+), 4 deletions(-)

commit ce2e9e402aaed90db0b49f6a3ea9d65b8208fcce
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 23:18:33 2026 +0300

    docs(flow): start panels on one bundled dispatch, full-roster the fallback

 skills/flow/review-panel.md | 22 ++++++++++++++--------
 1 file changed, 14 insertions(+), 8 deletions(-)

## Session narrative

A `/flow-fast` creating run for KAN-788: make one bundled panel dispatch the starting shape, slot-delta-scoped reviews and throwaway mutation copies explicit, and the full-roster shape the separation fallback. Inline execution, class small, compact panel — so the run's own review dogfooded the amended rule: round 0 dispatched primary+principles as one bundle, and the round's findings (the missing verbatim-moves.txt acknowledgement, a drifted restatement of the dropped per-role re-run shape in brainstorm-planner.md's Decide step, render-slot-prompt refusing the mutation-alone dispatch the new tree prints, and stale "never bundled" vocabulary) were fixed inline and confirmed by a bundled delta re-run, whose own three Minors closed inline at the round's close. Where it struggled: the mutation proof's hand-built patch hunks were refused twice on hunk-header counts before being generated from the file's real bytes; three reproducers were refused on their pinned citations because the fix had edited the very lines they demonstrated, and were re-authored against the defect-present tree per the re-author rule, both legs proved with prove-reproducer.sh; and check-verbatim-moves.sh had to be re-run and the acknowledgement file extended after each further prose edit, since the fix round's own wording was itself run-loaded corpus.
