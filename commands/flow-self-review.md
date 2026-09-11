---
name: /flow-self-review
id: flow-self-review
category: flow
description: Run a self-review pass a /flow run deferred, inline on this session's model, from the saved context bundle
---

**Model:** whatever this session runs on — that is the stronger-model choice, so pick it with
`/model` before invoking this command.

Use the **flow-self-review** skill — installed globally, so let your harness resolve it by name
rather than assuming a project-local path.

Follow that skill exactly. **Standalone, not a pipeline stage** — it takes one change name, writes
no per-change state file, and marks no `flow stage` call. It reads the saved context bundle a
`/flow` run's `## self review: defer` left behind, runs the five-angle reasoning pass inline, files
and rates findings, writes the report, deletes the bundle, and lands both on the default branch.

**Input:** one change name, required. Any other argument is reported rather than ignored.

**When done:** nothing further to run — the report and the deleted bundle are already committed
and pushed.
