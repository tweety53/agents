# self-review-backlog-flow-6

## Why

The next five tickets in KAN's "Backlog Flow" status, oldest first, filed by earlier self-review
passes, are planned here as one self-review fix change. Four are already fixed in the tree
(`design.md`, **Context**); one names a gap the tree still has:

- **KAN-702** — kan-573's cross-repo visual-trigger match used `origin/main` as the frontend's
  merge base instead of the worktree's recorded one, pulling the files of three archived changes
  into the match set. The step's `<merge-base>` still says nothing about where a satellite
  worktree's base comes from.

## What changes

- **Visual verification**'s cross-repo match (`skills/flow/verify-and-handoff.md`) binds
  `<merge-base>` to each worktree's own merge base from the run's working notes, never a
  hand-picked ref.
