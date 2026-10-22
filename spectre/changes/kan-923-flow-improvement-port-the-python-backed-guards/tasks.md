# kan-923-flow-improvement-port-the-python-backed-guards

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Ports seven Python guards into `flow-guard` (tasks 1–7) and deletes the shared Python modules with
the citation sweep (task 8). `design.md` is canonical for every decision; each task cites its entry
by ID.

**The Python and its wrapper at `bbfe685e` are each port's specification.** The `.py` docstring and
the `.sh` header state the contract; their bodies are the behaviour. A port reproduces arguments,
environment overrides (each wrapper's `*_ROOT`, `CHECK_TASK_RECORDS_ARCHIVED` and the like), every
output line byte for byte, the stdout/stderr split and every exit code — the wrapper's own refusals
(missing sibling, bad usage) included, except the `python3` probes, which have no counterpart. Where
a header and a body disagree, follow the header and record it as KAN-918's
`header-wins-on-exit-code` did; never pick silently.

**Parity floors** are in `design.md`'s **Measurements** (**Decision:**
carry-prior-port-decisions — KAN-760's `parity-by-case-count`).

**Every port task (1–7) follows the same five steps:** port the harness's cases to a Go table test
first (red — the guard is unregistered), port the guard, run the test green, replace the wrapper's
body with the shim, `git rm` the harness and the task's own entry `.py`. Each Go subtest is named
after the harness case's `ok:` (or `PASS:`) label, so parity is a `--- PASS` count. The Python
docstring's and the wrapper's comments move into the Go file beside the code they explain, updated
where the mechanism changed (KAN-760's `flow-guard-binary-rationale-in-go`). Each port registers its
basename in `guard.Registry` from its own file's `init()`, as `basemoved.go` does.

**Grammar** — a `lib/plan_grammar.py` function, a `check-task-commit-fields.py` parse function or
`plan-dispatch-bundles.py`'s `parse_tasks` that a port needs is reused from `planparse.go` when a twin
already exists there, and added there otherwise, never inside the port's own file (**Decision:**
grammar-in-planparse).

**Shim template** — the wrapper's header comment block kept, corrected only where the port makes a
statement false (Python and bash plumbing), then:

```bash verified:copied from the tail of scripts/check-ask-silence.sh @ bbfe685e
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "<name>: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec <name> 2 "<name>:" "$@"
```

`<name>` is the basename without `.sh`, written literally. The two `FLOW_GUARD_REPO_ROOT` lines and
the `flow_guard_cwd_checkout` line are kept only by a shim whose wrapper derived the repository root
from its own location (**Decision:** shims-export-repo-root); every other shim drops them.

**Test isolation:** every Go test calls `t.Parallel()`; fixture trees (git repos included) are built
once in `TestMain` or a `sync.Once` helper (`helpers_test.go` has the existing ones) and each case
copies its own into `t.TempDir()`; deadlines are injected on `Env`, never through the environment
(KAN-760's `inject-deadlines-in-process`). No test reads or writes the operator's real `$HOME`.

**Per-task verify** is `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
./internal/guard/ -run '^<Test>$' -count=1 -race`, then the shimmed script run once for real as its
step 5 names. The full package and `scripts/run-guard-tests.sh` run in `flow.verify`.

## Review Focus

- A shim run from an installed skill directory (`~/.claude/skills/flow/scripts/<name>.sh`, a
  symlink) must behave as the wrapper did — the no-argument scans resolve the repository root from
  the shim's real location, never the caller's working directory.
- `plan-dispatch-bundles` and `plan-dispatch-groups` print what `/flow`'s Decide step and
  `gather-dispatch-context` (`gatherdispatch.go`) read: their stdout must stay byte-identical, or a
  recorded decision stops matching its plan.
- The shared grammar: one Go twin per Python function, every port reading the same one — a task
  line or fence the shape guard accepts must be the one the bundler and the commit-fields guard see.
- Python's `str.splitlines()`, `re` and `sorted()` semantics reproduced exactly where they decide an
  output (line boundaries, Unicode classes, ordering).

---

- [ ] 1. Port check-plan-shape

**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`, `stats/internal/guard/planparse.go`, `scripts/check-plan-shape.sh`, `scripts/check-plan-shape.py`, `scripts/test-check-plan-shape.sh`
**Allowed-collateral:** `stats/internal/guard/helpers_test.go`
**Tests:** `TestCheckPlanShape`
**Regression:** fails if any of the harness's 131 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_plan_shape_test.go 2>/dev/null | grep -cE '^func Test' @ bbfe685e -->
**After:** none
**Commit:** `feat(guard): port check-plan-shape to Go`
**Build:** green

**Decision:** scope-seven-guards
**Decision:** plan-shape-subguards-in-process
**Decision:** shims-export-repo-root

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-plan-shape.sh`, one subtest
    per `ok:` label, the `CHECK_PLAN_SHAPE_ROOT` scans and the unlanded-archive scans included. Run —
    expect failure.
  - [ ] **Step 2: Port**, registering `check-plan-shape`: every check of `check-plan-shape.py`, then
    `plan-suite-steps`, `plan-script-paths` and `plan-task-id-rule` in-process per file, and the
    no-argument scan over `<spec-root>/changes/*/tasks.md` plus `unlanded-archives`, with
    `check_one`'s exit combination. The `plan_grammar.py` and commit-fields parse twins it needs go
    in `planparse.go`.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestCheckPlanShape$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckPlanShape/'` — at least 131.
  - [ ] **Step 4: Shim and delete** — shim template with the repo-root lines, code 2;
    `git rm scripts/test-check-plan-shape.sh scripts/check-plan-shape.py`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/check-plan-shape.sh` (no argument) and `scripts/check-plan-shape.sh
    spectre/changes/kan-923-flow-improvement-port-the-python-backed-guards/tasks.md` exit and print
    as the wrapper did at `bbfe685e`, directly and through a symlink to the shim in a temp
    directory.

- [ ] 2. Port check-task-build-green

**Files:** `stats/internal/guard/taskbuildgreen.go`, `stats/internal/guard/check_task_build_green_test.go`, `stats/internal/guard/planparse.go`, `scripts/check-task-build-green.sh`, `scripts/check-task-build-green.py`, `scripts/test-check-task-build-green.sh`
**Allowed-collateral:** `stats/internal/guard/helpers_test.go`
**Tests:** `TestCheckTaskBuildGreen`
**Regression:** fails if any of the harness's 88 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_task_build_green_test.go 2>/dev/null | grep -cE '^func Test' @ bbfe685e -->
**After:** Task 1
**Commit:** `feat(guard): port check-task-build-green to Go`
**Build:** green

**Decision:** scope-seven-guards
**Decision:** shims-export-repo-root

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-task-build-green.sh`, one
    subtest per `ok:` label, the `CHECK_TASK_BUILD_GREEN_ROOT` scans included. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-task-build-green`: the one-file check and the
    no-argument scan plus `unlanded-archives` in-process, reusing `planparse.go`'s grammar.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestCheckTaskBuildGreen$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckTaskBuildGreen/'` — at least 88.
  - [ ] **Step 4: Shim and delete** — shim template with the repo-root lines, code 2;
    `git rm scripts/test-check-task-build-green.sh scripts/check-task-build-green.py`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/check-task-build-green.sh` (no argument) exits and prints as the wrapper did at
    `bbfe685e`, directly and through a symlink.

- [ ] 3. Port plan-dispatch-bundles

**Files:** `stats/internal/guard/plandispatchbundles.go`, `stats/internal/guard/plan_dispatch_bundles_test.go`, `stats/internal/guard/planparse.go`, `scripts/plan-dispatch-bundles.sh`, `scripts/test-plan-dispatch-bundles.sh`
**Allowed-collateral:** `stats/internal/guard/helpers_test.go`
**Tests:** `TestPlanDispatchBundles`
**Regression:** fails if any of the harness's 50 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/plan_dispatch_bundles_test.go 2>/dev/null | grep -cE '^func Test' @ bbfe685e -->
**After:** Task 1, 2
**Commit:** `feat(guard): port plan-dispatch-bundles to Go`
**Build:** green

**Decision:** scope-seven-guards
**Decision:** shared-modules-deleted-last

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-plan-dispatch-bundles.sh`, one
    subtest per `ok:` label, the `PLAN_DISPATCH_BUNDLES_ROOT` scans included. Run — expect failure.
  - [ ] **Step 2: Port**, registering `plan-dispatch-bundles`; `parse_tasks` becomes the
    `planparse.go` twin later tasks reuse.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestPlanDispatchBundles$'
    -count=1 -race -v | grep -c -- '--- PASS: TestPlanDispatchBundles/'` — at least 50.
  - [ ] **Step 4: Shim and delete** — shim template (repo-root lines when the wrapper's no-argument
    scan derives the root from its own location), code 2; `git rm
    scripts/test-plan-dispatch-bundles.sh`. `scripts/plan-dispatch-bundles.py` stays until task 8.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/plan-dispatch-bundles.sh
    spectre/changes/kan-923-flow-improvement-port-the-python-backed-guards/tasks.md` prints what
    `python3 scripts/plan-dispatch-bundles.py` prints on the same file, byte for byte, directly and
    through a symlink.

- [ ] 4. Port plan-dispatch-groups

**Files:** `stats/internal/guard/plandispatchgroups.go`, `stats/internal/guard/plan_dispatch_groups_test.go`, `scripts/plan-dispatch-groups.sh`, `scripts/plan-dispatch-groups.py`, `scripts/test-plan-dispatch-groups.sh`
**Allowed-collateral:** `stats/internal/guard/helpers_test.go`
**Tests:** `TestPlanDispatchGroups`
**Regression:** fails if any of the harness's 14 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/plan_dispatch_groups_test.go 2>/dev/null | grep -cE '^func Test' @ bbfe685e -->
**After:** Task 3
**Commit:** `feat(guard): port plan-dispatch-groups to Go`
**Build:** green

**Decision:** scope-seven-guards

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-plan-dispatch-groups.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [ ] **Step 2: Port**, registering `plan-dispatch-groups`, calling task 3's bundle computation
    in-process; bundle violations pass through unchanged, as the Python did.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestPlanDispatchGroups$'
    -count=1 -race -v | grep -c -- '--- PASS: TestPlanDispatchGroups/'` — at least 14.
  - [ ] **Step 4: Shim and delete** — shim template, code 2; `git rm
    scripts/test-plan-dispatch-groups.sh scripts/plan-dispatch-groups.py`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/plan-dispatch-groups.sh
    spectre/changes/kan-923-flow-improvement-port-the-python-backed-guards/tasks.md` prints what the
    base's Python printed on the same file, directly and through a symlink.

- [ ] 5. Port generate-relocation-comparison

**Files:** `stats/internal/guard/relocationcomparison.go`, `stats/internal/guard/generate_relocation_comparison_test.go`, `scripts/generate-relocation-comparison.sh`, `scripts/generate-relocation-comparison.py`, `scripts/test-generate-relocation-comparison.sh`
**Allowed-collateral:** `stats/internal/guard/helpers_test.go`
**Tests:** `TestGenerateRelocationComparison`
**Regression:** fails if any of the harness's 35 `PASS:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/generate_relocation_comparison_test.go 2>/dev/null | grep -cE '^func Test' @ bbfe685e -->
**After:** Task 3
**Commit:** `feat(guard): port generate-relocation-comparison to Go`
**Build:** green

**Decision:** scope-seven-guards

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-generate-relocation-comparison.sh`,
    one subtest per `PASS:` label (the symlink-path exit-2 cases included). Run — expect failure.
  - [ ] **Step 2: Port**, registering `generate-relocation-comparison`, reusing task 3's
    `parse_tasks` twin and `planparse.go`'s **Files:** parsing; every git call carries the pins the
    Python's did.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run
    '^TestGenerateRelocationComparison$' -count=1 -race -v | grep -c -- '--- PASS:
    TestGenerateRelocationComparison/'` — at least 35.
  - [ ] **Step 4: Shim and delete** — shim template, code 2; `git rm
    scripts/test-generate-relocation-comparison.sh scripts/generate-relocation-comparison.py`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`; the shim run on
    `spectre/changes/archive/kan-300-myflow-reviewing-a-relocation-through-a-diff-is/tasks.md`, a plan
    declaring `**Relocation:** yes`, produces the comparison the base's Python produced on it, byte
    for byte, directly and through a symlink.
    <!-- verified: grep -l 'Relocation:\*\* yes' spectre/changes/archive/*/tasks.md @ bbfe685e -->

- [ ] 6. Port check-task-records

**Files:** `stats/internal/guard/taskrecords.go`, `stats/internal/guard/check_task_records_test.go`, `stats/internal/guard/planparse.go`, `scripts/check-task-records.sh`, `scripts/check-task-records.py`, `scripts/test-check-task-records.sh`
**Allowed-collateral:** `stats/internal/guard/helpers_test.go`
**Tests:** `TestCheckTaskRecords`
**Regression:** fails if any of the harness's 23 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_task_records_test.go 2>/dev/null | grep -cE '^func Test' @ bbfe685e -->
**After:** Task 1, 2, 3
**Commit:** `feat(guard): port check-task-records to Go`
**Build:** green

**Decision:** scope-seven-guards
**Decision:** shims-export-repo-root

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-task-records.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-task-records`: `unlanded-archives` in-process in place
    of `CHECK_TASK_RECORDS_ARCHIVED`, the commit-fields `check_files` and parse twins from
    `planparse.go` / `taskcommitfields.go`.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestCheckTaskRecords$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckTaskRecords/'` — at least 23.
  - [ ] **Step 4: Shim and delete** — shim template (repo-root lines when the wrapper derived the
    root from its own location), code 2; `git rm scripts/test-check-task-records.sh
    scripts/check-task-records.py`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/check-task-records.sh .` exits and prints as the wrapper did at `bbfe685e`, directly
    and through a symlink.

- [ ] 7. Port check-markdown-integrity

**Files:** `stats/internal/guard/markdownintegrity.go`, `stats/internal/guard/check_markdown_integrity_test.go`, `scripts/check-markdown-integrity.sh`, `scripts/check-markdown-integrity.py`, `scripts/test-check-markdown-integrity.sh`, `.flow/project.md`
**Allowed-collateral:** `stats/internal/guard/helpers_test.go`
**Tests:** `TestCheckMarkdownIntegrity`
**Regression:** fails if any of the harness's 25 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_markdown_integrity_test.go 2>/dev/null | grep -cE '^func Test' @ bbfe685e -->
**After:** none
**Commit:** `feat(guard): port check-markdown-integrity to Go`
**Build:** green

**Decision:** scope-seven-guards
**Decision:** markdown-integrity-shim

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-markdown-integrity.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-markdown-integrity`.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestCheckMarkdownIntegrity$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckMarkdownIntegrity/'` — at least 25.
  - [ ] **Step 4: Shim and delete** — new `scripts/check-markdown-integrity.sh` from the shim
    template (`chmod +x`), its header the `.py` docstring's contract paragraph; `.flow/project.md`'s
    `## lint` line `scripts/check-markdown-integrity.py` becomes `scripts/check-markdown-integrity.sh`;
    `git rm scripts/test-check-markdown-integrity.sh scripts/check-markdown-integrity.py`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/check-markdown-integrity.sh` on this tree exits and prints as the base's `.py` did,
    directly and through a symlink; `scripts/check-guard-symlinks.sh` exits 0.

- [ ] 8. Delete the shared Python modules and repoint their citations

**Files:** `scripts/plan-dispatch-bundles.py`, `scripts/lib/plan_grammar.py`, `scripts/check-task-commit-fields.py`, `.flow/project.md`
**Allowed-collateral:** `.flow/*.md`, `scripts/*.sh`, `scripts/lib/*.sh`, `hooks/*.py`, `mods/**/*.ts`, `skills/**/*.md`, `rules/*.mdc`, `README.md`, `CONTRIBUTING.md`, `stats/internal/guard/*.go`, `stats/cmd/flow/*.go`
**Tests:** none — deletion and citation sweep; the lint guards are the check
**Regression:** none — deleted modules no code imports, prose and comments only
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7
**Commit:** `refactor(scripts): delete the shared Python plan modules`
**Build:** green

**Decision:** shared-modules-deleted-last
**Decision:** carry-prior-port-decisions

  - [ ] **Step 1: Confirm no importer.** `grep -lE 'plan_grammar|check-task-commit-fields\.py|plan-dispatch-bundles\.py' scripts/*.py scripts/lib/*.py`
    prints only the three files themselves; any other hit is a port this plan missed — stop and
    report.
  - [ ] **Step 2: Delete.** `git rm scripts/plan-dispatch-bundles.py scripts/lib/plan_grammar.py
    scripts/check-task-commit-fields.py`.
  - [ ] **Step 3: Repoint.** `git grep -n -F -e plan_grammar.py -e check-task-commit-fields.py -e
    plan-dispatch-bundles.py -e check-plan-shape.py -e check-task-build-green.py -e
    check-task-records.py -e check-markdown-integrity.py -e plan-dispatch-groups.py -e
    generate-relocation-comparison.py -e test-check-plan-shape.sh -e test-check-task-build-green.sh
    -e test-check-task-records.sh -e test-plan-dispatch-bundles.sh -e
    test-check-markdown-integrity.sh -e test-plan-dispatch-groups.sh -e
    test-generate-relocation-comparison.sh -- ':!spectre/changes/archive' ':!docs/self-review'
    ':!spectre/changes/kan-923-*'`; repoint each citation to the Go file or test that now holds
    what it cites — a Go port's own "Moved from"/history comment naming the Python it ports stays.
    `.flow/project.md`'s Bash + Python + Go paragraph lists the seven new Go guards and drops the
    `.py` examples it names.
    `unverified: the file set is known only after tasks 1–7 land; widen **Files:** by a correction if a hit falls outside the collateral globs`
  - [ ] **Step 4: Verify.** Step 3's grep prints only the Go ports' own history comments; every
    guard in `.flow/project.md`'s `## lint` exits 0; `cd stats && make -s web-dist && go vet ./...
    && gofmt -l .` is clean.
