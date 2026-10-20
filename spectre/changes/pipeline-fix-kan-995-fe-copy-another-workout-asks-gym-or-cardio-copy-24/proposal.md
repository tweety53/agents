# pipeline-fix-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy-24

## Why

gymie's kan-995 run (on `--base feature/cardio`) ended its final report on a "Still open:" list of
two items the operator then had to tell the session to do:

- The self-review fix planner recorded decision `k995-rows-left`
  (`spectre/changes/archive/self-review-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy/design.md`):
  the harvester fix went forward only and left three already mis-stamped ledger rows wrong, the
  repair framed as the operator's option. The repair needed one small store command, later landed as
  `b2555b98` (`flowd restamp-dispatch`).
- The self-review pass recorded "feature/cardio lacks develop's devRestart isolation fix 98da4aad;
  a branch decision for the operator — noted"
  (gymie `docs/self-review/kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy-self-review.md`
  on `origin/feature/cardio`, line 15), and the session relayed it as an open decision. The fix was a
  clean merge of develop into `feature/cardio`, verified green and pushed.

`skills/flow-self-review/SKILL.md` lets a pass record "an observation with nothing to fix" as
`noted`, and nothing stops an observation that does name an action from taking that route; step 3
bars a planner from dropping a finding but says nothing of a fix that leaves the defect's damage in
place; and `skills/flow/cleanup.md`'s no-open-item rule did not name an item left as the operator's
call.

## What changes

- **Step 2 of `/flow-self-review`:** an observation that names an action the run can take is a
  fixable finding, never `noted` and never an operator decision — unless the action is `big`, reaches
  production, or deletes what cannot be restored.
- **Step 3:** a fix repairs what its defect already wrote wrong, in the same plan; and **The base
  sync** — a `--base` change's `<base>` missing a commit the project's `<default-branch>` carries is
  synced by this session (clean merge, the repository's lint and test green, pushed without asking),
  and offered for filing like a `big` finding only when the merge conflicts or the run is red.
- **`land-self-review-report.sh`** refuses a `noted` finding line that leaves its action to the
  operator — `LAND-DEFERRED-FINDING`, exit 4, nothing committed — beside the existing
  `LAND-UNTRACED-FINDING`; `integrate.md` and `/flow-fast` treat it as they treat that refusal.
- **`skills/flow/cleanup.md`:** the Finished handoff's no-open-item rule names "left as the
  operator's call".
