## Principles Review — kan-362-myflow-guard-test-suite-takes-119s-sequentially

**Standards files:** Both resolved (`CLAUDE.md`, `AGENTS.md`). Neither declares project-specific standards ("This project has not declared one yet") — Hard Invariants section is empty by design; generic lint-fix-priority and no-new-suppression policy applies (diff introduces no new suppressions).

### Assessment: **Principles-compliant, with one Minor finding**

### Findings

**Minor — `parallel_replay_failures` is unused by both real callers (KISS / speculative generality).**
- `scripts/lib/parallel.sh:209` exports `parallel_replay_failures`, documented and unit-tested (`scripts/test-lib-parallel.sh:78`) as part of the library's public interface.
- Neither real caller uses it: `scripts/run-guard-tests.sh:130-140` reimplements its own list-order/failure-only replay loop over `PARALLEL_RC`/`PARALLEL_OUTFILES` (needed because it also prints a per-harness `---- name ----` header, which the bare function doesn't provide — a defensible reason not to reuse it there). `scripts/test-check-installed-citations.sh` never calls it either — its CHECK phase replays through per-case `check_*` functions instead, since each case needs contextual pass/fail judgment, not raw replay.
- **Reproducer:** `grep -rln parallel_replay_failures scripts/ .flow/` -> only `scripts/lib/parallel.sh` and its own `scripts/test-lib-parallel.sh`. Verified live.
- Not blocking — the function is cheap, arguably the "obvious" primitive a spawn/capture library should expose alongside `parallel_run`, and it's given a real test rather than shipped as an unverified extra. But it is exactly the "interface serving two callers" question the brief raises, and the honest answer is it serves zero callers today. Worth a one-line note in the header (why it's there / who it's for) rather than leaving a reader to discover this by grep.

### Judgment questions — verdicts

1. **Source-time EXIT/INT/TERM traps in `scripts/lib/parallel.sh`, unlike `coverage.sh`/`owned-corpus.sh` (verified: `grep -n trap` on both returns nothing).** Sound, not a violation. Reproduced the hazard directly: a caller that sources the file after installing its own `trap cleanup EXIT` has that trap silently clobbered (`MY OWN CLEANUP RAN` never printed in a live test). But the header documents the exact chaining recipe, and the one real caller that needs it (`test-check-installed-citations.sh`) follows it correctly and is itself proven under a real `kill -INT` (verified: both harnesses pass, including the "real SIGINT" sub-case). `run-guard-tests.sh` installs no competing trap, so it needs nothing. Two callers today, both handled correctly and testedly — documentation is a sufficient answer at this scale; flag only that a third, undocumented-reading caller would silently lose cleanup.

2. **Abstraction timing.** Right-sized. Every exported symbol (`parallel_job_count`, `parallel_run`, `PARALLEL_OUTFILES`, `PARALLEL_RC`, `PARALLEL_CLEANUP_DIRS`, `JOBS`) is consumed by both real callers — verified by grep above — except `parallel_replay_failures` (Minor finding above). No speculative generality otherwise: no config knobs, no unused parameters, no mode enum.

3. **Test design for the temp-dir-observing tests.** Sound, and strengthened, not weakened, by the fix. Original design (scan global `$TMPDIR`) misfired under concurrency — reproduced and fixed in commit `2e3ed4f`, giving each observing child a private `mktemp -d` `TMPDIR` (verified present in `scripts/test-lib-parallel.sh:286,348` and `scripts/test-check-installed-citations.sh`'s trap-chain sub-cases). Ran both harnesses standalone — all cases pass, including the two SIGINT sub-cases that exercise a real signal to a real child process.

4. **Runner output contract.** Matches `design.md`'s `replay-failures-only` decision exactly: one `ok`/`FAIL` line with wall time, a summary, replay only on failure, contiguous per-harness. Simple, no unrequested modes. `RUN_GUARD_TESTS_ROOT` idiom verified copied verbatim from `CHECK_REFERENCES_ROOT` (`scripts/check-references.sh:46-56`).

### Verified live (reproducers run, not just read)
- `scripts/test-lib-parallel.sh`, `scripts/test-check-installed-citations.sh`, `scripts/test-run-guard-tests.sh` — each run standalone, all pass.
- Trap-clobber hazard reproduced with a throwaway naive-caller script.
- `parallel_replay_failures` dead-caller claim confirmed by grep.
- 42 `scripts/test-*.sh` harnesses on disk, matching the design doc's byte-identical claim.
