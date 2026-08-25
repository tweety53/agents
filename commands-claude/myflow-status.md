---
model: sonnet
description: Status — where every open change actually is (read-only)
---

Use the **myflow-status** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Accepts **any** state and never blocks. **Read-only** — it never commits, never advances a state, never creates a worktree, and never writes the state file.

Reports each open change's state, PR, absolute worktree path, last update, and the next command — including which `/myflow-finish` run comes next, since that depends on whether the branch is merged.

Also follow the myflow rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path. It is a stub: **load `skills/myflow-contracts/pipeline.md` first**, which is canonical for the states, transitions and the finish contract.

**Input:** the change name, from `$ARGUMENTS` or the conversation — and nothing else. **This command takes no flags.** If the name is omitted, run `spectre list --json` and use the sole relevant open change, asking which when there are several. Report any argument that is not a change name rather than ignoring it.

**When done:** run whichever next command it reported.
