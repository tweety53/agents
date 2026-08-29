# KAN-362 — parallelise the guard test suite

## Context

**Why this change exists, what it changes, and every measurement behind it:** `proposal.md`, beside
this file — canonical for all three, and not restated here.

**Bash 3.2 is this repository's floor.** `scripts/test-check-installed-citations.sh`'s own header
records it, citing `scripts/test-check-finish-preflight.sh`: indexed arrays only, no associative
arrays, and therefore no `wait -n` (Bash 4.3+). This is why the primitive below is built on
`xargs -P` — which is also the mechanism the 49.6s in `proposal.md` was actually measured with,
rather than a second mechanism nobody has timed.

## Components

| File | Role |
|---|---|
| `scripts/lib/parallel.sh` *(new)* | Sole owner of bounded-concurrency spawn, per-job output capture, deterministic replay, aggregate exit code |
| `scripts/run-guard-tests.sh` *(new)* | Discovers `scripts/test-*.sh`, runs them through the primitive, reports |
| `scripts/test-lib-parallel.sh` *(new)* | Harness for the primitive |
| `scripts/test-run-guard-tests.sh` *(new)* | Harness for the runner |
| `scripts/test-check-installed-citations.sh` | Its 14 cases become jobs on the same primitive |
| `.flow/project.md` | `## test` drops from 40 lines to 1; the runtime note is rewritten |

<!-- measured: awk '/^## test$/{f=1} f&&/^```/{c++; if(c==2) exit} f&&c==1&&/^scripts\/test-/' .flow/project.md | wc -l @ c38f24e -->

Each job's stdout and stderr are captured to its own index-named file under a `mktemp -d`.
**Replay is in list order, never completion order**, so a failure reads exactly as it does in
today's sequential transcript.

**Nested use is allowed to oversubscribe.** The outer runner runs
`test-check-installed-citations.sh` as one of its jobs, and that harness forks 14 of its own. No
budget is shared across the two levels: the nesting only bites at the tail of the outer run, which
is precisely where the pool has spare capacity, and a cross-level budget would be machinery serving
a case that does not hurt.

## Decisions

### Discover harnesses by glob rather than by list

**ID:** glob-discovers-harnesses
**Status:** active
**Chosen:** `run-guard-tests.sh` globs `scripts/test-*.sh`. Verified byte-identical to the 40 lines
<!-- measured: diff <(awk '/^## test$/{f=1} f&&/^```/{c++; if(c==2) exit} f&&c==1&&/^scripts\/test-/' .flow/project.md | sort) <(ls scripts/test-*.sh | sort) @ c38f24e -->
`## test` lists today. This also closes a standing hazard: a new `test-*.sh` nobody adds to
`.flow/project.md` is silently never run.
**Considered:** *an explicit array inside the runner* — rejected, it moves the same manual-edit
requirement to a different file and keeps the drift hazard. *The runner parsing `## test`'s own
fenced block* — rejected, it is circular (the file names the runner that reads it) and would
duplicate `check-task-commit-fields.py`'s parser.

### Quiet on pass, full captured output on failure

**ID:** replay-failures-only
**Status:** active
**Chosen:** one `ok`/`FAIL` line per harness with its wall time, then a summary; on failure, the
failing harnesses' captured streams replayed in full and contiguously.
**Considered:** *replaying all 40 harnesses' output in list order regardless of outcome* —
rejected, roughly a thousand lines of noise on every green run, in a transcript an agent reads.
*Streaming live with a per-harness line prefix* — rejected, lines from N harnesses still
interleave and the multi-line assertion output several harnesses emit gets shredded across the
prefixes.

### Job count derived from the core count

**ID:** jobs-from-core-count
**Status:** active
**Chosen:** `sysctl -n hw.ncpu` on macOS, `nproc` on Linux, fallback 4; `JOBS=` overrides.
**Considered:** *a fixed `-P 8`* — rejected, it bakes a guess about the machine into the script,
over-subscribing a 2-core CI runner and under-using a 16-core one. *`-P 0` (unbounded)* — rejected,
40 concurrent `git init` and `mktemp -d` trees thrash a laptop, and nothing has measured it.

### `test-check-installed-citations.sh` is parallelised, not given a sandbox injection point

