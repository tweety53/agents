# Work end to end — rationale

The incidents behind `rules/work-end-to-end.mdc`, moved here verbatim. `setup.sh` does not install
this file, and no dispatched agent reads it.

## The incidents behind it

- kan-1037 (2026-10-27): the user spent about 380k uncached tokens answering mid-run questions with
  the option already marked recommended.
- kan-1038 (2026-10-27): asked whether to fix all eleven self-review findings or only the blocker,
  the user answered "recommended — don't ask such things again".
- KAN-1039 (2026-10-27), the user's own statement: the only thing they care about is well-working
  apps and tooling end to end, no matter how many fix rounds or tooling updates and fixes are
  needed; the only interactive parts are planning and manual testing of the implementation with the
  fixes that follow, and that only when they explicitly asked to stop after implementation.
- kan-1044 (2026-10-28): the run stopped on two contradicting guards (`check-plan-shape` F21 against
  `check-task-commit-planning-paths`) to ask whether to fix the `agents` repo; the user answered
  "resolve the conflict yourself" and asked that such stops never happen again.
