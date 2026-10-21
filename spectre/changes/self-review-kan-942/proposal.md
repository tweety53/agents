# self-review-kan-942

## Why

**KAN-942 (flow-fix)** — the visual verifier attributes every departure on a view the merge base
carries no baseline PNG for to the change under verification. `skills/flow/visual-verify-verifier.md`
(**Every departure is attributed at the merge base**, lines 231-235 at `b76c2cf1`) states: "A view
this change creates has no rendering at the merge base, and neither has one the merge base carries
no baseline PNG for: every departure on either is this change's". KAN-758's visual-verify attempt
4 counted some 39 over-tolerance band pairs as that change's; every one came from geometry already
shipped at the merge base (pin-t3, pin-p3, j1, g2/g7, k872) or from frame artefacts — false
departures that cost extra verify rounds.

## What changes

- A view this change creates stays wholly the change's.
- A view the merge base already draws but carries no baseline PNG for is captured at the merge
  base — the change's capture spec run against a stack started from detached merge-base worktrees
  — before any departure on it is attributed, then the stack, its derived workspace and the
  worktrees are torn down, nothing committed.
- A merge-base capture that cannot run leaves the departure the change's, its line naming
  `merge-base capture failed — <reason>`.
