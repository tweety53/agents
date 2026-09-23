# Self-review context bundle for kan-637-flow-stats-app-a-panel-s-deferred-findings-have

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-637-flow-stats-app-a-panel-s-deferred-findings-have/tasks.md (absent)
skipped: spectre/changes/archive/kan-637-flow-stats-app-a-panel-s-deferred-findings-have/design.md (absent)
skipped: spectre/changes/archive/kan-637-flow-stats-app-a-panel-s-deferred-findings-have/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-637-flow-stats-app-a-panel-s-deferred-findings-have.md

# SDD ledger — kan-637-flow-stats-app-a-panel-s-deferred-findings-have

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-0-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-23T20:51:39Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 390eca2cd727ceca7c6c9f05fb45693b557871b3
- Outcome: completed
- Started: 2026-09-23T21:11:01Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-23T21:11:38Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-637-flow-stats-app-a-panel-s-deferred-findings-have-panel.md

# Review panel — kan-637-flow-stats-app-a-panel-s-deferred-findings-have

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow-contracts/jira-followups.md:325 | declares "two loading sites, one contract" but the same file still asserts at :325-326 "Every site that joins is in /flow's integrate run", at :240/:398 that the outstanding list "reaches the planning commit's message… on every route" (false for deferred findings — they never reach a planning commit), and at :293-295/:314 a retry premise ("Run 1 is re-entered…") the panel site has no equivalent of — the ask fires once per close, newly-deferred only, so a partial join there is never re-attempted |   |
| F2 | primary | Important | skills/flow-contracts/SKILL.md:29 | three normative files still state the repealed exclusivity: skills/flow-contracts/SKILL.md:29 ("Loaded by /flow's integrate run and no other command"), rules/flow-manual-review.mdc:56 ("integrate run 1 only"), skills/flow-contracts/jira-integration.md:248-249 ("loaded only by bare /flow run 1") — any of them read normatively cancels the new site |   |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh

## Pass log

### Round 0

- roster: compact — 61
- diff size: 48 changed lines, cap in force 600 — under cap, dispatched unasked
- docs-only reduction: every touched path ends .md/.mdc — pass 1 reduced to primary alone (guard exit 0)
- no addition this round — the resolved list ran alone

### Round 1

- fix round 1 ran inline — parent applied the fix, no subagent; F1 reconciled the contract's per-site statements, F2 updated the three stale exclusivity claims; setup.sh global re-run from main after a worktree-run misinstall was repaired
- re-run cap check vs primary's held sha a26df25: under cap; docs-only guard exit 0 — branch stayed docs-only, reduced roster holds, primary re-runs alone on the rerun pair
- delta re-run clean: F1 fixed, F2 fixed, no new finding at any severity; two cosmetic notes left in the report file only
fix-mutation: skills/flow-contracts/jira-followups.md — none — docs-only fix — no executable behaviour changed; the observable is the finding's reproducer, exit 1 pre-fix, exit 0 post-fix
fix-mutation: rules/flow-manual-review.mdc — none — docs-only fix — observable is check-installed-rules.sh green from main and the worktree skip, plus F2's reproducer exit 1→0
fix-mutations-total: 2
## git log --stat

commit 390eca2cd727ceca7c6c9f05fb45693b557871b3
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Sep 24 00:11:20 2026 +0300

    docs(flow): reconcile the follow-up contract with the panel's second loading site

 rules/flow-manual-review.mdc              |  2 +-
 skills/flow-contracts/SKILL.md            |  2 +-
 skills/flow-contracts/jira-followups.md   | 42 ++++++++++++++++++++-----------
 skills/flow-contracts/jira-integration.md |  5 ++--
 4 files changed, 32 insertions(+), 19 deletions(-)

commit a26df25d641071b229b89fc199eaba8e7c92afff
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Sep 23 23:48:21 2026 +0300

    docs(flow): load the follow-up contract at the deferred-findings site

 skills/flow-contracts/jira-followups.md | 15 ++++++++++++++-
 1 file changed, 14 insertions(+), 1 deletion(-)

commit e7b37897edf5edb648609bf58558f56002e90a0d
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Sep 23 23:46:31 2026 +0300

    docs(flow): file a deferred finding's follow-up at the panel's round close

 skills/flow/review-panel.md | 33 ++++++++++++++++++++++++++++++++-
 1 file changed, 32 insertions(+), 1 deletion(-)

## Session narrative

This `/flow-fast` run resolved KAN-637 (a review panel's deferred findings have no durable
follow-up) to a two-task docs change: `skills/flow/review-panel.md` gained **Deferred findings
file their follow-up at round close** — a round close leaving findings newly `deferred` loads
`skills/flow-contracts/jira-followups.md` as a second loading site and offers the filing once,
with the contract's own join/append machinery on a yes — and the contract file named its second
site and defined the per-site outstanding items. Both tasks landed before the panel; the plan's
decide roll came out `small`/inline/compact, and the docs-only reduction narrowed pass 1 to
`primary` alone. Where it struggled: pass 1 raised two Important findings the plan had missed —
the contract's own body still asserted the repealed integrate-run exclusivity in four passages,
and three other normative files (contracts SKILL.md, `rules/flow-manual-review.mdc`,
`jira-integration.md`) still carried the old loader-exclusivity claim — one of which lives in an
installed always-on rule. Fixing it meant running `setup.sh global`, and the first attempt ran it
from the worktree, repointing the operator's `~/.zcode`/`~/.claude` rule symlinks at a disposable
worktree; that was repaired by re-running the installer from the main checkout as the guard
instructs, and the fixed rule text reaches the installed managed blocks only after a later
`setup.sh global` from main once this branch lands. F2's first reproducer was bounced once for a
`demonstrates:` citation that did not resolve; the raising slot re-authored it and both
reproducers then drove the fix verification (exit 1 pre-fix, exit 0 post-fix). The store-side
open-findings registry the issue marks "longer term" was deliberately left out.