**ID:** citations-parallel-not-injected
**Status:** active
**Chosen:** its 14 independent cases become jobs on the same primitive. `check-installed-citations.py`
is not touched.
**Considered:** *KAN-362's own proposal — a new env var letting a caller hand the guard a sandbox
already populated by `setup.sh`, skipping its two runs* — rejected on two grounds: it adds a
test-only bypass to `ensure_within_sandbox`, the containment invariant that function exists
specifically to keep checked rather than assumed; and at 39% CPU utilisation the wall clock is
idle-dominated, so it is not established that `setup.sh` is even the dominant term. *Profiling each
case before deciding* — rejected, the 39% figure already answers the design question.

### One shared primitive rather than two copies

**ID:** shared-parallel-lib
**Status:** active
**Chosen:** `scripts/lib/parallel.sh`, owning spawn, capture, replay and exit code once.
**Considered:** *duplicating the `xargs` and capture logic in each caller* — rejected. Two callers
is exactly the threshold, and `scripts/lib/coverage.sh`'s own header records the five-copy
`resolve_file` drift this repository already suffered when a shared concern was not owned once.

### `## test` keeps three commands

**ID:** test-section-stays-three-commands
**Status:** active
**Chosen:** `scripts/run-guard-tests.sh`, `cd stats && go test ./... -race -count=1`, and
`cd stats/web && npm test` remain three separate entries.
**Considered:** *folding all three into the runner so `## test` names one command* — rejected.
`check-task-commit-fields.py`'s `read_single_test_command` returns a command only when the section
names exactly one; handing it a single command would make that guard invoke the entire suite to
probe for `RESULT <name>: pass` and `COUNT: <N>` before falling back to skipped-not-verified
anyway. Three entries preserve today's behaviour exactly, and teaching the runner that protocol is
a separate change nobody has asked for.

### The runner is scoped by `RUN_GUARD_TESTS_ROOT`

**ID:** runner-root-override
**Status:** active
**Chosen:** opt-in, honoured only when set, set-but-empty exits 2 — matching `CHECK_REFERENCES_ROOT`
and `CHECK_GUARD_SYMLINKS_ROOT` verbatim. This is load-bearing, not decoration:
`test-run-guard-tests.sh` is itself matched by the runner's own glob, so without a fixture override
its harness would re-enter the real suite.
**Considered:** *a positional `<scripts-dir>` argument* — rejected, the two guards this repository
already scopes this way both use the env-var idiom, and an argument would make the runner's normal
invocation something a call site could narrow.

### `test-run-reproducer.sh` is left alone

**ID:** reproducer-waits-untouched
**Status:** active
**Chosen:** its 17.7s of `sleep` stands, and sets the parallel floor.
**Considered:** *lowering its timeout override* — rejected for this change. It exercises
`run-reproducer.sh`'s real timeout and kill paths, so shrinking the waits risks weakening what the
harness proves, and KAN-362 itself scopes it out as "worth checking, not assumed here".

## Open questions

None.

## Testing

`test-lib-parallel.sh`:

- a clean run exits 0 and replays nothing;
- one failing job exits non-zero and replays exactly that job's own captured stdout and stderr;
- replay follows list order under deliberately shuffled completion order;
- jobs whose output interleaves in real time are still replayed contiguously;
- `JOBS=1` and `JOBS=<n>` produce identical output.

`test-run-guard-tests.sh`, against a fixture root via `RUN_GUARD_TESTS_ROOT`:

- discovery finds every `test-*.sh` in the fixture and nothing else;
- one `ok`/`FAIL` line per harness, plus the summary counts;
- non-zero exit on any failure;
- `RUN_GUARD_TESTS_ROOT` set but empty exits 2.

Every assertion is proven to fail when its behaviour is broken, per KAN-197's mutation discipline.

## Verification

- The new suite reproduces today's 40/40 pass, and its wall clock is measured against the ~20s
  target `proposal.md` states.
- `check-contract-budget.sh` — `.flow/project.md` shrinks, so its 26450-byte ceiling is unaffected.
- `check-guard-symlinks.sh` — neither new script is a shipped guard, so no `skills/*/scripts/`
  symlink is required.
- `check-references.sh` and `check-vocabulary.sh` over the new prose.
