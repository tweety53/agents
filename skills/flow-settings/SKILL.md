---
name: flow-settings
description: Read and change the harness-wide flow defaults — default model and reviewer slots. Standalone, not a pipeline stage. Use for /flow-settings.
allowed-tools: Bash(myflow:*), Bash(jq:*)
license: MIT
compatibility: Requires the myflow CLI and jq.
metadata:
  author: gymie
  version: "1.0"
---

Read and change the harness-wide settings record `myflow settings get`/`set` manage: the default
model (`defaultModel`) every `/flow` run's implementer, fixer and reviewer roles use unless a
session overrides it, and the reviewer slots (`reviewers`) the review panel dispatches by default.

**This is a standalone command, not a pipeline stage.** It takes no change name, reads and writes
no per-change state file, and marks no `myflow stage` call. It changes the harness-wide store, not
any one change's record.

**No flags.** Per **Command surface** (`skills/flow-contracts/pipeline.md`), no `/myflow-*` or
`/flow-*` command accepts a flag; that rule extends to this command as part of the same family. The
only input is the operator's answers to the questions this skill asks interactively.

**Announce at start:** "Using flow-settings."

## Workflow

### 1. Read current settings

```bash
CURRENT="$(myflow settings get)"
```

`myflow settings get` prints one line of JSON: `defaultModel` (a string) and `reviewers` (an array
of strings). A non-zero exit means the store could not be reached — report the CLI's stderr
verbatim and stop; there is no per-harness fallback file for this record the way a per-change state
file has one.

Print the current values plainly before asking anything:

```
Current flow settings:
  default model: sonnet
  reviewers:     <comma-separated list from Reviewers, verbatim>
```

### 2. Offer to change each field

Ask about only the two fields the settings store actually holds — `defaultModel` and `reviewers`.
Use **AskUserQuestion**, one field at a time, starting from the current value read in step 1:

- **Default model** — offer the harness's known model identifiers, read from
  `<agents repo>/stats/internal/store/settings.go`'s `ValidModels` map at the time this skill runs,
  plus "keep current". This is the one value every `/flow` run's implementer, fixer and reviewer
  roles default to; a session can still override it via plain-language instruction, recorded per-run
  rather than here.
- **Reviewers** — offer the exact reviewer-slot ids read from `<agents repo>/stats/internal/store/settings.go`'s
  `ValidReviewers` map at the time this skill runs, never a copy of that list written into this
  file — `ValidReviewers` is that store's own enum and the only place it is canonical. Three of its
  slots are dispatched by every `/flow` run by default; the remaining two are on-demand-only,
  dispatched only when the operator asks. Offer the full set as a multi-select seeded with the
  current list, plus "keep current".

If the operator keeps both fields unchanged, say so and stop — do not call `settings set` for a
no-op write.

### 3. Write the change

```bash
myflow settings set -model "<defaultModel>" -reviewers "<comma,separated,list>"
```

Both flags are required by the CLI even when only one field changed — `settings set` writes the
whole record, replacing what was recorded before. Pass the value just confirmed for the field that
changed, and the value read in step 1 for the field that did not.

A non-zero exit means the store rejected the write (an invalid model or reviewer name) or could not
be reached. Print the CLI's stderr verbatim — it names the specific bad value on a rejection — and
do not report success.

On success, report exactly what changed against the values read in step 1: which field(s) changed,
old value → new value. A field left unchanged is not mentioned as a change.

## Guardrails

- **Never** touch a per-change state file, `<project>/spectre/changes/`, or any `myflow
  state`/`myflow stage` call — this command's write is scoped to the harness-wide settings record
  alone.
- **Never** commit, stage, push, merge, or create a worktree or branch.
- **No flags** — the only input is the operator's interactive answers.
- **Never** call `settings set` with a value the operator did not just confirm, and never call it at
  all when nothing changed.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| View and change the harness-wide flow defaults | `/flow-settings` |
