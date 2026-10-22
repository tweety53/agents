# kan-921-flow-improvement-port-the-guard-test-runner-run

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Ports the guard test runner and its parallel library into `flow-guard` (task 1), retires the
bash coverage library behind golden Go cases (task 2), and sweeps every live mention of the five
deleted files (task 3). `design.md` is canonical for every decision; each task cites its entry by
ID.

**The scripts at `2a3a0297` are the specification.** `scripts/run-guard-tests.sh`'s header is the
contract and its body the behaviour; `scripts/lib/parallel.sh` supplies capture, replay and the
`JOBS` rules. The port reproduces every environment override, every stdout and stderr line byte
for byte, the stdout/stderr split, the duration file and every exit code. Where a script and its
header disagree, stop and report — never pick one silently.

**Shim template** — the header comment block kept, corrected only where the port makes a
statement false, then:

```bash verified:copied from the tail of scripts/check-references.sh @ 2a3a0297
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "run-guard-tests: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec run-guard-tests 2 "run-guard-tests:" "$@"
```

**Test isolation:** every Go test calls `t.Parallel()`; fixture trees and fixture git
repositories are built per case in `t.TempDir()`; overrides pass on `Env`, never through the
process environment; no test touches the operator's real `$HOME` or this repository's tree.

## Review Focus

- Replay order is list order, never completion order, under any `JOBS`.
- The tree gate compares porcelain status lines, names only lines present after and absent
  before, and prints the removed-entries fallback for a pure-deletion delta.
- A watched root that is not a work tree — a bare repository included — skips the gate with its
  one stderr line rather than failing.
- The duration file sorts longest first, a harness with no recorded time first of all, and is
  rewritten after every run.

---

- [x] 1. Port the guard test runner and its parallel library to Go

**Files:** `stats/internal/guard/runguardtests.go`, `stats/internal/guard/run_guard_tests_test.go`, `scripts/run-guard-tests.sh`, `scripts/lib/parallel.sh`, `scripts/test-lib-parallel.sh`, `scripts/test-run-guard-tests.sh`
**Tests:** `TestRunGuardTests`
**Regression:** fails if any of `test-run-guard-tests.sh`'s 24 `ok` behaviours or
`test-lib-parallel.sh`'s ported cases 1–6 regress — discovery, ok/FAIL lines and counts, exit
codes, set-but-empty overrides, replay-failures-only, companion presence, the tree gate, list-order
replay, contiguous capture, `JOBS` refusal — or SIGINT stops removing the runner's temp dir.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/run_guard_tests_test.go 2>/dev/null | grep -cE '^func Test' @ 2a3a0297 -->
**After:** none
**Commit:** `feat(guard): port run-guard-tests and lib/parallel to Go`
**Build:** green

**Decision:** runner-in-flow-guard
**Decision:** goroutine-pool
**Decision:** elapsed-unix-seconds
**Decision:** drop-bash-mechanism-cases

  - [x] **Step 1: Failing test.** In `run_guard_tests_test.go`, port every case of
    `scripts/test-run-guard-tests.sh` (cases 1–7 and their mutation proofs, one subtest per `ok`
    label) and cases 1–6 of `scripts/test-lib-parallel.sh` (with 5b), plus one subtest sending
    SIGINT mid-run and asserting the temp dir is gone. Run
    `cd stats && go test ./internal/guard/ -run '^TestRunGuardTests$' -count=1` — expect failure
    (unregistered).
  - [x] **Step 2: Port**, registering `run-guard-tests` from `runguardtests.go`'s `init()`: test
    root from `RUN_GUARD_TESTS_ROOT` else `<FLOW_GUARD_REPO_ROOT>/scripts`; watched root from
    `GUARD_TESTS_REPO_ROOT` else `git rev-parse --show-toplevel` of the repo root; every check,
    message and exit code in the script's order. The script's and `parallel.sh`'s still-true
    reasoning comments move beside the code they explain; bash-only mechanics (bash 3.2, traps,
    `__run_one`, `TIME_DIR`) are not carried.
  - [x] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestRunGuardTests$'
    -count=1 -race -v | grep -c -- '--- PASS: TestRunGuardTests/'` — at least 31, and the run
    exits 0.
  - [x] **Step 4: Shim and delete.** Replace `scripts/run-guard-tests.sh`'s body with the shim
    template; `git rm scripts/lib/parallel.sh scripts/test-lib-parallel.sh
    scripts/test-run-guard-tests.sh`.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/run-guard-tests.sh` on this worktree prints the summary line in the bash's shape and
    exits as the bash did at `2a3a0297` (run that ref's runner from a `git worktree`-free copy:
    `git show 2a3a0297:scripts/run-guard-tests.sh` and `…:scripts/lib/parallel.sh` into a temp
    `scripts/` + `scripts/lib/`, `RUN_GUARD_TESTS_ROOT` at a fixture), and the same fixture through
    the new shim prints byte-identical stdout and stderr, timings aside.

- [x] 2. Retire lib/coverage.sh behind golden Go cases

