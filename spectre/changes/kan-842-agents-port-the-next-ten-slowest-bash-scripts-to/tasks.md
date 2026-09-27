# kan-842-agents-port-the-next-ten-slowest-bash-scripts-to

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Ports ten more bash scripts into `flow-guard`, in dependency order: the before measurement
(task 1), the Go twins of two shared bash libraries (task 2), the ten ports (tasks 3–12), the
citation sweep (task 13), and the after measurement (task 14). `design.md` is canonical for every
decision; each task cites its entry by ID.

**The script at `c5379c0a` is each port's specification.** Its header comment states the
contract; its body is the behaviour. A port reproduces arguments, environment overrides, every
output line byte for byte, the stdout/stderr split, every side effect on git/the filesystem, and
every exit code. Where the source and its header disagree, stop and report — never pick one
silently.

**Parity floors** are in `design.md`'s **Measurements** (**Decision:**
carry-prior-port-decisions — KAN-760's `parity-by-case-count`).

**Every port task (3–12) follows the same five steps:** port the harness's cases to a Go table
test first (red — the guard is unregistered), port the script, run the test green, replace the
script's body with the shim, `git rm` the harness. Each Go subtest is named after the harness
case's `ok:` label, so parity is a `--- PASS` count. The script body's comments move into the Go
file beside the code they explain, updated where the mechanism changed (KAN-760's
`flow-guard-binary-rationale-in-go`). The Go file is
`stats/internal/guard/<name without - and without check->.go`; the test file is
`stats/internal/guard/<name with - replaced by _>_test.go` — the only name
`scripts/run-guard-tests.sh`'s companion rule accepts for a `check-*` shim, used for the three
non-`check-` scripts too for one convention. Each port registers its basename in `guard.Registry`
from its own file's `init()`, as `basemoved.go` does.

**Shim template** — the header comment block kept verbatim, corrected only where the port makes a
statement false (`$SCRIPT_DIR/<lib>` sourcing, bash-specific plumbing), then:

