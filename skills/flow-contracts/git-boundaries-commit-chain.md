# The guarded two-commit chain

Loaded by `/flow`'s integrate run 1 (**3. Commit the staged work**, `skills/flow/integrate.md`) and
by its implement phase only when the state file records a `prUrl` (**Stage, excluding the planning
paths**, `skills/flow/verify-and-handoff.md`). **Git boundaries** (`skills/flow-contracts/git-boundaries.md`)
is canonical for every other git boundary.

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
       || git -C <abs-worktree> commit -m "chore(spectre): plan <name>"; }
```

`<module>` is derived from the reshaped diff — the module carrying the change's substance, or a
broader area where it spans several, never a list. That is the same rule the creating run's
writing-plans stage applies to each task's `**Commit:**` field. The
planning message's subject is `chore(spectre): plan <name>`, never derived from the diff. At integrate its
message also lists anything the operator integrated over
(**1. Check for unfinished work**, `skills/flow/integrate.md`).

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
