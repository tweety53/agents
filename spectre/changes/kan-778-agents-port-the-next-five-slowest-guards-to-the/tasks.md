# kan-778-agents-port-the-next-five-slowest-guards-to-the

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Ports five more guards into `flow-guard`, in dependency order: the before measurement (task 1),
the `coverage.sh` Go twin two ports share (task 2), the five ports (tasks 3–7), the citation
sweep for the deleted files (task 8), and the after measurement (task 9). `design.md` is
canonical for every decision; each task cites its entry by ID.

**The guard at `3e48ecac` is each port's specification.** Its header comment states the contract;
its body is the behaviour. A port reproduces arguments, environment overrides, every verdict line
byte for byte, the stdout/stderr split and every exit code. Where the source and its header
disagree, stop and report — never pick one silently.

**Parity floors** (`ok` lines on a green harness run at base) are in `design.md`'s
**Measurements** (**Decision:** carry-kan-760-port-decisions — KAN-760's `parity-by-case-count`).

**Every port task (3–7) follows KAN-760's five-step shape:** port the harness's cases to a Go
table test first (red — the guard is unregistered), port the guard, run the test green, replace
the script's body with the shim, `git rm` the harness. Each Go subtest is named after the harness
case's `ok:` label, so parity is a `--- PASS` count. The script body's comments move into the Go
file beside the code they explain, updated where the mechanism changed (KAN-760's
`flow-guard-binary-rationale-in-go`). The test file is
`stats/internal/guard/<name with - replaced by _>_test.go` — the only name
`scripts/run-guard-tests.sh`'s companion rule accepts for a shim guard.

**Shim template** — the header comment block kept verbatim, corrected only where the port makes a
statement false (python3 probes, `$SCRIPT_DIR/<lib>` sourcing), then:

```bash verified:copied from the tail of scripts/check-panel-reproducer-exit-contract.sh @ 3e48ecac
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "<name>: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec <name> 2 "<name>:" "$@"
```

`<name>` is the guard's basename without `.sh`, written literally. All five use cannot-answer
code 2.
<!-- verified: each port's implementer read its guard's header at 3e48ecac — 2 is the cannot-answer exit for all five (check-plan-provenance's 3 and 4 are verdict codes) -->

**Test isolation:** every Go test calls `t.Parallel()`; fixture trees are built once in
`TestMain` (or a `sync.Once` helper) and each case copies its own into `t.TempDir()`; a stub
`flow` on PATH becomes `Env.Findings` where the guard reads findings; deadlines are injected on
`Env`, never through the environment (KAN-760's `inject-deadlines-in-process`).

Live verification: tasks 1 and 9 run the real suite on this machine and record before/after.

---

- [x] 1. Live verification: before timings

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** none
**Build:** green

**Decision:** suite-median-below-before

  - [x] **Step 1: Suite, before.** On this machine at `3e48ecac`, nothing else heavy running:
    `sysctl -n vm.loadavg` then `FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p
    scripts/run-guard-tests.sh`, three times; record each run's real/user/sys, load, harness
    count and the slowest five harnesses.
  - [x] **Step 2: Go package, before.** `cd stats && /usr/bin/time -p go test
    ./internal/guard/... -count=1` three times; record real/user/sys.
  - [x] **Step 3: Record** a **Before** table under `design.md`'s **Measurements** → **Suite
    before/after**, each figure tagged `measured:` with the command and `@ 3e48ecac`.

This task commits nothing; its figures are committed with the change's artifacts.

- [x] 2. coverage.sh's Go twin

**Files:** `stats/internal/guard/coverage.go`, `stats/internal/guard/coverage_test.go`
**Tests:** `TestCoverageParity`
**Regression:** `TestCoverageParity` fails if `coverage.go`'s rendered members-and-counts
fragment or its declared/undeclared-zero verdict differs from `scripts/lib/coverage.sh`'s for the
same inputs.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/coverage_test.go 2>/dev/null | grep -cE '^func Test' @ 3e48ecac -->
**After:** Task 1
**Commit:** `feat(stats): add the Go twin of lib/coverage.sh`
**Build:** green

**Decision:** coverage-go-twin

  - [x] **Step 1: Failing test.** `TestCoverageParity`: a table of member sets — none recorded,
    all non-zero, an undeclared zero, a declared zero, several members in and out of declaration
    order. Each row runs `bash -c '. scripts/lib/coverage.sh; coverage_declare …;
    coverage_record …; coverage_report; coverage_verdict …'` once and the Go functions in-process,
    and compares output and verdict byte for byte. Run `cd stats && go test ./internal/guard/ -run
    TestCoverageParity -count=1` — expect a compile failure.
  - [x] **Step 2: Port** `coverage_declare`/`coverage_record`/`coverage_report`/
    `coverage_verdict` as a small `coverage` type in `coverage.go`, its header citing
    `scripts/lib/coverage.sh` as the source of truth it mirrors.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
    ./internal/guard/ -run TestCoverageParity -count=1 -race`.

- [ ] 3. Port check-panel-reproducers

**Files:** `stats/internal/guard/panelreproducers.go`, `stats/internal/guard/check_panel_reproducers_test.go`, `scripts/check-panel-reproducers.sh`, `scripts/test-check-panel-reproducers.sh`, `skills/flow/scripts/reproducer-metachars.sh`, `stats/internal/guard/panelexitcontract.go`
**Tests:** `TestCheckPanelReproducers`
**Regression:** fails if any of the harness's 55 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_panel_reproducers_test.go 2>/dev/null | grep -cE '^func Test' @ 3e48ecac -->
**After:** Task 1
**Commit:** `feat(stats): port check-panel-reproducers to Go`
**Build:** green

**Decision:** scope-five-next-guards

  - [ ] **Step 1: Failing test.** Port every case, including the metacharacter loop, one subtest
    per `ok:` label; the banned set comes from `metachars.go` (already parity-tested against
    `scripts/reproducer-metachars.sh`). Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-panel-reproducers`. It calls the run-reproducer Go
    function in-process instead of exec'ing `$SCRIPT_DIR/run-reproducer.sh`, as
    `panelexitcontract.go` does; findings via `Env.Findings`, nil → exec `flow record findings`.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run
    '^TestCheckPanelReproducers$' -count=1 -race -v | grep -c -- '--- PASS:
    TestCheckPanelReproducers/'` — at least 55.
  - [ ] **Step 4: Shim and delete** — shim template; `git rm scripts/test-check-panel-reproducers.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/check-guard-symlinks.sh`.

Correction (2026-09-27): Step 2 declared the port calls the run-reproducer Go function in-process;
the guard at `3e48ecac` never runs run-reproducer — it sources `reproducer-metachars.sh` and checks each
reproducer lexically. Shipped: `metachars.go`'s `reproducerMetachars` in place of sourcing
`$SCRIPT_DIR/reproducer-metachars.sh`; findings via `Env.Findings`, nil → exec `flow record findings`.
The shim no longer needs `reproducer-metachars.sh`, so `scripts/check-guard-symlinks.sh` rule 6 flagged
`skills/flow/scripts/reproducer-metachars.sh` as dead weight and the commit deletes it.

Correction (2026-09-27): the guard's header and body disagreed on findings output the real
`flow record findings` never emits — `[null]`, `{}`, empty stdout at exit 0, two values, a ref-less
finding: the body passed several at exit 0 (`REPRODUCERS-OK`), which the header's exit-0 contract rules
out. The operator decided: refuse every shape that is not one array of objects with exit 2 (`jq failed —
cannot determine anything`), in this guard and check-unfinished-work alike; subtests pin each shape, and
`pcParseFindings`' doc comment in `panelexitcontract.go` states it — hence that path is added to
`**Files:**`. Case 20's label is kept for label parity; with the set compiled in, it asserts the set
still binds (exit 1 on `;`), as its comment says.

- [ ] 4. Port check-unfinished-work

**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`, `scripts/test-check-unfinished-work.sh`
**Tests:** `TestCheckUnfinishedWork`
**Regression:** fails if any of the harness's 102 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_unfinished_work_test.go 2>/dev/null | grep -cE '^func Test' @ 3e48ecac -->
**After:** Task 1
**Commit:** `feat(stats): port check-unfinished-work to Go`
**Build:** green

**Decision:** scope-five-next-guards

  - [ ] **Step 1: Failing test.** Port every case, one subtest per `ok:` label; the stub `flow`
    becomes `Env.Findings` (an error for the store-unreachable case, which must exit 2). Run —
    expect failure.
  - [ ] **Step 2: Port**, registering `check-unfinished-work`; plan resolution through
    `changeplan.go` and `specroot.go`, never a second copy. The optional third argument
    (canonical worktree) and every verdict line kept.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckUnfinishedWork$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckUnfinishedWork/'` — at least 102.
  - [ ] **Step 4: Shim and delete** — shim template; `git rm scripts/test-check-unfinished-work.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/check-guard-symlinks.sh`.

Correction (2026-09-27): findings output that is not one array of objects — empty stdout at exit 0, an
object, a null element, two values — is refused with exit 2 (`jq failed — cannot determine anything`),
the operator's decision recorded in task 3's Correction. The bash body read empty stdout and `{}` as zero
findings (CLEAR), an object's values as findings, and two values as CLEAR. The jq/grep diagnostic lines
bash printed before its own refusal on four paths are not reproduced (KAN-760's `pcJQFailed` idiom);
exit code, stdout and the guard's own lines match.

- [ ] 5. Port check-references

**Files:** `stats/internal/guard/references.go`, `stats/internal/guard/check_references_test.go`, `scripts/check-references.sh`, `scripts/test-check-references.sh`, `stats/internal/guard/guard.go`, `stats/cmd/flow-guard/main.go`
**Tests:** `TestCheckReferences`
**Regression:** fails if any of the harness's 41 `ok` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_references_test.go 2>/dev/null | grep -cE '^func Test' @ 3e48ecac -->
**After:** Task 2
**Commit:** `feat(stats): port check-references to Go`
**Build:** green

**Decision:** scope-five-next-guards
**Decision:** coverage-go-twin

  - [ ] **Step 1: Failing test.** Port every case, one subtest per `ok` label; fixtures are
    small trees in `t.TempDir()`, never this repository's own tree. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-references`; per-member coverage through
    `coverage.go`; the association shapes (`is_associated`), the suppression marker and the
    `file:line` report kept byte for byte.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckReferences$' -count=1 -race
    -v | grep -c -- '--- PASS: TestCheckReferences/'` — at least 41.
  - [ ] **Step 4: Shim and delete** — shim template; `git rm scripts/test-check-references.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/check-references.sh`
    on this tree exits 0 with the same member counts it printed at base.

Correction (2026-09-27): the shim template declared `flow_guard_exec` alone after the loader; the
guard defaults its root to its own checkout (`BASH_SOURCE`), which the cached Go binary cannot see,
so the shim also exports `FLOW_GUARD_REPO_ROOT` (`cd "$(dirname …)/.." && pwd`, logical, as the bash
did). Telling a set-but-empty `CHECK_REFERENCES_ROOT` from an unset one needs `Env.LookupEnv`, added to
`guard.go` and wired to `os.LookupEnv` in `stats/cmd/flow-guard/main.go` — hence the widened
`**Files:**`. File and path-set order is collated by exec'ing `sort` under the caller's locale, as the
bash's `sort`/`sort -u` did.

Correction (2026-09-27): an unreadable file in the scan set exited 1 at `3e48ecac` — the read loop's
redirect failing under `set -e`, with only bash's own error on stderr — an accident, not the contract. The
port refuses with 2, this guard's cannot-answer code, and names the file (`check-references: cannot read
<file>: …`); a subtest pins it.

- [ ] 6. Port check-plan-provenance

**Files:** `stats/internal/guard/planprovenance.go`, `stats/internal/guard/check_plan_provenance_test.go`, `scripts/check-plan-provenance.sh`, `scripts/check-plan-provenance.py`, `scripts/test-check-plan-provenance.sh`
**Tests:** `TestCheckPlanProvenance`
**Regression:** fails if any of the harness's 665 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_plan_provenance_test.go 2>/dev/null | grep -cE '^func Test' @ 3e48ecac -->
**After:** Task 1
**Commit:** `feat(stats): port check-plan-provenance to Go`
**Build:** green

**Decision:** scope-five-next-guards

  - [ ] **Step 1: Failing test.** Port every case, one subtest per `ok:` label; group them into
    table-driven subtests by the harness's own sections so the file stays navigable. Run — expect
    failure.
  - [ ] **Step 2: Port** `check-plan-provenance.py` into `planprovenance.go` (split into
    `planprovenance_*.go` files by concern when one file grows hard to navigate), registering
    `check-plan-provenance`. `CHECK_PLAN_PROVENANCE_ROOT`, argv, the stdout/stderr shape and all
    exit codes (0, 1, 2 environment, 3 containment, 4 content-classification) kept. Python `re`
    patterns ported to Go `regexp` only where RE2 matches identically; any pattern using
    lookaround or backreferences is rewritten as code with a subtest pinning it.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckPlanProvenance$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckPlanProvenance/'` — at least 665.
  - [ ] **Step 4: Shim and delete** — shim template (the python3 probe paragraphs in the header
    go); `git rm scripts/check-plan-provenance.py scripts/test-check-plan-provenance.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-plan-provenance.sh` over this change's own `tasks.md` exits 0, as the Python
    guard did at base.

Correction (2026-09-27): the shim template declared `flow_guard_exec` alone after the loader; the
Python guard defaulted its root to its own file's location, which the cached Go binary cannot see, so
the shim resolves the checkout (`cd -P … && pwd`) and exports it only when unset:
`export CHECK_PLAN_PROVENANCE_ROOT="${CHECK_PLAN_PROVENANCE_ROOT-$root}"` — an explicit value, empty
included, passes through. Twelve harness labels were relabelled: three timing and one exit-code label
lose their measured figure, and eight python3-interpreter cases become their missing/broken/working
`go` equivalents run through the real shim (KAN-760's case-56 precedent); one case is added (shim
default root). Every Python pattern using lookaround or Unicode `\s`/`\w` was rewritten as code.

Correction (2026-09-27): the gated review found the Unicode classes of the rewritten matchers unpinned —
20+ mutants (ASCII-only `\s`/`\w`, the `{0,20}` bound, `[A-Z]`, `str.strip()` sites) survived, one failing
open. `ppCasesUnicodeMatchers` adds 88 rows whose exit codes and lines are the Python guard's at `3e48ecac`;
19 of the 20 mutants now fail the suite. The survivor, `ppLeadingSpace` ASCII-only, is equivalent: the
list-gap trial it gates is only used to find a list marker, and a remainder starting with whitespace is
never one. `ppIsWord` reads Go's Unicode tables, so it differs from a newer host Python on characters later
Unicode versions added — only ever by reporting a claim Python would not. `ppRepr` now escapes every
character `str.isprintable()` rejects, pinned against python3's `repr()`.

- [x] 7. Port check-installed-citations

**Files:** `stats/internal/guard/installedcitations.go`, `stats/internal/guard/check_installed_citations_test.go`, `scripts/check-installed-citations.sh`, `scripts/check-installed-citations.py`, `scripts/test-check-installed-citations.sh`
**Tests:** `TestCheckInstalledCitations`
**Regression:** fails if any of the harness's 61 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_installed_citations_test.go 2>/dev/null | grep -cE '^func Test' @ 3e48ecac -->
**After:** Task 2
**Commit:** `feat(stats): port check-installed-citations to Go`
**Build:** green

**Decision:** scope-five-next-guards
**Decision:** coverage-go-twin

  - [x] **Step 1: Failing test.** Port every case, one subtest per `ok:` label. Run — expect
    failure.
  - [x] **Step 2: Port** the Python classifier and the wrapper's coverage handling into one Go
    guard, registering `check-installed-citations`: the sentinel-prefix hand-off between `.py`
    and `.sh` disappears; coverage is decided in-process through `coverage.go`, and the sandbox
    `mkdtemp`/subprocess the classifier runs becomes `os.MkdirTemp`/`exec.Command` with the same
    arguments and cleanup.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckInstalledCitations$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckInstalledCitations/'` — at least 61.
  - [x] **Step 4: Shim and delete** — shim template (the wrapper/sentinel paragraphs in the header
    go); `git rm scripts/check-installed-citations.py scripts/test-check-installed-citations.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-installed-citations.sh` on this tree exits 0 with the same member counts it
    printed at base.

Correction (2026-09-27): the shim exports `FLOW_GUARD_REPO_ROOT` (`cd -P … && pwd`, physical, as the
Python's `realpath` did) beside the template, for the same reason as task 5's; the header's
wrapper/sentinel explanation was replaced by the exit contract from the `.py` docstring. Three harness-only
labels (CHECK PHASE build-order replay, EXIT-trap chaining on clean exit and on SIGINT) tested the
harness's own trap plumbing, not the guard; three rows stand in for them (no sandbox left behind on a clean
run, on a refused run; `setup.sh` runs with HOME and cwd inside the guard's sandbox), keeping the count at 61.

- [x] 8. Repoint citations of the deleted files

**Files:** `.flow/project.md`, `.flow/project-rationale.md`, `.gitignore`, `scripts/check-markdown-integrity.py`, `scripts/check-plan-shape.py`, `scripts/check-task-build-green.py`, `scripts/check-task-build-green.sh`, `scripts/check-task-commit-fields.py`, `scripts/generate-relocation-comparison.py`, `scripts/lib/plan_grammar.py`, `scripts/lib/parallel.sh`, `scripts/plan-dispatch-bundles.py`, `scripts/plan-dispatch-bundles.sh`, `scripts/plan-dispatch-groups.py`, `skills/flow/SKILL.md`, `skills/flow/review-panel.md`, `skills/flow-contracts/SKILL.md`, `skills/flow-contracts/build-green-rationale.md`, `skills/flow-contracts/plan-provenance-guard.md`, `skills/flow-contracts/plan-provenance-guard-rationale.md`, `stats/internal/guard/taskcommitfields.go`
**Allowed-collateral:** `scripts/*.sh`, `scripts/lib/*.sh`, `scripts/*.py`, `skills/**/*.md`, `stats/internal/records/render_test.go`
**Tests:** none — citation sweep; the lint guards are the check
**Regression:** none — prose and comments only
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 3, 4, 5, 6, 7
**Commit:** `docs(scripts): repoint citations of the ported guards to their Go sources`
**Build:** green

**Decision:** carry-kan-760-port-decisions

  - [x] **Step 1: Find.** `grep -rlF -e check-plan-provenance.py -e check-installed-citations.py
    -e test-check-plan-provenance.sh -e test-check-installed-citations.sh -e
    test-check-references.sh -e test-check-unfinished-work.sh -e test-check-panel-reproducers.sh
    --exclude-dir=archive --exclude-dir=.worktrees --exclude-dir=node_modules --exclude-dir=.git
    --exclude-dir=self-review .` — every hit outside this change's own directory is
    declared above: a non-test file in the files field, a harness comment (`scripts/test-*.sh`, `scripts/lib/test-git-shim.sh`,
    `stats/internal/records/render_test.go`) under the collateral globs.
    <!-- measured: this grep, 42 files, 10 of them the port tasks' own, the other 32 in Files: above (plus skills/flow/SKILL.md, step 2) @ 3e48ecac; docs/self-review/ is history, left as is -->
  - [x] **Step 2: Repoint** each citation to the Go file or Go test that now holds what it
    cites (`stats/internal/guard/planprovenance.go`, `installedcitations.go`, the
    `check_*_test.go` files); a sentence describing Python or bash plumbing that no longer exists
    is corrected, not repointed. `.flow/project.md`'s "Bash + Python" paragraph states which
    guards are Go now. `skills/flow/SKILL.md`'s sentence that `check-unfinished-work.sh` needs
    `lib/change-plan.sh` as a sibling is corrected to what the shim needs (`lib/flow-guard.sh`).
  - [x] **Step 3: Verify.** Step 1's grep returns only `spectre/changes/kan-778-*` and
    `docs/self-review/`; every guard in `.flow/project.md`'s `## lint` exits 0.

Correction (2026-09-27): Step 3 declared the grep returns only `spectre/changes/kan-778-*` and
`docs/self-review/`; it also returns the Go ports' own headers and tests, which name the deleted
`.py` files and harnesses as the history they were ported from, pinned to `3e48ecac` (KAN-760's
`flow-guard-binary-rationale-in-go`), and `check_panel_reproducers_test.go`'s fixture string. Those
hits are intended. Beyond Step 1's grep, the sweep also corrected statements the grep cannot match:
libraries and harness comments still naming a ported guard as a caller that sources them.

- [ ] 9. Live verification: after timings and parity

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8
**Build:** green

**Decision:** suite-median-below-before
**Decision:** guard-package-under-20s

  - [ ] **Step 1: Suite, after.** Task 1 step 1's command, three times, on the branch head;
    same fields recorded.
  - [ ] **Step 2: Go package, after.** Task 1 step 2's command, three times.
  - [ ] **Step 3: Parity.** `go test ./internal/guard/ -count=1 -v | grep -c -- '--- PASS:
    Test<Name>/'` per port against its floor in `design.md`.
  - [ ] **Step 4: Record** an **After** table beside **Before**, same columns, each figure tagged
    `measured:` with the command and `@ spectre/kan-778-agents-port-the-next-five-slowest-guards-to-the`;
    name the slowest remaining harness and the next slice's guards.
  - [ ] **Step 5: Judge.** Failure looks like: suite median not below the Before median; any
    port's `--- PASS` count below its floor; the Go package median above 20s real; any harness
    red. Any of these is reported, not recorded as success.
