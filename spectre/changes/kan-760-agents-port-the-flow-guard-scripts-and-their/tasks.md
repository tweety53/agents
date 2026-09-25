# kan-760-agents-port-the-flow-guard-scripts-and-their

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Ports five bash guards to Go behind a new `flow-guard` binary, in dependency order: the binary
and its build (task 1), the suite runner that tests the branch's own binary (task 2), the helpers
three ports share (task 3), the five ports (tasks 4–8), the installer (task 9), and the
before/after measurement (task 10). `design.md` is canonical for every decision; each task cites
its entry by ID.

**The bash guard at `0747740` is each port's specification.** Its header comment states the
contract; its body is the behaviour. A port reproduces arguments, environment overrides, every
verdict line byte for byte, the stdout/stderr split and every exit code. Where the bash source and
its header disagree, stop and report — never pick one silently.

**Baseline, measured before any edit:**

- Suite: 120.7s real, 191.3s user, 249.3s sys; per-harness figures are in `design.md`'s
  **Measurements**.
  <!-- measured: /usr/bin/time -p scripts/run-guard-tests.sh @ 0747740 -->
- `ok:` lines on a green run — the parity floor for each port (**Decision:** parity-by-case-count):
  cleanup-complete 304, task-commit-fields 288, run-reproducer 109, gather-dispatch-context 88,
  panel-reproducer-exit-contract 59.
  <!-- measured: scripts/test-<name>.sh | grep -cE '^ok' @ 0747740 -->
- `stats/` holds 1025 Go `func Test` declarations; `stats/internal/guard/` does not exist.
  <!-- measured: find stats -name '*_test.go' -not -path '*/node_modules/*' | xargs grep -hcE '^func Test' | awk '{s+=$1} END{print s}' @ 0747740 -->
- Helper usage: `sha256-hex.sh` is sourced by cleanup-complete, run-reproducer and
  gather-dispatch-context; `spec-root.sh` by cleanup-complete and task-commit-fields;
  `change-plan.sh` by task-commit-fields and gather-dispatch-context.
  <!-- measured: grep -oE 'lib/[a-z0-9-]+\.sh' scripts/<name>.sh @ 0747740 -->
- `check-task-commit-fields.py` is loaded as a module by `check-task-records.py` and
  `check-plan-shape.py`, so it is not deleted.
  <!-- measured: grep -n COMMIT_FIELDS_PATH scripts/check-task-records.py; head -20 scripts/check-plan-shape.py @ 0747740 -->

**Every port task (4–8) follows the same five-step shape:** port the bash harness's cases to a Go
table test first (red — `Run` does not exist yet), port the guard, run the test green, replace
the bash guard's body with the shim, delete the bash harness. Each Go subtest is named after the
bash case's `ok:` label, so the parity count in task 10 is a `--- PASS` count.

