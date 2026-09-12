# Temporary artifacts registry

Every artifact the pipeline creates, with what creates it, where it lives, and what removes it.

**Loaded by `/flow`'s implement phase, `/flow`'s archive run and `/flow-fast`** — wherever an artifact this
table names is created or removed.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/artifacts-registry-rationale.md`;
**a `/flow*` run never loads it.**

## Temporary artifacts registry

Every artifact the pipeline creates, with what creates it, where it lives, and what removes it.

| Artifact | Created by | Location | Removed by |
|----------|-----------|----------|-----------|
| Per-task and review diffs | `/flow`'s implement phase | `<abs-worktree>/.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Panel slot verbatim reports | `/flow`'s review panel | `<abs-worktree>/.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Panel record | `/flow`'s implement phase | the store | nothing — the store is the terminal record |
| Self-review context bundle | run 2 step 9, on `defer` | `<project>/docs/self-review/<name>-context.md`, committed on `chore/archive-<name>` | `/flow-self-review <name>`, in the same commit as the report |
| SDD ledger | `/flow`'s implement phase | the store | nothing — the store is the terminal record |
| Rendered ledger and panel record | `flow record render` | `<abs-worktree>/.superpowers/sdd/` in the worktree; again under `<landing-worktree>` at run 2 step 9, for the self-review bundle | with the worktree, at run 2; with the landing worktree, at run 2 step 11 |
| Brainstorm design document | `/flow`'s creating run | `<abs-worktree>/.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Dispatch context bundle | `/flow`'s implement phase | `<abs-worktree>/.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Proposal artifact source | `/flow`'s creating run | the state directory | run 2, unconditionally |
| Worktree | `/flow`'s `flow.kickoff` | per the `worktrees` keys | run 2, after its existing checks |
| Local branch | `/flow`'s implement phase | the repository | run 2, `git branch -d` |
| Remote branch | finish run 1 | `origin` | run 2, without a further prompt |
| Archive branch | finish run 2 | the repository and `origin` | nothing in this pipeline — run 2 is terminal and the pull request outlives it |
| Change directory | `/flow`'s creating run | `<project>/spectre/changes/<name>/` | moved to the archive, never deleted |
| Workspace database and bucket | the project's `create` command, on first start in a worktree | inside the project's shared data services | run 2, the project's `remove` command |
| Claimed cache index | `/flow`'s implement phase, by probing, when it exports the workspace's variables | one of the shared cache's fixed indices | nothing in this pipeline — see below |
| State file | every command | the state directory | never — it is the terminal record |
| Bugbot's or Mutation's throwaway worktree copy | `/flow`'s review panel | sibling of the apply worktree, `<worktree>-<slot>-<round>` | the review panel itself, immediately after that slot's dispatch closes — never survives to run 2 |

**A change's spec edits are not an artifact and carry no row.** the implement phase's implementer writes
them directly into `<project>/spectre/specs/<capability>.md` on the change's branch, in the task
commit that implements the requirement, where they are ordinary source: the merge lands them and
there is nothing temporary to remove.

**This table is the one place a cleanup rule is stated.** Everything else that mentions a removal
points here rather than restating it. **Worktree cleanup**
(`skills/flow-contracts/finish-contract-run2.md`) is the *procedure* for the rows removed there, not a
second statement of the rule. See **Temporary artifacts registry**
(`skills/flow-contracts/artifacts-registry-rationale.md`) for why a stale second copy would be dangerous.

**An artifact no row accounts for is a defect in the registry**, corrected by adding the row —
never left unaccounted for on the grounds that something probably removes it. See **Temporary
artifacts registry** (`skills/flow-contracts/artifacts-registry-rationale.md`) for the incident that
established this.

**The proposal artifact source is not a record.** `/flow`'s creating run writes
`<state-dir>/<name>-proposal-artifact.html` so a revision round can republish to the same URL; the
published page outlives the file, and nothing in the repository preserves a copy. Run 2 deletes it
whether or not it exists — a `/flow-fast` change, which publishes none, has nothing to delete and
says so. The deletion is disclosed the same way the worktree removal is.

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
