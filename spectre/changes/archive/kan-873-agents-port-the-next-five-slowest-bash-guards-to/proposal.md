# kan-873-agents-port-the-next-five-slowest-bash-guards-to

## Why

- Every new guard and every extension of one is Go in `flow-guard` (`.flow/project.md`), so a guard
  still in bash must be ported before it can be extended; each port also takes a bash harness off
  `scripts/run-guard-tests.sh`.
- KAN-842 left five bash guards at the top of the remaining harness ranking (design.md
  **Measurements**).
- `scripts/test-flow-addr-declaration.sh` is red on `main` (`9cd35da8`): `flow lesson resolve`
  registers a second `FLOW_RECORDS_ADDR` `-addr` outside the one seam.

## What changes

- Five scripts run as Go ports inside `flow-guard`: `land-self-review-report`,
  `check-self-review-report`, `break-and-prove`, `prepare-workspace`, `check-panel-docs-only` —
  each CLI contract unchanged: arguments, environment overrides, output lines, stdout/stderr split,
  side effects, exit codes.
- Each `scripts/<name>.sh` becomes a `flow_guard_exec` shim keeping its header; its bash harness is
  replaced by an in-process Go test, every case ported.
- `lib/post-mutation-check.sh` and its harness are deleted; its Go twin is pinned by its own tests.
- `flow lesson resolve` registers its address through `registerRecordConnFlags`, refuses `-C`, and
  `.flow/project.md` names it in the record family; `test-flow-addr-declaration.sh` passes.
- Before/after suite timings are recorded in `design.md`.