**Shim template** — every ported `scripts/<name>.sh` keeps its header comment block verbatim
(contracts cite gate guards' headers as canonical) and replaces everything after it with (the
body's comments move to the Go file — task 11, **Decision:** flow-guard-binary-rationale-in-go):

```bash verified:authored for this plan (Decision: separate-flow-guard-binary, missing-binary-exits-2)
set -euo pipefail
if ! command -v flow-guard >/dev/null 2>&1; then
  echo "<name>: flow-guard not on PATH — run make -C <agents repo>/stats install-guard" >&2
  exit 2
fi
exec flow-guard <name> "$@"
```

`<name>` is the guard's basename without `.sh`, written literally; `<agents repo>` is written
literally as shown, the same placeholder the repository's prose already uses.

**Test isolation:** every Go test calls `t.Parallel()`; fixture git repositories are built once
per package in `TestMain` and each case copies its own (`cp -R` or `git clone --local`) into
`t.TempDir()`; nothing writes outside `t.TempDir()`. Tests inject deadlines on `Env` as
sub-second `time.Duration`s, never through the environment (**Decision:**
inject-deadlines-in-process).

Live verification: task 10 runs the real suite on this machine and records before/after.

---

- [x] 1. The flow-guard binary and its build

**Files:** `stats/cmd/flow-guard/main.go`, `stats/cmd/flow-guard/main_test.go`, `stats/internal/guard/guard.go`, `stats/Makefile`, `scripts/test-make-build.sh`, `stats/README.md`
**Tests:** `TestUnknownGuardExits2`, `TestNoGuardNameExits2`, `build target emits bin/flow-guard`
**Regression:** `TestUnknownGuardExits2` fails if the dispatcher stops refusing an unregistered
name with exit 2; `TestNoGuardNameExits2` fails if a bare `flow-guard` stops printing usage and
exiting 2; `build target emits bin/flow-guard` fails if `make build` stops writing
`bin/flow-guard`.
**Baseline:** before=0 after=3
<!-- measured: { cat stats/cmd/flow-guard/main_test.go 2>/dev/null | grep -E '^func Test'; grep -E '^assert_[a-z_]+ "build target emits bin/flow-guard"' scripts/test-make-build.sh; } | grep -c . @ 0747740 -->
**After:** none
**Commit:** `feat(stats): add the flow-guard binary and its build targets`
**Build:** green

**Decision:** separate-flow-guard-binary
**Decision:** install-via-make-and-setup

  - [x] **Step 1: Failing tests.** In `stats/cmd/flow-guard/main_test.go` write
    `TestUnknownGuardExits2` (calling `run([]string{"no-such-guard"}, …)` returns 2, stderr
    contains `unknown guard`) and `TestNoGuardNameExits2` (`run(nil, …)` returns 2, stderr carries
    the usage). In `scripts/test-make-build.sh` add, beside the existing `build target emits
    bin/flow` assertion, `assert_recipe_matches "build target emits bin/flow-guard"` matching
    `go build -o bin/flow-guard ./cmd/flow-guard`. Run: `cd stats && go test ./cmd/flow-guard/
    -run 'TestUnknownGuardExits2|TestNoGuardNameExits2'` — expect a compile failure;
    `scripts/test-make-build.sh` — expect the new assertion to FAIL.
  - [x] **Step 2: `stats/internal/guard/guard.go`.** The package's shared types — the only API
    later tasks build on:

```go unverified:compile it; adjust field names only if gofmt/vet demand, and update every later task's use to match
// Package guard holds the Go ports of the flow guard scripts. Each guard
// keeps its bash script's CLI contract; scripts/<name>.sh execs flow-guard.
package guard

// Env is everything a guard reads from its process, injected so tests run
// in-process and in parallel.
type Env struct {
    Getenv func(string) string // os.Getenv in production
    Dir    string              // working directory the bash guard would run in
    // Deadlines override the guards' integer-second env knobs in tests only;
    // zero means "parse the env knob as the bash guard does".
    ReproducerBound, ReproducerGrace, SurvivorsTimeout, SurvivorsKillGrace time.Duration
    // Findings returns `flow record findings -change <name>`'s JSON; nil
    // means exec the `flow` CLI on PATH, as the bash guard does.
    Findings func(change string) ([]byte, error)
}

// Func is one guard's entry point: args exclude the guard name; the return
// value is the process exit code.
type Func func(args []string, env Env, stdout, stderr io.Writer) int

// Registry maps a guard's basename (without .sh) to its Func. Each port
// task adds its own entry in its own file's init().
var Registry = map[string]Func{}
```
  - [x] **Step 3: `stats/cmd/flow-guard/main.go`.** `main` calls
    `os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))`; `run` prints usage and returns 2
    on no arguments, prints `flow-guard: unknown guard <name>` and returns 2 on an unregistered
    name, else calls `guard.Registry[name](args[1:], guard.Env{Getenv: os.Getenv, Dir: cwd},
    stdout, stderr)`. Header comment in the style of `stats/cmd/flow/main.go`'s.
  - [x] **Step 4: `stats/Makefile`.** `build` gains `go build -o bin/flow-guard
    ./cmd/flow-guard` after the `bin/flow` line. Add `install-guard`: `mkdir -p
    $(HOME)/.local/bin` and `go build -o $(HOME)/.local/bin/flow-guard ./cmd/flow-guard`, with a
    comment stating it never touches the dev daemon; make `restart` depend on `install-guard`
    (`restart: build install-guard`). Add `install-guard` to `.PHONY` if the file declares one.
  - [x] **Step 5: `stats/README.md`.** One short section, `flow-guard`: what it is (Go ports of
    guard scripts, invoked through `scripts/<name>.sh` shims), how it is installed (`make
    install-guard`, `make restart`, `setup.sh global`), and that `design.md` of this change holds
    the measurements — no figures copied.
  - [x] **Step 6: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./cmd/flow-guard/
    -run 'TestUnknownGuardExits2|TestNoGuardNameExits2' -count=1`; `scripts/test-make-build.sh`;
    `scripts/check-vocabulary.sh`; `scripts/check-references.sh`; `scripts/check-contract-budget.sh`.

Correction (2026-09-25): `run` takes no stdin — `run(args []string, stdout, stderr io.Writer) int` —
since `Func` takes none and none of the five guards reads it; `main` calls
`os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))`.

- [x] 2. The suite runner tests the branch's own flow-guard

**Files:** `scripts/run-guard-tests.sh`, `scripts/test-run-guard-tests.sh`, `scripts/test-go-guards.sh`, `.flow/project.md`
**Tests:** `case 8:`, `case 9:`, `case 10:`
**Regression:** `case 8:` fails if a shim guard whose Go test exists is refused as missing its
companion; `case 9:` fails if a shim guard with neither companion is no longer refused; `case
10:` fails if the harnesses stop seeing a `flow-guard` built from the runner's own tree first on
PATH — the branch's shims would then run the installed binary.
**Baseline:** before=38 after=41
<!-- measured: grep -cE 'pass "case [0-9]+[a-z]?:' scripts/test-run-guard-tests.sh @ 0747740 -->
**After:** Task 1
**Commit:** `feat(scripts): build flow-guard for the guard suite and accept Go companions`
**Build:** green

**Decision:** go-tests-replace-harnesses

  - [x] **Step 1: Failing cases** in `scripts/test-run-guard-tests.sh`, in its existing
    `RUN_GUARD_TESTS_ROOT` fixture style: `case 8:` a fixture `check-x.sh` whose body is the
    shim (contains `exec flow-guard check-x`) plus `../stats/internal/guard/check_x_test.go`
    relative to the fixture root, no `test-check-x.sh` — the runner does not refuse it; `case 9:`
    the same shim with neither companion — refused, exit 1, the stderr names `check-x.sh`; `case
    10:` a fixture harness that runs `command -v flow-guard` and prints it — the path printed is
    under the runner's temp directory, not `~/.local/bin`. Run the harness — expect all three to
    FAIL.
  - [x] **Step 2: Companion rule.** In `scripts/run-guard-tests.sh`'s companion scan (the loop
    filling `MISSING_COMPANIONS`), accept a guard as covered when `test-<name>.sh` exists **or**
    the guard's file contains `exec flow-guard <name>` and
    `$TEST_ROOT/../stats/internal/guard/<name with - replaced by _>_test.go` exists. Update the
    header's COMPANION PRESENCE paragraph to state the second form.
  - [x] **Step 3: Build flow-guard first.** Before launching harnesses, build
    `go build -o "$tmp/bin/flow-guard" ./cmd/flow-guard` from `$TEST_ROOT/../stats` into a
    `mktemp -d` directory, export `PATH="$tmp/bin:$PATH"`, and remove the directory on exit
    through `scripts/lib/parallel.sh`'s existing cleanup (its header: it owns the traps — add no
    competing trap; if it offers no hook, remove the directory after the summary and on the
    failure path explicitly). A failed build is exit 2 with the build output on stderr. Skip the
    build (one stderr line) when `$TEST_ROOT/../stats/cmd/flow-guard` does not exist, so fixture
    roots without it still run. Header paragraph: why — a shim would otherwise run the installed
    binary, not the branch's.
  - [x] **Step 4: `scripts/test-go-guards.sh`.** A harness: `cd "$SCRIPT_DIR/../stats" && go test
    ./internal/guard/... -count=1`, header stating it is the companion harness for every ported
    guard.
    <!-- measured: scripts/check-guard-symlinks.sh and scripts/check-vocabulary.sh both exit 0 with scripts/test-go-guards.sh present @ 1c35fe3 -->
  - [x] **Step 5: `.flow/project.md`'s `## test`** — one sentence after the paragraph on
    `run-guard-tests.sh`: it builds `flow-guard` from the tree and puts it first on PATH, and
    `test-go-guards.sh` runs the Go guard tests.
  - [x] **Step 6: Verify.** `scripts/test-run-guard-tests.sh`; `scripts/check-vocabulary.sh`;
    `scripts/check-references.sh`; `scripts/check-guard-symlinks.sh`;
    `scripts/check-contract-budget.sh`.

