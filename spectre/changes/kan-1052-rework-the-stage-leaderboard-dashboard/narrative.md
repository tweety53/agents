# kan-1052-rework-the-stage-leaderboard-dashboard — session narrative

## 2026-10-31 — creating run

Implementation resumed at the plan's four tasks. Groups [1,4] and [2] ran as one wave in throwaway worktrees and were cherry-picked in plan order; the first launch of group [1,4] was refused because its dispatch row was keyed `task-1-implementer` instead of the group's `task-1+4-implementer` — the row was closed aborted and re-begun under the group key. Task 3 ran on the canonical worktree against the UI-test stack started once beforehand.

Task 3's implementer reported, and the pass-1 full suite confirmed, five deterministic Experiments capture failures (each page 1px shorter than its baseline). They were not this branch's: bd8a8112, already in the merge base, changed `self-review-fix`'s allowed pairs and dropped its escalation line without re-recording the Experiments PNGs. full-suite-fix-1 recaptured them after root-causing; no tolerance was touched.

Panel round 0 raised 13 findings — 7 Important, all mutation survivors (untested previous-window error, not-recorded/unmeasured states, command-filtered previous total, Period chip preserving other chips, project spend order, swallowed projects error) and 6 Minors, among them call-efficiency's discarded projects aggregate on the previous-window call, fixed by a `projects=1` query parameter the view sends on its current-window call only. panel-fix-1 (two chunks) fixed all 13 on top of the pushed branch. Round 1 re-ran every role on sonnet; the mutation re-run wrote its verdict lines only in its message, so one `-cite` send-back added them to the report. Round 1's three mutation Minors (sub-cent threshold boundaries, `projects=0`) were fixed inline by the parent; primary's Minor on `projects` being accepted on every view was withdrawn as the shared allowlist's existing design (as for `class`).

Two zsh footguns cost a retry each: unquoted `$FR` holding several `-fix-report` flags did not word-split (fixed with an array), and `"$PWD:stats/web"` parsed as the `:s` modifier (fixed with `${PWD}:`). The mutation slot symlinked the canonical `node_modules` into its throwaway copy; the parent removed the link before the copy's removal so the shared directory was not emptied.

Live check against `scripts/live-store-copy.sh`: rows 10 → 54, projects absent → five projects by spend, project filter narrows rows but not projects, Spend tile equals the row sum. Origin/main moved 3 commits during the panel, overlapping styles.css, StateBoard.tsx and views.test.tsx — auto-resolved Continue; integrate's sync settles it.
