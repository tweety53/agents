# Live verification — kan-1052-rework-the-stage-leaderboard-dashboard

Window: from=2026-10-01T14:32:11Z to=2026-10-31T14:32:11Z (30 days ending at the run instant).

1. **Before** — protected daemon on 127.0.0.1:4173 (merge-base build), public read-only GET
   `/api/v1/stats/stage-leaderboard?from=…&to=…`: `rows` = 10, `projects` = 0 (absent).
2. **After** — `scripts/live-store-copy.sh up kan-1052` (flowd built from this worktree on 4175, copy of the dev store),
   same request plus `&projects=1` (the view's current-window call): `rows` = 54 (exceeds ten), sum of
   `totalCostUsd` = 6074.13; `projects` = gymie-7c1f238a 4390.06, agents-a740d89c 1462.36,
   flow-eval-07042433 209.05, spectre-e2f26422 9.11, gymie-frontend-3b0a684c 3.54 — ordered by spend,
   none zero-spend.
3. With `&project=gymie-7c1f238a`: `rows` = 38 (narrowed); `projects` identical to step 2's list and totals.
4. `http://127.0.0.1:4175/#/stage-leaderboard`: Spend tile `$6,074` (`−$1,622 (−21%) vs the previous 30d`) —
   equals step 2's row sum 6074.13; Top 3: Implementing tasks (flow.sdd-tdd) $2,302 38%, Code review
   (flow.review-panel) $968 16%, Checking the UI (flow.visual-verify) $929 15% — 69% combined; Biggest
   riser: Self-review (flow.self-review) +$864, $48.95 → $913 (+1.8k%).

Result: matched — none of the "not working" looks (rows still ten, projects empty, step 3 narrowed, Spend = top-ten sum).
Copy torn down with `scripts/live-store-copy.sh down kan-1052`.