```bash verified:copied from the tail of scripts/check-references.sh @ d71a2327, unchanged @ c5379c0a
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
exit: `1` for `check-installed-rules`, `2` for the other nine (`design.md` **Context**). The two
`FLOW_GUARD_REPO_ROOT` lines are kept only where the bash derived a root or sibling path from its
own location (`$SCRIPT_DIR/..`, `$SCRIPT_DIR/<sibling>`); a script that reads only its arguments
and cwd drops them.

**Test isolation:** every Go test calls `t.Parallel()`; fixture trees (git repos included) are
built once in `TestMain` or a `sync.Once` helper (`helpers_test.go` has the existing ones) and each
case copies its own into `t.TempDir()`; a stub `flow` on PATH becomes the injected `Env` hook the
existing ports use where the script reads the store; deadlines are injected on `Env`, never
through the environment (KAN-760's `inject-deadlines-in-process`). No test reads or writes the
operator's real `$HOME`: `check-installed-rules` cases set `CHECK_INSTALLED_RULES_HOME` to a temp
directory, as its harness does.

**Per-task verify** is `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
./internal/guard/ -run '^<Test>$' -count=1 -race`, then the shimmed script run once for real as
its step 5 names. The full package and `scripts/run-guard-tests.sh` run in task 14 and
`flow.verify`.

Live verification: tasks 1 and 14 run the real suite on this machine and record before/after.

## Review Focus

- A shim run from an installed skill directory (`~/.claude/skills/flow/scripts/<name>.sh`, a
  symlink) must resolve the same repository root the bash did — each port's step 5 runs the shim
  through a symlink in a temp directory where the bash also worked there.
- Locale-sensitive ordering: any `sort`/`sort -u`/`comm` in a bash body keeps the caller's
  collation (KAN-778 task 5's exec'd `sort`), pinned by a subtest with mixed-case names.
- `recover-guard-incident --apply` mutates git (aborts a revert, restores stash files): every
  refusal exit leaves `git status --porcelain`, `HEAD` and the stash list exactly as the bash did,
  and the abort-before-restore order is pinned by a subtest.
- `plan-class` prints the rolls `/flow`'s Decide step records: a subtest pins the three output
  lines byte for byte against the bash at `c5379c0a` for a fixed change name, so the name-derived
  rolls do not move.
- Exec'd siblings (`plan-dispatch-bundles.sh`, `plan-dispatch-groups.sh`,
  `check-visual-trigger.sh`) are resolved beside the shim through `FLOW_GUARD_REPO_ROOT`, with the
  same arguments, stdout/stderr capture and exit handling as the bash.

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

  - [ ] **Step 1: Suite, before.** On this machine at `c5379c0a`, in a checkout without the main
    checkout's untracked `skills/flow/scripts/guard-autosquash.sh` (`design.md` **Context**):
    `sysctl -n vm.loadavg` then `FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p
    scripts/run-guard-tests.sh`, three times; record each run's real/user/sys, load, exit, harness
    count and the slowest five harnesses (`grep '(Ns)'` of each log, sorted descending).
  - [ ] **Step 2: Go package, before.** `cd stats && /usr/bin/time -p go test
    ./internal/guard/... -count=1` three times; record real/user/sys and load.
  - [ ] **Step 3: Record** a **Before** table under `design.md`'s **Measurements** → **Suite
    before/after**, KAN-841's columns, each figure tagged `measured:` with the command and
    `@ c5379c0a`.

This task commits nothing; its figures are committed with the change's artifacts.

Correction (2026-09-27): the plan's wave order would dispatch this task beside the first wave of
port implementers; it runs instead after task 13 lands and immediately before task 14, still at
`c5379c0a` in a detached checkout, so Before and After are both measured on an otherwise idle
machine rather than Before under three concurrent `go test -race` runs — a Before inflated by
that load would bias `suite-median-below-before` toward passing.

- [ ] 2. Go twins of panel-touched-paths and owned-corpus

**Files:** `stats/internal/guard/paneltouchedpaths.go`, `stats/internal/guard/ownedcorpus.go`, `stats/internal/guard/libtwins_test.go`
**Tests:** `TestPanelTouchedPathsParity`, `TestOwnedCorpusParity`
**Regression:** each parity test fails if its Go twin's output, return status or stderr differs
from `scripts/lib/panel-touched-paths.sh` or `scripts/lib/owned-corpus.sh` for the same inputs.
**Baseline:** before=4 after=6
<!-- measured: cat stats/internal/guard/libtwins_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** none
**Commit:** `feat(stats): add Go twins of panel-touched-paths and owned-corpus`
**Build:** green

**Decision:** kan842-helper-twins

  - [ ] **Step 1: Failing tests.** Read each library's header for its functions and contract.
    For each function, a table of inputs covering every branch of its body (panel-touched-paths:
    `panel_resolve_git` with and without an override, `panel_validate_worktree` on a missing
    argument, a non-directory, a non-repository and an unknown merge base, `panel_touched_paths`
    over committed, staged, unstaged and untracked paths including a rename and a path with a
    space; owned-corpus: every corpus member kind the header names, an absent root, a symlinked
    member). Each row runs the bash library once via `bash -c '. scripts/lib/<lib>.sh; <fn>
    <args>'` against a fixture in `t.TempDir()` and the Go function in-process, comparing stdout,
    stderr and status byte for byte. Run `cd stats && go test ./internal/guard/ -run
    '^(TestPanelTouchedPathsParity|TestOwnedCorpusParity)$' -count=1` — expect a compile failure.
  - [ ] **Step 2: Port** each library's functions into its Go file, its header citing the bash
    library as the source of truth it mirrors and the parity test that pins them.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
    ./internal/guard/ -run '^(TestPanelTouchedPathsParity|TestOwnedCorpusParity)$' -count=1
    -race`.

Correction (2026-09-27): `lib/owned-corpus.sh`'s header says a symlink loop is refused as "cannot
look through"; measured, `/usr/bin/find -L` skips a loop silently and exits 0. The Go twin and
`TestOwnedCorpusParity` follow the measured behaviour; the header is corrected in task 13.

- [ ] 3. Port check-workspace-isolation

**Files:** `stats/internal/guard/workspaceisolation.go`, `stats/internal/guard/check_workspace_isolation_test.go`, `scripts/check-workspace-isolation.sh`, `scripts/test-check-workspace-isolation.sh`
**Tests:** `TestCheckWorkspaceIsolation`
**Regression:** fails if any of the harness's 152 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_workspace_isolation_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** none
**Commit:** `feat(stats): port check-workspace-isolation to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-workspace-isolation.sh`,
    one subtest per `ok:` label; fixtures are small trees in `t.TempDir()`. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-workspace-isolation`; file resolution through
    `resolvefile.go`; the body's awk program (the `#SUMMARY` emitter, the resource and command
    table parsers) ported as Go code, each violation message byte for byte;
    `CHECK_WORKSPACE_ISOLATION_PRINT_ROWS` read as the bash does.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckWorkspaceIsolation$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckWorkspaceIsolation/'` — at least 152.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-workspace-isolation.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-workspace-isolation.sh` on this tree exits with the code and output it printed
    at `c5379c0a`, directly and through a symlink to it in a temp directory.

