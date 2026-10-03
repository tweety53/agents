# kan-873-agents-port-the-next-five-slowest-bash-guards-to

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Clears the two base reds (tasks 1–2), ports five bash guards into `flow-guard` (tasks 3–7), sweeps
the citations of the deleted files (task 8) and records the after measurement (task 9). `design.md`
is canonical for every decision; each task cites its entry by ID.

**The script at `9cd35da8` is each port's specification.** Its header comment states the
contract; its body is the behaviour. A port reproduces arguments, environment overrides, every
output line byte for byte, the stdout/stderr split, every side effect on git/the filesystem, and
every exit code. Where the source and its header disagree, stop and report — never pick one
silently.

**Parity floors** are in `design.md`'s **Measurements** (**Decision:**
carry-prior-port-decisions — KAN-760's `parity-by-case-count`).

**Every port task (3–7) follows the same five steps:** port the harness's cases to a Go table test
first (red — the guard is unregistered), port the script, run the test green, replace the script's
body with the shim, `git rm` the harness. Each Go subtest is named after the harness case's `ok:`
label, so parity is a `--- PASS` count. The script body's comments move into the Go file beside the
code they explain, updated where the mechanism changed (KAN-760's
`flow-guard-binary-rationale-in-go`). The Go file is
`stats/internal/guard/<name without - and without check->.go`; the test file is
`stats/internal/guard/<name with - replaced by _>_test.go` — the only name
`scripts/run-guard-tests.sh`'s companion rule accepts for a `check-*` shim, used for the
non-`check-` scripts too for one convention. Each port registers its basename in `guard.Registry`
from its own file's `init()`, as `basemoved.go` does.

**Shim template** — the header comment block kept verbatim, corrected only where the port makes a
statement false (`$SCRIPT_DIR/<lib>` sourcing, bash-specific plumbing), then:

```bash verified:copied from the tail of scripts/check-references.sh @ 9cd35da8
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "<name>: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit <code>
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec <name> <code> "<name>:" "$@"
```

`<name>` is the basename without `.sh`, written literally. `<code>` is the script's cannot-answer
exit: `4` for `break-and-prove`, `2` for the other four (`design.md` **Context**). The two
`FLOW_GUARD_REPO_ROOT` lines are kept only where the bash derived a root or sibling path from its
own location (`$SCRIPT_DIR/..`, `$SCRIPT_DIR/<sibling>`); a script that reads only its arguments
and cwd drops them. Where the bash resolved its own symlink before deriving that root, the shim
exports `FLOW_GUARD_SELF="${BASH_SOURCE[0]}"` instead and the Go side resolves it with
`guardSelfDir`, as KAN-842's `check-workspace-isolation` does.

**Git config pins.** Every git call a port makes carries the pins its bash call carried (R1–R3 of
`scripts/test-git-config-pins.sh`, whose case 7 audits `stats/internal/guard/*.go`); git runs
through `gitExec` (`postmutationcheck.go`) or the port's existing equivalent.

