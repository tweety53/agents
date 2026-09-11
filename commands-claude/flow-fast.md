---
model: sonnet
description: Minimal-ceremony /flow variant — one invocation from Jira key to landed change, a git worktree for isolation only, every flow.* stage marked, Jira transitions kept, nothing else
---

Use the **flow-fast** skill — installed globally, so let your harness resolve it by name rather
than assuming a project-local path.

Follow that skill exactly. One invocation runs from the Jira key to the landed change: resolve the
issue and name per **Transitions** (`skills/flow-contracts/jira-integration.md`), move it to In Progress, create a git worktree on a branch named after the change (git isolation only
— no workspace setup, database or bucket), implement inline in this session, run the project's
`## lint` and the tests the change touches, print the change summary, land by the project's
`## default landing route` (asking only when none is declared), move the issue to In Review, and —
on merge and push — remove the worktree and branch and move the issue to Done. Every `flow.*`
stage `/flow` marks is marked, most as an empty pair, so the stats views see one pipeline.

No spectre artifacts, no state file, no decision record, no subagent, no review panel, no guard
script. Asks no model, planning-effort or review question.

Also follow the flow rule (`flow-manual-review.mdc`) — installed globally, so let your harness
resolve it rather than assuming a project-local path — for the Jira contract it points at; the
pipeline's states do not apply here, since `/flow-fast` writes none.

**Input:** the change description or Jira key, from `$ARGUMENTS` or the conversation — and nothing
else. **This command takes no flags.** Re-invoked with a name whose worktree still exists, the
argument is fix instructions, or — bare — the run resumes at cleanup once an open PR has merged.

**When done:** on merge and push, nothing further — the commit is on the default branch. On open
PR or manual, merge or land it, then re-run `/flow-fast <name>` bare to clean up.
