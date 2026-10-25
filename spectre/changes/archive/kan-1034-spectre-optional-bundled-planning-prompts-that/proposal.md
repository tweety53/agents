# kan-1034-spectre-optional-bundled-planning-prompts-that

## Why

Flow's planning prompts — Explore and Plan agents, the four `design.md` plan sections, the writing-plans task shape, the plan guards' expectations and plan mode — live only in the flow skills (KAN-1033). A spectre user without flow gets the bare CLI and none of them.

## What changes

- spectre offers two modes: **tool only** (`go install` the binary) and **tool + planning prompts** (install the spectre plugin or copy its skill folders), which adds the `spectre-planning` skill.
- `spectre-planning` is the one source of the generic planning prompt text; flow loads it instead of carrying its own copy.
- In the prompts mode, planning enters plan mode on its own: Claude Code's `EnterPlanMode`, and Cursor's Plan mode (switched to, or asked of the user, before exploring).
- Flow-only parts — stage marks, Decide, `flow record`, Jira, `plan-agent-pair.sh`, mockups, withdrawal — stay in flow.

## Fix round 1

- The spectre manifest test's comment says `UPSTREAM.md` names its source commit; since integrate it names the skill's path on agents `main` and the ticket, so the comment is reworded to match.
