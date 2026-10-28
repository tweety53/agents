# Live verification — kan-1048-rework-live-board-and-run-view

Per `design.md` `## Live check`. Both figures were measured on scratch copies of the dev store made with `scripts/live-store-copy.sh` (port 4175), never on the dev stack itself, with one request: `GET /api/v1/stats/state-board?from=1970-01-01T00:00:00Z&to=2026-10-28T23:14:52Z`.

## Before

Measured at merge base 6a53dd04f89c7eefe9712dde4987b445a6f44bef. The branch was later rebased onto 7b10e5691e3eb1e419a5b7cb0f3b65fd434bcf65, whose only `stats/` changes from 6a53dd04 are in `stats/internal/guard`, so flowd's state-board path is the same source and the figure stands. The daemon was built from the main checkout at that sha, against copy `flow_k1048b_live`.

- rows: 720
- rows carrying `currentStage`: 0
- rows carrying `doneTasks`: 0
- rows carrying `activity`: 0

## After

Measured at the branch HEAD. The daemon was built from the change worktree, against copy `flow_k1048a_live`.

- rows: 720 (same as before)
- rows carrying `currentStage`: 2 (`kan-1048-rework-live-board-and-run-view` → `flow.verify`, `kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets` → `flow.visual-verify`, the two changes with an open stage run)
- rows carrying `doneTasks`: 369
- rows carrying `activity`: 0. The copy predates any activity hook report.
- Probe: `PUT /api/v1/sessions/3a0e20fd-e67c-43ad-a5d9-91cb39893821/activity` with `{"state":"question","detail":"probe"}` returned 204. Reading back `…&change=kan-1048-rework-live-board-and-run-view` then gave 1 row with `currentStage` `flow.verify`, `activity` `question` and `activityDetail` `probe`.

## Failure look

The failure would be one of two things. Either no row carries `currentStage` while `flow status` shows an open stage, or the row count differs from before. Neither occurred: matched.
