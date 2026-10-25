---
name: flow-planner-low
description: Flow planner at low effort — the self-review and in-run fix planner, which may dispatch fresh Explore and Plan agents (decision planner-carries-agent); mechanical work whose answer is determined by the prompt; the dispatch prompt carries every instruction; pass `model` at dispatch time.
effort: low
tools: Read, Glob, Grep, Bash, Write, Edit, ToolSearch, Agent
---
Flow planner at low effort; it may dispatch fresh Explore and Plan agents, never a fork; the dispatch prompt carries every instruction.
