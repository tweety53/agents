# kan-922-flow-improvement-port-the-bash-python-tooling

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Ports `scripts/workspace.sh`, `scripts/load-sets.sh` and `scripts/installer-sandbox-diff.sh` into
`flow-guard` (one task each). `design.md` is canonical for every decision; each task cites its
entry by ID.

**The script at `e111a883` is each port's specification.** Its header states the contract, its body
the behaviour. A port reproduces arguments, environment handling, every output line byte for byte,
the stdout/stderr split and every exit code. Header comments move into the Go file beside the code
they explain, corrected where the mechanism changed. Each port registers its name in
`guard.Registry` from its own file's `init()`, as `plandispatchgroups.go` does, and reads its
process environment and tool paths from the `Env` the registry hands it, so tests inject both.

**Shim template** — the script's header comment block kept (corrected where it describes bash
plumbing), then, `<name>` the basename without `.sh`:

```bash verified:copied from the tail of scripts/plan-dispatch-groups.sh @ e111a883
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "<name>: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec <name> 2 "<name>:" "$@"
```

**Test isolation:** every Go test calls `t.Parallel()`; fixtures live in `t.TempDir()`; no test reads
or writes the operator's real `$HOME`.

**Per-task verify** is `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
./internal/guard/ -run '^<Test>$' -count=1 -race -v`, then the shim run once for real as step 5
names. The full package and `scripts/run-guard-tests.sh` run in `flow.verify`.

## Review Focus

- `workspace survivors`' exit/stdout contract (`.flow/project.md` `## workspace isolation`,
  `skills/flow-contracts/project-configuration.md`): exit 0 + empty stdout is the only verifying
  result; a failing `docker exec` must be non-zero, never an empty exit 0.
- `workspace remove --force` refuses any database not ending `_uitest` before any `docker` call.
- `installer-sandbox-diff` mode strings and symlink normalization must match the bash manifest
  byte for byte, including a tree path holding glob metacharacters.

- [x] 1. Port workspace

**Files:** `stats/internal/guard/workspace.go`, `stats/internal/guard/workspace_test.go`, `scripts/workspace.sh`, `scripts/test-workspace.sh`
**Tests:** `TestWorkspace`
**Regression:** fails if any of the harness's 21 `ok:` behaviours regress (container cases skip when `flow-postgres` is unreachable).
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/workspace_test.go 2>/dev/null | grep -cE '^func Test' @ e111a883 -->
**After:** none
**Commit:** `refactor(guard): port workspace.sh to Go`
**Build:** green

**Decision:** carry-prior-port-decisions
**Decision:** exec-external-tools
**Decision:** workspace-validation-without-docker

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-workspace.sh` to
    `TestWorkspace`, one subtest per `ok:` label. Refusal cases (empty/missing id, non-`[a-z0-9-]`
    id, unknown flag, `--force` on a non-`_uitest` id, `--force` on `create`) always run; the
    round-trip cases skip with `SKIP: flow-postgres compose stack not reachable` when
    `docker exec flow-postgres true` fails, and use a `wstest-<pid>` id cleaned up in `t.Cleanup`.
    Run — expect failure (`workspace` unregistered).
  - [x] **Step 2: Port** `workspace` into `stats/internal/guard/workspace.go`: same usage text,
    `ERROR:` lines, exit codes and `docker exec` argument vectors as the script; `survivors` keeps
    `PGCONNECT_TIMEOUT=5` and `PGOPTIONS=-c statement_timeout=5000`, prints each non-blank line of
    psql's output, and exits non-zero with psql's code when `docker exec` fails.
  - [x] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestWorkspace$' -count=1
    -race -v | grep -c -- '--- PASS: TestWorkspace/'` — 21 with `flow-postgres` up.
  - [x] **Step 4: Shim and delete** — shim template, `<name>` = `workspace`;
    `git rm scripts/test-workspace.sh`.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/workspace.sh survivors wstest-none; echo $?` prints nothing and `0`;
    `scripts/workspace.sh create 'Bad'; echo $?` prints the `[a-z0-9-]` error and `2`.

- [x] 2. Port load-sets

**Files:** `stats/internal/guard/loadsets.go`, `stats/internal/guard/loadsets_test.go`, `scripts/load-sets.sh`
**Tests:** `TestLoadSets`
**Regression:** fails if a row's format, a section header, the MISSING-to-stderr report or the always-0 exit changes.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/loadsets_test.go 2>/dev/null | grep -cE '^func Test' @ e111a883 -->
**After:** none
**Commit:** `refactor(guard): port load-sets.sh to Go`
**Build:** green

**Decision:** carry-prior-port-decisions
**Decision:** load-sets-self-locating
**Decision:** load-sets-cannot-answer-2

  - [x] **Step 1: Failing test.** `TestLoadSets` subtests: (a) on a fixture repo in `t.TempDir()`
    holding a few of the listed files of known sizes and a fixture global file, stdout equals the
    output `bash scripts/load-sets.sh <fixture> <global>` produces at `e111a883` (capture it once
    into the test as the expected literal); (b) each absent file is one
    `load-sets: MISSING <repo-relative path>` stderr line and counts 0; (c) exit 0 with every file
    missing. Run — expect failure.
  - [x] **Step 2: Port** `load-sets` into `stats/internal/guard/loadsets.go`: the same arrays in the
    same order, `printf '%-50s %7d bytes  ~%6d tok\n'` as `fmt` `%-50s %7d bytes  ~%6d tok`,
    integer `bytes/4`; `<repo>` defaults to `FLOW_GUARD_REPO_ROOT`, `<global>` to
    `$HOME/.claude/CLAUDE.md`; always exit 0. A MISSING line is printed each time a missing file is
    summed, as the script does.
  - [x] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestLoadSets$' -count=1
    -race -v`.
  - [x] **Step 4: Shim** — shim template, `<name>` = `load-sets`, plus
    `FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"` and `export FLOW_GUARD_REPO_ROOT` before the exec,
    as `scripts/check-plan-shape.sh` does.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `diff <(git show e111a883:scripts/load-sets.sh | bash -s -- "$PWD") <(scripts/load-sets.sh "$PWD")`
    prints nothing, stderr compared the same way.

