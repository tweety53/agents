**Findings**

None at high confidence.

**Investigation notes (not findings):**
- `git -C <worktree> status --porcelain scripts/` and `git diff scripts/` both clean — no concurrent guard-script mutation during this pass.
- Verified `TEST_GIT_SHIM_REAL_GIT` is captured at source time, before either test file touches `PATH`, and confirmed via the new chaining case (13) that a shim built while a prior shim sits on `PATH` still execs real git, not the chained shim — reproducer: `bash scripts/test-check-base-moved.sh` (case 13 passes).
- Checked every `assert_shim_fired` call site: `SHIM_DIR` is always a fresh `mktemp -d` per case (no stale-sentinel reuse across cases), and `OUT=$(...)`/`RC=$?` capture is synchronous, so the sentinel is never asserted before the guard subprocess has exited.
- Checked match-arg uniqueness for all `shim_failing_git` calls (`status`, `merge-base`, `HEAD^{commit}`, `rev-list`, `--cached`, `${RECORDED_BASE}..HEAD`) against every git invocation in `check-finish-preflight.sh`/`check-base-moved.sh` — each matches exactly one call site, no cross-firing.
- Under one batch of ~10-12 duplicate concurrent full-suite runs I did see `assert_shim_fired`/unrelated `merge-base` failures (exit-1-vs-other-nonzero flakiness in an unrelated case with no shim at all), but this did not reproduce across three follow-up batches of 30 runs each (90 runs total, 0 failures) once system load settled — consistent with transient contention from other panel slots on this shared machine at that moment, not a defect in this delta. Flagging as calibration-consistent noise per the brief's guidance, not reporting as a finding since it has no reliable reproducer.

VERDICT: clean
