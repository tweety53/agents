Only the two guards source the shared lib — no other silent caller. Verification complete: both guards' usage messages, tests, header preservation, mutation test, and search for other callers all check out.

**PASS**

- Live-ran both guards with no args in the worktree: stderr now prints the full wrapped `<base-ref>` explanation, exit status still 2.
- Ran both test harnesses: `check-finish-preflight: all cases pass`, `check-base-moved: all cases pass`, including the new `usage message states the base-ref rule` case.
- Mutation test: changed `prefers` -> `prefer` in `check-finish-preflight.sh`, re-ran harness, new case failed with `usage message no longer states the base-ref rule`; restored file, confirmed `git diff --stat` shows no residual change.
- `grep -c "prefers refs/remotes/origin/<base-ref>"` returns exactly 1 per guard — single copy per `usage-message-is-the-only-site`.
- `git diff` against merge base shows only the four hunks in `final-review.diff`; both guards' header comment blocks are untouched, satisfying `both-guards-share-the-defect` and the header-preservation constraint.
- `finish-contract.md`'s `origin/$BASE` composition instruction is untouched (grep confirms wording unchanged) — `contract-guidance-stays` honored.
- Searched for other silent callers: only `check-base-moved.sh` and `check-finish-preflight.sh` source `resolve-remote-base.sh`/`resolve_remote_base`. `prepare-archive-branch.sh` also takes a `<base>` argument but is bare-only by contract and its header already documents the `origin/<base>` fast-forward fully (not silent) — not the same defect, correctly out of scope.
- Both tasks' declared `Files`, `Tests`, `Commit` fields match the two commits (`6e4340b4`, `dc890c44`) exactly; no scope creep, no behavior change beyond the usage-message text.
