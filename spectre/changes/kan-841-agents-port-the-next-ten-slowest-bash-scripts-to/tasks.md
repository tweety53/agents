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

- [x] 1. Live verification: before timings

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** none
**Build:** green

**Decision:** suite-median-below-before

  - [x] **Step 1: Suite, before.** On this machine at `d71a2327`, nothing else heavy running:
    `sysctl -n vm.loadavg` then `FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p
    scripts/run-guard-tests.sh`, three times; record each run's real/user/sys, load, harness
    count and the slowest five harnesses (`grep '(Ns)'` of each log, sorted descending).
  - [x] **Step 2: Go package, before.** `cd stats && /usr/bin/time -p go test
    ./internal/guard/... -count=1` three times; record real/user/sys.
  - [x] **Step 3: Record** a **Before** table under `design.md`'s **Measurements** → **Suite
    before/after**, KAN-778's columns, each figure tagged `measured:` with the command and
    `@ d71a2327`.

This task commits nothing; its figures are committed with the change's artifacts.

- [x] 2. Go twins of resolve-file, project-section and post-mutation-check

**Files:** `stats/internal/guard/resolvefile.go`, `stats/internal/guard/projectsection.go`, `stats/internal/guard/postmutationcheck.go`, `stats/internal/guard/libtwins_test.go`, `stats/internal/guard/gatherdispatch.go`
**Tests:** `TestResolveFileParity`, `TestProjectSectionParity`, `TestPostMutationCheckParity`, `TestGitExecSignalStatus`
**Regression:** each parity test fails if its Go twin's output, return status or side effect
differs from `scripts/lib/resolve-file.sh`, `scripts/lib/project-section.sh` or
`scripts/lib/post-mutation-check.sh` for the same inputs.
**Baseline:** before=0 after=4
<!-- measured: cat stats/internal/guard/libtwins_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** none
**Commit:** `feat(stats): add Go twins of resolve-file, project-section and post-mutation-check`
**Build:** green

