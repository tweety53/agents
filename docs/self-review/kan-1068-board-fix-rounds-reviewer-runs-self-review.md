# Self-review report for kan-1068-board-fix-rounds-reviewer-runs

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** R7's main fix rounds counted no verify or visual-verify loop — those fixes record under the verifier role as `verify-fix-<N>` / `visual-verify-fix-<N>`, never `panel-fix`; each loop's highest round now adds in, its test seeded with the real role — fixed: 739ff762
- **[flow-fix]** The board's self-review C/I/M counts matched a change name across projects; the lateral now matches `project_key` too — fixed: 739ff762
- **[flow-fix]** The verbatim-moves FAIL reply named the spectre acknowledgement path a `/flow-interactive` run may not write, so R11's brainstorm.md clarification was dropped; the reply now names the interactive change root — fixed: 91973379
- **[flow-fix]** `flow.visual-verify` could be ended `completed` before its check ran (R3), which satisfies the landing gate; the skill now makes the end its own call after the screenshot, and R3's early row is repaired to `superseded` — fixed: a288d963
- **[flow-fix]** A GitHub SSH outage left two requests unpushed for one turn; external, cleared on its own — noted
- **[flow-fix]** R2's and R12's failed visual-verify runs carry no reason; `-reason` landed only in R13 — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** R17's live probe took three dispatches, two refused at dispatch — `flow-medium` carries no Agent tool, and an open experiment pins `explorer` to sonnet — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The live UI check caught three real layout defects before they landed (R2, R12 twice); keep the screenshot-against-the-ask on every `ui paths` request — noted
- **[flow-improvement]** One worktree with a merge and push per request landed every request on main without conflict — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** Explorer cost attribution relied on the planner happening to name `explore-<n>` in its agent description; brainstorm-planner.md now requires it — fixed: a288d963

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

_none — this angle produced no findings._

## What would make a good flow-eval case — `flow-eval`

_none — this angle produced no findings._

## Instruction gaps that cost fix rounds, and the generic prompt update that would have avoided them — `flow-prompt-tuning`

_none — this angle produced no findings._
