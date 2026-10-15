# self-review-drop-rule

## Why

kan-982's self-review planner dropped two real findings (a `Task-Id:` guard, a commit-scope check) because its first design would misfire, and the parent offered them for filing; the operator answered "fix them now" and asked that the pipeline prevent this.

## What changes

`/flow-self-review` step 3: the planner never drops a non-`big` finding over an imprecise first design — the objection becomes a design constraint. A finding stays unfixed only when `big` or disproved with evidence; any other drop is re-planned, never offered for filing.