**Test isolation:** every Go test calls `t.Parallel()`; fixture trees (git repos included) are
built once in `TestMain` or a `sync.Once` helper (`helpers_test.go` has the existing ones) and each
case copies its own into `t.TempDir()`; deadlines are injected on `Env`, never through the
environment (KAN-760's `inject-deadlines-in-process`). No test reads or writes the operator's real
`$HOME`, and no test pushes anywhere but a bare repository in `t.TempDir()`.

**Per-task verify** is `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
./internal/guard/ -run '^<Test>$' -count=1 -race`, then the shimmed script run once for real as
its step 5 names. The full package and `scripts/run-guard-tests.sh` run in task 9 and
`flow.verify`.

Live verification: task 9 runs the real suite on this machine and records the after figures
beside design.md's **Before**.

## Review Focus

- A shim run from an installed skill directory (`~/.claude/skills/flow/scripts/<name>.sh`, a
  symlink) must resolve the same repository root the bash did — each port's step 5 runs the shim
  through a symlink in a temp directory where the bash also worked there.
- `land-self-review-report` commits and pushes: its branch re-assertions run before every git
  write in the bash's order, and every refusal exit leaves the index, `HEAD` and the remote exactly
  as the bash did — pinned by subtests on a bare remote.
- `break-and-prove` mutates and restores a file: the restore runs on every exit path past the
  mutation, exit 3 takes precedence over exit 2 as the header states, and the tree-drift check is
  `postmutationcheck.go`'s.
- `check-self-review-report`'s frozen five-angle list and its `find`-error-is-exit-2 rule keep the
  bash's verdicts; the angle table is read through `coverage.go`.
- Locale-sensitive ordering: any `sort`/`sort -u`/`comm` in a bash body keeps the caller's
  collation, pinned by a subtest with mixed-case names.

---

- [ ] 1. Route flow lesson resolve through the record-family seam

**Files:** `stats/cmd/flow/lesson.go`, `stats/cmd/flow/lesson_test.go`, `.flow/project.md`
**Tests:** `TestRunLessonResolveRefusesDirFlag`
**Regression:** `TestRunLessonResolveRefusesDirFlag` fails when `-C` is accepted again; reverting the
commit also turns `scripts/test-flow-addr-declaration.sh`'s `pass-unmutated` red.
**Baseline:** before=5 after=6
<!-- measured: cat stats/cmd/flow/lesson_test.go | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `fix(stats): register flow lesson resolve's address through the record seam`
**Build:** green

**Decision:** lesson-resolve-through-record-seam

  - [ ] **Step 1: Reproduce** before touching anything: `bash scripts/test-flow-addr-declaration.sh`
    prints `FAIL pass-unmutated: … found 2` and exits 1.
    <!-- measured: bash scripts/test-flow-addr-declaration.sh, exit 1, FAIL pass-unmutated @ 9cd35da8 -->
  - [ ] **Step 2: Failing test.** `TestRunLessonResolveRefusesDirFlag`: `runLessonResolve` with
    `-topic x -C /tmp` exits 2, stderr carries `flow: lesson resolve takes no -C`, and no store call
    is made. Run — expect failure.
  - [ ] **Step 3: Fix.** Replace lesson.go's own `-addr`/`-timeout` registrations with a
    `recordIdentityFlags` value registered through `registerRecordConnFlags` (a column-1-tab call
    line, as the harness's assertion 3 reads it); after parsing, a non-empty `dir` prints the line
    above plus `lessonUsage` and exits 2. Rewrite the comment that explained the old registration.
    In `.flow/project.md` `## workspace isolation`, the sentence becomes "The record family
    (`flow record`, `flow self-review bundle`, `flow lesson resolve`) resolves …", the parenthesis
    and verb kept on one physical line.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./cmd/flow/ && go test ./cmd/flow/
    -run '^TestRunLessonResolve' -count=1 -race`; `bash scripts/test-flow-addr-declaration.sh`
    exits 0 with six `ok` lines; `scripts/check-references.sh` exits 0.

- [ ] 2. Fix the SIGPIPE race in test-make-build.sh

**Files:** `scripts/test-make-build.sh`
**Tests:** none — repairs existing tests: `build target still compiles every package` and every other `assert_recipe_matches`/`assert_makefile_matches_all` case of scripts/test-make-build.sh
**Regression:** none — harness-internal; step 3 records what the race looked like
**Baseline:** before=0 after=0
<!-- predicted: no Go test is added by this task -->
**After:** none
**Commit:** `fix(scripts): read the make recipe without a pipe in test-make-build.sh`
**Build:** green

  - [ ] **Step 1: Reproduce** before touching anything: run 2 of design.md's base ranking failed
    `build target still compiles every package` while printing a recipe that contains `go build
    ./...`. Mechanism: under `set -o pipefail`, `printf '%s\n' "$RECIPE" | grep -Eq` lets `grep -q`
    exit on its first match while `printf` is still writing; `printf` dies of SIGPIPE, the pipeline
    returns 141, and the `if` reads a match as a miss. Show it: a loop of 2000 iterations of
    `set -o pipefail; printf '%s\n' "$RECIPE" | grep -Eq '<case 3 pattern>'` with `RECIPE` padded
    to 200 KiB after the matching line counts non-zero exits; record the count.
    `unverified: the count on this machine; a zero count at 200 KiB means raise the padding until the race shows, never treat zero as proof`
  - [ ] **Step 2: Fix** both `grep -q` sites — `assert_recipe_matches` and
    `assert_makefile_matches_all` — to read from a here-string (`grep -Eq "$pattern" <<<"$RECIPE"`,
    `<<<"$MAKEFILE_TEXT"`), no pipe; `assert_makefile_lacks` reads the whole input and stays.
    Comment the reason beside `assert_recipe_matches`.
  - [ ] **Step 3: Prove the seam.** Step 1's loop with the here-string form: zero non-zero exits;
    with the pipe form restored: step 1's count again. Record both counts in the commit body.
  - [ ] **Step 4: Verify.** `bash scripts/test-make-build.sh` exits 0, ten runs in a row.

- [ ] 3. Port check-panel-docs-only

**Files:** `stats/internal/guard/paneldocsonly.go`, `stats/internal/guard/check_panel_docs_only_test.go`, `scripts/check-panel-docs-only.sh`, `scripts/test-check-panel-docs-only.sh`
**Tests:** `TestCheckPanelDocsOnly`
**Regression:** fails if any of the harness's 23 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_panel_docs_only_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port check-panel-docs-only to Go`
**Build:** green

**Decision:** scope-five-next-guards

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-panel-docs-only.sh`, one
    subtest per `ok:` label; fixtures are small git repos in `t.TempDir()`. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-panel-docs-only`; the touched-path set through
    `paneltouchedpaths.go`, never re-derived.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckPanelDocsOnly$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckPanelDocsOnly/'` — at least 23.
  - [ ] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines dropped
    unless step 2 needs a sibling; `git rm scripts/test-check-panel-docs-only.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/check-panel-docs-only.sh
    "$PWD" 9cd35da8` on this tree exits 1 printing the first non-doc path, directly and through a
    symlink to it in a temp directory; with no arguments it exits 2 with the usage line it printed
    at `9cd35da8`.

- [ ] 4. Port prepare-workspace

**Files:** `stats/internal/guard/prepareworkspace.go`, `stats/internal/guard/prepare_workspace_test.go`, `scripts/prepare-workspace.sh`, `scripts/test-prepare-workspace.sh`
**Tests:** `TestPrepareWorkspace`
**Regression:** fails if any of the harness's 25 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/prepare_workspace_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port prepare-workspace to Go`
**Build:** green

**Decision:** scope-five-next-guards

**Decision:** prepare-workspace-execs-sibling

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-prepare-workspace.sh`, one
    subtest per `ok:` label; the cases that replace or remove the sibling guard build their own
    scripts directory in `t.TempDir()` and point the guard at it the way the shim's
    `FLOW_GUARD_REPO_ROOT`/`FLOW_GUARD_SELF` would. Run — expect failure.
  - [ ] **Step 2: Port**, registering `prepare-workspace`; the workspace id and every derived value
    through the Go code `check-workspace-isolation` and `stats/cmd/flow/workspaceid.go` already
    share where one exists, else ported from the bash; `check-workspace-isolation.sh` exec'd beside
    the shim, its stdout relayed verbatim and its stderr inherited, exit codes mapped exactly as the
    header states (126 included).
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestPrepareWorkspace$' -count=1 -race
    -v | grep -c -- '--- PASS: TestPrepareWorkspace/'` — at least 25.
  - [ ] **Step 4: Shim and delete** — shim template, code 2, the sibling resolved as the bash did
    (`FLOW_GUARD_REPO_ROOT`, or `FLOW_GUARD_SELF` if the bash resolved its own symlink);
    `git rm scripts/test-prepare-workspace.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/prepare-workspace.sh
    "$PWD"` on this worktree prints the same `KEY=value` lines and exit code it printed at
    `9cd35da8`, directly and through a symlink to it in a temp directory.

- [ ] 5. Port break-and-prove and delete lib/post-mutation-check.sh

**Files:** `stats/internal/guard/breakandprove.go`, `stats/internal/guard/break_and_prove_test.go`, `stats/internal/guard/post_mutation_check_test.go`, `stats/internal/guard/libtwins_test.go`, `stats/internal/guard/postmutationcheck.go`, `scripts/break-and-prove.sh`, `scripts/test-break-and-prove.sh`, `scripts/lib/post-mutation-check.sh`, `scripts/test-lib-post-mutation-check.sh`
**Tests:** `TestBreakAndProve`, `TestPostMutationCheck`
**Regression:** `TestBreakAndProve` fails if any of the harness's 38 `ok:` behaviours regress;
`TestPostMutationCheck` fails if any of `test-lib-post-mutation-check.sh`'s 4 behaviours regress.
**Baseline:** before=8 after=9
<!-- measured: cat stats/internal/guard/break_and_prove_test.go stats/internal/guard/post_mutation_check_test.go stats/internal/guard/libtwins_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port break-and-prove to Go and drop the bash post-mutation library`
**Build:** green

**Decision:** scope-five-next-guards

**Decision:** delete-post-mutation-check-lib

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-break-and-prove.sh`, one subtest
    per `ok:` label; port the 4 cases of `scripts/test-lib-post-mutation-check.sh` into
    `TestPostMutationCheck` in `post_mutation_check_test.go`, each pinning `snapshotTreeState` /
    `checkTreeRestored` output, stderr and status as fixed expected values (the bash library's
    output on the same fixture, captured once at `9cd35da8` and written into the test). Run —
    expect `TestBreakAndProve` to fail.
  - [ ] **Step 2: Port**, registering `break-and-prove`; tree snapshot and restore check through
    `postmutationcheck.go`; `--clean`, `--sed` and `--patch` as the bash parses them; the restore
    on every exit path past the mutation; exit 3 over exit 2 as the header states; 126/127 from
    the test command mapped to 4.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^(TestBreakAndProve|TestPostMutationCheck)$'
    -count=1 -race -v | grep -c -- '--- PASS: TestBreakAndProve/'` — at least 38, and
    `TestPostMutationCheck` passes all 4.
  - [ ] **Step 4: Shim and delete** — shim template, code 4; remove `TestPostMutationCheckParity`
    from `libtwins_test.go`; correct `postmutationcheck.go`'s header, which names the bash library
    as the source of truth; `git rm scripts/test-break-and-prove.sh
    scripts/lib/post-mutation-check.sh scripts/test-lib-post-mutation-check.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `mutate_and_verify_test.go`'s
    `bashAtBase` read of `lib/post-mutation-check.sh` still passes (it reads the file at the base
    commit, not the tree); `scripts/break-and-prove.sh` with no arguments exits 4 with the usage
    line it printed at `9cd35da8`, directly and through a symlink to it in a temp directory.

- [ ] 6. Port land-self-review-report

**Files:** `stats/internal/guard/landselfreviewreport.go`, `stats/internal/guard/land_self_review_report_test.go`, `scripts/land-self-review-report.sh`, `scripts/test-land-self-review-report.sh`
**Tests:** `TestLandSelfReviewReport`
**Regression:** fails if any of the harness's 43 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/land_self_review_report_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port land-self-review-report to Go`
**Build:** green

**Decision:** scope-five-next-guards

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-land-self-review-report.sh`, one
    subtest per `ok:` label; each case's remote is a bare repository in `t.TempDir()`. Run — expect
    failure.
  - [ ] **Step 2: Port**, registering `land-self-review-report`; the chain in the header's order —
    branch re-assertions before the add, before the commit and before the pull/push pair; the
    foreign-staged refusal (exit 3); git's own exit code passed through unmasked; every
    `LAND-*` line byte for byte; the bash's `g()` wrapper pins carried onto each git call.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestLandSelfReviewReport$' -count=1
    -race -v | grep -c -- '--- PASS: TestLandSelfReviewReport/'` — at least 43.
  - [ ] **Step 4: Shim and delete** — shim template, code 2; `git rm
    scripts/test-land-self-review-report.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `bash
    scripts/test-git-config-pins.sh` and `bash scripts/test-protect-main-checkout.sh` exit 0;
    `scripts/land-self-review-report.sh` with no arguments exits 2 with the usage line it printed
    at `9cd35da8`, directly and through a symlink to it in a temp directory.

- [ ] 7. Port check-self-review-report

**Files:** `stats/internal/guard/selfreviewreport.go`, `stats/internal/guard/check_self_review_report_test.go`, `scripts/check-self-review-report.sh`, `scripts/test-check-self-review-report.sh`
**Tests:** `TestCheckSelfReviewReport`
**Regression:** fails if any of the harness's 63 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_self_review_report_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port check-self-review-report to Go`
**Build:** green

**Decision:** scope-five-next-guards

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-self-review-report.sh`,
    one subtest per `ok:` label; fixtures are small `docs/self-review/` trees in `t.TempDir()`. Run
    — expect failure.
  - [ ] **Step 2: Port**, registering `check-self-review-report`; the angle table through
    `coverage.go`; the frozen `five-angle-reports.txt` list; an unreadable list or a failing
    directory walk is exit 2, never an empty set; the default directory resolved from the repo root
    as the bash's `$SCRIPT_DIR/..` did.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckSelfReviewReport$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckSelfReviewReport/'` — at least 63.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-self-review-report.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-self-review-report.sh` on this tree exits with the code and output it printed at
    `9cd35da8`, directly and through a symlink to it in a temp directory.

- [ ] 8. Repoint citations of the deleted files

**Files:** `.flow/project.md`
**Allowed-collateral:** `.flow/*.md`, `scripts/*.sh`, `scripts/lib/*.sh`, `scripts/*.py`, `hooks/*.py`, `skills/**/*.md`, `rules/*.mdc`, `README.md`, `CONTRIBUTING.md`, `stats/internal/guard/*.go`
**Tests:** none — citation sweep; the lint guards are the check
**Regression:** none — prose and comments only
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7
**Commit:** `docs(scripts): repoint citations of the ported scripts to their Go sources`
**Build:** green

**Decision:** carry-prior-port-decisions

  - [ ] **Step 1: Find.** `grep -rlF -e test-check-panel-docs-only.sh -e test-prepare-workspace.sh
    -e test-break-and-prove.sh -e test-land-self-review-report.sh -e
    test-check-self-review-report.sh -e post-mutation-check.sh --exclude-dir=archive
    --exclude-dir=.worktrees --exclude-dir=node_modules --exclude-dir=.git
    --exclude-dir=self-review .`; then grep the same tree for citations of a ported script's body
    comments and of its bash plumbing (`sources lib/…`).
    `unverified: the file set is known only after tasks 3–7 land; widen **Files:** by a correction if a hit falls outside the collateral globs`
  - [ ] **Step 2: Repoint** each citation to the Go file or Go test that now holds what it cites; a
    sentence describing bash plumbing that no longer exists is corrected, not repointed.
    `.flow/project.md`'s paragraph naming which guards are Go lists the five new ones; library
    headers naming a ported script as a caller that sources them are corrected.
  - [ ] **Step 3: Verify.** Step 1's grep returns only `spectre/changes/kan-873-*`,
    `docs/self-review/` and the Go ports' own history comments; every guard in `.flow/project.md`'s
    `## lint` exits 0.

- [ ] 9. Live verification: after timings and parity

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8
**Build:** green

  - [ ] **Step 1: Suite, after.** `FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh`,
    three times on the branch head; record each run's wall, exit, harness count and slowest five.
  - [ ] **Step 2: Parity.** `cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '---
    PASS: Test<Name>/'` per port against its floor in `design.md`.
  - [ ] **Step 3: Record** an **After** table beside design.md's **Before**, same columns, each
    figure tagged `measured:` with the command and
    `@ branch spectre/kan-873-agents-port-the-next-five-slowest-bash-guards-to`; name the slowest
    remaining bash harness and the next slice.
  - [ ] **Step 4: Judge.** Failure looks like: any harness red; any port's `--- PASS` count below
    its floor; the suite median above the Before median. Any of these is reported, not recorded as
    success.
