---
name: /myflow-finish
id: myflow-finish
category: myflow
description: Finish — integrate the branch, then archive and clean up once it is merged
---

**Model:** Sonnet (or your default) is fine here — Opus is reserved for `/myflow-start`'s brainstorming stage. Cursor doesn't yet support a per-command model frontmatter field, so this is a recommendation, not an enforced switch.

Use the **myflow-finish** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Accepts **`IN_PROGRESS`**. It runs **twice**, and the branch's merge status alone decides which run happens:

- **Run 1 — branch not merged.** Asks up front how it should land: open a pull request (default), merge and push, or leave it to you. Commits the staged work, takes that route, and **stops** — the state stays `IN_PROGRESS`.
- **Run 2 — branch merged.** Verifies the merge, syncs delta specs, archives the change, **commits and pushes the archive**, removes the worktrees and branches, and writes **`FINISHED`**.

**Runs no tests, linters, or coverage check** before integrating — that happened during `/myflow-do`.

Worktree removal runs four gating checks first (no uncommitted tracked changes, no untracked-and-unignored files, no commits that exist only here, local stack stopped) and **leaves everything alone if any of them fails**. It then discloses the ignored files `--force` will destroy and asks before removing.

Also follow the myflow rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path. It is a stub: **load `skills/myflow-contracts/pipeline.md` first**, which is canonical for the states, transitions, git boundaries and the finish contract.

**Input:** the change name, from `$ARGUMENTS` or the conversation — and nothing else. **This command takes no flags.** If the name is omitted, run `openspec list --json` and use the sole relevant open change, asking which when there are several. Report any argument that is not a change name rather than ignoring it.

**When done:** run 1 hands back to `/myflow-finish <name>` — the same command — once the branch is merged. Run 2 is terminal.
