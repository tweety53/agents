# faster-integrate

## worktree-cleanup-preserved

### Why

On KAN-924, `remove-change-worktrees.sh` stopped run 2's cleanup with exit 3 for a `.env` byte-identical to the main checkout's `.env`. **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`) asks only about an entry that is irreplaceable and unpreserved; an entry that still exists, unchanged, in the main checkout is preserved, so the stop asked a question the contract never asks.

### What changes

- Check 4 marks an unclassified entry preserved when it is a regular file byte-identical to the same path in the repository's main checkout (the first entry of `git worktree list --porcelain`); its `UNCLASSIFIED:` line ends `— preserved: identical in <main-checkout>`.
- The disclosure stop (exit 3) fires only for an unpreserved unclassified entry or a wave-group copy; preserved entries are shown and the removal proceeds in the same call.
