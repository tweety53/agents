# Self-review context bundle for kan-864-agents-flow-base-branch-override

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-864-agents-flow-base-branch-override/tasks.md (absent)
skipped: spectre/changes/archive/kan-864-agents-flow-base-branch-override/design.md (absent)
skipped: spectre/changes/archive/kan-864-agents-flow-base-branch-override/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-864-agents-flow-base-branch-override.md

# SDD ledger — kan-864-agents-flow-base-branch-override

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: opus effort=medium
- Commit: no commit
- Outcome: completed
- Started: 2026-10-01T11:25:35Z
- Tokens: input 76, output 451, cache read 2448944, cache creation 101668

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: opus effort=low
- Commit: no commit
- Diff base: d5494921
- Outcome: completed
- Started: 2026-10-01T11:31:45Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: opus effort=low
- Commit: no commit
- Diff base: d5494921
- Outcome: completed
- Started: 2026-10-01T11:31:45Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-864-agents-flow-base-branch-override-panel.md

# Review panel — kan-864-agents-flow-base-branch-override

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/guard/verbatimmoves.go:156 | Three base readers (verbatimmoves vmDefaultBase, selfreview changeBranchLog, check-task-records.py default) still resolve origin/HEAD and ignore branch.<cur>.flowBase |   |
| F2 | primary | Important | skills/flow-contracts/pipeline.md:56 | Canonical Command surface (pipeline.md) and commands-claude/flow.md, flow-fast.md still say no command accepts a flag; --base exception lives only in two SKILL.md files |   |
| F3 | primary | Minor | stats/internal/guard/kickoffworktree.go:167 | flowBase config is written before flow state add-worktree, breaking persisted-before-anything-else ordering |   |
| F4 | primary | Minor | skills/flow-fast/scripts/resolve-base-branch.sh:1 | symlink and verbatim-moves.txt not in Task 3 Files |   |
| F5 | principles | Important | stats/internal/guard/verbatimmoves.go:156 | Single Source of Truth: change base defined in two disagreeing places (same three readers) |   |
| F6 | principles | Important | skills/flow-contracts/pipeline.md:56 | Single Source of Truth: accepted-arguments contract contradicts itself (pipeline.md vs SKILLs) |   |
| F7 | principles | Minor | skills/flow-fast/SKILL.md:115 | flow-fast redefines <default-branch> to mean the --base value; rename to <base> |   |
| F8 | principles | Minor | stats/internal/guard/kickoffworktree.go:65 | empty-base check after the byte loop; fold into the usage guard |   |
| F9 | principles | Minor | stats/internal/guard/verbatimmoves.go:163 | R1 fixed by copying the recorded-flowBase rule; verbatimmoves.go shares the guard package with resolvebasebranch.go and could share one helper |   |
| F10 | primary | Minor | stats/internal/guard/verbatimmoves.go:162 | none of the three new recorded-base read sites carries a test |   |

findings-total: 10
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 withdrawn the symlink and verbatim-moves.txt are what check-guard-symlinks and check-verbatim-moves require of the Files Task 3 names
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed

reproducers-total: 10
finding-reproducer: F1 .superpowers/sdd/reproducers/0-P1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-P1-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-P1-3.sh
finding-reproducer: F4 none — plan-record deviation only
finding-reproducer: F5 .superpowers/sdd/reproducers/0-R1-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-R2-1.sh
finding-reproducer: F7 none — naming
finding-reproducer: F8 none — nit
finding-reproducer: F9 none — duplication only; nothing behaves differently
finding-reproducer: F10 none — missing test coverage

## Pass log

### Round 0

- roster: compact — 29
- diff size: 170 lines, under cap; docs-only: exit 1 (scripts/kickoff-worktree.sh) — resolved roster primary+principles dispatched; no operator-named addition this round — the resolved list ran alone
## git log --stat

commit 2eab868ea088aae69767106fbfc0347d3460c321
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Oct 1 14:34:40 2026 +0300

    fix(guard): review Minors

 scripts/test-check-task-records.sh               | 12 ++++++++++
 stats/internal/guard/resolve_base_branch_test.go | 20 +++++++++++++++++
 stats/internal/guard/resolvebasebranch.go        | 12 ++++++++--
 stats/internal/guard/verbatimmoves.go            |  4 ++--
 stats/internal/selfreview/bundle_test.go         | 28 ++++++++++++++++++++++++
 5 files changed, 72 insertions(+), 4 deletions(-)

