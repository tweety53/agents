# kan-922-flow-improvement-port-the-bash-python-tooling — session narrative

## 2026-10-22 — creating run

Implementation resumed at STARTED with the plan already gated; inline execution, three tasks in plan order.

- Task 2's port was written before its test; the code was moved aside, the test run RED against an unregistered guard, and the code restored — the TDD order was recovered, not skipped.
- Task 3's first manifest pin failed on symlink targets: Go's `cmd.Dir` sets no `PWD`, so the stand-in installer's `pwd` printed the symlink-resolved `/private/var/...` path and `<TREE>` normalization missed it. The port now sets `PWD` to the tree, as the script's `cd` did; recorded as a Correction on task 3.
- The first inline `dispatch begin` for task 2 was refused because the CLI checked the main checkout (the session's cwd), which carries another session's uncommitted `kickoff-worktree` edits; re-run from the worktree it recorded. The main checkout's edits were left untouched.
- The pass-1 full suite failed `TestCheckSelfReviewReport` on the base's own KAN-923 self-review report (a non-canonical none-marker and disposition); fixed as full-suite-fix-1.
- `break-and-prove.sh` ran its test command from its own cwd, so the first mutation proofs failed with `go.mod not found` on both legs; re-run with `go -C <stats>`.
- Round 0 raised two Importants (workspace ignored `Env` for docker; nothing pinned the survivors failure exit) and five Minors; fixed in panel-fix-1 except F7 (withdrawn: per-row MISSING lines are the report's contract). Round 1 re-ran all four roles clean.
