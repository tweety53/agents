# Design — pipeline-fix-kan-1060-flow-cost-visual-verify-now-costs-more-than-47

## Context

The failing assertion (`enforce_loop_dispatch_hook_test.go:267`, `incident calls = [], want 1`)
reads the argv log the `incidentStub` script writes. The hook runs that stub through
`record_incident`, `subprocess.run(..., timeout=10)`, and swallows `OSError`, `ValueError` and
`SubprocessError` — `TimeoutExpired` included.

**Cause, established.** Two candidates were tested on this machine (darwin 25.5):

- ETXTBSY (golang/go#22315) — **disproved here.** A script held open for write by the executing
  process itself still execs on Darwin: `subprocess.run(['./s'])` with `open('s','a')` held returns
  rc 0 and the script's output. Darwin does not raise "text file busy" for an exec of a file open
  for write, so the swallowed `OSError` cannot have come from it on this platform.
- Starvation past the 10 s bound — **confirmed by forcing it.** With `incidentStub`'s script
  starting `sleep 11` before it logs, one isolated
  `go test -count=1 -run TestEnforceLoopDispatchHookRecordsADenial ./internal/repotest/` fails with
  the production message, `checker rc 1: incident calls = [], want 1 (out={"hookSpecificOutput"…`.
  With the fix below in place and the same `sleep 11`, the run passes; removing only the
  `FLOW_HOOK_UNBOUNDED=1` injection from `runHook` fails it again. Under the full `-race` suite the
  stub's `/bin/sh` is starved past 10 s, killed, and the swallow hides it.

The same exposure exists in the baseline hook's incident cases
(`enforce_agent_baseline_hook_test.go:559`, `:600`), which call `runHook` without
`FLOW_HOOK_UNBOUNDED=1`; kan-1033 added the seam to that hook but passed it only from its
store-stub callers.

## Decisions

**ID:** cause-is-the-bound
**Status:** active
**Chosen:** the 10 s `record_incident` bound under CPU starvation, forced and reproduced as above.
**Considered:** ETXTBSY from parallel `os.WriteFile` + fork — rejected: it does not occur on Darwin
(checked), where the failure was seen.

**ID:** lift-bound-via-existing-seam
**Status:** active
**Chosen:** reuse kan-1033's seam — `bound()` in the loop hook, `FLOW_HOOK_UNBOUNDED=1` injected by
`runHook` for every hook test. The bound is production's fail-open, not under test; lifting it in
tests removes wall-clock time from every hook case — a deterministic clock, not a wider tolerance.
**Considered:** a longer timeout — rejected by `fix-determinism-at-the-source`. Executing stubs via
`/bin/sh <script>` or writing them before parallel subtests fork — rejected: they address ETXTBSY,
which is disproved. A shared Python module for `bound()` — rejected: a three-line function in two
standalone hook scripts.

**ID:** seam-in-runhook-not-per-call
**Status:** active
**Chosen:** inject in `runHook`'s base env so every hook case is covered, including the baseline
hook's two incident cases that lack it today. The existing per-call `FLOW_HOOK_UNBOUNDED=1`
arguments become redundant but harmless; left in place to keep the diff minimal.

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — a 22-line diff reusing an existing seam, with the cause forced and the
proof recorded.
**Considered:** `flow-medium` — the hook ships to every session, but production never sets the
variable, so its runtime behaviour is unchanged.

**ID:** fixer-grouping
**Status:** active
**Chosen:** one fixer for task 1, one commit, `opus` on `flow-low`: two files, one mechanism.

## Open questions

None.
