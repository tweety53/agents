# Self-review context bundle for kan-873-agents-port-the-next-five-slowest-bash-guards-to

found: 6 of 7 sources; skipped: 1 of 7 sources
skipped: change summary (absent)

## .superpowers/sdd/ledgers/kan-873-agents-port-the-next-five-slowest-bash-guards-to.md

# SDD ledger — kan-873-agents-port-the-next-five-slowest-bash-guards-to

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — implementer

- Task: no task
- Role: implementer
- Key: task-1-implementer
- Model: opus effort=medium
- Commit: 9eb4e659
- Outcome: completed
- Started: 2026-10-03T10:44:30Z
- Tokens: input 240, output 27267, cache read 17335591, cache creation 299430

## Dispatch 2 — implementer

- Task: no task
- Role: implementer
- Key: task-2-implementer
- Model: opus effort=high
- Commit: bc158bd6
- Outcome: completed
- Started: 2026-10-03T10:44:30Z
- Tokens: not measured

## Dispatch 3 — implementer

- Task: no task
- Role: implementer
- Key: task-3-implementer
- Model: opus effort=high
- Commit: ae96f2b3
- Outcome: completed
- Started: 2026-10-03T10:44:30Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Key: task-1+4+7-reviewer
- Model: opus effort=medium
- Commit: no commit
- Outcome: fix
- Started: 2026-10-03T11:03:13Z
- Tokens: input 122, output 5252, cache read 5651273, cache creation 223391

## Dispatch 5 — reviewer

- Task: 5
- Role: reviewer
- Key: task-5-reviewer
- Model: opus effort=high
- Commit: no commit
- Outcome: clean
- Started: 2026-10-03T11:14:13Z
- Tokens: input 98, output 2930, cache read 4439185, cache creation 165442

## Dispatch 6 — reviewer

- Task: no task
- Role: reviewer
- Key: task-3+6-reviewer
- Model: opus effort=high
- Commit: no commit
- Outcome: fix
- Started: 2026-10-03T11:14:13Z
- Tokens: not measured

## Dispatch 7 — implementer

- Task: no task
- Role: implementer
- Key: task-7-implementer-fix-1
- Model: opus effort=default
- Commit: e102add0
- Outcome: completed
- Started: 2026-10-03T11:16:42Z
- Tokens: not measured

## Dispatch 8 — reviewer

- Task: 7
- Role: reviewer
- Key: task-7-reviewer-fix-1
- Model: sonnet effort=medium
- Commit: no commit
- Outcome: clean
- Started: 2026-10-03T11:17:03Z
- Tokens: not measured

## Dispatch 9 — implementer

- Task: no task
- Role: implementer
- Key: task-6-implementer-fix-1
- Model: opus effort=default
- Commit: 889c4035
- Outcome: completed
- Started: 2026-10-03T11:26:00Z
- Tokens: not measured

## Dispatch 10 — reviewer

- Task: 6
- Role: reviewer
- Key: task-6-reviewer-fix-1
- Model: sonnet effort=medium
- Commit: no commit
- Outcome: clean
- Started: 2026-10-03T11:27:19Z
- Tokens: not measured

## Dispatch 11 — implementer

- Task: 10
- Role: implementer
- Key: task-10-implementer
- Model: opus effort=default
- Commit: 8961e3d2
- Outcome: completed
- Started: 2026-10-03T11:27:19Z
- Tokens: not measured

## Dispatch 12 — implementer

- Task: no task
- Role: implementer
- Key: task-8-implementer
- Model: sonnet effort=medium
- Commit: ea3a5808
- Outcome: completed
- Started: 2026-10-03T11:29:22Z
- Tokens: input 38, output 286, cache read 915274, cache creation 62670

## Dispatch 13 — reviewer

- Task: 10
- Role: reviewer
- Key: task-10-reviewer
- Model: opus effort=medium
- Commit: no commit
- Outcome: clean
- Started: 2026-10-03T11:29:22Z
- Tokens: not measured

## Dispatch 14 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: opus effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-03T11:39:18Z
- Tokens: input 158, output 775, cache read 11884321, cache creation 282028

## Dispatch 15 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-0
- Model: opus effort=medium
- Commit: 509d807cd8d3f4fcb9141e152847f343fff96f11
- Outcome: completed
- Started: 2026-10-03T11:56:41Z
- Tokens: input 118, output 5054, cache read 3778629, cache creation 116453

## Dispatch 16 — reviewer

- Task: no task
- Role: reviewer
- Key: panel-1-primary
- Model: sonnet effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-10-03T12:04:41Z
- Tokens: input 10, output 41, cache read 91697, cache creation 26977

## Dispatch 17 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary-retry
- Model: sonnet effort=low
- Commit: no commit
- Diff base: 2c37d2890925279497a80e211c87e958df641342
- Outcome: completed
- Started: 2026-10-03T12:05:38Z
- Tokens: input 10, output 46, cache read 71435, cache creation 36668

## Dispatch 18 — verifier

- Task: no task
- Role: verifier
- Key: verify
- Model: opus effort=medium
- Commit: no commit
- Outcome: completed
- Started: 2026-10-03T12:06:33Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-873-agents-port-the-next-five-slowest-bash-guards-to-panel.md

# Review panel — kan-873-agents-port-the-next-five-slowest-bash-guards-to

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/guard/breakandprove.go:347 | a --clean command that backgrounds a child stalls each leg until that child exits. |   |
| F2 | primary | Minor | stats/internal/guard/prepareworkspace.go:55 | the sibling guard is looked up in $FLOW_GUARD_REPO_ROOT/scripts, not beside the shim. |   |
| F3 | primary | Minor | docs/prompt-audit-2026-09-29/audit-finish.md:573 | cites scripts/test-check-self-review-report.sh, which this change deletes. |   |
| F4 | principles | Minor | stats/internal/guard/selfreviewreport.go:425 | DRY: re-states crSort's exec'd-sort setup (references.go:280-306). |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-P1-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-P2-1.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-P3-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-R1-1.sh

## Pass log

### Round 0

