# self-review-kan-990-watch-on-youtube-link

## Why

The self-review pass over kan-990 found three non-`big` defects, each held against the tree
(`design.md`, **Context**):

- gymie's `## apps` Regression suite row does not include a change that touches a `ui paths` path,
  though `capture` re-baselines that screen's PNGs in `gymie-playwright`. kan-990 got no
  playwright worktree, so `verify` and `capture` ran in the main checkout and the capture was
  denied — a wasted 20.9-minute verifier dispatch.
- `integrate-change.sh` counts the canonical worktree's `link.md` (the `## Merge order` record) as
  implementation, so a canonical worktree with no other implementation path is asked a subject;
  the parent supplied the plan subject and gymie landed two `chore(spectre): plan
  kan-990-watch-on-youtube-link` commits.
- The verifier key suffix `-<worktree basename>` never tells worktrees apart: every worktree is
  `<repo>-worktrees/<name>`. With two UI worktrees the second verifier's idempotent `begin`
  overwrites the first's row and `-diff-base`, and `check-visual-verify-dispatched.sh` reports the
  first worktree VISUAL-VERIFY-MISSING.

## What changes

- gymie's Regression suite row includes any change touching a `## visual verification` `ui paths`
  path — committed on gymie's kan-990 change branch.
- `integrate-change.sh prepare` asks no subject for a canonical worktree whose only implementation
  path is a `link.md`, and `commit` lands it as `chore(spectre): merge order`;
  `finish-contract-run1.md` and `integrate.md` say so.
- The `verify`, `visual-verify` and `tooling-analysis` dispatch keys are suffixed with the
  repository name — the main checkout's basename — instead of the worktree basename.
