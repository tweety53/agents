# self-review-kan-979-tune-failure-modes-prompt

## Why

The self-review pass over kan-979 found three `flow-fix` defects, none `big`:

- An `In-run pipeline fix:` narrative line can land off the format **Pipeline defects found mid-run**
  (`skills/flow-contracts/pipeline.md`) fixes, carrying no sha and no blast radius, so the pass cannot
  read the fix's sha off the bundle. Nothing checks the line.
- `flow record summary` exists but no skill, script, hook or rule calls it, so every self-review bundle's
  first source is `skipped: change summary (absent)`.
- A fix run's implementer dispatch, review, pass rows and findings can all go unrecorded on a combined
  "fix, then land" invocation, and integrate lands the change anyway.

## What changes

- `integrate-change.sh commit` stops on `STOP: narrative` when an `In-run pipeline fix:` line is off the
  format, before reshaping anything; `skills/flow/integrate.md` names the stop.
- `skills/flow-contracts/pipeline.md` records every printed summary with `flow record summary`;
  `skills/flow/verify-and-handoff.md` cites it at the `IN_PROGRESS` block.
- `check-unfinished-work` reports OUTSTANDING when a ticked task has no implementer dispatch row, so
  integrate's unfinished-work gate catches an unrecorded fix round.
