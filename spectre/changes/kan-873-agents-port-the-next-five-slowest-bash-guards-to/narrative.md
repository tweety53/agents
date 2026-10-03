# kan-873-agents-port-the-next-five-slowest-bash-guards-to — session narrative

## 2026-10-03 — creating run

- The operator asked for a Jira task and an unattended run "till the handoff", taking every
  recommended option. The run read that as the `IN_PROGRESS` handoff and continued past the
  `STARTED` plan gate into implementation in the same session.
- The base was red before any edit: `test-flow-addr-declaration.sh` failed on every run, because
  `79c2f025` added a second `-addr` registration in `flow lesson resolve`. `test-make-build.sh` also
  raced under `pipefail` (SIGPIPE from `grep -q` reading a pipe). Both were fixed in this change,
  as tasks 1 and 2, rather than worked around.
- The class was raised from regular to big partway through planning, once the five ports' parity
  surface was measured. The largest is `break-and-prove`, at 137 subtests against a floor of 38.
- A defect outside the five ports surfaced: under macOS `/bin/bash` 3.2, a shim whose
  `lib/flow-guard.sh` is missing exits 1 before its `|| { …; exit <code>; }` block runs. Task 10
  fixed the shim template in all 68 shims and added a test that runs each one without its library.
- Gated per-task reviews raised two Importants:
  - task 7: a symlinked target without a trailing slash;
  - task 6: the test read the operator's `~/.gitconfig`.

  Both were fixed and re-reviewed clean.
- Panel round 0 raised one Important and three Minors:
  - Important: a `--clean` command that backgrounds a child stalled each `break-and-prove` leg.
  - Minor: `prepare-workspace` looked its sibling up through a literal `scripts/`.
  - Minor: a dated audit file still cited a deleted harness.
  - Minor: a duplicated sort setup.

  One fix commit closed all four.
- **Time sink:** each reviewer's reproducers pinned their `# premise:` lines to the exact lines the
  fix rewrote, so the pinned re-runs could only refuse. They were re-authored without those
  premises and proved both ways with `prove-reproducer.sh` against the pre-fix commit. One
  reproducer (`0-P3-1.sh`) could never have passed as written: its body tested for the deleted file
  rather than the citation.
- **Record gap:** the first round-1 re-review dispatch was recorded without `-slot`. `flow record
  dispatch begin` does not update the slot on a repeat call with the same key, so
  `check-panel-findings-closed.sh` could not see the re-run. The re-review was dispatched again
  under the `-retry` key with the slot set.
- **Plan-record mismatch:** the fix subagent added the fix commit's paths to tasks 7 and 8's
  `**Files:**` fields. `check-task-records.sh` correctly refused this, because those tasks' own
  commits do not touch those paths. The additions were reverted, since the fix commit carries them.
- **Flakes:** the first verify run of `scripts/run-guard-tests.sh` failed in two harnesses this
  change does not touch, `test-run-guard-tests.sh` (case 2's summary count) and
  `test-check-done-when-paths.sh`. Both passed alone, and the whole suite passed on its one
  re-run. The cause was not investigated.

## 2026-10-03 — integrate run

- **Preflight:** `RUN1`; main checkout `STAGED-CLEAN` and `DRIFT-CLEAN`.
- **Unfinished-work gate:** `CLEAR`; visual verify not required (no UI paths).
- **Base:** `origin/main` had not moved since the recorded merge base — no rebase.
- **Route:** merge and push, from the project's configured default.
