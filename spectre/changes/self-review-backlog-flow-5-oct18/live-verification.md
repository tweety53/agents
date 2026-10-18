# Live verification

Workspace `flowd` on `flow_self_review_74e2` (port 7403), `FLOW_STATE_DIR` a scratch copy holding one journal entry (`live-check-change.journal.stage`, 472 bytes, written by `flow stage begin` against an unreachable store).

- **Before** (merge base `52438f0d`, built from a detached worktree): `flowd journal replay trigger=startup journals=1 applied=1 refused=0`; the entry left the copy (0 bytes).
- **After** (this branch): `flowd journal replay skipped: not the dev workspace's flow database`; the entry stayed (472 bytes).
- Result: matched. The protected dev `flowd` was not restarted; its startup replay line is checked once the landing restarts it.
