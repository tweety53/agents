# self-review-kan-1048-rework-live-board-and-run-view

## Why

The self-review pass of `kan-1048-rework-live-board-and-run-view` found seven fixable defects:

- 1-A: `check-verify-green`'s `^verify-(fix-)?[0-9]+` also matches a plain `verify-<repo>` key whose
  repository name starts with a digit — `verify-3d-viewer` groups as `verifyd-viewer`, so a blocked
  `verify-3d-viewer` is never superseded by its re-run.
- 1-B: `TestRunVerifyListShimNamesItsSiblings` covers `runverifylist.go` alone; every other Go guard
  that execs a sibling script (`rebuildfullappsuitezip.go`, `integratechange*.go`, …) can drop its
  shim's `$SCRIPT_DIR/<x>.sh` line unnoticed, and `setup.sh` then never installs the sibling.
- 1-D: a mutation reproducer whose edit fails exits 0 on unmutated code, which the runner reads as a
  surviving mutant; four of kan-1048's seven carried no landing check, and the
  `skills/flow/review-panel.md` rule that every mutation confirms its edit landed is prose alone.
- 1-E: `check-plan-shape` accepted an `**Allowed-collateral:**` entry under
  `spectre/changes/<name>/visual-verification/`, which `check-task-commit-planning-paths` refuses in
  a task commit; `skills/flow/verify-fix-loop.md` step 1 never says where the evidence directory goes.
- 1-F: `flowd` stamped kan-1048's round-0 panel ends at its own close time, since the running daemon
  predated kan-1044's restamp fix; 19 rows dated 2026-10-28 carry such an end.
- 1-J: `hooks/flow-session-activity.py` maps a subagent's `PostToolUse` — fired under the parent's
  `session_id` — to `working`, overwriting an open `question`.
- 8-A: visual-verify reported a one-frame tooltip linger off a single screencast frame that never
  reproduced, costing a whole fix round.

## What changes

- The re-run key regex anchors its number with `(-|$)`.
- One repotest test checks every Go guard's sibling execs against its shim.
- `run-reproducer.sh` refuses a declared mutation reproducer that does not run under `set -e`; the
  panel prose and the test-sanity brief say so.
- `check-plan-shape` refuses a `**Files:**` or `**Allowed-collateral:**` entry under a planning path
  other than `visual-verification.md`; the verify-fix loop names where the evidence directory goes.
- `flowd` startup restamps stamped ends over the last seven days; `flowd restamp-ends` repairs the
  dev store's 2026-10-28 rows.
- The activity hook ignores tool events carrying `agent_id`.
- The verifier confirms a departure one frame long before reporting it.
