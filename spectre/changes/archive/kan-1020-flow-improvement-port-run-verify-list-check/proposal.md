# kan-1020-flow-improvement-port-run-verify-list-check

## Why

Three guards still carry their logic in bash — `scripts/run-verify-list.sh`, `scripts/check-narrative-runs.sh`
and `scripts/rebuild-full-app-suite-zip.sh` — each with a bash harness the guard-test runner spends a job on.
Two libraries, `scripts/lib/resolve-file.sh` and `scripts/lib/spec-root.sh`, outlived every script that sourced
them: their Go twins (`resolvefile.go`, `specroot.go`) are the only live copies, yet the bash files and the
comments naming them as live stay in the tree.

## What changes

- Each of the three scripts becomes a thin `flow_guard_exec` shim; its logic runs in `flow-guard`
  (`stats/internal/guard/`), with byte-for-byte stdout, stderr and exit-code parity.
- Each bash harness (`scripts/test-run-verify-list.sh`, `scripts/test-check-narrative-runs.sh`,
  `scripts/test-rebuild-full-app-suite-zip.sh`) becomes a Go test that also runs the bash guard as it stood at
  `8c54d6d3` and compares results.
- `run-verify-list` runs its `test` list on the guard-test runner's job pool (`rgtRun`).
- `scripts/lib/resolve-file.sh` and `scripts/lib/spec-root.sh` are deleted; the one test still sourcing the live
  file reads it at `8c54d6d3`, and comments naming either as live are updated.
