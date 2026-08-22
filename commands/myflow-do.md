---
name: /myflow-do
id: myflow-do
category: myflow
description: Do — implement the plan with TDD and the review panel
---

**Model:** keep this **session** on Sonnet (or your default). See "Model policy" in `skills/myflow-contracts/pipeline.md`, which is canonical. Cursor doesn't yet support a per-command model frontmatter field, so the session setting is a recommendation rather than an enforced switch — but the subagent models are set at dispatch time and apply in every harness.

Use the **myflow-do** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Accepts **`STARTED`** (first run — creates the worktree) or **`IN_PROGRESS`** (fix run — resumes the **existing** worktree). Ends at **`IN_PROGRESS`** from `STARTED`; from `IN_PROGRESS` it writes the state back **unchanged**, because a fix never moves the state.

Produces **both** the staged diff **and** the run instructions, so reviewing the code and running the apps are one human gate.

**No commits, push, merge, or PR** — with one exception: if the state file records a `prUrl`, a PR is already open and a staged-only fix would be invisible on it, so the fix is committed and pushed to that branch instead.

Runs the project's lint and test commands before handing off, because **nothing runs them later** — `/myflow-finish` has no verification gate.

Also follow the myflow rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path. It is a stub: **load `skills/myflow-contracts/pipeline.md` first**, which is canonical for the states, transitions, git boundaries and the finish contract.

**Input:** the change name, from `$ARGUMENTS` or the conversation — and nothing else. **This command takes no flags.** If the name is omitted, run `openspec list --json` and use the sole relevant open change, asking which when there are several. Report any argument that is not a change name rather than ignoring it.

**When done:** review the staged diff and run the apps. Re-run `/myflow-do <name>` to fix anything you find, then `/myflow-finish <name>`.
