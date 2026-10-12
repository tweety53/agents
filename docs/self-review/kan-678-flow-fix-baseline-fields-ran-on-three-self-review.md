# Self-review report for kan-678-flow-fix-baseline-fields-ran-on-three

**Deferred:** reasoning pass run on Muse Code (Muse Spark) from the in-run bundle

Resume-to-land pass: this run found the change's two commits already on the branch, the worktree clean, and no prior report committed; the flow store was unreachable for the whole run (all stage marks journaled), so ledger/panel/store sources were unavailable. The pass covers the diff, the plan, the targeted harness, and this run's own observations, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the guard-test path assumes writable home caches: test-check-plan-shape.sh failed 6 cases until FLOW_GUARD_CACHE_DIR and GOCACHE were redirected to temp dirs — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** the cache redirection cost three harness re-runs before the suite went green; same root as the flow-fix finding above — declined

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** F12 landed with case plus mutation pairs (cases 35/36) per the companion-task convention, and the suite passes unmodified — declined

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

_none — this angle produced no findings._
