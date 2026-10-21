# self-review-kan-926

## Why

KAN-926 (flow-fix, found by kan-916's self-review pass): in kan-916's integrate run,
`check-main-checkout-drift.sh` reported `DRIFT-DIRTY` — four uncommitted files in the main
checkout. They were verified and landed on `main` from a throwaway worktree (`670bd7ad`). Git then
refused to fast-forward the main checkout, because its identical uncommitted copies would be
overwritten, and the operator had to reset it by hand.

The defect still reproduces at `b76c2cf1`: run 2's `refresh-main-checkout.sh` refuses any tracked
change with `REFRESH-REFUSED: <path> has tracked changes — nothing touched`, and git's own
`merge --ff-only` refuses an unstaged copy identical to `origin/main` with `Your local changes to
the following files would be overwritten by merge`.

## What changes

- `refresh-main-checkout.sh` (its Go guard) discards a dirty main checkout's tracked changes when
  every changed path is byte-identical to `origin/<base>` — worktree copy, and index copy wherever
  staged, no unmerged entry — by restoring exactly those paths to `HEAD`, prints
  `REFRESH-DISCARDED: <path> <n> tracked paths identical to origin/<base>`, and fast-forwards as
  before. Any path that differs keeps the existing `has tracked changes` refusal with nothing
  touched.
- The script's header contract states the landed-copies rule and the new output line.