- [x] 3. Port installer-sandbox-diff

**Files:** `stats/internal/guard/installersandboxdiff.go`, `stats/internal/guard/installersandboxdiff_test.go`, `scripts/installer-sandbox-diff.sh`, `scripts/test-installer-sandbox-diff.sh`
**Tests:** `TestInstallerSandboxDiff`
**Regression:** fails if any of the harness's 15 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/installersandboxdiff_test.go 2>/dev/null | grep -cE '^func Test' @ e111a883 -->
**After:** none
**Commit:** `refactor(guard): port installer-sandbox-diff.sh to Go`
**Build:** green

**Decision:** carry-prior-port-decisions
**Decision:** exec-external-tools

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-installer-sandbox-diff.sh`
    (cases 1–9) to `TestInstallerSandboxDiff`, one subtest per `ok:` label, with the harness's own
    fixture `setup.sh` trees built under `t.TempDir()` and `TMPDIR` injected through `Env`. Run —
    expect failure.
  - [x] **Step 2: Port** `installer-sandbox-diff` into
    `stats/internal/guard/installersandboxdiff.go`: tree resolution and `ERROR:` texts, `mktemp`
    work dir under `$TMPDIR` (or `/tmp`) named `installer-sandbox-diff.XXXXXX`, `bash ./setup.sh
    global` per tree with `HOME` set and `MANAGED_BLOCK_POSTPROCESS` removed, the 15-line log tail
    on failure, the manifest (`l`/`d`/`f` lines, `LC_ALL=C` byte-sorted relative paths, literal
    tree→`<TREE>` replacement, mode from `syscall.Stat_t.Mode` in octal on Darwin and permission
    bits in octal elsewhere, `cksum` output exec'd), `diff -u` exec'd, the identical/differs lines,
    and the kept-sandbox stderr line on every non-zero exit.
  - [x] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestInstallerSandboxDiff$'
    -count=1 -race -v | grep -c -- '--- PASS: TestInstallerSandboxDiff/'` — at least 15.
  - [x] **Step 4: Shim and delete** — shim template, `<name>` = `installer-sandbox-diff`;
    `git rm scripts/test-installer-sandbox-diff.sh`.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/installer-sandbox-diff.sh "$PWD" "$PWD"` exits 0 and prints the identical line;
    the manifests it builds match the ones `git show e111a883:scripts/installer-sandbox-diff.sh`
    builds for the same tree (compare a kept sandbox's `old.manifest` from each, forcing a keep
    with a one-byte edit in a copy of the tree).

Correction (2026-10-22): the plan declared one subtest per harness `ok:` label and a count of at
least 15; `TestInstallerSandboxDiff` carries all 15 labels in 9 per-case subtests (a case's exit and
output labels share one), plus 3 the harness lacked — a non-directory argument, a failing
installer's log tail and kept sandbox, and an exact manifest pin — 12 in all. The port also sets
`PWD` to the tree for the installer run, as the script's `cd` did; without it an installer reading
`pwd` sees the symlink-resolved path and link targets miss the `<TREE>` normalization.
