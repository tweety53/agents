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

#### Before

| Run | real (s) | user (s) | sys (s) | load (1/5/15) | exit | harnesses |
|-----|---------:|---------:|--------:|---------------|-----:|-----------|
| suite 1 | 118.14 | 198.10 | 277.12 | 6.16 4.98 4.62 | 0 | 84 (84 passed) |
| suite 2 | 110.11 | 196.96 | 282.89 | 22.31 12.11 7.51 | 0 | 84 (84 passed) |
| suite 3 | 110.68 | 196.05 | 280.32 | 23.72 16.20 9.70 | 0 | 84 (84 passed) |
| go 1 | 14.08 | 16.51 | 27.42 | 25.46 20.12 12.12 | 0 | — |
| go 2 | 15.42 | 16.72 | 27.75 | 23.62 19.96 12.20 | 0 | — |
| go 3 | 13.04 | 15.92 | 26.61 | 23.67 20.14 12.40 | 0 | — |

- Median real: suite **110.68s**; Go package **14.08s**.
- Load rose from 6 to ~23 after run 1 from other sessions on the host, not the measurement.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p scripts/run-guard-tests.sh x3 @ 3e48ecac -->
<!-- measured: cd stats && /usr/bin/time -p go test ./internal/guard/... -count=1 x3 @ 3e48ecac -->

Slowest five harnesses (runner's per-harness `(Ns)`):

| Rank | Suite 1 | Suite 2 | Suite 3 |
|-----:|---------|---------|---------|
| 1 | test-check-installed-citations.sh 98s | test-check-installed-citations.sh 109s | test-check-installed-citations.sh 109s |
| 2 | test-check-plan-provenance.sh 92s | test-check-plan-provenance.sh 98s | test-check-plan-provenance.sh 106s |
| 3 | test-check-references.sh 79s | test-check-references.sh 83s | test-check-references.sh 86s |
| 4 | test-setup.sh 62s | test-setup.sh 62s | test-setup.sh 66s |
| 5 | test-check-panel-reproducers.sh 47s | test-check-panel-reproducers.sh 48s | test-compose-mockup-frames.sh 50s |

<!-- measured: grep '(Ns)' of each run's log, sorted descending @ 3e48ecac -->

#### After

| Run | real (s) | user (s) | sys (s) | load (1/5/15) | exit | harnesses |
|-----|---------:|---------:|--------:|---------------|-----:|-----------|
| suite 1 | 53.95 | 114.94 | 161.15 | 2.97 4.13 3.71 | 0 | 79 (79 passed) |
| suite 2 | 52.33 | 117.90 | 165.94 | 12.69 6.95 4.80 | 0 | 79 (79 passed) |
| suite 3 | 52.29 | 119.09 | 167.97 | 14.08 8.49 5.52 | 0 | 79 (79 passed) |
| go 1 | 11.83 | 16.84 | 33.29 | 14.78 9.78 6.20 | 0 | — |
| go 2 | 12.01 | 16.67 | 30.46 | 15.05 9.99 6.31 | 0 | — |
| go 3 | 10.69 | 16.77 | 31.07 | 14.62 10.12 6.42 | 0 | — |

- Median real: suite **52.33s** (Before 110.68s); Go package **11.83s** (Before 14.08s).
- Harness count fell 84 -> 79: the five ported harnesses were git rm'd.
- Go exit: each run printed `ok … stats/internal/guard`; the 0 is read from that line.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p scripts/run-guard-tests.sh x3 @ spectre/kan-778-agents-port-the-next-five-slowest-guards-to-the -->
<!-- measured: cd stats && /usr/bin/time -p go test ./internal/guard/... -count=1 x3 @ spectre/kan-778-agents-port-the-next-five-slowest-guards-to-the -->

Slowest five harnesses (runner's per-harness `(Ns)`):

| Rank | Suite 1 | Suite 2 | Suite 3 |
|-----:|---------|---------|---------|
| 1 | test-setup.sh 39s | test-setup.sh 34s | test-setup.sh 32s |
| 2 | test-compose-mockup-frames.sh 25s | test-compose-mockup-frames.sh 24s | test-compose-mockup-frames.sh 23s |
| 3 | test-check-guard-symlinks.sh 24s | test-check-dispatch-paragraphs.sh 23s | test-generate-relocation-comparison.sh 22s |
| 4 | test-check-dispatch-paragraphs.sh 24s | test-generate-relocation-comparison.sh 22s | test-check-stage-mark-calls.sh 22s |
| 5 | test-mutate-and-verify.sh 23s | test-check-stage-mark-calls.sh 22s | test-check-guard-symlinks.sh 22s |

<!-- measured: grep '(Ns)' of each run's log, sorted descending @ spectre/kan-778-agents-port-the-next-five-slowest-guards-to-the -->

Parity against the floors above:

| Port | `--- PASS` count | floor | result |
|---|---:|---:|---|
| TestCheckPanelReproducers | 66 | 55 | pass |
| TestCheckUnfinishedWork | 122 | 102 | pass |
| TestCheckReferences | 42 | 41 | pass |
| TestCheckPlanProvenance | 1439 raw / 996 leaf | 665 | pass |
| TestCheckInstalledCitations | 61 | 61 | pass |

- Zero `--- FAIL` lines; `go test -v` exit 0.
- TestCheckPlanProvenance leaf count: of the 1439 `--- PASS: TestCheckPlanProvenance/<path>` names, counted those with no other passed name prefixed by `<path>/` (i.e. not a case-group parent) = 996.

<!-- measured: cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '--- PASS: Test<Name>/' per port @ spectre/kan-778-agents-port-the-next-five-slowest-guards-to-the -->

#### Judgement

- Suite median 52.33s < Before 110.68s — met (−53%).
- Go package median 11.83s ≤ 20s — met.
- Every port ≥ its floor — met (installed-citations exactly at 61).
- Every harness green — met (79/79, each run).
- Slowest remaining harness: `test-setup.sh` (median 34s). Next slice, guards only: `check-dispatch-paragraphs`,
  `check-guard-symlinks`, `check-stage-mark-calls` (22–23s each), then `check-panel-fix-single-dispatch` and
  `mutate-and-verify` (21–23s); `setup.sh`, `compose-mockup-frames` and `generate-relocation-comparison`
  are not guards — whether they are in a port's scope is the operator's call.

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

### Deferred panel findings go to KNOWN-BUGS.md, unasked

**ID:** deferred-findings-to-known-bugs
**Status:** active
**Chosen:** a review-panel round close that defers findings appends them to
`<project>/KNOWN-BUGS.md` under `## Deferred review findings` (`skills/flow-contracts/known-bugs.md`
canonical) and asks nothing; no Jira follow-up is filed at that close — the operator's instruction
during this run, which deferred seven Minors (F6–F12).
**Considered:** keeping the Jira filing prompt at round close — the operator declined it and asked
for the prompt to go; a new per-change Markdown file — `KNOWN-BUGS.md` already carries the
recorded-not-repaired entries a next change reads.

## Open questions
