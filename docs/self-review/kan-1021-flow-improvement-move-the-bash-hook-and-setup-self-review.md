# Self-review report for kan-1021-flow-improvement-move-the-bash-hook-and-setup

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them. The first pass's report was lost after its findings were planned; the pass was re-run over the saved bundle.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `project-get.sh` prints the citation check inside its fence; an `eval` of the body ran a bare `bash` and hung — fixed: e80c85a6
- **[flow-fix]** flow-* agents carry no Skill tool yet were told to invoke the superpowers skills; the paragraphs now name the installed SKILL.md files, guard-pinned — fixed: 2212f3f2
- **[flow-fix]** integrate run 1's reshape base ignored self-review loop commits, folding the self-review plan into the change's plan commit; this branch's folded commit was split back — fixed: 6241891c
- **[flow-fix]** Absence sweeps covering a later task's files misfired on tasks 5 and 6; plan-shape guard F18 now refuses them — fixed: 22d9f1ed
- **[flow-fix]** One land attempt of loop 27 hit `go test`'s 10-minute timeout in the guard package; the suite passed alone in 3.5 minutes and on the immediate retry — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** An off-by-one reproducer citation cost a slot round-trip; the parent now corrects a verbatim-elsewhere line in place — fixed: 2246560e
- **[flow-cost]** F4's re-run verdict outside the fix hunk closed with one `-cite` send-back (31s), as designed — noted
- **[flow-cost]** G1's and the planner's cache reads match their scope; no reducible step — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The panel's primary slot caught cross-project regression F1 with a reproducer and five mutation proofs — noted
- **[flow-improvement]** Four groups ran in two waves of throwaway worktrees and cherry-picked in plan order with no conflicts — noted
- **[flow-improvement]** One Go subtest per harness case made "coverage did not drop" mechanically checkable — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Subagent rows reading `output not measured` is documented harvester behaviour — noted
- **[flow-stats-app]** The lost session's planner dispatch row was closed once by hand; no recurring derivation to move — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** Wave 2 waited on all of G1 though it needed only Task 1's helper; a grouping tradeoff, no defect — noted
- **[flow-speed]** Verify's full-suite re-run was required, since the fix round changed code — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case: F1, `pfChecks` dropped the replaced code's `exists` gate (the operator chose not to mine it) — noted
  - Severity: Important, primary slot
  - Location: `stats/internal/guard/pipelinefix.go:357`, `:379`; bad at fe86e1d9, fixed by fd52c163
  - Why: a port that silently drops an existence guard, visible only against the old behaviour and a foreign-project caller
- **[flow-eval]** impl case: Task 10, rewire pipeline-fix check selection to the Go packages (the operator chose not to mine it) — noted
  - Task: tasks.md task 10
  - Code: `stats/internal/guard/pipelinefix.go:315`, `:381`
  - Why: a small task whose naive implementation falls into the F1 trap
