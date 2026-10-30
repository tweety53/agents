# pipeline-fix-kan-1061-flow-add-flow-interactive-fast-conversational-45

## Why

`scripts/plan-agent-pair.sh <seed> <n>` (Go port `stats/internal/guard/planagentpair.go`) prints
the first pair of the `plan-designer` row of **Allowed pairs** for every planning dispatch, while
**Explore and design agents** (`skills/flow/brainstorm-planner.md`) sends both an Explore
dispatch (`Role: explorer`) and a Plan dispatch (`Role: plan-designer`) to it for their pair, and
`hooks/enforce-agent-baseline.py` admits each on its own Role's row. Under open experiment 1
(`plan-designer-xhigh-explorer-low`) the script printed `opus xhigh` for an Explore dispatch and
the hook refused it with `DISPATCH-PAIR-REFUSED: explorer runs on sonnet/Explore@low`: script and
hook contradict, and no Explore dispatch can follow the contract while the two rows differ.

## What changes

- `plan-agent-pair.sh` takes a third argument, `<role>` — `explorer` or `plan-designer` — and
  prints the first pair of that row, an open experiment's override of that row applied, exactly
  the row the hook admits the dispatch on. Any other role, or two arguments, is a usage error,
  exit 2.
- Every contract citing its usage names the role: `skills/flow/brainstorm-planner.md`,
  `skills/flow/brainstorm.md`, `skills/flow-plan/SKILL.md` and the PLANNING DELEGATION paragraph
  of `skills/flow-self-review/SKILL.md`.
- `stats/internal/guard/plan_agent_pair_test.go` pins experiment 1's overrides: `explorer` prints
  `sonnet low`, `plan-designer` `opus xhigh`.
