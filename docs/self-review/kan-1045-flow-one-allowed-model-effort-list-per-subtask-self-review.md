# Self-review report for kan-1045-flow-one-allowed-model-effort-list-per-subtask

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `.flow/project.md` named a ported guard's Go companion test `<name>_test.go`, not the `-`→`_` form `run-guard-tests.sh` wants; it cost panel F1 (Critical) — fixed: 7ae38348
- **[flow-fix]** `flow tasks files` could only add, so narrowing a `**Files:**` field forced the hand edit `implement.md` forbids; it gains `remove` — fixed: a5ea828f
- **[flow-fix]** The findings merger was skipped in both panel rounds and nothing refused it; `flow record dispatch begin` now refuses a panel fix or a live re-review before its round's merger — fixed: 596e591c
- **[flow-fix]** `implement.md` and `pipeline-fix.sh` still cited the deleted self-review pairs table; they cite **Allowed pairs** — fixed: f0996ca6
- **[flow-fix]** flow-fast sent dispatches to the closed list for the `description` only, not the `Role:` line the hook now requires — fixed: a96ea925
- **[flow-fix]** The fix review's Minors: the merger gate's sentence now names a live reviewer row, and flow-fast's `micro` panel names the merger — fixed: 83289e90
- **[flow-fix]** The `panel-fix-0` begin and a dirty-tree begin were refused; both guards held as designed and the back-fill route was used — noted
- **[flow-fix]** The plan underestimated callers (`pipelinefix.go`, `plan-agent-pair.sh`'s root, installed citations); each was recorded as a Correction paragraph — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Every repotest hook run did its own `go build` of flow-guard under a fresh HOME (6.2 s vs 0.76 s for the two new tests); `runHook` now shares one guard cache — fixed: 0288aba9
- **[flow-cost]** Inline panel-fix-1 took 22m24s and task 5 read 8.0M cache tokens; no pipeline change cuts them without quality loss — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The citation guard's false positive was fixed at the guard with an anchored exclusion and a reporting test, not allow markers — noted
- **[flow-improvement]** All eight round-0 findings closed in one fix round, re-reviewed clean on sonnet/low, with two mutation proofs — noted
- **[flow-improvement]** Three explorers ran in parallel (54s–1m9s) — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** `flow tasks files remove` replaces the hand `**Files:**` narrowing — fixed: a5ea828f

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** The daemon keeps its own `ValidModels` check apart from the table, by recorded decision `daemon-keeps-valid-models`; a warm `check-dispatch-pair.sh` call is 0.03–0.05 s, so no daemon endpoint is needed — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** The repotest hook tests' wall-clock, by the shared guard cache — fixed: 0288aba9
- **[flow-speed]** The mutation slot starts after the test-sanity slot ends, by design (same CPUs) — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — panel F1, the guard companion test named `modelpolicy_test.go` instead of `check_dispatch_pair_test.go`; not mined, flow-eval mining paused by the operator — noted
  - Bad code: `stats/internal/guard/modelpolicy_test.go` at `eca38159`
  - Fixed code: `stats/internal/guard/check_dispatch_pair_test.go` at `96ef6eb0`
  - Why: a repo convention only the whole guard suite exposes; tests whether a reviewer knows the companion rule rather than reading the diff alone
- **[flow-eval]** impl case — task 1, the **Allowed pairs** table, its parser and the `check-dispatch-pair` guard; not mined, flow-eval mining paused by the operator — noted
  - Code: `stats/internal/guard/modelpolicy.go:100` `policyAllows`
  - Why: a self-contained parser and guard with an exact exit-code contract and fixture-testable marker semantics

## Generic implementer/reviewer prompt updates — `flow-prompt-tuning`

_none — this angle produced no findings._
