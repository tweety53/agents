# Self-review context bundle for kan-808-flow-automation-guard-agents-md-and-claude-md

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-808-flow-automation-guard-agents-md-and-claude-md/tasks.md (absent)
skipped: spectre/changes/archive/kan-808-flow-automation-guard-agents-md-and-claude-md/design.md (absent)
skipped: spectre/changes/archive/kan-808-flow-automation-guard-agents-md-and-claude-md/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-808-flow-automation-guard-agents-md-and-claude-md.md

# SDD ledger — kan-808-flow-automation-guard-agents-md-and-claude-md

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: timed-out
- Started: 2026-09-28T22:03:46Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: exp-failure-modes
- Key: panel-0-exp-failure-modes
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-28T22:03:46Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles-retry
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-28T22:21:52Z
- Tokens: not measured

## Dispatch 4 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 0499954b
- Outcome: completed
- Started: 2026-09-28T22:39:55Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-808-flow-automation-guard-agents-md-and-claude-md-panel.md

# Review panel — kan-808-flow-automation-guard-agents-md-and-claude-md

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | exp-failure-modes | important | stats/internal/guard/handnotes.go:77 | any read error on a harness file is silently read as absence — an existing-but-unreadable file yields a green HAND-NOTES-SINGLE exit 0 while real drift goes unreported |   |
| F2 | exp-failure-modes | minor | scripts/lib/flow-guard.sh:0 | a guard run inside setup.sh rewrite window can read a torn file and report a false DRIFT — loud, recoverable, never a wrong green; fix is installer-side |   |
| F3 | primary | minor | stats/internal/guard/handnotes.go:110 | base hand section is re-split per compared file and the line loop runs one dead index past both ends |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-exp-failure-modes-1.sh
finding-reproducer: F2 none — timing-dependent race; the non-atomic write is setup.sh's, untouched by this diff
finding-reproducer: F3 none — no behavior difference

## Pass log

### Round 0

- wall-clock breach: primary+principles elapsed 950s against the 15-minute ceiling — attempt closed timed-out, bundle re-dispatched once (panel-0-primary+principles-retry)
- roster: compact — rolled 76
- diff size 414, cap not exceeded — proceed; docs-only exit 1 (scripts/check-hand-notes-in-step.sh) — resolved roster ran
- no addition this round — the resolved list ran alone
- F1 defect identity dedupe: the same handnotes.go read-error-as-absence Important was raised by exp-failure-modes (dispatch 2) and, on the ceiling re-dispatch, by primary and principles (dispatch 3) — recorded once under the first raiser; join recorded here

### Round 1

- F1 reproducer re-run flipped demonstrated → not demonstrated (run-reproducer exit 1, sha 351c2225…, pre-fix pin honored) and the fix diff touches handnotes.go — the first re-run refused ambiguous (premise pinned line 77, moved by the fix), the reproducer was re-authored onto tree-stable premises and proved both legs by prove-reproducer.sh at ef53a8f0
- F2 closed on the path condition: exemption-form Minor, fix diff touches setup.sh install_managed_block which the finding names; its -location scripts/lib/flow-guard.sh:0 was recorded imprecisely — the note is canonical and names setup.sh
- F3 closed on the path condition: exemption-form Minor, fix diff touches handnotes.go compare loop
- auto-resolved: the base branch has moved and touches setup.sh, a path this change also touched (overlap — rebase not confirmed conflict-free) → Stop — I'll rebase or reorder first
## git log --stat

commit 845f2718a2d5d8c0052a03b88080c8f871f9a03f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 01:44:58 2026 +0300

    docs(project): drop the one-guard-reads-outside claim the new guard falsifies

 .flow/project.md | 16 ++++++++++------
 1 file changed, 10 insertions(+), 6 deletions(-)

commit 3296cf6fb548b77781893a6b49db60ac6b7f8ec4
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 01:44:58 2026 +0300

    fix(setup): make the managed-block rewrite atomic

 setup.sh | 18 +++++++++++++-----
 1 file changed, 13 insertions(+), 5 deletions(-)

commit 0499954b64688ba60baa55b14f1fbb8baa7b1ed6
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 01:44:58 2026 +0300

    fix(guard): refuse on an unreadable harness file and tighten the compare loop

 stats/internal/guard/check_hand_notes_test.go | 53 +++++++++++++++++++++++++++
 stats/internal/guard/handnotes.go             | 23 ++++++++++--
 2 files changed, 72 insertions(+), 4 deletions(-)

commit ef53a8f019692f99f781adccaaf9ad1eba183cef
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 00:59:14 2026 +0300

    docs(project): run check-hand-notes-in-step in the lint list

 .flow/project.md | 1 +
 1 file changed, 1 insertion(+)

commit 986c6df91a44f1b0015892433826ccb8d17f0216
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 00:57:03 2026 +0300

    feat(scripts): shim check-hand-notes-in-step onto flow-guard

 scripts/check-hand-notes-in-step.sh | 61 +++++++++++++++++++++++++++++++++++++
 1 file changed, 61 insertions(+)

commit 85ce49384f7fa13591adac5f7dc7fe3c2626ec16
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 00:56:25 2026 +0300

    feat(guard): fail lint when the global instruction files' hand notes drift

 stats/internal/guard/check_hand_notes_test.go | 202 ++++++++++++++++++++++++++
 stats/internal/guard/handnotes.go             | 150 +++++++++++++++++++
 2 files changed, 352 insertions(+)

## Session narrative

This invocation resumed a run interrupted mid-verification: the branch already carried the
implementation (Go guard + table tests, the flow_guard_exec shim, the lint-list wiring) and the
closed panel round with its three fix commits, but the self-review bundle and the landing had not
happened. This session re-ran the worktree's full 28-command lint list and the targeted
`internal/guard` race tests. Three hits needed work, and one was the change proving itself: the
new `check-hand-notes-in-step.sh` caught a real one-sided hand append on this machine —
`~/.claude/CLAUDE.md`'s Models note carried a "Code review — runs on opus" bullet that
`~/.zcode/AGENTS.md`'s copy lacked — fixed by mirroring the bullet byte-for-byte into AGENTS.md
exactly as task 3 prescribed. The other two were fresh-worktree artifacts (missing `node_modules`
and unbuilt SPA `dist`) fixed by `npm ci` and `npm run build`. The struggle worth recording: the
guard's first verdict was against the operator's live home files, not fixtures — the fix edits
files outside the repository, which is exactly the blast radius the plan's task 3 accepted.
