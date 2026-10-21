# kan-918-flow-improvement-port-the-git-snapshot-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Ports five bash guards into `flow-guard` (tasks 1–5), sweeps the citations of the deleted
harnesses and the bash plumbing (task 6) and records the after measurement (task 7). `design.md`
is canonical for every decision; each task cites its entry by ID.

**The script at `c9b8fde8` is each port's specification.** Its header comment states the
contract; its body is the behaviour. A port reproduces arguments, environment overrides, every
output line byte for byte, the stdout/stderr split, every side effect on git/the filesystem, and
every exit code. Where the source and its header disagree, stop and report — never pick one
silently.

**Parity floors** are in `design.md`'s **Measurements** (**Decision:**
carry-prior-port-decisions — KAN-760's `parity-by-case-count`).

**Every port task (1–5) follows the same five steps:** port the harness's cases to a Go table test
first (red — the guard is unregistered), port the script, run the test green, replace the script's
body with the shim, `git rm` the harness. Each Go subtest is named after the harness case's `ok:`
label, so parity is a `--- PASS` count. The script body's comments move into the Go file beside the
code they explain, updated where the mechanism changed (KAN-760's
`flow-guard-binary-rationale-in-go`). The Go file is
`stats/internal/guard/<name without - and without check->.go`; the test file is
`stats/internal/guard/<name with - replaced by _>_test.go` — the only name
`scripts/run-guard-tests.sh`'s companion rule accepts for a `check-*` shim, used for
`guard-autosquash` too for one convention. Each port registers its basename in `guard.Registry`
from its own file's `init()`, as `basemoved.go` does.

**Shim template** — the header comment block kept verbatim, corrected only where the port makes a
statement false (bash-specific plumbing), then:

```bash verified:copied from the tail of scripts/check-base-moved.sh @ c9b8fde8
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "<name>: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec <name> 2 "<name>:" "$@"
```

`<name>` is the basename without `.sh`, written literally. No shim exports `FLOW_GUARD_REPO_ROOT`:
none of the five derives a path from its own location (`design.md` **Context**).

**Git config pins.** Every git call a port makes carries the pins its bash call carried (R1–R3 of
`scripts/test-git-config-pins.sh`, whose case 7 audits `stats/internal/guard/*.go`); git runs
through `gitExec` (`postmutationcheck.go`) or the port's existing equivalent.

