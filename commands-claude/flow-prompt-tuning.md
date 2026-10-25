---
description: Apply the generic implementer/reviewer prompt updates the self-review's prompt-tuning angle saved, after an A/B replay proves each; record each case in docs/prompt-tuning/LEDGER.md and remove its folder
---

Use the **flow-prompt-tuning** skill — installed globally, so let your harness resolve it by name
rather than assuming a project-local path.

Follow that skill exactly, from **Workflow** (`skills/flow-prompt-tuning/SKILL.md`). **Standalone, not a pipeline stage** — it takes no argument, writes no
state file, and marks no `flow stage` call. It walks the pending cases under
`<agents repo>/docs/prompt-tuning/`, groups those proposing the same update, A/B-replays each group on opus,
applies the updates that pass, records every case in the ledger, deletes the case folders and
lands the result by the agents repository's default landing route.

**Input:** none. Any argument is reported rather than ignored.

**When done:** nothing further to run — the updates and the ledger are already landed.
