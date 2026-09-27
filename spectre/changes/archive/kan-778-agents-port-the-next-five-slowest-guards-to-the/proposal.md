# kan-778-agents-port-the-next-five-slowest-guards-to-the

## Why

- KAN-760 ported five guards to the Go `flow-guard` binary, but its suite criterion was not met:
  the `scripts/run-guard-tests.sh` wall is now bounded by unported harnesses —
  `test-check-installed-citations.sh` (104/118/167s), `test-check-plan-provenance.sh`
  (93/113/175s), `test-check-references.sh` (81/101/165s), `test-check-panel-reproducers.sh`
  (56s). <!-- measured: KAN-760 design.md Measurements, /usr/bin/time -p scripts/run-guard-tests.sh x3 @ f4205935 -->
- KAN-760's `follow-ups-at-integrate` decision named this slice; KAN-776 records that it was never
  filed.

## What changes

- Five more guards run as Go ports inside `flow-guard`: `check-panel-reproducers`,
  `check-unfinished-work`, `check-references`, `check-plan-provenance`,
  `check-installed-citations` — each CLI contract unchanged: arguments, environment overrides,
  verdict lines, stdout/stderr split, exit codes.
- Each `scripts/<name>.sh` becomes a `flow_guard_exec` shim keeping its header;
  `check-plan-provenance.py` and `check-installed-citations.py` are deleted, their reasoning moved
  into the Go files and every citation of them repointed.
- `scripts/lib/coverage.sh`'s record/report/verdict logic gains a Go twin for the two ported
  guards that source it; the bash library stays for the four bash guards still using it.
- The five bash harnesses are replaced by in-process Go tests, every case ported.
- `go test ./internal/guard/...` runs in at most 20s real (KAN-760: 10s, for five guards).
- Before/after suite timings are measured on the same machine and recorded in `design.md`; the
  next slice is named from the after measurement.