Correction (2026-09-27): Step 4's template `FLOW_GUARD_REPO_ROOT` resolves to `<repo>/skills/flow`
when the shim runs through `skills/flow/scripts/check-workspace-isolation.sh`; the bash resolved its
own symlink first (`resolve_file`, harness case 16), and only when no root argument was given (case
17). Shipped: the shim exports `FLOW_GUARD_SELF="${BASH_SOURCE[0]}"` in place of the two
`FLOW_GUARD_REPO_ROOT` lines (code 2), and the Go guard resolves it with `resolveFile` only when there
are no arguments. Case 15a (the awk validator printing no `#SUMMARY`) has no Go analogue; its three
labels now run the real shim in a tree with no `stats/` (exit 2, "cannot build flow-guard").
Divergences left outside the parity fixtures, named in `workspaceisolation.go`: invalid UTF-8 under a
UTF-8 locale is read where macOS awk aborted (exit 2); a heading whose space is U+00A0 is not the
section (`ccHeading`); awk `-v` escape processing of a backslash in the root is not reproduced.

- [x] 4. Port check-task-reviewer-single-dispatch

**Files:** `stats/internal/guard/taskreviewersingledispatch.go`, `stats/internal/guard/check_task_reviewer_single_dispatch_test.go`, `scripts/check-task-reviewer-single-dispatch.sh`, `scripts/test-check-task-reviewer-single-dispatch.sh`
**Tests:** `TestCheckTaskReviewerSingleDispatch`
**Regression:** fails if any of the harness's 11 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_task_reviewer_single_dispatch_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** none
**Commit:** `feat(stats): port check-task-reviewer-single-dispatch to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** dispatches-via-env-hook

  - [x] **Step 1: Failing test.** Port every case of
    `scripts/test-check-task-reviewer-single-dispatch.sh`, one subtest per `ok:` label; the stub
    `flow` the harness puts on PATH becomes `Env.Dispatches`. Run — expect failure.
  - [x] **Step 2: Port**, registering `check-task-reviewer-single-dispatch`; the store read as
    `panelfixsingledispatch.go`'s `pfdRead` does it; `plan-dispatch-bundles.sh` and
    `plan-dispatch-groups.sh` exec'd beside the shim with stdout and stderr combined, as the bash's
    `2>&1` captured them; `jq` filters become `encoding/json` decoding of the same fields.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run
    '^TestCheckTaskReviewerSingleDispatch$' -count=1 -race -v | grep -c -- '--- PASS:
    TestCheckTaskReviewerSingleDispatch/'` — at least 11.
  - [x] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT` (siblings resolve
    from it), code 2; `git rm scripts/test-check-task-reviewer-single-dispatch.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-task-reviewer-single-dispatch.sh` with no arguments exits 2 with the line it
    printed at `c5379c0a`.

Correction (2026-09-27): the script's header and body disagree on an unreadable decision row
(`[1]`): the header treats it as class `big`, the bash body died under `set -e` with an
undocumented exit 5. The port first followed the header (exit 0, `big`); the gated review found
`big` is the looser class for this guard (it tolerates one bundle per group, small/regular one per
run), so that reading passed what the bash blocked, and a top-level JSON object read as `big` too.
Fix commits `1b2cf11c`, `19c94536` and `d968f3da` make both a cannot-answer (exit 2, "decision
output was JSON but not an array of decision rows") — inside the 0/1/2 contract, blocking where the
bash blocked; a failed `flow record decisions` call and output that is not JSON at all (the bash's
`jq empty` gate) still read as `big`, as in the bash. The header's sentences are corrected to match. The `flow record decisions` read has no `Env` hook (`dispatches-via-env-hook`
names dispatches only) and execs `flow` on PATH through `pcFlow`; tests run with no `flow` on PATH,
except the cases that put a stub `flow` there.

- [ ] 5. Port recover-guard-incident

**Files:** `stats/internal/guard/recoverguardincident.go`, `stats/internal/guard/recover_guard_incident_test.go`, `scripts/recover-guard-incident.sh`, `scripts/test-recover-guard-incident.sh`
**Tests:** `TestRecoverGuardIncident`
**Regression:** fails if any of the harness's 67 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/recover_guard_incident_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** none
**Commit:** `feat(stats): port recover-guard-incident to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-recover-guard-incident.sh`,
    one subtest per `ok:` label, plus the Review Focus rows (every refusal leaves porcelain, HEAD
    and stash list unchanged; every abort precedes every restore). Run — expect failure.
  - [ ] **Step 2: Port**, registering `recover-guard-incident`; git runs as child processes with
    the arguments the bash passed; each restore writes `git show "stash@{0}^3:<f>"`'s bytes to
    the file, unstaged; preconditions checked in the header's order, the first failure naming its
    cause on stderr with stdout empty.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestRecoverGuardIncident$' -count=1
    -race -v | grep -c -- '--- PASS: TestRecoverGuardIncident/'` — at least 67.
  - [ ] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines
    dropped; `git rm scripts/test-recover-guard-incident.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/recover-guard-incident.sh --bogus` exits 2 with the line it printed at `c5379c0a`.

