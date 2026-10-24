# self-review-kan-1012-templates-step-1-be-template-module-seeds-search

## Why

The self-review pass of `kan-1012-templates-step-1-be-template-module-seeds-search` found three
fixable defects:

- F1: re-running gymie's whole-stack `devStart` over a worktree whose stack is (partly) up refuses
  itself — `withoutStaleOAuthPinnedEntries` drops the running gateway's recorded entry for being in
  the binding, and `resolveWholeBlock` then reads the worktree's own gateway on 8080 as an unplaced
  holder (`cannotDiscoverRefusal`), contradicting the documented recovery "re-running the same
  command is the recovery".
- F2: the per-task commit recipe in `skills/flow/implement.md` (`git add -- <files> ':(exclude)…'`)
  exited 0 having staged only the one tracked path; 15 new paths stayed untracked and the
  pathspec-scoped `git commit --` failed.
- F3: `write-panel-diff` has no exclusion, so kan-1012's vendored `handoff_*` directories filled the
  panel's `final-review.diff`, tripping the docs-only check and the size cap.

## What changes

- gymie's `withoutStaleOAuthPinnedEntries` keeps an in-binding pinned entry that already equals
  today's deterministic port, so a re-run over its own stack resolves instead of refusing.
- A new `stage-task-files.sh` guard stages a task's files with the excludes ahead of the paths and
  refuses, exit 1, when git leaves any named path unstaged; implement.md's recipe calls it.
- `write-panel-diff.sh` leaves out every path a worktree's `## panel diff exclude` globs match; the
  key is documented in `project-configuration.md`, and gymie declares `handoff_*/**`.
