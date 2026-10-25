---
name: flow-planner-medium
description: Flow planner at medium effort — the self-review and in-run fix planner, which may dispatch fresh Explore and Plan agents (decision planner-carries-agent); the default for a scoped task with a stated plan; the dispatch prompt carries every instruction; pass `model` at dispatch time.
effort: medium
tools: Read, Glob, Grep, Bash, Write, Edit, ToolSearch, Agent
---
Flow planner at medium effort; it may dispatch fresh Explore and Plan agents, never a fork; the dispatch prompt carries every instruction.
