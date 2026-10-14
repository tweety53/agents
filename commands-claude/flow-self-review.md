---
description: Run a change's self-review pass, a one-shot opus dispatch, from the context bundle an older /flow or /flow-fast run saved; plan in full, fix and land every non-big finding, file the big ones
---

Use the **flow-self-review** skill — installed globally, so let your harness resolve it by name
rather than assuming a project-local path.

Follow that skill exactly. **Standalone, not a pipeline stage** — it takes one change name, writes
no per-change state file, and marks no `flow stage` call. It reads the context bundle an
older `/flow` or `/flow-fast` run saved — newer runs run the same pass in-run — dispatches the six-angle reasoning pass, plans, fixes and
lands every finding that is not big, files the big ones, rates, records each finding in the flow
store, writes the report, deletes the bundle, and lands both on the default branch.

The pass, the planning, the fixes and the reviews each run as a one-shot dispatch on the pair the
skill names, never on this session's model.

**Input:** one change name, required. Any other argument is reported rather than ignored.

**When done:** nothing further to run — the report and the deleted bundle are already committed
and pushed.