**Decision:** shared-helper-go-twins

  - [x] **Step 1: Failing tests.** Read each library's header for its functions and contract. For
    each, a table of inputs covering every branch of its body (resolve-file: plain file, symlink
    chain, relative symlink, dangling link, directory; project-section: present section, absent,
    empty body, fenced body, heading-level edge, trailing prose; post-mutation-check: clean tree,
    dirty tracked file, new untracked file, drift under a mutation). Each row runs the bash
    library once via `bash -c '. scripts/lib/<lib>.sh; <fn> <args>'` against a fixture in
    `t.TempDir()` and the Go function in-process, comparing stdout, stderr and status byte for
    byte. Run `cd stats && go test ./internal/guard/ -run 'Parity$' -count=1` — expect a compile
    failure.
  - [x] **Step 2: Port** each library's functions into its Go file, its header citing the bash
    library as the source of truth it mirrors and the parity test that pins them.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
    ./internal/guard/ -run '^(TestResolveFileParity|TestProjectSectionParity|TestPostMutationCheckParity)$'
    -count=1 -race`.

Correction (2026-09-27): the plan declared three new twins. `resolveFile` and `projectSection`
already existed, untested, inside `gatherdispatch.go`; they were moved unchanged into
`resolvefile.go`/`projectsection.go` (helper `gdcPhysicalDir` renamed `physicalDir`) instead of
written a second time, so `gatherdispatch.go` joins **Files:**. `project_section` on a missing
file is left without a parity row: bash prints `cat`'s error, the twin prints nothing, and every
caller checks existence first.

- [x] 3. Port check-stage-mark-calls

**Files:** `stats/internal/guard/stagemarkcalls.go`, `stats/internal/guard/check_stage_mark_calls_test.go`, `stats/internal/guard/guard.go`, `scripts/check-stage-mark-calls.sh`, `scripts/test-check-stage-mark-calls.sh`
**Tests:** `TestCheckStageMarkCalls`
**Regression:** fails if any of the harness's 74 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_stage_mark_calls_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 1
**Commit:** `feat(stats): port check-stage-mark-calls to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-stage-mark-calls.sh`,
    one subtest per `ok:` label; fixtures are small trees in `t.TempDir()`. Run — expect failure.
  - [x] **Step 2: Port**, registering `check-stage-mark-calls`; per-member coverage through
    `coverage.go`; the stage-key set it reads from the stats Go sources (`go` calls in the body)
    read the same way.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckStageMarkCalls$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckStageMarkCalls/'` — at least 74.
  - [x] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-stage-mark-calls.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-stage-mark-calls.sh` on this tree exits 0 with the same output it printed at
    `d71a2327`, directly and through a symlink to it in a temp directory.

Correction (2026-09-27): the injected `Env.StageKeys` hook lives on `Env` in `guard.go`, which
joins **Files:**. Step 5's "through a symlink to it in a temp directory" cannot hold: the bash at
`d71a2327` exits 2 there too (it cannot find `lib/coverage.sh`), and the shim matches it. One
documented divergence: under a UTF-8 locale macOS awk aborts on invalid UTF-8 and the bash counted
that file as zero calls; the port checks it (the two agree under `LC_ALL=C`).

- [x] 4. Port check-guard-symlinks

**Files:** `stats/internal/guard/guardsymlinks.go`, `stats/internal/guard/check_guard_symlinks_test.go`, `scripts/check-guard-symlinks.sh`, `scripts/test-check-guard-symlinks.sh`
**Tests:** `TestCheckGuardSymlinks`, `TestShimSiblingsDeclared`
**Regression:** fails if any of the harness's 118 `ok:` behaviours regress.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/check_guard_symlinks_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 2
**Commit:** `feat(stats): port check-guard-symlinks to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** shared-helper-go-twins

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-guard-symlinks.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [x] **Step 2: Port**, registering `check-guard-symlinks`; link resolution through
    `resolvefile.go` (task 2), coverage through `coverage.go`; every awk program in the body
    (`CITATION_AWK`, `DELEGATE_AWK`, `RULE3_AWK`) ported as Go code with its rule's subtests
    pinning it. Rule 2's sibling derivation, which greps a guard's source for
    `$SCRIPT_DIR/<name>`, keeps reading the Go source for a shimmed guard, as the bash does since
    KAN-760 — now including the ten scripts this change shims.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckGuardSymlinks$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckGuardSymlinks/'` — at least 118.
  - [x] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-guard-symlinks.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-guard-symlinks.sh` on this tree exits 0 with the same output it printed at
    `d71a2327`, directly and through a symlink to it in a temp directory.

Correction (2026-09-27): step 2's premise was false — at `d71a2327` rule 2 reads
`scripts/<guard>`, which for a shimmed guard is the shim, not its Go source, and no KAN-760 commit
changed that. The port keeps the bash behaviour (parity): a shimmed guard contributes no
`$SCRIPT_DIR/` siblings. The task-4 review found the citation, rule 3 and delegation scans read past
a NUL where the bash's awk ended the line; fixed to parity and pinned against the live bash. A NUL
in a rule-4 guard source still differs on stderr only: bash adds its own `ignored null byte`
warning naming the script's path, which a Go binary cannot emit.

- [x] 5. Port check-dispatch-paragraphs

**Files:** `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`, `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
**Tests:** `TestCheckDispatchParagraphs`
**Regression:** fails if any of the harness's 183 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_dispatch_paragraphs_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 1
**Commit:** `feat(stats): port check-dispatch-paragraphs to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-dispatch-paragraphs.sh`
    (3542 lines), one subtest per `ok:` label, grouped into table-driven subtests by the harness's
    own sections. Run — expect failure.
    <!-- measured: wc -l < scripts/test-check-dispatch-paragraphs.sh @ d71a2327 -->
  - [x] **Step 2: Port**, registering `check-dispatch-paragraphs`; the site table (`SITE_PATHS`,
    `SITE_VARIANTS`, `SITE_MIN_BLOCKS`, `ENTRY_*`) becomes a Go table in the same order;
    `CHECK_DISPATCH_PARAGRAPHS_ROOT` and `CHECK_GUARD_SYMLINKS_ROOT` overrides kept, set-but-empty
    told from unset through `Env.LookupEnv`.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckDispatchParagraphs$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckDispatchParagraphs/'` — at least 183.
  - [x] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-dispatch-paragraphs.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-dispatch-paragraphs.sh` on this tree exits 0 with the same output it printed at
    `d71a2327`.

Correction (2026-09-27): step 2 named a `CHECK_GUARD_SYMLINKS_ROOT` override; the script at
`d71a2327` never reads it (its header cites it only as the precedent it mirrors), so only
`CHECK_DISPATCH_PARAGRAPHS_ROOT` is kept. The old-vs-new diff matched 93 of 94 runs byte for
byte; the one difference is bash's own `ignored null byte` warning on the NUL fixture, which a Go
binary cannot emit. The task-5 review found a second, recorded in `dispatchparagraphs.go`: macOS BSD
grep 2.6.0 in a UTF-8 locale finds no non-ASCII label in a file holding a NUL byte, so the bash
reported the em-dash `**VERBATIM REPORT — THE FACT:**` label missing where the port — matching the
header, GNU grep and `LC_ALL=C` — finds it.

- [x] 6. Port mutate-and-verify

**Files:** `stats/internal/guard/mutateandverify.go`, `stats/internal/guard/mutate_and_verify_test.go`, `scripts/mutate-and-verify.sh`, `scripts/test-mutate-and-verify.sh`, `stats/cmd/flow-guard/main.go`, `stats/cmd/flow-guard/main_test.go`
**Tests:** `TestMutateAndVerify`
**Regression:** fails if any of the harness's 44 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/mutate_and_verify_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 2
**Commit:** `feat(stats): port mutate-and-verify to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** shared-helper-go-twins

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-mutate-and-verify.sh`, one
    subtest per `ok:` label, plus the Review Focus rows (tree and HEAD unchanged after each
    refusal exit; a worktree path with a space; the real shim with no `go` on PATH exits 4).
    Run — expect failure.
  - [x] **Step 2: Port**, registering `mutate-and-verify`; exit codes 0/2/3/4 per the header;
    `MUTATE_AND_VERIFY_MAX_NEW_FAILURES` kept with its parsing and default; the post-mutation
    drift check through `postmutationcheck.go` (task 2); harnesses are run as child processes
    exactly as the bash ran them (same argv, cwd, inherited environment).
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestMutateAndVerify$' -count=1
    -race -v | grep -c -- '--- PASS: TestMutateAndVerify/'` — at least 44.
  - [x] **Step 4: Shim and delete** — shim template, code 4 (the loader's own failure exits 4
    too); `FLOW_GUARD_REPO_ROOT` lines dropped (the script resolves its root from cwd with
    `git rev-parse`); `git rm scripts/test-mutate-and-verify.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/mutate-and-verify.sh`
    with no arguments exits 4 with the usage line it printed at `d71a2327`.

