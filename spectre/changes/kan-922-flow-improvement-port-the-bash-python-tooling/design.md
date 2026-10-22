## Context

- Eighth slice of KAN-760's port. Base `e111a883`. Pattern and per-port conventions are KAN-923's
  (`spectre/changes/archive/kan-923-flow-improvement-port-the-python-backed-guards/tasks.md`
  preamble).
- `plan-dispatch-groups` and `generate-relocation-comparison`, two of the ticket's five rows, are
  already Go: `d8045f82 refactor(guard): port the Python-backed guards to Go` replaced both wrappers
  with shims and deleted their `.py` files and harnesses. Recorded as fixed; no task.
  <!-- verified: git show --stat d8045f82; cat scripts/plan-dispatch-groups.sh scripts/generate-relocation-comparison.sh @ e111a883 -->
- `workspace.sh`'s `create`/`remove`/`survivors` triple is the contract `.flow/project.md`'s
  `## workspace isolation` names; the shim keeps the path, so that table does not change.
- No file outside the two harnesses themselves cites `test-workspace.sh` or
  `test-installer-sandbox-diff.sh`.
  <!-- verified: grep -rn 'test-workspace\.sh\|test-installer-sandbox-diff\.sh' --exclude-dir=spectre --exclude-dir=docs . @ e111a883 -->

## Measurements

| Harness | floor |
|---|--:|
| `scripts/test-workspace.sh` | 21 |
| `scripts/test-installer-sandbox-diff.sh` | 15 |

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) bash scripts/test-<name>.sh, grep -c '^ *ok' of each log, both exit 0, flow-postgres reachable @ e111a883 -->

## Decisions

### Already-ported helpers get no task

**ID:** two-already-ported
**Status:** active
**Chosen:** record `plan-dispatch-groups` and `generate-relocation-comparison` as fixed by `d8045f82` and plan nothing for them.
**Considered:** re-porting or re-auditing them — the shims, Go files and deleted harnesses are already on base; there is nothing left to port.

### Carry KAN-923's port conventions

**ID:** carry-prior-port-decisions
**Status:** active
**Chosen:** parity by case count (one Go subtest per harness `ok:` line), registration from each file's `init()`, the KAN-923 shim template, `t.Parallel()` tests that never touch the real `$HOME`.
**Considered:** new conventions for non-guard helpers — no reason the helpers differ from the guards in how they are built or tested.

### External tools stay exec'd

**ID:** exec-external-tools
**Status:** active
**Chosen:** `workspace` runs `docker exec` exactly as the script does; `installer-sandbox-diff` runs `bash ./setup.sh global`, `cksum` and `diff -u` through `os/exec`, and prints modes as the script's `stat` did (Darwin: full `st_mode` octal; elsewhere: permission bits octal).
**Considered:** a Go Postgres client, a Go CRC/unified-diff — more code and a parity risk for no gain; the tools are what the byte-for-byte output is defined by.

### load-sets keeps its own location as the default repo

**ID:** load-sets-self-locating
**Status:** active
**Chosen:** the `load-sets.sh` shim exports `FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"` and the Go side uses it when no `<repo>` argument is given, as `check-plan-shape.sh` does.
**Considered:** the caller's working directory — changes the script's documented default.

### load-sets' cannot-answer exit is 2

**ID:** load-sets-cannot-answer-2
**Status:** active
**Chosen:** the shim passes `2` to `flow_guard_exec`, so a failed build exits 2; the Go report itself still always exits 0.
**Considered:** `0` — would turn a build failure into an empty report that reads as a measurement.

### workspace validation cases run without the container

**ID:** workspace-validation-without-docker
**Status:** active
**Chosen:** the id/usage/`--force` refusal cases run always (they exit before any `docker` call); the round-trip cases skip with the harness's `SKIP` reason when `flow-postgres` is unreachable.
**Considered:** skipping the whole test as the harness did — would leave the refusals unchecked on every machine without the stack.

## Open questions

## Live check

None beyond the round-trip cases: `TestWorkspace`'s container cases run `create`/`remove`/`survivors` against the real `flow-postgres` with a throwaway `wstest-` id, which is the live system; a refusal of `survivors` to print a created database, or a non-empty print after `remove`, is the failure shape.
