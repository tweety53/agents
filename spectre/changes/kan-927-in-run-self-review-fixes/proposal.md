# kan-927-in-run-self-review-fixes

## Why

KAN-829 and KAN-875 were meant to have every non-`big` self-review finding fixed during the run, quickly and cheaply. What shipped:
- Integrate only saves a bundle ("Self-review is always deferred").
- The fixing pass is a manual command on the default branch that nobody runs. About 16 bundles are pending.
- Every fix and review runs on `opus` at `flow-high`.
- In-run pipeline fixes are recorded only in the narrative. `self_review_findings` holds 1 `fixed` row in 92, and no UI shows any of it.

## What changes

- `/flow`'s integrate run 1 and `/flow-fast`'s **5. Verify** run the six-angle self-review pass themselves. Every non-`big` finding is fixed and recorded before the change lands.
- Fixes run cheap: one `flow-medium` fixer (`opus`, or `sonnet` for trivial fixes), an `opus` `flow-medium` first review, `sonnet` `flow-medium` re-reviews, and only the touched files' lint and tests.
- Fixes for an agents-repo change land on the change branch. Other projects' changes get one `<agents repo>` branch landed by its default route.
- Every in-run pipeline fix is also a `self_review_findings` row.
- `stats/web` gains a "Self-review fixes" view: every finding with its angle, outcome and commit link, filterable by outcome, plus per-change fixed/filed/declined counts.
- `/flow-self-review` remains the fallback for bundles saved before this change.