**Files:** `stats/internal/guard/coverage.go`, `stats/internal/guard/coverage_test.go`, `scripts/lib/coverage.sh`, `scripts/test-lib-coverage.sh`
**Tests:** `TestCoverage`
**Regression:** fails if `coverage.go`'s rendered fragment, verdict lines, verdict code or a
rejected call's message drifts from what `coverage.sh` printed at `2a3a0297` for any of the
parity table's 15 cases or the `test-lib-coverage.sh` cases it adds.
**Baseline:** before=0 after=1
<!-- measured: grep -cE '^func TestCoverage\(' stats/internal/guard/coverage_test.go @ 2a3a0297 -->
**After:** none
**Commit:** `refactor(guard): retire lib/coverage.sh behind golden Go cases`
**Build:** green

**Decision:** coverage-golden-strings

  - [x] **Step 1: Capture.** With `coverage.sh` still present, have `TestCoverageParity` print
    each case's bash output (a temporary `t.Log`), run it, and keep the strings.
  - [x] **Step 2: Golden table.** Rename to `TestCoverage`; each case carries its captured string
    as `want`, compared against the Go output; drop the `exec.Command` and `shq`. Add the
    `test-lib-coverage.sh` cases the table lacks (a member with a space, a member with a leading
    `-`, the empty corpus, a declared-zero member now non-zero), their `want` captured from
    `coverage.sh` the same way.
  - [x] **Step 3: Move the reasoning and delete.** `coverage.sh`'s header reasoning (KAN-73, the
    three things it owns, reject-never-coerce, declaration as an explicit call) moves into
    `coverage.go`'s doc comment, which now says it is the one definition;
    `check_guard_symlinks_test.go` and `check_stage_mark_calls_test.go` drop `lib/coverage.sh`
    from their fixture-copy lists; `git rm scripts/lib/coverage.sh scripts/test-lib-coverage.sh`.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
    ./internal/guard/ -run '^(TestCoverage|TestCheckGuardSymlinks|TestCheckStageMarkCalls)$'
    -count=1 -race`.

Correction (2026-10-22): step 3 declared that `check_guard_symlinks_test.go` and
`check_stage_mark_calls_test.go` drop `lib/coverage.sh` from their fixture-copy lists. Neither
copies the working tree's file: both write `historicalFile(…, "d71a2327", …)` — the bash guard and
the `coverage.sh` it sourced at that ref — into a temp tree for their bash-parity cases, so the
copy stays and those two files are untouched. Measured: dropping the entry from
`check_guard_symlinks_test.go` fails `TestCheckGuardSymlinks/port:_a_prose_citation_after_a_NUL_is_not_seen/matches_the_bash`.
The `test-lib-coverage.sh` cases the table lacked were the leading-dash, spaced member's three
(5a–5c); the empty corpus and the declared-zero-now-non-zero member were already table cases.
The "declared zero without a reason" case now declares with an empty reason: the one-argument
call was a bash arity distinction the Go API has no form for.

- [x] 3. Sweep mentions of the deleted files

**Files:** `.flow/project.md`, `scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh`, `scripts/check-self-review-report.sh`, `stats/internal/guard/installedcitations.go`, `stats/internal/guard/references.go`, `stats/internal/guard/check_installed_citations_test.go`, `skills/flow-self-review/SKILL.md`, `stats/internal/guard/coverage.go`, `stats/internal/guard/runguardtests.go`, `stats/internal/guard/run_guard_tests_test.go`
**Tests:** none — repairs existing tests: a comment in `check_installed_citations_test.go` only
**Regression:** none — prose only; `scripts/check-references.sh` and the absence sweep below
cover it.
**After:** Task 1, 2
**Commit:** `docs(scripts): name the Go runner and coverage.go in place of the deleted bash`
**Build:** green

  - [x] **Step 1: Rewrite** each mention to name the Go file (`runguardtests.go`, `coverage.go`)
    or drop the clause where it only contrasted the bash library; `.flow/project.md`'s runner
    paragraph says the runner runs harnesses concurrently in `flow-guard`.
  - [x] **Step 2: Verify.** `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh`
    exit 0; `cd stats && gofmt -l . && go vet ./internal/guard/`.
  - [x] **Step 3: Absence sweep.** `git grep -n -F -e 'lib/parallel.sh' -e 'lib/coverage.sh'
    -e 'coverage.sh' -e 'parallel.sh' -e 'test-lib-parallel' -e 'test-lib-coverage'
    -e 'test-run-guard-tests' -- ':!spectre/changes/archive' ':!spectre/changes/kan-921-flow-improvement-port-the-guard-test-runner-run'`
    prints nothing.

Correction (2026-10-22): the absence sweep also prints the two `historicalFile` fixture-copy lines
task 2's correction keeps (`check_guard_symlinks_test.go`'s `"lib/coverage.sh"` and
`check_stage_mark_calls_test.go`'s `"scripts/lib/coverage.sh"`, each copying the file at
`d71a2327`); those two lines are the only output allowed. `**Files:**` widened by `coverage.go`, `runguardtests.go` and
`run_guard_tests_test.go`: tasks 1 and 2 left comments naming the deleted paths for provenance,
and the sweep reached them.
