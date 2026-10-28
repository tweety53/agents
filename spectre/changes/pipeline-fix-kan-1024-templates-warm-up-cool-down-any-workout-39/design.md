# Design — pipeline-fix-kan-1024-templates-warm-up-cool-down-any-workout-39

## Context

See `proposal.md`.

## Decisions

**ID:** shim-exports-script-dir
**Chosen:** export `FLOW_GUARD_SCRIPT_DIR` in `scripts/close-task.sh`, with the
`: "$SCRIPT_DIR/resolve-visual-screenshots.sh"` sibling line check-guard-symlinks rule 2 reads,
copied from `scripts/rebuild-full-app-suite-zip.sh`.
**Rejected:** close-task deriving the script dir in Go (e.g. from the executable path) — the shims
are the established place this variable is set (three already do); a Go-side derivation is a
second mechanism for one value.

**ID:** rebuild-worktree-stays-canonical
**Chosen:** keep `absCanonical` as the rebuild's `<worktree>` argument. Reason:
`resolve-visual-screenshots.sh` takes `<project root>` and reads its `.flow/project.md`, which
lives only in the canonical worktree; a regression checkout (the playwright repo in kan-1024) has
none, and `TestCloseTaskRebuildsASuiteZipInTheRegressionCheckout` pins exactly that. Passing the
zip's own worktree would make the resolver exit 2. The kan-1024 re-run passing with only the
variable set confirms it.

**ID:** first-review-effort
**Chosen:** flow-low — a four-line shim export and one test, verified failing without it.

**ID:** fixer-groups
**Chosen:** one group, task 1.
