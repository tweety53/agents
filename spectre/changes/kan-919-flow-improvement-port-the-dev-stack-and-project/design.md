## Context

The guard migration ports a bash guard by writing its logic in `stats/internal/guard/<name>.go`,
registering it in `guard.Registry`, reducing `scripts/<name>.sh` to a `flow_guard_exec` shim, and
carrying every bash harness case into a Go table test before deleting the harness — c20a7ab2
(mutation-reproducer-pin, normative-inventory, python-suppressions) is the latest instance and the
shape every task here follows. A deleted library a Go parity test still pins is read from history
with `bashLibAt` (`stats/internal/guard/libtwins_test.go`), as KAN-887 did for `change-plan.sh`.

Callers are unchanged: `scripts/run-verify-list.sh`, the `skills/flow/*.md` prose, the
`skills/flow/scripts/` symlinks, `stats/Makefile`'s `check-uitest-overrides` target and
`.flow/project.md`'s `## lint` invoke the shims by the same paths and arguments.

## Decisions

### One task per guard, then one deletion task

**ID:** task-per-guard
**Status:** active
**Chosen:** Tasks 1–3 port one guard each, shim and harness deletion included; Task 4 deletes the libraries once all three are ported — each port is reviewable and revertable alone, and the deletion cannot land while a bash caller remains.
**Considered:** one commit porting all three, as c20a7ab2 did — rejected: the three guards share no code beyond the Go helpers that already exist, so a combined commit buys nothing and makes a parity regression harder to bisect.

### Delete visual-table-cells.awk too

**ID:** delete-visual-table-cells
**Status:** active
**Chosen:** delete `scripts/lib/visual-table-cells.awk` with the two shell libraries — `check-spec-reach.sh` and `check-dev-stack-fresh.sh` are its only callers, and its Go twin already lives in `visualsection.go`.
**Considered:** keep it as the ticket does not name it — rejected: once both callers are Go it is dead code whose only reader is its own parity test.

### Keep resolve-file.sh

**ID:** keep-resolve-file
**Status:** active
**Chosen:** keep `scripts/lib/resolve-file.sh`; only its `Sourced by` header drops `check-dev-stack-fresh.sh`, which cites but never sources it.
**Considered:** delete it — rejected: `check-plan-shape.sh` and `plan-dispatch-bundles.sh` still source it, and the ticket assigns its deletion to whichever of this and the Python-backed-guards ticket lands last.

### Parity tests read the deleted libraries at the merge base

**ID:** parity-at-merge-base
**Status:** active
**Chosen:** `TestStripBOMParity`, `TestSanitizeDisplayParity` and `TestVisualTableCellsParity` read their bash through `historicalFile` at `b1b0d601`, this change's merge base, where the files last stood — the parity-at-historical-ref pattern `bashLibAt` already implements.
**Considered:** delete the parity tests with the libraries — rejected: they are the only record that the Go twins match the bash every ported caller relied on; reusing `libTwinBase` (`ef6f87cd`) — rejected: `strip-bom.sh` changed after it (a9bb6d12).

### Harness cases carried by label, the port's own exec of external tools

**ID:** harness-cases-by-label
**Status:** active
**Chosen:** each Go test carries every case of its bash harness as it stands at `b1b0d601`, one subtest per harness case label, asserting the harness's exit code and output patterns; `check-spec-reach`'s Go port runs `node` and `npm` the way the bash did, and its test puts the harness's fake binaries on `PATH` through the guard's `Env`.
**Considered:** live byte-for-byte diffing of the bash against the Go on every case — rejected: the bash is deleted in the same commit, and the c20a7ab2 precedent carries harness cases rather than running both.

## Open questions

## Live check

Run each shim against this repository before Task 1 and after Task 4 and compare stdout, stderr
and exit code:

- `scripts/check-spec-reach.sh .` — before: `Spec reach: not configured`, exit 0.
- `scripts/check-uitest-overrides.sh` (from `stats/`, via `make -s check-uitest-overrides`) — before: `UITEST-OVERRIDES-OK: every UITEST_* variable is override`, exit 0.
- `scripts/check-dev-stack-fresh.sh .` — before, with no stack serving 4174: `check-dev-stack-fresh: fingerprint: cd stats/web && npm run build && curl -sf http://127.0.0.1:4174/ | cmp -s - ../internal/web/dist/index.html`, the build's own output, then `stack-stale: …`, exit 1.

Not working looks like any difference in those lines or exit codes, a `flow-guard: unknown guard`
line, or exit 2 from a shim that could not load `lib/flow-guard.sh`.
