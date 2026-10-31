# Review panel — kan-1052-rework-the-stage-leaderboard-dashboard

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | primary | Minor | stats/web/src/views/StageLeaderboard.tsx:115 | the built view and `StageLeaderboardProjects` keep $0.00 costed rows, so a zero-cost stage shows as a ranked row, counts in "<N> stages ran" and "Other N stages", and a zero-spend project gets a chip. |   |
| F2 | 0 | test-sanity | Minor | stats/web/src/views/views.test.tsx:733 | `expect(table).toBeTruthy()` asserts a `within(...)` wrapper captured at the top of the test, which is always truthy; it guards nothing. |   |
| F3 | 0 | call-efficiency | Minor | stats/internal/api/stats.go:512 | every load and every Period/Project chip click therefore runs the projects aggregate twice (once wasted), and a Project chip click re-runs it for an unchanged window although it ignores `project`. |   |
| F4 | 0 | mutation | Important | stats/web/src/views/StageLeaderboard.tsx:242 | A failed previous-window request is never shown: mutating the state to always `cur` (blank table instead of the error) survives; no test fails the previous-window fetch. |   |
| F5 | 0 | mutation | Important | stats/web/src/views/StageLeaderboard.tsx:114 | Dropping `!cur.data.unmeasured` survives: no test answers the leaderboard unmeasured, so tiles/table over unmeasured data go uncaught. |   |
| F6 | 0 | mutation | Important | stats/web/src/views/StageLeaderboard.tsx:114 | Dropping `cur.data.recorded` survives: no not-recorded leaderboard case. |   |
| F7 | 0 | mutation | Important | stats/web/src/views/StageLeaderboard.tsx:126 | Spend tile's previous total ignoring the Command chip survives: the command-filter test never reads the Spend delta, so a wrong "vs the previous" figure under a command filter passes. |   |
| F8 | 0 | mutation | Important | stats/web/src/views/StageLeaderboard.tsx:155 | A Period chip resetting Project/Command (`...defaults()` instead of `...f`) survives: the test picks the period before any other chip. |   |
| F9 | 0 | mutation | Important | stats/internal/store/aggregate.go:617 | Spend-descending order of projects (`ORDER BY 2 DESC, 1` → `ORDER BY 1, 2`) survives: the test's fixture has A (spend 4) sorting before B (spend 2) alphabetically too. |   |
| F10 | 0 | mutation | Important | stats/internal/api/stats.go:515 | Swallowing the `StageLeaderboardProjects` error (200 with empty projects) survives: `statsFake` returns `aggErr` from both methods, so no test fails the projects call alone. |   |
| F11 | 0 | mutation | Minor | stats/web/src/views/StageLeaderboard.tsx:66 | `prev < 0.005` → `prev <= 0` survives (sub-cent previous spend reads as growth, not "new"). |   |
| F12 | 0 | mutation | Minor | stats/web/src/views/StageLeaderboard.tsx:132 | Command note shown even with a Command chip on survives (tiles not checked under a command filter). |   |
| F13 | 0 | mutation | Minor | stats/web/src/views/StageLeaderboard.tsx:70 | The `k` percent format (≥1000%) is never exercised. |   |
| F14 | 1 | mutation | Minor | stats/web/src/views/StageLeaderboard.tsx:117 | Threshold 0.005 → 0.5 survives: the only sub-cent row tested is 0.001, and no fixture has a value between 0.005 and 0.5. |   |
| F15 | 1 | mutation | Minor | stats/web/src/views/StageLeaderboard.tsx:117 | `>=` → `>` at exactly 0.005 also survives. |   |
| F16 | 1 | mutation | Minor | stats/internal/api/stats.go:515 | `== "1"` → `!= ""` survives, so `projects=0` would run the aggregate. |   |
| F17 | 1 | primary | Minor | stats/internal/api/stats.go:165 | the `projects` param is now accepted and ignored on every view, same as `class`. |   |

findings-total: 17
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 withdrawn the shared query allowlist accepts every view parameter on every view by design, as it already does for class, and an ignored parameter changes no response

