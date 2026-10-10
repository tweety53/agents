# fix-stale-merge-base

## stale-merge-base

### Why

KAN-924's integrate run rebased its worktrees onto the base, then stopped before landing. The rebased merge base is a this-run-only value, so the next integrate run read the pre-rebase merge base from the state file. `check-base-moved.sh` reported the base's gained commits as already carried — `CLEAR`, no sync, no `<rebased-merge-base>` — and the reshape would have taken the stale merge base, folding a commit already on the base (gymie-playwright `ca4c165`) into the change's implementation commit. The operator caught it by hand.

### What changes

- `check-base-moved.sh`'s carried `CLEAR` line names the branch's actual fork point, `git merge-base HEAD <ref>`, as `merge base now <sha>`.
- That `<sha>` is the worktree's `<rebased-merge-base>` for the rest of the run, the reshape included, exactly as a clean sync's `REBASED` sha is.
