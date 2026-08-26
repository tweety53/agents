---
name: /flow
id: flow
category: myflow
description: Single-command pipeline — brainstorm, implement behind a fixed review panel, and integrate, pausing only at the human gates
---

**Model:** keep this **session** on Sonnet (or your default). See "Model policy" in
`skills/flow-contracts/model-policy.md` for the per-harness enforcement notes that still apply —
its three-role table does not; see `skills/flow/SKILL.md`'s own **Model resolution**, which is
canonical for `/flow`. Cursor doesn't yet support a per-command model frontmatter field, so the
session setting is a recommendation rather than an enforced switch — but the subagent models are
set at dispatch time and apply in every harness.

Use the **flow** skill — installed globally, so let your harness resolve it by name rather than
assuming a project-local path.

Follow that skill exactly. Accepts **no state** (creates a change), **`STARTED`** (resumes a
creating run that stopped before implementation), or **`IN_PROGRESS`**. On a creating run it writes
`STARTED` immediately, then runs brainstorming (unchanged, fully interactive) and, in the same
invocation, implementation behind a fixed 3-slot review panel, ending at `IN_PROGRESS`. Re-invoked
with an argument at `IN_PROGRESS`, the argument is fix instructions. Re-invoked bare at
`IN_PROGRESS`, it asks how to land the branch; merge-and-push continues in the same invocation
through archive to `FINISHED`, while open PR and manual stop and hand off.

Publishes no proposal artifact — the operator is present for the brainstorming dialogue that
produces the design. Asks no planning-effort, model, or review-panel-roster question on a creating
run — the panel is fixed at 3 required slots (Primary, Principles, Code review (low)) on every run;
Bugbot and Security are dispatched only when you explicitly ask for either, at any point in the run.

Also follow the myflow rule (`myflow-manual-review.mdc`) — installed globally, so let your harness
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
