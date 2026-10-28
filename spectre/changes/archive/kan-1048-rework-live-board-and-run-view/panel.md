# Review panel — kan-1048-rework-live-board-and-run-view

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | call-efficiency | important | stats/web/src/views/StateBoard.tsx:202 | refetches the whole-history all-projects state board every 5 s and does not pause while the tab is hidden |   |
| F2 | 0 | call-efficiency | important | stats/web/src/hooks/useRunDetail.ts:143 | fetches every change's board row in the project to read one; the state-board view ignores the change parameter |   |
| F3 | 0 | call-efficiency | minor | hooks/flow-session-activity.py:77 | spawns a process, PUTs and rewrites the row on every PostToolUse even when state and detail are unchanged |   |
| F4 | 0 | test-sanity | minor | stats/internal/api/activity_test.go:21 | the board-carries-the-stored-value assertion reads rows the test's own fake mirrored; ActivitySince/CurrentStageStartedAt never asserted |   |
| F5 | 0 | test-sanity | minor | stats/web/src/format.test.ts:12 | formatSpan is tested under describe(formatRelative) |   |
| F6 | 0 | primary | minor | stats/web/src/format.ts:73 | formatModel cannot handle date-suffixed model ids: claude-haiku-4-5-20251001 renders haiku-4-5-20251001 |   |
| F7 | 0 | primary | minor | stats/web/src/hooks/useRunDetail.ts:143 | useRunDetail fetches the project's whole state board to read one row | supersedes F2 |
| F8 | 0 | principles | minor | stats/internal/store/activity.go:15 | the activity state list is duplicated across SQL CHECK, Go, Python and TS and the 500-byte cap in 2 |   |
| F9 | 0 | mutation | important | stats/internal/store/activity.go:19 | replacing the updated_at-only-on-state-change CASE with now() survives; ActivitySince untested |   |
| F10 | 0 | mutation | important | stats/internal/store/aggregate.go:115 | flipping the open-run pick started_at DESC to ASC survives; no test has two open runs |   |
| F11 | 0 | mutation | important | stats/internal/store/aggregate.go:127 | c.updated_at < $2 to <= survives; the half-open boundary is untested |   |
| F12 | 0 | mutation | important | hooks/flow-session-activity.py:29 | MAX_DETAIL_BYTES 500 to 501 survives; no test sends a detail over 500 bytes |   |
| F13 | 0 | mutation | minor | hooks/flow-session-activity.py:63 | dropping the URL-quoting of the session id survives |   |
| F14 | 0 | mutation | minor | hooks/flow-session-activity.py:30 | STORE_TIMEOUT_SECONDS 2 to 20 survives |   |
| F15 | 0 | failure-modes | minor | hooks/flow-session-activity.py:30 | a parallel tool's PostToolUse working can overwrite a question while the question is still open |   |
| F16 | 0 | failure-modes | minor | stats/web/src/hooks/useRunDetail.ts:150 | the auxiliary state-board fetch sits inside Promise.all; its failure drops the stage runs and aggregate that loaded |   |
| F17 | 0 | failure-modes | minor | stats/internal/store/activity.go:15 | a detail containing U+0000 makes Postgres reject the insert; activity goes stale |   |
| F18 | 0 | failure-modes | minor | stats/internal/store/aggregate.go:115 | a session that dies without Stop leaves question or permission on the open stage run indefinitely |   |
| F19 | 1 | mutation | minor | stats/web/src/views/StateBoard.tsx:208 | removing document.removeEventListener(visibilitychange, tick) passes all 197 tests; no test unmounts the board and checks the listener is gone |   |
| F20 | 1 | failure-modes | minor | stats/web/src/hooks/useRunDetail.ts:145 | the board-fetch catch swallows the error with no log or indicator, so the header silently lacks state |   |

findings-total: 20
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 withdrawn a hook runs as one process per event by the harness design, and the upsert leaves updated_at unchanged on a repeated state
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 withdrawn each language boundary needs its own copy of the state list, and the SQL CHECK enforces it at the store
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 withdrawn a timeout constant is a timing bound no deterministic test can pin without a sleep-based harness
finding-status: F15 withdrawn the answered question's own PostToolUse is the only signal that clears question, so ignoring working would strand it
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 withdrawn the 6h stale-run sweeper closes the open run, and the board reads activity only through an open run
finding-status: F19 fixed
finding-status: F20 withdrawn the header showing no state is the designed fallback a failing board fetch takes, pinned by its own test

