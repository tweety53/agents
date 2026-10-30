---
model: opus
description: Conversational /flow session — each request the operator types is implemented inline, verified, landed by merge and push and, for UI, checked on the live page, until an explicit stop self-reviews, cleans up and finishes
---

Use the **flow-interactive** skill — installed globally, so let your harness resolve it by name
rather than assuming a project-local path.

Follow that skill exactly. One invocation is one session: one change, one Jira issue and one
worktree, and a stream of requests this session works inline, each recorded in flowd, landed on
the base branch by merge and push and reinstalled before the next. An explicit stop runs one
self-review pass, removes the worktree and branch, moves the issue to Done per **Transitions**
(`skills/flow-contracts/jira-integration.md`) and marks the change `FINISHED`. No spectre artifacts and no state file.

**Input:** empty, a first request, a Jira key, or the name of an open interactive session, from
`$ARGUMENTS` or the conversation — and nothing else. **This command takes no flags.** A key or name
of an open session resumes it.

**When done:** type the next request, or tell the session to stop. Every landed request is already
on the base branch; nothing further to merge.
