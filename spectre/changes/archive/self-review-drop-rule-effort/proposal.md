# self-review-drop-rule-effort

## Why

Operator, 2026-10-15: a self-review finding that is hard to solve may raise the planner to high effort; xhigh needs the operator's approval first.

## What changes

`/flow-self-review` step 3: a re-plan may run on `flow-high`; an `xhigh` re-plan is asked first, and a refusal leaves the finding offered for filing.
