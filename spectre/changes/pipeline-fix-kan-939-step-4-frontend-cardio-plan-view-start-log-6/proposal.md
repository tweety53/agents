# pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-6

## Why

The pipeline-fix loop (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`) and
the self-review fix loop (step 3 of `skills/flow-self-review/SKILL.md`) are prose only, and the
parent session of kan-939 skipped their steps: it dispatched fixers on `flow-high` — reserved for
the no-progress escalation — with no planner, and landed three docs fixes on the agents repo's
`main` (b8e0e063, bb273948, 3fd761da) with no review dispatch at all. Nothing recorded the loops'
dispatches: flowd holds no `pipeline-fix-*` or `self-review-*` row for any loop to date. The
operator ordered skipping made impossible, mechanically.

## What changes

- A new `PreToolUse` hook, `hooks/enforce-loop-dispatch.py`, beside `enforce-agent-baseline.py`,
  hands every Agent/Task call whose prompt carries a `Key: pipeline-fix-<k>-…` or
  `Key: self-review-…` line to a new Go guard, `scripts/check-loop-dispatch.sh`. The guard reads
  the loop's pairs from one table in `skills/flow-self-review/SKILL.md`, refuses a dispatch off
  its row (a `flow-high` one unless its escalation is recorded), one while an earlier loop
  dispatch is still open, a reused key, and any fix or review before a completed plan — and
  records the admitted dispatch in flowd under the loop's own change.
- The agents repo's existing landing-push guard (`hooks/protect-main-checkout.py`, which already
  refuses an unarchived change) also refuses a landing push of a `pipeline-fix-*` or
  `self-review-*` branch unless a new Go guard, `scripts/check-loop-landing.sh`, finds in flowd a
  completed plan, a completed fix, every row closed and a last review closed `clean`. Forward-only:
  the three commits already landed stay as history.
- A new `scripts/pipeline-fix.sh` runs the loops' fixed mechanics — `start`, `end`, `escalate`,
  `verify`, `land`, `drop` — and the two skill files cite it in place of the commands they
  restated.