Correction (2026-09-27): `flow-guard`'s own failure path (`os.Getwd` failing) answered
`cannotAnswer`'s 2, which `mutate-and-verify` reads as "refused"; `cannotAnswer` now returns 4 for
it, pinned by a row in `TestCannotAnswerIsTheGuardsOwnCode`, so `stats/cmd/flow-guard/main.go` and
`main_test.go` join **Files:**. Signals: the bash's EXIT trap restored the touched files and
reported on SIGTERM/SIGHUP before dying of the signal; the port handles SIGHUP/SIGINT/SIGTERM the
same way and re-raises (pinned through the real shim). On SIGINT it restores at once where
`/bin/bash` 3.2 waited for the harness — a `ponytail:` comment names it. The trap is shared with
task 11 as `trapExitSignals` (`provereproducer.go`) and also takes
SIGPIPE: bash 3.2's trap restored on a closed stdout where bash 5.3 died leaving the mutation, and
the port restores and exits 141 at once (the Go runtime cannot re-raise a SIGPIPE it did not
raise), and the main flow halts at the failed write, so no later harness starts (pinned with a
second harness). A sent SIGABRT, SIGFPE, SIGSYS or SIGTRAP, which the Go runtime answers with a
stack dump and exit 2 (read as "refused, nothing was mutated"), is trapped the same way and exits
128+n; SIGABRT is pinned. A sent SIGSEGV, SIGBUS or SIGILL is trapped the same way on linux, but
never reaches `signal.Notify` on darwin, nor does darwin's SIGEMT — the runtime crashes with exit 2
and the mutation stays — accepted and named in the `ponytail:` comment, a trap in the shim being
the upgrade path; so is a SIGPIPE ignored at start, which is trapped anyway (exit 141 after the
restore) where bash ran on to its own exit. `git apply`'s output is written once git has exited,
so a closed stdout neither hangs the restore nor hides an apply that ran (pinned with a `git`
first on PATH that prints while applying). The patch is never applied once a signal is being handled
(pinned on `mvRun.apply`). Other signals bash's trap caught (SIGUSR1, ...) the Go runtime ignores, and a SIGTERM ignored at start is not seen as ignored — both accepted
and named in the `ponytail:` comment. A harness with no `#!` line reads as unanswered (exit 4) where
bash ran it as a script — accepted, every harness this runs has one. The mechanisms bash got from
`$(...)` and `[ -le ]` are pinned against the bash at `d71a2327`: NUL bytes are dropped from
harness output (bash 5's "ignored null byte" warning is not reproduced), a bare `FAIL: ` takes no
part in the difference, and an out-of-range bound falls to FLAG (bash's own `integer expected`
line is not reproduced). `mvGitRun` and task 2's `pmcGit` became one `gitExec` in
`postmutationcheck.go`; the post-mutation-check twins keep running
`git -C <worktree>`, so a missing worktree prints git's own error and returns 128 as the bash did,
pinned by a parity row in `libtwins_test.go` (fix commits touch these three files, so
they stay out of **Files:**, which records the task commit).

- [x] 7. Port prepare-archive-branch

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

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-prepare-archive-branch.sh`,
    one subtest per `ok:` label, plus the Review Focus rows (tree and HEAD unchanged after each
    refusal exit; a landing path with a space). Run — expect failure.
  - [x] **Step 2: Port**, registering `prepare-archive-branch`; exit codes 0/1/2/3 and the
    post-run drift meaning of 2 per the header; drift detection through `postmutationcheck.go`
    (task 2); every `git` call kept in the bash's order, since its step numbering (2b …) is
    cited by `skills/flow-contracts/finish-contract-run2.md`.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestPrepareArchiveBranch$' -count=1
    -race -v | grep -c -- '--- PASS: TestPrepareArchiveBranch/'` — at least 97.
  - [x] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines dropped;
    `git rm scripts/test-prepare-archive-branch.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/prepare-archive-branch.sh` with no arguments exits 2 with the line it printed at
    `d71a2327`.

Correction (2026-09-27): paths join the caller's directory uncleaned, so `lnk/..` resolves
physically as the bash's `[ -e ]` and `git -C` did. The bash's `export LC_ALL=C` is not carried to
git: parsed git output is locale-independent, and the git stderr passed through prints in the
caller's locale — visible only with a localized git. Each refusal after HEAD has moved, the rename
and directory-entry classification, and the byte-wise branch-name check are pinned against the
bash at `d71a2327`, status and branch included. Two defects the bash had and the port keeps (a
relative landing path given from outside the main checkout; porcelain-quoted paths never
classified) are recorded in `KNOWN-BUGS.md`.

- [x] 8. Port check-base-moved

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

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-base-moved.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [x] **Step 2: Port**, registering `check-base-moved`; `scripts/lib/resolve-remote-base.sh`'s
    function ported to `resolveremotebase.go` (shared with task 12); the `MOVED`/`CLEAR` verdict
    lines and the sorted path lists kept byte for byte, sort collation per Review Focus.
    `scripts/lib/resolve-remote-base.sh` stays until task 12 removes its last caller.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckBaseMoved$' -count=1 -race
    -v | grep -c -- '--- PASS: TestCheckBaseMoved/'` — at least 63.
  - [x] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines dropped;
    `git rm scripts/test-check-base-moved.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/check-base-moved.sh`
    run against this worktree and `d71a2327` prints the same verdict as the bash at `d71a2327`.

Correction (2026-09-27): the port's private `fromDir`/git-output helpers duplicated
`changeplan.go`'s; they are folded into the existing `capture`, which `changeplan.go` now shares
(a fix commit touches it, so it stays out of **Files:**, which records the task commit). The path lists carry `core.quotePath` octal escapes (`"\303\244.txt"`)
exactly as the bash printed them, pinned by a non-ASCII fixture, and the usage text cites
`base_ref_usage_message` at `d71a2327` since task 12 deletes that lib.

- [x] 9. Port check-panel-fix-single-dispatch

**Files:** `stats/internal/guard/panelfixsingledispatch.go`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `scripts/check-panel-fix-single-dispatch.sh`, `scripts/test-check-panel-fix-single-dispatch.sh`, `stats/internal/guard/guard.go`
**Tests:** `TestCheckPanelFixSingleDispatch`
**Regression:** fails if any of the harness's 20 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_panel_fix_single_dispatch_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 1, 3
**Commit:** `feat(stats): port check-panel-fix-single-dispatch to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [x] **Step 1: Failing test.** Port every case of
    `scripts/test-check-panel-fix-single-dispatch.sh`, one subtest per `ok:` label; the stub
    `flow` the harness puts on PATH becomes the injected findings hook on `Env` the existing
    ports use (`panelreproducers.go`). Run — expect failure.
  - [x] **Step 2: Port**, registering `check-panel-fix-single-dispatch`; the `jq` filters become
    `encoding/json` decoding of the same fields; `CANONICAL_KEY_RE` ported as a Go `regexp` only
    where RE2 matches identically, with a subtest pinning it.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckPanelFixSingleDispatch$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckPanelFixSingleDispatch/'` — at least 20.
  - [x] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines
    dropped; `git rm scripts/test-check-panel-fix-single-dispatch.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-panel-fix-single-dispatch.sh` with no arguments exits 2 with the line it
    printed at `d71a2327`.

Correction (2026-09-27): the guard reads `flow record dispatches` as well as findings, so
`guard.go` gains an `Env.Dispatches` hook and joins **Files:**. Header and body disagreed on one
edge: malformed findings JSON on a chunked round exits 5 (or 1) in the body, 2 ("jq … failing")
in the header. The port follows the header's cannot-answer contract — exit 2, `findings rows were
not readable JSON -- cannot answer` — surfaced to the operator at the handoff. Dispatch rows given
as a JSON object (never printed by `flow`) are refused with exit 2 where jq walked its values.
A non-string key prints as compact JSON where jq pretty-printed it, and bash 5's "ignored null
byte" warning (3.2 prints none) is not reproduced — both accepted, as are the `cd` builtin's own
diagnostics (`cd: …: Permission denied`, `cd: -x: invalid option` and its usage line), which the
port does not print beside the guard's own refusal line.
The mechanisms the bash got implicitly from jq and `$(...)` — a NUL in a key dropped, trailing
newlines trimmed, a boolean `round` at round 0, an empty findings array, 64-bit wrap, a null row —
are pinned by a `port:` subtest that runs the bash at `d71a2327` beside the port. A multi-value
findings stream cannot answer (exit 2). A worktree without search permission is refused as
vanished, as the bash's `cd` refused it; the helper `pcAbs` replaces the identical `abs` closures
in `panelexitcontract.go` and `panelreproducers.go` (touched by a fix commit, so out
of **Files:**, which records the task commit) — their own
search-permission gap is out of scope and recorded in `KNOWN-BUGS.md`.

- [x] 10. Port check-model-keys

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

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-model-keys.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [x] **Step 2: Port**, registering `check-model-keys`; section reads through
    `projectsection.go` (task 2); the valid-model set it reads from
    `stats/internal/store/settings.go` read from that file's text as the bash does, never by
    importing the store package, so a fixture tree can supply its own.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckModelKeys$' -count=1 -race
    -v | grep -c -- '--- PASS: TestCheckModelKeys/'` — at least 30.
  - [x] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-model-keys.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/check-model-keys.sh`
    on this tree exits 0 with the same output it printed at `d71a2327`.

Correction (2026-09-27): the script's header says it checks "both keys"; its loop at `d71a2327`
checks only `self review model`. The port follows the code, and the header question is surfaced to
the operator at the handoff. `lib/project-section.sh`'s header and `gatherdispatch.go`'s comment
named `check-model-keys.sh` as a caller; fix commits drop it from both, which stay out
of **Files:**, since it records the task commit.

- [x] 11. Port prove-reproducer

**Files:** `stats/internal/guard/provereproducer.go`, `stats/internal/guard/prove_reproducer_test.go`, `scripts/prove-reproducer.sh`, `scripts/test-prove-reproducer.sh`, `scripts/lib/reproducer-path.sh`, `stats/internal/guard/guard.go`
**Tests:** `TestProveReproducer`
**Regression:** fails if any of the harness's 8 cases regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/prove_reproducer_test.go 2>/dev/null | grep -cE '^func Test' @ d71a2327 -->
**After:** Task 1, 3, 6, 9
**Commit:** `feat(stats): port prove-reproducer to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** in-process-run-reproducer

**Decision:** sole-user-helpers-move-into-go

  - [x] **Step 1: Failing test.** Port `case_1`…`case_8` of `scripts/test-prove-reproducer.sh`,
    one subtest per case named after the case's own comment, plus the Review Focus rows (tree
    and HEAD unchanged after a refused leg; a worktree path with a space). Run — expect failure.
  - [x] **Step 2: Port**, registering `prove-reproducer`; both legs call the Go run-reproducer
    in-process (`runreproducer.go`) with the argv the bash passed to `run-reproducer.sh`, its
    returned code mapped to exactly the verdicts the bash read from the shim's exit; the scratch
    worktree created and removed as the bash did; `scripts/lib/reproducer-path.sh`'s check ported
    into `provereproducer.go` and the library `git rm`'d.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestProveReproducer$' -count=1 -race
    -v | grep -c -- '--- PASS: TestProveReproducer/'` — at least 8.
  - [x] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines
    dropped; `git rm scripts/test-prove-reproducer.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/prove-reproducer.sh`
    with no arguments exits 2 with the usage line it printed at `d71a2327`.

Correction (2026-09-27): signals are handled as the bash's trap did — on SIGINT/SIGTERM/SIGHUP
the scratch worktree is removed at once and the signal re-raised (143 for SIGTERM), with signals
ignored at entry left ignored (Go sees an inherited ignore for SIGHUP and SIGINT only — a SIGTERM
ignored at start is handled, accepted); no per-step checkpoint. A SIGINT sent to the guard's pid
alone ends the run at once where bash waited for its leg — accepted, a Ctrl-C reaches the whole
process group; the subtest sends it that way. `guard.go` joins **Files:** only to drop the
`Env.Signals` hook an earlier revision added. `mkdir -p` and `cp -p` run as child processes so
their stderr reaches the caller byte for byte, the directory taken by `dirname`, uncleaned
(`a/./r.sh` creates `a/.`). The exec-bit check on the copy stays: it fails on a noexec TMPDIR,
and is pinned against the bash through a `cp` first on PATH that drops the mode. A SIGPIPE on the
verdict write exits 141, as both bash versions did, and a sent SIGABRT, SIGFPE, SIGSYS or SIGTRAP
exits 128+n (the trap is task 6's, shared), as does a sent SIGSEGV, SIGBUS or SIGILL on linux; on
darwin those crash the Go runtime, leaving the scratch — accepted, as recorded under task 6. The legs' run-reproducer temp files go under the scratch base, so
the cleanup removes them on a signal too, as the bash's separate run-reproducer process did. A
comment fix to task 9's test rode this task's second fix commit.

- [x] 12. Port check-finish-preflight

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

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-finish-preflight.sh`,
    one subtest per `ok:` label; cases that stub `check-worktree-location.sh` keep stubbing it as a
    sibling file in the fixture's scripts directory. Run — expect failure.
  - [x] **Step 2: Port**, registering `check-finish-preflight`; base resolution through
    `resolveremotebase.go` (task 8); `check-worktree-location.sh` exec'd from
    `$FLOW_GUARD_REPO_ROOT/scripts/` with the bash's argv, its exit code and stderr handled as
    the bash did (the ancestor test's "only exit 1 means not an ancestor" kept);
    `git rm scripts/lib/resolve-remote-base.sh`, whose last caller this was.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckFinishPreflight$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckFinishPreflight/'` — at least 58.
  - [x] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT` (the sibling
    exec needs it), code 2; `git rm scripts/test-check-finish-preflight.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-finish-preflight.sh` against this worktree prints the same verdict the bash
    printed at `d71a2327`, directly and through a symlink to it in a temp directory.

Correction (2026-09-27): `scripts/lib/base-ref-usage.sh` loses its last caller with this port and
is removed by a fix commit, so it stays out of **Files:**, which records the task commit. The dirty-file count, the physical main-checkout path and the
signal-killed child's `exited 143` wording are pinned against the bash at `d71a2327`.

- [x] 13. Repoint citations of the deleted files

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

  - [x] **Step 1: Find.** `grep -rlF -e test-check-stage-mark-calls.sh -e
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
  - [x] **Step 2: Repoint** each citation to the Go file or Go test that now holds what it cites;
    a sentence describing bash plumbing that no longer exists is corrected, not repointed.
    `.flow/project.md`'s paragraph naming which guards are Go lists the ten new ones; library
    headers naming a ported script as a caller that sources them are corrected.
  - [x] **Step 3: Verify.** Step 1's grep returns only `spectre/changes/kan-841-*`,
    `docs/self-review/` and the Go ports' own history comments; every guard in
    `.flow/project.md`'s `## lint` exits 0.

- [x] 14. Live verification: after timings and parity

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13
**Build:** green

**Decision:** suite-median-below-before

**Decision:** guard-package-under-30s

  - [x] **Step 1: Suite, after.** Task 1 step 1's command, three times, on the branch head; same
    fields recorded.
  - [x] **Step 2: Go package, after.** Task 1 step 2's command, three times.
  - [x] **Step 3: Parity.** `cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '---
    PASS: Test<Name>/'` per port against its floor in `design.md`.
  - [x] **Step 4: Record** an **After** table beside **Before**, same columns, each figure tagged
    `measured:` with the command and
    `@ branch spectre/kan-841-agents-port-the-next-ten-slowest-bash-scripts-to`; name the slowest
    remaining harness and the next slice.
  - [x] **Step 5: Judge.** Failure looks like: suite median not below the Before median; any
    port's `--- PASS` count below its floor; the Go package median above 30s real; any harness
    red. Any of these is reported, not recorded as success.
