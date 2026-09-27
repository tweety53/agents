# kan-842-agents-port-the-next-ten-slowest-bash-scripts-to

## Why

- KAN-841 left `scripts/run-guard-tests.sh` bounded, past the excluded installer, Python wrappers
  and flow-guard's own harnesses, by bash harnesses whose scripts carry their logic in bash.
  <!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh, per-harness (Ns) @ c5379c0a -->
- KAN-841's After judgement named the remaining candidates; KAN-842 files the next ten.

## What changes

- Ten scripts run as Go ports inside `flow-guard`: `check-workspace-isolation`,
  `check-task-reviewer-single-dispatch`, `recover-guard-incident`, `plan-class`,
  `check-installed-rules`, `resolve-base-branch`, `check-visual-verify-dispatched`,
  `check-contract-budget`, `check-task-commit-planning-paths`, `check-panel-citation-trigger` —
  each CLI contract unchanged: arguments, environment overrides, output lines, stdout/stderr
  split, side effects, exit codes.
- Each `scripts/<name>.sh` becomes a `flow_guard_exec` shim keeping its header; its ten bash
  harnesses are replaced by in-process Go tests, every case ported.
- `lib/panel-touched-paths.sh` and `lib/owned-corpus.sh` gain Go twins with parity tests; the bash
  libraries stay for their remaining bash callers.
- `lib/sha256-hex.sh`, whose last bash caller is `plan-class.sh`, is deleted; `sha256.go` already
  holds its Go form.
- `go test ./internal/guard/...` runs in at most 40s real.
- Before/after suite timings are recorded in `design.md`; the next slice is named from the after
  measurement.
