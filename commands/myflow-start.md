---
name: /myflow-start
id: myflow-start
category: myflow
description: Start — brainstorm, write the OpenSpec artifacts, publish the proposal artifact
---

**Model:** Opus (or your harness's strongest model) — brainstorming and design benefit most from stronger reasoning. Cursor doesn't yet support a per-command model frontmatter field, so switch manually in the composer picker before running this.

Use the **myflow-start** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Accepts **no change** (creates one) or **`STARTED`** (revises the existing proposal, republishing the artifact to the **same** URL). Ends at **`STARTED`**.

Runs Superpowers **#1** (brainstorming, with its design approval gate) and **#3** (writing-plans), woven into the OpenSpec artifacts, then publishes the proposal artifact the human reads.

**Never** writes code, creates a worktree, or creates a branch.

Also follow the myflow rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path. It is a stub: **load `skills/myflow-contracts/pipeline.md` first**, which is canonical for the states, transitions and the finish contract.

**Input:** the change name, from `$ARGUMENTS` or the conversation — and nothing else. **This command takes no flags.** If the name is omitted, run `openspec list --json` and use the sole relevant open change, asking which when there are several. Report any argument that is not a change name rather than ignoring it.

**When done:** read the artifact. Re-run `/myflow-start <name>` to revise the plan, or run `/myflow-do <name>` to implement it.
