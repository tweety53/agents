# kan-921-flow-improvement-port-the-guard-test-runner-run

## Context

`scripts/run-guard-tests.sh` at `2a3a0297` is the runner's specification: its header states the
contract (glob discovery, companion presence, the tree gate, `RUN_GUARD_TESTS_ROOT` and
`GUARD_TESTS_REPO_ROOT`, the output shape, the flow-guard cache and telemetry overrides) and its
body is the behaviour, including the longest-first ordering read from
`<git-common-dir>/guard-test-durations`. `scripts/lib/parallel.sh` supplies the bounded spawn,
merged stdout+stderr capture, list-order replay (`---- <name> ----` headers, each block followed by
a blank line), the `JOBS` refusal and the aggregate exit code.

Findings that shape the plan:

- `lib/parallel.sh` has one caller left, the runner (KAN-778 ported the other).
- `lib/coverage.sh` has **no** bash caller left: every guard that sourced it is Go and uses
  `coverage.go`. Its only consumers are `scripts/test-lib-coverage.sh` and `TestCoverageParity`
  (`stats/internal/guard/coverage_test.go`), which diffs `coverage.go` against it; two Go fixtures
  (`check_guard_symlinks_test.go`, `check_stage_mark_calls_test.go`) still copy it into their
  fixture trees.
- Live prose naming the deleted files: `.flow/project.md`, the header comments of
  `scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh` and
  `scripts/check-self-review-report.sh`, Go comments in `coverage.go`, `installedcitations.go`,
  `references.go` and `check_installed_citations_test.go`, and
  `skills/flow-self-review/SKILL.md`.

The runner, its library and the coverage library are one change because deleting `coverage.sh`
and `parallel.sh` is only safe once their last consumers are Go, and the sweep of their mentions
is one absence check over all five deleted paths.

## Decisions

### Where the runner lives

**ID:** runner-in-flow-guard
**Status:** active
**Chosen:** `Registry["run-guard-tests"]` in `stats/internal/guard/runguardtests.go`, behind the
standard shim (`scripts/check-references.sh`'s shape) — the shim and build cache already exist and
the project rule puts guard logic in `flow-guard`.
**Considered:** a separate `stats/cmd/run-guard-tests` binary — rejected, a second build path and
cache for one caller.

### How the jobs run

**ID:** goroutine-pool
**Status:** active
**Chosen:** one goroutine per harness gated by a semaphore of `JOBS` (else `runtime.NumCPU()`),
each running `bash <harness>` with stdout and stderr written to one buffer; results indexed by
list position, so replay walks list order by construction. A `JOBS` set but not a positive integer
prints `parallel_job_count: JOBS must be a positive integer, got '<v>'` (`got 0` for zero) and
`run-guard-tests: could not run the harnesses (see above)` on stderr, exit 2, as today.
**Considered:** keeping `xargs -P` behind Go — rejected, it keeps the bash library the ticket
deletes.

### Per-harness wall time

**ID:** elapsed-unix-seconds
**Status:** active
**Chosen:** `end.Unix() - start.Unix()` — what `$(date +%s)` arithmetic computed.
**Considered:** `int(time.Since(start).Seconds())` — rejected, it differs from the bash figure
whenever a run straddles a second boundary.

### Which parallel.sh cases are ported

**ID:** drop-bash-mechanism-cases
**Status:** active
**Chosen:** port `test-lib-parallel.sh` cases 1–6 (clean run quiet, one failure replayed alone,
list order under shuffled completion, interleaved output kept contiguous, `JOBS=1` ≡ `JOBS=n`,
labelled replay headers, `JOBS` refusal) as runner-level Go cases; cases 7–9 (capture dir under
`mktemp -d`, trap cleanup on SIGINT, `set -e` neutralisation) are not ported, save one Go case
proving SIGINT removes the runner's temp dir.
**Considered:** porting 7–9 as Go equivalents — rejected, they would test `os/exec` and bash
traps that no longer exist, not this code.

### Coverage's expected outputs

**ID:** coverage-golden-strings
**Status:** active
**Chosen:** `TestCoverageParity` becomes `TestCoverage`: the same table, each case's expected
output the literal string `coverage.sh` printed at `2a3a0297` (captured before the delete), plus
every `test-lib-coverage.sh` case the table lacks. `coverage.sh`'s header reasoning (KAN-73, the
three things it owns, reject-never-coerce) moves into `coverage.go`'s doc comment.
**Considered:** keeping `coverage.sh` as a test-only oracle — rejected, the ticket deletes it and
an oracle nothing ships is dead weight.

## Open questions

## Live check

Run `scripts/run-guard-tests.sh` on the real tree before the change (bash, at `2a3a0297`) and
after (Go): record each run's summary line and exit code. After must report the before harness
count minus the three deleted harnesses, `0 failed`, exit 0, in the same summary shape. Not
working looks like a shim build error, `no test-*.sh harnesses found` / exit 2, a count unchanged
from before (the deletes did not land), or a summary line in a different shape.
