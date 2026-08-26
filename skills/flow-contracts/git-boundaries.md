# Git boundaries

Which git actions each command may take, and the guarded two-commit chain that enforces the split
between implementation and planning artifacts.

**Loaded by `/myflow-do`, `/myflow-finish` and `/myflow-fast`** — at the step that commits or stages.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/git-boundaries-rationale.md`;
**a `/myflow-*` run never loads it.**

## Git boundaries

| Command | Condition | Allowed git actions |
|---------|-----------|---------------------|
| `/myflow-start` | — | **None — stages planning artifacts and never commits** |
| `/myflow-do` | from `STARTED` | Create branch/worktree + **commits each task** (fixups fold in) — no push, merge, or PR |
| `/myflow-do` | at `IN_PROGRESS`, no `prUrl` | Resume **existing** worktree + **commits fixups** the same way — no push, merge, or PR |
| `/myflow-do` | at `IN_PROGRESS`, `prUrl` recorded | **Commits twice and pushes** to the PR branch — implementation, then planning artifacts; the one exception |
| `/myflow-finish` | run 1 | **Commits twice** — implementation, then planning artifacts — and pushes; opens a PR or merges, by the operator's choice |
| `/myflow-finish` | run 2, before self-review | **Commits** the archive on `chore/archive-<name>` — never `<base>` — and removes worktrees and branches |
| `/myflow-finish` | run 2, during self-review | **Commits** the self-review report on `chore/archive-<name>` — a second, separate commit, and still no push |
| `/myflow-finish` | run 2, after self-review | **Pushes** `chore/archive-<name>` once, carrying both commits, and opens its pull request — never pushes `<base>` |
| `/myflow-status` | — | None — read-only |

**The planning paths** are the two that
**Handoff output** (`skills/flow-contracts/pipeline.md`) names. `/myflow-do` clears them from the index and only
then stages with them excluded by pathspec — an exclusion governs what an
`add` adds and cannot retract what an earlier step staged, so the clearing pass is what makes the
rule hold rather than merely assert it. Its staging area therefore carries implementation only, and
`/myflow-finish` is what commits them.

**A capability spec is implementation, not planning.** `<project>/spectre/specs/<capability>.md`
states what the system must do, so changing it changes the product exactly as code does: the
implementer writes and commits it on the change branch, in the task commit that implements the
requirement, and it never reaches the planning commit. Within the spectre tree the boundary is the
directory — `<project>/spectre/changes/` is planning and `<project>/spectre/specs/` is not — which is
why every pathspec below names the change folder rather than the tree.

`git add -A` respects `<project>/.gitignore`. Never force-add.

**Both commits are guarded, and an empty one is skipped rather than failed.** Each commit is
preceded by a staged-changes test, and the whole sequence is one `&&` chain, run as a single
command. See **Git boundaries** (`skills/flow-contracts/git-boundaries-rationale.md`) for the ordinary
cases this guards against and why it is a chain rather than `set -e`.

```bash
git -C <abs-worktree> reset -q -- spectre/changes/ docs/superpowers/ \
  && git -C <abs-worktree> add -A -- . ':(exclude)spectre/changes/' ':(exclude)docs/superpowers/' \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "<type>(<module>): <what the implementation does>"; } \
  && git -C <abs-worktree> add -A \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "chore(spectre): plan and session records"; }
```

`<module>` is derived from the reshaped diff — the module carrying the change's substance, or a
broader area where it spans several, never a list. That is the same rule `/myflow-start`'s
writing-plans stage applies to each task's `**Commit:**` field. The
planning message is a **fixed literal**, never derived — every planning commit stages the same two
trees in every change, so there is nothing about it that varies.

**A skipped commit is reported, and a FAILED commit — one a hook rejects — is a git failure: report
git's own output and stop.** See **Git boundaries** (`skills/flow-contracts/git-boundaries-rationale.md`)
for what an unguarded sequence would do instead.

**A planning path that is a tracked symlink stops the run, and is never worked around.** When either
of the two is a symlink — or `<project>/spectre/` is, putting `<project>/spectre/changes/` behind
one — the
`git add -A -- . ':(exclude)spectre/changes/' ':(exclude)docs/superpowers/'` call exits 128 with
`fatal: pathspec … is beyond a symbolic link` and stages **nothing at all**. Report that message,
name the path, and stop at `IN_PROGRESS`. The only way to stage past it is a bare `git add -A`,
which puts the planning artifacts into the implementation commit — the one outcome this split
exists to prevent — so the fix belongs in the repository, by making the path a real directory.
