# self-review-kan-urgent-5

## Why

Five Jira tickets, planned as self-review findings; four hold against the tree at `ebc10e49`
(`design.md`, **Context**), and KAN-984 is already fixed:

- KAN-963: `protect-main-checkout.py` judges a write's path with `abspath`, which collapses `..`
  before symlinks resolve, so `cat > <link into main/sub>/../f` writes into the main checkout and is
  allowed.
- KAN-972: `check-task-commit-fields` requires a task's declared `Commit:` subject on every listed
  commit of a cross-repo task, so KAN-924 Task 29's frontend and playwright commits had to drop their
  module scope to share one subject.
- KAN-976: the same guard judges one commit per repository, so a gated review's on-top fix commit,
  carrying its task's `Task-Id:` trailer, is judged alone and fails on the declared files and tests
  it never touched.
- KAN-987: `.flow/project.md`'s `## worktree setup` runs `npm ci` and the SPA build on every creating
  run, though a docs-only change never compiles `stats/internal/web`.

## What changes

- `protect-main-checkout.py` judges a path after its symlinks resolve.
- A cross-repo task's commits each may name their own repository's scope, so long as one carries the
  declared subject exactly.
- The task-fields guard joins every commit from the branch's merge base whose `Task-Id:` trailer
  names the task into one verdict.
- `## worktree setup` is gone; `make -s web-dist` builds the SPA lazily, opening the `## test` and
  `## lint` lines that need it.
