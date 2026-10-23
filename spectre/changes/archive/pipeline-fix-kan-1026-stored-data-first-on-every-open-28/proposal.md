# pipeline-fix-kan-1026-stored-data-first-on-every-open-28

## Why

Run 2 cleanup of gymie's kan-1026 stopped on the disclosure stop (exit 3) for
`UNCLASSIFIED: /Users/tweety53/Projects/gymie-worktrees/kan-1026-stored-data-first-on-every-open — .env`.
That `.env` was a symlink to `/Users/tweety53/Projects/gymie/.env`, the main checkout's own file,
made by gymie's dev tooling `linkEnvFile` (`/Users/tweety53/Projects/gymie/.flow/project.md:104`);
removing it loses nothing. Check 4's preservation rule (**Worktree cleanup**,
`skills/flow-contracts/finish-contract-run2.md`) covers only a regular file with a distinct regular
file at the same path in the main checkout, so every gymie worktree's cleanup stops on its linked
`.env` and asks the operator a question about nothing.

## What changes

- `remove-change-worktrees.sh` (Go: `stats/internal/guard/removechangeworktrees.go`) treats an
  unclassified entry at the worktree root that is a symlink resolving to the regular file at the
  same path in the main checkout as preserved: its line ends
  `— preserved: linked to <main-checkout>`, it never makes the disclosure stop, and it is relayed
  per check 4.
- `TestRemoveChangeWorktreesDisclose` gains the case `an entry linked to the main checkout's proceeds`,
  which exits 3 without the fix.
- The preservation sentence in `skills/flow-contracts/finish-contract-run2.md` and the script's
  header name the symlink case and its label.
