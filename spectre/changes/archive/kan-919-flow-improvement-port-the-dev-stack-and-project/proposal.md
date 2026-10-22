# kan-919-flow-improvement-port-the-dev-stack-and-project

## Why

`scripts/check-dev-stack-fresh.sh`, `scripts/check-spec-reach.sh` and
`scripts/check-uitest-overrides.sh` are still bash, and they are the last bash consumers of
`scripts/lib/strip-bom.sh`, `scripts/lib/sanitize-display.sh` and
`scripts/lib/visual-table-cells.awk` (KAN-919, part of the bash → Go guard migration inventory of
2026-10-07). Every other guard has a Go twin of these helpers already (`stripBOM`,
`sanitizeDisplay`, `vtSplitCells`/`vtTrimCell`/`vtFoldCell` in
`stats/internal/guard/visualsection.go`), so the bash copies are a second statement of the same
logic kept alive by three callers.

## What changes

- Each of the three guards runs as Go in `stats/internal/guard/`, registered in `guard.Registry`;
  its `scripts/<name>.sh` becomes a thin `flow_guard_exec` shim with the same arguments, stdout,
  stderr and exit codes.
- Every case of each bash harness (`scripts/test-check-<name>.sh`) is carried into a Go table test,
  and the harness is deleted.
- `scripts/lib/strip-bom.sh`, `scripts/lib/sanitize-display.sh` and
  `scripts/lib/visual-table-cells.awk` are deleted, with their citations; the Go parity tests that
  sourced them run them as they stood at the merge base.
- `scripts/lib/resolve-file.sh` stays — `check-plan-shape.sh` and `plan-dispatch-bundles.sh` still
  source it; its header stops naming `check-dev-stack-fresh.sh`, which never sourced it.
