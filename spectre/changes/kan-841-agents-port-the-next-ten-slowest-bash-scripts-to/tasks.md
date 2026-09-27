# kan-841-agents-port-the-next-ten-slowest-bash-scripts-to

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Ports ten more bash scripts into `flow-guard`, in dependency order: the before measurement
(task 1), the Go twins of three shared bash libraries (task 2), the ten ports (tasks 3–12), the
citation sweep (task 13), and the after measurement (task 14). `design.md` is canonical for every
decision; each task cites its entry by ID.

**The script at `d71a2327` is each port's specification.** Its header comment states the
contract; its body is the behaviour. A port reproduces arguments, environment overrides, every
output line byte for byte, the stdout/stderr split, every side effect on git/the filesystem, and
every exit code. Where the source and its header disagree, stop and report — never pick one
silently.

**Parity floors** are in `design.md`'s **Measurements** (**Decision:**
carry-prior-port-decisions — KAN-760's `parity-by-case-count`).

**Every port task (3–12) follows the same five steps:** port the harness's cases to a Go table
test first (red — the guard is unregistered), port the script, run the test green, replace the
script's body with the shim, `git rm` the harness. Each Go subtest is named after the harness
case's `ok:` label (`case_N` function name for `prove-reproducer`), so parity is a `--- PASS`
count. The script body's comments move into the Go file beside the code they explain, updated
where the mechanism changed (KAN-760's `flow-guard-binary-rationale-in-go`). The Go file is
`stats/internal/guard/<name without - and without check->.go`; the test file is
`stats/internal/guard/<name with - replaced by _>_test.go` — the only name
`scripts/run-guard-tests.sh`'s companion rule accepts for a `check-*` shim, used for the three
tools too for one convention. Each port registers its basename in `guard.Registry` (`guard.go`).

**Shim template** — the header comment block kept verbatim, corrected only where the port makes a
statement false (`$SCRIPT_DIR/<lib>` sourcing, bash-specific plumbing), then:

```bash verified:copied from the tail of scripts/check-references.sh @ d71a2327
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
exit: `4` for `mutate-and-verify`, `2` for the other nine (`design.md` **Context**). The two
`FLOW_GUARD_REPO_ROOT` lines are kept only where the bash derived a root or sibling path from its
own location (`$SCRIPT_DIR/..`, `$SCRIPT_DIR/<sibling>`); a script that reads only its arguments
and cwd drops them.

**Test isolation:** every Go test calls `t.Parallel()`; fixture trees (git repos included) are
built once in `TestMain` or a `sync.Once` helper (`helpers_test.go` has the existing ones) and each
case copies its own into `t.TempDir()`; a stub `flow` on PATH becomes the injected `Env` hook the
existing ports use where the script reads the store; deadlines are injected on `Env`, never
through the environment (KAN-760's `inject-deadlines-in-process`).

**Per-task verify** is `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
./internal/guard/ -run '^<Test>$' -count=1 -race`, then the shimmed script run once for real as
its step 5 names. The full package and `scripts/run-guard-tests.sh` run in task 14 and
`flow.verify`.

Live verification: tasks 1 and 14 run the real suite on this machine and record before/after.

## Review Focus

- A shim run from an installed skill directory (`~/.claude/skills/flow/scripts/<name>.sh`, a
  symlink) must resolve the same repository root the bash did — each port's step 5 runs the shim
  through a symlink in a temp directory.
- Locale-sensitive ordering: any `sort`/`sort -u`/`comm` in a bash body keeps the caller's
  collation (KAN-778 task 5's exec'd `sort`), pinned by a subtest with mixed-case names.
- Scripts that mutate git state (`mutate-and-verify`, `prepare-archive-branch`,
  `prove-reproducer`) leave the tree exactly as the bash did on every exit path — each has a
  subtest per refusal exit asserting `git status --porcelain` and `HEAD` after.
- Paths with spaces in worktree/landing arguments — one subtest per git-mutating script.
- Missing `go` on PATH through the real shim exits the shim's cannot-answer code — pinned for
  `mutate-and-verify` (code 4, the only non-2) by one subtest in task 6, shaped as
  `check_task_commit_fields_test.go`'s "case 56".

---

- [ ] 1. Live verification: before timings

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** none
**Build:** green

**Decision:** suite-median-below-before

  - [ ] **Step 1: Suite, before.** On this machine at `d71a2327`, nothing else heavy running:
    `sysctl -n vm.loadavg` then `FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p
    scripts/run-guard-tests.sh`, three times; record each run's real/user/sys, load, harness
    count and the slowest five harnesses (`grep '(Ns)'` of each log, sorted descending).
  - [ ] **Step 2: Go package, before.** `cd stats && /usr/bin/time -p go test
    ./internal/guard/... -count=1` three times; record real/user/sys.
  - [ ] **Step 3: Record** a **Before** table under `design.md`'s **Measurements** → **Suite
    before/after**, KAN-778's columns, each figure tagged `measured:` with the command and
    `@ d71a2327`.

This task commits nothing; its figures are committed with the change's artifacts.

- [ ] 2. Go twins of resolve-file, project-section and post-mutation-check

**Files:** `stats/internal/guard/resolvefile.go`, `stats/internal/guard/projectsection.go`, `stats/internal/guard/postmutationcheck.go`, `stats/internal/guard/libtwins_test.go`
**Tests:** `TestResolveFileParity`, `TestProjectSectionParity`, `TestPostMutationCheckParity`
**Regression:** each parity test fails if its Go twin's output, return status or side effect
differs from `scripts/lib/resolve-file.sh`, `scripts/lib/project-section.sh` or
`scripts/lib/post-mutation-check.sh` for the same inputs.
**Baseline:** before=0 after=3
<!-- measured: cat stats/internal/guard/libtwins_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** none
**Commit:** `feat(stats): add Go twins of resolve-file, project-section and post-mutation-check`
**Build:** green

**Decision:** shared-helper-go-twins

  - [ ] **Step 1: Failing tests.** Read each library's header for its functions and contract. For
    each, a table of inputs covering every branch of its body (resolve-file: plain file, symlink
    chain, relative symlink, dangling link, directory; project-section: present section, absent,
    empty body, fenced body, heading-level edge, trailing prose; post-mutation-check: clean tree,
    dirty tracked file, new untracked file, drift under a mutation). Each row runs the bash
    library once via `bash -c '. scripts/lib/<lib>.sh; <fn> <args>'` against a fixture in
    `t.TempDir()` and the Go function in-process, comparing stdout, stderr and status byte for
    byte. Run `cd stats && go test ./internal/guard/ -run 'Parity$' -count=1` — expect a compile
    failure.
  - [ ] **Step 2: Port** each library's functions into its Go file, its header citing the bash
    library as the source of truth it mirrors and the parity test that pins them.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
    ./internal/guard/ -run '^(TestResolveFileParity|TestProjectSectionParity|TestPostMutationCheckParity)$'
    -count=1 -race`.

- [ ] 3. Port check-stage-mark-calls

**Files:** `stats/internal/guard/stagemarkcalls.go`, `stats/internal/guard/check_stage_mark_calls_test.go`, `scripts/check-stage-mark-calls.sh`, `scripts/test-check-stage-mark-calls.sh`
**Tests:** `TestCheckStageMarkCalls`
**Regression:** fails if any of the harness's 74 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_stage_mark_calls_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 1
**Commit:** `feat(stats): port check-stage-mark-calls to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-stage-mark-calls.sh`,
    one subtest per `ok:` label; fixtures are small trees in `t.TempDir()`. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-stage-mark-calls`; per-member coverage through
    `coverage.go`; the stage-key set it reads from the stats Go sources (`go` calls in the body)
    read the same way.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckStageMarkCalls$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckStageMarkCalls/'` — at least 74.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-stage-mark-calls.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-stage-mark-calls.sh` on this tree exits 0 with the same output it printed at
    `d71a2327`, directly and through a symlink to it in a temp directory.

- [ ] 4. Port check-guard-symlinks

**Files:** `stats/internal/guard/guardsymlinks.go`, `stats/internal/guard/check_guard_symlinks_test.go`, `scripts/check-guard-symlinks.sh`, `scripts/test-check-guard-symlinks.sh`
**Tests:** `TestCheckGuardSymlinks`
**Regression:** fails if any of the harness's 118 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_guard_symlinks_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 2
**Commit:** `feat(stats): port check-guard-symlinks to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** shared-helper-go-twins

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-guard-symlinks.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-guard-symlinks`; link resolution through
    `resolvefile.go` (task 2), coverage through `coverage.go`; every awk program in the body
    (`CITATION_AWK`, `DELEGATE_AWK`, `RULE3_AWK`) ported as Go code with its rule's subtests
    pinning it. Rule 2's sibling derivation, which greps a guard's source for
    `$SCRIPT_DIR/<name>`, keeps reading the Go source for a shimmed guard, as the bash does since
    KAN-760 — now including the ten scripts this change shims.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckGuardSymlinks$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckGuardSymlinks/'` — at least 118.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-guard-symlinks.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-guard-symlinks.sh` on this tree exits 0 with the same output it printed at
    `d71a2327`, directly and through a symlink to it in a temp directory.

- [ ] 5. Port check-dispatch-paragraphs

**Files:** `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`, `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
**Tests:** `TestCheckDispatchParagraphs`
**Regression:** fails if any of the harness's 183 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_dispatch_paragraphs_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 1
**Commit:** `feat(stats): port check-dispatch-paragraphs to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-dispatch-paragraphs.sh`
    (3542 lines), one subtest per `ok:` label, grouped into table-driven subtests by the harness's
    own sections. Run — expect failure.
    <!-- measured: wc -l < scripts/test-check-dispatch-paragraphs.sh @ d71a2327 -->
  - [ ] **Step 2: Port**, registering `check-dispatch-paragraphs`; the site table (`SITE_PATHS`,
    `SITE_VARIANTS`, `SITE_MIN_BLOCKS`, `ENTRY_*`) becomes a Go table in the same order;
    `CHECK_DISPATCH_PARAGRAPHS_ROOT` and `CHECK_GUARD_SYMLINKS_ROOT` overrides kept, set-but-empty
    told from unset through `Env.LookupEnv`.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckDispatchParagraphs$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckDispatchParagraphs/'` — at least 183.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-dispatch-paragraphs.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-dispatch-paragraphs.sh` on this tree exits 0 with the same output it printed at
    `d71a2327`.

- [ ] 6. Port mutate-and-verify

**Files:** `stats/internal/guard/mutateandverify.go`, `stats/internal/guard/mutate_and_verify_test.go`, `scripts/mutate-and-verify.sh`, `scripts/test-mutate-and-verify.sh`
**Tests:** `TestMutateAndVerify`
**Regression:** fails if any of the harness's 44 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/mutate_and_verify_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 2
**Commit:** `feat(stats): port mutate-and-verify to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** shared-helper-go-twins

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-mutate-and-verify.sh`, one
    subtest per `ok:` label, plus the Review Focus rows (tree and HEAD unchanged after each
    refusal exit; a worktree path with a space; the real shim with no `go` on PATH exits 4).
    Run — expect failure.
  - [ ] **Step 2: Port**, registering `mutate-and-verify`; exit codes 0/2/3/4 per the header;
    `MUTATE_AND_VERIFY_MAX_NEW_FAILURES` kept with its parsing and default; the post-mutation
    drift check through `postmutationcheck.go` (task 2); harnesses are run as child processes
    exactly as the bash ran them (same argv, cwd, inherited environment).
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestMutateAndVerify$' -count=1
    -race -v | grep -c -- '--- PASS: TestMutateAndVerify/'` — at least 44.
  - [ ] **Step 4: Shim and delete** — shim template, code 4 (the loader's own failure exits 4
    too); `FLOW_GUARD_REPO_ROOT` lines dropped (the script resolves its root from cwd with
    `git rev-parse`); `git rm scripts/test-mutate-and-verify.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/mutate-and-verify.sh`
    with no arguments exits 4 with the usage line it printed at `d71a2327`.

- [ ] 7. Port prepare-archive-branch

**Files:** `stats/internal/guard/preparearchivebranch.go`, `stats/internal/guard/prepare_archive_branch_test.go`, `scripts/prepare-archive-branch.sh`, `scripts/test-prepare-archive-branch.sh`
**Tests:** `TestPrepareArchiveBranch`
**Regression:** fails if any of the harness's 97 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/prepare_archive_branch_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 2
**Commit:** `feat(stats): port prepare-archive-branch to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** shared-helper-go-twins

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-prepare-archive-branch.sh`,
    one subtest per `ok:` label, plus the Review Focus rows (tree and HEAD unchanged after each
    refusal exit; a landing path with a space). Run — expect failure.
  - [ ] **Step 2: Port**, registering `prepare-archive-branch`; exit codes 0/1/2/3 and the
    post-run drift meaning of 2 per the header; drift detection through `postmutationcheck.go`
    (task 2); every `git` call kept in the bash's order, since its step numbering (2b …) is
    cited by `skills/flow-contracts/finish-contract-run2.md`.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestPrepareArchiveBranch$' -count=1
    -race -v | grep -c -- '--- PASS: TestPrepareArchiveBranch/'` — at least 97.
  - [ ] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines dropped;
    `git rm scripts/test-prepare-archive-branch.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/prepare-archive-branch.sh` with no arguments exits 2 with the line it printed at
    `d71a2327`.

- [ ] 8. Port check-base-moved

**Files:** `stats/internal/guard/basemoved.go`, `stats/internal/guard/resolveremotebase.go`, `stats/internal/guard/check_base_moved_test.go`, `scripts/check-base-moved.sh`, `scripts/test-check-base-moved.sh`
**Tests:** `TestCheckBaseMoved`
**Regression:** fails if any of the harness's 63 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_base_moved_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 1
**Commit:** `feat(stats): port check-base-moved to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** sole-user-helpers-move-into-go

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-base-moved.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-base-moved`; `scripts/lib/resolve-remote-base.sh`'s
    function ported to `resolveremotebase.go` (shared with task 12); the `MOVED`/`CLEAR` verdict
    lines and the sorted path lists kept byte for byte, sort collation per Review Focus.
    `scripts/lib/resolve-remote-base.sh` stays until task 12 removes its last caller.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckBaseMoved$' -count=1 -race
    -v | grep -c -- '--- PASS: TestCheckBaseMoved/'` — at least 63.
  - [ ] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines dropped;
    `git rm scripts/test-check-base-moved.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/check-base-moved.sh`
    run against this worktree and `d71a2327` prints the same verdict as the bash at `d71a2327`.

- [ ] 9. Port check-panel-fix-single-dispatch

**Files:** `stats/internal/guard/panelfixsingledispatch.go`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `scripts/check-panel-fix-single-dispatch.sh`, `scripts/test-check-panel-fix-single-dispatch.sh`
**Tests:** `TestCheckPanelFixSingleDispatch`
**Regression:** fails if any of the harness's 20 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_panel_fix_single_dispatch_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 1
**Commit:** `feat(stats): port check-panel-fix-single-dispatch to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [ ] **Step 1: Failing test.** Port every case of
    `scripts/test-check-panel-fix-single-dispatch.sh`, one subtest per `ok:` label; the stub
    `flow` the harness puts on PATH becomes the injected findings hook on `Env` the existing
    ports use (`panelreproducers.go`). Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-panel-fix-single-dispatch`; the `jq` filters become
    `encoding/json` decoding of the same fields; `CANONICAL_KEY_RE` ported as a Go `regexp` only
    where RE2 matches identically, with a subtest pinning it.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckPanelFixSingleDispatch$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckPanelFixSingleDispatch/'` — at least 20.
  - [ ] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines
    dropped; `git rm scripts/test-check-panel-fix-single-dispatch.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-panel-fix-single-dispatch.sh` with no arguments exits 2 with the line it
    printed at `d71a2327`.

- [ ] 10. Port check-model-keys

**Files:** `stats/internal/guard/modelkeys.go`, `stats/internal/guard/check_model_keys_test.go`, `scripts/check-model-keys.sh`, `scripts/test-check-model-keys.sh`
**Tests:** `TestCheckModelKeys`
**Regression:** fails if any of the harness's 30 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_model_keys_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 2
**Commit:** `feat(stats): port check-model-keys to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** shared-helper-go-twins

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-model-keys.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-model-keys`; section reads through
    `projectsection.go` (task 2); the valid-model set it reads from
    `stats/internal/store/settings.go` read from that file's text as the bash does, never by
    importing the store package, so a fixture tree can supply its own.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckModelKeys$' -count=1 -race
    -v | grep -c -- '--- PASS: TestCheckModelKeys/'` — at least 30.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-model-keys.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/check-model-keys.sh`
    on this tree exits 0 with the same output it printed at `d71a2327`.

- [ ] 11. Port prove-reproducer

**Files:** `stats/internal/guard/provereproducer.go`, `stats/internal/guard/prove_reproducer_test.go`, `scripts/prove-reproducer.sh`, `scripts/test-prove-reproducer.sh`, `scripts/lib/reproducer-path.sh`
**Tests:** `TestProveReproducer`
**Regression:** fails if any of the harness's 8 cases regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/prove_reproducer_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 1
**Commit:** `feat(stats): port prove-reproducer to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** in-process-run-reproducer

**Decision:** sole-user-helpers-move-into-go

  - [ ] **Step 1: Failing test.** Port `case_1`…`case_8` of `scripts/test-prove-reproducer.sh`,
    one subtest per case named after the case's own comment, plus the Review Focus rows (tree
    and HEAD unchanged after a refused leg; a worktree path with a space). Run — expect failure.
  - [ ] **Step 2: Port**, registering `prove-reproducer`; both legs call the Go run-reproducer
    in-process (`runreproducer.go`) with the argv the bash passed to `run-reproducer.sh`, its
    returned code mapped to exactly the verdicts the bash read from the shim's exit; the scratch
    worktree created and removed as the bash did; `scripts/lib/reproducer-path.sh`'s check ported
    into `provereproducer.go` and the library `git rm`'d.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestProveReproducer$' -count=1 -race
    -v | grep -c -- '--- PASS: TestProveReproducer/'` — at least 8.
  - [ ] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines
    dropped; `git rm scripts/test-prove-reproducer.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/prove-reproducer.sh`
    with no arguments exits 2 with the usage line it printed at `d71a2327`.

- [ ] 12. Port check-finish-preflight

**Files:** `stats/internal/guard/finishpreflight.go`, `stats/internal/guard/check_finish_preflight_test.go`, `scripts/check-finish-preflight.sh`, `scripts/test-check-finish-preflight.sh`, `scripts/lib/resolve-remote-base.sh`
**Tests:** `TestCheckFinishPreflight`
**Regression:** fails if any of the harness's 58 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_finish_preflight_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 8
**Commit:** `feat(stats): port check-finish-preflight to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** exec-unported-siblings

**Decision:** sole-user-helpers-move-into-go

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-finish-preflight.sh`,
    one subtest per `ok:` label; cases that stub `check-worktree-location.sh` keep stubbing it as a
    sibling file in the fixture's scripts directory. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-finish-preflight`; base resolution through
    `resolveremotebase.go` (task 8); `check-worktree-location.sh` exec'd from
    `$FLOW_GUARD_REPO_ROOT/scripts/` with the bash's argv, its exit code and stderr handled as
    the bash did (the ancestor test's "only exit 1 means not an ancestor" kept);
    `git rm scripts/lib/resolve-remote-base.sh`, whose last caller this was.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckFinishPreflight$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckFinishPreflight/'` — at least 58.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT` (the sibling
    exec needs it), code 2; `git rm scripts/test-check-finish-preflight.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-finish-preflight.sh` against this worktree prints the same verdict the bash
    printed at `d71a2327`, directly and through a symlink to it in a temp directory.

- [ ] 13. Repoint citations of the deleted files

**Files:** `.flow/project.md`
**Allowed-collateral:** `.flow/*.md`, `scripts/*.sh`, `scripts/lib/*.sh`, `scripts/*.py`, `skills/**/*.md`, `rules/*.mdc`, `README.md`, `CONTRIBUTING.md`, `stats/internal/guard/*.go`, `stats/internal/records/*_test.go`
**Tests:** none — citation sweep; the lint guards are the check
**Regression:** none — prose and comments only
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 3, 4, 5, 6, 7, 8, 9, 10, 11, 12
**Commit:** `docs(scripts): repoint citations of the ported scripts to their Go sources`
**Build:** green

**Decision:** carry-prior-port-decisions

  - [ ] **Step 1: Find.** `grep -rlF -e test-check-stage-mark-calls.sh -e
    test-check-guard-symlinks.sh -e test-check-dispatch-paragraphs.sh -e test-mutate-and-verify.sh
    -e test-prepare-archive-branch.sh -e test-check-base-moved.sh -e
    test-check-panel-fix-single-dispatch.sh -e test-check-model-keys.sh -e
    test-prove-reproducer.sh -e test-check-finish-preflight.sh -e lib/resolve-remote-base.sh -e
    lib/reproducer-path.sh --exclude-dir=archive --exclude-dir=.worktrees
    --exclude-dir=node_modules --exclude-dir=.git --exclude-dir=self-review .`; then grep the same
    tree for citations of a ported script's body comments (`<name>.sh`'s `# Step`, `rule <n>`
    and header-paragraph citations) and of its bash plumbing (`sources lib/…`).
    `unverified: the file set is known only after tasks 3–12 land; widen **Files:** by a
    correction if a hit falls outside the collateral globs`
  - [ ] **Step 2: Repoint** each citation to the Go file or Go test that now holds what it cites;
    a sentence describing bash plumbing that no longer exists is corrected, not repointed.
    `.flow/project.md`'s paragraph naming which guards are Go lists the ten new ones; library
    headers naming a ported script as a caller that sources them are corrected.
  - [ ] **Step 3: Verify.** Step 1's grep returns only `spectre/changes/kan-841-*`,
    `docs/self-review/` and the Go ports' own history comments; every guard in
    `.flow/project.md`'s `## lint` exits 0.

- [ ] 14. Live verification: after timings and parity

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13
**Build:** green

**Decision:** suite-median-below-before

**Decision:** guard-package-under-30s

  - [ ] **Step 1: Suite, after.** Task 1 step 1's command, three times, on the branch head; same
    fields recorded.
  - [ ] **Step 2: Go package, after.** Task 1 step 2's command, three times.
  - [ ] **Step 3: Parity.** `cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '---
    PASS: Test<Name>/'` per port against its floor in `design.md`.
  - [ ] **Step 4: Record** an **After** table beside **Before**, same columns, each figure tagged
    `measured:` with the command and
    `@ branch spectre/kan-841-agents-port-the-next-ten-slowest-bash-scripts-to`; name the slowest
    remaining harness and the next slice.
  - [ ] **Step 5: Judge.** Failure looks like: suite median not below the Before median; any
    port's `--- PASS` count below its floor; the Go package median above 30s real; any harness
    red. Any of these is reported, not recorded as success.
