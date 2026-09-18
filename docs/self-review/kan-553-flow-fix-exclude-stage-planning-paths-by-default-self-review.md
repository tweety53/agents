# kan-553-flow-fix-exclude-stage-planning-paths-by-default — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-553-flow-fix-exclude-stage-planning-paths-by-default-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** panel F3 (deferred): merge commits are counted but never diffed (diff-tree -r without -m prints nothing for merges), so a Task-Id evil merge sweeping planning paths answers CLEAN; close with a -m semantics decision and a harness case — filed: KAN-607
- **[flow-fix]** panel F5 (deferred): the guard's fourth exit-2 branch (unresolvable HEAD) has no harness case; add the no-commits sandbox case — filed: KAN-608
- **[flow-fix]** the flow-high subagent type the harness contract names is not registered in this harness's Agent tool, so every panel dispatch silently ran on general-purpose; align the contract with the harness or register the type, and name the fallback in the pass log — filed: KAN-609
- **[flow-fix]** panel F4 (deferred): the harness asserting the bite is stronger than the plan wording promised is the intended behaviour, nothing to change — declined
- **[flow-fix]** the NUL-split walk bug and the command-substitution NUL strip were both fixed in-run and are pinned by the rewritten walk and its harness — declined
- **[flow-fix]** origin/main moved twice mid-run forcing two rebases and a staleness re-run; inherent concurrency, and the staleness re-run handled it as designed — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the harness's single-commit cases all passed while the run's own four-commit branch exposed the walk bug (every sha after the first carried git's entry separator); every log-walking guard gets a multi-commit harness case with a violation in a later commit — filed: KAN-611

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the round-0 dispatches were recorded on the decision's pair (sonnet/high) where the harness mapping calls for the dispatched glm-5.3-flash, and the store accepted the wrong pair without comment; flowd could validate a dispatch row's recorded model against the harness mapping at write time — filed: KAN-610
- **[flow-stats-app]** all five dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