reproducers-total: 20
finding-reproducer: F1 .superpowers/sdd/reproducers/0-call-efficiency-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-call-efficiency-2.sh
finding-reproducer: F3 none — cost depends on tool-call rate
finding-reproducer: F4 none — test-strength finding
finding-reproducer: F5 none — test naming
finding-reproducer: F6 none — cosmetic
finding-reproducer: F7 none — performance note
finding-reproducer: F8 none — cross-language duplication
finding-reproducer: F9 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/0-mutation-4.sh
finding-reproducer: F13 .superpowers/sdd/reproducers/0-mutation-5.sh
finding-reproducer: F14 none — timing bound
finding-reproducer: F15 none — needs a racing harness
finding-reproducer: F16 none — needs a failing endpoint
finding-reproducer: F17 none — hook output invisible
finding-reproducer: F18 none — needs a dead session
finding-reproducer: F19 .superpowers/sdd/reproducers/1-mutation-1.sh
finding-reproducer: F20 none — intended behaviour, covered by a test

## Pass log

### Round 0

- diff size 3807 lines, under cap; roster compact primary+principles (opus/low) + test-sanity+call-efficiency (sonnet/low) + mutation+failure-modes (sonnet/low); docs-only exit 1 (hooks/flow-session-activity.py) — resolved roster unchanged; standards: none resolved
- dispatched: panel-0-primary+principles on opus/low; panel-0-test-sanity+call-efficiency on sonnet/low
- dispatched: panel-0-mutation+failure-modes on sonnet/low
- parent renamed mutation reproducers R-0-mutation-<n>.sh to 0-mutation-<n>.sh (guard shape); content unchanged
- verdict: fix (0 Critical, 5 Important)

### Round 1

- panel-fix-1 inline (opus/low, decision execution inline) at 7603021f: fixed F1 F2 F4 F5 F6 F7 F9 F10 F11 F12 F13 F16 F17 and the full-suite failure (sweeper abandoned UI-test fixture open runs); F11 reproducer re-authored (escaped $2 never matched, so its edit never applied) and proved both ways, sha 72e2929889f5; auto-decided withdrawals F3 F8 F14 F15 F18
- round-boundary base check: MOVED, 14 commits, overlap stats/cmd/flowd/main.go — auto-resolved Continue; integrate sync settles it
- dispatched: panel-1-primary+principles on opus/low; panel-1-test-sanity+call-efficiency on sonnet/low; panel-1-mutation+failure-modes on sonnet/low
- round 1 re-run raised Minors only: F19 fixed inline (unmount listener assertion, break-and-prove held), F20 withdrawn; client_test stub signature updated for LiveStateBoard(change) in the same inline commit (stats-go build failure)
- verdict: clean
fix-mutation: stats/web/src/views/StateBoard.tsx — visibility check document.hidden replaced by false — state board stops refetching while the tab is hidden
fix-mutation: stats/web/src/hooks/useRunDetail.ts — change dropped from the state-board request — run detail header and kpis
fix-mutation: stats/web/src/hooks/useRunDetail.ts — catch on the board fetch removed — run detail still shows its stage runs when the board row fails to load
fix-mutation: stats/web/src/format.ts — date-suffix group removed from formatModel — formatModel shortens a model id to family and version
fix-mutation: stats/internal/store/aggregate.go — change predicate made always-true — TestLiveStateBoardLiveFieldEdges
fix-mutation: stats/internal/store/activity.go — updated_at CASE replaced by now() — TestLiveStateBoardLiveFieldEdges
fix-mutation: stats/internal/store/aggregate.go — open-run pick started_at DESC to ASC — TestLiveStateBoardLiveFieldEdges
fix-mutation: stats/internal/store/aggregate.go — updated_at < to changed to <= — TestLiveStateBoardLiveFieldEdges
fix-mutation: stats/internal/api/stats.go — change query value not passed to the store — TestSessionActivityPut
fix-mutation: stats/cmd/flowd/main.go — sweepEnabled returns true always — TestSweepEnabled
fix-mutation: hooks/flow-session-activity.py — MAX_DETAIL_BYTES 500 to 501 — TestFlowSessionActivityHook
fix-mutation: hooks/flow-session-activity.py — session id URL-quoting removed — TestFlowSessionActivityHook
fix-mutation: hooks/flow-session-activity.py — NUL strip removed — TestFlowSessionActivityHook
fix-mutations-total: 13

### Round 2

- dispatched: panel-2-primary on opus/low over 080b58724efeafedada350bfb72b0a024c61626b; panel-2-test-sanity+call-efficiency on sonnet/low over 080b58724efeafedada350bfb72b0a024c61626b; panel-2-mutation+failure-modes on sonnet/low over 080b58724efeafedada350bfb72b0a024c61626b
- in-run fix 1 re-review (late-fix.diff since 080b5872): primary + extras clean; tooltip flushSync recorded unforceable in one run (React flushes the discrete update before paint), mutation slot concurs
- verdict: clean