commit f2c69a2633679d01220643c91c097f165c3423ee
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Oct 1 14:31:18 2026 +0300

    fix(flow): every base reader honours the recorded base; the flag contract names --base

 commands-claude/flow-fast.md                       |  4 +--
 commands-claude/flow.md                            |  2 +-
 scripts/check-task-records.py                      | 22 +++++++++++++--
 skills/flow-contracts/pipeline.md                  |  4 ++-
 skills/flow-fast/SKILL.md                          | 32 +++++++++++-----------
 .../verbatim-moves.txt                             | 31 +++++++++++++++++++++
 stats/internal/guard/kickoffworktree.go            | 12 +++-----
 stats/internal/guard/verbatimmoves.go              | 15 +++++++---
 stats/internal/selfreview/git.go                   |  9 ++++--
 9 files changed, 95 insertions(+), 36 deletions(-)

commit d549492162a9d41cc6d98ffb7d67b658eea57ccf
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Oct 1 14:24:44 2026 +0300

    feat(flow): a creating run takes --base <branch>

 skills/flow-fast/SKILL.md                                      | 10 ++++++++--
 skills/flow-fast/scripts/resolve-base-branch.sh                |  1 +
 skills/flow-plan/SKILL.md                                      |  2 +-
 skills/flow/SKILL.md                                           |  5 +++++
 skills/flow/brainstorm.md                                      |  6 +++++-
 skills/flow/cross-repo-worktrees.md                            |  6 +++++-
 .../verbatim-moves.txt                                         | 10 ++++++++++
 7 files changed, 35 insertions(+), 5 deletions(-)

commit cc4aabb1eb70b1b370afd409ae0d19b11082b148
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Oct 1 14:21:25 2026 +0300

    feat(guard): kickoff-worktree creates from and records a named base

 scripts/kickoff-worktree.sh                   | 12 +++++--
 stats/internal/guard/kickoff_worktree_test.go | 34 +++++++++++++++++-
 stats/internal/guard/kickoffworktree.go       | 52 ++++++++++++++++++++++-----
 3 files changed, 86 insertions(+), 12 deletions(-)

commit f11cbfd068b435ad486d78671a080b945a9d3945
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Oct 1 14:20:32 2026 +0300

    feat(guard): resolve-base-branch honours a recorded per-branch base

 scripts/resolve-base-branch.sh                   |  5 +++++
 stats/internal/guard/resolve_base_branch_test.go | 19 +++++++++++++++++++
 stats/internal/guard/resolvebasebranch.go        |  8 ++++++++
 3 files changed, 32 insertions(+)

## Session narrative

This /flow-fast run lets a flow change name its base branch (`--base <branch>`). It came out of moving gymie's offline-mode work onto `feature/offline-mode`. Kickoff records the base on the change's branch as `branch.<branch>.flowBase`, and `resolve-base-branch.sh` reads it before `origin/HEAD`. Three tasks did that: the resolver, kickoff, and the skill prose. Where it struggled:
- **The plan overclaimed.** It said every downstream consumer would follow without its own edit. The panel found three readers that bypass the resolver: the verbatim-moves lint base, the self-review branch log, and the task-records default. It also found the canonical "no flags" contract in `pipeline.md` and the command stubs. The fix round covered all of these, and the round-1 Minors added shared-helper reuse and tests proven to fail without the fix.
- **Two contracts disagree.** `check-verbatim-moves` (a lint step) requires `spectre/changes/<change>/verbatim-moves.txt` for any reworded or new run-loaded prose, and /flow-fast says it never writes `spectre/`. The repository's precedent (`rule-prod-read-only-gets`, `handoff-markdown`) is a bare `verbatim-moves.txt`, so the run followed it.
- **Left as recommendations:**
  - Recovering when the recorded base is deleted on origin (`git config --unset branch.<b>.flowBase`) is undocumented.
  - `/flow-self-review` checks for the default branch, so a `--base` change's bundle landed on the feature branch needs a checkout of that base.
  - run 2's main-checkout refresh will report REFRESH-REFUSED on such changes.
