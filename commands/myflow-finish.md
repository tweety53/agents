---
name: /myflow-finish
id: myflow-finish
category: myflow
description: Finish — integrate the branch, then archive and clean up once it is merged
---

**Model:** Sonnet (or your default) is fine here — Opus is reserved for `/myflow-start`'s brainstorming stage. Cursor doesn't yet support a per-command model frontmatter field, so this is a recommendation, not an enforced switch.

Use the **myflow-finish** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Accepts **`IN_PROGRESS`**. It runs **twice**, and the branch's merge status alone decides which run happens:

- **Run 1 — branch not merged.** Checks each worktree in the resolved set (see **Resolving a change's worktrees**, `skills/myflow-contracts/worktree-resolution.md`) for **unfinished work** first — before the landing question and before any git action — and offers exactly three courses: stop and finish it first *(recommended)*, continue anyway, or file **or join** a Jira follow-up and then continue. Then asks how it should land: open a pull request (default), merge and push, or leave it to you. Makes **two commits** — the implementation first, the planning artifacts second — takes that route, moves the linked issue to **In Review** whichever route was taken, and **stops** — the state stays `IN_PROGRESS`.
- **Run 2 — branch merged.** Verifies the merge, archives the change, **commits and pushes the archive**, removes the worktrees, the local branch, the **remote branch** and the proposal artifact source, then **verifies the cleanup** and writes **`FINISHED`**.

**Runs no tests, linters, or coverage check** before integrating — that happened during `/myflow-do`.

Every rule about what run 2 removes — which artifact, when, and on what condition — is stated once under **Temporary artifacts registry** (`skills/myflow-contracts/artifacts-registry.md`), which also points at the removal procedure and its safety checks. None of it is repeated here.

Also follow the myflow rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path. It is a stub: **load `skills/myflow-contracts/pipeline.md` first**, which is canonical for the states, transitions and the finish contract.

**Input:** the change name, from `$ARGUMENTS` or the conversation — and nothing else. **This command takes no flags.** If the name is omitted, run `spectre list --json` and use the sole relevant open change, asking which when there are several. Report any argument that is not a change name rather than ignoring it.

**When done:** run 1 hands back to `/myflow-finish <name>` — the same command — once the branch is merged. Run 2 is terminal.
