# The guarded commit chain

Loaded by `/flow`'s integrate run 1 (**3. Commit the staged work**, `skills/flow/integrate.md`) and
by its implement phase only when the state file records a `prUrl` (**Stage, excluding the planning
paths**, `skills/flow/verify-and-handoff.md`). **Git boundaries** (`skills/flow-contracts/git-boundaries.md`)
is canonical for every other git boundary.

**The commit is guarded, and an empty one is skipped rather than failed.** The commit is
preceded by a staged-changes test, and the whole sequence is one `&&` chain, run as a single
command.

```bash
git -C <abs-worktree> add -A -- . \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "<type>(<module>): <what the implementation does>"; }
```

`<module>` is derived from the reshaped diff — the module carrying the change's substance, or a
broader area where it spans several, never a list. That is the same rule the creating run's
writing-plans stage applies to each task's `**Commit:**` field. At integrate its
message also lists anything the operator integrated over
(**1. Check for unfinished work**, `skills/flow/integrate.md`).

**The change directory is left out by the ignore file alone.** `<project>/spectre/changes/` is
gitignored, so `add -A` stages nothing under it; `flow-guard commit-split` refuses, exit 2 naming the
path, a worktree where it is not ignored.

**A skipped commit is reported, and a FAILED commit — one a hook rejects — is a git failure: report
git's own output and stop.**

**A planning path that is a tracked symlink stops the run, and is never worked around.** When
`<project>/spectre/changes/` is a symlink — or `<project>/spectre/` is, putting it behind one — the
`git add -A -- .` call exits 128 with
`fatal: pathspec … is beyond a symbolic link` and stages **nothing at all**. Report that message,
name the path, and stop at `IN_PROGRESS`. The fix belongs in the repository, by making the path a real directory.
