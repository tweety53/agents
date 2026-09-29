---
model: opus
description: Single-command pipeline — brainstorm, implement behind the review panel resolved from the settings store, and integrate, pausing only at the human gates
---

Use the **flow** skill — installed globally, so let your harness resolve it by name rather than
assuming a project-local path.

Follow that skill exactly. Accepts **no state** (creates a change), **`STARTED`** (resumes a
creating run that stopped before implementation), or **`IN_PROGRESS`**. On a creating run it writes
`STARTED` immediately, then runs brainstorming (fully interactive, in this session), ending at
`STARTED` at the plan gate with a `/clear` handoff. Resumed at `STARTED` once planned, it runs
implementation, review and verification orchestrated by this session — implementers and panel
slots dispatched on the resolved default model — behind the review panel the recorded decision
names, ending at `IN_PROGRESS`. Re-invoked with an argument at `IN_PROGRESS`, the
argument is fix instructions. A plain problem report typed with no /flow at all, in the session
that ran the last /flow <name>, is the same fix run (**A plain message at IN_PROGRESS**,
skills/flow/SKILL.md).
Re-invoked bare at `IN_PROGRESS`, it lands the branch by the project's `## default landing route`,
asking how only when none is declared; merge-and-push continues in the
same invocation through archive to `FINISHED`, while open PR and manual stop and hand off.

Publishes no proposal artifact — the operator is present for the brainstorming dialogue that
produces the design. Asks no planning-effort, model, or review-panel-roster question on a creating
run — the roster is the recorded decision's, the settings store's reviewer list only on a
`micro` decision (`skills/flow/review-panel.md` is canonical for it); a slot beyond that list is added only by an
explicit operator instruction, at any point in the run.

Also follow the flow rule (`flow-manual-review.mdc`) — rendered into the managed block of
`~/.claude/CLAUDE.md`, so it is already in context. It is a stub: **load
`skills/flow-contracts/pipeline.md` first**, which is canonical for the states and transitions —
git boundaries and the finish contract live in their own files, which it names; `/flow`'s own stage keys are in `skills/flow/SKILL.md`'s own
**Stage keys**, cited rather than repeated here.

**Input:** the change name or a description/Jira key to seed a new change, from `$ARGUMENTS` or the
conversation — and nothing else. **This command takes no flags.** If omitted at `IN_PROGRESS`, resolve
it per **Change name resolution (all `/flow*` commands)** (`skills/flow-contracts/pipeline.md`).
Report any argument that is not a change name, description, or fix instruction rather than ignoring
it.

**When done:** at `STARTED` after the plan gate, `/clear`, then `/flow <name>` to implement. At `IN_PROGRESS` with a fresh staged diff — after an implementation run or a fix — review
the staged diff — the stack is running — then re-run `/flow <name>` (or `/flow <name> <fix>`) as needed. At
`IN_PROGRESS` after choosing open PR or manual, there is nothing new to review — wait for the merge
or your manual steps, then re-run bare to archive. At `FINISHED`, nothing further.
