# Self-review report for kan-916-per-worktree-lsp

**Deferred:** reasoning pass run on claude-opus-5-5 from docs/self-review/kan-916-per-worktree-lsp-context.md — the pass covers what the bundle holds and nothing beyond it.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `flow suite record`/`list` followed the worktree's isolated `FLOW_ADDR`, so the verify run's suite runtimes were dropped (no daemon on 127.0.0.1:4373) and no worktree run ever fed the cross-run medians `.flow/project.md` `## test` relies on; suite now resolves `FLOW_RECORDS_ADDR` like the record family — fixed: e9901f64
- **[flow-fix]** the MUTATION PROOF paragraph presented `git checkout --` as the flip's restore, which twice reverted uncommitted fixes in this run; it now cites `break-and-prove.sh`, whose restore is a pre-mutation snapshot — fixed: aee79446
- **[flow-fix]** landing main-checkout drift from a throwaway worktree (`670bd7ad`) left the main checkout's identical uncommitted copies blocking its fast-forward, for the operator to reset by hand — filed: KAN-926

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** 22 of 26 dispatches carry no `agent_id` and no token metrics (`Tokens: not measured`), so the run's cost per role and model cannot be measured — filed: KAN-925

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

_none — this angle produced no findings._

**Rating:** 3/5 — fine