- auto-resolved: convergence confirm → approve the design and move on
- auto-resolved: which five guards → the five slowest by 3-run median harness time (KAN-842 exclusions)
- auto-resolved: Proceed to implementation? → Yes
- roster: compact — 16
- diff size: 8345 lines, over cap — proceeded automatically
- docs-only: exit 1 — first non-doc path scripts/aside-planning-artifacts.sh; roster primary+principles dispatched
- no addition this round — the resolved list ran alone.
- standards: CLAUDE.md, AGENTS.md
- panel-fix-0 (opus/medium) fixed F1-F4 in 509d807c; reproducers re-authored (premise lines pinned rewritten lines; P3 body could never pass) and proved both directions against 2c37d289 via prove-reproducer.sh; fix-round-0 diff 2c37d289..509d807c
fix-mutation: stats/internal/guard/breakandprove.go — clean.Stdout set back to the EPIPE wrapper stdout — TestBreakAndProveCleanBackgroundChild
fix-mutation: stats/internal/guard/prepareworkspace.go — reverted to the FLOW_GUARD_REPO_ROOT/scripts lookup and REPO_ROOT export — TestPrepareWorkspaceShimSibling
fix-mutation: stats/internal/guard/selfreviewreport.go — crSortSep NUL separator flipped to newline — TestCheckSelfReviewReport/a_report_name_carrying_a_newline_stays_one_path
fix-mutation: stats/internal/guard/references.go — crSort's unique branch dropped -u — TestPanelTouchedPathsParity/touched:_en_US.UTF-8_collation
fix-mutation: docs/prompt-audit-2026-09-29/audit-finish.md — none — a citation edit with no executable behaviour; check-references.sh exits 0
fix-mutations-total: 5

### Round 1

- primary re-ran alone on sonnet/low (rerun pair) on fix-round-0.diff + F1 site: F1 fixed, nothing new; principles raised only a Minor so not re-run
- primary re-dispatched under panel-1-primary-retry with -slot primary: the first round-1 dispatch row lacked its slot tag, so the closed guard could not see it; verdict F1 fixed, nothing new
## spectre/changes/archive/kan-873-agents-port-the-next-five-slowest-bash-guards-to/tasks.md

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

- [x] 1. Route flow lesson resolve through the record-family seam

**Files:** `stats/cmd/flow/lesson.go`, `stats/cmd/flow/lesson_test.go`, `.flow/project.md`, `stats/cmd/flow/record.go`, `stats/cmd/flow/state.go`
**Tests:** `TestRunLessonResolveRefusesDirFlag`
**Regression:** `TestRunLessonResolveRefusesDirFlag` fails when `-C` is accepted again; reverting the
commit also turns `scripts/test-flow-addr-declaration.sh`'s `pass-unmutated` red.
**Baseline:** before=5 after=6
<!-- measured: cat stats/cmd/flow/lesson_test.go | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `fix(stats): register flow lesson resolve's address through the record seam`
**Build:** green

