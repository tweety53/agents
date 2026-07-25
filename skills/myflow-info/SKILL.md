---
name: myflow-info
description: Explain the myflow pipeline — stages, gates, commands, and flags — by reading the current rule file. Read-only reference. Use for /myflow-info.
allowed-tools: Bash(cat:*), Read
license: MIT
compatibility: Requires rules/myflow-manual-review.mdc.
metadata:
  author: gymie
  version: "1.0"
---

Explain the myflow pipeline: stages, who acts at each gate, which command advances what, and the available flags. **Read-only** — touches no change, no git state, no files.

**Announce at start:** "Using myflow-info."

## Source of truth — do not embed a copy

This skill holds **presentation instructions only**. The pipeline content lives in the rule file and must be read at invocation time so this command can never drift out of date:

```bash
cat .cursor/rules/myflow-manual-review.mdc
```

Also read the agents-data README for the Basic Workflow step map and the typical-flow walkthrough — on this machine `/Users/tweety53/Projects/agents/README.md` (**machine-specific**: it is the sync target of `.cursor/hooks/sync-agents-data.sh`, overridable via `AGENTS_DATA`; if that path does not exist, skip it and use the rule file alone rather than failing).

**Never** answer from memory, and **never** paste a hardcoded pipeline table into this skill file. If the rule file is missing, say so and stop — do not reconstruct it.

## Workflow

### 1. Read the sources

Read the rule file's **Pipeline stages**, **State file**, **Stage transitions**, gate sections, and **Opt-out** list.

### 2. Render the overview (no argument)

```
## myflow pipeline

start → do → manual review (Gate B) → manual test (Gate C) → review → PR review (Gate D) → finish (Gate E)

### Stages
<the Pipeline stages table, from the rule file>

### Who acts at each gate
| Gate | Who | What |
|------|-----|------|
| A | You | Approve the plan after /myflow-start (`/myflow-start-done` or `/myflow-start-fix`) |
| B | You | Review the staged diff in the worktree IDE (`/myflow-do-manual-review`, then `/myflow-do-done`) |
| C | You | Run the apps and work the manual-test checklist (`/myflow-manual-test-done`) |
| D | You | Review the PR on the forge and **merge it** — never automated (`/myflow-review-done`) |
| E | Agent | /myflow-finish verifies the merge, syncs specs, archives |

### Commands
<the Stage transitions table, plus /myflow-status and /myflow-info>

### Flags
<the Opt-out list>
```

Close with a one-line pointer: "Run `/myflow-status` to see where your changes actually are."

### 3. Detail view (argument given)

`/myflow-info <stage-or-command>` — print only that stage or command: what it requires, what it does, what it writes to the state file, and what runs next. Match loosely (`do-fix`, `/myflow-do-fix`, and `awaiting-do-review` all resolve). Unrecognized argument → list the valid stage and command names.

## Guardrails

- **Never** modify anything — no files, no git, no state.
- **Never** answer from memory; always re-read the rule file first.
- **Never** duplicate the pipeline tables into this skill file.
- Keep the output scannable — tables over prose.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Explain the whole pipeline | `/myflow-info` |
| Explain one stage or command | `/myflow-info <stage-or-command>` |
| See where changes actually are | `/myflow-status` |
