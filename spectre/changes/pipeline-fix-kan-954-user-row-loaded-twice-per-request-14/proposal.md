# pipeline-fix-kan-954-user-row-loaded-twice-per-request-14

## Why

A loop planner's `first-review-effort` decision is silently ignored when it writes the effort the
way `skills/flow-self-review/SKILL.md` step 3 item 1 names it — "the first review's effort,
`medium` or `low`". `lpFirstReviewType` (`stats/internal/guard/loop.go`) matches only a backticked
`flow-low` or `flow-medium` on the decision's `**Chosen:**` line, and on anything else falls back to
`flow-medium`. In this session two planners wrote `**Chosen:** low — …` (gymie
`self-review-kan-954-user-row-loaded-twice-per-request`) and `**Chosen:** the first review runs at
effort` followed by a backticked `low` (`pipeline-fix-kan-954-user-row-loaded-twice-per-request-13`),
and `scripts/check-loop-dispatch.sh` refused each `flow-low` review dispatch with `PAIR — … runs on
flow-medium, as the loop's design.md first-review-effort decision records (flow-medium when none)`.
The same misread already sent two archived self-reviews' recorded `low` reviews to `flow-medium`
(`self-review-kan-980-no-flow-roles`, `self-review-kan-959-call-efficiency-skill-and-reviewer-slot`).

## What changes

- `lpFirstReviewType` reads the effort from the `first-review-effort` entry alone — up to the next
  `**ID:**` — taking a bare `low` or `medium` that opens the `**Chosen:**` value, else the line's
  first backticked `low`, `medium`, `flow-low` or `flow-medium`, mapped to `flow-low`/`flow-medium`.
- A recorded `first-review-effort` decision whose `**Chosen:**` line names no effort — or that has
  no `**Chosen:**` line — refuses the `review-1` dispatch with a new `EFFORT` refusal naming the
  forms to write, instead of silently running it on `flow-medium`. A plan with no such decision
  still runs `review-1` on `flow-medium`, as the skill states.
- `scripts/check-loop-dispatch.sh`'s header lists the `EFFORT` refusal.
- `TestCheckLoopDispatchFirstReviewEffort` covers both forms, the precedence between them, the
  refusal, and an entry with no `**Chosen:**` line.
