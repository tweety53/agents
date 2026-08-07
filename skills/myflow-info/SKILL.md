---
name: myflow-info
description: Explain the myflow pipeline — the three states, the three commands, and the shape of the finish stages — by reading skills/myflow-contracts/pipeline.md. Read-only reference. Use for /myflow-info.
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

**What this command no longer recites, and where it went.** `pipeline.md` still answers *what the
two runs are and which stages each has* — the level-1 stage table and the level-2 expansions of every
finish stage stayed in the core for exactly that reason. The **procedures** — the three preflight
signals, the unfinished-work courses, the landing routes, run 2's numbered steps and the
worktree-cleanup checks — moved to `skills/myflow-contracts/finish-contract.md`, which only
`/myflow-finish` loads.

**This command does not read that file**, and that is deliberate: it is the one read-only command in
the pipeline, and making it pay the full finish load would undo the split for the command that
benefits least. Name the file as where the detail lives, and stop there.

For a question about the state file's shape, the self-heal rules, project configuration, Jira, or
follow-up issues, read the matching contract file beside it rather than paraphrasing.

## What to say

Scale the answer to the question. A specific question ("when does it commit?") gets a specific
answer, not the whole pipeline.

For a general "how does this work", lead with the shape — the three-line command-to-state summary
under **States** (`skills/myflow-contracts/pipeline.md`), shown as it was read this invocation. This
skill carries no copy of it: the one it used to hold had already drifted from the contract's wording
about which `/myflow-finish` run is terminal, which is exactly what a frozen copy does and exactly
what the guardrail below forbids presenting.

When the question is about the flow itself — what runs when, and in what order — present what you
read rather than a summary of it: the state diagram and the level-1 stage table, one row per
command, under **Pipeline flow** (`skills/myflow-contracts/pipeline.md`). This skill deliberately
carries no copy of either.

A question about what one command does is answered from that command's row and the level-2
expansion of whichever stage the question is about — one expansion per stage that hides
substructure, in that same section.

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
- **Never** present a remembered diagram, stage table or state summary — show the ones read during
  this invocation, from **States** and from **Pipeline flow** (`skills/myflow-contracts/pipeline.md`).
  A block this command does not read is one it can never present, which is why this skill holds a
  copy of none of them.
- **No flags.** The only argument is an optional topic to explain.