**Test isolation:** every Go test calls `t.Parallel()`; fixture trees (git repos included) are
built once in `TestMain` or a `sync.Once` helper (`helpers_test.go` has the existing ones) and each
case copies its own into `t.TempDir()`; deadlines are injected on `Env`, never through the
environment (KAN-760's `inject-deadlines-in-process`). No test reads or writes the operator's real
`$HOME`.

**Per-task verify** is `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
./internal/guard/ -run '^<Test>$' -count=1 -race`, then the shimmed script run once for real as
its step 5 names. The full package and `scripts/run-guard-tests.sh` run in task 7 and
`flow.verify`.

## Review Focus

- A shim run from an installed skill directory (`~/.claude/skills/flow/scripts/<name>.sh`, a
  symlink) must behave as the bash did — each port's step 5 runs the shim through a symlink in a
  temp directory.
- `check-plan-unchanged` and `check-tree-markers` write a snapshot then verify against it: the
  snapshot file's bytes and location stay the bash's, so a snapshot written before the port
  verifies after it.
- `guard-autosquash` is exec'd by `fold-fixup` (`foldfixup.go`): its refusal lines stay on stderr
  and its exit codes stay 0/1/2, or fold-fixup misreads them.
- `check-worktree-processes` parses `lsof -F` output: physical-path comparison (`pwd -P`
  semantics) and the missing-`lsof` exit 2 keep the bash's verdicts.
- Locale-sensitive ordering: any `sort`/`sort -u`/`comm` in a bash body keeps the caller's
  collation.

---

- [ ] 1. Port check-plan-unchanged

**Files:** `stats/internal/guard/planunchanged.go`, `stats/internal/guard/check_plan_unchanged_test.go`, `scripts/check-plan-unchanged.sh`, `scripts/test-check-plan-unchanged.sh`
**Tests:** `TestCheckPlanUnchanged`
**Regression:** fails if any of the harness's 30 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_plan_unchanged_test.go 2>/dev/null | grep -cE '^func Test' @ c9b8fde8 -->
**After:** none
**Commit:** `feat(guard): port check-plan-unchanged to Go`
**Build:** green

**Decision:** scope-five-issue-guards

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-plan-unchanged.sh`, one
    subtest per `ok:` label (cases 14a/14b on an archived change included). Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-plan-unchanged`; snapshot file format and location
    byte-identical to the bash's.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestCheckPlanUnchanged$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckPlanUnchanged/'` — at least 30.
  - [ ] **Step 4: Shim and delete** — shim template, code 2;
    `git rm scripts/test-check-plan-unchanged.sh`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`; a snapshot then
    verify of this worktree through `scripts/check-plan-unchanged.sh` exits as the bash did at
    `c9b8fde8`, directly and through a symlink to it in a temp directory.

- [ ] 2. Port check-tree-markers

**Files:** `stats/internal/guard/treemarkers.go`, `stats/internal/guard/check_tree_markers_test.go`, `scripts/check-tree-markers.sh`, `scripts/test-check-tree-markers.sh`
**Tests:** `TestCheckTreeMarkers`
**Regression:** fails if any of the harness's 30 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_tree_markers_test.go 2>/dev/null | grep -cE '^func Test' @ c9b8fde8 -->
**After:** none
**Commit:** `feat(guard): port check-tree-markers to Go`
**Build:** green

**Decision:** scope-five-issue-guards

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-tree-markers.sh`, one
    subtest per `ok:` label (the absolute-path, empty-path and empty-marker refusals included). Run
    — expect failure.
  - [ ] **Step 2: Port**, registering `check-tree-markers`; markers-file parsing and every exit-2
    message byte-identical.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestCheckTreeMarkers$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckTreeMarkers/'` — at least 30.
  - [ ] **Step 4: Shim and delete** — shim template, code 2;
    `git rm scripts/test-check-tree-markers.sh`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`; a snapshot then
    verify on a temp markers file through `scripts/check-tree-markers.sh` exits as the bash did at
    `c9b8fde8`, directly and through a symlink.

- [ ] 3. Port guard-autosquash

**Files:** `stats/internal/guard/guardautosquash.go`, `stats/internal/guard/guard_autosquash_test.go`, `scripts/guard-autosquash.sh`, `scripts/test-guard-autosquash.sh`, `stats/internal/guard/foldfixup.go`
**Tests:** `TestGuardAutosquash`
**Regression:** fails if any of the harness's 12 `ok:` behaviours regress; the existing TestFoldFixup, run unchanged, fails if
fold-fixup stops reading the shim's verdicts.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/guard_autosquash_test.go 2>/dev/null | grep -cE '^func Test' @ c9b8fde8 -->
**After:** none
**Commit:** `feat(guard): port guard-autosquash to Go`
**Build:** green

**Decision:** fold-fixup-execs-autosquash-shim

Correction (2026-10-21): the plan declared `TestFoldFixup` under **Tests:**; the commit leaves that existing test unchanged and runs it, so it moved to **Regression:**.

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-guard-autosquash.sh`, one
    subtest per `ok:` label (`targets` and `after` subcommands, usage exits). Run — expect failure.
  - [ ] **Step 2: Port**, registering `guard-autosquash`; `targets`/`after` subcommands, the
    `usage` and `cannot_answer` lines byte-identical.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestGuardAutosquash$'
    -count=1 -race -v | grep -c -- '--- PASS: TestGuardAutosquash/'` — at least 12; `go test
    ./internal/guard/ -run '^TestFoldFixup' -count=1 -race` passes.
  - [ ] **Step 4: Shim and delete** — shim template, code 2;
    `git rm scripts/test-guard-autosquash.sh`; `foldfixup.go`'s "guard-autosquash stays bash"
    comment corrected to name the shim it execs.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/guard-autosquash.sh targets <this worktree> HEAD` exits as the bash did at
    `c9b8fde8`, directly and through a symlink.

- [ ] 4. Port check-panel-diff-size

**Files:** `stats/internal/guard/paneldiffsize.go`, `stats/internal/guard/check_panel_diff_size_test.go`, `scripts/check-panel-diff-size.sh`, `scripts/test-check-panel-diff-size.sh`
**Tests:** `TestCheckPanelDiffSize`
**Regression:** fails if any of the harness's 8 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_panel_diff_size_test.go 2>/dev/null | grep -cE '^func Test' @ c9b8fde8 -->
**After:** none
**Commit:** `feat(guard): port check-panel-diff-size to Go`
**Build:** green

**Decision:** scope-five-issue-guards

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-panel-diff-size.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-panel-diff-size`.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestCheckPanelDiffSize$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckPanelDiffSize/'` — at least 8.
  - [ ] **Step 4: Shim and delete** — shim template, code 2;
    `git rm scripts/test-check-panel-diff-size.sh`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/check-panel-diff-size.sh` on this worktree with the arguments its header's usage names
    exits as the bash did at `c9b8fde8`, directly and through a symlink.

- [ ] 5. Port check-worktree-processes

**Files:** `stats/internal/guard/worktreeprocesses.go`, `stats/internal/guard/check_worktree_processes_test.go`, `scripts/check-worktree-processes.sh`, `scripts/test-check-worktree-processes.sh`
**Tests:** `TestCheckWorktreeProcesses`
**Regression:** fails if any of the harness's 27 `ok:` behaviours regress, the KAN-197
always-CLEAR mutation case included.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_worktree_processes_test.go 2>/dev/null | grep -cE '^func Test' @ c9b8fde8 -->
**After:** none
**Commit:** `feat(guard): port check-worktree-processes to Go`
**Build:** green

**Decision:** worktree-processes-execs-lsof

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-worktree-processes.sh`,
    one subtest per `ok:` label; real child processes with their cwd in a `t.TempDir()` worktree,
    every path compared physically (`filepath.EvalSymlinks`), each child killed in `t.Cleanup`; the
    missing-`lsof` case runs with a `PATH` on `Env` that holds no `lsof`. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-worktree-processes`; `lsof` resolved through `Env`'s
    `PATH`, its `-F` output parsed as the bash did.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run
    '^TestCheckWorktreeProcesses$' -count=1 -race -v | grep -c -- '--- PASS:
    TestCheckWorktreeProcesses/'` — at least 27.
  - [ ] **Step 4: Shim and delete** — shim template, code 2;
    `git rm scripts/test-check-worktree-processes.sh`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/check-worktree-processes.sh <this worktree>` exits as the bash did at `c9b8fde8`,
    directly and through a symlink.

- [ ] 6. Repoint citations of the deleted harnesses and the bash plumbing

**Files:** `.flow/project.md`
**Allowed-collateral:** `.flow/*.md`, `scripts/*.sh`, `scripts/lib/*.sh`, `scripts/*.py`, `hooks/*.py`, `skills/**/*.md`, `rules/*.mdc`, `README.md`, `CONTRIBUTING.md`, `stats/internal/guard/*.go`
**Tests:** none — citation sweep; the lint guards are the check
**Regression:** none — prose and comments only
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5
**Commit:** `docs(scripts): repoint citations of the ported guards to their Go sources`
**Build:** green

**Decision:** carry-prior-port-decisions

  - [ ] **Step 1: Find.** `grep -rlF -e test-check-plan-unchanged.sh -e test-check-tree-markers.sh
    -e test-guard-autosquash.sh -e test-check-panel-diff-size.sh -e
    test-check-worktree-processes.sh --exclude-dir=archive --exclude-dir=.worktrees
    --exclude-dir=node_modules --exclude-dir=.git --exclude-dir=self-review .`; then grep the same
    tree for citations of a ported script's body comments.
    `unverified: the file set is known only after tasks 1–5 land; widen **Files:** by a correction if a hit falls outside the collateral globs`
  - [ ] **Step 2: Repoint** each citation to the Go file or Go test that now holds what it cites; a
    sentence describing bash plumbing that no longer exists is corrected, not repointed.
    `.flow/project.md`'s paragraph naming which guards are Go lists the five new ones.
  - [ ] **Step 3: Verify.** Step 1's grep returns only `spectre/changes/kan-918-*`,
    `docs/self-review/` and the Go ports' own history comments; every guard in `.flow/project.md`'s
    `## lint` exits 0.

- [ ] 7. Live verification: after timings and parity

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6
**Build:** green

  - [ ] **Step 1: Suite, after.** `FLOW_GUARD_CACHE_DIR=$(mktemp -d) scripts/run-guard-tests.sh`,
    three times on the branch head; record each run's wall, exit, harness count and slowest five.
  - [ ] **Step 2: Parity.** `cd stats && go test ./internal/guard/ -run '^Test<Name>$' -count=1 -v
    | grep -c -- '--- PASS: Test<Name>/'` per port against its floor in `design.md`.
  - [ ] **Step 3: Record** an **After** table beside design.md's **Before**, same columns, each
    figure tagged `measured:` with the command and
    `@ branch spectre/kan-918-flow-improvement-port-the-git-snapshot-and`.
  - [ ] **Step 4: Judge.** Failure looks like: any harness red; any port's `--- PASS` count below
    its floor; the suite median above the Before median. Any of these is reported, not recorded as
    success.
