# Final review panel — kan-100-myflow-get-rid-of-staging-use-commits

**Roster:** light (required: Primary, Principles, Code review (low))
**Optional slots:** none — no triggers fired requiring Security/Adversarial/Lens B/Lens C beyond
what the required three already covered; this diff is documentation/contract-editing plus two
Python guard scripts with no auth, migrations, or concurrency surface.

## Pass 1 (full roster)

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Primary | Important | `openspec/changes/kan-100-myflow-get-rid-of-staging-use-commits/tasks.md` (tasks 3.3/3.4) | Stale `Tests:`/`Baseline:` field numbering vs. real 14-case (at the time) test harness |
| F2 | Primary | Important | `openspec/changes/kan-100-myflow-get-rid-of-staging-use-commits/tasks.md` (task 6.2) | Checked-off task contradicted its own literal checkbox text (guard never added to `## lint`, correctly) |
| F3 | Code review (low) | Important (reported as "medium") | `scripts/check-task-commit-fields.py:515` (`parse_task_fields`) | No fence-tracking; a field-looking line inside a fenced example block would be misparsed as real data |
| F4 | Principles | Important | `scripts/check-task-commit-fields.py` (`_commit_reverted`) | Initial `git revert` call ran outside the guarded region; a failed revert skipped cleanup |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

## Fix round 1 (targeted — one fix subagent, all 4 findings)

- F1: corrected tasks 3.3/3.4's `Tests:`/`Baseline:` fields to match the real (by then 14-case)
  harness, and again after F3/F4 added two more cases (final: 16).
- F2: rewrote task 6.2's checkboxes and `Verify:` line to state the corrected outcome honestly
  (guard not added to `## lint`, why, and where it's actually covered).
- F3: added fence-tracking to `parse_task_fields` (reusing the existing `FENCE_LINE_RE`), with a new
  Case 15 proving a field-looking line inside a fence is not parsed as real data.
- F4: moved the initial `git revert` call inside a guarded region so a failed revert also triggers
  cleanup before re-raising, with a new Case 16 forcing a revert conflict and asserting the worktree
  is left clean.

Re-review (scoped, one dispatch) verified all four ADDRESSED, no new breakage, re-ran
`scripts/test-check-task-commit-fields.sh` (16/16), `scripts/test-check-task-build-green.sh`
(17/17), `scripts/check-task-build-green.sh` (exit 0) directly rather than trusting the fix report.

## Result

**Clean — zero open findings at any severity.**
