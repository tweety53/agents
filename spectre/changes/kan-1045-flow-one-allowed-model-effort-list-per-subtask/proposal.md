# kan-1045-flow-one-allowed-model-effort-list-per-subtask

## Why

Which model and effort a flow dispatch may run on is stated in about ten places across three
languages — `PINNED_ROLES`, `FLOW_HIGH_MARKERS` and `FLOW_XHIGH_TYPES` in
`hooks/enforce-agent-baseline.py`, the self-review pairs table `stats/internal/guard/loop.go`
parses, the constants in `planagentpair.go` and `planclass.go`'s `pcExtra`, and prose in
brainstorm-planner.md, review-panel.md, visual-verify*.md, implement.md, pipeline.md,
review-panel-late-fix.md and flow-self-review/SKILL.md. Several roles (the merger, the brainstorm
explorer and planner, the self-review re-review) appear in no list at all, and most dispatches are
checked against nothing (KAN-1045).

## What changes

- One table, **Allowed pairs** in `skills/flow-contracts/model-policy.md`: role id → allowed
  `<model>/<subagent_type>` pairs, with escalations as rows that require a prompt marker.
- Every flow dispatch prompt carries a `Role: <role id>` line.
- A new Go guard, `check-dispatch-pair`, admits or refuses a dispatch against the table; the
  agent-baseline hook calls it for every dispatch in a flow run and refuses a missing `Role:` line,
  an unknown role, a pair outside its role's rows, or a marked row without its marker.
- `check-loop-dispatch`, `plan-agent-pair` and `plan-class`'s extra dispatches read the table
  instead of hard-coding pairs; the hook's three hard-coded rules and the self-review pairs table
  are deleted.
- Every prose statement of a role's pair becomes a citation of **Allowed pairs**.
