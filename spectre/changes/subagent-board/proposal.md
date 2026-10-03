# subagent-board

## Why

Operator, 2026-10-03: the subagents a session runs should stay visible in the status-line shape
(`<emoji> Task <x>/<n> (<few words>) — <state>`), not only as the engine's own one-row-per-agent
footer, and a `/flow` run's dispatches should carry the plan's task numbers into that view —
task groups included.

## What changes

- `mods/subagent-board/`: a Claude Code function-hooks plugin. A band above the prompt lists up to
  five subagents as `<model-effort> <emoji> Task <x>/<n> (<desc>) — <state>`, a running row's emoji
  naming its kind (🔍 review, 🔨 fix, 👀 visual verification, 🧪 tests and lint); the band hides once
  none runs. The hint line under the prompt gains the change's tracker key, the `/flow` phase, the
  running stage's emoji and a per-emoji tally, read from `flow stage` marks.
- `setup.sh global`: links each `mods/<name>/` into `~/.claude/skills/`, where Claude Code
  auto-loads a plugin folder.
- `skills/flow/implement.md` **Dispatch sites**: every dispatch's Agent-tool `description` is
  `Task 3/22 (...)`, `Tasks 3+4+7/22 (...)` for a group, or the dispatch's key and a few words;
  `/flow-fast` follows it.