Correction (2026-09-27): the gated review found the repo-dir resolution looser than `cd "$1"`:
a lexical `filepath.Join` accepted an empty argument (bash 5: "null directory", exit 2; with
`--apply` the port aborted the cwd's revert), a missing component before `..`, and a directory
without search permission. Fix commit `d3810105` checks existence on the uncleaned path and search
permission, and refuses an empty argument, each with `not a directory: <arg>`, exit 2 — pinned by
three port subtests. Known and deferred to `KNOWN-BUGS.md`: `ls-tree` quotes non-ASCII planning
paths, so their `--apply` restore fails after the abort, in the bash and the port alike.

- [ ] 6. Port plan-class

**Files:** `stats/internal/guard/planclass.go`, `stats/internal/guard/plan_class_test.go`, `scripts/plan-class.sh`, `scripts/test-plan-class.sh`, `scripts/lib/sha256-hex.sh`
**Tests:** `TestPlanClass`
**Regression:** fails if any of the harness's 25 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/plan_class_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** Task 2
**Commit:** `feat(stats): port plan-class to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** kan842-helper-twins

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-plan-class.sh`, one subtest per
    `ok:` label, plus the Review Focus row pinning the three output lines against the bash at
    `c5379c0a` for a fixed change name. Run — expect failure.
  - [ ] **Step 2: Port**, registering `plan-class`; hashing through `sha256.go`, the
    `micro` touched-path read through `paneltouchedpaths.go` (task 2); `git rm
    scripts/lib/sha256-hex.sh`.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestPlanClass$' -count=1 -race -v |
    grep -c -- '--- PASS: TestPlanClass/'` — at least 25.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-plan-class.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/plan-class.sh spectre/changes/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to/tasks.md 1`
    prints the three lines the bash printed for the same plan at `c5379c0a`.

Correction (2026-09-27): three cases no harness covered diverge from the bash: an unreadable
tasks file exits 2 (`plan-class.sh: cannot read …`) where the bash printed a malformed answer; a
`<repos>` beyond int64 classifies the same without bash's `[` error line; `owned_corpus_files`'
wrong-argument-count refusal has no Go form (the function takes one root). `\b` in the
`**Build:** red` match follows the caller's locale, as grep's did. `sha256.go:10` and
`helpers_test.go:53` still cite the deleted `lib/sha256-hex.sh`; repointed in task 13.

- [ ] 7. Port check-installed-rules

**Files:** `stats/internal/guard/installedrules.go`, `stats/internal/guard/check_installed_rules_test.go`, `scripts/check-installed-rules.sh`, `scripts/test-check-installed-rules.sh`
**Tests:** `TestCheckInstalledRules`
**Regression:** fails if any of the harness's 20 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_installed_rules_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** none
**Commit:** `feat(stats): port check-installed-rules to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-installed-rules.sh`, one
    subtest per `ok:` label, each with `CHECK_INSTALLED_RULES_HOME` set to its own temp directory.
    Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-installed-rules`; `HOME` and
    `CHECK_INSTALLED_RULES_HOME` read through `Env.Getenv`/`Env.LookupEnv` as the bash reads them;
    the front-matter awk ported as Go code; the `.git`-is-a-file early exit kept.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckInstalledRules$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckInstalledRules/'` — at least 20.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code **1**;
    `git rm scripts/test-check-installed-rules.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `CHECK_INSTALLED_RULES_HOME=$(mktemp -d) scripts/check-installed-rules.sh` exits with the code
    and line it printed at `c5379c0a`.

Correction (2026-09-27): tasks 7 and 11 shims load `$SCRIPT_DIR/lib/flow-guard.sh` (task 8's
correction). check-installed-rules orders its `*.mdc`/`*.md` globs through the exec'd `crSort` —
bash's glob order measured equal to `sort`'s under en_US.UTF-8 and unlike byte order — so a missing
`sort` on PATH is a new refusal (exit 1); an unset `FLOW_GUARD_REPO_ROOT` is refused (exit 1) as in
the sibling ports. Untested edges differ: a directory as `CHECK_INSTALLED_RULES_SETUP_SH` gives a
different refusal line (both exit 1), and glob characters in a `managed_files` element are not
expanded. An unset `HOME` with no override first read as empty (a green verdict where the bash died
under `set -u`, exit 1); fix commit `f5d76c2f` refuses it, exit 1, pinned by a port subtest.

- [x] 8. Port resolve-base-branch

**Files:** `stats/internal/guard/resolvebasebranch.go`, `stats/internal/guard/resolve_base_branch_test.go`, `scripts/resolve-base-branch.sh`, `scripts/test-resolve-base-branch.sh`, `skills/flow-status/scripts/lib`
**Tests:** `TestResolveBaseBranch`
**Regression:** fails if any of the harness's 39 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/resolve_base_branch_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** none
**Commit:** `feat(stats): port resolve-base-branch to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-resolve-base-branch.sh`, one
    subtest per `ok:` label; fixture repos with and without an `origin` remote built once and
    copied per case. Run — expect failure.
  - [x] **Step 2: Port**, registering `resolve-base-branch`; git as child processes with the
    bash's arguments; exit 3 (no `origin`) kept distinct from exit 2, as the header's "WHY EXIT 3
    IS SEPARATE" paragraph requires.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestResolveBaseBranch$' -count=1
    -race -v | grep -c -- '--- PASS: TestResolveBaseBranch/'` — at least 39.
  - [x] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines
    dropped; `git rm scripts/test-resolve-base-branch.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/resolve-base-branch.sh <dir>` on a branch checkout whose origin/HEAD is main prints
    `main` and exits 0, as at `c5379c0a` (a detached worktree exits 1 on both).

Correction (2026-09-27): the template shim's `$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh`
failed from `skills/flow-status/scripts/`, which carried only the `resolve-base-branch.sh` link and
no `lib/` (exit 2, "cannot load lib/flow-guard.sh"); the bash worked there, and
`check-guard-symlinks.sh` rule 2 cannot see the `$(dirname …)` spelling. Shipped instead: the shim
loads `$SCRIPT_DIR/lib/flow-guard.sh` (the `check-panel-fix-single-dispatch.sh` spelling, visible to
rule 2), and `skills/flow-status/scripts/lib -> ../../../scripts/lib` is added — rule 2 exits 1
without it, 0 with it. Step 5 was measured in a detached wave worktree, where the bash and the shim
both exit 1; its expectation is restated for a branch checkout above.

- [ ] 9. Port check-visual-verify-dispatched

**Files:** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`, `scripts/test-check-visual-verify-dispatched.sh`
**Tests:** `TestCheckVisualVerifyDispatched`
**Regression:** fails if any of the harness's 18 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_visual_verify_dispatched_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** none
**Commit:** `feat(stats): port check-visual-verify-dispatched to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** dispatches-via-env-hook

  - [ ] **Step 1: Failing test.** Port every case of
    `scripts/test-check-visual-verify-dispatched.sh`, one subtest per `ok:` label; the stub `flow`
    becomes `Env.Dispatches`. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-visual-verify-dispatched`; `check-visual-trigger.sh`
    exec'd beside the shim and its three exit codes read as-is, as the header states; the store
    read as `pfdRead` does it.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckVisualVerifyDispatched$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckVisualVerifyDispatched/'` — at least 18.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT` (the trigger guard
    resolves from it), code 2; `git rm scripts/test-check-visual-verify-dispatched.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-visual-verify-dispatched.sh` with no arguments exits 2 with the line it printed
    at `c5379c0a`.

- [ ] 10. Port check-contract-budget

**Files:** `stats/internal/guard/contractbudget.go`, `stats/internal/guard/check_contract_budget_test.go`, `scripts/check-contract-budget.sh`, `scripts/test-check-contract-budget.sh`
**Tests:** `TestCheckContractBudget`
**Regression:** fails if any of the harness's 39 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_contract_budget_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** Task 2
**Commit:** `feat(stats): port check-contract-budget to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** kan842-helper-twins

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-contract-budget.sh`, one
    subtest per `ok:` label. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-contract-budget`; the corpus through
    `ownedcorpus.go` (task 2); `CHECK_CONTRACT_BUDGET_ROOT` read as the bash reads it; the
    budget-row and word-count arithmetic reproduced exactly, including `wc`'s whitespace rules.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckContractBudget$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckContractBudget/'` — at least 39.
  - [ ] **Step 4: Shim and delete** — shim template with `FLOW_GUARD_REPO_ROOT`, code 2;
    `git rm scripts/test-check-contract-budget.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-contract-budget.sh` on this tree exits 0 with the output it printed at
    `c5379c0a`, directly and through a symlink to it in a temp directory.

- [x] 11. Port check-task-commit-planning-paths

**Files:** `stats/internal/guard/taskcommitplanningpaths.go`, `stats/internal/guard/check_task_commit_planning_paths_test.go`, `scripts/check-task-commit-planning-paths.sh`, `scripts/test-check-task-commit-planning-paths.sh`
**Tests:** `TestCheckTaskCommitPlanningPaths`
**Regression:** fails if any of the harness's 19 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_task_commit_planning_paths_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** none
**Commit:** `feat(stats): port check-task-commit-planning-paths to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

  - [x] **Step 1: Failing test.** Port every case of
    `scripts/test-check-task-commit-planning-paths.sh`, one subtest per `ok:` label. Run — expect
    failure.
  - [x] **Step 2: Port**, registering `check-task-commit-planning-paths`; the spec root through
    `specroot.go`; git as child processes with the bash's arguments.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckTaskCommitPlanningPaths$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckTaskCommitPlanningPaths/'` — at least 19.
  - [x] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines
    dropped; `git rm scripts/test-check-task-commit-planning-paths.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-task-commit-planning-paths.sh` with no arguments exits 2 with the line it
    printed at `c5379c0a`.

- [ ] 12. Port check-panel-citation-trigger

**Files:** `stats/internal/guard/panelcitationtrigger.go`, `stats/internal/guard/check_panel_citation_trigger_test.go`, `scripts/check-panel-citation-trigger.sh`, `scripts/test-check-panel-citation-trigger.sh`
**Tests:** `TestCheckPanelCitationTrigger`
**Regression:** fails if any of the harness's 24 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_panel_citation_trigger_test.go 2>/dev/null | grep -cE '^func Test' @ c5379c0a -->
**After:** Task 2
**Commit:** `feat(stats): port check-panel-citation-trigger to Go`
**Build:** green

**Decision:** scope-ten-next-scripts

**Decision:** kan842-helper-twins

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-panel-citation-trigger.sh`,
    one subtest per `ok:` label. Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-panel-citation-trigger`; git resolution, worktree
    validation and the touched-path read through `paneltouchedpaths.go` (task 2); the first path
    ending `.md` or `.mdc` decides exit 0.
  - [ ] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckPanelCitationTrigger$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckPanelCitationTrigger/'` — at least 24.
  - [ ] **Step 4: Shim and delete** — shim template, code 2, `FLOW_GUARD_REPO_ROOT` lines
    dropped; `git rm scripts/test-check-panel-citation-trigger.sh`.
  - [ ] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-panel-citation-trigger.sh "$PWD" c5379c0a` exits the code
    `git show c5379c0a:scripts/check-panel-citation-trigger.sh` run with its libraries at
    `c5379c0a` exits for the same arguments on this branch.

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

  - [ ] **Step 1: Find.** `grep -rlF -e test-check-workspace-isolation.sh -e
    test-check-task-reviewer-single-dispatch.sh -e test-recover-guard-incident.sh -e
    test-plan-class.sh -e test-check-installed-rules.sh -e test-resolve-base-branch.sh -e
    test-check-visual-verify-dispatched.sh -e test-check-contract-budget.sh -e
    test-check-task-commit-planning-paths.sh -e test-check-panel-citation-trigger.sh -e
    lib/sha256-hex.sh --exclude-dir=archive --exclude-dir=.worktrees --exclude-dir=node_modules
    --exclude-dir=.git --exclude-dir=self-review .`; then grep the same tree for citations of a
    ported script's body comments and of its bash plumbing (`sources lib/…`).
    `unverified: the file set is known only after tasks 3–12 land; widen **Files:** by a
    correction if a hit falls outside the collateral globs`
  - [ ] **Step 2: Repoint** each citation to the Go file or Go test that now holds what it cites;
    a sentence describing bash plumbing that no longer exists is corrected, not repointed.
    `.flow/project.md`'s paragraph naming which guards are Go lists the ten new ones; library
    headers naming a ported script as a caller that sources them are corrected.
  - [ ] **Step 3: Verify.** Step 1's grep returns only `spectre/changes/kan-842-*`,
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

**Decision:** guard-package-under-40s

  - [ ] **Step 1: Suite, after.** Task 1 step 1's command, three times, on the branch head; same
    fields recorded.
  - [ ] **Step 2: Go package, after.** Task 1 step 2's command, three times.
  - [ ] **Step 3: Parity.** `cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '---
    PASS: Test<Name>/'` per port against its floor in `design.md`.
  - [ ] **Step 4: Record** an **After** table beside **Before**, same columns, each figure tagged
    `measured:` with the command and
    `@ branch spectre/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to`; name the slowest
    remaining harness and the next slice.
  - [ ] **Step 5: Judge.** Failure looks like: suite median not below the Before median; any
    port's `--- PASS` count below its floor; the Go package median above 40s real; any harness
    red. Any of these is reported, not recorded as success.
