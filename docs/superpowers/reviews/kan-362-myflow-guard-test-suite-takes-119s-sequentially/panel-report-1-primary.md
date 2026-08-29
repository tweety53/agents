Delta re-review of the fix round (`git diff 907ee16..4c9128d`).

## PASS

All four findings genuinely closed; no new defects; tree clean.

**F2 (Major, `.flow/project.md`)** — HOLDS. Note now states reproduced 3-of-8, root-caused as shared `$TMPDIR` namespace collision (not oversubscription), fixed in `2e3ed4f`, 24 clean runs since — matches a `measured:` provenance-tagged comment in the file's established format (`.flow/project.md:110`). Root cause is real: `2e3ed4f` gives each temp-dir-observing sub-case (test-lib-parallel.sh cases 7-8, test-check-installed-citations.sh trap-chain sub-cases) its own private `TMPDIR` instead of scanning the shared namespace.

**F1 (Minor, `scripts/lib/parallel.sh`)** — HOLDS.
- `run-guard-tests.sh` now calls `parallel_replay_failures "${NAMES[@]}"` instead of duplicating the walk; diffed the removed loop against the new call — byte-identical output shape (`---- name ----` header, cat, blank line, list order).
- Verified via `scripts/test-run-guard-tests.sh` (fixture-based, isolated) — all replay/order/header assertions pass unmodified.
- The `return 0` bug claim is real, confirmed by reverting it: `/bin/bash /tmp/repro-f1b.sh` (caller under `set -euo pipefail`, `result="$(parallel_replay_failures)"` after a failing job, no labels) aborted before `"got here"` printed (script exit 1) — the `if [ "$#" -gt 0 ]` test evaluating false was the loop's/function's last exit status, killing the command-substitution caller. With `return 0` restored, the same repro reaches `"got here"` cleanly (exit 0).

**F3 (Minor, `parallel_run` under `set -euo pipefail`)** — HOLDS, verified hardest per instructions:
- (a) Prompt's exact reproducer: `parallel_run "exit 0" "exit 3"` bare under `set -euo pipefail` aborts the caller at the call site with script exit 1 (parallel_run's own `return`, not a pipeline abort) — matches the documented contract exactly.
- (b) Reverted the conditional restore to unconditional `set -e` and re-ran `scripts/test-lib-parallel.sh`: broke immediately (exit 1, only 2 of 36 assertions ran) — confirms the fixer's claim that unconditional restore breaks the `set +e`-bracketed call sites.
- (c) `$-` errexit detection verified correct under `/bin/bash` 3.2.57: caller's `-e` state after the call always matches what the caller itself set (`set +e` then explicit `set -e` -> `ehB`; `set +e` alone, no restore -> stays `hB`, not forced back on).
- (d) No other shell-option leakage: `set +e`/restore touches only errexit; `pipefail` and other flags are never modified by this code, confirmed by reading the diff and by direct `$-` inspection before/after calls.
- All three real call sites (`run-guard-tests.sh:88`, `test-check-installed-citations.sh:935,1023`) still bracket with `set +e`/`set -e` as before and pass in isolated harness runs.

**F4 (Minor, surviving mutant)** — HOLDS. Mutated the CHECK-phase loop to count `${#CMDS[@]}-1` down to `0` (reversed order, same case coverage) — harness failed with `FAIL: CHECK PHASE order broken: expected [0 1 2 ...] got [51 50 49 ...]`. Reverted; harness passes clean again (PASS, exit 0).

**No weakened tests, no new defects**: diffed all test files — every change is additive (new sub-cases, new assertions, private TMPDIR isolation); no existing assertion was loosened or removed. Ran `test-lib-parallel.sh` (36/36 ok), `test-check-installed-citations.sh` (all cases + new order assertion, exit 0), `test-run-guard-tests.sh` (all fixture cases, exit 0) individually via `/bin/bash` (3.2.57) — did not re-run the full 42-harness suite, per instructions.

Bash 3.2 floor respected: only indexed arrays (`labels=("$@")`, `CHECK_ORDER+=(...)`), no associative arrays, no `wait -n`.

`git status --short` in the worktree shows only pre-existing untracked spectre/docs artifacts — no leftover mutations from this review.
