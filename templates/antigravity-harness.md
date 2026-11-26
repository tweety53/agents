## Antigravity Harness Mapping

The `/flow*` skills and contracts under `~/.gemini/config/plugins/flow/skills/` are authored once
and shared across harnesses. When running in Antigravity, apply this native tool, subagent, and
path mapping wherever a skill or contract names a Claude Code primitive:

| Skill / contract reference | Antigravity native equivalent |
|---|---|
| `Bash` | `run_command` |
| `Read`, `Write`, `Edit`, `MultiEdit` | `read_file`, `write_file`, `edit_file` |
| `AskUserQuestion` | `ask_question` (multiple-choice) or plain text (yes/no) |
| `Skill` tool / `/flow*` slash command | `read_file` on `~/.gemini/config/plugins/flow/skills/<name>/SKILL.md` (load `flow-contracts/pipeline.md` first for `/flow`) |
| `Agent` / `Task` (write/command roles: `flow-{low,medium,high,xhigh}`, `general-purpose`) | `invoke_subagent` with `TypeName: "self"` (or a `define_subagent` with `EnableWriteTools: true`) |
| `Agent` / `Task` (read-only roles: `Explore`, `Plan`, `flow-planner-*`, read-only reviewers) | `invoke_subagent` with `TypeName: "research"` |
| Effort / model tier (`low`, `medium`, `high`, `xhigh`) | `Model`: `low` → `"flash"`, `medium` → `"inherit"`, `high`/`xhigh` → `"pro"` (or `"inherit"` if already on that tier) |
| `subagent_type: "fork"` | Not used in `/flow*` runs — always spawn a fresh subagent with explicit file paths in `Prompt` |
| `~/.claude/rules/agent-baseline.md` | `~/.gemini/config/plugins/flow/rules-full/agent-baseline.md` |
| `~/.claude/rules/<name>.md` | `~/.gemini/config/plugins/flow/rules-full/<name>.md` |
| `~/.claude/skills/<name>/...` | `~/.gemini/config/plugins/flow/skills/<name>/...` |
| `<project>/CLAUDE.md` | `<project>/AGENTS.md` (also read `<project>/CLAUDE.md` if present) |

### Subagent baseline propagation

Include these two sentences verbatim in every `invoke_subagent` `Prompt`:

> Before anything else, read `~/.gemini/config/plugins/flow/rules-full/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.