reproducers-total: 17
finding-reproducer: F1 none — the UI-test fixture has no zero-cost costed run
finding-reproducer: F2 none — the assertion is a no-op; deleting it is the fix
finding-reproducer: F3 .superpowers/sdd/reproducers/0-call-efficiency-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-mutation-4.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/0-mutation-5.sh
finding-reproducer: F9 .superpowers/sdd/reproducers/0-mutation-6.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/0-mutation-7.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/0-mutation-8.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/0-mutation-9.sh
finding-reproducer: F13 .superpowers/sdd/reproducers/0-mutation-10.sh
finding-reproducer: F14 .superpowers/sdd/reproducers/1-mutation-1.sh
finding-reproducer: F15 none — same site as F14; a row at exactly 0.005 covers it
finding-reproducer: F16 none — the contract names only projects=1
finding-reproducer: F17 none — accepted-parameter behaviour, no runtime defect

## Pass log

### Round 0

- diff size: 1699 lines, under cap — proceed
- docs-only: exit 1 (stats/internal/api/stats_test.go) — resolved roster primary+principles unchanged; roster: full; no addition this round — the resolved list ran alone; standards: CLAUDE.md
- dispatched: panel-0-primary+principles on opus/low; panel-0-test-sanity+call-efficiency on opus/low
- dispatched: panel-0-mutation+failure-modes on opus/low
- verdict: fix (0 Critical, 7 Important)

### Round 1

- panel-fix-1: 13 open findings from round 0 (7 Important mutation survivors, 6 Minor) in two chunks; reads final-review.diff and the slot reports
- round 1 re-run: every pass-1 role re-runs (full-suite-fix-1 commit ac75face adds a delta; mutation, primary, test-sanity, call-efficiency raised findings) on panel.rerun_dispatch sonnet/low, reading fix-round-1.diff
- dispatched: panel-1-primary+principles on sonnet/low over e2e9bbf3e4d48a254c24b1ba80efd0d4330cff14; panel-1-test-sanity+call-efficiency on sonnet/low over e2e9bbf3e4d48a254c24b1ba80efd0d4330cff14; panel-1-mutation+failure-modes on sonnet/low
- base sync: origin/main moved 3 commits, overlapping stats/web/src/styles.css, StateBoard.tsx, views.test.tsx — auto-resolved Continue (review as is); integrate sync settles it
- verdict: clean
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — `prev.status !== "ready"` → `false` (0-mutation-1.sh) — stage leaderboard shows a failed previous-window request
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — `!cur.data.unmeasured` → `true` (0-mutation-2.sh) — stage leaderboard reads not-recorded and unmeasured as Panel does
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — `cur.data.recorded` → `true` (0-mutation-3.sh) — stage leaderboard reads not-recorded and unmeasured as Panel does
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — prevRows filter(byCommand) → filter(() => true) (0-mutation-4.sh) — stage leaderboard Spend tile compares the chosen command's previous spend
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — Period chip `...f` → `...defaults()` (0-mutation-5.sh) — stage leaderboard Period chip keeps the Project and Command chips
fix-mutation: stats/internal/store/aggregate.go — `ORDER BY 2 DESC, 1` → `ORDER BY 1, 2` (0-mutation-6.sh) — TestStageLeaderboardProjects
fix-mutation: stats/internal/api/stats.go — projects error branch removed (0-mutation-7.sh) — TestStatsStageLeaderboardProjectsError
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — rows `.filter(costed)` removed — stage leaderboard drops stages and projects under half a cent
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — projects `.filter(costed)` removed — stage leaderboard drops stages and projects under half a cent
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — `projects: true` removed from the current-window call — stage leaderboard asks for projects on its current-window call only
fix-mutation: stats/internal/api/stats.go — `&& firstValue(values, "projects") == "1"` removed — TestStatsStageLeaderboardProjects
fix-mutation: stats/web/src/api.ts — `qs.set("projects", "1")` → `"0"` — buildStatsViewQuery sends projects=1 only when projects is asked for
fix-mutation: stats/web/src/views/views.test.tsx — none — F2 deletes a no-op assertion; no behaviour
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — `prev < 0.005` → `prev <= 0` — stage leaderboard reads a sub-cent previous spend as new (views.test.tsx)
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — cmdNote condition always true — stage leaderboard reads a sub-cent previous spend as new (views.test.tsx)
fix-mutation: stats/web/src/views/StageLeaderboard.tsx — `p >= 1000` → `p >= 1e9` — stage leaderboard reads a sub-cent previous spend as new (views.test.tsx)
fix-mutations-total: 16
