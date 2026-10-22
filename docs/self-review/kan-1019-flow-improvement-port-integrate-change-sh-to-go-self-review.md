# Self-review report for kan-1019-flow-improvement-port-integrate-change-sh-to-go

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 3 — fine

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The design's bash-vs-shim live check was deferred and never fully run; no guard tied a design's `## Live check` to `live-verification.md` — fixed: 7287cc2c
- **[flow-fix]** `generate-relocation-comparison` exited 2 when the branch deleted a planned file, losing this plan's comparison — fixed: d310870b
- **[flow-fix]** The three Important parity fixes (F1–F3) had only untracked reproducers, no suite test — fixed: bbb9433f
- **[flow-fix]** The panel fix round never requires a suite test for a fixed Critical/Important finding (blast radius 75) — filed: KAN-1025

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Round-0 reviewer read 8.27M cache tokens over 15m31s, largely verbatim-moved harness text; the relocation comparison fix d310870b now marks it moved — noted
- **[flow-cost]** The sonnet/low test-sanity slot is allowed by the model rules — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The untouched bash harness served as the parity oracle; the port passed 215/215 first try — noted
- **[flow-improvement]** Harness cases were lifted verbatim into Go raw strings, so all 215 labels became subtests — noted
- **[flow-improvement]** The panel caught 4 parity defects the oracle could not see (process cwd, two trimming reads, a hard-coded env list) — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** `check-plan-shape` did not check a plan's Baseline chain; task 4's before=1 under task 3's after=2 surfaced only at commit — fixed: 76359518
- **[flow-automation]** Task 5's file-list widening was caught by the plan's own absence sweep — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Parallel panel slots were all stamped with the round's last finish; each now ends at its own report time, rows 7101/7102 repaired — fixed: 9b4b9a0c
- **[flow-stats-app]** Reviewer rows reading `output not measured` is documented harvester behaviour — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** `narrative_format` built 6 fixtures serially (32.3s of a 67.5s run); now parallel per-fixture cases — fixed: d78ccef8

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case: task 1, port integrate-change.sh to Go behind a flow_guard_exec shim (the operator chose not to mine it) — noted
  - Task: tasks.md task 1
  - Code: `stats/internal/guard/integratechange.go:181`, `:231`; `scripts/integrate-change.sh:42-47`, commit 1de701ea
  - Why: 215 existing assertions grade it pass/fail; known hidden parity traps (cwd, `IFS=` reads, child env)
- **[flow-eval]** review case: F1 Important, the Go process keeps the caller's cwd so cleanup from a worktree is HELD by itself (the operator chose not to mine it) — noted
  - Bad: 87c448a8 — shim exec'd without `cd /`; `icPhys(c.env.Dir, mainArg)`
  - Fixed: f594df8d — `FLOW_GUARD_CALLER_DIR="$PWD"` + `cd /` in the shim; `icPhys(dir, …)`
  - Why: process-state parity bug invisible to the 215-assertion oracle and to in-process tests with `Dir: "/"`
