---
model: sonnet
description: Reduced-ceremony /flow variant — direct-write brainstorm, inline TDD implementation, a fixed primary+simple-reviewer panel, seven guards, a one-run merge-and-push finish, pausing only at the IN_PROGRESS review gate
---

Use the **flow-fast** skill — installed globally, so let your harness resolve it by name rather
than assuming a project-local path.

Follow that skill exactly. Accepts **no state** (creates a change), **`STARTED`** (resumes a
creating run that stopped before implementation), or **`IN_PROGRESS`**. On a creating run it writes
`STARTED` immediately, then reads the project context and writes `proposal.md`/`design.md`/
`tasks.md` directly — no `superpowers:brainstorming`, no `superpowers:writing-plans`, no
design-approval gate, asking only for a true blocker — then implementation inline, task by task,
TDD, targeted tests and lint only, then a fixed two-slot review panel (`primary` +
`simple-reviewer`, Critical/Major fixed inline, every Minor deferred, `simple-reviewer` re-run
alone on a fix delta), ending at `IN_PROGRESS`. Re-invoked with an argument at `IN_PROGRESS`, the
argument is fix instructions. Re-invoked bare at `IN_PROGRESS`, it lands by **merge-and-push
only** — no landing question, no pull-request or manual route — merging, archiving on the base
branch and pushing once in this one run, through to `FINISHED`.

Publishes no proposal artifact, no rendered ledger or panel record, and no self-review. Asks no
planning-effort, model, or review-panel-roster question — none of those is dynamic for
`/flow-fast`: execution is always inline, and the roster is always `primary` + `simple-reviewer`,
never the settings-store list.

Also follow the flow rule (`flow-manual-review.mdc`) — installed globally, so let your harness
resolve it rather than assuming a project-local path. It is a stub: **load
`skills/flow-contracts/pipeline.md` first**, which is canonical for the states, transitions, git
boundaries and the finish contract; `/flow-fast`'s own stage keys are in
`skills/flow-fast/SKILL.md`'s own **Stage keys**, cited rather than repeated here.

**Input:** the change name or a description/Jira key to seed a new change, from `$ARGUMENTS` or the
conversation — and nothing else. **This command takes no flags.** If omitted at `IN_PROGRESS`, run
`spectre list --json` and use the sole relevant open change, asking which when there are several.
Report any argument that is not a change name, description, or fix instruction rather than ignoring
it.

**When done:** at `IN_PROGRESS` with a fresh staged diff — after a creating run or a fix — review
the staged diff and run the apps, then re-run `/flow-fast <name>` (or `/flow-fast <name> <fix>`) as
needed, or re-run it bare to land the branch and finish in one run. At `FINISHED`, nothing further.
A change started under `/flow-fast` may finish under `/flow` and vice versa — both read and write
the same state record.
