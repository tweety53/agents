---
model: sonnet
description: Fast — brainstorm, implement, and integrate in one command, pausing only at the human gates
---

Use the **myflow-fast** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Accepts **no state** (creates a change) or **`IN_PROGRESS`**. On a
creating run it runs brainstorming (unchanged, fully interactive) and then, in the same
invocation, implementation and the review panel, ending at `IN_PROGRESS`. Re-invoked with an
argument at `IN_PROGRESS`, the argument is fix instructions. Re-invoked bare at `IN_PROGRESS`, it
asks how to land the branch; merge-and-push continues in the same invocation through archive to
`FINISHED`, while open PR and manual stop and hand off.

Publishes no proposal artifact — the operator is present for the brainstorming dialogue that
produces the design.

Also follow the myflow rule (`myflow-manual-review.mdc`) — installed globally, so let your harness
resolve it rather than assuming a project-local path. It is a stub: **load
`skills/myflow-contracts/pipeline.md` first**, which is canonical for the states, transitions, git
boundaries and the finish contract.

**Input:** the change name or a description/Jira key to seed a new change, from `$ARGUMENTS` or
the conversation — and nothing else. **This command takes no flags.** If omitted at `IN_PROGRESS`,
run `spectre list --json` and use the sole relevant open change, asking which when there are
several. Report any argument that is not a change name, description, or fix instruction rather
than ignoring it.

**When done:** at `IN_PROGRESS` with a fresh staged diff — after a creating run or a fix — review
the staged diff and run the apps, then re-run `/myflow-fast <name>` (or `/myflow-fast <name>
<fix>`) as needed. At `IN_PROGRESS` after choosing open PR or manual, there is nothing new to
review — wait for the merge or your manual steps, then re-run bare to archive. At `FINISHED`,
nothing further.