Correction (2026-09-25): `case 9:` passed before the change — it pins a regression rather than
showing RED; breaking the companion rule on purpose failed it (`got 0`), restored it passed.

- [x] 3. Shared helpers: sha256, spec root, change plan

**Files:** `stats/internal/guard/sha256.go`, `stats/internal/guard/specroot.go`, `stats/internal/guard/changeplan.go`, `stats/internal/guard/helpers_test.go`
**Tests:** `TestSHA256Hex`, `TestSpecRoot`, `TestChangePlan`
**Regression:** each fails if its helper's Go port diverges from the bash helper's output on the
cases the test pins.
**Baseline:** before=0 after=3
<!-- measured: cat stats/internal/guard/helpers_test.go 2>/dev/null | grep -cE '^func Test' @ 0747740 -->
**After:** Task 1
**Commit:** `feat(stats): port the guard helpers sha256, spec root and change plan to Go`
**Build:** green

  - [x] **Step 1: Scope.** For each of `scripts/lib/sha256-hex.sh`, `scripts/lib/spec-root.sh`,
    `scripts/lib/change-plan.sh`, list the functions the five guards actually call (`grep` each
    guard for the function names). Port only those.
  - [x] **Step 2: Failing tests** in `helpers_test.go`: `TestSHA256Hex` (known vectors, including
    the empty file); `TestSpecRoot` and `TestChangePlan` as table tests whose rows are the
    relevant cases of `scripts/test-lib-change-plan.sh` and any spec-root cases in the harnesses
    of the guards that use it — each row carrying the bash helper's own expected output. Run
    `cd stats && go test ./internal/guard/ -run 'TestSHA256Hex|TestSpecRoot|TestChangePlan'` —
    expect compile failure.
  - [x] **Step 3: Port** into the three files, `sha256` through `crypto/sha256` (no subprocess).
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
    ./internal/guard/ -run 'TestSHA256Hex|TestSpecRoot|TestChangePlan' -count=1 -race`.

- [x] 4. Port run-reproducer

**Files:** `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`, `stats/internal/guard/metachars.go`, `scripts/run-reproducer.sh`, `scripts/test-run-reproducer.sh`, `scripts/test-check-panel-reproducers.sh`
**Tests:** `TestRunReproducer`, `TestMetacharsMatchBashSource`
**Regression:** `TestRunReproducer` fails if any of the bash harness's 109 `ok:` behaviours
regress; `TestMetacharsMatchBashSource` fails if the Go banned-character set drifts from
`scripts/reproducer-metachars.sh`, which bash scripts still source.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/runreproducer_test.go 2>/dev/null | grep -cE '^func Test' @ 0747740 -->
**After:** Task 2, 3
**Commit:** `feat(stats): port run-reproducer to Go`
**Build:** green

**Decision:** inject-deadlines-in-process
**Decision:** helper-parity-tests
**Decision:** parity-by-case-count

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-run-reproducer.sh` to
    `TestRunReproducer`, one subtest per `ok:` label, calling `guard.Registry["run-reproducer"]`
    in-process. The timeout/grace/double-fork cases set `Env.ReproducerBound`/`ReproducerGrace`
    to sub-second values instead of waiting out `sleep 30` fixtures; each such case still
    asserts the survivor was killed (process gone), so a shorter deadline never weakens the
    assertion. `TestMetacharsMatchBashSource` parses the set out of
    `../../../scripts/reproducer-metachars.sh` and compares it to `metachars.go`'s. Run — expect
    failure.
  - [x] **Step 2: Port** `scripts/run-reproducer.sh` (and the `reproducer-metachars.sh`,
    `lib/coverage.sh` parts it sources) to `runreproducer.go`/`metachars.go`, registering
    `run-reproducer`. The reproducer runs under `exec.CommandContext` with `SysProcAttr{Setpgid:
    true}`; the bound, the SIGTERM→SIGKILL grace and the survivor sweep kill the process group,
    driven by timers, never a poll loop. `RUN_REPRODUCER_BOUND_SECONDS`,
    `RUN_REPRODUCER_GRACE_SECONDS`, `RUN_REPRODUCER_BOUND_FILE` keep the bash parsing and
    defaults when `Env`'s durations are zero.
  - [x] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run
    'TestRunReproducer|TestMetacharsMatchBashSource' -count=1 -race -v | grep -c -- '--- PASS:
    TestRunReproducer/'` — at least 109.
  - [x] **Step 4: Shim and delete.** Replace `scripts/run-reproducer.sh`'s body with the shim
    template; `git rm scripts/test-run-reproducer.sh`.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`; `make -C stats
    build` then with `stats/bin` first on PATH run the callers' harnesses:
    `scripts/test-prove-reproducer.sh`, `scripts/test-mutate-and-verify.sh`,
    `scripts/test-check-panel-reproducers.sh`, `scripts/test-check-mutation-reproducer-pin.sh`,
    `scripts/test-break-and-prove.sh`; `scripts/check-guard-symlinks.sh`;
    `scripts/check-references.sh`; `scripts/check-vocabulary.sh`.

Correction (2026-09-25): shipped differently from the plan in four measured points.
- `scripts/test-check-panel-reproducers.sh` case 19 grepped `run-reproducer.sh` for its
  `source reproducer-metachars.sh` line, which the shim removes; it now asserts the shim line and
  the Go parity test, `ok:` label unchanged — `**Files:**` widened.
- Step 2's "`lib/coverage.sh` parts it sources" is `reproducer-metachars.sh` and
  `lib/sha256-hex.sh`; `run-reproducer.sh` names `coverage.sh` only in comments.
  <!-- measured: grep -n 'coverage.sh' scripts/run-reproducer.sh @ 0747740 -->
- `SysProcAttr{Setsid: true}`, not `Setpgid` — the bash guard's `os.setsid()` (new session and group).
- The bound and the grace are timers; the descendant snapshot (200ms) and the survivor liveness
  check (10ms, under the grace deadline) stay periodic — a self-detaching child is visible only
  while its parent lives, and no portable wait exists on a pid the guard did not start.
- The shim exits 4, not 2, when `flow-guard` is missing — run-reproducer's own cannot-answer code
  (2 is its "refused"); **Decision:** missing-binary-cannot-answer-code.
- Test bound: macOS's first exec of a freshly written script costs ~0.2s (0.201s cold, 0.009s
  warm), so timeout cases fire the bound through `RUN_REPRODUCER_BOUND_FILE` when the fixture is
  ready, with a 30s backstop bound — no assertion loosened.

- [x] 5. Port check-panel-reproducer-exit-contract

**Files:** `stats/internal/guard/panelexitcontract.go`, `stats/internal/guard/check_panel_reproducer_exit_contract_test.go`, `scripts/check-panel-reproducer-exit-contract.sh`, `scripts/test-check-panel-reproducer-exit-contract.sh`
**Tests:** `TestCheckPanelReproducerExitContract`
**Regression:** fails if any of the bash harness's 59 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_panel_reproducer_exit_contract_test.go 2>/dev/null | grep -cE '^func Test' @ 0747740 -->
**After:** Task 4
**Commit:** `feat(stats): port check-panel-reproducer-exit-contract to Go`
**Build:** green

**Decision:** parity-by-case-count

  - [x] **Step 1: Failing test.** Port every case of the bash harness, one subtest per `ok:`
    label. The stub `flow` on PATH becomes `Env.Findings` returning the canned JSON (or an error
    for the store-unreachable case); the stub runner becomes a fake reproducer script the real
    in-process runner executes. Cases 16 and 17 keep using the real runner against a real
    reproducer, positive and inverted. Run — expect failure.
  - [x] **Step 2: Port**, registering `check-panel-reproducer-exit-contract`; it calls the
    run-reproducer Go function in-process instead of exec'ing `$SCRIPT_DIR/run-reproducer.sh`.
    With `Env.Findings` nil it execs `flow record findings -change <name>` exactly as the bash
    guard does.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run TestCheckPanelReproducerExitContract
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckPanelReproducerExitContract/'` — at least
    59.
  - [x] **Step 4: Shim and delete** — shim template; `git rm
    scripts/test-check-panel-reproducer-exit-contract.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet`; `scripts/test-check-panel-reproducers.sh` with
    `stats/bin` first on PATH; `scripts/check-guard-symlinks.sh`; `scripts/check-references.sh`.

