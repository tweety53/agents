# pipeline-fix-kan-999-group-session-workout-names-25

## Why

`/flow` run 2 cleanup of gymie kan-999 stopped with `STOP: disclose`: `remove-change-worktrees.sh`
check 4 printed `UNCLASSIFIED: <worktree> — .env` and exited 3, `integrate-change.sh` relayed it, and
the run asked the operator whether to delete the worktree. The `.env` is the worktree's own copy of
the project's gitignored local dev env file — the main checkout holds `.env` with the same keys, the
worktree's carries its isolated ports and `ALLOWED_ORIGINS` — so it differs and was not
`preserved: identical`. The answer to that ask is always proceed; the operator's instruction is
"update flow not to ask such obvious questions".

## What changes

- Check 4 marks an unclassified entry preserved when it is a regular file and the same path in the
  repository's main checkout is a distinct regular file, whether or not the bytes match. The line
  ends `— preserved: identical in <main-checkout>` when they do and
  `— preserved: differs from <main-checkout>` when they do not; either is shown and never stops the
  removal.
- An entry with no counterpart in the main checkout, or one the main checkout reaches through a
  symlink into the worktree, is still unpreserved and still makes the disclosure stop.
- `finish-contract-run2.md` and the script's header state the new rule; the Go guard test pins it.
