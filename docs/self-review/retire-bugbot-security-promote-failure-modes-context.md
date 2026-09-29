# Self-review context bundle for retire-bugbot-security-promote-failure-modes

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/retire-bugbot-security-promote-failure-modes/tasks.md (absent)
skipped: spectre/changes/archive/retire-bugbot-security-promote-failure-modes/design.md (absent)
skipped: spectre/changes/archive/retire-bugbot-security-promote-failure-modes/narrative.md (absent)

## .superpowers/sdd/ledgers/retire-bugbot-security-promote-failure-modes.md

# SDD ledger — retire-bugbot-security-promote-failure-modes

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: opus effort=medium
- Commit: no commit
- Outcome: completed
- Started: 2026-09-29T09:13:36Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: exp-failure-modes
- Key: panel-0-exp-failure-modes
- Model: opus effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-09-29T09:13:36Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: opus effort=low
- Commit: no commit
- Diff base: 097f4da4912968fba190462624f9c2fadd94c5ba
- Outcome: completed
- Started: 2026-09-29T09:23:43Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Slot: exp-failure-modes
- Key: panel-1-exp-failure-modes
- Model: opus effort=low
- Commit: no commit
- Diff base: 097f4da4912968fba190462624f9c2fadd94c5ba
- Outcome: completed
- Started: 2026-09-29T09:23:43Z
- Tokens: not measured

## Dispatch 5 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: opus effort=low
- Commit: no commit
- Diff base: 097f4da4912968fba190462624f9c2fadd94c5ba
- Outcome: completed
- Started: 2026-09-29T09:25:22Z
- Tokens: not measured
## .superpowers/sdd/reviews/retire-bugbot-security-promote-failure-modes-panel.md

# Review panel — retire-bugbot-security-promote-failure-modes

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | .superpowers/sdd/retire-bugbot-security-promote-failure-modes/tasks.md:4 | Task 1 Files: names skills/flow/SKILL-rationale.md, which no commit touches |   |
| F2 | primary | Minor | skills/flow/failure-modes-reviewer-prompt.md:16 | failure-modes and principles prompts disclaim bug hunt/security audit as other slots' angles; no slot owns them now |   |
| F3 | primary | Minor | skills/flow/review-panel.md:332 | default-panel grouping rule yields primary+principles+failure-modes · mutation, planner static row says primary+principles · failure-modes+mutation, yet called the same static logic |   |
| F4 | primary | Minor | skills/flow/SKILL.md:95 | a stored reviewers list carrying a retired id resolves unchanged onto a default panel with no roster row or prompt |   |
| F5 | principles | Minor | stats/cmd/flow/settings.go:117 | -reviewers usage text hard-codes ValidReviewers; settings.go comment hard-codes its size |   |
| F6 | principles | Minor | skills/flow/review-panel.md:332 | grouping contradiction between review-panel default rule and planner static row (raised independently) |   |
| F7 | exp-failure-modes | Important | stats/internal/store/settings.go:37 | a settings row stored before this change carrying bugbot/security has no handling: it resolves to ids with no prompt, and keep-current in /flow-settings is rejected by PutSettings |   |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed

reproducers-total: 7
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 none — two docs contradict; no command runs either rule
finding-reproducer: F4 none — the live row is clean; demonstrating needs a store write
finding-reproducer: F5 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-exp-failure-modes-1.sh

## Pass log

### Round 0

- roster: compact — 34
- diff-size: 596 changed lines, under cap — proceed
- docs-only: exit 1, first non-doc path scripts/check-cleanup-complete.sh — resolved roster runs unchanged
- no addition this round — the resolved list ran alone.
- citation check: scripts/check-references.sh captured to citation-check.md
- context bundle: bundle rebuilt — no cached bundle
## git log --stat

