# Self-review report for kan-1033-plan-mode-planning

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The parent ended its turn after the panel to answer two mid-run operator messages; `pipeline.md` now says a mid-run operator message is never a stop — fixed: ba68637c
- **[flow-fix]** An in-run pipeline fix committed straight on `spectre/<name>` with no loop rows was refused by nothing; `check-untasked-commits` now lists a `Flow-Key: pipeline-fix-` commit on a change branch — fixed: 4246416e
- **[flow-fix]** `SKILL-rationale.md` said the Agent tool has no dispatch-time `effort`, contradicting the planning texts and the tool's actual parameters; the rationale clause is corrected — fixed: ace3ff62
- **[flow-fix]** `check-guard-symlinks`' placeholder regex missed a literal-prefixed placeholder (`fp-<literal-token>`), dropping that citation from the required set — fixed: 25c85d21
- **[flow-fix]** The agent-baseline hook's test failed a deny case open under the full `-race` suite (5 s subprocess bound starved); the tests inject `FLOW_HOOK_UNBOUNDED=1`, proved by a 6 s stub — fixed: 5249db0c
- **[flow-fix]** Task 4's verify missed guards its own files needed (citations, symlinks, the setup count); an existing planner rule already requires them, a planner slip — noted
- **[flow-fix]** design.md claimed `loop.go` needed no change; recorded as a Correction as the contract asks — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Round 4 existed only to confirm the flake fix's measured rate, which the determinism rule asks up front; small cost — noted
- **[flow-cost]** Panel round 0 (9–12 min per bundle) is within the norm for a ~1000-line diff — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The full-suite fix ran concurrently with panel round 0, costing no wall-clock — noted
- **[flow-improvement]** The mutation slot caught unpinned PLANNING DELEGATION phrases, and its re-run caught a first fix that iterated the table it pinned — noted
- **[flow-improvement]** The flake fix was proved by breaking it, with before/after run counts — noted
- **[flow-improvement]** Every plan deviation landed as a dated Correction — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** The inline verify row was hand-composed and back-filled wrong twice; `run-verify-list.sh` now records it, and rows 7647/7650/7651 were retimed — fixed: 40c74b57
- **[flow-automation]** A bare `cat > file` with no stdin hung a Bash call for 120 s; the Bash hook now denies it — fixed: ab7bcc77

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Re-run rounds' pass rows were hand-recorded and missed until the unfinished-work guard flagged them; `panel-round-begins.sh` now records the round's pass row — fixed: 3f2f652e

## What can be sped up — `flow-speed`

- **[flow-speed]** 120 s lost to the hung Bash call, removed by ab7bcc77 — noted
- **[flow-speed]** Six full `-race` suite runs to measure the flake are what the determinism rule requires — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F5, PLANNING DELEGATION phrases unpinned; not offered for mining — noted
  - severity: Important; location: `stats/internal/guard/dispatchparagraphs.go:83`
  - reviewer: "removing \"plan-agent-pair.sh\" from the list passes TestCheckDispatchParagraphs"
  - fixed in: 4effe8c3
  - why: a mutation-only gap whose first fix iterated the table it pinned
- **[flow-eval]** review case — F10, unreadable design.md branch untested; not offered for mining — noted
  - severity: Important; location: `stats/internal/guard/planshape.go:406`
  - fixed in: 3a0fe02b
  - why: crisp error-path coverage case with a reproducer
- **[flow-eval]** impl case — Task 3, plan-mode design sections in `check-plan-shape`; not offered for mining — noted
  - code: `/Users/tweety53/Projects/agents-worktrees/kan-1033-plan-mode-planning/stats/internal/guard/planshape.go:380`
  - why: header-scoped rule with three edge cases the panel pinned
