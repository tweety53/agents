# faster-integrate

## worktree-cleanup-preserved

### Why

On KAN-924, `remove-change-worktrees.sh` stopped run 2's cleanup with exit 3 for a `.env` byte-identical to the main checkout's `.env`. **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`) asks only about an entry that is irreplaceable and unpreserved; an entry that still exists, unchanged, in the main checkout is preserved, so the stop asked a question the contract never asks.

### What changes

- Check 4 marks an unclassified entry preserved when it is a regular file byte-identical to the same path in the repository's main checkout (the first entry of `git worktree list --porcelain`); its `UNCLASSIFIED:` line ends `— preserved: identical in <main-checkout>`.
- The disclosure stop (exit 3) fires only for an unpreserved unclassified entry or a wave-group copy; preserved entries are shown and the removal proceeds in the same call.

## self-review-minor-no-rereview

### Why

The self-review fix loop re-reviewed after every fix round, Minor-only rounds included, while the review panel already closes a Minor-only round without re-running a slot (**Panel re-runs**, `skills/flow/review-panel.md`). The extra re-review dispatch cost a round on every self-review pass whose review raised only Minors.

### What changes

- Step 3 of `skills/flow-self-review/SKILL.md`: the review grades each finding Critical, Important or Minor; a review whose every finding is Minor is closed by the one fix dispatch, folded into each finding's commit as before, with no re-review — step 4's lint and tests are its verification.
- Critical and Important findings keep the fix → re-review loop, which now ends at a clean or Minor-only review.
