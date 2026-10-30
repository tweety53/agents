# Live verification — group-allowed-pairs-roles

Against a copy of the dev store (`scripts/live-store-copy.sh up group-allowed-pairs-roles`, flowd from this worktree on 127.0.0.1:4175, database `flow_group_allowed_pairs_roles_live`).

## 1. `GET /api/v1/roles` serves group and does

- Before (dev flowd on 4173, the merge-base build): entries carry no `group` or `does` key (`null` for `main`, `explorer`, `plan-designer`), and `panel-extra` is served at index 10.
- After (4175): groups in first-appearance order `Session, Planning, Implementation, Review, Visual verification, Self-review, Pipeline fix`; 0 entries with an empty group; 0 unmarked entries with an empty `does`; 0 marked entries with a non-empty `does`; `panel-extra` absent; `panel-sanity-efficiency` at index 12, `panel-mutation-failure` at index 13.
- Failure look (missing `group`, empty `group`, `panel-extra` still served): not seen — matched.

## 2. Experiment stats select the two bundle roles

- `flow experiment open -name live-check -cell 'panel-mutation-failure=opus/flow-low' -cell 'panel-sanity-efficiency=opus/flow-low'` → experiment 5.
- Read at once, both windows were null: an experiment opened now has a before window of zero width (it reaches back as far as the after window reaches forward), so the design's expected non-null before mean cannot hold for a freshly opened experiment. On the copy only, its start was moved back 60 days (`update experiments set started_at = now() - interval '60 days' where id = 5`).
- Before figure, window 2026-07-02 .. 2026-08-31: `findingsPerDispatch` before = null — the copy holds no `panel-0-mutation+failure-modes` / `panel-0-test-sanity+call-efficiency` reviewer dispatch in that window (those bundles post-date it).
- After figure, window 2026-08-31 .. now: `findingsPerDispatch` after = 1.3137 over the copy's 25 `panel-0-mutation+failure-modes` and 77 `panel-0-test-sanity+call-efficiency` reviewer dispatches of the last 120 days — the selectors match real dispatch keys.
- Failure look (a null mean on the side that holds those dispatches): not seen — matched.

## 3. Teardown

`scripts/live-store-copy.sh down group-allowed-pairs-roles`.
