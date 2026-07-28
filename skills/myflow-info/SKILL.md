---
name: myflow-info
description: Explain the myflow pipeline — the three states, the three commands, and the finish contract — by reading skills/myflow-contracts/pipeline.md. Read-only reference. Use for /myflow-info.
allowed-tools: Bash(cat:*), Read
license: MIT
compatibility: Requires skills/myflow-contracts/pipeline.md.
metadata:
  author: gymie
  version: "2.0"
---

Explain the myflow pipeline: the states, who acts at each gate, and which command does what.
**Read-only** — touches no change, no git state, no files.

**Announce at start:** "Using myflow-info."

## Source of truth — do not embed a copy

This skill holds **presentation instructions only**. The pipeline content lives in
`skills/myflow-contracts/pipeline.md` and must be read at invocation time so this command can never
drift out of date:

```bash
# global install first, then a project-local install as fallback
cat ~/.claude/skills/myflow-contracts/pipeline.md 2>/dev/null \
  || cat ~/.cursor/skills/myflow-contracts/pipeline.md 2>/dev/null \
  || cat skills/myflow-contracts/pipeline.md
```

**Never answer from memory.** The pipeline is versioned, and this command exists precisely so the
answer reflects the installed contract rather than a remembered one. If none of those paths
resolve, say so and stop — a confidently remembered pipeline is the failure this command prevents.

For a question about the state file's shape, the self-heal rules, project configuration, or Jira,
read the matching contract file beside it rather than paraphrasing.

## What to say

Scale the answer to the question. A specific question ("when does it commit?") gets a specific
answer, not the whole pipeline.

For a general "how does this work", lead with the shape:

```text
/myflow-start  → STARTED      you: read the proposal artifact
/myflow-do     → IN_PROGRESS  you: review the staged diff and run the apps
/myflow-finish → FINISHED     terminal (it integrates on its first run)
```

Then the points that are load-bearing and least guessable:

- **Each command ends in the state named after it.** The human gate is a property of the state, so
  no command exists whose only job is to record that a review happened.
- **`/myflow-do` emits both the staged diff and the manual test guide**, so reviewing and testing
  are one sitting.
- **Every command is re-entrant.** Re-run `/myflow-start` to revise the proposal, `/myflow-do` to
  fix something. A fix never moves the state.
- **`/myflow-finish` runs twice** — once to integrate (open a PR by default, merge, or leave it to
  you), and again once the branch is merged, to sync specs, archive, push, and remove the
  worktrees.
- **No command takes a flag.** The only argument is the optional change name.
- **Nothing runs tests or linters before integration** — that happened during `/myflow-do`.

Also available: `/myflow-status` for where the open changes actually are.

## Guardrails

- **Never** write, stage, or commit anything.
- **Never** advance or repair a change's state — that is not this command's job.
- **Never** describe a state, command, or flag that is not in `pipeline.md`.
- **No flags.** The only argument is an optional topic to explain.
