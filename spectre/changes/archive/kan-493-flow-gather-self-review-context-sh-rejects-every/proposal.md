# kan-493-flow-gather-self-review-context-sh-rejects-every

## Why

`gather-self-review-context.sh`'s optional `<repo-root>` override validates the supplied path by
deriving a repository root from `git -C <root> rev-parse --git-common-dir` and requiring the
`dirname` of that to equal `<root>`. For any linked worktree — which `<landing-worktree>` always
is — `--git-common-dir` resolves to the *main checkout's* `.git`, never the worktree's own
directory, so the equality check can never pass and the script exits 2. `skills/flow/archive.md`
step 9 documents exactly that invocation (the fourth argument must be the landing worktree, since
run 2's archive step moves the archived path physically under it, off the main checkout), so
every cross-repo change's self-review step is unrunnable as documented. Confirmed live on the
KAN-459 `/flow` run; worked around there by reading the four context files directly instead of
invoking the script. Also covers the parallel flow-automation finding for the same bug — one
issue, one fix.

## What changes

- The `<repo-root>` override validates with `git -C <root> rev-parse --show-toplevel` equal to
  `<root>` instead of the git-common-dir equality: a main checkout root and a linked worktree
  root both pass; a repository subdirectory, a non-repository directory, a relative path and a
  symlink-divergent path still exit 2.
- The script's own comments that state the old refusal (the header NOTE on `<repo-root>` and the
  `REPO_ROOT_OVERRIDE` block) are updated to state the new contract; the process-cwd derivation
  in `validate_archived_path()` step 1 keeps `--git-common-dir` unchanged (F23 — there the main
  checkout *is* the correct anchor).
- The harness gains one case that passes a linked worktree as `<repo-root>` with cwd outside the
  repository.