**Decision:** lesson-resolve-through-record-seam

  - [x] **Step 1: Reproduce** before touching anything: `bash scripts/test-flow-addr-declaration.sh`
    prints `FAIL pass-unmutated: … found 2` and exits 1.
    <!-- measured: bash scripts/test-flow-addr-declaration.sh, exit 1, FAIL pass-unmutated @ 9cd35da8 -->
  - [x] **Step 2: Failing test.** `TestRunLessonResolveRefusesDirFlag`: `runLessonResolve` with
    `-topic x -C /tmp` exits 2, stderr carries `flow: lesson resolve takes no -C`, and no store call
    is made. Run — expect failure.
  - [x] **Step 3: Fix.** Replace lesson.go's own `-addr`/`-timeout` registrations with a
    `recordIdentityFlags` value registered through `registerRecordConnFlags` (a column-1-tab call
    line, as the harness's assertion 3 reads it); after parsing, a non-empty `dir` prints the line
    above plus `lessonUsage` and exits 2. Rewrite the comment that explained the old registration.
    In `.flow/project.md` `## workspace isolation`, the sentence becomes "The record family
    (`flow record`, `flow self-review bundle`, `flow lesson resolve`) resolves …", the parenthesis
    and verb kept on one physical line.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./cmd/flow/ && go test ./cmd/flow/
    -run '^TestRunLessonResolve' -count=1 -race`; `bash scripts/test-flow-addr-declaration.sh`
    exits 0 with six `ok` lines; `scripts/check-references.sh` exits 0.

Correction (2026-10-03): step 3's sentence shipped as "The record family (`flow record`, `flow lesson
resolve`, `flow self-review bundle`) resolves …" — appending the verb last turned the harness's
`mutate-declaration-drops-verb` case into a no-op (its sed anchors on `` `flow self-review bundle`) ``).
`**Files:**` widened by `stats/cmd/flow/record.go` and `stats/cmd/flow/state.go`: their doc comments
named `flow self-review bundle` as the only other seam user.

- [x] 2. Fix the SIGPIPE race in test-make-build.sh

**Files:** `scripts/test-make-build.sh`
**Tests:** none — repairs existing tests: `build target still compiles every package` and every other `assert_recipe_matches`/`assert_makefile_matches_all` case of scripts/test-make-build.sh
**Regression:** none — harness-internal; step 3 records what the race looked like
**Baseline:** before=0 after=0
<!-- predicted: no Go test is added by this task -->
**After:** none
**Commit:** `fix(scripts): read the make recipe without a pipe in test-make-build.sh`
**Build:** green

  - [x] **Step 1: Reproduce** before touching anything: run 2 of design.md's base ranking failed
    `build target still compiles every package` while printing a recipe that contains `go build
    ./...`. Mechanism: under `set -o pipefail`, `printf '%s\n' "$RECIPE" | grep -Eq` lets `grep -q`
    exit on its first match while `printf` is still writing; `printf` dies of SIGPIPE, the pipeline
    returns 141, and the `if` reads a match as a miss. Show it: a loop of 2000 iterations of
    `set -o pipefail; printf '%s\n' "$RECIPE" | grep -Eq '<case 3 pattern>'` with `RECIPE` padded
    to 200 KiB after the matching line counts non-zero exits; record the count.
    <!-- measured: 2000 non-zero of 2000 at 200 KiB (pipe form), 0 of 2000 (here-string); at the real 160-byte recipe 36 of 12000 under 12 concurrent loops, 0 of 2000 idle @ dc4f234c -->
  - [x] **Step 2: Fix** both `grep -q` sites — `assert_recipe_matches` and
    `assert_makefile_matches_all` — to read from a here-string (`grep -Eq "$pattern" <<<"$RECIPE"`,
    `<<<"$MAKEFILE_TEXT"`), no pipe; `assert_makefile_lacks` reads the whole input and stays.
    Comment the reason beside `assert_recipe_matches`.
  - [x] **Step 3: Prove the seam.** Step 1's loop with the here-string form: zero non-zero exits;
    with the pipe form restored: step 1's count again. Record both counts in the commit body.
  - [x] **Step 4: Verify.** `bash scripts/test-make-build.sh` exits 0, ten runs in a row.

- [x] 3. Port check-panel-docs-only

**Files:** `stats/internal/guard/paneldocsonly.go`, `stats/internal/guard/check_panel_docs_only_test.go`, `scripts/check-panel-docs-only.sh`, `scripts/test-check-panel-docs-only.sh`
**Tests:** `TestCheckPanelDocsOnly`
**Regression:** fails if any of the harness's 23 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_panel_docs_only_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port check-panel-docs-only to Go`
**Build:** green

**Decision:** scope-five-next-guards

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-panel-docs-only.sh`, one
    subtest per `ok:` label; fixtures are small git repos in `t.TempDir()`. Run — expect failure.
  - [x] **Step 2: Port**, registering `check-panel-docs-only`; the touched-path set through
    `paneltouchedpaths.go`, never re-derived.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckPanelDocsOnly$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckPanelDocsOnly/'` — at least 23.
  - [x] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines dropped
    unless step 2 needs a sibling; `git rm scripts/test-check-panel-docs-only.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/check-panel-docs-only.sh
    "$PWD" 9cd35da8` on this tree exits 1 printing the first non-doc path, directly and through a
    symlink to it in a temp directory; with no arguments it exits 2 with the usage line it printed
    at `9cd35da8`.

Correction (2026-10-03): step 5's symlink check runs through a symlink in a temp directory that also
links `lib/` beside it; a lone symlink has no `lib/` and refused at `9cd35da8` too (exit 1 there,
the shim's cannot-answer exit 2 now). The gated review's Minor — `panelGit` read the operator's
`~/.gitconfig` under test — is fixed at the branch tip (`GIT_CONFIG_GLOBAL` forwarded from `Env`).

- [x] 4. Port prepare-workspace

**Files:** `stats/internal/guard/prepareworkspace.go`, `stats/internal/guard/prepare_workspace_test.go`, `scripts/prepare-workspace.sh`, `scripts/test-prepare-workspace.sh`
**Tests:** `TestPrepareWorkspace`, `TestPrepareWorkspaceShimSibling`
**Regression:** fails if any of the harness's 25 `ok:` behaviours regress.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/prepare_workspace_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port prepare-workspace to Go`
**Build:** green

**Decision:** scope-five-next-guards

**Decision:** prepare-workspace-execs-sibling

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-prepare-workspace.sh`, one
    subtest per `ok:` label; the cases that replace or remove the sibling guard build their own
    scripts directory in `t.TempDir()` and point the guard at it the way the shim's
    `FLOW_GUARD_REPO_ROOT`/`FLOW_GUARD_SELF` would. Run — expect failure.
  - [x] **Step 2: Port**, registering `prepare-workspace`; the workspace id and every derived value
    through the Go code `check-workspace-isolation` and `stats/cmd/flow/workspaceid.go` already
    share where one exists, else ported from the bash; `check-workspace-isolation.sh` exec'd beside
    the shim, its stdout relayed verbatim and its stderr inherited, exit codes mapped exactly as the
    header states (126 included).
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestPrepareWorkspace$' -count=1 -race
    -v | grep -c -- '--- PASS: TestPrepareWorkspace/'` — at least 25.
  - [x] **Step 4: Shim and delete** — shim template, code 2, the sibling resolved as the bash did
    (`FLOW_GUARD_REPO_ROOT`, or `FLOW_GUARD_SELF` if the bash resolved its own symlink);
    `git rm scripts/test-prepare-workspace.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `scripts/prepare-workspace.sh
    "$PWD"` on this worktree prints the same `KEY=value` lines and exit code it printed at
    `9cd35da8`, directly and through a symlink to it in a temp directory.

Correction (2026-10-03): the shim spells `"$(dirname -- "${BASH_SOURCE[0]}")/check-workspace-isolation.sh"`
in its port note so `check-guard-symlinks.sh` rule 2 still sees the sibling
(`remove-change-worktrees.sh`'s precedent). The workspace id reuses `ccWorkspaceID`
(`cleanupcomplete.go`); `stats/cmd/flow/workspaceid.go` is package main. A fixture flake (a
`t.TempDir()` named after a label holding `=`, 2/5 runs) was pinned with `os.MkdirTemp`, 0/15 after.

- [x] 5. Port break-and-prove and delete lib/post-mutation-check.sh

**Files:** `stats/internal/guard/breakandprove.go`, `stats/internal/guard/break_and_prove_test.go`, `stats/internal/guard/post_mutation_check_test.go`, `stats/internal/guard/libtwins_test.go`, `stats/internal/guard/postmutationcheck.go`, `scripts/break-and-prove.sh`, `scripts/test-break-and-prove.sh`, `scripts/lib/post-mutation-check.sh`, `scripts/test-lib-post-mutation-check.sh`
**Tests:** `TestBreakAndProve`, `TestPostMutationCheck`, `TestBreakAndProveCleanBackgroundChild`
**Regression:** `TestBreakAndProve` fails if any of the harness's 38 `ok:` behaviours regress;
`TestPostMutationCheck` fails if any of `test-lib-post-mutation-check.sh`'s 4 behaviours regress.
**Baseline:** before=8 after=11
<!-- measured: cat stats/internal/guard/break_and_prove_test.go stats/internal/guard/post_mutation_check_test.go stats/internal/guard/libtwins_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port break-and-prove to Go and drop the bash post-mutation library`
**Build:** green

**Decision:** scope-five-next-guards

**Decision:** delete-post-mutation-check-lib

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-break-and-prove.sh`, one subtest
    per `ok:` label; port the 4 cases of `scripts/test-lib-post-mutation-check.sh` into
    `TestPostMutationCheck` in `post_mutation_check_test.go`, each pinning `snapshotTreeState` /
    `checkTreeRestored` output, stderr and status as fixed expected values (the bash library's
    output on the same fixture, captured once at `9cd35da8` and written into the test). Run —
    expect `TestBreakAndProve` to fail.
  - [x] **Step 2: Port**, registering `break-and-prove`; tree snapshot and restore check through
    `postmutationcheck.go`; `--clean`, `--sed` and `--patch` as the bash parses them; the restore
    on every exit path past the mutation; exit 3 over exit 2 as the header states; 126/127 from
    the test command mapped to 4.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^(TestBreakAndProve|TestPostMutationCheck)$'
    -count=1 -race -v | grep -c -- '--- PASS: TestBreakAndProve/'` — at least 38, and
    `TestPostMutationCheck` passes all 4.
  - [x] **Step 4: Shim and delete** — shim template, code 4; remove `TestPostMutationCheckParity`
    from `libtwins_test.go`; correct `postmutationcheck.go`'s header, which names the bash library
    as the source of truth; `git rm scripts/test-break-and-prove.sh
    scripts/lib/post-mutation-check.sh scripts/test-lib-post-mutation-check.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `mutate_and_verify_test.go`'s
    `bashAtBase` read of `lib/post-mutation-check.sh` still passes (it reads the file at the base
    commit, not the tree); `scripts/break-and-prove.sh` with no arguments exits 4 with the usage
    line it printed at `9cd35da8`, directly and through a symlink to it in a temp directory.

Correction (2026-10-03): step 4's shim drops the `FLOW_GUARD_REPO_ROOT` lines (the port reads only
its arguments and cwd). Step 5's symlink check runs through a symlink in a temp directory that also
links `lib/` beside it — every skill layout's shape; a bare symlink without `lib/` refuses with
`cannot load lib/flow-guard.sh` (exit 4), where `9cd35da8` printed the usage line. `TestPostMutationCheck`
carries the harness's 4 cases plus `TestPostMutationCheckParity`'s 10 rows as fixed values. Bash's
`<script>: line N:` diagnostic prefix becomes `break-and-prove: `; a test command found on PATH but
not runnable exits 127 rather than 126 (both map to 4).

- [x] 6. Port land-self-review-report

**Files:** `stats/internal/guard/landselfreviewreport.go`, `stats/internal/guard/land_self_review_report_test.go`, `scripts/land-self-review-report.sh`, `scripts/test-land-self-review-report.sh`, `skills/flow-self-review/scripts/lib`
**Tests:** `TestLandSelfReviewReport`
**Regression:** fails if any of the harness's 43 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/land_self_review_report_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port land-self-review-report to Go`
**Build:** green

**Decision:** scope-five-next-guards

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-land-self-review-report.sh`, one
    subtest per `ok:` label; each case's remote is a bare repository in `t.TempDir()`. Run — expect
    failure.
  - [x] **Step 2: Port**, registering `land-self-review-report`; the chain in the header's order —
    branch re-assertions before the add, before the commit and before the pull/push pair; the
    foreign-staged refusal (exit 3); git's own exit code passed through unmasked; every
    `LAND-*` line byte for byte; the bash's `g()` wrapper pins carried onto each git call.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestLandSelfReviewReport$' -count=1
    -race -v | grep -c -- '--- PASS: TestLandSelfReviewReport/'` — at least 43.
  - [x] **Step 4: Shim and delete** — shim template, code 2; `git rm
    scripts/test-land-self-review-report.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `bash
    scripts/test-git-config-pins.sh` and `bash scripts/test-protect-main-checkout.sh` exit 0;
    `scripts/land-self-review-report.sh` with no arguments exits 2 with the usage line it printed
    at `9cd35da8`, directly and through a symlink to it in a temp directory.

Correction (2026-10-03): `**Files:**` widened by `skills/flow-self-review/scripts/lib` — the shim
sources `lib/flow-guard.sh` beside itself and that skill directory had no `lib` symlink
(`check-guard-symlinks.sh` rule 2). Harness case 7's rejected push now targets a bare remote
refusing through its own `pre-receive` hook, per the plan's isolation rule.

Correction (2026-10-03): the gated review's Important — the guard's commit, pull and push read the
operator's `~/.gitconfig` under test (a global `core.hooksPath` failed 11 subtests) — is fixed by
forwarding `GIT_CONFIG_GLOBAL` from `Env`; a missing `sort` exits 127, the shell's status under
pipefail. Step 5's symlink check runs with `lib/` linked beside the shim, as task 3's does.

- [x] 7. Port check-self-review-report

**Files:** `stats/internal/guard/selfreviewreport.go`, `stats/internal/guard/check_self_review_report_test.go`, `scripts/check-self-review-report.sh`, `scripts/test-check-self-review-report.sh`
**Tests:** `TestCheckSelfReviewReport`, `TestCheckReferences`
**Regression:** fails if any of the harness's 63 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_self_review_report_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** none
**Commit:** `feat(stats): port check-self-review-report to Go`
**Build:** green

**Decision:** scope-five-next-guards

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-self-review-report.sh`,
    one subtest per `ok:` label; fixtures are small `docs/self-review/` trees in `t.TempDir()`. Run
    — expect failure.
  - [x] **Step 2: Port**, registering `check-self-review-report`; the angle table through
    `coverage.go`; the frozen `five-angle-reports.txt` list; an unreadable list or a failing
    directory walk is exit 2, never an empty set; the default directory resolved from the repo root
    as the bash's `$SCRIPT_DIR/..` did.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckSelfReviewReport$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckSelfReviewReport/'` — at least 63.
  - [x] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-self-review-report.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-self-review-report.sh` on this tree exits with the code and output it printed at
    `9cd35da8`, directly and through a symlink to it in a temp directory.

Correction (2026-10-03): the gated review's Important — a symlinked target without a trailing `/`
was walked, where `find -P` saw an empty corpus — is fixed, with a finding line carrying an
invalid UTF-8 byte now malformed under a UTF-8 locale and the report list sorted `sort -z`. The
usage line prints the literal `check-self-review-report.sh`, not `$0` (the shim execs flow-guard,
which never sees the invoked path), and an unreadable report loses bash's own
`<script>: line N: <file>: Permission denied` prefix line; both exits are unchanged.

- [x] 8. Repoint citations of the deleted files

**Files:** `.flow/project.md`
**Allowed-collateral:** `.flow/*.md`, `scripts/*.sh`, `scripts/lib/*.sh`, `scripts/*.py`, `hooks/*.py`, `skills/**/*.md`, `rules/*.mdc`, `README.md`, `CONTRIBUTING.md`, `stats/internal/guard/*.go`
**Tests:** none — citation sweep; the lint guards are the check
**Regression:** none — prose and comments only
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7, 10
**Commit:** `docs(scripts): repoint citations of the ported scripts to their Go sources`
**Build:** green

**Decision:** carry-prior-port-decisions

  - [x] **Step 1: Find.** `grep -rlF -e test-check-panel-docs-only.sh -e test-prepare-workspace.sh
    -e test-break-and-prove.sh -e test-land-self-review-report.sh -e
    test-check-self-review-report.sh -e post-mutation-check.sh --exclude-dir=archive
    --exclude-dir=.worktrees --exclude-dir=node_modules --exclude-dir=.git
    --exclude-dir=self-review .`; then grep the same tree for citations of a ported script's body
    comments and of its bash plumbing (`sources lib/…`).
    `unverified: the file set is known only after tasks 3–7 land; widen **Files:** by a correction if a hit falls outside the collateral globs`
  - [x] **Step 2: Repoint** each citation to the Go file or Go test that now holds what it cites; a
    sentence describing bash plumbing that no longer exists is corrected, not repointed.
    `.flow/project.md`'s paragraph naming which guards are Go lists the five new ones; library
    headers naming a ported script as a caller that sources them are corrected.
  - [x] **Step 3: Verify.** Step 1's grep returns only `spectre/changes/kan-873-*`,
    `docs/self-review/` and the Go ports' own history comments; every guard in `.flow/project.md`'s
    `## lint` exits 0.

- [x] 9. Live verification: after timings and parity

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8, 10
**Build:** green

  - [x] **Step 1: Suite, after.** `FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh`,
    three times on the branch head; record each run's wall, exit, harness count and slowest five.
  - [x] **Step 2: Parity.** `cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '---
    PASS: Test<Name>/'` per port against its floor in `design.md`.
  - [x] **Step 3: Record** an **After** table beside design.md's **Before**, same columns, each
    figure tagged `measured:` with the command and
    `@ branch spectre/kan-873-agents-port-the-next-five-slowest-bash-guards-to`; name the slowest
    remaining bash harness and the next slice.
  - [x] **Step 4: Judge.** Failure looks like: any harness red; any port's `--- PASS` count below
    its floor; the suite median above the Before median. Any of these is reported, not recorded as
    success.

- [x] 10. Shims refuse a missing lib/flow-guard.sh with their cannot-answer code under bash 3.2

**Files:** `stats/internal/guard/shim_template_test.go`
**Allowed-collateral:** `scripts/*.sh`
**Tests:** `TestShimMissingLibCannotAnswer`
**Regression:** `TestShimMissingLibCannotAnswer` fails for every shim whose source line loses the
readability check: `/bin/bash` 3.2 then exits 1 without the cannot-load line.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/shim_template_test.go 2>/dev/null | grep -cE '^func Test' @ 9cd35da8 -->
**After:** Task 3, 4, 5, 6, 7
**Commit:** `fix(scripts): shims refuse a missing flow-guard library under bash 3.2`
**Build:** green

Found by task 5's implementer and confirmed by the parent: under `/bin/bash` 3.2, a shim whose
`lib/flow-guard.sh` is missing exits 1 at the failed `.` without running its `|| { …; exit <code>; }`
block, so neither the cannot-load line nor the cannot-answer code reaches the caller; bash 5 runs
the block.
<!-- measured: ln -s scripts/check-panel-docs-only.sh <tmp>/x.sh; /bin/bash <tmp>/x.sh → rc 1, bash 5.3 → rc 2 @ ae96f2b3 -->

  - [x] **Step 1: Failing test.** `TestShimMissingLibCannotAnswer`: for every `scripts/*.sh` that
    calls `flow_guard_exec`, symlink it alone into `t.TempDir()`, run it with `/bin/bash`, and
    assert the exit is the code its cannot-load block names and stderr carries `cannot load
    lib/flow-guard.sh`. Run — expect failures.
  - [x] **Step 2: Fix** every shim's source line `. "<dir>/lib/flow-guard.sh" || {` to
    `[ -r "<dir>/lib/flow-guard.sh" ] && . "<dir>/lib/flow-guard.sh" || {`, the `<dir>` spelling
    kept so `check-guard-symlinks.sh` rule 2 still reads the sibling.
  - [x] **Step 3: Verify.** The test green; `scripts/check-guard-symlinks.sh`,
    `bash scripts/test-lib-flow-guard.sh` exit 0.
## spectre/changes/archive/kan-873-agents-port-the-next-five-slowest-bash-guards-to/design.md

# kan-873-agents-port-the-next-five-slowest-bash-guards-to

## Context

- Fifth slice of KAN-760's guard port (KAN-778, KAN-841, KAN-842, KAN-850 before it). Base
  `9cd35da8`. The pattern and every per-port convention are KAN-842's (`tasks.md` preamble).
- Ranking: three runs of `scripts/run-guard-tests.sh` at base, per-harness `(Ns)`, KAN-842's
  exclusions (`setup.sh`, Python-backed scripts, flow-guard's own harnesses, `workspace.sh`) plus
  every script already a `flow_guard_exec` shim.
  <!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh x3, grep '(Ns)' per harness @ 9cd35da8 -->

  | Script | run 1 | run 2 | run 3 | median |
  |---|--:|--:|--:|--:|
  | `land-self-review-report` | 10 | 11 | 11 | 11 |
  | `check-self-review-report` | 11 | 8 | 11 | 11 |
  | `break-and-prove` | 10 | 10 | 11 | 10 |
  | `prepare-workspace` | 10 | 9 | 8 | 9 |
  | `check-panel-docs-only` | 15 | 7 | 8 | 8 |
  | `check-worktree-location` | 9 | 7 | 7 | 7 |
  | `check-spec-reach` | 10 | 6 | 7 | 7 |
  | `check-normative-inventory` | 10 | 6 | 7 | 7 |

- Every base run exited 1 on one harness, `test-flow-addr-declaration.sh`: case `pass-unmutated` —
  "expected exactly one FLOW_RECORDS_ADDR -addr registration under
  stats/cmd/flow, found 2". The second is `stats/cmd/flow/lesson.go:66`, added by `79c2f025`.
  Suite wall 79s, 65s, 75s.
  <!-- measured: grep FAIL of each run's log @ 9cd35da8; grep -n '"addr", resolveRecordsAddr()' stats/cmd/flow/*.go -->
- Cannot-answer codes, from each header at base: `break-and-prove` 4, the other four 2.
  <!-- verified: the exit-code paragraph of each script's header @ 9cd35da8 -->
- Sibling helpers at base, each with a Go twin already:
  - `lib/coverage.sh` (`check-self-review-report`) → `coverage.go`; also sourced by
    `check-installed-citations.sh`, `check-guard-symlinks.sh` — stays.
  - `lib/post-mutation-check.sh` (`break-and-prove`) → `postmutationcheck.go`; no other runtime
    caller — deleted (`delete-post-mutation-check-lib`).
  - `lib/panel-touched-paths.sh` (`check-panel-docs-only`) → `paneltouchedpaths.go`; also
    `write-panel-diff.sh` — stays.
  - `check-workspace-isolation.sh` (`prepare-workspace`) — already a shim; exec'd as a sibling
    (`prepare-workspace-execs-sibling`).
  <!-- verified: grep -l 'lib/<name>.sh' scripts/*.sh, grep -n 'SCRIPT_DIR/' of each script @ 9cd35da8 -->

## Measurements

### Parity floors at base

| Harness | floor |
|---|--:|
| `test-land-self-review-report.sh` | 43 |
| `test-check-self-review-report.sh` | 63 |
| `test-break-and-prove.sh` | 38 |
| `test-prepare-workspace.sh` | 25 |
| `test-check-panel-docs-only.sh` | 23 |
| `test-lib-post-mutation-check.sh` | 4 |

- Floor is the harness's `^ok` line count on a green run; all six exited 0.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) bash scripts/test-<name>.sh, grep -c '^ok' of each log, all exit 0 @ 9cd35da8 -->

Each Go test carries at least its harness's floor in subtests (`parity-by-case-count`).

### Suite before/after

**Before** is the ranking's three runs; task 9 records **After** in the same columns.

**Before** — `@ 9cd35da8`:

| Run | suite wall | exit | harnesses | slowest five |
|---|--:|--:|--:|---|
| 1 | 79s | 1 | 46 | go-guards 78s, setup 37s, generate-relocation-comparison 37s, aside-planning-artifacts 22s, check-plan-shape 20s |
| 2 | 65s | 1 | 46 | go-guards 64s, generate-relocation-comparison 27s, setup 16s, lib-flow-guard 15s, run-guard-tests 14s |
| 3 | 75s | 1 | 46 | go-guards 73s, generate-relocation-comparison 29s, setup 18s, check-plan-shape 18s, lib-flow-guard 16s |
| **median** | **75s** | | | |

- Exit 1 in every run is `test-flow-addr-declaration.sh` (**Context**); run 2 also failed
  `test-make-build.sh`'s `build target still compiles every package` with a recipe that contains
  the line — a `pipefail` SIGPIPE race (`tasks.md` task 2).

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh x3, the runner's own "N harnesses, … Ns wall" line and grep '(Ns)' sorted descending @ 9cd35da8 -->

**After** — `@ branch spectre/kan-873-agents-port-the-next-five-slowest-bash-guards-to` (`ea3a5808`):

| Run | suite wall | exit | harnesses | slowest five |
|---|--:|--:|--:|---|
| 1 | 68s | 0 | 40 | go-guards 68s, generate-relocation-comparison 26s, setup 19s, lib-flow-guard 16s, check-plan-shape 16s |
| 2 | 71s | 0 | 40 | go-guards 71s, generate-relocation-comparison 22s, lib-flow-guard 12s, check-plan-shape 12s, run-guard-tests 11s |
| 3 | 54s | 0 | 40 | go-guards 53s, generate-relocation-comparison 16s, lib-flow-guard 13s, lib-parallel 11s, check-plan-shape 11s |
| **median** | **68s** | | | |

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh x3, the runner's own "N harnesses, … Ns wall" line and grep '(Ns)' sorted descending @ branch spectre/kan-873-agents-port-the-next-five-slowest-bash-guards-to -->

Parity (`--- PASS: Test<Name>/`, floor in brackets): CheckPanelDocsOnly 26 (23), PrepareWorkspace
32 (25), BreakAndProve 137 (38), PostMutationCheck 14 (4), LandSelfReviewReport 49 (43),
CheckSelfReviewReport 75 (63) — every port at or above its floor.

<!-- measured: cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '--- PASS: Test<Name>/' @ branch spectre/kan-873-agents-port-the-next-five-slowest-bash-guards-to -->

Judgement: every run green (the two base reds fixed); suite median 68s against 75s; six bash
harnesses gone (46 → 40). `test-go-guards.sh` remains the critical path. Every remaining harness at
or above 10s median is Python-backed (`generate-relocation-comparison`, `check-plan-shape`,
`check-task-records`), already a shim (`aside-planning-artifacts`) or a runner harness
(`lib-parallel`, `run-guard-tests`); the next slice re-ranks the sub-10s bash tail by this
change's three-run median method first.

## Decisions

### Carry the prior slices' port decisions unchanged

**ID:** carry-prior-port-decisions
**Status:** active
**Chosen:** KAN-760's `flow-guard-binary-rationale-in-go`, `go-tests-replace-harnesses`,
`parity-by-case-count`, `inject-deadlines-in-process`, `helper-parity-tests`,
`port-base-moves-into-go`, `guard-binary-built-from-checkout`, KAN-841's
`shared-helper-go-twins`, `exec-unported-siblings`, and KAN-842's
`exclude-workspace-sh` govern this slice as written; the base for `port-base-moves-into-go` is
`9cd35da8` — the operator's instruction (recommended options).
**Considered:** re-deciding each per script — no script here raises a case those decisions do not
cover.

### Scope: the five next bash guards

**ID:** scope-five-next-guards
**Status:** active
**Chosen:** port `land-self-review-report`, `check-self-review-report`, `break-and-prove`,
`prepare-workspace` and `check-panel-docs-only` — the five highest medians of three base runs
(**Context**).
**Considered:** a single run's ranking, as KAN-842 used — run 1 alone ties five scripts at 10s and
puts `check-panel-docs-only` first at 15s, a figure the next two runs halve; the median separates
the ties.

### lib/post-mutation-check.sh is deleted

**ID:** delete-post-mutation-check-lib
**Status:** active
**Chosen:** once `break-and-prove` is ported, delete `scripts/lib/post-mutation-check.sh` and
`scripts/test-lib-post-mutation-check.sh`; `TestPostMutationCheckParity` stops comparing against
the bash library and pins `snapshotTreeState`/`checkTreeRestored` against fixed expected output,
carrying the harness's four cases — KAN-842's `sha256-hex.sh` precedent.
**Considered:** keeping the library — a parity test against a library nothing runs pins the Go
twin to dead code.

### prepare-workspace execs its sibling guard

**ID:** prepare-workspace-execs-sibling
**Status:** active
**Chosen:** the Go port execs `check-workspace-isolation.sh` beside the shim, through
`FLOW_GUARD_REPO_ROOT`, as the bash did — its missing- and unrunnable-sibling cases keep their
meaning and output.
**Considered:** calling the registered Go guard in-process — saves one process start but makes the
sibling cases unreachable, so their harness cases could not be ported.

### flow lesson resolve joins the record family's seam

**ID:** lesson-resolve-through-record-seam
**Status:** active
**Chosen:** `runLessonResolve` registers `-addr`/`-timeout`/`-C` through
`registerRecordConnFlags`, exits 2 with `flow: lesson resolve takes no -C` when `-C` is given (the
read spans every project), and `.flow/project.md`'s record-family sentence names
`flow lesson resolve` — the harness's five assertions hold unchanged.
**Considered:** loosening the harness to allow a second registration — weakens a guard (Lint Fix
Priority); accepting `-C` silently — a flag that does nothing, which `79c2f025`'s own comment
rejects.

### prepare-workspace's sibling path comes from the shim

**ID:** prepare-workspace-sibling-path-exported
**Status:** active — supersedes `prepare-workspace-execs-sibling`'s "through `FLOW_GUARD_REPO_ROOT`"
**Chosen:** the shim exports `FLOW_GUARD_WORKSPACE_ISOLATION="$SCRIPT_DIR/check-workspace-isolation.sh"`
and the Go port execs that path, as `check-finish-preflight`'s `FLOW_GUARD_WORKTREE_LOCATION` and
`fold-fixup`'s `FLOW_GUARD_AUTOSQUASH` do — the bash's `$SCRIPT_DIR`, whatever the directory is
named.
**Considered:** `<FLOW_GUARD_REPO_ROOT>/scripts` — agrees with the bash only while the shim's
directory is named `scripts` (panel round 0, P2); `FLOW_GUARD_SELF` with `guardSelfDir` — resolves
the shim's symlink, which the bash's `$SCRIPT_DIR` never did.

## Open questions
## spectre/changes/archive/kan-873-agents-port-the-next-five-slowest-bash-guards-to/narrative.md

# kan-873-agents-port-the-next-five-slowest-bash-guards-to — session narrative

## 2026-10-03 — creating run

- The operator asked for a Jira task and an unattended run "till the handoff", taking every
  recommended option. The run read that as the `IN_PROGRESS` handoff and continued past the
  `STARTED` plan gate into implementation in the same session.
- The base was red before any edit: `test-flow-addr-declaration.sh` failed on every run, because
  `79c2f025` added a second `-addr` registration in `flow lesson resolve`. `test-make-build.sh` also
  raced under `pipefail` (SIGPIPE from `grep -q` reading a pipe). Both were fixed in this change,
  as tasks 1 and 2, rather than worked around.
- The class was raised from regular to big partway through planning, once the five ports' parity
  surface was measured. The largest is `break-and-prove`, at 137 subtests against a floor of 38.
- A defect outside the five ports surfaced: under macOS `/bin/bash` 3.2, a shim whose
  `lib/flow-guard.sh` is missing exits 1 before its `|| { …; exit <code>; }` block runs. Task 10
  fixed the shim template in all 68 shims and added a test that runs each one without its library.
- Gated per-task reviews raised two Importants:
  - task 7: a symlinked target without a trailing slash;
  - task 6: the test read the operator's `~/.gitconfig`.

  Both were fixed and re-reviewed clean.
- Panel round 0 raised one Important and three Minors:
  - Important: a `--clean` command that backgrounds a child stalled each `break-and-prove` leg.
  - Minor: `prepare-workspace` looked its sibling up through a literal `scripts/`.
  - Minor: a dated audit file still cited a deleted harness.
  - Minor: a duplicated sort setup.

  One fix commit closed all four.
- **Time sink:** each reviewer's reproducers pinned their `# premise:` lines to the exact lines the
  fix rewrote, so the pinned re-runs could only refuse. They were re-authored without those
  premises and proved both ways with `prove-reproducer.sh` against the pre-fix commit. One
  reproducer (`0-P3-1.sh`) could never have passed as written: its body tested for the deleted file
  rather than the citation.
- **Record gap:** the first round-1 re-review dispatch was recorded without `-slot`. `flow record
  dispatch begin` does not update the slot on a repeat call with the same key, so
  `check-panel-findings-closed.sh` could not see the re-run. The re-review was dispatched again
  under the `-retry` key with the slot set.
- **Plan-record mismatch:** the fix subagent added the fix commit's paths to tasks 7 and 8's
  `**Files:**` fields. `check-task-records.sh` correctly refused this, because those tasks' own
  commits do not touch those paths. The additions were reverted, since the fix commit carries them.
- **Flakes:** the first verify run of `scripts/run-guard-tests.sh` failed in two harnesses this
  change does not touch, `test-run-guard-tests.sh` (case 2's summary count) and
  `test-check-done-when-paths.sh`. Both passed alone, and the whole suite passed on its one
  re-run. The cause was not investigated.

## 2026-10-03 — integrate run

- **Preflight:** `RUN1`; main checkout `STAGED-CLEAN` and `DRIFT-CLEAN`.
- **Unfinished-work gate:** `CLEAR`; visual verify not required (no UI paths).
- **Base:** `origin/main` had not moved since the recorded merge base — no rebase.
- **Route:** merge and push, from the project's configured default.
## git log --stat

commit 8460e63b5ec59c237a6c015c257bc47e34bb3051
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 15:21:14 2026 +0300

    feat(stats): port prepare-workspace, check-self-review-report, break-and-prove, check-panel-docs-only and land-self-review-report to Go

 .flow/project.md                                   |    4 +-
 docs/prompt-audit-2026-09-29/audit-finish.md       |    2 +-
 scripts/aside-planning-artifacts.sh                |    2 +-
 scripts/break-and-prove.sh                         |  281 +-----
 scripts/check-archive-scope.sh                     |    2 +-
 scripts/check-base-moved.sh                        |    2 +-
 scripts/check-cleanup-complete.sh                  |    2 +-
 scripts/check-dispatch-paragraphs.sh               |    2 +-
 scripts/check-done-when-paths.sh                   |    2 +-
 scripts/check-fast-route-record.sh                 |    2 +-
 scripts/check-finish-preflight.sh                  |    2 +-
 scripts/check-foreign-staged.sh                    |    2 +-
 scripts/check-guard-symlinks.sh                    |    2 +-
 scripts/check-hand-notes-in-step.sh                |    2 +-
 scripts/check-installed-citations.sh               |    2 +-
 scripts/check-installed-rules.sh                   |    2 +-
 scripts/check-late-fix-trigger.sh                  |    2 +-
 scripts/check-main-checkout-drift.sh               |    2 +-
 scripts/check-panel-citation-trigger.sh            |    2 +-
 scripts/check-panel-docs-only.sh                   |   41 +-
 scripts/check-panel-findings-closed.sh             |    2 +-
 scripts/check-panel-fix-single-dispatch.sh         |    2 +-
 scripts/check-panel-reproducer-exit-contract.sh    |    2 +-
 scripts/check-panel-reproducers.sh                 |    2 +-
 scripts/check-plan-provenance.sh                   |    2 +-
 scripts/check-planning-commit-location.sh          |    2 +-
 scripts/check-references.sh                        |    2 +-
 scripts/check-review-gate.sh                       |    2 +-
 scripts/check-self-review-report.sh                |  432 +--------
 scripts/check-stage-mark-calls.sh                  |    2 +-
 scripts/check-task-commit-fields.sh                |    2 +-
 scripts/check-task-commit-planning-paths.sh        |    2 +-
 scripts/check-task-reviewer-single-dispatch.sh     |    2 +-
 scripts/check-unfinished-work.sh                   |    2 +-
 scripts/check-verbatim-moves.sh                    |    2 +-
 scripts/check-verify-report.sh                     |    2 +-
 scripts/check-visual-preflight.sh                  |    2 +-
 scripts/check-visual-trigger.sh                    |    2 +-
 scripts/check-visual-verification.sh               |    2 +-
 scripts/check-visual-verify-dispatched.sh          |    2 +-
 scripts/check-workspace-isolation.sh               |    2 +-
 scripts/classify-untracked.sh                      |    2 +-
 scripts/close-task.sh                              |    2 +-
 scripts/commit-archive.sh                          |    2 +-
 scripts/commit-split.sh                            |    2 +-
 scripts/compose-mockup-frames.sh                   |    2 +-
 scripts/fold-fixup.sh                              |    2 +-
 scripts/gather-dispatch-context.sh                 |    2 +-
 scripts/kickoff-worktree.sh                        |    2 +-
 scripts/land-self-review-report.sh                 |   86 +-
 scripts/lib/panel-touched-paths.sh                 |    8 +-
 scripts/lib/post-mutation-check.sh                 |   81 --
 scripts/measure-visual-properties.sh               |    2 +-
 scripts/mutate-and-verify.sh                       |    2 +-
 scripts/plan-class.sh                              |    2 +-
 scripts/prepare-archive-branch.sh                  |    2 +-
 scripts/prepare-workspace.sh                       |  240 +----
 scripts/project-get.sh                             |    2 +-
 scripts/prove-reproducer.sh                        |    2 +-
 scripts/recover-guard-incident.sh                  |    2 +-
 scripts/refresh-main-checkout.sh                   |    2 +-
 scripts/refresh-plan-base.sh                       |    2 +-
 scripts/remove-change-worktrees.sh                 |    2 +-
 scripts/render-slot-prompt.sh                      |    2 +-
 scripts/reshape-branch.sh                          |    2 +-
 scripts/resolve-base-branch.sh                     |    2 +-
 scripts/resolve-visual-screenshots.sh              |    2 +-
 scripts/run-reproducer.sh                          |    2 +-
 scripts/sync-onto-base.sh                          |    2 +-
 scripts/sync-panel-base.sh                         |    2 +-
 scripts/test-break-and-prove.sh                    |  294 ------
 scripts/test-check-panel-docs-only.sh              |  247 -----
 scripts/test-check-self-review-report.sh           | 1010 --------------------
 scripts/test-flow-active-change-hook.sh            |    2 +-
 scripts/test-git-config-pins.sh                    |    2 +-
 scripts/test-installer-sandbox-diff.sh             |    2 +-
 scripts/test-land-self-review-report.sh            |  417 --------
 scripts/test-lib-post-mutation-check.sh            |   89 --
 scripts/test-make-build.sh                         |   10 +-
 scripts/test-prepare-workspace.sh                  |  342 -------
 scripts/throwaway-worktree.sh                      |    2 +-
 scripts/write-panel-diff.sh                        |    2 +-
 skills/flow-self-review/scripts/lib                |    1 +
 stats/cmd/flow/lesson.go                           |   34 +-
 stats/cmd/flow/lesson_test.go                      |   34 +
 stats/cmd/flow/record.go                           |    4 +-
 stats/cmd/flow/state.go                            |    2 +-
 stats/internal/guard/break_and_prove_test.go       |  523 ++++++++++
 stats/internal/guard/breakandprove.go              |  508 ++++++++++
 stats/internal/guard/check_panel_docs_only_test.go |  264 +++++
 .../guard/check_self_review_report_test.go         |  495 ++++++++++
 stats/internal/guard/helpers_test.go               |    9 +
 .../internal/guard/land_self_review_report_test.go |  549 +++++++++++
 stats/internal/guard/landselfreviewreport.go       |  221 +++++
 stats/internal/guard/libtwins_test.go              |   97 --
 stats/internal/guard/paneldocsonly.go              |   49 +
 stats/internal/guard/paneltouchedpaths.go          |    8 +
 stats/internal/guard/post_mutation_check_test.go   |  151 +++
 stats/internal/guard/postmutationcheck.go          |   17 +-
 stats/internal/guard/prepare_workspace_test.go     |  296 ++++++
 stats/internal/guard/prepareworkspace.go           |  290 ++++++
 stats/internal/guard/references.go                 |   18 +-
 stats/internal/guard/selfreviewreport.go           |  434 +++++++++
 stats/internal/guard/shim_template_test.go         |   69 ++
 104 files changed, 4110 insertions(+), 3679 deletions(-)

commit 58741e2d7dafb3077e7aeee90078d614ccbfa440
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 15:21:15 2026 +0300

    chore(spectre): plan

 .../narrative.md                                                   | 7 +++++++
 1 file changed, 7 insertions(+)

commit a77dd87162f2447b585493e314a051a62dda4297
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 15:22:39 2026 +0300

    chore(spectre): archive kan-873-agents-port-the-next-five-slowest-bash-guards-to

 .../design.md                                      |   0
 .../ledger.md                                      | 205 +++++++++++++++++++++
 .../narrative.md                                   |   0
 .../panel.md                                       |  47 +++++
 .../proposal.md                                    |   0
 .../tasks.md                                       |   0
 6 files changed, 252 insertions(+)

## Session narrative

Run 2 ran as the merge-and-push continuation of a clean run 1 (preflight `RUN1`, unfinished-work `CLEAR`, base unmoved, route from the project default). The first two stage-begin marks of run 2 were issued with the token held in a shell variable, which the CLI rejected as one malformed flag; they were re-issued with the literal token. Worktree cleanup's disclosure stop listed only Python bytecode caches and a TypeScript build-info file as unclassified — regeneratable, so the run proceeded. The workspace `remove` found no `flow_kan_873_557e` database to drop, and the survivor report came back empty.
