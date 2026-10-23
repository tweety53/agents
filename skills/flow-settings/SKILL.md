---
name: flow-settings
description: Read and change the harness-wide flow defaults — the reviewer slots and the learning note. Standalone, not a pipeline stage. Use for /flow-settings.
allowed-tools: Bash(flow:*), Bash(jq:*)
license: MIT
compatibility: Requires the flow CLI and jq.
---

Read and change the harness-wide settings record `flow settings get`/`set` manage: the reviewer
slots (`reviewers`) the review panel dispatches by default, and `learning` — whether an
implementation run prints a learning note (**Learning**, `skills/flow-contracts/learning.md`). Models are not a setting — the Decide
step chooses every dispatch's model (**Model and effort**, `skills/flow/brainstorm-planner.md`).

**This is a standalone command, not a pipeline stage.** It takes no change name, reads and writes
no per-change state file, and marks no `flow stage` call. It changes the harness-wide store, not
any one change's record.

**No flags.** Per **Command surface** (`skills/flow-contracts/pipeline.md`), no `/flow*` command
accepts a flag; that rule extends to this command as part of the same family. The
only input is the operator's answers to the questions this skill asks interactively.

**Announce at start:** "Using flow-settings."

## Workflow

### 1. Read current settings

```bash
CURRENT="$(flow settings get)"
```

`flow settings get` prints one line of JSON: `reviewers` (an array of strings) and `learning` (a
boolean). A non-zero exit means the store could not be reached — report the CLI's stderr verbatim and stop;
there is no per-harness fallback file for this record the way a per-change state file has one.

Print the current values plainly before asking anything:

```
Current flow settings:
  reviewers:  <comma-separated list from Reviewers, verbatim>
  learning:   <on | off>
```

Then run step 3's conductor-depth check against the **stored** roster and print its line when it
fires — a roster saved before the check existed surfaces its gap here too, not only when the
operator changes something.

### 2. Offer to change the reviewers and the learning note

Ask about both fields the settings store holds with **AskUserQuestion**, starting from the current
values read in step 1. **Learning** — on or off, plus "keep current".

- **Reviewers** — offer the exact reviewer-slot ids read from `<agents repo>/stats/internal/store/settings.go`'s
  `ValidReviewers` map at the time this skill runs, never a copy of that list written into this
  file — `ValidReviewers` is that store's own enum and the only place it is canonical. **This list is
  the review panel**: every id it holds is dispatched by every `/flow` run until this command changes
  it again, none held back as a fixed floor — resolution is canonical in `skills/flow/SKILL.md`'s
  Model resolution, dispatch in `skills/flow/review-panel.md`'s roster. A per-run operator
  instruction can add a slot for a single run without touching this list; an id's presence here is
  what makes every run dispatch it. Offer the full
  set as a multi-select seeded with the current list, plus "keep current", and say plainly when
  asking that the selection made here becomes every subsequent run's panel, not just this one. The
  CLI never sends an empty roster: an empty `-reviewers` is read as "not given", so selecting zero
  slots with nothing else changed fails at step 3 with exit 2 (`<agents repo>/stats/cmd/flow/settings.go`'s
  `needs -reviewers, -learning or both` check), distinct from the store's own exit-1 rejection —
  warn the operator before they try it that zero slots does not turn review off.
If the operator keeps both unchanged, say so and stop — do not call `settings set` for a no-op
write. Silence on the ask keeps the current values — the seeded selection, unchanged (**Unanswered
mid-run asks**, `skills/flow-contracts/pipeline.md`).

### 3. Write the change

```bash
flow settings set [-reviewers "<comma,separated,list>"] [-learning on|off]
```

Pass only the fields the operator changed: `settings set` reads the current record, overrides the
given fields and writes it back, so the other field is kept.

**Before the write, tell the operator which slots cannot dispatch as themselves here.** The
roster's slots spawn as **The roster** (`skills/flow/review-panel.md`) spawns them — on
`flow-low`, or on `flow-<effort>` where the panel is decided — and a conductor whose harness
offers neither substitutes a harness-provided type (`general-purpose`) for those slots. Compare
each confirmed slot's spawn type against the agent types this session's own harness offers at
conductor depth, and when the type is absent print one line ahead of the write, naming every
affected slot in roster order:

```
⚠ roster: no `flow-low` agent type at conductor depth — primary, principles will be substituted (general-purpose) at panel time
```

The write proceeds either way: never a gate. Never
block the save on it and never drop a slot to silence it.

A non-zero exit means the store rejected the write (an invalid reviewer name) or could not
be reached. Print the CLI's stderr verbatim — it names the specific bad value on a rejection — and
do not report success.

On success, report exactly what changed against the values read in step 1: old → new, per field.

## Guardrails

- **Never** commit, stage, push, merge, or create a worktree or branch.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| View and change the harness-wide flow defaults | `/flow-settings` |
