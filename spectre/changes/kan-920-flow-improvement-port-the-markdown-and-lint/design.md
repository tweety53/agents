# kan-920-flow-improvement-port-the-markdown-and-lint

## Context

- Seventh slice of KAN-760's guard port (KAN-778, KAN-841, KAN-842, KAN-850, KAN-873, KAN-918
  before it). Base `75261910`. The pattern and every per-port convention are KAN-918's, carried
  unchanged (`tasks.md` preamble).
- All three scripts derive the repository root from their own location (`$SCRIPT_DIR/..`), so each
  shim exports `FLOW_GUARD_REPO_ROOT` and calls `flow_guard_cwd_checkout`, as
  `scripts/check-dispatch-paragraphs.sh` does. Two honour an opt-in override —
  `CHECK_NORMATIVE_INVENTORY_ROOT`, `CHECK_MUTATION_REPRODUCER_PIN_ROOT` — refused when set but
  empty; the Go port reads it from `Env` exactly as `dispatchparagraphs.go` reads
  `CHECK_DISPATCH_PARAGRAPHS_ROOT`. `check-python-suppressions` takes path arguments instead.
  <!-- verified: grep -n 'SCRIPT_DIR\|_ROOT' scripts/check-{normative-inventory,mutation-reproducer-pin,python-suppressions}.sh; sed -n 130,145p stats/internal/guard/dispatchparagraphs.go @ 75261910 -->
- Cannot-answer code is `2` for all three.
  <!-- verified: each script's header "Exit codes" paragraph @ 75261910 -->
- `scripts/lib/owned-corpus.sh` is sourced only by `check-normative-inventory.sh`;
  `scripts/lib/parallel.sh` names it in a comment and does not source it, so the issue's premise
  that a second bash consumer exists does not hold. `asksilence.go`'s `asCorpus` is already its Go
  twin, pinned by `TestAskSilenceCorpusParity` through `bashLib`.
  <!-- verified: grep -rn 'owned-corpus' --exclude-dir=archive . ; sed -n 60,70p scripts/lib/parallel.sh @ 75261910 -->
- `scripts/lib/owned-corpus.sh` at `libTwinBase` (`ef6f87cd`) is byte-identical to the base, so a
  parity test can run it from history once it is deleted.
  <!-- verified: git diff --stat ef6f87cd 75261910 -- scripts/lib/owned-corpus.sh prints nothing @ 75261910 -->

## Measurements

### Parity floors at base

| Harness | floor |
|---|--:|
| `test-check-normative-inventory.sh` | 20 |
| `test-check-mutation-reproducer-pin.sh` | 20 |
| `test-check-python-suppressions.sh` | 19 |

- Floor is the harness's `^ok` line count on a green run; all three exited 0.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) bash scripts/test-check-<name>.sh, grep -c '^ok' of each log, all exit 0 @ 75261910 -->

Each Go test carries at least its harness's floor in subtests (`parity-by-case-count`).

**After** — subtests per port: recorded by task 5.

## Decisions

### Carry the prior slices' port decisions unchanged

**ID:** carry-prior-port-decisions
**Status:** active
**Chosen:** KAN-760's `flow-guard-binary-rationale-in-go`, `go-tests-replace-harnesses`,
`parity-by-case-count`, `inject-deadlines-in-process`, `helper-parity-tests`,
`guard-binary-built-from-checkout`, KAN-841's `shared-helper-go-twins`, KAN-873's shim template and
five-step port, and KAN-918's `header-wins-on-exit-code` govern this slice as written; the base is
`75261910`.
**Considered:** re-deciding each per script — no script here raises a case those decisions do not
cover.

### Scope: the three guards the issue names, plus their corpus library

**ID:** scope-three-guards-and-owned-corpus
**Status:** active
**Chosen:** port exactly the three KAN-920 lists, one task each, and delete
`scripts/lib/owned-corpus.sh` in the `check-normative-inventory` task, since that task removes its
last bash consumer.
**Considered:** leaving `owned-corpus.sh` for the test-runner ticket as the issue suggests — its
premise (that `parallel.sh` sources it) is false (`## Context`), so the file would sit with no
consumer and a Go twin pinned against it.

### One Go corpus walk, shared by check-ask-silence and check-normative-inventory

**ID:** shared-owned-corpus-walk
**Status:** active
**Chosen:** move `asCorpus` and `askSilenceScopes` out of `asksilence.go` into
`stats/internal/guard/ownedcorpus.go` as `ownedCorpus(root, prefix, stderr)` and
`ownedCorpusScopes`; `check-ask-silence` passes its own `check-ask-silence` prefix (its messages
unchanged), `check-normative-inventory` passes `owned_corpus_files` (the bash library's own prefix,
so its stderr stays byte-identical). `TestAskSilenceCorpusParity` keeps pinning the walk against the
bash library, now through `bashLibAt` at `libTwinBase`.
**Considered:** a second walk in `normativeinventory.go` — two Go definitions of one corpus is the
drift `asksilence.go`'s own comment calls a real defect; dropping the parity test with the bash
library — loses the only check that is not laid out from the Go scope list itself.

### Python-suppressions skips binary files the way grep -I does

**ID:** python-suppressions-binary-skip
**Status:** active
**Chosen:** a file whose content holds a NUL byte is treated as binary and yields no hit, which is
how GNU and BSD grep's `-I` classify a file; every other file is matched line by line with a
case-insensitive Go regexp of the same alternation.
**Considered:** exec'ing `grep` from the port — keeps grep's exact heuristic but leaves the guard a
bash pipeline in Go clothing and adds a PATH dependency the bash had implicitly; no fixture in the
harness depends on a heuristic finer than NUL detection.

### No suite-timing table

**ID:** no-suite-timings
**Status:** active
**Chosen:** record parity floors and the after subtest counts only; no before/after
`run-guard-tests.sh` wall-clock table.
**Considered:** KAN-918's three-run before/after table — KAN-920 does not ask for it, and KAN-918's
own runs showed the suite wall moving with machine load (313–411s on an unchanged tree), so a table
for three small harnesses would measure noise.

## Live check

`flow.verify` runs each of the three shims once on this repository through
`~/.claude/skills/flow/scripts/<name>.sh` where installed, else `scripts/<name>.sh`, with no
arguments, and compares exit code, stdout and stderr with the same call against the base's bash
(`git show 75261910:scripts/<name>.sh` run from a temp copy of the base tree); any difference is a
failure. For `check-normative-inventory` the stdout comparison is the whole inventory, byte for
byte, and the stderr count line must name the same sentence and file counts as the base's. A shim that prints nothing and exits 0 where the bash printed an inventory or a verdict
line is the failure shape to watch for.

## Open questions
