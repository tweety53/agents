# Self-review report for kan-1037-stats-weekly-costs-dashboard

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** /flow put recommended-option questions to the operator past the plan gate and stopped at milestones (in-run pipeline fix) — fixed: 2f091d8
- **[flow-fix]** `run-verify-list.sh` run lint then test under one key ended the verify row `skipped` when the test group was cache-skipped, and signal seven of `check-unfinished-work` accepted only `completed`; it now accepts `skipped`, and the stale Recording sentences are corrected — fixed: 274850e2
- **[flow-fix]** A reused `verify` key keeps its first startedAt, so a re-run after a new code commit never satisfied signal seven; the doc and the reason line now name a fresh `verify-<k>` key — fixed: fd5d8189
- **[flow-fix]** A pipe to `head` SIGPIPE-killed `run-verify-list` mid-list, leaving the row unended; SIGPIPE is now caught with `signal.Notify` so child commands keep the default — fixed: f8c6e682
- **[flow-fix]** `TestProveReproducerSignals` could hang on a blocking FIFO open once the reader was gone; the release is now non-blocking with a bounded retry, proved by a test that hung against the old open — fixed: a67977aa
- **[flow-fix]** The cleanup guard matched only `-wave-group-<n>` throwaway copies, so a leftover `-mutation-<n>` copy passed run 2; it now matches both — fixed: 236fcbc3
- **[flow-fix]** A mutation slot symlinked the shared `node_modules` into its copy, and `npm ci` there emptied it; the brief forbids the symlink and the Makefile refuses `npm ci` over one — fixed: c059964b
- **[flow-fix]** The commit step's `STOP: narrative` for the missing in-run fix line cost one retry; the guard worked as designed — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** `flow stage` took the change only as a positional while every `flow record` takes `-change`; it now accepts `-change` — fixed: 64cb0130
- **[flow-cost]** About 380k uncached tokens went on recommended-option asks; covered by the in-run fix 2f091d8 — noted
- **[flow-cost]** Panel round 0's two Criticals were compile and test breaks caught by the full suite running alongside pass 1, as designed — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The mutation slot found ten real test gaps in round 0, each killed and recorded — noted
- **[flow-improvement]** Visual verification surfaced two spec gaps (empty weeks dropped, a lone point drawing no line) before landing — noted
- **[flow-improvement]** The live check matched the store's overall cost to a psql sum — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** `stage_runs.stage_key` was a dead column that misled this change's planner; a migration drops it — fixed: f6cc8591

## What can be sped up — `flow-speed`

- **[flow-speed]** The verify hang, the serial verify re-runs and the `npm ci` restore were costs of the fixed flow-fix findings above — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1 (Critical, round 0), not mined (mining ask suspended by the operator) — noted
  - location: `stats/internal/api/stats.go:50`
  - reviewer: "StatsStore gained WeeklyCosts but fakeStore (internal/web/embed_test.go) and stubStageStore (internal/client/client_test.go) did not; go vet/test ./... fail to compile"
  - why: a cross-package interface-implementer miss with a crisp compile oracle
- **[flow-eval]** review case — F6 (Important, round 0), not mined (mining ask suspended by the operator) — noted
  - location: `stats/web/src/components/TimeSeriesPanel.tsx:90`
  - reviewer: "header comment promises a null breaks the line; the code draws across it"
  - why: comment-versus-code contract drift, checkable with a gap fixture
- **[flow-eval]** impl case — Task 1 (weekly Wed-to-Wed cost totals), not mined (mining ask suspended by the operator) — noted
  - code: `stats/internal/store/aggregate.go` `WeeklyCosts`
  - why: a Wednesday-anchored UTC bucket with a precedence-ordered category CASE and boundary-row tests

## Instruction gaps that cost fix rounds — `flow-prompt-tuning`

_none — this angle produced no findings._
