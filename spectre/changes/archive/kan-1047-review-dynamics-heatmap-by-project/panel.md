# Review panel — kan-1047-review-dynamics-heatmap-by-project

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | test-sanity | Important | stats/internal/store/reviewdynamics_test.go:334 | TestReviewDynamicsClassAndProjects claims the Projects list ignores the filters, but every call passes the fixture's only project, so the check still passes if the Projects query applies the project filter. |   |
| F2 | 0 | mutation | Important | stats/internal/store/reviewdynamics.go:150 | Projects query drops d.role = 'reviewer' and no test fails. |   |
| F3 | 0 | mutation | Important | stats/internal/store/reviewdynamics.go:150 | Projects query d.started_at < $2 -> <= survives. |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-test-sanity-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-mutation-2.sh

## Pass log

### Round 0

- diff size: 422 lines under cap; docs-only: exit 1 (stats/internal/api/promptchanges.go) — roster primary+principles unchanged; roster: full; standards: CLAUDE.md, AGENTS.md; no operator slot addition this round — the resolved list ran alone
- mid-run operator request added task 2 (two-column grid, prompt-change marker filter); re-decided as decision 362, class small unchanged
- operator mid-run request added task 3 (drop the since-last-prompt-change preset); re-decided, class small unchanged
- dispatched: panel-0-primary+principles on opus/low; panel-0-test-sanity+call-efficiency on opus/low
- dispatched: panel-0-mutation+failure-modes on sonnet/low
- reproducers 0-mutation-1/2 repaired by the parent: their EXIT trap restored a relative path after cd stats, leaving the mutant in the tree; go -C stats keeps the cwd
- parent repaired F1 reproducer (declared mutation-reproducer but exited 1 on defect; flipped to the mutation convention) and F2 reproducer (single quotes inside a single-quoted sed, so the mutation never landed)
- verdict: fix (0 Critical, 3 Important)

### Round 1

- rerun: test-sanity (F1) and mutation (F2, F3) on fix-round-1.diff, rerun pair sonnet/low; primary, principles, call-efficiency, failure-modes raised nothing and keep their results
- dispatched: panel-1-test-sanity+mutation on sonnet/low over ba3fa2bf1c7a4cb791d6d86eee8d4a80fedca4f0
- fix-mutation rows for panel-fix-1 recorded late, after the round-1 re-review began (the store refuses round 1 then), so they sit under round 2; the round-1 re-review independently re-ran all three mutants and saw each killed
- late mutation proof for stats/internal/store/reviewdynamics.go: parent recorded the rows after launching the round-1 re-review, which independently re-killed all three mutants
- verdict: clean
fix-mutation: stats/internal/store/reviewdynamics.go — Projects query filtered by project — TestReviewDynamicsClassAndProjects
fix-mutation: stats/internal/store/reviewdynamics.go — Projects query role filter dropped — TestReviewDynamicsClassAndProjects
fix-mutation: stats/internal/store/reviewdynamics.go — Projects query upper bound < to <= — TestReviewDynamicsClassAndProjects
fix-mutations-total: 3

### Round 2

- verdict: clean
fix-mutation: stats/internal/store/reviewdynamics.go — Projects query filtered by project — TestReviewDynamicsClassAndProjects
fix-mutation: stats/internal/store/reviewdynamics.go — Projects query role filter dropped — TestReviewDynamicsClassAndProjects
fix-mutation: stats/internal/store/reviewdynamics.go — Projects query upper bound < to <= — TestReviewDynamicsClassAndProjects
fix-mutations-total: 3

### Round 3

- in-run fix 1 (flow.visual-verify) re-review: primary + always-on extras over late-fix.diff since 93a7daa74578f6f065944820d12003d08eac04a4
- dispatched: panel-3-primary on opus/low over 93a7daa74578f6f065944820d12003d08eac04a4; panel-3-test-sanity+call-efficiency on opus/low over 93a7daa74578f6f065944820d12003d08eac04a4
- base moved 3 commits (in-run pipeline fix landed on main, no overlap); rebased before launch; late-fix range from 63c409c6a73340c23e192232c95efb7d14f448c6; rows begun before the rebase carry the pre-rebase sha 93a7daa7 as diff-base
- dispatched: panel-3-mutation+failure-modes on sonnet/low over 63c409c6a73340c23e192232c95efb7d14f448c6
- visual-verify after in-run fix 1: no `specs` command is declared, so a scoped re-run would run the whole suite; dispatched the closing unscoped -fix-1-full run directly
- verdict: clean

### Round 4

- in-run fix 2 (flow.visual-verify) re-review: primary + always-on extras over late-fix.diff since 94b34a37b6b7ffe7ef7a5f00f06ca4a4cfda85b3
- dispatched: panel-4-primary on opus/low over 94b34a37b6b7ffe7ef7a5f00f06ca4a4cfda85b3; panel-4-test-sanity+call-efficiency on opus/low over 94b34a37b6b7ffe7ef7a5f00f06ca4a4cfda85b3; panel-4-mutation+failure-modes on sonnet/low over 94b34a37b6b7ffe7ef7a5f00f06ca4a4cfda85b3
- visual-verify after in-run fix 2: no specs command declared, so the closing unscoped -fix-2-full run is dispatched directly
- verdict: clean

### Round 5

- in-run fix 3 (flow.visual-verify) re-review: primary + always-on extras over late-fix.diff since db38114067363291982c8c216da6ca08b4a4d04b
- dispatched: panel-5-primary on opus/low over db38114067363291982c8c216da6ca08b4a4d04b; panel-5-test-sanity+call-efficiency on opus/low over db38114067363291982c8c216da6ca08b4a4d04b; panel-5-mutation+failure-modes on sonnet/low over db38114067363291982c8c216da6ca08b4a4d04b
- verdict: clean
