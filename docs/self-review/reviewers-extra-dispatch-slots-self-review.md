# Self-review report for reviewers-extra-dispatch-slots

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 3 — fine

The pass covers what the bundle holds and nothing beyond it.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** An unquoted `--include=*.go` grep failed under zsh's no-match glob error; no pipeline text prescribes that form, so there is no pipeline change to make — noted
- **[flow-fix]** The bundle skipped 5 of 8 sources, as expected for `/flow-fast`, which writes no spectre artifacts and prints its change summary after self-review — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Most context went to reading `review-panel.md` and the self-review skill in full, though `/flow-fast` cites section ranges; an agent choice, not a pipeline defect — noted
- **[flow-cost]** The primary reviewer read 321k cached tokens for an 81-line diff, normal priming for a decided panel — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Test-first with the red run shown; the drift Minor closed by a pin test mutation-checked against the docs table; the rejected schema-extension alternative recorded with its reason — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Extra-dispatch descriptions live as a Go map rather than stored decision data; the alternative was rejected with a reason and the pin test covers drift — noted

## What can be sped up — `flow-speed`

_none — this angle produced no findings._

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case, not chosen for mining — noted
  - task: 1. Count extra-dispatch slots as reviewers, with their descriptions
  - commit: 3194861e fix(store): show extra-dispatch reviewer slots and their descriptions
  - code: /Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/reviewers-extra-dispatch-slots/aggregate.go (roster_slots UNION over `$.extra_dispatch.slots[*]`)
  - tests: `TestReviewersIncludesExtraDispatchSlots`, zero rows before the fix, passing after
  - why: a silent filter hiding records stored at a second JSON location; fixing it needs the data shape, not just the code
