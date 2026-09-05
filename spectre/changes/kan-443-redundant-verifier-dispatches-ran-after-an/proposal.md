# kan-443-redundant-verifier-dispatches-ran-after-an

## Why

KAN-443 (a `myflow-cost` self-review finding from KAN-423) states that three verifier dispatches
ran after an already-clean re-review where one would do, and asks for verifier dispatch to be
deduped or gated on the prior review round being clean.

The store rows for KAN-423 (`flow-postgres`, `dispatches` table, `dispatch_key` column, session
token `mf-kan423-1788562979`) do not support that reading:

| Seq | `dispatch_key` | Stage | Started (UTC) | Ended | Cache-read tokens |
|-----|----------------|-------|---------------|-------|-------------------|
| 31 | `verify-gymie-worktrees` | `flow.verify`, backend worktree: lint, test, `check-spec-reach.sh` | 09:07:08 | 09:13:37 | 298 217 |
| 32 | `verify-gymie-frontend-worktrees` | `flow.verify`, frontend worktree: lint, test | 09:07:08 | 09:13:37 | 423 432 |
| 33 | `visual-verify-gymie-worktrees` | `flow.visual-verify`, backend worktree: stack probe, Playwright spec authored and captured, PNGs read, `test:visual` diagnosed | 09:14:25 | 09:25:00 | 7 454 966 |

- Dispatches 31 and 32 are the "one verifier per worktree" rule of **The verifier dispatch**
  (`skills/flow/verify-and-handoff.md`); KAN-423 had two worktrees. Each ran a different
  project's lint and test commands.
- Dispatch 33 is `flow.visual-verify`, a different stage with different commands; its cost is the
  visual work itself (its report is `visual-verification.md` in the KAN-423 archive).
- No `verify-2` re-dispatch happened. No dispatch repeated another's checks.
- Nothing in the pipeline dispatches a fixed count of verifiers, and no verifier is triggered by
  the panel's outcome: verification always follows a clean panel by construction (KAN-433 moved
  lint and test out of the conductor into the verifier). "After an already-clean re-review" is
  the normal sequence, not a symptom. A gate on the prior round being clean would gate nothing.

The finding was produced by a self-review subagent reading the rendered ledger, where the three
rows are indistinguishable: `RenderLedger` (`stats/internal/records/render.go`) prints `Slot` and
`Diff base` but never `Dispatch.Key`, although the store holds `dispatch_key` for every row since
KAN-288. Three `Role: verifier` rows with only their token counts to tell them apart read as
three of the same thing. That is the defect this change fixes; the requested gate is not built,
and this proposal is where the original premise is closed — no Jira comment is posted.

## What changes

- The rendered SDD ledger carries a `- Key: <dispatch_key>` line per dispatch, placed after the
  `- Slot:` line and conditional on the key being non-empty, in the same shape `Slot` and `Diff
  base` already use. Rows recorded without a key render no `Key` line.
- One new Go test asserts the line renders for a keyed dispatch and is absent for a keyless one.
- Nothing changes in `skills/flow/verify-and-handoff.md`, `skills/flow/review-panel.md`, or any
  verifier dispatch rule.
