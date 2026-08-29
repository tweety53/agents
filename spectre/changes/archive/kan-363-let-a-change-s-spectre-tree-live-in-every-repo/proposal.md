# kan-363-let-a-change-s-spectre-tree-live-in-every-repo

KAN-363 · Let a change's spectre tree live in every repo it touches, with the trees referencing
each other.

## Why

A cross-repo change keeps its spectre tree in one repo only. KAN-343 spanned `gymie` and
`gymie-frontend`; the whole tree lived in `gymie`, and the frontend repo carried no record of what
its branch was part of.

Measured cost on that change:

- `check-unfinished-work.sh` reported `OUTSTANDING` for the frontend worktree — `no plan at
  <worktree>/spectre/changes/<name>/tasks.md`, for a plan that is not there and never will be. A
  human had to wave the gate through.
- `check-task-commit-fields.sh` reached no verdict in either repo: no `tasks.md` in the frontend,
  and in `gymie` a glob matching three open root changes, which it refuses to guess between. All
  12 tasks' `Files:`/`Tests:`/`Commit:` fields were checked by hand.
- Cloning `gymie-frontend` alone, nothing said the branch belonged to KAN-343 except the branch
  name and the `Task-Id:` commit trailers.
- The two repos had to be merged backend-first by hand, so `develop` never carried a UI reading
  fields the API did not yet serve. Nothing recorded that ordering.

The plan itself already knew the layout — KAN-343's `tasks.md` carries a per-task
`**Repository:**` field and a preamble naming which tasks live where. The guards and the second
repo did not.

## What changes

- **`link.md`**, a new optional file in a spectre change directory, in both directions: a satellite
  repo's copy names the canonical change, the canonical's copy names its parts and the order they
  land in. Every field has a grammar; none is free text.
- **`spectre link <peer>:<id>`**, which writes both sides in one run, guarded, refusing rather than
  landing on top of an in-flight edit in the peer's working tree.
- **`spectre validate`** resolves links, reports one-sided links, archive skew and merge-order
  mismatches; **`spectre list`** shows a satellite with its canonical's progress.
- **The guards follow the link.** `check-unfinished-work.sh` counts a satellite worktree against
  the canonical plan instead of reporting a missing one; `check-task-commit-fields.sh` resolves the
  task through the link instead of exiting without a verdict. `check-cleanup-complete.sh` needs no
  change — its check is a directory-existence test rather than a scaffold-file test, so it already
  reports a leftover satellite directory correctly; it gains the coverage pinning that down.
- **`/flow` runs `spectre link`** for every worktree beyond the canonical one, and **run 1 lands
  the repos in the recorded merge order**, stopping on the first failure.

Two repositories are touched: `spectre` (tree format, CLI) and `agents` (guards, contracts).
spectre ships first.
