# Self-review report for kan-918-flow-improvement-port-the-git-snapshot-and

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** PLAN FIELDS told the fixer to widen a task's `**Tests:**`/`**Files:**`/`**Baseline:**` on every route, though the on-top Flow-Key commit is outside every task and `check-task-records` rule 2 refuses the widened fields; now scoped to the `fold-fixup.sh` route — fixed: 163b0c80
- **[flow-fix]** The live check ran while the detached lint and test groups were still running, wrote `live-verification.md` into the watched tree and failed `run-guard-tests.sh`'s watched-tree check (272 s re-run); `verify-and-handoff.md` now starts it once both groups have exited — fixed: df48e0b9
- **[flow-fix]** `check-plan-shape` passed a `**Tests:**` name (`TestFoldFixup`) already declared in a base-tree test file outside the task's `**Files:**`, which `check-task-commit-fields` refused at commit; new rule F15, measured at zero false positives over 195 archived plans — fixed: 24878355
- **[flow-fix]** Round-1 slot `panel-1-primary+principles` handed back in 14 s without reading its brief, the only place its REPORT FILE path lived, so no report file was written; each slot's Agent call now carries its REPORT FILE path — fixed: 0035900f
- **[flow-fix]** The pins-audit diff(1) false positive and the regression its first fix introduced were both fixed in-run (e4772ed0, a3066c76); the mutation re-run's `-cite` send-back worked as designed — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Six whole-suite timing runs (~36 min) settled nothing, as design.md predicted; the convention lives only in archived KAN-760-slice plans and no live skill prescribes it, so there is nothing in the pipeline to change — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The mutation slot caught 11 surviving mutants in round 0, a regression in a fix (F1) and a test the batch rewrite left out (F20) — noted
- **[flow-improvement]** The live check compared byte for byte through the symlinked install layout against a `git archive` of the base — noted
- **[flow-improvement]** The two header/body exit-code splits were recorded as decision `header-wins-on-exit-code` rather than chosen silently — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Recording a round's findings took one `flow record finding` per row, hand-looped in shell (21 rows, a zsh `${7:+-task $7}` word-splitting retry); `flow record finding -from <jsonl>` records them from one file — fixed: 57b447e9

## What can be sped up — `flow-speed`

- **[flow-speed]** `scripts/test-integrate-change.sh` alone set the guard suite's wall (313–411 s), its ~40 cases run serially on one `$WORK`; cases now run in parallel subshells with their own `WORK`, ~6½ min to ~1 min — fixed: 8620f8b9

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case — Task 3, port guard-autosquash to Go; operator chose not to mine — noted
  - Task: `/Users/tweety53/Projects/agents-worktrees/kan-918-flow-improvement-port-the-git-snapshot-and/spectre/changes/archive/kan-918-flow-improvement-port-the-git-snapshot-and/tasks.md:505`
  - Code: `/Users/tweety53/Projects/agents-worktrees/kan-918-flow-improvement-port-the-git-snapshot-and/stats/internal/guard/guardautosquash.go:60`
  - Why: exact-parity port, one batched `cat-file --batch-check` with `^{commit}` peeling pinned by a mutation test, a deliberate fail-open decision
- **[flow-eval]** review case — F1 (Important, mutation slot), the pins-audit regex let `exec.Command(gitBin, "diff", …)` pass; operator chose not to mine — noted
  - Location: `scripts/test-git-config-pins.sh:196`; fixed a3066c76035dcfd187673267543e6adf6b377764
  - Files: `/Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/kan-918-flow-improvement-port-the-git-snapshot-and/test-git-config-pins.bad.sh`, `…/test-git-config-pins.fixed.sh`
  - Why: a false-positive fix silently opened a hole in a security-style audit; only a mutation probe caught it
