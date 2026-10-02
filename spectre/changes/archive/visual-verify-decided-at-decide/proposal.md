## Why

`VISUAL-VERIFY-MISSING` stopped integrate on changes with no visible surface (gymie kan-863,
kan-866), because the "is visual verification needed?" judgement was made nowhere visible. The full
account is `skills/flow/SKILL-rationale.md`'s "Decide, step 5 visual verification" entry.

## What changes

- **Decide** step 5 (`skills/flow/brainstorm-planner.md`) records `visual` in `decision.json`.
- `flow decision render` (`stats/cmd/flow/decision.go`) validates and renders it.
- **Visual verification** (`skills/flow/verify-and-handoff.md`) honours a `skipped` decision.
- `check-visual-verify-dispatched.sh` (its header is the contract) accepts it.
- `trsdLastClass` (`stats/internal/guard/taskreviewersingledispatch.go`) read the oldest decision's
  class — `flow record decisions` lists newest first. Renamed `trsdNewestClass`; it takes the first.
- `rules/agent-baseline.md` **Reporting back**: every noticed defect is fixed; outside the task,
  cheap and fast.
