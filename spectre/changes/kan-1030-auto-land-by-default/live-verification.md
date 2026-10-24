# kan-1030-auto-land-by-default — live verification

Exercise: `flowd` built from the merge base (6c9c7d53) and from this branch, each run on port 4199
against its own scratch database (`flow_live1030b` / `flow_live1030a` in `flow-postgres`), with
`FLOW_STATE_DIR` in the session scratchpad; the dev daemon on 4173 was not touched.

Steps: `flow state set live-1030` at `STARTED` with `"reviewBeforeLanding": true`, `flow state get`;
then `state set` at `IN_PROGRESS` without the field, `flow state get`.

- Before (merge base): the first `state set` printed `⚠ flow: store unreachable — wrote local journal`
  (the unknown field refused) and the first `get` found no record — the failure look.
- After (branch): first get `{"state":"STARTED","reviewBeforeLanding":true}`; second get
  `{"state":"IN_PROGRESS"}` with the field omitted — matched.
