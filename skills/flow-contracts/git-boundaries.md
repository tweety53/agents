# Git boundaries

Which git actions each command may take, the planning commits, and the guarded two-commit chain
that enforces the split between implementation and planning artifacts.

**Loaded by `/flow`'s implement phase, bare `/flow` and `/flow-fast`** — at the step that commits or stages.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/git-boundaries-rationale.md`;
**a `/flow*` run never loads it.**

## Git boundaries

| Command | Condition | Allowed git actions |
|---------|-----------|---------------------|
| `/flow`'s creating run | — | **Creates the worktree and pushes its empty branch** at kickoff; makes the plan-gate planning commit once the plan gate answers **Yes** (**Planning commits** below), and nothing else |
| `/flow`'s implement phase | from `STARTED` | Resume the kickoff worktree + **commits each task** (fixups fold in) and the **Planning commits** below, **pushing after each** — no merge or PR |
| `/flow`'s implement phase | at `IN_PROGRESS`, no `prUrl` | Resume **existing** worktree + **commits fixups** and planning commits the same way, pushing after each — no merge or PR |
| `/flow`'s implement phase | at `IN_PROGRESS`, `prUrl` recorded | **Commits twice and pushes `--force-with-lease`** to the PR branch — implementation, then whatever planning delta the last planning commit left |
| bare `/flow` | run 1 | **Reshapes** — keeps every planning commit, collapses task and fixup commits — then **commits twice** — implementation, then the planning delta — and pushes `--force-with-lease`; opens a PR or merges, by the operator's choice |
| bare `/flow` | run 2, before self-review | **Commits** the archive on `chore/archive-<name>` — never `<base>` — in the landing worktree, and removes worktrees and branches |
| bare `/flow` | run 2, during self-review | **Commits** the self-review report, or the context bundle on `## self review: defer`, on `chore/archive-<name>` — a second, separate commit, in the landing worktree, and still no push |
| bare `/flow` | run 2, after self-review | **Pushes** `chore/archive-<name>` once, carrying both commits, from the landing worktree, and opens its pull request — never pushes `<base>` |
| `/flow-status` | — | None — read-only |
| `/flow-plan` | change captured | **Commits once** — the planning artifacts, `chore(spectre): plan` — on `spectre/<name>` in the change worktree, and pushes it (**Capturing a new change**, `skills/flow-plan/SKILL.md`); nothing else, ever |

**No command writes the main checkout.** `/flow` creates `<project>/.worktrees/<name>` inside
`flow.kickoff`, `/flow-fast` inside its kickoff, `/flow-plan` through that same kickoff at
capture — it reads the main checkout before then and writes nothing — and every write, stage,
commit and push in the table above happens in a worktree. The main checkout is never checked out,
staged, committed or written, whatever branch it sits on.

## Planning commits

**Planning artifacts are committed in their own commits, at fixed boundaries, and never inside a
task or fixup commit.** Every planning commit is pathspec-scoped to the change folder, carries no
`Task-Id:` trailer — which is what keeps it outside `check-task-commit-planning-paths.sh`'s
task-commit check — and is pushed plain (**Branch backup** below):

| Boundary | Carries | Subject |
|----------|---------|---------|
| The plan gate answers **Yes** (**Plan review gate**, `skills/flow/brainstorm-planner.md`) | `<project>/spectre/changes/<name>/` | `chore(spectre): plan` |
| After each `spectre link` in `flow.isolate-workspace` (`skills/flow/implement.md`), before the next link | each `<project>/spectre/changes/<name>/link.md` the link wrote — the canonical worktree's and the satellite's — in the worktree holding it | `chore(spectre): link <peer>` |
| Before every reviewer dispatch — a gated per-task reviewer bundle, a panel round, a re-run | `<project>/spectre/changes/<name>/` | `chore(spectre): plan` |
| `flow.document-fix` has appended a fix run's tasks, before the first implementer dispatch | `<project>/spectre/changes/<name>/` | `chore(spectre): plan` |
| `flow.write-in-progress` has appended the narrative, before the handoff | `<project>/spectre/changes/<name>/` | `chore(spectre): plan` |

```bash
check-planning-commit-location.sh <abs-worktree> <name> \
  && git -C <abs-worktree> add -A -- <path> \
  && { git -C <abs-worktree> diff --cached --quiet -- <path> \
       || git -C <abs-worktree> commit -m "<subject>" -- <path>; } \
  && git -C <abs-worktree> push origin <branch>
```

