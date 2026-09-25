# kan-760-agents-port-the-flow-guard-scripts-and-their

## Why

- `scripts/run-guard-tests.sh` (86 bash harnesses) takes 120s wall, 191s user, 249s sys on 10
  cores at `0747740` — kernel time above user time: the cost is process spawning and waiting, not
  compute. <!-- measured: /usr/bin/time -p scripts/run-guard-tests.sh @ 0747740, 2026-09-25 -->
- The five slowest harnesses sum to 318s of the suite's 1171s summed harness time, and the slowest
  (`test-check-panel-reproducer-exit-contract.sh`, 102s under load) bounds the wall clock.
  <!-- measured: per-harness lines of the same run @ 0747740 -->
- Standalone, four of the five spend most of their wall time waiting, not computing — 0.2s poll
  loops and real timeout/grace deadlines driven case after case in series.
  <!-- measured: /usr/bin/time -p scripts/test-<name>.sh, one at a time @ 0747740 -->
- Bash has reached its floor (KAN-760: earlier bash-side work already took the suite 152s → ~100s).

## What changes

- A new `flow-guard` binary (`stats/cmd/flow-guard`) runs Go ports of five guards:
  `check-cleanup-complete`, `check-panel-reproducer-exit-contract`, `check-task-commit-fields`,
  `run-reproducer`, `gather-dispatch-context` — each with its CLI contract unchanged: arguments,
  environment overrides, verdict lines, stdout/stderr split, exit codes.
- Each ported `scripts/<name>.sh` keeps its header comment and becomes a shim that execs
  `flow-guard <name>`; the old body's reasoning moves into the guard's Go file, and every
  citation of it is repointed there. Skills, `skills/*/scripts/` symlinks and every `$SCRIPT_DIR/<name>.sh`
  caller are unchanged.
- The five bash harnesses are replaced by in-process Go tests in `stats/internal/guard/`, every
  case ported, run in parallel on shared fixtures; `run-guard-tests.sh` runs that package as one
  harness.
- `make build`, a new `make install-guard`, `make restart` and `setup.sh global` build
  `flow-guard`.
- Before/after suite timings are measured on the same machine and recorded in `design.md`.
- Expected suite wall ≈ 60s, bounded by `test-check-panel-reproducers.sh`; the remaining guards
  are follow-up KAN slices, filed at integrate.