commit 874be09d789c2072223308d4327a6f304f769bc2
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 12:23:06 2026 +0300

    fix(review-panel): close the panel's round-0 findings
    
    Strip retired reviewer ids from a stored flow_settings row by migration,
    derive the CLI's reviewer list from ValidReviewers, keep the default
    panel's floor bundle alone like the planner's static grouping, and stop
    the failure-modes and principles prompts disclaiming angles no slot owns
    any more.

 skills/flow/failure-modes-reviewer-prompt.md       |  4 +--
 skills/flow/principles-reviewer-prompt.md          |  4 +--
 skills/flow/review-panel.md                        |  5 ++-
 stats/cmd/flow/settings.go                         |  3 +-
 .../0033_flow_settings_drop_retired_reviewers.sql  | 15 +++++++++
 stats/internal/store/settings.go                   |  8 +++--
 stats/internal/store/settings_test.go              | 36 ++++++++++++++++++++++
 7 files changed, 64 insertions(+), 11 deletions(-)

commit 097f4da4912968fba190462624f9c2fadd94c5ba
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 12:12:31 2026 +0300

    feat(store): accept failure-modes and reject bugbot and security reviewer ids
    
    The settings-store vocabulary follows the panel roster: bugbot and
    security are retired, failure-modes is the promoted slot. Historical
    rows carrying a retired id still load unchanged.

 stats/cmd/flow/settings.go            |  2 +-
 stats/internal/api/settings_test.go   |  4 ++--
 stats/internal/store/settings.go      | 16 +++++++++-------
 stats/internal/store/settings_test.go | 21 +++++++++++++++++++--
 4 files changed, 31 insertions(+), 12 deletions(-)

commit ebdfdde14c410a015605e93b1b8a7b34cad2145f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 12:11:14 2026 +0300

    feat(review-panel): retire bugbot and security, promote failure-modes to a slot
    
    Security raised no finding in 22 dispatches and bugbot yielded 0.31
    Critical+Important findings per dispatch against 0.49-0.55 for primary,
    principles and mutation. The experimental failure-modes prompt yielded
    1.17 over 12 dispatches, so it joins the regular and big rosters as a
    permanent diff-reading slot.

 README.md                                          |   4 +-
 scripts/check-cleanup-complete.sh                  |   2 +-
 skills/flow-contracts/artifacts-registry.md        |   2 +-
 skills/flow-contracts/model-policy-rationale.md    |   8 +-
 skills/flow-contracts/model-policy.md              |   6 +-
 skills/flow-settings/SKILL.md                      |   2 +-
 skills/flow/SKILL.md                               |   4 +-
 skills/flow/brainstorm-planner.md                  |   8 +-
 skills/flow/bugbot-reviewer-prompt.md              | 103 ---------------------
 ...e-modes.md => failure-modes-reviewer-prompt.md} |  20 ++--
 skills/flow/implement.md                           |   4 -
 skills/flow/primary-reviewer-prompt.md             |   4 +-
 skills/flow/review-panel-optional-slots.md         |  17 ++--
 skills/flow/review-panel.md                        |  47 +++++-----
 skills/flow/security-reviewer-prompt.md            |  94 -------------------
 15 files changed, 59 insertions(+), 266 deletions(-)

## Session narrative

The operator asked which reviewers are most effective and to delete the rest. The session queried the flow store read-only: security had raised no finding in 22 bundled dispatches, and bugbot yielded 0.31 Critical+Important findings per dispatch against 0.49–0.55 for primary, principles and mutation. exp-failure-modes yielded 1.17 over 12 dispatches. The operator chose to keep primary, principles and mutation, promote failure-modes, and land through /flow-fast, unlinked to Jira. The implementation deleted both prompt files and moved the experimental prompt to `skills/flow/failure-modes-reviewer-prompt.md`. It rewired every live reference: the roster table, the planner tree (failure-modes joins regular and big, grouped `failure-modes+mutation`), the default-panel grouping, the throwaway-worktree prose, the contracts, the README and the settings store's `ValidReviewers`. It left archived spectre history and rationale quotes untouched. Where the run struggled: the panel's failure-modes slot caught that a stored settings row still carrying a retired id had no handling, a gap the first pass missed; it was fixed by migration 0033. Two reproducers had declared the very text the fix replaced as their premise, so they failed loudly after the fix and had to be repaired. F1's pre-fix proof leg could not run, because the plan file is untracked. The session also closed the panel stage before check-panel-findings-closed.sh confirmed closure; the guard then required a principles re-run for fixed Minors, which ran after verify had opened, and its result was clean. The zsh shell also broke a word-split variable used to hold a command.
