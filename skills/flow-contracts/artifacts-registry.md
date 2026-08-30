# Temporary artifacts registry

Every artifact the pipeline creates, with what creates it, where it lives, and what removes it.

**Loaded by `/myflow-do`, `/myflow-finish` run 2 and `/myflow-fast`** — wherever an artifact this
table names is created or removed.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/artifacts-registry-rationale.md`;
**a `/flow*` run never loads it.**

## Temporary artifacts registry

Every artifact the pipeline creates, with what creates it, where it lives, and what removes it.

| Artifact | Created by | Location | Removed by |
|----------|-----------|----------|-----------|
| Per-task and review diffs | `/myflow-do` | `<abs-worktree>/.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Panel slot verbatim reports | `/flow`'s review panel | `<abs-worktree>/.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Panel record | `/myflow-do` | the store | nothing — the store is the terminal record |
| SDD ledger | `/myflow-do` | the store | nothing — the store is the terminal record |
| Rendered ledger and panel record | `flow record render` | `<project>/docs/superpowers/` | nothing — they are committed and archived with the change |
| Dispatch context bundle | `/myflow-do` | `<abs-worktree>/.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Proposal artifact source | `/myflow-start` | the state directory | run 2, only if run 1's copy under `<project>/docs/superpowers/artifacts/` exists |
| Worktree | `/myflow-do` | per the `worktrees` keys | run 2, after its existing checks |
| Local branch | `/myflow-do` | the repository | run 2, `git branch -d` |
| Remote branch | finish run 1 | `origin` | run 2, without a further prompt |
| Archive branch | finish run 2 | the repository and `origin` | nothing in this pipeline — run 2 is terminal and the pull request outlives it |
| Change directory | `/myflow-start` | `<project>/spectre/changes/<name>/` | moved to the archive, never deleted |
| Workspace database and bucket | the project's `create` command, on first start in a worktree | inside the project's shared data services | run 2, the project's `remove` command |
| Claimed cache index | `/myflow-do`, by probing, when it exports the workspace's variables | one of the shared cache's fixed indices | nothing in this pipeline — see below |
| State file | every command | the state directory | never — it is the terminal record |
| Bugbot's throwaway worktree copy | `/flow`'s review panel | sibling of the apply worktree, `<worktree>-bugbot-<round>` | the review panel itself, immediately after that Bugbot dispatch closes — never survives to run 2 |

**A change's spec edits are not an artifact and carry no row.** `/myflow-do`'s implementer writes
them directly into `<project>/spectre/specs/<capability>.md` on the change's branch, in the task
commit that implements the requirement, where they are ordinary source: the merge lands them and
there is nothing temporary to remove.

**This table is the one place a cleanup rule is stated.** Everything else that mentions a removal
points here rather than restating it. **Worktree cleanup**
(`skills/flow-contracts/finish-contract.md`) is the *procedure* for the rows removed there, not a
second statement of the rule. See **Temporary artifacts registry**
(`skills/flow-contracts/artifacts-registry-rationale.md`) for why a stale second copy would be dangerous.

**An artifact no row accounts for is a defect in the registry**, corrected by adding the row —
never left unaccounted for on the grounds that something probably removes it. See **Temporary
artifacts registry** (`skills/flow-contracts/artifacts-registry-rationale.md`) for the incident that
established this.

**Where the proposal artifact source comes from, and what produces the copy its row tests.**
`/myflow-start` writes `<state-dir>/<name>-proposal-artifact.html` so a revision round can republish
to the same URL, and the preserved copy its row requires lives under
`<project>/docs/superpowers/artifacts/`. **Finish run 1 is what puts it there**, by copying it before
it stages — see **Run 1 — the branch is not merged**
(`skills/flow-contracts/finish-contract.md`), which is canonical for that copy, for the change-name
and containment checks it makes first, and for the skip when a change published no artifact. The
condition is therefore reachable in both directions: a change whose artifact run 1 copied is deleted
at run 2, and a `/myflow-fast` change, which publishes none, is not. No preserved copy → leave the
file and say so. The deletion is disclosed the same way the worktree removal is. See **Temporary
artifacts registry** (`skills/flow-contracts/artifacts-registry-rationale.md`) for why the row is
conditional.

**The workspace row belongs only to a project that declares isolation, and for every other project
it is a row about nothing — which is why it names no database, no bucket and no service.** A project
declares the commands that create these resources, that remove them, and that report which of them
survived, in
**Project configuration** (`skills/flow-contracts/project-configuration.md`).
Which resources there are, and how each derived value is derived, is stated under
**What the id derives** (`skills/flow-contracts/workspace-isolation.md`).

**This is the one row whose removal is verified by asking rather than by looking: a survivor is
established from the project's own survivor report, never inferred from the removal's exit code**
— stated once under **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`), with
the report's output and exit-code contract under **Project configuration**
(`skills/flow-contracts/project-configuration.md`). A report that could not reach its service is
skipped rather than failed. See **Temporary artifacts registry**
(`skills/flow-contracts/artifacts-registry-rationale.md`) for why asking, not looking, is required here.

**Nothing removes the archive branch either, on `origin` or in the repository.** See **Temporary
artifacts registry** (`skills/flow-contracts/artifacts-registry-rationale.md`) for why, and for
design.md's open question `archive-branch-cleanup`.

**Nothing removes the claimed cache index, and nothing in this pipeline can.** It is not written
into the state file, and the project's `remove` command does not touch it either — stated as a
property of the `cache index` resource word under **Project configuration**
(`skills/flow-contracts/project-configuration.md`). See **Temporary artifacts registry**
(`skills/flow-contracts/artifacts-registry-rationale.md`) for why: guessing an index to sweep risks flushing
another workspace's.
