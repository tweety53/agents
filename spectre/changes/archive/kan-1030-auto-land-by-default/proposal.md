# kan-1030-auto-land-by-default

## Why

- Every `/flow` change stops at the `IN_PROGRESS` review handoff and waits for a second, bare `/flow`
  before it lands, even when the operator has nothing to review. That costs one extra operator turn
  per change for no gain.
- The operator still needs to be able to keep the review gate for a particular change.

## What changes

- By default, a creating, resuming or fix run that writes `IN_PROGRESS` skips the review handoff and
  goes straight on into integrate in the same invocation. It lands by the project's
  `## default landing route` and, on merge-and-push, continues through cleanup to `FINISHED`.
- A new boolean on the change's state record, `reviewBeforeLanding`, restores today's stop at the
  `IN_PROGRESS` handoff for that change.
- The operator sets it, or clears it, with a plain-language instruction given at the start of
  planning or at any moment of the run (for example during implementation). The run records it at
  once with a state write, and the latest instruction wins.
- The plan gate and its `/clear` handoff are unchanged, and so is every stop integrate already makes.
