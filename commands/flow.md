---
name: /flow
id: flow
category: flow
description: Single-command pipeline — brainstorm, implement behind the review panel resolved from the settings store, and integrate, pausing only at the human gates
---

**Model:** planning and implementation both run in subagents whose models are set at dispatch time
(**Model resolution**, `skills/flow/SKILL.md`), so the session model is not load-bearing.

Use the **flow** skill — installed globally, so let your harness resolve it by name rather than
assuming a project-local path.

Follow that skill exactly. Accepts **no state** (creates a change), **`STARTED`** (resumes a
creating run that stopped before implementation), or **`IN_PROGRESS`**. On a creating run it writes
`STARTED` immediately, then runs brainstorming (unchanged, fully interactive, now in a planner
subagent on the configured planning model) and, in the same invocation, implementation behind the
review panel resolved from the settings store, ending at `IN_PROGRESS`. Re-invoked with an argument
at `IN_PROGRESS`, the argument is fix instructions.
Re-invoked bare at `IN_PROGRESS`, it asks how to land the branch; merge-and-push continues in the
same invocation through archive to `FINISHED`, while open PR and manual stop and hand off.

Publishes no proposal artifact — the operator is present for the brainstorming dialogue that
produces the design. Asks no planning-effort, model, or review-panel-roster question on a creating
run — the roster is resolved from the settings store's reviewer list instead
(`skills/flow/review-panel.md` is canonical for it); a slot beyond that list is added only by an
explicit operator instruction, at any point in the run.

Also follow the flow rule (`flow-manual-review.mdc`) — installed globally, so let your harness
resolve it rather than assuming a project-local path. It is a stub: **load
`skills/flow-contracts/pipeline.md` first**, which is canonical for the states, transitions, git
boundaries and the finish contract; `/flow`'s own stage keys are in `skills/flow/SKILL.md`'s own
**Stage keys**, cited rather than repeated here.

**Input:** the change name or a description/Jira key to seed a new change, from `$ARGUMENTS` or the
conversation — and nothing else. **This command takes no flags.** If omitted at `IN_PROGRESS`, run
`spectre list --json` and use the sole relevant open change, asking which when there are several.
Report any argument that is not a change name, description, or fix instruction rather than ignoring
it.

**When done:** at `IN_PROGRESS` with a fresh staged diff — after a creating run or a fix — review
the staged diff and run the apps, then re-run `/flow <name>` (or `/flow <name> <fix>`) as needed. At
`IN_PROGRESS` after choosing open PR or manual, there is nothing new to review — wait for the merge
or your manual steps, then re-run bare to archive. At `FINISHED`, nothing further.