Correction (2026-09-25): the test file was planned as `panelexitcontract_test.go`; `scripts/run-guard-tests.sh`'s
companion rule (task 2, **Decision:** go-tests-replace-harnesses) accepts only
`stats/internal/guard/<name with - replaced by _>_test.go` for a shim guard, so it is renamed to match.

Correction (2026-09-25): the bash guard also calls `flow state get`, which `Env` has no hook
for; the tests keep a stub `flow` on `Env`'s PATH for it and use `Env.Findings` for findings —
case 12 runs through the real `flow record findings` call. Subtest count 60 against the floor of 59.

- [x] 6. Port gather-dispatch-context

**Files:** `stats/internal/guard/gatherdispatch.go`, `stats/internal/guard/gatherdispatch_test.go`, `scripts/gather-dispatch-context.sh`, `scripts/test-gather-dispatch-context.sh`
**Tests:** `TestGatherDispatchContext`
**Regression:** fails if any of the bash harness's 88 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/gatherdispatch_test.go 2>/dev/null | grep -cE '^func Test' @ 0747740 -->
**After:** Task 2, 3
**Commit:** `feat(stats): port gather-dispatch-context to Go`
**Build:** green

**Decision:** parity-by-case-count

  - [x] **Step 1: Failing test** — every harness case, one subtest per `ok:` label; the
    no-hash-tool case becomes a case asserting the Go port needs no hash tool (it uses
    `crypto/sha256`) — state that substitution in the subtest's comment. Run — expect failure.
  - [x] **Step 2: Port**, registering `gather-dispatch-context`, including the parts of
    `lib/resolve-file.sh`, `lib/within-root.sh`, `lib/lexical-normalize.sh`,
    `lib/project-section.sh` it sources, in `gatherdispatch.go`.
  - [x] **Step 3: Green** — `--- PASS: TestGatherDispatchContext/` count at least 88.
  - [x] **Step 4: Shim and delete** — `git rm scripts/test-gather-dispatch-context.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet`; `scripts/check-guard-symlinks.sh`;
    `scripts/check-references.sh`; `scripts/check-vocabulary.sh`.

