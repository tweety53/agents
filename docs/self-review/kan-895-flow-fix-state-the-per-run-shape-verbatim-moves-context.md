# Self-review context bundle for kan-895-flow-fix-state-the-per-run-shape-verbatim-moves

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-895-flow-fix-state-the-per-run-shape-verbatim-moves/tasks.md (absent)
skipped: spectre/changes/archive/kan-895-flow-fix-state-the-per-run-shape-verbatim-moves/design.md (absent)
skipped: spectre/changes/archive/kan-895-flow-fix-state-the-per-run-shape-verbatim-moves/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-895-flow-fix-state-the-per-run-shape-verbatim-moves.md

# SDD ledger — kan-895-flow-fix-state-the-per-run-shape-verbatim-moves

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-06T19:35:02Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-895-flow-fix-state-the-per-run-shape-verbatim-moves-panel.md

# Review panel — kan-895-flow-fix-state-the-per-run-shape-verbatim-moves

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | stats/internal/guard/verbatimmoves.go:172 | vmFastChangeRoot silently names the alphabetically-first .superpowers/sdd/*/tasks.md when several exist; its comment asserts the one-root invariant nothing enforces |   |
| F2 | principles | Minor | stats/internal/guard/verbatimmoves.go:156 | the FAIL advice-sentence template is duplicated across the two reply branches, differing only in the path argument |   |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 .superpowers/sdd/reproducers/0-0-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-0-2.sh

## Pass log

### Round 0

- roster: compact — 36
- diff size: 82 lines, under cap — proceeding
- docs-only: exit 1 — first non-documentation path stats/internal/guard/check_verbatim_moves_test.go — resolved roster runs unchanged
- no addition this round — the resolved list ran alone
- dispatch given opus/medium per operator model policy — ledger row carries the zcode harness mapping (glm-5.3-flash/high), the only pair this harness records
## git log --stat

commit 61b4ecc902700be60a825010e5f551fed1179131
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 22:38:42 2026 +0300

    docs(flow): root the flow-fast acknowledgement path citation

 skills/flow-fast/SKILL.md | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)

commit dfd8e32feff8cb14e3c98c4a8e5dcd40579247d7
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 22:35:44 2026 +0300

    fix(guard): review Minors

 stats/internal/guard/verbatimmoves.go | 12 +++++++-----
 1 file changed, 7 insertions(+), 5 deletions(-)

commit d8febc887801e1a9aeeda9de6f8d5d38dfe4f992
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 22:16:06 2026 +0300

    docs(flow): state the per-run-shape verbatim-moves acknowledgement paths
    
    The SKILL's verbatim-moves sentence named only this run's <changeRoot>
    home and gestured at a spectre one it never names, so a run reading the
    FAIL output's copy-this line and this sentence together could not tell
    which of the two paths applies to the run shape it is in. The sentence now
    names both paths with the run shape that owns each — the spectre change
    directory on a /flow run, <changeRoot> on this one, never spectre here.

 skills/flow-fast/SKILL.md | 6 ++++--
 1 file changed, 4 insertions(+), 2 deletions(-)

commit 9076b5e2d6369150771c775c5a1e269921f00fdb
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 22:14:54 2026 +0300

    fix(guard): name the run shape's own verbatim-moves path in the FAIL reply
    
    The FAIL reply listed both acknowledgement homes as an undifferentiated
    either/or, so a flow-fast run reading it could not tell which path applies
    to the run shape it is in — kan-800 wrote the spectre one against this
    run's own never-write-spectre guardrail, and kan-777 before it hand-rolled
    citation-shaped sentences instead. The reply now detects the run shape by
    the plan only a flow-fast run writes there — tasks.md under
    .superpowers/sdd/<change>/ — and names that change's own
    verbatim-moves.txt; with no such root it names the /flow path alone.

 stats/internal/guard/check_verbatim_moves_test.go | 51 +++++++++++++++++++++++
 stats/internal/guard/verbatimmoves.go             | 25 ++++++++++-
 2 files changed, 75 insertions(+), 1 deletion(-)

## Session narrative

This run fixed KAN-895: the verbatim-moves acknowledgement path is now unambiguous at the point of need, in both places a run reads it. The guard (stats/internal/guard/verbatimmoves.go) detects the run shape by the one artifact only a flow-fast run writes there — tasks.md under .superpowers/sdd/<change>/, a /flow run keeping its tasks.md in its spectre change root — and the FAIL reply names the detected shape's own acknowledgement path instead of the old undifferentiated either-or; the flow-fast SKILL's verbatim-moves sentence names both homes with the run shape owning each. The guard change went test-first: the new TestCheckVerbatimMovesFailReplyNamesRunShapePath failed on the old either-or reply in both subtests, then passed. The SKILL edit was deliberately acked through the very mechanism the change documents — its two sentences are listed in this run's own .superpowers/sdd/<change>/verbatim-moves.txt, which the new reply named correctly on its first live FAIL. Two approaches were tried and revised: the SKILL sentence first used <worktree>/ as the flow-fast path root, which check-installed-citations.sh rejected (names no root — the citation grammar accepts <project>/ and <agents repo>/ and the SKILL's own established token is <abs-worktree>), so the sentence was re-rooted on <abs-worktree> and the ack line re-cut from the fresh FAIL output; and the panel round raised two Minors — an over-asserted one-root comment and a duplicated advice template — both fixed inline at round close, collapsing the two Fprintf branches onto one computed path. The review panel (compact, primary+principles) ran on opus per the operator model policy; its ledger row carries the zcode harness mapping (glm-5.3-flash/high), the only pair this harness' store records. SPA tests were not run: the diff names no stats/web file, so nothing there was in scope.
