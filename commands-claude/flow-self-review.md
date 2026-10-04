---
description: Run a change's self-review pass, inline on this session's model, from the context bundle /flow or /flow-fast saved; fix and land every non-big finding, file the big ones
---

Use the **flow-self-review** skill — installed globally, so let your harness resolve it by name
rather than assuming a project-local path.

Follow that skill exactly. **Standalone, not a pipeline stage** — it takes one change name, writes
no per-change state file, and marks no `flow stage` call. It reads the context bundle a
`/flow` or `/flow-fast` run saved, runs the six-angle reasoning pass inline, fixes and
lands every finding that is not big, files the big ones, rates, records each finding in the flow
store, writes the report, deletes the bundle, and lands both on the default branch.

The pass runs on whatever model this session is already on.

**Input:** one change name, required. Any other argument is reported rather than ignored.

**When done:** nothing further to run — the report and the deleted bundle are already committed
and pushed.
