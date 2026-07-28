---
model: sonnet
description: Info — explain the pipeline by reading the installed contract (read-only)
---

Use the **myflow-info** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Accepts **any** state and never blocks. **Read-only** — writes nothing at all.

It reads `skills/myflow-contracts/pipeline.md` at invocation time and answers from it, never from memory, so the explanation always matches the installed contract.

Also follow the myflow rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path. It is a stub: **load `skills/myflow-contracts/pipeline.md` first**, which is canonical for the states, transitions, git boundaries and the finish contract.

**Input:** the change name, from `$ARGUMENTS` or the conversation — and nothing else. **This command takes no flags.** If the name is omitted, run `openspec list --json` and use the sole relevant open change, asking which when there are several. Report any argument that is not a change name rather than ignoring it.

**When done:** nothing — this command only explains.
