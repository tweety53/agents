# kan-362-myflow-guard-test-suite-takes-119s-sequentially

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Four tasks: the shared primitive, the runner built on it, the citations harness moved onto it, and
the project configuration rewritten against a measurement only the first three make possible.
`design.md` is canonical for the decisions; `proposal.md` for every measurement behind them.

**Baseline, measured before any edit:**

- 40 harnesses matched by `scripts/test-*.sh`, byte-identical to the 40 lines `.flow/project.md`'s
  `## test` section lists.
  <!-- measured: diff <(awk '/^## test$/{f=1} f&&/^```/{c++; if(c==2) exit} f&&c==1&&/^scripts\/test-/' .flow/project.md | sort) <(ls scripts/test-*.sh | sort) @ c38f24e -->
- `scripts/test-check-installed-citations.sh`: 17.649s wall, 6.90s CPU, 14 guard invocations.
  <!-- measured: time scripts/test-check-installed-citations.sh; grep -c run_guard scripts/test-check-installed-citations.sh @ c38f24e -->
- `.flow/project.md` is 23779 bytes against `check-contract-budget.sh`'s 26450-byte ceiling.
  <!-- measured: wc -c .flow/project.md; grep -n 'project.md' scripts/check-contract-budget.sh @ c38f24e -->
- 0 files under `scripts/` reference a parallel-execution primitive; none exists.
  <!-- measured: grep -rln 'xargs -P' scripts/ @ c38f24e -->

- [x] 1. `scripts/lib/parallel.sh` — the spawn, capture and replay primitive

One owner for bounded-concurrency spawn, per-job output capture, deterministic replay and the
aggregate exit code, per `design.md`'s `shared-parallel-lib`.

**Bash 3.2 is the floor** — indexed arrays only, no associative arrays, and **no `wait -n`**. Build
on `xargs -P`, which is the mechanism `proposal.md`'s 49.6s was measured with.

Each job's stdout and stderr go to its own index-named file under a `mktemp -d`, removed on exit
including on interrupt. **Replay is in list order, never completion order.**

Job count per `design.md`'s `jobs-from-core-count`: `sysctl -n hw.ncpu`, then `nproc`, then 4;
`JOBS=` overrides. A `JOBS=` that is set but not a positive integer is refused with exit 2 rather
than coerced — this repository's guards reject rather than fold input into a reassuring default, as
`scripts/lib/coverage.sh`'s header records for its own count argument.

Write `scripts/test-lib-parallel.sh` **first, RED before GREEN**, covering every bullet under
`design.md`'s **Testing** for this file. Prove the list-order bullet with jobs whose completion
order is deliberately shuffled by differing `sleep` durations, not by hoping the scheduler varies.
Per KAN-197's mutation discipline, prove each assertion fails when its behaviour is broken.

**Files:** `scripts/lib/parallel.sh`, `scripts/test-lib-parallel.sh`
**Tests:** `scripts/test-lib-parallel.sh`
**Regression:** reverting this commit removes the only implementation of ordered failure replay, so
both callers added later lose failure attribution and a parallel run's diagnostics become the
interleaved output KAN-362 names as parallelism's one real cost.
**Baseline:** before=40 after=41 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 1 -->
**Commit:** `feat(scripts): add the parallel spawn, capture and replay primitive`
**Build:** green

- [x] 2. `scripts/run-guard-tests.sh` — the runner

Discovers `scripts/test-*.sh` and runs them through task 1's primitive, per `design.md`'s
`glob-discovers-harnesses` and `replay-failures-only`.

Output shape: one `ok`/`FAIL` line per harness with its wall time, then a summary naming the total,
the passed count, the failed count and the wall clock; then, only on failure, each failing harness's
captured streams replayed in full and contiguously. Non-zero exit naming every failure.

`RUN_GUARD_TESTS_ROOT` per `design.md`'s `runner-root-override` — opt-in, honoured only when set,
set-but-empty exits 2. Copy the idiom from `scripts/check-references.sh`'s own block rather than
writing a third variant of it.

**This override is what makes the task's own harness possible at all:**
`scripts/test-run-guard-tests.sh` is itself matched by the runner's glob, so a harness that did not
point the runner at a fixture root would re-enter the real suite.

