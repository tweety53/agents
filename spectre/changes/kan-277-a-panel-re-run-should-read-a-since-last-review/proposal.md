# kan-277-a-panel-re-run-should-read-a-since-last-review

## Why

`scripts/check-panel-diff-size.sh` gates the review panel's over-cap prompt on the full
branch diff (`merge-base..working-tree`), on every round — pass 1 and every re-run alike.
The delta mechanism already in `skills/flow/review-panel.md` (landed in `72f4f0e`, after
this finding's own run) scopes most Full-re-run slots to their own `since-last-review`
delta, so the full-branch measurement increasingly overstates what any single slot
actually reads as rounds accumulate. On KAN-77 this meant the same over-cap prompt fired
from round 5 onward — 3,793 changed lines measured against an 886-line actual delta by the
last round — while the true per-slot reading burden stayed well under the cap the whole
time. Self-review finding, angle `myflow-cost`, from KAN-77 (`docs/self-review/kan-77-sdd-ledger-canonical-path-self-review.md`), filed as KAN-277.

Targeted re-runs (`fix-round-N.diff`) have no cap check today at all — every re-run slot
reads that one diff, so a single number would gate correctly there too, and adding it
costs nothing extra.

## What changes

- `check-panel-diff-size.sh` gains no interface change — it already measures one point
  against the working tree (or, when clean, equivalently against `HEAD`), which is exactly
  what a scoped delta needs.
- `skills/flow/review-panel.md` changes what it feeds the cap check and what it gates on,
  per re-run mode:
  - **Pass 1** — unchanged: every slot reads the full diff, so the full-branch count is
    already the per-slot max.
  - **Full re-run** — the gating number becomes `max(full-branch count, if Primary or any
    slot with no held last-reviewed sha is dispatched this round; largest scoped-delta
    count among returning slots)`, computed by calling the unmodified script once per
    distinct starting point (merge-base, and each distinct last-reviewed sha in play this
    round). The full-branch count is still reported alongside, for context.
  - **Targeted re-run** — a new cap check, gated on `fix-round-N.diff`'s own size (the one
    diff every re-run slot reads this round).
- `<abs-worktree>/.superpowers/sdd/final-review-panel.md` records both the gating number
  and the full-branch count on every round the check runs, not a single number.
