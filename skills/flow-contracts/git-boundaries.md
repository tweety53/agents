# Git boundaries

Which git actions each command may take, and the guarded two-commit chain that enforces the split
between implementation and planning artifacts.

**Loaded by `/flow`'s implement phase, bare `/flow` and `/flow-fast`** — at the step that commits or stages.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/git-boundaries-rationale.md`;
**a `/flow*` run never loads it.**

## Git boundaries

| Command | Condition | Allowed git actions |
|---------|-----------|---------------------|
| `/flow`'s creating run | — | **Creates the worktree and pushes its empty branch** at kickoff; stages planning artifacts there and never commits |
| `/flow`'s implement phase | from `STARTED` | Resume the kickoff worktree + **commits each task** (fixups fold in), **pushing after each** — no merge or PR |
| `/flow`'s implement phase | at `IN_PROGRESS`, no `prUrl` | Resume **existing** worktree + **commits fixups** the same way, pushing after each — no merge or PR |
| `/flow`'s implement phase | at `IN_PROGRESS`, `prUrl` recorded | **Commits twice and pushes `--force-with-lease`** to the PR branch — implementation, then planning artifacts; the one planning-commit exception |
| bare `/flow` | run 1 | **Commits twice** — implementation, then planning artifacts — and pushes `--force-with-lease`; opens a PR or merges, by the operator's choice |
| bare `/flow` | run 2, before self-review | **Commits** the archive on `chore/archive-<name>` — never `<base>` — in the landing worktree, and removes worktrees and branches |
| bare `/flow` | run 2, during self-review | **Commits** the self-review report, or the context bundle on `## self review: defer`, on `chore/archive-<name>` — a second, separate commit, in the landing worktree, and still no push |
| bare `/flow` | run 2, after self-review | **Pushes** `chore/archive-<name>` once, carrying both commits, from the landing worktree, and opens its pull request — never pushes `<base>` |
| `/flow-status` | — | None — read-only |
| `/flow-plan` | staging note captured | **Commits once** — the note, its plan and its decision — on `plan-<stem>` in its research worktree, and pushes that commit to `<default-branch>` (**Landing the note**, `skills/flow-plan/SKILL.md`); nothing else, ever |

**No command touches the main checkout.** `/flow` creates `<project>/.worktrees/<name>` inside
`flow.kickoff`, `/flow-fast` inside its kickoff, `/flow-plan` a `_plan-<stem>` research worktree
at its start, and every read, write, stage, commit and push in the table above happens in a
worktree. The main checkout is never checked out, staged, committed or written, whatever branch
it sits on.

## Branch backup

A change's branch lives on the remote from the moment its worktree exists: `git worktree add` is
followed by `git push -u origin <branch>`, and every commit a run makes on the branch — a task
commit, a fixup, a panel fix — is followed by `git push origin <branch>` from the same parent
call that made or verified it. A lost worktree is then rebuilt with `git worktree add <path>
origin/<branch>`. The one rewrite is integrate's reshape (`reset --soft` to the merge base, then
the two commits), so that run's push is `git push --force-with-lease origin <branch>`; every other
push is plain. `/flow-plan`'s research branch is the exception — it is landed onto
`<default-branch>` and deleted in the same session, so it is never pushed under its own name.

**The planning paths** are the two that
**Handoff output** (`skills/flow-contracts/pipeline.md`) names. `/flow`'s implement phase clears them from the index and only
then stages with them excluded by pathspec — an exclusion governs what an
`add` adds and cannot retract what an earlier step staged, so the clearing pass is what makes the
rule hold rather than merely assert it. Its staging area therefore carries implementation only, and
bare `/flow` is what commits them.

**A capability spec is implementation, not planning.** `<project>/spectre/specs/<capability>.md`
states what the system must do, so changing it changes the product exactly as code does: the
implementer writes and commits it on the change branch, in the task commit that implements the
requirement, and it never reaches the planning commit. Within the spectre tree the boundary is the
directory — `<project>/spectre/changes/` is planning and `<project>/spectre/specs/` is not — which is
why every pathspec below names the change folder rather than the tree.

**`link.md` is implementation too, carved out by file, not directory.** It states which
repositories a change spans, its branch and merge order — structural, not the
`proposal.md`/`design.md`/`tasks.md` planning narrative — and in a satellite repository it is the
change's only content, so the directory boundary above would leave that branch with no
implementation commit. A `:(exclude)` pathspec beats a positive one in the same `add`, so it needs
its own `add` after the exclude, guarded so a `link.md`-less change still commits cleanly —
`<agents repo>/scripts/commit-split.sh` guards it.

`git add -A` respects `<project>/.gitignore`. Never force-add.

**Both commits are guarded, and an empty one is skipped rather than failed.** Each commit is
preceded by a staged-changes test, and the whole sequence is one `&&` chain, run as a single
command. See **Git boundaries** (`skills/flow-contracts/git-boundaries-rationale.md`) for the ordinary
cases this guards against and why it is a chain rather than `set -e`.

```bash
git -C <abs-worktree> reset -q -- spectre/changes/ docs/superpowers/ \
  && git -C <abs-worktree> add -A -- . ':(exclude)spectre/changes/' ':(exclude)docs/superpowers/' \
  && { git -C <abs-worktree> add -A -- 'spectre/changes/<id>/link.md' 2>/dev/null || true; } \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "<type>(<module>): <what the implementation does>"; } \
  && git -C <abs-worktree> add -A \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "chore(spectre): plan and session records"; }
```

`<module>` is derived from the reshaped diff — the module carrying the change's substance, or a
broader area where it spans several, never a list. That is the same rule the creating run's
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