Write `scripts/test-run-guard-tests.sh` **first, RED before GREEN**, covering every bullet under
`design.md`'s **Testing** for this file, with the same mutation discipline as task 1.

**Files:** `scripts/run-guard-tests.sh`, `scripts/test-run-guard-tests.sh`
**Tests:** `scripts/test-run-guard-tests.sh`
**Regression:** reverting this commit returns the suite to 40 sequential invocations and restores
the tool-timeout problem `.flow/project.md` currently works around in prose.
**Baseline:** before=41 after=42 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 2 -->
**Commit:** `feat(scripts): run the guard harnesses concurrently`
**Build:** green

- [x] 3. `scripts/test-check-installed-citations.sh` runs its cases on the primitive

Its 14 cases are independent by construction — each calls `new_fixture_repo`, which builds a fresh
`mktemp -d` fixture — so they become jobs on task 1's primitive, per `design.md`'s
`citations-parallel-not-injected`.

**Do not touch `scripts/check-installed-citations.py` or `scripts/check-installed-citations.sh`.**
The rejected alternative was a sandbox injection point in the guard; the whole point of this task is
that the harness changes and the guard does not.

The harness's existing `ok:`/`FAIL:` line per assertion, its final tail line and its non-zero exit
on any failure must all survive unchanged — a reader of its output should not be able to tell it
became concurrent except from the wall clock.

**Chain the EXIT trap, and correct the claim that says you need not.** `scripts/lib/parallel.sh`
installs `EXIT`/`INT`/`TERM` traps at source time, and this harness already installs
`trap cleanup EXIT` guarding the `SANDBOXES` fixture repos. Sourcing one into the other silently
drops whichever trap was installed first — either 14 fixture-repo trees leak, or the library's own
`mktemp -d` leaks. Task 1's per-task review reproduced both directions. Chain them per the
library header's own instructions, and fix that header's next sentence, "Neither current caller
needs to", which is false as of this task.

**Files:** `scripts/test-check-installed-citations.sh`, `scripts/lib/parallel.sh`
**Tests:** `scripts/test-check-installed-citations.sh`
**Regression:** reverting this commit returns that harness to 17.6s, which — with
`test-run-reproducer.sh` at 17.7s — makes it a joint setter of the parallel critical path rather
than a harness that finishes inside it, and restores the unchained EXIT trap that leaks one of the
two cleanups.
**Baseline:** before=42 after=42 harnesses matched by `scripts/test-*.sh`
<!-- predicted: this task adds no harness, it converts one -->
**Commit:** `perf(scripts): run the installed-citations cases concurrently`
**Build:** green

- [x] 4. `.flow/project.md` — one command, and a note stating the measured truth

`## test`'s 40 harness lines collapse to `scripts/run-guard-tests.sh`. The two `stats` commands stay
separate entries, per `design.md`'s `test-section-stays-three-commands` — that decision records
exactly what folding them in would break, and this task must not do it.

**Measure the full suite before writing the note**, sequentially and through the runner, and tag
both figures `measured:` with the branch as the ref. The note replaces today's "split the run
across more than one invocation … or raise the tool timeout" workaround: that prose exists only
because the suite did not fit, and it must not survive a change that makes it fit.

The three per-guard prose notes below `## test` and `## lint` that name
`check-installed-citations.sh`'s subprocess cost are about the **guard**, not the harness, and task
3 did not change the guard — leave them alone.

Then run this repository's own `## lint` list. `check-contract-budget.sh`'s 26450-byte ceiling is a
ceiling and the file shrinks, so no budget row is edited; if any guard does report a hit, fix the
offending line rather than the guard.

**Files:** `.flow/project.md`
**Tests:** none added — this task edits configuration prose, and the runner it points at is already
covered by `scripts/test-run-guard-tests.sh` from task 2
**Regression:** reverting this commit leaves the runner built and unused, with every `/flow` verify
stage still running the 40 sequential lines and still hitting the tool timeout.
**Baseline:** before=42 after=42 harnesses matched by `scripts/test-*.sh`
<!-- predicted: this task adds no harness -->
**Commit:** `refactor(project-config): point ## test at the concurrent runner`
**Build:** green
