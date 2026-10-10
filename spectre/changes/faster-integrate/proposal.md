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

## integrate-change-script

### Why

A bare `/flow` at `IN_PROGRESS` was far too slow. On gymie's KAN-924 the parent spent about fifty tool calls and loaded about ten large contract files — `integrate.md`, `cleanup.md`, both finish contracts, `worktree-resolution.md`, `git-boundaries-commit-chain.md`, `session-records.md` and more — only to call existing guards one at a time, per worktree, across three repositories. None of those calls needed judgement.

### What changes

- `scripts/integrate-change.sh` (symlinked into `skills/flow/scripts/`) runs every mechanical step of run 1 and run 2 in four subcommands — `prepare`, `commit`, `land`, `cleanup` — sequencing the existing guards and making the stage marks. It re-implements no guard: `finish-contract-run1.md` and `finish-contract-run2.md` stay canonical for every step, and each guard's header for its verdicts. Each call ends on one `NEXT:` or `STOP: <kind>` line; a stop closes the open mark `stopped`.
- `skills/flow/integrate.md` and `skills/flow/cleanup.md` keep only the calls, the judgement the parent owns — commit subjects, the narrative, the self-review pass, the landing question, the Jira transitions — and what to do at each stop. The finish contracts, `artifacts-registry.md`, `sync-onto-base.md` and `unfinished-work-gate.md` load only at the stop that needs them. The finish session's definite load set drops from about 25.6k to about 8.1k tokens (`scripts/load-sets.sh`).
- `check-stage-mark-calls.sh` holds an `integrate-change.sh` call to `stage begin`'s session-token and harness rules, since the token the skill writes there is the one the script's marks carry.
- `scripts/test-integrate-change.sh` drives the real script and the real guards over a two-repository fixture: the merge-and-push happy path through `FINISHED`, and the foreign-staged stop.