`<path>` is the table's **Carries** cell with its `<project>/` prefix dropped — relative to the worktree root — and an empty delta skips the commit rather than failing
it, exactly as the two-commit chain below skips. `<peer>` is the peer name the link command was
given. **`spectre link` is never run with `--force`**: it refuses while the canonical change
directory carries uncommitted modifications, and the answer to that refusal is the planning commit
above, never an override. Integrate's reshape (**Branch backup** below) keeps every planning commit
as the separate commit it was made as, so the landed branch carries each one.

**A planning commit lands only in the change's own worktree, on `spectre/<name>`.**
`check-planning-commit-location.sh <abs-worktree> <name>` answers `PLANNING-COMMIT-LOCATION-OK` (exit
0), or prints `PLANNING-COMMIT-MAIN-CHECKOUT` and/or `PLANNING-COMMIT-WRONG-BRANCH` (exit 1), or
cannot answer (exit 2); anything but 0 stops the run before anything is staged, reporting the
guard's lines. It matters most for a link commit: `spectre link` runs from a repository's primary
checkout, which is where a link commit made from the wrong directory would land.
`commit-split.sh` and `reshape-branch.sh` run it themselves. **When the guard cannot be located**,
check by hand before committing: `git -C <abs-worktree> rev-parse --path-format=absolute --git-dir`
must differ from `git -C <abs-worktree> rev-parse --path-format=absolute --git-common-dir`, and
`git -C <abs-worktree> branch --show-current` must print `spectre/<name>`; otherwise stop without
committing.

## Branch backup

A change's branch lives on the remote from the moment its worktree exists: `git worktree add` is
followed by `git push -u origin <branch>`, and every commit a run makes on the branch — a task
commit, a fixup, a panel fix — is followed by `git push origin <branch>` from the same parent
call that made or verified it. A lost worktree is then rebuilt with `git worktree add <path>
origin/<branch>`. The one rewrite is integrate's reshape (`reshape-branch.sh`, then
the two commits), so that run's push is `git push --force-with-lease origin <branch>`; every other
push is plain.

**The planning path** is the one **Handoff output** (`skills/flow-contracts/pipeline.md`) names.
`/flow`'s implement phase clears it from the index and only then stages with it excluded by
pathspec — an exclusion governs what an
`add` adds and cannot retract what an earlier step staged, so the clearing pass is what makes the
rule hold rather than merely assert it. Its staging area therefore carries implementation only;
planning reaches the branch through **Planning commits** below and nothing else.

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

**A commit a run instructs defaults to the pathspec-scoped form.** `git commit -m "<subject>" -- <paths>`:
the commit carries only the paths it names, whatever else the index holds. A plain commit takes
the whole staged tree with it, so wherever a run's instructions know the paths a commit should carry,
the commit names them. The guarded two-commit chain below is the deliberate exception: its
`add -A` is what picks up operator edits and new files at the human gate, which a
pathspec-scoped commit would drop.

**Both commits are guarded, and an empty one is skipped rather than failed.** Each commit is
preceded by a staged-changes test, and the whole sequence is one `&&` chain, run as a single
command. See **Git boundaries** (`skills/flow-contracts/git-boundaries-rationale.md`) for the ordinary
cases this guards against and why it is a chain rather than `set -e`.

```bash
git -C <abs-worktree> reset -q -- spectre/changes/ \
  && git -C <abs-worktree> add -A -- . ':(exclude)spectre/changes/' \
  && { git -C <abs-worktree> add -A -- 'spectre/changes/<id>/link.md' 2>/dev/null || true; } \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "<type>(<module>): <what the implementation does>"; } \
  && git -C <abs-worktree> add -A \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "chore(spectre): plan"; }
```

`<module>` is derived from the reshaped diff — the module carrying the change's substance, or a
broader area where it spans several, never a list. That is the same rule the creating run's
writing-plans stage applies to each task's `**Commit:**` field. The
planning message is a **fixed literal**, never derived — every planning commit stages the same two
trees in every change, so there is nothing about it that varies.

**A skipped commit is reported, and a FAILED commit — one a hook rejects — is a git failure: report
git's own output and stop.** See **Git boundaries** (`skills/flow-contracts/git-boundaries-rationale.md`)
for what an unguarded sequence would do instead.

**A planning path that is a tracked symlink stops the run, and is never worked around.** When
`<project>/spectre/changes/` is a symlink — or `<project>/spectre/` is, putting it behind one — the
`git add -A -- . ':(exclude)spectre/changes/'` call exits 128 with
`fatal: pathspec … is beyond a symbolic link` and stages **nothing at all**. Report that message,
name the path, and stop at `IN_PROGRESS`. The only way to stage past it is a bare `git add -A`,
which puts the planning artifacts into the implementation commit — the one outcome this split
exists to prevent — so the fix belongs in the repository, by making the path a real directory.
