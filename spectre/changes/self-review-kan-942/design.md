## Context

A standalone self-review fix loop over one finding, KAN-942, in worktree
`/Users/tweety53/Projects/agents-worktrees/self-review-kan-942` on branch `self-review-kan-942`,
based at `b76c2cf1` (origin/main). The brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-942-design.md` in this worktree.

**Reachable at the base.** The sentence the proposal quotes stands at lines 231-235 of
`skills/flow/visual-verify-verifier.md` at `b76c2cf1`; no commit since 2026-09-01 over that file
touched it, and no archived change records a rejected merge-base capture. The operator asked for no
Go-guard work on this loop, so the fix is the verifier's own step text; no guard pins the
attribution prose today (`scripts/check-verify-report.sh` reads no attribution field).

The project already runs a merge base's code beside a change's own stack for its before figures —
gymie's `.flow/project.md` (**A before figure runs the merge base's backend**) adds a detached
`../<name>-before` worktree, runs the stack from it, and removes the workspace it derived before
removing the worktree (kan-1001). The fix generalises that shape for the capture.

## Decisions

### Capture at the merge base when it has no baseline

**ID:** merge-base-capture
**Status:** active
**Chosen:** a view the merge base already draws but carries no baseline PNG for is captured at the merge base — detached `<name>-merge-base` worktrees beside the change's own in each repository the stack and capture run from, the change's spec, sidecar and fixtures copied in, the stack started there on its own derived workspace, `capture` run, the same element measured and judged as the baseline-PNG case is; then stack, derived workspace and worktrees torn down, nothing committed. One capture per spec serves every departure on its views. A capture that cannot run leaves the departure the change's, named `merge-base capture failed — <reason>`. The finding's own fix direction, and the only route that measures merge-base geometry where no baseline exists.
**Considered:** attributing by whether the diff touched an element's source files — maps rendered elements to files by judgment and measures no geometry; stashing the change and capturing in its own worktree — mutates the worktree and the served stack mid-verify; committing a merge-base baseline for the spec — writes files outside the change's scope into the repository; reporting the missing baseline and leaving attribution to the parent — leaves the false departures standing.

### Worktree name

**ID:** merge-base-worktree-name
**Status:** active
**Chosen:** `<name>-merge-base`, beside the change's own worktree in each repository — a stack whose start pairs sibling worktrees by name (gymie's `resolveSiblingCheckout`) pairs these, and a workspace id derived from the name never collides with the change's.
**Considered:** a `mktemp -d` path — a sibling-pairing stack would not find its partners and fall back to the main checkouts; gymie's `<name>-before` — collides with a before-figure worktree a verify of the same change may hold.

### Prose only, no guard

**ID:** prose-only
**Status:** active
**Chosen:** the fix is the verifier's step text alone, its new sentences acknowledged in `verbatim-moves.txt`.
**Considered:** a guard that refuses a report attributing a departure on a baseline-less existing view to the change — the operator asked for no Go-guard work on this loop, and new guard logic lands in Go only.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one prose paragraph in one file, applied from a verified diff, with no code or guard to break.
**Considered:** `flow-medium` — buys depth on code paths this change does not touch.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-942-fix`, Task 1.
**Considered:** none other — the plan has one task.

## Open questions