- [x] 7. Port check-task-commit-fields

**Files:** `stats/internal/guard/taskcommitfields.go`, `stats/internal/guard/check_task_commit_fields_test.go`, `scripts/check-task-commit-fields.sh`, `scripts/test-check-task-commit-fields.sh`
**Tests:** `TestCheckTaskCommitFields`, `TestTaskFieldParseMatchesPython`
**Regression:** `TestCheckTaskCommitFields` fails if any of the bash harness's 288 `ok:`
behaviours regress; `TestTaskFieldParseMatchesPython` fails if the Go task-field parse drifts
from `scripts/check-task-commit-fields.py`, which `check-task-records.py` and
`check-plan-shape.py` still load.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/check_task_commit_fields_test.go 2>/dev/null | grep -cE '^func Test' @ 0747740 -->
**After:** Task 2, 3
**Commit:** `feat(stats): port check-task-commit-fields to Go`
**Build:** green

**Decision:** helper-parity-tests
**Decision:** parity-by-case-count

  - [x] **Step 1: Failing tests.** Port all 139 numbered cases, one subtest per `ok:` label.
    `TestTaskFieldParseMatchesPython` runs `python3` once over every plan fixture the Go test
    uses, importing `check-task-commit-fields.py` the way `check-task-records.py` does
    (`load_module`), dumps each task's parsed fields as JSON, and compares with the Go parse.
    Run — expect failure.
  - [x] **Step 2: Port** the wrapper and `check-task-commit-fields.py` (with the `plan_grammar`
    parts it imports) to `taskcommitfields.go`, registering `check-task-commit-fields`.
    `check-task-commit-fields.py` stays in place, unchanged.
  - [x] **Step 3: Green** — `--- PASS: TestCheckTaskCommitFields/` count at least 288.
  - [x] **Step 4: Shim and delete** — `git rm scripts/test-check-task-commit-fields.sh`. Keep
    `scripts/check-task-commit-fields.py`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet`; `scripts/test-check-task-records.sh`,
    `scripts/test-check-plan-shape.sh`, `scripts/test-check-unfinished-work.sh`,
    `scripts/test-lib-change-plan.sh` with `stats/bin` first on PATH;
    `scripts/check-guard-symlinks.sh`; `scripts/check-references.sh`.

Correction (2026-09-25): the test file was planned as `taskcommitfields_test.go`; `scripts/run-guard-tests.sh`'s
companion rule (task 2, **Decision:** go-tests-replace-harnesses) accepts only
`stats/internal/guard/<name with - replaced by _>_test.go` for a shim guard, so it is renamed to match.

Correction (2026-09-25): shipped differently from the plan in these measured points.
- Step 3's count is `grep -cE -- '--- PASS: TestCheckTaskCommitFields/[^/ ]+/'` (288); the plain
  prefix also counts case-group lines (403).
- No `TestMain`: Go allows one per package and four ports share it, so each top-level test builds
  its base repo once in its own temp dir.
- Case 56 tested the wrapper's missing `lib/plan_grammar.py`, unreachable once compiled in; it now
  runs the shim with no `flow-guard` on PATH (exit 2, install message). The wrapper's other exit-2
  refusals (missing grammar/spec-root/change-plan module, missing or broken python3) are gone —
  nothing left to be missing.
- An unreadable (non-UTF-8) plan exits 1 with Python's final `UnicodeDecodeError` line, as the
  `.py` actually does — its header's "exit 2" was never true.
- Case 124's short timeout is a `tcfRunMeasuredAt` argument (600s prod, 1s test), not an `Env` field.
- Python's Unicode `\d`/`\w`/`\b` are ASCII in the port: a non-ASCII letter or digit glued to a
  task id, Case label or build keyword can parse differently; no plan in the corpus has one.
- Review fix: a `[` inside a glob class is escaped as Python's `re` reads it, closing a false pass
  on `Allowed-collateral: docs/[x[:alpha:]*[y]` (`TestTcfFnmatchMatchesPython`).

- [ ] 8. Port check-cleanup-complete

**Files:** `stats/internal/guard/cleanupcomplete.go`, `stats/internal/guard/check_cleanup_complete_test.go`, `scripts/check-cleanup-complete.sh`, `scripts/test-check-cleanup-complete.sh`
**Tests:** `TestCheckCleanupComplete`
**Regression:** fails if any of the bash harness's 304 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_cleanup_complete_test.go 2>/dev/null | grep -cE '^func Test' @ 0747740 -->
**After:** Task 2, 3
**Commit:** `feat(stats): port check-cleanup-complete to Go`
**Build:** green

**Decision:** inject-deadlines-in-process
**Decision:** parity-by-case-count

  - [ ] **Step 1: Failing test** — every harness case, one subtest per `ok:` label; the
    `skip:` path (locale) stays a `t.Skip` with the same reason. Survivor-timeout cases set
    `Env.SurvivorsTimeout`/`SurvivorsKillGrace` sub-second and still assert the survivor was
    terminated. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-cleanup-complete`. `CHECK_CLEANUP_SURVIVORS_TIMEOUT`
    keeps its bash parsing, range check and 60s default when `Env.SurvivorsTimeout` is zero; the
    survivor wait uses `exec.CommandContext` and timers, never a poll loop.
  - [ ] **Step 3: Green** — `--- PASS: TestCheckCleanupComplete/` count at least 304.
  - [ ] **Step 4: Shim and delete** — the shim keeps the header's hand-verification paragraph
    verbatim (`pipeline.md`'s **Hand-verifying a guard verdict** cites it); `git rm
    scripts/test-check-cleanup-complete.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet`; `scripts/test-check-worktree-processes.sh`,
    `scripts/test-resolve-base-branch.sh`, `scripts/test-check-workspace-isolation.sh`,
    `scripts/test-workspace.sh` with `stats/bin` first on PATH; `scripts/check-guard-symlinks.sh`;
    `scripts/check-references.sh`.

Correction (2026-09-25): the test file was planned as `cleanupcomplete_test.go`; `scripts/run-guard-tests.sh`'s
companion rule (task 2, **Decision:** go-tests-replace-harnesses) accepts only
`stats/internal/guard/<name with - replaced by _>_test.go` for a shim guard, so it is renamed to match.

Correction (2026-09-25): Step 1's "sub-second" and Step 2's "`exec.CommandContext`" shipped as
`exec.Command` + `Setpgid` (bash's `set -m`) + a `time.Timer`. Cases 28, 28b, 28d fire the bound
on readiness and 29 never fires it, through a new `Env.SurvivorsExpire` channel in
`stats/internal/guard/guard.go` — a wall-clock bound, even 3s, raced macOS's ~0.2s first exec
under the package's parallel `-race` load. Review fixes: survivors run as `bash` (argv[0]), and
`TestCCSurvivorsTimeoutParsesEnvKnob` pins `CHECK_CLEANUP_SURVIVORS_TIMEOUT`'s parsing.

- [x] 9. setup.sh global builds flow-guard

**Files:** `setup.sh`, `scripts/test-setup.sh`, `scripts/check-installed-citations.py`
**Tests:** `global installs flow-guard`
**Regression:** fails if `setup.sh global` stops building `flow-guard` into
`$HOME/.local/bin`.
**Baseline:** before=144 after=145
<!-- measured: grep -cE '^assert_[a-z_]+ "' scripts/test-setup.sh @ 0747740 -->
**After:** Task 1
**Commit:** `feat(setup): build flow-guard in setup.sh global`
**Build:** green

**Decision:** install-via-make-and-setup

  - [x] **Step 1: Failing assertion** in `scripts/test-setup.sh`'s global run: `assert_…
    "global installs flow-guard"` — `$home/.local/bin/flow-guard` exists and is executable. Run —
    expect FAIL.
  - [x] **Step 2: `setup.sh`** — in `install_global`, a function `install_flow_guard` that runs
    `go build -o "$HOME/.local/bin/flow-guard" ./cmd/flow-guard` from the repository's `stats/`;
    no Go toolchain is `die "setup.sh global needs Go to build flow-guard"` (operator's choice:
    a machine without Go fails setup). Add it to the finish banner's summary lines the way other
    installs are listed.
    <!-- measured: a sandbox HOME moves go's default cache — cold build 2.82s real, 0.46s with the real GOCACHE exported; test-setup.sh exports GOCACHE="$(go env GOCACHE)" before sandboxing @ 416330f -->
  - [x] **Step 3: Verify.** `scripts/test-setup.sh`; `scripts/test-installer-sandbox-diff.sh`;
    `scripts/check-installed-citations.sh`; `scripts/check-vocabulary.sh`.

Correction (2026-09-25): the plan declared `setup.sh` and `scripts/test-setup.sh` only. The same
cold-cache cost hit `scripts/check-installed-citations.py`'s two sandboxed `setup.sh` runs (2.98s →
4.55s real), so its `run_setup` now passes the real HOME's `GOCACHE` too, folded into this task's
commit; `**Files:**` widened to match.

- [ ] 10. Live verification: before/after timings

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 4, 5, 6, 7, 8, 9, 11
**Commit:** none — this task commits nothing to this repository; its figures go into `design.md`'s **Measurements**, committed with the change's artifacts
**Build:** green

  - [ ] **Step 1: Suite, after.** On this machine, nothing else heavy running:
    `/usr/bin/time -p scripts/run-guard-tests.sh` three times; record each run's real/user/sys
    and harness count, and the slowest three harnesses.
  - [ ] **Step 2: Go package, after.** `cd stats && /usr/bin/time -p go test ./internal/guard/...
    -count=1` three times (warm build cache); record real/user/sys.
  - [ ] **Step 3: Parity.** `go test ./internal/guard/ -count=1 -v | grep -c -- '--- PASS:
    Test<Name>/'` per port against the floors in the Baseline block above.
  - [ ] **Step 4: Record** an **After** table in `design.md`'s **Measurements**, same columns as
    **Before**, each figure tagged `measured:` with the command and `@ branch
    spectre/kan-760-agents-port-the-flow-guard-scripts-and-their`.
  - [ ] **Step 5: Judge.** Failure looks like: suite real not below the Before 120.7s; any
    port's `--- PASS` count below its floor; `go test ./internal/guard/...` above 10s real; any
    harness red. Any of these is reported, not recorded as success.
    <!-- predicted: suite ≈ 60s real, bounded by test-check-panel-reproducers.sh; guard package < 10s — confirmed by steps 1–2 -->

- [ ] 11. Move the ported guards' rationale into Go and repoint its citations

**Files:** `rules/commit-scope-is-the-module.mdc`, `scripts/check-cleanup-complete.sh`, `scripts/check-mutation-reproducer-pin.sh`, `scripts/check-panel-reproducer-exit-contract.sh`, `scripts/check-panel-reproducers.sh`, `scripts/check-task-commit-fields.py`, `scripts/check-task-commit-fields.sh`, `scripts/check-unfinished-work.sh`, `scripts/check-workspace-isolation.sh`, `scripts/gather-dispatch-context.sh`, `scripts/lib/change-plan.sh`, `scripts/lib/lexical-normalize.sh`, `scripts/lib/plan_grammar.py`, `scripts/lib/project-section.sh`, `scripts/lib/reproducer-path.sh`, `scripts/lib/sanitize-display.sh`, `scripts/lib/sha256-hex.sh`, `scripts/lib/spec-root.sh`, `scripts/lib/visual-table-cells.awk`, `scripts/lib/within-root.sh`, `scripts/reproducer-metachars.sh`, `scripts/prove-reproducer.sh`, `scripts/resolve-base-branch.sh`, `scripts/run-reproducer.sh`, `scripts/test-check-mutation-reproducer-pin.sh`, `scripts/test-check-task-build-green.sh`, `scripts/test-check-unfinished-work.sh`, `scripts/test-check-workspace-isolation.sh`, `scripts/test-lib-change-plan.sh`, `scripts/test-prove-reproducer.sh`, `scripts/test-resolve-base-branch.sh`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-contracts/pipeline-rationale.md`, `skills/flow-contracts/project-configuration-rationale.md`, `stats/internal/guard/check_cleanup_complete_test.go`, `stats/internal/guard/cleanupcomplete.go`, `stats/internal/guard/gatherdispatch.go`, `stats/internal/guard/panelexitcontract.go`, `stats/internal/guard/runreproducer.go`, `stats/internal/guard/taskcommitfields.go`
**Tests:** `scripts/check-mutation-reproducer-pin.sh`, `TestCheckCleanupComplete`
**Regression:** `scripts/check-mutation-reproducer-pin.sh` exits 2 while it reads a line the
`run-reproducer.sh` shim removed; `TestCheckCleanupComplete` fails if a moved comment's code edit
changes behaviour.
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task; check-mutation-reproducer-pin.sh goes from exit 2 to exit 0 -->
**After:** Task 2, 3, 4, 5, 6, 7, 8
**Commit:** `docs(stats): move the ported guards' rationale into Go and repoint its citations`
**Build:** green

**Decision:** flow-guard-binary-rationale-in-go

  - [ ] **Step 1: Inventory.** For each of the five guards, diff its `0747740` bash body against its
    shim (`git show 0747740:scripts/<name>.sh`) and list every comment block below the header. The
    sweep in `<worktree>/.superpowers/sdd/reviewer-report-task-8.md`'s `## Dangling-reference sweep`
    is the citation list to repoint; confirm each entry and extend it with `grep -rn` for each
    guard's basename, its deleted `test-<name>.sh`, and each moved comment's own title across
    `scripts/`, `skills/`, `rules/`, `.flow/`, `README.md`, `CONTRIBUTING.md` (never
    `spectre/changes/archive/`).
  - [ ] **Step 2: Move.** Place each comment beside the Go code it explains, in Go comment style,
    updated only where the mechanism changed (python3 shim → `Setsid`, poll → timer, a sourced lib →
    the Go helper) — reasoning, rejected alternatives and recorded incidents kept whole. Correct a kept
    header only where the port made a statement false; say so in the shim's first comment line after
    the header.
  - [ ] **Step 3: Repoint.** Every citation names the Go file (`stats/internal/guard/<file>.go`, the
    moved comment's title) or the Go test (`TestX/<case>`). `lib/*.sh` headers saying "sourced by
    <guard>" drop that guard. `rules/commit-scope-is-the-module.mdc:35` and
    `skills/flow-contracts/pipeline-rationale.md:171` stop naming the `.py` as the enforcer.
  - [ ] **Step 4: Lint.** `scripts/check-mutation-reproducer-pin.sh` reads the pinned line from
    `stats/internal/guard/runreproducer.go` (or wherever step 2 put it); prove it bites by breaking
    the pin and capturing its failure, then restoring. Its harness
    `scripts/test-check-mutation-reproducer-pin.sh` stays green.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./internal/guard/
    -count=1 -race`; every `## lint` command in `.flow/project.md`; with a tree-built `flow-guard`
    first on PATH, `scripts/test-check-mutation-reproducer-pin.sh`, `scripts/test-check-panel-reproducers.sh`,
    `scripts/test-check-workspace-isolation.sh`, `scripts/test-check-unfinished-work.sh`,
    `scripts/test-resolve-base-branch.sh`, `scripts/test-prove-reproducer.sh`,
    `scripts/test-check-task-build-green.sh`, `scripts/test-lib-change-plan.sh`.

Correction (2026-09-25): `**Files:**` narrowed to the 39 paths the commit touched — the other 17
declared paths held no stale citation — and widened by `scripts/reproducer-metachars.sh`, whose
"sourced by both callers" line named run-reproducer. `**Tests:**` drops `scripts/check-references.sh`
and `TestRunReproducer`, which the commit's diff does not name; both still run green. Comments with
no Go counterpart (the bash 3.2 floor, the FIFO/fd 9 sentinel, `SECONDS`, the job-notification
redirect, the three-way grep status, awk failing) were dropped, each listed in the implementer
report. `scripts/lib/lexical-normalize.sh` and `scripts/lib/within-root.sh` now have no bash
caller; their headers say so, and deleting them is out of this change's scope.
Review fix (2026-09-25): four citations still dangled — `case_92`/`case_93` (one subtest,
`case_92-93`), `test-check-task-build-green.sh`'s "that harness's case 56", `prove-reproducer.sh`'s
"its existing harness" (the deleted `test-run-reproducer.sh`; `**Files:**` widened) and
`lib/within-root.sh`'s present-tense "Sourced by" — repointed.
