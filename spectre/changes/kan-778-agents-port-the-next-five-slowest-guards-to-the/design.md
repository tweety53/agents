## Context

- Second slice of KAN-760's guard port, the one its `follow-ups-at-integrate` decision names.
  Base `3e48ecac`.
- `flow-guard` (`stats/cmd/flow-guard`, `stats/internal/guard/`) and the shim library
  `scripts/lib/flow-guard.sh` exist; a guard is added by one Go file registered in
  `guard.Registry` and a shim calling `flow_guard_exec <name> "$@"`.
- Sibling helpers the five guards source, and their Go state at base:
  - `scripts/lib/change-plan.sh`, `scripts/lib/spec-root.sh` (`check-unfinished-work`) — ported
    as `changeplan.go`, `specroot.go`.
  - `scripts/reproducer-metachars.sh` (`check-panel-reproducers`) — ported as `metachars.go`,
    parity-tested.
  - `scripts/lib/coverage.sh` (`check-references`, `check-installed-citations`) — not ported; also
    sourced by `check-guard-symlinks.sh`, `check-self-review-report.sh`,
    `check-stage-mark-calls.sh`, `check-vocabulary.sh`, which stay bash.
- `check-plan-provenance.py` and `check-installed-citations.py` are standalone — no other script
  imports them as a module; five Python guards, `.flow/project.md`, `.flow/project-rationale.md`,
  `scripts/lib/test-git-shim.sh`, `scripts/lib/plan_grammar.py`, `scripts/plan-dispatch-bundles.sh`
  and four `skills/flow-contracts/` files cite them by name.
  <!-- verified: grep -rl 'check-installed-citations.py\|check-plan-provenance.py' outside spectre/changes/archive @ 3e48ecac -->

## Measurements

### Parity floors at base

| Harness | `ok` lines (floor) | real (5 run concurrently) |
|---|---|---|
| `test-check-panel-reproducers.sh` | 55 | 27.2s |
| `test-check-unfinished-work.sh` | 102 | 25.6s |
| `test-check-references.sh` | 41 | 36.8s |
| `test-check-plan-provenance.sh` | 665 | 33.9s |
| `test-check-installed-citations.sh` | 61 | 36.7s |

<!-- measured: FLOW_GUARD_CACHE_DIR=<tmp> /usr/bin/time -p bash scripts/test-<name>.sh, the five concurrently, grep -c '^ok' of each log, all exit 0 @ 3e48ecac; load 7.48 after -->

Each Go test file carries at least its harness's floor in cases (`parity-by-case-count`).

### Suite before/after

Task 1 records **Before**; the last task records **After**. Shape as KAN-760's: x3 each, load
average read before each run.

## Decisions

### Carry KAN-760's port decisions unchanged

**ID:** carry-kan-760-port-decisions
**Status:** active
**Chosen:** KAN-760's `flow-guard-binary-rationale-in-go`, `go-tests-replace-harnesses`,
`parity-by-case-count`, `inject-deadlines-in-process`, `helper-parity-tests`,
`port-base-moves-into-go` and `guard-binary-built-from-checkout` (archived
`spectre/changes/archive/kan-760-agents-port-the-flow-guard-scripts-and-their/design.md`) govern
this slice as written; the base for `port-base-moves-into-go` is `3e48ecac` — the operator's
choice.
**Considered:** re-deciding each per guard — no guard here raises a case KAN-760's decisions do not
cover.

### Scope: all five named guards

**ID:** scope-five-next-guards
**Status:** active
**Chosen:** port `check-panel-reproducers`, `check-unfinished-work`, `check-references`,
`check-plan-provenance`, `check-installed-citations` in one change — the operator's choice.
**Considered:** four now with `check-plan-provenance` (3343-line `.py`, 4813-line harness) as its
own slice — a smaller review, but leaves the second-slowest harness bounding the wall; the three
slowest harnesses only — leaves two known slices for later.

### coverage.sh gets a Go twin; the bash library stays

**ID:** coverage-go-twin
**Status:** active
**Chosen:** `stats/internal/guard/coverage.go` implements `coverage_declare`/`coverage_record`/
`coverage_report`/`coverage_verdict` for the two ported guards; `scripts/lib/coverage.sh` stays for
the four bash guards that source it; a Go test runs `coverage.sh` and `coverage.go` over the same
member/count inputs and fails on any difference in rendered fragment or verdict (per
`helper-parity-tests`).
**Considered:** porting the four other bash guards too — out of this slice's scope; no parity
test — the two copies drift silently, the KAN-73 failure `coverage.sh` exists to prevent.

### The guard test package runs in at most 20s

**ID:** guard-package-under-20s
**Status:** active
**Chosen:** `go test ./internal/guard/... -count=1` runs in ≤20s real on this machine, every case
kept, fixed at the source of the wall time exactly as KAN-760's `guard-package-under-10s` states —
which this supersedes for the grown package — never by dropping a case, `-short`, or a loosened
assertion — the operator's choice.
**Considered:** keeping ≤10s — twice the guards in the same budget; no package budget — the Go
package could become the next ceiling unseen.

### Suite success: after median below before median

**ID:** suite-median-below-before
**Status:** active
**Chosen:** `scripts/run-guard-tests.sh` is timed x3 with `sysctl -n vm.loadavg` before each run,
at base and at the last implementation commit; the change meets its goal when the after median is
below the before median; the slowest remaining harness is named and the next slice identified from
it, filed at integrate per KAN-760's `follow-ups-at-integrate` — the operator's choice.
**Considered:** record only — no criterion to fail; slowest harness < 30s — a threshold nothing
measured yet supports.

## Open questions
