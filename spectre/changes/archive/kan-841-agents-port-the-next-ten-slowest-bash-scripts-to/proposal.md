# kan-841-agents-port-the-next-ten-slowest-bash-scripts-to

## Why

- KAN-778 cut the `scripts/run-guard-tests.sh` median from 110.68s to 52.33s; the wall is now
  bounded by ten bash harnesses at 14–25s each, whose scripts carry their logic in bash.
  <!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh, per-harness (Ns) @ d71a2327 -->
- KAN-778's judgement named this slice; KAN-841 files it.

## What changes

- Ten scripts run as Go ports inside `flow-guard`: `check-stage-mark-calls`,
  `check-guard-symlinks`, `check-dispatch-paragraphs`, `mutate-and-verify`,
  `prepare-archive-branch`, `check-base-moved`, `check-panel-fix-single-dispatch`,
  `check-model-keys`, `prove-reproducer`, `check-finish-preflight` — each CLI contract unchanged:
  arguments, environment overrides, output lines, stdout/stderr split, exit codes.
- Each `scripts/<name>.sh` becomes a `flow_guard_exec` shim keeping its header; its ten bash
  harnesses are replaced by in-process Go tests, every case ported.
- `lib/resolve-file.sh`, `lib/project-section.sh`, `lib/post-mutation-check.sh` gain Go twins
  with parity tests; the bash libraries stay for their remaining bash callers.
- `lib/resolve-remote-base.sh` and `lib/reproducer-path.sh`, used only by ported scripts, move
  into Go and are deleted.
- `go test ./internal/guard/...` runs in at most 30s real.
- Before/after suite timings are recorded in `design.md`; the next slice is named from the after
  measurement.
